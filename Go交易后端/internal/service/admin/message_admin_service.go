package admin

import (
	"errors"
	"fmt"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// MessageAdminService 站内信管理服务
type MessageAdminService struct{}

// NewMessageAdminService 创建站内信管理服务实例
func NewMessageAdminService() *MessageAdminService {
	return &MessageAdminService{}
}

// MessageListQueryRequest 管理端消息列表查询请求
type MessageListQueryRequest struct {
	PageInfo
	Type       *int8  `json:"type" form:"type"`
	Status     *int8  `json:"status" form:"status"`
	Priority   *int8  `json:"priority" form:"priority"`
	IsBatch    *int8  `json:"is_batch" form:"is_batch"`
	Keyword    string `json:"keyword" form:"keyword"`         // 标题关键词搜索
	StartTime  int64  `json:"start_time" form:"start_time"`
	EndTime    int64  `json:"end_time" form:"end_time"`
}

// MessageStatisticsResponse 消息统计响应
type MessageStatisticsResponse struct {
	TotalMessages   int64            `json:"total_messages"`
	TotalReceivers  int64            `json:"total_receivers"`
	ReadCount       int64            `json:"read_count"`
	UnreadCount     int64            `json:"unread_count"`
	ReadRate        float64          `json:"read_rate"`
	TypeStatistics  []TypeStatistics `json:"type_statistics"`
}

// TypeStatistics 类型统计
type TypeStatistics struct {
	Type  int8  `json:"type"`
	Count int64 `json:"count"`
}

// ReceiverStatistics 接收者统计
type ReceiverStatistics struct {
	MessageID     uint  `json:"message_id"`
	TotalCount    int64 `json:"total_count"`
	ReadCount     int64 `json:"read_count"`
	UnreadCount   int64 `json:"unread_count"`
	ReadRate      float64 `json:"read_rate"`
}

// GetMessageList 获取消息列表
func (s *MessageAdminService) GetMessageList(req MessageListQueryRequest) ([]model.Message, int64, error) {
	db := database.DB.Model(&model.Message{})

	// 条件筛选
	if req.Type != nil {
		db = db.Where("type = ?", *req.Type)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if req.Priority != nil {
		db = db.Where("priority = ?", *req.Priority)
	}
	if req.IsBatch != nil {
		db = db.Where("is_batch = ?", *req.IsBatch)
	}
	if req.Keyword != "" {
		db = db.Where("title LIKE ?", "%"+req.Keyword+"%")
	}
	if req.StartTime > 0 {
		db = db.Where("create_time >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		db = db.Where("create_time <= ?", req.EndTime)
	}

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	var messages []model.Message
	err := db.Order("create_time DESC").
		Limit(req.GetLimit()).
		Offset(req.GetOffset()).
		Find(&messages).Error

	return messages, total, err
}

// GetMessageDetail 获取消息详情
func (s *MessageAdminService) GetMessageDetail(messageID uint) (*model.Message, error) {
	var message model.Message
	err := database.DB.Where("id = ?", messageID).First(&message).Error
	return &message, err
}

// CreateMessage 创建消息（管理端）
func (s *MessageAdminService) CreateMessage(message *model.Message, receiverIDs []uint) error {
	if message.Title == "" || message.Content == "" {
		return errors.New("标题和内容不能为空")
	}

	if len(receiverIDs) == 0 {
		return errors.New("接收者列表不能为空")
	}

	// 设置时间
	message.CreateTime = time.Now().Unix()
	message.UpdateTime = time.Now().Unix()
	message.Status = 1

	// 判断是否群发
	if len(receiverIDs) > 1 {
		message.IsBatch = 1
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存消息
	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存消息失败: %v", err)
	}

	// 批量创建用户消息关联
	userMessages := make([]model.UserMessage, 0, len(receiverIDs))
	createTime := time.Now().Unix()

	for _, receiverID := range receiverIDs {
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
	if err := tx.CreateInBatches(userMessages, 100).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("批量创建用户消息关联失败: %v", err)
	}

	return tx.Commit().Error
}

// UpdateMessage 更新消息
func (s *MessageAdminService) UpdateMessage(message *model.Message) error {
	message.UpdateTime = time.Now().Unix()
	
	// 只更新允许修改的字段
	return database.DB.Model(&model.Message{}).
		Where("id = ?", message.ID).
		Updates(map[string]interface{}{
			"title":       message.Title,
			"content":     message.Content,
			"type":        message.Type,
			"priority":    message.Priority,
			"attachment":  message.Attachment,
			"update_time": message.UpdateTime,
		}).Error
}

// DeleteMessage 删除消息（硬删除）
func (s *MessageAdminService) DeleteMessage(messageID uint) error {
	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户消息关联
	if err := tx.Where("message_id = ?", messageID).Delete(&model.UserMessage{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除用户消息关联失败: %v", err)
	}

	// 删除消息
	if err := tx.Delete(&model.Message{}, messageID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除消息失败: %v", err)
	}

	return tx.Commit().Error
}

// WithdrawMessage 撤回消息
func (s *MessageAdminService) WithdrawMessage(messageID uint) error {
	return database.DB.Model(&model.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"status":      2,
			"update_time": time.Now().Unix(),
		}).Error
}

// GetReceiverList 获取消息的接收者列表
func (s *MessageAdminService) GetReceiverList(messageID uint, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 统计总数
	var total int64
	database.DB.Model(&model.UserMessage{}).
		Where("message_id = ?", messageID).
		Count(&total)

	// 查询接收者列表
	type ReceiverInfo struct {
		UserID     uint   `json:"user_id"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		Phone      string `json:"phone"`
		IsRead     int8   `json:"is_read"`
		ReadTime   int64  `json:"read_time"`
		CreateTime int64  `json:"create_time"`
	}

	var receivers []ReceiverInfo
	err := database.DB.Table("user_messages um").
		Select("um.user_id, u.account as username, u.email, u.phone, um.is_read, um.read_time, um.create_time").
		Joins("LEFT JOIN users u ON um.user_id = u.id").
		Where("um.message_id = ?", messageID).
		Order("um.create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&receivers).Error

	// 转换为map
	result := make([]map[string]interface{}, len(receivers))
	for i, r := range receivers {
		result[i] = map[string]interface{}{
			"user_id":     r.UserID,
			"username":    r.Username,
			"email":       r.Email,
			"phone":       r.Phone,
			"is_read":     r.IsRead,
			"read_time":   r.ReadTime,
			"create_time": r.CreateTime,
		}
	}

	return result, total, err
}

// GetMessageStatistics 获取消息统计信息
func (s *MessageAdminService) GetMessageStatistics() (*MessageStatisticsResponse, error) {
	stats := &MessageStatisticsResponse{}

	// 总消息数
	database.DB.Model(&model.Message{}).
		Where("status = 1").
		Count(&stats.TotalMessages)

	// 总接收者数
	database.DB.Model(&model.UserMessage{}).
		Where("is_deleted = 0").
		Count(&stats.TotalReceivers)

	// 已读数
	database.DB.Model(&model.UserMessage{}).
		Where("is_read = 1 AND is_deleted = 0").
		Count(&stats.ReadCount)

	// 未读数
	stats.UnreadCount = stats.TotalReceivers - stats.ReadCount

	// 计算阅读率
	if stats.TotalReceivers > 0 {
		stats.ReadRate = float64(stats.ReadCount) / float64(stats.TotalReceivers) * 100
	}

	// 按类型统计
	database.DB.Model(&model.Message{}).
		Select("type, COUNT(*) as count").
		Where("status = 1").
		Group("type").
		Scan(&stats.TypeStatistics)

	return stats, nil
}

// GetMessageReceiverStatistics 获取单条消息的接收统计
func (s *MessageAdminService) GetMessageReceiverStatistics(messageID uint) (*ReceiverStatistics, error) {
	stats := &ReceiverStatistics{
		MessageID: messageID,
	}

	// 总接收者数
	database.DB.Model(&model.UserMessage{}).
		Where("message_id = ?", messageID).
		Count(&stats.TotalCount)

	// 已读数
	database.DB.Model(&model.UserMessage{}).
		Where("message_id = ? AND is_read = 1", messageID).
		Count(&stats.ReadCount)

	// 未读数
	stats.UnreadCount = stats.TotalCount - stats.ReadCount

	// 计算阅读率
	if stats.TotalCount > 0 {
		stats.ReadRate = float64(stats.ReadCount) / float64(stats.TotalCount) * 100
	}

	return stats, nil
}

// BatchSendToUsers 批量发送给指定用户
func (s *MessageAdminService) BatchSendToUsers(message *model.Message, userIDs []uint) error {
	return s.CreateMessage(message, userIDs)
}

// SendToAllUsers 发送给所有用户
func (s *MessageAdminService) SendToAllUsers(message *model.Message) error {
	// 查询所有用户ID
	var userIDs []uint
	if err := database.DB.Model(&model.User{}).Pluck("id", &userIDs).Error; err != nil {
		return fmt.Errorf("查询用户列表失败: %v", err)
	}

	if len(userIDs) == 0 {
		return errors.New("没有找到用户")
	}

	return s.CreateMessage(message, userIDs)
}

// SendByUserLevel 按用户等级发送
func (s *MessageAdminService) SendByUserLevel(message *model.Message, levelID uint) error {
	// 查询指定等级的用户
	var userIDs []uint
	if err := database.DB.Model(&model.User{}).
		Where("level_id = ?", levelID).
		Pluck("id", &userIDs).Error; err != nil {
		return fmt.Errorf("查询用户列表失败: %v", err)
	}

	if len(userIDs) == 0 {
		return errors.New("该等级下没有找到用户")
	}

	return s.CreateMessage(message, userIDs)
}

// ===== 消息模板管理 =====

// GetTemplateList 获取消息模板列表
func (s *MessageAdminService) GetTemplateList() ([]model.MessageTemplate, error) {
	var templates []model.MessageTemplate
	err := database.DB.Order("create_time DESC").Find(&templates).Error
	return templates, err
}

// CreateTemplate 创建消息模板
func (s *MessageAdminService) CreateTemplate(template *model.MessageTemplate) error {
	template.CreateTime = time.Now().Unix()
	template.UpdateTime = time.Now().Unix()
	return database.DB.Create(template).Error
}

// UpdateTemplate 更新消息模板
func (s *MessageAdminService) UpdateTemplate(template *model.MessageTemplate) error {
	template.UpdateTime = time.Now().Unix()
	return database.DB.Save(template).Error
}

// DeleteTemplate 删除消息模板
func (s *MessageAdminService) DeleteTemplate(templateID uint) error {
	return database.DB.Delete(&model.MessageTemplate{}, templateID).Error
}

