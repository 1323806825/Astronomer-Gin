package model

import "time"

// UserFavorite 用户收藏表
type UserFavorite struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'用户ID'"`
	ArticleID  uint64    `json:"article_id" gorm:"not null;comment:'文章ID'"`
	CreateTime time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'收藏时间'"`
}

func (f *UserFavorite) TableName() string {
	return "user_favorite"
}
