package model

// ArticleStar 文章点赞表
type ArticleStar struct {
	ID        uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    string `gorm:"column:user_id;type:varchar(36);not null;index:idx_user_id" json:"userId"`
	ArticleID uint   `gorm:"column:article_id;not null" json:"articleId"`
}

func (a *ArticleStar) TableName() string {
	return "article_star"
}
