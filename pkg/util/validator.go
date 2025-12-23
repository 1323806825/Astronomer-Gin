package util

import (
	"astronomer-gin/pkg/constant"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ==================== 参数验证工具 ====================

// ValidateParam 通用参数验证
type ValidateParam struct {
	value  interface{}
	errors []*constant.BizError
}

// NewValidate 创建验证器
func NewValidate(value interface{}) *ValidateParam {
	return &ValidateParam{
		value:  value,
		errors: make([]*constant.BizError, 0),
	}
}

// Required 验证必填
func (v *ValidateParam) Required(err *constant.BizError) *ValidateParam {
	if v.value == nil || v.value == "" {
		v.errors = append(v.errors, err)
	}
	return v
}

// MinLength 验证最小长度
func (v *ValidateParam) MinLength(min int, err *constant.BizError) *ValidateParam {
	if str, ok := v.value.(string); ok {
		if utf8.RuneCountInString(str) < min {
			v.errors = append(v.errors, err)
		}
	}
	return v
}

// MaxLength 验证最大长度
func (v *ValidateParam) MaxLength(max int, err *constant.BizError) *ValidateParam {
	if str, ok := v.value.(string); ok {
		if utf8.RuneCountInString(str) > max {
			v.errors = append(v.errors, err)
		}
	}
	return v
}

// Pattern 验证正则表达式
func (v *ValidateParam) Pattern(pattern string, err *constant.BizError) *ValidateParam {
	if str, ok := v.value.(string); ok {
		matched, _ := regexp.MatchString(pattern, str)
		if !matched {
			v.errors = append(v.errors, err)
		}
	}
	return v
}

// GetError 获取第一个错误
func (v *ValidateParam) GetError() *constant.BizError {
	if len(v.errors) > 0 {
		return v.errors[0]
	}
	return nil
}

// IsValid 是否验证通过
func (v *ValidateParam) IsValid() bool {
	return len(v.errors) == 0
}

// ==================== 内容验证工具 ====================

// ValidateTitle 验证标题
func ValidateTitle(title string) *constant.BizError {
	if strings.TrimSpace(title) == "" {
		return constant.ErrTitleRequired
	}
	if utf8.RuneCountInString(title) > constant.MaxTitleLength {
		return constant.ErrTitleTooLong
	}
	return nil
}

// ValidateContent 验证内容
func ValidateContent(content string) *constant.BizError {
	if strings.TrimSpace(content) == "" {
		return constant.ErrContentRequired
	}
	if utf8.RuneCountInString(content) > constant.MaxContentLength {
		return constant.ErrContentTooLong
	}
	return nil
}

// ValidateComment 验证评论
func ValidateComment(comment string) *constant.BizError {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return constant.ErrCommentContentRequired
	}
	length := utf8.RuneCountInString(comment)
	if length < constant.MinCommentLength {
		return constant.ErrCommentTooShort
	}
	if length > constant.MaxCommentLength {
		return constant.ErrCommentTooLong
	}
	return nil
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) *constant.BizError {
	username = strings.TrimSpace(username)
	length := utf8.RuneCountInString(username)

	if length < constant.MinUsernameLength || length > constant.MaxUsernameLength {
		return constant.ErrUsernameInvalid
	}

	// 用户名只能包含中英文、数字、下划线
	matched, _ := regexp.MatchString(`^[\p{Han}a-zA-Z0-9_]+$`, username)
	if !matched {
		return constant.ErrUsernameInvalid
	}

	return nil
}

// ValidatePassword 验证密码
func ValidatePassword(password string) *constant.BizError {
	length := len(password)

	if length < constant.MinPasswordLength || length > constant.MaxPasswordLength {
		return constant.ErrPasswordFormatInvalid
	}

	// 密码必须包含字母和数字
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasDigit {
		return constant.ErrPasswordFormatInvalid
	}

	return nil
}

// ValidatePhone 验证手机号
func ValidatePhone(phone string) *constant.BizError {
	// 简单的中国手机号验证：1开头，11位数字
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	if !matched {
		return constant.ErrPhoneInvalid
	}
	return nil
}

// ==================== 敏感词过滤工具 ====================

// 敏感词列表（实际项目应该从配置文件或数据库加载）
var sensitiveWords = []string{
	"反动", "暴力", "色情", "赌博", "毒品",
	// 实际项目中应该有更完整的敏感词库
}

// 注意：ContainsSensitiveWord 和 FilterSensitiveWord 已移至 sensitive_word_filter.go
// 使用新的 DFA 算法实现，性能更好
// 请使用: util.ContainsSensitiveWord() 和 util.ReplaceSensitiveWord()

// ==================== 数据脱敏工具 ====================

// MaskPhone 手机号脱敏（保留前3后4位）
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// MaskEmail 邮箱脱敏
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	if len(username) <= 2 {
		return "**@" + parts[1]
	}

	return username[:1] + "***" + username[len(username)-1:] + "@" + parts[1]
}

// MaskIDCard 身份证脱敏（保留前6后4位）
func MaskIDCard(idCard string) string {
	if len(idCard) != 18 {
		return idCard
	}
	return idCard[:6] + "********" + idCard[14:]
}
