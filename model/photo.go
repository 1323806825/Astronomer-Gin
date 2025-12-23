package model

import "time"

// Photo 图片表
type Photo struct {
	ID        int       `json:"id" gorm:"primary_key;auto_increment"`
	UserID    int       `json:"userID" gorm:"column:user_id"`
	ArticleID int       `json:"articleID" gorm:"column:article_id"`
	Photo     string    `json:"photo" gorm:"type:varchar(500)"`
	Date      time.Time `json:"date"`
	Title     string    `json:"title" gorm:"type:varchar(255)"`
	Position  string    `json:"position" gorm:"type:varchar(500)"`
}

func (p *Photo) TableName() string {
	return "blogphoto"
}
