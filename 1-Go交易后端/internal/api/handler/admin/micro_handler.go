package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/response"
)

// ============ 秒合约管理 ============

// GetMicroNumberList 获取秒合约金额配置列表
func GetMicroNumberList(c *gin.Context) {
	var list []model.MicroNumber
	if err := database.DB.Order("id ASC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// CreateMicroNumber 创建秒合约金额配置
func CreateMicroNumber(c *gin.Context) {
	var req struct {
		CurrencyID uint    `json:"currency_id" binding:"required"`
		Number     float64 `json:"number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	item := model.MicroNumber{
		CurrencyID: req.CurrencyID,
		Number:     req.Number,
		CreatedAt:  time.Now(),
	}

	if err := database.DB.Create(&item).Error; err != nil {
		response.Error(c, "创建失败")
		return
	}

	response.Success(c, "创建成功", item)
}

// DeleteMicroNumber 删除秒合约金额配置
func DeleteMicroNumber(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := database.DB.Delete(&model.MicroNumber{}, id).Error; err != nil {
		response.Error(c, "删除失败")
		return
	}

	response.Success(c, "删除成功")
}

// GetMicroSecondsList 获取秒合约时间配置列表
func GetMicroSecondsList(c *gin.Context) {
	var list []model.MicroSeconds
	if err := database.DB.Order("seconds ASC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// CreateMicroSeconds 创建秒合约时间配置
func CreateMicroSeconds(c *gin.Context) {
	var req struct {
		Seconds     int     `json:"seconds" binding:"required"`
		ProfitRatio float64 `json:"profit_ratio" binding:"required"`
		Status      int8    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	item := model.MicroSeconds{
		Seconds:     uint(req.Seconds),
		ProfitRatio: req.ProfitRatio,
		Status:      req.Status,
		CreatedAt:   time.Now(),
	}

	if err := database.DB.Create(&item).Error; err != nil {
		response.Error(c, "创建失败")
		return
	}

	response.Success(c, "创建成功", item)
}

// UpdateMicroSeconds 更新秒合约时间配置
func UpdateMicroSeconds(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Seconds     int     `json:"seconds"`
		ProfitRatio float64 `json:"profit_ratio"`
		Status      int8    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Seconds > 0 {
		updates["seconds"] = req.Seconds
	}
	if req.ProfitRatio > 0 {
		updates["profit_ratio"] = req.ProfitRatio
	}
	updates["status"] = req.Status

	if err := database.DB.Model(&model.MicroSeconds{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, "更新失败")
		return
	}

	response.Success(c, "更新成功")
}

// ToggleMicroSecondsStatus 切换秒合约时间配置状态
func ToggleMicroSecondsStatus(c *gin.Context) {
	var req struct {
		ID     uint `json:"id" binding:"required"`
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Model(&model.MicroSeconds{}).Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
		response.Error(c, "更新失败")
		return
	}

	response.Success(c, "更新成功")
}

// DeleteMicroSeconds 删除秒合约时间配置
func DeleteMicroSeconds(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := database.DB.Delete(&model.MicroSeconds{}, id).Error; err != nil {
		response.Error(c, "删除失败")
		return
	}

	response.Success(c, "删除成功")
}

// MicroOrderListItem 秒合约订单列表项
type MicroOrderListItem struct {
	model.MicroOrder
	AccountNumber string `json:"account_number"`
	CurrencyName  string `json:"currency_name"`
	LegalName     string `json:"legal_name"`
}

// GetMicroOrderList 获取秒合约订单列表
func GetMicroOrderList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		CurrencyID    uint   `json:"currency_id"`
		Status        *int   `json:"status"`
		Result        string `json:"result"` // win/lose/draw
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	list, total, err := GetMicroOrderListAdvanced(req.AccountNumber, req.CurrencyID, req.Status, req.Result, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetMicroOrderListAdvanced 获取秒合约订单列表（带筛选）
func GetMicroOrderListAdvanced(accountNumber string, currencyID uint, status *int, result string, page, pageSize int) (list []MicroOrderListItem, total int64, err error) {
	db := database.DB.Table("micro_orders AS mo").
		Select("mo.*, u.account_number, c1.name AS currency_name, 'USDT' AS legal_name").
		Joins("LEFT JOIN users u ON mo.user_id = u.id").
		Joins("LEFT JOIN currency c1 ON mo.currency_id = c1.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID > 0 {
		db = db.Where("mo.currency_id = ?", currencyID)
	}
	if status != nil {
		db = db.Where("mo.status = ?", *status)
	}
	if result != "" {
		switch result {
		case "win":
			db = db.Where("mo.profit_result = 1")
		case "lose":
			db = db.Where("mo.profit_result = 2")
		case "draw":
			db = db.Where("mo.profit_result = 0")
		}
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("mo.id DESC").Find(&list).Error
	return
}

// UpdateMicroOrder 编辑秒合约订单（风控）
func UpdateMicroOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		PreProfitResult *int8    `json:"pre_profit_result"` // 预设结果: 0-不干预, 1-赢, 2-输
		EndPrice        *float64 `json:"end_price"`         // 强制设置结算价格
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.PreProfitResult != nil {
		updates["pre_profit_result"] = *req.PreProfitResult
	}
	if req.EndPrice != nil {
		updates["end_price"] = *req.EndPrice
	}

	if len(updates) == 0 {
		response.BadRequest(c, "没有要更新的内容")
		return
	}

	if err := database.DB.Model(&model.MicroOrder{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, "更新失败")
		return
	}

	response.Success(c, "更新成功")
}

// BatchMicroRisk 批量设置秒合约风控
func BatchMicroRisk(c *gin.Context) {
	var req struct {
		OrderIDs        []uint `json:"order_ids" binding:"required"`
		PreProfitResult int8   `json:"pre_profit_result"` // 0-不干预, 1-赢, 2-输
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if len(req.OrderIDs) == 0 {
		response.BadRequest(c, "请选择订单")
		return
	}

	if err := database.DB.Model(&model.MicroOrder{}).
		Where("id IN ? AND status = 0", req.OrderIDs).
		Update("pre_profit_result", req.PreProfitResult).Error; err != nil {
		response.Error(c, "设置失败")
		return
	}

	response.Success(c, "设置成功")
}

// GetMicroOrderDetail 获取秒合约订单详情
func GetMicroOrderDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order model.MicroOrder
	if err := database.DB.Where("id = ?", id).First(&order).Error; err != nil {
		response.Error(c, "订单不存在")
		return
	}

	// 获取用户信息
	var user model.User
	database.DB.Select("id, account_number, phone, email").Where("id = ?", order.UserID).First(&user)

	result := map[string]interface{}{
		"order": order,
		"user":  user,
	}

	response.Success(c, "获取成功", result)
}
