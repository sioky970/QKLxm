package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/response"
)

// ============ 杠杆交易管理 ============

// LeverOrderListItem 杠杆订单列表项
type LeverOrderListItem struct {
	model.LeverTransaction
	AccountNumber string  `json:"account_number"`
	CurrencyName  string  `json:"currency_name"`
	LegalName     string  `json:"legal_name"`
	RiskRate      float64 `json:"risk_rate"` // 风险率
}

// GetLeverOrderList 获取杠杆订单列表
func GetLeverOrderList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		CurrencyID    uint   `json:"currency_id"`
		LegalID       uint   `json:"legal_id"`
		Status        *int   `json:"status"` // 0-持仓中, 1-已平仓
		Type          string `json:"type"`   // buy/sell
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

	list, total, err := GetLeverOrderListAdvanced(req.AccountNumber, req.CurrencyID, req.LegalID, req.Status, req.Type, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetLeverOrderListAdvanced 获取杠杆订单列表（带筛选）
func GetLeverOrderListAdvanced(accountNumber string, currencyID, legalID uint, status *int, orderType string, page, pageSize int) (list []LeverOrderListItem, total int64, err error) {
	db := database.DB.Table("lever_transaction AS lt").
		Select("lt.*, u.account_number, c1.name AS currency_name, c2.name AS legal_name").
		Joins("LEFT JOIN users u ON lt.user_id = u.id").
		Joins("LEFT JOIN currency c1 ON lt.currency = c1.id").
		Joins("LEFT JOIN currency c2 ON lt.legal = c2.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID > 0 {
		db = db.Where("lt.currency = ?", currencyID)
	}
	if legalID > 0 {
		db = db.Where("lt.legal = ?", legalID)
	}
	if status != nil {
		db = db.Where("lt.status = ?", *status)
	}
	if orderType != "" {
		if orderType == "buy" {
			db = db.Where("lt.type = 1")
		} else if orderType == "sell" {
			db = db.Where("lt.type = 2")
		}
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("lt.id DESC").Find(&list).Error
	return
}

// GetLeverOrderDetail 获取杠杆订单详情
func GetLeverOrderDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order model.LeverTransaction
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

// GetLeverMultipleListFull 获取杠杆倍数配置列表（统一配置）
func GetLeverMultipleListFull(c *gin.Context) {
	var list []model.LeverMultiple
	// 只查询currency_id为null的全局配置
	if err := database.DB.Where("currency_id IS NULL").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// CreateLeverMultiple 创建杠杆倍数配置（全局统一）
func CreateLeverMultiple(c *gin.Context) {
	var req struct {
		Type  string `json:"type" binding:"required"` // multiple-倍数, hand-手数
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 将type转换为int：multiple=1, hand=2
	var typeInt int
	if req.Type == "multiple" {
		typeInt = 1
	} else if req.Type == "hand" {
		typeInt = 2
	}

	// 全局配置，currency_id为nil
	item := model.LeverMultiple{
		Type:       typeInt,
		Value:      req.Value,
		CurrencyID: nil, // 全局配置
	}

	if err := database.DB.Create(&item).Error; err != nil {
		response.Error(c, "创建失败")
		return
	}

	response.Success(c, "创建成功", item)
}

// DeleteLeverMultiple 删除杠杆倍数配置
func DeleteLeverMultiple(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := database.DB.Delete(&model.LeverMultiple{}, id).Error; err != nil {
		response.Error(c, "删除失败")
		return
	}

	response.Success(c, "删除成功")
}

// CloseLeverOrder 管理员强制平仓
func CloseLeverOrder(c *gin.Context) {
	var req struct {
		ID         uint    `json:"id" binding:"required"`
		ClosePrice float64 `json:"close_price"` // 平仓价格，为0则使用当前价格
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var order model.LeverTransaction
		if err := tx.Where("id = ? AND status = 0", req.ID).First(&order).Error; err != nil {
			return err
		}

		closePrice := req.ClosePrice
		if closePrice <= 0 {
			closePrice = order.UpdatePrice // 使用UpdatePrice代替CurrentPrice
		}

		// 计算盈亏
		var profits float64
		if order.Type == 1 { // 做多
			profits = (closePrice - order.Price) * order.Number * float64(order.Multiple)
		} else { // 做空
			profits = (order.Price - closePrice) * order.Number * float64(order.Multiple)
		}

		// 更新订单状态
		now := time.Now().Unix()
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":        1,
			"current_price": closePrice,
			"fact_profits":  profits,
			"complete_time": now,
			"update_time":   now,
		}).Error; err != nil {
			return err
		}

		// 退还保证金 + 盈亏到用户账户
		returnAmount := order.CautionMoney + profits
		if returnAmount > 0 {
			if err := tx.Model(&model.UsersWallet{}).
				Where("user_id = ? AND currency = ?", order.UserID, order.LegalID).
				Update("lever_balance", gorm.Expr("lever_balance + ?", returnAmount)).Error; err != nil {
				return err
			}
		}

		// 记录日志
		accountLog := model.AccountLog{
			UserID:      order.UserID,
			Value:       profits,
			CurrencyID:  order.LegalID,
			Type:        50, // 杠杆平仓
			Info:        "管理员强制平仓",
			CreatedTime: time.Now().Unix(),
		}
		return tx.Create(&accountLog).Error
	}); err != nil {
		response.Error(c, "平仓失败: "+err.Error())
		return
	}

	response.Success(c, "平仓成功")
}
