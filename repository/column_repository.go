package repository

import (
	"astronomer-gin/model"
	"fmt"

	"gorm.io/gorm"
)

// ColumnRepository 专栏仓储接口
type ColumnRepository interface {
	// GetList 获取专栏列表（分页）
	GetList(page, pageSize int) ([]*model.ArticleColumn, int64, error)

	// GetByID 根据ID获取专栏
	GetByID(id uint64) (*model.ArticleColumn, error)

	// GetByUserID 获取用户的专栏列表
	GetByUserID(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error)

	// GetHotColumns 获取热门专栏（按订阅数排序）
	GetHotColumns(limit int) ([]*model.ArticleColumn, error)

	// Create 创建专栏
	Create(column *model.ArticleColumn) error

	// Update 更新专栏
	Update(column *model.ArticleColumn) error

	// Delete 删除专栏
	Delete(id uint64) error

	// IncrementArticleCount 增加专栏文章数
	IncrementArticleCount(columnID uint64) error

	// DecrementArticleCount 减少专栏文章数
	DecrementArticleCount(columnID uint64) error

	// CheckSubscribed 检查用户是否订阅了专栏
	CheckSubscribed(userID string, columnID uint64) (bool, error)

	// Subscribe 订阅专栏
	Subscribe(userID string, columnID uint64) error

	// Unsubscribe 取消订阅专栏
	Unsubscribe(userID string, columnID uint64) error

	// GetSubscribers 获取专栏的订阅者列表
	GetSubscribers(columnID uint64, page, pageSize int) ([]string, int64, error)

	// GetSubscribedColumns 获取用户订阅的专栏列表
	GetSubscribedColumns(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error)

	// AddArticle 将文章添加到专栏
	AddArticle(columnID, articleID uint64, sortOrder int) error

	// RemoveArticle 从专栏移除文章
	RemoveArticle(columnID, articleID uint64) error

	// GetColumnArticles 获取专栏的文章列表
	GetColumnArticles(columnID uint64, page, pageSize int) ([]*model.ArticleV3, int64, error)

	// GetArticlePosition 获取文章在专栏中的位置
	GetArticlePosition(columnID, articleID uint64) (int, error)

	// UpdateArticlePosition 更新文章在专栏中的位置
	UpdateArticlePosition(columnID, articleID uint64, sortOrder int) error
}

type columnRepository struct {
	db *gorm.DB
}

// NewColumnRepository 创建专栏仓储实例
func NewColumnRepository(db *gorm.DB) ColumnRepository {
	return &columnRepository{db: db}
}

// GetList 获取专栏列表（分页）
func (r *columnRepository) GetList(page, pageSize int) ([]*model.ArticleColumn, int64, error) {
	var columns []*model.ArticleColumn
	var total int64

	offset := (page - 1) * pageSize

	// 只获取正常状态的专栏
	if err := r.db.Model(&model.ArticleColumn{}).
		Where("status = ?", 1).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计专栏总数失败: %w", err)
	}

	if err := r.db.Where("status = ?", 1).
		Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&columns).Error; err != nil {
		return nil, 0, fmt.Errorf("查询专栏列表失败: %w", err)
	}

	return columns, total, nil
}

// GetByID 根据ID获取专栏
func (r *columnRepository) GetByID(id uint64) (*model.ArticleColumn, error) {
	var column model.ArticleColumn
	if err := r.db.Where("id = ? AND status = ?", id, 1).First(&column).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("专栏不存在")
		}
		return nil, fmt.Errorf("查询专栏失败: %w", err)
	}
	return &column, nil
}

// GetByUserID 获取用户的专栏列表
func (r *columnRepository) GetByUserID(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error) {
	var columns []*model.ArticleColumn
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.Model(&model.ArticleColumn{}).
		Where("user_id = ? AND status = ?", userID, 1).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计用户专栏总数失败: %w", err)
	}

	if err := r.db.Where("user_id = ? AND status = ?", userID, 1).
		Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&columns).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户专栏列表失败: %w", err)
	}

	return columns, total, nil
}

// GetHotColumns 获取热门专栏（按订阅数排序）
func (r *columnRepository) GetHotColumns(limit int) ([]*model.ArticleColumn, error) {
	var columns []*model.ArticleColumn

	if err := r.db.Where("status = ?", 1).
		Order("subscriber_count DESC, article_count DESC").
		Limit(limit).
		Find(&columns).Error; err != nil {
		return nil, fmt.Errorf("查询热门专栏失败: %w", err)
	}

	return columns, nil
}

