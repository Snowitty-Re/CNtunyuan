package dto

import (
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// FileResponse 文件响应
type FileResponse struct {
	ID           string    `json:"id"`
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FileType     string    `json:"file_type"`
	MimeType     string    `json:"mime_type"`
	Size         int64     `json:"size"`
	SizeReadable string    `json:"size_readable"`
	URL          string    `json:"url"`
	StorageType  string    `json:"storage_type"`
	UploaderID   string    `json:"uploader_id"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	Page       int    `form:"page,default=1" binding:"min=1"`
	PageSize   int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword    string `form:"keyword"`
	FileType   string `form:"file_type"`
	UploaderID string `form:"uploader_id"`
}

// FileListResponse 文件列表响应
type FileListResponse = PageResult[FileResponse]

// FileStatsResponse 文件统计响应
type FileStatsResponse struct {
	TotalCount int64 `json:"total_count"`
	TotalSize  int64 `json:"total_size"`
	ImageCount int64 `json:"image_count"`
	ImageSize  int64 `json:"image_size"`
	AudioCount int64 `json:"audio_count"`
	AudioSize  int64 `json:"audio_size"`
	VideoCount int64 `json:"video_count"`
	VideoSize  int64 `json:"video_size"`
	DocCount   int64 `json:"doc_count"`
	DocSize    int64 `json:"doc_size"`
}

// BindFileRequest 绑定文件请求
type BindFileRequest struct {
	EntityType string `json:"entity_type" binding:"required"`
	EntityID   string `json:"entity_id" binding:"required"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	Files []FileResponse `json:"files"`
}

// ToFileResponse 转换为文件响应
func ToFileResponse(file *entity.File) FileResponse {
	return FileResponse{
		ID:           file.ID,
		FileName:     file.FileName,
		OriginalName: file.OriginalName,
		FileType:     string(file.FileType),
		MimeType:     file.MimeType,
		Size:         file.Size,
		SizeReadable: FormatFileSize(file.Size),
		URL:          file.URL,
		StorageType:  string(file.StorageType),
		UploaderID:   file.UploaderID,
		EntityType:   file.EntityType,
		EntityID:     file.EntityID,
		Description:  file.Description,
		CreatedAt:    file.CreatedAt,
	}
}

// NewFileListResponse 创建文件列表响应
func NewFileListResponse(list []FileResponse, total int64, page, pageSize int) FileListResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return FileListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
