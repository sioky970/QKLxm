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

// 交割合约手续费率: 0.03%
const DeliveryFeeRate = 0.0003

// 最小开仓保证金: 50 USDT
const DeliveryMinMargin = 50.0

// 交割执行延迟时间(秒)
const DeliverySettlementDelay = 24 * 60 * 60 // 24小时

// ====================================
// DeliveryTradingService 交割合约交易服务
// ====================================

type DeliveryTradingService struct{}

func NewDeliveryTradingService() *DeliveryTradingService {
	return &DeliveryTradingService{}
}

// 全局服务实例
var deliveryTradingService *DeliveryTradingService

func GetDeliveryTradingService() *DeliveryTradingService {
	if deliveryTradingService == nil {
		deliveryTradingService = NewDeliveryTradingService()
	}
	return deliveryTradingService
}

// ====================================
// 请求/响应结构体
// ====================================

// OpenDeliveryPositionRequest 开仓请求
type OpenDeliveryPositionRequest struct {
	ContractID   uint    `json:"contract_id" binding:"required"` // 合约ID
	Side         int8    `json:"side" binding:"required"`       // 1=做多, 2=做空
	OrderType    int8    `json:"order_type" binding:"required"`  // 1=市价, 2=限价
	Price        float64 `json:"price"`                        // 限价价格(限价单时必填)
	Size         float64 `json:"size" binding:"required"`      // 合约数量
	Leverage     int     `json:"leverage" binding:"required"`   // 杠杆倍数
	StopLoss     *float64 `json:"stop_loss"`                  // 止损价格
	TakeProfit   *float64 `json:"take_profit"`                // 止盈价格
}

// OpenDeliveryPositionResponse 开仓响应
type OpenDeliveryPositionResponse struct {
	PositionID   uint    `json:"position_id"`
	OrderID      uint    `json:"order_id"`
	Status       int8    `json:"status"`
	Side         int8    `json:"side"`
	EntryPrice   float64 `json:"entry_price"`
	Margin       float64 `json:"margin"`
	Fee          float64 `json:"fee"`
	Size         float64 `json:"size"`
	Leverage     int     `json:"leverage"`
	StopLoss     *float64 `json:"stop_loss"`
	TakeProfit   *float64 `json:"take_profit"`
	CreateTime   int64   `json:"create_time"`
}

// CloseDeliveryPositionRequest 平仓请求
type CloseDeliveryPositionRequest struct {
	PositionID   uint   `json:"position_id" binding:"required"` // 持仓ID
	CloseType    int8   `json:"close_type" binding:"required"`   // 1=手动平仓, 2=止盈止损, 3=强平
	ClosePrice   *float64 `json:"close_price"`                  // 平仓价格(市价单为空)
}

// CloseDeliveryPositionResponse 平仓响应
type CloseDeliveryPositionResponse struct {
	PositionID   uint    `json:"position_id"`
	ClosePrice   float64 `json:"close_price"`
	PnL         float64 `json:"pnl"`
	ReturnValue  float64 `json:"return_value"`
	Fee         float64 `json:"fee"`
	CloseType   int8    `json:"close_type"`
	CloseTime   int64   `json:"close_time"`
}

// DeliveryPositionListRequest 持仓列表请求
type DeliveryPositionListRequest struct {
	ContractID  uint   `json:"contract_id"`
	Status      int8   `json:"status"` // -1=全部, 0=已平仓, 1=持仓中
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}

// DeliveryPositionDetailResponse 持仓详情响应
type DeliveryPositionDetailResponse struct {
	ID                  uint     `json:"id"`
	ContractID          uint     `json:"contract_id"`
	Symbol              string   `json:"symbol"`
	Side                int8     `json:"side"`                       // 1=做多, 2=做空
	Status              int8     `json:"status"`                     // 1=持仓, 0=已平仓
	Size                float64  `json:"size"`                       // 持仓数量
	EntryPrice          float64  `json:"entry_price"`                // 开仓价格
	CurrentPrice        float64  `json:"current_price"`              // 当前价格
	UnrealizedPnL       float64  `json:"unrealized_pnl"`           // 未实现盈亏
	PnLRate             float64  `json:"pnl_rate"`                  // 收益率
	Margin              float64  `json:"margin"`                     // 保证金
	Leverage            int      `json:"leverage"`                   // 杠杆倍数
	StopLossPrice       *float64 `json:"stop_loss_price"`           // 止损价格
	TakeProfitPrice     *float64 `json:"take_profit_price"`         // 止盈价格
	DeliveryStatus      int8     `json:"delivery_status"`           // 交割状态
	CreateTime         int64    `json:"create_time"`
	UpdateTime         int64    `json:"update_time"`
}

