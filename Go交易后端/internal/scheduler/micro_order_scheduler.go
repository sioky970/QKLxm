package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/pkg/utils"
	"exchange-go/internal/service"
	"exchange-go/internal/websocket"

	"gorm.io/gorm"
)

// ============================================================================
// 秒合约自动结算调度器
// ============================================================================

// MicroOrderScheduler 秒合约自动结算调度器
type MicroOrderScheduler struct {
	riskEngine *service.RiskEngine
	running    uint32 // Use atomic operations
	stopChan   chan struct{}
	startOnce  sync.Once
	stopOnce   sync.Once
}

// NewMicroOrderScheduler 创建秒合约调度器实例
func NewMicroOrderScheduler() *MicroOrderScheduler {
	return &MicroOrderScheduler{
		riskEngine: service.NewRiskEngine(),
		stopChan:   make(chan struct{}),
	}
}

// Start 启动自动结算任务
func (s *MicroOrderScheduler) Start() {
	// Use atomic to check if already running
	if !atomic.CompareAndSwapUint32(&s.running, 0, 1) {
		logger.Warn("秒合约自动结算任务已在运行中")
		return
	}

	logger.Info("启动秒合约自动结算任务...")

	// 每秒检查一次待结算订单
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		atomic.StoreUint32(&s.running, 0)
		logger.Info("秒合约自动结算任务已完全停止")
	}()

	for {
		select {
		case <-ticker.C:
			s.processExpiredOrders()
		case <-s.stopChan:
			logger.Info("收到停止信号，秒合约自动结算任务正在停止...")
			return
		}
	}
}

// Stop 停止自动结算任务
func (s *MicroOrderScheduler) Stop() {
	s.stopOnce.Do(func() {
		if atomic.LoadUint32(&s.running) == 1 {
			close(s.stopChan)
			logger.Info("发送停止信号到秒合约自动结算任务")
		}
	})
}

// processExpiredOrders 处理到期订单
func (s *MicroOrderScheduler) processExpiredOrders() {
	now := time.Now()

	// 查找所有状态为0(进行中)且已到期的订单
	var expiredOrders []model.MicroOrder
	result := database.DB.Where("status = ? AND DATE_ADD(created_at, INTERVAL seconds SECOND) <= ?", 0, now).
		Find(&expiredOrders)

	if result.Error != nil {
		logger.Errorf("查询到期秒合约订单失败: %v", result.Error)
		return
	}

	if len(expiredOrders) == 0 {
		return
	}

	logger.Infof("发现 %d 个到期订单，开始结算...", len(expiredOrders))

	successCount := 0
	failCount := 0

	for _, order := range expiredOrders {
		// 二次检查订单是否真的到期
		expectedEndTime := order.CreatedAt.Add(time.Duration(order.Seconds) * time.Second)
		if now.Before(expectedEndTime) {
			continue
		}

		// 执行结算
		if err := s.settleOrder(&order); err != nil {
			logger.Errorf("订单结算失败: orderID=%d, error=%v", order.ID, err)
			failCount++
		} else {
			successCount++
		}
	}

	if successCount > 0 || failCount > 0 {
		logger.Infof("秒合约结算完成: 成功=%d, 失败=%d", successCount, failCount)
	}
}

