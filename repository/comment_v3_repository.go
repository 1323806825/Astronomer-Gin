package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
	"time"
)

// CommentV3Repository 企业级评论Repository接口
type CommentV3Repository interface {
	// ==================== 基础CRUD ====================
	Create(comment *model.CommentV3) error
	FindByID(id uint64) (*model.CommentV3, error)
	Update(comment *model.CommentV3) error
	UpdateFields(id uint64, fields map[string]interface{}) error
	Delete(id uint64) error     // 硬删除
	SoftDelete(id uint64) error // 软删除
	CheckOwnership(id uint64, userID string) bool

	// ==================== 评论查询 ====================
	// 获取根评论列表（一级评论）
	FindRootComments(targetType int8, targetID uint64, page, pageSize int, sortBy string) ([]model.CommentV3, int64, error)
	// 获取子评论列表（某个评论的所有回复）
	FindSubComments(parentID uint64, page, pageSize int) ([]model.CommentV3, int64, error)
	// 获取评论树（根评论及其所有子评论）
	FindCommentTree(rootID uint64) ([]model.CommentV3, error)
	// 获取用户的所有评论
	FindByUserID(userID string, page, pageSize int) ([]model.CommentV3, int64, error)
	// 获取对象的所有评论
	FindByTarget(targetType int8, targetID uint64, page, pageSize int) ([]model.CommentV3, int64, error)

	// ==================== 楼层管理 ====================
	// 获取下一个楼层号
	GetNextFloorNumber(targetType int8, targetID uint64) (int, error)
	// 获取下一个子楼层号
	GetNextSubFloorNumber(parentID uint64) (int, error)
	// 根据楼层号查询评论
	FindByFloorNumber(targetType int8, targetID uint64, floorNumber int) (*model.CommentV3, error)

	// ==================== 回复链管理 ====================
	// 构建回复链路
	BuildReplyChain(parentID uint64) (model.JSONUint64List, error)
	// 计算评论深度
	CalculateDepth(parentID uint64) (int, error)

	// ==================== 热评管理 ====================
	// 获取热门评论
	FindHotComments(targetType int8, targetID uint64, limit int) ([]model.CommentV3, error)
	// 创建热评记录
	CreateHotList(hotList *model.CommentHotList) error
	// 批量更新热评榜单
	BatchUpdateHotList(targetType int8, targetID uint64, hotComments []model.CommentV3) error
	// 获取热评榜单
	FindHotList(targetType int8, targetID uint64, limit int) ([]model.CommentHotList, error)
	// 清空热评榜单
	ClearHotList(targetType int8, targetID uint64) error

	// ==================== 统计字段更新 ====================
	IncrementLikeCount(id uint64) error
	DecrementLikeCount(id uint64) error
	IncrementDislikeCount(id uint64) error
	DecrementDislikeCount(id uint64) error
	IncrementReplyCount(id uint64) error
	DecrementReplyCount(id uint64) error
	IncrementTotalReplyCount(id uint64) error
	DecrementTotalReplyCount(id uint64) error
	UpdateHotScore(id uint64, score float64) error
	UpdateQualityScore(id uint64, score float64) error

	// ==================== 互动管理 ====================
	CreateInteraction(interaction *model.CommentInteraction) error
	DeleteInteraction(commentID uint64, userID string, actionType int8) error
	FindInteraction(commentID uint64, userID string, actionType int8) (*model.CommentInteraction, error)
	IsLiked(commentID uint64, userID string) bool
	IsDisliked(commentID uint64, userID string) bool

	// ==================== 举报管理 ====================
	CreateReport(report *model.CommentReport) error
	UpdateReport(report *model.CommentReport) error
	FindReportByID(id uint64) (*model.CommentReport, error)
	FindReportsByCommentID(commentID uint64) ([]model.CommentReport, error)
	FindPendingReports(page, pageSize int) ([]model.CommentReport, int64, error)
	GetReportCount(commentID uint64) (int64, error)

	// ==================== 盖楼管理 ====================
	CreateFloorBuilding(building *model.CommentFloorBuilding) error
	UpdateFloorBuilding(building *model.CommentFloorBuilding) error
	FindFloorBuilding(targetType int8, targetID uint64, userID string) (*model.CommentFloorBuilding, error)
	IncrementFloorCount(targetType int8, targetID uint64, userID string) error

	// ==================== UP主追评 ====================
	CreateAuthorReply(reply *model.CommentAuthorReply) error
	FindAuthorReplies(commentID uint64) ([]model.CommentAuthorReply, error)
	DeleteAuthorReply(id uint64) error

	// ==================== 表情包管理 ====================
	CreateEmotion(emotion *model.CommentEmotion) error
	FindEmotionByID(id uint64) (*model.CommentEmotion, error)
	FindAllEmotions() ([]model.CommentEmotion, error)
	FindHotEmotions(limit int) ([]model.CommentEmotion, error)
	IncrementEmotionUseCount(id uint64) error

	// ==================== 敏感词管理 ====================
	CreateSensitiveWord(word *model.CommentSensitiveWord) error
	UpdateSensitiveWord(word *model.CommentSensitiveWord) error
	FindSensitiveWordByWord(word string) (*model.CommentSensitiveWord, error)
	FindAllSensitiveWords() ([]model.CommentSensitiveWord, error)
	FindEnabledSensitiveWords() ([]model.CommentSensitiveWord, error)
	DeleteSensitiveWord(id uint64) error

	// ==================== 统计管理 ====================
	CreateStats(stats *model.CommentStats) error
	UpdateStats(stats *model.CommentStats) error
	FindStatsByTarget(targetType int8, targetID uint64) (*model.CommentStats, error)
	IncrementTotalCommentCount(targetType int8, targetID uint64) error
	IncrementTodayCommentCount(targetType int8, targetID uint64) error
	IncrementRootCommentCount(targetType int8, targetID uint64) error
	ResetDailyStats() error

	// ==================== 折叠规则 ====================
	CreateFoldRule(rule *model.CommentFoldRule) error
	UpdateFoldRule(rule *model.CommentFoldRule) error
	FindFoldRuleByID(id uint64) (*model.CommentFoldRule, error)
	FindAllFoldRules() ([]model.CommentFoldRule, error)
	FindEnabledFoldRules() ([]model.CommentFoldRule, error)
	DeleteFoldRule(id uint64) error

	// ==================== 置顶/精选 ====================
	PinComment(id uint64) error
	UnpinComment(id uint64) error
	FeatureComment(id uint64) error
	UnfeatureComment(id uint64) error
	FindPinnedComments(targetType int8, targetID uint64) ([]model.CommentV3, error)
	FindFeaturedComments(targetType int8, targetID uint64, limit int) ([]model.CommentV3, error)

	// ==================== 批量操作 ====================
	BatchDelete(ids []uint64) error
	BatchUpdateStatus(ids []uint64, status int8) error
	BatchFold(ids []uint64) error
}

