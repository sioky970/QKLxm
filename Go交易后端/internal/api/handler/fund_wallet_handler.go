package handler

import (
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// FundWalletHandler 资金钱包处理器
type FundWalletHandler struct{}

var fundWalletHandler = &FundWalletHandler{}

// GetFundWalletHandler 获取资金钱包处理器实例
func GetFundWalletHandler() *FundWalletHandler {
	return fundWalletHandler
}

// GetFundWalletBalance 获取资金钱包余额
// @Summary 获取资金钱包余额
// @Description 获取用户资金钱包的余额信息
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=service.FundWalletBalanceResponse} "获取成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/balance [get]
func (h *FundWalletHandler) GetFundWalletBalance(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	uid := userID.(uint)
	currencyID := c.Query("currency_id")
	cID := uint64(1)
	if currencyID != "" {
		tmp, err := strconv.ParseUint(currencyID, 10, 64)
		if err == nil {
			cID = tmp
		}
	}

	wallet, err := service.GetFundWalletService().GetOrCreateFundWallet(uint64(uid), cID)
	if err != nil {
		response.Error(c, "获取资金钱包失败: "+err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"wallet_id":        wallet.ID,
		"user_id":          wallet.UserID,
		"currency_id":      wallet.CurrencyID,
		"currency_name":    wallet.CurrencyName,
		"available_balance": wallet.AvailableBalance.String(),
		"locked_balance":   wallet.LockedBalance.String(),
		"total_balance":    wallet.GetTotalBalance().String(),
		"total_deposit":    wallet.TotalDeposit.String(),
		"total_withdraw":   wallet.TotalWithdraw.String(),
	})
}

// GetAllFundWallets 获取用户所有资金钱包
// @Summary 获取所有资金钱包
// @Description 获取用户所有币种的资金钱包余额
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/list [get]
func (h *FundWalletHandler) GetAllFundWallets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	uid := userID.(uint)
	wallets, err := service.GetFundWalletService().GetFundWalletByUserID(uint64(uid))
	if err != nil {
		response.Error(c, "获取资金钱包列表失败: "+err.Error())
		return
	}

	// 构建响应
	walletsData := make([]gin.H, 0, len(wallets))
	for _, w := range wallets {
		walletsData = append(walletsData, gin.H{
			"wallet_id":        w.ID,
			"currency_id":      w.CurrencyID,
			"currency_name":    w.CurrencyName,
			"available_balance": w.AvailableBalance.String(),
			"locked_balance":   w.LockedBalance.String(),
			"total_balance":    w.GetTotalBalance().String(),
		})
	}

	response.Success(c, "获取成功", walletsData)
}

// FundWalletTransferRequest 资金钱包划转请求
type FundWalletTransferRequest struct {
	FromWallet string `json:"from_wallet" binding:"required"` // 转出钱包: fund, spot, contract, delivery
	ToWallet   string `json:"to_wallet" binding:"required"`   // 转入钱包: fund, spot, contract, delivery
	Amount     string `json:"amount" binding:"required"`       // 划转金额
	CurrencyID uint   `json:"currency_id"`                     // 币种ID，默认1(USDT)
}

// TransferToFund 转入资金钱包
// @Summary 转入资金钱包
// @Description 将现货/合约/交割账户的资金转入资金钱包
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Param request body FundWalletTransferRequest true "划转请求"
// @Success 200 {object} response.Response "划转成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/transfer-in [post]
func (h *FundWalletHandler) TransferToFund(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req FundWalletTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 解析金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "金额必须为有效数字且大于0")
		return
	}

	uid := uint64(userID.(uint))
	clientIP := c.ClientIP()

	// 确定转入方式
	var result *service.TransferResult
	switch req.FromWallet {
	case "spot":
		result = service.GetFundWalletService().DepositToFundWallet(uid, amount, clientIP)
	case "contract":
		result = service.GetFundWalletService().Transfer(uid, model.WalletTypeContract, model.WalletTypeFund, amount, clientIP)
	case "delivery":
		result = service.GetFundWalletService().Transfer(uid, model.WalletTypeDelivery, model.WalletTypeFund, amount, clientIP)
	default:
		response.BadRequest(c, "无效的转出钱包类型")
		return
	}

	if !result.Success {
		response.Error(c, result.ErrorMsg)
		return
	}

	response.Success(c, "划转成功", gin.H{
		"transfer_no":   result.TransferNo,
		"from_wallet":   result.FromWallet,
		"to_wallet":     result.ToWallet,
		"amount":        result.Amount,
		"to_balance":    result.ToBalance,
	})
}

