package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/email"
	"astronomer-gin/pkg/jwt"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserServiceV2 企业级用户服务接口
type UserServiceV2 interface {
	Register(phone, password, username string) error
	Login(phone, password string) (string, *model.User, error)
	GetUserInfo(phone string) (*model.User, error)
	GetUserInfoByID(userID string) (*model.User, error)
	UpdateUserInfo(phone string, updates map[string]interface{}) error
	ChangePassword(phone, oldPassword, newPassword string) error

	// 缓存管理
	RefreshUserCache(phone string) error
	ClearUserCache(phone string) error
}

type userServiceV2 struct {
	userRepo    repository.UserRepository
	cacheHelper *util.CacheHelper
}

func NewUserServiceV2(userRepo repository.UserRepository) UserServiceV2 {
	return &userServiceV2{
		userRepo:    userRepo,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
	}
}

// Register 用户注册（企业级实现）
func (s *userServiceV2) Register(phone, password, username string) error {
	// 1. 参数验证
	if err := util.ValidatePhone(phone); err != nil {
		return err
	}
	if err := util.ValidatePassword(password); err != nil {
		return err
	}
	if err := util.ValidateUsername(username); err != nil {
		return err
	}

	// 2. 敏感词检查
	if util.ContainsSensitiveWord(username) {
		return constant.ErrUsernameInvalid
	}

	// 3. 检查手机号是否已存在
	if s.userRepo.ExistsByPhone(phone) {
		return constant.ErrPhoneRegistered
	}

	// 4. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return constant.ErrSystemError
	}

	// 5. 创建用户
	now := time.Now()
	user := &model.User{
		Phone:      phone,
		Password:   string(hashedPassword),
		Username:   username,
		CreateTime: &now,
	}

	if err := s.userRepo.Create(user); err != nil {
		return constant.ErrRegisterFailed
	}

	// 6. 异步发送欢迎邮件
	go func() {
		// 使用phone作为email（如果有单独的email字段，使用user.Email）
		if err := email.SendWelcomeEmail(phone, username); err != nil {
			// 记录错误但不影响注册流程
			fmt.Printf("发送欢迎邮件失败: %v\n", err)
		}
	}()

	return nil
}

// Login 用户登录（企业级实现）
func (s *userServiceV2) Login(phone, password string) (string, *model.User, error) {
	// 1. 参数验证
	if err := util.ValidatePhone(phone); err != nil {
		return "", nil, err
	}

	// 2. 检查登录失败锁定
	lockKey := fmt.Sprintf("login:lock:%s", phone)
	if s.cacheHelper.Exists(lockKey) {
		return "", nil, constant.ErrAccountLocked
	}

	// 3. 查找用户
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		// 记录登录失败次数
		s.recordLoginFail(phone)
		return "", nil, constant.ErrUserNotExist
	}

	// 4. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 记录登录失败次数
		s.recordLoginFail(phone)
		return "", nil, constant.ErrPasswordIncorrect
	}

	// 5. 清除登录失败记录
	failKey := fmt.Sprintf("login:fail:%s", phone)
	s.cacheHelper.Delete(failKey)

	// 6. 生成token
	token, err := jwt.Sign(user.Phone, user.ID)
	if err != nil {
		return "", nil, constant.ErrSystemError
	}

	// 7. 数据脱敏
	userCopy := *user
	userCopy.Password = "" // 不返回密码
	userCopy.Phone = util.MaskPhone(user.Phone)

	// 8. 更新缓存
	s.setUserCache(user)

	return token, &userCopy, nil
}

// GetUserInfo 获取用户信息（带缓存）
func (s *userServiceV2) GetUserInfo(phone string) (*model.User, error) {
	// 1. 先从缓存获取
	cacheKey := constant.CacheKeyUserInfo + phone
	var user model.User

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&user,
		time.Duration(constant.CacheExpireMedium)*time.Second,
		func() (interface{}, error) {
			// 缓存未命中，从数据库查询
			return s.userRepo.FindByPhone(phone)
		},
	)

	if err != nil {
		return nil, constant.ErrUserNotExist
	}

	// 2. 数据脱敏
	user.Password = ""

	return &user, nil
}

