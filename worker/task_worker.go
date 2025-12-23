package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"astronomer-gin/pkg/queue"
)

// TaskWorker 异步任务消费者
type TaskWorker struct {
	mqClient    *queue.RabbitMQClient
	handler     TaskHandler
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	isRunning   bool
	mu          sync.Mutex
}

// TaskHandler 任务处理器接口
type TaskHandler interface {
	Handle(ctx context.Context, taskType string, data []byte) error
}

// Task 任务结构
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewTaskWorker 创建任务消费者
func NewTaskWorker(mqClient *queue.RabbitMQClient, handler TaskHandler, workerCount int) *TaskWorker {
	ctx, cancel := context.WithCancel(context.Background())

	if workerCount <= 0 {
		workerCount = 5 // 默认5个worker
	}

	return &TaskWorker{
		mqClient:    mqClient,
		handler:     handler,
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动任务消费者
func (w *TaskWorker) Start() error {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.isRunning = true
	w.mu.Unlock()

	log.Printf("Starting %d task workers...", w.workerCount)

	// 启动多个worker goroutine
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	log.Println("Task workers started successfully")
	return nil
}

// Stop 停止任务消费者
func (w *TaskWorker) Stop() {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	log.Println("Stopping task workers...")
	w.cancel()
	w.wg.Wait()

	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()

	log.Println("Task workers stopped")
}

// IsRunning 检查worker是否在运行
func (w *TaskWorker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isRunning
}

// worker 单个worker处理逻辑
func (w *TaskWorker) worker(id int) {
	defer w.wg.Done()

	log.Printf("Worker %d started", id)

	// 创建消费者通道
	messages, err := w.mqClient.Consume(w.ctx)
	if err != nil {
		log.Printf("Worker %d failed to start consuming: %v", id, err)
		return
	}

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return

		case msg, ok := <-messages:
			if !ok {
				log.Printf("Worker %d: message channel closed", id)
				return
			}

			// 处理消息
			w.processMessage(id, msg)
		}
	}
}

// processMessage 处理单个消息
func (w *TaskWorker) processMessage(workerID int, msg []byte) {
	startTime := time.Now()

	// 解析任务
	var task Task
	if err := json.Unmarshal(msg, &task); err != nil {
		log.Printf("Worker %d: failed to unmarshal task: %v", workerID, err)
		return
	}

	log.Printf("Worker %d: processing task %s (type: %s)", workerID, task.ID, task.Type)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	// 调用处理器处理任务
	if err := w.handler.Handle(ctx, task.Type, msg); err != nil {
		log.Printf("Worker %d: task %s failed: %v", workerID, task.ID, err)
		// 这里可以根据错误类型决定是否重试
		return
	}

	duration := time.Since(startTime)
	log.Printf("Worker %d: task %s completed in %v", workerID, task.ID, duration)
}

// DefaultTaskHandler 默认任务处理器
type DefaultTaskHandler struct{}

// Handle 实现TaskHandler接口
func (h *DefaultTaskHandler) Handle(ctx context.Context, taskType string, data []byte) error {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	switch taskType {
	case "file_uploaded":
		return h.handleFileUploaded(ctx, task)
	case "email_notification":
		return h.handleEmailNotification(ctx, task)
	case "data_sync":
		return h.handleDataSync(ctx, task)
	default:
		log.Printf("Unknown task type: %s", taskType)
		return nil
	}
}

// handleFileUploaded 处理文件上传完成任务
func (h *DefaultTaskHandler) handleFileUploaded(ctx context.Context, task Task) error {
	log.Printf("Processing file uploaded task: %v", task.Data)

	// 这里可以添加实际的业务逻辑，例如：
	// - 图片压缩
	// - 生成缩略图
	// - 病毒扫描
	// - 更新数据库记录
	// - 发送通知

	filename, _ := task.Data["filename"].(string)
	objectName, _ := task.Data["object_name"].(string)

	log.Printf("File uploaded: %s (stored as: %s)", filename, objectName)

	return nil
}

// handleEmailNotification 处理邮件通知任务
func (h *DefaultTaskHandler) handleEmailNotification(ctx context.Context, task Task) error {
	log.Printf("Processing email notification task: %v", task.Data)

	// 这里可以添加发送邮件的逻辑

	return nil
}

// handleDataSync 处理数据同步任务
func (h *DefaultTaskHandler) handleDataSync(ctx context.Context, task Task) error {
	log.Printf("Processing data sync task: %v", task.Data)

	// 这里可以添加数据同步的逻辑

	return nil
}
