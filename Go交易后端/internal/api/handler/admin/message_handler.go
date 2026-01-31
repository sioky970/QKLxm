package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// ============ 站内信管理 ============

// GetMessageList 获取消息列表
func GetMessageList(c *gin.Context) {
	var req admin.MessageListQueryRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Message.GetMessageList(req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetMessageDetail 获取消息详情
func GetMessageDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	message, err := adminService.Message.GetMessageDetail(uint(id))
	if err != nil {
		response.Error(c, "消息不存在")
		return
	}

	response.Success(c, "获取成功", message)
}

// CreateMessage 创建消息
func CreateMessage(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Type        int8   `json:"type"`
		Priority    int8   `json:"priority"`
		ReceiverIDs []uint `json:"receiver_ids" binding:"required"`
		Attachment  string `json:"attachment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取管理员ID（从JWT中）
	adminID, _ := c.Get("admin_id")

	message := &model.Message{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		SenderID:   adminID.(uint),
		SenderType: 1, // 管理员
		Attachment: req.Attachment,
	}

	if err := adminService.Message.CreateMessage(message, req.ReceiverIDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "发送成功")
}

// UpdateMessage 更新消息
func UpdateMessage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Type       int8   `json:"type"`
		Priority   int8   `json:"priority"`
		Attachment string `json:"attachment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	message := &model.Message{
		ID:         uint(id),
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		Attachment: req.Attachment,
	}

	if err := adminService.Message.UpdateMessage(message); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteMessage 删除消息
func DeleteMessage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := adminService.Message.DeleteMessage(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// WithdrawMessage 撤回消息
func WithdrawMessage(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Message.WithdrawMessage(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "撤回成功")
}

// BatchSendMessage 批量发送消息
func BatchSendMessage(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Type        int8   `json:"type"`
		Priority    int8   `json:"priority"`
		ReceiverIDs []uint `json:"receiver_ids" binding:"required"`
		Attachment  string `json:"attachment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	adminID, _ := c.Get("admin_id")

	message := &model.Message{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		SenderID:   adminID.(uint),
		SenderType: 1,
		Attachment: req.Attachment,
	}

	if err := adminService.Message.BatchSendToUsers(message, req.ReceiverIDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "批量发送成功")
}

// SendToAllUsers 发送给所有用户
func SendToAllUsers(c *gin.Context) {
	var req struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Type       int8   `json:"type"`
		Priority   int8   `json:"priority"`
		Attachment string `json:"attachment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	adminID, _ := c.Get("admin_id")

	message := &model.Message{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		SenderID:   adminID.(uint),
		SenderType: 1,
		Attachment: req.Attachment,
	}

	if err := adminService.Message.SendToAllUsers(message); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "群发成功")
}

// SendByUserLevel 按用户等级发送
func SendByUserLevel(c *gin.Context) {
	var req struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Type       int8   `json:"type"`
		Priority   int8   `json:"priority"`
		LevelID    uint   `json:"level_id" binding:"required"`
		Attachment string `json:"attachment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	adminID, _ := c.Get("admin_id")

	message := &model.Message{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
		SenderID:   adminID.(uint),
		SenderType: 1,
		Attachment: req.Attachment,
	}

	if err := adminService.Message.SendByUserLevel(message, req.LevelID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "发送成功")
}

// GetReceiverList 获取消息接收者列表
func GetReceiverList(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	receivers, total, err := adminService.Message.GetReceiverList(uint(id), page, pageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, receivers, total, page, pageSize)
}

// GetMessageStatistics 获取消息统计
func GetMessageStatistics(c *gin.Context) {
	stats, err := adminService.Message.GetMessageStatistics()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", stats)
}

// GetMessageReceiverStatistics 获取单条消息的接收统计
func GetMessageReceiverStatistics(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	stats, err := adminService.Message.GetMessageReceiverStatistics(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", stats)
}

// ============ 消息模板管理 ============

// GetTemplateList 获取模板列表
func GetTemplateList(c *gin.Context) {
	templates, err := adminService.Message.GetTemplateList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", templates)
}

// CreateTemplate 创建模板
func CreateTemplate(c *gin.Context) {
	var template model.MessageTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if template.Code == "" || template.Title == "" || template.Content == "" {
		response.BadRequest(c, "模板代码、标题和内容不能为空")
		return
	}

	if err := adminService.Message.CreateTemplate(&template); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功")
}

// UpdateTemplate 更新模板
func UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var template model.MessageTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	template.ID = uint(id)

	if err := adminService.Message.UpdateTemplate(&template); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteTemplate 删除模板
func DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := adminService.Message.DeleteTemplate(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}
