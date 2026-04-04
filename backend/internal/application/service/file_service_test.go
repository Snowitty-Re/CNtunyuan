package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/testutil"
	"github.com/Snowitty-Re/CNtunyuan/pkg/filesecurity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockStorageService 存储服务 mock
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Upload(ctx context.Context, reader io.Reader, filename string, size int64, contentType string) (*entity.File, error) {
	args := m.Called(ctx, reader, filename, size, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.File), args.Error(1)
}

func (m *MockStorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorageService) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockStorageService) GetURL(ctx context.Context, path string) string {
	args := m.Called(ctx, path)
	return args.String(0)
}

func (m *MockStorageService) Exists(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorageService) GetType() entity.StorageType {
	args := m.Called()
	return args.Get(0).(entity.StorageType)
}

// MockVirusScanner 病毒扫描器 mock
type MockVirusScanner struct {
	mock.Mock
}

func (m *MockVirusScanner) Scan(file io.Reader, filename string) (*filesecurity.ScanResult, error) {
	args := m.Called(file, filename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesecurity.ScanResult), args.Error(1)
}

func (m *MockVirusScanner) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockVirusScanner) GetName() string {
	args := m.Called()
	return args.String(0)
}

// MockFileRepository 文件仓储 mock
type MockFileRepository struct {
	mock.Mock
}

// Implement Repository[entity.File] interface
func (m *MockFileRepository) Create(ctx context.Context, entity *entity.File) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFileRepository) Update(ctx context.Context, entity *entity.File) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileRepository) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileRepository) FindByID(ctx context.Context, id string) (*entity.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.File), args.Error(1)
}

func (m *MockFileRepository) FindAll(ctx context.Context) ([]entity.File, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.File), args.Error(1)
}

func (m *MockFileRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// Implement FileRepository specific methods
func (m *MockFileRepository) FindByUploader(ctx context.Context, uploaderID string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	args := m.Called(ctx, uploaderID, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.File]), args.Error(1)
}

func (m *MockFileRepository) FindByType(ctx context.Context, fileType entity.FileType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	args := m.Called(ctx, fileType, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.File]), args.Error(1)
}

func (m *MockFileRepository) FindByEntity(ctx context.Context, entityType string, entityID string) ([]entity.File, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.File), args.Error(1)
}

func (m *MockFileRepository) FindByURLOrPath(ctx context.Context, fileURL string, filePath string) (*entity.File, error) {
	args := m.Called(ctx, fileURL, filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.File), args.Error(1)
}

func (m *MockFileRepository) FindByStorageType(ctx context.Context, storageType entity.StorageType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	args := m.Called(ctx, storageType, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.File]), args.Error(1)
}

func (m *MockFileRepository) Search(ctx context.Context, keyword string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	args := m.Called(ctx, keyword, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.File]), args.Error(1)
}

func (m *MockFileRepository) UpdateEntity(ctx context.Context, fileID string, entityType string, entityID string) error {
	args := m.Called(ctx, fileID, entityType, entityID)
	return args.Error(0)
}

func (m *MockFileRepository) GetStats(ctx context.Context) (*entity.FileStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.FileStats), args.Error(1)
}

func (m *MockFileRepository) GetTotalSize(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) CountByType(ctx context.Context, fileType entity.FileType) (int64, error) {
	args := m.Called(ctx, fileType)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// Test Helpers
// ============================================================================

// createMultipartFileHeader 创建 multipart.FileHeader 用于测试
// 创建一个有效的 JPEG 文件（以 FF D8 开头）
func createMultipartFileHeader(filename string, content []byte, contentType string) *multipart.FileHeader {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// 创建表单文件
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}

	_, err = part.Write(content)
	if err != nil {
		panic(err)
	}

	writer.Close()

	// 解析 multipart 消息
	reader := multipart.NewReader(&b, writer.Boundary())
	form, err := reader.ReadForm(int64(len(content)) + 1024)
	if err != nil {
		panic(err)
	}

	files := form.File["file"]
	if len(files) == 0 {
		panic("no files in form")
	}

	return files[0]
}

