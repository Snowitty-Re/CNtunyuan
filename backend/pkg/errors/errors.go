// Package errors 提供统一错误处理
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 通用错误码 (0-999)
	CodeSuccess           ErrorCode = 0
	CodeUnknown           ErrorCode = 1
	CodeInternal          ErrorCode = 2
	CodeInvalidParam      ErrorCode = 400
	CodeUnauthorized      ErrorCode = 401
	CodeForbidden         ErrorCode = 403
	CodeNotFound          ErrorCode = 404
	CodeMethodNotAllowed  ErrorCode = 405
	CodeTimeout           ErrorCode = 408
	CodeConflict          ErrorCode = 409
	CodeTooManyRequests   ErrorCode = 429
	CodeServiceUnavailable ErrorCode = 503

	// 业务错误码 (1000-9999)
	CodeUserNotFound      ErrorCode = 1000
	CodeUserExists        ErrorCode = 1001
	CodeInvalidPassword   ErrorCode = 1002
	CodeInvalidToken      ErrorCode = 1003
	CodeTokenExpired      ErrorCode = 1004
	CodeInvalidCaptcha    ErrorCode = 1005
	CodeAccountDisabled   ErrorCode = 1006
	CodeAccountLocked     ErrorCode = 1007

	// 组织相关 (2000-2099)
	CodeOrgNotFound       ErrorCode = 2000
	CodeOrgExists         ErrorCode = 2001
	CodeOrgHasChildren    ErrorCode = 2002
	CodeOrgHasUsers       ErrorCode = 2003

	// 案件相关 (3000-3099)
	CodeCaseNotFound      ErrorCode = 3000
	CodeCaseExists        ErrorCode = 3001
	CodeCaseClosed        ErrorCode = 3002

	// 任务相关 (4000-4099)
	CodeTaskNotFound      ErrorCode = 4000
	CodeTaskExists        ErrorCode = 4001
	CodeTaskAssigned      ErrorCode = 4002
	CodeTaskCompleted     ErrorCode = 4003

	// 方言相关 (5000-5099)
	CodeDialectNotFound   ErrorCode = 5000
	CodeDialectExists     ErrorCode = 5001

	// 文件相关 (6000-6099)
	CodeFileNotFound      ErrorCode = 6000
	CodeFileTooLarge      ErrorCode = 6001
	CodeInvalidFileType   ErrorCode = 6002
	CodeFileUploadFailed  ErrorCode = 6003
)

// HTTPStatus 返回错误码对应的 HTTP 状态码
func (c ErrorCode) HTTPStatus() int {
	switch c {
	case CodeSuccess:
		return http.StatusOK
	case CodeInvalidParam:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeConflict:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeInternal, CodeUnknown:
		return http.StatusInternalServerError
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		if c >= 1000 && c < 6000 {
			// 业务错误通常返回 200，通过 code 区分
			return http.StatusOK
		}
		return http.StatusInternalServerError
	}
}

// String 返回错误码描述
func (c ErrorCode) String() string {
	switch c {
	case CodeSuccess:
		return "success"
	case CodeUnknown:
		return "unknown error"
	case CodeInternal:
		return "internal server error"
	case CodeInvalidParam:
		return "invalid parameter"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeForbidden:
		return "forbidden"
	case CodeNotFound:
		return "not found"
	case CodeTimeout:
		return "request timeout"
	case CodeTooManyRequests:
		return "too many requests"
	case CodeUserNotFound:
		return "user not found"
	case CodeUserExists:
		return "user already exists"
	case CodeInvalidPassword:
		return "invalid password"
	case CodeInvalidToken:
		return "invalid token"
	case CodeTokenExpired:
		return "token expired"
	default:
		return "error"
	}
}

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
	Err     error     `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail 添加详细错误信息
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// WithError 添加原始错误
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// Is 检查错误码是否匹配
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// New 创建新错误
func New(code ErrorCode, message string) *AppError {
	if message == "" {
		message = code.String()
	}
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建格式化错误
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	if message == "" {
		message = code.String()
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf 包装格式化错误
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// IsCode 检查错误是否匹配指定错误码
func IsCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*AppError); ok {
		return e.Code == code
	}
	return false
}

// GetCode 获取错误码
func GetCode(err error) ErrorCode {
	if err == nil {
		return CodeSuccess
	}
	if e, ok := err.(*AppError); ok {
		return e.Code
	}
	return CodeUnknown
}

// 预定义常用错误
var (
	ErrInvalidParam     = New(CodeInvalidParam, "")
	ErrUnauthorized     = New(CodeUnauthorized, "")
	ErrForbidden        = New(CodeForbidden, "")
	ErrNotFound         = New(CodeNotFound, "")
	ErrInternal         = New(CodeInternal, "")
	ErrUserNotFound     = New(CodeUserNotFound, "")
	ErrUserExists       = New(CodeUserExists, "")
	ErrInvalidPassword  = New(CodeInvalidPassword, "")
	ErrInvalidToken     = New(CodeInvalidToken, "")
	ErrTokenExpired     = New(CodeTokenExpired, "")
	ErrAccountDisabled  = New(CodeAccountDisabled, "")
	ErrAccountLocked    = New(CodeAccountLocked, "")
	ErrTooManyRequests  = New(CodeTooManyRequests, "")
	ErrFileTooLarge     = New(CodeFileTooLarge, "")
	ErrInvalidFileType  = New(CodeInvalidFileType, "")
)
