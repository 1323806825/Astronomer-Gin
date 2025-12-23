package uuid

import (
	"github.com/google/uuid"
)

// New 生成新的UUID (v4)
func New() string {
	return uuid.New().String()
}

// IsValid 验证UUID是否有效
func IsValid(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// Parse 解析UUID字符串
func Parse(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// MustParse 解析UUID，如果失败则panic
func MustParse(id string) uuid.UUID {
	return uuid.MustParse(id)
}

// Nil 返回空UUID
func Nil() string {
	return uuid.Nil.String()
}
