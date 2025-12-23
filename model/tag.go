package model

import "time"

// Tag 标签表
type Tag struct {
	ID         uint      `gorm:"column:id;comment:'主键';primaryKey;not null" json:"id"`
	TagName    string    `gorm:"column:tag_name;varchar(32);comment:'标签名称';not null" json:"tagName"`
	Sort       uint      `gorm:"column:sort;comment:'排序';default null" json:"sort"`
	CreateTime time.Time `gorm:"column:create_time;comment:'创建时间';not null" json:"createTime"`
}

func (t *Tag) TableName() string {
	return "tag"
}