// settleOrder 结算单个订单
func (s *MicroOrderScheduler) settleOrder(order *model.MicroOrder) error {
	// 再次检查订单状态（防止重复结算）
	var currentOrder model.MicroOrder
	if err := database.DB.First(&currentOrder, order.ID).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}
	if currentOrder.Status != 0 {
		// 已结算，跳过
		return nil
	}

	// 获取当前价格（用于记录结算价格）
	currentPrice, err := s.getCurrentPrice(order.CurrencyID)
	if err != nil {
		return fmt.Errorf("获取当前价格失败: %w", err)
	}

	// 计算风控结果
	// 优先级: 账户风控 > 币种风控 > 全局风控 > 默认50%概率
	var profitResult int8
	logger.Infof("[settleOrder] 开始结算订单: orderID=%d, userID=%d, PreProfitResult=%d", order.ID, order.UserID, order.PreProfitResult)

	if order.PreProfitResult != 0 {
		// 已有预设结果，直接使用
		profitResult = order.PreProfitResult
		logger.Infof("[settleOrder] 使用订单预设风控结果: orderID=%d, profitResult=%d", order.ID, profitResult)
	} else {
		// 通过风控引擎计算
		logger.Infof("[settleOrder] 通过风控引擎计算结果: orderID=%d", order.ID)
		riskResult := s.riskEngine.CalculateMicroRiskResult(order)
		profitResult = int8(riskResult)
		logger.Infof("[settleOrder] 风控引擎计算完成: orderID=%d, profitResult=%d", order.ID, profitResult)
	}

	// 计算盈亏金额
	var factProfits float64
	if profitResult == 1 { // 盈利
		factProfits = order.Number * order.ProfitRatio
	} else { // 亏损
		factProfits = -order.Number
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 更新订单状态
	updateData := map[string]interface{}{
		"status":        1, // 1=已结算
		"end_price":     currentPrice,
		"fact_profits":  factProfits,
		"profit_result": profitResult,
		"updated_at":    now,
	}

	if err := tx.Model(&model.MicroOrder{}).Where("id = ? AND status = 0", order.ID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 获取法币ID (USDT)
	legalID, err := service.GetCurrencyService().GetDefaultLegalCurrencyID()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取默认计价币失败: %w", err)
	}

	// 记录账户日志
	if err := s.recordAccountLog(tx, order, currentPrice, factProfits, profitResult, legalID); err != nil {
		tx.Rollback()
		return fmt.Errorf("记录账户日志失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新 UserAssets 表（不再使用UsersWallet）
	// 正确的结算逻辑：
	// - 盈利：解锁本金（从锁定移到可用），然后额外增加盈利
	// - 亏损：只减少锁定余额（本金损失）
	// 使用合约钱包进行结算
	if profitResult == 1 { // 盈利
		// 解锁本金（从锁定移到可用）
		if err := service.GetUserAssetsService().UnlockContractUsdt(order.UserID, order.Number); err != nil {
			logger.Errorf("[settleOrder] 解锁合约钱包USDT本金失败: orderID=%d, userID=%d, amount=%.8f, err=%v",
				order.ID, order.UserID, order.Number, err)
		} else {
			// 额外增加盈利到可用余额
			profitAmount := order.Number * order.ProfitRatio
			if err := service.GetUserAssetsService().UpdateContractUsdtBalance(order.UserID, profitAmount, false); err != nil {
				logger.Errorf("[settleOrder] 增加盈利到合约钱包可用余额失败: orderID=%d, userID=%d, profit=%.8f, err=%v",
					order.ID, order.UserID, profitAmount, err)
			}
		}
	} else { // 亏损
		// 直接减少锁定余额（本金损失）
		if err := service.GetUserAssetsService().UpdateContractUsdtBalance(order.UserID, -order.Number, true); err != nil {
			logger.Errorf("[settleOrder] 减少合约钱包锁定余额（亏损）失败: orderID=%d, userID=%d, amount=%.8f, err=%v",
				order.ID, order.UserID, order.Number, err)
		}
	}

	resultStr := "亏损"
	if profitResult == 1 {
		resultStr = "盈利"
	}

	logger.Infof("订单结算成功: orderID=%d, userID=%d, openPrice=%.8f, endPrice=%.8f, result=%s, profit=%.2f",
		order.ID, order.UserID, order.OpenPrice, currentPrice, resultStr, factProfits)

	// WebSocket推送结算结果 - 使用安全goroutine
	utils.SafeGoWithName("broadcastSettlement", func() {
		s.broadcastSettlement(order.UserID, order.ID, profitResult, factProfits, currentPrice)
	})

	// WebSocket推送余额变更 - 使用安全goroutine
	utils.SafeGoWithName("broadcastBalanceUpdate", func() {
		service.GetWalletService().BroadcastBalanceUpdate(order.UserID)
	})

	return nil
}

// getCurrentPrice 获取当前价格
func (s *MicroOrderScheduler) getCurrentPrice(currencyID uint) (float64, error) {
	legalID, err := service.GetCurrencyService().GetDefaultLegalCurrencyID()
	if err != nil {
		return 0, err
	}

	var quotation model.CurrencyQuotation
	err = database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).
		First(&quotation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有行情数据，尝试从市场数据获取
			var marketHour model.MarketHour
			err = database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).
				Order("timestamp DESC").
				First(&marketHour).Error
			if err != nil {
				return 0, fmt.Errorf("未找到价格数据")
			}
			return marketHour.Close, nil
		}
		return 0, err
	}

	return quotation.Price, nil
}

