package router

import (
	"astronomer-gin/handler"
	"astronomer-gin/handler/admin"
	"astronomer-gin/handler/chat"
	"astronomer-gin/handler/favorite"
	"astronomer-gin/handler/follow"
	"astronomer-gin/handler/notification"
	"astronomer-gin/handler/search"
	"astronomer-gin/handler/trending"
	"astronomer-gin/handler/upload"
	"astronomer-gin/handler/user"
	"astronomer-gin/middleware"
	"astronomer-gin/pkg/database"
	"astronomer-gin/repository"
	"astronomer-gin/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// ==================== 全局中间件 ====================
	// 1. 请求追踪（生成Request ID）
	r.Use(middleware.RequestID())

	// 2. 跨域处理
	r.Use(middleware.CORS())

	// 3. 日志记录
	r.Use(middleware.Logger())

	// 4. 全局错误处理（panic恢复）
	r.Use(middleware.ErrorHandler())

	// 获取数据库连接
	db := database.GetDB()

	// 初始化Repository层
	userRepo := repository.NewUserRepository(db)
	blogRepo := repository.NewBlogRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	followRepo := repository.NewFollowRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// 初始化V3 Repository层
	articleV3Repo := repository.NewArticleV3Repository(db)
	commentV3Repo := repository.NewCommentV3Repository(db)
	likeRepo := repository.NewLikeRepository(db)
	columnRepo := repository.NewColumnRepository(db)

	// 初始化Service层（使用V2版本）
	userService := service.NewUserServiceV2(userRepo)
	favoriteService := service.NewFavoriteServiceV2(favoriteRepo, blogRepo, notifyRepo)
	followService := service.NewFollowServiceV2(followRepo, userRepo, notifyRepo)
	notifyService := service.NewNotificationServiceV2(notifyRepo)
	uploadService := service.NewUploadServiceV2()
	searchService := service.NewSearchServiceV2(blogRepo, userRepo)
	trendingService := service.NewTrendingServiceV2(blogRepo, userRepo)
	chatService := service.NewChatServiceV2(chatRepo, followRepo, userRepo)

	// 初始化V3 Service层（企业级功能）
	articleV3Service := service.NewArticleV3Service(articleV3Repo, userRepo, followRepo, likeRepo, favoriteRepo, db)
	commentV3Service := service.NewCommentV3Service(commentV3Repo, articleV3Repo, userRepo, likeRepo, notifyRepo, db)
	columnService := service.NewColumnService(columnRepo, userRepo, notifyRepo, articleV3Repo)

	// 初始化Handler层
	userHandler := user.NewUserHandler(userService)
	favoriteHandler := favorite.NewFavoriteHandler(favoriteService, userService)
	followHandler := follow.NewFollowHandler(followService, userService)
	notifyHandler := notification.NewNotificationHandler(notifyService, userService)
	uploadHandler := upload.NewUploadHandler(uploadService)
	searchHandler := search.NewSearchHandler(searchService)
	trendingHandler := trending.NewTrendingHandler(trendingService)
	syncHandler := admin.NewSyncHandler(blogRepo)
	chatHandler := chat.NewChatHandler(chatService, userService)

	// 初始化V3 Handler层（企业级功能）
	articleV3Handler := handler.NewArticleV3Handler(articleV3Service)
	commentV3Handler := handler.NewCommentV3Handler(commentV3Service)
	columnHandler := handler.NewColumnHandler(columnService)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==================== 注册V3路由（企业级功能） ====================
	articleV3Handler.RegisterRoutes(r)
	commentV3Handler.RegisterRoutes(r)

	// ==================== V3版本的用户路由（兼容前端） ====================
	apiV3 := r.Group("/api/v3")
	{
		// 用户相关路由（无需认证）
		userV3Public := apiV3.Group("/user")
		{
			userV3Public.POST("/register", middleware.RegisterRateLimit(), userHandler.Register)
			userV3Public.POST("/login", middleware.LoginRateLimit(), userHandler.Login)
			userV3Public.GET("/captcha", userHandler.GetCaptcha)
		}

		// 用户相关路由（需要认证）- 必须在动态路由之前注册
		userV3Auth := apiV3.Group("/user")
		userV3Auth.Use(middleware.AuthMiddleware())
		{
			userV3Auth.GET("/current", userHandler.GetUserInfo) // 获取当前登录用户信息
			userV3Auth.GET("/info", userHandler.GetUserInfo)
			userV3Auth.PUT("/update", userHandler.UpdateUserInfo)
		}

		// 用户公开路由（动态路由，必须在静态路由之后）
		userV3PublicDynamic := apiV3.Group("/user")
		{
			userV3PublicDynamic.GET("/:id", userHandler.GetUserProfile) // 获取用户资料（公开）
		}

		// 上传功能
		uploadV3Auth := apiV3.Group("/upload")
		uploadV3Auth.Use(middleware.AuthMiddleware())
		{
			uploadV3Auth.POST("/image", uploadHandler.UploadImage)
			uploadV3Auth.POST("/file", uploadHandler.UploadFile)
			uploadV3Auth.POST("/multiple", uploadHandler.UploadMultiple)
			uploadV3Auth.DELETE("/delete", uploadHandler.DeleteFile)
		}

		// 搜索功能
		searchV3Public := apiV3.Group("/search")
		{
			searchV3Public.GET("/articles", searchHandler.SearchArticles)
			searchV3Public.GET("/users", searchHandler.SearchUsers)
			searchV3Public.GET("/all", searchHandler.SearchAll)
		}

		// 通知功能
		notifyV3Auth := apiV3.Group("/notifications")
		notifyV3Auth.Use(middleware.AuthMiddleware())
		{
			notifyV3Auth.GET("", notifyHandler.GetNotifications)
			notifyV3Auth.GET("/unread", notifyHandler.GetUnreadCount)
			notifyV3Auth.PUT("/:id/read", notifyHandler.MarkAsRead)
			notifyV3Auth.PUT("/read-all", notifyHandler.MarkAllAsRead)
			notifyV3Auth.DELETE("/:id", notifyHandler.DeleteNotification)
		}

		// ==================== 专栏功能（完整路由） ====================
		// 专栏公开路由
		columnV3Public := apiV3.Group("/columns")
		{
			columnV3Public.GET("", columnHandler.GetList)                        // 获取专栏列表
			columnV3Public.GET("/hot", columnHandler.GetHotColumns)              // 获取热门专栏
			columnV3Public.GET("/:id", columnHandler.GetDetail)                  // 获取专栏详情
			columnV3Public.GET("/:id/articles", columnHandler.GetColumnArticles) // 获取专栏文章列表
		}

		// 专栏认证路由
		columnV3Auth := apiV3.Group("/columns")
		columnV3Auth.Use(middleware.AuthMiddleware())
		{
			columnV3Auth.POST("", columnHandler.Create)                                                // 创建专栏
			columnV3Auth.PUT("/:id", columnHandler.Update)                                             // 更新专栏
			columnV3Auth.DELETE("/:id", columnHandler.Delete)                                          // 删除专栏
			columnV3Auth.POST("/:id/subscribe", columnHandler.Subscribe)                               // 订阅专栏
			columnV3Auth.DELETE("/:id/subscribe", columnHandler.Unsubscribe)                           // 取消订阅
			columnV3Auth.GET("/subscribed", columnHandler.GetSubscribedColumns)                        // 获取我订阅的专栏
			columnV3Auth.POST("/:id/articles", columnHandler.AddArticle)                               // 添加文章到专栏
			columnV3Auth.DELETE("/:id/articles/:articleId", columnHandler.RemoveArticle)               // 从专栏移除文章
			columnV3Auth.PUT("/:id/articles/:articleId/position", columnHandler.UpdateArticlePosition) // 更新文章位置
		}

		// 用户专栏路由（使用 :id 与其他用户路由保持一致）
		apiV3.GET("/user/:id/columns", columnHandler.GetUserColumns) // 获取用户的专栏列表

		// ==================== 关注/好友/拉黑功能 ====================
		// 关注相关路由（公开 - 查看粉丝/关注列���）
		followV3Public := apiV3.Group("/user")
		{
			followV3Public.GET("/:id/followers", followHandler.GetFollowers)
			followV3Public.GET("/:id/following", followHandler.GetFollowing)
		}

		// 关注相关路由（需要认证）
		followV3Auth := apiV3.Group("/user")
		followV3Auth.Use(middleware.AuthMiddleware())
		{
			followV3Auth.POST("/:id/follow", middleware.FollowRateLimit(), followHandler.FollowUser)
			followV3Auth.DELETE("/:id/follow", followHandler.UnfollowUser)
			followV3Auth.GET("/:id/is-following", followHandler.CheckFollowing)

			// 好友相关
			followV3Auth.GET("/friends", followHandler.GetFriendsList)
			followV3Auth.GET("/:id/is-friend", followHandler.CheckFriend)
			followV3Auth.GET("/friends/count", followHandler.GetFriendsCount)

			// 拉黑相关
			followV3Auth.POST("/:id/block", followHandler.BlockUser)
			followV3Auth.DELETE("/:id/block", followHandler.UnblockUser)
			followV3Auth.GET("/blocked", followHandler.GetBlockList)
		}

		// ==================== 收藏功能 ====================
		favoriteV3Auth := apiV3.Group("/user")
		favoriteV3Auth.Use(middleware.AuthMiddleware())
		{
			favoriteV3Auth.GET("/favorites", favoriteHandler.GetUserFavorites)
		}

		// ==================== 热门榜单功能 ====================
		// 热门榜单（公开）
		trendingV3Public := apiV3.Group("/trending")
		{
			trendingV3Public.GET("/articles", trendingHandler.GetTrendingArticles)
			trendingV3Public.GET("/users", trendingHandler.GetTrendingUsers)
		}

		// 热门榜单管理（需要认证）
		trendingV3Auth := apiV3.Group("/trending")
		trendingV3Auth.Use(middleware.AuthMiddleware())
		{
			trendingV3Auth.POST("/refresh", trendingHandler.RefreshTrending)
		}

		// ==================== 私信聊天功能 ====================
		// 私信功能（需要认证，只有好友才能互发私信）
		chatV3Auth := apiV3.Group("/chat")
		chatV3Auth.Use(middleware.AuthMiddleware())
		{
			chatV3Auth.POST("/send", chatHandler.SendMessage)
			chatV3Auth.GET("/history/:id", chatHandler.GetChatHistory)
			chatV3Auth.GET("/sessions", chatHandler.GetChatSessions)
			chatV3Auth.GET("/unread", chatHandler.GetUnreadCount)
			chatV3Auth.PUT("/read/:id", chatHandler.MarkAsRead)
			chatV3Auth.DELETE("/message/:id", chatHandler.DeleteMessage)
			chatV3Auth.DELETE("/conversation/:id", chatHandler.DeleteChatWithUser)
		}

		// ==================== WebSocket 实时通信 ====================
		// WebSocket连接（需要认证）
		wsV3Auth := apiV3.Group("/ws")
		wsV3Auth.Use(middleware.AuthMiddleware())
		{
			wsV3Auth.GET("/chat", chatHandler.HandleWebSocket)
		}

		// ==================== 管理员功能 ====================
		// 管理员操作（需要认证）
		adminV3Auth := apiV3.Group("/admin")
		adminV3Auth.Use(middleware.AuthMiddleware())
		{
			adminV3Auth.POST("/sync/articles", syncHandler.SyncArticlesToES)
		}
	}

	return r
}
