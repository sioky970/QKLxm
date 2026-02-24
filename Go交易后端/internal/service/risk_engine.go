package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/types"

	"gorm.io/gorm"
)

// RiskEngine currency-based risk control (default quote: USDT).
type RiskEngine struct{}

// NewRiskEngine creates a new RiskEngine.
func NewRiskEngine() *RiskEngine {
	return &RiskEngine{}
}

// ApplyRiskControlForOrder applies risk control and returns a forced result when applicable.
func (e *RiskEngine) ApplyRiskControlForOrder(order *model.MicroOrder) (types.RiskResult, bool, error) {
	if order == nil {
		return types.RiskResultNone, false, errors.New("order is nil")
	}

	// 1) User-level risk control. (Priority 1)
	if order.UserID > 0 {
		userRisk, err := e.getUserRiskResult(order.UserID)
		if err != nil {
			return types.RiskResultNone, false, err
		}
		if userRisk != types.RiskResultNone {
			logger.Debugf("apply user risk: userID=%d, risk=%d", order.UserID, userRisk)
			return userRisk, true, nil
		}
	}

	// 2) Global Risk Control (Priority 2)
	riskConfig, err := NewRiskControlService().GetRiskConfig()
	if err == nil && riskConfig != nil {
		switch riskConfig.Mode {
		case RiskModeGroup: // 2: 群控
			logger.Debugf("apply global group risk: %d", riskConfig.GroupResult)
			return riskConfig.GroupResult, true, nil
		case RiskModeMoney: // 3: 金额控制
			// 查找匹配的金额规则
			for _, rule := range riskConfig.MoneyRules {
				if (rule.Min == 0 || order.Number >= rule.Min) && (rule.Max == 0 || order.Number <= rule.Max) {
					// 匹配到规则，按概率计算结果
					rand.Seed(time.Now().UnixNano())
					if rand.Intn(100) < rule.ProfitProbability {
						logger.Debugf("apply money risk (rule match): amount=%.2f, result=profit", order.Number)
						return types.RiskResultProfit, true, nil
					}
					logger.Debugf("apply money risk (rule match): amount=%.2f, result=loss", order.Number)
					return types.RiskResultLoss, true, nil
				}
			}
			// 未匹配到金额规则，使用默认胜率
			rand.Seed(time.Now().UnixNano())
			randNum := rand.Intn(100)
			if randNum < riskConfig.ProbabilityProfitPercent {
				logger.Debugf("apply global probability risk (default): result=profit")
				return types.RiskResultProfit, true, nil
			} else if randNum < riskConfig.ProbabilityProfitPercent+riskConfig.ProbabilityBalancePercent {
				logger.Debugf("apply global probability risk (default): result=none (balance)")
				return types.RiskResultNone, true, nil
			}
			logger.Debugf("apply global probability risk (default): result=loss")
			return types.RiskResultLoss, true, nil
		}
	}

	// 3) Currency-level risk control. (Priority 3)
	var currency model.Currency
	if err := database.DB.Where("id = ?", order.CurrencyID).First(&currency).Error; err != nil {
		return types.RiskResultNone, false, fmt.Errorf("load currency failed: %w", err)
	}

	// Time window risk.
	if currency.RiskTimeEnabled == 1 && isInTimeWindow(time.Now(), currency.RiskTimeStart, currency.RiskTimeEnd) {
		if currency.RiskTimeResult != 0 {
			logger.Debugf("apply currency time risk: currencyID=%d, result=%d", order.CurrencyID, currency.RiskTimeResult)
			return types.RiskResult(currency.RiskTimeResult), true, nil
		}
	}

	// Money range risk.
	if currency.RiskMoneyEnabled == 1 {
		amount := order.Number
		min := currency.RiskMoneyMin
		max := currency.RiskMoneyMax
		inRange := (min == 0 || amount >= min) && (max == 0 || amount <= max)
		if inRange && currency.RiskMoneyResult != 0 {
			logger.Debugf("apply currency money risk: currencyID=%d, result=%d", order.CurrencyID, currency.RiskMoneyResult)
			return types.RiskResult(currency.RiskMoneyResult), true, nil
		}
	}

	// Probability risk.
	if currency.RiskProbEnabled == 1 {
		profitProb := currency.RiskProfitProbability
		if profitProb < 0 {
			profitProb = 0
		} else if profitProb > 100 {
			profitProb = 100
		}
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(100) < profitProb {
			logger.Debugf("apply currency probability risk: currencyID=%d, result=profit", order.CurrencyID)
			return types.RiskResultProfit, true, nil
		}
		logger.Debugf("apply currency probability risk: currencyID=%d, result=loss", order.CurrencyID)
		return types.RiskResultLoss, true, nil
	}

	return types.RiskResultNone, false, nil
}