// Create 创建专栏
func (r *columnRepository) Create(column *model.ArticleColumn) error {
	if err := r.db.Create(column).Error; err != nil {
		return fmt.Errorf("创建专栏失败: %w", err)
	}
	return nil
}

// Update 更新专栏
func (r *columnRepository) Update(column *model.ArticleColumn) error {
	if err := r.db.Save(column).Error; err != nil {
		return fmt.Errorf("更新专栏失败: %w", err)
	}
	return nil
}

// Delete 删除专栏（软删除，设置状态为隐藏）
func (r *columnRepository) Delete(id uint64) error {
	if err := r.db.Model(&model.ArticleColumn{}).
		Where("id = ?", id).
		Update("status", 2).Error; err != nil {
		return fmt.Errorf("删除专栏失败: %w", err)
	}
	return nil
}

// IncrementArticleCount 增加专栏文章数
func (r *columnRepository) IncrementArticleCount(columnID uint64) error {
	if err := r.db.Model(&model.ArticleColumn{}).
		Where("id = ?", columnID).
		UpdateColumn("article_count", gorm.Expr("article_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("增加专栏文章数失败: %w", err)
	}
	return nil
}

// DecrementArticleCount 减少专栏文章数
func (r *columnRepository) DecrementArticleCount(columnID uint64) error {
	if err := r.db.Model(&model.ArticleColumn{}).
		Where("id = ?", columnID).
		Where("article_count > ?", 0).
		UpdateColumn("article_count", gorm.Expr("article_count - ?", 1)).Error; err != nil {
		return fmt.Errorf("减少专栏文章数失败: %w", err)
	}
	return nil
}

// CheckSubscribed 检查用户是否订阅了专栏
func (r *columnRepository) CheckSubscribed(userID string, columnID uint64) (bool, error) {
	var count int64
	err := r.db.Model(&model.ColumnSubscription{}).
		Where("user_id = ? AND column_id = ?", userID, columnID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查订阅状态失败: %w", err)
	}
	return count > 0, nil
}

// Subscribe 订阅专栏
func (r *columnRepository) Subscribe(userID string, columnID uint64) error {
	subscription := &model.ColumnSubscription{
		ColumnID: columnID,
		UserID:   userID,
	}

	// 开启事务
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否已订阅
		var count int64
		if err := tx.Model(&model.ColumnSubscription{}).
			Where("user_id = ? AND column_id = ?", userID, columnID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("检查订阅状态失败: %w", err)
		}

		if count > 0 {
			return fmt.Errorf("已经订阅过该专栏")
		}

		// 添加订阅记录
		if err := tx.Create(subscription).Error; err != nil {
			return fmt.Errorf("创建订阅记录失败: %w", err)
		}

		// 增加订阅数
		if err := tx.Model(&model.ArticleColumn{}).
			Where("id = ?", columnID).
			UpdateColumn("subscriber_count", gorm.Expr("subscriber_count + ?", 1)).Error; err != nil {
			return fmt.Errorf("更新订阅数失败: %w", err)
		}

		return nil
	})
}

// Unsubscribe 取消订阅专栏
func (r *columnRepository) Unsubscribe(userID string, columnID uint64) error {
	// 开启事务
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除订阅记录
		result := tx.Where("user_id = ? AND column_id = ?", userID, columnID).
			Delete(&model.ColumnSubscription{})

		if result.Error != nil {
			return fmt.Errorf("删除订阅记录失败: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("未订阅该专栏")
		}

		// 减少订阅数
		if err := tx.Model(&model.ArticleColumn{}).
			Where("id = ?", columnID).
			Where("subscriber_count > ?", 0).
			UpdateColumn("subscriber_count", gorm.Expr("subscriber_count - ?", 1)).Error; err != nil {
			return fmt.Errorf("更新订阅数失败: %w", err)
		}

		return nil
	})
}

// GetSubscribers 获取专栏的订阅者列表
func (r *columnRepository) GetSubscribers(columnID uint64, page, pageSize int) ([]string, int64, error) {
	var userIDs []string
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.Model(&model.ColumnSubscription{}).
		Where("column_id = ?", columnID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计订阅者总数失败: %w", err)
	}

	if err := r.db.Model(&model.ColumnSubscription{}).
		Where("column_id = ?", columnID).
		Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询订阅者列表失败: %w", err)
	}

	return userIDs, total, nil
}

