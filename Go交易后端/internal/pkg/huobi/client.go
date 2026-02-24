package huobi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"exchange-go/config"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/logger"

	"github.com/gorilla/websocket"
)

// Client 火币WebSocket客户端
type Client struct {
	wsURL           string
	conn            *websocket.Conn
	mu              sync.RWMutex
	isConnected     bool
	reconnectDelay  time.Duration
	maxReconnect    int
	subscriptions   []string
	stopChan        chan struct{}
	tickerCallbacks []func(*TickerData)
	klineCallbacks  []func(*KlineData)
}

// TickerData 行情数据
type TickerData struct {
	Symbol    string  `json:"symbol"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int64   `json:"count"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	LastPrice float64 `json:"lastPrice"`
	PrevClose float64 `json:"prevClose"`
	Timestamp int64   `json:"ts"`
}

// KlineData K线数据
type KlineData struct {
	Symbol    string  `json:"symbol"`
	Period    string  `json:"period"`
	ID        int64   `json:"id"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int64   `json:"count"`
	Timestamp int64   `json:"ts"`
}

// DepthData 深度数据
type DepthData struct {
	Symbol    string      `json:"symbol"`
	Bids      [][]float64 `json:"bids"`
	Asks      [][]float64 `json:"asks"`
	Timestamp int64       `json:"ts"`
}

// HuobiMessage 火币消息结构
type HuobiMessage struct {
	Ping   int64           `json:"ping,omitempty"`
	Pong   int64           `json:"pong,omitempty"`
	Status string          `json:"status,omitempty"`
	Subbed string          `json:"subbed,omitempty"`
	Ch     string          `json:"ch,omitempty"`
	Ts     int64           `json:"ts,omitempty"`
	Tick   json.RawMessage `json:"tick,omitempty"`
	ErrMsg string          `json:"err-msg,omitempty"`
}

// TickerTick 行情Tick
type TickerTick struct {
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int64   `json:"count"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	LastPrice float64 `json:"lastPrice"`
	PrevClose float64 `json:"prevClose"`
}

