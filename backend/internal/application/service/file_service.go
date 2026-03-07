package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	domainService "github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

var (
	ErrFileNotFound       = fmt.Errorf("file not found")
	ErrFileTypeNotAllowed = fmt.Errorf("file type not allowed")
	ErrFileTooLarge       = fmt.Errorf("file too large")
)

// FileAppService 文件应用服务
type FileAppService struct {
	fileRepo       repository.FileRepository
	storageService domainService.StorageService
	maxFileSize    int64
	allowedTypes   []string
}

// NewFileAppService 创建文件应用服务
func NewFileAppService(
	fileRepo repository.FileRepository,
	storageService domainService.StorageService,
	maxFileSize int64,
	allowedTypes []string,
) *FileAppService {
	return &FileAppService{
		fileRepo:       fileRepo,
		storageService: storageService,
		maxFileSize:    maxFileSize,
		allowedTypes:   allowedTypes,
	}
}

// UploadFile 上传单个文件
func (s *FileAppService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, uploaderID string) (*dto.FileResponse, error) {
	logger.Info("UploadFile service called",
		logger.String("filename", header.Filename),
		logger.Int64("size", header.Size),
		logger.String("uploader_id", uploaderID),
		logger.Int64("max_file_size", s.maxFileSize),
	)

	// 检查文件大小
	if s.maxFileSize > 0 && header.Size > s.maxFileSize {
		logger.Warn("File too large", logger.Int64("size", header.Size), logger.Int64("max", s.maxFileSize))
		return nil, ErrFileTooLarge
	}

	// 上传文件到存储
	uploadedFile, err := s.storageService.Upload(ctx, file, header.Filename, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		logger.Error("Failed to upload file to storage", logger.Err(err))
		return nil, err
	}

	logger.Info("File uploaded to storage",
		logger.String("path", uploadedFile.Path),
		logger.String("url", uploadedFile.URL),
	)

	// 设置上传者
	uploadedFile.UploaderID = uploaderID

	// 保存到数据库
	if err := s.fileRepo.Create(ctx, uploadedFile); err != nil {
		// 删除已上传的文件
		s.storageService.Delete(ctx, uploadedFile.Path)
		logger.Error("Failed to save file record to database", logger.Err(err))
		return nil, err
	}

	logger.Info("File record saved to database", logger.String("file_id", uploadedFile.ID))

	resp := dto.ToFileResponse(uploadedFile)
	return &resp, nil
}

// UploadFiles 批量上传文件
func (s *FileAppService) UploadFiles(ctx context.Context, files []multipart.File, headers []*multipart.FileHeader, uploaderID string) ([]dto.FileResponse, error) {
	responses := make([]dto.FileResponse, 0, len(files))

	for i, file := range files {
		resp, err := s.UploadFile(ctx, file, headers[i], uploaderID)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// GetByID 根据ID获取文件
func (s *FileAppService) GetByID(ctx context.Context, id string) (*dto.FileResponse, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrFileNotFound
	}

	resp := dto.ToFileResponse(file)
	return &resp, nil
}

// GetFile 获取文件内容
func (s *FileAppService) GetFile(ctx context.Context, id string) (io.ReadCloser, *entity.File, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, ErrFileNotFound
	}

	reader, err := s.storageService.Download(ctx, file.Path)
	if err != nil {
		return nil, nil, err
	}

	return reader, file, nil
}

// List 文件列表
func (s *FileAppService) List(ctx context.Context, req *dto.FileListRequest) (*dto.FileListResponse, error) {
	query := repository.NewFileQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.FileType = entity.FileType(req.FileType)
	query.UploaderID = req.UploaderID

	result, err := s.fileRepo.Search(ctx, req.Keyword, repository.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]dto.FileResponse, len(result.List))
	for i, file := range result.List {
		list[i] = dto.ToFileResponse(&file)
	}

	resp := dto.NewFileListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// Delete 删除文件
func (s *FileAppService) Delete(ctx context.Context, id string) error {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return ErrFileNotFound
	}

	// 软删除数据库记录
	if err := s.fileRepo.SoftDelete(ctx, id); err != nil {
		return err
	}

	// 删除物理文件
	if err := s.storageService.Delete(ctx, file.Path); err != nil {
		logger.Warn("Failed to delete physical file", logger.String("path", file.Path), logger.Err(err))
	}

	return nil
}

// BindToEntity 绑定文件到实体
func (s *FileAppService) BindToEntity(ctx context.Context, fileID string, entityType string, entityID string) error {
	return s.fileRepo.UpdateEntity(ctx, fileID, entityType, entityID)
}

// GetFilesByEntity 获取实体的文件
func (s *FileAppService) GetFilesByEntity(ctx context.Context, entityType string, entityID string) ([]dto.FileResponse, error) {
	files, err := s.fileRepo.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	list := make([]dto.FileResponse, len(files))
	for i, file := range files {
		list[i] = dto.ToFileResponse(&file)
	}

	return list, nil
}

// GetStats 获取文件统计
func (s *FileAppService) GetStats(ctx context.Context) (*dto.FileStatsResponse, error) {
	stats, err := s.fileRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.FileStatsResponse{
		TotalCount: stats.TotalCount,
		TotalSize:  stats.TotalSize,
		ImageCount: stats.ImageCount,
		ImageSize:  stats.ImageSize,
		AudioCount: stats.AudioCount,
		AudioSize:  stats.AudioSize,
		VideoCount: stats.VideoCount,
		VideoSize:  stats.VideoSize,
		DocCount:   stats.DocCount,
		DocSize:    stats.DocSize,
	}, nil
}
