package handler

import (
	"exchange-go/internal/api/middleware"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"
	"fmt"

	"github.com/gin-gonic/gin"
)

var userService = service.NewUserService()
var walletService = service.NewWalletService()
var spotTradingService = service.NewSpotTradingService()
var futuresTradingService = service.NewFuturesTradingService()
var microTradingService = service.NewMicroTradingService()
var newsService = service.NewNewsService()
var messageService = service.NewMessageService()
var currencyService = service.NewCurrencyService()

// ============ 用户模块 ============

// UserRegister 用户注册
// @Summary 用户注册
// @Description 新用户注册接口
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册参数"
// @Success 200 {object} response.Response "注册成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /user/register [post]
func UserRegister(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := userService.Register(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "注册成功")
}

// UserLogin 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT Token
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录参数"
// @Success 200 {object} response.Response{data=object} "登录成功，返回token和用户信息"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "账号或密码错误"
// @Router /user/login [post]
func UserLogin(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := userService.Login(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "登录成功", gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

// SendVerifyCode 发送验证码
func SendVerifyCode(c *gin.Context) {
	target := c.PostForm("target")
	codeType := c.PostForm("type") // email 或 mobile

	if target == "" {
		response.BadRequest(c, "请输入手机号或邮箱")
		return
	}
	if codeType == "" {
		codeType = "mobile"
	}

	code, err := service.SendVerifyCode(target, codeType)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	// 开发模式返回验证码
	if code != "" {
		response.Success(c, "验证码已发送", gin.H{"code": code})
	} else {
		response.Success(c, "验证码已发送")
	}
}

// ResetPassword 重置密码
func ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := userService.ResetPassword(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "密码重置成功")
}

// UserInfo 获取用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /user/info [get]
func UserInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	user, hasPayPassword, err := userService.GetUserInfoWithPayPasswordStatus(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"user":             user,
		"has_pay_password": hasPayPassword,
	})
}

// UserUpdate 更新用户信息
func UserUpdate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := userService.UpdateUserInfo(userID, updates); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := userService.ChangePassword(userID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "密码修改成功")
}

// ChangePayPassword 修改支付密码
func ChangePayPassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")

	if newPassword == "" {
		response.BadRequest(c, "请输入新支付密码")
		return
	}

	if err := userService.ChangePayPassword(userID, oldPassword, newPassword); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "支付密码修改成功")
}

// GetUserCashInfo 获取用户提现信息
func GetUserCashInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	cash, err := userService.GetUserCashInfo(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", cash)
}

// SaveUserCashInfo 保存用户提现信息
func SaveUserCashInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.UserCashInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := userService.SaveUserCashInfo(userID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "保存成功")
}

// ============ 行情模块 ============

var marketService = service.NewMarketService()

// NewQuotation 最新行情
func NewQuotation(c *gin.Context) {
	quotations, err := marketService.GetNewQuotation()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", quotations)
}

// MarketData 市场数据
func MarketData(c *gin.Context) {
	var req service.MarketDataRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	data, err := marketService.GetMarketData(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", data)
}

// GetKline K线数据
func GetKline(c *gin.Context) {
	var req service.KlineRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := marketService.GetKline(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	if result.Code == -1 {
		response.Error(c, result.Msg)
		return
	}

	response.Success(c, "获取成功", result.Data)
}

// CurrencyQuotation 币种行情
func CurrencyQuotation(c *gin.Context) {
	currencyID := c.Query("currency_id")
	legalID := c.Query("legal_id")

	var currencyIDUint, legalIDUint uint = 0, 0
	fmt.Sscanf(currencyID, "%d", &currencyIDUint)
	fmt.Sscanf(legalID, "%d", &legalIDUint)

	quotation, err := marketService.GetCurrencyQuotation(currencyIDUint, legalIDUint)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", quotation)
}

// ============ 钱包模块 ============

// WalletRecharge 充值
// @Summary 钱包充值
// @Description 申请充值到指定币种钱包
// @Tags 钱包模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.RechargeRequest true "充值参数"
// @Success 200 {object} response.Response "充值申请已提交"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/recharge [post]
func WalletRecharge(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := walletService.Recharge(userID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "充值申请已提交")
}

// WalletWithdraw 提现
func WalletWithdraw(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := walletService.Withdraw(userID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "提现申请已提交")
}

// WalletLogs 钱包流水
func WalletLogs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	currencyID := c.Query("currency_id")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageInt := 1
	pageSizeInt := 20
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)

	var currencyIDUint uint = 0
	if currencyID != "" {
		fmt.Sscanf(currencyID, "%d", &currencyIDUint)
	}

	logs, total, err := walletService.GetWalletLogs(userID, currencyIDUint, pageInt, pageSizeInt)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"list":  logs,
		"total": total,
		"page":  pageInt,
		"size":  pageSizeInt,
	})
}

