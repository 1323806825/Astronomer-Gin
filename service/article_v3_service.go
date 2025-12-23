package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/repository"
	"fmt"
	"github.com/russross/blackfriday/v2"
	"gorm.io/gorm"
	"time"
)

// ArticleV3Service 企业级文章服务接口
type ArticleV3Service interface {
	// ==================== 文章发布 ====================
	// 创建并发布文章
	CreateArticle(req *CreateArticleRequest) (*model.ArticleV3, error)
	// 更新文章
	UpdateArticle(articleID uint64, userID string, req *UpdateArticleRequest) error
	// 删除文章（软删除）
	DeleteArticle(articleID uint64, userID string) error
	// 获取文章详情（含内容）
	GetArticleDetail(articleID uint64, viewerID string) (*ArticleDetailResponse, error)
	// 获取文章列表
	GetArticleList(req *ArticleListRequest) (*ArticleListResponse, error)

	// ==================== 草稿管理 ====================
	// 保存草稿（自动保存）
	SaveDraft(userID string, req *SaveDraftRequest) (*model.ArticleDraft, error)
	// 更新草稿
	UpdateDraft(draftID uint64, userID string, req *SaveDraftRequest) error
	// 发布草稿
	PublishDraft(draftID uint64, userID string) (*model.ArticleV3, error)
	// 获取用户草稿列表
	GetUserDrafts(userID string, page, pageSize int) ([]model.ArticleDraft, int64, error)
	// 获取草稿详情
	GetDraftDetail(draftID uint64, userID string) (*model.ArticleDraft, error)
	// 删除草稿
	DeleteDraft(draftID uint64, userID string) error

	// ==================== 分类管理 ====================
	// 创建分类
	CreateCategory(name string, parentID uint64, icon string, sortOrder int) (*model.ArticleCategory, error)
	// 获取所有分类（树形结构）
	GetCategoryTree() ([]CategoryTreeNode, error)
	// 获取分类下的文章
	GetArticlesByCategory(categoryID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)

	// ==================== 专栏管理 ====================
	// 创建专栏
	CreateColumn(userID string, req *CreateColumnRequest) (*model.ArticleColumn, error)
	// 更新专栏
	UpdateColumn(columnID uint64, userID string, req *UpdateColumnRequest) error
	// 添加文章到专栏
	AddArticleToColumn(columnID, articleID uint64, userID string, sortOrder int) error
	// 从专栏移除文章
	RemoveArticleFromColumn(columnID, articleID uint64, userID string) error
	// 获取专栏详情
	GetColumnDetail(columnID uint64, userID string) (*ColumnDetailResponse, error)
	// 获取专栏文章列表
	GetColumnArticles(columnID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)
	// 订阅专栏
	SubscribeColumn(columnID uint64, userID string) error
	// 取消订阅专栏
	UnsubscribeColumn(columnID uint64, userID string) error

	// ==================== 话题管理 ====================
	// 创建话题
	CreateTopic(name, description, userID string) (*model.Topic, error)
	// 获取热门话题
	GetHotTopics(limit int) ([]model.Topic, error)
	// 获取话题详情
	GetTopicDetail(topicID uint64, userID string) (*TopicDetailResponse, error)
	// 获取话题下的文章
	GetArticlesByTopic(topicID uint64, page, pageSize int) ([]model.ArticleV3, int64, error)
	// 关注话题
	FollowTopic(topicID uint64, userID string) error
	// 取关话题
	UnfollowTopic(topicID uint64, userID string) error

	// ==================== 互动功能 ====================
	// 点赞文章
	LikeArticle(articleID uint64, userID string) error
	// 取消点赞
	UnlikeArticle(articleID uint64, userID string) error
	// 收藏文章
	FavoriteArticle(articleID uint64, userID string) error
	// 取消收藏
	UnfavoriteArticle(articleID uint64, userID string) error
	// 增加浏览量
	IncrementViewCount(articleID uint64, userID string) error

	// ==================== 推荐与排序 ====================
	// 获取精选文章
	GetFeaturedArticles(limit int) ([]model.ArticleV3, error)
	// 获取热门文章
	GetHotArticles(limit int) ([]model.ArticleV3, error)
	// 获取推荐文章
	GetRecommendedArticles(userID string, limit int) ([]model.ArticleV3, error)
	// 获取相关文章
	GetRelatedArticles(articleID uint64, limit int) ([]model.ArticleV3, error)

	// ==================== 热度计算 ====================
	// 计算文章热度分数
	CalculateHotScore(articleID uint64) (float64, error)
	// 批量更新热度分数（定时任务）
	BatchUpdateHotScores() error
	// 标记为热门文章
	MarkAsHot(articleID uint64) error
	// 取消热门标记
	UnmarkAsHot(articleID uint64) error

	// ==================== 版本历史 ====================
	// 获取文章历史版本
	GetArticleHistory(articleID uint64) ([]model.ArticleHistory, error)
	// 回滚到指定版本
	RollbackToVersion(articleID uint64, userID string, version int) error

	// ==================== 统计分析 ====================
	// 获取文章详细统计
	GetArticleStats(articleID uint64) (*model.ArticleStatsDetail, error)
	// 获取用户文章统计
	GetUserArticleStats(userID string) (*UserArticleStats, error)
}

