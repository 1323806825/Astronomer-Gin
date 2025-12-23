package trending

import (
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TrendingHandler struct {
	trendingService service.TrendingServiceV2
}

func NewTrendingHandler(trendingService service.TrendingServiceV2) *TrendingHandler {
	return &TrendingHandler{
		trendingService: trendingService,
	}
}

// GetTrendingArticles 获取热门文章榜单
func (h *TrendingHandler) GetTrendingArticles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// 获取热门文章
	articles, err := h.trendingService.GetTrendingArticles(limit)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":  articles,
		"total": len(articles),
		"limit": limit,
	})
}

// GetTrendingUsers 获取热门用户榜单
func (h *TrendingHandler) GetTrendingUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// 获取热门用户
	users, err := h.trendingService.GetTrendingUsers(limit)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":  users,
		"total": len(users),
		"limit": limit,
	})
}

// RefreshTrending 手动刷新热门榜单（管理员接口）
func (h *TrendingHandler) RefreshTrending(c *gin.Context) {
	// 刷新热门数据
	err := h.trendingService.RefreshTrendingData()
	if err != nil {
		util.InternalServerError(c, "刷新失败："+err.Error())
		return
	}

	util.SuccessWithMessage(c, "热门榜单刷新成功", nil)
}
