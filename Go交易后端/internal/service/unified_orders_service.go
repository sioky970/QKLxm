package service

import (
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"fmt"
	"time"
)

// UnifiedOrdersService 统一订单查询服务
type UnifiedOrdersService struct{}

// NewUnifiedOrdersService 创建统一订单服务实例
func NewUnifiedOrdersService() *UnifiedOrdersService {
	return &UnifiedOrdersService{}
}

// TradeType 交易类型常量
const (
	TradeTypeSpot     = "spot"     // 现货交易
	TradeTypeContract = "contract" // 永续合约
	TradeTypeMicro    = "micro"    // 秒合约
)

// UnifiedOrderListRequest 统一订单列表请求
type UnifiedOrderListRequest struct {
	TradeType  string `json:"trade_type"`  // 交易类型: spot, contract, micro, 空则查询全部
	CurrencyID uint   `json:"currency_id"` // 币种ID筛选
	Status     *int   `json:"status"`      // 状态筛选
	Side       string `json:"side"`        // 交易方向筛选
	Page       int    `json:"page"`        // 页码
	PageSize   int    `json:"page_size"`   // 每页数量
}

// UnifiedOrder 统一订单结构
type UnifiedOrder struct {
	ID           uint    `json:"id"`
	TradeType    string  `json:"trade_type"`    // 交易类型: spot, contract, micro
	TradeTypeCN  string  `json:"trade_type_cn"` // 交易类型中文
	Symbol       string  `json:"symbol"`        // 交易对
	CurrencyID   uint    `json:"currency_id"`   // 币种ID
	CurrencyName string  `json:"currency_name"` // 币种名称
	Side         string  `json:"side"`          // 买/卖方向
	SideCN       string  `json:"side_cn"`       // 方向中文
	OrderType    string  `json:"order_type"`    // 订单类型: market/limit
	OrderTypeCN  string  `json:"order_type_cn"` // 订单类型中文
	Price        float64 `json:"price"`         // 价格
	Quantity     float64 `json:"quantity"`      // 数量
	DealQuantity float64 `json:"deal_quantity"` // 已成交数量
	Amount       float64 `json:"amount"`        // 金额
	Fee          float64 `json:"fee"`           // 手续费
	PnL          float64 `json:"pnl"`           // 盈亏
	Status       int     `json:"status"`        // 状态码
	StatusCN     string  `json:"status_cn"`     // 状态中文
	Leverage     int     `json:"leverage"`      // 杠杆倍数（合约专用）
	Duration     int     `json:"duration"`      // 持续时间（秒合约专用）
	CreateTime   int64   `json:"create_time"`   // 创建时间戳
	UpdateTime   int64   `json:"update_time"`   // 更新时间戳
}

