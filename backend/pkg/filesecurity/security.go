// Package filesecurity 文件上传安全检查
package filesecurity

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// FileSecurityChecker 文件安全检查器
type FileSecurityChecker struct {
	allowedExtensions map[string]bool
	allowedMimeTypes  map[string]bool
	maxFileSize       int64
	checkFileHeader   bool
}

// NewFileSecurityChecker 创建文件安全检查器
func NewFileSecurityChecker(allowedTypes []string, maxFileSize int64) *FileSecurityChecker {
	checker := &FileSecurityChecker{
		allowedExtensions: make(map[string]bool),
		allowedMimeTypes:  make(map[string]bool),
		maxFileSize:       maxFileSize,
		checkFileHeader:   true,
	}

	for _, t := range allowedTypes {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, ".") {
			checker.allowedExtensions[t] = true
			checker.allowedExtensions[t[1:]] = true
		} else {
			checker.allowedExtensions[t] = true
			checker.allowedExtensions["."+t] = true
		}
	}

	checker.allowedMimeTypes = map[string]bool{
		"image/jpeg":         true,
		"image/jpg":          true,
		"image/png":          true,
		"image/gif":          true,
		"image/webp":         true,
		"audio/mpeg":         true,
		"audio/mp3":          true,
		"audio/wav":          true,
		"audio/wave":         true,
		"audio/x-wav":        true,
		"video/mp4":          true,
		"video/mpeg":         true,
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"text/plain": true,
	}

	return checker
}

// SecurityCheckResult 安全检查结果
type SecurityCheckResult struct {
	Passed       bool     `json:"passed"`
	ErrorMessage string   `json:"error_message,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	FileType     string   `json:"file_type,omitempty"`
	FileSize     int64    `json:"file_size,omitempty"`
	MimeType     string   `json:"mime_type,omitempty"`
}

// CheckFile 检查上传文件
func (c *FileSecurityChecker) CheckFile(header *multipart.FileHeader) (*SecurityCheckResult, error) {
	result := &SecurityCheckResult{
		Passed:   true,
		Warnings: make([]string, 0),
	}

	if header.Size > c.maxFileSize {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("文件大小超过限制 (最大 %d MB)", c.maxFileSize/1024/1024)
		return result, nil
	}
	result.FileSize = header.Size

	ext := strings.ToLower(filepath.Ext(header.Filename))
	result.FileType = ext

	extWithoutDot := strings.TrimPrefix(ext, ".")
	if ext == "" || (!c.allowedExtensions[ext] && !c.allowedExtensions[extWithoutDot]) {
		result.Passed = false
		if ext == "" {
			result.ErrorMessage = "文件必须有扩展名"
		} else {
			result.ErrorMessage = fmt.Sprintf("不支持的文件类型: %s", ext)
		}
		return result, nil
	}

	file, err := header.Open()
	if err != nil {
		result.Passed = false
		result.ErrorMessage = "无法读取文件"
		return result, err
	}
	defer file.Close()

	mimeType, err := c.detectMimeType(file)
	if err != nil {
		result.Passed = false
		result.ErrorMessage = "无法检测文件类型"
		return result, err
	}
	result.MimeType = mimeType

	if !c.allowedMimeTypes[mimeType] {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("不支持的文件格式: %s", mimeType)
		return result, nil
	}

	if c.checkFileHeader {
		if err := c.checkFileHeaderValidity(file, ext); err != nil {
			result.Passed = false
			result.ErrorMessage = err.Error()
			return result, nil
		}
	}

	if err := c.checkFileName(header.Filename); err != nil {
		result.Warnings = append(result.Warnings, err.Error())
	}

	return result, nil
}

func (c *FileSecurityChecker) detectMimeType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}
	return http.DetectContentType(buffer[:n]), nil
}

func (c *FileSecurityChecker) checkFileHeaderValidity(file multipart.File, ext string) error {
	buffer := make([]byte, 8)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}
	header := buffer[:n]

	switch ext {
	case ".jpg", ".jpeg":
		if !isJPEGHeader(header) {
			return fmt.Errorf("文件头不是有效的JPEG格式")
		}
	case ".png":
		if !isPNGHeader(header) {
			return fmt.Errorf("文件头不是有效的PNG格式")
		}
	case ".gif":
		if !isGIFHeader(header) {
			return fmt.Errorf("文件头不是有效的GIF格式")
		}
	case ".pdf":
		if !isPDFHeader(header) {
			return fmt.Errorf("文件头不是有效的PDF格式")
		}
	case ".mp3":
		if !isMP3Header(header) {
			return fmt.Errorf("文件头不是有效的MP3格式")
		}
	}
	return nil
}

func (c *FileSecurityChecker) checkFileName(filename string) error {
	dangerousChars := []string{"..", "<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("文件名包含危险字符: %s", char)
		}
	}
	if strings.HasPrefix(filepath.Base(filename), ".") {
		return fmt.Errorf("文件名不能以点开始")
	}
	if filename == "" || filename == "." {
		return fmt.Errorf("文件名不能为空")
	}
	return nil
}

func isJPEGHeader(header []byte) bool {
	if len(header) < 2 {
		return false
	}
	return header[0] == 0xFF && header[1] == 0xD8
}

func isPNGHeader(header []byte) bool {
	if len(header) < 8 {
		return false
	}
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return bytes.Equal(header[:8], pngSignature)
}

func isGIFHeader(header []byte) bool {
	if len(header) < 6 {
		return false
	}
	return bytes.Equal(header[:4], []byte("GIF8")) &&
		(header[4] == '7' || header[4] == '9') &&
		header[5] == 'a'
}

func isPDFHeader(header []byte) bool {
	if len(header) < 5 {
		return false
	}
	return bytes.Equal(header[:5], []byte("%PDF-"))
}

func isMP3Header(header []byte) bool {
	if len(header) < 2 {
		return false
	}
	if len(header) >= 3 && bytes.Equal(header[:3], []byte("ID3")) {
		return true
	}
	return header[0] == 0xFF && (header[1]&0xE0) == 0xE0
}
