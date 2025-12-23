package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"astronomer-gin/model"
	"astronomer-gin/repository"
)

// NotificationHandler 通知任务处理器
type NotificationHandler struct {
	notifyRepo repository.NotificationRepository
	userRepo   repository.UserRepository
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notifyRepo repository.NotificationRepository, userRepo repository.UserRepository) *NotificationHandler {
	return &NotificationHandler{
		notifyRepo: notifyRepo,
		userRepo:   userRepo,
	}
}

// Handle 实现TaskHandler接口
func (h *NotificationHandler) Handle(ctx context.Context, taskType string, data []byte) error {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 只处理通知类型的任务
	if task.Type != "notification" {
		log.Printf("NotificationHandler: skipping non-notification task type: %s", task.Type)
		return nil
	}

	// 根据通知类型分发处理
	notifyType, ok := task.Data["notify_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid notify_type in task data")
	}

	switch notifyType {
	case "follow":
		return h.handleFollowNotification(ctx, task.Data)
	case "like_article":
		return h.handleLikeArticleNotification(ctx, task.Data)
	case "comment":
		return h.handleCommentNotification(ctx, task.Data)
	case "reply_comment":
		return h.handleReplyCommentNotification(ctx, task.Data)
	default:
		log.Printf("Unknown notification type: %s", notifyType)
		return nil
	}
}

// handleFollowNotification 处理关注通知
func (h *NotificationHandler) handleFollowNotification(ctx context.Context, data map[string]interface{}) error {
	userID := data["user_id"].(string)
	followerID := data["follower_id"].(string)
	followerUsername := data["follower_username"].(string)

	log.Printf("Processing follow notification: user %s followed by %s", userID, followerID)

	// 创建通知
	notification := &model.Notification{
		UserID:       userID,
		Type:         model.NotificationTypeFollow,
		FromUserID:   followerID,
		FromUsername: followerUsername,
		Content:      followerUsername + " 关注了你",
		RelatedID:    followerID,
		RelatedType:  "user",
		IsRead:       false,
		CreateTime:   time.Now(),
	}

	if err := h.notifyRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create follow notification: %w", err)
	}

	log.Printf("Follow notification created successfully for user %s", userID)
	return nil
}

// handleLikeArticleNotification 处理文章点赞通知
func (h *NotificationHandler) handleLikeArticleNotification(ctx context.Context, data map[string]interface{}) error {
	authorID := data["author_id"].(string)
	likerID := data["liker_id"].(string)
	likerUsername := data["liker_username"].(string)
	articleID := uint64(data["article_id"].(float64))

	log.Printf("Processing like article notification: article %d liked by %s", articleID, likerID)

	// 不给自己发通知
	if authorID == likerID {
		log.Printf("Skipping self-like notification")
		return nil
	}

	// 创建通知
	notification := &model.Notification{
		UserID:       authorID,
		Type:         model.NotificationTypeLikeArticle,
		FromUserID:   likerID,
		FromUsername: likerUsername,
		Content:      likerUsername + " 赞了你的文章",
		RelatedID:    fmt.Sprintf("%d", articleID),
		RelatedType:  "article",
		IsRead:       false,
		CreateTime:   time.Now(),
	}

	if err := h.notifyRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create like article notification: %w", err)
	}

	log.Printf("Like article notification created successfully for user %s", authorID)
	return nil
}

// handleCommentNotification 处理评论文章通知
func (h *NotificationHandler) handleCommentNotification(ctx context.Context, data map[string]interface{}) error {
	authorID := data["author_id"].(string)
	commenterID := data["commenter_id"].(string)
	commenterUsername := data["commenter_username"].(string)
	articleID := uint64(data["article_id"].(float64))
	commentID := uint64(data["comment_id"].(float64))

	log.Printf("Processing comment notification: article %d commented by %s", articleID, commenterID)

	// 不给自己发通知
	if authorID == commenterID {
		log.Printf("Skipping self-comment notification")
		return nil
	}

	// 创建通知
	notification := &model.Notification{
		UserID:       authorID,
		Type:         model.NotificationTypeComment,
		FromUserID:   commenterID,
		FromUsername: commenterUsername,
		Content:      commenterUsername + " 评论了你的文章",
		RelatedID:    fmt.Sprintf("%d", commentID),
		RelatedType:  "comment",
		IsRead:       false,
		CreateTime:   time.Now(),
	}

	if err := h.notifyRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create comment notification: %w", err)
	}

	log.Printf("Comment notification created successfully for user %s", authorID)
	return nil
}

// handleReplyCommentNotification 处理回复评论通知
func (h *NotificationHandler) handleReplyCommentNotification(ctx context.Context, data map[string]interface{}) error {
	targetUserID := data["target_user_id"].(string)
	replyUserID := data["reply_user_id"].(string)
	replyUsername := data["reply_username"].(string)
	commentID := uint64(data["comment_id"].(float64))

	log.Printf("Processing reply comment notification: comment %d replied by %s", commentID, replyUserID)

	// 不给自己发通知
	if targetUserID == replyUserID {
		log.Printf("Skipping self-reply notification")
		return nil
	}

	// 创建通知
	notification := &model.Notification{
		UserID:       targetUserID,
		Type:         model.NotificationTypeReply,
		FromUserID:   replyUserID,
		FromUsername: replyUsername,
		Content:      replyUsername + " 回复了你的评论",
		RelatedID:    fmt.Sprintf("%d", commentID),
		RelatedType:  "comment",
		IsRead:       false,
		CreateTime:   time.Now(),
	}

	if err := h.notifyRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create reply notification: %w", err)
	}

	log.Printf("Reply comment notification created successfully for user %s", targetUserID)
	return nil
}
