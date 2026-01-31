package service

import (
	"errors"
	"fmt"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// ====================================
// 常量定义
// ====================================

// 手续费率: 0.0031%
const ContractFeeRate = 0.000031

// 最小开仓保证金: 10 USDT
const MinMargin = 10.0

// ====================================
// FuturesTradingService 合约交易服务
// ====================================

type FuturesTradingService struct{}

func NewFuturesTradingService() *FuturesTradingService {
	return &FuturesTradingService{}
}

// 全局服务实例
var futuresTradingService *FuturesTradingService

func GetFuturesTradingService() *FuturesTradingService {
	if futuresTradingService == nil {
		futuresTradingService = NewFuturesTradingService()
	}
	return futuresTradingService
}

// ====================================
// 请求/响应结构体
// ====================================

// OpenPositionRequest 开仓请求
type OpenPositionRequest struct {
	CurrencyID uint    `json:"currency_id" binding:"required"` // 币种ID
	LegalID    uint    `json:"legal_id" binding:"required"`    // 法币ID(默认USDT=3)
	Type       int8    `json:"type" binding:"required"`        // 方向:1=做多,2=做空
	OrderType  int8    `json:"order_type" binding:"required"`  // 订单类型:1=市价,2=限价
	Margin     float64 `json:"margin" binding:"required"`      // 初始保证金(USDT)
	Leverage   int     `json:"leverage" binding:"required"`    // 杠杆倍数
	LimitPrice float64 `json:"limit_price"`                    // 限价价格(限价单时必填)
}

// OpenPositionResponse 开仓响应
type OpenPositionResponse struct {
	OrderID      uint    `json:"order_id"`
	Status       int8    `json:"status"`
	OrderType    int8    `json:"order_type"`
	Type         int8    `json:"type"`
	Margin       float64 `json:"margin"`        // 实际保证金(扣除手续费后)
	Fee          float64 `json:"fee"`           // 手续费
	EntryPrice   float64 `json:"entry_price"`   // 开仓价格
	Leverage     int     `json:"leverage"`      // 杠杆倍数
	PositionSize float64 `json:"position_size"` // 仓位价值
	CreateTime   int64   `json:"create_time"`
}

// ClosePositionRequest 平仓请求
type ClosePositionRequest struct {
	PositionID uint `json:"position_id" binding:"required"` // 仓位ID
}

// ClosePositionResponse 平仓响应
type ClosePositionResponse struct {
	PositionID  uint    `json:"position_id"`
	ClosePrice  float64 `json:"close_price"`  // 平仓价格
	PnL         float64 `json:"pnl"`          // 盈亏
	ReturnValue float64 `json:"return_value"` // 返还金额
	CloseType   int8    `json:"close_type"`   // 平仓类型
	CloseTime   int64   `json:"close_time"`
}

// ContractChaseOrderRequest 合约追单请求(限价单转市价)
type ContractChaseOrderRequest struct {
	OrderID uint `json:"order_id" binding:"required"` // 订单ID
}

// SetTPSLRequest 设置止盈止损请求
type SetTPSLRequest struct {
	PositionID       uint     `json:"position_id" binding:"required"` // 仓位ID
	TakeProfitAmount *float64 `json:"take_profit_amount"`             // 止盈金额(USDT)
	StopLossAmount   *float64 `json:"stop_loss_amount"`               // 止损金额(USDT)
}

// PositionListRequest 持仓列表请求
type PositionListRequest struct {
	CurrencyID uint `json:"currency_id"`
	LegalID    uint `json:"legal_id"`
	Status     int8 `json:"status"` // -1=全部, 0=挂单中, 1=持仓中
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
}

// PositionDetailResponse 持仓详情响应
type PositionDetailResponse struct {
	ID               uint     `json:"id"`
	UserID           uint     `json:"user_id"`
	CurrencyID       uint     `json:"currency_id"`
	LegalID          uint     `json:"legal_id"`
	Type             int8     `json:"type"`               // 1=做多,2=做空
	OrderType        int8     `json:"order_type"`         // 1=市价,2=限价
	Status           int8     `json:"status"`             // 0=挂单,1=持仓,2=平仓中,3=已平仓
	EntryPrice       float64  `json:"entry_price"`        // 开仓价格
	CurrentPrice     float64  `json:"current_price"`      // 当前价格
	LimitPrice       *float64 `json:"limit_price"`        // 限价价格
	Margin           float64  `json:"margin"`             // 实际保证金
	OriginMargin     float64  `json:"origin_margin"`      // 原始保证金
	Fee              float64  `json:"fee"`                // 手续费
	Leverage         int      `json:"leverage"`           // 杠杆倍数
	PositionSize     float64  `json:"position_size"`      // 仓位价值
	UnrealizedPnL    float64  `json:"unrealized_pnl"`     // 未实现盈亏
	PnLRate          float64  `json:"pnl_rate"`           // 收益率
	LiquidationPrice float64  `json:"liquidation_price"`  // 爆仓价格
	TakeProfitAmount *float64 `json:"take_profit_amount"` // 止盈金额
	StopLossAmount   *float64 `json:"stop_loss_amount"`   // 止损金额
	CreateTime       int64    `json:"create_time"`
}

// ====================================
// 核心交易方法
// ====================================

// OpenPosition 开仓(市价单/限价单)
func (s *FuturesTradingService) OpenPosition(userID uint, req *OpenPositionRequest) (*OpenPositionResponse, error) {
	// 1. 参数校验
	if req.Margin < MinMargin {
		return nil, fmt.Errorf("最小开仓保证金为 %.2f USDT", MinMargin)
	}
	if req.Leverage <= 0 {
		return nil, errors.New("杠杆倍数必须大于0")
	}
	if req.Type != 1 && req.Type != 2 {
		return nil, errors.New("交易方向无效(1=做多,2=做空)")
	}
	if req.OrderType != model.ContractOrderTypeMarket && req.OrderType != model.ContractOrderTypeLimit {
		return nil, errors.New("订单类型无效(1=市价,2=限价)")
	}
	if req.OrderType == model.ContractOrderTypeLimit && req.LimitPrice <= 0 {
		return nil, errors.New("限价单必须设置限价价格")
	}

	// 2. 获取当前市场价格
	currentPrice, err := s.getCurrentMarketPrice(req.CurrencyID, req.LegalID)
	if err != nil {
		return nil, fmt.Errorf("获取市场价格失败: %w", err)
	}

	// 3. 计算手续费: 手续费 = 初始保证金 × 杠杆倍数 × 0.0031%
	fee := req.Margin * float64(req.Leverage) * ContractFeeRate
	actualMargin := req.Margin - fee

	if actualMargin <= 0 {
		return nil, errors.New("保证金不足以支付手续费")
	}

	// 4. 检查用户余额
	wallet, err := GetWalletService().GetWalletByCurrency(userID, req.LegalID)
	if err != nil {
		return nil, fmt.Errorf("获取钱包失败: %w", err)
	}
	if wallet.UsdtBalance < req.Margin {
		return nil, fmt.Errorf("USDT余额不足，当前余额: %.2f", wallet.UsdtBalance)
	}

	// 5. 确定开仓价格和状态
	var entryPrice float64
	var status int8
	var transactionTime float64

	if req.OrderType == model.ContractOrderTypeMarket {
		// 市价单: 立即以当前价格成交
		entryPrice = currentPrice
		status = model.ContractStatusActive // 交易中
		transactionTime = float64(time.Now().Unix())
	} else {
		// 限价单: 挂单等待
		entryPrice = 0 // 未成交
		status = model.ContractStatusPending
		transactionTime = 0
	}

	// 6. 计算仓位价值
	positionSize := req.Margin * float64(req.Leverage)

	// 7. 事务处理
	var position model.LeverTransaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 创建仓位记录（不再更新UsersWallet，只使用UserAssets）
		now := time.Now().Unix()
		limitPrice := req.LimitPrice
		position = model.LeverTransaction{
			Type:               req.Type,
			OrderType:          req.OrderType,
			UserID:             userID,
			CurrencyID:         req.CurrencyID,
			LegalID:            req.LegalID,
			Currency:           req.CurrencyID,
			Legal:              req.LegalID,
			OriginPrice:        currentPrice,
			Price:              entryPrice,
			UpdatePrice:        currentPrice,
			OriginCautionMoney: req.Margin,
			CautionMoney:       actualMargin,
			ActualMargin:       &actualMargin,
			TradeFee:           fee,
			Multiple:           req.Leverage,
			Status:             status,
			CreateTime:         now,
			TransactionTime:    transactionTime,
			UpdateTime:         float64(now),
		}

		// 设置限价
		if req.OrderType == model.ContractOrderTypeLimit {
			position.LimitPrice = &limitPrice
		}

		if err := tx.Create(&position).Error; err != nil {
			return fmt.Errorf("创建仓位失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 同步更新 UserAssets 表 - 扣除合约钱包USDT可用余额
	if err := GetUserAssetsService().UpdateContractUsdtBalance(userID, -req.Margin, false); err != nil {
		logger.Errorf("[OpenPosition] 同步更新合约钱包UserAssets失败: userID=%d, margin=%.8f, err=%v",
			userID, req.Margin, err)
		// 不影响主流程，仅记录错误
	}

	logger.Infof("[合约开仓] userID=%d, positionID=%d, type=%d, orderType=%d, margin=%.4f, fee=%.4f, leverage=%d",
		userID, position.ID, req.Type, req.OrderType, req.Margin, fee, req.Leverage)

	return &OpenPositionResponse{
		OrderID:      position.ID,
		Status:       status,
		OrderType:    req.OrderType,
		Type:         req.Type,
		Margin:       actualMargin,
		Fee:          fee,
		EntryPrice:   entryPrice,
		Leverage:     req.Leverage,
		PositionSize: positionSize,
		CreateTime:   position.CreateTime,
	}, nil
}

// ClosePosition 手动平仓
func (s *FuturesTradingService) ClosePosition(userID uint, req *ClosePositionRequest) (*ClosePositionResponse, error) {
	// 1. 获取仓位
	var position model.LeverTransaction
	if err := database.DB.Where("id = ? AND user_id = ?", req.PositionID, userID).First(&position).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("仓位不存在")
		}
		return nil, err
	}

	// 2. 检查仓位状态
	if position.Status != model.ContractStatusActive {
		return nil, errors.New("仓位状态不允许平仓")
	}

	// 3. 检查用户风控设置
	riskControl, err := s.getUserRiskControl(userID)
	if err == nil && riskControl != nil {
		// 强制爆仓模式: 任何平仓都触发爆仓
		if riskControl.ForceLiquidationEnabled == 1 {
			return s.executeLiquidation(&position, model.CloseTypeLiquidation)
		}
	}

	// 4. 获取当前价格
	currentPrice, err := s.getCurrentMarketPrice(position.CurrencyID, position.LegalID)
	if err != nil {
		return nil, fmt.Errorf("获取市场价格失败: %w", err)
	}

	// 5. 计算盈亏
	pnl := s.calculatePnL(position, currentPrice)

	// 6. 检查风控亏损模式
	if riskControl != nil && riskControl.ForceLossEnabled == 1 {
		// 亏损模式: 强制变为亏损
		if pnl > 0 {
			// 将盈利转为亏损
			pnl = -pnl
		}
	}

	// 7. 计算返还金额
	returnValue := position.CautionMoney + pnl
	if returnValue < 0 {
		returnValue = 0
	}

	// 8. 执行平仓
	closeType := model.CloseTypeManual
	resp, err := s.executeClose(&position, currentPrice, pnl, returnValue, closeType)
	if err != nil {
		return nil, err
	}

	logger.Infof("[合约平仓] userID=%d, positionID=%d, closePrice=%.4f, pnl=%.4f, return=%.4f",
		userID, position.ID, currentPrice, pnl, returnValue)

	return resp, nil
}

// ChaseOrder 追单(限价单转市价立即成交)
func (s *FuturesTradingService) ChaseOrder(userID uint, req *ContractChaseOrderRequest) (*OpenPositionResponse, error) {
	// 1. 获取订单
	var position model.LeverTransaction
	if err := database.DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&position).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	// 2. 检查订单状态(只有挂单中的限价单才能追单)
	if position.Status != model.ContractStatusPending {
		return nil, errors.New("只有挂单中的订单才能追单")
	}
	if position.OrderType != model.ContractOrderTypeLimit {
		return nil, errors.New("只有限价单才能追单")
	}

	// 3. 获取当前市场价格
	currentPrice, err := s.getCurrentMarketPrice(position.CurrencyID, position.LegalID)
	if err != nil {
		return nil, fmt.Errorf("获取市场价格失败: %w", err)
	}

	// 4. 更新订单为市价单并成交
	now := time.Now()
	err = database.DB.Model(&position).Updates(map[string]interface{}{
		"order_type":       model.ContractOrderTypeMarket,
		"price":            currentPrice,
		"status":           model.ContractStatusActive,
		"transaction_time": float64(now.Unix()),
		"update_time":      float64(now.Unix()),
	}).Error

	if err != nil {
		return nil, fmt.Errorf("追单失败: %w", err)
	}

	logger.Infof("[合约追单] userID=%d, orderID=%d, price=%.4f", userID, position.ID, currentPrice)

	return &OpenPositionResponse{
		OrderID:      position.ID,
		Status:       model.ContractStatusActive,
		OrderType:    model.ContractOrderTypeMarket,
		Type:         position.Type,
		Margin:       position.CautionMoney,
		Fee:          position.TradeFee,
		EntryPrice:   currentPrice,
		Leverage:     position.Multiple,
		PositionSize: position.OriginCautionMoney * float64(position.Multiple),
		CreateTime:   position.CreateTime,
	}, nil
}