// GetWithdrawalList 获取用户提现记录列表
func GetWithdrawalList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageInt := 1
	pageSizeInt := 20
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)

	list, total, err := walletService.GetWithdrawalList(userID, pageInt, pageSizeInt)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"list":  list,
		"total": total,
		"page":  pageInt,
		"size":  pageSizeInt,
	})
}

// GetWithdrawConfig 获取提现配置
func GetWithdrawConfig(c *gin.Context) {
	config, err := walletService.GetWithdrawConfig()
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, "获取成功", config)
}

// WalletInfo 钱包详情
func WalletInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	currencyID := c.Query("currency_id")
	if currencyID == "" {
		response.BadRequest(c, "请选择币种")
		return
	}

	var currencyIDUint uint
	fmt.Sscanf(currencyID, "%d", &currencyIDUint)

	wallet, err := walletService.GetWalletByCurrency(userID, currencyIDUint)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", wallet)
}

// WalletList 钱包列表
// @Summary 获取用户所有钱包
// @Description 获取用户各种币种的钱包列表
// @Tags 钱包模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=array} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/list [get]
func WalletList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	list, err := walletService.GetWalletList(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// GetAssetOverview 获取资产概览
// @Summary 获取用户资产概览
// @Description 获取用户总资产、今日盈亏和各币种资产列表（用于首页和资产页）
// @Tags 钱包模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/asset-overview [get]
func GetAssetOverview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	overview, err := walletService.GetAssetOverview(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", overview)
}

// ============ 币币交易模块 ============

// SpotSubmit 币币交易下单
// @Summary 币币交易下单
// @Description 提交币币交易订单（限价单/市价单）
// @Tags 币币交易
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.SpotSubmitOrderRequest true "下单参数"
// @Success 200 {object} response.Response{data=object} "下单成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /transaction/submit [post]
func SpotSubmit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.SpotSubmitOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := spotTradingService.SubmitOrder(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "下单成功", result)
}

// SpotCancel 取消币币交易订单
func SpotCancel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := spotTradingService.CancelOrder(userID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "撤单成功")
}

// SpotChase 追单：修改订单价格并成交
func SpotChase(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.ChaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := spotTradingService.ChaseOrder(userID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "追单成功")
}

// SpotList 币币交易订单列表
func SpotList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.SpotOrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 添加详细日志：打印接收到的参数
	logger.Infof("[SpotList] 接收参数 - UserID: %d, Status: %v, Side: %v, CurrencyID: %d, Page: %d",
		userID, req.Status, req.Side, req.CurrencyID, req.Page)

	orders, total, err := spotTradingService.GetOrderList(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	logger.Infof("[SpotList] 返回结果 - Total: %d, Count: %d", total, len(orders))
	response.SuccessWithPage(c, orders, total, req.Page, req.PageSize)
}

// SpotHistory 币币交易历史
func SpotHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	currencyIDStr := c.Query("currency_id")
	legalIDStr := c.Query("legal_id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	var currencyID, legalID uint = 0, 0
	fmt.Sscanf(currencyIDStr, "%d", &currencyID)
	fmt.Sscanf(legalIDStr, "%d", &legalID)

	page := 1
	pageSize := 20
	fmt.Sscanf(pageStr, "%d", &page)
	fmt.Sscanf(pageSizeStr, "%d", &pageSize)

	trades, total, err := spotTradingService.GetTradeHistory(userID, currencyID, legalID, page, pageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, trades, total, page, pageSize)
}

