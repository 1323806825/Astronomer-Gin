# Astronomer-Gin

这是从go-kratos微服务架构重构为Gin单体应用的Astronomer项目。

## 项目结构

```
Astronomer-Gin/
├── config/                 # 配置文件
│   ├── config.yaml        # 主配置文件
│   └── config.go          # 配置加载逻辑
├── handler/               # HTTP请求处理器（Controller层）
│   ├── user/              # 用户相关handler
│   ├── blog/              # 博客相关handler
│   └── comment/           # 评论相关handler
├── service/               # 业务逻辑层（Service层）
│   ├── user_service.go    # 用户业务逻辑
│   ├── blog_service.go    # 博客业务逻辑
│   └── comment_service.go # 评论业务逻辑
├── repository/            # 数据访问层（Repository层）
│   ├── user_repository.go    # 用户数据访问
│   ├── blog_repository.go    # 博客数据访问
│   └── comment_repository.go # 评论数据访问
├── middleware/            # Gin中间件
│   ├── auth.go            # JWT认证中间件
│   ├── cors.go            # CORS跨域中间件
│   └── logger.go          # 日志中间件
├── model/                 # 数据模型
│   ├── user.go            # 用户模型
│   ├── article.go         # 文章模型
│   ├── comment.go         # 评论模型
│   └── ...
├── pkg/                   # 公共包
│   ├── database/          # 数据库连接
│   ├── redis/             # Redis连接
│   ├── jwt/               # JWT工具
│   ├── captcha/           # 验证码生成
│   ├── constant/          # 常量定义
│   └── util/              # 工具函数
├── router/                # 路由配置
│   └── router.go          # 路由注册与依赖注入
├── logs/                  # 日志目录
├── main.go                # 程序入口
└── README.md              # 项目说明

```

## 功能模块

### 1. 用户模块 (User)
- 用户注册
- 用户登录
- 获取用户信息
- 更新用户信息
- 验证码生成

### 2. 博客模块 (Blog)
- 创建博客
- 获取博客列表（支持分页和标签筛选）
- 获取博客详情
- 更新博客
- 删除博客
- 博客点赞/取消点赞

### 3. 评论模块 (Comment)
- 创建一级评论
- 创建二级评论（回复评论）
- 获取文章评论列表
- 获取二级评论列表
- 评论点赞/取消点赞

## 技术栈

- **Web框架**: Gin v1.11.0
- **ORM**: GORM v1.31.1
- **数据库**: MySQL
- **缓存**: Redis
- **认证**: JWT (golang-jwt/jwt/v4)
- **验证码**: base64Captcha
- **密码加密**: bcrypt
- **配置解析**: YAML

## 配置说明

修改 `config/config.yaml` 文件：

```yaml
server:
  port: 8080              # 服务端口
  mode: debug             # 运行模式: debug, release, test

database:
  host: 116.198.234.46    # 数据库地址
  port: 3305              # 数据库端口
  username: root          # 数据库用户名
  password: root          # 数据库密码
  dbname: astronomer      # 数据库名称

redis:
  addr: 116.198.234.46:6376  # Redis地址
  password: ""               # Redis密码
  db: 0                      # Redis数据库

jwt:
  secret_key: astronomer     # JWT密钥
  expire_hours: 24           # Token过期时间(小时)
```

## 运行项目

### 方式1: 直接运行源码
```bash
go run main.go
```

### 方式2: 编译后运行
```bash
# 编译
go build -o astronomer-gin.exe main.go

# 运行
./astronomer-gin.exe
```

## API接口

### 用户相关
- `POST /api/v1/user/register` - 用户注册
- `POST /api/v1/user/login` - 用户登录
- `GET /api/v1/user/captcha` - 获取验证码
- `GET /api/v1/user/info` - 获取用户信息 (需认证)
- `PUT /api/v1/user/info` - 更新用户信息 (需认证)

### 博客相关
- `GET /api/v1/blog/list` - 获取博客列表
- `GET /api/v1/blog/:id` - 获取博客详情
- `POST /api/v1/blog` - 创建博客 (需认证)
- `PUT /api/v1/blog/:id` - 更新博客 (需认证)
- `DELETE /api/v1/blog/:id` - 删除博客 (需认证)
- `POST /api/v1/blog/:id/like` - 点赞博客 (需认证)

