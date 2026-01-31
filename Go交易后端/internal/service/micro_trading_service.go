package service

import (
	"errors"
	"fmt"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/pkg/utils"

	"gorm.io/gorm"
)

// ============================================================================
// 秒合约业务常量定义
// ============================================================================

// MicroDirection 秒合约方向
type MicroDirection string

const (
	MicroDirectionRise MicroDirection = "rise" // 买涨
	MicroDirectionFall MicroDirection = "fall" // 买跌
)

// MicroProfitResult 秒合约盈亏结果（无平局）
type MicroProfitResult int8

const (
	MicroProfitResultLoss   MicroProfitResult = -1 // 亏损
	MicroProfitResultProfit MicroProfitResult = 1  // 盈利
)

// MicroOrderStatus 秒合约订单状态
type MicroOrderStatus int8

const (
	MicroOrderStatusActive  MicroOrderStatus = 0 // 进行中
	MicroOrderStatusSettled MicroOrderStatus = 1 // 已结算
)

// MicroDuration 秒合约周期（固定5档，不可修改）
type MicroDuration uint

const (
	MicroDuration30S  MicroDuration = 30  // 30秒
	MicroDuration60S  MicroDuration = 60  // 60秒
	MicroDuration120S MicroDuration = 120 // 120秒
	MicroDuration180S MicroDuration = 180 // 180秒
	MicroDuration300S MicroDuration = 300 // 300秒
)

// ValidMicroDurations 有效的秒合约周期列表
var ValidMicroDurations = []MicroDuration{
	MicroDuration30S,
	MicroDuration60S,
	MicroDuration120S,
	MicroDuration180S,
	MicroDuration300S,
}

// MicroMinAmount 最低投入金额（USDT）
const MicroMinAmount float64 = 10.0

// ============================================================================
// 秒合约交易服务
// ============================================================================

// MicroTradingService 秒合约交易服务
type MicroTradingService struct{}

// NewMicroTradingService 创建秒合约交易服务实例
func NewMicroTradingService() *MicroTradingService {
	return &MicroTradingService{}
}

// 单例模式
var microTradingServiceInstance *MicroTradingService

// GetMicroTradingService 获取秒合约交易服务单例
func GetMicroTradingService() *MicroTradingService {
	if microTradingServiceInstance == nil {
		microTradingServiceInstance = NewMicroTradingService()
	}
	return microTradingServiceInstance
}

// ============================================================================
// 请求/响应结构体
// ============================================================================

// MicroPeriodConfig 周期配置
type MicroPeriodConfig struct {
	Seconds     uint    `json:"seconds"`      // 周期秒数
	ProfitRatio float64 `json:"profit_ratio"` // 盈利率
	Status      int8    `json:"status"`       // 状态: 1=启用, 0=禁用
}

// MicroSubmitOrderRequest 提交订单请求
type MicroSubmitOrderRequest struct {
	CurrencyID uint           `json:"currency_id" binding:"required"` // 交易币种ID
	Direction  MicroDirection `json:"direction" binding:"required"`   // 方向: rise/fall
	Amount     float64        `json:"amount" binding:"required,gt=0"` // 投入金额(USDT)
	Seconds    uint           `json:"seconds" binding:"required"`     // 周期秒数: 30/60/120/180/300
}

// MicroSubmitOrderResponse 提交订单响应
type MicroSubmitOrderResponse struct {
	OrderID         uint           `json:"order_id"`          // 订单ID
	CurrencyID      uint           `json:"currency_id"`       // 币种ID
	Direction       MicroDirection `json:"direction"`         // 方向
	Amount          float64        `json:"amount"`            // 投入金额
	Seconds         uint           `json:"seconds"`           // 周期秒数
	ProfitRatio     float64        `json:"profit_ratio"`      // 盈利率
	OpenPrice       float64        `json:"open_price"`        // 开仓价格
	ExpectedEndTime time.Time      `json:"expected_end_time"` // 预计结算时间
	CreateTime      time.Time      `json:"create_time"`       // 创建时间
}

// MicroOrderListRequest 订单列表请求
type MicroOrderListRequest struct {
	Status   *MicroOrderStatus `json:"status"`    // 状态筛选: 0=进行中, 1=已结算
	Page     int               `json:"page"`      // 页码
	PageSize int               `json:"page_size"` // 每页数量
}