type commentV3Repository struct {
	db *gorm.DB
}

// NewCommentV3Repository 创建CommentV3Repository实例
func NewCommentV3Repository(db *gorm.DB) CommentV3Repository {
	return &commentV3Repository{db: db}
}

// ==================== 基础CRUD实现 ====================

func (r *commentV3Repository) Create(comment *model.CommentV3) error {
	return r.db.Create(comment).Error
}

func (r *commentV3Repository) FindByID(id uint64) (*model.CommentV3, error) {
	var comment model.CommentV3
	err := r.db.Where("id = ? AND delete_time IS NULL", id).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentV3Repository) Update(comment *model.CommentV3) error {
	return r.db.Save(comment).Error
}

func (r *commentV3Repository) UpdateFields(id uint64, fields map[string]interface{}) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).Updates(fields).Error
}

func (r *commentV3Repository) Delete(id uint64) error {
	return r.db.Delete(&model.CommentV3{}, id).Error
}

func (r *commentV3Repository) SoftDelete(id uint64) error {
	now := time.Now()
	return r.db.Model(&model.CommentV3{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.CommentStatusDeleted,
			"delete_time": &now,
		}).Error
}

func (r *commentV3Repository) CheckOwnership(id uint64, userID string) bool {
	var count int64
	r.db.Model(&model.CommentV3{}).Where("id = ? AND user_id = ?", id, userID).Count(&count)
	return count > 0
}

