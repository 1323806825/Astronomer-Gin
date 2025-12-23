package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

// ==================== 专栏管理实现 ====================

func (r *articleV3Repository) CreateColumn(column *model.ArticleColumn) error {
	return r.db.Create(column).Error
}

func (r *articleV3Repository) UpdateColumn(column *model.ArticleColumn) error {
	return r.db.Save(column).Error
}

func (r *articleV3Repository) FindColumnByID(id uint64) (*model.ArticleColumn, error) {
	var column model.ArticleColumn
	err := r.db.Where("id = ?", id).First(&column).Error
	if err != nil {
		return nil, err
	}
	return &column, nil
}

func (r *articleV3Repository) FindColumnsByUserID(userID uint64, page, pageSize int) ([]model.ArticleColumn, int64, error) {
	var columns []model.ArticleColumn
	var total int64

	query := r.db.Model(&model.ArticleColumn{}).Where("user_id = ? AND status = ?", userID, 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&columns).Error; err != nil {
		return nil, 0, err
	}

	return columns, total, nil
}

func (r *articleV3Repository) FindAllColumns(page, pageSize int) ([]model.ArticleColumn, int64, error) {
	var columns []model.ArticleColumn
	var total int64

	query := r.db.Model(&model.ArticleColumn{}).Where("status = ?", 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("subscriber_count DESC, create_time DESC").
		Limit(pageSize).Offset(offset).Find(&columns).Error; err != nil {
		return nil, 0, err
	}

	return columns, total, nil
}

func (r *articleV3Repository) AddArticleToColumn(columnID, articleID uint64, sortOrder int) error {
	rel := &model.ArticleColumnRel{
		ColumnID:  columnID,
		ArticleID: articleID,
		SortOrder: sortOrder,
	}
	return r.db.Create(rel).Error
}

func (r *articleV3Repository) RemoveArticleFromColumn(columnID, articleID uint64) error {
	return r.db.Where("column_id = ? AND article_id = ?", columnID, articleID).
		Delete(&model.ArticleColumnRel{}).Error
}

func (r *articleV3Repository) FindArticlesByColumnID(columnID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	return r.FindByColumnID(columnID, page, pageSize)
}

func (r *articleV3Repository) IncrementColumnArticleCount(id uint64) error {
	return r.db.Model(&model.ArticleColumn{}).Where("id = ?", id).
		UpdateColumn("article_count", gorm.Expr("article_count + 1")).Error
}

func (r *articleV3Repository) DecrementColumnArticleCount(id uint64) error {
	return r.db.Model(&model.ArticleColumn{}).Where("id = ?", id).
		UpdateColumn("article_count", gorm.Expr("GREATEST(article_count - 1, 0)")).Error
}

func (r *articleV3Repository) IncrementColumnSubscriberCount(id uint64) error {
	return r.db.Model(&model.ArticleColumn{}).Where("id = ?", id).
		UpdateColumn("subscriber_count", gorm.Expr("subscriber_count + 1")).Error
}

func (r *articleV3Repository) DecrementColumnSubscriberCount(id uint64) error {
	return r.db.Model(&model.ArticleColumn{}).Where("id = ?", id).
		UpdateColumn("subscriber_count", gorm.Expr("GREATEST(subscriber_count - 1, 0)")).Error
}

// ==================== 话题管理实现 ====================

func (r *articleV3Repository) CreateTopic(topic *model.Topic) error {
	return r.db.Create(topic).Error
}

func (r *articleV3Repository) UpdateTopic(topic *model.Topic) error {
	return r.db.Save(topic).Error
}

func (r *articleV3Repository) FindTopicByID(id uint64) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.Where("id = ?", id).First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *articleV3Repository) FindTopicByName(name string) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.Where("name = ?", name).First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *articleV3Repository) FindAllTopics(page, pageSize int) ([]model.Topic, int64, error) {
	var topics []model.Topic
	var total int64

	query := r.db.Model(&model.Topic{}).Where("status = ?", 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("article_count DESC, follow_count DESC").
		Limit(pageSize).Offset(offset).Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

func (r *articleV3Repository) FindHotTopics(limit int) ([]model.Topic, error) {
	var topics []model.Topic
	err := r.db.Where("is_hot = ? AND status = ?", true, 1).
		Order("article_count DESC, follow_count DESC").
		Limit(limit).
		Find(&topics).Error
	return topics, err
}

func (r *articleV3Repository) AddArticleToTopic(articleID, topicID uint64) error {
	rel := &model.ArticleTopicRel{
		ArticleID: articleID,
		TopicID:   topicID,
	}
	return r.db.Create(rel).Error
}

func (r *articleV3Repository) RemoveArticleFromTopic(articleID, topicID uint64) error {
	return r.db.Where("article_id = ? AND topic_id = ?", articleID, topicID).
		Delete(&model.ArticleTopicRel{}).Error
}

func (r *articleV3Repository) FindTopicsByArticleID(articleID uint64) ([]model.Topic, error) {
	var topics []model.Topic
	err := r.db.Table("topic").
		Joins("INNER JOIN article_topic_rel ON topic.id = article_topic_rel.topic_id").
		Where("article_topic_rel.article_id = ?", articleID).
		Find(&topics).Error
	return topics, err
}

func (r *articleV3Repository) IncrementTopicArticleCount(id uint64) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", id).
		UpdateColumn("article_count", gorm.Expr("article_count + 1")).Error
}

func (r *articleV3Repository) DecrementTopicArticleCount(id uint64) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", id).
		UpdateColumn("article_count", gorm.Expr("GREATEST(article_count - 1, 0)")).Error
}