// DeliveryContractInfo 交割合约信息
type DeliveryContractInfo struct {
	ID          uint    `json:"id"`
	Symbol      string  `json:"symbol"`
	BaseAsset   string  `json:"base_asset"`
	QuoteAsset  string  `json:"quote_asset"`
	ContractSize float64 `json:"contract_size"`
	MaxLeverage int     `json:"max_leverage"`
	DeliveryDate time.Time `json:"delivery_date"`
	Status      int8    `json:"status"`
}

// DeliveryAccountBalance 交割合约账户余额
type DeliveryAccountBalance struct {
	UsdtBalance     float64 `json:"usdt_balance"`      // USDT余额
	UsdtLocked      float64 `json:"usdt_locked"`       // 冻结余额
	DeliveryMargin  float64 `json:"delivery_margin"`   // 交割合约保证金
	AvailableMargin float64 `json:"available_margin"`  // 可用保证金
	DeliveryPnL     float64 `json:"delivery_pnl"`    // 交割合约盈亏
}

// ====================================
// 核心交易方法
// ====================================

// OpenPosition 开仓(市价单/限价单)
func (s *DeliveryTradingService) OpenPosition(userID uint, req *OpenDeliveryPositionRequest) (*OpenDeliveryPositionResponse, error) {
	// 1. 参数校验
	if req.Size <= 0 {
		return nil, errors.New("合约数量必须大于0")
	}
	if req.Leverage <= 0 {
		return nil, errors.New("杠杆倍数必须大于0")
	}
	if req.Side != 1 && req.Side != 2 {
		return nil, errors.New("交易方向无效(1=做多, 2=做空)")
	}
	if req.OrderType != 1 && req.OrderType != 2 {
		return nil, errors.New("订单类型无效(1=市价, 2=限价)")
	}
	if req.OrderType == 2 && req.Price <= 0 {
		return nil, errors.New("限价单必须提供有效价格")
	}

	var openResponse *OpenDeliveryPositionResponse
	
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 2. 获取交割合约配置
		var contract model.DeliveryContract
		if err := tx.First(&contract, req.ContractID).Error; err != nil {
			return fmt.Errorf("交割合约不存在: %w", err)
		}
		
		if contract.Status != 1 {
			return errors.New("交割合约已禁用")
		}
		
		// 检查杠杆倍数
		if req.Leverage > contract.MaxLeverage {
			return fmt.Errorf("杠杆倍数超出最大值: %d", contract.MaxLeverage)
		}

		// 3. 获取账户余额
		accountSummary, err := GetUserAssetsService().GetDeliveryAccountSummary(userID)
		if err != nil {
			return fmt.Errorf("获取交割合约账户失败: %w", err)
		}

		// 4. 计算所需保证金
		requiredMargin := s.calculateRequiredMargin(req.Price, req.Size, req.Leverage, contract)
		
		if accountSummary.AvailableMargin < requiredMargin {
			return fmt.Errorf("保证金不足: 可用 %.8f, 需要 %.8f", 
				accountSummary.AvailableMargin, requiredMargin)
		}

		// 5. 获取当前价格
		currentPrice, err := s.getCurrentPrice(contract, req.Price, req.OrderType)
		if err != nil {
			return fmt.Errorf("获取价格失败: %w", err)
		}

		// 6. 计算手续费
		fee := requiredMargin * DeliveryFeeRate

		// 7. 创建持仓记录
		position := &model.DeliveryPosition{
			UserID:       userID,
			ContractID:   req.ContractID,
			Symbol:       contract.Symbol,
			Side:         req.Side,
			Size:         req.Size,
			EntryPrice:   currentPrice,
			Leverage:     req.Leverage,
			Margin:       requiredMargin,
			CurrentPrice: currentPrice,
			TakeProfitPrice: req.TakeProfit,
			StopLossPrice: req.StopLoss,
			Status:       1,
			DeliveryStatus: 0,
		}

		position.CreateTime = time.Now()
		position.UpdateTime = time.Now()
		
		if err := tx.Create(position).Error; err != nil {
			return fmt.Errorf("创建持仓失败: %w", err)
		}

		// 8. 冻结保证金
		if err := GetUserAssetsService().FreezeDeliveryMargin(userID, requiredMargin); err != nil {
			return fmt.Errorf("冻结保证金失败: %w", err)
		}

		// 9. 创建订单记录(可选，用于历史追踪)
		orderID, err := s.createDeliveryOrder(tx, userID, position.ID, req, currentPrice, fee)
		if err != nil {
			logger.Warnf("[Delivery] 创建订单记录失败: %v", err)
		}

		logger.Infof("[Delivery] 开仓成功: userID=%d, positionID=%d, orderID=%d, symbol=%s, side=%d, size=%.8f, price=%.8f, leverage=%d, margin=%.8f, fee=%.8f",
			userID, position.ID, orderID, position.Symbol, position.Side, position.Size, position.EntryPrice, position.Leverage, position.Margin, fee)

		// 构建响应
		openResponse = &OpenDeliveryPositionResponse{
			PositionID: position.ID,
			OrderID:    orderID,
			Status:     1,
			Side:       req.Side,
			EntryPrice: currentPrice,
			Margin:     requiredMargin,
			Fee:        fee,
			Size:       req.Size,
			Leverage:   req.Leverage,
			StopLoss:   req.StopLoss,
			TakeProfit: req.TakeProfit,
			CreateTime: position.CreateTime.Unix(),
		}

		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return openResponse, nil
}

