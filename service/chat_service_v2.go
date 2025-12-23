package service

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	wsLib "astronomer-gin/pkg/websocket"
	"astronomer-gin/repository"
	"fmt"
	"log"
	"time"
)

// ChatServiceV2 私信服务接口
type ChatServiceV2 interface {
	// 发送私信
	SendMessage(fromUserID, toUserID string, content string, messageType int) error

	// 获取聊天记录
	GetChatHistory(userID, targetUserID string, page, pageSize int) ([]model.UserChat, int64, error)

	// 获取会话列表（包含用户信息）
	GetChatSessions(userID string) ([]model.ChatSession, error)

	// 获取未读消息数
	GetUnreadCount(userID string) (int64, error)

	// 标记已读
	MarkAsRead(userID, fromUserID string) error

	// 删除私信
	DeleteMessage(messageID uint64, userID string) error
	DeleteChatWithUser(userID, targetUserID string) error
}

type chatServiceV2 struct {
	chatRepo   repository.ChatRepository
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
}

// NewChatServiceV2 创建私信服务V2实例
func NewChatServiceV2(
	chatRepo repository.ChatRepository,
	followRepo repository.FollowRepository,
	userRepo repository.UserRepository,
) ChatServiceV2 {
	return &chatServiceV2{
		chatRepo:   chatRepo,
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// SendMessage 发送私信（带WebSocket实时推送）
func (s *chatServiceV2) SendMessage(fromUserID, toUserID string, content string, messageType int) error {
	// 1. 参数验证
	if fromUserID == toUserID {
		return fmt.Errorf("不能给自己发送私信")
	}

	if content == "" {
		return constant.ErrParamInvalid
	}

	if len(content) > 1000 {
		return fmt.Errorf("消息内容不能超过1000字")
	}

	// 2. 检查接收者是否存在
	toUser, err := s.userRepo.FindByID(toUserID)
	if err != nil || toUser == nil {
		return fmt.Errorf("接收者不存在")
	}

	// 3. 检查是否被拉黑
	if s.followRepo.IsBlocked(toUserID, fromUserID) {
		return fmt.Errorf("对方已将你拉黑，无法发送私信")
	}

	// 4. 检查是否为好友（互相关注才能发私信，类似抖音）
	if !s.followRepo.IsFriend(fromUserID, toUserID) {
		return fmt.Errorf("只能给好友发送私信，请先互相关注")
	}

	// 5. 创建私信记录
	chat := &model.UserChat{
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		Content:     content,
		MessageType: messageType,
		IsRead:      false,
	}

	// 6. 保存到数据库
	if err := s.chatRepo.SendMessage(chat); err != nil {
		return err
	}

	// 7. 通过WebSocket实时推送给接收者（如果在线）
	hub := wsLib.GetHub()
	if hub.IsOnline(toUserID) {
		// 获取发送者信息
		fromUser, _ := s.userRepo.FindByID(fromUserID)

		message := &wsLib.WSMessage{
			Type: wsLib.MessageTypeNewMessage,
			Data: wsLib.ChatMessageData{
				ID:           chat.ID,
				FromUserID:   fromUserID,
				ToUserID:     toUserID,
				Content:      content,
				MessageType:  messageType,
				CreateTime:   chat.CreateTime,
				FromUsername: fromUser.Username,
				FromAvatar:   fromUser.Icon,
			},
			Timestamp: time.Now().Unix(),
		}

		if success := hub.SendToUser(toUserID, message); success {
			log.Printf("✅ 实时推送消息成功: %d -> %d", fromUserID, toUserID)
		} else {
			log.Printf("⚠️  实时推送消息失败（接收者不在线）: %d -> %d", fromUserID, toUserID)
		}
	}

	return nil
}

// GetChatHistory 获取聊天记录
func (s *chatServiceV2) GetChatHistory(userID, targetUserID string, page, pageSize int) ([]model.UserChat, int64, error) {
	// 参数验证
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > constant.MaxPageSize {
		pageSize = constant.DefaultPageSize
	}

	// 检查是否为好友
	if !s.followRepo.IsFriend(userID, targetUserID) {
		return nil, 0, fmt.Errorf("只能查看好友的聊天记录")
	}

	// 获取聊天记录
	chats, total, err := s.chatRepo.GetChatHistory(userID, targetUserID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 标记来自对方的消息为已读
	_ = s.chatRepo.MarkAsRead(userID, targetUserID)

	return chats, total, nil
}

// GetChatSessions 获取会话列表
func (s *chatServiceV2) GetChatSessions(userID string) ([]model.ChatSession, error) {
	// 获取会话列表
	sessions, err := s.chatRepo.GetChatSessions(userID)
	if err != nil {
		return nil, err
	}

	// 填充用户信息
	for i := range sessions {
		user, err := s.userRepo.FindByID(sessions[i].UserID)
		if err == nil && user != nil {
			sessions[i].Username = user.Username
			sessions[i].Avatar = user.Icon
		}
	}

	return sessions, nil
}

// GetUnreadCount 获取未读消息数
func (s *chatServiceV2) GetUnreadCount(userID string) (int64, error) {
	return s.chatRepo.GetUnreadCount(userID)
}

// MarkAsRead 标记已读
func (s *chatServiceV2) MarkAsRead(userID, fromUserID string) error {
	return s.chatRepo.MarkAsRead(userID, fromUserID)
}

// DeleteMessage 删除私信
func (s *chatServiceV2) DeleteMessage(messageID uint64, userID string) error {
	if messageID == 0 {
		return constant.ErrParamInvalid
	}
	return s.chatRepo.DeleteMessage(messageID, userID)
}

// DeleteChatWithUser 删除与某人的所有聊天记录
func (s *chatServiceV2) DeleteChatWithUser(userID, targetUserID string) error {
	if targetUserID != "" {
		return constant.ErrParamInvalid
	}
	return s.chatRepo.DeleteChatWithUser(userID, targetUserID)
}
