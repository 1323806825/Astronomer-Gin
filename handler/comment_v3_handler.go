package handler

import (
	"astronomer-gin/middleware"
	"astronomer-gin/pkg/response"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CommentV3Handler 企业级评论处理器
type CommentV3Handler struct {
	commentService service.CommentV3Service
}

// NewCommentV3Handler 创建评论处理器实例
func NewCommentV3Handler(commentService service.CommentV3Service) *CommentV3Handler {
	return &CommentV3Handler{
		commentService: commentService,
	}
}

// RegisterRoutes 注册路由
func (h *CommentV3Handler) RegisterRoutes(r *gin.Engine) {
	v3 := r.Group("/api/v3")
	{
		// 公开接口（无需认证）
		v3.GET("/comments/:id", h.GetCommentDetail)       // 评论详情
		v3.GET("/comments/root", h.GetRootComments)       // 根评论列表
		v3.GET("/comments/:id/replies", h.GetSubComments) // 子评论列表
		v3.GET("/comments/:id/tree", h.GetCommentTree)    // 评论树
		v3.GET("/comments/hot", h.GetHotComments)         // 热评列表
		v3.GET("/comments/stats", h.GetCommentStats)      // 评论统计

		// 需要认证的接口
		auth := v3.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// 评论发表
			auth.POST("/comments/root", h.CreateRootComment)   // 发表根评论
			auth.POST("/comments/reply", h.CreateReplyComment) // 发表回复评论
			auth.DELETE("/comments/:id", h.DeleteComment)      // 删除评论

			// 互动功能
			auth.POST("/comments/:id/like", h.LikeComment)           // 点赞评论
			auth.DELETE("/comments/:id/like", h.UnlikeComment)       // 取消点赞
			auth.POST("/comments/:id/dislike", h.DislikeComment)     // 点踩评论
			auth.DELETE("/comments/:id/dislike", h.UndislikeComment) // 取消点踩

			// 举报功能
			auth.POST("/comments/:id/report", h.ReportComment) // 举报评论
			auth.GET("/reports/pending", h.GetPendingReports)  // 获取待审核举报（管理员）
			auth.POST("/reports/:id/handle", h.HandleReport)   // 审核举报（管理员）

			// UP主功能
			auth.POST("/comments/:id/author-reply", h.AddAuthorReply) // UP主追评
			auth.POST("/comments/:id/pin", h.PinComment)              // 置顶评论
			auth.DELETE("/comments/:id/pin", h.UnpinComment)          // 取消置顶
			auth.POST("/comments/:id/feature", h.FeatureComment)      // 精选评论
			auth.DELETE("/comments/:id/feature", h.UnfeatureComment)  // 取消精选

			// 用户统计
			auth.GET("/comments/my-stats", h.GetUserCommentStats) // 我的评论统计

			// 管理功能（管理员）
			auth.POST("/comments/batch-delete", h.BatchDeleteComments) // 批量删除评论
			auth.POST("/comments/batch-fold", h.BatchFoldComments)     // 批量折叠评论
			auth.GET("/sensitive-words", h.GetSensitiveWords)          // 获取敏感词列表
			auth.POST("/sensitive-words", h.AddSensitiveWord)          // 添加敏感词
		}
	}
}

// ==================== 评论发表接口 ====================

// CreateRootComment 发表根评论
func (h *CommentV3Handler) CreateRootComment(c *gin.Context) {
	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}
	req.UserID = userID.(string)

	// 获取IP地址
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	comment, err := h.commentService.CreateRootComment(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, comment)
}

// CreateReplyComment 发表回复评论
func (h *CommentV3Handler) CreateReplyComment(c *gin.Context) {
	var req service.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	req.UserID = userID.(string)
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	comment, err := h.commentService.CreateReplyComment(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, comment)
}

