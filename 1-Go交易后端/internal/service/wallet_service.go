package service

import (
	"errors"
	"fmt"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/websocket"

	"gorm.io/gorm"
)

// WalletService 钱包服务
type WalletService struct{}

// NewWalletService 创建钱包服务实例
func NewWalletService() *WalletService {
	return &WalletService{}
}

// WalletType 钱包类型
type WalletType string

const (
	WalletTypeLegal  WalletType = "legal"  // 法币账户
	WalletTypeChange WalletType = "change" // 币币账户
	WalletTypeLever  WalletType = "lever"  // 合约账户
	WalletTypeMicro  WalletType = "micro"  // 秒合约账户
)

// WalletBalanceInfo 钱包余额信息
type WalletBalanceInfo struct {
	ID               uint    `json:"id"`
	CurrencyID       uint    `json:"currency_id"`
	CurrencyName     string  `json:"currency_name"`
	CurrencyLogo     string  `json:"currency_logo"`
	Balance          float64 `json:"balance"`
	LockedBalance    float64 `json:"locked_balance"`
	AvailableBalance float64 `json:"available_balance"`
	UsdtValue        float64 `json:"usdt_value"`
}

// WalletListResponse 钱包列表响应
type WalletListResponse struct {
	LegalWallet  *WalletSummary `json:"legal_wallet"`
	ChangeWallet *WalletSummary `json:"change_wallet"`
	LeverWallet  *WalletSummary `json:"lever_wallet"`
	MicroWallet  *WalletSummary `json:"micro_wallet"`
	ExRate       float64        `json:"ex_rate"`
}

// WalletSummary 钱包汇总
type WalletSummary struct {
	Balance   []WalletBalanceInfo `json:"balance"`
	Total     float64             `json:"total"`
	UsdtTotal float64             `json:"usdt_total"`
}

// TransferRequest 划转请求
type TransferRequest struct {
	From       WalletType `json:"from" binding:"required"`
	To         WalletType `json:"to" binding:"required"`
	CurrencyID uint       `json:"currency_id" binding:"required"`
	Amount     float64    `json:"amount" binding:"required,gt=0"`
}

// RechargeRequest 充值请求
type RechargeRequest struct {
	CurrencyID uint    `json:"currency_id" binding:"required"`
	Account    string  `json:"account" binding:"required"`     // 用户账号
	Amount     float64 `json:"amount" binding:"required,gt=0"` // 充值金额
	Hash       string  `json:"hash"`                           // 交易哈希
	Type       string  `json:"type"`                           // 地址类型
	Image      string  `json:"pic"`                            // 图片凭证
}

