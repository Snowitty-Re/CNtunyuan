package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	fileService *service.FileAppService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(fileService *service.FileAppService) *UploadHandler {
	return &UploadHandler{fileService: fileService}
}

// RegisterRoutes 注册路由
func (h *UploadHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	upload := router.Group("/upload")
	upload.Use(authMiddleware.Required())
	{
		upload.POST("", h.Upload)
		upload.POST("/batch", h.UploadBatch)
		upload.GET("/:id", h.GetByID)
		upload.GET("/:id/download", h.Download)
		upload.GET("/entity/:type/:id", h.GetFilesByEntity)
		upload.DELETE("/:id", h.Delete)
		upload.PUT("/:id/bind", h.BindToEntity)
		upload.GET("/stats", h.GetStats)
	}
}

// Upload 单文件上传
func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	userID := middleware.GetUserID(c)

	resp, err := h.fileService.UploadFile(c.Request.Context(), file, header, userID)
	if err != nil {
		switch err {
		case service.ErrFileTooLarge:
			response.BadRequest(c, "file too large")
		default:
			logger.Error("Failed to upload file", logger.Err(err))
			response.InternalServerError(c, "failed to upload file")
		}
		return
	}

	response.Success(c, resp)
}

// UploadBatch 批量上传
func (h *UploadHandler) UploadBatch(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "invalid form data")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.BadRequest(c, "files are required")
		return
	}

	userID := middleware.GetUserID(c)
	fileReaders := make([]multipart.File, 0, len(files))
	fileHeaders := make([]*multipart.FileHeader, 0, len(files))

	for _, header := range files {
		f, err := header.Open()
		if err != nil {
			response.InternalServerError(c, "failed to read file")
			return
		}
		defer f.Close()
		fileReaders = append(fileReaders, f)
		fileHeaders = append(fileHeaders, header)
	}

	responses, err := h.fileService.UploadFiles(c.Request.Context(), fileReaders, fileHeaders, userID)
	if err != nil {
		logger.Error("Failed to upload files", logger.Err(err))
		response.InternalServerError(c, "failed to upload files")
		return
	}

	response.Success(c, dto.UploadResponse{Files: responses})
}

// GetByID 获取文件信息
func (h *UploadHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "file id is required")
		return
	}

	resp, err := h.fileService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrFileNotFound {
			response.NotFound(c, "file not found")
			return
		}
		response.InternalServerError(c, "failed to get file")
		return
	}

	response.Success(c, resp)
}

// Download 下载文件
func (h *UploadHandler) Download(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "file id is required")
		return
	}

	reader, file, err := h.fileService.GetFile(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrFileNotFound {
			response.NotFound(c, "file not found")
			return
		}
		response.InternalServerError(c, "failed to get file")
		return
	}
	defer reader.Close()

	// 设置下载头
	filename := file.OriginalName
	if filename == "" {
		filename = file.FileName
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))

	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
}

// GetFilesByEntity 获取实体的文件
func (h *UploadHandler) GetFilesByEntity(c *gin.Context) {
	entityType := c.Param("type")
	entityID := c.Param("id")

	if entityType == "" || entityID == "" {
		response.BadRequest(c, "entity type and id are required")
		return
	}

	files, err := h.fileService.GetFilesByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		response.InternalServerError(c, "failed to get files")
		return
	}

	response.Success(c, files)
}

// Delete 删除文件
func (h *UploadHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "file id is required")
		return
	}

	if err := h.fileService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrFileNotFound {
			response.NotFound(c, "file not found")
			return
		}
		logger.Error("Failed to delete file", logger.Err(err))
		response.InternalServerError(c, "failed to delete file")
		return
	}

	response.NoContent(c)
}

// BindToEntity 绑定文件到实体
func (h *UploadHandler) BindToEntity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "file id is required")
		return
	}

	var req dto.BindFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.fileService.BindToEntity(c.Request.Context(), id, req.EntityType, req.EntityID); err != nil {
		if err == service.ErrFileNotFound {
			response.NotFound(c, "file not found")
			return
		}
		response.InternalServerError(c, "failed to bind file")
		return
	}

	response.Success(c, nil)
}

// GetStats 获取文件统计
func (h *UploadHandler) GetStats(c *gin.Context) {
	stats, err := h.fileService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}
