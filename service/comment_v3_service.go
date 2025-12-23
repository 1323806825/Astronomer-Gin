package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/repository"
	"fmt"
	"gorm.io/gorm"
)

// CommentV3Service 企业级评论服务接口
type CommentV3Service interface {
	// ==================== 评论发表 ====================
	// 发表根评论
	CreateRootComment(req *CreateCommentRequest) (*model.CommentV3, error)
	// 发表回复评论
	CreateReplyComment(req *CreateReplyRequest) (*model.CommentV3, error)
	// 删除评论
	DeleteComment(commentID uint64, userID string) error
	// 获取评论详情
	GetCommentDetail(commentID uint64) (*model.CommentV3, error)

	// ==================== 评论查询 ====================
	// 获取根评论列表
	GetRootComments(targetType int8, targetID uint64, page, pageSize int, sortBy string) (*CommentListResponse, error)
	// 获取子评论列表
	GetSubComments(parentID uint64, page, pageSize int) (*CommentListResponse, error)
	// 获取评论树
	GetCommentTree(rootID uint64) ([]model.CommentV3, error)
	// 获取热评
	GetHotComments(targetType int8, targetID uint64, limit int) ([]model.CommentV3, error)

	// ==================== 互动功能 ====================
	// 点赞评论
	LikeComment(commentID uint64, userID string) error
	// 取消点赞
	UnlikeComment(commentID uint64, userID string) error
	// 点踩评论
	DislikeComment(commentID uint64, userID string) error
	// 取消点踩
	UndislikeComment(commentID uint64, userID string) error

	// ==================== 举报功能 ====================
	// 举报评论
	ReportComment(req *ReportCommentRequest) error
	// 审核举报
	HandleReport(reportID uint64, adminID, result string, approved bool) error
	// 获取待审核举报
	GetPendingReports(page, pageSize int) ([]model.CommentReport, int64, error)

	// ==================== UP主功能 ====================
	// UP主追评
	AddAuthorReply(commentID uint64, authorID string, content string) error
	// 置顶评论
	PinComment(commentID uint64, authorID string) error
	// 取消置顶
	UnpinComment(commentID uint64, authorID string) error
	// 精选评论
	FeatureComment(commentID uint64, authorID string) error
	// 取消精选
	UnfeatureComment(commentID uint64, authorID string) error

	// ==================== 热评管理 ====================
	// 计算评论热度分数
	CalculateCommentHotScore(commentID uint64) (float64, error)
	// 更新热评榜单（定时任务）
	UpdateHotCommentList(targetType int8, targetID uint64) error
	// 批量更新热度分数
	BatchUpdateHotScores(targetType int8, targetID uint64) error

	// ==================== 统计功能 ====================
	// 获取评论统计
	GetCommentStats(targetType int8, targetID uint64) (*model.CommentStats, error)
	// 获取用户评论统计
	GetUserCommentStats(userID string) (*UserCommentStats, error)

	// ==================== 管理功能 ====================
	// 批量删除评论
	BatchDeleteComments(commentIDs []uint64, adminID string) error
	// 批量折叠评论
	BatchFoldComments(commentIDs []uint64) error
	// 获取敏感词列表
	GetSensitiveWords() ([]model.CommentSensitiveWord, error)
	// 添加敏感词
	AddSensitiveWord(word string, level int8, action int8) error
}

// ==================== 请求/响应结构体 ====================

// CreateCommentRequest 创建根评论请求
type CreateCommentRequest struct {
	TargetType  int8     `json:"target_type" binding:"required"`
	TargetID    uint64   `json:"target_id" binding:"required"`
	UserID      string   `json:"user_id" binding:"required"`
	Content     string   `json:"content" binding:"required,min=1,max=5000"`
	ContentType int8     `json:"content_type"`
	Images      []string `json:"images"`
	AtUserIDs   []uint64 `json:"at_user_ids"`
	IP          string   `json:"ip"`
	IPLocation  string   `json:"ip_location"`
	DeviceType  string   `json:"device_type"`
	UserAgent   string   `json:"user_agent"`
}

// CreateReplyRequest 创建回复评论请求
type CreateReplyRequest struct {
	ParentID         uint64   `json:"parent_id" binding:"required"`
	ReplyToCommentID uint64   `json:"reply_to_comment_id"`
	ReplyToUserID    string   `json:"reply_to_user_id"`
	UserID           string   `json:"user_id" binding:"required"`
	Content          string   `json:"content" binding:"required,min=1,max=5000"`
	ContentType      int8     `json:"content_type"`
	Images           []string `json:"images"`
	AtUserIDs        []uint64 `json:"at_user_ids"`
	IP               string   `json:"ip"`
	IPLocation       string   `json:"ip_location"`
	DeviceType       string   `json:"device_type"`
	UserAgent        string   `json:"user_agent"`
}

