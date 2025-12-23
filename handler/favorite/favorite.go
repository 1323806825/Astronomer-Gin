package favorite

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	favoriteService service.FavoriteServiceV2
	userService     service.UserServiceV2
}

func NewFavoriteHandler(favoriteService service.FavoriteServiceV2, userService service.UserServiceV2) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteService: favoriteService,
		userService:     userService,
	}
}

// FavoriteArticle 收藏文章
func (h *FavoriteHandler) FavoriteArticle(c *gin.Context) {
	phone, _ := c.Get("phone")
	articleID := c.Param("id")

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	aid, _ := strconv.ParseUint(articleID, 10, 64)

	// 收藏文章
	if err := h.favoriteService.FavoriteArticle(user.ID, aid); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "收藏成功", nil)
}

// UnfavoriteArticle 取消收藏
func (h *FavoriteHandler) UnfavoriteArticle(c *gin.Context) {
	phone, _ := c.Get("phone")
	articleID := c.Param("id")

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	aid, _ := strconv.ParseUint(articleID, 10, 64)

	// 取消收藏
	if err := h.favoriteService.UnfavoriteArticle(user.ID, aid); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "取消收藏成功", nil)
}

// GetUserFavorites 获取用户的收藏列表
func (h *FavoriteHandler) GetUserFavorites(c *gin.Context) {
	phone, _ := c.Get("phone")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取收藏列表
	articles, total, err := h.favoriteService.GetUserFavorites(user.ID, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     articles,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CheckFavorited 检查是否已收藏
func (h *FavoriteHandler) CheckFavorited(c *gin.Context) {
	phone, _ := c.Get("phone")
	articleID := c.Param("id")

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	aid, _ := strconv.ParseUint(articleID, 10, 64)

	// 检查是否已收藏
	isFavorited := h.favoriteService.IsFavorited(user.ID, aid)

	util.Success(c, gin.H{
		"is_favorited": isFavorited,
	})
}
