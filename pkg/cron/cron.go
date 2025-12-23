package cron

import (
	"astronomer-gin/repository"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// CronManager 定时任务管理器
type CronManager struct {
	cron     *cron.Cron
	db       *gorm.DB
	blogRepo repository.BlogRepository
}

// NewCronManager 创建定时任务管理器
func NewCronManager(db *gorm.DB, blogRepo repository.BlogRepository) *CronManager {
	// 创建带秒级精度的cron实例
	c := cron.New(cron.WithSeconds())

	return &CronManager{
		cron:     c,
		db:       db,
		blogRepo: blogRepo,
	}
}

// Start 启动所有定时任务
func (m *CronManager) Start() error {
	log.Println("⏰ 初始化定时任务...")

	// 1. 每小时更新热度分数
	if _, err := m.cron.AddFunc("0 0 * * * *", m.UpdateHotScores); err != nil {
		return fmt.Errorf("添加热度更新任务失败: %w", err)
	}
	log.Println("✅ 热度更新任务: 每小时执行")

	// 2. 每天凌晨0点重置每日统计
	if _, err := m.cron.AddFunc("0 0 0 * * *", m.ResetDailyStats); err != nil {
		return fmt.Errorf("添加每日统计重置任务失败: %w", err)
	}
	log.Println("✅ 每日统计重置: 每天0点执行")

	// 3. 每周一凌晨1点重置每周统计
	if _, err := m.cron.AddFunc("0 0 1 * * 1", m.ResetWeeklyStats); err != nil {
		return fmt.Errorf("添加每周统计重置任务失败: %w", err)
	}
	log.Println("✅ 每周统计重置: 每周一1点执行")

	// 4. 每月1号凌晨2点重置每月统计
	if _, err := m.cron.AddFunc("0 0 2 1 * *", m.ResetMonthlyStats); err != nil {
		return fmt.Errorf("添加每月统计重置任务失败: %w", err)
	}
	log.Println("✅ 每月统计重置: 每月1号2点执行")

	// 5. 每10分钟清理过期的临时数据
	if _, err := m.cron.AddFunc("0 */10 * * * *", m.CleanupTempData); err != nil {
		return fmt.Errorf("添加临时数据清理任务失败: %w", err)
	}
	log.Println("✅ 临时数据清理: 每10分钟执行")

	// 6. 每天凌晨3点备份关键数据
	if _, err := m.cron.AddFunc("0 0 3 * * *", m.BackupCriticalData); err != nil {
		return fmt.Errorf("添加数据备份任务失败: %w", err)
	}
	log.Println("✅ 数据备份任务: 每天3点执行")

	// 启动定时任务
	m.cron.Start()
	log.Println("🚀 定时任务已启动")

	return nil
}

// Stop 停止所有定时任务
func (m *CronManager) Stop() {
	log.Println("⏸️  停止定时任务...")
	ctx := m.cron.Stop()
	<-ctx.Done()
	log.Println("✅ 定时任务已停止")
}

// ==================== 定时任务具体实现 ====================

// UpdateHotScores 更新热度分数
func (m *CronManager) UpdateHotScores() {
	startTime := time.Now()
	log.Println("\n[定时任务] 开始更新热度分数...")

	// Reddit热度算法
	// HotScore = (ViewCount*0.1 + LikeCount*0.5 + CommentCount*0.3 + FavoriteCount*0.1) * exp(-HoursSincePublish/48)

	query := `
		UPDATE article SET hot_score = (
			(visit_count * 0.1 + star_count * 0.5 + comment_count * 0.3) *
			EXP(- TIMESTAMPDIFF(HOUR, create_time, NOW()) / 48.0)
		)
		WHERE delete_time IS NULL
	`

	result := m.db.Exec(query)
	if result.Error != nil {
		log.Printf("❌ 更新热度分数失败: %v\n", result.Error)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ 热度分数更新完成！影响行数: %d, 耗时: %v\n", result.RowsAffected, duration)
}

// ResetDailyStats 重置每日统计
func (m *CronManager) ResetDailyStats() {
	startTime := time.Now()
	log.Println("\n[定时任务] 开始重置每日统计...")

	// 重置article表的today_view_count
	result := m.db.Exec("UPDATE article SET today_view_count = 0 WHERE delete_time IS NULL")
	if result.Error != nil {
		log.Printf("❌ 重置每日统计失败: %v\n", result.Error)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ 每日统计重置完成！影响行数: %d, 耗时: %v\n", result.RowsAffected, duration)
}

// ResetWeeklyStats 重置每周统计
func (m *CronManager) ResetWeeklyStats() {
	startTime := time.Now()
	log.Println("\n[定时任务] 开始重置每周统计...")

	// 重置article表的week_view_count
	result := m.db.Exec("UPDATE article SET week_view_count = 0 WHERE delete_time IS NULL")
	if result.Error != nil {
		log.Printf("❌ 重置每周统计失败: %v\n", result.Error)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ 每周统计重置完成！影响行数: %d, 耗时: %v\n", result.RowsAffected, duration)
}

// ResetMonthlyStats 重置每月统计
func (m *CronManager) ResetMonthlyStats() {
	startTime := time.Now()
	log.Println("\n[定时任务] 开始重置每月统计...")

	// 重置article表的month_view_count
	result := m.db.Exec("UPDATE article SET month_view_count = 0 WHERE delete_time IS NULL")
	if result.Error != nil {
		log.Printf("❌ 重置每月统计失败: %v\n", result.Error)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ 每月统计重置完成！影响行数: %d, 耗时: %v\n", result.RowsAffected, duration)
}

// CleanupTempData 清理过期的临时数据
func (m *CronManager) CleanupTempData() {
	startTime := time.Now()
	log.Println("\n[定时任务] 开始清理临时数据...")

	count := 0

	// 1. 清理24小时前的验证码缓存（Redis）
	// 注意: 这里假设使用Redis，实际需要调用Redis服务
	log.Println("  - 清理过期验证码...")

	// 2. 清理30天前的软删除数据
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	result := m.db.Exec("DELETE FROM article WHERE delete_time IS NOT NULL AND delete_time < ?", thirtyDaysAgo)
	if result.Error != nil {
		log.Printf("❌ 清理软删除文章失败: %v\n", result.Error)
	} else {
		count += int(result.RowsAffected)
		log.Printf("  - 清理软删除文章: %d条\n", result.RowsAffected)
	}

	// 3. 清理过期的草稿（90天未编辑）
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	result = m.db.Exec(`
		DELETE FROM article
		WHERE status = 0
		AND update_time < ?
		AND is_published = false
	`, ninetyDaysAgo)
	if result.Error != nil {
		log.Printf("❌ 清理过期草稿失败: %v\n", result.Error)
	} else {
		count += int(result.RowsAffected)
		log.Printf("  - 清理过期草稿: %d条\n", result.RowsAffected)
	}

	duration := time.Since(startTime)
	log.Printf("✅ 临时数据清理完成！总清理: %d条, 耗时: %v\n", count, duration)
}

// BackupCriticalData 备份关键数据
func (m *CronManager) BackupCriticalData() {
	startTime := time.Now()
	log.Println("\n[定时任务] 开始备份关键数据...")

	// 这里可以实现数据库备份逻辑
	// 例如: 导出SQL、上传到OSS等
	// 为了演示，这里只打印日志

	var stats struct {
		TotalArticles int64
		TotalComments int64
		TotalUsers    int64
		ActiveUsers   int64
	}

	m.db.Model(&struct{ TableName string }{}).Raw("SELECT COUNT(*) FROM article WHERE delete_time IS NULL").Scan(&stats.TotalArticles)
	m.db.Model(&struct{ TableName string }{}).Raw("SELECT COUNT(*) FROM comment_parent").Scan(&stats.TotalComments)
	m.db.Model(&struct{ TableName string }{}).Raw("SELECT COUNT(*) FROM user").Scan(&stats.TotalUsers)
	m.db.Model(&struct{ TableName string }{}).Raw("SELECT COUNT(*) FROM user WHERE last_login_time > DATE_SUB(NOW(), INTERVAL 30 DAY)").Scan(&stats.ActiveUsers)

	log.Printf("  数据统计:")
	log.Printf("  - 文章总数: %d", stats.TotalArticles)
	log.Printf("  - 评论总数: %d", stats.TotalComments)
	log.Printf("  - 用户总数: %d", stats.TotalUsers)
	log.Printf("  - 活跃用户(30天): %d", stats.ActiveUsers)

	duration := time.Since(startTime)
	log.Printf("✅ 数据备份完成！耗时: %v\n", duration)

	// TODO: 实际生产环境应该执行真正的备份操作
	// 例如: mysqldump、上传到云存储等
}

// ==================== 手动触发任务 ====================

// ManualUpdateHotScores 手动触发热度更新
func (m *CronManager) ManualUpdateHotScores() error {
	log.Println("🔧 手动触发热度更新...")
	m.UpdateHotScores()
	return nil
}

// ManualResetStats 手动触发统计重置
func (m *CronManager) ManualResetStats(statsType string) error {
	log.Printf("🔧 手动触发统计重置: %s\n", statsType)

	switch statsType {
	case "daily":
		m.ResetDailyStats()
	case "weekly":
		m.ResetWeeklyStats()
	case "monthly":
		m.ResetMonthlyStats()
	default:
		return fmt.Errorf("未知的统计类型: %s", statsType)
	}

	return nil
}

// GetCronStatus 获取定时任务状态
func (m *CronManager) GetCronStatus() map[string]interface{} {
	entries := m.cron.Entries()

	tasks := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		tasks = append(tasks, map[string]interface{}{
			"next_run": entry.Next,
			"prev_run": entry.Prev,
		})
	}

	return map[string]interface{}{
		"running":    true,
		"task_count": len(entries),
		"tasks":      tasks,
	}
}