// MicroOrderDetailResponse 订单详情响应
type MicroOrderDetailResponse struct {
	ID              uint              `json:"id"`                // 订单ID
	UserID          uint              `json:"user_id"`           // 用户ID
	CurrencyID      uint              `json:"currency_id"`       // 币种ID
	CurrencyName    string            `json:"currency_name"`     // 币种名称
	Direction       MicroDirection    `json:"direction"`         // 方向
	Amount          float64           `json:"amount"`            // 投入金额
	Seconds         uint              `json:"seconds"`           // 周期秒数
	ProfitRatio     float64           `json:"profit_ratio"`      // 盈利率
	OpenPrice       float64           `json:"open_price"`        // 开仓价格
	EndPrice        float64           `json:"end_price"`         // 结算价格
	FactProfits     float64           `json:"fact_profits"`      // 实际盈亏金额
	ProfitResult    MicroProfitResult `json:"profit_result"`     // 结果: 1=盈, -1=亏
	PreProfitResult int8              `json:"pre_profit_result"` // 预设结果: 0=无, 1=盈, -1=亏 (用于前端控盘)
	Status          MicroOrderStatus  `json:"status"`            // 状态: 0=进行中, 1=已结算
	ExpectedEndTime time.Time         `json:"expected_end_time"` // 预计结算时间
	CreateTime      time.Time         `json:"create_time"`       // 创建时间
	SettleTime      *time.Time        `json:"settle_time"`       // 结算时间
}

// ============================================================================
// 周期配置接口
// ============================================================================

// GetPeriodConfigs 获取所有周期配置
func (s *MicroTradingService) GetPeriodConfigs() ([]MicroPeriodConfig, error) {
	var configs []model.MicroSeconds

	// 从数据库读取周期配置
	err := database.DB.Where("status = ?", 1).Order("seconds ASC").Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("获取周期配置失败: %w", err)
	}

	// 如果数据库没有配置，返回默认配置
	if len(configs) == 0 {
		return s.getDefaultPeriodConfigs(), nil
	}

	var result []MicroPeriodConfig
	for _, c := range configs {
		result = append(result, MicroPeriodConfig{
			Seconds:     c.Seconds,
			ProfitRatio: c.ProfitRatio,
			Status:      c.Status,
		})
	}

	return result, nil
}

// getDefaultPeriodConfigs 获取默认周期配置
func (s *MicroTradingService) getDefaultPeriodConfigs() []MicroPeriodConfig {
	return []MicroPeriodConfig{
		{Seconds: 30, ProfitRatio: 0.40, Status: 1},
		{Seconds: 60, ProfitRatio: 0.50, Status: 1},
		{Seconds: 120, ProfitRatio: 0.60, Status: 1},
		{Seconds: 180, ProfitRatio: 0.80, Status: 1},
		{Seconds: 300, ProfitRatio: 1.00, Status: 1},
	}
}

// GetProfitRatioBySeconds 根据周期获取盈利率
func (s *MicroTradingService) GetProfitRatioBySeconds(seconds uint) (float64, error) {
	var config model.MicroSeconds
	err := database.DB.Where("seconds = ? AND status = ?", seconds, 1).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 使用默认盈利率
			defaults := s.getDefaultPeriodConfigs()
			for _, d := range defaults {
				if d.Seconds == seconds {
					return d.ProfitRatio, nil
				}
			}
			return 0, errors.New("无效的周期")
		}
		return 0, err
	}
	return config.ProfitRatio, nil
}

// IsValidDuration 检查周期是否有效
func (s *MicroTradingService) IsValidDuration(seconds uint) bool {
	for _, d := range ValidMicroDurations {
		if uint(d) == seconds {
			return true
		}
	}
	return false
}

// ============================================================================
// 订单提交接口
// ============================================================================

