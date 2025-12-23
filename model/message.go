package model

import "time"

// UserMessage 用户消息表
type UserMessage struct {
	ID              uint      `gorm:"column:id;comment:'主键';primaryKey;not null" json:"id"`
	ReceivedUserID  uint      `gorm:"column:received_user_id;comment:'接收人用户id';default null" json:"receivedUserId"`
	ArticleID       uint      `gorm:"column:article_id;comment:'文章id';default null" json:"articleId"`
	ArticleTitle    string    `gorm:"column:article_title;comment:'文章标题';varchar(150);default null" json:"articleTitle"`
	CommentParentID uint      `gorm:"column:comment_parent_id;comment:'评论id';default null" json:"commentParentId"`
	CommentSubTwoID uint      `gorm:"column:comment_sub_two_id;comment:'评论id';default null" json:"commentSubTwoId"`
	SendUserID      uint      `gorm:"column:send_user_id;comment:'发送人用户id';default null" json:"sendUserId"`
	SendUserName    string    `gorm:"column:send_user_name;varchar(150);comment:'发送人昵称';default null" json:"sendUserName"`
	SendUserIcon    string    `gorm:"column:send_user_icon;varchar(255);comment:'发送人头像';default null" json:"sendUserIcon"`
	MessageType     int64     `gorm:"column:message_type;comment:'消息类型(1->回复我的, 2->赞了文章, 3->赞了评论, 4->系统消息)';default null" json:"messageType"`
	MessageContent  string    `gorm:"column:message_content;comment:'消息内容';varchar(1000);default null" json:"messageContent"`
	CreateTime      time.Time `gorm:"column:create_time;comment:'创建时间';not null" json:"createTime"`
}

func (u *UserMessage) TableName() string {
	return "user_message"
}