// ==================== 请求/响应结构体 ====================

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title       string   `json:"title" binding:"required,min=1,max=200"`
	Summary     string   `json:"summary" binding:"max=500"`
	Content     string   `json:"content" binding:"required,min=10"`
	CoverImage  string   `json:"cover_image"`
	ContentType int8     `json:"content_type"`
	CategoryID  uint64   `json:"category_id"`
	ColumnID    uint64   `json:"column_id"`
	Tags        []string `json:"tags"`
	Topics      []string `json:"topics"`
	Visibility  int8     `json:"visibility"`
	IsPaid      bool     `json:"is_paid"`
	Price       float64  `json:"price"`
	Keywords    string   `json:"keywords"`
	Description string   `json:"description"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title       *string   `json:"title"`
	Summary     *string   `json:"summary"`
	Content     *string   `json:"content"`
	CoverImage  *string   `json:"cover_image"`
	CategoryID  *uint64   `json:"category_id"`
	ColumnID    *uint64   `json:"column_id"`
	Tags        *[]string `json:"tags"`
	Topics      *[]string `json:"topics"`
	Visibility  *int8     `json:"visibility"`
	Keywords    *string   `json:"keywords"`
	Description *string   `json:"description"`
}

// SaveDraftRequest 保存草稿请求
type SaveDraftRequest struct {
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"`
	CoverImage string   `json:"cover_image"`
	CategoryID uint64   `json:"category_id"`
	ColumnID   uint64   `json:"column_id"`
	Tags       []string `json:"tags"`
	Topics     []string `json:"topics"`
}

// ArticleDetailResponse 文章详情响应（扁平化结构，匹配前端期望）
type ArticleDetailResponse struct {
	// 文章基本信息
	ID            uint64   `json:"id"`
	UserID        string   `json:"user_id"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	CoverImage    string   `json:"cover_image"`
	ContentType   int8     `json:"content_type"`
	CategoryID    uint64   `json:"category_id"`
	ColumnID      uint64   `json:"column_id"`
	Tags          []string `json:"tags"`
	Topics        []string `json:"topics"`
	Visibility    int8     `json:"visibility"`
	Status        int8     `json:"status"`
	ViewCount     uint64   `json:"view_count"`
	LikeCount     uint64   `json:"like_count"`
	CommentCount  uint64   `json:"comment_count"`
	FavoriteCount uint64   `json:"favorite_count"`
	ShareCount    uint64   `json:"share_count"`
	CreateTime    string   `json:"create_time"`
	UpdateTime    string   `json:"update_time"`
	PublishTime   string   `json:"publish_time"`

	// 文章内容
	Content   string `json:"content"`
	WordCount int    `json:"word_count"`
	ReadTime  int    `json:"read_time"`

	// 作者信息
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
	AuthorBio    string `json:"author_bio"`

	// 用户互动状态
	IsLiked     bool `json:"is_liked"`
	IsFavorited bool `json:"is_favorited"`
	IsFollowing bool `json:"is_following"`

	// 相关数据
	Categories      []map[string]interface{} `json:"categories"`
	RelatedArticles []ArticleListItem        `json:"related_articles"`
}

// ArticleAuthorInfo 作者信息
type ArticleAuthorInfo struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	Intro         string `json:"intro"`
	FollowerCount int    `json:"follower_count"`
}

// ArticleListRequest 文章列表请求
type ArticleListRequest struct {
	CategoryID uint64 `json:"category_id"`
	ColumnID   uint64 `json:"column_id"`
	TopicID    uint64 `json:"topic_id"`
	UserID     string `json:"user_id"`
	Keyword    string `json:"keyword"`
	SortBy     string `json:"sort_by"` // hot, time, like
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Status     int8   `json:"status"`
}

// ArticleListResponse 文章列表响应
type ArticleListResponse struct {
	Articles []ArticleListItem `json:"articles"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// ArticleListItem 文章列表项（包含作者信息）
type ArticleListItem struct {
	*model.ArticleV3
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
}

// CreateColumnRequest 创建专栏请求
type CreateColumnRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	CoverImage  string `json:"cover_image"`
	SortType    int8   `json:"sort_type"`
}