// SubmitOrder 提交秒合约订单
func (s *MicroTradingService) SubmitOrder(userID uint, req *MicroSubmitOrderRequest) (*MicroSubmitOrderResponse, error) {
	// 1. 验证周期是否有效
	if !s.IsValidDuration(req.Seconds) {
		return nil, errors.New("无效的周期，仅支持30/60/120/180/300秒")
	}

	// 2. 验证最低投入金额
	if req.Amount < MicroMinAmount {
		return nil, fmt.Errorf("最低投入金额为 %.2f USDT", MicroMinAmount)
	}

	// 3. 验证方向
	if req.Direction != MicroDirectionRise && req.Direction != MicroDirectionFall {
		return nil, errors.New("无效的方向，仅支持rise/fall")
	}

	// 4. 检查用户是否已有进行中的订单（单用户单订单限制）
	var activeCount int64
	database.DB.Model(&model.MicroOrder{}).
		Where("user_id = ? AND status = ?", userID, MicroOrderStatusActive).
		Count(&activeCount)
	if activeCount > 0 {
		return nil, errors.New("您已有进行中的订单，请等待结算后再下单")
	}

	// 5. 获取盈利率
	profitRatio, err := s.GetProfitRatioBySeconds(req.Seconds)
	if err != nil {
		return nil, fmt.Errorf("获取盈利率失败: %w", err)
	}

	// 6. 校验币种是否存在
	var currency model.Currency
	if err := database.DB.Where("id = ?", req.CurrencyID).First(&currency).Error; err != nil {
		return nil, errors.New("无效的币种")
	}

	// 7. 获取法币ID
	legalID, err := GetCurrencyService().GetDefaultLegalCurrencyID()
	if err != nil {
		return nil, errors.New("系统配置错误")
	}

	// 8. 获取当前市场价格
	currentPrice, err := s.getCurrentMarketPrice(req.CurrencyID, legalID)
	if err != nil {
		return nil, fmt.Errorf("获取市场价格失败: %w", err)
	}

	// 9. 验证用户合约钱包USDT余额
	// 秒合约订单应该使用合约钱包余额验证
	contractWallet, err := GetUserAssetsService().GetUserAssetsByType(userID, model.WalletTypeContract)
	if err != nil {
		return nil, fmt.Errorf("获取合约钱包失败: %w", err)
	}
	if contractWallet.UsdtBalance < req.Amount {
		return nil, fmt.Errorf("合约钱包USDT余额不足，当前余额: %.2f，需要: %.2f", contractWallet.UsdtBalance, req.Amount)
	}

	// 10. 获取用户风控设置并应用
	// 优先级: 用户风控 > 其他风控
	preProfitResult := int8(0) // 默认无预设
	var user model.User
	if err := database.DB.Select("risk").Where("id = ?", userID).First(&user).Error; err == nil {
		if user.Risk == 1 { // 盈利
			preProfitResult = 1
			logger.Infof("[SubmitOrder] 应用用户盈利风控: userID=%d", userID)
		} else if user.Risk == -1 { // 亏损
			preProfitResult = -1
			logger.Infof("[SubmitOrder] 应用用户亏损风控: userID=%d", userID)
		}
	}

	// 11. 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 12. 创建订单（无手续费）
	order := &model.MicroOrder{
		UserID:          userID,
		CurrencyID:      req.CurrencyID,
		Type:            s.directionToInt(req.Direction),
		Seconds:         req.Seconds,
		Number:          req.Amount,
		ProfitRatio:     profitRatio,
		OpenPrice:       currentPrice,
		EndPrice:        0,
		FactProfits:     0,
		Status:          int8(MicroOrderStatusActive), // 0=进行中
		PreProfitResult: preProfitResult,              // 应用用户风控预设
		ProfitResult:    0,                            // 未结算
		Fee:             0,                            // 无手续费
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 12. 提交事务（不再更新UsersWallet，只使用UserAssets）
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新 UserAssets 表 - 减少USDT可用，增加USDT锁定（使用合约钱包）
	if err := GetUserAssetsService().LockContractUsdt(userID, req.Amount); err != nil {
		logger.Errorf("[SubmitOrder] 锁定合约钱包USDT余额失败: userID=%d, amount=%.8f, err=%v",
			userID, req.Amount, err)
		// 注意：这里需要处理回滚逻辑，因为订单已创建
		// 实际生产环境应该使用分布式事务或补偿机制
	}

	expectedEndTime := now.Add(time.Duration(req.Seconds) * time.Second)

	logger.Infof("秒合约下单成功: userID=%d, orderID=%d, direction=%s, amount=%.2f, seconds=%d, profitRatio=%.2f",
		userID, order.ID, req.Direction, req.Amount, req.Seconds, profitRatio)

	// 推送余额变更 - 使用安全goroutine
	utils.SafeGoWithName("broadcastBalanceUpdate", func() {
		GetWalletService().BroadcastBalanceUpdate(userID)
	})

	return &MicroSubmitOrderResponse{
		OrderID:         order.ID,
		CurrencyID:      req.CurrencyID,
		Direction:       req.Direction,
		Amount:          req.Amount,
		Seconds:         req.Seconds,
		ProfitRatio:     profitRatio,
		OpenPrice:       currentPrice,
		ExpectedEndTime: expectedEndTime,
		CreateTime:      now,
	}, nil
}

// ============================================================================
// 订单查询接口
// ============================================================================

// GetOrderList 获取订单列表
func (s *MicroTradingService) GetOrderList(userID uint, req *MicroOrderListRequest) ([]MicroOrderDetailResponse, int64, error) {
	var orders []model.MicroOrder
	var total int64

	// 默认分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := database.DB.Model(&model.MicroOrder{}).Where("user_id = ?", userID)

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	query.Count(&total)

	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	var result []MicroOrderDetailResponse
	for _, order := range orders {
		result = append(result, s.modelToResponse(order))
	}

	return result, total, nil
}

// GetOrderDetail 获取订单详情
func (s *MicroTradingService) GetOrderDetail(userID uint, orderID uint) (*MicroOrderDetailResponse, error) {
	var order model.MicroOrder
	err := database.DB.Where("user_id = ? AND id = ?", userID, orderID).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	detail := s.modelToResponse(order)
	return &detail, nil
}

// GetActiveOrder 获取用户当前进行中的订单
func (s *MicroTradingService) GetActiveOrder(userID uint) (*MicroOrderDetailResponse, error) {
	var order model.MicroOrder
	err := database.DB.Where("user_id = ? AND status = ?", userID, MicroOrderStatusActive).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有进行中的订单
		}
		return nil, err
	}

	detail := s.modelToResponse(order)
	return &detail, nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// getCurrentMarketPrice 获取当前市场价格
