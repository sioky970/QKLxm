package scheduler

import (
	"time"

	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/service"
)

// ContractScheduler 合约交易定时任务调度器
type ContractScheduler struct {
	futuresService *service.FuturesTradingService
	running        bool
	stopChan       chan bool
}

// NewContractScheduler 创建合约调度器实例
func NewContractScheduler() *ContractScheduler {
	return &ContractScheduler{
		futuresService: service.GetFuturesTradingService(),
		running:        false,
		stopChan:       make(chan bool),
	}
}

// Start 启动合约交易定时任务
func (s *ContractScheduler) Start() {
	if s.running {
		logger.Warn("合约交易定时任务已在运行中")
		return
	}

	s.running = true
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
			logger.Info("合约交易定时任务已停止")
			s.running = false
			return
		}
	}
}

// Stop 停止定时任务
func (s *ContractScheduler) Stop() {
	if s.running {
		s.stopChan <- true
	}
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
	return s.running
}