// DeleteComment 删除评论
func (h *CommentV3Handler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.DeleteComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 评论查询接口 ====================

// GetCommentDetail 获取评论详情
func (h *CommentV3Handler) GetCommentDetail(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	comment, err := h.commentService.GetCommentDetail(commentID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, comment)
}

// GetRootComments 获取根评论列表
func (h *CommentV3Handler) GetRootComments(c *gin.Context) {
	targetType := parseInt8(c.Query("target_type"))
	targetID := parseUint64(c.Query("target_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sortBy := c.DefaultQuery("sort_by", "hot") // hot, like, time

	if targetType == 0 || targetID == 0 {
		response.BadRequest(c, "target_type 和 target_id 不能为空")
		return
	}

	result, err := h.commentService.GetRootComments(targetType, targetID, page, pageSize, sortBy)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetSubComments 获取子评论列表
func (h *CommentV3Handler) GetSubComments(c *gin.Context) {
	parentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的父评论ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.commentService.GetSubComments(parentID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetCommentTree 获取评论树
func (h *CommentV3Handler) GetCommentTree(c *gin.Context) {
	rootID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的根评论ID")
		return
	}

	tree, err := h.commentService.GetCommentTree(rootID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, tree)
}

// GetHotComments 获取热评列表
func (h *CommentV3Handler) GetHotComments(c *gin.Context) {
	targetType := parseInt8(c.Query("target_type"))
	targetID := parseUint64(c.Query("target_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if targetType == 0 || targetID == 0 {
		response.BadRequest(c, "target_type 和 target_id 不能为空")
		return
	}

	comments, err := h.commentService.GetHotComments(targetType, targetID, limit)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, comments)
}

// GetCommentStats 获取评论统计
func (h *CommentV3Handler) GetCommentStats(c *gin.Context) {
	targetType := parseInt8(c.Query("target_type"))
	targetID := parseUint64(c.Query("target_id"))

	if targetType == 0 || targetID == 0 {
		response.BadRequest(c, "target_type 和 target_id 不能为空")
		return
	}

	stats, err := h.commentService.GetCommentStats(targetType, targetID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// ==================== 互动功能接口 ====================

// LikeComment 点赞评论
func (h *CommentV3Handler) LikeComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.LikeComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnlikeComment 取消点赞
func (h *CommentV3Handler) UnlikeComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.UnlikeComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DislikeComment 点踩评论
func (h *CommentV3Handler) DislikeComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.DislikeComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UndislikeComment 取消点踩
func (h *CommentV3Handler) UndislikeComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.UndislikeComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 举报功能接口 ====================

// ReportComment 举报评论
func (h *CommentV3Handler) ReportComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	var req service.ReportCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	req.CommentID = commentID
	req.ReporterUserID = userID.(string)

	if err := h.commentService.ReportComment(&req); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetPendingReports 获取待审核举报
func (h *CommentV3Handler) GetPendingReports(c *gin.Context) {
	// TODO: 验证管理员权限

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reports, total, err := h.commentService.GetPendingReports(page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"reports":   reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// HandleReport 审核举报
func (h *CommentV3Handler) HandleReport(c *gin.Context) {
	// TODO: 验证管理员权限

	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的举报ID")
		return
	}

	var req struct {
		Result   string `json:"result" binding:"required"`
		Approved bool   `json:"approved"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	adminID, _ := c.Get("user_id")
	if err := h.commentService.HandleReport(reportID, adminID.(string), req.Result, req.Approved); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== UP主功能接口 ====================

// AddAuthorReply UP主追评
func (h *CommentV3Handler) AddAuthorReply(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.AddAuthorReply(commentID, userID.(string), req.Content); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// PinComment 置顶评论
func (h *CommentV3Handler) PinComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.PinComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnpinComment 取消置顶
func (h *CommentV3Handler) UnpinComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.UnpinComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// FeatureComment 精选评论
func (h *CommentV3Handler) FeatureComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.FeatureComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnfeatureComment 取消精选
func (h *CommentV3Handler) UnfeatureComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.commentService.UnfeatureComment(commentID, userID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 统计功能接口 ====================

// GetUserCommentStats 获取用户评论统计
func (h *CommentV3Handler) GetUserCommentStats(c *gin.Context) {
	userID, _ := c.Get("user_id")

	stats, err := h.commentService.GetUserCommentStats(userID.(string))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// ==================== 管理功能接口 ====================

// BatchDeleteComments 批量删除评论
func (h *CommentV3Handler) BatchDeleteComments(c *gin.Context) {
	// TODO: 验证管理员权限

	var req struct {
		CommentIDs []uint64 `json:"comment_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	adminID, _ := c.Get("user_id")
	if err := h.commentService.BatchDeleteComments(req.CommentIDs, adminID.(string)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchFoldComments 批量折叠评论
func (h *CommentV3Handler) BatchFoldComments(c *gin.Context) {
	// TODO: 验证管理员权限

	var req struct {
		CommentIDs []uint64 `json:"comment_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.commentService.BatchFoldComments(req.CommentIDs); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetSensitiveWords 获取敏感词列表
func (h *CommentV3Handler) GetSensitiveWords(c *gin.Context) {
	// TODO: 验证管理员权限

	words, err := h.commentService.GetSensitiveWords()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, words)
}

// AddSensitiveWord 添加敏感词
func (h *CommentV3Handler) AddSensitiveWord(c *gin.Context) {
	// TODO: 验证管理员权限

	var req struct {
		Word   string `json:"word" binding:"required"`
		Level  int8   `json:"level" binding:"required,min=1,max=3"`
		Action int8   `json:"action" binding:"required,min=1,max=3"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.commentService.AddSensitiveWord(req.Word, req.Level, req.Action); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 辅助方法 ====================

func parseUint64(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}

func parseInt8(s string) int8 {
	val, _ := strconv.ParseInt(s, 10, 8)
	return int8(val)
}