// KlineTick K线Tick
type KlineTick struct {
	ID     int64   `json:"id"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Amount float64 `json:"amount"`
	Vol    float64 `json:"vol"`
	Count  int64   `json:"count"`
}

// NewClient 创建火币WebSocket客户端
func NewClient() *Client {
	wsURL := config.GlobalConfig.Huobi.WsURL
	if wsURL == "" {
		wsURL = "wss://api.huobi.pro/ws"
	}

	return &Client{
		wsURL:          wsURL,
		reconnectDelay: 5 * time.Second,
		maxReconnect:   -1, // 无限重连
		subscriptions:  make([]string, 0),
		stopChan:       make(chan struct{}),
	}
}

// Connect 连接到火币WebSocket
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("连接火币WebSocket失败: %w", err)
	}

	c.conn = conn
	c.isConnected = true
	logger.Info("火币WebSocket连接成功")

	return nil
}

// Start 启动客户端
func (c *Client) Start() error {
	if err := c.Connect(); err != nil {
		return err
	}

	// 启动消息接收协程
	go c.receiveMessages()

	// 启动心跳检测
	go c.heartbeat()

	return nil
}

// Stop 停止客户端
func (c *Client) Stop() {
	close(c.stopChan)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.isConnected = false
	logger.Info("火币WebSocket客户端已停止")
}

// Subscribe 订阅频道
func (c *Client) Subscribe(channel string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return fmt.Errorf("未连接到火币WebSocket")
	}

	msg := map[string]string{
		"sub": channel,
		"id":  fmt.Sprintf("sub_%d", time.Now().UnixNano()),
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("订阅失败: %w", err)
	}

	c.subscriptions = append(c.subscriptions, channel)
	logger.Infof("订阅频道: %s", channel)
	return nil
}

// SubscribeTicker 订阅行情
func (c *Client) SubscribeTicker(symbol string) error {
	channel := fmt.Sprintf("market.%s.ticker", symbol)
	return c.Subscribe(channel)
}

// SubscribeKline 订阅K线
func (c *Client) SubscribeKline(symbol, period string) error {
	channel := fmt.Sprintf("market.%s.kline.%s", symbol, period)
	return c.Subscribe(channel)
}

// SubscribeDepth 订阅深度
func (c *Client) SubscribeDepth(symbol string, depth int) error {
	channel := fmt.Sprintf("market.%s.depth.step%d", symbol, depth)
	return c.Subscribe(channel)
}

// receiveMessages 接收消息
func (c *Client) receiveMessages() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			c.mu.RLock()
			conn := c.conn
			isConnected := c.isConnected
			c.mu.RUnlock()

			if !isConnected || conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Errorf("读取消息失败: %v", err)
				c.handleDisconnect()
				continue
			}

			// 解压gzip数据
			data, err := c.decompress(message)
			if err != nil {
				logger.Errorf("解压消息失败: %v", err)
				continue
			}

			c.handleMessage(data)
		}
	}
}

// decompress 解压gzip数据
func (c *Client) decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// handleMessage 处理消息
func (c *Client) handleMessage(data []byte) {
	var msg HuobiMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Errorf("解析消息失败: %v, 数据: %s", err, string(data))
		return
	}

	// 处理心跳
	if msg.Ping > 0 {
		c.sendPong(msg.Ping)
		return
	}

	// 处理订阅确认
	if msg.Subbed != "" {
		logger.Infof("订阅确认: %s", msg.Subbed)
		return
	}

	// 处理错误
	if msg.ErrMsg != "" {
		logger.Errorf("火币错误: %s", msg.ErrMsg)
		return
	}

	// 处理数据
	if msg.Ch != "" && len(msg.Tick) > 0 {
		c.handleTick(msg.Ch, msg.Tick, msg.Ts)
	}
}

// handleTick 处理Tick数据
func (c *Client) handleTick(channel string, tick json.RawMessage, ts int64) {
	parts := strings.Split(channel, ".")
	if len(parts) < 3 {
		return
	}

	symbol := parts[1]
	dataType := parts[2]

	switch dataType {
	case "ticker":
		c.handleTickerData(symbol, tick, ts)
	case "kline":
		if len(parts) >= 4 {
			period := parts[3]
			c.handleKlineData(symbol, period, tick, ts)
		}
	case "depth":
		c.handleDepthData(symbol, tick, ts)
	}
}

// handleTickerData 处理行情数据
func (c *Client) handleTickerData(symbol string, tick json.RawMessage, ts int64) {
	var tickData TickerTick
	if err := json.Unmarshal(tick, &tickData); err != nil {
		logger.Errorf("解析Ticker数据失败: %v", err)
		return
	}

	data := &TickerData{
		Symbol:    symbol,
		Open:      tickData.Open,
		High:      tickData.High,
		Low:       tickData.Low,
		Close:     tickData.Close,
		Amount:    tickData.Amount,
		Vol:       tickData.Vol,
		Count:     tickData.Count,
		Bid:       tickData.Bid,
		BidSize:   tickData.BidSize,
		Ask:       tickData.Ask,
		AskSize:   tickData.AskSize,
		LastPrice: tickData.LastPrice,
		PrevClose: tickData.PrevClose,
		Timestamp: ts,
	}

	// 存储到Redis
	c.storeTickerToRedis(data)

	// 调用回调
	for _, callback := range c.tickerCallbacks {
		go callback(data)
	}
}

// handleKlineData 处理K线数据
func (c *Client) handleKlineData(symbol, period string, tick json.RawMessage, ts int64) {
	var tickData KlineTick
	if err := json.Unmarshal(tick, &tickData); err != nil {
		logger.Errorf("解析K线数据失败: %v", err)
		return
	}

	data := &KlineData{
		Symbol:    symbol,
		Period:    period,
		ID:        tickData.ID,
		Open:      tickData.Open,
		Close:     tickData.Close,
		Low:       tickData.Low,
		High:      tickData.High,
		Amount:    tickData.Amount,
		Vol:       tickData.Vol,
		Count:     tickData.Count,
		Timestamp: ts,
	}

	// 存储到Redis
	c.storeKlineToRedis(data)

	// 调用回调
	for _, callback := range c.klineCallbacks {
		go callback(data)
	}
}

// handleDepthData 处理深度数据
func (c *Client) handleDepthData(symbol string, tick json.RawMessage, ts int64) {
	var depthData struct {
		Bids [][]float64 `json:"bids"`
		Asks [][]float64 `json:"asks"`
	}
	if err := json.Unmarshal(tick, &depthData); err != nil {
		logger.Errorf("解析深度数据失败: %v", err)
		return
	}

	data := &DepthData{
		Symbol:    symbol,
		Bids:      depthData.Bids,
		Asks:      depthData.Asks,
		Timestamp: ts,
	}

	// 存储到Redis
	c.storeDepthToRedis(data)
}

// storeTickerToRedis 存储行情到Redis
func (c *Client) storeTickerToRedis(data *TickerData) {
	key := fmt.Sprintf("ticker:%s", data.Symbol)
	if err := cache.Set(key, data, 60*time.Second); err != nil {
		logger.Errorf("存储Ticker到Redis失败: %v", err)
	}
}

// storeKlineToRedis 存储K线到Redis
func (c *Client) storeKlineToRedis(data *KlineData) {
	key := fmt.Sprintf("kline:%s:%s", data.Symbol, data.Period)
	// 使用List存储K线数据，保留最近1000条
	jsonData, _ := json.Marshal(data)
	if err := cache.LPush(key, string(jsonData)); err != nil {
		logger.Errorf("存储K线到Redis失败: %v", err)
	}
	// 限制长度
	cache.RDB.LTrim(cache.Ctx, key, 0, 999)
}

// storeDepthToRedis 存储深度到Redis
func (c *Client) storeDepthToRedis(data *DepthData) {
	key := fmt.Sprintf("depth:%s", data.Symbol)
	if err := cache.Set(key, data, 30*time.Second); err != nil {
		logger.Errorf("存储深度到Redis失败: %v", err)
	}
}

// sendPong 发送Pong响应（优化：添加写入超时，避免阻塞）
func (c *Client) sendPong(ping int64) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return
	}

	// 设置写入超时，避免阻塞心跳响应
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	msg := map[string]int64{"pong": ping}
	if err := conn.WriteJSON(msg); err != nil {
		logger.Errorf("发送Pong失败: %v", err)
	}

	// 清除写入超时
	conn.SetWriteDeadline(time.Time{})
}

// heartbeat 心跳检测
func (c *Client) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.mu.RLock()
			isConnected := c.isConnected
			c.mu.RUnlock()

			if !isConnected {
				c.reconnect()
			}
		}
	}
}

// handleDisconnect 处理断开连接
func (c *Client) handleDisconnect() {
	c.mu.Lock()
	c.isConnected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	logger.Warn("火币WebSocket连接断开，准备重连...")
	go c.reconnect()
}

// reconnect 重新连接
func (c *Client) reconnect() {
	reconnectCount := 0
	for {
		select {
		case <-c.stopChan:
			return
		default:
			if c.maxReconnect > 0 && reconnectCount >= c.maxReconnect {
				logger.Error("火币WebSocket重连次数已达上限")
				return
			}

			logger.Infof("尝试重连火币WebSocket... (第%d次)", reconnectCount+1)

			if err := c.Connect(); err != nil {
				logger.Errorf("重连失败: %v", err)
				reconnectCount++
				time.Sleep(c.reconnectDelay)
				continue
			}

			// 重新订阅所有频道
			c.mu.Lock()
			subs := make([]string, len(c.subscriptions))
			copy(subs, c.subscriptions)
			c.subscriptions = c.subscriptions[:0]
			c.mu.Unlock()

			for _, sub := range subs {
				if err := c.Subscribe(sub); err != nil {
					logger.Errorf("重新订阅失败: %v", err)
				}
				time.Sleep(100 * time.Millisecond) // 避免订阅过快
			}

			logger.Info("火币WebSocket重连成功")
			return
		}
	}
}

// OnTicker 注册Ticker回调
func (c *Client) OnTicker(callback func(*TickerData)) {
	c.tickerCallbacks = append(c.tickerCallbacks, callback)
}

// OnKline 注册K线回调
func (c *Client) OnKline(callback func(*KlineData)) {
	c.klineCallbacks = append(c.klineCallbacks, callback)
}

// GetTickerFromRedis 从Redis获取行情数据
func GetTickerFromRedis(symbol string) (*TickerData, error) {
	key := fmt.Sprintf("ticker:%s", symbol)
	var data TickerData
	if err := cache.Get(key, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetDepthFromRedis 从Redis获取深度数据
func GetDepthFromRedis(symbol string) (*DepthData, error) {
	key := fmt.Sprintf("depth:%s", symbol)
	var data DepthData
	if err := cache.Get(key, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetKlineFromRedis 从Redis获取K线数据
func GetKlineFromRedis(symbol, period string, limit int) ([]KlineData, error) {
	key := fmt.Sprintf("kline:%s:%s", symbol, period)
	results, err := cache.LRange(key, 0, int64(limit-1))
	if err != nil {
		return nil, err
	}

	var klines []KlineData
	for _, item := range results {
		var kline KlineData
		if err := json.Unmarshal([]byte(item), &kline); err != nil {
			continue
		}
		klines = append(klines, kline)
	}
	return klines, nil
}