// UpdateColumnRequest 更新专栏请求
type UpdateColumnRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	CoverImage  *string `json:"cover_image"`
	SortType    *int8   `json:"sort_type"`
	IsFinished  *bool   `json:"is_finished"`
}

// ColumnDetailResponse 专栏详情响应
type ColumnDetailResponse struct {
	Column       *model.ArticleColumn `json:"column"`
	Author       *ArticleAuthorInfo   `json:"author"`
	ArticleCount int                  `json:"article_count"`
	IsSubscribed bool                 `json:"is_subscribed"`
}

// TopicDetailResponse 话题详情响应
type TopicDetailResponse struct {
	Topic      *model.Topic `json:"topic"`
	IsFollowed bool         `json:"is_followed"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	*model.ArticleCategory
	Children []CategoryTreeNode `json:"children"`
}

// UserArticleStats 用户文章统计
type UserArticleStats struct {
	TotalArticles  int64   `json:"total_articles"`
	TotalViews     uint64  `json:"total_views"`
	TotalLikes     uint64  `json:"total_likes"`
	TotalComments  uint64  `json:"total_comments"`
	TotalFavorites uint64  `json:"total_favorites"`
	AvgHotScore    float64 `json:"avg_hot_score"`
	FeaturedCount  int     `json:"featured_count"`
	HotCount       int     `json:"hot_count"`
}

// ==================== Service实现 ====================

type articleV3Service struct {
	articleRepo  repository.ArticleV3Repository
	userRepo     repository.UserRepository
	followRepo   repository.FollowRepository
	likeRepo     repository.LikeRepository
	favoriteRepo repository.FavoriteRepository
	db           *gorm.DB
}

// NewArticleV3Service 创建ArticleV3Service实例
func NewArticleV3Service(
	articleRepo repository.ArticleV3Repository,
	userRepo repository.UserRepository,
	followRepo repository.FollowRepository,
	likeRepo repository.LikeRepository,
	favoriteRepo repository.FavoriteRepository,
	db *gorm.DB,
) ArticleV3Service {
	return &articleV3Service{
		articleRepo:  articleRepo,
		userRepo:     userRepo,
		followRepo:   followRepo,
		likeRepo:     likeRepo,
		favoriteRepo: favoriteRepo,
		db:           db,
	}
}

// ==================== 文章发布实现 ====================

// CreateArticle 创建并发布文章
func (s *articleV3Service) CreateArticle(req *CreateArticleRequest) (*model.ArticleV3, error) {
	// 1. 参数验证
	if err := s.validateArticleContent(req.Title, req.Content); err != nil {
		return nil, err
	}

	// 2. 敏感词检查
	if err := s.checkSensitiveWords(req.Title, req.Content); err != nil {
		return nil, err
	}

	// 3. 构建文章对象
	now := time.Now()
	article := &model.ArticleV3{
		Title:       req.Title,
		Summary:     req.Summary,
		CoverImage:  req.CoverImage,
		ContentType: req.ContentType,
		CategoryID:  req.CategoryID,
		ColumnID:    req.ColumnID,
		Tags:        model.JSONStringList(req.Tags),
		Topics:      model.JSONStringList(req.Topics),
		Status:      model.ArticleV3StatusPublished,
		Visibility:  req.Visibility,
		IsPaid:      req.IsPaid,
		Price:       req.Price,
		Keywords:    req.Keywords,
		Description: req.Description,
		PublishTime: &now,
	}

	// 4. 生成内容对象
	content := &model.ArticleContent{
		Content:     req.Content,
		ContentHTML: string(blackfriday.Run([]byte(req.Content))),
		WordCount:   len([]rune(req.Content)),
		ReadTime:    s.calculateReadTime(req.Content),
	}

	// 5. 事务创建文章和内容
	if err := s.articleRepo.CreateWithContent(article, content); err != nil {
		return nil, fmt.Errorf("创建文章失败: %w", err)
	}

	// 6. 创建历史版本
	s.createHistoryVersion(article.ID, article.Title, req.Content, "", model.ChangeTypeCreate, article.UserID)

	// 7. 更新分类计数
	if req.CategoryID > 0 {
		s.articleRepo.IncrementCategoryArticleCount(req.CategoryID)
	}

	// 8. 处理话题
	if len(req.Topics) > 0 {
		s.handleTopics(article.ID, req.Topics)
	}

	// 9. 处理标签
	if len(req.Tags) > 0 {
		s.handleTags(req.Tags)
	}

	// 10. 初始化统计详情
	s.initializeStatsDetail(article.ID)

	return article, nil
}

// UpdateArticle 更新文章
func (s *articleV3Service) UpdateArticle(articleID uint64, userID string, req *UpdateArticleRequest) error {
	// 1. 检查权限
	if !s.articleRepo.CheckOwnership(articleID, userID) {
		return constant.ErrPermissionDenied
	}

	// 2. 获取原文章
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return fmt.Errorf("文章不存在: %w", err)
	}

	// 3. 构建更新字段
	updates := make(map[string]interface{})
	oldContent := ""

	if req.Title != nil {
		if err := s.validateTitle(*req.Title); err != nil {
			return err
		}
		updates["title"] = *req.Title
	}

	if req.Content != nil {
		if err := s.validateContent(*req.Content); err != nil {
			return err
		}
		// 更新内容表
		content, _ := s.articleRepo.FindContentByArticleID(articleID)
		if content != nil {
			content.Content = *req.Content
			content.ContentHTML = string(blackfriday.Run([]byte(*req.Content)))
			content.WordCount = len([]rune(*req.Content))
			content.ReadTime = s.calculateReadTime(*req.Content)
			s.articleRepo.UpdateContent(content)
		}
	}

	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}

	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}

	if req.CategoryID != nil {
		// 更新分类计数
		if article.CategoryID > 0 {
			s.articleRepo.DecrementCategoryArticleCount(article.CategoryID)
		}
		if *req.CategoryID > 0 {
			s.articleRepo.IncrementCategoryArticleCount(*req.CategoryID)
		}
		updates["category_id"] = *req.CategoryID
	}

	if req.Tags != nil {
		updates["tags"] = model.JSONStringList(*req.Tags)
		s.handleTags(*req.Tags)
	}

	if req.Topics != nil {
		updates["topics"] = model.JSONStringList(*req.Topics)
		s.handleTopics(articleID, *req.Topics)
	}

	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}

	if req.Keywords != nil {
		updates["keywords"] = *req.Keywords
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	// 4. 更新文章
	if err := s.articleRepo.UpdateFields(articleID, updates); err != nil {
		return fmt.Errorf("更新文章失败: %w", err)
	}

	// 5. 创建历史版本
	newTitle := article.Title
	if req.Title != nil {
		newTitle = *req.Title
	}
	newContent := oldContent
	if req.Content != nil {
		newContent = *req.Content
	}
	s.createHistoryVersion(articleID, newTitle, newContent, "用户编辑", model.ChangeTypeEdit, userID)

	return nil
}

// DeleteArticle 删除文章（软删除）
func (s *articleV3Service) DeleteArticle(articleID uint64, userID string) error {
	// 1. 检查权限
	if !s.articleRepo.CheckOwnership(articleID, userID) {
		return constant.ErrPermissionDenied
	}

	// 2. 获取文章
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return fmt.Errorf("文章不存在: %w", err)
	}

	// 3. 软删除
	if err := s.articleRepo.SoftDelete(articleID); err != nil {
		return fmt.Errorf("删除文章失败: %w", err)
	}

	// 4. 更新分类计数
	if article.CategoryID > 0 {
		s.articleRepo.DecrementCategoryArticleCount(article.CategoryID)
	}

	// 5. 更新专栏计数
	if article.ColumnID > 0 {
		s.articleRepo.DecrementColumnArticleCount(article.ColumnID)
	}

	return nil
}

// GetArticleDetail 获取文章详情
func (s *articleV3Service) GetArticleDetail(articleID uint64, viewerID string) (*ArticleDetailResponse, error) {
	// 1. 获取文章和内容
	article, content, err := s.articleRepo.FindByIDWithContent(articleID)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %w", err)
	}

	// 2. 检查可见性权限
	if !s.checkArticleVisibility(article, viewerID) {
		return nil, constant.ErrPermissionDenied
	}

	// 3. 获取作者信息
	author, _ := s.userRepo.FindByID(article.UserID)
	authorName := "未知"
	authorAvatar := ""
	authorBio := ""
	if author != nil {
		authorName = author.Username
		authorAvatar = author.Icon
		authorBio = author.Intro
	}

	// 4. 获取分类
	var categories []map[string]interface{}
	if article.CategoryID > 0 {
		if cat, err := s.articleRepo.FindCategoryByID(article.CategoryID); err == nil {
			categories = append(categories, map[string]interface{}{
				"id":   cat.ID,
				"name": cat.Name,
				"icon": cat.Icon,
			})
		}
	}

	// 5. 获取相关文章（带作者信息）
	relatedArticlesRaw, _ := s.articleRepo.FindRelatedArticles(articleID, 5)
	relatedArticles := make([]ArticleListItem, 0, len(relatedArticlesRaw))
	for _, relArticle := range relatedArticlesRaw {
		relAuthor, _ := s.userRepo.FindByID(relArticle.UserID)
		relAuthorName := "未知"
		relAuthorAvatar := ""
		if relAuthor != nil {
			relAuthorName = relAuthor.Username
			relAuthorAvatar = relAuthor.Icon
		}
		relatedArticles = append(relatedArticles, ArticleListItem{
			ArticleV3:    &relArticle,
			AuthorName:   relAuthorName,
			AuthorAvatar: relAuthorAvatar,
		})
	}

	// 6. 检查用户互动状态
	isLiked := false
	isFavorited := false
	isFollowing := false
	if viewerID != "" {
		isLiked = s.likeRepo.IsArticleLiked(viewerID, articleID)
		isFavorited = s.favoriteRepo.IsFavorited(viewerID, articleID)
		isFollowing = s.followRepo.IsFollowing(viewerID, article.UserID)
	}

	// 7. 格式化时间
	createTime := ""
	updateTime := ""
	publishTime := ""
	if !article.CreateTime.IsZero() {
		createTime = article.CreateTime.Format("2006-01-02 15:04:05")
	}
	if !article.UpdateTime.IsZero() {
		updateTime = article.UpdateTime.Format("2006-01-02 15:04:05")
	}
	if article.PublishTime != nil {
		publishTime = article.PublishTime.Format("2006-01-02 15:04:05")
	}

	// 8. 返回扁平化结构
	return &ArticleDetailResponse{
		// 文章基本信息
		ID:            article.ID,
		UserID:        article.UserID,
		Title:         article.Title,
		Summary:       article.Summary,
		CoverImage:    article.CoverImage,
		ContentType:   article.ContentType,
		CategoryID:    article.CategoryID,
		ColumnID:      article.ColumnID,
		Tags:          []string(article.Tags),
		Topics:        []string(article.Topics),
		Visibility:    article.Visibility,
		Status:        article.Status,
		ViewCount:     article.ViewCount,
		LikeCount:     article.LikeCount,
		CommentCount:  article.CommentCount,
		FavoriteCount: article.FavoriteCount,
		ShareCount:    article.ShareCount,
		CreateTime:    createTime,
		UpdateTime:    updateTime,
		PublishTime:   publishTime,

		// 文章内容
		Content:   content.Content,
		WordCount: content.WordCount,
		ReadTime:  content.ReadTime,

		// 作者信息
		AuthorID:     article.UserID,
		AuthorName:   authorName,
		AuthorAvatar: authorAvatar,
		AuthorBio:    authorBio,

		// 用户互动状态
		IsLiked:     isLiked,
		IsFavorited: isFavorited,
		IsFollowing: isFollowing,

		// 相关数据
		Categories:      categories,
		RelatedArticles: relatedArticles,
	}, nil
}

// GetArticleList 获取文章列表
func (s *articleV3Service) GetArticleList(req *ArticleListRequest) (*ArticleListResponse, error) {
	// 1. 构建查询参数
	params := &repository.ArticleQueryParams{
		CategoryID: req.CategoryID,
		ColumnID:   req.ColumnID,
		TopicID:    req.TopicID,
		UserID:     req.UserID,
		Keyword:    req.Keyword,
		Status:     model.ArticleV3StatusPublished,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 2. 设置排序
	switch req.SortBy {
	case "hot":
		params.SortBy = "hot_score"
		params.SortOrder = "DESC"
	case "like":
		params.SortBy = "like_count"
		params.SortOrder = "DESC"
	default:
		params.SortBy = "publish_time"
		params.SortOrder = "DESC"
	}

	// 3. 查询文章列表
	articles, total, err := s.articleRepo.FindList(params)
	if err != nil {
		return nil, fmt.Errorf("查询文章列表失败: %w", err)
	}

	// 4. 为每篇文章添加作者信息
	articleItems := make([]ArticleListItem, 0, len(articles))
	for _, article := range articles {
		author, _ := s.userRepo.FindByID(article.UserID)
		authorName := "未知"
		authorAvatar := ""
		if author != nil {
			authorName = author.Username
			authorAvatar = author.Icon
		}

		articleItems = append(articleItems, ArticleListItem{
			ArticleV3:    &article,
			AuthorName:   authorName,
			AuthorAvatar: authorAvatar,
		})
	}

	return &ArticleListResponse{
		Articles: articleItems,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 继续在下一个文件...