// createValidJPEGContent 创建有效的 JPEG 文件内容
func createValidJPEGContent(size int) []byte {
	// JPEG 文件头: FF D8 FF
	content := make([]byte, size)
	content[0] = 0xFF
	content[1] = 0xD8
	content[2] = 0xFF
	// 添加 JFIF 标记使其被识别为 image/jpeg
	copy(content[3:], []byte{0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00})
	// 填充其余字节
	for i := 11; i < size; i++ {
		content[i] = byte(i % 256)
	}
	return content
}

// createValidPNGContent 创建有效的 PNG 文件内容
func createValidPNGContent(size int) []byte {
	// PNG 文件头: 89 50 4E 47 0D 0A 1A 0A
	content := make([]byte, size)
	copy(content, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	for i := 8; i < size; i++ {
		content[i] = byte(i % 256)
	}
	return content
}

// ============================================================================
// Test Setup
// ============================================================================

func setupFileServiceTest(t *testing.T) (*FileAppService, *MockFileRepository, *MockStorageService, *MockVirusScanner) {
	mockRepo := new(MockFileRepository)
	mockStorage := new(MockStorageService)
	mockScanner := new(MockVirusScanner)

	// 默认配置：允许 jpg, png, gif 类型，最大 10MB
	allowedTypes := []string{"jpg", "jpeg", "png", "gif"}
	maxFileSize := int64(10 * 1024 * 1024) // 10MB

	service := NewFileAppService(mockRepo, mockStorage, maxFileSize, allowedTypes)
	service.SetVirusScanner(mockScanner)

	// 配置 mock scanner 默认返回可用且干净
	mockScanner.On("IsAvailable").Return(true).Maybe()
	mockScanner.On("GetName").Return("MockScanner").Maybe()

	return service, mockRepo, mockStorage, mockScanner
}

// ============================================================================
// Tests for UploadFile
// ============================================================================

func TestFileAppService_UploadFile_Success(t *testing.T) {
	service, mockRepo, mockStorage, mockScanner := setupFileServiceTest(t)

	// 创建有效的 JPEG 文件内容
	content := createValidJPEGContent(1024)
	header := createMultipartFileHeader("test.jpg", content, "image/jpeg")

	// 创建 multipart.File
	file, err := header.Open()
	require.NoError(t, err)
	defer file.Close()

	uploaderID := "user-123"

	// 设置 mock 期望
	uploadedFile := &entity.File{
		BaseEntity:   entity.BaseEntity{ID: "file-123"},
		FileName:     "test_123456.jpg",
		OriginalName: "test.jpg",
		FileType:     entity.FileTypeImage,
		MimeType:     "image/jpeg",
		Size:         int64(len(content)),
		Path:         "/uploads/test_123456.jpg",
		URL:          "http://localhost/uploads/test_123456.jpg",
		StorageType:  entity.StorageTypeLocal,
		UploaderID:   uploaderID,
	}

	mockScanner.On("Scan", mock.Anything, "test.jpg").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil)

	mockStorage.On("Upload", mock.Anything, mock.Anything, "test.jpg", int64(len(content)), mock.Anything).
		Return(uploadedFile, nil)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.File")).Return(nil)

	// 执行测试
	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, uploaderID)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uploadedFile.ID, resp.ID)
	assert.Equal(t, uploadedFile.OriginalName, resp.OriginalName)
	assert.Equal(t, string(uploadedFile.FileType), resp.FileType)
	assert.Equal(t, uploadedFile.URL, resp.URL)
	assert.Equal(t, uploaderID, resp.UploaderID)

	// 验证 mock 调用
	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockScanner.AssertExpectations(t)
}

