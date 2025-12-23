package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *model.Notification) error
	FindByID(id uint64) (*model.Notification, error)
	GetList(userID string, page, pageSize int) ([]model.Notification, int64, error)
	GetUnreadCount(userID string) (int64, error)
	MarkAsRead(id uint64) error
	MarkAllAsRead(userID string) error
	Delete(id uint64) error
	DeleteByRelated(relatedType string, relatedID uint64) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create 创建通知
func (r *notificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

// FindByID 根据ID查找通知
func (r *notificationRepository) FindByID(id uint64) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetList 获取通知列表（分页）
func (r *notificationRepository) GetList(userID string, page, pageSize int) ([]model.Notification, int64, error) {
	var total int64
	var notifications []model.Notification

	// 获取总数
	if err := r.db.Model(&model.Notification{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetUnreadCount 获取未读通知数量
func (r *notificationRepository) GetUnreadCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记单条通知为已读
func (r *notificationRepository) MarkAsRead(id uint64) error {
	return r.db.Model(&model.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

// MarkAllAsRead 标记所有通知为已读
func (r *notificationRepository) MarkAllAsRead(userID string) error {
	return r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// Delete 删除通知
func (r *notificationRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Notification{}, id).Error
}

// DeleteByRelated 根据关联内容删除通知（如文章被删除时，删除相关通知）
func (r *notificationRepository) DeleteByRelated(relatedType string, relatedID uint64) error {
	return r.db.Where("related_type = ? AND related_id = ?", relatedType, relatedID).
		Delete(&model.Notification{}).Error
}
