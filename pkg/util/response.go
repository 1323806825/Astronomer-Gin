package util

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构（企业级标准）
type Response struct {
	Code      int         `json:"code"`                 // 业务状态码
	Message   string      `json:"message"`              // 提示信息
	Data      interface{} `json:"data,omitempty"`       // 业务数据
	RequestID string      `json:"request_id,omitempty"` // 请求追踪ID
	Timestamp int64       `json:"timestamp"`            // 响应时间戳
}

// PageResponse 分页响应结构
type PageResponse struct {
	List      interface{} `json:"list"`       // 数据列表
	Total     int64       `json:"total"`      // 总数
	Page      int         `json:"page"`       // 当前页
	PageSize  int         `json:"page_size"`  // 每页大小
	TotalPage int         `json:"total_page"` // 总页数
}

// buildResponse 构建统一响应（内部方法）
func buildResponse(c *gin.Context, code int, message string, data interface{}) Response {
	requestID, _ := c.Get("request_id")
	resp := Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	if requestID != nil {
		resp.RequestID = requestID.(string)
	}
	return resp
}

// ==================== 成功响应 ====================

// Success 成功响应（无消息）
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, buildResponse(c, 200, "操作成功", data))
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, buildResponse(c, 200, message, data))
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPage := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPage++
	}

	pageData := PageResponse{
		List:      list,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}

	c.JSON(http.StatusOK, buildResponse(c, 200, "查询成功", pageData))
}

// ==================== 错误响应 ====================

// Error 通用错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, buildResponse(c, code, message, nil))
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, buildResponse(c, code, message, data))
}

// BadRequest 400 参数错误
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 40000, message, nil))
}

// Unauthorized 401 未授权
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 40100, message, nil))
}

// Forbidden 403 禁止访问
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 40300, message, nil))
}

// NotFound 404 资源未找到
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 40400, message, nil))
}

// Conflict 409 资源冲突
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 40900, message, nil))
}

// TooManyRequests 429 请求过于频繁
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 42900, message, nil))
}

// InternalServerError 500 服务器内部错误
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 50000, message, nil))
}

// ServiceUnavailable 503 服务不可用
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusOK, buildResponse(c, 50300, message, nil))
}