// ApplyRiskControl applies risk control in priority order: User > Global > Currency
func (e *RiskEngine) ApplyRiskControl(currencyID uint) error {
	return e.ApplyRiskControlForUser(0, currencyID)
}

// ApplyRiskControlForUser applies risk control in priority order: User > Global > Currency
func (e *RiskEngine) ApplyRiskControlForUser(userID, currencyID uint) error {
	order := &model.MicroOrder{
		UserID:     userID,
		CurrencyID: currencyID,
	}
	result, applied, err := e.ApplyRiskControlForOrder(order)
	if err != nil {
		return err
	}
	if applied && result == types.RiskResultLoss {
		return errors.New("forced_loss")
	}
	return nil
}
func applyFixedResult(orders *[]model.MicroOrder, result int8) int {
	if result == 0 {
		return 0
	}
	affected := 0
	for i := range *orders {
		order := &(*orders)[i]
		if order.PreProfitResult == 0 {
			order.PreProfitResult = result
			database.DB.Model(order).Update("pre_profit_result", result)
			affected++
		}
	}
	return affected
}

func applyMoneyResult(orders *[]model.MicroOrder, min, max float64, result int8) int {
	if result == 0 {
		return 0
	}
	affected := 0
	for i := range *orders {
		order := &(*orders)[i]
		if order.PreProfitResult != 0 {
			continue
		}
		if (min == 0 || order.Number >= min) && (max == 0 || order.Number <= max) {
			order.PreProfitResult = result
			database.DB.Model(order).Update("pre_profit_result", result)
			affected++
		}
	}
	return affected
}

func applyProbability(orders *[]model.MicroOrder, profitProbability int) int {
	if profitProbability <= 0 {
		return 0
	}
	if profitProbability > 100 {
		profitProbability = 100
	}
	affected := 0
	rand.Seed(time.Now().UnixNano())
	for i := range *orders {
		order := &(*orders)[i]
		if order.PreProfitResult != 0 {
			continue
		}
		if rand.Intn(100) < profitProbability {
			order.PreProfitResult = 1
		} else {
			order.PreProfitResult = -1
		}
		database.DB.Model(order).Update("pre_profit_result", order.PreProfitResult)
		affected++
	}
	return affected
}

func (e *RiskEngine) deduceOrderPrice(currencyID uint, orders []model.MicroOrder) {
	if len(orders) == 0 {
		return
	}

	legalID := getDefaultLegalID()
	if legalID == 0 {
		logger.Warn("default legal currency not found")
		return
	}

	var quotation model.CurrencyQuotation
	err := database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).
		First(&quotation).Error
	if err != nil {
		logger.Errorf("quotation not found: %v", err)
		return
	}

	currentPrice := quotation.Price
	if currentPrice <= 0 {
		logger.Warn("invalid current price, skip risk price deduction")
		return
	}

	for i := range orders {
		if orders[i].PreProfitResult == 0 {
			continue
		}
		targetPrice := calculateTargetPrice(currentPrice, orders[i].Type, types.RiskResult(orders[i].PreProfitResult))
		logger.Debugf(
			"risk order: orderID=%d, type=%d, preResult=%d, currentPrice=%.8f, targetPrice=%.8f",
			orders[i].ID, orders[i].Type, orders[i].PreProfitResult, currentPrice, targetPrice,
		)
	}
}