### 评论相关
- `GET /api/v1/comment/article/:articleId` - 获取文章评论列表
- `GET /api/v1/comment/sub/:parentId` - 获取二级评论列表
- `POST /api/v1/comment` - 创建评论 (需认证)
- `POST /api/v1/comment/sub` - 创建二级评论 (需认证)
- `POST /api/v1/comment/:id/like?type=parent|sub` - 点赞评论 (需认证)

### 健康检查
- `GET /health` - 健康检查

## 认证说明

需要认证的接口需要在请求头中携带token：

```
Authorization: Bearer <token>
```

或者在查询参数中携带：

```
?token=<token>
```

## 与原微服务架构的差异

### 原架构 (go-kratos微服务)
- 3个独立的微服务 (User, Blog, Comment)
- 使用gRPC和HTTP双协议
- Consul服务发现
- Wire依赖注入
- 复杂的项目结构

### 新架构 (Gin单体)
- 单一应用程序
- 仅使用HTTP协议
- 无需服务发现
- 简化的依赖管理
- 扁平化的项目结构
- 更易于开发和部署

## 架构设计

本项目采用经典的三层架构设计：

### Handler层 (Controller)
- 负责处理HTTP请求和响应
- 参数验证和绑定
- 调用Service层处理业务逻辑
- 返回统一的JSON响应

### Service层 (Business Logic)
- 封装业务逻辑
- 协调多个Repository
- 处理事务逻辑
- 数据转换和验证

### Repository层 (Data Access)
- 封装数据库访问操作
- 提供CRUD接口
- 隔离数据库实现细节
- 便于单元测试

### 依赖注入流程

```go
// router/router.go
db := database.GetDB()

// 初始化Repository层
userRepo := repository.NewUserRepository(db)
blogRepo := repository.NewBlogRepository(db)

// 初始化Service层（注入Repository��
userService := service.NewUserService(userRepo)
blogService := service.NewBlogService(blogRepo)

// 初始化Handler层（注入Service）
userHandler := user.NewUserHandler(userService)
blogHandler := blog.NewBlogHandler(blogService, userService)
```

## 开发说明

### 添加新功能

按照以下步骤添加新功能：

1. **定义数据模型** - 在 `model/` 中定义数据结构
```go
// model/example.go
type Example struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"type:varchar(100)"`
    CreatedAt time.Time
}
```

2. **创建Repository** - 在 `repository/` 中实现数据访问
```go
// repository/example_repository.go
type ExampleRepository interface {
    Create(example *model.Example) error
    FindByID(id uint) (*model.Example, error)
}

type exampleRepository struct {
    db *gorm.DB
}

func NewExampleRepository(db *gorm.DB) ExampleRepository {
    return &exampleRepository{db: db}
}
```

3. **实现Service** - 在 `service/` 中编写业务逻辑
```go
// service/example_service.go
type ExampleService interface {
    CreateExample(name string) error
}

type exampleService struct {
    exampleRepo repository.ExampleRepository
}

func NewExampleService(exampleRepo repository.ExampleRepository) ExampleService {
    return &exampleService{exampleRepo: exampleRepo}
}
```

4. **创建Handler** - 在 `handler/` 中处理HTTP请求
```go
// handler/example/example.go
type ExampleHandler struct {
    exampleService service.ExampleService
}

func NewExampleHandler(exampleService service.ExampleService) *ExampleHandler {
    return &ExampleHandler{exampleService: exampleService}
}

func (h *ExampleHandler) Create(c *gin.Context) {
    // 处理请求
}
```

5. **注册路由** - 在 `router/router.go` 中注册路由
```go
// 初始化
exampleRepo := repository.NewExampleRepository(db)
exampleService := service.NewExampleService(exampleRepo)
exampleHandler := example.NewExampleHandler(exampleService)

// 注册路由
api.POST("/example", exampleHandler.Create)
```

### 数据库迁移

项目使用GORM，可以使用AutoMigrate自动创建表：

```go
db.AutoMigrate(&model.User{}, &model.Article{}, &model.CommentParent{})
```

## 注意事项

1. 首次运行前请确保MySQL和Redis服务正常运行
2. 修改配置文件中的数据库连接信息
3. 生产环境请修改JWT密钥和其他敏感配置
4. 生产环境建议将mode设置为release

## License

MIT