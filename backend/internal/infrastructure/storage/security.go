// Package storage 文件上传安全检查
package storage

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
	// 允许的文件扩展名
	allowedExtensions map[string]bool
	// 允许的文件MIME类型
	allowedMimeTypes map[string]bool
	// 最大文件大小（字节）
	maxFileSize int64
	// 是否检查文件头
	checkFileHeader bool
}

// NewFileSecurityChecker 创建文件安全检查器
func NewFileSecurityChecker(allowedTypes []string, maxFileSize int64) *FileSecurityChecker {
	checker := &FileSecurityChecker{
		allowedExtensions: make(map[string]bool),
		allowedMimeTypes:  make(map[string]bool),
		maxFileSize:       maxFileSize,
		checkFileHeader:   true,
	}

	// 解析允许的类型
	for _, t := range allowedTypes {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		
		// 如果是扩展名格式（如 .jpg 或 jpg）
		if strings.HasPrefix(t, ".") {
			checker.allowedExtensions[t] = true
			checker.allowedExtensions[t[1:]] = true
		} else {
			checker.allowedExtensions[t] = true
			checker.allowedExtensions["."+t] = true
		}
	}

	// 设置默认允许的MIME类型
	checker.allowedMimeTypes = map[string]bool{
		"image/jpeg":                true,
		"image/jpg":                 true,
		"image/png":                 true,
		"image/gif":                 true,
		"image/webp":                true,
		"audio/mpeg":                true,
		"audio/mp3":                 true,
		"audio/wav":                 true,
		"audio/wave":                true,
		"audio/x-wav":               true,
		"video/mp4":                 true,
		"video/mpeg":                true,
		"application/pdf":           true,
		"application/msword":        true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"text/plain":                true,
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

// CheckFile 检查上传的文件
func (c *FileSecurityChecker) CheckFile(header *multipart.FileHeader) (*SecurityCheckResult, error) {
	result := &SecurityCheckResult{
		Passed:   true,
		Warnings: make([]string, 0),
	}

	// 检查文件大小
	if header.Size > c.maxFileSize {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("文件大小超过限制 (最大 %d MB)", c.maxFileSize/1024/1024)
		return result, nil
	}
	result.FileSize = header.Size

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	result.FileType = ext

	// 检查扩展名
	if !c.allowedExtensions[ext] && !c.allowedExtensions[ext[1:]] {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("不支持的文件类型: %s", ext)
		return result, nil
	}

	// 打开文件进行进一步检查
	file, err := header.Open()
	if err != nil {
		result.Passed = false
		result.ErrorMessage = "无法读取文件"
		return result, err
	}
	defer file.Close()

	// 检测MIME类型
	mimeType, err := c.detectMimeType(file)
	if err != nil {
		result.Passed = false
		result.ErrorMessage = "无法检测文件类型"
		return result, err
	}
	result.MimeType = mimeType

	// 检查MIME类型是否允许
	if !c.allowedMimeTypes[mimeType] {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("不支持的文件格式: %s", mimeType)
		return result, nil
	}

	// 文件头检查（防止伪造扩展名）
	if c.checkFileHeader {
		if err := c.checkFileHeaderValidity(file, ext, mimeType); err != nil {
			result.Passed = false
			result.ErrorMessage = err.Error()
			return result, nil
		}
	}

	// 检查文件名安全性
	if err := c.checkFileName(header.Filename); err != nil {
		result.Warnings = append(result.Warnings, err.Error())
	}

	return result, nil
}

// detectMimeType 检测文件的MIME类型
func (c *FileSecurityChecker) detectMimeType(file multipart.File) (string, error) {
	// 读取文件前512字节用于检测MIME类型
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	
	// 重置文件指针
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 检测MIME类型
	mimeType := http.DetectContentType(buffer[:n])
	return mimeType, nil
}

// checkFileHeaderValidity 检查文件头有效性
func (c *FileSecurityChecker) checkFileHeaderValidity(file multipart.File, ext, mimeType string) error {
	// 读取文件头进行验证
	buffer := make([]byte, 8)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}
	
	// 重置文件指针
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	header := buffer[:n]

	// 检查常见的文件签名
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

// checkFileName 检查文件名安全性
func (c *FileSecurityChecker) checkFileName(filename string) error {
	// 检查危险字符
	dangerousChars := []string{"..", "<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("文件名包含危险字符: %s", char)
		}
	}

	// 检查是否以点开头（隐藏文件）
	if strings.HasPrefix(filepath.Base(filename), ".") {
		return fmt.Errorf("文件名不能以点开始")
	}

	// 检查空文件名
	if filename == "" || filename == "." {
		return fmt.Errorf("文件名不能为空")
	}

	return nil
}

// 文件头检测函数

// isJPEGHeader 检查是否为JPEG文件头
func isJPEGHeader(header []byte) bool {
	if len(header) < 2 {
		return false
	}
	// JPEG文件以FF D8开始
	return header[0] == 0xFF && header[1] == 0xD8
}

// isPNGHeader 检查是否为PNG文件头
func isPNGHeader(header []byte) bool {
	if len(header) < 8 {
		return false
	}
	// PNG文件头: 89 50 4E 47 0D 0A 1A 0A
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return bytes.Equal(header[:8], pngSignature)
}

// isGIFHeader 检查是否为GIF文件头
func isGIFHeader(header []byte) bool {
	if len(header) < 6 {
		return false
	}
	// GIF87a 或 GIF89a
	return bytes.Equal(header[:4], []byte("GIF8")) && 
		(header[4] == '7' || header[4] == '9') && 
		header[5] == 'a'
}

// isPDFHeader 检查是否为PDF文件头
func isPDFHeader(header []byte) bool {
	if len(header) < 5 {
		return false
	}
	return bytes.Equal(header[:5], []byte("%PDF-"))
}

// isMP3Header 检查是否为MP3文件头
func isMP3Header(header []byte) bool {
	if len(header) < 2 {
		return false
	}
	// MP3文件以ID3标签或同步字开始
	// ID3标签: "ID3"
	// 同步字: 0xFFE
	if bytes.Equal(header[:3], []byte("ID3")) {
		return true
	}
	// 检查同步字 (11个1)
	return header[0] == 0xFF && (header[1]&0xE0) == 0xE0
}

// SanitizeFileName 清理文件名
func SanitizeFileName(filename string) string {
	// 移除路径
	filename = filepath.Base(filename)
	
	// 替换危险字符
	replacer := strings.NewReplacer(
		"..", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	
	filename = replacer.Replace(filename)
	
	// 确保不以点开头
	if strings.HasPrefix(filename, ".") {
		filename = "_" + filename
	}
	
	return filename
}
