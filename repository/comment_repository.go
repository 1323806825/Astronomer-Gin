package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type CommentRepository interface {
	// 一级评论
	CreateParentComment(comment *model.CommentParent) error
	FindParentCommentsByArticleID(articleID string) ([]model.CommentParent, error)
	IncrementParentGoodCount(id int64) error
	DecrementParentGoodCount(id int64) error

	// 二级评论
	CreateSubComment(comment *model.CommentSubTwo) error
	FindSubCommentsByParentID(parentID string) ([]model.CommentSubTwo, error)
	IncrementSubGoodCount(id int64) error
	DecrementSubGoodCount(id int64) error

	// 一级评论点赞
	CreateParentLike(like *model.CommentParentLike) error
	DeleteParentLike(commentID int64, userID string) error
	IsParentLiked(commentID int64, userID string) bool

	// 二级评论点赞
	CreateSubLike(like *model.CommentSubTwoLike) error
	DeleteSubLike(commentID int64, userID string) error
	IsSubLiked(commentID int64, userID string) bool
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// CreateParentComment 创建一级评论
func (r *commentRepository) CreateParentComment(comment *model.CommentParent) error {
	return r.db.Create(comment).Error
}

// FindParentCommentsByArticleID 根据文章ID查找一级评论列表
func (r *commentRepository) FindParentCommentsByArticleID(articleID string) ([]model.CommentParent, error) {
	var comments []model.CommentParent
	err := r.db.Where("article_id = ?", articleID).Order("comment_time DESC").Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// IncrementParentGoodCount 增加一级评论点赞数
func (r *commentRepository) IncrementParentGoodCount(id int64) error {
	return r.db.Model(&model.CommentParent{}).Where("id = ?", id).UpdateColumn("good_count", gorm.Expr("good_count + 1")).Error
}

// DecrementParentGoodCount 减少一级评论点赞数
func (r *commentRepository) DecrementParentGoodCount(id int64) error {
	return r.db.Model(&model.CommentParent{}).Where("id = ?", id).UpdateColumn("good_count", gorm.Expr("good_count - 1")).Error
}

// CreateSubComment 创建二级评论
func (r *commentRepository) CreateSubComment(comment *model.CommentSubTwo) error {
	return r.db.Create(comment).Error
}

// FindSubCommentsByParentID 根据父评论ID查找二级评论列表
func (r *commentRepository) FindSubCommentsByParentID(parentID string) ([]model.CommentSubTwo, error) {
	var subComments []model.CommentSubTwo
	err := r.db.Where("parent_comment_id = ?", parentID).Order("comment_time DESC").Find(&subComments).Error
	if err != nil {
		return nil, err
	}
	return subComments, nil
}

// IncrementSubGoodCount 增加二级评论点赞数
func (r *commentRepository) IncrementSubGoodCount(id int64) error {
	return r.db.Model(&model.CommentSubTwo{}).Where("id = ?", id).UpdateColumn("good_count", gorm.Expr("good_count + 1")).Error
}

// DecrementSubGoodCount 减少二级评论点赞数
func (r *commentRepository) DecrementSubGoodCount(id int64) error {
	return r.db.Model(&model.CommentSubTwo{}).Where("id = ?", id).UpdateColumn("good_count", gorm.Expr("good_count - 1")).Error
}

// CreateParentLike 创建一级评论点赞
func (r *commentRepository) CreateParentLike(like *model.CommentParentLike) error {
	return r.db.Create(like).Error
}

// DeleteParentLike 删除一级评论点赞
func (r *commentRepository) DeleteParentLike(commentID int64, userID string) error {
	return r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&model.CommentParentLike{}).Error
}

// IsParentLiked 检查是否已点赞一级评论
func (r *commentRepository) IsParentLiked(commentID int64, userID string) bool {
	var count int64
	r.db.Model(&model.CommentParentLike{}).Where("comment_id = ? AND user_id = ?", commentID, userID).Count(&count)
	return count > 0
}

// CreateSubLike 创建二级评论点赞
func (r *commentRepository) CreateSubLike(like *model.CommentSubTwoLike) error {
	return r.db.Create(like).Error
}

// DeleteSubLike 删除二级评论点赞
func (r *commentRepository) DeleteSubLike(commentID int64, userID string) error {
	return r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&model.CommentSubTwoLike{}).Error
}

// IsSubLiked 检查是否已点赞二级评论
func (r *commentRepository) IsSubLiked(commentID int64, userID string) bool {
	var count int64
	r.db.Model(&model.CommentSubTwoLike{}).Where("comment_id = ? AND user_id = ?", commentID, userID).Count(&count)
	return count > 0
}
