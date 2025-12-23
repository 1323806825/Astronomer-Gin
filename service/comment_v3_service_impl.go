package service

import (
	"astronomer-gin/model"
	"fmt"
	"math"
	"strings"
	"time"
)

// ==================== 举报功能实现 ====================

// ReportComment 举报评论
func (s *commentV3Service) ReportComment(req *ReportCommentRequest) error {
	// 1. 检查评论是否存在
	_, err := s.commentRepo.FindByID(req.CommentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 创建举报记录
	report := &model.CommentReport{
		CommentID:      req.CommentID,
		ReporterUserID: req.ReporterUserID,
		ReasonType:     req.ReasonType,
		ReasonDesc:     req.ReasonDesc,
		Status:         model.ReportStatusPending,
	}

	if err := s.commentRepo.CreateReport(report); err != nil {
		return fmt.Errorf("举报失败: %w", err)
	}

	// 3. 检查举报数，超过阈值自动折叠
	reportCount, _ := s.commentRepo.GetReportCount(req.CommentID)
	if reportCount >= 3 {
		// 自动折叠
		s.commentRepo.UpdateFields(req.CommentID, map[string]interface{}{
			"status": model.CommentStatusFolded,
		})
	}

	return nil
}

// HandleReport 审核举报
func (s *commentV3Service) HandleReport(reportID uint64, adminID, result string, approved bool) error {
	// 1. 获取举报记录
	report, err := s.commentRepo.FindReportByID(reportID)
	if err != nil {
		return fmt.Errorf("举报记录不存在")
	}

	// 2. 更新举报状态
	now := time.Now()
	report.HandleUserID = adminID
	report.HandleResult = result
	report.HandleTime = &now

	if approved {
		report.Status = model.ReportStatusApproved

		// 举报成立，删除评论
		s.commentRepo.SoftDelete(report.CommentID)
	} else {
		report.Status = model.ReportStatusRejected
	}

	if err := s.commentRepo.UpdateReport(report); err != nil {
		return fmt.Errorf("更新举报记录失败: %w", err)
	}

	return nil
}

// GetPendingReports 获取待审核举报
func (s *commentV3Service) GetPendingReports(page, pageSize int) ([]model.CommentReport, int64, error) {
	return s.commentRepo.FindPendingReports(page, pageSize)
}

// ==================== UP主功能实现 ====================

// AddAuthorReply UP主追评
func (s *commentV3Service) AddAuthorReply(commentID uint64, authorID, content string) error {
	// 1. 获取评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 验证是否为UP主
	if !s.isTargetAuthor(comment.TargetType, comment.TargetID, authorID) {
		return fmt.Errorf("只有UP主可以追评")
	}

	// 3. 创建追评
	reply := &model.CommentAuthorReply{
		CommentID:    commentID,
		AuthorUserID: authorID,
		Content:      content,
	}

	if err := s.commentRepo.CreateAuthorReply(reply); err != nil {
		return fmt.Errorf("追评失败: %w", err)
	}

	return nil
}

// PinComment 置顶评论
func (s *commentV3Service) PinComment(commentID uint64, authorID string) error {
	// 1. 获取评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 验证是否为UP主
	if !s.isTargetAuthor(comment.TargetType, comment.TargetID, authorID) {
		return fmt.Errorf("只有UP主可以置顶评论")
	}

	// 3. 置顶
	return s.commentRepo.PinComment(commentID)
}

// UnpinComment 取消置顶
func (s *commentV3Service) UnpinComment(commentID uint64, authorID string) error {
	// 1. 获取评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 验证是否为UP主
	if !s.isTargetAuthor(comment.TargetType, comment.TargetID, authorID) {
		return fmt.Errorf("只有UP主可以取消置顶")
	}

	// 3. 取消置顶
	return s.commentRepo.UnpinComment(commentID)
}

// FeatureComment 精选评论
func (s *commentV3Service) FeatureComment(commentID uint64, authorID string) error {
	// 1. 获取评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 验证是否为UP主
	if !s.isTargetAuthor(comment.TargetType, comment.TargetID, authorID) {
		return fmt.Errorf("只有UP主可以精选评论")
	}

	// 3. 精选
	return s.commentRepo.FeatureComment(commentID)
}

// UnfeatureComment 取消精选
func (s *commentV3Service) UnfeatureComment(commentID uint64, authorID string) error {
	// 1. 获取评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 验证是否为UP主
	if !s.isTargetAuthor(comment.TargetType, comment.TargetID, authorID) {
		return fmt.Errorf("只有UP主可以取消精选")
	}

	// 3. 取消精选
	return s.commentRepo.UnfeatureComment(commentID)
}

// ==================== 热评管理实现 ====================

// CalculateCommentHotScore 计算评论热度分数
func (s *commentV3Service) CalculateCommentHotScore(commentID uint64) (float64, error) {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return 0, err
	}

	// 热度算法（类似知乎）：
	// HotScore = (点赞数 - 踩数)*0.6 + 回复数*0.3 + 时间衰减*0.1

	likeWeight := float64(comment.LikeCount) * 0.6
	dislikeWeight := float64(comment.DislikeCount) * 0.1
	replyWeight := float64(comment.ReplyCount) * 0.3

	baseScore := likeWeight - dislikeWeight + replyWeight

	// 时间衰减（24小时内权重最高）
	hoursSinceCreate := time.Since(comment.CreateTime).Hours()
	timeDecay := math.Exp(-hoursSinceCreate/24.0) * 10.0

	hotScore := baseScore * timeDecay

	// 如果点赞数很高，给予额外加成
	if comment.LikeCount > 100 {
		hotScore *= 1.2
	}

	// 如果是作者评论，给予加成
	if comment.IsAuthor {
		hotScore *= 1.3
	}

	// 更新热度分数
	s.commentRepo.UpdateHotScore(commentID, hotScore)

	return hotScore, nil
}