// ==================== 评论查询实现 ====================

func (r *commentV3Repository) FindRootComments(targetType int8, targetID uint64, page, pageSize int, sortBy string) ([]model.CommentV3, int64, error) {
	var comments []model.CommentV3
	var total int64

	query := r.db.Model(&model.CommentV3{}).
		Where("target_type = ? AND target_id = ? AND parent_id = 0 AND status = ? AND delete_time IS NULL",
			targetType, targetID, model.CommentStatusNormal)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序方式
	orderSQL := "create_time DESC" // 默认按时间倒序
	switch sortBy {
	case "hot":
		orderSQL = "hot_score DESC, like_count DESC, create_time DESC"
	case "like":
		orderSQL = "like_count DESC, create_time DESC"
	case "time_asc":
		orderSQL = "create_time ASC"
	}

	// 置顶评论优先
	orderSQL = "is_pinned DESC, " + orderSQL

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order(orderSQL).Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentV3Repository) FindSubComments(parentID uint64, page, pageSize int) ([]model.CommentV3, int64, error) {
	var comments []model.CommentV3
	var total int64

	query := r.db.Model(&model.CommentV3{}).
		Where("parent_id = ? AND status = ? AND delete_time IS NULL",
			parentID, model.CommentStatusNormal)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time ASC").Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentV3Repository) FindCommentTree(rootID uint64) ([]model.CommentV3, error) {
	var comments []model.CommentV3
	err := r.db.Where("root_id = ? AND status = ? AND delete_time IS NULL",
		rootID, model.CommentStatusNormal).
		Order("depth ASC, create_time ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentV3Repository) FindByUserID(userID string, page, pageSize int) ([]model.CommentV3, int64, error) {
	var comments []model.CommentV3
	var total int64

	query := r.db.Model(&model.CommentV3{}).
		Where("user_id = ? AND status = ? AND delete_time IS NULL",
			userID, model.CommentStatusNormal)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentV3Repository) FindByTarget(targetType int8, targetID uint64, page, pageSize int) ([]model.CommentV3, int64, error) {
	var comments []model.CommentV3
	var total int64

	query := r.db.Model(&model.CommentV3{}).
		Where("target_type = ? AND target_id = ? AND status = ? AND delete_time IS NULL",
			targetType, targetID, model.CommentStatusNormal)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// ==================== 楼层管理实现 ====================

func (r *commentV3Repository) GetNextFloorNumber(targetType int8, targetID uint64) (int, error) {
	var maxFloor int
	err := r.db.Model(&model.CommentV3{}).
		Where("target_type = ? AND target_id = ? AND parent_id = 0", targetType, targetID).
		Select("COALESCE(MAX(floor_number), 0)").
		Scan(&maxFloor).Error

	if err != nil {
		return 0, err
	}

	return maxFloor + 1, nil
}

func (r *commentV3Repository) GetNextSubFloorNumber(parentID uint64) (int, error) {
	var maxSubFloor int
	err := r.db.Model(&model.CommentV3{}).
		Where("parent_id = ?", parentID).
		Select("COALESCE(MAX(sub_floor_number), 0)").
		Scan(&maxSubFloor).Error

	if err != nil {
		return 0, err
	}

	return maxSubFloor + 1, nil
}

func (r *commentV3Repository) FindByFloorNumber(targetType int8, targetID uint64, floorNumber int) (*model.CommentV3, error) {
	var comment model.CommentV3
	err := r.db.Where("target_type = ? AND target_id = ? AND floor_number = ? AND parent_id = 0",
		targetType, targetID, floorNumber).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// ==================== 回复链管理实现 ====================

func (r *commentV3Repository) BuildReplyChain(parentID uint64) (model.JSONUint64List, error) {
	var chain model.JSONUint64List

	if parentID == 0 {
		return chain, nil
	}

	// 递归查找父评论的回复链
	var parent model.CommentV3
	if err := r.db.Where("id = ?", parentID).First(&parent).Error; err != nil {
		return nil, err
	}

	// 复制父评论的回复链
	if parent.ReplyChain != nil {
		chain = append(chain, parent.ReplyChain...)
	}

	// 添加当前父评论ID
	chain = append(chain, parentID)

	return chain, nil
}

func (r *commentV3Repository) CalculateDepth(parentID uint64) (int, error) {
	if parentID == 0 {
		return 0, nil
	}

	var parent model.CommentV3
	if err := r.db.Where("id = ?", parentID).First(&parent).Error; err != nil {
		return 0, err
	}

	return parent.Depth + 1, nil
}

// ==================== 热评管理实现 ====================

func (r *commentV3Repository) FindHotComments(targetType int8, targetID uint64, limit int) ([]model.CommentV3, error) {
	var comments []model.CommentV3
	err := r.db.Where("target_type = ? AND target_id = ? AND is_hot = ? AND status = ? AND delete_time IS NULL",
		targetType, targetID, true, model.CommentStatusNormal).
		Order("hot_score DESC, like_count DESC").
		Limit(limit).
		Find(&comments).Error
	return comments, err
}

func (r *commentV3Repository) CreateHotList(hotList *model.CommentHotList) error {
	return r.db.Create(hotList).Error
}

func (r *commentV3Repository) BatchUpdateHotList(targetType int8, targetID uint64, hotComments []model.CommentV3) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 清空旧榜单
		if err := tx.Where("target_type = ? AND target_id = ?", targetType, targetID).
			Delete(&model.CommentHotList{}).Error; err != nil {
			return err
		}

		// 2. 插入新榜单
		for i, comment := range hotComments {
			hotList := &model.CommentHotList{
				TargetType:   targetType,
				TargetID:     targetID,
				CommentID:    comment.ID,
				UserID:       comment.UserID,
				Username:     comment.Username,
				Content:      comment.Content,
				LikeCount:    comment.LikeCount,
				RankPosition: i + 1,
				HotScore:     comment.HotScore,
			}
			if err := tx.Create(hotList).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *commentV3Repository) FindHotList(targetType int8, targetID uint64, limit int) ([]model.CommentHotList, error) {
	var hotList []model.CommentHotList
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("rank_position ASC").
		Limit(limit).
		Find(&hotList).Error
	return hotList, err
}

func (r *commentV3Repository) ClearHotList(targetType int8, targetID uint64) error {
	return r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Delete(&model.CommentHotList{}).Error
}

// ==================== 统计字段更新实现 ====================

func (r *commentV3Repository) IncrementLikeCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *commentV3Repository) DecrementLikeCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *commentV3Repository) IncrementDislikeCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("dislike_count", gorm.Expr("dislike_count + 1")).Error
}

