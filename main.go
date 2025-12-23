package main

import (
	"astronomer-gin/config"
	"astronomer-gin/pkg/cron"
	"astronomer-gin/pkg/database"
	"astronomer-gin/pkg/elasticsearch"
	"astronomer-gin/pkg/email"
	"astronomer-gin/pkg/minio"
	"astronomer-gin/pkg/queue"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/repository"
	"astronomer-gin/router"
	"astronomer-gin/worker"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	_ "astronomer-gin/docs" // Swagger生成的docs包
)

// @title           Astronomer-Gin API
// @version         2.2.0
// @description     企业级社交博客平台API - 支持博客发布、评论互动、用户关注、通知推送等功能
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/astronomer-gin/astronomer-gin
// @contact.email  support@astronomer.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Token. 格式: Bearer {token}

// @tag.name 用户模块
// @tag.description 用户注册、登录、信息管理等接口

// @tag.name 博客模块
// @tag.description 博客CRUD、点赞、草稿管理等接口

// @tag.name 评论模块
// @tag.description 评论发表、回复、点赞等接口

// @tag.name 收藏模块
// @tag.description 文章收藏相关接口

// @tag.name 关注模块
// @tag.description 用户关注、粉丝、拉黑等接口

// @tag.name 通知模块
// @tag.description 系统通知、消息推送等接口

// @tag.name 文件上传模块
// @tag.description 图片、文件上传相关接口

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 初始化Redis
	if err := redis.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}
	defer redis.Close()

	// 初始化ElasticSearch（可选）
	if err := elasticsearch.InitElasticsearch(&cfg.Elasticsearch); err != nil {
		log.Printf("⚠️  初始化ElasticSearch失败: %v (将使用MySQL搜索)", err)
	} else if cfg.Elasticsearch.Enabled {
		// 创建文章索引
		if err := elasticsearch.CreateArticleIndex(); err != nil {
			log.Printf("⚠️  创建文章索引失败: %v", err)
		}
	}

	// 初始化MinIO
	if err := minio.InitMinIO(&cfg.MinIO); err != nil {
		log.Fatalf("初始化MinIO失败: %v", err)
	}
	log.Println("MinIO客户端初始化成功")

	// 初始化邮件服务
	if err := email.InitEmail(&cfg.Email); err != nil {
		log.Printf("⚠️  初始化邮件服务失败: %v (将禁用邮件功能)", err)
	}

	// 初始化RabbitMQ
	if err := queue.InitRabbitMQ(&cfg.RabbitMQ); err != nil {
		log.Fatalf("初始化RabbitMQ失败: %v", err)
	}
	defer queue.Client.Close()

	// 启动异步任务Worker
	// 初始化Repository(Worker需要用到)
	db := database.GetDB()
	notifyRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)
	blogRepo := repository.NewBlogRepository(db)

	// 创建各种处理器
	notificationHandler := worker.NewNotificationHandler(notifyRepo, userRepo)
	statsHandler := worker.NewStatsHandler(blogRepo)
	combinedHandler := worker.NewCombinedHandler(notificationHandler, statsHandler)

	// 启动Worker（5个并发）
	taskWorker := worker.NewTaskWorker(queue.Client, combinedHandler, 5)
	if err := taskWorker.Start(); err != nil {
		log.Fatalf("启动Task Worker失败: %v", err)
	}
	defer taskWorker.Stop()
	log.Println("Task Worker启动成功 (5个并发worker,支持通知和统计任务)")

	// 初始化并启动定时任务
	cronManager := cron.NewCronManager(db, blogRepo)
	if err := cronManager.Start(); err != nil {
		log.Fatalf("启动定时任务失败: %v", err)
	}
	defer cronManager.Stop()

	// 设置路由
	r := router.SetupRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("服务器启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
