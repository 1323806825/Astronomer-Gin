package model

import "time"

// UserChat 用户私信表
type UserChat struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement;comment:'主键'"`
	FromUserID  string    `json:"from_user_id" gorm:"type:varchar(36);not null;index:idx_from_user;comment:'发送者ID'"`
	ToUserID    string    `json:"to_user_id" gorm:"type:varchar(36);not null;index:idx_to_user;comment:'接收者ID'"`
	Content     string    `json:"content" gorm:"type:text;not null;comment:'消息内容'"`
	MessageType int       `json:"message_type" gorm:"default:1;comment:'消息类型：1-文本 2-图片 3-语音'"`
	IsRead      bool      `json:"is_read" gorm:"default:false;index:idx_is_read;comment:'是否已读'"`
	CreateTime  time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'发送时间'"`
}

func (c *UserChat) TableName() string {
	return "user_chat"
}

// ChatSession 会话信息（用于返回给前端）
type ChatSession struct {
	UserID          string    `json:"user_id"`           // 对方用户ID
	Username        string    `json:"username"`          // 对方用户名
	Avatar          string    `json:"avatar"`            // 对方头像
	LastMessage     string    `json:"last_message"`      // 最后一条消息
	LastMessageTime time.Time `json:"last_message_time"` // 最后消息时间
	UnreadCount     int64     `json:"unread_count"`      // 未读消息数
}