// UpdateHotCommentList 更新热评榜单
func (s *commentV3Service) UpdateHotCommentList(targetType int8, targetID uint64) error {
	// 1. 获取所有评论
	comments, _, err := s.commentRepo.FindByTarget(targetType, targetID, 1, 1000)
	if err != nil {
		return err
	}

	// 2. 计算每条评论的热度
	for i := range comments {
		s.CalculateCommentHotScore(comments[i].ID)
	}

	// 3. 重新查询，按热度排序
	hotComments, _, err := s.commentRepo.FindRootComments(targetType, targetID, 1, 10, "hot")
	if err != nil {
		return err
	}

	// 4. 更新热评榜单
	return s.commentRepo.BatchUpdateHotList(targetType, targetID, hotComments)
}

// BatchUpdateHotScores 批量更新热度分数
func (s *commentV3Service) BatchUpdateHotScores(targetType int8, targetID uint64) error {
	// 1. 获取所有评论
	comments, _, err := s.commentRepo.FindByTarget(targetType, targetID, 1, 1000)
	if err != nil {
		return err
	}

	// 2. 批量计算热度
	for _, comment := range comments {
		s.CalculateCommentHotScore(comment.ID)
	}

	return nil
}

// ==================== 统计功能实现 ====================

// GetCommentStats 获取评论统计
func (s *commentV3Service) GetCommentStats(targetType int8, targetID uint64) (*model.CommentStats, error) {
	stats, err := s.commentRepo.FindStatsByTarget(targetType, targetID)
	if err != nil {
		// 如果不存在，创建新的统计记录
		stats = &model.CommentStats{
			TargetType: targetType,
			TargetID:   targetID,
		}
		s.commentRepo.CreateStats(stats)
	}

	return stats, nil
}