// SetTakeProfit 设置止盈止损
func (s *FuturesTradingService) SetTPSL(userID uint, req *SetTPSLRequest) error {
	// 1. 获取仓位
	var position model.LeverTransaction
	if err := database.DB.Where("id = ? AND user_id = ?", req.PositionID, userID).First(&position).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("仓位不存在")
		}
		return err
	}

	// 2. 检查仓位状态(只有持仓中才能设置止盈止损)
	if position.Status != model.ContractStatusActive {
		return errors.New("只有持仓中的仓位才能设置止盈止损")
	}

	// 3. 更新止盈止损
	updates := make(map[string]interface{})
	if req.TakeProfitAmount != nil {
		updates["take_profit_amount"] = *req.TakeProfitAmount
	}
	if req.StopLossAmount != nil {
		updates["stop_loss_amount"] = *req.StopLossAmount
	}
	updates["update_time"] = float64(time.Now().Unix())

	if err := database.DB.Model(&position).Updates(updates).Error; err != nil {
		return fmt.Errorf("设置止盈止损失败: %w", err)
	}

	logger.Infof("[合约止盈止损] userID=%d, positionID=%d, tp=%v, sl=%v",
		userID, position.ID, req.TakeProfitAmount, req.StopLossAmount)

	return nil
}

