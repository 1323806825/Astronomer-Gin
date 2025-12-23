package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/database"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// BlogServiceV2 企业级博客服务接口
type BlogServiceV2 interface {
	// 文章CRUD
	CreateArticle(userID string, title, preface, content, photo, tag string) (uint64, error)
	UpdateArticle(articleID uint64, userID string, updates map[string]interface{}) error
	DeleteArticle(articleID uint64, userID string) error
	GetArticleDetail(articleID uint64) (*model.Article, error)
	GetArticleList(page, pageSize int, tag string) ([]model.Article, int64, error)

	// 草稿功能
	SaveDraft(userID string, title, preface, content, photo, tag string) (uint64, error)
	UpdateDraft(draftID uint64, userID string, updates map[string]interface{}) error
	GetUserDrafts(userID string, page, pageSize int) ([]model.Article, int64, error)
	PublishDraft(draftID uint64, userID string) error

	// 点赞功能
	LikeArticle(articleID uint64, userID string, username string, notifyRepo repository.NotificationRepository) error
	UnlikeArticle(articleID uint64, userID string) error
	IsLiked(articleID uint64, userID string) bool

	// 缓存管理
	RefreshArticleCache(articleID uint64) error
	ClearArticleCache(articleID uint64) error
	ClearArticleListCache() error
}

type blogServiceV2 struct {
	blogRepo    repository.BlogRepository
	cacheHelper *util.CacheHelper
	db          *gorm.DB
}

func NewBlogServiceV2(blogRepo repository.BlogRepository) BlogServiceV2 {
	return &blogServiceV2{
		blogRepo:    blogRepo,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
		db:          database.GetDB(),
	}
}

// CreateArticle 创建文章（企业级实现）
func (s *blogServiceV2) CreateArticle(userID string, title, preface, content, photo, tag string) (uint64, error) {
	// 1. 参数验证
	if err := util.ValidateTitle(title); err != nil {
		return 0, err
	}
	if err := util.ValidateContent(content); err != nil {
		return 0, err
	}

	// 2. 敏感词检查
	if util.ContainsSensitiveWord(title) || util.ContainsSensitiveWord(content) {
		return 0, constant.ErrContentHasSensitiveWord
	}

	// 3. 创建文章
	now := time.Now()
	article := &model.Article{
		UserID:        userID,
		Title:         title,
		Preface:       preface,
		Content:       content,
		Photo:         photo,
		Tag:           tag,
		Status:        model.ArticleStatusPublished,
		CreateTime:    now,
		UpdateTime:    now,
		Visit:         0,
		GoodCount:     0,
		Appear:        true,
		Comment:       true,
		CommentCount:  0,
		FavoriteCount: 0,
	}

	if err := s.blogRepo.Create(article); err != nil {
		return 0, constant.ErrCreateArticleFailed
	}

	// 4. 清除文章列表缓存
	s.ClearArticleListCache()

	return article.ID, nil
}

// UpdateArticle 更新文章（企业级实现）
func (s *blogServiceV2) UpdateArticle(articleID uint64, userID string, updates map[string]interface{}) error {
	// 1. 检查文章所有权
	if !s.blogRepo.CheckOwnership(articleID, userID) {
		return constant.ErrNotArticleOwner
	}

	// 2. 参数验证
	if title, ok := updates["title"].(string); ok && title != "" {
		if err := util.ValidateTitle(title); err != nil {
			return err
		}
		if util.ContainsSensitiveWord(title) {
			return constant.ErrContentHasSensitiveWord
		}
	}

	if content, ok := updates["content"].(string); ok && content != "" {
		if err := util.ValidateContent(content); err != nil {
			return err
		}
		if util.ContainsSensitiveWord(content) {
			return constant.ErrContentHasSensitiveWord
		}
	}

	// 3. 添加更新时间
	updates["update_time"] = time.Now()

	// 4. 更新文章
	if err := s.blogRepo.UpdateFields(articleID, updates); err != nil {
		return constant.ErrUpdateArticleFailed
	}

	// 5. 清除缓存
	s.ClearArticleCache(articleID)
	s.ClearArticleListCache()

	return nil
}

