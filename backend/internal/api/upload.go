package api

import (
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	storageService *service.StorageService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(storageService *service.StorageService) *UploadHandler {
	return &UploadHandler{
		storageService: storageService,
	}
}

// Upload 上传单个文件
// @Summary 上传文件
// @Description 上传单个文件，支持图片、音频、视频等类型
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param type query string false "文件类型分类(images/audio/video/document)" default(images)
// @Success 200 {object} utils.Response{data=service.FileInfo}
// @Router /upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "获取文件失败: "+err.Error())
		return
	}

	// 检查文件大小 (最大 50MB)
	if file.Size > 50*1024*1024 {
		utils.BadRequest(c, "文件大小不能超过50MB")
		return
	}

	// 获取文件类型分类
	fileType := c.DefaultQuery("type", "images")

	// 验证文件类型
	if !isAllowedFileType(file, fileType) {
		utils.BadRequest(c, "不支持的文件类型")
		return
	}

	// 上传文件
	ctx := c.Request.Context()
	fileInfo, err := h.storageService.UploadFile(ctx, file, fileType)
	if err != nil {
		utils.ServerError(c, "上传失败: "+err.Error())
		return
	}

	utils.Success(c, fileInfo)
}

// UploadMultiple 上传多个文件
// @Summary 批量上传文件
// @Description 同时上传多个文件
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "文件列表"
// @Param type query string false "文件类型分类(images/audio/video/document)" default(images)
// @Success 200 {object} utils.Response{data=[]service.FileInfo}
// @Router /upload/batch [post]
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	// 获取表单
	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "获取表单失败: "+err.Error())
		return
	}

	// 获取文件列表
	files := form.File["files"]
	if len(files) == 0 {
		utils.BadRequest(c, "请选择文件")
		return
	}

	// 检查文件数量 (最多 20 个)
	if len(files) > 20 {
		utils.BadRequest(c, "一次最多上传20个文件")
		return
	}

	// 检查总大小 (最大 100MB)
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
		if file.Size > 50*1024*1024 {
			utils.BadRequest(c, "单个文件大小不能超过50MB")
			return
		}
	}
	if totalSize > 100*1024*1024 {
		utils.BadRequest(c, "总文件大小不能超过100MB")
		return
	}

	// 获取文件类型分类
	fileType := c.DefaultQuery("type", "images")

	// 上传文件
	ctx := c.Request.Context()
	fileInfos, err := h.storageService.UploadMultipleFiles(ctx, files, fileType)
	if err != nil {
		utils.ServerError(c, "上传失败: "+err.Error())
		return
	}

	utils.Success(c, fileInfos)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除已上传的文件
// @Tags 文件上传
// @Produce json
// @Param url query string true "文件URL"
// @Success 200 {object} utils.Response
// @Router /upload [delete]
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	fileURL := c.Query("url")
	if fileURL == "" {
		utils.BadRequest(c, "请提供文件URL")
		return
	}

	ctx := c.Request.Context()
	err := h.storageService.DeleteFile(ctx, fileURL)
	if err != nil {
		utils.ServerError(c, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// isAllowedFileType 检查文件类型是否允许
func isAllowedFileType(file *multipart.FileHeader, fileType string) bool {
	contentType := ""
	if file.Header != nil {
		if types, ok := file.Header["Content-Type"]; ok && len(types) > 0 {
			contentType = types[0]
		}
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))

	switch fileType {
	case "images":
		allowedExts := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
			".webp": true, ".bmp": true, ".svg": true,
		}
		allowedTypes := map[string]bool{
			"image/jpeg": true, "image/png": true, "image/gif": true,
			"image/webp": true, "image/bmp": true, "image/svg+xml": true,
		}
		return allowedExts[ext] || allowedTypes[contentType]

	case "audio":
		allowedExts := map[string]bool{
			".mp3": true, ".wav": true, ".ogg": true, ".m4a": true,
			".aac": true, ".flac": true, ".wma": true,
		}
		allowedTypes := map[string]bool{
			"audio/mpeg": true, "audio/wav": true, "audio/ogg": true,
			"audio/mp4": true, "audio/aac": true, "audio/flac": true,
		}
		return allowedExts[ext] || allowedTypes[contentType]

	case "video":
		allowedExts := map[string]bool{
			".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
			".flv": true, ".mkv": true, ".webm": true,
		}
		allowedTypes := map[string]bool{
			"video/mp4": true, "video/avi": true, "video/quicktime": true,
			"video/x-ms-wmv": true, "video/x-flv": true, "video/webm": true,
		}
		return allowedExts[ext] || allowedTypes[contentType]

	case "document":
		allowedExts := map[string]bool{
			".pdf": true, ".doc": true, ".docx": true, ".xls": true,
			".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true,
		}
		allowedTypes := map[string]bool{
			"application/pdf": true, "application/msword": true,
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
			"text/plain": true,
		}
		return allowedExts[ext] || allowedTypes[contentType]

	default:
		return false
	}
}
