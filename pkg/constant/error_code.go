package constant

// ==================== 错误码体系设计 ====================
// 错误码规则：XXYYZ
// XX: 模块码（10=用户, 20=博客, 30=评论, 40=关注, 50=收藏, 60=通知, 70=系统）
// YY: 错误类别（00=成功, 01=参数, 02=业务, 03=权限, 04=资源, 05=状态）
// Z: 具体错误序号（0-9）

// BizError 业务错误结构
type BizError struct {
	Code      int    // 错误码
	Message   string // 中文描述
	EnMessage string // 英文描述（国际化）
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return e.Message
}

// NewBizError 创建业务错误
func NewBizError(code int, message, enMessage string) *BizError {
	return &BizError{
		Code:      code,
		Message:   message,
		EnMessage: enMessage,
	}
}

// ==================== 通用错误码 (00xxx) ====================

var (
	// 成功
	ErrSuccess = NewBizError(200, "操作成功", "Success")

	// 参数错误 (001xx)
	ErrParamInvalid       = NewBizError(10, "参数错误", "Invalid parameter")
	ErrParamMissing       = NewBizError(11, "缺少必要参数", "Missing required parameter")
	ErrParamFormatInvalid = NewBizError(12, "参数格式错误", "Invalid parameter format")
	ErrValidationFailed   = NewBizError(13, "参数校验失败", "Validation failed")

	// 系统错误 (002xx)
	ErrSystemError        = NewBizError(20, "系统错误", "System error")
	ErrDatabaseError      = NewBizError(21, "数据库错误", "Database error")
	ErrCacheError         = NewBizError(22, "缓存错误", "Cache error")
	ErrNetworkError       = NewBizError(23, "网络错误", "Network error")
	ErrServiceUnavailable = NewBizError(24, "服务不可用", "Service unavailable")

	// 权限错误 (003xx)
	ErrUnauthorized     = NewBizError(30, "未授权", "Unauthorized")
	ErrTokenInvalid     = NewBizError(31, "Token无效", "Invalid token")
	ErrTokenExpired     = NewBizError(32, "Token已过期", "Token expired")
	ErrPermissionDenied = NewBizError(33, "权限不足", "Permission denied")
	ErrForbidden        = NewBizError(34, "禁止访问", "Forbidden")

	// 资源错误 (004xx)
	ErrResourceNotFound = NewBizError(40, "资源不存在", "Resource not found")
	ErrResourceConflict = NewBizError(41, "资源冲突", "Resource conflict")
	ErrResourceDeleted  = NewBizError(42, "资源已删除", "Resource deleted")

	// 限流错误 (005xx)
	ErrTooManyRequests = NewBizError(50, "请求过于频繁", "Too many requests")
	ErrRateLimitExceed = NewBizError(51, "超出访问频率限制", "Rate limit exceeded")
)

// ==================== 用户模块错误码 (10xxx) ====================

var (
	// 注册相关 (100xx)
	ErrUserAlreadyExists  = NewBizError(10001, "用户已存在", "User already exists")
	ErrPhoneRegistered    = NewBizError(10002, "手机号已注册", "Phone number already registered")
	ErrUsernameRegistered = NewBizError(10003, "用户名已注册", "Username already registered")
	ErrEmailRegistered    = NewBizError(10004, "邮箱已注册", "Email already registered")
	ErrRegisterFailed     = NewBizError(10005, "注册失败", "Registration failed")
	ErrCaptchaInvalid     = NewBizError(10006, "验证码错误", "Invalid captcha")
	ErrCaptchaExpired     = NewBizError(10007, "验证码已过期", "Captcha expired")

	// 登录相关 (101xx)
	ErrUserNotExist      = NewBizError(10101, "用户不存在", "User not found")
	ErrPasswordIncorrect = NewBizError(10102, "密码错误", "Incorrect password")
	ErrLoginFailed       = NewBizError(10103, "登录失败", "Login failed")
	ErrAccountDisabled   = NewBizError(10104, "账号已被禁用", "Account disabled")
	ErrAccountLocked     = NewBizError(10105, "账号已被锁定", "Account locked")

	// 账号操作 (102xx)
	ErrPasswordFormatInvalid = NewBizError(10201, "密码格式不正确", "Invalid password format")
	ErrOldPasswordIncorrect  = NewBizError(10202, "原密码错误", "Old password incorrect")
	ErrUpdateUserFailed      = NewBizError(10203, "更新用户信息失败", "Update user failed")
	ErrUsernameInvalid       = NewBizError(10204, "用户名格式不正确", "Invalid username format")
	ErrPhoneInvalid          = NewBizError(10205, "手机号格式不正确", "Invalid phone format")
)