func TestFileAppService_UploadFile_InvalidFileType(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	// 创建不支持的文件类型（例如 .exe）
	content := []byte("some executable content")

	// 手动创建 header，因为 exe 文件类型不支持
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, _ := writer.CreateFormFile("file", "test.exe")
	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(&b, writer.Boundary())
	form, _ := reader.ReadForm(10240)
	header := form.File["file"][0]

	file, err := header.Open()
	require.NoError(t, err)
	defer file.Close()

	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, "user-123")

	// 验证结果
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "不支持的文件类型")

	// 验证存储和仓储未被调用
	mockStorage.AssertNotCalled(t, "Upload")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestFileAppService_UploadFile_FileTooLarge(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	// 创建超过 10MB 的 JPEG 文件
	largeContent := createValidJPEGContent(11 * 1024 * 1024) // 11MB
	header := createMultipartFileHeader("large.jpg", largeContent, "image/jpeg")

	file, err := header.Open()
	require.NoError(t, err)
	defer file.Close()

	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, "user-123")

	// 验证结果
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "文件大小超过限制")

	// 验证存储和仓储未被调用
	mockStorage.AssertNotCalled(t, "Upload")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestFileAppService_UploadFile_VirusDetected(t *testing.T) {
	service, mockRepo, mockStorage, mockScanner := setupFileServiceTest(t)

	content := createValidJPEGContent(1024)
	header := createMultipartFileHeader("virus.jpg", content, "image/jpeg")

	file, err := header.Open()
	require.NoError(t, err)
	defer file.Close()

	// 设置 mock scanner 返回检测到病毒
	mockScanner.On("Scan", mock.Anything, "virus.jpg").Return(&filesecurity.ScanResult{
		Clean:     false,
		Infected:  true,
		VirusName: "TestVirus",
		Scanner:   "MockScanner",
	}, nil)

	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, "user-123")

	// 验证结果
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "检测到病毒")

	// 验证存储和仓储未被调用
	mockStorage.AssertNotCalled(t, "Upload")
	mockRepo.AssertNotCalled(t, "Create")
	mockScanner.AssertExpectations(t)
}

func TestFileAppService_UploadFile_StorageUploadFailed(t *testing.T) {
	service, mockRepo, mockStorage, mockScanner := setupFileServiceTest(t)

	content := createValidJPEGContent(1024)
	header := createMultipartFileHeader("test.jpg", content, "image/jpeg")

	file, err := header.Open()
	require.NoError(t, err)
	defer file.Close()

	// 设置 mock 期望：扫描通过，但存储上传失败
	mockScanner.On("Scan", mock.Anything, "test.jpg").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil)

	uploadError := errors.New("storage upload failed")
	mockStorage.On("Upload", mock.Anything, mock.Anything, "test.jpg", int64(len(content)), mock.Anything).
		Return(nil, uploadError)

	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, "user-123")

	// 验证结果
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, uploadError, err)

	// 验证仓储未被调用
	mockRepo.AssertNotCalled(t, "Create")
}

func TestFileAppService_UploadFile_DatabaseSaveFailed(t *testing.T) {
	service, mockRepo, mockStorage, mockScanner := setupFileServiceTest(t)

	content := createValidJPEGContent(1024)
	header := createMultipartFileHeader("test.jpg", content, "image/jpeg")

	file, err := header.Open()
	require.NoError(t, err)
	defer file.Close()

	uploadedFile := &entity.File{
		BaseEntity: entity.BaseEntity{ID: "file-123"},
		FileName:   "test.jpg",
		Path:       "/uploads/test.jpg",
	}

	dbError := errors.New("database error")

	// 设置 mock 期望
	mockScanner.On("Scan", mock.Anything, "test.jpg").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil)

	mockStorage.On("Upload", mock.Anything, mock.Anything, "test.jpg", int64(len(content)), mock.Anything).
		Return(uploadedFile, nil)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.File")).Return(dbError)

	// 期望删除已上传的文件
	mockStorage.On("Delete", mock.Anything, uploadedFile.Path).Return(nil)

	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, "user-123")

	// 验证结果
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, dbError, err)

	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Tests for UploadFiles (Multiple files)
// ============================================================================

