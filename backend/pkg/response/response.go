package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// New 创建响应
func New(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, New(0, "success", data))
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, New(0, message, data))
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, New(0, "created", data))
}

// NoContent 无内容响应
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, New(code, message, nil))
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "请求参数错误"
	}
	c.JSON(http.StatusBadRequest, New(http.StatusBadRequest, message, nil))
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "请先登录"
	}
	c.JSON(http.StatusUnauthorized, New(http.StatusUnauthorized, message, nil))
}

// Forbidden 禁止访问
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "权限不足"
	}
	c.JSON(http.StatusForbidden, New(http.StatusForbidden, message, nil))
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	c.JSON(http.StatusNotFound, New(http.StatusNotFound, message, nil))
}

// Conflict 资源冲突
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = "资源已存在"
	}
	c.JSON(http.StatusConflict, New(http.StatusConflict, message, nil))
}

// InternalServerError 服务器内部错误
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	c.JSON(http.StatusInternalServerError, New(http.StatusInternalServerError, message, nil))
}

// PageResult 分页结果
type PageResult[T any] struct {
	List       []T   `json:"list"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// Page 分页响应
func Page[T any](c *gin.Context, list []T, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, PageResult[T]{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

// GetOffset 获取偏移量
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}