func calculateTargetPrice(currentPrice float64, orderType int8, expectedResult types.RiskResult) float64 {
	rand.Seed(time.Now().UnixNano())
	fluctuation := 0.000001 + rand.Float64()*0.0001

	if expectedResult == types.RiskResultProfit {
		if orderType == 1 {
			return currentPrice * (1 + fluctuation)
		}
		return currentPrice * (1 - fluctuation)
	}
	if expectedResult == types.RiskResultLoss {
		if orderType == 1 {
			return currentPrice * (1 - fluctuation)
		}
		return currentPrice * (1 + fluctuation)
	}
	return currentPrice
}

func isInTimeWindow(now time.Time, start, end string) bool {
	startMin, okStart := parseHHMM(start)
	endMin, okEnd := parseHHMM(end)
	if !okStart || !okEnd {
		return false
	}
	nowMin := now.Hour()*60 + now.Minute()
	if startMin <= endMin {
		return nowMin >= startMin && nowMin <= endMin
	}
	return nowMin >= startMin || nowMin <= endMin
}

func parseHHMM(value string) (int, bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// getUserRiskResult 获取用户级别的风控结果
func (e *RiskEngine) getUserRiskResult(userID uint) (types.RiskResult, error) {
	var user model.User
	err := database.DB.Select("risk").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return types.RiskResultNone, err
	}

	return types.RiskResult(user.Risk), nil
}

// getGlobalRiskResult 获取全局风控结果
func (e *RiskEngine) getGlobalRiskResult() (types.RiskResult, error) {
	var setting model.Setting
	err := database.DB.Where("key = ?", "risk_global_result").First(&setting).Error
	if err != nil {
		// 如果没有全局设置，默认返回无预设
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.RiskResultNone, nil
		}
		return types.RiskResultNone, err
	}

	var result int
	fmt.Sscanf(setting.Value, "%d", &result)
	return types.RiskResult(result), nil
}

func getDefaultLegalID() uint {
	var currency model.Currency
	if err := database.DB.Where("name = ?", "USDT").First(&currency).Error; err != nil {
		return 0
	}
	return currency.ID
}

// ============================================================================
// 秒合约专用风控接口
// ============================================================================

// MicroRiskResult 秒合约风控结果 (无平局)
type MicroRiskResult int8

const (
	MicroRiskNone   MicroRiskResult = 0  // 无预设
	MicroRiskProfit MicroRiskResult = 1  // 强制盈利
	MicroRiskLoss   MicroRiskResult = -1 // 强制亏损
)

// CalculateMicroRiskResult 计算秒合约风控结果
// 优先级: 账户风控 > 币种风控 > 全局风控 > 默认50%概率
// 返回值: 1=盈利, -1=亏损 (无平局)
func (e *RiskEngine) CalculateMicroRiskResult(order *model.MicroOrder) MicroRiskResult {
	if order == nil {
		logger.Warnf("[CalculateMicroRiskResult] order is nil, returning random result")
		return e.defaultRandom()
	}

	logger.Infof("[CalculateMicroRiskResult] 开始计算风控: orderID=%d, userID=%d, currencyID=%d", order.ID, order.UserID, order.CurrencyID)

	// 1) 账户风控 (Priority 1) - 检查用户的risk字段
	if order.UserID > 0 {
		userRisk, err := e.getUserRiskResult(order.UserID)
		logger.Infof("[CalculateMicroRiskResult] 账户风控查询: userID=%d, risk=%d, err=%v", order.UserID, userRisk, err)
		if err == nil && userRisk != types.RiskResultNone {
			logger.Infof("[CalculateMicroRiskResult] 应用账户风控: userID=%d, result=%d", order.UserID, userRisk)
			if userRisk == types.RiskResultProfit {
				return MicroRiskProfit
			}
			return MicroRiskLoss
		}
		if err != nil {
			logger.Warnf("[CalculateMicroRiskResult] 获取用户风控失败: userID=%d, err=%v", order.UserID, err)
		}
	}

	// 2) 币种风控 (Priority 2)
	if order.CurrencyID > 0 {
		currencyRisk := e.getCurrencyRiskResult(order)
		logger.Infof("[CalculateMicroRiskResult] 币种风控查询: currencyID=%d, result=%d", order.CurrencyID, currencyRisk)
		if currencyRisk != MicroRiskNone {
			logger.Infof("[CalculateMicroRiskResult] 应用币种风控: currencyID=%d, result=%d", order.CurrencyID, currencyRisk)
			return currencyRisk
		}
	}

	// 3) 全局风控 (Priority 3)
	globalRisk := e.getGlobalMicroRiskResult(order)
	logger.Infof("[CalculateMicroRiskResult] 全局风控查询: result=%d", globalRisk)
	if globalRisk != MicroRiskNone {
		logger.Infof("[CalculateMicroRiskResult] 应用全局风控: result=%d", globalRisk)
		return globalRisk
	}

	// 4) 默认: 50%概率随机盈亏
	result := e.defaultRandom()
	logger.Infof("[CalculateMicroRiskResult] 应用默认随机风控: result=%d", result)
	return result
}