func (r *commentV3Repository) DecrementDislikeCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("dislike_count", gorm.Expr("GREATEST(dislike_count - 1, 0)")).Error
}

func (r *commentV3Repository) IncrementReplyCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("reply_count", gorm.Expr("reply_count + 1")).Error
}

func (r *commentV3Repository) DecrementReplyCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("reply_count", gorm.Expr("GREATEST(reply_count - 1, 0)")).Error
}

func (r *commentV3Repository) IncrementTotalReplyCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("total_reply_count", gorm.Expr("total_reply_count + 1")).Error
}

func (r *commentV3Repository) DecrementTotalReplyCount(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("total_reply_count", gorm.Expr("GREATEST(total_reply_count - 1, 0)")).Error
}

func (r *commentV3Repository) UpdateHotScore(id uint64, score float64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("hot_score", score).Error
}

func (r *commentV3Repository) UpdateQualityScore(id uint64, score float64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("quality_score", score).Error
}

// ==================== 互动管理实现 ====================

func (r *commentV3Repository) CreateInteraction(interaction *model.CommentInteraction) error {
	return r.db.Create(interaction).Error
}

func (r *commentV3Repository) DeleteInteraction(commentID uint64, userID string, actionType int8) error {
	return r.db.Where("comment_id = ? AND user_id = ? AND action_type = ?",
		commentID, userID, actionType).Delete(&model.CommentInteraction{}).Error
}

