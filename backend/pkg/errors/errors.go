// Package errors 提供统一错误处理
package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 通用错误码 (0-999)
	CodeSuccess            ErrorCode = 0
	CodeUnknown            ErrorCode = 1
	CodeInternal           ErrorCode = 2
	CodeInvalidParam       ErrorCode = 400
	CodeUnauthorized       ErrorCode = 401
	CodeForbidden          ErrorCode = 403
	CodeNotFound           ErrorCode = 404
	CodeMethodNotAllowed   ErrorCode = 405
	CodeTimeout            ErrorCode = 408
	CodeConflict           ErrorCode = 409
	CodeTooManyRequests    ErrorCode = 429
	CodeServiceUnavailable ErrorCode = 503

	// 业务错误码 (1000-9999)
	CodeUserNotFound    ErrorCode = 1000
	CodeUserExists      ErrorCode = 1001
	CodeInvalidPassword ErrorCode = 1002
	CodeInvalidToken    ErrorCode = 1003
	CodeTokenExpired    ErrorCode = 1004
	CodeInvalidCaptcha  ErrorCode = 1005
	CodeAccountDisabled ErrorCode = 1006
	CodeAccountLocked   ErrorCode = 1007

	// 组织相关 (2000-2099)
	CodeOrgNotFound    ErrorCode = 2000
	CodeOrgExists      ErrorCode = 2001
	CodeOrgHasChildren ErrorCode = 2002
	CodeOrgHasUsers    ErrorCode = 2003

	// 案件相关 (3000-3099)
	CodeCaseNotFound ErrorCode = 3000
	CodeCaseExists   ErrorCode = 3001
	CodeCaseClosed   ErrorCode = 3002

	// 任务相关 (4000-4099)
	CodeTaskNotFound  ErrorCode = 4000
	CodeTaskExists    ErrorCode = 4001
	CodeTaskAssigned  ErrorCode = 4002
	CodeTaskCompleted ErrorCode = 4003

	// 方言相关 (5000-5099)
	CodeDialectNotFound ErrorCode = 5000
	CodeDialectExists   ErrorCode = 5001

	// 文件相关 (6000-6099)
	CodeFileNotFound     ErrorCode = 6000
	CodeFileTooLarge     ErrorCode = 6001
	CodeInvalidFileType  ErrorCode = 6002
	CodeFileUploadFailed ErrorCode = 6003

	// 工作流错误 (7000-7099)
	CodeWorkflowNotFound       ErrorCode = 7000
	CodeWorkflowInvalidState   ErrorCode = 7001
	CodeWorkflowTransition     ErrorCode = 7002
	CodeWorkflowApproval       ErrorCode = 7003
	CodeWorkflowInstanceExists ErrorCode = 7004
	CodeWorkflowNodeNotFound   ErrorCode = 7005

	// 权限错误 (8000-8099)
	CodePermissionDenied      ErrorCode = 8000
	CodeDataPermissionDenied  ErrorCode = 8001
	CodeFieldPermissionDenied ErrorCode = 8002
	CodeRoleNotFound          ErrorCode = 8003
	CodeRoleExists            ErrorCode = 8004

	// 缓存错误 (9000-9099)
	CodeCacheError     ErrorCode = 9000
	CodeCacheMiss      ErrorCode = 9001
	CodeCacheSetFailed ErrorCode = 9002

	// 审计错误 (9100-9199)
	CodeAuditLogFailed ErrorCode = 9100
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
	case CodeForbidden, CodePermissionDenied, CodeDataPermissionDenied, CodeFieldPermissionDenied:
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
		if c >= 1000 && c < 10000 {
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
	case CodePermissionDenied:
		return "permission denied"
	case CodeDataPermissionDenied:
		return "data access denied"
	case CodeFieldPermissionDenied:
		return "field access denied"
	case CodeWorkflowNotFound:
		return "workflow not found"
	case CodeWorkflowInvalidState:
		return "invalid workflow state"
	case CodeCacheError:
		return "cache error"
	case CodeAuditLogFailed:
		return "audit log failed"
	default:
		return "error"
	}
}

// ErrorContext 错误上下文信息
type ErrorContext struct {
	TraceID    string                 `json:"trace_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	OrgID      string                 `json:"org_id,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	ClientIP   string                 `json:"client_ip,omitempty"`
}

// NewErrorContext 创建错误上下文
func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		Timestamp: time.Now(),
		Extra:     make(map[string]interface{}),
	}
}

