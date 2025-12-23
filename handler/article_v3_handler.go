package handler

import (
	"astronomer-gin/middleware"
	"astronomer-gin/pkg/response"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ArticleV3Handler 企业级文章处理器
type ArticleV3Handler struct {
	articleService service.ArticleV3Service
}

// NewArticleV3Handler 创建文章处理器实例
func NewArticleV3Handler(articleService service.ArticleV3Service) *ArticleV3Handler {
	return &ArticleV3Handler{
		articleService: articleService,
	}
}

// RegisterRoutes 注册路由
func (h *ArticleV3Handler) RegisterRoutes(r *gin.Engine) {
	v3 := r.Group("/api/v3")
	{
		// 公开接口（无需认证）
		v3.GET("/articles", h.GetArticleList)                // 文章列表
		v3.GET("/articles/:id", h.GetArticleDetail)          // 文章详情
		v3.GET("/articles/:id/history", h.GetArticleHistory) // 文章历史版本

		// 分类相关
		v3.GET("/categories", h.GetCategoryTree)                    // 分类树
		v3.GET("/categories/:id/articles", h.GetArticlesByCategory) // 分类下的文章

		// 话题相关
		v3.GET("/topics/hot", h.GetHotTopics)                // 热门话题
		v3.GET("/topics/:id", h.GetTopicDetail)              // 话题详情
		v3.GET("/topics/:id/articles", h.GetArticlesByTopic) // 话题下的文章

		// 需要认证的接口
		auth := v3.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// 文章管理
			auth.POST("/articles", h.CreateArticle)       // 发布文章
			auth.PUT("/articles/:id", h.UpdateArticle)    // 更新文章
			auth.DELETE("/articles/:id", h.DeleteArticle) // 删除文章

			// 草稿管理
			auth.POST("/drafts", h.SaveDraft)                // 保存草稿
			auth.GET("/drafts", h.GetUserDrafts)             // 我的草稿列表
			auth.GET("/drafts/:id", h.GetDraftDetail)        // 获取草稿详情
			auth.PUT("/drafts/:id", h.UpdateDraft)           // 更新草稿
			auth.POST("/drafts/:id/publish", h.PublishDraft) // 发布草稿
			auth.DELETE("/drafts/:id", h.DeleteDraft)        // 删除草稿

			// 话题管理
			auth.POST("/topics", h.CreateTopic)                // 创建话题
			auth.POST("/topics/:id/follow", h.FollowTopic)     // 关注话题
			auth.DELETE("/topics/:id/follow", h.UnfollowTopic) // 取消关注

			// 分类管理（管理员）
			auth.POST("/categories", h.CreateCategory) // 创建分类
		}
	}
}

// ==================== 文章相关接口 ====================

// CreateArticle 发布文章
func (h *ArticleV3Handler) CreateArticle(c *gin.Context) {
	var req service.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 从JWT中获取用户ID
	//userID, exists := c.Get("user_id")
	//if !exists {
	//	response.Unauthorized(c, "未登录")
	//	return
	//}
	//req.UserID = userID.(uint64)

	article, err := h.articleService.CreateArticle(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, article)
}

// UpdateArticle 更新文章
func (h *ArticleV3Handler) UpdateArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req service.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.UpdateArticle(articleID, userID.(string), &req); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteArticle 删除文章
func (h *ArticleV3Handler) DeleteArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.DeleteArticle(articleID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetArticleDetail 获取文章详情
func (h *ArticleV3Handler) GetArticleDetail(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var viewerID string
	if userID, exists := c.Get("user_id"); exists {
		viewerID = userID.(string)
	}

	detail, err := h.articleService.GetArticleDetail(articleID, viewerID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, detail)
}

// GetArticleList 获取文章列表
func (h *ArticleV3Handler) GetArticleList(c *gin.Context) {
	req := &service.ArticleListRequest{
		Page:       1,
		PageSize:   20,
		SortBy:     c.DefaultQuery("sort_by", "hot"),
		CategoryID: parseUint64(c.Query("category_id")),
		ColumnID:   parseUint64(c.Query("column_id")),
		Status:     parseInt8(c.Query("status")),
	}

	if page := c.Query("page"); page != "" {
		req.Page, _ = strconv.Atoi(page)
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		req.PageSize, _ = strconv.Atoi(pageSize)
	}

	result, err := h.articleService.GetArticleList(req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetArticleHistory 获取文章历史版本
func (h *ArticleV3Handler) GetArticleHistory(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效��文章ID")
		return
	}

	history, err := h.articleService.GetArticleHistory(articleID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, history)
}

// ==================== 草稿相关接口 ====================

// SaveDraft 保存草稿
func (h *ArticleV3Handler) SaveDraft(c *gin.Context) {
	var req service.SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	draft, err := h.articleService.SaveDraft(userID.(string), &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, draft)
}

// UpdateDraft 更新草稿
func (h *ArticleV3Handler) UpdateDraft(c *gin.Context) {
	draftID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的草稿ID")
		return
	}

	var req service.SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.UpdateDraft(draftID, userID.(string), &req); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetDraftDetail 获取草稿详情
func (h *ArticleV3Handler) GetDraftDetail(c *gin.Context) {
	draftID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的草稿ID")
		return
	}

	userID, _ := c.Get("user_id")
	draft, err := h.articleService.GetDraftDetail(draftID, userID.(string))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, draft)
}

