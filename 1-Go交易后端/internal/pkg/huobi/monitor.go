package huobi

import (
	"sync"
	"time"

	"exchange-go/internal/pkg/logger"
)

// Monitor 负责监控火币数据获取的状态
type Monitor struct {
	manager        *Manager
	scheduler      *HistoryKlineScheduler
	status         map[string]*SymbolStatus
	mutex          sync.RWMutex
	lastUpdateTime map[string]time.Time
}

// SymbolStatus 交易对状态信息
type SymbolStatus struct {
	Symbol           string
	LastKlineTime    int64
	LastTickerTime   int64
	ErrorCount       int
	LastError        string
	IsActive         bool
	LastUpdateTime   time.Time
}

// NewMonitor 创建新的监控实例
func NewMonitor(m *Manager, s *HistoryKlineScheduler) *Monitor {
	return &Monitor{
		manager:        m,
		scheduler:      s,
		status:         make(map[string]*SymbolStatus),
		lastUpdateTime: make(map[string]time.Time),
	}
}

// Start 启动监控
func (mon *Monitor) Start() {
	go mon.runHealthCheck()
	go mon.runStatusReport()
}

// runHealthCheck 运行健康检查
func (mon *Monitor) runHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mon.checkSymbolStatus()
		}
	}
}

// checkSymbolStatus 检查交易对状态
func (mon *Monitor) checkSymbolStatus() {
	mon.manager.mu.RLock()
	symbols := mon.manager.symbolMap
	mon.manager.mu.RUnlock()

	for symbol, info := range symbols {
		// 检查Redis中的最新K线数据
		klines, err := GetKlineFromRedis(symbol, "1min", 1)
		if err != nil || len(klines) == 0 {
			// 尝试从数据库获取最新数据
			dbKlines, err := GetKlineByPeriod(info.CurrencyID, info.LegalID, "1min", 1)
			if err != nil || len(dbKlines) == 0 {
				mon.updateSymbolStatus(symbol, false, "没有获取到K线数据")
				continue
			}
		}

		// 检查行情数据
		tickerData, err := GetTickerFromRedis(symbol)
		if err != nil || tickerData == nil {
			mon.updateSymbolStatus(symbol, false, "没有获取到行情数据")
			continue
		}

		// 更新状态
		mon.updateSymbolStatus(symbol, true, "")
	}
}

// updateSymbolStatus 更新交易对状态
func (mon *Monitor) updateSymbolStatus(symbol string, isActive bool, lastError string) {
	mon.mutex.Lock()
	defer mon.mutex.Unlock()

	status, exists := mon.status[symbol]
	if !exists {
		status = &SymbolStatus{
			Symbol: symbol,
		}
		mon.status[symbol] = status
	}

	status.IsActive = isActive
	status.LastError = lastError
	status.LastUpdateTime = time.Now()
	if !isActive {
		status.ErrorCount++
	} else {
		status.ErrorCount = 0
	}

	mon.lastUpdateTime[symbol] = time.Now()
}

// runStatusReport 定期报告状态
func (mon *Monitor) runStatusReport() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mon.reportStatus()
		}
	}
}

// reportStatus 报告当前状态
func (mon *Monitor) reportStatus() {
	mon.mutex.RLock()
	defer mon.mutex.RUnlock()

	totalSymbols := len(mon.status)
	activeSymbols := 0
	inactiveSymbols := 0

	for _, status := range mon.status {
		if status.IsActive {
			activeSymbols++
		} else {
			inactiveSymbols++
		}
	}

	logger.Infof("火币数据监控报告: 总计%d个交易对, 活跃%d个, 不活跃%d个", totalSymbols, activeSymbols, inactiveSymbols)

	if inactiveSymbols > 0 {
		logger.Warnf("以下交易对处于非活跃状态:")
		for _, status := range mon.status {
			if !status.IsActive {
				logger.Warnf("- %s: %s (错误次数: %d)", status.Symbol, status.LastError, status.ErrorCount)
			}
		}
	}
}

// GetSymbolStatus 获取特定交易对状态
func (mon *Monitor) GetSymbolStatus(symbol string) *SymbolStatus {
	mon.mutex.RLock()
	defer mon.mutex.RUnlock()

	if status, exists := mon.status[symbol]; exists {
		return status
	}
	return nil
}

// GetAllStatuses 获取所有交易对状态
func (mon *Monitor) GetAllStatuses() map[string]*SymbolStatus {
	mon.mutex.RLock()
	defer mon.mutex.RUnlock()

	result := make(map[string]*SymbolStatus, len(mon.status))
	for k, v := range mon.status {
		result[k] = v
	}
	return result
}