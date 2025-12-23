package handler

import (
	"astronomer-gin/pkg/response"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ColumnHandler struct {
	columnService service.ColumnService
}

func NewColumnHandler(columnService service.ColumnService) *ColumnHandler {
	return &ColumnHandler{
		columnService: columnService,
	}
}

// GetList 获取专栏列表
// @Summary 获取专栏列表
// @Description 获取所有专栏的分页列表
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} object{code=int,data=object{columns=[]model.ArticleColumn,total=int,page=int,page_size=int}}
// @Failure 400 {object} object{code=int,message=string}
// @Router /api/v3/columns [get]
func (h *ColumnHandler) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	columns, total, err := h.columnService.GetList(page, pageSize)
	if err != nil {
		response.ServerError(c, "获取专栏列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"columns":   columns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDetail 获取专栏详情
// @Summary 获取专栏详情
// @Description 根据ID获取专栏的详细信息
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Success 200 {object} object{code=int,data=service.ColumnDetailResponse}
// @Failure 404 {object} object{code=int,message=string}
// @Router /api/v3/columns/{id} [get]
func (h *ColumnHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 获取当前用户ID（如果已登录）
	currentUserID, _ := c.Get("user_id")
	userID := ""
	if currentUserID != nil {
		userID, _ = currentUserID.(string)
	}

	detail, err := h.columnService.GetByID(id, userID)
	if err != nil {
		response.NotFound(c, "专栏不存在")
		return
	}

	response.Success(c, detail)
}

// GetUserColumns 获取用户的专栏列表
// @Summary 获取用户的专栏列表
// @Description 获取指定用户创建的所有专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path string true "用户ID(UUID)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} object{code=int,data=object{columns=[]model.ArticleColumn,total=int}}
// @Failure 400 {object} object{code=int,message=string}
// @Router /api/v3/user/{id}/columns [get]
func (h *ColumnHandler) GetUserColumns(c *gin.Context) {
	userID := c.Param("id") // 使用 id 与其他用户路由保持一致
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	columns, total, err := h.columnService.GetByUserID(userID, page, pageSize)
	if err != nil {
		response.ServerError(c, "获取用户专栏列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"columns":   columns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetHotColumns 获取热门专栏
// @Summary 获取热门专栏
// @Description 按订阅数获取热门专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} object{code=int,data=[]model.ArticleColumn}
// @Router /api/v3/columns/hot [get]
func (h *ColumnHandler) GetHotColumns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	columns, err := h.columnService.GetHotColumns(limit)
	if err != nil {
		response.ServerError(c, "获取热门专栏失败: "+err.Error())
		return
	}

	response.Success(c, columns)
}

// Create 创建专栏
// @Summary 创建专栏
// @Description 创建新的专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param request body service.CreateColumnRequest true "专栏信息"
// @Success 200 {object} object{code=int,data=model.ArticleColumn}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns [post]
func (h *ColumnHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	var req service.CreateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	column, err := h.columnService.Create(userID.(string), &req)
	if err != nil {
		response.ServerError(c, "创建专栏失败: "+err.Error())
		return
	}

	response.Success(c, column)
}

// Update 更新专栏
// @Summary 更新专栏
// @Description 更新专栏信息
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Param request body service.UpdateColumnRequest true "专栏信息"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id} [put]
func (h *ColumnHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var req service.UpdateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.Update(userID.(string), id, &req); err != nil {
		response.ServerError(c, "更新专栏失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除专栏
// @Summary 删除专栏
// @Description 删除指定专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id} [delete]
func (h *ColumnHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.Delete(userID.(string), id); err != nil {
		response.ServerError(c, "删除专栏失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// Subscribe 订阅专栏
// @Summary 订阅专栏
// @Description 订阅指定专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id}/subscribe [post]
func (h *ColumnHandler) Subscribe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.Subscribe(userID.(string), id); err != nil {
		response.ServerError(c, "订阅失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// Unsubscribe 取消订阅专栏
// @Summary 取消订阅专栏
// @Description 取消订阅指定专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id}/subscribe [delete]
func (h *ColumnHandler) Unsubscribe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.Unsubscribe(userID.(string), id); err != nil {
		response.ServerError(c, "取消订阅失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetSubscribedColumns 获取用户订阅的专栏
// @Summary 获取用户订阅的专栏
// @Description 获取当前用户订阅的所有专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} object{code=int,data=object{columns=[]model.ArticleColumn,total=int}}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/subscribed [get]
func (h *ColumnHandler) GetSubscribedColumns(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	columns, total, err := h.columnService.GetSubscribedColumns(userID.(string), page, pageSize)
	if err != nil {
		response.ServerError(c, "获取订阅列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"columns":   columns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetColumnArticles 获取专栏的文章列表
// @Summary 获取专栏文章列表
// @Description 获取专栏中的所有文章
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} object{code=int,data=object{articles=[]model.ArticleV3,total=int}}
// @Failure 400 {object} object{code=int,message=string}
// @Router /api/v3/columns/{id}/articles [get]
func (h *ColumnHandler) GetColumnArticles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	articles, total, err := h.columnService.GetColumnArticles(id, page, pageSize)
	if err != nil {
		response.ServerError(c, "获取专栏文章失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"articles":  articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddArticle 添加文章到专栏
// @Summary 添加文章到专栏
// @Description 将文章添加到指定专栏
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Param request body object{article_id=int,sort_order=int} true "文章ID和排序"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id}/articles [post]
func (h *ColumnHandler) AddArticle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	idStr := c.Param("id")
	columnID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var req struct {
		ArticleID uint64 `json:"article_id" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.AddArticle(userID.(string), columnID, req.ArticleID, req.SortOrder); err != nil {
		response.ServerError(c, "添加文章失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveArticle 从专栏移除文章
// @Summary 从专栏移除文章
// @Description 从指定专栏移除文章
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Param articleId path int true "文章ID"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id}/articles/{articleId} [delete]
func (h *ColumnHandler) RemoveArticle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	columnIDStr := c.Param("id")
	columnID, err := strconv.ParseUint(columnIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	articleIDStr := c.Param("articleId") // 使用驼峰命名与其他路由保持一致
	articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.RemoveArticle(userID.(string), columnID, articleID); err != nil {
		response.ServerError(c, "移除文章失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateArticlePosition 更新文章位置
// @Summary 更新文章在专栏中的位置
// @Description 调整文章在专栏中的排序
// @Tags 专栏模块
// @Accept json
// @Produce json
// @Param id path int true "专栏ID"
// @Param articleId path int true "文章ID"
// @Param request body object{sort_order=int} true "新的排序位置"
// @Success 200 {object} object{code=int,message=string}
// @Failure 400 {object} object{code=int,message=string}
// @Security Bearer
// @Router /api/v3/columns/{id}/articles/{articleId}/position [put]
func (h *ColumnHandler) UpdateArticlePosition(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	columnIDStr := c.Param("id")
	columnID, err := strconv.ParseUint(columnIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	articleIDStr := c.Param("articleId") // 使用驼峰命名与其他路由保持一致
	articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var req struct {
		SortOrder int `json:"sort_order" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.columnService.UpdateArticlePosition(userID.(string), columnID, articleID, req.SortOrder); err != nil {
		response.ServerError(c, "更新文章位置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}
