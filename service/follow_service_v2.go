package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/database"
	"astronomer-gin/pkg/queue"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// FollowServiceV2 企业级关注服务接口
type FollowServiceV2 interface {
	// 关注相关
	FollowUser(userID string, followUserID string, username string) error
	UnfollowUser(userID string, followUserID string) error
	IsFollowing(userID string, followUserID string) bool
	GetFollowers(userID string, page, pageSize int) ([]model.User, int64, error)
	GetFollowing(userID string, page, pageSize int) ([]model.User, int64, error)

	// 好友相关（互关即好友）
	IsFriend(userID, targetUserID string) bool
	GetFriendsList(userID string, page, pageSize int) ([]model.User, int64, error)
	GetFriendsCount(userID string) (int64, error)

	// 拉黑相关
	BlockUser(userID, blockUserID string) error
	UnblockUser(userID, blockUserID string) error
	IsBlocked(userID, blockUserID string) bool
	GetBlockList(userID string) ([]model.User, error)

	// 缓存管理
	RefreshFollowCache(userID string) error
	ClearFollowCache(userID string) error
}

type followServiceV2 struct {
	followRepo  repository.FollowRepository
	userRepo    repository.UserRepository
	notifyRepo  repository.NotificationRepository
	cacheHelper *util.CacheHelper
	db          *gorm.DB
}

func NewFollowServiceV2(followRepo repository.FollowRepository, userRepo repository.UserRepository, notifyRepo repository.NotificationRepository) FollowServiceV2 {
	return &followServiceV2{
		followRepo:  followRepo,
		userRepo:    userRepo,
		notifyRepo:  notifyRepo,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
		db:          database.GetDB(),
	}
}

// FollowUser 关注用户（企业级实现）
func (s *followServiceV2) FollowUser(userID, followUserID string, username string) error {
	// 1. 参数验证
	if userID == followUserID {
		return constant.ErrCannotFollowSelf
	}

	// 2. 检查被关注用户是否存在
	_, err := s.userRepo.FindByID(followUserID)
	if err != nil {
		return constant.ErrUserNotExist
	}

	// 3. 检查是否已关注
	if s.followRepo.IsFollowing(userID, followUserID) {
		return constant.ErrAlreadyFollowed
	}

	// 4. 检查是否被拉黑
	if s.followRepo.IsBlocked(followUserID, userID) {
		return constant.ErrBlocked
	}

	// 5. 使用事务创建关注关系并更新计数
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 5.1 创建关注关系
		follow := &model.UserFollow{
			UserID:       userID,
			FollowUserID: followUserID,
		}
		if err := tx.Create(follow).Error; err != nil {
			return err
		}

		// 5.2 更新关注者的关注数
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
			return err
		}

		// 5.3 更新被关注者的粉丝数
		if err := tx.Model(&model.User{}).Where("id = ?", followUserID).
			UpdateColumn("followed_count", gorm.Expr("followed_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return constant.ErrFollowFailed
	}

	// 6. 异步发送关注通知（不影响主流程）
	if queue.Client != nil {
		go func() {
			task := queue.CreateTask(queue.TaskTypeNotification, map[string]interface{}{
				"notify_type":       "follow",
				"user_id":           followUserID,
				"follower_id":       userID,
				"follower_username": username,
			})
			ctx := context.Background()
			if err := queue.Client.PublishTask(ctx, task); err != nil {
				log.Printf("Failed to publish follow notification task: %v", err)
			}
		}()
	}

	// 7. 清除相关缓存
	s.ClearFollowCache(userID)
	s.ClearFollowCache(followUserID)

	// 8. 清除双方用户信息缓存（因为following_count和followed_count已更新）
	s.cacheHelper.Delete(fmt.Sprintf("%s%s", constant.CacheKeyUserInfo, userID))
	s.cacheHelper.Delete(fmt.Sprintf("%s%s", constant.CacheKeyUserInfo, followUserID))

	return nil
}