// WithTraceID 添加追踪ID
func (ec *ErrorContext) WithTraceID(traceID string) *ErrorContext {
	ec.TraceID = traceID
	return ec
}

// WithUserID 添加用户ID
func (ec *ErrorContext) WithUserID(userID string) *ErrorContext {
	ec.UserID = userID
	return ec
}

// WithOrgID 添加组织ID
func (ec *ErrorContext) WithOrgID(orgID string) *ErrorContext {
	ec.OrgID = orgID
	return ec
}

// WithRequestInfo 添加请求信息
func (ec *ErrorContext) WithRequestInfo(method, path string) *ErrorContext {
	ec.Method = method
	ec.Path = path
	return ec
}

// WithClientIP 添加客户端IP
func (ec *ErrorContext) WithClientIP(ip string) *ErrorContext {
	ec.ClientIP = ip
	return ec
}

// WithExtra 添加额外信息
func (ec *ErrorContext) WithExtra(key string, value interface{}) *ErrorContext {
	ec.Extra[key] = value
	return ec
}

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode    `json:"code"`
	Message string       `json:"message"`
	Detail  string       `json:"detail,omitempty"`
	Err     error        `json:"-"`
	Context *ErrorContext `json:"context,omitempty"`
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

// WithContext 添加上下文
func (e *AppError) WithContext(ctx *ErrorContext) *AppError {
	e.Context = ctx
	return e
}

// Is 检查错误码是否匹配
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// ShouldLog 判断是否应该记录日志
func (e *AppError) ShouldLog() bool {
	// 业务错误通常不需要记录错误日志
	if e.Code >= 1000 && e.Code < 5000 {
		return false
	}
	return true
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

// WrapWithContext 包装错误并添加上下文
func WrapWithContext(ctx context.Context, err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	appErr := Wrap(err, code, message)
	
	// 从 context 中提取上下文信息
	errorCtx := NewErrorContext()
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		errorCtx.WithTraceID(traceID)
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		errorCtx.WithUserID(userID)
	}
	if orgID, ok := ctx.Value("org_id").(string); ok {
		errorCtx.WithOrgID(orgID)
	}
	
	appErr.Context = errorCtx
	return appErr
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

// IsNotFound 判断是否是未找到错误
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	code := GetCode(err)
	return code == CodeNotFound || code == CodeUserNotFound || 
		   code == CodeOrgNotFound || code == CodeCaseNotFound ||
		   code == CodeTaskNotFound || code == CodeDialectNotFound ||
		   code == CodeFileNotFound || code == CodeWorkflowNotFound
}

// IsPermissionDenied 判断是否是权限错误
func IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	code := GetCode(err)
	return code == CodeForbidden || code == CodePermissionDenied ||
		   code == CodeDataPermissionDenied || code == CodeFieldPermissionDenied
}

// 预定义常用错误
var (
	ErrInvalidParam    = New(CodeInvalidParam, "")
	ErrUnauthorized    = New(CodeUnauthorized, "")
	ErrForbidden       = New(CodeForbidden, "")
	ErrNotFound        = New(CodeNotFound, "")
	ErrInternal        = New(CodeInternal, "")
	ErrUserNotFound    = New(CodeUserNotFound, "")
	ErrUserExists      = New(CodeUserExists, "")
	ErrInvalidPassword = New(CodeInvalidPassword, "")
	ErrInvalidToken    = New(CodeInvalidToken, "")
	ErrTokenExpired    = New(CodeTokenExpired, "")
	ErrAccountDisabled = New(CodeAccountDisabled, "")
	ErrAccountLocked   = New(CodeAccountLocked, "")
	ErrTooManyRequests = New(CodeTooManyRequests, "")
	ErrFileTooLarge    = New(CodeFileTooLarge, "")
	ErrInvalidFileType = New(CodeInvalidFileType, "")
	ErrPermissionDenied = New(CodePermissionDenied, "")
	ErrDataPermissionDenied = New(CodeDataPermissionDenied, "")
	
	// 工作流错误
	ErrWorkflowNotFound     = New(CodeWorkflowNotFound, "")
	ErrWorkflowInvalidState = New(CodeWorkflowInvalidState, "")
	ErrWorkflowTransition   = New(CodeWorkflowTransition, "")
	
	// 缓存错误
	ErrCacheError = New(CodeCacheError, "")
	ErrCacheMiss  = New(CodeCacheMiss, "")
)
