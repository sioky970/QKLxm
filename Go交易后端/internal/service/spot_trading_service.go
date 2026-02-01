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

// 全局服务实例
var (
	currencyService *CurrencyService
	walletService   *WalletService
)

// GetCurrencyService 获取币种服务实例
func GetCurrencyService() *CurrencyService {
	if currencyService == nil {
		currencyService = NewCurrencyService()
	}
	return currencyService
}

// GetWalletService 获取钱包服务实例
func GetWalletService() *WalletService {
	if walletService == nil {
		walletService = NewWalletService()
	}
	return walletService
}

// SpotTradingService 币币交易服务
type SpotTradingService struct{}

// NewSpotTradingService 创建币币交易服务实例
func NewSpotTradingService() *SpotTradingService {
	return &SpotTradingService{}
}

// OrderSide 订单方向
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"  // 买入
	OrderSideSell OrderSide = "sell" // 卖出
)

// SpotSubmitOrderRequest 提交订单请求
type SpotSubmitOrderRequest struct {
	CurrencyID   uint      `json:"currency_id" binding:"required"`
	LegalID      uint      `json:"legal_id" binding:"required"`
	Type         OrderType `json:"type" binding:"required"`          // limit 或 market
	Side         OrderSide `json:"side" binding:"required"`          // buy 或 sell
	Price        float64   `json:"price" binding:"required,gt=0"`    // 价格
	Quantity     float64   `json:"quantity" binding:"required,gt=0"` // 数量
	TriggerPrice *float64  `json:"trigger_price"`                    // 触发价格(条件单)
	TriggerType  *string   `json:"trigger_type"`                     // 触发类型(条件单)
}

// SpotSubmitOrderResponse 提交订单响应
type SpotSubmitOrderResponse struct {
	OrderID     uint        `json:"order_id"`
	OrderNo     string      `json:"order_no"`
	Status      OrderStatus `json:"status"`
	AvgPrice    float64     `json:"avg_price"`
	FilledQty   float64     `json:"filled_qty"`
	UnfilledQty float64     `json:"unfilled_qty"`
	TotalAmount float64     `json:"total_amount"`
	CreateTime  time.Time   `json:"create_time"`
}

// CancelOrderRequest 撤销订单请求
type CancelOrderRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

// ChaseOrderRequest 追单请求
type ChaseOrderRequest struct {
	OrderID  uint    `json:"order_id" binding:"required"`
	NewPrice float64 `json:"new_price" binding:"required,gt=0"`
}

// SpotOrderListRequest 订单列表请求
type SpotOrderListRequest struct {
	CurrencyID uint         `json:"currency_id"`
	LegalID    uint         `json:"legal_id"`
	Side       *OrderSide   `json:"side"`
	Status     *OrderStatus `json:"status"`
	Page       int          `json:"page" default:"1"`
	PageSize   int          `json:"page_size" default:"20"`
}

// SpotOrderDetailResponse 订单详情响应
type SpotOrderDetailResponse struct {
	ID         uint    `json:"id"`
	OrderNo    string  `json:"order_no"`
	UserID     uint    `json:"user_id"`
	CurrencyID uint    `json:"currency_id"`
	LegalID    uint    `json:"legal_id"`
	Symbol     string  `json:"symbol"`      // 交易对名称，如 "BTC/USDT"
	Type       string  `json:"type"`        // "limit" 或 "market"
	Side       string  `json:"side"`        // "buy" 或 "sell"
	Price      float64 `json:"price"`       // 订单价格
	Number     float64 `json:"number"`      // 订单数量（前端使用这个字段名）
	DealNumber float64 `json:"deal_number"` // 成交数量
	OrderValue float64 `json:"order_value"` // 订单价值(USDT)
	Status     int     `json:"status"`      // 0=未成交, 1=部分成交, 2=已成交, 3=已撤销
	Time       int64   `json:"time"`        // 创建时间戳（Unix秒）
	CreateTime int64   `json:"create_time"` // 创建时间戳（兼容字段）
}

