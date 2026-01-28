package huobi

import (
	"fmt"
	"sync"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/websocket"
)

// Manager manages Huobi market data subscriptions and cache sync.
type Manager struct {
	client     *Client
	symbols    []string
	symbolMap  map[string]SymbolInfo // huobi symbol -> local symbol info
	mu         sync.RWMutex
	syncTicker *time.Ticker
	stopChan   chan struct{}
	monitor    *Monitor
}

// SymbolInfo describes mapping between Huobi symbol and local currency IDs.
type SymbolInfo struct {
	HuobiSymbol  string // huobi format, e.g. btcusdt
	LocalSymbol  string // local format, e.g. BTC/USDT
	CurrencyID   uint
	LegalID      uint
	CurrencyName string
	LegalName    string
}

// NewManager creates a new Huobi manager.
func NewManager() *Manager {
	return &Manager{
		client:    NewClient(),
		symbols:   []string{},
		symbolMap: make(map[string]SymbolInfo),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the Huobi manager.
func (m *Manager) Start() error {
	if err := m.initSymbolMap(); err != nil {
		logger.Warnf("初始化交易对映射失败: %v", err)
	}

	if err := m.client.Start(); err != nil {
		return fmt.Errorf("启动火币WebSocket客户端失败: %w", err)
	}

	m.client.OnTicker(m.onTickerUpdate)
	m.client.OnKline(m.onKlineUpdate)

	time.Sleep(1 * time.Second)
	m.subscribeSymbols()

	m.syncTicker = time.NewTicker(5 * time.Second)
	go m.syncToDatabase()

	// 立即执行一次同步，确保启动时有数据
	time.Sleep(2 * time.Second) // 等待WebSocket连接建立并接收数据
	m.doSyncToDatabase()
	logger.Info("初始数据同步完成")

	// 初始化并启动监控器
	historyScheduler := GetHistoryKlineScheduler()
	m.monitor = NewMonitor(m, historyScheduler)
	m.monitor.Start()

	logger.Info("火币数据管理器启动成功")
	return nil
}

// Stop stops the Huobi manager.
func (m *Manager) Stop() {
	close(m.stopChan)
	if m.syncTicker != nil {
		m.syncTicker.Stop()
	}
	m.client.Stop()
	logger.Info("火币数据管理器已停止")
}

// RefreshSymbols reloads enabled currencies and restarts subscriptions.
func (m *Manager) RefreshSymbols() error {
	symbols, symbolMap, err := loadEnabledSymbolMap()
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.symbols = symbols
	m.symbolMap = symbolMap
	m.mu.Unlock()

	if m.client != nil {
		m.client.Stop()
	}
	m.client = NewClient()
	m.client.OnTicker(m.onTickerUpdate)
	m.client.OnKline(m.onKlineUpdate)

	if err := m.client.Start(); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	m.subscribeSymbols()
	logger.Infof("订阅已刷新，共%d个交易对", len(symbolMap))

	// 重新初始化监控器以适应新的交易对
	if m.monitor != nil {
		historyScheduler := GetHistoryKlineScheduler()
		m.monitor = NewMonitor(m, historyScheduler)
		m.monitor.Start()
	}

	return nil
}

func (m *Manager) initSymbolMap() error {
	symbols, symbolMap, err := loadEnabledSymbolMap()
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.symbols = symbols
	m.symbolMap = symbolMap
	m.mu.Unlock()

	logger.Infof("初始化交易对映射完成，共%d个交易对", len(symbolMap))
	return nil
}

func (m *Manager) subscribeSymbols() {
	for _, symbol := range m.symbols {
		if err := m.client.SubscribeTicker(symbol); err != nil {
			logger.Errorf("订阅%s行情失败: %v", symbol, err)
		}
		time.Sleep(500 * time.Millisecond)

		// 改为订陕5分钟K线，不再订陕1分钟
		if err := m.client.SubscribeKline(symbol, "5min"); err != nil {
			logger.Errorf("订阅%s K线失败: %v", symbol, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (m *Manager) onTickerUpdate(data *TickerData) {
	m.mu.RLock()
	info, ok := m.symbolMap[data.Symbol]
	m.mu.RUnlock()
	if !ok {
		return
	}

	change := "0.00%"
	if data.PrevClose > 0 {
		changeRate := (data.Close - data.PrevClose) / data.PrevClose * 100
		if changeRate >= 0 {
			change = fmt.Sprintf("+%.2f%%", changeRate)
		} else {
			change = fmt.Sprintf("%.2f%%", changeRate)
		}
	} else if data.Open > 0 {
		changeRate := (data.Close - data.Open) / data.Open * 100
		if changeRate >= 0 {
			change = fmt.Sprintf("+%.2f%%", changeRate)
		} else {
			change = fmt.Sprintf("%.2f%%", changeRate)
		}
	}

	enhancedData := map[string]interface{}{
		"symbol":        data.Symbol,
		"currency_id":   info.CurrencyID,
		"legal_id":      info.LegalID,
		"currency_name": info.CurrencyName,
		"legal_name":    info.LegalName,
		"price":         data.Close,
		"now_price":     data.Close,
		"open":          data.Open,
		"high":          data.High,
		"low":           data.Low,
		"close":         data.Close,
		"volume":        data.Vol,
		"amount":        data.Amount,
		"change":        change,
		"bid":           data.Bid,
		"ask":           data.Ask,
		"timestamp":     data.Timestamp,
	}

	key := fmt.Sprintf("market:ticker:%d:%d", info.CurrencyID, info.LegalID)
	if err := cache.Set(key, enhancedData, 60*time.Second); err != nil {
		logger.Errorf("存储行情数据失败: %v", err)
	}

	key2 := fmt.Sprintf("market:ticker:%s", data.Symbol)
	cache.Set(key2, enhancedData, 60*time.Second)

	// 推送到WebSocket客户端
	m.broadcastMarketData(enhancedData)
}

func (m *Manager) onKlineUpdate(data *KlineData) {
	m.mu.RLock()
	info, ok := m.symbolMap[data.Symbol]
	m.mu.RUnlock()
	if !ok {
		return
	}

	klineKey := fmt.Sprintf("market:kline:%d:%d:%s", info.CurrencyID, info.LegalID, data.Period)
	cache.Set(klineKey+":latest", data, 60*time.Second)
}

func (m *Manager) syncToDatabase() {
	for {
		select {
		case <-m.stopChan:
			return
		case <-m.syncTicker.C:
			m.doSyncToDatabase()
		}
	}
}

func (m *Manager) doSyncToDatabase() {
	m.mu.RLock()
	symbols := m.symbolMap
	m.mu.RUnlock()

	if len(symbols) == 0 {
		logger.Debug("没有需要同步的交易对")
		return
	}

	successCount := 0
	for symbol, info := range symbols {
		ticker, err := GetTickerFromRedis(symbol)
		if err != nil || ticker == nil {
			logger.Debugf("从Redis获取%s行情失败: %v", symbol, err)
			continue
		}

		if ticker.Close == 0 {
			logger.Debugf("%s价格为0，跳过", symbol)
			continue
		}

		change := "0.00%"
		if ticker.PrevClose > 0 {
			changeRate := (ticker.Close - ticker.PrevClose) / ticker.PrevClose * 100
			if changeRate >= 0 {
				change = fmt.Sprintf("+%.2f%%", changeRate)
			} else {
				change = fmt.Sprintf("%.2f%%", changeRate)
			}
		} else if ticker.Open > 0 {
			changeRate := (ticker.Close - ticker.Open) / ticker.Open * 100
			if changeRate >= 0 {
				change = fmt.Sprintf("+%.2f%%", changeRate)
			} else {
				change = fmt.Sprintf("%.2f%%", changeRate)
			}
		}

		var quotation model.CurrencyQuotation
		result := database.DB.Where("currency_id = ? AND legal_id = ?", info.CurrencyID, info.LegalID).First(&quotation)

		if result.RowsAffected == 0 {
			quotation = model.CurrencyQuotation{
				CurrencyID: info.CurrencyID,
				LegalID:    info.LegalID,
				Price:      ticker.Close,
				NowPrice:   ticker.Close,
				Change:     change,
				Volume:     ticker.Vol,
				AddTime:    time.Now().Unix(),
			}
			if err := database.DB.Create(&quotation).Error; err != nil {
				logger.Errorf("创建%s行情失败: %v", symbol, err)
				continue
			}
			logger.Infof("创建%s行情成功: 价格=%.2f, 涨跌幅=%s", symbol, ticker.Close, change)
		} else {
			if err := database.DB.Model(&quotation).Updates(map[string]interface{}{
				"now_price": ticker.Close,
				"change":    change,
				"volume":    ticker.Vol,
				"add_time":  time.Now().Unix(),
			}).Error; err != nil {
				logger.Errorf("更新%s行情失败: %v", symbol, err)
				continue
			}
		}
		successCount++
	}

	if successCount > 0 {
		logger.Debugf("同步%d个交易对行情到数据库", successCount)
	}
}

// AddSymbol adds a new symbol to subscribe.
func (m *Manager) AddSymbol(symbol string) error {
	m.mu.Lock()
	m.symbols = append(m.symbols, symbol)
	m.mu.Unlock()

	if err := m.client.SubscribeTicker(symbol); err != nil {
		return err
	}
	// 订陕5分钟K线，不再订陕1分钟
	return m.client.SubscribeKline(symbol, "5min")
}

// GetMarketData fetches ticker data from Redis.
func (m *Manager) GetMarketData(currencyID, legalID uint) (map[string]interface{}, error) {
	key := fmt.Sprintf("market:ticker:%d:%d", currencyID, legalID)
	var data map[string]interface{}
	if err := cache.Get(key, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// GetAllMarketData fetches ticker data for all enabled symbols.
func (m *Manager) GetAllMarketData() ([]map[string]interface{}, error) {
	m.mu.RLock()
	symbols := m.symbolMap
	m.mu.RUnlock()

	var result []map[string]interface{}
	for _, info := range symbols {
		key := fmt.Sprintf("market:ticker:%d:%d", info.CurrencyID, info.LegalID)
		var data map[string]interface{}
		if err := cache.Get(key, &data); err != nil {
			continue
		}
		result = append(result, data)
	}
	return result, nil
}

var globalManager *Manager
var managerOnce sync.Once

// GetManager returns the global manager instance.
func GetManager() *Manager {
	managerOnce.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// broadcastMarketData 广播行情数据到WebSocket客户端
func (m *Manager) broadcastMarketData(data map[string]interface{}) {
	websocket.BroadcastMarketData(data)
}
