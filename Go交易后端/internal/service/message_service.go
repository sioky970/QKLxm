package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// MessageService 站内信服务
type MessageService struct{}

// NewMessageService 创建站内信服务实例
func NewMessageService() *MessageService {
	return &MessageService{}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       int8   `json:"type"`
	Priority   int8   `json:"priority"`
	ReceiverID uint   `json:"receiver_id"`           // 单个接收者（单发）
	ReceiverIDs []uint `json:"receiver_ids"`         // 多个接收者（群发）
	Attachment string `json:"attachment,omitempty"`
}

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	Page     int  `json:"page" form:"page"`
	PageSize int  `json:"page_size" form:"page_size"`
	Type     *int8 `json:"type" form:"type"`         // 消息类型筛选
	IsRead   *int8 `json:"is_read" form:"is_read"`   // 已读状态筛选
	Priority *int8 `json:"priority" form:"priority"` // 优先级筛选
}

// MessageDetailResponse 消息详情响应
type MessageDetailResponse struct {
	model.Message
	IsRead   int8  `json:"is_read"`
	ReadTime int64 `json:"read_time"`
}

// SendMessage 发送站内信（单发）
func (s *MessageService) SendMessage(senderID uint, senderType int8, req *SendMessageRequest) error {
	if req.ReceiverID == 0 {
		return errors.New("接收者ID不能为空")
	}

	// 创建消息记录
	message := model.Message{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		SenderID:   senderID,
		SenderType: senderType,
		IsBatch:    0,
		Attachment: req.Attachment,
		Status:     1,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存消息
	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存消息失败: %v", err)
	}

	// 创建用户消息关联
	userMessage := model.UserMessage{
		UserID:     req.ReceiverID,
		MessageID:  message.ID,
		IsRead:     0,
		IsDeleted:  0,
		ReadTime:   0,
		CreateTime: time.Now().Unix(),
	}

	if err := tx.Create(&userMessage).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建用户消息关联失败: %v", err)
	}

	return tx.Commit().Error
}

// SendBatchMessage 批量发送站内信（群发）
func (s *MessageService) SendBatchMessage(senderID uint, senderType int8, req *SendMessageRequest) error {
	if len(req.ReceiverIDs) == 0 {
		return errors.New("接收者列表不能为空")
	}

	// 创建消息记录
	message := model.Message{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		SenderID:   senderID,
		SenderType: senderType,
		IsBatch:    1,
		Attachment: req.Attachment,
		Status:     1,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存消息
	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存消息失败: %v", err)
	}

	// 批量创建用户消息关联
	userMessages := make([]model.UserMessage, 0, len(req.ReceiverIDs))
	createTime := time.Now().Unix()

	for _, receiverID := range req.ReceiverIDs {
		userMessages = append(userMessages, model.UserMessage{
			UserID:     receiverID,
			MessageID:  message.ID,
			IsRead:     0,
			IsDeleted:  0,
			ReadTime:   0,
			CreateTime: createTime,
		})
	}

	// 批量插入
	if len(userMessages) > 0 {
		if err := tx.CreateInBatches(userMessages, 100).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("批量创建用户消息关联失败: %v", err)
		}
	}

	return tx.Commit().Error
}

// SendToAllUsers 发送给所有用户
func (s *MessageService) SendToAllUsers(senderID uint, senderType int8, req *SendMessageRequest) error {
	// 查询所有用户ID
	var userIDs []uint
	if err := database.DB.Model(&model.User{}).Pluck("id", &userIDs).Error; err != nil {
		return fmt.Errorf("查询用户列表失败: %v", err)
	}

	if len(userIDs) == 0 {
		return errors.New("没有找到用户")
	}

	// 使用批量发送
	req.ReceiverIDs = userIDs
	return s.SendBatchMessage(senderID, senderType, req)
}

