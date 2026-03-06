// Package response 提供统一响应格式
package response

import (
	"net/http"

	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	if err == nil {
		Success(c, nil)
		return
	}

	// 检查是否是应用错误
	if appErr, ok := err.(*errors.AppError); ok {
		status := appErr.Code.HTTPStatus()
		c.JSON(status, Response{
			Code:    int(appErr.Code),
			Message: appErr.Message,
			Data:    appErr.Detail,
		})
		return
	}

	// 普通错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    int(errors.CodeInternal),
		Message: "internal server error",
		Data:    err.Error(),
	})
}

// ErrorCode 使用错误码返回错误
func ErrorCode(c *gin.Context, code errors.ErrorCode) {
	status := code.HTTPStatus()
	c.JSON(status, Response{
		Code:    int(code),
		Message: code.String(),
	})
}

// ErrorCodeWithMessage 使用错误码和自定义消息返回错误
func ErrorCodeWithMessage(c *gin.Context, code errors.ErrorCode, message string) {
	status := code.HTTPStatus()
	c.JSON(status, Response{
		Code:    int(code),
		Message: message,
	})
}

// ErrorCodeWithDetail 使用错误码、消息和详情返回错误
func ErrorCodeWithDetail(c *gin.Context, code errors.ErrorCode, message, detail string) {
	status := code.HTTPStatus()
	c.JSON(status, Response{
		Code:    int(code),
		Message: message,
		Data:    detail,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	ErrorCodeWithMessage(c, errors.CodeInvalidParam, message)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	ErrorCodeWithMessage(c, errors.CodeUnauthorized, message)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "forbidden"
	}
	ErrorCodeWithMessage(c, errors.CodeForbidden, message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "resource not found"
	}
	ErrorCodeWithMessage(c, errors.CodeNotFound, message)
}

// InternalServerError 500 错误
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	ErrorCodeWithMessage(c, errors.CodeInternal, message)
}

// TooManyRequests 429 限流错误
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "too many requests"
	}
	ErrorCodeWithMessage(c, errors.CodeTooManyRequests, message)
}

// Created 201 创建成功
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

// NoContent 204 无内容
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Conflict 409 冲突
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = "resource conflict"
	}
	ErrorCodeWithMessage(c, errors.CodeConflict, message)
}

// PaginatedData 分页数据结构
type PaginatedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Pagination 分页参数
type Pagination struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// DefaultPagination 返回默认分页参数
func DefaultPagination() Pagination {
	return Pagination{
		Page:     1,
		PageSize: 10,
	}
}

// Normalize 规范化分页参数
func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 返回偏移量
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 返回限制数
func (p *Pagination) Limit() int {
	return p.PageSize
}

// SuccessPaginated 成功响应带分页数据
func SuccessPaginated(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PaginatedData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}
