package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"
)

// ============ 法币交易管理 ============

// LegalOrderListItem 法币订单列表项
type LegalOrderListItem struct {
	model.LegalDeal
	BuyerAccount  string `json:"buyer_account"`
	SellerAccount string `json:"seller_account"`
	CurrencyName  string `json:"currency_name"`
	Type          string `json:"type"`
	CurrencyID    uint   `json:"currency_id"`
}

// GetLegalOrderList 获取法币订单列表
func GetLegalOrderList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		SellerNumber  string `json:"seller_number"`
		Type          string `json:"type"`   // buy/sell
		Status        *int   `json:"status"` // 0-待付款, 1-已付款, 2-已完成, 3-已取消
		CurrencyID    uint   `json:"currency_id"`
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

	list, total, err := GetLegalOrderListAdvanced(req.AccountNumber, req.SellerNumber, req.Type, req.Status, req.CurrencyID, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetLegalOrderListAdvanced 获取法币订单列表（带筛选）
func GetLegalOrderListAdvanced(accountNumber, sellerNumber, orderType string, status *int, currencyID uint, page, pageSize int) (list []LegalOrderListItem, total int64, err error) {
	db := database.DB.Table("legal_deal AS ld").
		Select("ld.*, u1.account_number AS buyer_account, u2.account_number AS seller_account, c.name AS currency_name, lds.type AS type, lds.currency_id AS currency_id").
		Joins("LEFT JOIN users u1 ON ld.user_id = u1.id").
		Joins("LEFT JOIN users u2 ON ld.seller_id = u2.id").
		Joins("LEFT JOIN legal_deal_send lds ON ld.legal_deal_send_id = lds.id").
		Joins("LEFT JOIN currency c ON lds.currency_id = c.id")

	if accountNumber != "" {
		db = db.Where("u1.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if sellerNumber != "" {
		db = db.Where("u2.account_number LIKE ?", "%"+sellerNumber+"%")
	}
	if orderType != "" {
		db = db.Where("lds.type = ?", orderType)
	}
	if status != nil {
		db = db.Where("ld.status = ?", *status)
	}
	if currencyID > 0 {
		db = db.Where("lds.currency_id = ?", currencyID)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("ld.id DESC").Find(&list).Error
	return
}

// GetLegalStatistics 法币统计
func GetLegalStatistics(c *gin.Context) {
	var result struct {
		TodayBuyUsdt     float64 `json:"today_buy_usdt"`
		TodaySellUsdt    float64 `json:"today_sell_usdt"`
		TotalBuyUsdt     float64 `json:"total_buy_usdt"`
		TotalSellUsdt    float64 `json:"total_sell_usdt"`
		LockBalance      float64 `json:"lock_balance"`
		AvailableBalance float64 `json:"available_balance"`
	}

	todayStart := time.Now().Truncate(24 * time.Hour).Unix()

	// 今日买入
	database.DB.Table("legal_deal ld").
		Joins("JOIN legal_deal_send lds ON ld.legal_deal_send_id = lds.id").
		Where("ld.status = 2 AND lds.type = 'buy' AND ld.create_time >= ?", todayStart).
		Select("COALESCE(SUM(ld.number), 0)").
		Scan(&result.TodayBuyUsdt)

	// 今日卖出
	database.DB.Table("legal_deal ld").
		Joins("JOIN legal_deal_send lds ON ld.legal_deal_send_id = lds.id").
		Where("ld.status = 2 AND lds.type = 'sell' AND ld.create_time >= ?", todayStart).
		Select("COALESCE(SUM(ld.number), 0)").
		Scan(&result.TodaySellUsdt)

	// 总买入
	database.DB.Table("legal_deal ld").
		Joins("JOIN legal_deal_send lds ON ld.legal_deal_send_id = lds.id").
		Where("ld.status = 2 AND lds.type = 'buy'").
		Select("COALESCE(SUM(ld.number), 0)").
		Scan(&result.TotalBuyUsdt)

	// 总卖出
	database.DB.Table("legal_deal ld").
		Joins("JOIN legal_deal_send lds ON ld.legal_deal_send_id = lds.id").
		Where("ld.status = 2 AND lds.type = 'sell'").
		Select("COALESCE(SUM(ld.number), 0)").
		Scan(&result.TotalSellUsdt)

	response.Success(c, "获取成功", result)
}

// CancelLegalOrder 管理员取消法币订单
func CancelLegalOrder(c *gin.Context) {
	var req struct {
		DealID uint `json:"deal_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var deal model.LegalDeal
		if err := tx.Where("id = ?", req.DealID).First(&deal).Error; err != nil {
			return err
		}

		if deal.IsSure == 2 || deal.IsSure == 3 {
			return gorm.ErrInvalidData
		}

		// 更新订单状态为取消
		if err := tx.Model(&deal).Updates(map[string]interface{}{
			"is_sure":     3, // 3表示取消
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
			return err
		}

		// 如果已付款，需要退还买家资金
		if deal.IsSure == 1 {
			// 退还逻辑根据实际业务处理
		}

		// 解冻卖家资产
		var send model.LegalDealSend
		if err := tx.Where("id = ?", deal.LegalDealSendID).First(&send).Error; err == nil {
			// 增加剩余数量
			tx.Model(&send).Update("surplus_num", gorm.Expr("surplus_num + ?", deal.Number))
		}

		return nil
	}); err != nil {
		response.Error(c, "取消失败")
		return
	}

	response.Success(c, "取消成功")
}

// ConfirmLegalPay 管理员确认付款
func ConfirmLegalPay(c *gin.Context) {
	var req struct {
		DealID uint `json:"deal_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Model(&model.LegalDeal{}).Where("id = ? AND status = 0", req.DealID).Updates(map[string]interface{}{
		"is_pay":      1,
		"status":      1,
		"update_time": time.Now().Unix(),
	}).Error; err != nil {
		response.Error(c, "确认失败")
		return
	}

	response.Success(c, "确认成功")
}

// ConfirmLegalReceive 管理员确认收款（完成交易）
func ConfirmLegalReceive(c *gin.Context) {
	var req struct {
		DealID uint `json:"deal_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var deal model.LegalDeal
		if err := tx.Where("id = ? AND status = 1", req.DealID).First(&deal).Error; err != nil {
			return err
		}

		// 更新订单状态
		if err := tx.Model(&deal).Updates(map[string]interface{}{
			"is_sure":     1,
			"status":      2,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
			return err
		}

		// 获取发布信息和币种名称
		var send model.LegalDealSend
		if err := tx.Where("id = ?", deal.LegalDealSendID).First(&send).Error; err != nil {
			return err
		}

		// 查询币种名称
		var currency model.Currency
		if err := tx.Where("id = ?", send.CurrencyID).First(&currency).Error; err != nil {
			return err
		}

		// 转移币到买家账户（使用UserAssets）
		if err := service.GetUserAssetsService().UpdateCurrencyBalance(deal.UserID, currency.Name, deal.Number, false); err != nil {
			return err
		}

		// 记录日志
		accountLog := model.AccountLog{
			UserID:      deal.UserID,
			Value:       deal.Number,
			CurrencyID:  send.CurrencyID,
			Type:        60, // 法币买入
			Info:        "法币交易完成",
			CreatedTime: time.Now().Unix(),
		}
		return tx.Create(&accountLog).Error
	}); err != nil {
		response.Error(c, "确认失败")
		return
	}

	response.Success(c, "交易完成")
}

// GetLegalSendList 获取法币发布列表
func GetLegalSendList(c *gin.Context) {
	var req struct {
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
		Type     string `json:"type"`
		Status   *int   `json:"status"` // is_shelves: 0-下架, 1-上架
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

	db := database.DB.Table("legal_deal_send AS lds").
		Select("lds.*, u.account_number, c.name AS currency_name").
		Joins("LEFT JOIN users u ON lds.seller_id = u.id").
		Joins("LEFT JOIN currency c ON lds.currency_id = c.id")

	if req.Type != "" {
		db = db.Where("lds.type = ?", req.Type)
	}
	if req.Status != nil {
		db = db.Where("lds.is_shelves = ?", *req.Status)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	offset := req.PageSize * (req.Page - 1)
	if err := db.Limit(req.PageSize).Offset(offset).Order("lds.id DESC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// ============ C2C交易管理 ============

// C2COrderListItem C2C订单列表项
type C2COrderListItem struct {
	model.C2cDeal
	BuyerAccount  string `json:"buyer_account"`
	SellerAccount string `json:"seller_account"`
	CurrencyName  string `json:"currency_name"`
	Type          string `json:"type"`
	CurrencyID    uint   `json:"currency_id"`
}

// GetC2COrderList 获取C2C订单列表
func GetC2COrderList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		SellerNumber  string `json:"seller_number"`
		Type          string `json:"type"`
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

	db := database.DB.Table("c2c_deal AS cd").
		Select("cd.*, u1.account_number AS buyer_account, u2.account_number AS seller_account, c.name AS currency_name, cds.type AS type, cds.currency_id AS currency_id").
		Joins("LEFT JOIN users u1 ON cd.user_id = u1.id").
		Joins("LEFT JOIN users u2 ON cd.seller_id = u2.id").
		Joins("LEFT JOIN c2c_deal_send cds ON cd.legal_deal_send_id = cds.id").
		Joins("LEFT JOIN currency c ON cds.currency_id = c.id")

	if req.AccountNumber != "" {
		db = db.Where("u1.account_number LIKE ?", "%"+req.AccountNumber+"%")
	}
	if req.SellerNumber != "" {
		db = db.Where("u2.account_number LIKE ?", "%"+req.SellerNumber+"%")
	}
	if req.Type != "" {
		db = db.Where("cds.type = ?", req.Type)
	}

	var total int64
	db.Count(&total)

	var list []C2COrderListItem
	offset := req.PageSize * (req.Page - 1)
	if err := db.Limit(req.PageSize).Offset(offset).Order("cd.id DESC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetC2CStatistics C2C统计
func GetC2CStatistics(c *gin.Context) {
	var result struct {
		BuyTotal  float64 `json:"buy_total"`
		SellTotal float64 `json:"sell_total"`
	}

	database.DB.Table("c2c_deal cd").
		Joins("JOIN c2c_deal_send cds ON cd.legal_deal_send_id = cds.id").
		Where("cd.is_sure = 1 AND cds.type = 'buy'").
		Select("COALESCE(SUM(cd.number), 0)").
		Scan(&result.BuyTotal)

	database.DB.Table("c2c_deal cd").
		Joins("JOIN c2c_deal_send cds ON cd.legal_deal_send_id = cds.id").
		Where("cd.is_sure = 1 AND cds.type = 'sell'").
		Select("COALESCE(SUM(cd.number), 0)").
		Scan(&result.SellTotal)

	response.Success(c, "获取成功", result)
}

// BackC2CSend 撤回C2C发布
func BackC2CSend(c *gin.Context) {
	var req struct {
		SendID uint `json:"send_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var send model.C2cDealSend
		if err := tx.Where("id = ?", req.SendID).First(&send).Error; err != nil {
			return err
		}

		// 更新发布状态为下架
		if err := tx.Model(&send).Update("is_shelves", 0).Error; err != nil {
			return err
		}

		// 退还剩余数量到用户账户（使用UserAssets）
		if send.SurplusNumber > 0 {
			// 查询币种名称
			var currency model.Currency
			if err := tx.Where("id = ?", send.CurrencyID).First(&currency).Error; err != nil {
				return err
			}
			if err := service.GetUserAssetsService().UpdateCurrencyBalance(send.SellerID, currency.Name, send.SurplusNumber, false); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		response.Error(c, "撤回失败")
		return
	}

	response.Success(c, "撤回成功")
}

// DeleteC2CSend 删除C2C发布
func DeleteC2CSend(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := database.DB.Delete(&model.C2cDealSend{}, id).Error; err != nil {
		response.Error(c, "删除失败")
		return
	}

	response.Success(c, "删除成功")
}

// ============ 商家管理 ============

// SellerListItem 商家列表项
type SellerListItem struct {
	model.Seller
	AccountNumber string `json:"account_number"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	RealName      string `json:"real_name"`
}

// GetSellerListAdvanced 获取商家列表（带筛选）
func GetSellerListAdvanced(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		Status        *int   `json:"status"` // 0-待审核, 1-已通过, 2-已拒绝
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

	db := database.DB.Table("seller AS s").
		Select("s.*, u.account_number, u.phone, u.email, COALESCE(ur.name, '') AS real_name").
		Joins("LEFT JOIN users u ON s.user_id = u.id").
		Joins("LEFT JOIN user_real ur ON s.user_id = ur.user_id")

	if req.AccountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+req.AccountNumber+"%")
	}
	if req.Status != nil {
		db = db.Where("s.status = ?", *req.Status)
	}

	var total int64
	db.Count(&total)

	var list []SellerListItem
	offset := req.PageSize * (req.Page - 1)
	if err := db.Limit(req.PageSize).Offset(offset).Order("s.id DESC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// CreateSeller 添加商家
func CreateSeller(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	seller := model.Seller{
		UserID:     req.UserID,
		Status:     int(req.Status),
		CreateTime: time.Now().Unix(),
	}

	if err := database.DB.Create(&seller).Error; err != nil {
		response.Error(c, "添加失败")
		return
	}

	// 如果直接通过，更新用户的商家标识
	if req.Status == 1 {
		database.DB.Model(&model.User{}).Where("id = ?", req.UserID).Update("is_seller", 1)
	}

	response.Success(c, "添加成功", seller)
}

// ReviewSeller 审核商家
func ReviewSeller(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Status int8   `json:"status" binding:"required"` // 1-通过, 2-拒绝
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var seller model.Seller
		if err := tx.Where("id = ?", req.ID).First(&seller).Error; err != nil {
			return err
		}

		// 更新商家状态
		updates := map[string]interface{}{
			"status":      req.Status,
			"update_time": time.Now().Unix(),
		}
		if req.Reason != "" {
			updates["reason"] = req.Reason
		}
		if err := tx.Model(&seller).Updates(updates).Error; err != nil {
			return err
		}

		// 如果通过，更新用户的商家标识
		if req.Status == 1 {
			return tx.Model(&model.User{}).Where("id = ?", seller.UserID).Update("is_seller", 1).Error
		}

		return nil
	}); err != nil {
		response.Error(c, "审核失败")
		return
	}

	msg := "审核通过"
	if req.Status == 2 {
		msg = "已拒绝"
	}
	response.Success(c, msg)
}

// DeleteSeller 删除商家
func DeleteSeller(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var seller model.Seller
		if err := tx.Where("id = ?", id).First(&seller).Error; err != nil {
			return err
		}

		// 删除商家记录
		if err := tx.Delete(&seller).Error; err != nil {
			return err
		}

		// 取消用户的商家标识
		return tx.Model(&model.User{}).Where("id = ?", seller.UserID).Update("is_seller", 0).Error
	}); err != nil {
		response.Error(c, "删除失败")
		return
	}

	response.Success(c, "删除成功")
}