// GetUserInfoByID 根据ID获取用户信息（带缓存）
func (s *userServiceV2) GetUserInfoByID(userID string) (*model.User, error) {
	// 1. 先从缓存获取
	cacheKey := fmt.Sprintf("%s%s", constant.CacheKeyUserInfo, userID)
	var user model.User

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&user,
		time.Duration(constant.CacheExpireMedium)*time.Second,
		func() (interface{}, error) {
			// 缓存未命中，从数据库查询
			return s.userRepo.FindByID(userID)
		},
	)

	if err != nil {
		return nil, constant.ErrUserNotExist
	}

	// 2. 数据脱敏
	user.Password = ""

	return &user, nil
}

// UpdateUserInfo 更新用户信息（企业级实现）
func (s *userServiceV2) UpdateUserInfo(phone string, updates map[string]interface{}) error {
	// 1. 查找用户
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		return constant.ErrUserNotExist
	}

	// 2. 验证更新内容
	if username, ok := updates["username"].(string); ok && username != "" {
		if err := util.ValidateUsername(username); err != nil {
			return err
		}
		if util.ContainsSensitiveWord(username) {
			return constant.ErrUsernameInvalid
		}
	}

	// 3. 禁止直接更新密码（应使用ChangePassword）
	delete(updates, "password")

	// 4. 更新用户信息
	if err := s.userRepo.UpdateFields(user.ID, updates); err != nil {
		return constant.ErrUpdateUserFailed
	}

	// 5. 清除所有相关缓存
	s.ClearUserCache(phone)                                                       // 清除基于 phone 的缓存
	s.cacheHelper.Delete(fmt.Sprintf("%s%s", constant.CacheKeyUserInfo, user.ID)) // 清除基于 userID 的缓存

	return nil
}

// ChangePassword 修改密码（企业级实现）
func (s *userServiceV2) ChangePassword(phone, oldPassword, newPassword string) error {
	// 1. 参数验证
	if err := util.ValidatePassword(newPassword); err != nil {
		return err
	}

	// 2. 查找用户
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		return constant.ErrUserNotExist
	}

	// 3. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return constant.ErrOldPasswordIncorrect
	}

	// 4. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return constant.ErrSystemError
	}

	// 5. 更新密码
	updates := map[string]interface{}{
		"password": string(hashedPassword),
	}

	if err := s.userRepo.UpdateFields(user.ID, updates); err != nil {
		return constant.ErrUpdateUserFailed
	}

	// 6. 清除缓存
	s.ClearUserCache(phone)

	return nil
}

// RefreshUserCache 刷新用户缓存
func (s *userServiceV2) RefreshUserCache(phone string) error {
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		return err
	}

	return s.setUserCache(user)
}

// ClearUserCache 清除用户缓存
func (s *userServiceV2) ClearUserCache(phone string) error {
	cacheKey := constant.CacheKeyUserInfo + phone
	return s.cacheHelper.Delete(cacheKey)
}

// ==================== 私有辅助方法 ====================

// recordLoginFail 记录登录失败
func (s *userServiceV2) recordLoginFail(phone string) {
	failKey := fmt.Sprintf("login:fail:%s", phone)
	count, _ := s.cacheHelper.Incr(failKey)

	// 第一次失败，设置过期时间（5分钟）
	if count == 1 {
		s.cacheHelper.Expire(failKey, time.Minute*5)
	}

	// 失败5次，锁定账号5分钟
	if count >= 5 {
		lockKey := fmt.Sprintf("login:lock:%s", phone)
		s.cacheHelper.SetString(lockKey, "1", time.Duration(constant.LockTimeLogin)*time.Second)
	}
}

// setUserCache 设置用户缓存
func (s *userServiceV2) setUserCache(user *model.User) error {
	cacheKey := constant.CacheKeyUserInfo + user.Phone
	return s.cacheHelper.Set(cacheKey, user, time.Duration(constant.CacheExpireMedium)*time.Second)
}