// ClosePosition 平仓
func (s *DeliveryTradingService) ClosePosition(userID uint, req *CloseDeliveryPositionRequest) (*CloseDeliveryPositionResponse, error) {
	var closeResponse *CloseDeliveryPositionResponse
	
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 获取持仓信息
		var position model.DeliveryPosition
		if err := tx.First(&position, req.PositionID).Error; err != nil {
			return fmt.Errorf("持仓不存在: %w", err)
		}

		if position.UserID != userID {
			return errors.New("无权操作该持仓")
		}

		if position.Status != 1 {
			return errors.New("持仓已平仓")
		}

		// 2. 获取合约信息
		var contract model.DeliveryContract
		if err := tx.First(&contract, position.ContractID).Error; err != nil {
			return fmt.Errorf("获取合约信息失败: %w", err)
		}

		// 3. 获取平仓价格
		var closePrice float64
		var priceErr error
		if req.ClosePrice != nil {
			closePrice, priceErr = s.getCurrentPrice(contract, *req.ClosePrice, 1) // 限价平仓
		} else {
			closePrice, priceErr = s.getCurrentPrice(contract, 0, 1) // 市价平仓
		}
		if priceErr != nil {
			return fmt.Errorf("获取平仓价格失败: %w", priceErr)
		}

		// 4. 计算盈亏
		pnl := s.calculatePnL(&position, closePrice)
		fee := position.Margin * DeliveryFeeRate
		returnValue := position.Margin + pnl - fee

		// 5. 更新持仓状态
		position.Status = 0
		position.CurrentPrice = closePrice
		position.UnrealizedPnL = pnl
		position.UpdateTime = time.Now()

		if err := tx.Save(&position).Error; err != nil {
			return fmt.Errorf("更新持仓失败: %w", err)
		}

		// 6. 解冻保证金并更新余额
		if err := GetUserAssetsService().UnfreezeDeliveryMargin(userID, position.Margin); err != nil {
			return fmt.Errorf("解冻保证金失败: %w", err)
		}

		// 7. 更新盈亏
		if err := GetUserAssetsService().UpdateDeliveryPnL(userID, pnl-fee); err != nil {
			return fmt.Errorf("更新盈亏失败: %w", err)
		}

		// 8. 记录平仓历史
		if err := s.createDeliveryCloseHistory(tx, &position, closePrice, pnl, fee, req.CloseType); err != nil {
			logger.Warnf("[Delivery] 创建平仓历史失败: %v", err)
		}

		logger.Infof("[Delivery] 平仓成功: userID=%d, positionID=%d, symbol=%s, pnl=%.8f, return=%.8f, fee=%.8f",
			userID, position.ID, position.Symbol, pnl, returnValue, fee)

		// 构建响应
		closeResponse = &CloseDeliveryPositionResponse{
			PositionID: position.ID,
			ClosePrice: closePrice,
			PnL:       pnl,
			ReturnValue: returnValue,
			Fee:       fee,
			CloseType: req.CloseType,
			CloseTime: time.Now().Unix(),
		}

		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return closeResponse, nil
}