// TransferFromFund 转出资金钱包
// @Summary 转出资金钱包
// @Description 将资金钱包的资金转到现货/合约/交割账户
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Param request body FundWalletTransferRequest true "划转请求"
// @Success 200 {object} response.Response "划转成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/transfer-out [post]
func (h *FundWalletHandler) TransferFromFund(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req FundWalletTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 解析金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "金额必须为有效数字且大于0")
		return
	}

	uid := uint64(userID.(uint))
	clientIP := c.ClientIP()

	// 确定转出方式
	var result *service.TransferResult
	switch req.ToWallet {
	case "spot":
		result = service.GetFundWalletService().WithdrawFromFundWallet(uid, amount, clientIP)
	case "contract":
		result = service.GetFundWalletService().Transfer(uid, model.WalletTypeFund, model.WalletTypeContract, amount, clientIP)
	case "delivery":
		result = service.GetFundWalletService().Transfer(uid, model.WalletTypeFund, model.WalletTypeDelivery, amount, clientIP)
	default:
		response.BadRequest(c, "无效的转入钱包类型")
		return
	}

	if !result.Success {
		response.Error(c, result.ErrorMsg)
		return
	}

	response.Success(c, "划转成功", gin.H{
		"transfer_no":   result.TransferNo,
		"from_wallet":   result.FromWallet,
		"to_wallet":     result.ToWallet,
		"amount":        result.Amount,
		"from_balance":  result.FromBalance,
	})
}

// GetFundWalletOverview 获取资金钱包总览
// @Summary 获取资金钱包总览
// @Description 获取资金钱包的总资产、累计充值、24小时转入转出等统计信息
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/overview [get]
func (h *FundWalletHandler) GetFundWalletOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	uid := userID.(uint)
	overview, err := service.GetFundWalletService().GetOverview(uint64(uid))
	if err != nil {
		response.Error(c, "获取资金钱包总览失败: "+err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"total_balance":    overview.TotalBalance.String(),
		"total_deposit":   overview.TotalDeposit.String(),
		"total_withdraw":  overview.TotalWithdraw.String(),
		"transfer_in_24h": overview.TransferIn24h.String(),
		"transfer_out_24h": overview.TransferOut24h.String(),
	})
}

