package user

import (
	"astronomer-gin/pkg/captcha"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	uuidPkg "astronomer-gin/pkg/uuid"
	"astronomer-gin/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserServiceV2
}

func NewUserHandler(userService service.UserServiceV2) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 新用户注册账号，需要提供手机号、用户名、密码和验证码
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body object{phone=string,username=string,password=string,captchaId=string,captchaVal=string} true "注册信息"
// @Success 200 {object} object{code=int,message=string,data=object} "注册成功"
// @Failure 400 {object} object{code=int,message=string} "参数错误或验证码错误"
// @Failure 500 {object} object{code=int,message=string} "服务器内部错误"
// @Router /user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Phone      string `json:"phone" binding:"required"`
		Username   string `json:"username" binding:"required"`
		Password   string `json:"password" binding:"required"`
		CaptchaID  string `json:"captchaId" binding:"required"`
		CaptchaVal string `json:"captchaVal" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	// 验证验证码
	if !captcha.VerifyCaptcha(req.CaptchaID, req.CaptchaVal) {
		util.BadRequest(c, constant.CaptchaError)
		return
	}

	// 调用service层注册用户
	if err := h.userService.Register(req.Phone, req.Password, req.Username); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	// 注册成功后自动登录，返回token和用户信息
	token, user, err := h.userService.Login(req.Phone, req.Password)
	if err != nil {
		util.InternalServerError(c, "注册成功但自动登录失败")
		return
	}

	util.SuccessWithMessage(c, constant.RegisterSuccess, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"phone":    user.Phone,
			"username": user.Username,
			"icon":     user.Icon,
			"sex":      user.Sex,
		},
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取Token，可选验证码验证
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body object{phone=string,password=string,captchaId=string,captchaVal=string} true "登录信息"
// @Success 200 {object} object{code=int,message=string,data=object{token=string,userInfo=object}} "登录成功，返回token和用户信息"
// @Failure 400 {object} object{code=int,message=string} "参数错误、验证码错误或登录失败"
// @Router /user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Phone      string `json:"phone" binding:"required"`
		Password   string `json:"password" binding:"required"`
		CaptchaID  string `json:"captchaId"`
		CaptchaVal string `json:"captchaVal"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	// 如果提供了验证码，进行验证
	if req.CaptchaID != "" && req.CaptchaVal != "" {
		if !captcha.VerifyCaptcha(req.CaptchaID, req.CaptchaVal) {
			util.BadRequest(c, constant.CaptchaError)
			return
		}
	}

	// 调用service层登录
	token, user, err := h.userService.Login(req.Phone, req.Password)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.LoginSuccess, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"phone":    user.Phone,
			"username": user.Username,
			"icon":     user.Icon,
			"sex":      user.Sex,
		},
	})
}

// GetCaptcha 获取验证码
// @Summary 获取图形验证码
// @Description 生成图形验证码，返回验证码ID和Base64图片
// @Tags 用户模块
// @Produce json
// @Success 200 {object} object{code=int,data=object{captchaId=string,picPath=string}} "返回验证码ID和图片"
// @Failure 500 {object} object{code=int,message=string} "服务器内部错误"
// @Router /user/captcha [get]
func (h *UserHandler) GetCaptcha(c *gin.Context) {
	captchaInfo, err := captcha.GetCaptcha(c.Request.Context())
	if err != nil {
		util.InternalServerError(c, constant.Fail)
		return
	}

	util.Success(c, captchaInfo)
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	phone, _ := c.Get("phone")

	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"id":             user.ID,
		"phone":          user.Phone,
		"username":       user.Username,
		"avatar":         user.Icon,  // icon -> avatar (前端字段)
		"bio":            user.Intro, // intro -> bio (前端字段)
		"sex":            user.Sex,
		"note":           user.Note,
		"followingCount": user.FollowingCount,
		"followedCount":  user.FollowedCount,
		"createTime":     user.CreateTime,
	})
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	phone, _ := c.Get("phone")

	var req struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"` // 前端使用 avatar
		Bio      string `json:"bio"`    // 前端使用 bio (个人简介)
		Sex      int    `json:"sex"`
		Note     string `json:"note"` // 备注（可选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, constant.ParamError)
		return
	}

	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	// avatar 映射到 icon 字段
	if req.Avatar != "" {
		updates["icon"] = req.Avatar
	}
	// bio 映射到 intro 字段
	if req.Bio != "" {
		updates["intro"] = req.Bio
	}
	if req.Sex >= 0 {
		updates["sex"] = req.Sex
	}
	if req.Note != "" {
		updates["note"] = req.Note
	}

	if err := h.userService.UpdateUserInfo(phone.(string), updates); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, constant.UpdateSuccess, nil)
}

// GetUserProfile 获取用户资料（公开接口，通过用户ID）
// @Summary 获取用户资料
// @Description 获取指定用户的公开资料信息
// @Tags 用户模块
// @Produce json
// @Param id path string true "用户ID(UUID)"
// @Success 200 {object} object{code=int,data=object} "返回用户资料"
// @Failure 400 {object} object{code=int,message=string} "参数错误"
// @Failure 404 {object} object{code=int,message=string} "用户不存在"
// @Router /user/{id} [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		util.BadRequest(c, "用户ID不能为空")
		return
	}

	// 验证UUID格式
	if !uuidPkg.IsValid(userID) {
		util.BadRequest(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetUserInfoByID(userID)
	if err != nil {
		util.NotFound(c, "用户不存在")
		return
	}

	// 返回与前端期望字段名匹配的数据
	util.Success(c, gin.H{
		"id":              user.ID,
		"username":        user.Username,
		"avatar":          user.Icon,  // icon -> avatar
		"bio":             user.Intro, // intro -> bio
		"sex":             user.Sex,
		"article_count":   0,                   // TODO: 需要从文章表统计
		"follower_count":  user.FollowedCount,  // followedCount -> follower_count
		"following_count": user.FollowingCount, // followingCount -> following_count
		"like_count":      0,                   // TODO: 需要从点赞表统计
		"favorite_count":  0,                   // TODO: 需要从收藏表统计
		"create_time":     user.CreateTime,
	})
}