// CancelOrder 撤销限价单
func (s *FuturesTradingService) CancelOrder(userID uint, orderID uint) error {
	// 1. 获取订单
	var position model.LeverTransaction
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&position).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}

	// 2. 检查订单状态
	if position.Status != model.ContractStatusPending {
		return errors.New("只有挂单中的订单才能撤销")
	}

	// 3. 事务处理: 退还保证金并更新状态
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 更新订单状态为已撤单（不再更新UsersWallet，只使用UserAssets）
		if err := tx.Model(&position).Updates(map[string]interface{}{
			"status":        model.ContractStatusCancel,
			"complete_time": float64(time.Now().Unix()),
		}).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	logger.Infof("[合约撤单] userID=%d, orderID=%d, returnMargin=%.4f", userID, orderID, position.OriginCautionMoney)
	return nil
}

// GetPositionList 获取持仓列表
func (s *FuturesTradingService) GetPositionList(userID uint, req *PositionListRequest) ([]PositionDetailResponse, int64, error) {
	var positions []model.LeverTransaction
	var total int64

	query := database.DB.Model(&model.LeverTransaction{}).Where("user_id = ?", userID)

	// 过滤条件
	if req.CurrencyID > 0 {
		query = query.Where("currency = ?", req.CurrencyID)
	}
	if req.LegalID > 0 {
		query = query.Where("legal = ?", req.LegalID)
	}
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}

	query.Count(&total)

	// 分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&positions).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应结构
	var result []PositionDetailResponse
	for _, pos := range positions {
		result = append(result, s.toPositionDetail(pos))
	}

	return result, total, nil
}

