package redis

import (
	"astronomer-gin/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	log.Println("Redis连接成功")
	return nil
}

// GetClient 获取Redis客户端
func GetClient() *redis.Client {
	return Client
}

// Set 设置键值对
func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Del 删除键
func Del(keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(keys ...string) (int64, error) {
	return Client.Exists(ctx, keys...).Result()
}

// Close 关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
