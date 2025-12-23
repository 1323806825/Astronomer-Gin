package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"fmt"
	"time"
)

// NotificationServiceV2 企业级通知服务接口
type NotificationServiceV2 interface {
	GetNotifications(userID string, page, pageSize int) ([]model.Notification, int64, error)
	GetUnreadCount(userID string) (int64, error)
	MarkAsRead(userID string, notificationID uint64) error
	MarkAllAsRead(userID string) error
	DeleteNotification(userID string, notificationID uint64) error

	// 缓存管理
	RefreshNotificationCache(userID string) error
	ClearNotificationCache(userID string) error
}

type notificationServiceV2 struct {
	notifyRepo  repository.NotificationRepository
	cacheHelper *util.CacheHelper
}

func NewNotificationServiceV2(notifyRepo repository.NotificationRepository) NotificationServiceV2 {
	return &notificationServiceV2{
		notifyRepo:  notifyRepo,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
	}
}

// GetNotifications 获取通知列表（带缓存）
func (s *notificationServiceV2) GetNotifications(userID string, page, pageSize int) ([]model.Notification, int64, error) {
	// 1. 参数验证
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 从缓存获取
	cacheKey := fmt.Sprintf("%s%d:page:%d:size:%d", constant.CacheKeyNotification, userID, page, pageSize)

	type CachedData struct {
		Notifications []model.Notification
		Total         int64
	}

	var cached CachedData
	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&cached,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			notifications, total, err := s.notifyRepo.GetList(userID, page, pageSize)
			if err != nil {
				return nil, err
			}
			return CachedData{Notifications: notifications, Total: total}, nil
		},
	)

	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return cached.Notifications, cached.Total, nil
}

// GetUnreadCount 获取未读通知数量（带缓存）
func (s *notificationServiceV2) GetUnreadCount(userID string) (int64, error) {
	cacheKey := fmt.Sprintf("%sunread:%d", constant.CacheKeyNotification, userID)
	var count int64

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&count,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			return s.notifyRepo.GetUnreadCount(userID)
		},
	)

	if err != nil {
		return 0, constant.ErrDatabaseQuery
	}

	return count, nil
}

// MarkAsRead 标记通知为已读（企业级实现）
func (s *notificationServiceV2) MarkAsRead(userID string, notificationID uint64) error {
	// 1. 验证通知是否存在且属于该用户
	notification, err := s.notifyRepo.FindByID(notificationID)
	if err != nil {
		return constant.ErrNotificationNotFound
	}

	if notification.UserID != userID {
		return constant.ErrUnauthorized
	}

	// 2. 标记为已读
	if err := s.notifyRepo.MarkAsRead(notificationID); err != nil {
		return constant.ErrUpdateFail
	}

	// 3. 清除缓存
	s.ClearNotificationCache(userID)

	return nil
}

// MarkAllAsRead 标记所有通知为已读（企业级实现）
func (s *notificationServiceV2) MarkAllAsRead(userID string) error {
	// 1. 标记所有通知为已读
	if err := s.notifyRepo.MarkAllAsRead(userID); err != nil {
		return constant.ErrUpdateFail
	}

	// 2. 清除缓存
	s.ClearNotificationCache(userID)

	return nil
}

// DeleteNotification 删除通知（企业级实现）
func (s *notificationServiceV2) DeleteNotification(userID string, notificationID uint64) error {
	// 1. 验证通知是否存在且属于该用户
	notification, err := s.notifyRepo.FindByID(notificationID)
	if err != nil {
		return constant.ErrNotificationNotFound
	}

	if notification.UserID != userID {
		return constant.ErrUnauthorized
	}

	// 2. 删除通知
	if err := s.notifyRepo.Delete(notificationID); err != nil {
		return constant.ErrDeleteFail
	}

	// 3. 清除缓存
	s.ClearNotificationCache(userID)

	return nil
}

// RefreshNotificationCache 刷新通知缓存
func (s *notificationServiceV2) RefreshNotificationCache(userID string) error {
	// 清除所有相关缓存，下次请求时会重新加载
	return s.ClearNotificationCache(userID)
}

// ClearNotificationCache 清除通知缓存
func (s *notificationServiceV2) ClearNotificationCache(userID string) error {
	// 删除该用户所有通知相关缓存
	pattern := fmt.Sprintf("%s%d:*", constant.CacheKeyNotification, userID)
	return s.cacheHelper.DeleteByPattern(pattern)
}