func TestFileAppService_UploadFiles_Success(t *testing.T) {
	service, mockRepo, mockStorage, mockScanner := setupFileServiceTest(t)

	// 创建两个文件
	content1 := createValidJPEGContent(1024)
	content2 := createValidPNGContent(1024)

	header1 := createMultipartFileHeader("file1.jpg", content1, "image/jpeg")
	header2 := createMultipartFileHeader("file2.png", content2, "image/png")

	file1, _ := header1.Open()
	file2, _ := header2.Open()
	defer file1.Close()
	defer file2.Close()

	files := []multipart.File{file1, file2}
	headers := []*multipart.FileHeader{header1, header2}
	uploaderID := "user-123"

	uploadedFile1 := &entity.File{
		BaseEntity:   entity.BaseEntity{ID: "file-1"},
		FileName:     "file1.jpg",
		OriginalName: "file1.jpg",
		FileType:     entity.FileTypeImage,
		MimeType:     "image/jpeg",
		Size:         int64(len(content1)),
		Path:         "/uploads/file1.jpg",
		URL:          "http://localhost/uploads/file1.jpg",
		StorageType:  entity.StorageTypeLocal,
		UploaderID:   uploaderID,
	}

	uploadedFile2 := &entity.File{
		BaseEntity:   entity.BaseEntity{ID: "file-2"},
		FileName:     "file2.png",
		OriginalName: "file2.png",
		FileType:     entity.FileTypeImage,
		MimeType:     "image/png",
		Size:         int64(len(content2)),
		Path:         "/uploads/file2.png",
		URL:          "http://localhost/uploads/file2.png",
		StorageType:  entity.StorageTypeLocal,
		UploaderID:   uploaderID,
	}

	// 设置 mock 期望（每个文件都会被扫描和上传）
	mockScanner.On("Scan", mock.Anything, "file1.jpg").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil).Once()
	mockScanner.On("Scan", mock.Anything, "file2.png").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil).Once()

	mockStorage.On("Upload", mock.Anything, mock.Anything, "file1.jpg", int64(len(content1)), mock.Anything).
		Return(uploadedFile1, nil).Once()
	mockStorage.On("Upload", mock.Anything, mock.Anything, "file2.png", int64(len(content2)), mock.Anything).
		Return(uploadedFile2, nil).Once()

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.File")).Return(nil).Twice()

	// 执行测试前需要重新打开文件，因为之前的 Open 已经被 scanner 读取过了
	file1, _ = header1.Open()
	file2, _ = header2.Open()
	files = []multipart.File{file1, file2}

	ctx := testutil.Context()
	responses, err := service.UploadFiles(ctx, files, headers, uploaderID)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, responses, 2)
	assert.Equal(t, "file-1", responses[0].ID)
	assert.Equal(t, "file-2", responses[1].ID)

	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockScanner.AssertExpectations(t)
}

func TestFileAppService_UploadFiles_PartialFailure(t *testing.T) {
	service, mockRepo, mockStorage, mockScanner := setupFileServiceTest(t)

	content1 := createValidJPEGContent(1024)
	content2 := createValidPNGContent(1024)

	header1 := createMultipartFileHeader("file1.jpg", content1, "image/jpeg")
	header2 := createMultipartFileHeader("file2.png", content2, "image/png")

	file1, _ := header1.Open()
	file2, _ := header2.Open()
	defer file1.Close()
	defer file2.Close()

	files := []multipart.File{file1, file2}
	headers := []*multipart.FileHeader{header1, header2}

	uploadedFile1 := &entity.File{
		BaseEntity: entity.BaseEntity{ID: "file-1"},
		Path:       "/uploads/file1.jpg",
	}

	// 设置 mock 期望：第一个文件成功，第二个文件失败
	mockScanner.On("Scan", mock.Anything, "file1.jpg").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil).Once()
	mockScanner.On("Scan", mock.Anything, "file2.png").Return(&filesecurity.ScanResult{
		Clean:   true,
		Scanner: "MockScanner",
	}, nil).Once()

	mockStorage.On("Upload", mock.Anything, mock.Anything, "file1.jpg", int64(len(content1)), mock.Anything).
		Return(uploadedFile1, nil).Once()

	uploadError := errors.New("upload failed for second file")
	mockStorage.On("Upload", mock.Anything, mock.Anything, "file2.png", int64(len(content2)), mock.Anything).
		Return(nil, uploadError).Once()

	// 第一个文件保存成功
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.File")).Return(nil).Once()

	// 重新打开文件
	file1, _ = header1.Open()
	file2, _ = header2.Open()
	files = []multipart.File{file1, file2}

	ctx := testutil.Context()
	responses, err := service.UploadFiles(ctx, files, headers, "user-123")

	// 验证结果 - 应该返回错误，因为没有事务回滚机制
	require.Error(t, err)
	assert.Nil(t, responses)
	assert.Equal(t, uploadError, err)
}

