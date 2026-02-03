package scheduler

import (
	"sync"
	"sync/atomic"
	"time"

	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/service"
)

// ContractScheduler 合约交易定时任务调度器
type ContractScheduler struct {
	futuresService *service.FuturesTradingService
	running        uint32 // Use atomic operations
	stopChan       chan struct{}
	startOnce      sync.Once
	stopOnce       sync.Once
}

// NewContractScheduler 创建合约调度器实例
func NewContractScheduler() *ContractScheduler {
	return &ContractScheduler{
		futuresService: service.GetFuturesTradingService(),
		stopChan:       make(chan struct{}),
	}
}

// Start 启动合约交易定时任务
func (s *ContractScheduler) Start() {
	// Use atomic to check if already running
	if !atomic.CompareAndSwapUint32(&s.running, 0, 1) {
		logger.Warn("合约交易定时任务已在运行中")
		return
	}

	logger.Info("启动合约交易定时任务...")

	// 爆仓检测: 每2秒检查一次
	liquidationTicker := time.NewTicker(2 * time.Second)

	// 止盈止损检测: 每2秒检查一次
	tpslTicker := time.NewTicker(2 * time.Second)

	// 限价单检测: 每1秒检查一次
	limitOrderTicker := time.NewTicker(1 * time.Second)

	defer func() {
		liquidationTicker.Stop()
		tpslTicker.Stop()
		limitOrderTicker.Stop()
		atomic.StoreUint32(&s.running, 0)
		logger.Info("合约交易定时任务已完全停止")
	}()

	for {
		select {
		case <-liquidationTicker.C:
			s.checkLiquidations()
		case <-tpslTicker.C:
			s.checkTPSL()
		case <-limitOrderTicker.C:
			s.checkLimitOrders()
		case <-s.stopChan:
			logger.Info("收到停止信号，合约交易定时任务正在停止...")
			return
		}
	}
}

// Stop 停止定时任务
func (s *ContractScheduler) Stop() {
	s.stopOnce.Do(func() {
		if atomic.LoadUint32(&s.running) == 1 {
			close(s.stopChan)
			logger.Info("发送停止信号到合约交易定时任务")
		}
	})
}

// checkLiquidations 检查爆仓条件
func (s *ContractScheduler) checkLiquidations() {
	if err := s.futuresService.CheckAllLiquidations(); err != nil {
		logger.Errorf("爆仓检测任务执行失败: %v", err)
	}
}

// checkTPSL 检查止盈止损条件
func (s *ContractScheduler) checkTPSL() {
	if err := s.futuresService.CheckAllTPSL(); err != nil {
		logger.Errorf("止盈止损检测任务执行失败: %v", err)
	}
}

// checkLimitOrders 检查限价单成交条件
func (s *ContractScheduler) checkLimitOrders() {
	if err := s.futuresService.CheckAllLimitOrders(); err != nil {
		logger.Errorf("限价单检测任务执行失败: %v", err)
	}
}

// IsRunning 检查是否在运行
func (s *ContractScheduler) IsRunning() bool {
	return atomic.LoadUint32(&s.running) == 1
}
