package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
	"time"
)

// ArticleV3Repository 企业级文章Repository接口
type ArticleV3Repository interface {
	// ==================== 基础CRUD ====================
	Create(article *model.ArticleV3) error
	CreateWithContent(article *model.ArticleV3, content *model.ArticleContent) error
	FindByID(id uint64) (*model.ArticleV3, error)
	FindByIDWithContent(id uint64) (*model.ArticleV3, *model.ArticleContent, error)
	Update(article *model.ArticleV3) error
	UpdateFields(id uint64, fields map[string]interface{}) error
	Delete(id uint64) error     // 硬删除
	SoftDelete(id uint64) error // 软删除
	CheckOwnership(id uint64, userID string) bool

	// ==================== 列表查询 ====================
	// 查询文章列表（支持分类、专栏、状态、可见性、排序等复杂查询）
	FindList(params *ArticleQueryParams) ([]model.ArticleV3, int64, error)
	FindByUserID(userID uint64, page, pageSize int, status int8) ([]model.ArticleV3, int64, error)
	FindByCategoryID(categoryID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)
	FindByColumnID(columnID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)
	FindByTopicID(topicID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)
	FindByIDs(ids []uint64) ([]model.ArticleV3, error)

	// ==================== 推荐与排序 ====================
	FindFeaturedArticles(limit int) ([]model.ArticleV3, error)           // 精选文章
	FindHotArticles(limit int) ([]model.ArticleV3, error)                // 热门文章
	FindRecommendedArticles(limit int) ([]model.ArticleV3, error)        // 推荐文章
	FindByHotScore(page, pageSize int) ([]model.ArticleV3, int64, error) // 按热度排序

	// ==================== 统计字段更新 ====================
	IncrementViewCount(id uint64) error
	IncrementRealViewCount(id uint64) error
	IncrementLikeCount(id uint64) error
	DecrementLikeCount(id uint64) error
	IncrementCommentCount(id uint64) error
	DecrementCommentCount(id uint64) error
	IncrementShareCount(id uint64) error
	IncrementFavoriteCount(id uint64) error
	DecrementFavoriteCount(id uint64) error

	// ==================== 热度计算 ====================
	UpdateHotScore(id uint64, score float64) error
	UpdateQualityScore(id uint64, score float64) error
	BatchUpdateHotScores(updates map[uint64]float64) error

	// ==================== 内容管理 ====================
	CreateContent(content *model.ArticleContent) error
	UpdateContent(content *model.ArticleContent) error
	FindContentByArticleID(articleID uint64) (*model.ArticleContent, error)

	// ==================== 草稿管理 ====================
	CreateDraft(draft *model.ArticleDraft) error
	UpdateDraft(draft *model.ArticleDraft) error
	FindDraftByID(id uint64) (*model.ArticleDraft, error)
	FindUserDrafts(userID string, page, pageSize int) ([]model.ArticleDraft, int64, error)
	DeleteDraft(id uint64) error
	PublishDraft(draftID uint64, article *model.ArticleV3) error // 发布草稿

	// ==================== 版本历史 ====================
	CreateHistory(history *model.ArticleHistory) error
	FindHistoryByArticleID(articleID uint64) ([]model.ArticleHistory, error)
	FindHistoryByVersion(articleID uint64, version int) (*model.ArticleHistory, error)

	// ==================== 分类管理 ====================
	CreateCategory(category *model.ArticleCategory) error
	UpdateCategory(category *model.ArticleCategory) error
	FindCategoryByID(id uint64) (*model.ArticleCategory, error)
	FindAllCategories() ([]model.ArticleCategory, error)
	FindCategoriesByParentID(parentID uint64) ([]model.ArticleCategory, error)
	IncrementCategoryArticleCount(id uint64) error
	DecrementCategoryArticleCount(id uint64) error

	// ==================== 专栏管理 ====================
	CreateColumn(column *model.ArticleColumn) error
	UpdateColumn(column *model.ArticleColumn) error
	FindColumnByID(id uint64) (*model.ArticleColumn, error)
	FindColumnsByUserID(userID uint64, page, pageSize int) ([]model.ArticleColumn, int64, error)
	FindAllColumns(page, pageSize int) ([]model.ArticleColumn, int64, error)
	AddArticleToColumn(columnID, articleID uint64, sortOrder int) error
	RemoveArticleFromColumn(columnID, articleID uint64) error
	FindArticlesByColumnID(columnID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)
	IncrementColumnArticleCount(id uint64) error
	DecrementColumnArticleCount(id uint64) error
	IncrementColumnSubscriberCount(id uint64) error
	DecrementColumnSubscriberCount(id uint64) error
	// 专栏订阅相关
	SubscribeColumn(userID string, columnID uint64) error
	UnsubscribeColumn(userID string, columnID uint64) error
	IsColumnSubscribed(userID string, columnID uint64) bool
	GetUserSubscribedColumns(userID string, page, pageSize int) ([]model.ArticleColumn, int64, error)

	// ==================== 话题管理 ====================
	CreateTopic(topic *model.Topic) error
	UpdateTopic(topic *model.Topic) error
	FindTopicByID(id uint64) (*model.Topic, error)
	FindTopicByName(name string) (*model.Topic, error)
	FindAllTopics(page, pageSize int) ([]model.Topic, int64, error)
	FindHotTopics(limit int) ([]model.Topic, error)
	AddArticleToTopic(articleID, topicID uint64) error
	RemoveArticleFromTopic(articleID, topicID uint64) error
	FindTopicsByArticleID(articleID uint64) ([]model.Topic, error)
	IncrementTopicArticleCount(id uint64) error
	DecrementTopicArticleCount(id uint64) error
	IncrementTopicFollowCount(id uint64) error
	DecrementTopicFollowCount(id uint64) error
	// 话题关注相关
	FollowTopic(userID string, topicID uint64) error
	UnfollowTopic(userID string, topicID uint64) error
	IsTopicFollowed(userID string, topicID uint64) bool
	GetUserFollowedTopics(userID string, page, pageSize int) ([]model.Topic, int64, error)

	// ==================== 标签管理 ====================
	CreateTag(tag *model.ArticleTag) error
	FindOrCreateTag(name string) (*model.ArticleTag, error)
	FindTagByName(name string) (*model.ArticleTag, error)
	FindAllTags() ([]model.ArticleTag, error)
	FindHotTags(limit int) ([]model.ArticleTag, error)
	IncrementTagArticleCount(name string) error
	DecrementTagArticleCount(name string) error

	// ==================== 相关推荐 ====================
	CreateRelation(relation *model.ArticleRelation) error
	BatchCreateRelations(relations []model.ArticleRelation) error
	FindRelatedArticles(articleID uint64, limit int) ([]model.ArticleV3, error)

	// ==================== 统计详情 ====================
	CreateStatsDetail(stats *model.ArticleStatsDetail) error
	UpdateStatsDetail(stats *model.ArticleStatsDetail) error
	FindStatsDetailByArticleID(articleID uint64) (*model.ArticleStatsDetail, error)
	ResetDailyStats() error // 重置每日统计（定时任务）

	// ==================== 搜索 ====================
	SearchArticles(keyword string, page, pageSize int) ([]model.ArticleV3, int64, error)
}

