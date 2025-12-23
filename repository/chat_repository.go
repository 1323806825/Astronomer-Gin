package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type ChatRepository interface {
	// 发送私信
	SendMessage(chat *model.UserChat) error

	// 获取与某人的聊天记录
	GetChatHistory(userID, targetUserID string, page, pageSize int) ([]model.UserChat, int64, error)

	// 获取会话列表（所有聊过天的人）
	GetChatSessions(userID string) ([]model.ChatSession, error)

	// 获取未读消息数
	GetUnreadCount(userID string) (int64, error)
	GetUnreadCountWithUser(userID, fromUserID string) (int64, error)

	// 标记消息已读
	MarkAsRead(userID, fromUserID string) error
	MarkMessageAsRead(messageID string) error

	// 删除私信
	DeleteMessage(messageID uint64, userID string) error
	DeleteChatWithUser(userID, targetUserID string) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

// SendMessage 发送私信
func (r *chatRepository) SendMessage(chat *model.UserChat) error {
	return r.db.Create(chat).Error
}

// GetChatHistory 获取与某人的聊天记录
func (r *chatRepository) GetChatHistory(userID, targetUserID string, page, pageSize int) ([]model.UserChat, int64, error) {
	var total int64
	var chats []model.UserChat

	// 查询条件：双方互发的消息
	query := r.db.Model(&model.UserChat{}).Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		userID, targetUserID, targetUserID, userID,
	)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("create_time ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&chats).Error; err != nil {
		return nil, 0, err
	}

	return chats, total, nil
}

// GetChatSessions 获取会话列表
func (r *chatRepository) GetChatSessions(userID string) ([]model.ChatSession, error) {
	// 查询所有与我相关的消息（我发送的和我接收的）
	var sessions []model.ChatSession

	// SQL: 找出所有与我聊过天的人，以及最后一条消息
	sql := `
		SELECT
			CASE
				WHEN from_user_id = ? THEN to_user_id
				ELSE from_user_id
			END as user_id,
			MAX(create_time) as last_message_time,
			(SELECT content FROM user_chat WHERE
				(from_user_id = user_id AND to_user_id = ?) OR
				(from_user_id = ? AND to_user_id = user_id)
				ORDER BY create_time DESC LIMIT 1) as last_message,
			(SELECT COUNT(*) FROM user_chat WHERE
				from_user_id = user_id AND to_user_id = ? AND is_read = false) as unread_count
		FROM user_chat
		WHERE from_user_id = ? OR to_user_id = ?
		GROUP BY user_id
		ORDER BY last_message_time DESC
	`

	if err := r.db.Raw(sql, userID, userID, userID, userID, userID, userID).Scan(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetUnreadCount 获取总未读消息数
func (r *chatRepository) GetUnreadCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserChat{}).
		Where("to_user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

// GetUnreadCountWithUser 获取与某人的未读消息数
func (r *chatRepository) GetUnreadCountWithUser(userID, fromUserID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserChat{}).
		Where("to_user_id = ? AND from_user_id = ? AND is_read = false", userID, fromUserID).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记来自某人的所有消息为已读
func (r *chatRepository) MarkAsRead(userID, fromUserID string) error {
	return r.db.Model(&model.UserChat{}).
		Where("to_user_id = ? AND from_user_id = ? AND is_read = false", userID, fromUserID).
		Update("is_read", true).Error
}

// MarkMessageAsRead 标记单条消息为已读
func (r *chatRepository) MarkMessageAsRead(messageID string) error {
	return r.db.Model(&model.UserChat{}).
		Where("id = ?", messageID).
		Update("is_read", true).Error
}

// DeleteMessage 删除单条私信
func (r *chatRepository) DeleteMessage(messageID uint64, userID string) error {
	// 只能删除自己发送的消息
	return r.db.Where("id = ? AND from_user_id = ?", messageID, userID).
		Delete(&model.UserChat{}).Error
}

// DeleteChatWithUser 删除与某人的所有聊天记录
func (r *chatRepository) DeleteChatWithUser(userID, targetUserID string) error {
	return r.db.Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		userID, targetUserID, targetUserID, userID,
	).Delete(&model.UserChat{}).Error
}
