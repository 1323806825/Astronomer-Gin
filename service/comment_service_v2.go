package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/database"
	"astronomer-gin/pkg/queue"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"
	"astronomer-gin/repository"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// CommentServiceV2 企业级评论服务接口
type CommentServiceV2 interface {
	// 评论CRUD
	CreateComment(articleID int64, comment, username, phone, commentAddr, userID string) (int64, error)
	GetCommentList(articleID string) ([]model.CommentParent, error)
	CreateSubComment(parentCommentID int64, toUserID string, comment, username, phone, toUsername, toPhone, commentAddr string, userID string) (int64, error)
	GetSubCommentList(parentID string) ([]model.CommentSubTwo, error)
	LikeComment(commentID int64, userID string, commentType string) error

	// 缓存管理
	RefreshCommentCache(articleID string) error
	ClearCommentCache(articleID string) error
}

type commentServiceV2 struct {
	commentRepo repository.CommentRepository
	blogRepo    repository.BlogRepository
	userRepo    repository.UserRepository
	cacheHelper *util.CacheHelper
	db          *gorm.DB
}

func NewCommentServiceV2(commentRepo repository.CommentRepository, blogRepo repository.BlogRepository, userRepo repository.UserRepository) CommentServiceV2 {
	return &commentServiceV2{
		commentRepo: commentRepo,
		blogRepo:    blogRepo,
		userRepo:    userRepo,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
		db:          database.GetDB(),
	}
}

// CreateComment 创建一级评论（企业级实现）
func (s *commentServiceV2) CreateComment(articleID int64, comment, username, phone, commentAddr, userID string) (int64, error) {
	// 1. 参数验证
	if err := util.ValidateContent(comment); err != nil {
		return 0, err
	}

	// 2. 敏感词检查
	if util.ContainsSensitiveWord(comment) {
		return 0, constant.ErrContentHasSensitiveWord
	}

	// 3. 检查文章是否存在
	article, err := s.blogRepo.FindByID(uint64(articleID))
	if err != nil {
		return 0, constant.ErrArticleNotFound
	}

	// 4. 检查文章是否允许评论
	if !article.Comment {
		return 0, constant.ErrArticleCommentDisabled
	}

	var commentID int64

	// 5. 使用事务创建评论并更新文章评论数
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 5.1 创建评论
		commentParent := &model.CommentParent{
			ArticleID:   articleID,
			Comment:     comment,
			CommentTime: time.Now().Format("2006-01-02 15:04:05"),
			Username:    username,
			UserID:      userID,
			Phone:       phone,
			GoodCount:   0,
			CommentAddr: commentAddr,
		}

		if err := tx.Create(commentParent).Error; err != nil {
			return err
		}

		commentID = commentParent.ID

		// 5.2 更新文章评论数
		if err := tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, constant.ErrCreateCommentFailed
	}

	// 6. 异步发送评论通知（不影响主流程）
	if userID != article.UserID && queue.Client != nil {
		go func() {
			task := queue.CreateTask(queue.TaskTypeNotification, map[string]interface{}{
				"notify_type":        "comment",
				"author_id":          article.UserID,
				"commenter_id":       userID,
				"commenter_username": username,
				"article_id":         float64(articleID),
				"comment_id":         float64(commentID),
			})
			ctx := context.Background()
			if err := queue.Client.PublishTask(ctx, task); err != nil {
				log.Printf("Failed to publish comment notification task: %v", err)
			}
		}()
	}

	// 7. 清除评论列表缓存
	s.ClearCommentCache(fmt.Sprintf("%d", articleID))

	return commentID, nil
}

// GetCommentList 获取文章评论列表（带缓存）
func (s *commentServiceV2) GetCommentList(articleID string) ([]model.CommentParent, error) {
	// 1. 从缓存获取
	cacheKey := fmt.Sprintf("%s%s", constant.CacheKeyCommentList, articleID)
	var comments []model.CommentParent

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&comments,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			return s.commentRepo.FindParentCommentsByArticleID(articleID)
		},
	)

	if err != nil {
		return nil, constant.ErrDatabaseQuery
	}

	return comments, nil
}

