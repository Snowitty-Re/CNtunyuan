package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
// @Description 文件上传管理接口，支持单文件/批量上传、下载、绑定实体等功能
// @Tags 文件管理
// @BasePath /api/v1
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
		upload.GET("/stats", middleware.RequireAdmin(), h.GetStats)
	}
}

// Upload 单文件上传
// @Summary 单文件上传
// @Description 上传单个文件，支持图片、音频、视频等格式，文件大小限制50MB
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Success 200 {object} response.Response{data=dto.FileResponse} "上传成功"
// @Failure 400 {object} response.Response "请求参数错误（未提供文件或文件过大）"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload [post]
// @Security BearerAuth
func (h *UploadHandler) Upload(c *gin.Context) {
	logger.Info("Upload handler called",
		logger.String("content_type", c.ContentType()),
	)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("Failed to get form file", logger.Err(err))
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	userID := middleware.GetUserID(c)
	logger.Info("Uploading file",
		logger.String("filename", header.Filename),
		logger.Int64("size", header.Size),
		logger.String("user_id", userID),
	)

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

	logger.Info("File uploaded successfully", logger.String("file_id", resp.ID))
	response.Success(c, resp)
}

// UploadBatch 批量上传
// @Summary 批量文件上传
// @Description 同时上传多个文件，支持图片、音频、视频等格式
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "要上传的文件列表（支持多个）"
// @Success 200 {object} response.Response{data=dto.UploadResponse} "上传成功"
// @Failure 400 {object} response.Response "请求参数错误（未提供文件）"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/batch [post]
// @Security BearerAuth
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

	// 收集所有文件句柄用于最终统一关闭
	var openedFiles []multipart.File
	defer func() {
		for _, f := range openedFiles {
			f.Close()
		}
	}()

	for _, header := range files {
		f, err := header.Open()
		if err != nil {
			response.InternalServerError(c, "failed to read file")
			return
		}
		openedFiles = append(openedFiles, f)
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
// @Summary 获取文件详情
// @Description 根据文件ID获取文件详细信息（元数据，不包含文件内容）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path string true "文件ID"
// @Success 200 {object} response.Response{data=dto.FileResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误（文件ID不能为空）"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/{id} [get]
// @Security BearerAuth
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
// @Summary 下载文件
// @Description 根据文件ID下载文件内容，返回二进制文件流
// @Tags 文件管理
// @Accept json
// @Produce octet-stream
// @Param id path string true "文件ID"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} response.Response "请求参数错误（文件ID不能为空）"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/{id}/download [get]
// @Security BearerAuth
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

	// 设置下载头（安全处理文件名，防止头注入）
	filename := file.OriginalName
	if filename == "" {
		filename = file.FileName
	}
	// 清理文件名中的危险字符
	safeName := strings.Map(func(r rune) rune {
		if r == '"' || r == '\\' || r == '\n' || r == '\r' {
			return '_'
		}
		return r
	}, filename)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", safeName))
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))

	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
}

// GetFilesByEntity 获取实体的文件
// @Summary 获取实体关联的文件列表
// @Description 根据实体类型和ID获取关联的所有文件列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param type path string true "实体类型（如：missing_person, dialect, user等）"
// @Param id path string true "实体ID"
// @Success 200 {object} response.Response{data=[]dto.FileResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误（实体类型或ID不能为空）"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/entity/{type}/{id} [get]
// @Security BearerAuth
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
// @Summary 删除文件
// @Description 根据文件ID删除文件（物理删除文件及数据库记录）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path string true "文件ID"
// @Success 204 "删除成功，无返回内容"
// @Failure 400 {object} response.Response "请求参数错误（文件ID不能为空）"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/{id} [delete]
// @Security BearerAuth
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
// @Summary 绑定文件到实体
// @Description 将已上传的文件绑定到指定实体（如走失人员、方言等）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path string true "文件ID"
// @Param body body dto.BindFileRequest true "绑定请求参数"
// @Success 200 {object} response.Response "绑定成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/{id}/bind [put]
// @Security BearerAuth
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
// @Summary 获取文件统计信息
// @Description 获取系统文件统计数据，包括总数量、总大小、各类型文件统计等
// @Tags 文件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.FileStatsResponse} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /upload/stats [get]
// @Security BearerAuth
func (h *UploadHandler) GetStats(c *gin.Context) {
	stats, err := h.fileService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}
