package model

// CommentParent 一级评论
type CommentParent struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`                        // 评论ID
	ArticleID   int64  `gorm:"column:article_id;type:varchar(20)" json:"article_id"`      // 文章ID
	Comment     string `gorm:"column:comment;type:longtext" json:"comment"`               // 评论内容
	CommentTime string `gorm:"column:comment_time;type:longtext" json:"comment_time"`     // 评论时间
	Username    string `gorm:"column:username;type:varchar(255)" json:"username"`         // 评论人名称
	UserID      string `gorm:"column:user_id;type:varchar(36)" json:"userId"`             // 用户ID
	Phone       string `gorm:"column:phone;type:varchar(255)" json:"phone"`               // 评论人手机号
	GoodCount   uint64 `gorm:"column:good_count;type:bigint unsigned" json:"good_count"`  // 父级评论点赞量
	CommentAddr string `gorm:"column:comment_addr;type:varchar(255)" json:"comment_addr"` // 评论人IP地址
}

func (c *CommentParent) TableName() string {
	return "comment_parent"
}

// CommentSubTwo 二级评论
type CommentSubTwo struct {
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"id"`                                 // 评论ID
	ParentCommentID int64  `gorm:"column:parent_comment_id;type:varchar(20)" json:"parent_comment_id"` // 父级评论ID
	Comment         string `gorm:"column:comment;type:longtext" json:"comment"`                        // 评论内容
	CommentTime     string `gorm:"column:comment_time;type:longtext" json:"comment_time"`              // 评论时间
	Username        string `gorm:"column:username;type:varchar(255)" json:"username"`                  // 评论人名称
	UserID          string `gorm:"column:user_id;type:varchar(36)" json:"userId"`                      // 用户ID
	Phone           string `gorm:"column:phone;type:varchar(255)" json:"phone"`                        // 评论人手机号
	GoodCount       uint64 `gorm:"column:good_count;type:bigint unsigned" json:"good_count"`           // 子级评论点赞量
	ToUsername      string `gorm:"column:to_username;type:varchar(255)" json:"to_username"`            // 被回复人用户名
	ToPhone         string `gorm:"column:to_phone;type:varchar(255)" json:"to_phone"`                  // 被回复人手机号
	ToUserID        string `gorm:"column:to_user_id;type:varchar(36)" json:"toUserId"`                 // 被回复人用户ID
	CommentAddr     string `gorm:"column:comment_addr;type:varchar(255)" json:"comment_addr"`          // 评论人IP地址
}

func (c *CommentSubTwo) TableName() string {
	return "comment_sub_two"
}

// CommentParentLike 一级评论点赞表
type CommentParentLike struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID int64  `gorm:"column:comment_id" json:"comment_id"`
	UserID    string `gorm:"column:user_id;type:varchar(36)" json:"user_id"`
}

func (c *CommentParentLike) TableName() string {
	return "comment_parent_like"
}

// CommentSubTwoLike 二级评论点赞表
type CommentSubTwoLike struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID int64  `gorm:"column:comment_id" json:"comment_id"`
	UserID    string `gorm:"column:user_id;type:varchar(36)" json:"user_id"`
}

func (c *CommentSubTwoLike) TableName() string {
	return "comment_sub_two_like"
}