// ==================== 博��模块错误码 (20xxx) ====================

var (
	// 文章操作 (200xx)
	ErrArticleNotFound        = NewBizError(20001, "文章不存在", "Article not found")
	ErrCreateArticleFailed    = NewBizError(20002, "创建文章失败", "Create article failed")
	ErrUpdateArticleFailed    = NewBizError(20003, "更新文章失败", "Update article failed")
	ErrDeleteArticleFailed    = NewBizError(20004, "删除文章失败", "Delete article failed")
	ErrArticleDeleted         = NewBizError(20005, "文章已删除", "Article deleted")
	ErrArticleCommentDisabled = NewBizError(20006, "文章不允许评论", "Article comment disabled")

	// 文章内容 (201xx)
	ErrTitleRequired           = NewBizError(20101, "标题不能为空", "Title is required")
	ErrTitleTooLong            = NewBizError(20102, "标题过长", "Title too long")
	ErrContentRequired         = NewBizError(20103, "内容不能为空", "Content is required")
	ErrContentTooLong          = NewBizError(20104, "内容过长", "Content too long")
	ErrContentHasSensitiveWord = NewBizError(20105, "内容包含敏感词", "Content contains sensitive words")

	// 文章权限 (202xx)
	ErrNotArticleOwner     = NewBizError(20201, "无权操作此文章", "Not article owner")
	ErrArticleNotPublished = NewBizError(20202, "文章未发布", "Article not published")

	// 点赞操作 (203xx)
	ErrLikeFailed   = NewBizError(20301, "点赞失败", "Like failed")
	ErrAlreadyLiked = NewBizError(20302, "已经点赞过了", "Already liked")
	ErrNotLiked     = NewBizError(20303, "还未点赞", "Not liked yet")

	// 草稿操作 (204xx)
	ErrDraftNotFound      = NewBizError(20401, "草稿不存在", "Draft not found")
	ErrSaveDraftFailed    = NewBizError(20402, "保存草稿失败", "Save draft failed")
	ErrPublishDraftFailed = NewBizError(20403, "发布草稿失败", "Publish draft failed")
)

// ==================== 评论模块错误码 (30xxx) ====================

var (
	// 评论操作 (300xx)
	ErrCommentNotFound     = NewBizError(30001, "评论不存在", "Comment not found")
	ErrCreateCommentFailed = NewBizError(30002, "发表评论失败", "Create comment failed")
	ErrDeleteCommentFailed = NewBizError(30003, "删除评论失败", "Delete comment failed")
	ErrCommentDeleted      = NewBizError(30004, "评论已删除", "Comment deleted")

	// 评论内容 (301xx)
	ErrCommentContentRequired = NewBizError(30101, "评论内容不能为空", "Comment content is required")
	ErrCommentTooLong         = NewBizError(30102, "评论内容过长", "Comment too long")
	ErrCommentTooShort        = NewBizError(30103, "评论内容过短", "Comment too short")

	// 评论权限 (302xx)
	ErrNotCommentOwner        = NewBizError(30201, "无权操作此评论", "Not comment owner")
	ErrArticleNotAllowComment = NewBizError(30202, "文章不允许评论", "Article does not allow comments")

	// 评论点赞 (303xx)
	ErrLikeCommentFailed = NewBizError(30301, "点赞评论失败", "Like comment failed")
)

// ==================== 关注模块错误码 (40xxx) ====================