// PublishDraft 发布草稿
func (h *ArticleV3Handler) PublishDraft(c *gin.Context) {
	draftID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的草稿ID")
		return
	}

	userID, _ := c.Get("user_id")
	article, err := h.articleService.PublishDraft(draftID, userID.(string))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, article)
}

// GetUserDrafts 获取用户草稿列表
func (h *ArticleV3Handler) GetUserDrafts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID, _ := c.Get("user_id")
	drafts, total, err := h.articleService.GetUserDrafts(userID.(string), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"drafts":    drafts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteDraft 删除草稿
func (h *ArticleV3Handler) DeleteDraft(c *gin.Context) {
	draftID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的草稿ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.DeleteDraft(draftID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 分类相关接口 ====================

// CreateCategory 创建分类
func (h *ArticleV3Handler) CreateCategory(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		ParentID  uint64 `json:"parent_id"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	category, err := h.articleService.CreateCategory(req.Name, req.ParentID, req.Icon, req.SortOrder)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, category)
}

// GetCategoryTree 获取分类树
func (h *ArticleV3Handler) GetCategoryTree(c *gin.Context) {
	tree, err := h.articleService.GetCategoryTree()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, tree)
}

// GetArticlesByCategory 获取分类下的文章
func (h *ArticleV3Handler) GetArticlesByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的分类ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	articles, total, err := h.articleService.GetArticlesByCategory(categoryID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"articles":  articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ==================== 专栏相关接口 ====================

// CreateColumn 创建专栏
func (h *ArticleV3Handler) CreateColumn(c *gin.Context) {
	var req service.CreateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	column, err := h.articleService.CreateColumn(userID.(string), &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, column)
}

// UpdateColumn 更新专栏
func (h *ArticleV3Handler) UpdateColumn(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}

	var req service.UpdateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.UpdateColumn(columnID, userID.(string), &req); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetColumnDetail 获取专栏详情
func (h *ArticleV3Handler) GetColumnDetail(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}
	userID, _ := c.Get("user_id")
	detail, err := h.articleService.GetColumnDetail(columnID, userID.(string))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, detail)
}

// GetColumnArticles 获取专栏文章列表
func (h *ArticleV3Handler) GetColumnArticles(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	articles, total, err := h.articleService.GetColumnArticles(columnID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"articles":  articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddArticleToColumn 添加文章到专栏
func (h *ArticleV3Handler) AddArticleToColumn(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}

	var req struct {
		ArticleID uint64 `json:"article_id" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.AddArticleToColumn(columnID, req.ArticleID, userID.(string), req.SortOrder); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveArticleFromColumn 从专栏移除文章
func (h *ArticleV3Handler) RemoveArticleFromColumn(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}

	articleID, err := strconv.ParseUint(c.Param("articleId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.RemoveArticleFromColumn(columnID, articleID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// SubscribeColumn 订阅专栏
func (h *ArticleV3Handler) SubscribeColumn(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.SubscribeColumn(columnID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnsubscribeColumn 取消订阅专栏
func (h *ArticleV3Handler) UnsubscribeColumn(c *gin.Context) {
	columnID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的专栏ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.UnsubscribeColumn(columnID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 话题相关接口 ====================

// CreateTopic 创建话题
func (h *ArticleV3Handler) CreateTopic(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	topic, err := h.articleService.CreateTopic(req.Name, req.Description, userID.(string))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, topic)
}

// GetHotTopics 获取热门话题
func (h *ArticleV3Handler) GetHotTopics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	topics, err := h.articleService.GetHotTopics(limit)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, topics)
}

// GetTopicDetail 获取话题详情
func (h *ArticleV3Handler) GetTopicDetail(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的话题ID")
		return
	}
	userID, _ := c.Get("user_id")
	detail, err := h.articleService.GetTopicDetail(topicID, userID.(string))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, detail)
}

// GetArticlesByTopic 获取话题下的文章
func (h *ArticleV3Handler) GetArticlesByTopic(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的话题ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	articles, total, err := h.articleService.GetArticlesByTopic(topicID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"articles":  articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// FollowTopic 关注话题
func (h *ArticleV3Handler) FollowTopic(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的话题ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.FollowTopic(topicID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnfollowTopic 取消关注话题
func (h *ArticleV3Handler) UnfollowTopic(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的话题ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.articleService.UnfollowTopic(topicID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