// WithdrawRequest 提现请求
type WithdrawRequest struct {
	CurrencyID  uint    `json:"currency_id" binding:"required"`
	Address     string  `json:"address" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	PayPassword string  `json:"pay_password" binding:"required"`
}

// GetWalletList 获取用户钱包列表
func (s *WalletService) GetWalletList(userID uint) (*WalletListResponse, error) {
	// 获取用户所有钱包
	var wallets []model.UsersWallet
	result := database.DB.Where("user_id = ?", userID).Find(&wallets)
	if result.Error != nil {
		return nil, result.Error
	}

	// 获取所有币种信息
	var currencies []model.Currency
	database.DB.Where("is_display = 1").Find(&currencies)
	currencyMap := make(map[uint]model.Currency)
	for _, c := range currencies {
		currencyMap[c.ID] = c
	}

	// 初始化返回结构
	resp := &WalletListResponse{
		LegalWallet:  &WalletSummary{Balance: []WalletBalanceInfo{}},
		ChangeWallet: &WalletSummary{Balance: []WalletBalanceInfo{}},
		LeverWallet:  &WalletSummary{Balance: []WalletBalanceInfo{}},
		MicroWallet:  &WalletSummary{Balance: []WalletBalanceInfo{}},
		ExRate:       6.5, // TODO: 从配置读取
	}

	// 遍历钱包，分类汇总
	for _, w := range wallets {
		currency, ok := currencyMap[w.CurrencyID]
		if !ok {
			continue
		}

		// 统一USDT余额（不再区分子账户）
		info := WalletBalanceInfo{
			ID:               w.ID,
			CurrencyID:       w.CurrencyID,
			CurrencyName:     currency.Name,
			CurrencyLogo:     currency.Logo,
			Balance:          w.UsdtBalance,
			LockedBalance:    w.LockUsdtBalance,
			AvailableBalance: w.UsdtBalance,
			UsdtValue:        w.UsdtBalance * currency.Rate,
		}

		// 所有余额都显示在币币账户中（保持API兼容）
		resp.ChangeWallet.Balance = append(resp.ChangeWallet.Balance, info)
		resp.ChangeWallet.UsdtTotal += info.UsdtValue
	}

	return resp, nil
}

// AssetOverviewResponse 资产概览响应
type AssetOverviewResponse struct {
	TotalBalance    float64         `json:"total_balance"`     // 总资产(USDT)
	TotalUsdValue   float64         `json:"total_usd_value"`   // 总资产(USD)
	TodayProfit     float64         `json:"today_profit"`      // 今日盈亏(USDT)
	TodayProfitRate string          `json:"today_profit_rate"` // 今日盈亏率
	Assets          []AssetItemInfo `json:"assets"`            // 各币种资产列表
}

// AssetItemInfo 单个资产项信息
type AssetItemInfo struct {
	CurrencyID   uint    `json:"currency_id"`   // 币种ID
	CurrencyName string  `json:"currency_name"` // 币种名称
	Symbol       string  `json:"symbol"`        // 币种符号(如BTC, ETH)
	Logo         string  `json:"logo"`          // 币种图标
	Balance      float64 `json:"balance"`       // 持有数量
	UsdtValue    float64 `json:"usdt_value"`    // USDT价值
	UsdValue     float64 `json:"usd_value"`     // USD价值
	Price        float64 `json:"price"`         // 当前价格(USDT)
	Sort         int     `json:"sort"`          // 排序值(越大越靠前)
}

// GetAssetOverview 获取资产概览（用于首页和资产页）
// 统一使用 UserAssets 新表，保持数据一致性
func (s *WalletService) GetAssetOverview(userID uint) (*AssetOverviewResponse, error) {
	// 统一使用 UserAssets 表读取，与 GetAllAssetsWithBalance 保持一致
	return GetUserAssetsService().GetAssetOverviewFromUserAssets(userID)
}

// GetAllAssetsWithBalance 获取所有启用币种及用户余额（包括余额为0的）
// 统一使用 UserAssets 表
func (s *WalletService) GetAllAssetsWithBalance(userID uint) (*AssetOverviewResponse, error) {
	return GetUserAssetsService().GetAssetOverviewFromUserAssets(userID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetWalletByCurrency 获取用户指定币种钱包
func (s *WalletService) GetWalletByCurrency(userID, currencyID uint) (*model.UsersWallet, error) {
	var wallet model.UsersWallet
	result := database.DB.Where("user_id = ? AND currency = ?", userID, currencyID).First(&wallet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 钱包不存在，自动创建
			logger.Infof("用户钱包不存在，自动创建: userID=%d, currencyID=%d", userID, currencyID)
			newWallet := &model.UsersWallet{
				UserID:          userID,
				CurrencyID:      currencyID,
				UsdtBalance:     0,
				LockUsdtBalance: 0,
				Address:         "",
				Status:          1,
				CreateTime:      time.Now().Unix(),
			}
			if err := database.DB.Create(newWallet).Error; err != nil {
				return nil, fmt.Errorf("创建钱包失败: %w", err)
			}
			return newWallet, nil
		}
		return nil, result.Error
	}
	return &wallet, nil
}

// Transfer 账户间划转（已统一为USDT余额，不再支持划转）
func (s *WalletService) Transfer(userID uint, req *TransferRequest) error {
	return errors.New("已统一为USDT余额，不再支持账户间划转")
}

// Withdraw 提现申请
func (s *WalletService) Withdraw(userID uint, req *WithdrawRequest) error {
	// 验证支付密码
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.PayPassword == "" {
		return errors.New("请先设置支付密码")
	}

	if MakePayPassword(req.PayPassword) != user.PayPassword {
		return errors.New("支付密码错误")
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取钱包并加锁
	var wallet model.UsersWallet
	result := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("user_id = ? AND currency = ?", userID, req.CurrencyID).
		First(&wallet)
	if result.Error != nil {
		tx.Rollback()
		return errors.New("钱包不存在")
	}

	// 检查余额（使用统一USDT余额）
	if wallet.UsdtBalance < req.Amount {
		tx.Rollback()
		return errors.New("余额不足")
	}

	// 获取币种信息
	var currency model.Currency
	if err := database.DB.First(&currency, req.CurrencyID).Error; err != nil {
		tx.Rollback()
		return errors.New("币种不存在")
	}

	// 扣减余额，冻结
	if err := tx.Model(&wallet).Updates(map[string]interface{}{
		"legal_balance":      gorm.Expr("legal_balance - ?", req.Amount),
		"lock_legal_balance": gorm.Expr("lock_legal_balance + ?", req.Amount),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("扣减余额失败: %w", err)
	}

	// 创建提现记录
	withdrawRecord := &model.UsersWalletOut{
		UserID:     userID,
		CurrencyID: req.CurrencyID,
		Address:    req.Address,
		Number:     req.Amount,
		Status:     1, // 待审核
		CreateTime: time.Now().Unix(),
	}
	if err := tx.Create(withdrawRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建提现记录失败: %w", err)
	}

	// 记录日志
	s.addAccountLog(tx, userID, req.CurrencyID, -req.Amount, "申请提币扣除余额", 200)
	s.addAccountLog(tx, userID, req.CurrencyID, req.Amount, "申请提币冻结余额", 201)

	tx.Commit()
	logger.Infof("用户提现申请成功: userID=%d, currencyID=%d, amount=%f", userID, req.CurrencyID, req.Amount)
	return nil
}

// GetWalletLogs 获取钱包流水
func (s *WalletService) GetWalletLogs(userID uint, currencyID uint, page, pageSize int) ([]model.AccountLog, int64, error) {
	var logs []model.AccountLog
	var total int64

	query := database.DB.Model(&model.AccountLog{}).Where("user_id = ?", userID)
	if currencyID > 0 {
		query = query.Where("currency = ?", currencyID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	result := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return logs, total, nil
}

// GetRechargeAddress 获取充值地址
func (s *WalletService) GetRechargeAddress(userID, currencyID uint) (string, error) {
	var wallet model.UsersWallet
	result := database.DB.Where("user_id = ? AND currency = ?", userID, currencyID).First(&wallet)
	if result.Error != nil {
		return "", errors.New("钱包不存在")
	}

	if wallet.Address == "" {
		// TODO: 调用区块链接口生成地址
		return "", errors.New("暂无充值地址，请联系客服")
	}

	return wallet.Address, nil
}

// Recharge 充值申请
func (s *WalletService) Recharge(userID uint, req *RechargeRequest) error {
	// 创建充值申请记录
	chargeReq := &model.ChargeReq{
		UID:         int(userID),
		CurrencyID:  req.CurrencyID,
		Amount:      req.Amount,
		UserAccount: req.Account,
		Status:      1, // 待审核
		ToAddress:   req.Type,
		Image:       req.Image,
		CreatedAt:   time.Now(),
		Remark:      "",
	}

	if err := database.DB.Create(chargeReq).Error; err != nil {
		logger.Errorf("创建充值申请失败: userID=%d, err=%v", userID, err)
		return fmt.Errorf("创建充值申请失败: %w", err)
	}

	logger.Infof("用户充值申请成功: userID=%d, currencyID=%d, amount=%f", userID, req.CurrencyID, req.Amount)
	return nil
}

// addAccountLog 添加账户日志
func (s *WalletService) addAccountLog(tx *gorm.DB, userID, currencyID uint, value float64, info string, logType int) {
	log := &model.AccountLog{
		UserID:      userID,
		Value:       value,
		Info:        info,
		Type:        logType,
		CurrencyID:  currencyID,
		CreatedTime: time.Now().Unix(),
	}
	tx.Create(log)
}

// CreateUserWallets 创建用户钱包 (注册时调用)
func (s *WalletService) CreateUserWallets(userID uint) error {
	var currencies []model.Currency
	database.DB.Where("is_display = 1").Find(&currencies)

	for _, currency := range currencies {
		wallet := &model.UsersWallet{
			UserID:          userID,
			CurrencyID:      currency.ID,
			UsdtBalance:     0,
			LockUsdtBalance: 0,
			Status:          1,
			CreateTime:      time.Now().Unix(),
		}
		if err := database.DB.Create(wallet).Error; err != nil {
			logger.Errorf("创建钱包失败: userID=%d, currencyID=%d, err=%v", userID, currency.ID, err)
		}
	}

	return nil
}

// TodayProfitLossResponse 今日盈亏响应
type TodayProfitLossResponse struct {
	TodayProfit      float64 `json:"today_profit"`       // 今日总盈利
	TodayLoss        float64 `json:"today_loss"`         // 今日总亏损
	NetProfitLoss    float64 `json:"net_profit_loss"`    // 今日净盈亏
	MicroProfit      float64 `json:"micro_profit"`       // 秒合约盈利
	MicroLoss        float64 `json:"micro_loss"`         // 秒合约亏损
	LeverProfit      float64 `json:"lever_profit"`       // 杠杆交易盈利
	LeverLoss        float64 `json:"lever_loss"`         // 杠杆交易亏损
	SpotProfit       float64 `json:"spot_profit"`        // 币币交易盈利
	SpotLoss         float64 `json:"spot_loss"`          // 币币交易亏损
	ProfitOrderCount int64   `json:"profit_order_count"` // 盈利订单数
	LossOrderCount   int64   `json:"loss_order_count"`   // 亏损订单数
	TotalOrderCount  int64   `json:"total_order_count"`  // 总订单数
	WinRate          float64 `json:"win_rate"`           // 胜率
	StartTime        int64   `json:"start_time"`         // 开始时间（北京时间0点）
	EndTime          int64   `json:"end_time"`           // 结束时间（当前时间）
}

// GetTodayProfitLoss 获取今日盈亏统计（北京时间）
func (s *WalletService) GetTodayProfitLoss(userID uint) (*TodayProfitLossResponse, error) {
	// 使用北京时间（UTC+8）
	beijingLocation := time.FixedZone("CST", 8*3600)
	now := time.Now().In(beijingLocation)

	// 今日0点（北京时间）
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, beijingLocation)
	todayStartUnix := todayStart.Unix()
	nowUnix := now.Unix()

	result := &TodayProfitLossResponse{
		StartTime: todayStartUnix,
		EndTime:   nowUnix,
	}

	// 1. 统计秒合约盈亏（已结算的订单）
	var microProfitSum, microLossSum float64
	var microProfitCount, microLossCount int64

	// 秒合约盈利
	database.DB.Model(&model.MicroOrder{}).
		Where("user_id = ? AND status = 2 AND fact_profits > 0 AND complete_time >= ?", userID, todayStartUnix).
		Select("COALESCE(SUM(fact_profits), 0)").
		Scan(&microProfitSum)

	database.DB.Model(&model.MicroOrder{}).
		Where("user_id = ? AND status = 2 AND fact_profits > 0 AND complete_time >= ?", userID, todayStartUnix).
		Count(&microProfitCount)

	// 秒合约亏损
	database.DB.Model(&model.MicroOrder{}).
		Where("user_id = ? AND status = 2 AND fact_profits < 0 AND complete_time >= ?", userID, todayStartUnix).
		Select("COALESCE(SUM(ABS(fact_profits)), 0)").
		Scan(&microLossSum)

	database.DB.Model(&model.MicroOrder{}).
		Where("user_id = ? AND status = 2 AND fact_profits < 0 AND complete_time >= ?", userID, todayStartUnix).
		Count(&microLossCount)

	result.MicroProfit = microProfitSum
	result.MicroLoss = microLossSum

	// 2. 统计杠杆交易盈亏（已平仓的订单）
	var leverProfitSum, leverLossSum float64
	var leverProfitCount, leverLossCount int64

	// 杠杆交易盈利
	database.DB.Model(&model.LeverTransaction{}).
		Where("user_id = ? AND status = 3 AND fact_profits > 0 AND handle_time >= ?", userID, todayStartUnix).
		Select("COALESCE(SUM(fact_profits), 0)").
		Scan(&leverProfitSum)

	database.DB.Model(&model.LeverTransaction{}).
		Where("user_id = ? AND status = 3 AND fact_profits > 0 AND handle_time >= ?", userID, todayStartUnix).
		Count(&leverProfitCount)

	// 杠杆交易亏损
	database.DB.Model(&model.LeverTransaction{}).
		Where("user_id = ? AND status = 3 AND fact_profits < 0 AND handle_time >= ?", userID, todayStartUnix).
		Select("COALESCE(SUM(ABS(fact_profits)), 0)").
		Scan(&leverLossSum)

	database.DB.Model(&model.LeverTransaction{}).
		Where("user_id = ? AND status = 3 AND fact_profits < 0 AND handle_time >= ?", userID, todayStartUnix).
		Count(&leverLossCount)

	result.LeverProfit = leverProfitSum
	result.LeverLoss = leverLossSum

	// 3. 统计币币交易盈亏（从成交记录计算）
	var spotProfitSum, spotLossSum float64
	var spotProfitCount, spotLossCount int64

	// 币币交易通过 transaction_complete 表统计
	// 假设买入时记录为正，卖出收益为盈利
	database.DB.Model(&model.TransactionComplete{}).
		Where("user_id = ? AND create_time >= ? AND profit > 0", userID, todayStartUnix).
		Select("COALESCE(SUM(profit), 0)").
		Scan(&spotProfitSum)

	database.DB.Model(&model.TransactionComplete{}).
		Where("user_id = ? AND create_time >= ? AND profit > 0", userID, todayStartUnix).
		Count(&spotProfitCount)

	database.DB.Model(&model.TransactionComplete{}).
		Where("user_id = ? AND create_time >= ? AND profit < 0", userID, todayStartUnix).
		Select("COALESCE(SUM(ABS(profit)), 0)").
		Scan(&spotLossSum)

	database.DB.Model(&model.TransactionComplete{}).
		Where("user_id = ? AND create_time >= ? AND profit < 0", userID, todayStartUnix).
		Count(&spotLossCount)

	result.SpotProfit = spotProfitSum
	result.SpotLoss = spotLossSum

	// 4. 汇总统计
	result.TodayProfit = microProfitSum + leverProfitSum + spotProfitSum
	result.TodayLoss = microLossSum + leverLossSum + spotLossSum
	result.NetProfitLoss = result.TodayProfit - result.TodayLoss

	result.ProfitOrderCount = microProfitCount + leverProfitCount + spotProfitCount
	result.LossOrderCount = microLossCount + leverLossCount + spotLossCount
	result.TotalOrderCount = result.ProfitOrderCount + result.LossOrderCount

	// 计算胜率
	if result.TotalOrderCount > 0 {
		result.WinRate = float64(result.ProfitOrderCount) / float64(result.TotalOrderCount) * 100
	}

	logger.Infof("用户今日盈亏统计: userID=%d, profit=%.2f, loss=%.2f, net=%.2f, winRate=%.2f%%",
		userID, result.TodayProfit, result.TodayLoss, result.NetProfitLoss, result.WinRate)

	return result, nil
}

// DailyProfitLoss 单日盈亏数据
type DailyProfitLoss struct {
	Date          string  `json:"date"`            // 日期（YYYY-MM-DD）
	Profit        float64 `json:"profit"`          // 当日盈利
	Loss          float64 `json:"loss"`            // 当日亏损
	NetProfitLoss float64 `json:"net_profit_loss"` // 净盈亏
	MicroProfit   float64 `json:"micro_profit"`    // 秒合约盈利
	MicroLoss     float64 `json:"micro_loss"`      // 秒合约亏损
	LeverProfit   float64 `json:"lever_profit"`    // 杠杆盈利
	LeverLoss     float64 `json:"lever_loss"`      // 杠杆亏损
	SpotProfit    float64 `json:"spot_profit"`     // 币币盈利
	SpotLoss      float64 `json:"spot_loss"`       // 币币亏损
	OrderCount    int64   `json:"order_count"`     // 订单总数
	WinRate       float64 `json:"win_rate"`        // 胜率
	Timestamp     int64   `json:"timestamp"`       // 日期时间戳（当天0点）
}

// SevenDaysProfitLossResponse 七天盈亏折线图响应
type SevenDaysProfitLossResponse struct {
	DailyData       []DailyProfitLoss `json:"daily_data"`        // 每日数据
	TotalProfit     float64           `json:"total_profit"`      // 7天总盈利
	TotalLoss       float64           `json:"total_loss"`        // 7天总亏损
	TotalNetProfit  float64           `json:"total_net_profit"`  // 7天净盈亏
	AvgDailyProfit  float64           `json:"avg_daily_profit"`  // 日均盈利
	AvgDailyLoss    float64           `json:"avg_daily_loss"`    // 日均亏损
	MaxDailyProfit  float64           `json:"max_daily_profit"`  // 单日最高盈利
	MaxDailyLoss    float64           `json:"max_daily_loss"`    // 单日最高亏损
	TotalOrderCount int64             `json:"total_order_count"` // 7天总订单数
	AvgWinRate      float64           `json:"avg_win_rate"`      // 平均胜率
	ProfitDays      int               `json:"profit_days"`       // 盈利天数
	LossDays        int               `json:"loss_days"`         // 亏损天数
	StartDate       string            `json:"start_date"`        // 开始日期
	EndDate         string            `json:"end_date"`          // 结束日期
}

// GetSevenDaysProfitLoss 获取最近7天盈亏统计（北京时间）
func (s *WalletService) GetSevenDaysProfitLoss(userID uint) (*SevenDaysProfitLossResponse, error) {
	// 使用北京时间（UTC+8）
	beijingLocation := time.FixedZone("CST", 8*3600)
	now := time.Now().In(beijingLocation)

	// 今日0点
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, beijingLocation)

	// 初始化响应
	result := &SevenDaysProfitLossResponse{
		DailyData: make([]DailyProfitLoss, 0, 7),
	}

	var totalWinRate float64
	var validDays int // 有交易的天数

	// 遍历最近7天
	for i := 6; i >= 0; i-- {
		dayStart := today.AddDate(0, 0, -i)
		dayEnd := dayStart.Add(24 * time.Hour)
		dayStartUnix := dayStart.Unix()
		dayEndUnix := dayEnd.Unix()

		dailyData := DailyProfitLoss{
			Date:      dayStart.Format("2006-01-02"),
			Timestamp: dayStartUnix,
		}

		// 1. 统计秒合约盈亏
		var microProfit, microLoss float64
		var microProfitCount, microLossCount int64

		database.DB.Model(&model.MicroOrder{}).
			Where("user_id = ? AND status = 2 AND fact_profits > 0 AND complete_time >= ? AND complete_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Select("COALESCE(SUM(fact_profits), 0)").
			Scan(&microProfit)

		database.DB.Model(&model.MicroOrder{}).
			Where("user_id = ? AND status = 2 AND fact_profits > 0 AND complete_time >= ? AND complete_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Count(&microProfitCount)

		database.DB.Model(&model.MicroOrder{}).
			Where("user_id = ? AND status = 2 AND fact_profits < 0 AND complete_time >= ? AND complete_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Select("COALESCE(SUM(ABS(fact_profits)), 0)").
			Scan(&microLoss)

		database.DB.Model(&model.MicroOrder{}).
			Where("user_id = ? AND status = 2 AND fact_profits < 0 AND complete_time >= ? AND complete_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Count(&microLossCount)

		dailyData.MicroProfit = microProfit
		dailyData.MicroLoss = microLoss

		// 2. 统计杠杆交易盈亏
		var leverProfit, leverLoss float64
		var leverProfitCount, leverLossCount int64

		database.DB.Model(&model.LeverTransaction{}).
			Where("user_id = ? AND status = 3 AND fact_profits > 0 AND handle_time >= ? AND handle_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Select("COALESCE(SUM(fact_profits), 0)").
			Scan(&leverProfit)

		database.DB.Model(&model.LeverTransaction{}).
			Where("user_id = ? AND status = 3 AND fact_profits > 0 AND handle_time >= ? AND handle_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Count(&leverProfitCount)

		database.DB.Model(&model.LeverTransaction{}).
			Where("user_id = ? AND status = 3 AND fact_profits < 0 AND handle_time >= ? AND handle_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Select("COALESCE(SUM(ABS(fact_profits)), 0)").
			Scan(&leverLoss)

		database.DB.Model(&model.LeverTransaction{}).
			Where("user_id = ? AND status = 3 AND fact_profits < 0 AND handle_time >= ? AND handle_time < ?",
				userID, dayStartUnix, dayEndUnix).
			Count(&leverLossCount)

		dailyData.LeverProfit = leverProfit
		dailyData.LeverLoss = leverLoss

		// 3. 统计币币交易盈亏
		var spotProfit, spotLoss float64
		var spotProfitCount, spotLossCount int64

		database.DB.Model(&model.TransactionComplete{}).
			Where("user_id = ? AND create_time >= ? AND create_time < ? AND profit > 0",
				userID, dayStartUnix, dayEndUnix).
			Select("COALESCE(SUM(profit), 0)").
			Scan(&spotProfit)

		database.DB.Model(&model.TransactionComplete{}).
			Where("user_id = ? AND create_time >= ? AND create_time < ? AND profit > 0",
				userID, dayStartUnix, dayEndUnix).
			Count(&spotProfitCount)

		database.DB.Model(&model.TransactionComplete{}).
			Where("user_id = ? AND create_time >= ? AND create_time < ? AND profit < 0",
				userID, dayStartUnix, dayEndUnix).
			Select("COALESCE(SUM(ABS(profit)), 0)").
			Scan(&spotLoss)

		database.DB.Model(&model.TransactionComplete{}).
			Where("user_id = ? AND create_time >= ? AND create_time < ? AND profit < 0",
				userID, dayStartUnix, dayEndUnix).
			Count(&spotLossCount)

		dailyData.SpotProfit = spotProfit
		dailyData.SpotLoss = spotLoss

		// 4. 汇总单日数据
		dailyData.Profit = microProfit + leverProfit + spotProfit
		dailyData.Loss = microLoss + leverLoss + spotLoss
		dailyData.NetProfitLoss = dailyData.Profit - dailyData.Loss

		profitCount := microProfitCount + leverProfitCount + spotProfitCount
		lossCount := microLossCount + leverLossCount + spotLossCount
		dailyData.OrderCount = profitCount + lossCount

		// 计算单日胜率
		if dailyData.OrderCount > 0 {
			dailyData.WinRate = float64(profitCount) / float64(dailyData.OrderCount) * 100
			totalWinRate += dailyData.WinRate
			validDays++
		}

		// 统计盈亏天数
		if dailyData.NetProfitLoss > 0 {
			result.ProfitDays++
		} else if dailyData.NetProfitLoss < 0 {
			result.LossDays++
		}

		// 更新最大值
		if dailyData.Profit > result.MaxDailyProfit {
			result.MaxDailyProfit = dailyData.Profit
		}
		if dailyData.Loss > result.MaxDailyLoss {
			result.MaxDailyLoss = dailyData.Loss
		}

		// 累计总计
		result.TotalProfit += dailyData.Profit
		result.TotalLoss += dailyData.Loss
		result.TotalOrderCount += dailyData.OrderCount

		result.DailyData = append(result.DailyData, dailyData)
	}

	// 计算汇总数据
	result.TotalNetProfit = result.TotalProfit - result.TotalLoss
	result.AvgDailyProfit = result.TotalProfit / 7
	result.AvgDailyLoss = result.TotalLoss / 7

	if validDays > 0 {
		result.AvgWinRate = totalWinRate / float64(validDays)
	}

	// 设置日期范围
	if len(result.DailyData) > 0 {
		result.StartDate = result.DailyData[0].Date
		result.EndDate = result.DailyData[len(result.DailyData)-1].Date
	}

	logger.Infof("用户7天盈亏统计: userID=%d, totalProfit=%.2f, totalLoss=%.2f, netProfit=%.2f, profitDays=%d, lossDays=%d",
		userID, result.TotalProfit, result.TotalLoss, result.TotalNetProfit, result.ProfitDays, result.LossDays)

	return result, nil
}

// BroadcastBalanceUpdate 广播用户余额变更到WebSocket
// 当用户余额发生变化时调用此函数，实时推送给前端
func (s *WalletService) BroadcastBalanceUpdate(userID uint) {
	// 获取最新的资产概览数据
	assetOverview, err := s.GetAllAssetsWithBalance(userID)
	if err != nil {
		logger.Errorf("[WS] 获取用户余额失败: userID=%d, err=%v", userID, err)
		return
	}

	// 构建WebSocket消息数据
	data := map[string]interface{}{
		"user_id":           userID,
		"total_balance":     assetOverview.TotalBalance,
		"total_usd_value":   assetOverview.TotalUsdValue,
		"today_profit":      assetOverview.TodayProfit,
		"today_profit_rate": assetOverview.TodayProfitRate,
		"assets":            assetOverview.Assets,
		"update_time":       time.Now().Unix(),
	}

	// 推送到wallet频道（用户需要订阅 wallet:userID 频道）
	channelName := fmt.Sprintf("wallet:%d", userID)
	hub := websocket.GetHub()
	hub.SendToChannel(channelName, data)

	logger.Infof("[WS] 余额变更推送成功: userID=%d, channel=%s, balance=%.8f",
		userID, channelName, assetOverview.TotalBalance)
}

// BroadcastOrderUpdate 推送订单状态变更
func (s *WalletService) BroadcastOrderUpdate(userID uint, orderID uint, status int, dealNumber float64, price float64) {
	data := map[string]interface{}{
		"order_id":    orderID,
		"status":      status, // 0=待成交, 1=部分成交, 2=已成交, 3=已撤销
		"deal_number": dealNumber,
		"price":       price,
		"update_time": time.Now().Unix(),
	}

	// 推送到order频道
	channelName := fmt.Sprintf("order:%d", userID)
	hub := websocket.GetHub()
	hub.SendToChannel(channelName, data)

	logger.Infof("[WS] 订单状态推送: userID=%d, orderID=%d, status=%d", userID, orderID, status)
}