// LeverSubmit 合约下单(兼容旧API)
// @Summary 合约交易下单
// @Description 提交合约交易订单，开仓或加仓
// @Tags 合约交易
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.OpenPositionRequest true "下单参数"
// @Success 200 {object} response.Response{data=object} "下单成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /lever/submit [post]
func LeverSubmit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.OpenPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := futuresTradingService.OpenPosition(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "下单成功", result)
}

// LeverClose 合约平仓(兼容旧API)
func LeverClose(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.ClosePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := futuresTradingService.ClosePosition(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "平仓成功", result)
}

// LeverPosition 合约持仓
func LeverPosition(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.PositionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	positions, total, err := futuresTradingService.GetPositionList(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, positions, total, req.Page, req.PageSize)
}

// LeverHistory 合约历史(兼容旧API，复用持仓列表查询已平仓订单)
func LeverHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	currencyIDStr := c.Query("currency_id")
	legalIDStr := c.Query("legal_id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	var currencyID, legalID uint = 0, 0
	fmt.Sscanf(currencyIDStr, "%d", &currencyID)
	fmt.Sscanf(legalIDStr, "%d", &legalID)

	page := 1
	pageSize := 20
	fmt.Sscanf(pageStr, "%d", &page)
	fmt.Sscanf(pageSizeStr, "%d", &pageSize)

	// 使用持仓列表查询已平仓订单
	req := &service.PositionListRequest{
		CurrencyID: currencyID,
		LegalID:    legalID,
		Status:     3, // 已平仓
		Page:       page,
		PageSize:   pageSize,
	}

	orders, total, err := futuresTradingService.GetPositionList(userID, req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, orders, total, page, pageSize)
}

// MicroSubmit 秒合约下单
// @Summary 秒合约交易
// @Description 提交秒合约订单，预测涨跌
// @Tags 秒合约
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.MicroSubmitOrderRequest true "下单参数"
// @Success 200 {object} response.Response{data=object} "下单成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /micro/submit [post]
func MicroSubmit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.MicroSubmitOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := microTradingService.SubmitOrder(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "下单成功", result)
}

// MicroList 秒合约订单列表
// @Summary 获取秒合约订单列表
// @Description 获取用户的秒合约订单列表，可按状态筛选
// @Tags 秒合约
// @Produce json
// @Security BearerAuth
// @Param status query int false "状态: 0=进行中, 1=已结算"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=[]service.MicroOrderDetailResponse}
// @Router /micro/orders [get]
func MicroList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.MicroOrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	orders, total, err := microTradingService.GetOrderList(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, orders, total, req.Page, req.PageSize)
}

// MicroPeriods 获取秒合约周期配置
// @Summary 获取秒合约周期配置
// @Description 获取所有可用的秒合约周期及对应盈利率
// @Tags 秒合约
// @Produce json
// @Success 200 {object} response.Response{data=[]service.MicroPeriodConfig}
// @Router /micro/periods [get]
func MicroPeriods(c *gin.Context) {
	configs, err := microTradingService.GetPeriodConfigs()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", configs)
}

// MicroActiveOrder 获取用户当前进行中的订单
// @Summary 获取当前进行中的秒合约订单
// @Description 获取用户当前唯一进行中的秒合约订单（如有）
// @Tags 秒合约
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.MicroOrderDetailResponse}
// @Router /micro/active [get]
func MicroActiveOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	order, err := microTradingService.GetActiveOrder(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	if order == nil {
		response.Success(c, "暂无进行中的订单", nil)
		return
	}

	response.Success(c, "获取成功", order)
}

// MicroOrderDetail 获取秒合约订单详情
// @Summary 获取秒合约订单详情
// @Description 根据订单ID获取秒合约订单详情
// @Tags 秒合约
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} response.Response{data=service.MicroOrderDetailResponse}
// @Router /micro/order/{id} [get]
func MicroOrderDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	orderIDStr := c.Param("id")
	var orderID uint
	fmt.Sscanf(orderIDStr, "%d", &orderID)
	if orderID == 0 {
		response.BadRequest(c, "订单ID无效")
		return
	}

	order, err := microTradingService.GetOrderDetail(userID, orderID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", order)
}