func (r *articleV3Repository) IncrementTopicFollowCount(id uint64) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", id).
		UpdateColumn("follow_count", gorm.Expr("follow_count + 1")).Error
}

func (r *articleV3Repository) DecrementTopicFollowCount(id uint64) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", id).
		UpdateColumn("follow_count", gorm.Expr("GREATEST(follow_count - 1, 0)")).Error
}

// ==================== 标签管理实现 ====================

func (r *articleV3Repository) CreateTag(tag *model.ArticleTag) error {
	return r.db.Create(tag).Error
}

func (r *articleV3Repository) FindOrCreateTag(name string) (*model.ArticleTag, error) {
	var tag model.ArticleTag

	// 先尝试查找
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err == nil {
		return &tag, nil
	}

	// 如果不存在则创建
	if err == gorm.ErrRecordNotFound {
		tag = model.ArticleTag{Name: name}
		if err := r.db.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	}

	return nil, err
}

func (r *articleV3Repository) FindTagByName(name string) (*model.ArticleTag, error) {
	var tag model.ArticleTag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *articleV3Repository) FindAllTags() ([]model.ArticleTag, error) {
	var tags []model.ArticleTag
	err := r.db.Order("article_count DESC, follow_count DESC").Find(&tags).Error
	return tags, err
}

