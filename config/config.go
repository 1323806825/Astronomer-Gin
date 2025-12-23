package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	MinIO         MinIOConfig         `yaml:"minio"`
	RabbitMQ      RabbitMQConfig      `yaml:"rabbitmq"`
	JWT           JWTConfig           `yaml:"jwt"`
	Admin         AdminConfig         `yaml:"admin"`
	Log           LogConfig           `yaml:"log"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Email         EmailConfig         `yaml:"email"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver       string `yaml:"driver"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	Charset      string `yaml:"charset"`
	ParseTime    bool   `yaml:"parse_time"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

// GetDSN 获取数据库连接字符串
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset, d.ParseTime)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// MinIOConfig MinIO对象存储配置
type MinIOConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	UseSSL     bool   `yaml:"use_ssl"`
	BucketName string `yaml:"bucket_name"`
	PublicURL  string `yaml:"public_url"`
}

// RabbitMQConfig RabbitMQ消息队列配置
type RabbitMQConfig struct {
	URL          string `yaml:"url"`
	ExchangeName string `yaml:"exchange_name"`
	QueueName    string `yaml:"queue_name"`
	RoutingKey   string `yaml:"routing_key"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey   string `yaml:"secret_key"`
	ExpireHours int    `yaml:"expire_hours"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// ElasticsearchConfig ElasticSearch配置
type ElasticsearchConfig struct {
	Addresses []string `yaml:"addresses"` // ES集群地址列表
	Username  string   `yaml:"username"`  // ES用户名（如果启用了安全认证）
	Password  string   `yaml:"password"`  // ES密码
	Enabled   bool     `yaml:"enabled"`   // 是否启用ES（false则降级到MySQL）
}

// EmailConfig 邮件服务配置
type EmailConfig struct {
	Enabled   bool                `yaml:"enabled"`   // 是否启用邮件服务
	SMTPHost  string              `yaml:"smtp_host"` // SMTP服务器地址
	SMTPPort  int                 `yaml:"smtp_port"` // SMTP端口
	Username  string              `yaml:"username"`  // 发件人邮箱
	Password  string              `yaml:"password"`  // 邮箱授权码
	FromName  string              `yaml:"from_name"` // 发件人名称
	UseSSL    bool                `yaml:"use_ssl"`   // 是否使用SSL
	Templates EmailTemplateConfig `yaml:"templates"` // 邮件模板配置
}

// EmailTemplateConfig 邮件模板配置
type EmailTemplateConfig struct {
	WelcomeSubject string `yaml:"welcome_subject"` // 欢迎邮件主题
	CommentSubject string `yaml:"comment_subject"` // 评论通知主题
	LikeSubject    string `yaml:"like_subject"`    // 点赞通知主题
	FollowSubject  string `yaml:"follow_subject"`  // 关注通知主题
}

var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}
