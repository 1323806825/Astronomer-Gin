package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

// ==================== UP主追评实现 ====================

func (r *commentV3Repository) CreateAuthorReply(reply *model.CommentAuthorReply) error {
	return r.db.Create(reply).Error
}

func (r *commentV3Repository) FindAuthorReplies(commentID uint64) ([]model.CommentAuthorReply, error) {
	var replies []model.CommentAuthorReply
	err := r.db.Where("comment_id = ?", commentID).
		Order("create_time ASC").
		Find(&replies).Error
	return replies, err
}

func (r *commentV3Repository) DeleteAuthorReply(id uint64) error {
	return r.db.Delete(&model.CommentAuthorReply{}, id).Error
}

// ==================== 表情包管理实现 ====================

func (r *commentV3Repository) CreateEmotion(emotion *model.CommentEmotion) error {
	return r.db.Create(emotion).Error
}

func (r *commentV3Repository) FindEmotionByID(id uint64) (*model.CommentEmotion, error) {
	var emotion model.CommentEmotion
	err := r.db.Where("id = ?", id).First(&emotion).Error
	if err != nil {
		return nil, err
	}
	return &emotion, nil
}

func (r *commentV3Repository) FindAllEmotions() ([]model.CommentEmotion, error) {
	var emotions []model.CommentEmotion
	err := r.db.Order("sort_order ASC, use_count DESC").Find(&emotions).Error
	return emotions, err
}

func (r *commentV3Repository) FindHotEmotions(limit int) ([]model.CommentEmotion, error) {
	var emotions []model.CommentEmotion
	err := r.db.Where("is_hot = ?", true).
		Order("use_count DESC").
		Limit(limit).
		Find(&emotions).Error
	return emotions, err
}

