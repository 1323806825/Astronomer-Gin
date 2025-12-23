package notification

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService service.NotificationServiceV2
	userService         service.UserServiceV2
}

func NewNotificationHandler(notificationService service.NotificationServiceV2, userService service.UserServiceV2) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		userService:         userService,
	}
}

// GetNotifications 获取通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	phone, _ := c.Get("phone")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取通知列表
	notifications, total, err := h.notificationService.GetNotifications(user.ID, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     notifications,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetUnreadCount 获取未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取未读数量
	count, err := h.notificationService.GetUnreadCount(user.ID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"unread_count": count,
	})
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	phone, _ := c.Get("phone")
	notificationID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	nid, _ := strconv.ParseUint(notificationID, 10, 64)

	// 标记为已读
	if err := h.notificationService.MarkAsRead(user.ID, nid); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "标记成功", nil)
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 全部标记为已读
	if err := h.notificationService.MarkAllAsRead(user.ID); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "全部标记成功", nil)
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	phone, _ := c.Get("phone")
	notificationID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	nid, _ := strconv.ParseUint(notificationID, 10, 64)

	// 删除通知
	if err := h.notificationService.DeleteNotification(user.ID, nid); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除成功", nil)
}