// GetPendingOrders 获取挂单列表(限价单未成交)
func (s *FuturesTradingService) GetPendingOrders(userID uint, currencyID, legalID uint) ([]PositionDetailResponse, error) {
	var positions []model.LeverTransaction

	query := database.DB.Where("user_id = ? AND status = ?", userID, model.ContractStatusPending)

	if currencyID > 0 {
		query = query.Where("currency = ?", currencyID)
	}
	if legalID > 0 {
		query = query.Where("legal = ?", legalID)
	}

	if err := query.Order("id DESC").Find(&positions).Error; err != nil {
		return nil, err
	}

	var result []PositionDetailResponse
	for _, pos := range positions {
		result = append(result, s.toPositionDetail(pos))
	}

	return result, nil
}

// ====================================
// 定时任务方法
// ====================================

// CheckAllLiquidations 检查所有仓位的爆仓条件
func (s *FuturesTradingService) CheckAllLiquidations() error {
	// 获取所有持仓中的仓位
	var positions []model.LeverTransaction
	if err := database.DB.Where("status = ?", model.ContractStatusActive).Find(&positions).Error; err != nil {
		return err
	}

	for _, position := range positions {
		// 获取当前价格
		currentPrice, err := s.getCurrentMarketPrice(position.CurrencyID, position.LegalID)
		if err != nil {
			continue
		}

		// 计算盈亏
		pnl := s.calculatePnL(position, currentPrice)

		// 检查是否达到爆仓条件: 亏损达到保证金的100%
		if pnl <= -position.CautionMoney {
			// 执行爆仓
			_, err := s.executeLiquidation(&position, model.CloseTypeLiquidation)
			if err != nil {
				logger.Errorf("[爆仓检测] 执行爆仓失败: positionID=%d, err=%v", position.ID, err)
			} else {
				logger.Infof("[爆仓检测] 执行爆仓成功: positionID=%d, userID=%d", position.ID, position.UserID)
			}
		}
	}

	return nil
}

