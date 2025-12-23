package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"astronomer-gin/repository"
)

// StatsHandler 统计任务处理器
type StatsHandler struct {
	blogRepo repository.BlogRepository
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(blogRepo repository.BlogRepository) *StatsHandler {
	return &StatsHandler{
		blogRepo: blogRepo,
	}
}

// Handle 实现TaskHandler接口
func (h *StatsHandler) Handle(ctx context.Context, taskType string, data []byte) error {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 根据统计类型分发处理
	statsType, ok := task.Data["stats_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid stats_type in task data")
	}

	switch statsType {
	case "view":
		return h.handleViewIncrement(ctx, task.Data)
	case "like":
		return h.handleLikeUpdate(ctx, task.Data)
	default:
		log.Printf("Unknown stats type: %s", statsType)
		return nil
	}
}

// handleViewIncrement 处理浏览量增加
func (h *StatsHandler) handleViewIncrement(ctx context.Context, data map[string]interface{}) error {
	articleID := uint64(data["article_id"].(float64))

	log.Printf("Processing view increment for article %d", articleID)

	// 更新文章浏览量
	if err := h.blogRepo.IncrementVisit(articleID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	log.Printf("View count incremented successfully for article %d", articleID)
	return nil
}

// handleLikeUpdate 处理点赞数更新
func (h *StatsHandler) handleLikeUpdate(ctx context.Context, data map[string]interface{}) error {
	articleID := uint64(data["article_id"].(float64))
	increment := int(data["increment"].(float64)) // 1 或 -1

	log.Printf("Processing like update for article %d (increment: %d)", articleID, increment)

	// 更新点赞数
	if increment > 0 {
		if err := h.blogRepo.IncrementGoodCount(articleID); err != nil {
			return fmt.Errorf("failed to increment like count: %w", err)
		}
	} else {
		if err := h.blogRepo.DecrementGoodCount(articleID); err != nil {
			return fmt.Errorf("failed to decrement like count: %w", err)
		}
	}

	log.Printf("Like count updated successfully for article %d", articleID)
	return nil
}
