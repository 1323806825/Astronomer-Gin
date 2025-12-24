package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/elasticsearch"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"fmt"
	"log"
	"time"
)

// SearchServiceV2 企业级搜索服务接口
type SearchServiceV2 interface {
	// 搜索文章
	SearchArticles(keyword string, page, pageSize int) ([]model.Article, int64, error)

	// 搜索用户（带关注状态）
	SearchUsers(keyword string, page, pageSize int, currentUserID string) ([]model.User, int64, error)

	// 综合搜索
	SearchAll(keyword string, page, pageSize int, currentUserID string) (map[string]interface{}, error)
}

type searchServiceV2 struct {
	blogRepo    repository.BlogRepository
	userRepo    repository.UserRepository
	followRepo  repository.FollowRepository
	cacheHelper *util.CacheHelper
}

// NewSearchServiceV2 创建搜索服务V2实例
func NewSearchServiceV2(blogRepo repository.BlogRepository, userRepo repository.UserRepository, followRepo repository.FollowRepository) SearchServiceV2 {
	return &searchServiceV2{
		blogRepo:    blogRepo,
		userRepo:    userRepo,
		followRepo:  followRepo,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
	}
}

// SearchArticles 搜索文章（ES优先，降级到MySQL）
func (s *searchServiceV2) SearchArticles(keyword string, page, pageSize int) ([]model.Article, int64, error) {
	// 1. 参数验证
	if keyword == "" {
		return nil, 0, constant.ErrParamInvalid
	}

	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 优先使用ElasticSearch
	if elasticsearch.IsEnabled() {
		esArticles, total, err := elasticsearch.SearchArticles(keyword, page, pageSize)
		if err != nil {
			// ES失败，降级到MySQL
			log.Printf("⚠️  ES搜索失败，降级到MySQL: %v", err)
			return s.searchArticlesFromMySQL(keyword, page, pageSize)
		}

		// 将ES文档转换为Article模型
		if len(esArticles) == 0 {
			return []model.Article{}, 0, nil
		}

		// 批量查询完整文章信息
		var articleIDs []uint64
		for _, doc := range esArticles {
			articleIDs = append(articleIDs, doc.ID)
		}

		articles, err := s.blogRepo.FindByIDs(articleIDs)
		if err != nil {
			return nil, 0, err
		}

		// 按ES返回的顺序排序
		sortedArticles := make([]model.Article, 0, len(articleIDs))
		articleMap := make(map[uint64]model.Article)
		for _, article := range articles {
			articleMap[article.ID] = article
		}
		for _, id := range articleIDs {
			if article, ok := articleMap[id]; ok {
				sortedArticles = append(sortedArticles, article)
			}
		}

		log.Printf("✅ ES搜索成功: 关键词=%s, 结果数=%d", keyword, len(sortedArticles))
		return sortedArticles, total, nil
	}

	// 3. ES未启用，直接使用MySQL
	log.Printf("ℹ️  ES未启用，使用MySQL搜索")
	return s.searchArticlesFromMySQL(keyword, page, pageSize)
}

// searchArticlesFromMySQL MySQL搜索（降级方案）
func (s *searchServiceV2) searchArticlesFromMySQL(keyword string, page, pageSize int) ([]model.Article, int64, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("search:mysql:article:%s:page:%d:size:%d", keyword, page, pageSize)

	// 尝试从缓存获取
	type CachedData struct {
		Articles []model.Article
		Total    int64
	}

	var cached CachedData
	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&cached,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从MySQL搜索
			articles, total, err := s.blogRepo.SearchArticles(keyword, page, pageSize)
			if err != nil {
				return nil, err
			}
			return CachedData{Articles: articles, Total: total}, nil
		},
	)

	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return cached.Articles, cached.Total, nil
}

// SearchUsers 搜索用户（带缓存和关注状态）
func (s *searchServiceV2) SearchUsers(keyword string, page, pageSize int, currentUserID string) ([]model.User, int64, error) {
	// 1. 参数验证
	if keyword == "" {
		return nil, 0, constant.ErrParamInvalid
	}

	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 从数据库搜索（不使用缓存，因为关注状态因人而异）
	users, total, err := s.userRepo.SearchUsers(keyword, page, pageSize)
	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	// 3. 数据脱敏
	for i := range users {
		users[i].Phone = util.MaskPhone(users[i].Phone)
	}

	// 4. 如果用户已登录，检查关注状态
	if currentUserID != "" {
		for i := range users {
			// 跳过自己
			if users[i].ID == currentUserID {
				users[i].IsFollowed = false
				continue
			}
			// 检查是否关注
			users[i].IsFollowed = s.followRepo.IsFollowing(currentUserID, users[i].ID)
		}
	}

	return users, total, nil
}

// SearchAll 综合搜索（文章+用户）
func (s *searchServiceV2) SearchAll(keyword string, page, pageSize int, currentUserID string) (map[string]interface{}, error) {
	// 1. 参数验证
	if keyword == "" {
		return nil, constant.ErrParamInvalid
	}

	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 并发搜索文章和用户
	type ArticleResult struct {
		Articles []model.Article
		Total    int64
		Err      error
	}

	type UserResult struct {
		Users []model.User
		Total int64
		Err   error
	}

	articleChan := make(chan ArticleResult)
	userChan := make(chan UserResult)

	// 搜索文章
	go func() {
		articles, total, err := s.SearchArticles(keyword, page, pageSize)
		articleChan <- ArticleResult{Articles: articles, Total: total, Err: err}
	}()

	// 搜索用户
	go func() {
		users, total, err := s.SearchUsers(keyword, page, pageSize, currentUserID)
		userChan <- UserResult{Users: users, Total: total, Err: err}
	}()

	// 等待结果
	articleRes := <-articleChan
	userRes := <-userChan

	// 3. 组合结果
	result := map[string]interface{}{
		"articles": map[string]interface{}{
			"list":  articleRes.Articles,
			"total": articleRes.Total,
		},
		"users": map[string]interface{}{
			"list":  userRes.Users,
			"total": userRes.Total,
		},
		"page":     page,
		"pageSize": pageSize,
		"keyword":  keyword,
	}

	return result, nil
}
