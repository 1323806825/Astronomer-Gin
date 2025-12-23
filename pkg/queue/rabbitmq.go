package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"astronomer-gin/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Client *RabbitMQClient

// RabbitMQClient RabbitMQ客户端封装
type RabbitMQClient struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	queueName    string
	routingKey   string
}

// TaskType 任务类型
type TaskType string

const (
	TaskTypeNotification TaskType = "notification" // 通知任务
	TaskTypeEmail        TaskType = "email"        // 邮件任务
	TaskTypeSMS          TaskType = "sms"          // 短信任务
	TaskTypeImageProcess TaskType = "image"        // 图片处理任务
	TaskTypeStats        TaskType = "stats"        // 统计任务
)

// Task 任务消息结构
type Task struct {
	ID        string                 `json:"id"`         // 任务ID
	Type      TaskType               `json:"type"`       // 任务类型
	Data      map[string]interface{} `json:"data"`       // 任务数据
	Retry     int                    `json:"retry"`      // 已重试次数
	MaxRetry  int                    `json:"max_retry"`  // 最大重试次数
	CreatedAt time.Time              `json:"created_at"` // 创建时间
}

// InitRabbitMQ 初始化RabbitMQ客户端
func InitRabbitMQ(cfg *config.RabbitMQConfig) error {
	// 连接RabbitMQ
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 创建channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// 声明exchange
	err = ch.ExchangeDeclare(
		cfg.ExchangeName, // name
		"direct",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// 声明queue
	q, err := ch.QueueDeclare(
		cfg.QueueName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 绑定queue到exchange
	err = ch.QueueBind(
		q.Name,           // queue name
		cfg.RoutingKey,   // routing key
		cfg.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// 设置Qos
	err = ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	Client = &RabbitMQClient{
		conn:         conn,
		channel:      ch,
		exchangeName: cfg.ExchangeName,
		queueName:    cfg.QueueName,
		routingKey:   cfg.RoutingKey,
	}

	log.Println("RabbitMQ客户端初始化成功")
	return nil
}

// PublishTask 发布任务到队列
func (r *RabbitMQClient) PublishTask(ctx context.Context, task *Task) error {
	// 序列化任务
	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 发布消息
	err = r.channel.PublishWithContext(ctx,
		r.exchangeName, // exchange
		r.routingKey,   // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	log.Printf("Task published: %s (type: %s)", task.ID, task.Type)
	return nil
}

// TaskHandler 任务处理器函数类型
type TaskHandler func(task *Task) error

// Consume 消费消息（返回消息通道）
func (r *RabbitMQClient) Consume(ctx context.Context) (<-chan []byte, error) {
	msgs, err := r.channel.Consume(
		r.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}

	// 创建字节切片通道
	byteChan := make(chan []byte)

	// 在goroutine中转换消息
	go func() {
		defer close(byteChan)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				// 发送消息体到通道
				select {
				case byteChan <- d.Body:
					// 自动确认消息
					d.Ack(false)
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return byteChan, nil
}

// ConsumeTask 消费任务（阻塞方法）
func (r *RabbitMQClient) ConsumeTask(handler TaskHandler) error {
	msgs, err := r.channel.Consume(
		r.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	log.Printf("开始消费队列: %s", r.queueName)

	// 处理消息
	for d := range msgs {
		var task Task
		err := json.Unmarshal(d.Body, &task)
		if err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			d.Nack(false, false) // 拒绝消息，不重新入队
			continue
		}

		log.Printf("Processing task: %s (type: %s)", task.ID, task.Type)

		// 调用处理器
		err = handler(&task)
		if err != nil {
			log.Printf("Failed to process task %s: %v", task.ID, err)

			// 检查是否需要重试
			if task.Retry < task.MaxRetry {
				task.Retry++
				// 重新发布任务
				ctx := context.Background()
				if publishErr := r.PublishTask(ctx, &task); publishErr != nil {
					log.Printf("Failed to republish task: %v", publishErr)
				}
			}

			d.Nack(false, false) // 拒绝消息
		} else {
			log.Printf("Task processed successfully: %s", task.ID)
			d.Ack(false) // 确认消息
		}
	}

	return nil
}

// Close 关闭连接
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// CreateTask 创建任务（辅助函数）
func CreateTask(taskType TaskType, data map[string]interface{}) *Task {
	return &Task{
		ID:        generateTaskID(),
		Type:      taskType,
		Data:      data,
		Retry:     0,
		MaxRetry:  3,
		CreatedAt: time.Now(),
	}
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
