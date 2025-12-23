package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ==================== 评论主表（统一评论表） ====================

// CommentV3 企业级评论模型（统一表，支持无限层级）
type CommentV3 struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	// 所属对象
	TargetType int8   `gorm:"type:tinyint;not null;index:idx_target;comment:'1-文章 2-视频 3-问答 4-动态'" json:"target_type"`
	TargetID   uint64 `gorm:"not null;index:idx_target" json:"target_id"`

	// 用户信息
	UserID     string `gorm:"type:varchar(36);not null;index:idx_user" json:"user_id"`
	Username   string `gorm:"type:varchar(100);comment:'用户名（冗余）'" json:"username,omitempty"`
	UserAvatar string `gorm:"type:varchar(500);comment:'用户头像（冗余）'" json:"user_avatar,omitempty"`

	// 评论结构（核心！）
	ParentID         uint64 `gorm:"default:0;index:idx_parent;comment:'父评论ID（0表示根评论）'" json:"parent_id"`
	RootID           uint64 `gorm:"default:0;index:idx_root;comment:'根评论ID（方便查询整个评论树）'" json:"root_id"`
	ReplyToUserID    string `gorm:"type:varchar(36);default:''" json:"reply_to_user_id"`
	ReplyToCommentID uint64 `gorm:"default:0" json:"reply_to_comment_id"`

	// 楼层信息（关键！）
	FloorNumber    int            `gorm:"default:0;index:idx_floor;comment:'楼层号（1楼、2楼...）'" json:"floor_number"`
	SubFloorNumber int            `gorm:"default:0;comment:'子楼层号（1-1、1-2...）'" json:"sub_floor_number"`
	ReplyChain     JSONUint64List `gorm:"type:varchar(1000);comment:'回复链路（JSON数组：[id1, id2, id3]）'" json:"reply_chain"`
	Depth          int            `gorm:"default:0;comment:'评论深度（0-根评论 1-一级回复 2-二级回复...）'" json:"depth"`

	// 评论内容
	Content     string         `gorm:"type:text;not null" json:"content"`
	ContentType int8           `gorm:"type:tinyint;default:1;comment:'1-文本 2-图片 3-表情包'" json:"content_type"`
	Images      JSONStringList `gorm:"type:varchar(1000);comment:'图片URL（JSON数组）'" json:"images,omitempty"`
	AtUserIDs   JSONUint64List `gorm:"type:varchar(500);comment:'@的用户ID列表（JSON）'" json:"at_user_ids,omitempty"`

	// 评论状态
	Status     int8 `gorm:"type:tinyint;default:1;comment:'1-正常 2-审核中 3-已删除 4-已折叠 5-已屏蔽'" json:"status"`
	IsPinned   bool `gorm:"default:false;comment:'是否置顶（UP主置顶）'" json:"is_pinned"`
	IsAuthor   bool `gorm:"default:false;comment:'是否作者评论'" json:"is_author"`
	IsHot      bool `gorm:"default:false;index:idx_hot;comment:'是否热评'" json:"is_hot"`
	IsFeatured bool `gorm:"default:false;comment:'是否精选评论'" json:"is_featured"`

	// 互动数据
	LikeCount       int `gorm:"default:0" json:"like_count"`
	DislikeCount    int `gorm:"default:0" json:"dislike_count"`
	ReplyCount      int `gorm:"default:0;comment:'直接回复数'" json:"reply_count"`
	TotalReplyCount int `gorm:"default:0;comment:'总回复数（包含子回复）'" json:"total_reply_count"`

	// 热度计算
	HotScore     float64 `gorm:"type:decimal(10,2);default:0;index:idx_hot" json:"hot_score"`
	QualityScore float64 `gorm:"type:decimal(5,2);default:0" json:"quality_score"`

	// IP与设备
	IP         string `gorm:"type:varchar(50)" json:"ip,omitempty"`
	IPLocation string `gorm:"type:varchar(100)" json:"ip_location,omitempty"`
	DeviceType string `gorm:"type:varchar(50)" json:"device_type,omitempty"`
	UserAgent  string `gorm:"type:text" json:"user_agent,omitempty"`

	// 审核相关
	AuditStatus int8   `gorm:"type:tinyint;default:0;comment:'0-待审核 1-通过 2-不通过'" json:"audit_status"`
	AuditReason string `gorm:"type:varchar(200)" json:"audit_reason,omitempty"`
	RiskLevel   int8   `gorm:"type:tinyint;default:0;comment:'0-正常 1-低风险 2-中风险 3-高风险'" json:"risk_level"`

	// 时间戳
	CreateTime time.Time  `gorm:"autoCreateTime;index:idx_target" json:"create_time"`
	UpdateTime time.Time  `gorm:"autoUpdateTime" json:"update_time"`
	DeleteTime *time.Time `json:"delete_time,omitempty"` // 软删除

	// 扩展字段
	ExtInfo JSONMap `gorm:"type:json" json:"ext_info,omitempty"`
}

func (CommentV3) TableName() string {
	return "comment_v3"
}

