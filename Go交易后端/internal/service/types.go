package service

// 通用订单类型定义
// 本文件包含所有交易服务共享的类型定义

// OrderType 订单类型
type OrderType string

const (
	OrderTypeLimit      OrderType = "limit"       // 限价单
	OrderTypeMarket     OrderType = "market"      // 市价单
	OrderTypeStopLoss   OrderType = "stop_loss"   // 止损单
	OrderTypeTakeProfit OrderType = "take_profit" // 止盈单
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // 挂单中/待成交
	OrderStatusPartial   OrderStatus = "partial"   // 部分成交
	OrderStatusActive    OrderStatus = "active"    // 交易中
	OrderStatusClosing   OrderStatus = "closing"   // 平仓中
	OrderStatusClosed    OrderStatus = "closed"    // 已平仓
	OrderStatusCompleted OrderStatus = "completed" // 完全成交
	OrderStatusCancelled OrderStatus = "cancelled" // 已撤单
	OrderStatusFailed    OrderStatus = "failed"    // 失败
	OrderStatusOpened    OrderStatus = "opened"    // 交易中(秒合约专用)
)

// SubmitOrderRequest 提交订单请求（通用）
type SubmitOrderRequest struct {
	CurrencyID   uint    `json:"currency_id" binding:"required"`
	LegalID      uint    `json:"legal_id" binding:"required"`
	Price        float64 `json:"price"`
	Quantity     float64 `json:"quantity"`
	TriggerPrice *float64 `json:"trigger_price"`
}

// SubmitOrderResponse 提交订单响应（通用）
type SubmitOrderResponse struct {
	OrderID    uint        `json:"order_id"`
	OrderNo    string      `json:"order_no"`
	Status     OrderStatus `json:"status"`
	CreateTime interface{} `json:"create_time"`
}
