package model

import "time"

// UserFollow 用户关注表
type UserFollow struct {
	ID           uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'关注者ID'"`
	FollowUserID string    `json:"follow_user_id" gorm:"type:varchar(36);not null;index:idx_follow_user_id;comment:'被关注者ID'"`
	CreateTime   time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'关注时间'"`
}

func (f *UserFollow) TableName() string {
	return "user_follow"
}

// UserBlock 用户拉黑表
type UserBlock struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'用户ID'"`
	BlockUserID string    `json:"block_user_id" gorm:"type:varchar(36);not null;comment:'被拉黑的用户ID'"`
	CreateTime  time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'拉黑时间'"`
}

func (b *UserBlock) TableName() string {
	return "user_block"
}
