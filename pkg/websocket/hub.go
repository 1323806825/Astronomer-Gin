package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client WebSocket客户端
type Client struct {
	UserID     string          // 用户ID
	Conn       *websocket.Conn // WebSocket连接
	Send       chan []byte     // 发送消息通道
	Hub        *Hub            // 连接管理器
	mu         sync.Mutex      // 保护并发写入
	LastActive time.Time       // 最后活跃时间
}

// Hub WebSocket连接管理器（单例）
type Hub struct {
	clients    map[string]*Client // 用户ID -> 客户端映射
	broadcast  chan *WSMessage    // 广播消息
	register   chan *Client       // 注册客户端
	unregister chan *Client       // 注销客户端
	mu         sync.RWMutex       // 保护clients map
}

var (
	instance *Hub
	once     sync.Once
)

// GetHub 获取Hub单例
func GetHub() *Hub {
	once.Do(func() {
		instance = &Hub{
			clients:    make(map[string]*Client),
			broadcast:  make(chan *WSMessage, 256),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		go instance.Run()
	})
	return instance
}

// Run 启动Hub
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second) // 30秒清理一次
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 如果用户已经有连接，关闭旧连接
			if oldClient, exists := h.clients[client.UserID]; exists {
				close(oldClient.Send)
				oldClient.Conn.Close()
				log.Printf("⚠️  用户 %d 已有连接，关闭旧连接", client.UserID)
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("✅ 用户 %d 上线，当前在线人数: %d", client.UserID, len(h.clients))

			// 通知好友该用户上线
			h.NotifyOnlineStatus(client.UserID, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, exists := h.clients[client.UserID]; exists {
				delete(h.clients, client.UserID)
				close(client.Send)
				log.Printf("👋 用户 %d 下线，当前在线人数: %d", client.UserID, len(h.clients))
			}
			h.mu.Unlock()

			// 通知好友该用户下线
			h.NotifyOnlineStatus(client.UserID, false)

		case message := <-h.broadcast:
			// 广播消息（暂时未使用，预留）
			h.BroadcastToAll(message)

		case <-ticker.C:
			// 定期清理超时连接（5分钟无活动）
			h.CleanupInactiveClients(5 * time.Minute)
		}
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SendToUser 向指定用户发送消息
func (h *Hub) SendToUser(userID string, message *WSMessage) bool {
	h.mu.RLock()
	client, exists := h.clients[userID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 序列化消息失败: %v", err)
		return false
	}

	select {
	case client.Send <- data:
		return true
	default:
		log.Printf("⚠️  用户 %d 发送队列已满", userID)
		return false
	}
}

// BroadcastToAll 广播消息给所有在线用户
func (h *Hub) BroadcastToAll(message *WSMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 序列化广播消息失败: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.clients {
		select {
		case client.Send <- data:
		default:
			log.Printf("⚠️  广播到用户 %d 失败，发送队列已满", userID)
		}
	}
}

// IsOnline 检查用户是否在线
func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.clients[userID]
	return exists
}

// GetOnlineCount 获取在线人数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetOnlineUsers 获取所有在线用户ID
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// NotifyOnlineStatus 通知好友用户上线/下线状态
func (h *Hub) NotifyOnlineStatus(userID string, isOnline bool) {
	// TODO: 查询用户的好友列表，向好友推送上线/下线通知
	// 这里需要注入 FollowRepository，暂时省略
	message := &WSMessage{
		Type: MessageTypeOnline,
		Data: OnlineStatusData{
			UserID:   userID,
			IsOnline: isOnline,
		},
		Timestamp: time.Now().Unix(),
	}

	if !isOnline {
		message.Type = MessageTypeOffline
	}

	// 实际应该只通知好友，这里暂时不实现
	log.Printf("📢 用户 %d %s", userID, map[bool]string{true: "上线", false: "下线"}[isOnline])
}

// CleanupInactiveClients 清理不活跃的客户端
func (h *Hub) CleanupInactiveClients(timeout time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for userID, client := range h.clients {
		if now.Sub(client.LastActive) > timeout {
			log.Printf("🧹 清理不活跃连接: 用户 %d", userID)
			client.Conn.Close()
			delete(h.clients, userID)
			close(client.Send)
		}
	}
}

// writePump 写入泵（从Send通道读取并发送到WebSocket）
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second) // 心跳间隔
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Send通道已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("❌ 写入WebSocket失败: %v", err)
				return
			}

		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			pong := &WSMessage{
				Type:      MessageTypePong,
				Timestamp: time.Now().Unix(),
			}
			data, _ := json.Marshal(pong)
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

// readPump 读取泵（从WebSocket读取并处理消息）
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	// 设置读取参数
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket读取错误: %v", err)
			}
			break
		}

		// 更新活跃时间
		c.LastActive = time.Now()

		// 处理接收到的消息
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("❌ 解析WebSocket消息失败: %v", err)
			continue
		}

		// 处理不同类型的消息
		c.handleMessage(&wsMsg)
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(msg *WSMessage) {
	switch msg.Type {
	case MessageTypePing:
		// 响应心跳
		pong := &WSMessage{
			Type:      MessageTypePong,
			Timestamp: time.Now().Unix(),
		}
		data, _ := json.Marshal(pong)
		c.Send <- data

	case MessageTypeTyping:
		// 正在输入，转发给目标用户
		if typingData, ok := msg.Data.(map[string]interface{}); ok {
			if toUserID, ok := typingData["to_user_id"].(string); ok {
				c.Hub.SendToUser(toUserID, &WSMessage{
					Type: MessageTypeTyping,
					Data: TypingData{
						UserID:   c.UserID,
						ToUserID: toUserID,
						IsTyping: typingData["is_typing"].(bool),
					},
					Timestamp: time.Now().Unix(),
				})
			}
		}

	case MessageTypeRead:
		// 已读回执，转发给发送者
		if readData, ok := msg.Data.(map[string]interface{}); ok {
			if fromUserID, ok := readData["from_user_id"].(string); ok {
				c.Hub.SendToUser(fromUserID, &WSMessage{
					Type: MessageTypeReadReceipt,
					Data: ReadData{
						UserID:     c.UserID,
						FromUserID: fromUserID,
					},
					Timestamp: time.Now().Unix(),
				})
			}
		}

	default:
		log.Printf("⚠️  未知消息类型: %s", msg.Type)
	}
}

// Start 启动客户端读写协程
func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}
