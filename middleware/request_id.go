package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求追踪中间件
// 为每个请求生成唯一的Request ID，用于日志追踪和问题排查
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取或生成新的Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 存储到context中
		c.Set("request_id", requestID)

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
