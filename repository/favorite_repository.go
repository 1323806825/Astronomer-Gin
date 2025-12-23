package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type FavoriteRepository interface {
	Create(favorite *model.UserFavorite) error
	Delete(userID string, articleID uint64) error
	IsFavorited(userID string, articleID uint64) bool
	GetUserFavorites(userID string, page, pageSize int) ([]uint64, int64, error)
	GetFavoriteCount(articleID uint64) (int64, error)
	GetUserFavoriteCount(userID string) (int64, error)
}

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

// Create 创建收藏
func (r *favoriteRepository) Create(favorite *model.UserFavorite) error {
	return r.db.Create(favorite).Error
}

// Delete 取消收藏
func (r *favoriteRepository) Delete(userID string, articleID uint64) error {
	return r.db.Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(&model.UserFavorite{}).Error
}

// IsFavorited 检查是否已收藏
func (r *favoriteRepository) IsFavorited(userID string, articleID uint64) bool {
	var count int64
	r.db.Model(&model.UserFavorite{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Count(&count)
	return count > 0
}

// GetUserFavorites 获取用户收藏的文章ID列表（分页）
func (r *favoriteRepository) GetUserFavorites(userID string, page, pageSize int) ([]uint64, int64, error) {
	var total int64
	var favorites []model.UserFavorite

	// 获取总数
	if err := r.db.Model(&model.UserFavorite{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	// 提取文章ID
	articleIDs := make([]uint64, len(favorites))
	for i, fav := range favorites {
		articleIDs[i] = fav.ArticleID
	}

	return articleIDs, total, nil
}

// GetFavoriteCount 获取文章的收藏数
func (r *favoriteRepository) GetFavoriteCount(articleID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserFavorite{}).
		Where("article_id = ?", articleID).
		Count(&count).Error
	return count, err
}

// GetUserFavoriteCount 获取用户的收藏总数
func (r *favoriteRepository) GetUserFavoriteCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserFavorite{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
