package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/repository"
	"fmt"
	"math"
	"time"
)

// ==================== 草稿管理实现 ====================

// SaveDraft 保存草稿
func (s *articleV3Service) SaveDraft(userID string, req *SaveDraftRequest) (*model.ArticleDraft, error) {
	draft := &model.ArticleDraft{
		UserID:        userID,
		Title:         req.Title,
		Summary:       req.Summary,
		Content:       req.Content,
		CoverImage:    req.CoverImage,
		CategoryID:    req.CategoryID,
		ColumnID:      req.ColumnID,
		Tags:          model.JSONStringList(req.Tags),
		Topics:        model.JSONStringList(req.Topics),
		AutoSaveCount: 1,
	}

	now := time.Now()
	draft.LastEditTime = &now

	if err := s.articleRepo.CreateDraft(draft); err != nil {
		return nil, fmt.Errorf("保存草稿失败: %w", err)
	}

	return draft, nil
}

// UpdateDraft 更新草稿
func (s *articleV3Service) UpdateDraft(draftID uint64, userID string, req *SaveDraftRequest) error {
	// 1. 检查草稿所有权
	draft, err := s.articleRepo.FindDraftByID(draftID)
	if err != nil {
		return fmt.Errorf("草稿不存在: %w", err)
	}

	if draft.UserID != userID {
		return constant.ErrPermissionDenied
	}

	// 2. 更新草稿
	draft.Title = req.Title
	draft.Summary = req.Summary
	draft.Content = req.Content
	draft.CoverImage = req.CoverImage
	draft.CategoryID = req.CategoryID
	draft.ColumnID = req.ColumnID
	draft.Tags = model.JSONStringList(req.Tags)
	draft.Topics = model.JSONStringList(req.Topics)
	draft.AutoSaveCount++

	now := time.Now()
	draft.LastEditTime = &now

	if err := s.articleRepo.UpdateDraft(draft); err != nil {
		return fmt.Errorf("更新草稿失败: %w", err)
	}

	return nil
}

// PublishDraft 发布草稿
func (s *articleV3Service) PublishDraft(draftID uint64, userID string) (*model.ArticleV3, error) {
	// 1. 获取草稿
	draft, err := s.articleRepo.FindDraftByID(draftID)
	if err != nil {
		return nil, fmt.Errorf("草稿不存在: %w", err)
	}

	if draft.UserID != userID {
		return nil, constant.ErrPermissionDenied
	}

	if draft.IsPublished {
		return nil, fmt.Errorf("草稿已发布")
	}

	// 2. 验证内容
	if err := s.validateArticleContent(draft.Title, draft.Content); err != nil {
		return nil, err
	}

	// 3. 创建文章
	now := time.Now()
	article := &model.ArticleV3{
		UserID:      userID,
		Title:       draft.Title,
		Summary:     draft.Summary,
		CoverImage:  draft.CoverImage,
		ContentType: model.ArticleContentTypeText,
		CategoryID:  draft.CategoryID,
		ColumnID:    draft.ColumnID,
		Tags:        draft.Tags,
		Topics:      draft.Topics,
		Status:      model.ArticleV3StatusPublished,
		Visibility:  model.ArticleVisibilityPublic,
		PublishTime: &now,
	}

	// 4. 使用事务发布
	if err := s.articleRepo.PublishDraft(draftID, article); err != nil {
		return nil, fmt.Errorf("发布草稿失败: %w", err)
	}

	// 5. 创建内容记录
	content := &model.ArticleContent{
		ArticleID: article.ID,
		Content:   draft.Content,
		WordCount: len([]rune(draft.Content)),
		ReadTime:  s.calculateReadTime(draft.Content),
	}
	s.articleRepo.CreateContent(content)

	// 6. 处理分类、话题、标签
	if article.CategoryID > 0 {
		s.articleRepo.IncrementCategoryArticleCount(article.CategoryID)
	}

	if len(draft.Topics) > 0 {
		s.handleTopics(article.ID, draft.Topics)
	}

	if len(draft.Tags) > 0 {
		s.handleTags(draft.Tags)
	}

	// 7. 创建历史版本
	s.createHistoryVersion(article.ID, article.Title, draft.Content, "从草稿发布", model.ChangeTypePublish, userID)

	return article, nil
}

