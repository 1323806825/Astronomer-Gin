package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// CombinedHandler 组合任务处理器,可以处理多种类型的任务
type CombinedHandler struct {
	notificationHandler *NotificationHandler
	statsHandler        *StatsHandler
}

// NewCombinedHandler 创建组合处理器
func NewCombinedHandler(notificationHandler *NotificationHandler, statsHandler *StatsHandler) *CombinedHandler {
	return &CombinedHandler{
		notificationHandler: notificationHandler,
		statsHandler:        statsHandler,
	}
}

// Handle 实现TaskHandler接口
func (h *CombinedHandler) Handle(ctx context.Context, taskType string, data []byte) error {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	log.Printf("CombinedHandler processing task type: %s", task.Type)

	// 根据任务类型分发给不同的处理器
	switch task.Type {
	case "notification":
		return h.notificationHandler.Handle(ctx, taskType, data)
	case "stats":
		return h.statsHandler.Handle(ctx, taskType, data)
	case "image":
		// 图片处理任务
		return h.handleImageTask(ctx, task)
	default:
		log.Printf("Unknown task type: %s", task.Type)
		return nil
	}
}

// handleImageTask 处理图片任务（示例）
func (h *CombinedHandler) handleImageTask(ctx context.Context, task Task) error {
	log.Printf("Processing image task: %v", task.Data)
	// 这里可以添加图片处理逻辑
	// - 生成缩略图
	// - 图片压缩
	// - 水印添加等
	return nil
}