// ============================================================================
// Tests for GetByID
// ============================================================================

func TestFileAppService_GetByID_Success(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	fileID := "file-123"
	now := time.Now()

	expectedFile := &entity.File{
		BaseEntity: entity.BaseEntity{
			ID:        fileID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		FileName:     "test.jpg",
		OriginalName: "test.jpg",
		FileType:     entity.FileTypeImage,
		MimeType:     "image/jpeg",
		Size:         1024,
		Path:         "/uploads/test.jpg",
		URL:          "http://localhost/uploads/test.jpg",
		StorageType:  entity.StorageTypeLocal,
		UploaderID:   "user-123",
	}

	mockRepo.On("FindByID", mock.Anything, fileID).Return(expectedFile, nil)

	ctx := testutil.Context()
	resp, err := service.GetByID(ctx, fileID, "uploader-id", true)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, fileID, resp.ID)
	assert.Equal(t, expectedFile.FileName, resp.FileName)
	assert.Equal(t, expectedFile.URL, resp.URL)

	mockRepo.AssertExpectations(t)
}

func TestFileAppService_GetByID_NotFound(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	fileID := "non-existent-file"

	mockRepo.On("FindByID", mock.Anything, fileID).Return(nil, errors.New("record not found"))

	ctx := testutil.Context()
	resp, err := service.GetByID(ctx, fileID, "uploader-id", true)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrFileNotFound, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Tests for Delete
// ============================================================================

func TestFileAppService_Delete_Success(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	fileID := "file-123"
	filePath := "/uploads/test.jpg"

	expectedFile := &entity.File{
		BaseEntity: entity.BaseEntity{ID: fileID},
		Path:       filePath,
	}

	mockRepo.On("FindByID", mock.Anything, fileID).Return(expectedFile, nil)
	mockRepo.On("SoftDelete", mock.Anything, fileID).Return(nil)
	mockStorage.On("Delete", mock.Anything, filePath).Return(nil)

	ctx := testutil.Context()
	err := service.Delete(ctx, fileID, "uploader-id", true)

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestFileAppService_Delete_NotFound(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	fileID := "non-existent-file"

	mockRepo.On("FindByID", mock.Anything, fileID).Return(nil, errors.New("record not found"))

	ctx := testutil.Context()
	err := service.Delete(ctx, fileID, "uploader-id", true)

	require.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)

	// 验证其他方法未被调用
	mockRepo.AssertNotCalled(t, "SoftDelete")
	mockStorage.AssertNotCalled(t, "Delete")
}

func TestFileAppService_Delete_PhysicalDeleteFailed(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	fileID := "file-123"
	filePath := "/uploads/test.jpg"

	expectedFile := &entity.File{
		BaseEntity: entity.BaseEntity{ID: fileID},
		Path:       filePath,
	}

	mockRepo.On("FindByID", mock.Anything, fileID).Return(expectedFile, nil)
	mockRepo.On("SoftDelete", mock.Anything, fileID).Return(nil)

	// 物理删除失败，但软删除成功
	deleteError := errors.New("physical delete failed")
	mockStorage.On("Delete", mock.Anything, filePath).Return(deleteError)

	ctx := testutil.Context()
	err := service.Delete(ctx, fileID, "uploader-id", true)

	// 应该返回 nil，因为软删除成功了
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// ============================================================================
// Tests for GetFile
// ============================================================================

func TestFileAppService_GetFile_Success(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	fileID := "file-123"
	filePath := "/uploads/test.jpg"

	expectedFile := &entity.File{
		BaseEntity: entity.BaseEntity{ID: fileID},
		Path:       filePath,
	}

	mockContent := io.NopCloser(strings.NewReader("file content"))

	mockRepo.On("FindByID", mock.Anything, fileID).Return(expectedFile, nil)
	mockStorage.On("Download", mock.Anything, filePath).Return(mockContent, nil)

	ctx := testutil.Context()
	reader, file, err := service.GetFile(ctx, fileID, "uploader-id", true)

	require.NoError(t, err)
	assert.NotNil(t, reader)
	assert.NotNil(t, file)
	assert.Equal(t, fileID, file.ID)

	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestFileAppService_GetFile_NotFound(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	fileID := "non-existent-file"

	mockRepo.On("FindByID", mock.Anything, fileID).Return(nil, errors.New("record not found"))

	ctx := testutil.Context()
	reader, file, err := service.GetFile(ctx, fileID, "uploader-id", true)

	require.Error(t, err)
	assert.Nil(t, reader)
	assert.Nil(t, file)
	assert.Equal(t, ErrFileNotFound, err)

	mockStorage.AssertNotCalled(t, "Download")
}

func TestFileAppService_GetFile_DownloadFailed(t *testing.T) {
	service, mockRepo, mockStorage, _ := setupFileServiceTest(t)

	fileID := "file-123"
	filePath := "/uploads/test.jpg"

	expectedFile := &entity.File{
		BaseEntity: entity.BaseEntity{ID: fileID},
		Path:       filePath,
	}

	downloadError := errors.New("download failed")

	mockRepo.On("FindByID", mock.Anything, fileID).Return(expectedFile, nil)
	mockStorage.On("Download", mock.Anything, filePath).Return(nil, downloadError)

	ctx := testutil.Context()
	reader, file, err := service.GetFile(ctx, fileID, "uploader-id", true)

	require.Error(t, err)
	assert.Nil(t, reader)
	assert.Nil(t, file)
	assert.Equal(t, downloadError, err)
}

// ============================================================================
// Tests for List
// ============================================================================

func TestFileAppService_List_Success(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	req := &dto.FileListRequest{
		Page:       1,
		PageSize:   10,
		Keyword:    "test",
		FileType:   "image",
		UploaderID: "user-123",
	}

	now := time.Now()
	files := []entity.File{
		{
			BaseEntity: entity.BaseEntity{
				ID:        "file-1",
				CreatedAt: now,
			},
			FileName:   "test1.jpg",
			FileType:   entity.FileTypeImage,
			Size:       1024,
			UploaderID: "user-123",
		},
		{
			BaseEntity: entity.BaseEntity{
				ID:        "file-2",
				CreatedAt: now,
			},
			FileName:   "test2.png",
			FileType:   entity.FileTypeImage,
			Size:       2048,
			UploaderID: "user-123",
		},
	}

	pageResult := &repository.PageResult[entity.File]{
		List:       files,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockRepo.On("Search", mock.Anything, req.Keyword, repository.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}).Return(pageResult, nil)

	ctx := testutil.Context()
	resp, err := service.List(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.List, 2)
	assert.Equal(t, int64(2), resp.Total)
	assert.Equal(t, 1, resp.Page)

	mockRepo.AssertExpectations(t)
}

func TestFileAppService_List_SearchFailed(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	req := &dto.FileListRequest{
		Page:     1,
		PageSize: 10,
		Keyword:  "test",
	}

	searchError := errors.New("search failed")

	mockRepo.On("Search", mock.Anything, req.Keyword, repository.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}).Return(nil, searchError)

	ctx := testutil.Context()
	resp, err := service.List(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, searchError, err)
}

// ============================================================================
// Tests for BindToEntity
// ============================================================================

func TestFileAppService_BindToEntity_Success(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	fileID := "file-123"
	entityType := "task"
	entityID := "task-456"

	mockRepo.On("FindByID", mock.Anything, fileID).Return(&entity.File{
		BaseEntity: entity.BaseEntity{ID: fileID},
		UploaderID: "uploader-id",
	}, nil)
	mockRepo.On("UpdateEntity", mock.Anything, fileID, entityType, entityID).Return(nil)

	ctx := testutil.Context()
	err := service.BindToEntity(ctx, fileID, entityType, entityID, "uploader-id", true)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFileAppService_BindToEntity_Failed(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	fileID := "file-123"
	entityType := "task"
	entityID := "task-456"

	bindError := errors.New("bind failed")
	mockRepo.On("FindByID", mock.Anything, fileID).Return(&entity.File{
		BaseEntity: entity.BaseEntity{ID: fileID},
		UploaderID: "uploader-id",
	}, nil)
	mockRepo.On("UpdateEntity", mock.Anything, fileID, entityType, entityID).Return(bindError)

	ctx := testutil.Context()
	err := service.BindToEntity(ctx, fileID, entityType, entityID, "uploader-id", true)

	require.Error(t, err)
	assert.Equal(t, bindError, err)
}

// ============================================================================
// Tests for GetFilesByEntity
// ============================================================================

func TestFileAppService_GetFilesByEntity_Success(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	entityType := "task"
	entityID := "task-456"

	now := time.Now()
	files := []entity.File{
		{
			BaseEntity: entity.BaseEntity{
				ID:        "file-1",
				CreatedAt: now,
			},
			FileName:   "attachment1.jpg",
			FileType:   entity.FileTypeImage,
			EntityType: entityType,
			EntityID:   entityID,
		},
		{
			BaseEntity: entity.BaseEntity{
				ID:        "file-2",
				CreatedAt: now,
			},
			FileName:   "attachment2.pdf",
			FileType:   entity.FileTypeDocument,
			EntityType: entityType,
			EntityID:   entityID,
		},
	}

	mockRepo.On("FindByEntity", mock.Anything, entityType, entityID).Return(files, nil)

	ctx := testutil.Context()
	resp, err := service.GetFilesByEntity(ctx, entityType, entityID, "uploader-id", true)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "file-1", resp[0].ID)
	assert.Equal(t, "file-2", resp[1].ID)

	mockRepo.AssertExpectations(t)
}

func TestFileAppService_GetFilesByEntity_Failed(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	entityType := "task"
	entityID := "task-456"

	findError := errors.New("find failed")
	mockRepo.On("FindByEntity", mock.Anything, entityType, entityID).Return(nil, findError)

	ctx := testutil.Context()
	resp, err := service.GetFilesByEntity(ctx, entityType, entityID, "uploader-id", true)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, findError, err)
}

// ============================================================================
// Tests for GetStats
// ============================================================================

func TestFileAppService_GetStats_Success(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	expectedStats := &entity.FileStats{
		TotalCount: 100,
		TotalSize:  1024 * 1024 * 100, // 100MB
		ImageCount: 50,
		ImageSize:  1024 * 1024 * 50,
		AudioCount: 20,
		AudioSize:  1024 * 1024 * 20,
		VideoCount: 15,
		VideoSize:  1024 * 1024 * 20,
		DocCount:   15,
		DocSize:    1024 * 1024 * 10,
	}

	mockRepo.On("GetStats", mock.Anything).Return(expectedStats, nil)

	ctx := testutil.Context()
	resp, err := service.GetStats(ctx)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedStats.TotalCount, resp.TotalCount)
	assert.Equal(t, expectedStats.TotalSize, resp.TotalSize)
	assert.Equal(t, expectedStats.ImageCount, resp.ImageCount)
	assert.Equal(t, expectedStats.AudioCount, resp.AudioCount)

	mockRepo.AssertExpectations(t)
}

