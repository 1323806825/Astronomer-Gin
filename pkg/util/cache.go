package util

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheHelper Redis缓存辅助类
type CacheHelper struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheHelper 创建缓存助手
func NewCacheHelper(client *redis.Client) *CacheHelper {
	return &CacheHelper{
		client: client,
		ctx:    context.Background(),
	}
}

// Get 获取缓存（自动反序列化）
func (c *CacheHelper) Get(key string, dest interface{}) error {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// GetString 获取字符串缓存
func (c *CacheHelper) GetString(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

// Set 设置缓存（自动序列化）
func (c *CacheHelper) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(c.ctx, key, data, expiration).Err()
}

// SetString 设置字符串缓存
func (c *CacheHelper) SetString(key string, value string, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func (c *CacheHelper) Delete(keys ...string) error {
	return c.client.Del(c.ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *CacheHelper) Exists(key string) bool {
	result, _ := c.client.Exists(c.ctx, key).Result()
	return result > 0
}

// Incr 自增
func (c *CacheHelper) Incr(key string) (int64, error) {
	return c.client.Incr(c.ctx, key).Result()
}

// Decr 自减
func (c *CacheHelper) Decr(key string) (int64, error) {
	return c.client.Decr(c.ctx, key).Result()
}

// Expire 设置过期时间
func (c *CacheHelper) Expire(key string, expiration time.Duration) error {
	return c.client.Expire(c.ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *CacheHelper) TTL(key string) (time.Duration, error) {
	return c.client.TTL(c.ctx, key).Result()
}

// GetOrSet 获取缓存，不存在则执行回调函数并设置
// 用于实现缓存穿透保护
func (c *CacheHelper) GetOrSet(key string, dest interface{}, expiration time.Duration, loader func() (interface{}, error)) error {
	// 先尝试从缓存获取
	err := c.Get(key, dest)
	if err == nil {
		return nil
	}

	// 缓存未命中，执行加载函数
	if err == redis.Nil {
		data, err := loader()
		if err != nil {
			return err
		}

		// 设置缓存
		if err := c.Set(key, data, expiration); err != nil {
			return err
		}

		// 反序列化到目标对象
		jsonData, _ := json.Marshal(data)
		return json.Unmarshal(jsonData, dest)
	}

	return err
}

// DeleteByPattern 根据模式删除缓存（谨慎使用）
func (c *CacheHelper) DeleteByPattern(pattern string) error {
	var cursor uint64
	var err error

	for {
		var keys []string
		keys, cursor, err = c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(c.ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

// HSet 设置哈希字段
func (c *CacheHelper) HSet(key, field string, value interface{}) error {
	return c.client.HSet(c.ctx, key, field, value).Err()
}

// HGet 获取哈希字段
func (c *CacheHelper) HGet(key, field string) (string, error) {
	return c.client.HGet(c.ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (c *CacheHelper) HGetAll(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}

// HDel 删除哈希字段
func (c *CacheHelper) HDel(key string, fields ...string) error {
	return c.client.HDel(c.ctx, key, fields...).Err()
}

// ZAdd 添加到有序集合
func (c *CacheHelper) ZAdd(key string, score float64, member interface{}) error {
	return c.client.ZAdd(c.ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRevRange 按分数倒序获取有序集合范围
func (c *CacheHelper) ZRevRange(key string, start, stop int64) ([]string, error) {
	return c.client.ZRevRange(c.ctx, key, start, stop).Result()
}

// ZRem 从有序集合删除成员
func (c *CacheHelper) ZRem(key string, members ...interface{}) error {
	return c.client.ZRem(c.ctx, key, members...).Err()
}
