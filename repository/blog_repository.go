package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type BlogRepository interface {
	Create(article *model.Article) error
	FindByID(id uint64) (*model.Article, error)
	FindList(page, pageSize int, tag string, status int) ([]model.Article, int64, error)
	FindByIDs(ids []uint64) ([]model.Article, error)
	FindUserDrafts(userID string, page, pageSize int) ([]model.Article, int64, error)
	Update(article *model.Article) error
	UpdateFields(id uint64, fields map[string]interface{}) error
	Delete(id uint64) error
	IncrementVisit(id uint64) error
	IncrementGoodCount(id uint64) error
	DecrementGoodCount(id uint64) error
	IncrementCommentCount(id int64) error
	IncrementFavoriteCount(id uint64) error
	DecrementFavoriteCount(id uint64) error
	CheckOwnership(id uint64, userID string) bool

	// 搜索
	SearchArticles(keyword string, page, pageSize int) ([]model.Article, int64, error)

	// Like相关
	CreateLike(like *model.ArticleStar) error
	DeleteLike(articleID uint64, userID string) error
	IsLiked(articleID uint64, userID string) bool
}

type blogRepository struct {
	db *gorm.DB
}

func NewBlogRepository(db *gorm.DB) BlogRepository {
	return &blogRepository{db: db}
}

// Create 创建文章
func (r *blogRepository) Create(article *model.Article) error {
	return r.db.Create(article).Error
}

// FindByID 根据ID查找文章
func (r *blogRepository) FindByID(id uint64) (*model.Article, error) {
	var article model.Article
	err := r.db.Where("id = ?", id).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// FindList 查找文章列表（支持分页、标签筛选和状态过滤）
func (r *blogRepository) FindList(page, pageSize int, tag string, status int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).Where("appear = ? AND status = ?", true, status)

	if tag != "" {
		query = query.Where("tag LIKE ?", "%"+tag+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// FindByIDs 根据ID列表查找文章
func (r *blogRepository) FindByIDs(ids []uint64) ([]model.Article, error) {
	var articles []model.Article
	if len(ids) == 0 {
		return articles, nil
	}

	err := r.db.Where("id IN ? AND status = ?", ids, model.ArticleStatusPublished).
		Order("create_time DESC").
		Find(&articles).Error
	return articles, err
}

// FindUserDrafts 查找用户的草稿列表
func (r *blogRepository) FindUserDrafts(userID string, page, pageSize int) ([]model.Article, int64, error) {
	var drafts []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).Where("user_id = ? AND status = ?", userID, model.ArticleStatusDraft)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("update_time DESC").Limit(pageSize).Offset(offset).Find(&drafts).Error; err != nil {
		return nil, 0, err
	}

	return drafts, total, nil
}

// Update 更新文章
func (r *blogRepository) Update(article *model.Article) error {
	return r.db.Save(article).Error
}

// UpdateFields 更新文章指定字段
func (r *blogRepository) UpdateFields(id uint64, fields map[string]interface{}) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除文章
func (r *blogRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Article{}, id).Error
}

// IncrementVisit 增加访问量
func (r *blogRepository) IncrementVisit(id uint64) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("visit", gorm.Expr("visit + 1")).Error
}

// IncrementGoodCount 增加点赞数
func (r *blogRepository) IncrementGoodCount(id uint64) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("good_count", gorm.Expr("good_count + 1")).Error
}

// DecrementGoodCount 减少点赞数
func (r *blogRepository) DecrementGoodCount(id uint64) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("good_count", gorm.Expr("good_count - 1")).Error
}

// IncrementCommentCount 增加评论数
func (r *blogRepository) IncrementCommentCount(id int64) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// CheckOwnership 检查文章所有权
func (r *blogRepository) CheckOwnership(id uint64, userID string) bool {
	var count int64
	r.db.Model(&model.Article{}).Where("id = ? AND user_id = ?", id, userID).Count(&count)
	return count > 0
}

// CreateLike 创建点赞
func (r *blogRepository) CreateLike(like *model.ArticleStar) error {
	return r.db.Create(like).Error
}

// DeleteLike 删除点赞
func (r *blogRepository) DeleteLike(articleID uint64, userID string) error {
	return r.db.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&model.ArticleStar{}).Error
}

// IsLiked 检查是否已点赞
func (r *blogRepository) IsLiked(articleID uint64, userID string) bool {
	var count int64
	r.db.Model(&model.ArticleStar{}).Where("article_id = ? AND user_id = ?", articleID, userID).Count(&count)
	return count > 0
}

// IncrementFavoriteCount 增加收藏数
func (r *blogRepository) IncrementFavoriteCount(id uint64) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error
}

// DecrementFavoriteCount 减少收藏数
func (r *blogRepository) DecrementFavoriteCount(id uint64) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count - 1")).Error
}

// SearchArticles 搜索文章（通过标题和内容）
func (r *blogRepository) SearchArticles(keyword string, page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	// 构建搜索条件
	query := r.db.Model(&model.Article{}).Where("status = ?", model.ArticleStatusPublished)

	if keyword != "" {
		// 搜索标题或内容包含关键词的文章
		likeKeyword := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", likeKeyword, likeKeyword)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按相关度排序（优先标题匹配，然后按更新时间排序）
	offset := (page - 1) * pageSize
	if err := query.Order("update_time DESC").Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}