// GetUserCommentStats 获取用户评论统计
func (s *commentV3Service) GetUserCommentStats(userID string) (*UserCommentStats, error) {
	// 获取用户所有评论
	comments, total, err := s.commentRepo.FindByUserID(userID, 1, 10000)
	if err != nil {
		return nil, err
	}

	stats := &UserCommentStats{
		TotalComments: total,
	}

	var totalLikes int64
	for _, comment := range comments {
		totalLikes += int64(comment.LikeCount)
		stats.TotalReplies += int64(comment.ReplyCount)

		if comment.IsHot {
			stats.HotCommentCount++
		}
		if comment.IsPinned {
			stats.PinnedCount++
		}
		if comment.IsFeatured {
			stats.FeaturedCount++
		}
	}

	stats.TotalLikes = totalLikes

	if total > 0 {
		stats.AvgLikeCount = float64(totalLikes) / float64(total)
	}

	return stats, nil
}

// ==================== 管理功能实现 ====================

// BatchDeleteComments 批量删除评论
func (s *commentV3Service) BatchDeleteComments(commentIDs []uint64, adminID string) error {
	// TODO: 验证管理员权限（需要集成权限系统）
	// 当前简单实现：信任调用者已验证权限
	// 优化方向：集成RBAC权限系统，验证adminID是否有管理员权限
	return s.commentRepo.BatchDelete(commentIDs)
}

// BatchFoldComments 批量折叠评论
func (s *commentV3Service) BatchFoldComments(commentIDs []uint64) error {
	return s.commentRepo.BatchFold(commentIDs)
}

// GetSensitiveWords 获取敏感词列表
func (s *commentV3Service) GetSensitiveWords() ([]model.CommentSensitiveWord, error) {
	return s.commentRepo.FindAllSensitiveWords()
}

// AddSensitiveWord 添加敏感词
func (s *commentV3Service) AddSensitiveWord(word string, level int8, action int8) error {
	sensitiveWord := &model.CommentSensitiveWord{
		Word:      word,
		Level:     level,
		Action:    action,
		IsEnabled: true,
	}

	return s.commentRepo.CreateSensitiveWord(sensitiveWord)
}

// ==================== 辅助方法 ====================

// validateCommentContent 验证评论内容
func (s *commentV3Service) validateCommentContent(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("评论内容不能为空")
	}

	if len([]rune(content)) > 5000 {
		return fmt.Errorf("评论内容不能超过5000字")
	}

	return nil
}

// checkCommentSensitiveWords 检查评论敏感词
func (s *commentV3Service) checkCommentSensitiveWords(content string) error {
	// 1. 获取启用的敏感词列表
	words, err := s.commentRepo.FindEnabledSensitiveWords()
	if err != nil {
		return nil // 如果查询失败，不阻止评论
	}

	// 2. 检查每个敏感词
	for _, word := range words {
		if strings.Contains(content, word.Word) {
			// 根据动作处理
			switch word.Action {
			case model.SensitiveWordActionBlock:
				// 拦截
				return fmt.Errorf("评论包含敏感词")
			case model.SensitiveWordActionReplace:
				// 替换（这里只是检查，实际替换在前端或中间件处理）
				continue
			case model.SensitiveWordActionAudit:
				// 需要人工审核（不阻止发布，但标记为待审核）
				continue
			}
		}
	}

	return nil
}

// isTargetAuthor 检查用户是否为目标对象的作者
func (s *commentV3Service) isTargetAuthor(targetType int8, targetID uint64, userID string) bool {
	switch targetType {
	case model.CommentTargetTypeArticle:
		// 检查是否为文章作者
		return s.articleRepo.CheckOwnership(targetID, userID)
	case model.CommentTargetTypeVideo:
		// TODO: 检查是否为视频作者（需要视频模块支持）
		// 待实现：videoRepo.CheckOwnership(targetID, userID)
		return false
	case model.CommentTargetTypeQA:
		// TODO: 检查是否为问答作者（需要问答模块支持）
		// 待实现：qaRepo.CheckOwnership(targetID, userID)
		return false
	case model.CommentTargetTypeDynamic:
		// TODO: 检查是否为动态作者（需要动态模块支持）
		// 待实现：dynamicRepo.CheckOwnership(targetID, userID)
		return false
	default:
		return false
	}
}

