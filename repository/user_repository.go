package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByID(id string) (*model.User, error)
	FindByPhone(phone string) (*model.User, error)
	Update(user *model.User) error
	UpdateFields(id string, fields map[string]interface{}) error
	Delete(id string) error
	ExistsByPhone(phone string) bool
	SearchUsers(keyword string, page, pageSize int) ([]model.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByPhone 根据手机号查找用户
func (r *userRepository) FindByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateFields 更新用户指定字段
func (r *userRepository) UpdateFields(id string, fields map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

// ExistsByPhone 检查手机号是否已存在
func (r *userRepository) ExistsByPhone(phone string) bool {
	var count int64
	r.db.Model(&model.User{}).Where("phone = ?", phone).Count(&count)
	return count > 0
}

// SearchUsers 搜索用户（通过用户名或备注）
func (r *userRepository) SearchUsers(keyword string, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 构建搜索条件
	query := r.db.Model(&model.User{})

	if keyword != "" {
		// 搜索用户名或备注包含关键词的用户
		likeKeyword := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR note LIKE ?", likeKeyword, likeKeyword)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按创建时间倒序
	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