// DeleteArticle 删除文章（软删除）
func (s *blogServiceV2) DeleteArticle(articleID uint64, userID string) error {
	// 1. 检查文章所有权
	if !s.blogRepo.CheckOwnership(articleID, userID) {
		return constant.ErrNotArticleOwner
	}

	// 2. 软删除（更新状态为已删除）
	updates := map[string]interface{}{
		"status":      model.ArticleStatusDeleted,
		"update_time": time.Now(),
	}

	if err := s.blogRepo.UpdateFields(articleID, updates); err != nil {
		return constant.ErrDeleteArticleFailed
	}

	// 3. 清除缓存
	s.ClearArticleCache(articleID)
	s.ClearArticleListCache()

	return nil
}

// GetArticleDetail 获取文章详情（带缓存）
func (s *blogServiceV2) GetArticleDetail(articleID uint64) (*model.Article, error) {
	// 1. 从缓存获取
	cacheKey := fmt.Sprintf("%s%d", constant.CacheKeyArticle, articleID)
	var article model.Article

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&article,
		time.Duration(constant.CacheExpireMedium)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			art, err := s.blogRepo.FindByID(articleID)
			if err != nil {
				return nil, err
			}

			// 检查文章状态
			if art.Status == model.ArticleStatusDeleted {
				return nil, constant.ErrArticleDeleted
			}

			return art, nil
		},
	)

	if err != nil {
		return nil, constant.ErrArticleNotFound
	}

	// 2. 异步增加访问量（不阻塞响应）
	go s.blogRepo.IncrementVisit(articleID)

	return &article, nil
}

// GetArticleList 获取文章列表（带缓存）
func (s *blogServiceV2) GetArticleList(page, pageSize int, tag string) ([]model.Article, int64, error) {
	// 1. 参数验证和默认值
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 2. 构建缓存键
	cacheKey := fmt.Sprintf("%spage:%d:size:%d:tag:%s", constant.CacheKeyArticleList, page, pageSize, tag)

	// 3. 尝试从缓存获取
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
			// 从数据库查询
			articles, total, err := s.blogRepo.FindList(page, pageSize, tag, model.ArticleStatusPublished)
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

// SaveDraft 保存草稿
func (s *blogServiceV2) SaveDraft(userID string, title, preface, content, photo, tag string) (uint64, error) {
	// 1. 标题至少需要填写
	if err := util.ValidateTitle(title); err != nil {
		return 0, err
	}

	// 2. 创建草稿
	now := time.Now()
	draft := &model.Article{
		UserID:     userID,
		Title:      title,
		Preface:    preface,
		Content:    content,
		Photo:      photo,
		Tag:        tag,
		Status:     model.ArticleStatusDraft,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := s.blogRepo.Create(draft); err != nil {
		return 0, constant.ErrSaveDraftFailed
	}

	return draft.ID, nil
}

// UpdateDraft 更新草稿
func (s *blogServiceV2) UpdateDraft(draftID uint64, userID string, updates map[string]interface{}) error {
	// 1. 检查所有权
	if !s.blogRepo.CheckOwnership(draftID, userID) {
		return constant.ErrNotArticleOwner
	}

	// 2. 检查是否是草稿状态
	article, err := s.blogRepo.FindByID(draftID)
	if err != nil {
		return constant.ErrDraftNotFound
	}

	if article.Status != model.ArticleStatusDraft {
		return constant.ErrDraftNotFound
	}

	// 3. 更新草稿
	updates["update_time"] = time.Now()
	if err := s.blogRepo.UpdateFields(draftID, updates); err != nil {
		return constant.ErrSaveDraftFailed
	}

	return nil
}

// GetUserDrafts 获取用户草稿列表
func (s *blogServiceV2) GetUserDrafts(userID string, page, pageSize int) ([]model.Article, int64, error) {
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	drafts, total, err := s.blogRepo.FindUserDrafts(userID, page, pageSize)
	if err != nil {
		return nil, 0, constant.ErrDatabaseQuery
	}

	return drafts, total, nil
}

// PublishDraft 发布草稿（使用事务）
func (s *blogServiceV2) PublishDraft(draftID uint64, userID string) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询草稿
		var draft model.Article
		if err := tx.Where("id = ? AND user_id = ? AND status = ?", draftID, userID, model.ArticleStatusDraft).
			First(&draft).Error; err != nil {
			return constant.ErrDraftNotFound
		}

		// 2. 验证内容
		if err := util.ValidateContent(draft.Content); err != nil {
			return err
		}

		// 3. 敏感词检查
		if util.ContainsSensitiveWord(draft.Title) || util.ContainsSensitiveWord(draft.Content) {
			return constant.ErrContentHasSensitiveWord
		}

		// 4. 更新状态为已发布
		if err := tx.Model(&draft).Updates(map[string]interface{}{
			"status":      model.ArticleStatusPublished,
			"update_time": time.Now(),
		}).Error; err != nil {
			return constant.ErrPublishDraftFailed
		}

		// 5. 清除缓存（事务提交后才会生效）
		s.ClearArticleListCache()

		return nil
	})
}

