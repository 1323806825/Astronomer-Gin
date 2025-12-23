package elasticsearch

import (
	"astronomer-gin/config"
	"context"
	"fmt"
	"log"

	"github.com/olivere/elastic/v7"
)

var Client *elastic.Client

// InitElasticsearch 初始化ElasticSearch客户端
func InitElasticsearch(cfg *config.ElasticsearchConfig) error {
	if !cfg.Enabled {
		log.Println("ElasticSearch未启用，将使用MySQL搜索")
		return nil
	}

	options := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Addresses...),
		elastic.SetSniff(false), // 禁用嗅探（Docker环境常需要）
	}

	// 如果配置了用户名密码
	if cfg.Username != "" && cfg.Password != "" {
		options = append(options, elastic.SetBasicAuth(cfg.Username, cfg.Password))
	}

	var err error
	Client, err = elastic.NewClient(options...)
	if err != nil {
		return fmt.Errorf("连接ElasticSearch失败: %w", err)
	}

	// 测试连接
	ctx := context.Background()
	info, code, err := Client.Ping(cfg.Addresses[0]).Do(ctx)
	if err != nil {
		return fmt.Errorf("Ping ElasticSearch失败: %w", err)
	}

	log.Printf("✅ ElasticSearch连接成功: %s (version: %s, code: %d)",
		cfg.Addresses[0], info.Version.Number, code)

	return nil
}

// GetClient 获取ES客户端
func GetClient() *elastic.Client {
	return Client
}

// IsEnabled 检查ES是否启用
func IsEnabled() bool {
	return Client != nil
}