// ============ 新闻模块（前台） ============

// GetNewsList 获取新闻列表
func GetNewsList(c *gin.Context) {
	categoryID := c.DefaultQuery("category_id", "0")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")

	var categoryIDUint uint = 0
	var pageInt, pageSizeInt int = 1, 10
	fmt.Sscanf(categoryID, "%d", &categoryIDUint)
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)

	if pageInt <= 0 {
		pageInt = 1
	}
	if pageSizeInt <= 0 || pageSizeInt > 100 {
		pageSizeInt = 10
	}

	list, total, err := newsService.GetPublishedNewsList(categoryIDUint, pageInt, pageSizeInt)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, pageInt, pageSizeInt)
}

// GetNewsDetail 获取新闻详情
func GetNewsDetail(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.BadRequest(c, "新闻ID不能为空")
		return
	}

	var newsID uint
	fmt.Sscanf(id, "%d", &newsID)

	news, err := newsService.GetNewsDetail(newsID)
	if err != nil {
		response.Error(c, "新闻不存在或已下架")
		return
	}

	response.Success(c, "获取成功", news)
}

// GetNewsCategoryList 获取新闻分类列表
func GetNewsCategoryList(c *gin.Context) {
	list, err := newsService.GetNewsCategoryList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// GetRecommendedNews 获取推荐新闻
func GetRecommendedNews(c *gin.Context) {
	limit := c.DefaultQuery("limit", "5")
	var limitInt int = 5
	fmt.Sscanf(limit, "%d", &limitInt)

	if limitInt <= 0 || limitInt > 20 {
		limitInt = 5
	}

	list, err := newsService.GetRecommendedNews(limitInt)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// GetHotNews 获取热门新闻
func GetHotNews(c *gin.Context) {
	limit := c.DefaultQuery("limit", "5")
	var limitInt int = 5
	fmt.Sscanf(limit, "%d", &limitInt)

	if limitInt <= 0 || limitInt > 20 {
		limitInt = 5
	}

	list, err := newsService.GetHotNews(limitInt)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// ============ 站内信模块 ============

// GetMessageList 获取用户的站内信列表
func GetMessageList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.MessageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	messages, total, err := messageService.GetUserMessageList(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, messages, total, req.Page, req.PageSize)
}

// GetMessageDetail 获取消息详情
func GetMessageDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	messageID := c.Query("message_id")
	if messageID == "" {
		response.BadRequest(c, "消息ID不能为空")
		return
	}

	var msgID uint
	fmt.Sscanf(messageID, "%d", &msgID)

	message, err := messageService.GetMessageDetail(userID, msgID)
	if err != nil {
		response.Error(c, "消息不存在")
		return
	}

	response.Success(c, "获取成功", message)
}

// MarkMessageAsRead 标记消息为已读
func MarkMessageAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		MessageID uint `json:"message_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := messageService.MarkAsRead(userID, req.MessageID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "标记成功")
}

// MarkMessageAsUnread 标记消息为未读
func MarkMessageAsUnread(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		MessageID uint `json:"message_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := messageService.MarkAsUnread(userID, req.MessageID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "标记成功")
}

