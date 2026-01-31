package service

import (
	"sync"
	"sync/atomic"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/pkg/utils"
)

// MatchingEngine 价格监控撮合引擎（做市商模式）
// 用户与平台直接兑换，不存在买卖对手盘
// - 市价单：立即按当前市场价成交
// - 限价单：监控市场价格，达到目标价时自动成交
type MatchingEngine struct {
	running   uint32 // Use atomic operations
	stopChan  chan struct{}
	mu        sync.Mutex
	startOnce sync.Once
	stopOnce  sync.Once
}

var (
	matchingEngine     *MatchingEngine
	matchingEngineOnce sync.Once
)

// GetMatchingEngine 获取撮合引擎单例
func GetMatchingEngine() *MatchingEngine {
	matchingEngineOnce.Do(func() {
		matchingEngine = &MatchingEngine{
			stopChan: make(chan struct{}),
		}
	})
	return matchingEngine
}

// Start 启动价格监控任务
func (e *MatchingEngine) Start() {
	// Use atomic to check if already running
	if !atomic.CompareAndSwapUint32(&e.running, 0, 1) {
		logger.Warn("[MatchingEngine] 价格监控任务已在运行中")
		return
	}

	logger.Info("[MatchingEngine] 启动价格监控任务（每秒检查限价单）...")

	// 每秒检查一次限价单是否满足成交条件
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		atomic.StoreUint32(&e.running, 0)
		logger.Info("[MatchingEngine] 价格监控任务已完全停止")
	}()

	for {
		select {
		case <-ticker.C:
			e.checkPendingOrders()
		case <-e.stopChan:
			logger.Info("[MatchingEngine] 收到停止信号，价格监控任务正在停止...")
			return
		}
	}
}

// Stop 停止价格监控任务
func (e *MatchingEngine) Stop() {
	e.stopOnce.Do(func() {
		if atomic.LoadUint32(&e.running) == 1 {
			close(e.stopChan)
			logger.Info("[MatchingEngine] 发送停止信号到价格监控任务")
		}
	})
}

// checkPendingOrders 检查所有待成交的限价单
func (e *MatchingEngine) checkPendingOrders() {
	// 查询所有待成交的限价单（status=0）
	var pendingOrders []model.Transaction
	if err := database.DB.Where("status = 0").Find(&pendingOrders).Error; err != nil {
		logger.Errorf("[MatchingEngine] 查询待成交订单失败: %v", err)
		return
	}

	if len(pendingOrders) == 0 {
		return
	}

	// 逐个检查订单是否满足成交条件
	for _, order := range pendingOrders {
		// 获取当前市场价格
		marketPrice, err := e.getMarketPrice(uint(order.Currency), order.LegalID)
		if err != nil {
			continue // 获取价格失败，跳过
		}

		// 检查是否满足成交条件
		if e.shouldFill(&order, marketPrice) {
			logger.Infof("[MatchingEngine] 订单%d满足成交条件: 订单价格=%.8f, 市场价格=%.8f",
				order.ID, order.Price, marketPrice)

			// 执行成交（使用市场价格）
			e.executeOrder(&order, marketPrice)
		}
	}
}

// shouldFill 判断订单是否满足成交条件
func (e *MatchingEngine) shouldFill(order *model.Transaction, marketPrice float64) bool {
	if order.Type == 1 { // 买入：市场价 <= 限价时成交（用户希望低价买入）
		return marketPrice <= order.Price
	} else { // 卖出：市场价 >= 限价时成交（用户希望高价卖出）
		return marketPrice >= order.Price
	}
}

// getMarketPrice 获取当前市场价格
func (e *MatchingEngine) getMarketPrice(currencyID, legalID uint) (float64, error) {
	// 优先从行情表获取最新价格
	var quotation model.CurrencyQuotation
	err := database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).
		First(&quotation).Error

	if err == nil && quotation.Price > 0 {
		return quotation.Price, nil
	}

	// 备用：从小时K线获取收盘价
	var marketHour model.MarketHour
	err = database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).
		Order("timestamp DESC").
		First(&marketHour).Error

	if err == nil && marketHour.Close > 0 {
		return marketHour.Close, nil
	}

	return 0, err
}

// ExecuteMarketOrder 执行市价单（立即成交）
// 调用时机：用户提交市价单时
func (e *MatchingEngine) ExecuteMarketOrder(order *model.Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 获取当前市场价格
	marketPrice, err := e.getMarketPrice(uint(order.Currency), order.LegalID)
	if err != nil || marketPrice <= 0 {
		logger.Errorf("[MatchingEngine] 获取市场价格失败，无法执行市价单: %v", err)
		return err
	}

	logger.Infof("[MatchingEngine] 执行市价单: orderID=%d, 市场价格=%.8f", order.ID, marketPrice)
	return e.executeOrder(order, marketPrice)
}

