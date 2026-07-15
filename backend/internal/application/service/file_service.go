package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	domainService "github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	apperrors "github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/filesecurity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

var (
	ErrFileNotFound       = fmt.Errorf("file not found")
	ErrFileTypeNotAllowed = fmt.Errorf("file type not allowed")
	ErrFileTooLarge       = fmt.Errorf("file too large")
	ErrFileForbidden      = fmt.Errorf("no permission to access file")
)

// FileAppService 文件应用服务
type FileAppService struct {
	fileRepo        repository.FileRepository
	storageService  domainService.StorageService
	securityChecker *filesecurity.FileSecurityChecker
	virusScanner    filesecurity.VirusScanner
	maxFileSize     int64
	allowedTypes    []string
}

// NewFileAppService 创建文件应用服务
func NewFileAppService(
	fileRepo repository.FileRepository,
	storageService domainService.StorageService,
	maxFileSize int64,
	allowedTypes []string,
) *FileAppService {
	return &FileAppService{
		fileRepo:        fileRepo,
		storageService:  storageService,
		securityChecker: filesecurity.NewFileSecurityChecker(allowedTypes, maxFileSize),
		virusScanner:    filesecurity.NewNoOpScanner(), // 默认使用空扫描器
		maxFileSize:     maxFileSize,
		allowedTypes:    allowedTypes,
	}
}

// SetVirusScanner 设置病毒扫描器
func (s *FileAppService) SetVirusScanner(scanner filesecurity.VirusScanner) {
	s.virusScanner = scanner
}

// UploadFile 上传单个文件
func (s *FileAppService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, uploaderID string) (*dto.FileResponse, error) {
	logger.Info("UploadFile service called",
		logger.String("filename", header.Filename),
		logger.Int64("size", header.Size),
		logger.String("uploader_id", uploaderID),
		logger.Int64("max_file_size", s.maxFileSize),
	)

	// 关闭传入的文件句柄，因为我们需要重新打开进行安全检查
	// 安全检查会在内部重新打开文件
	file.Close()

	// 执行安全检查（文件类型、MIME、病毒扫描）
	securityResult, err := filesecurity.PerformSecurityCheck(header, s.securityChecker, s.virusScanner)
	if err != nil {
		logger.Error("Security check failed", logger.Err(err), logger.String("filename", header.Filename))
		return nil, apperrors.New(apperrors.CodeInvalidFileType, "文件安全检查失败")
	}

	if !securityResult.Passed {
		logger.Warn("Security check failed",
			logger.String("filename", header.Filename),
			logger.String("error", securityResult.ErrorMessage),
		)
		if strings.Contains(securityResult.ErrorMessage, "文件大小超过限制") {
			return nil, apperrors.New(apperrors.CodeFileTooLarge, securityResult.ErrorMessage)
		}
		return nil, apperrors.New(apperrors.CodeInvalidFileType, securityResult.ErrorMessage)
	}

	logger.Info("Security check passed",
		logger.String("filename", header.Filename),
		logger.String("mime_type", securityResult.FileCheck.MimeType),
		logger.String("scanner", securityResult.VirusScan.Scanner),
	)

	// 重新打开文件进行上传
	uploadFile, err := header.Open()
	if err != nil {
		logger.Error("Failed to reopen file for upload", logger.Err(err))
		return nil, err
	}
	defer uploadFile.Close()

	// 上传文件到存储（使用重新打开的文件句柄）
	uploadedFile, err := s.storageService.Upload(ctx, uploadFile, header.Filename, header.Size, header.Header.Get("Content-Type"))
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
func (s *FileAppService) GetByID(ctx context.Context, id string, userID string, isManager bool) (*dto.FileResponse, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrFileNotFound
	}
	if !s.canReadFile(file, userID, isManager) {
		return nil, ErrFileForbidden
	}

	resp := dto.ToFileResponse(file)
	return &resp, nil
}

// GetFile 获取文件内容
func (s *FileAppService) GetFile(ctx context.Context, id string, userID string, isManager bool) (io.ReadCloser, *entity.File, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, ErrFileNotFound
	}
	if !s.canReadFile(file, userID, isManager) {
		return nil, nil, ErrFileForbidden
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
	query.Keyword = strings.TrimSpace(req.Keyword)
	query.FileType = entity.FileType(strings.TrimSpace(req.FileType))
	query.UploaderID = strings.TrimSpace(req.UploaderID)
	query.EntityType = strings.TrimSpace(req.EntityType)
	query.StorageType = entity.StorageType(strings.TrimSpace(req.StorageType))

	result, err := s.fileRepo.List(ctx, query)
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
func (s *FileAppService) Delete(ctx context.Context, id string, userID string, isManager bool) error {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return ErrFileNotFound
	}
	if !s.canWriteFile(file, userID, isManager) {
		return ErrFileForbidden
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
func (s *FileAppService) BindToEntity(ctx context.Context, fileID string, entityType string, entityID string, userID string, isManager bool) error {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return ErrFileNotFound
	}
	if !s.canWriteFile(file, userID, isManager) {
		return ErrFileForbidden
	}

	return s.fileRepo.UpdateEntity(ctx, fileID, entityType, entityID)
}

// GetFilesByEntity 获取实体的文件
func (s *FileAppService) GetFilesByEntity(ctx context.Context, entityType string, entityID string, userID string, isManager bool) ([]dto.FileResponse, error) {
	if entityType == "user" && entityID != userID && !isManager {
		return nil, ErrFileForbidden
	}

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

func (s *FileAppService) canWriteFile(file *entity.File, userID string, isManager bool) bool {
	return isManager || file.UploaderID == userID
}

func (s *FileAppService) canReadFile(file *entity.File, userID string, isManager bool) bool {
	// Manager 或上传者可读；不再对“任意已绑定实体文件”放行（防 IDOR）
	return s.canWriteFile(file, userID, isManager)
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