// AuditResult 审核结果
type AuditResult struct {
	Status    int8
	RiskLevel int8
}

// auditCommentContent 审核评论内容
func (s *commentV3Service) auditCommentContent(comment *model.CommentV3) AuditResult {
	result := AuditResult{
		Status:    model.CommentAuditStatusApproved, // 默认通过
		RiskLevel: model.RiskLevelNormal,
	}

	// 1. 长度检查
	contentLength := len([]rune(comment.Content))
	if contentLength < 5 {
		// 内容太短，可能是无意义评论
		result.RiskLevel = model.RiskLevelLow
	}

	// 2. 敏感词检查
	words, _ := s.commentRepo.FindEnabledSensitiveWords()
	hasSensitiveWord := false
	highLevelSensitiveWord := false

	for _, word := range words {
		if strings.Contains(comment.Content, word.Word) {
			hasSensitiveWord = true
			if word.Level >= model.SensitiveWordLevelSerious {
				highLevelSensitiveWord = true
				break
			}
		}
	}

	if highLevelSensitiveWord {
		result.Status = model.CommentAuditStatusRejected
		result.RiskLevel = model.RiskLevelHigh
	} else if hasSensitiveWord {
		result.RiskLevel = model.RiskLevelMedium
	}

	// 3. 重复内容检查
	if s.isRepeatedContent(comment.Content) {
		result.RiskLevel = model.RiskLevelMedium
	}

	// 4. URL链接检查（可能是广告）
	if strings.Contains(comment.Content, "http://") || strings.Contains(comment.Content, "https://") {
		result.RiskLevel = model.RiskLevelMedium
	}

	// 5. 全大写检查（可能是恶意刷屏）
	if s.isAllUpperCase(comment.Content) {
		result.RiskLevel = model.RiskLevelLow
	}

	return result
}

// isRepeatedContent 检查是否为重复内容
func (s *commentV3Service) isRepeatedContent(content string) bool {
	// 简单检查：连续3个相同字符
	runes := []rune(content)
	if len(runes) < 3 {
		return false
	}

	repeatedCount := 0
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			repeatedCount++
			if repeatedCount >= 3 {
				return true
			}
		} else {
			repeatedCount = 0
		}
	}

	return false
}

// isAllUpperCase 检查是否全为大写
func (s *commentV3Service) isAllUpperCase(content string) bool {
	upperCount := 0
	totalLetters := 0

	for _, r := range content {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			totalLetters++
			if r >= 'A' && r <= 'Z' {
				upperCount++
			}
		}
	}

	if totalLetters == 0 {
		return false
	}

	return float64(upperCount)/float64(totalLetters) > 0.8
}

// handleFloorBuilding 处理盖楼
func (s *commentV3Service) handleFloorBuilding(targetType int8, targetID uint64, userID string, commentID uint64) {
	// 1. 查找是否已存在盖楼记录
	building, err := s.commentRepo.FindFloorBuilding(targetType, targetID, userID)

	if err != nil {
		// 不存在，创建新记录
		now := time.Now()
		building = &model.CommentFloorBuilding{
			TargetType:       targetType,
			TargetID:         targetID,
			UserID:           userID,
			CommentIDs:       model.JSONUint64List{commentID},
			FloorCount:       1,
			FirstCommentTime: &now,
			LastCommentTime:  &now,
		}
		s.commentRepo.CreateFloorBuilding(building)
	} else {
		// 已存在，更新记录
		building.CommentIDs = append(building.CommentIDs, commentID)
		building.FloorCount++
		now := time.Now()
		building.LastCommentTime = &now
		s.commentRepo.UpdateFloorBuilding(building)
	}
}