// updateUserWallet 更新用户钱包余额（已迁移到UserAssets）
// 注意：此方法不再更新UsersWallet，只返回计算后的返还金额，由调用方更新UserAssets
func (s *MicroOrderScheduler) updateUserWallet(tx *gorm.DB, order *model.MicroOrder, legalID uint, profitResult int8) (float64, error) {
	// 计算返还金额
	// 投入金额 = order.Number (无手续费)
	var returnAmount float64

	if profitResult == 1 { // 盈利
		// 返还本金 + 盈利
		returnAmount = order.Number + (order.Number * order.ProfitRatio)
	} else { // 亏损 (profitResult == -1)
		// 亏损不返还本金
		returnAmount = 0
	}

	// 不再更新UsersWallet，只返回返还金额
	// UserAssets的更新由调用方处理
	return returnAmount, nil
}

// recordAccountLog 记录账户日志
func (s *MicroOrderScheduler) recordAccountLog(tx *gorm.DB, order *model.MicroOrder, endPrice, profit float64, profitResult int8, legalID uint) error {
	resultStr := "亏损"
	if profitResult == 1 {
		resultStr = "盈利"
	}

	dirStr := "买涨"
	if order.Type == 2 {
		dirStr = "买跌"
	}

	logType := 301 // 秒合约结算
	info := fmt.Sprintf("秒合约结算[%s]: 订单ID=%d, 方向=%s, 周期=%ds, 投入=%.2f, 开盘价=%.8f, 结算价=%.8f, 盈亏=%.2f",
		resultStr, order.ID, dirStr, order.Seconds, order.Number, order.OpenPrice, endPrice, profit)

	accountLog := &model.AccountLog{
		UserID:      order.UserID,
		Value:       profit,
		Info:        info,
		Type:        logType,
		CurrencyID:  legalID,
		CreatedTime: time.Now().Unix(),
	}

	return tx.Create(accountLog).Error
}

// broadcastSettlement WebSocket推送结算结果
func (s *MicroOrderScheduler) broadcastSettlement(userID, orderID uint, profitResult int8, profit, endPrice float64) {
	resultStr := "loss"
	if profitResult == 1 {
		resultStr = "profit"
	}

	message := map[string]interface{}{
		"type": "micro_order_settled",
		"data": map[string]interface{}{
			"order_id":      orderID,
			"result":        resultStr,
			"profit_result": profitResult,
			"profit":        profit,
			"end_price":     endPrice,
			"settle_time":   time.Now().Format(time.RFC3339),
		},
	}

	// 推送到 micro:用户ID 频道
	channelName := fmt.Sprintf("micro:%d", userID)
	hub := websocket.GetHub()
	hub.SendToChannel(channelName, message)
}

// GetPendingOrdersCount 获取待结算订单数量
func (s *MicroOrderScheduler) GetPendingOrdersCount() (int64, error) {
	var count int64
	err := database.DB.Model(&model.MicroOrder{}).Where("status = 0").Count(&count).Error
	return count, err
}
