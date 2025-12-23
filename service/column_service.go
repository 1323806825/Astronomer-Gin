package service

import (
	"astronomer-gin/model"
	"astronomer-gin/repository"
	"fmt"
	"log"
)

// ColumnService 专栏服务接口
type ColumnService interface {
	// GetList 获取专栏列表
	GetList(page, pageSize int) ([]*model.ArticleColumn, int64, error)

	// GetByID 获取专栏详情
	GetByID(id uint64, currentUserID string) (*ColumnDetailResponse, error)

	// GetByUserID 获取用户的专栏列表
	GetByUserID(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error)

	// GetHotColumns 获取热门专栏
	GetHotColumns(limit int) ([]*model.ArticleColumn, error)

	// Create 创建专栏
	Create(userID string, req *CreateColumnRequest) (*model.ArticleColumn, error)

	// Update 更新专栏
	Update(userID string, columnID uint64, req *UpdateColumnRequest) error

	// Delete 删除专栏
	Delete(userID string, columnID uint64) error

	// Subscribe 订阅专栏
	Subscribe(userID string, columnID uint64) error

	// Unsubscribe 取消订阅
	Unsubscribe(userID string, columnID uint64) error

	// GetSubscribedColumns 获取用户订阅的专栏
	GetSubscribedColumns(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error)

	// AddArticle 添加文章到专栏
	AddArticle(userID string, columnID, articleID uint64, sortOrder int) error

	// RemoveArticle 从专栏移除文章
	RemoveArticle(userID string, columnID, articleID uint64) error

	// GetColumnArticles 获取专栏的文章列表
	GetColumnArticles(columnID uint64, page, pageSize int) ([]*model.ArticleV3, int64, error)

	// UpdateArticlePosition 更新文章在专栏中的位置
	UpdateArticlePosition(userID string, columnID, articleID uint64, sortOrder int) error
}

type columnService struct {
	columnRepo       repository.ColumnRepository
	userRepo         repository.UserRepository
	notificationRepo repository.NotificationRepository
	articleRepo      repository.ArticleV3Repository
}

// NewColumnService 创建专栏服务实例
func NewColumnService(
	columnRepo repository.ColumnRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	articleRepo repository.ArticleV3Repository,
) ColumnService {
	return &columnService{
		columnRepo:       columnRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		articleRepo:      articleRepo,
	}
}

// GetList 获取专栏列表
func (s *columnService) GetList(page, pageSize int) ([]*model.ArticleColumn, int64, error) {
	return s.columnRepo.GetList(page, pageSize)
}

// GetByID 获取专栏详情
func (s *columnService) GetByID(id uint64, currentUserID string) (*ColumnDetailResponse, error) {
	// 获取专栏基本信息
	column, err := s.columnRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 获取作者信息
	author, err := s.userRepo.FindByID(column.UserID)
	if err != nil {
		log.Printf("获取专栏作者信息失败", "error", err, "user_id", column.UserID)
		// 作者信息获取失败不影响专栏详情返回
		author = &model.User{
			ID:       column.UserID,
			Username: "未知用户",
		}
	}

	// 检查当前用户是否订阅
	isSubscribed := false
	if currentUserID != "" {
		isSubscribed, _ = s.columnRepo.CheckSubscribed(currentUserID, id)
	}

	return &ColumnDetailResponse{
		Column: column,
		Author: &ArticleAuthorInfo{
			UserID:   author.ID,
			Username: author.Username,
			Avatar:   author.Icon,
			Intro:    author.Intro,
		},
		ArticleCount: int(column.ArticleCount),
		IsSubscribed: isSubscribed,
	}, nil
}

// GetByUserID 获取用户的专栏列表
func (s *columnService) GetByUserID(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error) {
	return s.columnRepo.GetByUserID(userID, page, pageSize)
}

