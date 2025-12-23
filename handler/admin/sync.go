package admin

import (
	"astronomer-gin/pkg/elasticsearch"
	"astronomer-gin/repository"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SyncHandler 数据同步处理器
type SyncHandler struct {
	blogRepo repository.BlogRepository
}

// NewSyncHandler 创建数据同步处理器
func NewSyncHandler(blogRepo repository.BlogRepository) *SyncHandler {
	return &SyncHandler{
		blogRepo: blogRepo,
	}
}

// SyncArticlesToES 同步文章到ElasticSearch
func (h *SyncHandler) SyncArticlesToES(c *gin.Context) {
	// 检查ES是否启用
	if !elasticsearch.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    http.StatusServiceUnavailable,
			"message": "ElasticSearch未启用",
		})
		return
	}

	// 批量获取所有已发布文章
	page := 1
	pageSize := 100 // 每次批量处理100篇
	totalSynced := 0

	for {
		// 获取一批文章（status=1表示已发布）
		articles, total, err := h.blogRepo.FindList(page, pageSize, "", 1)
		if err != nil {
			log.Printf("❌ 获取文章失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "获取文章失败: " + err.Error(),
			})
			return
		}

		if len(articles) == 0 {
			break
		}

		// 批量索引到ES
		if err := elasticsearch.BulkIndexArticles(articles); err != nil {
			log.Printf("❌ 批量索引失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "批量索引失败: " + err.Error(),
			})
			return
		}

		totalSynced += len(articles)
		log.Printf("✅ 已同步 %d/%d 篇文章", totalSynced, total)

		// 如果已经处理完所有文章，退出循环
		if int64(totalSynced) >= total {
			break
		}

		page++
	}

	log.Printf("🎉 数据同步完成！共同步 %d 篇文章", totalSynced)
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "数据同步成功",
		"data": gin.H{
			"total": totalSynced,
		},
	})
}
