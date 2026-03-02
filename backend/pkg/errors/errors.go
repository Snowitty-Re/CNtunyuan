package errors

import (
	"errors"
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 系统级错误 (1-999)
	CodeInternal     ErrorCode = 500 // 内部错误
	CodeDatabase     ErrorCode = 501 // 数据库错误
	CodeCache        ErrorCode = 502 // 缓存错误
	CodeNetwork      ErrorCode = 503 // 网络错误
	CodeTimeout      ErrorCode = 504 // 超时错误
	CodeUnavailable  ErrorCode = 503 // 服务不可用

	// 参数错误 (1000-1999)
	CodeInvalidParams   ErrorCode = 1000 // 参数错误
	CodeMissingParams   ErrorCode = 1001 // 缺少参数
	CodeInvalidFormat   ErrorCode = 1002 // 格式错误
	CodeOutOfRange      ErrorCode = 1003 // 超出范围

	// 认证授权错误 (2000-2999)
	CodeUnauthorized    ErrorCode = 2000 // 未认证
	CodeForbidden       ErrorCode = 2001 // 无权限
	CodeTokenExpired    ErrorCode = 2002 // Token过期
	CodeTokenInvalid    ErrorCode = 2003 // Token无效
	CodeLoginFailed     ErrorCode = 2004 // 登录失败
	CodeAccountLocked   ErrorCode = 2005 // 账号锁定
	CodeAccountDisabled ErrorCode = 2006 // 账号禁用

	// 资源错误 (3000-3999)
	CodeNotFound        ErrorCode = 3000 // 资源不存在
	CodeAlreadyExists   ErrorCode = 3001 // 资源已存在
	CodeConflict        ErrorCode = 3002 // 资源冲突
	CodeTooManyRequests ErrorCode = 3003 // 请求过于频繁

	// 业务错误 (4000-4999)
	CodeBusinessError   ErrorCode = 4000 // 业务错误
	CodeOperationFailed ErrorCode = 4001 // 操作失败
	CodeStateInvalid    ErrorCode = 4002 // 状态无效
)

// Error 自定义错误类型
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
	Err     error     `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// New 创建新错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf 格式化包装错误
func Wrapf(code ErrorCode, format string, err error, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// WithDetail 添加详情
func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail
	return e
}

// Is 判断错误类型
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// 快捷创建方法

func Internal(message string) *Error {
	return New(CodeInternal, message)
}

func Database(message string, err error) *Error {
	return Wrap(CodeDatabase, message, err)
}

func InvalidParams(message string) *Error {
	return New(CodeInvalidParams, message)
}

func MissingParams(param string) *Error {
	return New(CodeMissingParams, fmt.Sprintf("缺少必要参数: %s", param))
}

func Unauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(CodeForbidden, message)
}

func NotFound(resource string) *Error {
	return New(CodeNotFound, fmt.Sprintf("%s不存在", resource))
}

func AlreadyExists(resource string) *Error {
	return New(CodeAlreadyExists, fmt.Sprintf("%s已存在", resource))
}

func Conflict(message string) *Error {
	return New(CodeConflict, message)
}

func BusinessError(message string) *Error {
	return New(CodeBusinessError, message)
}

func Timeout(message string) *Error {
	return New(CodeTimeout, message)
}

func TooManyRequests(message string) *Error {
	return New(CodeTooManyRequests, message)
}

// IsErrorCode 判断错误码
func IsErrorCode(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return CodeInternal
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Message
	}
	return err.Error()
}