// GetHotColumns 获取热门专栏
func (s *columnService) GetHotColumns(limit int) ([]*model.ArticleColumn, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	return s.columnRepo.GetHotColumns(limit)
}

// Create 创建专栏
func (s *columnService) Create(userID string, req *CreateColumnRequest) (*model.ArticleColumn, error) {
	// 验证用户是否存在
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户专栏数量限制（企业级功能：限制每个用户最多创建的专栏数）
	columns, _, err := s.columnRepo.GetByUserID(userID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("检查用户专栏数量失败: %w", err)
	}

	// 普通用户最多创建10个专栏，VIP用户可以创建更多
	maxColumns := 10
	if user.Role == "vip" || user.Role == "admin" || user.Role == "super_admin" {
		maxColumns = 50
	}

	if len(columns) >= maxColumns {
		return nil, fmt.Errorf("专栏数量已达上限（%d个）", maxColumns)
	}

	// 创建专栏
	column := &model.ArticleColumn{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		SortType:    req.SortType,
		Status:      1,
	}

	if column.SortType == 0 {
		column.SortType = 1 // 默认自定义排序
	}

	if err := s.columnRepo.Create(column); err != nil {
		return nil, err
	}

	log.Println("创建专栏成功", "column_id", column.ID, "user_id", userID, "name", req.Name)

	return column, nil
}

// Update 更新专栏
func (s *columnService) Update(userID string, columnID uint64, req *UpdateColumnRequest) error {
	// 获取专栏
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return err
	}

	// 权限检查：只有作者可以更新专栏
	if column.UserID != userID {
		return fmt.Errorf("无权限更新此专栏")
	}

	// 更新字段
	if req.Name != nil && *req.Name != "" {
		column.Name = *req.Name
	}
	if req.Description != nil {
		column.Description = *req.Description
	}
	if req.CoverImage != nil {
		column.CoverImage = *req.CoverImage
	}
	if req.SortType != nil && *req.SortType > 0 {
		column.SortType = *req.SortType
	}
	if req.IsFinished != nil {
		column.IsFinished = *req.IsFinished
	}

	if err := s.columnRepo.Update(column); err != nil {
		return err
	}

	log.Println("更新专栏成功", "column_id", columnID, "user_id", userID)

	return nil
}

// Delete 删除专栏
func (s *columnService) Delete(userID string, columnID uint64) error {
	// 获取专栏
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return err
	}

	// 权限检查：只有作者可以删除专栏
	if column.UserID != userID {
		return fmt.Errorf("无权限删除此专栏")
	}

	// 删除专栏（软删除）
	if err := s.columnRepo.Delete(columnID); err != nil {
		return err
	}

	log.Println("删除专栏成功", "column_id", columnID, "user_id", userID)

	return nil
}

// Subscribe 订阅专栏
func (s *columnService) Subscribe(userID string, columnID uint64) error {
	// 获取专栏信息
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return err
	}

	// 不能订阅自己的专栏
	if column.UserID == userID {
		return fmt.Errorf("不能订阅自己的专栏")
	}

	// 订阅专栏
	if err := s.columnRepo.Subscribe(userID, columnID); err != nil {
		return err
	}

	// 发送通知给专栏作者（企业级功能：整合通知系统）
	go func() {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			log.Printf("获取订阅者信息失败", "error", err, "user_id", userID)
			return
		}

		notification := &model.Notification{
			UserID:       column.UserID,
			Type:         6, // 订阅专栏通知（需要在常量中定义）
			FromUserID:   userID,
			FromUsername: user.Username,
			Content:      fmt.Sprintf("%s 订阅了你的专栏《%s》", user.Username, column.Name),
			RelatedID:    columnID,
			RelatedType:  "column",
		}

		if err := s.notificationRepo.Create(notification); err != nil {
			log.Printf("创建订阅通知失败", "error", err)
		}
	}()

	log.Println("订阅专栏成功", "column_id", columnID, "user_id", userID)

	return nil
}

