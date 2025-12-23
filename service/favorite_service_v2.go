package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/database"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// FavoriteServiceV2 企业级收藏服务接口
type FavoriteServiceV2 interface {
	FavoriteArticle(userID string, articleID uint64) error
	UnfavoriteArticle(userID string, articleID uint64) error
	IsFavorited(userID string, articleID uint64) bool
	GetUserFavorites(userID string, page, pageSize int) ([]model.Article, int64, error)

	// 缓存管理
	RefreshFavoriteCache(userID string) error
	ClearFavoriteCache(userID string) error
}

type favoriteServiceV2 struct {
	favoriteRepo repository.FavoriteRepository
	blogRepo     repository.BlogRepository
	notifyRepo   repository.NotificationRepository
	cacheHelper  *util.CacheHelper
	db           *gorm.DB
}

func NewFavoriteServiceV2(favoriteRepo repository.FavoriteRepository, blogRepo repository.BlogRepository, notifyRepo repository.NotificationRepository) FavoriteServiceV2 {
	return &favoriteServiceV2{
		favoriteRepo: favoriteRepo,
		blogRepo:     blogRepo,
		notifyRepo:   notifyRepo,
		cacheHelper:  util.NewCacheHelper(redis.GetClient()),
		db:           database.GetDB(),
	}
}

// FavoriteArticle 收藏文章（企业级实现）
func (s *favoriteServiceV2) FavoriteArticle(userID string, articleID uint64) error {
	// 1. 检查文章是否存在
	article, err := s.blogRepo.FindByID(articleID)
	if err != nil {
		return constant.ErrArticleNotFound
	}

	// 2. 检查是否已收藏
	if s.favoriteRepo.IsFavorited(userID, articleID) {
		return constant.ErrAlreadyFavorited
	}

	// 3. 使用事务创建收藏记录并更新文章收藏数
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 3.1 创建收藏记录
		favorite := &model.UserFavorite{
			UserID:     userID,
			ArticleID:  articleID,
			CreateTime: time.Now(),
		}

		if err := tx.Create(favorite).Error; err != nil {
			return err
		}

		// 3.2 增加文章收藏数
		if err := tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return constant.ErrFavoriteFailed
	}

	// 4. 清除缓存
	s.ClearFavoriteCache(userID)

	// 5. 这里可以添加收藏通知（可选）
	_ = article

	return nil
}

// UnfavoriteArticle 取消收藏（企业级实现）
func (s *favoriteServiceV2) UnfavoriteArticle(userID string, articleID uint64) error {
	// 1. 检查是否已收藏
	if !s.favoriteRepo.IsFavorited(userID, articleID) {
		return constant.ErrNotFavorited
	}

	// 2. 使用事务删除收藏记录并更新文章收藏数
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 2.1 删除收藏记录
		if err := tx.Where("user_id = ? AND article_id = ?", userID, articleID).
			Delete(&model.UserFavorite{}).Error; err != nil {
			return err
		}

		// 2.2 减少文章收藏数
		if err := tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return constant.ErrUnfavoriteFailed
	}

	// 3. 清除缓存
	s.ClearFavoriteCache(userID)

	return nil
}

// IsFavorited 检查是否已收藏
func (s *favoriteServiceV2) IsFavorited(userID string, articleID uint64) bool {
	return s.favoriteRepo.IsFavorited(userID, articleID)
}

// GetUserFavorites 获取用户的收藏列表（带缓存）
func (s *favoriteServiceV2) GetUserFavorites(userID string, page, pageSize int) ([]model.Article, int64, error) {
	// 1. 参数验证
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 从缓存获取
	cacheKey := fmt.Sprintf("%s%d:page:%d:size:%d", constant.CacheKeyFavorite, userID, page, pageSize)

	type CachedData struct {
		Articles []model.Article
		Total    int64
	}

	var cached CachedData
	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&cached,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			articleIDs, total, err := s.favoriteRepo.GetUserFavorites(userID, page, pageSize)
			if err != nil {
				return nil, err
			}

			if len(articleIDs) == 0 {
				return CachedData{Articles: []model.Article{}, Total: total}, nil
			}

			// 根据ID列表查询文章详情
			articles, err := s.blogRepo.FindByIDs(articleIDs)
			if err != nil {
				return nil, err
			}

			return CachedData{Articles: articles, Total: total}, nil
		},
	)

	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return cached.Articles, cached.Total, nil
}

// RefreshFavoriteCache 刷新收藏缓存
func (s *favoriteServiceV2) RefreshFavoriteCache(userID string) error {
	// 清除所有相关缓存，下次请求时会重新加载
	return s.ClearFavoriteCache(userID)
}

// ClearFavoriteCache 清除收藏缓存
func (s *favoriteServiceV2) ClearFavoriteCache(userID string) error {
	// 删除该用户所有收藏相关缓存
	pattern := fmt.Sprintf("%s%d:*", constant.CacheKeyFavorite, userID)
	return s.cacheHelper.DeleteByPattern(pattern)
}