func (r *commentV3Repository) FindInteraction(commentID uint64, userID string, actionType int8) (*model.CommentInteraction, error) {
	var interaction model.CommentInteraction
	err := r.db.Where("comment_id = ? AND user_id = ? AND action_type = ?",
		commentID, userID, actionType).First(&interaction).Error
	if err != nil {
		return nil, err
	}
	return &interaction, nil
}

func (r *commentV3Repository) IsLiked(commentID uint64, userID string) bool {
	var count int64
	r.db.Model(&model.CommentInteraction{}).
		Where("comment_id = ? AND user_id = ? AND action_type = ?",
			commentID, userID, model.InteractionTypeLike).Count(&count)
	return count > 0
}

func (r *commentV3Repository) IsDisliked(commentID uint64, userID string) bool {
	var count int64
	r.db.Model(&model.CommentInteraction{}).
		Where("comment_id = ? AND user_id = ? AND action_type = ?",
			commentID, userID, model.InteractionTypeDislike).Count(&count)
	return count > 0
}

// ==================== 举报管理实现 ====================

func (r *commentV3Repository) CreateReport(report *model.CommentReport) error {
	return r.db.Create(report).Error
}

func (r *commentV3Repository) UpdateReport(report *model.CommentReport) error {
	return r.db.Save(report).Error
}

func (r *commentV3Repository) FindReportByID(id uint64) (*model.CommentReport, error) {
	var report model.CommentReport
	err := r.db.Where("id = ?", id).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *commentV3Repository) FindReportsByCommentID(commentID uint64) ([]model.CommentReport, error) {
	var reports []model.CommentReport
	err := r.db.Where("comment_id = ?", commentID).
		Order("create_time DESC").
		Find(&reports).Error
	return reports, err
}

func (r *commentV3Repository) FindPendingReports(page, pageSize int) ([]model.CommentReport, int64, error) {
	var reports []model.CommentReport
	var total int64

	query := r.db.Model(&model.CommentReport{}).Where("status = ?", model.ReportStatusPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *commentV3Repository) GetReportCount(commentID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.CommentReport{}).
		Where("comment_id = ?", commentID).Count(&count).Error
	return count, err
}

// ==================== 盖楼管理实现 ====================

func (r *commentV3Repository) CreateFloorBuilding(building *model.CommentFloorBuilding) error {
	return r.db.Create(building).Error
}

func (r *commentV3Repository) UpdateFloorBuilding(building *model.CommentFloorBuilding) error {
	return r.db.Save(building).Error
}

func (r *commentV3Repository) FindFloorBuilding(targetType int8, targetID uint64, userID string) (*model.CommentFloorBuilding, error) {
	var building model.CommentFloorBuilding
	err := r.db.Where("target_type = ? AND target_id = ? AND user_id = ?",
		targetType, targetID, userID).First(&building).Error
	if err != nil {
		return nil, err
	}
	return &building, nil
}

func (r *commentV3Repository) IncrementFloorCount(targetType int8, targetID uint64, userID string) error {
	return r.db.Model(&model.CommentFloorBuilding{}).
		Where("target_type = ? AND target_id = ? AND user_id = ?", targetType, targetID, userID).
		UpdateColumn("floor_count", gorm.Expr("floor_count + 1")).Error
}

// 继续在下一个文件...
