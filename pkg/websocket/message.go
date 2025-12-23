package websocket

import "time"

// MessageType WebSocket消息类型
const (
	// 客户端 -> 服务端
	MessageTypeChat   = "chat"   // 发送消息（通过HTTP API，WS仅接收）
	MessageTypeTyping = "typing" // 正在输入
	MessageTypeRead   = "read"   // 已读消息
	MessageTypePing   = "ping"   // 心跳ping
	MessageTypeAuth   = "auth"   // 认证（如果需要）

	// 服务端 -> 客户端
	MessageTypeNewMessage  = "new_message"  // 新消息通知
	MessageTypeOnline      = "online"       // 用户上线
	MessageTypeOffline     = "offline"      // 用户下线
	MessageTypeReadReceipt = "read_receipt" // 已读回执
	MessageTypePong        = "pong"         // 心跳pong
	MessageTypeError       = "error"        // 错误消息
)

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type      string      `json:"type"`                 // 消息类型
	Data      interface{} `json:"data"`                 // 消息数据
	Timestamp int64       `json:"timestamp,omitempty"`  // 时间戳
	MessageID string      `json:"message_id,omitempty"` // 消息ID（用于追踪）
}

// ChatMessageData 聊天消息数据
type ChatMessageData struct {
	ID          uint64    `json:"id"`           // 消息ID
	FromUserID  string    `json:"from_user_id"` // 发送者ID
	ToUserID    string    `json:"to_user_id"`   // 接收者ID
	Content     string    `json:"content"`      // 消息内容
	MessageType int       `json:"message_type"` // 消息类型（1-文本 2-图片 3-语音）
	CreateTime  time.Time `json:"create_time"`  // 发送时间

	// 发送者信息（冗余，便于前端显示）
	FromUsername string `json:"from_username,omitempty"`
	FromAvatar   string `json:"from_avatar,omitempty"`
}

// TypingData 正在输入数据
type TypingData struct {
	UserID   string `json:"user_id"`    // 用户ID
	ToUserID string `json:"to_user_id"` // 接收者ID
	IsTyping bool   `json:"is_typing"`  // 是否正在输入
}

// ReadData 已读数据
type ReadData struct {
	UserID     string   `json:"user_id"`               // 读取者ID
	FromUserID string   `json:"from_user_id"`          // 消息发送者ID
	MessageIDs []uint64 `json:"message_ids,omitempty"` // 已读的消息ID列表
}

// OnlineStatusData 在线状态数据
type OnlineStatusData struct {
	UserID   string `json:"user_id"`   // 用户ID
	IsOnline bool   `json:"is_online"` // 是否在线
}

// ErrorData 错误数据
type ErrorData struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
}
