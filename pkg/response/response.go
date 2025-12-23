package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// 业务状态码定义
const (
	CodeSuccess      = 200 // 成功
	CodeBadRequest   = 400 // 请求参数错误
	CodeUnauthorized = 401 // 未授权
	CodeForbidden    = 403 // 无权限
	CodeNotFound     = 404 // 资源不存在
	CodeServerError  = 500 // 服务器内部错误
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeBadRequest,
		Message: message,
		Data:    nil,
	})
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeUnauthorized,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeForbidden,
		Message: message,
		Data:    nil,
	})
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeNotFound,
		Message: message,
		Data:    nil,
	})
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeServerError,
		Message: message,
		Data:    nil,
	})
}

// Error 自定义错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
