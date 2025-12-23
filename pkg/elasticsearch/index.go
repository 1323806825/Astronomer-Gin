package elasticsearch

import (
	"astronomer-gin/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

const (
	// 索引名称
	ArticleIndex = "articles"
	UserIndex    = "users"
)

// ArticleDocument ES文章文档结构
type ArticleDocument struct {
	ID            uint64    `json:"id"`
	UserID        string    `json:"user_id"` // UUID字符串
	Title         string    `json:"title"`
	Preface       string    `json:"preface"`
	Content       string    `json:"content"`
	Photo         string    `json:"photo"`
	Tag           string    `json:"tag"`
	Visit         int64     `json:"visit"`
	GoodCount     int64     `json:"good_count"`
	CommentCount  int64     `json:"comment_count"`
	FavoriteCount int64     `json:"favorite_count"`
	CreateTime    time.Time `json:"create_time"`
	UpdateTime    time.Time `json:"update_time"`
}

// 文章索引mapping（包含中文分词器IK）
const articleMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "ik_smart_analyzer": {
          "type": "custom",
          "tokenizer": "ik_smart"
        },
        "ik_max_word_analyzer": {
          "type": "custom",
          "tokenizer": "ik_max_word"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "long"
      },
      "user_id": {
        "type": "long"
      },
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "preface": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "content": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "photo": {
        "type": "keyword"
      },
      "tag": {
        "type": "keyword"
      },
      "visit": {
        "type": "long"
      },
      "good_count": {
        "type": "long"
      },
      "comment_count": {
        "type": "long"
      },
      "favorite_count": {
        "type": "long"
      },
      "create_time": {
        "type": "date"
      },
      "update_time": {
        "type": "date"
      }
    }
  }
}`

// CreateArticleIndex 创建文章索引
func CreateArticleIndex() error {
	if Client == nil {
		return fmt.Errorf("ES客户端未初始化")
	}

	ctx := context.Background()

	// 检查索引是否存在
	exists, err := Client.IndexExists(ArticleIndex).Do(ctx)
	if err != nil {
		return fmt.Errorf("检查索引失败: %w", err)
	}

	if exists {
		fmt.Printf("索引 %s 已存在\n", ArticleIndex)
		return nil
	}

	// 创建索引
	createIndex, err := Client.CreateIndex(ArticleIndex).
		BodyString(articleMapping).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	if !createIndex.Acknowledged {
		return fmt.Errorf("创建索引未确认")
	}

	fmt.Printf("✅ 成功创建索引: %s\n", ArticleIndex)
	return nil
}

// IndexArticle 索引单篇文章
func IndexArticle(article *model.Article) error {
	if Client == nil {
		return nil // ES未启用，静默失败
	}

	ctx := context.Background()

	doc := ArticleDocument{
		ID:            article.ID,
		UserID:        article.UserID,
		Title:         article.Title,
		Preface:       article.Preface,
		Content:       article.Content,
		Photo:         article.Photo,
		Tag:           article.Tag,
		Visit:         int64(article.Visit),
		GoodCount:     int64(article.GoodCount),
		CommentCount:  int64(article.CommentCount),
		FavoriteCount: int64(article.FavoriteCount),
		CreateTime:    article.CreateTime,
		UpdateTime:    article.UpdateTime,
	}

	_, err := Client.Index().
		Index(ArticleIndex).
		Id(fmt.Sprintf("%d", article.ID)).
		BodyJson(doc).
		Do(ctx)

	return err
}

// UpdateArticle 更新文章索引
func UpdateArticle(articleID uint64, updates map[string]interface{}) error {
	if Client == nil {
		return nil // ES未启用，静默失败
	}

	ctx := context.Background()

	_, err := Client.Update().
		Index(ArticleIndex).
		Id(fmt.Sprintf("%d", articleID)).
		Doc(updates).
		Do(ctx)

	return err
}

// DeleteArticle 删除文章索引
func DeleteArticle(articleID uint64) error {
	if Client == nil {
		return nil // ES未启用，静默失败
	}

	ctx := context.Background()

	_, err := Client.Delete().
		Index(ArticleIndex).
		Id(fmt.Sprintf("%d", articleID)).
		Do(ctx)

	return err
}

// SearchArticles ES搜索文章
func SearchArticles(keyword string, page, pageSize int) ([]ArticleDocument, int64, error) {
	if Client == nil {
		return nil, 0, fmt.Errorf("ES客户端未初始化")
	}

	ctx := context.Background()

	// 构建多字段查询（标题权重3，前言权重2，内容权重1）
	query := elastic.NewBoolQuery().Should(
		elastic.NewMatchQuery("title", keyword).Boost(3.0),
		elastic.NewMatchQuery("preface", keyword).Boost(2.0),
		elastic.NewMatchQuery("content", keyword).Boost(1.0),
	).MinimumShouldMatch("1")

	// 高亮设置
	highlight := elastic.NewHighlight().
		Field("title").
		Field("content").
		PreTags("<em>").
		PostTags("</em>")

	// 执行搜索
	from := (page - 1) * pageSize
	searchResult, err := Client.Search().
		Index(ArticleIndex).
		Query(query).
		Highlight(highlight).
		From(from).
		Size(pageSize).
		Sort("_score", false).      // 按相关度排序
		Sort("create_time", false). // 相关度相同时按时间排序
		Pretty(true).
		Do(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("ES搜索失败: %w", err)
	}

	// 解析结果
	var articles []ArticleDocument
	for _, hit := range searchResult.Hits.Hits {
		var article ArticleDocument
		if err := json.Unmarshal(hit.Source, &article); err != nil {
			continue
		}
		articles = append(articles, article)
	}

	total := searchResult.TotalHits()

	return articles, total, nil
}

// BulkIndexArticles 批量索引文章
func BulkIndexArticles(articles []model.Article) error {
	if Client == nil {
		return nil // ES未启用，静默失败
	}

	if len(articles) == 0 {
		return nil
	}

	ctx := context.Background()
	bulkRequest := Client.Bulk()

	for _, article := range articles {
		doc := ArticleDocument{
			ID:            article.ID,
			UserID:        article.UserID,
			Title:         article.Title,
			Preface:       article.Preface,
			Content:       article.Content,
			Photo:         article.Photo,
			Tag:           article.Tag,
			Visit:         int64(article.Visit),
			GoodCount:     int64(article.GoodCount),
			CommentCount:  int64(article.CommentCount),
			FavoriteCount: int64(article.FavoriteCount),
			CreateTime:    article.CreateTime,
			UpdateTime:    article.UpdateTime,
		}

		req := elastic.NewBulkIndexRequest().
			Index(ArticleIndex).
			Id(fmt.Sprintf("%d", article.ID)).
			Doc(doc)

		bulkRequest.Add(req)
	}

	bulkResponse, err := bulkRequest.Do(ctx)
	if err != nil {
		return fmt.Errorf("批量索引失败: %w", err)
	}

	if bulkResponse.Errors {
		return fmt.Errorf("批量索引部分失败")
	}

	fmt.Printf("✅ 成功索引 %d 篇文章\n", len(articles))
	return nil
}
