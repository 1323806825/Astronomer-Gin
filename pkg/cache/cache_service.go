package cache

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
)

// CacheService Redis缓存服务
type CacheService struct {
	ctx context.Context
}

// NewCacheService 创建缓存服务
func NewCacheService() *CacheService {
	return &CacheService{
		ctx: context.Background(),
	}
}

// ==================== 文章缓存 ====================

// GetArticle 获取缓存的文章
func (s *CacheService) GetArticle(articleID uint64) (*model.Article, error) {
	key := fmt.Sprintf("article:%d", articleID)

	data, err := redis.GetClient().Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var article model.Article
	if err := json.Unmarshal([]byte(data), &article); err != nil {
		return nil, err
	}

	return &article, nil
}

// SetArticle 缓存文章（5分钟过期）
func (s *CacheService) SetArticle(article *model.Article) error {
	key := fmt.Sprintf("article:%d", article.ID)

	data, err := json.Marshal(article)
	if err != nil {
		return err
	}

	return redis.GetClient().Set(s.ctx, key, data, 5*time.Minute).Err()
}

// DeleteArticle 删除文章缓存
func (s *CacheService) DeleteArticle(articleID uint64) error {
	key := fmt.Sprintf("article:%d", articleID)
	return redis.GetClient().Del(s.ctx, key).Err()
}

// ==================== 热门文章缓存 ====================

