package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseCode 响应码
type ResponseCode int

const (
	CodeSuccess       ResponseCode = 200
	CodeCreated       ResponseCode = 201
	CodeBadRequest    ResponseCode = 400
	CodeUnauthorized  ResponseCode = 401
	CodeForbidden     ResponseCode = 403
	CodeNotFound      ResponseCode = 404
	CodeServerError   ResponseCode = 500
)

// Response 统一响应结构
type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data"`
}

// PageData 分页数据
type PageData struct {
	List       interface{} `json:"list"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应带消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeCreated,
		Message: "created",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code ResponseCode, message string) {
	c.JSON(int(code), Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	Error(c, CodeUnauthorized, message)
}

// Forbidden 禁止访问
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "forbidden"
	}
	Error(c, CodeForbidden, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "not found"
	}
	Error(c, CodeNotFound, message)
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	Error(c, CodeServerError, message)
}

// PageSuccess 分页成功响应
func PageSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData{
			List:       list,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}