func (r *commentV3Repository) IncrementEmotionUseCount(id uint64) error {
	return r.db.Model(&model.CommentEmotion{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

// ==================== 敏感词管理实现 ====================

func (r *commentV3Repository) CreateSensitiveWord(word *model.CommentSensitiveWord) error {
	return r.db.Create(word).Error
}

func (r *commentV3Repository) UpdateSensitiveWord(word *model.CommentSensitiveWord) error {
	return r.db.Save(word).Error
}

func (r *commentV3Repository) FindSensitiveWordByWord(word string) (*model.CommentSensitiveWord, error) {
	var sensitiveWord model.CommentSensitiveWord
	err := r.db.Where("word = ?", word).First(&sensitiveWord).Error
	if err != nil {
		return nil, err
	}
	return &sensitiveWord, nil
}

func (r *commentV3Repository) FindAllSensitiveWords() ([]model.CommentSensitiveWord, error) {
	var words []model.CommentSensitiveWord
	err := r.db.Order("level DESC, create_time DESC").Find(&words).Error
	return words, err
}

func (r *commentV3Repository) FindEnabledSensitiveWords() ([]model.CommentSensitiveWord, error) {
	var words []model.CommentSensitiveWord
	err := r.db.Where("is_enabled = ?", true).
		Order("level DESC").
		Find(&words).Error
	return words, err
}

func (r *commentV3Repository) DeleteSensitiveWord(id uint64) error {
	return r.db.Delete(&model.CommentSensitiveWord{}, id).Error
}

// ==================== 统计管理实现 ====================

func (r *commentV3Repository) CreateStats(stats *model.CommentStats) error {
	return r.db.Create(stats).Error
}

func (r *commentV3Repository) UpdateStats(stats *model.CommentStats) error {
	return r.db.Save(stats).Error
}

func (r *commentV3Repository) FindStatsByTarget(targetType int8, targetID uint64) (*model.CommentStats, error) {
	var stats model.CommentStats
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *commentV3Repository) IncrementTotalCommentCount(targetType int8, targetID uint64) error {
	// 先尝试更新
	result := r.db.Model(&model.CommentStats{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("total_comment_count", gorm.Expr("total_comment_count + 1"))

	if result.Error != nil {
		return result.Error
	}

	// 如果没有记录，则创建
	if result.RowsAffected == 0 {
		stats := &model.CommentStats{
			TargetType:        targetType,
			TargetID:          targetID,
			TotalCommentCount: 1,
		}
		return r.db.Create(stats).Error
	}

	return nil
}

func (r *commentV3Repository) IncrementTodayCommentCount(targetType int8, targetID uint64) error {
	result := r.db.Model(&model.CommentStats{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("today_comment_count", gorm.Expr("today_comment_count + 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		stats := &model.CommentStats{
			TargetType:        targetType,
			TargetID:          targetID,
			TodayCommentCount: 1,
		}
		return r.db.Create(stats).Error
	}

	return nil
}

func (r *commentV3Repository) IncrementRootCommentCount(targetType int8, targetID uint64) error {
	result := r.db.Model(&model.CommentStats{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("root_comment_count", gorm.Expr("root_comment_count + 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		stats := &model.CommentStats{
			TargetType:       targetType,
			TargetID:         targetID,
			RootCommentCount: 1,
		}
		return r.db.Create(stats).Error
	}

	return nil
}

func (r *commentV3Repository) ResetDailyStats() error {
	return r.db.Model(&model.CommentStats{}).
		Updates(map[string]interface{}{
			"today_comment_count": 0,
		}).Error
}

// ==================== 折叠规则实现 ====================

func (r *commentV3Repository) CreateFoldRule(rule *model.CommentFoldRule) error {
	return r.db.Create(rule).Error
}

func (r *commentV3Repository) UpdateFoldRule(rule *model.CommentFoldRule) error {
	return r.db.Save(rule).Error
}

func (r *commentV3Repository) FindFoldRuleByID(id uint64) (*model.CommentFoldRule, error) {
	var rule model.CommentFoldRule
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *commentV3Repository) FindAllFoldRules() ([]model.CommentFoldRule, error) {
	var rules []model.CommentFoldRule
	err := r.db.Order("create_time DESC").Find(&rules).Error
	return rules, err
}

func (r *commentV3Repository) FindEnabledFoldRules() ([]model.CommentFoldRule, error) {
	var rules []model.CommentFoldRule
	err := r.db.Where("is_enabled = ?", true).Find(&rules).Error
	return rules, err
}

func (r *commentV3Repository) DeleteFoldRule(id uint64) error {
	return r.db.Delete(&model.CommentFoldRule{}, id).Error
}

// ==================== 置顶/精选实现 ====================

func (r *commentV3Repository) PinComment(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("is_pinned", true).Error
}

func (r *commentV3Repository) UnpinComment(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("is_pinned", false).Error
}

func (r *commentV3Repository) FeatureComment(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("is_featured", true).Error
}

func (r *commentV3Repository) UnfeatureComment(id uint64) error {
	return r.db.Model(&model.CommentV3{}).Where("id = ?", id).
		UpdateColumn("is_featured", false).Error
}

func (r *commentV3Repository) FindPinnedComments(targetType int8, targetID uint64) ([]model.CommentV3, error) {
	var comments []model.CommentV3
	err := r.db.Where("target_type = ? AND target_id = ? AND is_pinned = ? AND status = ? AND delete_time IS NULL",
		targetType, targetID, true, model.CommentStatusNormal).
		Order("create_time DESC").
		Find(&comments).Error
	return comments, err
}

func (r *commentV3Repository) FindFeaturedComments(targetType int8, targetID uint64, limit int) ([]model.CommentV3, error) {
	var comments []model.CommentV3
	err := r.db.Where("target_type = ? AND target_id = ? AND is_featured = ? AND status = ? AND delete_time IS NULL",
		targetType, targetID, true, model.CommentStatusNormal).
		Order("quality_score DESC, like_count DESC").
		Limit(limit).
		Find(&comments).Error
	return comments, err
}

// ==================== 批量操作实现 ====================

func (r *commentV3Repository) BatchDelete(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Delete(&model.CommentV3{}, ids).Error
}

func (r *commentV3Repository) BatchUpdateStatus(ids []uint64, status int8) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.CommentV3{}).Where("id IN ?", ids).
		UpdateColumn("status", status).Error
}

func (r *commentV3Repository) BatchFold(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.CommentV3{}).Where("id IN ?", ids).
		UpdateColumn("status", model.CommentStatusFolded).Error
}