// GetUserDrafts 获取用户草稿列表
func (s *articleV3Service) GetUserDrafts(userID string, page, pageSize int) ([]model.ArticleDraft, int64, error) {
	return s.articleRepo.FindUserDrafts(userID, page, pageSize)
}

// GetDraftDetail 获取草稿详情
func (s *articleV3Service) GetDraftDetail(draftID uint64, userID string) (*model.ArticleDraft, error) {
	draft, err := s.articleRepo.FindDraftByID(draftID)
	if err != nil {
		return nil, fmt.Errorf("草稿不存在: %w", err)
	}

	// 验证草稿所有权
	if draft.UserID != userID {
		return nil, constant.ErrPermissionDenied
	}

	return draft, nil
}

// DeleteDraft 删除草稿
func (s *articleV3Service) DeleteDraft(draftID uint64, userID string) error {
	draft, err := s.articleRepo.FindDraftByID(draftID)
	if err != nil {
		return fmt.Errorf("草稿不存在: %w", err)
	}

	if draft.UserID != userID {
		return constant.ErrPermissionDenied
	}

	return s.articleRepo.DeleteDraft(draftID)
}

// ==================== 分类管理实现 ====================

// CreateCategory 创建分类
func (s *articleV3Service) CreateCategory(name string, parentID uint64, icon string, sortOrder int) (*model.ArticleCategory, error) {
	category := &model.ArticleCategory{
		Name:      name,
		ParentID:  parentID,
		Icon:      icon,
		SortOrder: sortOrder,
		IsShow:    true,
	}

	if err := s.articleRepo.CreateCategory(category); err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	return category, nil
}

