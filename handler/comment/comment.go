package comment

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService service.CommentServiceV2
	userService    service.UserServiceV2
}

func NewCommentHandler(commentService service.CommentServiceV2, userService service.UserServiceV2) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		userService:    userService,
	}
}

// CreateComment 创建一级评论
func (h *CommentHandler) CreateComment(c *gin.Context) {
	phone, _ := c.Get("phone")

	var req struct {
		ArticleID int64  `json:"articleId" binding:"required"`
		Comment   string `json:"comment" binding:"required"`
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

	// 创建评论
	commentID, err := h.commentService.CreateComment(
		req.ArticleID,
		req.Comment,
		user.Username,
		user.Phone,
		c.ClientIP(),
		user.ID,
	)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.InsertSuccess, gin.H{
		"id": commentID,
	})
}

// GetCommentList 获取文章评论列表
func (h *CommentHandler) GetCommentList(c *gin.Context) {
	articleID := c.Param("articleId")

	comments, err := h.commentService.GetCommentList(articleID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, comments)
}

// CreateSubComment 创建二级评论
func (h *CommentHandler) CreateSubComment(c *gin.Context) {
	phone, _ := c.Get("phone")

	var req struct {
		ParentCommentID int64  `json:"parentCommentId" binding:"required"`
		Comment         string `json:"comment" binding:"required"`
		ToUserID        string `json:"toUserId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 创建二级评论
	commentID, err := h.commentService.CreateSubComment(
		req.ParentCommentID,
		req.ToUserID,
		req.Comment,
		user.Username,
		user.Phone,
		"", // toUsername - service层会查询
		"", // toPhone - service层会查询
		c.ClientIP(),
		user.ID,
	)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.InsertSuccess, gin.H{
		"id": commentID,
	})
}

// GetSubCommentList 获取二级评论列表
func (h *CommentHandler) GetSubCommentList(c *gin.Context) {
	parentID := c.Param("parentId")

	subComments, err := h.commentService.GetSubCommentList(parentID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, subComments)
}

// LikeComment 点赞/取消点赞评论
func (h *CommentHandler) LikeComment(c *gin.Context) {
	commentID := c.Param("id")
	commentType := c.Query("type") // parent 或 sub
	phone, _ := c.Get("phone")

	// 获取用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	cid, _ := strconv.ParseInt(commentID, 10, 64)

	// 点赞/取消点赞
	if err := h.commentService.LikeComment(cid, user.ID, commentType); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.Success, nil)
}