// LikeArticle 点赞文章（使用事务）
func (s *blogServiceV2) LikeArticle(articleID uint64, userID, username string, notifyRepo repository.NotificationRepository) error {
	// 1. 检查是否已点赞
	if s.blogRepo.IsLiked(articleID, userID) {
		return constant.ErrAlreadyLiked
	}

	// 2. 获取文章信息
	article, err := s.blogRepo.FindByID(articleID)
	if err != nil {
		return constant.ErrArticleNotFound
	}

	// 3. 使用事务执行点赞
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 3.1 创建点赞记录
		like := &model.ArticleStar{
			UserID:    userID,
			ArticleID: uint(articleID),
		}
		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// 3.2 增加点赞数
		if err := tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("good_count", gorm.Expr("good_count + ?", 1)).Error; err != nil {
			return err
		}

		// 3.3 如果不是自己的文章，创建通知
		if userID != article.UserID {
			notification := &model.Notification{
				UserID:       article.UserID,
				Type:         model.NotificationTypeLikeArticle,
				FromUserID:   userID,
				FromUsername: username,
				Content:      username + " 赞了你的文章",
				RelatedID:    articleID,
				RelatedType:  "article",
				IsRead:       false,
				CreateTime:   time.Now(),
			}
			if err := tx.Create(notification).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return constant.ErrLikeFailed
	}

	// 4. 清除文章缓存
	s.ClearArticleCache(articleID)

	return nil
}

// UnlikeArticle 取消点赞文章（使用事务）
func (s *blogServiceV2) UnlikeArticle(articleID uint64, userID string) error {
	// 1. 检查是否已点赞
	if !s.blogRepo.IsLiked(articleID, userID) {
		return constant.ErrNotLiked
	}

	// 2. 使用事务执行取消点赞
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 2.1 删除点赞记录
		if err := tx.Where("article_id = ? AND user_id = ?", articleID, userID).
			Delete(&model.ArticleStar{}).Error; err != nil {
			return err
		}

		// 2.2 减少点赞数
		if err := tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("good_count", gorm.Expr("good_count - ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return constant.ErrLikeFailed
	}

	// 3. 清除文章缓存
	s.ClearArticleCache(articleID)

	return nil
}

// IsLiked 检查是否已点赞
func (s *blogServiceV2) IsLiked(articleID uint64, userID string) bool {
	return s.blogRepo.IsLiked(articleID, userID)
}

// ==================== 缓存管理 ====================

// RefreshArticleCache 刷新文章缓存
func (s *blogServiceV2) RefreshArticleCache(articleID uint64) error {
	article, err := s.blogRepo.FindByID(articleID)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s%d", constant.CacheKeyArticle, articleID)
	return s.cacheHelper.Set(cacheKey, article, time.Duration(constant.CacheExpireMedium)*time.Second)
}

// ClearArticleCache 清除文章缓存
func (s *blogServiceV2) ClearArticleCache(articleID uint64) error {
	cacheKey := fmt.Sprintf("%s%d", constant.CacheKeyArticle, articleID)
	return s.cacheHelper.Delete(cacheKey)
}

// ClearArticleListCache 清除文章列表缓存
func (s *blogServiceV2) ClearArticleListCache() error {
	// 删除所有文章列表缓存
	pattern := constant.CacheKeyArticleList + "*"
	return s.cacheHelper.DeleteByPattern(pattern)
}
