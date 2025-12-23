package model

import "time"

// Notification 通知表
type Notification struct {
	ID           uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'接收者ID'"`
	Type         int       `json:"type" gorm:"not null;comment:'通知类型：1-点赞文章 2-评论文章 3-回复评论 4-关注 5-点赞评论'"`
	FromUserID   string    `json:"from_user_id" gorm:"type:varchar(36);comment:'触发通知的用户ID'"`
	FromUsername string    `json:"from_username" gorm:"type:varchar(100);comment:'触发通知的用户名'"`
	Content      string    `json:"content" gorm:"type:varchar(500);comment:'通知内容'"`
	RelatedID    any       `json:"related_id" gorm:"type:varchar(50);comment:'关联ID（文章ID/评论ID/用户ID）'"`
	RelatedType  string    `json:"related_type" gorm:"type:varchar(50);comment:'关联类型：article/comment/user'"`
	IsRead       bool      `json:"is_read" gorm:"default:0;comment:'是否已读'"`
	CreateTime   time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'创建时间'"`
}

func (n *Notification) TableName() string {
	return "notification"
}

// 通知类型常量
const (
	NotificationTypeLikeArticle = 1 // 点赞文章
	NotificationTypeComment     = 2 // 评论文章
	NotificationTypeReply       = 3 // 回复评论
	NotificationTypeFollow      = 4 // 关注
	NotificationTypeLikeComment = 5 // 点赞评论
)