func TestFileAppService_GetStats_Failed(t *testing.T) {
	service, mockRepo, _, _ := setupFileServiceTest(t)

	statsError := errors.New("get stats failed")
	mockRepo.On("GetStats", mock.Anything).Return(nil, statsError)

	ctx := testutil.Context()
	resp, err := service.GetStats(ctx)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, statsError, err)
}

// ============================================================================
// Integration Tests with Real Database
// ============================================================================

func TestFileAppService_Integration_UploadAndGet(t *testing.T) {
	// 创建真实测试数据库
	testDB := testutil.NewTestDB(t)
	defer testDB.Close()

	// 创建仓储实现
	fileRepo := NewGormFileRepository(testDB.DB)

	// 创建 mock 存储服务
	mockStorage := new(MockStorageService)

	// 创建服务
	allowedTypes := []string{"jpg", "png"}
	maxFileSize := int64(10 * 1024 * 1024)
	service := NewFileAppService(fileRepo, mockStorage, maxFileSize, allowedTypes)

	// 使用空扫描器绕过病毒扫描
	// 设置 mock 存储返回
	content := createValidJPEGContent(1024)
	header := createMultipartFileHeader("test.jpg", content, "image/jpeg")

	uploadedFile := &entity.File{
		BaseEntity:   entity.BaseEntity{ID: "file-int-001"},
		FileName:     "test.jpg",
		OriginalName: "test.jpg",
		FileType:     entity.FileTypeImage,
		MimeType:     "image/jpeg",
		Size:         int64(len(content)),
		Path:         "/uploads/test.jpg",
		URL:          "http://localhost/uploads/test.jpg",
		StorageType:  entity.StorageTypeLocal,
		UploaderID:   "user-123",
	}

	mockStorage.On("Upload", mock.Anything, mock.Anything, "test.jpg", int64(len(content)), mock.Anything).
		Return(uploadedFile, nil)

	// 测试上传
	file, _ := header.Open()
	defer file.Close()

	ctx := testutil.Context()
	resp, err := service.UploadFile(ctx, file, header, "user-123")

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// 测试获取
	getResp, err := service.GetByID(ctx, resp.ID, "integration-user", true)
	require.NoError(t, err)
	assert.Equal(t, resp.ID, getResp.ID)
	assert.Equal(t, resp.FileName, getResp.FileName)
}

