package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ==================== 文章主表 ====================

// ArticleV3 企业级文章模型
type ArticleV3 struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID string `gorm:"type:varchar(36);not null;index:idx_user_id" json:"user_id"`

	// 基础信息
	Title       string `gorm:"type:varchar(200);not null" json:"title"`
	Summary     string `gorm:"type:varchar(500)" json:"summary"`
	CoverImage  string `gorm:"type:varchar(500)" json:"cover_image"`
	ContentType int8   `gorm:"type:tinyint;default:1;comment:'1-图文 2-视频 3-音频 4-问答'" json:"content_type"`

	// 分类与标签
	CategoryID uint64         `gorm:"default:0;index:idx_category" json:"category_id"`
	ColumnID   uint64         `gorm:"default:0;index:idx_column" json:"column_id"`
	Tags       JSONStringList `gorm:"type:varchar(500)" json:"tags"`   // JSON数组
	Topics     JSONStringList `gorm:"type:varchar(500)" json:"topics"` // JSON数组

	// 状态管理
	Status       int8 `gorm:"type:tinyint;default:1;comment:'1-已发布 2-审核中 3-审核失败 4-已下线 5-已删除'" json:"status"`
	Visibility   int8 `gorm:"type:tinyint;default:1;comment:'1-公开 2-仅粉丝 3-仅好友 4-私密 5-付费'" json:"visibility"`
	AllowComment bool `gorm:"default:true" json:"allow_comment"`
	AllowRepost  bool `gorm:"default:true" json:"allow_repost"`

	// 内容审核
	AuditStatus int8       `gorm:"type:tinyint;default:0;comment:'0-待审核 1-通过 2-驳回'" json:"audit_status"`
	AuditReason string     `gorm:"type:varchar(200)" json:"audit_reason"`
	AuditTime   *time.Time `json:"audit_time"`
	AuditUserID string     `gorm:"type:varchar(36)" json:"audit_user_id"`

	// 统计数据
	ViewCount     uint64 `gorm:"default:0" json:"view_count"`
	RealViewCount uint64 `gorm:"default:0;comment:'真实浏览量（去重）'" json:"real_view_count"`
	LikeCount     uint64 `gorm:"default:0" json:"like_count"`
	CommentCount  uint64 `gorm:"default:0" json:"comment_count"`
	ShareCount    uint64 `gorm:"default:0" json:"share_count"`
	FavoriteCount uint64 `gorm:"default:0" json:"favorite_count"`

	// 推荐与排序
	Weight       int     `gorm:"default:0;comment:'权重（影响排序）'" json:"weight"`
	HotScore     float64 `gorm:"type:decimal(10,2);default:0" json:"hot_score"`
	QualityScore float64 `gorm:"type:decimal(5,2);default:0;comment:'质量分数（AI评分）'" json:"quality_score"`
	IsFeatured   bool    `gorm:"default:false;index:idx_featured" json:"is_featured"`
	IsTop        bool    `gorm:"default:false" json:"is_top"`
	IsHot        bool    `gorm:"default:false;index:idx_hot" json:"is_hot"`
	IsRecommend  bool    `gorm:"default:false" json:"is_recommend"`

	// SEO优化
	Keywords    string `gorm:"type:varchar(200)" json:"keywords"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	Slug        string `gorm:"type:varchar(200);unique;index:idx_slug" json:"slug"`

	// 付费相关
	IsPaid      bool    `gorm:"default:false" json:"is_paid"`
	Price       float64 `gorm:"type:decimal(10,2);default:0" json:"price"`
	FreeContent string  `gorm:"type:text" json:"free_content"`

	// 时间戳
	PublishTime *time.Time `gorm:"index:idx_status_publish" json:"publish_time"`
	CreateTime  time.Time  `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime  time.Time  `gorm:"autoUpdateTime" json:"update_time"`
	DeleteTime  *time.Time `json:"delete_time,omitempty"` // 软删除

	// 扩展字段
	ExtInfo JSONMap `gorm:"type:json" json:"ext_info,omitempty"`
}

// TableName 指定表名
func (ArticleV3) TableName() string {
	return "article_v3"
}

