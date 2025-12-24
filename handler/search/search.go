package search

import (
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService service.SearchServiceV2
}

func NewSearchHandler(searchService service.SearchServiceV2) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchArticles 搜索文章
func (h *SearchHandler) SearchArticles(c *gin.Context) {
	keyword := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if keyword == "" {
		util.BadRequest(c, "搜索关键词不能为空")
		return
	}

	// 搜索文章
	articles, total, err := h.searchService.SearchArticles(keyword, page, pageSize)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     articles,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"keyword":  keyword,
	})
}

// SearchUsers 搜索用户
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if keyword == "" {
		util.BadRequest(c, "搜索关键词不能为空")
		return
	}

	// 获取当前登录用户ID（如果已登录）
	var currentUserID string
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(string)
	}

	// 搜索用户
	users, total, err := h.searchService.SearchUsers(keyword, page, pageSize, currentUserID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"keyword":  keyword,
	})
}

// SearchAll 综合搜索
func (h *SearchHandler) SearchAll(c *gin.Context) {
	keyword := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if keyword == "" {
		util.BadRequest(c, "搜索关键词不能为空")
		return
	}

	// 获取当前登录用户ID（如果已登录）
	var currentUserID string
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(string)
	}

	// 综合搜索
	result, err := h.searchService.SearchAll(keyword, page, pageSize, currentUserID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, result)
}