// GormFileRepository 用于集成测试的简单实现
type GormFileRepository struct {
	db *gorm.DB
}

func NewGormFileRepository(db *gorm.DB) *GormFileRepository {
	return &GormFileRepository{db: db}
}

// Implement Repository[entity.File]
func (r *GormFileRepository) Create(ctx context.Context, entity *entity.File) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *GormFileRepository) Update(ctx context.Context, entity *entity.File) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormFileRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.File{}, "id = ?", id).Error
}

func (r *GormFileRepository) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.File{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *GormFileRepository) FindByID(ctx context.Context, id string) (*entity.File, error) {
	var file entity.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *GormFileRepository) FindAll(ctx context.Context) ([]entity.File, error) {
	var files []entity.File
	err := r.db.WithContext(ctx).Find(&files).Error
	return files, err
}

func (r *GormFileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).Count(&count).Error
	return count, err
}

func (r *GormFileRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Implement FileRepository specific methods (minimal for tests)
func (r *GormFileRepository) FindByUploader(ctx context.Context, uploaderID string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	return nil, nil
}

func (r *GormFileRepository) FindByType(ctx context.Context, fileType entity.FileType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	return nil, nil
}

func (r *GormFileRepository) FindByEntity(ctx context.Context, entityType string, entityID string) ([]entity.File, error) {
	var files []entity.File
	err := r.db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", entityType, entityID).Find(&files).Error
	return files, err
}

func (r *GormFileRepository) FindByURLOrPath(ctx context.Context, fileURL string, filePath string) (*entity.File, error) {
	var file entity.File
	db := r.db.WithContext(ctx).Model(&entity.File{})
	switch {
	case fileURL != "" && filePath != "":
		db = db.Where("url = ? OR path = ?", fileURL, filePath)
	case fileURL != "":
		db = db.Where("url = ?", fileURL)
	default:
		db = db.Where("path = ?", filePath)
	}
	if err := db.First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *GormFileRepository) FindByStorageType(ctx context.Context, storageType entity.StorageType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	return nil, nil
}

func (r *GormFileRepository) Search(ctx context.Context, keyword string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64
	db := r.db.WithContext(ctx).Model(&entity.File{})
	if keyword != "" {
		db = db.Where("file_name LIKE ?", "%"+keyword+"%")
	}
	db.Count(&total)
	db.Offset((pagination.Page - 1) * pagination.PageSize).Limit(pagination.PageSize).Find(&files)
	return repository.NewPageResult(files, total, pagination.Page, pagination.PageSize), nil
}

func (r *GormFileRepository) UpdateEntity(ctx context.Context, fileID string, entityType string, entityID string) error {
	return r.db.WithContext(ctx).Model(&entity.File{}).Where("id = ?", fileID).Updates(map[string]interface{}{
		"entity_type": entityType,
		"entity_id":   entityID,
	}).Error
}

func (r *GormFileRepository) GetStats(ctx context.Context) (*entity.FileStats, error) {
	return &entity.FileStats{}, nil
}

func (r *GormFileRepository) GetTotalSize(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).Select("COALESCE(SUM(size), 0)").Scan(&total).Error
	return total, err
}

func (r *GormFileRepository) CountByType(ctx context.Context, fileType entity.FileType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).Where("file_type = ?", fileType).Count(&count).Error
	return count, err
}