// ==================== 评论互动表 ====================

// CommentInteraction 评论互动（点赞/踩/举报）
type CommentInteraction struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID  uint64    `gorm:"not null;uniqueIndex:uk_comment_user_action;index:idx_comment" json:"comment_id"`
	UserID     string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_comment_user_action;index:idx_user" json:"user_id"`
	ActionType int8      `gorm:"type:tinyint;not null;uniqueIndex:uk_comment_user_action;comment:'1-点赞 2-踩 3-举报'" json:"action_type"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CommentInteraction) TableName() string {
	return "comment_interaction"
}

// ==================== 评论热榜表 ====================

// CommentHotList 热评榜单（热评缓存）
type CommentHotList struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType int8   `gorm:"type:tinyint;not null;uniqueIndex:uk_target_comment" json:"target_type"`
	TargetID   uint64 `gorm:"not null;uniqueIndex:uk_target_comment;index:idx_target_rank" json:"target_id"`
	CommentID  uint64 `gorm:"not null;uniqueIndex:uk_target_comment" json:"comment_id"`

	// 热评信息（冗余，减少JOIN）
	UserID    string `gorm:"type:varchar(36)" json:"user_id,omitempty"`
	Username  string `gorm:"type:varchar(100)" json:"username,omitempty"`
	Content   string `gorm:"type:text" json:"content,omitempty"`
	LikeCount int    `json:"like_count,omitempty"`

	RankPosition int       `gorm:"index:idx_target_rank;comment:'排��位置'" json:"rank_position"`
	HotScore     float64   `gorm:"type:decimal(10,2)" json:"hot_score"`
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (CommentHotList) TableName() string {
	return "comment_hot_list"
}

// ==================== 评论举报表 ====================

// CommentReport 评论举报
type CommentReport struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID      uint64     `gorm:"not null;index:idx_comment" json:"comment_id"`
	ReporterUserID string     `gorm:"type:varchar(36);not null" json:"reporter_user_id"`
	ReasonType     int8       `gorm:"type:tinyint;comment:'1-垃圾广告 2-色情低俗 3-政治敏感 4-人身攻击 5-造谣传谣'" json:"reason_type"`
	ReasonDesc     string     `gorm:"type:varchar(500)" json:"reason_desc"`
	Status         int8       `gorm:"type:tinyint;default:0;index:idx_status;comment:'0-待处理 1-已处理-成立 2-已处理-不成立'" json:"status"`
	HandleResult   string     `gorm:"type:varchar(200)" json:"handle_result,omitempty"`
	HandleUserID   string     `gorm:"type:varchar(36)" json:"handle_user_id,omitempty"`
	HandleTime     *time.Time `json:"handle_time,omitempty"`
	CreateTime     time.Time  `gorm:"autoCreateTime" json:"create_time"`
}

func (CommentReport) TableName() string {
	return "comment_report"
}

// ==================== 评论盖楼表 ====================

// CommentFloorBuilding 评论盖楼记录（连续评论）
type CommentFloorBuilding struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType       int8           `gorm:"type:tinyint;not null" json:"target_type"`
	TargetID         uint64         `gorm:"not null;index:idx_target_user" json:"target_id"`
	UserID           string         `gorm:"type:varchar(36);not null;index:idx_target_user" json:"user_id"`
	CommentIDs       JSONUint64List `gorm:"type:varchar(1000);comment:'盖楼的评论ID列表（JSON）'" json:"comment_ids"`
	FloorCount       int            `gorm:"default:0" json:"floor_count"`
	FirstCommentTime *time.Time     `json:"first_comment_time"`
	LastCommentTime  *time.Time     `json:"last_comment_time"`
}

func (CommentFloorBuilding) TableName() string {
	return "comment_floor_building"
}

// ==================== UP主追评表 ====================

// CommentAuthorReply UP主追评（作者对评论的补充）
type CommentAuthorReply struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID    uint64    `gorm:"not null;index:idx_comment" json:"comment_id"`
	AuthorUserID string    `gorm:"type:varchar(36);not null" json:"author_user_id"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CommentAuthorReply) TableName() string {
	return "comment_author_reply"
}

// ==================== 评论表情包表 ====================

// CommentEmotion 评论表情包（神评配图）
type CommentEmotion struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"type:varchar(50);not null" json:"name"`
	ImageURL   string    `gorm:"type:varchar(500);not null" json:"image_url"`
	Category   string    `gorm:"type:varchar(50)" json:"category"`
	UseCount   uint64    `gorm:"default:0" json:"use_count"`
	IsHot      bool      `gorm:"default:false" json:"is_hot"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CommentEmotion) TableName() string {
	return "comment_emotion"
}

// ==================== 评论敏感词库 ====================