// GetFundWalletTransfers 获取划转记录
// @Summary 获取划转记录
// @Description 获取资金钱包的划转记录列表
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param transfer_type query int false "划转类型: 1=充值到资金钱包,2=从资金钱包提走,3=转入现货,4=从现货转入,5=转入合约,6=从合约转入,7=转入交割,8=从交割转入"
// @Success 200 {object} response.Response "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/transfers [get]
func (h *FundWalletHandler) GetFundWalletTransfers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	transferType, _ := strconv.Atoi(c.Query("transfer_type"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	uid := userID.(uint)
	var tType int8
	if transferType > 0 {
		tType = int8(transferType)
	}

	transfers, total, err := service.GetFundWalletService().GetTransferList(uint64(uid), page, pageSize, tType)
	if err != nil {
		response.Error(c, "获取划转记录失败: "+err.Error())
		return
	}

	// 构建响应
	transferList := make([]gin.H, 0, len(transfers))
	for _, t := range transfers {
		transferList = append(transferList, gin.H{
			"transfer_no":     t.TransferNo,
			"transfer_type":   t.TransferType,
			"from_wallet":     t.FromWalletType,
			"to_wallet":       t.ToWalletType,
			"amount":          t.Amount.String(),
			"fee":             t.Fee.String(),
			"balance_after":    t.BalanceAfter.String(),
			"status":          t.Status,
			"created_time":    t.CreatedTime,
			"completed_time":  t.CompletedTime,
		})
	}

	response.Success(c, "获取成功", gin.H{
		"list":  transferList,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetFundWalletTransferDetail 获取划转记录详情
// @Summary 获取划转记录详情
// @Description 根据划转单号获取单条划转记录的详细信息
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Param transfer_no path string true "划转单号"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/transfers/{transfer_no} [get]
func (h *FundWalletHandler) GetFundWalletTransferDetail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	transferNo := c.Param("transfer_no")
	if transferNo == "" {
		response.BadRequest(c, "划转单号不能为空")
		return
	}

	transfer, err := service.GetFundWalletService().GetTransferDetail(transferNo)
	if err != nil {
		response.Error(c, "获取划转记录失败: "+err.Error())
		return
	}

	// 验证是否属于当前用户
	if transfer.UserID != uint64(userID.(uint)) {
		response.Error(c, "无权查看此划转记录")
		return
	}

	response.Success(c, "获取成功", gin.H{
		"transfer_no":     transfer.TransferNo,
		"transfer_type":   transfer.TransferType,
		"from_wallet":     transfer.FromWalletType,
		"to_wallet":       transfer.ToWalletType,
		"currency_id":     transfer.CurrencyID,
		"currency_name":   transfer.CurrencyName,
		"amount":          transfer.Amount.String(),
		"fee":             transfer.Fee.String(),
		"balance_before":  transfer.BalanceBefore.String(),
		"balance_after":   transfer.BalanceAfter.String(),
		"status":          transfer.Status,
		"remark":          transfer.Remark,
		"client_ip":       transfer.ClientIP,
		"created_time":    transfer.CreatedTime,
		"completed_time":  transfer.CompletedTime,
	})
}

// GetAllWalletBalances 获取所有钱包余额
// @Summary 获取所有钱包余额
// @Description 获取用户所有类型钱包的余额总览
// @Tags 资金钱包
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/fund/all-balances [get]
func (h *FundWalletHandler) GetAllWalletBalances(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	uid := uint64(userID.(uint))

	// 获取各钱包余额
	spotAssets, _ := service.GetUserAssetsService().GetUserAssetsByType(uint(uid), model.WalletTypeSpot)
	contractAssets, _ := service.GetUserAssetsService().GetUserAssetsByType(uint(uid), model.WalletTypeContract)
	deliveryAssets, _ := service.GetUserAssetsService().GetUserAssetsByType(uint(uid), model.WalletTypeDelivery)
	fundWallet, _ := service.GetFundWalletService().GetOrCreateFundWallet(uid, 1)

	response.Success(c, "获取成功", gin.H{
		"spot": gin.H{
			"balance":      spotAssets.UsdtBalance,
			"locked":       spotAssets.UsdtLocked,
			"wallet_type":  "spot",
		},
		"contract": gin.H{
			"balance":      contractAssets.UsdtBalance,
			"locked":       contractAssets.UsdtLocked,
			"wallet_type":  "contract",
		},
		"delivery": gin.H{
			"balance":      deliveryAssets.UsdtBalance,
			"locked":       deliveryAssets.UsdtLocked,
			"wallet_type":  "delivery",
		},
		"fund": gin.H{
			"available_balance": fundWallet.AvailableBalance.String(),
			"locked_balance":    fundWallet.LockedBalance.String(),
			"total_balance":     fundWallet.GetTotalBalance().String(),
			"wallet_type":       "fund",
		},
		"total_all": spotAssets.UsdtBalance + contractAssets.UsdtBalance + deliveryAssets.UsdtBalance + fundWallet.GetTotalBalance().InexactFloat64(),
	})
}