// UnifiedOrderListResponse 统一订单列表响应
type UnifiedOrderListResponse struct {
	List     []UnifiedOrder `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// GetUnifiedOrderList 获取统一订单列表
func (s *UnifiedOrdersService) GetUnifiedOrderList(userID uint, req *UnifiedOrderListRequest) (*UnifiedOrderListResponse, error) {
	var allOrders []UnifiedOrder
	var total int64

	// 根据请求的交易类型查询对应的订单
	switch req.TradeType {
	case TradeTypeSpot:
		orders, count, err := s.getSpotOrders(userID, req)
		if err != nil {
			return nil, err
		}
		allOrders = orders
		total = count
	case TradeTypeContract:
		orders, count, err := s.getContractOrders(userID, req)
		if err != nil {
			return nil, err
		}
		allOrders = orders
		total = count
	case TradeTypeMicro:
		orders, count, err := s.getMicroOrders(userID, req)
		if err != nil {
			return nil, err
		}
		allOrders = orders
		total = count
	default:
		// 查询所有类型的订单
		spotOrders, spotCount, err := s.getSpotOrders(userID, req)
		if err != nil {
			return nil, err
		}
		contractOrders, contractCount, err := s.getContractOrders(userID, req)
		if err != nil {
			return nil, err
		}
		microOrders, microCount, err := s.getMicroOrders(userID, req)
		if err != nil {
			return nil, err
		}

		// 合并所有订单
		allOrders = append(allOrders, spotOrders...)
		allOrders = append(allOrders, contractOrders...)
		allOrders = append(allOrders, microOrders...)
		total = spotCount + contractCount + microCount

		// 按创建时间排序（降序）
		s.sortOrdersByTime(allOrders)

		// 分页处理
		offset := (req.Page - 1) * req.PageSize
		if offset >= len(allOrders) {
			allOrders = []UnifiedOrder{}
		} else {
			end := offset + req.PageSize
			if end > len(allOrders) {
				end = len(allOrders)
			}
			allOrders = allOrders[offset:end]
		}
	}

	return &UnifiedOrderListResponse{
		List:     allOrders,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// getSpotOrders 获取现货交易订单
func (s *UnifiedOrdersService) getSpotOrders(userID uint, req *UnifiedOrderListRequest) ([]UnifiedOrder, int64, error) {
	var orders []model.Transaction
	var total int64

	query := database.DB.Model(&model.Transaction{}).Where("from_user_id = ?", userID)

	// 币种筛选
	if req.CurrencyID > 0 {
		query = query.Where("currency = ?", req.CurrencyID)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 方向筛选
	if req.Side == "buy" {
		query = query.Where("type = 1")
	} else if req.Side == "sell" {
		query = query.Where("type = 2")
	}

	// 计数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	// 转换为统一格式
	var result []UnifiedOrder
	for _, order := range orders {
		result = append(result, s.convertSpotOrder(order))
	}

	return result, total, nil
}

// getContractOrders 获取永续合约订单
func (s *UnifiedOrdersService) getContractOrders(userID uint, req *UnifiedOrderListRequest) ([]UnifiedOrder, int64, error) {
	var orders []model.LeverTransaction
	var total int64

	query := database.DB.Model(&model.LeverTransaction{}).Where("user_id = ?", userID)

	// 币种筛选
	if req.CurrencyID > 0 {
		query = query.Where("currency = ?", req.CurrencyID)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 方向筛选
	if req.Side == "buy" || req.Side == "long" {
		query = query.Where("type = 1")
	} else if req.Side == "sell" || req.Side == "short" {
		query = query.Where("type = 2")
	}

	// 计数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	// 转换为统一格式
	var result []UnifiedOrder
	for _, order := range orders {
		result = append(result, s.convertContractOrder(order))
	}

	return result, total, nil
}

// getMicroOrders 获取秒合约订单
func (s *UnifiedOrdersService) getMicroOrders(userID uint, req *UnifiedOrderListRequest) ([]UnifiedOrder, int64, error) {
	var orders []model.MicroOrder
	var total int64

	query := database.DB.Model(&model.MicroOrder{}).Where("user_id = ?", userID)

	// 币种筛选
	if req.CurrencyID > 0 {
		query = query.Where("currency_id = ?", req.CurrencyID)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 方向筛选（秒合约: type=1买涨, type=2买跌）
	if req.Side == "buy" || req.Side == "rise" {
		query = query.Where("type = 1")
	} else if req.Side == "sell" || req.Side == "fall" {
		query = query.Where("type = 2")
	}

	// 计数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	// 转换为统一格式
	var result []UnifiedOrder
	for _, order := range orders {
		result = append(result, s.convertMicroOrder(order))
	}

	return result, total, nil
}

// convertSpotOrder 转换现货订单为统一格式
func (s *UnifiedOrdersService) convertSpotOrder(order model.Transaction) UnifiedOrder {
	// 获取币种名称
	currencyID := uint(order.Currency)
	currencyName := s.getCurrencyName(currencyID)
	symbol := fmt.Sprintf("%s/USDT", currencyName)

	// 方向
	side := "buy"
	sideCN := "买入"
	if order.Type == 2 {
		side = "sell"
		sideCN = "卖出"
	}

	// 订单类型
	orderType := "limit"
	orderTypeCN := "限价单"
	if order.Price == 0 {
		orderType = "market"
		orderTypeCN = "市价单"
	}

	// 状态
	statusCN := s.getSpotStatusCN(int(order.Status))

	return UnifiedOrder{
		ID:           order.ID,
		TradeType:    TradeTypeSpot,
		TradeTypeCN:  "现货交易",
		Symbol:       symbol,
		CurrencyID:   currencyID,
		CurrencyName: currencyName,
		Side:         side,
		SideCN:       sideCN,
		OrderType:    orderType,
		OrderTypeCN:  orderTypeCN,
		Price:        order.Price,
		Quantity:     order.Number,
		DealQuantity: order.DealNumber,
		Amount:       order.OrderValue,
		Fee:          0,
		PnL:          0,
		Status:       int(order.Status),
		StatusCN:     statusCN,
		Leverage:     1,
		Duration:     0,
		CreateTime:   order.Time,
		UpdateTime:   order.Time,
	}
}

// convertContractOrder 转换合约订单为统一格式
func (s *UnifiedOrdersService) convertContractOrder(order model.LeverTransaction) UnifiedOrder {
	// 获取币种名称
	currencyName := s.getCurrencyName(order.CurrencyID)
	symbol := fmt.Sprintf("%s/USDT", currencyName)

	// 方向
	side := "long"
	sideCN := "做多"
	if order.Type == 2 {
		side = "short"
		sideCN = "做空"
	}

	// 订单类型
	orderType := "market"
	orderTypeCN := "市价单"
	if order.OrderType == 2 {
		orderType = "limit"
		orderTypeCN = "限价单"
	}

	// 状态
	statusCN := s.getContractStatusCN(int(order.Status))

	return UnifiedOrder{
		ID:           order.ID,
		TradeType:    TradeTypeContract,
		TradeTypeCN:  "永续合约",
		Symbol:       symbol,
		CurrencyID:   order.CurrencyID,
		CurrencyName: currencyName,
		Side:         side,
		SideCN:       sideCN,
		OrderType:    orderType,
		OrderTypeCN:  orderTypeCN,
		Price:        order.Price,
		Quantity:     order.OriginCautionMoney,
		DealQuantity: order.OriginCautionMoney,
		Amount:       order.OriginCautionMoney * float64(order.Multiple),
		Fee:          order.TradeFee,
		PnL:          order.FactProfits,
		Status:       int(order.Status),
		StatusCN:     statusCN,
		Leverage:     order.Multiple,
		Duration:     0,
		CreateTime:   order.CreateTime,
		UpdateTime:   int64(order.UpdateTime),
	}
}

// convertMicroOrder 转换秒合约订单为统一格式
func (s *UnifiedOrdersService) convertMicroOrder(order model.MicroOrder) UnifiedOrder {
	// 获取币种名称
	currencyName := s.getCurrencyName(order.CurrencyID)
	symbol := fmt.Sprintf("%s/USDT", currencyName)

	// 方向
	side := "rise"
	sideCN := "买涨"
	if order.Type == 2 {
		side = "fall"
		sideCN = "买跌"
	}

	// 订单类型（秒合约都是市价单）
	orderType := "market"
	orderTypeCN := "市价单"

	// 状态
	statusCN := s.getMicroStatusCN(int(order.Status))

	return UnifiedOrder{
		ID:           order.ID,
		TradeType:    TradeTypeMicro,
		TradeTypeCN:  "交割合约",
		Symbol:       symbol,
		CurrencyID:   order.CurrencyID,
		CurrencyName: currencyName,
		Side:         side,
		SideCN:       sideCN,
		OrderType:    orderType,
		OrderTypeCN:  orderTypeCN,
		Price:        order.OpenPrice,
		Quantity:     order.Number,
		DealQuantity: order.Number,
		Amount:       order.Number,
		Fee:          order.Fee,
		PnL:          order.FactProfits,
		Status:       int(order.Status),
		StatusCN:     statusCN,
		Leverage:     1,
		Duration:     int(order.Seconds),
		CreateTime:   order.CreatedAt.Unix(),
		UpdateTime:   order.UpdatedAt.Unix(),
	}
}

// getCurrencyName 获取币种名称
func (s *UnifiedOrdersService) getCurrencyName(currencyID uint) string {
	var currency model.Currency
	if err := database.DB.Where("id = ?", currencyID).First(&currency).Error; err != nil {
		return "UNKNOWN"
	}
	return currency.Name
}

// getSpotStatusCN 获取现货订单状态中文
func (s *UnifiedOrdersService) getSpotStatusCN(status int) string {
	statusMap := map[int]string{
		0: "未成交",
		1: "部分成交",
		2: "已成交",
		3: "已撤销",
	}
	if cn, ok := statusMap[status]; ok {
		return cn
	}
	return "未知"
}

// getContractStatusCN 获取合约订单状态中文
func (s *UnifiedOrdersService) getContractStatusCN(status int) string {
	statusMap := map[int]string{
		0: "挂单中",
		1: "持仓中",
		2: "平仓中",
		3: "已平仓",
	}
	if cn, ok := statusMap[status]; ok {
		return cn
	}
	return "未知"
}

// getMicroStatusCN 获取秒合约订单状态中文
func (s *UnifiedOrdersService) getMicroStatusCN(status int) string {
	statusMap := map[int]string{
		0: "进行中",
		1: "已结算",
	}
	if cn, ok := statusMap[status]; ok {
		return cn
	}
	return "未知"
}

// sortOrdersByTime 按时间排序（降序）
func (s *UnifiedOrdersService) sortOrdersByTime(orders []UnifiedOrder) {
	for i := 0; i < len(orders)-1; i++ {
		for j := i + 1; j < len(orders); j++ {
			if orders[j].CreateTime > orders[i].CreateTime {
				orders[i], orders[j] = orders[j], orders[i]
			}
		}
	}
}

// GetOrderStatistics 获取订单统计
func (s *UnifiedOrdersService) GetOrderStatistics(userID uint) (*OrderStatistics, error) {
	stats := &OrderStatistics{}

	// 现货订单统计
	database.DB.Model(&model.Transaction{}).Where("from_user_id = ? AND status = 0", userID).Count(&stats.SpotPending)
	database.DB.Model(&model.Transaction{}).Where("from_user_id = ?", userID).Count(&stats.SpotTotal)

	// 合约订单统计
	database.DB.Model(&model.LeverTransaction{}).Where("user_id = ? AND status = 0", userID).Count(&stats.ContractPending)
	database.DB.Model(&model.LeverTransaction{}).Where("user_id = ? AND status = 1", userID).Count(&stats.ContractActive)
	database.DB.Model(&model.LeverTransaction{}).Where("user_id = ?", userID).Count(&stats.ContractTotal)

	// 秒合约订单统计
	database.DB.Model(&model.MicroOrder{}).Where("user_id = ? AND status = 0", userID).Count(&stats.MicroActive)
	database.DB.Model(&model.MicroOrder{}).Where("user_id = ?", userID).Count(&stats.MicroTotal)

	return stats, nil
}

// OrderStatistics 订单统计结构
type OrderStatistics struct {
	SpotPending     int64 `json:"spot_pending"`     // 现货待成交
	SpotTotal       int64 `json:"spot_total"`       // 现货总数
	ContractPending int64 `json:"contract_pending"` // 合约挂单中
	ContractActive  int64 `json:"contract_active"`  // 合约持仓中
	ContractTotal   int64 `json:"contract_total"`   // 合约总数
	MicroActive     int64 `json:"micro_active"`     // 秒合约进行中
	MicroTotal      int64 `json:"micro_total"`      // 秒合约总数
}

var unifiedOrdersService *UnifiedOrdersService

// GetUnifiedOrdersService 获取统一订单服务单例
func GetUnifiedOrdersService() *UnifiedOrdersService {
	if unifiedOrdersService == nil {
		unifiedOrdersService = NewUnifiedOrdersService()
	}
	return unifiedOrdersService
}

// GetOrderByID 根据ID和类型获取订单详情
func (s *UnifiedOrdersService) GetOrderByID(userID uint, orderID uint, tradeType string) (*UnifiedOrder, error) {
	switch tradeType {
	case TradeTypeSpot:
		var order model.Transaction
		if err := database.DB.Where("id = ? AND from_user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return nil, err
		}
		result := s.convertSpotOrder(order)
		return &result, nil

	case TradeTypeContract:
		var order model.LeverTransaction
		if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return nil, err
		}
		result := s.convertContractOrder(order)
		return &result, nil

	case TradeTypeMicro:
		var order model.MicroOrder
		if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return nil, err
		}
		result := s.convertMicroOrder(order)
		return &result, nil

	default:
		return nil, fmt.Errorf("无效的交易类型: %s", tradeType)
	}
}

// GetTodayOrderCount 获取今日订单数量
func (s *UnifiedOrdersService) GetTodayOrderCount(userID uint) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour).Unix()

	var spotCount, contractCount, microCount int64
	database.DB.Model(&model.Transaction{}).Where("from_user_id = ? AND time >= ?", userID, today).Count(&spotCount)
	database.DB.Model(&model.LeverTransaction{}).Where("user_id = ? AND create_time >= ?", userID, today).Count(&contractCount)
	database.DB.Model(&model.MicroOrder{}).Where("user_id = ? AND created_at >= ?", userID, time.Unix(today, 0)).Count(&microCount)

	return spotCount + contractCount + microCount, nil
}
