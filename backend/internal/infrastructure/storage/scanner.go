// Package storage 病毒扫描
package storage

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// VirusScanner 病毒扫描器接口
type VirusScanner interface {
	// Scan 扫描文件
	Scan(file io.Reader, filename string) (*ScanResult, error)
	// IsAvailable 扫描器是否可用
	IsAvailable() bool
	// GetName 获取扫描器名称
	GetName() string
}

// ScanResult 扫描结果
type ScanResult struct {
	Clean     bool   `json:"clean"`      // 是否干净
	Infected  bool   `json:"infected"`   // 是否感染
	VirusName string `json:"virus_name"` // 病毒名称
	Error     string `json:"error"`      // 错误信息
	Scanner   string `json:"scanner"`    // 扫描器名称
}

// NoOpScanner 空扫描器（默认，不进行扫描）
type NoOpScanner struct{}

// NewNoOpScanner 创建空扫描器
func NewNoOpScanner() VirusScanner {
	return &NoOpScanner{}
}

// Scan 扫描（总是返回干净）
func (s *NoOpScanner) Scan(file io.Reader, filename string) (*ScanResult, error) {
	return &ScanResult{
		Clean:   true,
		Scanner: "NoOp",
	}, nil
}

// IsAvailable 是否可用
func (s *NoOpScanner) IsAvailable() bool {
	return true
}

// GetName 获取名称
func (s *NoOpScanner) GetName() string {
	return "NoOp Scanner"
}

// ClamAVScanner ClamAV扫描器（通过TCP连接clamd守护进程）
type ClamAVScanner struct {
	address string
	timeout time.Duration
}

// NewClamAVScanner 创建ClamAV扫描器
func NewClamAVScanner(address string, timeoutSeconds int) VirusScanner {
	timeout := 30 * time.Second
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	return &ClamAVScanner{
		address: address,
		timeout: timeout,
	}
}

// Scan 扫描文件（通过clamd INSTREAM命令）
func (s *ClamAVScanner) Scan(file io.Reader, filename string) (*ScanResult, error) {
	conn, err := net.DialTimeout("tcp", s.address, s.timeout)
	if err != nil {
		return &ScanResult{
			Clean:   false,
			Scanner: "ClamAV",
			Error:   fmt.Sprintf("connection failed: %v", err),
		}, fmt.Errorf("failed to connect to clamd at %s: %w", s.address, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(s.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// 发送 INSTREAM 命令
	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return nil, fmt.Errorf("failed to send INSTREAM command: %w", err)
	}

	// 分块发送文件数据，每块最大 2048 字节
	buf := make([]byte, 2048)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			// 发送块长度（4字节大端序）
			chunkLen := uint32(n)
			lenBuf := []byte{
				byte(chunkLen >> 24),
				byte(chunkLen >> 16),
				byte(chunkLen >> 8),
				byte(chunkLen),
			}
			if _, err := conn.Write(lenBuf); err != nil {
				return nil, fmt.Errorf("failed to send chunk length: %w", err)
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return nil, fmt.Errorf("failed to send chunk data: %w", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("failed to read file: %w", readErr)
		}
	}

	// 发送结束标记（长度为0）
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return nil, fmt.Errorf("failed to send end marker: %w", err)
	}

	// 读取扫描结果
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	response := scanner.Text()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应: "stream: OK" 或 "stream: <virus_name> FOUND"
	result := &ScanResult{
		Scanner: "ClamAV",
	}

	response = strings.TrimSpace(response)
	response = strings.TrimSuffix(response, "\x00")

	if strings.HasSuffix(response, "OK") {
		result.Clean = true
		logger.Info("ClamAV scan clean", logger.String("file", filename))
	} else if strings.Contains(response, "FOUND") {
		result.Clean = false
		result.Infected = true
		// 提取病毒名称: "stream: Eicar-Signature FOUND" -> "Eicar-Signature"
		parts := strings.SplitN(response, ":", 2)
		if len(parts) == 2 {
			virusInfo := strings.TrimSpace(parts[1])
			virusInfo = strings.TrimSuffix(virusInfo, " FOUND")
			result.VirusName = strings.TrimSpace(virusInfo)
		}
		logger.Warn("ClamAV detected virus",
			logger.String("file", filename),
			logger.String("virus", result.VirusName),
		)
	} else if strings.Contains(response, "ERROR") {
		result.Clean = false
		result.Error = response
		return result, fmt.Errorf("clamd error: %s", response)
	} else {
		result.Clean = false
		result.Error = fmt.Sprintf("unexpected response: %s", response)
		return result, fmt.Errorf("unexpected clamd response: %s", response)
	}

	return result, nil
}

// IsAvailable 检查ClamAV服务是否可用
func (s *ClamAVScanner) IsAvailable() bool {
	conn, err := net.DialTimeout("tcp", s.address, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	// 发送PING命令验证clamd可用性
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return false
	}
	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return false
	}

	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	response := strings.TrimSpace(scanner.Text())
	response = strings.TrimSuffix(response, "\x00")
	return response == "PONG"
}

// GetName 获取名称
func (s *ClamAVScanner) GetName() string {
	return "ClamAV"
}

// SecurityScanResult 安全扫描完整结果
type SecurityScanResult struct {
	FileCheck    *SecurityCheckResult `json:"file_check"`
	VirusScan    *ScanResult          `json:"virus_scan"`
	Passed       bool                 `json:"passed"`
	ErrorMessage string               `json:"error_message,omitempty"`
}

// PerformSecurityCheck 执行完整安全扫描
func PerformSecurityCheck(
	fileHeader *multipart.FileHeader,
	securityChecker *FileSecurityChecker,
	virusScanner VirusScanner,
) (*SecurityScanResult, error) {
	result := &SecurityScanResult{
		Passed: true,
	}

	// 1. 文件安全检查
	fileCheck, err := securityChecker.CheckFile(fileHeader)
	if err != nil {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("文件检查失败: %v", err)
		return result, err
	}
	result.FileCheck = fileCheck

	if !fileCheck.Passed {
		result.Passed = false
		result.ErrorMessage = fileCheck.ErrorMessage
		return result, nil
	}

	// 2. 病毒扫描
	if virusScanner != nil && virusScanner.IsAvailable() {
		file, err := fileHeader.Open()
		if err != nil {
			result.Passed = false
			result.ErrorMessage = "无法打开文件进行病毒扫描"
			return result, err
		}
		defer file.Close()

		scanResult, err := virusScanner.Scan(file, fileHeader.Filename)
		if err != nil {
			result.Passed = false
			result.ErrorMessage = fmt.Sprintf("病毒扫描失败: %v", err)
			return result, err
		}
		result.VirusScan = scanResult

		if !scanResult.Clean {
			result.Passed = false
			if scanResult.Infected {
				result.ErrorMessage = fmt.Sprintf("检测到病毒: %s", scanResult.VirusName)
			} else {
				result.ErrorMessage = fmt.Sprintf("扫描错误: %s", scanResult.Error)
			}
			return result, nil
		}
	}

	return result, nil
}
