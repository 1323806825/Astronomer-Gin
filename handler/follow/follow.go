package follow

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	followService service.FollowServiceV2
	userService   service.UserServiceV2
}

func NewFollowHandler(followService service.FollowServiceV2, userService service.UserServiceV2) *FollowHandler {
	return &FollowHandler{
		followService: followService,
		userService:   userService,
	}
}

// FollowUser 关注用户
func (h *FollowHandler) FollowUser(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 关注用户
	if err := h.followService.FollowUser(user.ID, targetUserID, user.Username); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "关注成功", nil)
}

// UnfollowUser 取消关注
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 取消关注
	if err := h.followService.UnfollowUser(user.ID, targetUserID); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "取消关注成功", nil)
}

// GetFollowers 获取粉丝列表
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取粉丝列表
	users, total, err := h.followService.GetFollowers(userID, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetFollowing 获取关注列表
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	userID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	// 获取关注列表
	users, total, err := h.followService.GetFollowing(userID, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// BlockUser 拉黑用户
func (h *FollowHandler) BlockUser(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 拉黑用户
	if err := h.followService.BlockUser(user.ID, targetUserID); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "拉黑成功", nil)
}

// UnblockUser 取消拉黑
func (h *FollowHandler) UnblockUser(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 取消拉黑
	if err := h.followService.UnblockUser(user.ID, targetUserID); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "取消拉黑成功", nil)
}

// GetBlockList 获取拉黑列表
func (h *FollowHandler) GetBlockList(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取拉黑列表
	users, err := h.followService.GetBlockList(user.ID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list": users,
	})
}

// CheckFollowing 检查是否已关注
func (h *FollowHandler) CheckFollowing(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 检查是否已关注
	isFollowing := h.followService.IsFollowing(user.ID, targetUserID)

	util.Success(c, gin.H{
		"is_following": isFollowing,
	})
}

// ==================== 好友相关接口 ====================

// GetFriendsList 获取好友列表（互关的用户）
func (h *FollowHandler) GetFriendsList(c *gin.Context) {
	phone, _ := c.Get("phone")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取好友列表
	users, total, err := h.followService.GetFriendsList(user.ID, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CheckFriend 检查是否为好友
func (h *FollowHandler) CheckFriend(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 检查是否为好友
	isFriend := h.followService.IsFriend(user.ID, targetUserID)

	util.Success(c, gin.H{
		"is_friend": isFriend,
	})
}

// GetFriendsCount 获取好友数量
func (h *FollowHandler) GetFriendsCount(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取好友数量
	count, err := h.followService.GetFriendsCount(user.ID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"count": count,
	})
}
