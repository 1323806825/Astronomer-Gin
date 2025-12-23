package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type LikeRepository interface {
	// 文章点赞
	LikeArticle(userID string, articleID uint64) error
	UnlikeArticle(userID string, articleID uint64) error
	IsArticleLiked(userID string, articleID uint64) bool
	GetArticleLikeCount(articleID uint64) (int64, error)
	GetUserLikedArticles(userID string, page, pageSize int) ([]uint64, int64, error)

	// 批量检查点赞状态
	BatchCheckArticleLiked(userID string, articleIDs []uint64) (map[uint64]bool, error)
}

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

// LikeArticle 点赞文章
func (r *likeRepository) LikeArticle(userID string, articleID uint64) error {
	like := &model.ArticleStar{
		UserID:    userID,
		ArticleID: uint(articleID),
	}
	return r.db.Create(like).Error
}

// UnlikeArticle 取消点赞
func (r *likeRepository) UnlikeArticle(userID string, articleID uint64) error {
	return r.db.Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(&model.ArticleStar{}).Error
}

// IsArticleLiked 检查是否已点赞
func (r *likeRepository) IsArticleLiked(userID string, articleID uint64) bool {
	var count int64
	r.db.Model(&model.ArticleStar{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Count(&count)
	return count > 0
}

// GetArticleLikeCount 获取文章点赞数
func (r *likeRepository) GetArticleLikeCount(articleID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.ArticleStar{}).
		Where("article_id = ?", articleID).
		Count(&count).Error
	return count, err
}

// GetUserLikedArticles 获取用户点赞的文章列表
func (r *likeRepository) GetUserLikedArticles(userID string, page, pageSize int) ([]uint64, int64, error) {
	var total int64
	var likes []model.ArticleStar

	// 获取总数
	if err := r.db.Model(&model.ArticleStar{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	// 提取文章ID
	articleIDs := make([]uint64, len(likes))
	for i, like := range likes {
		articleIDs[i] = uint64(like.ArticleID)
	}

	return articleIDs, total, nil
}

// BatchCheckArticleLiked 批量检查文章点赞状态
func (r *likeRepository) BatchCheckArticleLiked(userID string, articleIDs []uint64) (map[uint64]bool, error) {
	if len(articleIDs) == 0 {
		return make(map[uint64]bool), nil
	}

	var likes []model.ArticleStar
	if err := r.db.Where("user_id = ? AND article_id IN ?", userID, articleIDs).
		Find(&likes).Error; err != nil {
		return nil, err
	}

	result := make(map[uint64]bool)
	for _, articleID := range articleIDs {
		result[articleID] = false
	}
	for _, like := range likes {
		result[uint64(like.ArticleID)] = true
	}

	return result, nil
}