// GetCategoryTree 获取分类树
func (s *articleV3Service) GetCategoryTree() ([]CategoryTreeNode, error) {
	// 1. 获取所有分类
	categories, err := s.articleRepo.FindAllCategories()
	if err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	// 2. 构建树形结构
	categoryMap := make(map[uint64]*CategoryTreeNode)
	var roots []CategoryTreeNode

	// 第一次遍历：创建所有节点
	for _, cat := range categories {
		node := CategoryTreeNode{
			ArticleCategory: &cat,
			Children:        []CategoryTreeNode{},
		}
		categoryMap[cat.ID] = &node
	}

	// 第二次遍历：构建父子关系
	for _, cat := range categories {
		node := categoryMap[cat.ID]
		if cat.ParentID == 0 {
			// 根节点
			roots = append(roots, *node)
		} else {
			// 子节点
			if parent, exists := categoryMap[cat.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	return roots, nil
}

// GetArticlesByCategory 获取分类下的文章
func (s *articleV3Service) GetArticlesByCategory(categoryID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	return s.articleRepo.FindByCategoryID(categoryID, page, pageSize)
}

// ==================== 专栏管理实现 ====================

// CreateColumn 创建专栏
func (s *articleV3Service) CreateColumn(userID string, req *CreateColumnRequest) (*model.ArticleColumn, error) {
	column := &model.ArticleColumn{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		SortType:    req.SortType,
		Status:      1, // 正常
	}

	if err := s.articleRepo.CreateColumn(column); err != nil {
		return nil, fmt.Errorf("创建专栏失败: %w", err)
	}

	return column, nil
}

// UpdateColumn 更新专栏
func (s *articleV3Service) UpdateColumn(columnID uint64, userID string, req *UpdateColumnRequest) error {
	// 1. 检查所有权
	column, err := s.articleRepo.FindColumnByID(columnID)
	if err != nil {
		return fmt.Errorf("专栏不存在: %w", err)
	}

	if column.UserID != userID {
		return constant.ErrPermissionDenied
	}

	// 2. 更新字段
	if req.Name != nil {
		column.Name = *req.Name
	}
	if req.Description != nil {
		column.Description = *req.Description
	}
	if req.CoverImage != nil {
		column.CoverImage = *req.CoverImage
	}
	if req.SortType != nil {
		column.SortType = *req.SortType
	}
	if req.IsFinished != nil {
		column.IsFinished = *req.IsFinished
	}

	if err := s.articleRepo.UpdateColumn(column); err != nil {
		return fmt.Errorf("更新专栏失败: %w", err)
	}

	return nil
}

// AddArticleToColumn 添加文章到专栏
func (s *articleV3Service) AddArticleToColumn(columnID, articleID uint64, userID string, sortOrder int) error {
	// 1. 检查专栏所有权
	column, err := s.articleRepo.FindColumnByID(columnID)
	if err != nil {
		return fmt.Errorf("专栏不存在: %w", err)
	}

	if column.UserID != userID {
		return constant.ErrPermissionDenied
	}

	// 2. 检查文章所有权
	if !s.articleRepo.CheckOwnership(articleID, userID) {
		return constant.ErrPermissionDenied
	}

	// 3. 添加到专栏
	if err := s.articleRepo.AddArticleToColumn(columnID, articleID, sortOrder); err != nil {
		return fmt.Errorf("添加文章到专栏失败: %w", err)
	}

	// 4. 更新专栏文章数
	s.articleRepo.IncrementColumnArticleCount(columnID)

	// 5. 更新文章的column_id
	s.articleRepo.UpdateFields(articleID, map[string]interface{}{"column_id": columnID})

	return nil
}

// RemoveArticleFromColumn 从专栏移除文章
func (s *articleV3Service) RemoveArticleFromColumn(columnID, articleID uint64, userID string) error {
	// 1. 检查专栏所有权
	column, err := s.articleRepo.FindColumnByID(columnID)
	if err != nil {
		return fmt.Errorf("专栏不存在: %w", err)
	}

	if column.UserID != userID {
		return constant.ErrPermissionDenied
	}

	// 2. 从专栏移除
	if err := s.articleRepo.RemoveArticleFromColumn(columnID, articleID); err != nil {
		return fmt.Errorf("移除文章失败: %w", err)
	}

	// 3. 更新专栏文章数
	s.articleRepo.DecrementColumnArticleCount(columnID)

	// 4. 清除文章的column_id
	s.articleRepo.UpdateFields(articleID, map[string]interface{}{"column_id": 0})

	return nil
}

// GetColumnDetail 获取专栏详情
func (s *articleV3Service) GetColumnDetail(columnID uint64, userID string) (*ColumnDetailResponse, error) {
	// 1. 获取专栏
	column, err := s.articleRepo.FindColumnByID(columnID)
	if err != nil {
		return nil, fmt.Errorf("专栏不存在: %w", err)
	}

	// 2. 获取作者信息
	author, _ := s.userRepo.FindByID(column.UserID)
	authorInfo := &ArticleAuthorInfo{
		UserID:   column.UserID,
		Username: author.Username,
		Avatar:   author.Icon,
		Intro:    author.Intro,
	}

	// 3. 检查订阅状态（如果用户已登录）
	isSubscribed := false
	if userID != "" {
		isSubscribed = s.articleRepo.IsColumnSubscribed(userID, columnID)
	}

	return &ColumnDetailResponse{
		Column:       column,
		Author:       authorInfo,
		ArticleCount: column.ArticleCount,
		IsSubscribed: isSubscribed,
	}, nil
}

// GetColumnArticles 获取专栏文章列表
func (s *articleV3Service) GetColumnArticles(columnID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	return s.articleRepo.FindByColumnID(columnID, page, pageSize)
}

// SubscribeColumn 订阅专栏
func (s *articleV3Service) SubscribeColumn(columnID uint64, userID string) error {
	// 检查是否已订阅
	if s.articleRepo.IsColumnSubscribed(userID, columnID) {
		return fmt.Errorf("已经订阅过该专栏")
	}

	// 创建订阅记录
	if err := s.articleRepo.SubscribeColumn(userID, columnID); err != nil {
		return err
	}

	// 增加订阅数
	return s.articleRepo.IncrementColumnSubscriberCount(columnID)
}

// UnsubscribeColumn 取消订阅专栏
func (s *articleV3Service) UnsubscribeColumn(columnID uint64, userID string) error {
	// 删除订阅记录
	if err := s.articleRepo.UnsubscribeColumn(userID, columnID); err != nil {
		return err
	}

	// 减少订阅数
	return s.articleRepo.DecrementColumnSubscriberCount(columnID)
}

// ==================== 话题管理实现 ====================

// CreateTopic 创建话题
func (s *articleV3Service) CreateTopic(name, description, userID string) (*model.Topic, error) {
	// 检查话题是否已存在
	if existing, _ := s.articleRepo.FindTopicByName(name); existing != nil {
		return existing, nil
	}

	topic := &model.Topic{
		Name:        name,
		Description: description,
		CreatorID:   userID,
		Status:      1, // 正常
	}

	if err := s.articleRepo.CreateTopic(topic); err != nil {
		return nil, fmt.Errorf("创建话题失败: %w", err)
	}

	return topic, nil
}

// GetHotTopics 获取热门话题
func (s *articleV3Service) GetHotTopics(limit int) ([]model.Topic, error) {
	return s.articleRepo.FindHotTopics(limit)
}

// GetTopicDetail 获取话题详情
func (s *articleV3Service) GetTopicDetail(topicID uint64, userID string) (*TopicDetailResponse, error) {
	topic, err := s.articleRepo.FindTopicByID(topicID)
	if err != nil {
		return nil, fmt.Errorf("话题不存在: %w", err)
	}

	isFollowed := false
	if userID != "" {
		isFollowed = s.articleRepo.IsTopicFollowed(userID, topicID)
	}

	return &TopicDetailResponse{
		Topic:      topic,
		IsFollowed: isFollowed,
	}, nil
}

// GetArticlesByTopic 获取话题下的文章
func (s *articleV3Service) GetArticlesByTopic(topicID uint64, page, pageSize int) ([]model.ArticleV3, int64, error) {
	return s.articleRepo.FindByTopicID(topicID, page, pageSize)
}

// FollowTopic 关注话题
func (s *articleV3Service) FollowTopic(topicID uint64, userID string) error {
	// 检查是否已关注
	if s.articleRepo.IsTopicFollowed(userID, topicID) {
		return fmt.Errorf("已经关注过该话题")
	}

	// 创建关注记录
	if err := s.articleRepo.FollowTopic(userID, topicID); err != nil {
		return err
	}

	// 增加关注数
	return s.articleRepo.IncrementTopicFollowCount(topicID)
}

// UnfollowTopic 取关话题
func (s *articleV3Service) UnfollowTopic(topicID uint64, userID string) error {
	// 删除关注记录
	if err := s.articleRepo.UnfollowTopic(userID, topicID); err != nil {
		return err
	}

	// 减少关注数
	return s.articleRepo.DecrementTopicFollowCount(topicID)
}

// ==================== 互动功能实现 ====================

// LikeArticle 点赞文章
func (s *articleV3Service) LikeArticle(articleID uint64, userID string) error {
	// 检查是否已点赞
	if s.likeRepo.IsArticleLiked(userID, articleID) {
		return fmt.Errorf("已经点赞过该文章")
	}

	// 创建点赞记录
	if err := s.likeRepo.LikeArticle(userID, articleID); err != nil {
		return err
	}

	// 增加点赞数
	return s.articleRepo.IncrementLikeCount(articleID)
}

// UnlikeArticle 取消点赞
func (s *articleV3Service) UnlikeArticle(articleID uint64, userID string) error {
	// 删除点赞记录
	if err := s.likeRepo.UnlikeArticle(userID, articleID); err != nil {
		return err
	}

	// 减少点赞数
	return s.articleRepo.DecrementLikeCount(articleID)
}

// FavoriteArticle 收藏文章
func (s *articleV3Service) FavoriteArticle(articleID uint64, userID string) error {
	// 检查是否已收藏
	if s.favoriteRepo.IsFavorited(userID, articleID) {
		return fmt.Errorf("已经收藏过该文章")
	}

	// 创建收藏记录
	favorite := &model.UserFavorite{
		UserID:    userID,
		ArticleID: articleID,
	}
	if err := s.favoriteRepo.Create(favorite); err != nil {
		return err
	}

	// 增加收藏数
	return s.articleRepo.IncrementFavoriteCount(articleID)
}

// UnfavoriteArticle 取消收藏
func (s *articleV3Service) UnfavoriteArticle(articleID uint64, userID string) error {
	// 删除收藏记录
	if err := s.favoriteRepo.Delete(userID, articleID); err != nil {
		return err
	}

	// 减少收藏数
	return s.articleRepo.DecrementFavoriteCount(articleID)
}

// IncrementViewCount 增加浏览量
func (s *articleV3Service) IncrementViewCount(articleID uint64, userID string) error {
	// 1. 增加总浏览量
	if err := s.articleRepo.IncrementViewCount(articleID); err != nil {
		return err
	}

	// 2. 如果是已登录用户，增加真实浏览量（去重）
	if userID != "" {
		// TODO: 使用Redis记录用户浏览历史，24小时内同一用户多次浏览只计数一次
		// 当前简单实现：直接增加计数，Redis去重功能待后续优化
		s.articleRepo.IncrementRealViewCount(articleID)
	}

	return nil
}

// ==================== 推荐与排序实现 ====================

// GetFeaturedArticles 获取精选文章
func (s *articleV3Service) GetFeaturedArticles(limit int) ([]model.ArticleV3, error) {
	return s.articleRepo.FindFeaturedArticles(limit)
}

// GetHotArticles 获取热门文章
func (s *articleV3Service) GetHotArticles(limit int) ([]model.ArticleV3, error) {
	return s.articleRepo.FindHotArticles(limit)
}

// GetRecommendedArticles 获取推���文章
func (s *articleV3Service) GetRecommendedArticles(userID string, limit int) ([]model.ArticleV3, error) {
	// TODO: 基于用户画像推荐（兴趣标签、阅读历史）
	return s.articleRepo.FindRecommendedArticles(limit)
}

// GetRelatedArticles 获取相关文章
func (s *articleV3Service) GetRelatedArticles(articleID uint64, limit int) ([]model.ArticleV3, error) {
	return s.articleRepo.FindRelatedArticles(articleID, limit)
}

// ==================== 热度计算实现 ====================

// CalculateHotScore 计算文章热度分数
func (s *articleV3Service) CalculateHotScore(articleID uint64) (float64, error) {
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return 0, err
	}

	// 热度算法（类似Reddit）：
	// HotScore = (浏览数*0.1 + 点赞数*0.5 + 评论数*0.3 + 收藏数*0.1) * 时间衰减

	viewWeight := float64(article.ViewCount) * 0.1
	likeWeight := float64(article.LikeCount) * 0.5
	commentWeight := float64(article.CommentCount) * 0.3
	favoriteWeight := float64(article.FavoriteCount) * 0.1

	baseScore := viewWeight + likeWeight + commentWeight + favoriteWeight

	// 时间衰减（48小时内权重最高）
	hoursSincePublish := time.Since(*article.PublishTime).Hours()
	timeDecay := math.Exp(-hoursSincePublish / 48.0) // 指数衰减

	hotScore := baseScore * timeDecay

	// 更新热度分数
	s.articleRepo.UpdateHotScore(articleID, hotScore)

	return hotScore, nil
}

// BatchUpdateHotScores 批量更新热度分数（定时任务）
func (s *articleV3Service) BatchUpdateHotScores() error {
	// 1. 获取最近7天发布的文章
	params := &repository.ArticleQueryParams{
		Status:    model.ArticleV3StatusPublished,
		Page:      1,
		PageSize:  1000,
		SortBy:    "publish_time",
		SortOrder: "DESC",
	}

	articles, _, err := s.articleRepo.FindList(params)
	if err != nil {
		return err
	}

	// 2. 批量计算热度
	updates := make(map[uint64]float64)
	for _, article := range articles {
		// 只更新最近7天的文章
		if time.Since(*article.PublishTime).Hours() > 168 {
			continue
		}

		score, _ := s.CalculateHotScore(article.ID)
		updates[article.ID] = score
	}

	// 3. 批量更新
	return s.articleRepo.BatchUpdateHotScores(updates)
}

// MarkAsHot 标记为热门
func (s *articleV3Service) MarkAsHot(articleID uint64) error {
	return s.articleRepo.UpdateFields(articleID, map[string]interface{}{"is_hot": true})
}

// UnmarkAsHot 取消热门标记
func (s *articleV3Service) UnmarkAsHot(articleID uint64) error {
	return s.articleRepo.UpdateFields(articleID, map[string]interface{}{"is_hot": false})
}

// ==================== 版本历史实现 ====================

// GetArticleHistory 获取文章历史版本
func (s *articleV3Service) GetArticleHistory(articleID uint64) ([]model.ArticleHistory, error) {
	return s.articleRepo.FindHistoryByArticleID(articleID)
}

// RollbackToVersion 回滚��指定版本
func (s *articleV3Service) RollbackToVersion(articleID uint64, userID string, version int) error {
	// 1. 检查权限
	if !s.articleRepo.CheckOwnership(articleID, userID) {
		return constant.ErrPermissionDenied
	}

	// 2. 获取历史版本
	history, err := s.articleRepo.FindHistoryByVersion(articleID, version)
	if err != nil {
		return fmt.Errorf("版本不存在: %w", err)
	}

	// 3. 回滚内容
	content, _ := s.articleRepo.FindContentByArticleID(articleID)
	if content != nil {
		content.Content = history.Content
		s.articleRepo.UpdateContent(content)
	}

	// 4. 更新文章标题
	s.articleRepo.UpdateFields(articleID, map[string]interface{}{"title": history.Title})

	// 5. 创建新版本记录
	s.createHistoryVersion(articleID, history.Title, history.Content,
		fmt.Sprintf("回滚到版本%d", version), model.ChangeTypeEdit, userID)

	return nil
}

// ==================== 统计分析实现 ====================

// GetArticleStats 获取文章详细统计
func (s *articleV3Service) GetArticleStats(articleID uint64) (*model.ArticleStatsDetail, error) {
	return s.articleRepo.FindStatsDetailByArticleID(articleID)
}

// GetUserArticleStats 获取用户文章统计
func (s *articleV3Service) GetUserArticleStats(userID string) (*UserArticleStats, error) {
	// 获取用户所有文章
	params := &repository.ArticleQueryParams{
		UserID:   userID,
		Status:   model.ArticleV3StatusPublished,
		Page:     1,
		PageSize: 10000,
	}

	articles, total, err := s.articleRepo.FindList(params)
	if err != nil {
		return nil, err
	}

	// 统计汇总
	stats := &UserArticleStats{
		TotalArticles: total,
	}

	var totalHotScore float64
	for _, article := range articles {
		stats.TotalViews += article.ViewCount
		stats.TotalLikes += article.LikeCount
		stats.TotalComments += article.CommentCount
		stats.TotalFavorites += article.FavoriteCount
		totalHotScore += article.HotScore

		if article.IsFeatured {
			stats.FeaturedCount++
		}
		if article.IsHot {
			stats.HotCount++
		}
	}

	if total > 0 {
		stats.AvgHotScore = totalHotScore / float64(total)
	}

	return stats, nil
}

// ==================== 辅助方法 ====================

// validateArticleContent 验证文章内容
func (s *articleV3Service) validateArticleContent(title, content string) error {
	if err := s.validateTitle(title); err != nil {
		return err
	}
	return s.validateContent(content)
}

// validateTitle 验证标题
func (s *articleV3Service) validateTitle(title string) error {
	if len(title) == 0 {
		return fmt.Errorf("标题不能为空")
	}
	if len([]rune(title)) > 200 {
		return fmt.Errorf("标题不能超过200字")
	}
	return nil
}

// validateContent 验证内容
func (s *articleV3Service) validateContent(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("内容不能为空")
	}
	if len([]rune(content)) < 10 {
		return fmt.Errorf("内容至少10字")
	}
	return nil
}

// checkSensitiveWords 检查敏感词
func (s *articleV3Service) checkSensitiveWords(title, content string) error {
	// TODO: 实现敏感词检查（基于敏感词库）
	// 优化方向：1) 集成DFA算法进行敏感词过滤
	//         2) 支持敏感词库动态更新
	//         3) 支持多级敏感词（警告、拒绝、替换）
	// 当前简单实现：暂不检查，待后续集成敏感词过滤服务
	return nil
}

// calculateReadTime 计算阅读时间
func (s *articleV3Service) calculateReadTime(content string) int {
	wordCount := len([]rune(content))
	// 假设每分钟阅读300字
	readTime := int(math.Ceil(float64(wordCount) / 300.0))
	if readTime < 1 {
		readTime = 1
	}
	return readTime
}

// handleTopics 处理话题
func (s *articleV3Service) handleTopics(articleID uint64, topicNames []string) {
	for _, topicName := range topicNames {
		// 查找或创建话题
		topic, err := s.articleRepo.FindTopicByName(topicName)
		if err != nil {
			// 话题不存在，创建
			topic = &model.Topic{
				Name:   topicName,
				Status: 1,
			}
			s.articleRepo.CreateTopic(topic)
		}

		// 关联文章和话题
		s.articleRepo.AddArticleToTopic(articleID, topic.ID)
		s.articleRepo.IncrementTopicArticleCount(topic.ID)
	}
}

// handleTags 处理标签
func (s *articleV3Service) handleTags(tagNames []string) {
	for _, tagName := range tagNames {
		// 查找或创建标签
		tag, err := s.articleRepo.FindOrCreateTag(tagName)
		if err == nil {
			s.articleRepo.IncrementTagArticleCount(tag.Name)
		}
	}
}

// createHistoryVersion 创建历史版本
func (s *articleV3Service) createHistoryVersion(articleID uint64, title, content, reason string, changeType int8, operatorID string) {
	// 获取当前最大版本号
	histories, _ := s.articleRepo.FindHistoryByArticleID(articleID)
	version := len(histories) + 1

	history := &model.ArticleHistory{
		ArticleID:    articleID,
		Version:      version,
		Title:        title,
		Content:      content,
		ChangeType:   changeType,
		ChangeReason: reason,
		OperatorID:   operatorID,
	}

	s.articleRepo.CreateHistory(history)
}

// checkArticleVisibility 检查文章可见性权限
func (s *articleV3Service) checkArticleVisibility(article *model.ArticleV3, viewerID string) bool {
	// 作者始终可见
	if article.UserID == viewerID {
		return true
	}

	switch article.Visibility {
	case model.ArticleVisibilityPublic:
		return true
	case model.ArticleVisibilityFollower:
		// 检查是否为粉丝（关注者）
		return s.followRepo.IsFollowing(viewerID, article.UserID)
	case model.ArticleVisibilityFriend:
		// 检查是否为好友（互相关注）
		return s.followRepo.IsFriend(viewerID, article.UserID)
	case model.ArticleVisibilityPrivate:
		return false
	case model.ArticleVisibilityPaid:
		// TODO: 检查是否已付费（需要付费系统支持）
		return false
	default:
		return false
	}
}

// initializeStatsDetail 初始化统计详情
func (s *articleV3Service) initializeStatsDetail(articleID uint64) {
	stats := &model.ArticleStatsDetail{
		ArticleID: articleID,
	}
	s.articleRepo.CreateStatsDetail(stats)
}