// getCurrencyRiskResult 获取币种级别风控结果
func (e *RiskEngine) getCurrencyRiskResult(order *model.MicroOrder) MicroRiskResult {
	var currency model.Currency
	if err := database.DB.Where("id = ?", order.CurrencyID).First(&currency).Error; err != nil {
		return MicroRiskNone
	}

	// 时间窗口风控
	if currency.RiskTimeEnabled == 1 && isInTimeWindow(time.Now(), currency.RiskTimeStart, currency.RiskTimeEnd) {
		if currency.RiskTimeResult == 1 {
			return MicroRiskProfit
		} else if currency.RiskTimeResult == -1 {
			return MicroRiskLoss
		}
	}

	// 金额范围风控
	if currency.RiskMoneyEnabled == 1 {
		amount := order.Number
		min := currency.RiskMoneyMin
		max := currency.RiskMoneyMax
		inRange := (min == 0 || amount >= min) && (max == 0 || amount <= max)
		if inRange {
			if currency.RiskMoneyResult == 1 {
				return MicroRiskProfit
			} else if currency.RiskMoneyResult == -1 {
				return MicroRiskLoss
			}
		}
	}

	// 概率风控
	if currency.RiskProbEnabled == 1 {
		profitProb := currency.RiskProfitProbability
		if profitProb < 0 {
			profitProb = 0
		} else if profitProb > 100 {
			profitProb = 100
		}
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(100) < profitProb {
			return MicroRiskProfit
		}
		return MicroRiskLoss
	}

	return MicroRiskNone
}

// getGlobalMicroRiskResult 获取全局风控结果(秒合约专用)
func (e *RiskEngine) getGlobalMicroRiskResult(order *model.MicroOrder) MicroRiskResult {
	riskConfig, err := NewRiskControlService().GetRiskConfig()
	if err != nil || riskConfig == nil {
		return MicroRiskNone
	}

	switch riskConfig.Mode {
	case RiskModeGroup: // 群控模式
		if riskConfig.GroupResult == types.RiskResultProfit {
			return MicroRiskProfit
		} else if riskConfig.GroupResult == types.RiskResultLoss {
			return MicroRiskLoss
		}
	case RiskModeMoney: // 金额控制模式
		for _, rule := range riskConfig.MoneyRules {
			if (rule.Min == 0 || order.Number >= rule.Min) && (rule.Max == 0 || order.Number <= rule.Max) {
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < rule.ProfitProbability {
					return MicroRiskProfit
				}
				return MicroRiskLoss
			}
		}
		// 未匹配到金额规则，使用默认胜率(取消平局)
		rand.Seed(time.Now().UnixNano())
		// 将原来的三分概率转为二分概率
		totalProb := riskConfig.ProbabilityProfitPercent + riskConfig.ProbabilityBalancePercent
		if totalProb <= 0 {
			totalProb = 50
		}
		profitThreshold := (riskConfig.ProbabilityProfitPercent * 100) / totalProb
		if rand.Intn(100) < profitThreshold {
			return MicroRiskProfit
		}
		return MicroRiskLoss
	}

	return MicroRiskNone
}

// defaultRandom 默认50%概率随机盈亏
func (e *RiskEngine) defaultRandom() MicroRiskResult {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(100) < 50 {
		return MicroRiskProfit
	}
	return MicroRiskLoss
}
