package huobi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

// KlinePeriods defines supported kline periods.
var KlinePeriods = map[string]string{
	"5min":  "5min",
	"15min": "15min",
	"60min": "60min",
	"4hour": "4hour",
	"1day":  "1day",
}

// KlinePeriodsPolling lists periods for polling tasks.
var KlinePeriodsPolling = []string{"5min", "15min", "60min", "4hour", "1day"}

// HuobiKlineResponse represents Huobi Kline API response.
type HuobiKlineResponse struct {
	Status string       `json:"status"`
	Ch     string       `json:"ch"`
	Ts     int64        `json:"ts"`
	Data   []HuobiKline `json:"data"`
}

// HuobiKline represents a single kline.
type HuobiKline struct {
	ID     int64   `json:"id"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Amount float64 `json:"amount"`
	Vol    float64 `json:"vol"`
	Count  int64   `json:"count"`
}

// HistoryKlineScheduler fetches and stores historical klines.
type HistoryKlineScheduler struct {
	httpClient *http.Client
	baseURL    string
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
	symbols    []string
	symbolMap  map[string]SymbolInfo
}

// NewHistoryKlineScheduler creates a new scheduler.
func NewHistoryKlineScheduler() *HistoryKlineScheduler {
	return &HistoryKlineScheduler{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.huobi.pro",
		running:   false,
		stopChan:  make(chan struct{}),
		symbols:   []string{},
		symbolMap: make(map[string]SymbolInfo),
	}
}

// Start starts the scheduler.
func (s *HistoryKlineScheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	manager := GetManager()
	if manager != nil {
		manager.mu.RLock()
		if len(manager.symbolMap) > 0 {
			symbolMap := make(map[string]SymbolInfo, len(manager.symbolMap))
			for k, v := range manager.symbolMap {
				symbolMap[k] = v
			}
			s.mu.Lock()
			s.symbolMap = symbolMap
			s.symbols = append([]string(nil), manager.symbols...)
			s.mu.Unlock()
		}
		manager.mu.RUnlock()
	}

	s.mu.RLock()
	hasSymbols := len(s.symbols) > 0
	s.mu.RUnlock()
	if !hasSymbols {
		symbols, symbolMap, err := loadEnabledSymbolMap()
		if err != nil {
			logger.Warnf("初始化历史K线交易对失败: %v", err)
		} else {
			s.mu.Lock()
			s.symbols = symbols
			s.symbolMap = symbolMap
			s.mu.Unlock()
		}
	}

	go s.fetchAllHistoryKlines()
	go s.startPolling()

	logger.Info("历史K线调度器启动成功")
	return nil
}

// Stop stops the scheduler.
func (s *HistoryKlineScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		close(s.stopChan)
		s.running = false
	}
	logger.Info("历史K线调度器已停止")
}

// RefreshSymbols reloads enabled symbols and triggers initial fetch.
func (s *HistoryKlineScheduler) RefreshSymbols() error {
	symbols, symbolMap, err := loadEnabledSymbolMap()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.symbols = symbols
	s.symbolMap = symbolMap
	s.mu.Unlock()

	if s.running {
		go s.fetchAllHistoryKlines()
	}
	logger.Infof("历史K线订阅已刷新，共%d个交易对", len(symbolMap))
	return nil
}

func (s *HistoryKlineScheduler) startPolling() {
	// 增加轮询间隔以避免API限制
	ticker5min := time.NewTicker(2 * time.Minute)
	ticker15min := time.NewTicker(5 * time.Minute)
	ticker60min := time.NewTicker(15 * time.Minute)
	ticker4hour := time.NewTicker(1 * time.Hour)
	ticker1day := time.NewTicker(2 * time.Hour)

	defer ticker5min.Stop()
	defer ticker15min.Stop()
	defer ticker60min.Stop()
	defer ticker4hour.Stop()
	defer ticker1day.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker5min.C:
			go s.fetchKlineForPeriod("5min", 100)
		case <-ticker15min.C:
			go s.fetchKlineForPeriod("15min", 100)
		case <-ticker60min.C:
			go s.fetchKlineForPeriod("60min", 100)
		case <-ticker4hour.C:
			go s.fetchKlineForPeriod("4hour", 100)
		case <-ticker1day.C:
			go s.fetchKlineForPeriod("1day", 100)
		}
	}
}

func (s *HistoryKlineScheduler) fetchAllHistoryKlines() {
	logger.Info("开始获取所有交易对历史K线数据...")

	s.mu.RLock()
	symbols := s.symbols
	s.mu.RUnlock()

	for _, symbol := range symbols {
		for _, period := range KlinePeriodsPolling {
			s.fetchAndStoreKline(symbol, period, 300)
			// 增加延迟以避免API限制
			time.Sleep(1 * time.Second)
		}
	}

	logger.Info("所有交易对历史K线数据获取完成")
}

func (s *HistoryKlineScheduler) fetchKlineForPeriod(period string, size int) {
	s.mu.RLock()
	symbols := s.symbols
	s.mu.RUnlock()

	for _, symbol := range symbols {
		s.fetchAndStoreKline(symbol, period, size)
		// 根据周期长度设置不同延迟，避免API限制
		switch period {
		case "5min", "15min":
			time.Sleep(500 * time.Millisecond)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func (s *HistoryKlineScheduler) fetchAndStoreKline(symbol, period string, size int) {
	klines, err := s.FetchKlineFromAPI(symbol, period, size)
	if err != nil {
		logger.Errorf("获取%s %s K线失败: %v", symbol, period, err)
		return
	}

	if len(klines) == 0 {
		logger.Debugf("获取%s %s K线数据为空", symbol, period)
		return
	}

	s.storeToRedis(symbol, period, klines)
	s.storeToDatabase(symbol, period, klines)
}

// FetchKlineFromAPI fetches klines from Huobi API.
func (s *HistoryKlineScheduler) FetchKlineFromAPI(symbol, period string, size int) ([]HuobiKline, error) {
	url := fmt.Sprintf("%s/market/history/kline?symbol=%s&period=%s&size=%d",
		s.baseURL, symbol, period, size)

	// 实现指数退避重试机制
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := s.httpClient.Get(url)
		if err != nil {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("HTTP请求失败: %w", err)
			}
			// 指数退避等待
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close() // 确保关闭响应体
			if i == maxRetries-1 {
				return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
			}
			// 指数退避等待
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close() // 确保关闭响应体
			if i == maxRetries-1 {
				return nil, fmt.Errorf("读取响应失败: %w", err)
			}
			// 指数退避等待
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		var result HuobiKlineResponse
		if err := json.Unmarshal(body, &result); err != nil {
			resp.Body.Close() // 确保关闭响应体
			if i == maxRetries-1 {
				return nil, fmt.Errorf("解析JSON失败: %w", err)
			}
			// 指数退避等待
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		if result.Status != "ok" {
			resp.Body.Close() // 确保关闭响应体
			if i == maxRetries-1 {
				return nil, fmt.Errorf("API返回错误状态: %s", result.Status)
			}
			// 指数退避等待
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		return result.Data, nil
	}

	return nil, fmt.Errorf("达到最大重试次数")
}

func (s *HistoryKlineScheduler) storeToRedis(symbol, period string, klines []HuobiKline) {
	key := fmt.Sprintf("kline:%s:%s", symbol, period)
	cache.RDB.Del(cache.Ctx, key)

	for _, kline := range klines {
		data := KlineData{
			Symbol:    symbol,
			Period:    period,
			ID:        kline.ID,
			Open:      kline.Open,
			Close:     kline.Close,
			Low:       kline.Low,
			High:      kline.High,
			Amount:    kline.Amount,
			Vol:       kline.Vol,
			Count:     kline.Count,
			Timestamp: kline.ID * 1000,
		}
		jsonData, _ := json.Marshal(data)
		cache.RPush(key, string(jsonData))
	}

	cache.Expire(key, 2*time.Hour)

	s.mu.RLock()
	info, ok := s.symbolMap[symbol]
	s.mu.RUnlock()

	if ok {
		key2 := fmt.Sprintf("market:kline:%d:%d:%s", info.CurrencyID, info.LegalID, period)
		cache.RDB.Del(cache.Ctx, key2)

		for _, kline := range klines {
			data := KlineData{
				Symbol:    symbol,
				Period:    period,
				ID:        kline.ID,
				Open:      kline.Open,
				Close:     kline.Close,
				Low:       kline.Low,
				High:      kline.High,
				Amount:    kline.Amount,
				Vol:       kline.Vol,
				Count:     kline.Count,
				Timestamp: kline.ID * 1000,
			}
			jsonData, _ := json.Marshal(data)
			cache.RPush(key2, string(jsonData))
		}
		cache.Expire(key2, 2*time.Hour)
	}
}

func (s *HistoryKlineScheduler) storeToDatabase(symbol, period string, klines []HuobiKline) {
	s.mu.RLock()
	info, ok := s.symbolMap[symbol]
	s.mu.RUnlock()
	if !ok {
		return
	}

	periodTypeMap := map[string]int8{
		"5min":  6,
		"15min": 1,
		"30min": 7,
		"60min": 2,
		"4hour": 11,
		"1day":  4,
		"1week": 8,
		"1mon":  9,
	}
	typeVal := periodTypeMap[period]
	if typeVal == 0 {
		typeVal = 5
	}

	for _, kline := range klines {
		signTime := fmt.Sprintf("%d", kline.ID)

		marketHour := model.MarketHour{
			CurrencyID: info.CurrencyID,
			LegalID:    info.LegalID,
			Type:       typeVal,
			Period:     period,
			Open:       kline.Open,
			Close:      kline.Close,
			High:       kline.High,
			Low:        kline.Low,
			Volume:     kline.Vol,
			Timestamp:  kline.ID,
			SignTime:   signTime,
			Sign:       2,
		}

		// 使用 Upsert 操作来提高性能
		var existing model.MarketHour
		result := database.DB.Where("currency_id = ? AND legal_id = ? AND period = ? AND day_time = ?",
			info.CurrencyID, info.LegalID, period, kline.ID).First(&existing)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Errorf("查询市场数据失败: %v", result.Error)
			continue
		}

		if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 记录不存在，创建新记录
			if err := database.DB.Create(&marketHour).Error; err != nil {
				logger.Errorf("创建市场数据失败: %v", err)
			}
		} else {
			// 记录存在，更新现有记录
			updates := map[string]interface{}{
				"start_price": kline.Open,
				"end_price":   kline.Close,
				"highest":     kline.High,
				"mminimum":    kline.Low,
				"number":      kline.Vol,
				"day_time":    kline.ID,
			}
			if err := database.DB.Model(&existing).Updates(updates).Error; err != nil {
				logger.Errorf("更新市场数据失败: %v", err)
			}
		}
	}
}

// GetHistoryKline returns historical klines, preferring Redis cache.
func (s *HistoryKlineScheduler) GetHistoryKline(symbol, period string, size int) ([]KlineData, error) {
	klines, err := GetKlineFromRedis(symbol, period, size)
	if err == nil && len(klines) > 0 {
		return klines, nil
	}

	huobiKlines, err := s.FetchKlineFromAPI(symbol, period, size)
	if err != nil {
		logger.Errorf("从火币API获取%s %s K线数据失败: %v", symbol, period, err)
		return nil, err
	}

	s.storeToRedis(symbol, period, huobiKlines)

	result := make([]KlineData, len(huobiKlines))
	for i, k := range huobiKlines {
		result[i] = KlineData{
			Symbol:    symbol,
			Period:    period,
			ID:        k.ID,
			Open:      k.Open,
			Close:     k.Close,
			Low:       k.Low,
			High:      k.High,
			Amount:    k.Amount,
			Vol:       k.Vol,
			Count:     k.Count,
			Timestamp: k.ID * 1000,
		}
	}
	return result, nil
}

var globalHistoryScheduler *HistoryKlineScheduler
var historySchedulerOnce sync.Once

// GetHistoryKlineScheduler returns the global scheduler.
func GetHistoryKlineScheduler() *HistoryKlineScheduler {
	historySchedulerOnce.Do(func() {
		globalHistoryScheduler = NewHistoryKlineScheduler()
	})
	return globalHistoryScheduler
}

// GetKlineByPeriod fetches klines from Redis or DB by period.
func GetKlineByPeriod(currencyID, legalID uint, period string, limit int) ([]KlineData, error) {
	key := fmt.Sprintf("market:kline:%d:%d:%s", currencyID, legalID, period)
	results, err := cache.LRange(key, 0, int64(limit-1))
	if err == nil && len(results) > 0 {
		var klines []KlineData
		for _, item := range results {
			var kline KlineData
			if err := json.Unmarshal([]byte(item), &kline); err != nil {
				continue
			}
			klines = append(klines, kline)
		}
		if len(klines) > 0 {
			return klines, nil
		}
	}

	var marketHours []model.MarketHour
	err = database.DB.Where("currency_id = ? AND legal_id = ? AND period = ?",
		currencyID, legalID, period).
		Order("day_time DESC").
		Limit(limit).
		Find(&marketHours).Error
	if err != nil {
		return nil, err
	}

	klines := make([]KlineData, len(marketHours))
	for i, mh := range marketHours {
		klines[i] = KlineData{
			Period:    mh.Period,
			ID:        mh.Timestamp,
			Open:      mh.Open,
			Close:     mh.Close,
			Low:       mh.Low,
			High:      mh.High,
			Vol:       mh.Volume,
			Timestamp: mh.Timestamp * 1000,
		}
	}

	return klines, nil
}
