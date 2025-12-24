package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"astronomer-gin/pkg/constant"
	minioPkg "astronomer-gin/pkg/minio"
	"astronomer-gin/pkg/queue"
	"astronomer-gin/pkg/redis"
	"astronomer-gin/pkg/util"

	"github.com/google/uuid"
)

// UploadServiceV2 企业级文件上传服务接口
type UploadServiceV2 interface {
	UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error)
	UploadImageWithVersions(ctx context.Context, file *multipart.FileHeader) (*ImageVersions, error)
	UploadFile(ctx context.Context, file *multipart.FileHeader, category string) (string, error)
	UploadMultiple(ctx context.Context, files []*multipart.FileHeader, category string) ([]string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}

type uploadServiceV2 struct {
	minioClient    *minioPkg.MinIOClient
	queueClient    *queue.RabbitMQClient
	cacheHelper    *util.CacheHelper
	imageProcessor *ImageProcessor
}

// NewUploadServiceV2 创建上传服务V2实例
func NewUploadServiceV2() UploadServiceV2 {
	service := &uploadServiceV2{
		minioClient: minioPkg.Client,
		queueClient: queue.Client,
		cacheHelper: util.NewCacheHelper(redis.GetClient()),
	}
	// 初始化图片处理器（传入自己的引用）
	service.imageProcessor = NewImageProcessor(service)
	return service
}

// UploadImage 上传图片（企业级实现）
func (s *uploadServiceV2) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// 1. 参数验证 - 文件类型
	if !isImageFile(file.Filename) {
		return "", constant.ErrInvalidImageFormat
	}

	// 2. 参数验证 - 文件大小（5MB）
	if file.Size > constant.MaxImageSize {
		return "", constant.ErrImageSizeExceeded
	}

	// 3. 上传文件
	return s.uploadFileToMinio(ctx, file, "images")
}

// UploadImageWithVersions 上传图片并生成多个版本
func (s *uploadServiceV2) UploadImageWithVersions(ctx context.Context, file *multipart.FileHeader) (*ImageVersions, error) {
	// 1. 参数验证 - 文件类型
	if !isImageFile(file.Filename) {
		return nil, constant.ErrInvalidImageFormat
	}

	// 2. 参数验证 - 文件大小（5MB）
	if file.Size > constant.MaxImageSize {
		return nil, constant.ErrImageSizeExceeded
	}

	// 3. 处理并上传多个版本
	versions, err := s.imageProcessor.ProcessAndUploadImage(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("处理图片失败: %w", err)
	}

	return versions, nil
}

// UploadFile 上传文件（企业级实现）
func (s *uploadServiceV2) UploadFile(ctx context.Context, file *multipart.FileHeader, category string) (string, error) {
	// 1. 参数验证 - 文件大小（20MB）
	if file.Size > constant.MaxFileSize {
		return "", constant.ErrFileSizeExceeded
	}

	// 2. 参数验证 - 分类
	if category == "" {
		category = "files"
	}

	// 3. 上传文件
	return s.uploadFileToMinio(ctx, file, category)
}

// UploadMultiple 批量上传文件（企业级实现）
func (s *uploadServiceV2) UploadMultiple(ctx context.Context, files []*multipart.FileHeader, category string) ([]string, error) {
	// 1. 参数验证
	if len(files) == 0 {
		return nil, constant.ErrNoFilesProvided
	}

	if len(files) > constant.MaxBatchUploadCount {
		return nil, constant.ErrTooManyFiles
	}

	// 2. 批量上传
	var fileURLs []string
	var errors []string

	for i, file := range files {
		fileURL, err := s.UploadFile(ctx, file, category)
		if err != nil {
			errors = append(errors, fmt.Sprintf("文件%d上传失败: %v", i+1, err))
			continue
		}
		fileURLs = append(fileURLs, fileURL)
	}

	// 3. 如果有失败的，返回详细错误
	if len(errors) > 0 {
		return fileURLs, fmt.Errorf("部分文件上传失败: %s", strings.Join(errors, "; "))
	}

	return fileURLs, nil
}

// uploadFileToMinio 上传文件到MinIO（内部方法）
func (s *uploadServiceV2) uploadFileToMinio(ctx context.Context, file *multipart.FileHeader, category string) (string, error) {
	// 1. 打开文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 2. 生成唯一文件名
	uniqueFilename := generateUniqueFilename(file.Filename)

	// 3. 生成对象路径
	objectName := minioPkg.GenerateObjectPath(category, uniqueFilename)

	// 4. 获取文件的Content-Type
	contentType := getContentType(file.Filename)

	// 5. 上传到MinIO
	fileURL, err := s.minioClient.UploadFile(ctx, objectName, src, file.Size, contentType)
	if err != nil {
		return "", constant.ErrUploadFailed
	}

	// 6. 如果是图片，发送异步任务生成缩略图（可选）
	if isImageFile(file.Filename) && s.queueClient != nil {
		go func() {
			task := queue.CreateTask(queue.TaskTypeImageProcess, map[string]interface{}{
				"object_name": objectName,
				"file_url":    fileURL,
				"action":      "generate_thumbnail",
			})
			_ = s.queueClient.PublishTask(context.Background(), task) // 忽略错误，不影响主流程
		}()
	}

	return fileURL, nil
}

// DeleteFile 删除文件（企业级实现）
func (s *uploadServiceV2) DeleteFile(ctx context.Context, fileURL string) error {
	// 1. 参数验证
	if fileURL == "" {
		return constant.ErrInvalidFileURL
	}

	// 2. 从URL中提取对象名
	objectName := extractObjectNameFromURL(fileURL)
	if objectName == "" {
		return constant.ErrInvalidFileURL
	}

	// 3. 从MinIO删除
	if err := s.minioClient.DeleteFile(ctx, objectName); err != nil {
		return constant.ErrDeleteFileFailed
	}

	return nil
}

// ==================== 辅助函数 ====================

// isImageFile 检查是否为图片文件
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	for _, validExt := range imageExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// getContentType 获取文件的Content-Type
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".zip":  "application/zip",
		".txt":  "text/plain",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}

// generateUniqueFilename 生成唯一文件名
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}

// extractObjectNameFromURL 从URL中提取对象名
func extractObjectNameFromURL(fileURL string) string {
	// 示例URL: http://localhost:9000/astronomer/images/2024/11/28/xxx.jpg
	// 需要提取: images/2024/11/28/xxx.jpg

	parts := strings.Split(fileURL, "/")
	if len(parts) < 3 {
		return ""
	}

	// 跳过协议、域名、bucket，取剩余部分
	for i, part := range parts {
		if part == "astronomer" && i+1 < len(parts) {
			return strings.Join(parts[i+1:], "/")
		}
	}

	return ""
}