// executeOrder 执行订单成交（核心逻辑）
func (e *MatchingEngine) executeOrder(order *model.Transaction, fillPrice float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 计算成交金额
	tradeAmount := order.Number * fillPrice

	// 获取币种名称
	currency, err := GetCurrencyService().GetCurrencyByID(uint(order.Currency))
	if err != nil {
		tx.Rollback()
		logger.Errorf("[MatchingEngine] 获取币种信息失败: %v", err)
		return err
	}
	currencyName := currency.Name

	// 获取用户钱包（用于验证用户存在）
	_, err = GetWalletService().GetWalletByCurrency(order.FromUserID, order.LegalID)
	if err != nil {
		tx.Rollback()
		logger.Errorf("[MatchingEngine] 获取钱包失败: %v", err)
		return err
	}

	if order.Type == 1 { // 买入：USDT → 加密货币
		// 1. 从锁定的USDT中扣除（使用UserAssets）
		if err := GetUserAssetsService().UpdateUsdtBalance(order.FromUserID, -tradeAmount, true); err != nil {
			tx.Rollback()
			logger.Errorf("[MatchingEngine] 扣除USDT锁定余额失败: %v", err)
			return err
		}
		// 增加币种余额
		if err := GetUserAssetsService().UpdateCurrencyBalance(order.FromUserID, currencyName, order.Number, false); err != nil {
			tx.Rollback()
			logger.Errorf("[MatchingEngine] 增加币种余额失败: %v", err)
			return err
		}

		logger.Infof("[MatchingEngine] 买入成交: orderID=%d, 花费USDT=%.8f, 获得%s=%.8f",
			order.ID, tradeAmount, currencyName, order.Number)

	} else { // 卖出：加密货币 → USDT
		// 1. 从锁定的币种中扣除，增加USDT可用余额（使用UserAssets）
		if err := GetUserAssetsService().UpdateCurrencyBalance(order.FromUserID, currencyName, -order.Number, true); err != nil {
			tx.Rollback()
			logger.Errorf("[MatchingEngine] 扣除币种锁定余额失败: %v", err)
			return err
		}
		if err := GetUserAssetsService().UpdateUsdtBalance(order.FromUserID, tradeAmount, false); err != nil {
			tx.Rollback()
			logger.Errorf("[MatchingEngine] 增加USDT余额失败: %v", err)
			return err
		}

		logger.Infof("[MatchingEngine] 卖出成交: orderID=%d, 卖出%s=%.8f, 获得USDT=%.8f",
			order.ID, currencyName, order.Number, tradeAmount)
	}

	// 2. 更新订单状态为已成交
	if err := tx.Model(&model.Transaction{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"status":      2, // 已成交
		"deal_number": order.Number,
		"price":       fillPrice, // 记录实际成交价格
	}).Error; err != nil {
		tx.Rollback()
		logger.Errorf("[MatchingEngine] 更新订单状态失败: %v", err)
		return err
	}

	// 3. 创建成交记录
	tradeRecord := &model.TransactionComplete{
		UserID:     order.FromUserID,
		FromUserID: 0, // 平台作为对手方
		Currency:   order.Currency,
		Legal:      int(order.LegalID),
		Price:      fillPrice,
		Number:     order.Number,
		Way:        int8(order.Type), // 1=买入, 2=卖出
		CreateTime: time.Now().Unix(),
	}
	if err := tx.Create(tradeRecord).Error; err != nil {
		tx.Rollback()
		logger.Errorf("[MatchingEngine] 创建成交记录失败: %v", err)
		return err
	}

	tx.Commit()

	// 事务提交后，同步更新 UserAssets 表（双写模式）
	if order.Type == 1 { // 买入：减少USDT锁定，增加币种可用
		GetUserAssetsService().UpdateUsdtBalance(order.FromUserID, -tradeAmount, true)
		GetUserAssetsService().UpdateCurrencyBalance(order.FromUserID, currencyName, order.Number, false)
	} else { // 卖出：减少币种锁定，增加USDT可用
		GetUserAssetsService().UpdateCurrencyBalance(order.FromUserID, currencyName, -order.Number, true)
		GetUserAssetsService().UpdateUsdtBalance(order.FromUserID, tradeAmount, false)
	}

	// WebSocket推送余额更新 - 使用安全goroutine
	utils.SafeGoWithName("broadcastBalanceUpdate", func() {
		GetWalletService().BroadcastBalanceUpdate(order.FromUserID)
	})

	// WebSocket推送订单状态变更 - 使用安全goroutine
	utils.SafeGoWithName("broadcastOrderUpdate", func() {
		GetWalletService().BroadcastOrderUpdate(order.FromUserID, order.ID, 2, order.Number, fillPrice)
	})

	logger.Infof("[MatchingEngine] 订单%d成交完成: %s %.8f @ %.8f USDT",
		order.ID, currencyName, order.Number, fillPrice)

	return nil
}

// MatchResult 撮合结果（保留兼容接口）
type MatchResult struct {
	Matched     bool
	FilledQty   float64
	FilledPrice float64
	FullFilled  bool
}

// TryMatch 尝试立即成交（用于市价单或满足条件的限价单）
func (e *MatchingEngine) TryMatch(order *model.Transaction) *MatchResult {
	result := &MatchResult{Matched: false}

	// 获取市场价格
	marketPrice, err := e.getMarketPrice(uint(order.Currency), order.LegalID)
	if err != nil || marketPrice <= 0 {
		logger.Warnf("[MatchingEngine] 订单%d: 无法获取市场价格，等待价格监控", order.ID)
		return result
	}

	// 检查是否满足成交条件
	if e.shouldFill(order, marketPrice) {
		if err := e.executeOrder(order, marketPrice); err == nil {
			result.Matched = true
			result.FilledQty = order.Number
			result.FilledPrice = marketPrice
			result.FullFilled = true
		}
	} else {
		logger.Infof("[MatchingEngine] 订单%d暂不满足成交条件: 订单价格=%.8f, 市场价格=%.8f, 等待价格变化",
			order.ID, order.Price, marketPrice)
	}

	return result
}