// ==================== 文章内容表（内容分离） ====================

// ArticleContent 文章内容（与主表分离，提升查询性能）
type ArticleContent struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID   uint64    `gorm:"not null;unique;index:idx_article" json:"article_id"`
	Content     string    `gorm:"type:longtext;not null;comment:'Markdown格式'" json:"content"`
	ContentHTML string    `gorm:"type:longtext;comment:'HTML格式'" json:"content_html"`
	TOC         JSONArray `gorm:"type:json;comment:'目录（自动生成）'" json:"toc"`
	WordCount   int       `gorm:"default:0" json:"word_count"`
	ReadTime    int       `gorm:"default:0;comment:'预计阅读时间（分钟）'" json:"read_time"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (ArticleContent) TableName() string {
	return "article_content"
}

// ==================== 文章草稿表 ====================

// ArticleDraft 文章草稿（支持自动保存）
type ArticleDraft struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string `gorm:"type:varchar(36);not null;index:idx_user" json:"user_id"`
	ArticleID uint64 `gorm:"default:0;index:idx_article;comment:'关联的文章ID（0表示新建）'" json:"article_id"`

	// 草稿内容
	Title      string         `gorm:"type:varchar(200)" json:"title"`
	Summary    string         `gorm:"type:varchar(500)" json:"summary"`
	CoverImage string         `gorm:"type:varchar(500)" json:"cover_image"`
	Content    string         `gorm:"type:longtext" json:"content"`
	CategoryID uint64         `json:"category_id"`
	ColumnID   uint64         `json:"column_id"`
	Tags       JSONStringList `gorm:"type:varchar(500)" json:"tags"`
	Topics     JSONStringList `gorm:"type:varchar(500)" json:"topics"`

	// 草稿管理
	AutoSaveCount int        `gorm:"default:0;comment:'自动保存次数'" json:"auto_save_count"`
	LastEditTime  *time.Time `json:"last_edit_time"`
	IsPublished   bool       `gorm:"default:false;index:idx_user" json:"is_published"`

	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (ArticleDraft) TableName() string {
	return "article_draft"
}

// ==================== 文章历史版本表 ====================

// ArticleHistory 文章历史版本（版本控制）
type ArticleHistory struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID uint64 `gorm:"not null;index:idx_article_version" json:"article_id"`
	Version   int    `gorm:"not null;index:idx_article_version" json:"version"`

	// 快照数据
	Title   string `gorm:"type:varchar(200)" json:"title"`
	Content string `gorm:"type:longtext" json:"content"`
	Summary string `gorm:"type:varchar(500)" json:"summary"`

	// 变更信息
	ChangeType   int8   `gorm:"type:tinyint;comment:'1-创建 2-编辑 3-发布 4-下线'" json:"change_type"`
	ChangeReason string `gorm:"type:varchar(200)" json:"change_reason"`
	OperatorID   string `gorm:"type:varchar(36)" json:"operator_id"`

	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (ArticleHistory) TableName() string {
	return "article_history"
}

// ==================== 文章分类表 ====================

// ArticleCategory 文章分类（支持多级分类）
type ArticleCategory struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(50);not null" json:"name"`
	ParentID     uint64    `gorm:"default:0;index:idx_parent" json:"parent_id"`
	Icon         string    `gorm:"type:varchar(200)" json:"icon"`
	SortOrder    int       `gorm:"default:0;index:idx_parent" json:"sort_order"`
	ArticleCount int       `gorm:"default:0" json:"article_count"`
	IsShow       bool      `gorm:"default:true" json:"is_show"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (ArticleCategory) TableName() string {
	return "article_category"
}

// ==================== 专栏表 ====================

// ArticleColumn 专栏（系列文章）
type ArticleColumn struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          string    `gorm:"type:varchar(36);not null;index:idx_user" json:"user_id"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name"`
	Description     string    `gorm:"type:varchar(500)" json:"description"`
	CoverImage      string    `gorm:"type:varchar(500)" json:"cover_image"`
	ArticleCount    int       `gorm:"default:0" json:"article_count"`
	SubscriberCount int       `gorm:"default:0" json:"subscriber_count"`
	IsFinished      bool      `gorm:"default:false" json:"is_finished"`
	SortType        int8      `gorm:"type:tinyint;default:1;comment:'1-自定义 2-时间正序 3-时间倒序'" json:"sort_type"`
	Status          int8      `gorm:"type:tinyint;default:1;comment:'1-正常 2-隐藏'" json:"status"`
	CreateTime      time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime      time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (ArticleColumn) TableName() string {
	return "article_column"
}

