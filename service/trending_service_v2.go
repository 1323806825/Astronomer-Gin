package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"context"
	"fmt"
	"strconv"
	"time"

	redisLib "github.com/go-redis/redis/v8"
)

// TrendingServiceV2 热门榜单服务接口
type TrendingServiceV2 interface {
	// 获取热门文章榜单
	GetTrendingArticles(limit int) ([]model.Article, error)

	// 获取热门用户榜单
	GetTrendingUsers(limit int) ([]model.User, error)

	// 更新文章热度分数
	UpdateArticleScore(articleID uint64) error

	// 更新用户热度分数
	UpdateUserScore(userID uint64) error

	// 批量更新热门榜单（定时任务）
	RefreshTrendingData() error
}

type trendingServiceV2 struct {
	blogRepo    repository.BlogRepository
	userRepo    repository.UserRepository
	redisClient *redisLib.Client
	cacheHelper *util.CacheHelper
}

// NewTrendingServiceV2 创建热门榜单服务V2实例
func NewTrendingServiceV2(blogRepo repository.BlogRepository, userRepo repository.UserRepository) TrendingServiceV2 {
	return &trendingServiceV2{
		blogRepo:    blogRepo,
		userRepo:    userRepo,
		redisClient: redis.GetClient(),
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
	}
}

const (
	// Redis ZSet键
	TrendingArticlesKey = "trending:articles" // 热门文章
	TrendingUsersKey    = "trending:users"    // 热门用户

	// 热度计算权重
	VisitWeight    = 1.0  // 访问量权重
	LikeWeight     = 5.0  // 点赞权重
	CommentWeight  = 10.0 // 评论权重
	FavoriteWeight = 8.0  // 收藏权重
	FollowerWeight = 3.0  // 粉丝权重（用户）

	// 榜单过期时间（1小时）
	TrendingExpire = 3600
)

// GetTrendingArticles 获取热门文章榜单
func (s *trendingServiceV2) GetTrendingArticles(limit int) ([]model.Article, error) {
	ctx := context.Background()

	// 1. 参数验证
	if limit <= 0 || limit > 100 {
		limit = 20 // 默认20条
	}

	// 2. 从Redis ZSet获取热门文章ID列表（按分数降序）
	articleIDs, err := s.redisClient.ZRevRange(ctx, TrendingArticlesKey, 0, int64(limit-1)).Result()
	if err != nil || len(articleIDs) == 0 {
		// 如果Redis中没有数据，触发刷新
		if err := s.RefreshTrendingData(); err != nil {
			return nil, constant.ErrDatabaseQuery
		}
		// 重新获取
		articleIDs, err = s.redisClient.ZRevRange(ctx, TrendingArticlesKey, 0, int64(limit-1)).Result()
		if err != nil || len(articleIDs) == 0 {
			return []model.Article{}, nil
		}
	}

	// 3. 转换ID为uint64
	var ids []uint64
	for _, idStr := range articleIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return []model.Article{}, nil
	}

	// 4. 批量查询文章详情
	articles, err := s.blogRepo.FindByIDs(ids)
	if err != nil {
		return nil, constant.ErrDatabaseQuery
	}

	// 5. 按照热度排序返回
	sortedArticles := make([]model.Article, 0, len(ids))
	articleMap := make(map[uint64]model.Article)
	for _, article := range articles {
		articleMap[article.ID] = article
	}
	for _, id := range ids {
		if article, ok := articleMap[id]; ok {
			sortedArticles = append(sortedArticles, article)
		}
	}

	return sortedArticles, nil
}

// GetTrendingUsers 获取热门用户榜单
func (s *trendingServiceV2) GetTrendingUsers(limit int) ([]model.User, error) {
	ctx := context.Background()

	// 1. 参数验证
	if limit <= 0 || limit > 100 {
		limit = 20 // 默认20条
	}

	// 2. 从Redis ZSet获取热门用户ID列表（按分数降序）
	userIDs, err := s.redisClient.ZRevRange(ctx, TrendingUsersKey, 0, int64(limit-1)).Result()
	if err != nil || len(userIDs) == 0 {
		// 如果Redis中没有数据，触发刷新
		if err := s.RefreshTrendingData(); err != nil {
			return nil, constant.ErrDatabaseQuery
		}
		// 重新获取
		userIDs, err = s.redisClient.ZRevRange(ctx, TrendingUsersKey, 0, int64(limit-1)).Result()
		if err != nil || len(userIDs) == 0 {
			return []model.User{}, nil
		}
	}

	// 3. 查询用户详情
	var users []model.User
	for _, idStr := range userIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		user, err := s.userRepo.FindByID(strconv.Itoa(int(uint(id))))
		if err != nil {
			continue
		}

		// 数据脱敏
		user.Phone = util.MaskPhone(user.Phone)
		users = append(users, *user)
	}

	return users, nil
}

