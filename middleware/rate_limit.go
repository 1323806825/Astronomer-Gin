package middleware

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	MaxRequests int           // 最大请求数
	Window      time.Duration // 时间窗口
	KeyPrefix   string        // Redis键前缀
}

// RateLimit 通用限流中间件（基于Redis）
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户标识（优先用户ID，其次IP）
		userKey := getUserKey(c)
		redisKey := fmt.Sprintf("rate_limit:%s:%s", config.KeyPrefix, userKey)

		// 获取Redis客户端
		rdb := redis.GetClient()
		if rdb == nil {
			// Redis不可用时放行请求（降级策略）
			c.Next()
			return
		}

		// 获取当前请求数
		count, err := rdb.Get(c, redisKey).Int()
		if err != nil && err.Error() != "redis: nil" {
			// Redis错误时放行请求
			c.Next()
			return
		}

		// 检查是否超出限流
		if count >= config.MaxRequests {
			util.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		// 增加计数
		pipe := rdb.Pipeline()
		pipe.Incr(c, redisKey)
		if count == 0 {
			// 第一次请求，设置过期时间
			pipe.Expire(c, redisKey, config.Window)
		}
		_, err = pipe.Exec(c)
		if err != nil {
			// Redis错误时放行请求
			c.Next()
			return
		}

		c.Next()
	}
}

// LoginRateLimit 登录限流（5次/分钟）
func LoginRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: constant.RateLimitLogin,
		Window:      time.Minute,
		KeyPrefix:   "login",
	})
}

// RegisterRateLimit 注册限流（3次/小时）
func RegisterRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: constant.RateLimitRegister,
		Window:      time.Hour,
		KeyPrefix:   "register",
	})
}

// CommentRateLimit 评论限流（10次/分钟）
func CommentRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: constant.RateLimitComment,
		Window:      time.Minute,
		KeyPrefix:   "comment",
	})
}

// ArticleRateLimit 发文限流（5次/小时）
func ArticleRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: constant.RateLimitArticle,
		Window:      time.Hour,
		KeyPrefix:   "article",
	})
}

// FollowRateLimit 关注限流（20次/分钟）
func FollowRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: constant.RateLimitFollow,
		Window:      time.Minute,
		KeyPrefix:   "follow",
	})
}

// getUserKey 获取用户唯一标识
func getUserKey(c *gin.Context) string {
	// 优先使用用户phone（已登录用户）
	if phone, exists := c.Get("phone"); exists {
		return phone.(string)
	}
	// 否则使用IP地址
	return c.ClientIP()
}
