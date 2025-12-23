package chat

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	wsLib "astronomer-gin/pkg/websocket"
	"astronomer-gin/service"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该验证Origin
		return true
	},
}

type ChatHandler struct {
	chatService service.ChatServiceV2
	userService service.UserServiceV2
}

func NewChatHandler(chatService service.ChatServiceV2, userService service.UserServiceV2) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		userService: userService,
	}
}

// HandleWebSocket WebSocket连接处理
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	phone, exists := c.Get("phone")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	client := &wsLib.Client{
		UserID:     user.ID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Hub:        wsLib.GetHub(),
		LastActive: time.Now(),
	}

	// 注册客户端
	client.Hub.Register(client)

	// 启动读写协程
	client.Start()

	log.Printf("🔌 WebSocket连接建立: 用户 %d (%s)", user.ID, user.Username)
}

// SendMessage 发送私信
func (h *ChatHandler) SendMessage(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取请求参数
	var req struct {
		ToUserID    string `json:"to_user_id" binding:"required"`
		Content     string `json:"content" binding:"required"`
		MessageType int    `json:"message_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 默认消息类型为文本
	if req.MessageType == 0 {
		req.MessageType = 1
	}

	// 发送私信
	if err := h.chatService.SendMessage(user.ID, req.ToUserID, req.Content, req.MessageType); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "发送成功", nil)
}

// GetChatHistory 获取与某人的聊天记录
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取聊天记录
	chats, total, err := h.chatService.GetChatHistory(user.ID, targetUserID, page, pageSize)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list":     chats,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetChatSessions 获取会话列表（所有聊过天的人）
func (h *ChatHandler) GetChatSessions(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取会话列表
	sessions, err := h.chatService.GetChatSessions(user.ID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"list": sessions,
	})
}

// GetUnreadCount 获取未读消息数
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	phone, _ := c.Get("phone")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 获取未读消息数
	count, err := h.chatService.GetUnreadCount(user.ID)
	if err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.Success(c, gin.H{
		"count": count,
	})
}

// MarkAsRead 标记消息已读
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	phone, _ := c.Get("phone")
	fromUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 标记已读
	if err := h.chatService.MarkAsRead(user.ID, fromUserID); err != nil {
		util.InternalServerError(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "标记成功", nil)
}

// DeleteMessage 删除单条私信
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	phone, _ := c.Get("phone")
	messageID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	msgID, _ := strconv.ParseUint(messageID, 10, 64)

	// 删除私信
	if err := h.chatService.DeleteMessage(msgID, user.ID); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除成功", nil)
}

// DeleteChatWithUser 删除与某人的所有聊天记录
func (h *ChatHandler) DeleteChatWithUser(c *gin.Context) {
	phone, _ := c.Get("phone")
	targetUserID := c.Param("id")

	// 获取当前用户信息
	user, err := h.userService.GetUserInfo(phone.(string))
	if err != nil {
		util.NotFound(c, constant.UserNotExist)
		return
	}

	// 删除聊天记录
	if err := h.chatService.DeleteChatWithUser(user.ID, targetUserID); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除成功", nil)
}