func (r *articleV3Repository) FindHotTags(limit int) ([]model.ArticleTag, error) {
	var tags []model.ArticleTag
	err := r.db.Where("is_hot = ?", true).
		Order("article_count DESC, follow_count DESC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (r *articleV3Repository) IncrementTagArticleCount(name string) error {
	return r.db.Model(&model.ArticleTag{}).Where("name = ?", name).
		UpdateColumn("article_count", gorm.Expr("article_count + 1")).Error
}

func (r *articleV3Repository) DecrementTagArticleCount(name string) error {
	return r.db.Model(&model.ArticleTag{}).Where("name = ?", name).
		UpdateColumn("article_count", gorm.Expr("GREATEST(article_count - 1, 0)")).Error
}

// ==================== 相关推荐实现 ====================

func (r *articleV3Repository) CreateRelation(relation *model.ArticleRelation) error {
	return r.db.Create(relation).Error
}

func (r *articleV3Repository) BatchCreateRelations(relations []model.ArticleRelation) error {
	if len(relations) == 0 {
		return nil
	}
	return r.db.Create(&relations).Error
}

func (r *articleV3Repository) FindRelatedArticles(articleID uint64, limit int) ([]model.ArticleV3, error) {
	var articles []model.ArticleV3

	// 通过关联表查询相关文章
	err := r.db.Table("article_v3").
		Joins("INNER JOIN article_relation ON article_v3.id = article_relation.related_article_id").
		Where("article_relation.article_id = ? AND article_v3.status = ? AND article_v3.delete_time IS NULL",
			articleID, model.ArticleStatusPublished).
		Order("article_relation.relevance_score DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// ==================== 统计详情实现 ====================

func (r *articleV3Repository) CreateStatsDetail(stats *model.ArticleStatsDetail) error {
	return r.db.Create(stats).Error
}

func (r *articleV3Repository) UpdateStatsDetail(stats *model.ArticleStatsDetail) error {
	return r.db.Save(stats).Error
}

func (r *articleV3Repository) FindStatsDetailByArticleID(articleID uint64) (*model.ArticleStatsDetail, error) {
	var stats model.ArticleStatsDetail
	err := r.db.Where("article_id = ?", articleID).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *articleV3Repository) ResetDailyStats() error {
	// 重置所有文章的今日统计数据
	return r.db.Model(&model.ArticleStatsDetail{}).
		Updates(map[string]interface{}{
			"today_view_count":    0,
			"today_like_count":    0,
			"today_comment_count": 0,
			"today_share_count":   0,
		}).Error
}

// ==================== 搜索实现 ====================

func (r *articleV3Repository) SearchArticles(keyword string, page, pageSize int) ([]model.ArticleV3, int64, error) {
	var articles []model.ArticleV3
	var total int64

	query := r.db.Model(&model.ArticleV3{}).
		Where("status = ? AND delete_time IS NULL", model.ArticleStatusPublished)

	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR keywords LIKE ?",
			likeKeyword, likeKeyword, likeKeyword)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按相关度和时间排序
	offset := (page - 1) * pageSize
	if err := query.Order("hot_score DESC, create_time DESC").
		Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// ==================== 专栏订阅实现 ====================

// SubscribeColumn 订阅专栏
func (r *articleV3Repository) SubscribeColumn(userID string, columnID uint64) error {
	subscription := &model.ColumnSubscription{
		UserID:   userID,
		ColumnID: columnID,
	}
	return r.db.Create(subscription).Error
}

// UnsubscribeColumn 取消订阅专栏
func (r *articleV3Repository) UnsubscribeColumn(userID string, columnID uint64) error {
	return r.db.Where("user_id = ? AND column_id = ?", userID, columnID).
		Delete(&model.ColumnSubscription{}).Error
}

// IsColumnSubscribed 检查是否已订阅专栏
func (r *articleV3Repository) IsColumnSubscribed(userID string, columnID uint64) bool {
	var count int64
	r.db.Model(&model.ColumnSubscription{}).
		Where("user_id = ? AND column_id = ?", userID, columnID).
		Count(&count)
	return count > 0
}

// GetUserSubscribedColumns 获取用户订阅的专栏列表
func (r *articleV3Repository) GetUserSubscribedColumns(userID string, page, pageSize int) ([]model.ArticleColumn, int64, error) {
	var subscriptions []model.ColumnSubscription
	var total int64

	// 获取总数
	if err := r.db.Model(&model.ColumnSubscription{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询订阅记录
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&subscriptions).Error; err != nil {
		return nil, 0, err
	}

	// 提取专栏ID
	columnIDs := make([]uint64, len(subscriptions))
	for i, sub := range subscriptions {
		columnIDs[i] = sub.ColumnID
	}

	if len(columnIDs) == 0 {
		return []model.ArticleColumn{}, 0, nil
	}

	// 查询专栏详情
	var columns []model.ArticleColumn
	if err := r.db.Where("id IN ?", columnIDs).Find(&columns).Error; err != nil {
		return nil, 0, err
	}

	return columns, total, nil
}

// ==================== 话题关注实现 ====================

// FollowTopic 关注话题
func (r *articleV3Repository) FollowTopic(userID string, topicID uint64) error {
	follow := &model.TopicFollow{
		UserID:  userID,
		TopicID: topicID,
	}
	return r.db.Create(follow).Error
}

// UnfollowTopic 取消关注话题
func (r *articleV3Repository) UnfollowTopic(userID string, topicID uint64) error {
	return r.db.Where("user_id = ? AND topic_id = ?", userID, topicID).
		Delete(&model.TopicFollow{}).Error
}

// IsTopicFollowed 检查是否已关注话题
func (r *articleV3Repository) IsTopicFollowed(userID string, topicID uint64) bool {
	var count int64
	r.db.Model(&model.TopicFollow{}).
		Where("user_id = ? AND topic_id = ?", userID, topicID).
		Count(&count)
	return count > 0
}

// GetUserFollowedTopics 获取用户关注的话题列表
func (r *articleV3Repository) GetUserFollowedTopics(userID string, page, pageSize int) ([]model.Topic, int64, error) {
	var follows []model.TopicFollow
	var total int64

	// 获取总数
	if err := r.db.Model(&model.TopicFollow{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询关注记录
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	// 提取话题ID
	topicIDs := make([]uint64, len(follows))
	for i, follow := range follows {
		topicIDs[i] = follow.TopicID
	}

	if len(topicIDs) == 0 {
		return []model.Topic{}, 0, nil
	}

	// 查询话题详情
	var topics []model.Topic
	if err := r.db.Where("id IN ?", topicIDs).Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}