// CheckAllTPSL 检查所有仓位的止盈止损条件
func (s *FuturesTradingService) CheckAllTPSL() error {
	// 获取所有设置了止盈止损的持仓中仓位
	var positions []model.LeverTransaction
	if err := database.DB.Where("status = ? AND (take_profit_amount IS NOT NULL OR stop_loss_amount IS NOT NULL)",
		model.ContractStatusActive).Find(&positions).Error; err != nil {
		return err
	}

	for _, position := range positions {
		// 获取当前价格
		currentPrice, err := s.getCurrentMarketPrice(position.CurrencyID, position.LegalID)
		if err != nil {
			continue
		}

		// 计算盈亏
		pnl := s.calculatePnL(position, currentPrice)

		// 检查止盈
		if position.TakeProfitAmount != nil && pnl >= *position.TakeProfitAmount {
			returnValue := position.CautionMoney + pnl
			_, err := s.executeClose(&position, currentPrice, pnl, returnValue, model.CloseTypeTakeProfit)
			if err != nil {
				logger.Errorf("[止盈检测] 执行止盈失败: positionID=%d, err=%v", position.ID, err)
			} else {
				logger.Infof("[止盈检测] 执行止盈成功: positionID=%d, pnl=%.4f", position.ID, pnl)
			}
			continue
		}

		// 检查止损
		if position.StopLossAmount != nil && pnl <= -*position.StopLossAmount {
			returnValue := position.CautionMoney + pnl
			if returnValue < 0 {
				returnValue = 0
			}
			_, err := s.executeClose(&position, currentPrice, pnl, returnValue, model.CloseTypeStopLoss)
			if err != nil {
				logger.Errorf("[止损检测] 执行止损失败: positionID=%d, err=%v", position.ID, err)
			} else {
				logger.Infof("[止损检测] 执行止损成功: positionID=%d, pnl=%.4f", position.ID, pnl)
			}
		}
	}

	return nil
}

