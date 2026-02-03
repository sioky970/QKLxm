package admin

import (
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// AdminFundWalletHandler 管理端资金钱包处理器
type AdminFundWalletHandler struct{}

var adminFundWalletHandler = &AdminFundWalletHandler{}

// GetAdminFundWalletHandler 获取管理端资金钱包处理器实例
func GetAdminFundWalletHandler() *AdminFundWalletHandler {
	return adminFundWalletHandler
}

// AdminFundWalletListItem 管理端资金钱包列表项
type AdminFundWalletListItem struct {
	ID               uint64 `json:"id"`
	UserID           uint64 `json:"user_id"`
	Username         string `json:"username"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	CurrencyID       uint64 `json:"currency_id"`
	CurrencyName     string `json:"currency_name"`
	AvailableBalance string `json:"available_balance"`
	LockedBalance    string `json:"locked_balance"`
	TotalBalance     string `json:"total_balance"`
	TotalDeposit     string `json:"total_deposit"`
	TotalWithdraw    string `json:"total_withdraw"`
	UpdateTime       int64  `json:"update_time"`
}

// GetFundWalletList 获取资金钱包列表（管理端）
// @Summary 获取资金钱包列表
// @Description 获取所有用户的资金钱包列表
// @Tags 管理端-资金钱包
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param user_id query int false "用户ID筛选"
// @Param currency_id query int false "币种ID筛选"
// @Success 200 {object} response.Response{data=AdminFundWalletListItem} "获取成功"
// @Router /admin/fund-wallets [get]
func (h *AdminFundWalletHandler) GetFundWalletList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userIDStr := c.Query("user_id")
	currencyIDStr := c.Query("currency_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询
	query := `SELECT 
		fw.id, fw.user_id, fw.currency_id, fw.currency_name,
		fw.available_balance, fw.locked_balance,
		fw.total_deposit, fw.total_withdraw, fw.update_time,
		u.phone, u.email
	FROM fund_wallets fw
	LEFT JOIN users u ON fw.user_id = u.id
	WHERE 1=1`

	var queryParams []interface{}

	if userIDStr != "" {
		query += " AND fw.user_id = ?"
		userID, _ := strconv.ParseUint(userIDStr, 10, 64)
		queryParams = append(queryParams, userID)
	}

	if currencyIDStr != "" {
		query += " AND fw.currency_id = ?"
		currencyID, _ := strconv.ParseUint(currencyIDStr, 10, 64)
		queryParams = append(queryParams, currencyID)
	}

	// 统计总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM fund_wallets fw WHERE 1=1"
	if userIDStr != "" {
		countQuery += " AND fw.user_id = ?"
	}
	if currencyIDStr != "" {
		countQuery += " AND fw.currency_id = ?"
	}

	var countParams []interface{}
	if userIDStr != "" {
		userID, _ := strconv.ParseUint(userIDStr, 10, 64)
		countParams = append(countParams, userID)
	}
	if currencyIDStr != "" {
		currencyID, _ := strconv.ParseUint(currencyIDStr, 10, 64)
		countParams = append(countParams, currencyID)
	}

	if err := database.DB.Raw(countQuery, countParams...).Scan(&total).Error; err != nil {
		response.Error(c, "统计总数失败: "+err.Error())
		return
	}

	// 添加排序和分页
	query += " ORDER BY fw.update_time DESC LIMIT ? OFFSET ?"
	queryParams = append(queryParams, pageSize, (page-1)*pageSize)

	// 执行查询
	type ResultRow struct {
		ID               uint64
		UserID           uint64
		CurrencyID       uint64
		CurrencyName     string
		AvailableBalance decimal.Decimal
		LockedBalance    decimal.Decimal
		TotalDeposit     decimal.Decimal
		TotalWithdraw    decimal.Decimal
		UpdateTime       int64
		Phone            string
		Email            string
	}

	var results []ResultRow
	if err := database.DB.Raw(query, queryParams...).Scan(&results).Error; err != nil {
		response.Error(c, "查询失败: "+err.Error())
		return
	}

	// 构建响应
	wallets := make([]AdminFundWalletListItem, 0, len(results))
	for _, r := range results {
		wallets = append(wallets, AdminFundWalletListItem{
			ID:               r.ID,
			UserID:           r.UserID,
			Username:         r.Phone,
			Phone:            r.Phone,
			Email:            r.Email,
			CurrencyID:       r.CurrencyID,
			CurrencyName:     r.CurrencyName,
			AvailableBalance: r.AvailableBalance.String(),
			LockedBalance:    r.LockedBalance.String(),
			TotalBalance:     r.AvailableBalance.Add(r.LockedBalance).String(),
			TotalDeposit:     r.TotalDeposit.String(),
			TotalWithdraw:    r.TotalWithdraw.String(),
			UpdateTime:       r.UpdateTime,
		})
	}

	response.Success(c, "获取成功", gin.H{
		"list":  wallets,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetFundWalletDetail 获取资金钱包详情（管理端）
// @Summary 获取资金钱包详情
// @Description 获取指定用户的资金钱包详细信息
// @Tags 管理端-资金钱包
// @Accept json
// @Produce json
// @Param id path int true "钱包ID"
// @Success 200 {object} response.Response "获取成功"
// @Router /admin/fund-wallets/{id} [get]
func (h *AdminFundWalletHandler) GetFundWalletDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的钱包ID")
		return
	}

	// 查询钱包信息
	var wallet model.FundWallet
	if err := database.DB.Where("id = ?", id).First(&wallet).Error; err != nil {
		response.Error(c, "钱包不存在")
		return
	}

	// 查询用户信息
	var user model.User
	if err := database.DB.Where("id = ?", wallet.UserID).First(&user).Error; err != nil {
		user.Phone = "未知用户"
	}

	response.Success(c, "获取成功", gin.H{
		"wallet": gin.H{
			"id":                wallet.ID,
			"user_id":           wallet.UserID,
			"currency_id":       wallet.CurrencyID,
			"currency_name":     wallet.CurrencyName,
			"available_balance": wallet.AvailableBalance.String(),
			"locked_balance":    wallet.LockedBalance.String(),
			"total_balance":     wallet.GetTotalBalance().String(),
			"total_deposit":    wallet.TotalDeposit.String(),
			"total_withdraw":   wallet.TotalWithdraw.String(),
			"version":          wallet.Version,
			"last_trade_time":  wallet.LastTradeTime,
			"create_time":      wallet.CreateTime,
			"update_time":      wallet.UpdateTime,
		},
		"user": gin.H{
			"id":       user.ID,
			"phone":    user.Phone,
			"email":    user.Email,
			"username": user.Phone,
		},
	})
}

// GetFundWalletTransfers 获取划转记录（管理端）
// @Summary 获取划转记录列表
// @Description 获取资金钱包的划转记录列表
// @Tags 管理端-资金钱包
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param user_id query int false "用户ID筛选"
// @Param transfer_type query int false "划转类型筛选"
// @Param status query int false "状态筛选"
// @Param start_time query int false "开始时间戳"
// @Param end_time query int false "结束时间戳"
// @Success 200 {object} response.Response "获取成功"
// @Router /admin/fund-wallets/transfers [get]
func (h *AdminFundWalletHandler) GetFundWalletTransfers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userIDStr := c.Query("user_id")
	transferTypeStr := c.Query("transfer_type")
	statusStr := c.Query("status")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.FundWalletTransfer{})

	if userIDStr != "" {
		userID, _ := strconv.ParseUint(userIDStr, 10, 64)
		query = query.Where("user_id = ?", userID)
	}

	if transferTypeStr != "" {
		transferType, _ := strconv.Atoi(transferTypeStr)
		query = query.Where("transfer_type = ?", transferType)
	}

	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		query = query.Where("status = ?", status)
	}

	if startTimeStr != "" {
		startTime, _ := strconv.ParseInt(startTimeStr, 10, 64)
		query = query.Where("created_time >= ?", startTime)
	}

	if endTimeStr != "" {
		endTime, _ := strconv.ParseInt(endTimeStr, 10, 64)
		query = query.Where("created_time <= ?", endTime)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, "统计总数失败: "+err.Error())
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var transfers []model.FundWalletTransfer
	if err := query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&transfers).Error; err != nil {
		response.Error(c, "查询失败: "+err.Error())
		return
	}

	// 构建响应
	transferList := make([]gin.H, 0, len(transfers))
	for _, t := range transfers {
		transferList = append(transferList, gin.H{
			"id":              t.ID,
			"transfer_no":     t.TransferNo,
			"user_id":         t.UserID,
			"transfer_type":   t.TransferType,
			"from_wallet":     t.FromWalletType,
			"to_wallet":       t.ToWalletType,
			"currency_id":     t.CurrencyID,
			"currency_name":   t.CurrencyName,
			"amount":         t.Amount.String(),
			"fee":             t.Fee.String(),
			"balance_before":  t.BalanceBefore.String(),
			"balance_after":   t.BalanceAfter.String(),
			"status":         t.Status,
			"remark":         t.Remark,
			"client_ip":      t.ClientIP,
			"created_time":   t.CreatedTime,
			"completed_time": t.CompletedTime,
		})
	}

	response.Success(c, "获取成功", gin.H{
		"list":  transferList,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// FundWalletStatistics 资金钱包统计
type FundWalletStatistics struct {
	TotalWallets      int64   `json:"total_wallets"`
	TotalBalance      string  `json:"total_balance"`
	TotalDeposit      string  `json:"total_deposit"`
	TotalWithdraw     string  `json:"total_withdraw"`
	TodayTransfers    int64   `json:"today_transfers"`
	TodayTransferAmt  string  `json:"today_transfer_amount"`
	PendingTransfers  int64   `json:"pending_transfers"`
}

// GetFundWalletStatistics 获取资金钱包统计（管理端）
// @Summary 获取资金钱包统计
// @Description 获取资金钱包的整体统计数据
// @Tags 管理端-资金钱包
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Router /admin/fund-wallets/statistics [get]
func (h *AdminFundWalletHandler) GetFundWalletStatistics(c *gin.Context) {
	var stats FundWalletStatistics

	// 统计钱包总数
	if err := database.DB.Model(&model.FundWallet{}).Count(&stats.TotalWallets).Error; err != nil {
		response.Error(c, "统计钱包总数失败: "+err.Error())
		return
	}

	// 统计总余额
	var totalBalance, totalDeposit, totalWithdraw decimal.Decimal
	database.DB.Model(&model.FundWallet{}).
		Select("COALESCE(SUM(available_balance + locked_balance), 0)").
		Scan(&totalBalance)

	database.DB.Model(&model.FundWallet{}).
		Select("COALESCE(SUM(total_deposit), 0)").
		Scan(&totalDeposit)

	database.DB.Model(&model.FundWallet{}).
		Select("COALESCE(SUM(total_withdraw), 0)").
		Scan(&totalWithdraw)

	// 今日统计
	todayStart := time.Now().Format("2006-01-02") + " 00:00:00"
	_ = todayStart
	todayTimestamp := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()).Unix()

	var todayAmount decimal.Decimal
	database.DB.Model(&model.FundWalletTransfer{}).
		Where("status = ? AND created_time >= ?", model.FundTransferSuccess, todayTimestamp).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&todayAmount)

	// 待处理划转
	database.DB.Model(&model.FundWalletTransfer{}).
		Where("status = ?", model.FundTransferPending).
		Count(&stats.PendingTransfers)

	stats.TotalBalance = totalBalance.String()
	stats.TotalDeposit = totalDeposit.String()
	stats.TotalWithdraw = totalWithdraw.String()
	stats.TodayTransferAmt = todayAmount.String()

	response.Success(c, "获取成功", stats)
}

// AdminAdjustFundWallet 管理端调整资金钱包余额
// @Summary 调整资金钱包余额
// @Description 管理员手动调整用户的资金钱包余额
// @Tags 管理端-资金钱包
// @Accept json
// @Produce json
// @Param request body object{user_id=number,amount=string,type=string,remark=string} true "调整请求"
// @Success 200 {object} response.Response "调整成功"
// @Router /admin/fund-wallets/adjust [post]
func (h *AdminFundWalletHandler) AdminAdjustFundWallet(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Amount string `json:"amount" binding:"required"`
		Type   string `json:"type" binding:"required"` // "add" 或 "subtract"
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "金额必须为有效数字且大于0")
		return
	}

	if req.Type != "add" && req.Type != "subtract" {
		response.BadRequest(c, "操作类型必须是 add 或 subtract")
		return
	}

	// 获取钱包
	wallet, err := service.GetFundWalletService().GetOrCreateFundWallet(req.UserID, 1)
	if err != nil {
		response.Error(c, "获取钱包失败: "+err.Error())
		return
	}

	now := time.Now().Unix()
	var newBalance decimal.Decimal
	balanceBefore := wallet.AvailableBalance

	if req.Type == "add" {
		newBalance = wallet.AvailableBalance.Add(amount)
		wallet.TotalDeposit = wallet.TotalDeposit.Add(amount)
	} else {
		newBalance = wallet.AvailableBalance.Sub(amount)
		if newBalance.LessThan(decimal.Zero) {
			response.Error(c, "余额不足，无法扣减")
			return
		}
		wallet.TotalWithdraw = wallet.TotalWithdraw.Add(amount)
	}

	wallet.AvailableBalance = newBalance
	wallet.UpdateTime = now

	if err := database.DB.Save(wallet).Error; err != nil {
		response.Error(c, "更新钱包失败: "+err.Error())
		return
	}

	// 记录调整日志
	transferNo := "ADJ" + time.Now().Format("20060102150405")
	record := &model.FundWalletTransfer{
		TransferNo:     transferNo,
		UserID:         req.UserID,
		TransferType:   model.FundTransferDeposit,
		FromWalletType: "admin",
		ToWalletType:   model.WalletTypeFund,
		CurrencyID:     1,
		CurrencyName:   "USDT",
		Amount:         amount,
		Fee:            decimal.Zero,
		BalanceBefore:  balanceBefore,
		BalanceAfter:   newBalance,
		Status:         model.FundTransferSuccess,
		Remark:         req.Remark,
		ClientIP:       c.ClientIP(),
		CreatedTime:    now,
		CompletedTime:  now,
	}
	database.DB.Create(record)

	response.Success(c, "调整成功", gin.H{
		"user_id":       req.UserID,
		"type":          req.Type,
		"amount":        amount.String(),
		"balance_before": balanceBefore.String(),
		"balance_after": newBalance.String(),
	})
}
