package model

import "time"

// Article 文章表
type Article struct {
	ID            uint64    `json:"id" gorm:"primaryKey;autoIncrement;comment:'主键'"`
	UserID        string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'用户id'"`
	Title         string    `json:"title" gorm:"size:150;not null;comment:'标题'"`
	Preface       string    `json:"preface" gorm:"size:255;comment:'简介'"`
	Photo         string    `json:"photo" gorm:"size:200;comment:'图片'"`
	Tag           string    `json:"tag" gorm:"size:200;comment:'标签'"`
	Status        int       `json:"status" gorm:"default:1;not null;index:idx_status;comment:'状态：0-草稿 1-已发布 2-已删除'"`
	CreateTime    time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'创建时间'"`
	UpdateTime    time.Time `json:"update_time" gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'"`
	Visit         uint64    `json:"visit" gorm:"not null;comment:'阅读量'"`
	Content       string    `json:"content" gorm:"type:text;comment:'文章内容'"`
	GoodCount     uint64    `json:"good_count" gorm:"not null;comment:'文章点赞量'"`
	Appear        bool      `json:"appear" gorm:"default:1;comment:'是否出现'"`
	Comment       bool      `json:"comment" gorm:"default:1;comment:'是否允许评论'"`
	CommentCount  uint64    `json:"comment_count" gorm:"not null;comment:'文章评论量'"`
	FavoriteCount uint64    `json:"favorite_count" gorm:"default:0;comment:'收藏数量'"`
}

func (a *Article) TableName() string {
	return "article"
}

// 文章状态常量
const (
	ArticleStatusDraft     = 0 // 草稿
	ArticleStatusPublished = 1 // 已发布
	ArticleStatusDeleted   = 2 // 已删除
)