// GetSubscribedColumns 获取用户订阅的专栏列表
func (r *columnRepository) GetSubscribedColumns(userID string, page, pageSize int) ([]*model.ArticleColumn, int64, error) {
	var columns []*model.ArticleColumn
	var total int64

	offset := (page - 1) * pageSize

	// 统计总数
	if err := r.db.Table("column_subscription").
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计订阅专栏总数失败: %w", err)
	}

	// 关联查询
	if err := r.db.Table("article_column").
		Joins("INNER JOIN column_subscription ON article_column.id = column_subscription.column_id").
		Where("column_subscription.user_id = ? AND article_column.status = ?", userID, 1).
		Order("column_subscription.create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&columns).Error; err != nil {
		return nil, 0, fmt.Errorf("查询订阅专栏列表失败: %w", err)
	}

	return columns, total, nil
}

// AddArticle 将文章添加到专栏
func (r *columnRepository) AddArticle(columnID, articleID uint64, sortOrder int) error {
	rel := &model.ArticleColumnRel{
		ColumnID:  columnID,
		ArticleID: articleID,
		SortOrder: sortOrder,
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否已存在
		var count int64
		if err := tx.Model(&model.ArticleColumnRel{}).
			Where("column_id = ? AND article_id = ?", columnID, articleID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("检查文章关联失败: %w", err)
		}

		if count > 0 {
			return fmt.Errorf("文章已在该专栏中")
		}

		// 添加关联
		if err := tx.Create(rel).Error; err != nil {
			return fmt.Errorf("创建文章关联失败: %w", err)
		}

		// 增加专栏文章数
		if err := r.IncrementArticleCount(columnID); err != nil {
			return err
		}

		// 更新文章的column_id
		if err := tx.Model(&model.ArticleV3{}).
			Where("id = ?", articleID).
			Update("column_id", columnID).Error; err != nil {
			return fmt.Errorf("更新文章专栏ID失败: %w", err)
		}

		return nil
	})
}

// RemoveArticle 从专栏移除文章
func (r *columnRepository) RemoveArticle(columnID, articleID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联
		result := tx.Where("column_id = ? AND article_id = ?", columnID, articleID).
			Delete(&model.ArticleColumnRel{})

		if result.Error != nil {
			return fmt.Errorf("删除文章关联失败: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("文章不在该专栏中")
		}

		// 减少专栏文章数
		if err := r.DecrementArticleCount(columnID); err != nil {
			return err
		}

		// 清除文章的column_id
		if err := tx.Model(&model.ArticleV3{}).
			Where("id = ? AND column_id = ?", articleID, columnID).
			Update("column_id", 0).Error; err != nil {
			return fmt.Errorf("清除文章专栏ID失败: %w", err)
		}

		return nil
	})
}

// GetColumnArticles 获取专栏的文章列表
func (r *columnRepository) GetColumnArticles(columnID uint64, page, pageSize int) ([]*model.ArticleV3, int64, error) {
	var articles []*model.ArticleV3
	var total int64

	offset := (page - 1) * pageSize

	// 统计总数
	if err := r.db.Table("article_column_rel").
		Where("column_id = ?", columnID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计专栏文章总数失败: %w", err)
	}

	// 关联查询文章，按专栏中的顺序排序
	if err := r.db.Table("article_v3").
		Joins("INNER JOIN article_column_rel ON article_v3.id = article_column_rel.article_id").
		Where("article_column_rel.column_id = ? AND article_v3.status = ?", columnID, 1).
		Order("article_column_rel.sort_order ASC, article_column_rel.add_time ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error; err != nil {
		return nil, 0, fmt.Errorf("查询专栏文章列表失败: %w", err)
	}

	return articles, total, nil
}

// GetArticlePosition 获取文章在专栏中的位置
func (r *columnRepository) GetArticlePosition(columnID, articleID uint64) (int, error) {
	var rel model.ArticleColumnRel
	if err := r.db.Where("column_id = ? AND article_id = ?", columnID, articleID).
		First(&rel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("文章不在该专栏中")
		}
		return 0, fmt.Errorf("查询文章位置失败: %w", err)
	}
	return rel.SortOrder, nil
}

// UpdateArticlePosition 更新文章在专栏中的位置
func (r *columnRepository) UpdateArticlePosition(columnID, articleID uint64, sortOrder int) error {
	result := r.db.Model(&model.ArticleColumnRel{}).
		Where("column_id = ? AND article_id = ?", columnID, articleID).
		Update("sort_order", sortOrder)

	if result.Error != nil {
		return fmt.Errorf("更新文章位置失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("文章不在该专栏中")
	}

	return nil
}