// CheckAllLimitOrders 检查所有限价单是否可以成交
func (s *FuturesTradingService) CheckAllLimitOrders() error {
	// 获取所有挂单中的限价单
	var orders []model.LeverTransaction
	if err := database.DB.Where("status = ? AND order_type = ?",
		model.ContractStatusPending, model.ContractOrderTypeLimit).Find(&orders).Error; err != nil {
		return err
	}

	for _, order := range orders {
		if order.LimitPrice == nil {
			continue
		}

		// 获取当前价格
		currentPrice, err := s.getCurrentMarketPrice(order.CurrencyID, order.LegalID)
		if err != nil {
			continue
		}

		// 检查是否达到限价条件
		shouldExecute := false
		if order.Type == 1 { // 做多
			// 做多: 当前价格 <= 限价价格
			shouldExecute = currentPrice <= *order.LimitPrice
		} else { // 做空
			// 做空: 当前价格 >= 限价价格
			shouldExecute = currentPrice >= *order.LimitPrice
		}

		if shouldExecute {
			// 执行成交
			now := time.Now()
			err := database.DB.Model(&order).Updates(map[string]interface{}{
				"price":            currentPrice,
				"status":           model.ContractStatusActive,
				"transaction_time": float64(now.Unix()),
				"update_time":      float64(now.Unix()),
			}).Error

			if err != nil {
				logger.Errorf("[限价单检测] 执行成交失败: orderID=%d, err=%v", order.ID, err)
			} else {
				logger.Infof("[限价单检测] 执行成交成功: orderID=%d, price=%.4f", order.ID, currentPrice)
				// TODO: WebSocket推送成交通知
			}
		}
	}

	return nil
}

// ====================================
// 辅助方法
// ====================================

// calculatePnL 计算盈亏
func (s *FuturesTradingService) calculatePnL(position model.LeverTransaction, currentPrice float64) float64 {
	if position.Price == 0 {
		return 0
	}

	// 仓位数量 = 保证金 × 杠杆 / 开仓价格
	quantity := position.CautionMoney * float64(position.Multiple) / position.Price

	if position.Type == 1 { // 做多
		return (currentPrice - position.Price) * quantity
	} else { // 做空
		return (position.Price - currentPrice) * quantity
	}
}

// calculateLiquidationPrice 计算爆仓价格
func (s *FuturesTradingService) calculateLiquidationPrice(position model.LeverTransaction) float64 {
	if position.Price == 0 || position.Multiple == 0 {
		return 0
	}

	// 爆仓价格 = 开仓价格 × (1 ± 1/杠杆)
	margin := 1.0 / float64(position.Multiple)

	if position.Type == 1 { // 做多
		// 做多爆仓价格 = 开仓价格 × (1 - 1/杠杆)
		return position.Price * (1 - margin)
	} else { // 做空
		// 做空爆仓价格 = 开仓价格 × (1 + 1/杠杆)
		return position.Price * (1 + margin)
	}
}

