package blog

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BlogHandler struct {
	blogService service.BlogServiceV2
	userService service.UserServiceV2
	notifyRepo  repository.NotificationRepository
}

func NewBlogHandler(blogService service.BlogServiceV2, userService service.UserServiceV2, notifyRepo repository.NotificationRepository) *BlogHandler {
	return &BlogHandler{
		blogService: blogService,
		userService: userService,
		notifyRepo:  notifyRepo,
	}
}

// CreateBlog 创建博客
func (h *BlogHandler) CreateBlog(c *gin.Context) {
	phone, _ := c.Get("phone")

	var req struct {
		Title   string `json:"title" binding:"required"`
		Preface string `json:"preface"`
		Content string `json:"content" binding:"required"`
		Photo   string `json:"photo"`
		Tag     string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 创建博客（直接发布）
	articleID, err := h.blogService.CreateArticle(user.ID, req.Title, req.Preface, req.Content, req.Photo, req.Tag)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.InsertSuccess, gin.H{
		"id": articleID,
	})
}

// GetBlogList 获取博客列表
func (h *BlogHandler) GetBlogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	tag := c.Query("tag")

	articles, total, err := h.blogService.GetArticleList(page, pageSize, tag)
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

// GetBlogDetail 获取博客详情
func (h *BlogHandler) GetBlogDetail(c *gin.Context) {
	id := c.Param("id")
	articleID, _ := strconv.ParseUint(id, 10, 64)

	article, err := h.blogService.GetArticleDetail(articleID)
	if err != nil {
		util.NotFound(c, err.Error())
		return
	}

	util.Success(c, article)
}

// UpdateBlog 更新博客
func (h *BlogHandler) UpdateBlog(c *gin.Context) {
	id := c.Param("id")
	phone, _ := c.Get("phone")
	articleID, _ := strconv.ParseUint(id, 10, 64)

	var req struct {
		Title   string `json:"title"`
		Preface string `json:"preface"`
		Content string `json:"content"`
		Photo   string `json:"photo"`
		Tag     string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Preface != "" {
		updates["preface"] = req.Preface
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Photo != "" {
		updates["photo"] = req.Photo
	}
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}

	// 更新博客
	if err := h.blogService.UpdateArticle(articleID, user.ID, updates); err != nil {
		util.Forbidden(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.UpdateSuccess, nil)
}

// DeleteBlog 删除博客
func (h *BlogHandler) DeleteBlog(c *gin.Context) {
	id := c.Param("id")
	phone, _ := c.Get("phone")
	articleID, _ := strconv.ParseUint(id, 10, 64)

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 删除博客
	if err := h.blogService.DeleteArticle(articleID, user.ID); err != nil {
		util.Forbidden(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.DeleteSuccess, nil)
}

// LikeBlog 点赞/取消点赞博客
func (h *BlogHandler) LikeBlog(c *gin.Context) {
	id := c.Param("id")
	phone, _ := c.Get("phone")
	articleID, _ := strconv.ParseUint(id, 10, 64)

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 点赞/取消点赞
	if err := h.blogService.LikeArticle(articleID, user.ID, user.Username, h.notifyRepo); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.Success, nil)
}

// SaveDraft 保存草稿
func (h *BlogHandler) SaveDraft(c *gin.Context) {
	phone, _ := c.Get("phone")

	var req struct {
		Title   string `json:"title" binding:"required"`
		Preface string `json:"preface"`
		Content string `json:"content"`
		Photo   string `json:"photo"`
		Tag     string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 保存草稿
	draftID, err := h.blogService.SaveDraft(user.ID, req.Title, req.Preface, req.Content, req.Photo, req.Tag)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "草稿保存成功", gin.H{
		"id": draftID,
	})
}

// GetDrafts 获取草稿列表
func (h *BlogHandler) GetDrafts(c *gin.Context) {
	phone, _ := c.Get("phone")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取草稿列表
	drafts, total, err := h.blogService.GetUserDrafts(user.ID, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     drafts,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// PublishDraft 发布草稿
func (h *BlogHandler) PublishDraft(c *gin.Context) {
	id := c.Param("id")
	phone, _ := c.Get("phone")
	draftID, _ := strconv.ParseUint(id, 10, 64)

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 发布草稿
	if err := h.blogService.PublishDraft(draftID, user.ID); err != nil {
		util.Forbidden(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "发布成功", nil)
}