// TradeHistoryResponse 交易历史响应
type TradeHistoryResponse struct {
	ID        uint      `json:"id"`
	OrderID   uint      `json:"order_id"`
	TradeType OrderSide `json:"trade_type"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Fee       float64   `json:"fee"`
	TradeTime time.Time `json:"trade_time"`
}

// SubmitOrder 提交订单
func (s *SpotTradingService) SubmitOrder(userID uint, req *SpotSubmitOrderRequest) (*SpotSubmitOrderResponse, error) {
	// 验证币种是否存在和启用（取消交易对验证）
	_, err := GetCurrencyService().GetCurrencyByID(req.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("币种不存在: %w", err)
	}
	
	_, err = GetCurrencyService().GetCurrencyByID(req.LegalID)
	if err != nil {
		return nil, fmt.Errorf("计价币种不存在: %w", err)
	}

	// 从 UserAssets 新表读取余额进行校验（保持与前端显示一致）
	userAssets, err := GetUserAssetsService().GetOrCreateUserAssets(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户资产失败: %w", err)
	}

	// 检查余额（使用 UserAssets 新表的数据）
	requiredAmount := req.Price * req.Quantity

	// 计算订单价值并校验余额
	var orderValue float64
	if req.Side == OrderSideSell {
		// 卖出：需要检查币种余额
		currency, _ := GetCurrencyService().GetCurrencyByID(req.CurrencyID)
		currencyBalance := userAssets.GetCurrencyBalance(currency.Name)
		orderValue = req.Quantity * req.Price
		if currencyBalance < req.Quantity {
			return nil, fmt.Errorf("%s余额不足，可用: %.8f, 需要: %.8f", currency.Name, currencyBalance, req.Quantity)
		}
	} else if req.Side == OrderSideBuy {
		// 买入：需要检查USDT余额
		orderValue = requiredAmount
		if userAssets.UsdtBalance < requiredAmount {
			return nil, fmt.Errorf("USDT余额不足，可用: %.8f, 需要: %.8f", userAssets.UsdtBalance, requiredAmount)
		}
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建订单
	order := &model.Transaction{
		FromUserID: userID,
		ToUserID:   0, // 待撮合时填充
		Currency:   int(req.CurrencyID),
		LegalID:    req.LegalID,
		Type:       s.orderSideToInt(req.Side),
		Price:      req.Price,
		Number:     req.Quantity,
		DealNumber: 0,          // 初始成交数量为0
		OrderValue: orderValue, // 订单价值
		Status:     0,          // 0=待成交
		Time:       time.Now().Unix(),
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 锁定相应资金（统一使用 UserAssets 表）
	lockedAmount := requiredAmount
	if req.Side == OrderSideSell {
		lockedAmount = req.Quantity
	}

	// 使用 UserAssets 锁定资金
	if req.Side == OrderSideBuy {
		// 买入：锁定USDT
		if err := GetUserAssetsService().UpdateUsdtBalance(userID, -lockedAmount, false); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("扣除USDT余额失败: %w", err)
		}
		if err := GetUserAssetsService().UpdateUsdtBalance(userID, lockedAmount, true); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("锁定USDT失败: %w", err)
		}
	} else {
		// 卖出：锁定币种
		currency, _ := GetCurrencyService().GetCurrencyByID(req.CurrencyID)
		if err := GetUserAssetsService().UpdateCurrencyBalance(userID, currency.Name, -lockedAmount, false); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("扣除币种余额失败: %w", err)
		}
		if err := GetUserAssetsService().UpdateCurrencyBalance(userID, currency.Name, lockedAmount, true); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("锁定币种失败: %w", err)
		}
	}

	tx.Commit()

	// 记录日志
	logger.Infof("用户提交订单成功: userID=%d, orderID=%d, type=%s, side=%s, price=%.8f, quantity=%.8f",
		userID, order.ID, req.Type, req.Side, req.Price, req.Quantity)

	// WebSocket实时推送余额变更
	go GetWalletService().BroadcastBalanceUpdate(userID)

	// 【事件驱动撮合】订单提交后立即尝试与对手盘撮合
	go func() {
		result := GetMatchingEngine().TryMatch(order)
		if result.Matched {
			logger.Infof("订单%d撮合成功: 成交数量=%.8f, 成交价格=%.8f, 完全成交=%v",
				order.ID, result.FilledQty, result.FilledPrice, result.FullFilled)
		}
	}()

	return &SpotSubmitOrderResponse{
		OrderID:     order.ID,
		OrderNo:     fmt.Sprintf("ORDER_%d", order.ID),
		Status:      OrderStatusPending,
		AvgPrice:    0,
		FilledQty:   0,
		UnfilledQty: req.Quantity,
		TotalAmount: requiredAmount,
		CreateTime:  time.Now(),
	}, nil
}

// CancelOrder 撤销订单
func (s *SpotTradingService) CancelOrder(userID uint, req *CancelOrderRequest) error {
	// 获取订单
	var order model.Transaction
	result := database.DB.First(&order, req.OrderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return result.Error
	}

	// 检查订单归属
	if order.FromUserID != userID {
		return errors.New("无权操作此订单")
	}

	// 检查订单状态
	if order.Status == 2 { // 2=已完成
		return errors.New("订单已完成，无法撤销")
	}
	if order.Status == 3 { // 3=已撤销
		return errors.New("订单已撤销")
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	if err := tx.Model(&order).Update("status", 3).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 释放锁定的资金（统一使用 UserAssets 表）
	var remainingAmount float64
	if order.Type == 1 { // 1=买入
		remainingAmount = (order.Number - order.DealNumber) * order.Price
	} else { // 2=卖出
		remainingAmount = order.Number - order.DealNumber
	}

	// 从锁定余额转回可用余额
	if order.Type == 1 { // 买入：解锁USDT
		if err := GetUserAssetsService().UpdateUsdtBalance(userID, -remainingAmount, true); err != nil {
			tx.Rollback()
			return fmt.Errorf("扣除锁定USDT失败: %w", err)
		}
		if err := GetUserAssetsService().UpdateUsdtBalance(userID, remainingAmount, false); err != nil {
			tx.Rollback()
			return fmt.Errorf("返还USDT余额失败: %w", err)
		}
	} else { // 卖出：解锁币种
		currency, _ := GetCurrencyService().GetCurrencyByID(uint(order.Currency))
		if err := GetUserAssetsService().UpdateCurrencyBalance(userID, currency.Name, -remainingAmount, true); err != nil {
			tx.Rollback()
			return fmt.Errorf("扣除锁定币种失败: %w", err)
		}
		if err := GetUserAssetsService().UpdateCurrencyBalance(userID, currency.Name, remainingAmount, false); err != nil {
			tx.Rollback()
			return fmt.Errorf("返还币种余额失败: %w", err)
		}
	}

	tx.Commit()

	logger.Infof("订单撤销成功: userID=%d, orderID=%d", userID, req.OrderID)

	// WebSocket实时推送余额变更
	go GetWalletService().BroadcastBalanceUpdate(userID)

	return nil
}

// GetOrderList 获取订单列表
func (s *SpotTradingService) GetOrderList(userID uint, req *SpotOrderListRequest) ([]SpotOrderDetailResponse, int64, error) {
	logger.Infof("[GetOrderList] 开始查询: userID=%d, status=%v, page=%d, pageSize=%d", userID, req.Status, req.Page, req.PageSize)

	var orders []model.Transaction
	var total int64

	// 使用 from_user_id 或 to_user_id 查询用户的订单
	query := database.DB.Model(&model.Transaction{}).Where("from_user_id = ? OR to_user_id = ?", userID, userID)

	if req.CurrencyID > 0 {
		query = query.Where("currency = ?", req.CurrencyID)
	}
	if req.LegalID > 0 {
		query = query.Where("legal = ?", req.LegalID)
	}
	if req.Side != nil {
		side := s.orderSideToInt(*req.Side)
		query = query.Where("type = ?", side)
	}
	if req.Status != nil {
		status := s.orderStatusToInt(*req.Status)
		query = query.Where("status = ?", status)
		logger.Infof("[GetOrderList] 过滤状态: status=%d", status)
	}

	query.Count(&total)
	logger.Infof("[GetOrderList] 总订单数: %d", total)

	offset := (req.Page - 1) * req.PageSize
	result := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&orders)
	if result.Error != nil {
		logger.Errorf("[GetOrderList] 查询失败: %v", result.Error)
		return nil, 0, result.Error
	}

	logger.Infof("[GetOrderList] 查询到订单数: %d", len(orders))

	var orderList []SpotOrderDetailResponse
	for i, order := range orders {
		logger.Infof("[GetOrderList] 处理订单[%d]: ID=%d", i, order.ID)
		orderDetail := s.modelToOrderDetail(order)
		orderList = append(orderList, orderDetail)
	}

	logger.Infof("[GetOrderList] 返回订单列表: %d条", len(orderList))
	return orderList, total, nil
}

// GetOrderDetail 获取订单详情
func (s *SpotTradingService) GetOrderDetail(userID uint, orderID uint) (*SpotOrderDetailResponse, error) {
	var order model.Transaction
	result := database.DB.Where("user_id = ? AND id = ?", userID, orderID).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, result.Error
	}

	detail := s.modelToOrderDetail(order)
	return &detail, nil
}

// GetTradeHistory 获取交易历史
func (s *SpotTradingService) GetTradeHistory(userID uint, currencyID, legalID uint, page, pageSize int) ([]TradeHistoryResponse, int64, error) {
	var trades []model.TransactionComplete
	var total int64

	query := database.DB.Model(&model.TransactionComplete{}).Where("user_id = ? OR from_user_id = ?", userID, userID)

	if currencyID > 0 {
		query = query.Where("currency = ?", currencyID)
	}
	if legalID > 0 {
		query = query.Where("legal = ?", legalID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	result := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&trades)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	var tradeList []TradeHistoryResponse
	for _, trade := range trades {
		tradeList = append(tradeList, TradeHistoryResponse{
			ID:        trade.ID,
			OrderID:   trade.ID, // 使用成交记录ID
			TradeType: s.intToOrderSide(int(trade.Way)),
			Price:     trade.Price,
			Quantity:  trade.Number,
			Fee:       0, // 模型中无Fee字段
			TradeTime: time.Unix(trade.CreateTime, 0),
		})
	}

	return tradeList, total, nil
}

// NOTE: MatchOrders 和 updateBalancesAfterTrade 已删除
// 当前系统使用做市商模式（Market Maker Mode），用户直接与平台兑换
// 不需要订单簿撮合逻辑，价格监控由 MatchingEngine 负责

// orderSideToInt 订单方向转整型
func (s *SpotTradingService) orderSideToInt(side OrderSide) int8 {
	switch side {
	case OrderSideBuy:
		return 1
	case OrderSideSell:
		return 2
	default:
		return 0
	}
}

// intToOrderSide 整型转订单方向
func (s *SpotTradingService) intToOrderSide(i int) OrderSide {
	switch i {
	case 1:
		return OrderSideBuy
	case 2:
		return OrderSideSell
	default:
		return OrderSideBuy
	}
}

// orderStatusToInt 订单状态转整型
func (s *SpotTradingService) orderStatusToInt(status OrderStatus) int8 {
	switch status {
	case OrderStatusPending:
		return 0
	case OrderStatusPartial:
		return 1
	case OrderStatusCompleted:
		return 2
	case OrderStatusCancelled:
		return 3
	default:
		return 0
	}
}

// modelToOrderDetail 模型转订单详情
func (s *SpotTradingService) modelToOrderDetail(order model.Transaction) SpotOrderDetailResponse {
	// 调试日志
	logger.Infof("[modelToOrderDetail] Order ID=%d, Currency=%d", order.ID, order.Currency)

	// 获取币种名称
	var currencyName string = "Unknown"
	currencyID := uint(order.Currency)

	if currencyID > 0 {
		currency, err := GetCurrencyService().GetCurrencyByID(currencyID)
		if err != nil {
			logger.Errorf("[modelToOrderDetail] GetCurrencyByID(%d) error: %v", currencyID, err)
		} else if currency == nil {
			logger.Warnf("[modelToOrderDetail] GetCurrencyByID(%d) returned nil", currencyID)
		} else {
			currencyName = currency.Name
			logger.Infof("[modelToOrderDetail] Found currency: ID=%d, Name=%s", currencyID, currencyName)
		}
	} else {
		logger.Warnf("[modelToOrderDetail] currency is 0, cannot get currency name")
	}

	// 构造交易对名称
	symbol := fmt.Sprintf("%s/USDT", currencyName)

	// 判断订单类型（市价/限价）
	orderType := "limit"
	if order.Price == 0 {
		orderType = "market"
	}

	// 判断交易方向
	side := "buy"
	if order.Type == 2 {
		side = "sell"
	}

	return SpotOrderDetailResponse{
		ID:         order.ID,
		OrderNo:    fmt.Sprintf("ORDER_%d", order.ID),
		UserID:     order.FromUserID,
		CurrencyID: currencyID,
		LegalID:    order.LegalID,
		Symbol:     symbol,
		Type:       orderType,
		Side:       side,
		Price:      order.Price,
		Number:     order.Number,
		DealNumber: order.DealNumber,
		OrderValue: order.OrderValue,
		Status:     int(order.Status),
		Time:       order.Time,
		CreateTime: order.Time,
	}
}

// intToOrderStatus 整型转订单状态
func (s *SpotTradingService) intToOrderStatus(i int8) OrderStatus {
	switch i {
	case 0:
		return OrderStatusPending
	case 1:
		return OrderStatusPartial
	case 2:
		return OrderStatusCompleted
	case 3:
		return OrderStatusCancelled
	default:
		return OrderStatusFailed
	}
}

// ChaseOrder 追单：修改订单价格并尝试成交
func (s *SpotTradingService) ChaseOrder(userID uint, req *ChaseOrderRequest) error {
	// 查询订单
	var order model.Transaction
	if err := database.DB.Where("id = ? AND from_user_id = ? AND status = 0", req.OrderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("订单不存在或已成交/撤销")
		}
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 更新订单价格和订单价值
	oldPrice := order.Price
	newOrderValue := req.NewPrice * order.Number

	if err := database.DB.Model(&order).Updates(map[string]interface{}{
		"price":       req.NewPrice,
		"order_value": newOrderValue,
	}).Error; err != nil {
		return fmt.Errorf("更新订单价格失败: %w", err)
	}

	logger.Infof("追单成功: orderID=%d, 旧价格=%.8f, 新价格=%.8f", order.ID, oldPrice, req.NewPrice)

	// 重新加载订单数据
	if err := database.DB.First(&order, order.ID).Error; err != nil {
		return err
	}

	// 尝试立即成交
	go s.tryInstantFill(&order)

	return nil
}

// canInstantFill 检查订单是否可以立即成交
func (s *SpotTradingService) canInstantFill(order *model.Transaction) bool {
	// 这里简化处理，实际应该检查市场价格
	// 对于买单：如果挂单价格 >= 市场价，可以成交
	// 对于卖单：如果挂单价格 <= 市场价，可以成交
	// 这里暂时返回true，表示始终尝试成交
	return true
}

// tryInstantFill 尝试立即成交订单
func (s *SpotTradingService) tryInstantFill(order *model.Transaction) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态为已成交
	if err := tx.Model(order).Updates(map[string]interface{}{
		"status":      2, // 已成交
		"deal_number": order.Number,
	}).Error; err != nil {
		tx.Rollback()
		logger.Errorf("更新订单状态失败: %v", err)
		return
	}

	// 获取币种名称
	currency, err := GetCurrencyService().GetCurrencyByID(uint(order.Currency))
	if err != nil {
		tx.Rollback()
		logger.Errorf("获取币种信息失败: %v", err)
		return
	}
	currencyName := currency.Name

	var tradedAmount float64
	if order.Type == 1 { // 买入：USDT -> 币
		tradedAmount = order.Number * order.Price

		// 从锁定USDT余额中扣除，增加币种余额
		if err := GetUserAssetsService().UpdateUsdtBalance(order.FromUserID, -tradedAmount, true); err != nil {
			tx.Rollback()
			logger.Errorf("扣除USDT锁定余额失败: %v", err)
			return
		}
		if err := GetUserAssetsService().UpdateCurrencyBalance(order.FromUserID, currencyName, order.Number, false); err != nil {
			tx.Rollback()
			logger.Errorf("增加币种余额失败: %v", err)
			return
		}

		logger.Infof("买单成交: orderID=%d, 扣除USDT=%.8f, 获得币种=%s 数量=%.8f",
			order.ID, tradedAmount, currencyName, order.Number)

	} else { // 卖出：币 -> USDT
		tradedAmount = order.Number * order.Price

		// 从锁定币种余额中扣除，增加USDT余额
		if err := GetUserAssetsService().UpdateCurrencyBalance(order.FromUserID, currencyName, -order.Number, true); err != nil {
			tx.Rollback()
			logger.Errorf("扣除币种锁定余额失败: %v", err)
			return
		}
		if err := GetUserAssetsService().UpdateUsdtBalance(order.FromUserID, tradedAmount, false); err != nil {
			tx.Rollback()
			logger.Errorf("增加USDT余额失败: %v", err)
			return
		}

		logger.Infof("卖单成交: orderID=%d, 扣除币种=%s 数量=%.8f, 获得USDT=%.8f",
			order.ID, currencyName, order.Number, tradedAmount)
	}

	tx.Commit()

	// WebSocket推送余额更新
	GetWalletService().BroadcastBalanceUpdate(order.FromUserID)

	logger.Infof("订单立即成交成功: orderID=%d", order.ID)
}
