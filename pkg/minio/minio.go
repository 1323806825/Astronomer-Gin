package minio

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"astronomer-gin/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *MinIOClient

// MinIOClient MinIO客户端封装
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

// InitMinIO 初始化MinIO客户端
func InitMinIO(cfg *config.MinIOConfig) error {
	// 创建MinIO客户端
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// 初始化全局客户端
	Client = &MinIOClient{
		client:     minioClient,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}

	// 确保Bucket存在
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		// 创建Bucket
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		// 设置Bucket为公开读取
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, cfg.BucketName)

		err = minioClient.SetBucketPolicy(ctx, cfg.BucketName, policy)
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return nil
}

// UploadFile 上传文件
func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// 上传文件到MinIO
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// 返回文件的公开URL
	fileURL := fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucketName, objectName)
	return fileURL, nil
}

// DeleteFile 删除文件
func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileURL 获取文件的临时访问URL（带过期时间）
func (m *MinIOClient) GetFileURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// ListFiles 列出所有文件
func (m *MinIOClient) ListFiles(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// GetFileInfo 获取文件信息
func (m *MinIOClient) GetFileInfo(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	objInfo, err := m.client.StatObject(ctx, m.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}
	return objInfo, nil
}

// GenerateUniqueFileName 生成唯一文件名
func GenerateUniqueFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}

// GenerateObjectPath 生成对象存储路径（按日期分类）
func GenerateObjectPath(category string, filename string) string {
	now := time.Now()
	return fmt.Sprintf("%s/%d/%02d/%02d/%s", category, now.Year(), now.Month(), now.Day(), filename)
}