// ReportCommentRequest 举报评论请求
type ReportCommentRequest struct {
	CommentID      uint64 `json:"comment_id" binding:"required"`
	ReporterUserID string `json:"reporter_user_id" binding:"required"`
	ReasonType     int8   `json:"reason_type" binding:"required"`
	ReasonDesc     string `json:"reason_desc" binding:"max=500"`
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments []model.CommentV3 `json:"comments"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// UserCommentStats 用户评论统计
type UserCommentStats struct {
	TotalComments   int64   `json:"total_comments"`
	TotalLikes      int64   `json:"total_likes"`
	TotalReplies    int64   `json:"total_replies"`
	AvgLikeCount    float64 `json:"avg_like_count"`
	HotCommentCount int     `json:"hot_comment_count"`
	PinnedCount     int     `json:"pinned_count"`
	FeaturedCount   int     `json:"featured_count"`
}

// ==================== Service实现 ====================

type commentV3Service struct {
	commentRepo repository.CommentV3Repository
	articleRepo repository.ArticleV3Repository
	userRepo    repository.UserRepository
	likeRepo    repository.LikeRepository
	notifyRepo  repository.NotificationRepository
	db          *gorm.DB
}

// NewCommentV3Service 创建CommentV3Service实例
func NewCommentV3Service(
	commentRepo repository.CommentV3Repository,
	articleRepo repository.ArticleV3Repository,
	userRepo repository.UserRepository,
	likeRepo repository.LikeRepository,
	notifyRepo repository.NotificationRepository,
	db *gorm.DB,
) CommentV3Service {
	return &commentV3Service{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
		userRepo:    userRepo,
		likeRepo:    likeRepo,
		notifyRepo:  notifyRepo,
		db:          db,
	}
}

// ==================== 评论发表实现 ====================

// CreateRootComment 发表根评论
func (s *commentV3Service) CreateRootComment(req *CreateCommentRequest) (*model.CommentV3, error) {
	// 1. 参数验证
	if err := s.validateCommentContent(req.Content); err != nil {
		return nil, err
	}

	// 2. 敏感词检查
	if err := s.checkCommentSensitiveWords(req.Content); err != nil {
		return nil, err
	}

	// 3. 获取用户信息
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 4. 获取楼层号
	floorNumber, err := s.commentRepo.GetNextFloorNumber(req.TargetType, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("获取楼层号失败: %w", err)
	}

	// 5. 检查是否为作者
	isAuthor := s.isTargetAuthor(req.TargetType, req.TargetID, req.UserID)

	// 6. 构建评论对象
	comment := &model.CommentV3{
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		UserID:         req.UserID,
		Username:       user.Username,
		UserAvatar:     user.Icon,
		ParentID:       0, // 根评论
		RootID:         0, // 自己就是根
		FloorNumber:    floorNumber,
		SubFloorNumber: 0,
		Depth:          0,
		Content:        req.Content,
		ContentType:    req.ContentType,
		Images:         model.JSONStringList(req.Images),
		AtUserIDs:      model.JSONUint64List(req.AtUserIDs),
		Status:         model.CommentStatusNormal,
		IsAuthor:       isAuthor,
		IP:             req.IP,
		IPLocation:     req.IPLocation,
		DeviceType:     req.DeviceType,
		UserAgent:      req.UserAgent,
		AuditStatus:    model.CommentAuditStatusPending, // 待审核
	}

	// 7. 内容审核
	auditResult := s.auditCommentContent(comment)
	comment.AuditStatus = auditResult.Status
	comment.RiskLevel = auditResult.RiskLevel

	// 如果风险等级高，直接折叠
	if comment.RiskLevel >= model.RiskLevelHigh {
		comment.Status = model.CommentStatusFolded
	}

	// 8. 创建评论
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	// 9. 设置根评论ID（自己）
	s.commentRepo.UpdateFields(comment.ID, map[string]interface{}{"root_id": comment.ID})

	// 10. 更新文章评论数
	if req.TargetType == model.CommentTargetTypeArticle {
		s.articleRepo.IncrementCommentCount(req.TargetID)
	}

	// 11. 更新评论统计
	s.commentRepo.IncrementTotalCommentCount(req.TargetType, req.TargetID)
	s.commentRepo.IncrementTodayCommentCount(req.TargetType, req.TargetID)
	s.commentRepo.IncrementRootCommentCount(req.TargetType, req.TargetID)

	// 12. 处理盖楼
	s.handleFloorBuilding(req.TargetType, req.TargetID, req.UserID, comment.ID)

	return comment, nil
}

// CreateReplyComment 发表回复评论
func (s *commentV3Service) CreateReplyComment(req *CreateReplyRequest) (*model.CommentV3, error) {
	// 1. 参数验证
	if err := s.validateCommentContent(req.Content); err != nil {
		return nil, err
	}

	// 2. 敏感词检查
	if err := s.checkCommentSensitiveWords(req.Content); err != nil {
		return nil, err
	}

	// 3. 获取父评论
	parentComment, err := s.commentRepo.FindByID(req.ParentID)
	if err != nil {
		return nil, fmt.Errorf("父评论不存在")
	}

	// 4. 获取用户信息
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 5. 获取子楼层号
	subFloorNumber, err := s.commentRepo.GetNextSubFloorNumber(req.ParentID)
	if err != nil {
		return nil, fmt.Errorf("获取子楼层号失败: %w", err)
	}

	// 6. 构建回复链路
	replyChain, err := s.commentRepo.BuildReplyChain(req.ParentID)
	if err != nil {
		return nil, fmt.Errorf("构建回复链失败: %w", err)
	}

	// 7. 计算深度
	depth, err := s.commentRepo.CalculateDepth(req.ParentID)
	if err != nil {
		return nil, fmt.Errorf("计算深度失败: %w", err)
	}

	// 8. 确定根评论ID
	rootID := parentComment.RootID
	if rootID == 0 {
		rootID = parentComment.ID // 父评论就是根评论
	}

	// 9. 检查是否为作者
	isAuthor := s.isTargetAuthor(parentComment.TargetType, parentComment.TargetID, req.UserID)

	// 10. 构建评论对象
	comment := &model.CommentV3{
		TargetType:       parentComment.TargetType,
		TargetID:         parentComment.TargetID,
		UserID:           req.UserID,
		Username:         user.Username,
		UserAvatar:       user.Icon,
		ParentID:         req.ParentID,
		RootID:           rootID,
		ReplyToUserID:    req.ReplyToUserID,
		ReplyToCommentID: req.ReplyToCommentID,
		FloorNumber:      parentComment.FloorNumber, // 继承父评论的楼层号
		SubFloorNumber:   subFloorNumber,            // 子楼层号
		ReplyChain:       replyChain,
		Depth:            depth,
		Content:          req.Content,
		ContentType:      req.ContentType,
		Images:           model.JSONStringList(req.Images),
		AtUserIDs:        model.JSONUint64List(req.AtUserIDs),
		Status:           model.CommentStatusNormal,
		IsAuthor:         isAuthor,
		IP:               req.IP,
		IPLocation:       req.IPLocation,
		DeviceType:       req.DeviceType,
		UserAgent:        req.UserAgent,
		AuditStatus:      model.CommentAuditStatusPending,
	}

	// 11. 内容审核
	auditResult := s.auditCommentContent(comment)
	comment.AuditStatus = auditResult.Status
	comment.RiskLevel = auditResult.RiskLevel

	if comment.RiskLevel >= model.RiskLevelHigh {
		comment.Status = model.CommentStatusFolded
	}

	// 12. 创建评论
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, fmt.Errorf("创建回复失败: %w", err)
	}

	// 13. 更新父评论回复数
	s.commentRepo.IncrementReplyCount(req.ParentID)

	// 14. 更新根评论总回复数
	s.commentRepo.IncrementTotalReplyCount(rootID)

	// 15. 更新文章评论数
	if parentComment.TargetType == model.CommentTargetTypeArticle {
		s.articleRepo.IncrementCommentCount(parentComment.TargetID)
	}

	// 16. 更新评论统计
	s.commentRepo.IncrementTotalCommentCount(parentComment.TargetType, parentComment.TargetID)
	s.commentRepo.IncrementTodayCommentCount(parentComment.TargetType, parentComment.TargetID)

	// 17. 处理盖楼
	s.handleFloorBuilding(parentComment.TargetType, parentComment.TargetID, req.UserID, comment.ID)

	return comment, nil
}

// DeleteComment 删除评论
func (s *commentV3Service) DeleteComment(commentID uint64, userID string) error {
	// 1. 检查权限
	if !s.commentRepo.CheckOwnership(commentID, userID) {
		return constant.ErrPermissionDenied
	}

	// 2. 获取评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 3. 软删除
	if err := s.commentRepo.SoftDelete(commentID); err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}

	// 4. 更新父评论回复数
	if comment.ParentID > 0 {
		s.commentRepo.DecrementReplyCount(comment.ParentID)
	}

	// 5. 更新根评论总回复数
	if comment.RootID > 0 && comment.RootID != commentID {
		s.commentRepo.DecrementTotalReplyCount(comment.RootID)
	}

	// 6. 更新文章评论数
	if comment.TargetType == model.CommentTargetTypeArticle {
		s.articleRepo.DecrementCommentCount(comment.TargetID)
	}

	return nil
}

// GetCommentDetail 获取评论详情
func (s *commentV3Service) GetCommentDetail(commentID uint64) (*model.CommentV3, error) {
	return s.commentRepo.FindByID(commentID)
}

// ==================== 评论查询实现 ====================

// GetRootComments 获取根评论列表
func (s *commentV3Service) GetRootComments(targetType int8, targetID uint64, page, pageSize int, sortBy string) (*CommentListResponse, error) {
	comments, total, err := s.commentRepo.FindRootComments(targetType, targetID, page, pageSize, sortBy)
	if err != nil {
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}

	return &CommentListResponse{
		Comments: comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetSubComments 获取子评论列表
func (s *commentV3Service) GetSubComments(parentID uint64, page, pageSize int) (*CommentListResponse, error) {
	comments, total, err := s.commentRepo.FindSubComments(parentID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询子评论失败: %w", err)
	}

	return &CommentListResponse{
		Comments: comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetCommentTree 获取评论树
func (s *commentV3Service) GetCommentTree(rootID uint64) ([]model.CommentV3, error) {
	return s.commentRepo.FindCommentTree(rootID)
}

// GetHotComments 获取热评
func (s *commentV3Service) GetHotComments(targetType int8, targetID uint64, limit int) ([]model.CommentV3, error) {
	return s.commentRepo.FindHotComments(targetType, targetID, limit)
}

// ==================== 互动功能实现 ====================

// LikeComment 点赞评论
func (s *commentV3Service) LikeComment(commentID uint64, userID string) error {
	// 1. 检查是否已点赞
	if s.commentRepo.IsLiked(commentID, userID) {
		return fmt.Errorf("已点赞")
	}

	// 2. 创建互动记录
	interaction := &model.CommentInteraction{
		CommentID:  commentID,
		UserID:     userID,
		ActionType: model.InteractionTypeLike,
	}

	if err := s.commentRepo.CreateInteraction(interaction); err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}

	// 3. 增加点赞数
	s.commentRepo.IncrementLikeCount(commentID)

	// 4. 重新计算热度
	s.CalculateCommentHotScore(commentID)

	return nil
}

// UnlikeComment 取消点赞
func (s *commentV3Service) UnlikeComment(commentID uint64, userID string) error {
	// 1. 删除互动记录
	if err := s.commentRepo.DeleteInteraction(commentID, userID, model.InteractionTypeLike); err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}

	// 2. 减少点赞数
	s.commentRepo.DecrementLikeCount(commentID)

	// 3. 重新计算热度
	s.CalculateCommentHotScore(commentID)

	return nil
}

// DislikeComment 点踩评论
func (s *commentV3Service) DislikeComment(commentID uint64, userID string) error {
	// 1. 检查是否已踩
	if s.commentRepo.IsDisliked(commentID, userID) {
		return fmt.Errorf("已踩")
	}

	// 2. 创建互动记录
	interaction := &model.CommentInteraction{
		CommentID:  commentID,
		UserID:     userID,
		ActionType: model.InteractionTypeDislike,
	}

	if err := s.commentRepo.CreateInteraction(interaction); err != nil {
		return fmt.Errorf("点踩失败: %w", err)
	}

	// 3. 增加踩数
	s.commentRepo.IncrementDislikeCount(commentID)

	return nil
}

// UndislikeComment 取消点踩
func (s *commentV3Service) UndislikeComment(commentID uint64, userID string) error {
	// 1. 删除互动记录
	if err := s.commentRepo.DeleteInteraction(commentID, userID, model.InteractionTypeDislike); err != nil {
		return fmt.Errorf("取消点踩失败: %w", err)
	}

	// 2. 减少踩数
	s.commentRepo.DecrementDislikeCount(commentID)

	return nil
}

// 继续在下一个文件...