// executeClose 执行平仓
func (s *FuturesTradingService) executeClose(position *model.LeverTransaction, closePrice, pnl, returnValue float64, closeType int8) (*ClosePositionResponse, error) {
	now := time.Now()

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 更新仓位状态
		if err := tx.Model(position).Updates(map[string]interface{}{
			"status":        model.ContractStatusClosed,
			"close_type":    closeType,
			"close_price":   closePrice,
			"fact_profits":  pnl,
			"handle_time":   float64(now.Unix()),
			"complete_time": float64(now.Unix()),
			"settled":       1,
		}).Error; err != nil {
			return err
		}

		// 不再更新UsersWallet，只使用UserAssets
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 更新 UserAssets 表 - 返还到合约钱包
	if returnValue > 0 {
		if err := GetUserAssetsService().UpdateContractUsdtBalance(position.UserID, returnValue, false); err != nil {
			logger.Errorf("[executeClose] 更新合约钱包UserAssets失败: userID=%d, returnValue=%.8f, err=%v",
				position.UserID, returnValue, err)
			// 不影响主流程，仅记录错误
		}
	}

	return &ClosePositionResponse{
		PositionID:  position.ID,
		ClosePrice:  closePrice,
		PnL:         pnl,
		ReturnValue: returnValue,
		CloseType:   closeType,
		CloseTime:   now.Unix(),
	}, nil
}

// executeLiquidation 执行爆仓
func (s *FuturesTradingService) executeLiquidation(position *model.LeverTransaction, closeType int8) (*ClosePositionResponse, error) {
	// 爆仓: 保证金全部清零
	currentPrice, _ := s.getCurrentMarketPrice(position.CurrencyID, position.LegalID)

	return s.executeClose(position, currentPrice, -position.CautionMoney, 0, closeType)
}

// getCurrentMarketPrice 获取当前市场价格
func (s *FuturesTradingService) getCurrentMarketPrice(currencyID, legalID uint) (float64, error) {
	quotation, err := GetMarketService().GetCurrencyQuotation(currencyID, legalID)
	if err != nil {
		return 0, err
	}
	return quotation.Price, nil
}

// getUserRiskControl 获取用户风控配置
func (s *FuturesTradingService) getUserRiskControl(userID uint) (*model.UserRiskControl, error) {
	var riskControl model.UserRiskControl
	err := database.DB.Where("user_id = ?", userID).First(&riskControl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &riskControl, nil
}

// toPositionDetail 转换为持仓详情
func (s *FuturesTradingService) toPositionDetail(pos model.LeverTransaction) PositionDetailResponse {
	// 获取当前价格
	currentPrice, _ := s.getCurrentMarketPrice(pos.CurrencyID, pos.LegalID)
	if currentPrice == 0 {
		currentPrice = pos.UpdatePrice
	}

	// 计算盈亏
	pnl := s.calculatePnL(pos, currentPrice)
	pnlRate := 0.0
	if pos.CautionMoney > 0 {
		pnlRate = pnl / pos.CautionMoney * 100
	}

	return PositionDetailResponse{
		ID:               pos.ID,
		UserID:           pos.UserID,
		CurrencyID:       pos.CurrencyID,
		LegalID:          pos.LegalID,
		Type:             pos.Type,
		OrderType:        pos.OrderType,
		Status:           pos.Status,
		EntryPrice:       pos.Price,
		CurrentPrice:     currentPrice,
		LimitPrice:       pos.LimitPrice,
		Margin:           pos.CautionMoney,
		OriginMargin:     pos.OriginCautionMoney,
		Fee:              pos.TradeFee,
		Leverage:         pos.Multiple,
		PositionSize:     pos.OriginCautionMoney * float64(pos.Multiple),
		UnrealizedPnL:    pnl,
		PnLRate:          pnlRate,
		LiquidationPrice: s.calculateLiquidationPrice(pos),
		TakeProfitAmount: pos.TakeProfitAmount,
		StopLossAmount:   pos.StopLossAmount,
		CreateTime:       pos.CreateTime,
	}
}

// ====================================
// 兼容旧接口的方法
// ====================================

// PositionSide 持仓方向
type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

// PositionStatus 持仓状态
type PositionStatus string

const (
	PositionStatusActive  PositionStatus = "active"
	PositionStatusClosing PositionStatus = "closing"
	PositionStatusClosed  PositionStatus = "closed"
)

// GetMarketService 获取行情服务实例
func GetMarketService() *MarketService {
	if marketService == nil {
		marketService = &MarketService{}
	}
	return marketService
}

var marketService *MarketService