// CommentSensitiveWord 评论敏感词库（内容审核）
type CommentSensitiveWord struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Word        string    `gorm:"type:varchar(100);not null;unique" json:"word"`
	Level       int8      `gorm:"type:tinyint;default:1;comment:'1-一般 2-严重 3-严重'" json:"level"`
	Action      int8      `gorm:"type:tinyint;default:1;comment:'1-替换 2-拦截 3-人工审核'" json:"action"`
	Replacement string    `gorm:"type:varchar(100)" json:"replacement,omitempty"`
	Category    string    `gorm:"type:varchar(50);comment:'政治/色情/广告'" json:"category"`
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CommentSensitiveWord) TableName() string {
	return "comment_sensitive_word"
}

// ==================== 评论统计表 ====================

// CommentStats 评论统计表（分离热数据）
type CommentStats struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType        int8       `gorm:"type:tinyint;not null;uniqueIndex:uk_target" json:"target_type"`
	TargetID          uint64     `gorm:"not null;uniqueIndex:uk_target" json:"target_id"`
	TotalCommentCount uint64     `gorm:"default:0" json:"total_comment_count"`
	TodayCommentCount int        `gorm:"default:0" json:"today_comment_count"`
	RootCommentCount  int        `gorm:"default:0" json:"root_comment_count"`
	AvgCommentLength  float64    `gorm:"type:decimal(10,2)" json:"avg_comment_length"`
	HotCommentCount   int        `gorm:"default:0" json:"hot_comment_count"`
	LastCommentTime   *time.Time `json:"last_comment_time"`
	UpdateTime        time.Time  `gorm:"autoUpdateTime" json:"update_time"`
}

func (CommentStats) TableName() string {
	return "comment_stats"
}

// ==================== 评论折叠规则表 ====================

// CommentFoldRule 评论折叠规则（智能折叠低质评论）
type CommentFoldRule struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleName   string    `gorm:"type:varchar(100)" json:"rule_name"`
	RuleType   int8      `gorm:"type:tinyint;comment:'1-关键词 2-低赞 3-举报数 4-用户等���'" json:"rule_type"`
	RuleConfig JSONMap   `gorm:"type:json;comment:'规则配置'" json:"rule_config"`
	IsEnabled  bool      `gorm:"default:true" json:"is_enabled"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CommentFoldRule) TableName() string {
	return "comment_fold_rule"
}

// ==================== 自定义JSON类型 ====================

// JSONUint64List JSON uint64数组类型（用于ID列表）
type JSONUint64List []uint64

func (j JSONUint64List) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONUint64List) Scan(value interface{}) error {
	if value == nil {
		*j = []uint64{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ==================== 常量定义 ====================

// 评论对象类型
const (
	CommentTargetTypeArticle = 1 // 文章
	CommentTargetTypeVideo   = 2 // 视频
	CommentTargetTypeQA      = 3 // 问答
	CommentTargetTypeDynamic = 4 // 动态
)

// 评论内容类型
const (
	CommentContentTypeText     = 1 // 文本
	CommentContentTypeImage    = 2 // 图片
	CommentContentTypeEmoticon = 3 // 表情包
)

// 评论状态
const (
	CommentStatusNormal   = 1 // 正常
	CommentStatusAuditing = 2 // 审核中
	CommentStatusDeleted  = 3 // 已删除
	CommentStatusFolded   = 4 // 已折叠
	CommentStatusBlocked  = 5 // 已屏蔽
)

// 审核状态
const (
	CommentAuditStatusPending  = 0 // 待审核
	CommentAuditStatusApproved = 1 // 通过
	CommentAuditStatusRejected = 2 // 不通过
)

// 风险等级
const (
	RiskLevelNormal = 0 // 正常
	RiskLevelLow    = 1 // 低风险
	RiskLevelMedium = 2 // 中风险
	RiskLevelHigh   = 3 // 高风险
)

// 互动类型
const (
	InteractionTypeLike    = 1 // 点赞
	InteractionTypeDislike = 2 // 踩
	InteractionTypeReport  = 3 // 举报
)

// 举报原因类型
const (
	ReportTypeSpam      = 1 // 垃圾广告
	ReportTypePorn      = 2 // 色情低俗
	ReportTypePolitical = 3 // 政治敏感
	ReportTypeAbuse     = 4 // 人身攻击
	ReportTypeFakeNews  = 5 // 造谣传谣
)

// 举报处理状态
const (
	ReportStatusPending  = 0 // 待处理
	ReportStatusApproved = 1 // 已处理-成立
	ReportStatusRejected = 2 // 已处理-不成立
)

// 敏感词等级
const (
	SensitiveWordLevelNormal  = 1 // 一般
	SensitiveWordLevelSerious = 2 // 严重
	SensitiveWordLevelVery    = 3 // 非常严重
)

// 敏感词处理动作
const (
	SensitiveWordActionReplace = 1 // 替换
	SensitiveWordActionBlock   = 2 // 拦截
	SensitiveWordActionAudit   = 3 // 人工审核
)

// 折叠规则类型
const (
	FoldRuleTypeKeyword   = 1 // 关键词
	FoldRuleTypeLowLike   = 2 // 低赞
	FoldRuleTypeReport    = 3 // 举报数
	FoldRuleTypeUserLevel = 4 // 用户等级
)
