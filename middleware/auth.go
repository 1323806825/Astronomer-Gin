package middleware

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/jwt"
	"astronomer-gin/pkg/util"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			// 尝试从查询参数获取token
			token = c.Query("token")
		}

		// 去除Bearer前缀
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			util.Unauthorized(c, constant.TokenInvalid)
			c.Abort()
			return
		}

		// 验证token
		claims, err := jwt.Verify(token)
		if err != nil {
			util.Unauthorized(c, constant.TokenInvalid)
			c.Abort()
			return
		}

		// 验证UserID是否存在
		if claims.UserID == "" {
			util.Unauthorized(c, constant.TokenInvalid)
			c.Abort()
			return
		}

		// 将用户信息存入上下文（直接使用JWT中的UserID，无需数据库查询）
		c.Set("phone", claims.Phone)
		c.Set("user_id", claims.UserID)
		c.Set("token", token)

		c.Next()
	}
}
