package middleware

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler 全局错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic堆栈
				requestID, _ := c.Get("request_id")
				stack := string(debug.Stack())

				// 使用日志记录（如果有配置zap）
				if logger, exists := c.Get("logger"); exists {
					if zapLogger, ok := logger.(*zap.Logger); ok {
						zapLogger.Error("Panic recovered",
							zap.String("request_id", fmt.Sprintf("%v", requestID)),
							zap.Any("error", err),
							zap.String("stack", stack),
						)
					}
				}

				// 返回统一错误响应
				util.InternalServerError(c, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()

		// 检查是否有业务错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 如果是业务错误类型
			if bizErr, ok := err.Err.(*constant.BizError); ok {
				util.Error(c, bizErr.Code, bizErr.Message)
				return
			}

			// 其他未知错误
			util.InternalServerError(c, err.Error())
		}
	}
}

// HandleBizError 处理业务错误（辅助函数）
// 在handler或service中使用：middleware.HandleBizError(c, constant.ErrUserNotExist)
func HandleBizError(c *gin.Context, bizErr *constant.BizError) {
	c.Error(bizErr)
	util.Error(c, bizErr.Code, bizErr.Message)
}