// GetPositions 获取持仓列表
func (s *DeliveryTradingService) GetPositions(userID uint, req *DeliveryPositionListRequest) ([]DeliveryPositionDetailResponse, error) {
	query := database.DB.Model(&model.DeliveryPosition{}).Where("user_id = ?", userID)

	// 合约筛选
	if req.ContractID > 0 {
		query = query.Where("contract_id = ?", req.ContractID)
	}

	// 状态筛选
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}

	// 分页
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PageSize

	var positions []model.DeliveryPosition
	err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&positions).Error
	if err != nil {
		return nil, err
	}

	var responses []DeliveryPositionDetailResponse
	for _, position := range positions {
		// 计算收益率
		pnlRate := 0.0
		if position.Margin > 0 {
			pnlRate = (position.UnrealizedPnL / position.Margin) * 100
		}

		response := DeliveryPositionDetailResponse{
			ID:                position.ID,
			ContractID:        position.ContractID,
			Symbol:            position.Symbol,
			Side:              position.Side,
			Status:            position.Status,
			Size:              position.Size,
			EntryPrice:        position.EntryPrice,
			CurrentPrice:      position.CurrentPrice,
			UnrealizedPnL:     position.UnrealizedPnL,
			PnLRate:           pnlRate,
			Margin:            position.Margin,
			Leverage:          position.Leverage,
			StopLossPrice:     position.StopLossPrice,
			TakeProfitPrice:   position.TakeProfitPrice,
			DeliveryStatus:    position.DeliveryStatus,
			CreateTime:        position.CreateTime.Unix(),
			UpdateTime:        position.UpdateTime.Unix(),
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetAccountBalance 获取交割合约账户余额
func (s *DeliveryTradingService) GetAccountBalance(userID uint) (*DeliveryAccountBalance, error) {
	summary, err := GetUserAssetsService().GetDeliveryAccountSummary(userID)
	if err != nil {
		return nil, err
	}

	return &DeliveryAccountBalance{
		UsdtBalance:     summary.UsdtBalance,
		UsdtLocked:      summary.UsdtLocked,
		DeliveryMargin:  summary.DeliveryMargin,
		AvailableMargin: summary.AvailableMargin,
		DeliveryPnL:     summary.DeliveryPnL,
	}, nil
}

// GetContracts 获取交割合约列表
func (s *DeliveryTradingService) GetContracts() ([]DeliveryContractInfo, error) {
	var contracts []model.DeliveryContract
	err := database.DB.Where("status = ?", 1).Order("delivery_date ASC").Find(&contracts).Error
	if err != nil {
		return nil, err
	}

	var contractInfos []DeliveryContractInfo
	for _, contract := range contracts {
		contractInfo := DeliveryContractInfo{
			ID:           contract.ID,
			Symbol:       contract.Symbol,
			BaseAsset:    contract.BaseAsset,
			QuoteAsset:   contract.QuoteAsset,
			ContractSize: contract.ContractSize,
			MaxLeverage:  contract.MaxLeverage,
			DeliveryDate: contract.DeliveryDate,
			Status:       contract.Status,
		}
		contractInfos = append(contractInfos, contractInfo)
	}

	return contractInfos, nil
}

// ====================================
// 辅助方法
// ====================================

// calculateRequiredMargin 计算所需保证金
func (s *DeliveryTradingService) calculateRequiredMargin(price, size float64, leverage int, contract model.DeliveryContract) float64 {
	// 基础保证金 = 价格 * 数量 * 合约乘数 / 杠杆倍数
	positionValue := price * size * contract.ContractSize
	margin := positionValue / float64(leverage)
	
	// 确保不低于最小保证金
	if margin < DeliveryMinMargin {
		return DeliveryMinMargin
	}
	
	return margin
}

// calculatePnL 计算盈亏
func (s *DeliveryTradingService) calculatePnL(position *model.DeliveryPosition, closePrice float64) float64 {
	var pnl float64
	if position.Side == 1 { // 做多
		pnl = (closePrice - position.EntryPrice) * position.Size * float64(position.Leverage)
	} else { // 做空
		pnl = (position.EntryPrice - closePrice) * position.Size * float64(position.Leverage)
	}
	
	// 考虑合约乘数
	// 这里假设contract乘数已经在价格计算中体现
	
	return pnl
}

// getCurrentPrice 获取当前价格
func (s *DeliveryTradingService) getCurrentPrice(contract model.DeliveryContract, fallbackPrice float64, orderType int8) (float64, error) {
	// 如果是限价单，使用指定价格
	if orderType == 2 && fallbackPrice > 0 {
		return fallbackPrice, nil
	}
	
	// 如果是市价单，需要从行情数据获取
	// 这里暂时返回fallbackPrice，实际应该从行情API获取
	if fallbackPrice > 0 {
		return fallbackPrice, nil
	}
	
	return 0, errors.New("无法获取当前价格")
}

// createDeliveryOrder 创建交割合约订单记录
func (s *DeliveryTradingService) createDeliveryOrder(tx *gorm.DB, userID uint, positionID uint, req *OpenDeliveryPositionRequest, price, fee float64) (uint, error) {
	// 这里可以创建订单记录表，暂时返回positionID作为订单ID
	// 实际项目中应该创建独立的订单表
	return positionID, nil
}

// createDeliveryCloseHistory 创建平仓历史记录
func (s *DeliveryTradingService) createDeliveryCloseHistory(tx *gorm.DB, position *model.DeliveryPosition, closePrice, pnl, fee float64, closeType int8) error {
	// 这里可以创建平仓历史记录表
	// 暂时记录到日志
	logger.Infof("[Delivery] 平仓历史: positionID=%d, closePrice=%.8f, pnl=%.8f, fee=%.8f, closeType=%d",
		position.ID, closePrice, pnl, fee, closeType)
	return nil
}