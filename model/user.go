package model

import (
	"time"

	"astronomer-gin/pkg/uuid"
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID             string     `json:"id" gorm:"type:varchar(36);primary_key"`
	Phone          string     `json:"phone" gorm:"column:phone;type:varchar(20);unique;not null"`
	Username       string     `json:"username" gorm:"column:username;type:varchar(255);not null"`
	Password       string     `json:"password" gorm:"column:password;type:varchar(255);not null"`
	Icon           string     `json:"icon" gorm:"column:icon;type:varchar(500)"`
	Sex            int        `json:"sex" gorm:"column:sex;comment:'性别(1->男 2->女)';default:1;not null"`
	Note           string     `json:"note" gorm:"column:note;type:varchar(500);comment:'备注'"`
	Intro          string     `json:"intro" gorm:"column:intro;type:varchar(500);comment:'个人简介'"`
	Role           string     `json:"role" gorm:"column:role;type:varchar(20);default:'user';comment:'角色:user-普通用户,admin-管理员,super_admin-超级管理员'"`
	FollowingCount int64      `json:"followingCount" gorm:"column:following_count;comment:'关注数量';default:0;not null"`
	FollowedCount  int64      `json:"followedCount" gorm:"column:followed_count;comment:'被关注数量';default:0;not null"`
	CreateTime     *time.Time `json:"createTime" gorm:"column:create_time;comment:'创建时间';not null"`
}

// BeforeCreate 创建前钩子 - 自动生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) TableName() string {
	return "user"
}
