package constant

// ==================== 通用常量 ====================

// 消息类型常量
const (
	ReplyComment  = 1 // 回复我的
	BlogLike      = 2 // 赞了文章
	CommentLike   = 3 // 赞了评论
	SystemMessage = 4 // 系统消息
	Zero          = 0
)

// 文章状态
const (
	ArticleStatusDraft     = 0 // 草稿
	ArticleStatusPublished = 1 // 已发布
	ArticleStatusDeleted   = 2 // 已删除
)

// 分页常量
const (
	DefaultPage     = 1   // 默认页码
	DefaultPageSize = 10  // 默认每页数量
	MaxPageSize     = 100 // 最大每页数量
)

// 内容长度限制
const (
	MaxTitleLength    = 100   // 标题最大长度
	MaxContentLength  = 50000 // 内容最大长度
	MaxCommentLength  = 500   // 评论最大长度
	MinCommentLength  = 1     // 评论最小长度
	MaxUsernameLength = 50    // 用户名最大长度
	MinUsernameLength = 2     // 用户名最小长度
	MaxPasswordLength = 50    // 密码最大长度
	MinPasswordLength = 6     // 密码最小长度
)

// 文件上传限制
const (
	MaxImageSize        = 5 * 1024 * 1024  // 图片最大5MB
	MaxFileSize         = 20 * 1024 * 1024 // 文件最大20MB
	MaxBatchUploadCount = 10               // 批量上传最多10个文件
)

// 缓存键前缀
const (
	CacheKeyUserInfo      = "user:info:"      // 用户信息缓存
	CacheKeyArticle       = "article:"        // 文章缓存
	CacheKeyArticleList   = "article:list:"   // 文章列表缓存
	CacheKeyUserArticles  = "user:articles:"  // 用户文章列表缓存
	CacheKeyHotArticles   = "hot:articles"    // 热门文章缓存
	CacheKeyCommentList   = "comment:list:"   // 评论列表缓存
	CacheKeyFollow        = "follow:"         // 关注缓存
	CacheKeyFavorite      = "favorite:"       // 收藏缓存
	CacheKeyNotification  = "notification:"   // 通知缓存
	CacheKeyFollowCount   = "follow:count:"   // 关注数缓存
	CacheKeyFavoriteCount = "favorite:count:" // 收藏数缓存
)

// 缓存过期时间（秒）
const (
	CacheExpireShort  = 300   // 5分钟
	CacheExpireMedium = 1800  // 30分钟
	CacheExpireLong   = 3600  // 1小时
	CacheExpireDay    = 86400 // 24小时
)

// 限流相关
const (
	RateLimitLogin    = 5  // 登录限流：5次/分钟
	RateLimitRegister = 3  // 注册限流：3次/小时
	RateLimitComment  = 10 // 评论限流：10次/分钟
	RateLimitArticle  = 5  // 发文限流：5次/小时
	RateLimitFollow   = 20 // 关注限流：20次/分钟
)

// 敏感操作锁定时间（秒）
const (
	LockTimeLogin   = 300 // 登录失败锁定5分钟
	LockTimeComment = 60  // 评论过快锁定1分钟
	LockTimeFollow  = 60  // 关注过快锁定1分钟
)

// ==================== 兼容旧代码的常量（逐步废弃）====================
// 响应消息常量
const (
	Success       = "操作成功"
	Fail          = "操作失败"
	InsertSuccess = "新增成功"
	InsertError   = "新增失败"
	UpdateSuccess = "更新成功"
	UpdateFail    = "更新失败"
	DeleteSuccess = "删除成功"
	DeleteError   = "删除失败"
	QuerySuccess  = "查询成功"
	QueryFail     = "查询失败"
	QueryEmpty    = "查询结果为空"

	LoginSuccess     = "登录成功"
	LoginFail        = "登录失败"
	RegisterSuccess  = "注册成功"
	RegisterFail     = "注册失败"
	PasswordError    = "密码错误"
	UserNotExist     = "用户不存在"
	UserAlreadyExist = "用户已存在"

	TokenInvalid     = "Token无效"
	TokenExpired     = "Token已过期"
	CaptchaError     = "验证码错误"
	ParamError       = "参数错误"
	PermissionDenied = "权限不足"
)

// HTTP状态码
const (
	SuccessRequest = 200
	BadRequest     = 400
	Unauthorized   = 401
	Forbidden      = 403
	NotFound       = 404
	ServerError    = 500
)