// Unsubscribe 取消订阅
func (s *columnService) Unsubscribe(userID string, columnID uint64) error {
	if err := s.columnRepo.Unsubscribe(userID, columnID); err != nil {
		return err
	}

	log.Println("取消订阅专栏成功", "column_id", columnID, "user_id", userID)

	return nil
}

// GetSubscribedColumns 获取用户订阅的专栏
func (s *columnService) GetSubscribedColumns(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error) {
	return s.columnRepo.GetSubscribedColumns(userID, page, pageSize)
}

// AddArticle 添加文章到专栏
func (s *columnService) AddArticle(userID string, columnID, articleID uint64, sortOrder int) error {
	// 获取专栏
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return err
	}

	// 权限检查：只有作者可以添加文章到专栏
	if column.UserID != userID {
		return fmt.Errorf("无权限向此专栏添加文章")
	}

	// 获取文章信息
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return fmt.Errorf("文章不存在")
	}

	// 检查文章作者是否为专栏作者
	if article.UserID != userID {
		return fmt.Errorf("只能将自己的文章添加到专栏")
	}

	// 添加文章到专栏
	if err := s.columnRepo.AddArticle(columnID, articleID, sortOrder); err != nil {
		return err
	}

	// 通知专栏订阅者（企业级功能：整合通知系统）
	go func() {
		subscribers, _, err := s.columnRepo.GetSubscribers(columnID, 1, 100)
		if err != nil {
			log.Printf("获取专栏订阅者失败", "error", err)
			return
		}

		for _, subscriberID := range subscribers {
			notification := &model.Notification{
				UserID:       subscriberID,
				Type:         7, // 专栏更新通知（需要在常量中定义）
				FromUserID:   userID,
				FromUsername: column.Name,
				Content:      fmt.Sprintf("专栏《%s》更新了新文章：%s", column.Name, article.Title),
				RelatedID:    articleID,
				RelatedType:  "article",
			}

			if err := s.notificationRepo.Create(notification); err != nil {
				log.Printf("创建专栏更新通知失败", "error", err, "subscriber_id", subscriberID)
			}
		}

		log.Println("已通知专栏订阅者", "column_id", columnID, "subscriber_count", len(subscribers))
	}()

	log.Println("添加文章到专栏成功", "column_id", columnID, "article_id", articleID, "user_id", userID)

	return nil
}

// RemoveArticle 从专栏移除文章
func (s *columnService) RemoveArticle(userID string, columnID, articleID uint64) error {
	// 获取专栏
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return err
	}

	// 权限检查：只有作者可以移除文章
	if column.UserID != userID {
		return fmt.Errorf("无权限从此专栏移除文章")
	}

	// 移除文章
	if err := s.columnRepo.RemoveArticle(columnID, articleID); err != nil {
		return err
	}

	log.Println("从专栏移除文章成功", "column_id", columnID, "article_id", articleID, "user_id", userID)

	return nil
}

// GetColumnArticles 获取专栏的文章列表
func (s *columnService) GetColumnArticles(columnID uint64, page, pageSize int) ([]*model.ArticleV3, int64, error) {
	return s.columnRepo.GetColumnArticles(columnID, page, pageSize)
}

// UpdateArticlePosition 更新文章在专栏中的位置
func (s *columnService) UpdateArticlePosition(userID string, columnID, articleID uint64, sortOrder int) error {
	// 获取专栏
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return err
	}

	// 权限检查：只有作者可以调整文章顺序
	if column.UserID != userID {
		return fmt.Errorf("无权限调整此专栏的文章顺序")
	}

	// 更新文章位置
	if err := s.columnRepo.UpdateArticlePosition(columnID, articleID, sortOrder); err != nil {
		return err
	}

	log.Println("更新文章位置成功", "column_id", columnID, "article_id", articleID, "sort_order", sortOrder)

	return nil
}