// ==================== 专栏-文章关联表 ====================

// ArticleColumnRel 专栏文章关联
type ArticleColumnRel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ColumnID  uint64    `gorm:"not null;uniqueIndex:uk_column_article;index:idx_column_sort" json:"column_id"`
	ArticleID uint64    `gorm:"not null;uniqueIndex:uk_column_article" json:"article_id"`
	SortOrder int       `gorm:"default:0;index:idx_column_sort" json:"sort_order"`
	AddTime   time.Time `gorm:"autoCreateTime" json:"add_time"`
}

func (ArticleColumnRel) TableName() string {
	return "article_column_rel"
}

// ==================== 话题表 ====================

// Topic 话题（类似微博话题）
type Topic struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null;unique;index:idx_name" json:"name"`
	Description  string    `gorm:"type:varchar(500)" json:"description"`
	CoverImage   string    `gorm:"type:varchar(500)" json:"cover_image"`
	ArticleCount uint64    `gorm:"default:0" json:"article_count"`
	FollowCount  uint64    `gorm:"default:0" json:"follow_count"`
	ViewCount    uint64    `gorm:"default:0" json:"view_count"`
	IsHot        bool      `gorm:"default:false;index:idx_hot" json:"is_hot"`
	IsRecommend  bool      `gorm:"default:false" json:"is_recommend"`
	Category     string    `gorm:"type:varchar(50)" json:"category"`
	CreatorID    string    `gorm:"type:varchar(36)" json:"creator_id"`
	Status       int8      `gorm:"type:tinyint;default:1;comment:'1-正常 2-隐藏 3-封禁'" json:"status"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (Topic) TableName() string {
	return "topic"
}

// ==================== 文章-话题关联表 ====================

// ArticleTopicRel 文章话题关联
type ArticleTopicRel struct {
	ArticleID  uint64    `gorm:"not null;primaryKey;index:idx_topic" json:"article_id"`
	TopicID    uint64    `gorm:"not null;primaryKey;index:idx_topic" json:"topic_id"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (ArticleTopicRel) TableName() string {
	return "article_topic_rel"
}

// ==================== 文章标签表 ====================