// ArticleQueryParams 文章查询参数（复杂查询）
type ArticleQueryParams struct {
	CategoryID  uint64
	ColumnID    uint64
	TopicID     uint64
	UserID      string
	Status      int8
	Visibility  int8
	ContentType int8
	IsFeatured  *bool
	IsHot       *bool
	IsRecommend *bool
	Tag         string
	Keyword     string
	SortBy      string // create_time, hot_score, view_count, like_count
	SortOrder   string // asc, desc
	Page        int
	PageSize    int
}

type articleV3Repository struct {
	db *gorm.DB
}

// NewArticleV3Repository 创建ArticleV3Repository实例
func NewArticleV3Repository(db *gorm.DB) ArticleV3Repository {
	return &articleV3Repository{db: db}
}

// ==================== 基础CRUD实现 ====================

func (r *articleV3Repository) Create(article *model.ArticleV3) error {
	return r.db.Create(article).Error
}

func (r *articleV3Repository) CreateWithContent(article *model.ArticleV3, content *model.ArticleContent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建文章主记录
		if err := tx.Create(article).Error; err != nil {
			return err
		}

		// 2. 创建内容记录
		content.ArticleID = article.ID
		if err := tx.Create(content).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *articleV3Repository) FindByID(id uint64) (*model.ArticleV3, error) {
	var article model.ArticleV3
	err := r.db.Where("id = ? AND delete_time IS NULL", id).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleV3Repository) FindByIDWithContent(id uint64) (*model.ArticleV3, *model.ArticleContent, error) {
	var article model.ArticleV3
	var content model.ArticleContent

	// 查询文章
	if err := r.db.Where("id = ? AND delete_time IS NULL", id).First(&article).Error; err != nil {
		return nil, nil, err
	}

	// 查询内容
	if err := r.db.Where("article_id = ?", id).First(&content).Error; err != nil {
		return &article, nil, err
	}

	return &article, &content, nil
}