// GetHotArticles 获取热门文章列表
func (s *CacheService) GetHotArticles() ([]model.Article, error) {
	key := "articles:hot:list"

	data, err := redis.GetClient().Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var articles []model.Article
	if err := json.Unmarshal([]byte(data), &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

// SetHotArticles 缓存热门文章（10分钟过期）
func (s *CacheService) SetHotArticles(articles []model.Article) error {
	key := "articles:hot:list"

	data, err := json.Marshal(articles)
	if err != nil {
		return err
	}

	return redis.GetClient().Set(s.ctx, key, data, 10*time.Minute).Err()
}

// ==================== 用户信息缓存 ====================

// GetUser 获取缓存的用户信息
func (s *CacheService) GetUser(userID uint64) (*model.User, error) {
	key := fmt.Sprintf("user:%d", userID)

	data, err := redis.GetClient().Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// SetUser 缓存用户信息（30分钟过期）
func (s *CacheService) SetUser(user *model.User) error {
	key := fmt.Sprintf("user:%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return redis.GetClient().Set(s.ctx, key, data, 30*time.Minute).Err()
}

// DeleteUser 删除用户缓存
func (s *CacheService) DeleteUser(userID uint64) error {
	key := fmt.Sprintf("user:%d", userID)
	return redis.GetClient().Del(s.ctx, key).Err()
}

// ==================== 评论缓存 ====================

// GetArticleComments 获取文章评论列表缓存
func (s *CacheService) GetArticleComments(articleID uint64, page int) ([]model.CommentParent, error) {
	key := fmt.Sprintf("comments:article:%d:page:%d", articleID, page)

	data, err := redis.GetClient().Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var comments []model.CommentParent
	if err := json.Unmarshal([]byte(data), &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// SetArticleComments 缓存文章评论列表（3分钟过期）
func (s *CacheService) SetArticleComments(articleID uint64, page int, comments []model.CommentParent) error {
	key := fmt.Sprintf("comments:article:%d:page:%d", articleID, page)

	data, err := json.Marshal(comments)
	if err != nil {
		return err
	}

	return redis.GetClient().Set(s.ctx, key, data, 3*time.Minute).Err()
}

// DeleteArticleComments 删除文章评论缓存（所有分页）
func (s *CacheService) DeleteArticleComments(articleID uint64) error {
	pattern := fmt.Sprintf("comments:article:%d:page:*", articleID)

	keys, err := redis.GetClient().Keys(s.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return redis.GetClient().Del(s.ctx, keys...).Err()
	}

	return nil
}

// ==================== 统计数据缓存 ====================

// GetArticleViewCount 获取文章浏览量
func (s *CacheService) GetArticleViewCount(articleID uint64) (int64, error) {
	key := fmt.Sprintf("article:view:%d", articleID)
	return redis.GetClient().Get(s.ctx, key).Int64()
}

// IncrArticleViewCount 增加文章浏览量（缓存中）
func (s *CacheService) IncrArticleViewCount(articleID uint64) error {
	key := fmt.Sprintf("article:view:%d", articleID)
	return redis.GetClient().Incr(s.ctx, key).Err()
}

// GetArticleLikeCount 获取文章点赞数
func (s *CacheService) GetArticleLikeCount(articleID uint64) (int64, error) {
	key := fmt.Sprintf("article:like:%d", articleID)
	return redis.GetClient().Get(s.ctx, key).Int64()
}

// IncrArticleLikeCount 增加文章点赞数
func (s *CacheService) IncrArticleLikeCount(articleID uint64) error {
	key := fmt.Sprintf("article:like:%d", articleID)
	return redis.GetClient().Incr(s.ctx, key).Err()
}

// DecrArticleLikeCount 减少文章点赞数
func (s *CacheService) DecrArticleLikeCount(articleID uint64) error {
	key := fmt.Sprintf("article:like:%d", articleID)
	return redis.GetClient().Decr(s.ctx, key).Err()
}

// ==================== 排行榜缓存 ====================

// AddToHotRanking 添加到热门排行榜（使用ZSet）
func (s *CacheService) AddToHotRanking(articleID uint64, score float64) error {
	key := "ranking:hot:articles"
	return redis.GetClient().ZAdd(s.ctx, key, &redisv8.Z{
		Score:  score,
		Member: articleID,
	}).Err()
}

// GetHotRanking 获取热门排行榜（前N名）
func (s *CacheService) GetHotRanking(limit int64) ([]uint64, error) {
	key := "ranking:hot:articles"

	result, err := redis.GetClient().ZRevRange(s.ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	articleIDs := make([]uint64, 0, len(result))
	for _, idStr := range result {
		var id uint64
		fmt.Sscanf(idStr, "%d", &id)
		articleIDs = append(articleIDs, id)
	}

	return articleIDs, nil
}

// ==================== 用户会话缓存 ====================

// SetUserSession 设置用户会话（Token）
func (s *CacheService) SetUserSession(userID uint64, token string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", token)
	return redis.GetClient().Set(s.ctx, key, userID, expiration).Err()
}

// GetUserSession 获取用户会话
func (s *CacheService) GetUserSession(token string) (uint64, error) {
	key := fmt.Sprintf("session:%s", token)
	return redis.GetClient().Get(s.ctx, key).Uint64()
}

// DeleteUserSession 删除用户会话（登出）
func (s *CacheService) DeleteUserSession(token string) error {
	key := fmt.Sprintf("session:%s", token)
	return redis.GetClient().Del(s.ctx, key).Err()
}

// ==================== 限流缓存 ====================

// CheckRateLimit 检查限流（滑动窗口）
func (s *CacheService) CheckRateLimit(key string, limit int64, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe := redis.GetClient().Pipeline()

	// 移除窗口外的记录
	pipe.ZRemRangeByScore(s.ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 计数当前窗口内的请求
	pipe.ZCard(s.ctx, key)

	// 添加当前请求
	pipe.ZAdd(s.ctx, key, &redisv8.Z{
		Score:  float64(now),
		Member: now,
	})

	// 设置过期时间
	pipe.Expire(s.ctx, key, window)

	cmds, err := pipe.Exec(s.ctx)
	if err != nil {
		return false, err
	}

	// 获取计数结果
	count := cmds[1].(*redisv8.IntCmd).Val()

	return count < limit, nil
}

// ==================== 缓存预热 ====================

// WarmupCache 缓存预热
func (s *CacheService) WarmupCache() error {
	// 这里可以实现缓存预热逻辑
	// 例如：预先加载热门文章、活跃用户等
	return nil
}

// ==================== 批量操作 ====================

// MGet 批量获取
func (s *CacheService) MGet(keys []string) ([]interface{}, error) {
	return redis.GetClient().MGet(s.ctx, keys...).Result()
}

// MSet 批量设置
func (s *CacheService) MSet(pairs map[string]interface{}) error {
	return redis.GetClient().MSet(s.ctx, pairs).Err()
}

// ==================== 缓存统计 ====================

// GetCacheStats 获取缓存统计信息
func (s *CacheService) GetCacheStats() (map[string]interface{}, error) {
	info, err := redis.GetClient().Info(s.ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"info": info,
	}, nil
}

// FlushCache 清空所有缓存（慎用！）
func (s *CacheService) FlushCache() error {
	return redis.GetClient().FlushDB(s.ctx).Err()
}