// UpdateArticleScore 更新文章热度分数
func (s *trendingServiceV2) UpdateArticleScore(articleID uint64) error {
	ctx := context.Background()

	// 1. 查询文章详情
	article, err := s.blogRepo.FindByID(articleID)
	if err != nil {
		return err
	}

	// 2. 只计算已发布的文章
	if article.Status != model.ArticleStatusPublished {
		return nil
	}

	// 3. 计算热度分数
	// score = visit * 1 + good_count * 5 + comment_count * 10 + favorite_count * 8
	score := float64(article.Visit)*VisitWeight +
		float64(article.GoodCount)*LikeWeight +
		float64(article.CommentCount)*CommentWeight +
		float64(article.FavoriteCount)*FavoriteWeight

	// 4. 时间衰减因子（文章越新，权重越高）
	// 超过7天的文章，分数逐渐衰减
	daysSincePublish := time.Since(article.CreateTime).Hours() / 24
	if daysSincePublish > 7 {
		decayFactor := 1.0 / (1.0 + (daysSincePublish-7)/7)
		score *= decayFactor
	}

	// 5. 更新到Redis ZSet
	err = s.redisClient.ZAdd(ctx, TrendingArticlesKey, &redisLib.Z{
		Score:  score,
		Member: fmt.Sprintf("%d", articleID),
	}).Err()

	if err != nil {
		return err
	}

	// 6. 设置过期时间
	s.redisClient.Expire(ctx, TrendingArticlesKey, time.Duration(TrendingExpire)*time.Second)

	return nil
}

// UpdateUserScore 更新用户热度分数
func (s *trendingServiceV2) UpdateUserScore(userID uint64) error {
	ctx := context.Background()

	// 1. 查询用户详情
	user, err := s.userRepo.FindByID(strconv.Itoa(int(uint(userID))))
	if err != nil {
		return err
	}

	// 2. 计算热度分数
	// score = following_count * 1 + followed_count * 3
	score := float64(user.FollowingCount) + float64(user.FollowedCount)*FollowerWeight

	// 3. 更新到Redis ZSet
	err = s.redisClient.ZAdd(ctx, TrendingUsersKey, &redisLib.Z{
		Score:  score,
		Member: fmt.Sprintf("%d", userID),
	}).Err()

	if err != nil {
		return err
	}

	// 4. 设置过期时间
	s.redisClient.Expire(ctx, TrendingUsersKey, time.Duration(TrendingExpire)*time.Second)

	return nil
}

// RefreshTrendingData 批量刷新热门榜单（定时任务）
func (s *trendingServiceV2) RefreshTrendingData() error {
	ctx := context.Background()

	// 1. 清空旧数据
	s.redisClient.Del(ctx, TrendingArticlesKey)
	s.redisClient.Del(ctx, TrendingUsersKey)

	// 2. 获取最近的文章（比如最近30天）
	// 这里简化实现，实际应该从数据库查询最近的文章
	articles, _, err := s.blogRepo.FindList(1, 500, "", model.ArticleStatusPublished)
	if err != nil {
		return err
	}

	// 3. 批量更新文章热度
	for _, article := range articles {
		// 只计算最近30天的文章
		if time.Since(article.CreateTime).Hours()/24 <= 30 {
			s.UpdateArticleScore(article.ID)
		}
	}

	// 4. 获取所有用户并更新热度（简化版，实际应该只更新活跃用户）
	// 这里我们只更新有粉丝的用户
	// 由于没有查询所有用户的方法，这里先跳过用户榜单的完整刷新
	// 实际使用时，可以在用户关注/取消关注时实时更新

	return nil
}
