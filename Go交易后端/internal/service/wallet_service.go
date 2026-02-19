package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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

// TransferRequest 划转请求（已废弃，请使用 WalletTransferService）
type TransferRequest struct {
	From       model.WalletType `json:"from" binding:"required"`
	To         model.WalletType `json:"to" binding:"required"`
	CurrencyID uint             `json:"currency_id" binding:"required"`
	Amount     float64          `json:"amount" binding:"required,gt=0"`
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
	CurrencyID   uint    `json:"currency_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	PayPassword  string  `json:"pay_password" binding:"required"`
	WithdrawType int8    `json:"withdraw_type"` // 提现类型:1=银行转账(默认),2=区块链提币
	// 银行转账字段
	Address string `json:"address"` // 银行卡号(兼容旧逻辑)
	// 区块链提币字段
	NetworkType  string `json:"network_type"`  // 网络类型:TRC20/ERC20/BRC20/BEP20/Polygon
	ChainAddress string `json:"chain_address"` // 区块链钱包地址
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Name    string  `json:"name"`    // 网络名称
	Fee     float64 `json:"fee"`     // 手续费
	MinFee  float64 `json:"min_fee"` // 最小手续费
	Enabled bool    `json:"enabled"` // 是否启用
}

// WithdrawConfigResponse 提现配置响应
type WithdrawConfigResponse struct {
	WithdrawMode      string        `json:"withdraw_mode"`       // 提现模式:bank/crypto/both
	BankEnabled       bool          `json:"bank_enabled"`        // 银行转账是否启用
	CryptoEnabled     bool          `json:"crypto_enabled"`      // 区块链提币是否启用
	Networks          []NetworkInfo `json:"networks"`            // 支持的网络列表
	WithdrawFeeRate   float64       `json:"withdraw_fee_rate"`   // 银行转账手续费率
	WithdrawMin       float64       `json:"withdraw_min"`        // 最小提现金额
	CryptoWithdrawFee float64       `json:"crypto_withdraw_fee"` // 区块链提币固定手续费
	UsdtCurrencyID    uint          `json:"usdt_currency_id"`    // USDT币种ID
}

// ValidateChainAddress 验证区块链地址格式
func ValidateChainAddress(networkType, address string) error {
	if address == "" {
		return errors.New("请输入区块链地址")
	}

	switch networkType {
	case "TRC20":
		// TRC20地址: T开头，34位
		if len(address) != 34 || address[0] != 'T' {
			return errors.New("TRC20地址格式错误，应为T开头的34位地址")
		}
	case "ERC20", "BEP20", "Polygon":
		// ERC20/BEP20/Polygon地址: 0x开头，42位
		if len(address) != 42 || address[:2] != "0x" {
			return errors.New(networkType + "地址格式错误，应为0x开头的42位地址")
		}
	case "BRC20":
		// BRC20地址: bc1开头(Bech32)或1/3开头(Legacy)
		if len(address) < 26 || len(address) > 62 {
			return errors.New("BRC20地址格式错误")
		}
		if address[:3] != "bc1" && address[0] != '1' && address[0] != '3' {
			return errors.New("BRC20地址格式错误，应为bc1/1/3开头")
		}
	default:
		return errors.New("不支持的网络类型: " + networkType)
	}

	return nil
}

// GetWalletList 获取用户钱包列表（已迁移到UserAssets）
func (s *WalletService) GetWalletList(userID uint) (*WalletListResponse, error) {
	// 统一使用 UserAssets 表获取用户资产
	assets, err := GetUserAssetsService().GetUserAssets(userID)
	if err != nil {
		return nil, err
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

	// 从UserAssets构建钱包列表
	for _, currency := range currencies {
		// 获取币种余额（从JSON字段）
		balance := assets.GetCurrencyBalance(currency.Name)
		locked := assets.GetCurrencyLocked(currency.Name)

		info := WalletBalanceInfo{
			ID:               uint(assets.ID),
			CurrencyID:       currency.ID,
			CurrencyName:     currency.Name,
			CurrencyLogo:     currency.Logo,
			Balance:          balance,
			LockedBalance:    locked,
			AvailableBalance: balance,
			UsdtValue:        balance * currency.Rate,
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

// GetWalletByCurrency 获取用户指定币种钱包（已迁移到UserAssets）
// 注意：此方法保留用于兼容旧代码，实际数据从UserAssets表读取
func (s *WalletService) GetWalletByCurrency(userID, currencyID uint) (*model.UserAssets, error) {
	// 统一使用 UserAssets 表
	assets, err := GetUserAssetsService().GetUserAssets(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户资产失败: %w", err)
	}
	return assets, nil
}

// Transfer 账户间划转（已统一为USDT余额，不再支持划转）
func (s *WalletService) Transfer(userID uint, req *TransferRequest) error {
	return errors.New("已统一为USDT余额，不再支持账户间划转")
}

// Withdraw 提现申请 (支持银行转账和区块链提币)
func (s *WalletService) Withdraw(userID uint, req *WithdrawRequest) error {
	// 默认为银行转账
	if req.WithdrawType == 0 {
		req.WithdrawType = 1
	}

	// 验证支付密码
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.PayPassword == "" {
		return errors.New("请先设置支付密码")
	}

	// 修正：使用 VerifyPayPassword 验证 bcrypt 哈希
	if !VerifyPayPassword(user.PayPassword, req.PayPassword) {
		return errors.New("支付密码错误")
	}

	// 获取币种信息
	var currency model.Currency
	if err := database.DB.First(&currency, req.CurrencyID).Error; err != nil {
		return errors.New("币种不存在")
	}

	// 获取用户实名信息
	var userReal model.UserReal
	database.DB.Where("user_id = ?", userID).First(&userReal)

	// 增加校验：必须实名认证通过才能提现
	if userReal.ReviewStatus != 2 {
		return errors.New("请先完成实名认证并通过审核")
	}

	// 创建提现记录
	withdrawRecord := &model.UsersWalletOut{
		UserID:       userID,
		CurrencyID:   req.CurrencyID,
		Number:       req.Amount,
		Status:       1, // 待审核
		CreateTime:   time.Now().Unix(),
		WithdrawType: req.WithdrawType,
		RealName:     userReal.Name,
		IDCard:       userReal.CardID,
	}

	// 根据提现类型进行不同处理
	if req.WithdrawType == 1 {
		// 银行转账：需要验证银行卡信息
		var userCash model.UserCashInfo
		if err := database.DB.Where("user_id = ?", userID).First(&userCash).Error; err != nil {
			// 尝试从KYC获取银行卡信息
			if userReal.BankCard == "" {
				return errors.New("请先绑定银行卡")
			}
			// 使用KYC中的银行卡信息
			withdrawRecord.BankAccount = userReal.BankCard
			withdrawRecord.BankName = "银行卡"
		} else {
			withdrawRecord.BankName = userCash.BankName
			withdrawRecord.BankBranch = userCash.BankBranch
			withdrawRecord.BankAccount = userCash.BankAccount
			if userCash.RealName != "" {
				withdrawRecord.RealName = userCash.RealName
			}
		}
		withdrawRecord.Address = withdrawRecord.BankAccount
	} else if req.WithdrawType == 2 {
		// 区块链提币：验证网络类型和地址
		if req.NetworkType == "" {
			return errors.New("请选择网络类型")
		}
		if err := ValidateChainAddress(req.NetworkType, req.ChainAddress); err != nil {
			return err
		}
		withdrawRecord.NetworkType = req.NetworkType
		withdrawRecord.ChainAddress = req.ChainAddress
		withdrawRecord.Address = req.ChainAddress // 兼容旧字段
	} else {
		return errors.New("不支持的提现类型")
	}

	// 开启事务
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 从资金钱包扣除余额（而不是从UserAssets）
		fundSvc := GetFundWalletService()
		amount := decimal.NewFromFloat(req.Amount)
		
		// 获取资金钱包并检查余额
		fundWallet, err := fundSvc.GetOrCreateFundWallet(uint64(userID), 1)
		if err != nil {
			return fmt.Errorf("获取资金钱包失败: %w", err)
		}

		// 检查资金钱包余额是否足够
		if fundWallet.AvailableBalance.LessThan(amount) {
			return fmt.Errorf("资金钱包余额不足:当前余额%s,需要%s", 
				fundWallet.AvailableBalance.String(), amount.String())
		}

		// 扣减资金钱包余额
		fundWallet.AvailableBalance = fundWallet.AvailableBalance.Sub(amount)
		fundWallet.LockedBalance = fundWallet.LockedBalance.Add(amount)
		fundWallet.TotalWithdraw = fundWallet.TotalWithdraw.Add(amount)
		fundWallet.UpdateTime = time.Now().Unix()

		if err := tx.Save(fundWallet).Error; err != nil {
			return fmt.Errorf("扣除资金钱包余额失败: %w", err)
		}

		// 2. 创建提现记录
		if err := tx.Create(withdrawRecord).Error; err != nil {
			return fmt.Errorf("创建提现记录失败: %w", err)
		}

		// 3. 记录账户日志 (使用 tx)
		logInfo := "申请提币扣除资金钱包余额"
		if req.WithdrawType == 2 {
			logInfo = fmt.Sprintf("申请区块链提币(%s)扣除资金钱包余额", req.NetworkType)
		}
		s.addAccountLog(tx, userID, req.CurrencyID, -req.Amount, logInfo, 200)
		s.addAccountLog(tx, userID, req.CurrencyID, req.Amount, "申请提币冻结资金钱包余额", 201)

		// 4. WebSocket 实时推送余额变更
		go s.BroadcastBalanceUpdate(userID)

		return nil
	})
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
// 注意：地址存储方式已变更，现在从currency表中获取或动态生成
func (s *WalletService) GetRechargeAddress(userID, currencyID uint) (string, error) {
	// 获取币种信息
	var currency model.Currency
	if err := database.DB.First(&currency, currencyID).Error; err != nil {
		return "", errors.New("币种不存在")
	}

	// 检查用户资产是否存在
	_, err := GetUserAssetsService().GetUserAssets(userID)
	if err != nil {
		return "", errors.New("用户资产不存在")
	}

	// TODO: 从数据库或区块链接口获取用户充值地址
	// 临时返回空，需要实现地址生成逻辑
	return "", errors.New("暂无充值地址，请联系客服")
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

// GetWithdrawalList 获取用户提现记录
func (s *WalletService) GetWithdrawalList(userID uint, page, pageSize int) ([]model.UsersWalletOut, int64, error) {
	var list []model.UsersWalletOut
	var total int64

	db := database.DB.Model(&model.UsersWalletOut{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetWithdrawConfig 获取提现配置
func (s *WalletService) GetWithdrawConfig() (*WithdrawConfigResponse, error) {
	resp := &WithdrawConfigResponse{
		WithdrawMode:      "both",
		BankEnabled:       true,
		CryptoEnabled:     true,
		Networks:          []NetworkInfo{},
		WithdrawFeeRate:   0,
		WithdrawMin:       10,
		CryptoWithdrawFee: 0,
	}

	// 从数据库读取配置
	var settings []model.Setting
	database.DB.Where("`key` IN (?)", []string{
		"withdraw_mode", "crypto_networks", "withdraw_fee_rate",
		"withdraw_min", "crypto_withdraw_fee",
	}).Find(&settings)

	for _, s := range settings {
		switch s.Key {
		case "withdraw_mode":
			resp.WithdrawMode = s.Value
		case "withdraw_fee_rate":
			if v, err := parseFloat(s.Value); err == nil {
				resp.WithdrawFeeRate = v
			}
		case "withdraw_min":
			if v, err := parseFloat(s.Value); err == nil {
				resp.WithdrawMin = v
			}
		case "crypto_withdraw_fee":
			if v, err := parseFloat(s.Value); err == nil {
				resp.CryptoWithdrawFee = v
			}
		case "crypto_networks":
			// 解析网络列表
			if s.Value != "" {
				networks := splitNetworks(s.Value)
				for _, n := range networks {
					resp.Networks = append(resp.Networks, NetworkInfo{
						Name:    n,
						Fee:     resp.CryptoWithdrawFee,
						MinFee:  resp.CryptoWithdrawFee,
						Enabled: true,
					})
				}
			}
		}
	}

	// 根据模式设置启用状态
	switch resp.WithdrawMode {
	case "bank":
		resp.BankEnabled = true
		resp.CryptoEnabled = false
	case "crypto":
		resp.BankEnabled = false
		resp.CryptoEnabled = true
	case "both":
		resp.BankEnabled = true
		resp.CryptoEnabled = true
	}

	// 如果没有配置网络，使用默认网络列表
	if len(resp.Networks) == 0 {
		defaultNetworks := []string{"TRC20", "ERC20", "BEP20", "Polygon", "BRC20"}
		for _, n := range defaultNetworks {
			resp.Networks = append(resp.Networks, NetworkInfo{
				Name:    n,
				Fee:     resp.CryptoWithdrawFee,
				MinFee:  resp.CryptoWithdrawFee,
				Enabled: true,
			})
		}
	}

	// 查询USDT的currency_id
	var usdtCurrency model.Currency
	if err := database.DB.Where("name = ?", "USDT").First(&usdtCurrency).Error; err == nil {
		resp.UsdtCurrencyID = usdtCurrency.ID
	}

	return resp, nil
}

// parseFloat 解析浮点数
func parseFloat(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// splitNetworks 分割网络列表
func splitNetworks(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
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
// 注意：已迁移到UserAssets表，需要为用户创建现货和合约两个钱包
func (s *WalletService) CreateUserWallets(userID uint) error {
	// 创建现货钱包
	_, err := GetUserAssetsService().CreateUserAssetsByType(userID, model.WalletTypeSpot)
	if err != nil {
		logger.Errorf("创建现货钱包失败: userID=%d, err=%v", userID, err)
		return err
	}
	logger.Infof("创建现货钱包成功: userID=%d", userID)

	// 创建合约钱包
	_, err = GetUserAssetsService().CreateUserAssetsByType(userID, model.WalletTypeContract)
	if err != nil {
		logger.Errorf("创建合约钱包失败: userID=%d, err=%v", userID, err)
		return err
	}
	logger.Infof("创建合约钱包成功: userID=%d", userID)

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

	// 获取现货、合约和交割钱包余额
	var spotBalance, contractBalance, deliveryBalance string = "0", "0", "0"
	walletTransferSvc := GetWalletTransferService()
	walletBalance, err := walletTransferSvc.GetUserAllWallets(uint64(userID))
	if err != nil {
		logger.Warnf("[WS] 获取钱包余额失败: userID=%d, err=%v", userID, err)
	} else {
		spotBalance = walletBalance.SpotBalance.StringFixed(4)
		contractBalance = walletBalance.ContractBalance.StringFixed(4)
		deliveryBalance = walletBalance.DeliveryBalance.StringFixed(4)
	}

	// 获取资金钱包余额
	var fundBalance string = "0"
	fundSvc := GetFundWalletService()
	fundWallet, err := fundSvc.GetOrCreateFundWallet(uint64(userID), 1)
	if err != nil {
		logger.Warnf("[WS] 获取资金钱包余额失败: userID=%d, err=%v", userID, err)
	} else {
		fundBalance = fundWallet.AvailableBalance.StringFixed(4)
	}

	// 构建WebSocket消息数据
	data := map[string]interface{}{
		"user_id":           userID,
		"total_balance":     assetOverview.TotalBalance,
		"total_usd_value":   assetOverview.TotalUsdValue,
		"today_profit":      assetOverview.TodayProfit,
		"today_profit_rate": assetOverview.TodayProfitRate,
		"spot_balance":      spotBalance,
		"contract_balance":  contractBalance,
		"delivery_balance":  deliveryBalance,
		"fund_balance":      fundBalance,
		"assets":            assetOverview.Assets,
		"update_time":       time.Now().Unix(),
	}

	// 推送到wallet频道（用户需要订阅 wallet:userID 频道）
	channelName := fmt.Sprintf("wallet:%d", userID)
	hub := websocket.GetHub()
	hub.SendToChannel(channelName, data)

	logger.Infof("[WS] 余额变更推送成功: userID=%d, channel=%s, spot=%s, contract=%s, delivery=%s, fund=%s, balance=%.8f",
		userID, channelName, spotBalance, contractBalance, deliveryBalance, fundBalance, assetOverview.TotalBalance)
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