var (
	// 关注操作 (400xx)
	ErrCannotFollowSelf    = NewBizError(40001, "不能关注自己", "Cannot follow yourself")
	ErrAlreadyFollowed     = NewBizError(40002, "已关注该用户", "Already followed")
	ErrNotFollowed         = NewBizError(40003, "未关注该用户", "Not followed")
	ErrFollowFailed        = NewBizError(40004, "关注失败", "Follow failed")
	ErrUnfollowFailed      = NewBizError(40005, "取消关注失败", "Unfollow failed")
	ErrCannotFollowBlocked = NewBizError(40006, "无法关注该用户", "Cannot follow this user")
	ErrBlocked             = NewBizError(40007, "该用户已拉黑你", "You have been blocked by this user")

	// 拉黑操作 (401xx)
	ErrCannotBlockSelf = NewBizError(40101, "不能拉黑自己", "Cannot block yourself")
	ErrAlreadyBlocked  = NewBizError(40102, "已拉黑该用户", "Already blocked")
	ErrNotBlocked      = NewBizError(40103, "未拉黑该用户", "Not blocked")
	ErrBlockFailed     = NewBizError(40104, "拉黑失败", "Block failed")
	ErrUnblockFailed   = NewBizError(40105, "取消拉黑失败", "Unblock failed")
)

// ==================== 收藏模块错误码 (50xxx) ====================

var (
	// 收藏操作 (500xx)
	ErrAlreadyFavorited    = NewBizError(50001, "已收藏该文章", "Already favorited")
	ErrNotFavorited        = NewBizError(50002, "未收藏该文章", "Not favorited")
	ErrFavoriteFailed      = NewBizError(50003, "收藏失败", "Favorite failed")
	ErrUnfavoriteFailed    = NewBizError(50004, "取消收藏失败", "Unfavorite failed")
	ErrCannotFavoriteDraft = NewBizError(50005, "不能收藏草稿", "Cannot favorite draft")
)

// ==================== 通知模块错误码 (60xxx) ====================

var (
	// 通知操作 (600xx)
	ErrNotificationNotFound     = NewBizError(60001, "通知不存在", "Notification not found")
	ErrMarkReadFailed           = NewBizError(60002, "标记已读失败", "Mark as read failed")
	ErrDeleteNotificationFailed = NewBizError(60003, "删除通知失败", "Delete notification failed")
	ErrNotNotificationOwner     = NewBizError(60004, "无权操作此通知", "Not notification owner")
)

// ==================== 系统模块错误码 (70xxx) ====================

var (
	// 数据库操作 (700xx)
	ErrDatabaseQuery  = NewBizError(70001, "数据库查询失败", "Database query failed")
	ErrDatabaseInsert = NewBizError(70002, "数据库插入失败", "Database insert failed")
	ErrDatabaseUpdate = NewBizError(70003, "数据库更新失败", "Database update failed")
	ErrDatabaseDelete = NewBizError(70004, "数据库删除失败", "Database delete failed")
	ErrUpdateFail     = NewBizError(70005, "更新失败", "Update failed")
	ErrDeleteFail     = NewBizError(70006, "删除失败", "Delete failed")

	// 缓存操作 (701xx)
	ErrCacheGet = NewBizError(70101, "缓存获取失败", "Cache get failed")
	ErrCacheSet = NewBizError(70102, "缓存设置失败", "Cache set failed")
	ErrCacheDel = NewBizError(70103, "缓存删除失败", "Cache delete failed")

	// 文件操作 (702xx)
	ErrFileUploadFailed   = NewBizError(70201, "文件上传失败", "File upload failed")
	ErrFileDeleteFailed   = NewBizError(70202, "文件删除失败", "File delete failed")
	ErrFileFormatInvalid  = NewBizError(70203, "文件格式不正确", "Invalid file format")
	ErrFileSizeExceeded   = NewBizError(70204, "文件大小超出限制", "File size exceeded")
	ErrInvalidImageFormat = NewBizError(70205, "不支持的图片格式", "Invalid image format")
	ErrImageSizeExceeded  = NewBizError(70206, "图片大小超出限制", "Image size exceeded")
	ErrUploadFailed       = NewBizError(70207, "上传失败", "Upload failed")
	ErrInvalidFileURL     = NewBizError(70208, "无效的文件URL", "Invalid file URL")
	ErrDeleteFileFailed   = NewBizError(70209, "删除文件失败", "Delete file failed")
	ErrNoFilesProvided    = NewBizError(70210, "未提供文件", "No files provided")
	ErrTooManyFiles       = NewBizError(70211, "文件数量超出限制", "Too many files")
)