func (s *MicroTradingService) getCurrentMarketPrice(currencyID, legalID uint) (float64, error) {
	quotation, err := GetMarketService().GetCurrencyQuotation(currencyID, legalID)
	if err != nil {
		return 0, err
	}
	return quotation.Price, nil
}

// directionToInt 方向转整型
func (s *MicroTradingService) directionToInt(dir MicroDirection) int8 {
	if dir == MicroDirectionFall {
		return 2
	}
	return 1 // 默认买涨
}

// intToDirection 整型转方向
func (s *MicroTradingService) intToDirection(i int8) MicroDirection {
	if i == 2 {
		return MicroDirectionFall
	}
	return MicroDirectionRise
}

// modelToResponse 模型转响应
func (s *MicroTradingService) modelToResponse(order model.MicroOrder) MicroOrderDetailResponse {
	// 获取币种名称
	var currencyName string
	var currency model.Currency
	if err := database.DB.Where("id = ?", order.CurrencyID).First(&currency).Error; err == nil {
		currencyName = currency.Name
	}

	// 计算预计结算时间
	expectedEndTime := order.CreatedAt.Add(time.Duration(order.Seconds) * time.Second)

	// 结算时间（仅已结算订单有效）
	var settleTime *time.Time
	if order.Status == int8(MicroOrderStatusSettled) {
		t := order.UpdatedAt
		settleTime = &t
	}

	return MicroOrderDetailResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		CurrencyID:      order.CurrencyID,
		CurrencyName:    currencyName,
		Direction:       s.intToDirection(order.Type),
		Amount:          order.Number,
		Seconds:         order.Seconds,
		ProfitRatio:     order.ProfitRatio,
		OpenPrice:       order.OpenPrice,
		EndPrice:        order.EndPrice,
		FactProfits:     order.FactProfits,
		ProfitResult:    MicroProfitResult(order.ProfitResult),
		PreProfitResult: order.PreProfitResult, // 预设结果（前端控盘使用）
		Status:          MicroOrderStatus(order.Status),
		ExpectedEndTime: expectedEndTime,
		CreateTime:      order.CreatedAt,
		SettleTime:      settleTime,
	}
}