// GetUserMessageList 获取用户的站内信列表
func (s *MessageService) GetUserMessageList(userID uint, req *MessageListRequest) ([]MessageDetailResponse, int64, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize

	// 构建查询
	query := database.DB.Table("user_messages um").
		Select("m.*, um.is_read, um.read_time").
		Joins("INNER JOIN messages m ON um.message_id = m.id").
		Where("um.user_id = ? AND um.is_deleted = 0 AND m.status = 1", userID)

	// 条件筛选
	if req.Type != nil {
		query = query.Where("m.type = ?", *req.Type)
	}
	if req.IsRead != nil {
		query = query.Where("um.is_read = ?", *req.IsRead)
	}
	if req.Priority != nil {
		query = query.Where("m.priority = ?", *req.Priority)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	var messages []MessageDetailResponse
	err := query.Order("m.priority DESC, m.create_time DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

// GetMessageDetail 获取消息详情并标记为已读
func (s *MessageService) GetMessageDetail(userID, messageID uint) (*MessageDetailResponse, error) {
	var result MessageDetailResponse

	// 查询消息详情
	err := database.DB.Table("user_messages um").
		Select("m.*, um.is_read, um.read_time").
		Joins("INNER JOIN messages m ON um.message_id = m.id").
		Where("um.user_id = ? AND um.message_id = ? AND um.is_deleted = 0 AND m.status = 1", userID, messageID).
		First(&result).Error

	if err != nil {
		return nil, err
	}

	// 如果未读，标记为已读
	if result.IsRead == 0 {
		s.MarkAsRead(userID, messageID)
	}

	return &result, nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(userID, messageID uint) error {
	return database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Updates(map[string]interface{}{
			"is_read":   1,
			"read_time": time.Now().Unix(),
		}).Error
}

// MarkAsUnread 标记消息为未读
func (s *MessageService) MarkAsUnread(userID, messageID uint) error {
	return database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Updates(map[string]interface{}{
			"is_read":   0,
			"read_time": 0,
		}).Error
}

// BatchMarkAsRead 批量标记为已读
func (s *MessageService) BatchMarkAsRead(userID uint, messageIDs []uint) error {
	if len(messageIDs) == 0 {
		return errors.New("消息ID列表不能为空")
	}

	return database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id IN ?", userID, messageIDs).
		Updates(map[string]interface{}{
			"is_read":   1,
			"read_time": time.Now().Unix(),
		}).Error
}

// MarkAllAsRead 标记所有消息为已读
func (s *MessageService) MarkAllAsRead(userID uint) error {
	return database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND is_read = 0", userID).
		Updates(map[string]interface{}{
			"is_read":   1,
			"read_time": time.Now().Unix(),
		}).Error
}

// DeleteMessage 删除消息（软删除）
func (s *MessageService) DeleteMessage(userID, messageID uint) error {
	return database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Update("is_deleted", 1).Error
}

// BatchDeleteMessages 批量删除消息
func (s *MessageService) BatchDeleteMessages(userID uint, messageIDs []uint) error {
	if len(messageIDs) == 0 {
		return errors.New("消息ID列表不能为空")
	}

	return database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id IN ?", userID, messageIDs).
		Update("is_deleted", 1).Error
}

// GetUnreadCount 获取未读消息数量
func (s *MessageService) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND is_read = 0 AND is_deleted = 0", userID).
		Count(&count).Error
	return count, err
}

// GetMessageStatistics 获取消息统计信息
func (s *MessageService) GetMessageStatistics(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总消息数
	var total int64
	database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND is_deleted = 0", userID).
		Count(&total)
	stats["total"] = total

	// 未读消息数
	var unread int64
	database.DB.Model(&model.UserMessage{}).
		Where("user_id = ? AND is_read = 0 AND is_deleted = 0", userID).
		Count(&unread)
	stats["unread"] = unread

	// 已读消息数
	stats["read"] = total - unread

	// 按类型统计
	type TypeCount struct {
		Type  int8  `json:"type"`
		Count int64 `json:"count"`
	}
	var typeCounts []TypeCount
	database.DB.Table("user_messages um").
		Select("m.type, COUNT(*) as count").
		Joins("INNER JOIN messages m ON um.message_id = m.id").
		Where("um.user_id = ? AND um.is_deleted = 0 AND m.status = 1", userID).
		Group("m.type").
		Scan(&typeCounts)
	stats["by_type"] = typeCounts

	return stats, nil
}

// SendFromTemplate 使用模板发送消息
func (s *MessageService) SendFromTemplate(senderID uint, senderType int8, templateCode string, variables map[string]string, receiverIDs []uint) error {
	// 查询模板
	var template model.MessageTemplate
	if err := database.DB.Where("code = ? AND status = 1", templateCode).First(&template).Error; err != nil {
		return fmt.Errorf("模板不存在或已禁用: %v", err)
	}

	// 替换变量
	title := template.Title
	content := template.Content
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		title = strings.ReplaceAll(title, placeholder, value)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	// 发送消息
	req := &SendMessageRequest{
		Title:       title,
		Content:     content,
		Type:        template.Type,
		Priority:    0,
		ReceiverIDs: receiverIDs,
	}

	if len(receiverIDs) == 1 {
		req.ReceiverID = receiverIDs[0]
		return s.SendMessage(senderID, senderType, req)
	}

	return s.SendBatchMessage(senderID, senderType, req)
}