// CreateSubComment 创建二级评论（企业级实现）
func (s *commentServiceV2) CreateSubComment(parentCommentID int64, toUserID string, comment, username, phone, toUsername, toPhone, commentAddr string, userID string) (int64, error) {
	// 1. 参数验证
	if err := util.ValidateContent(comment); err != nil {
		return 0, err
	}

	// 2. 敏感词检查
	if util.ContainsSensitiveWord(comment) {
		return 0, constant.ErrContentHasSensitiveWord
	}

	// 3. 如果有被回复用户，查询用户信息
	if toUserID != "" && toUsername == "" {
		toUser, err := s.userRepo.FindByID(toUserID)
		if err == nil {
			toUsername = toUser.Username
			toPhone = toUser.Phone
		}
	}

	// 4. 创建二级评论
	subComment := &model.CommentSubTwo{
		ParentCommentID: parentCommentID,
		Comment:         comment,
		CommentTime:     time.Now().Format("2006-01-02 15:04:05"),
		Username:        username,
		UserID:          userID,
		Phone:           phone,
		GoodCount:       0,
		ToUsername:      toUsername,
		ToPhone:         toPhone,
		ToUserID:        toUserID,
		CommentAddr:     commentAddr,
	}

	if err := s.commentRepo.CreateSubComment(subComment); err != nil {
		return 0, constant.ErrCreateCommentFailed
	}

	// 5. 异步发送回复通知（不影响主流程）
	if toUserID != "" && userID != toUserID && queue.Client != nil {
		go func() {
			task := queue.CreateTask(queue.TaskTypeNotification, map[string]interface{}{
				"notify_type":    "reply_comment",
				"target_user_id": toUserID,
				"reply_user_id":  userID,
				"reply_username": username,
				"comment_id":     float64(subComment.ID),
			})
			ctx := context.Background()
			if err := queue.Client.PublishTask(ctx, task); err != nil {
				log.Printf("Failed to publish reply comment notification task: %v", err)
			}
		}()
	}

	// 6. 清除子评论列表缓存
	s.ClearCommentCache(fmt.Sprintf("sub_%d", parentCommentID))

	return subComment.ID, nil
}

// GetSubCommentList 获取二级评论列表（带缓存）
func (s *commentServiceV2) GetSubCommentList(parentID string) ([]model.CommentSubTwo, error) {
	// 1. 从缓存获取
	cacheKey := fmt.Sprintf("%ssub_%s", constant.CacheKeyCommentList, parentID)
	var subComments []model.CommentSubTwo

	err := s.cacheHelper.GetOrSet(
		cacheKey,
		&subComments,
		time.Duration(constant.CacheExpireShort)*time.Second,
		func() (interface{}, error) {
			// 从数据库查询
			return s.commentRepo.FindSubCommentsByParentID(parentID)
		},
	)

	if err != nil {
		return nil, constant.ErrDatabaseQuery
	}

	return subComments, nil
}

// LikeComment 点赞/取消点赞评论（企业级实现，使用事务）
func (s *commentServiceV2) LikeComment(commentID int64, userID string, commentType string) error {
	if commentType == "parent" {
		// 一级评论点赞
		return s.db.Transaction(func(tx *gorm.DB) error {
			if s.commentRepo.IsParentLiked(commentID, userID) {
				// 已点赞，取消
				if err := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).
					Delete(&model.CommentParentLike{}).Error; err != nil {
					return err
				}
				if err := tx.Model(&model.CommentParent{}).Where("id = ?", commentID).
					UpdateColumn("good_count", gorm.Expr("good_count - ?", 1)).Error; err != nil {
					return err
				}
			} else {
				// 未点赞，添加
				like := &model.CommentParentLike{
					CommentID: commentID,
					UserID:    userID,
				}
				if err := tx.Create(like).Error; err != nil {
					return err
				}
				if err := tx.Model(&model.CommentParent{}).Where("id = ?", commentID).
					UpdateColumn("good_count", gorm.Expr("good_count + ?", 1)).Error; err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		// 二级评论点赞
		return s.db.Transaction(func(tx *gorm.DB) error {
			if s.commentRepo.IsSubLiked(commentID, userID) {
				// 已点赞，取消
				if err := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).
					Delete(&model.CommentSubTwoLike{}).Error; err != nil {
					return err
				}
				if err := tx.Model(&model.CommentSubTwo{}).Where("id = ?", commentID).
					UpdateColumn("good_count", gorm.Expr("good_count - ?", 1)).Error; err != nil {
					return err
				}
			} else {
				// 未点赞，添加
				like := &model.CommentSubTwoLike{
					CommentID: commentID,
					UserID:    userID,
				}
				if err := tx.Create(like).Error; err != nil {
					return err
				}
				if err := tx.Model(&model.CommentSubTwo{}).Where("id = ?", commentID).
					UpdateColumn("good_count", gorm.Expr("good_count + ?", 1)).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}
}

// RefreshCommentCache 刷新评论缓存
func (s *commentServiceV2) RefreshCommentCache(articleID string) error {
	comments, err := s.commentRepo.FindParentCommentsByArticleID(articleID)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s%s", constant.CacheKeyCommentList, articleID)
	return s.cacheHelper.Set(cacheKey, comments, time.Duration(constant.CacheExpireShort)*time.Second)
}

// ClearCommentCache 清除评论缓存
func (s *commentServiceV2) ClearCommentCache(articleID string) error {
	cacheKey := fmt.Sprintf("%s%s", constant.CacheKeyCommentList, articleID)
	return s.cacheHelper.Delete(cacheKey)
}