// ArticleTag 文章标签
type ArticleTag struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(50);not null;unique" json:"name"`
	ArticleCount int       `gorm:"default:0" json:"article_count"`
	FollowCount  int       `gorm:"default:0" json:"follow_count"`
	IsHot        bool      `gorm:"default:false" json:"is_hot"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (ArticleTag) TableName() string {
	return "article_tag"
}

// ==================== 文章相关推荐关联表 ====================

// ArticleRelation 文章相关推荐
type ArticleRelation struct {
	ArticleID        uint64    `gorm:"not null;primaryKey;index:idx_article_score" json:"article_id"`
	RelatedArticleID uint64    `gorm:"not null;primaryKey" json:"related_article_id"`
	RelevanceScore   float64   `gorm:"type:decimal(5,2);default:0;index:idx_article_score" json:"relevance_score"`
	RelationType     int8      `gorm:"type:tinyint;comment:'1-同作者 2-同分类 3-同话题 4-算法推荐'" json:"relation_type"`
	CreateTime       time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (ArticleRelation) TableName() string {
	return "article_relation"
}

// ==================== 文章统计详情表 ====================

// ArticleStatsDetail 文章统计详情（热数据分离）
type ArticleStatsDetail struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID         uint64    `gorm:"not null;unique" json:"article_id"`
	TodayViewCount    int       `gorm:"default:0" json:"today_view_count"`
	WeekViewCount     int       `gorm:"default:0" json:"week_view_count"`
	MonthViewCount    int       `gorm:"default:0" json:"month_view_count"`
	TodayLikeCount    int       `gorm:"default:0" json:"today_like_count"`
	TodayCommentCount int       `gorm:"default:0" json:"today_comment_count"`
	TodayShareCount   int       `gorm:"default:0" json:"today_share_count"`
	SourceStats       JSONMap   `gorm:"type:json;comment:'流量来源统计'" json:"source_stats,omitempty"`
	DeviceStats       JSONMap   `gorm:"type:json;comment:'设备统计'" json:"device_stats,omitempty"`
	RegionStats       JSONMap   `gorm:"type:json;comment:'地域统计'" json:"region_stats,omitempty"`
	AvgReadProgress   float64   `gorm:"type:decimal(5,2);default:0;comment:'平均阅读进度（%）'" json:"avg_read_progress"`
	AvgStayTime       int       `gorm:"default:0;comment:'平均停留时间（秒）'" json:"avg_stay_time"`
	UpdateTime        time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (ArticleStatsDetail) TableName() string {
	return "article_stats_detail"
}

// ==================== 自定义JSON类型（便于处理） ====================

// JSONStringList JSON字符串数组类型
type JSONStringList []string

func (j JSONStringList) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONStringList) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// JSONArray JSON数组类型
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = []interface{}{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// JSONMap JSON对象类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = map[string]interface{}{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ==================== 常量定义 ====================

// 文章内容类型
const (
	ArticleContentTypeText  = 1 // 图文
	ArticleContentTypeVideo = 2 // 视频
	ArticleContentTypeAudio = 3 // 音频
	ArticleContentTypeQA    = 4 // 问答
)

// 文章状态（V3版本）
const (
	ArticleV3StatusPublished   = 1 // 已发布
	ArticleV3StatusAuditing    = 2 // 审核中
	ArticleV3StatusAuditFailed = 3 // 审核失败
	ArticleV3StatusOffline     = 4 // 已下线
	ArticleV3StatusDeleted     = 5 // 已删除
)

// 文章可见性
const (
	ArticleVisibilityPublic   = 1 // 公开
	ArticleVisibilityFollower = 2 // 仅粉丝
	ArticleVisibilityFriend   = 3 // 仅好友
	ArticleVisibilityPrivate  = 4 // 私密
	ArticleVisibilityPaid     = 5 // 付费
)

// 审核状态
const (
	AuditStatusPending  = 0 // 待审核
	AuditStatusApproved = 1 // 通过
	AuditStatusRejected = 2 // 驳回
)

// 变更类型
const (
	ChangeTypeCreate  = 1 // 创建
	ChangeTypeEdit    = 2 // 编辑
	ChangeTypePublish = 3 // 发布
	ChangeTypeOffline = 4 // 下线
)

// 专栏排序方式
const (
	ColumnSortCustom   = 1 // 自定义
	ColumnSortTimeAsc  = 2 // 时间正序
	ColumnSortTimeDesc = 3 // 时间倒序
)

// 关联类型
const (
	RelationTypeSameAuthor   = 1 // 同作者
	RelationTypeSameCategory = 2 // 同分类
	RelationTypeSameTopic    = 3 // 同话题
	RelationTypeAlgorithm    = 4 // 算法推荐
)

// ==================== V3 专用关联表 ====================

// ColumnSubscription 专栏订阅表
type ColumnSubscription struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	ColumnID   uint64    `json:"column_id" gorm:"not null;index:idx_column_id;comment:'专栏ID'"`
	UserID     string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'用户ID'"`
	CreateTime time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'订阅时间'"`
}

func (c *ColumnSubscription) TableName() string {
	return "column_subscription"
}

// TopicFollow 话题关注表
type TopicFollow struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TopicID    uint64    `json:"topic_id" gorm:"not null;index:idx_topic_id;comment:'话题ID'"`
	UserID     string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_id;comment:'用户ID'"`
	CreateTime time.Time `json:"create_time" gorm:"default:CURRENT_TIMESTAMP;comment:'关注时间'"`
}

func (t *TopicFollow) TableName() string {
	return "topic_follow"
}