// BatchMarkMessagesAsRead 批量标记消息为已读
func BatchMarkMessagesAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		MessageIDs []uint `json:"message_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := messageService.BatchMarkAsRead(userID, req.MessageIDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "批量标记成功")
}

// MarkAllMessagesAsRead 标记所有消息为已读
func MarkAllMessagesAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	if err := messageService.MarkAllAsRead(userID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "全部标记为已读")
}

// DeleteMessage 删除消息
func DeleteMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		MessageID uint `json:"message_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := messageService.DeleteMessage(userID, req.MessageID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// BatchDeleteMessages 批量删除消息
func BatchDeleteMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		MessageIDs []uint `json:"message_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := messageService.BatchDeleteMessages(userID, req.MessageIDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "批量删除成功")
}

// GetUnreadMessageCount 获取未读消息数量
func GetUnreadMessageCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	count, err := messageService.GetUnreadCount(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{"count": count})
}

// GetMessageStatistics 获取消息统计
func GetMessageStatistics(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	stats, err := messageService.GetMessageStatistics(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", stats)
}

// ============ 今日盈亏统计 ============

// GetTodayProfitLoss 获取今日盈亏统计
// @Summary 获取今日盈亏统计
// @Description 获取当前用户今日盈亏详情（北京时间）
// @Tags 钱包模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.TodayProfitLossResponse} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/today-profit-loss [get]
func GetTodayProfitLoss(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	profitLoss, err := walletService.GetTodayProfitLoss(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", profitLoss)
}

// GetSevenDaysProfitLoss 获取7天盈亏折线图数据
// @Summary 获取7天盈亏折线图数据
// @Description 获取最近7天每日盈亏数据，用于绘制折线图（北京时间）
// @Tags 钱包模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.SevenDaysProfitLossResponse} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/seven-days-profit-loss [get]
func GetSevenDaysProfitLoss(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	sevenDaysData, err := walletService.GetSevenDaysProfitLoss(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", sevenDaysData)
}

// GetAllAssetsWithBalance 获取所有启用币种及用户余额
// @Summary 获取所有启用币种及用户余额
// @Description 获取所有启用的币种列表及用户在每个币种的余额（包括余额为0的币种）
// @Tags 钱包模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.AssetOverviewResponse} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Router /wallet/all-assets [get]
func GetAllAssetsWithBalance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	assets, err := walletService.GetAllAssetsWithBalance(userID)
	if err != nil {
		logger.Errorf("[GetAllAssetsWithBalance] 获取资产失败: err=%v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", assets)
}

// ============ 公开配置模块 ============

// GetRegisterConfig 获取注册配置
// @Summary 获取注册配置
// @Description 获取注册页面需要的公开配置项，如邀请码是否必填等
// @Tags 公开配置
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Router /config/register [get]
func GetRegisterConfig(c *gin.Context) {
	config, err := service.GetRegisterConfig()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", config)
}

// ============ 币种模块 ============

// GetCurrencyList 获取启用的币种列表
// @Summary 获取启用的币种列表
// @Description 获取系统中启用显示的所有币种信息
// @Tags 币种模块
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]service.CurrencyListResponse} "获取成功"
// @Router /currency/list [get]
func GetCurrencyList(c *gin.Context) {
	list, err := currencyService.GetCurrencyList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// ============ 统一订单模块 ============

var unifiedOrdersService = service.GetUnifiedOrdersService()

// GetUnifiedOrderList 获取统一订单列表
// @Summary 获取统一订单列表
// @Description 获取用户的所有类型订单（现货、永续合约、秒合约）
// @Tags 订单模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trade_type query string false "交易类型: spot(现货), contract(合约), micro(秒合约)"
// @Param currency_id query int false "币种ID筛选"
// @Param status query int false "状态筛选"
// @Param side query string false "方向筛选: buy/sell/long/short"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=service.UnifiedOrderListResponse} "获取成功"
// @Router /orders/all [get]
func GetUnifiedOrderList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.UnifiedOrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := unifiedOrdersService.GetUnifiedOrderList(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetOrderStatistics 获取订单统计
// @Summary 获取订单统计
// @Description 获取用户各类型订单的统计数据
// @Tags 订单模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.OrderStatistics} "获取成功"
// @Router /orders/statistics [get]
func GetOrderStatistics(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	stats, err := unifiedOrdersService.GetOrderStatistics(userID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", stats)
}

// GetOrderDetail 获取订单详情
// @Summary 获取订单详情
// @Description 根据订单ID和类型获取详情
// @Tags 订单模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id query int true "订单ID"
// @Param trade_type query string true "交易类型: spot/contract/micro"
// @Success 200 {object} response.Response{data=service.UnifiedOrder} "获取成功"
// @Router /orders/detail [get]
func GetOrderDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	orderIDStr := c.Query("order_id")
	tradeType := c.Query("trade_type")

	if orderIDStr == "" || tradeType == "" {
		response.BadRequest(c, "订单ID和交易类型不能为空")
		return
	}

	var orderID uint
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	order, err := unifiedOrdersService.GetOrderByID(userID, orderID, tradeType)
	if err != nil {
		response.Error(c, "订单不存在")
		return
	}

	response.Success(c, "获取成功", order)
}