func (r *articleV3Repository) Update(article *model.ArticleV3) error {
	return r.db.Save(article).Error
}

func (r *articleV3Repository) UpdateFields(id uint64, fields map[string]interface{}) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).Updates(fields).Error
}

func (r *articleV3Repository) Delete(id uint64) error {
	return r.db.Delete(&model.ArticleV3{}, id).Error
}

func (r *articleV3Repository) SoftDelete(id uint64) error {
	now := time.Now()
	return r.db.Model(&model.ArticleV3{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.ArticleStatusDeleted,
			"delete_time": &now,
		}).Error
}

func (r *articleV3Repository) CheckOwnership(id uint64, userID string) bool {
	var count int64
	r.db.Model(&model.ArticleV3{}).Where("id = ? AND user_id = ?", id, userID).Count(&count)
	return count > 0
}

// ==================== 列表查询实现 ====================

func (r *articleV3Repository) FindList(params *ArticleQueryParams) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	// 构建查询
	query := r.db.Model(&model.ArticleV3{}).Where("delete_time IS NULL")

	// 应用过滤条件
	if params.CategoryID > 0 {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if params.ColumnID > 0 {
		query = query.Where("column_id = ?", params.ColumnID)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Status > 0 {
		query = query.Where("status = ?", params.Status)
	}
	if params.Visibility > 0 {
		query = query.Where("visibility = ?", params.Visibility)
	}
	if params.ContentType > 0 {
		query = query.Where("content_type = ?", params.ContentType)
	}
	if params.IsFeatured != nil {
		query = query.Where("is_featured = ?", *params.IsFeatured)
	}
	if params.IsHot != nil {
		query = query.Where("is_hot = ?", *params.IsHot)
	}
	if params.IsRecommend != nil {
		query = query.Where("is_recommend = ?", *params.IsRecommend)
	}
	if params.Tag != "" {
		query = query.Where("tags LIKE ?", "%"+params.Tag+"%")
	}
	if params.Keyword != "" {
		likeKeyword := "%" + params.Keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", likeKeyword, likeKeyword)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "create_time"
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// 分页
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	if err := query.Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *articleV3Repository) FindByUserID(userID uint64, page, pageSize int, status int8) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	query := r.db.Model(&model.ArticleV3{}).Where("user_id = ? AND delete_time IS NULL", userID)
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *articleV3Repository) FindByCategoryID(categoryID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	query := r.db.Model(&model.ArticleV3{}).
		Where("category_id = ? AND status = ? AND delete_time IS NULL", categoryID, model.ArticleStatusPublished)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("hot_score DESC, create_time DESC").Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *articleV3Repository) FindByColumnID(columnID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	// 通过关联表查询
	query := r.db.Table("article_v3").
		Joins("INNER JOIN article_column_rel ON article_v3.id = article_column_rel.article_id").
		Where("article_column_rel.column_id = ? AND article_v3.status = ? AND article_v3.delete_time IS NULL",
			columnID, model.ArticleStatusPublished)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("article_column_rel.sort_order ASC").
		Limit(pageSize).Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *articleV3Repository) FindByTopicID(topicID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	// 通过关联表查询
	query := r.db.Table("article_v3").
		Joins("INNER JOIN article_topic_rel ON article_v3.id = article_topic_rel.article_id").
		Where("article_topic_rel.topic_id = ? AND article_v3.status = ? AND article_v3.delete_time IS NULL",
			topicID, model.ArticleStatusPublished)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("article_v3.create_time DESC").
		Limit(pageSize).Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *articleV3Repository) FindByIDs(ids []uint64) ([]model.ArticleV3, error) {
	var articles []model.ArticleV3
	if len(ids) == 0 {
		return articles, nil
	}

	err := r.db.Where("id IN ? AND status = ? AND delete_time IS NULL", ids, model.ArticleStatusPublished).
		Order("create_time DESC").
		Find(&articles).Error
	return articles, err
}

// ==================== 推荐与排序实现 ====================

func (r *articleV3Repository) FindFeaturedArticles(limit int) ([]model.ArticleV3, error) {
	var articles []model.ArticleV3
	err := r.db.Where("is_featured = ? AND status = ? AND delete_time IS NULL",
		true, model.ArticleStatusPublished).
		Order("publish_time DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *articleV3Repository) FindHotArticles(limit int) ([]model.ArticleV3, error) {
	var articles []model.ArticleV3
	err := r.db.Where("is_hot = ? AND status = ? AND delete_time IS NULL",
		true, model.ArticleStatusPublished).
		Order("hot_score DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *articleV3Repository) FindRecommendedArticles(limit int) ([]model.ArticleV3, error) {
	var articles []model.ArticleV3
	err := r.db.Where("is_recommend = ? AND status = ? AND delete_time IS NULL",
		true, model.ArticleStatusPublished).
		Order("quality_score DESC, hot_score DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *articleV3Repository) FindByHotScore(page, pageSize int) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	query := r.db.Model(&model.ArticleV3{}).
		Where("status = ? AND delete_time IS NULL", model.ArticleStatusPublished)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("hot_score DESC, create_time DESC").
		Limit(pageSize).Offset(offset).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// ==================== 统计字段更新实现 ====================

func (r *articleV3Repository) IncrementViewCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *articleV3Repository) IncrementRealViewCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("real_view_count", gorm.Expr("real_view_count + 1")).Error
}

func (r *articleV3Repository) IncrementLikeCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *articleV3Repository) DecrementLikeCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *articleV3Repository) IncrementCommentCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *articleV3Repository) DecrementCommentCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
}