// UnfollowUser 取消关注（企业级实现）
func (s *followServiceV2) UnfollowUser(userID, followUserID string) error {
	// 1. 检查是否已关注
	if !s.followRepo.IsFollowing(userID, followUserID) {
		return constant.ErrNotFollowed
	}

	// 2. 使用事务取消关注���更新计数
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 2.1 删除关注关系
		if err := tx.Where("user_id = ? AND follow_user_id = ?", userID, followUserID).
			Delete(&model.UserFollow{}).Error; err != nil {
			return err
		}

		// 2.2 更新关注者的关注数
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("following_count", gorm.Expr("following_count - ?", 1)).Error; err != nil {
			return err
		}

		// 2.3 更新被关注者的粉丝数
		if err := tx.Model(&model.User{}).Where("id = ?", followUserID).
			UpdateColumn("followed_count", gorm.Expr("followed_count - ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return constant.ErrUnfollowFailed
	}

	// 3. 清除相关缓存
	s.ClearFollowCache(userID)
	s.ClearFollowCache(followUserID)

	// 4. 清除双方用户信息缓存（因为following_count和followed_count已更新）
	s.cacheHelper.Delete(fmt.Sprintf("%s%s", constant.CacheKeyUserInfo, userID))
	s.cacheHelper.Delete(fmt.Sprintf("%s%s", constant.CacheKeyUserInfo, followUserID))

	return nil
}

// IsFollowing 检查是否已关注
func (s *followServiceV2) IsFollowing(userID, followUserID string) bool {
	return s.followRepo.IsFollowing(userID, followUserID)
}

// GetFollowers 获取粉丝列表（带缓存）
func (s *followServiceV2) GetFollowers(userID string, page, pageSize int) ([]model.User, int64, error) {
	// 1. 参数验证
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 从缓存获取
	cacheKey := fmt.Sprintf("%sfollowers:%d:page:%d:size:%d", constant.CacheKeyFollow, userID, page, pageSize)

	type CachedData struct {
		Users []model.User
		Total int64
	}

	var cached CachedData
	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&cached,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			userIDs, total, err := s.followRepo.GetFollowers(userID, page, pageSize)
			if err != nil {
				return nil, err
			}

			if len(userIDs) == 0 {
				return CachedData{Users: []model.User{}, Total: total}, nil
			}

			// 根据ID列表查询用户信息
			users := make([]model.User, 0, len(userIDs))
			for _, id := range userIDs {
				user, err := s.userRepo.FindByID(id)
				if err == nil {
					user.Password = "" // 数据脱敏
					users = append(users, *user)
				}
			}

			return CachedData{Users: users, Total: total}, nil
		},
	)

	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return cached.Users, cached.Total, nil
}

// GetFollowing 获取关注列表（带缓存）
func (s *followServiceV2) GetFollowing(userID string, page, pageSize int) ([]model.User, int64, error) {
	// 1. 参数验证
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 从缓存获取
	cacheKey := fmt.Sprintf("%sfollowing:%d:page:%d:size:%d", constant.CacheKeyFollow, userID, page, pageSize)

	type CachedData struct {
		Users []model.User
		Total int64
	}

	var cached CachedData
	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&cached,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			userIDs, total, err := s.followRepo.GetFollowing(userID, page, pageSize)
			if err != nil {
				return nil, err
			}

			if len(userIDs) == 0 {
				return CachedData{Users: []model.User{}, Total: total}, nil
			}

			// 根据ID列表查询用户信息
			users := make([]model.User, 0, len(userIDs))
			for _, id := range userIDs {
				user, err := s.userRepo.FindByID(id)
				if err == nil {
					user.Password = "" // 数据脱敏
					users = append(users, *user)
				}
			}

			return CachedData{Users: users, Total: total}, nil
		},
	)

	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return cached.Users, cached.Total, nil
}

// BlockUser 拉黑用户（企业级实现）
func (s *followServiceV2) BlockUser(userID, blockUserID string) error {
	// 1. 参数验证
	if userID == blockUserID {
		return constant.ErrCannotBlockSelf
	}

	// 2. 检查是否已拉黑
	if s.followRepo.IsBlocked(userID, blockUserID) {
		return constant.ErrAlreadyBlocked
	}

	// 3. 使用事务处理拉黑及相关操作
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 3.1 如果已关注，先取消关注
		if s.followRepo.IsFollowing(userID, blockUserID) {
			if err := tx.Where("user_id = ? AND follow_user_id = ?", userID, blockUserID).
				Delete(&model.UserFollow{}).Error; err != nil {
				return err
			}
			// 更新关注数
			tx.Model(&model.User{}).Where("id = ?", userID).
				UpdateColumn("following_count", gorm.Expr("following_count - ?", 1))
			tx.Model(&model.User{}).Where("id = ?", blockUserID).
				UpdateColumn("followed_count", gorm.Expr("followed_count - ?", 1))
		}

		// 3.2 如果对方关注了自己，删除对方的关注
		if s.followRepo.IsFollowing(blockUserID, userID) {
			if err := tx.Where("user_id = ? AND follow_user_id = ?", blockUserID, userID).
				Delete(&model.UserFollow{}).Error; err != nil {
				return err
			}
			// 更新关注数
			tx.Model(&model.User{}).Where("id = ?", blockUserID).
				UpdateColumn("following_count", gorm.Expr("following_count - ?", 1))
			tx.Model(&model.User{}).Where("id = ?", userID).
				UpdateColumn("followed_count", gorm.Expr("followed_count - ?", 1))
		}

		// 3.3 创建拉黑记录
		block := &model.UserBlock{
			UserID:      userID,
			BlockUserID: blockUserID,
		}
		if err := tx.Create(block).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return constant.ErrBlockFailed
	}

	// 4. 清除相关缓存
	s.ClearFollowCache(userID)
	s.ClearFollowCache(blockUserID)

	return nil
}

// UnblockUser 取消拉黑（企业级实现）
func (s *followServiceV2) UnblockUser(userID, blockUserID string) error {
	// 1. 检查是否已拉黑
	if !s.followRepo.IsBlocked(userID, blockUserID) {
		return constant.ErrNotBlocked
	}

	// 2. 删除拉黑记录
	if err := s.followRepo.Unblock(userID, blockUserID); err != nil {
		return constant.ErrUnblockFailed
	}

	// 3. 清除缓存
	s.ClearFollowCache(userID)

	return nil
}

// IsBlocked 检查是否已拉黑
func (s *followServiceV2) IsBlocked(userID, blockUserID string) bool {
	return s.followRepo.IsBlocked(userID, blockUserID)
}

// GetBlockList 获取拉黑列表（带缓存）
func (s *followServiceV2) GetBlockList(userID string) ([]model.User, error) {
	// 1. 从缓存获取
	cacheKey := fmt.Sprintf("%sblock:%d", constant.CacheKeyFollow, userID)
	var users []model.User

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&users,
		time.Duration(constant.CacheExpireMedium)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			userIDs, err := s.followRepo.GetBlockList(userID)
			if err != nil {
				return nil, err
			}

			if len(userIDs) == 0 {
				return []model.User{}, nil
			}

			// 根据ID列表查询用户信息
			resultUsers := make([]model.User, 0, len(userIDs))
			for _, id := range userIDs {
				user, err := s.userRepo.FindByID(id)
				if err == nil {
					user.Password = "" // 数据脱敏
					resultUsers = append(resultUsers, *user)
				}
			}

			return resultUsers, nil
		},
	)

	if err != nil {
		return nil, constant.ErrDatabaseQuery
	}

	return users, nil
}

// RefreshFollowCache 刷新关注缓存
func (s *followServiceV2) RefreshFollowCache(userID string) error {
	// 清除所有相关缓存，下次请求时会重新加载
	return s.ClearFollowCache(userID)
}

// ClearFollowCache 清除关注缓存
func (s *followServiceV2) ClearFollowCache(userID string) error {
	// 删除该用户所有关注相关缓存
	pattern := fmt.Sprintf("%s*:%d:*", constant.CacheKeyFollow, userID)
	return s.cacheHelper.DeleteByPattern(pattern)
}

// ==================== 好友相关方法 ====================

// IsFriend 检查是否为好友（互相关注）
func (s *followServiceV2) IsFriend(userID, targetUserID string) bool {
	return s.followRepo.IsFriend(userID, targetUserID)
}

// GetFriendsList 获取好友列表（带缓存）
func (s *followServiceV2) GetFriendsList(userID string, page, pageSize int) ([]model.User, int64, error) {
	// 1. 参数验证
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 从缓存获取
	cacheKey := fmt.Sprintf("%sfriends:%d:page:%d:size:%d", constant.CacheKeyFollow, userID, page, pageSize)

	type CachedData struct {
		Users []model.User
		Total int64
	}

	var cached CachedData
	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&cached,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询好友ID列表
			friendIDs, total, err := s.followRepo.GetFriendsList(userID, page, pageSize)
			if err != nil {
				return nil, err
			}

			if len(friendIDs) == 0 {
				return CachedData{Users: []model.User{}, Total: total}, nil
			}

			// 根据ID列表查询用户信息
			users := make([]model.User, 0, len(friendIDs))
			for _, id := range friendIDs {
				user, err := s.userRepo.FindByID(id)
				if err == nil {
					user.Password = "" // 数据脱敏
					users = append(users, *user)
				}
			}

			return CachedData{Users: users, Total: total}, nil
		},
	)

	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return cached.Users, cached.Total, nil
}

// GetFriendsCount 获取好友数量（带缓存）
func (s *followServiceV2) GetFriendsCount(userID string) (int64, error) {
	// 从缓存获取
	cacheKey := fmt.Sprintf("%sfriends_count:%d", constant.CacheKeyFollow, userID)
	var count int64

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&count,
		time.Duration(constant.CacheExpireMedium)*time.Second,
		func() (interface{}, error) {
			return s.followRepo.GetFriendsCount(userID)
		},
	)

	return count, err
}