func (r *articleV3Repository) IncrementShareCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *articleV3Repository) IncrementFavoriteCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error
}

func (r *articleV3Repository) DecrementFavoriteCount(id uint64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
}

// ==================== 热度计算实现 ====================

func (r *articleV3Repository) UpdateHotScore(id uint64, score float64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("hot_score", score).Error
}

func (r *articleV3Repository) UpdateQualityScore(id uint64, score float64) error {
	return r.db.Model(&model.ArticleV3{}).Where("id = ?", id).
		UpdateColumn("quality_score", score).Error
}

func (r *articleV3Repository) BatchUpdateHotScores(updates map[uint64]float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for id, score := range updates {
			if err := tx.Model(&model.ArticleV3{}).Where("id = ?", id).
				UpdateColumn("hot_score", score).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== 内容管理实现 ====================

func (r *articleV3Repository) CreateContent(content *model.ArticleContent) error {
	return r.db.Create(content).Error
}

func (r *articleV3Repository) UpdateContent(content *model.ArticleContent) error {
	return r.db.Save(content).Error
}

func (r *articleV3Repository) FindContentByArticleID(articleID uint64) (*model.ArticleContent, error) {
	var content model.ArticleContent
	err := r.db.Where("article_id = ?", articleID).First(&content).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// ==================== 草稿管理实现 ====================

func (r *articleV3Repository) CreateDraft(draft *model.ArticleDraft) error {
	return r.db.Create(draft).Error
}

func (r *articleV3Repository) UpdateDraft(draft *model.ArticleDraft) error {
	return r.db.Save(draft).Error
}

func (r *articleV3Repository) FindDraftByID(id uint64) (*model.ArticleDraft, error) {
	var draft model.ArticleDraft
	err := r.db.Where("id = ?", id).First(&draft).Error
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *articleV3Repository) FindUserDrafts(userID string, page, pageSize int) ([]model.ArticleDraft, int64, error) {
	var drafts []model.ArticleDraft
	var total int64

	query := r.db.Model(&model.ArticleDraft{}).Where("user_id = ? AND is_published = ?", userID, false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("update_time DESC").Limit(pageSize).Offset(offset).Find(&drafts).Error; err != nil {
		return nil, 0, err
	}

	return drafts, total, nil
}

func (r *articleV3Repository) DeleteDraft(id uint64) error {
	return r.db.Delete(&model.ArticleDraft{}, id).Error
}

func (r *articleV3Repository) PublishDraft(draftID uint64, article *model.ArticleV3) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建文章
		if err := tx.Create(article).Error; err != nil {
			return err
		}

		// 2. 标记草稿为已发布
		if err := tx.Model(&model.ArticleDraft{}).Where("id = ?", draftID).
			Updates(map[string]interface{}{
				"is_published": true,
				"article_id":   article.ID,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// ==================== 版本历史实现 ====================

func (r *articleV3Repository) CreateHistory(history *model.ArticleHistory) error {
	return r.db.Create(history).Error
}

func (r *articleV3Repository) FindHistoryByArticleID(articleID uint64) ([]model.ArticleHistory, error) {
	var histories []model.ArticleHistory
	err := r.db.Where("article_id = ?", articleID).Order("version DESC").Find(&histories).Error
	return histories, err
}

func (r *articleV3Repository) FindHistoryByVersion(articleID uint64, version int) (*model.ArticleHistory, error) {
	var history model.ArticleHistory
	err := r.db.Where("article_id = ? AND version = ?", articleID, version).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ==================== 分类管理实现 ====================

func (r *articleV3Repository) CreateCategory(category *model.ArticleCategory) error {
	return r.db.Create(category).Error
}

func (r *articleV3Repository) UpdateCategory(category *model.ArticleCategory) error {
	return r.db.Save(category).Error
}

func (r *articleV3Repository) FindCategoryByID(id uint64) (*model.ArticleCategory, error) {
	var category model.ArticleCategory
	err := r.db.Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *articleV3Repository) FindAllCategories() ([]model.ArticleCategory, error) {
	var categories []model.ArticleCategory
	err := r.db.Where("is_show = ?", true).Order("parent_id ASC, sort_order ASC").Find(&categories).Error
	return categories, err
}

func (r *articleV3Repository) FindCategoriesByParentID(parentID uint64) ([]model.ArticleCategory, error) {
	var categories []model.ArticleCategory
	err := r.db.Where("parent_id = ? AND is_show = ?", parentID, true).
		Order("sort_order ASC").Find(&categories).Error
	return categories, err
}

func (r *articleV3Repository) IncrementCategoryArticleCount(id uint64) error {
	return r.db.Model(&model.ArticleCategory{}).Where("id = ?", id).
		UpdateColumn("article_count", gorm.Expr("article_count + 1")).Error
}

func (r *articleV3Repository) DecrementCategoryArticleCount(id uint64) error {
	return r.db.Model(&model.ArticleCategory{}).Where("id = ?", id).
		UpdateColumn("article_count", gorm.Expr("GREATEST(article_count - 1, 0)")).Error
}

// 剩余方法继续在下一个文件中...
// （由于内容过长，我会将专栏、话题、标签等剩余部分放在下一部分）
