package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/types"

	"gorm.io/gorm"
)

// RiskControlService 风控服务
type RiskControlService struct{}

// NewRiskControlService 创建风控服务实例
func NewRiskControlService() *RiskControlService {
	return &RiskControlService{}
}

// RiskMode 风控模式
type RiskMode int

const (
	RiskModeNone        RiskMode = 0 // 无风控
	RiskModeUser        RiskMode = 1 // 用户点控
	RiskModeGroup       RiskMode = 2 // 全局群控
	RiskModeMoney       RiskMode = 3 // 金额控制
	RiskModeOrder       RiskMode = 4 // 单控
	RiskModeCurrency    RiskMode = 5 // 币种群控
	RiskModeProbability RiskMode = 6 // 概率控制（已合并到其他模式）
)

// MoneyRiskRule 金额风控规则
type MoneyRiskRule struct {
	Min               float64 `json:"min"`                // 最小金额
	Max               float64 `json:"max"`                // 最大金额 (0表示无上限)
	ProfitProbability int     `json:"profit_probability"` // 盈利概率(%)
	LossProbability   int     `json:"loss_probability"`   // 亏损概率(%)
}

// RiskConfig 风控配置
type RiskConfig struct {
	Mode                      RiskMode         `json:"mode"`                        // 风控模式
	GroupResult               types.RiskResult `json:"group_result"`                // 群控结果
	MoneyRules                []MoneyRiskRule  `json:"money_rules"`                 // 金额风控规则
	AdvanceSeconds            int              `json:"advance_seconds"`             // 提前风控秒数
	ProbabilityProfitPercent  int              `json:"probability_profit_percent"`  // 概率控制-盈利概率
	ProbabilityBalancePercent int              `json:"probability_balance_percent"` // 概率控制-平局概率
}

// GetRiskConfig 获取风控配置
func (s *RiskControlService) GetRiskConfig() (*RiskConfig, error) {
	// 尝试从缓存获取
	cacheKey := "risk:config"
	cachedConfig, err := cache.GetString(cacheKey)
	if err == nil && cachedConfig != "" {
		var config RiskConfig
		if err := json.Unmarshal([]byte(cachedConfig), &config); err == nil {
			return &config, nil
		}
	}

	// 从数据库获取
	config := &RiskConfig{
		Mode:                      RiskModeNone,
		GroupResult:               types.RiskResultLoss,
		AdvanceSeconds:            5, // 固定为5秒
		ProbabilityProfitPercent:  45,
		ProbabilityBalancePercent: 5,
	}

	// 获取风控模式
	var modeSetting model.Setting
	if err := database.DB.Where("key = ?", "risk_mode").First(&modeSetting).Error; err == nil {
		var mode int
		fmt.Sscanf(modeSetting.Value, "%d", &mode)
		config.Mode = RiskMode(mode)
	}

	// 获取群控结果
	var groupResultSetting model.Setting
	if err := database.DB.Where("key = ?", "risk_group_result").First(&groupResultSetting).Error; err == nil {
		var result int
		fmt.Sscanf(groupResultSetting.Value, "%d", &result)
		config.GroupResult = types.RiskResult(result)
	}

	// 获取金额风控规则
	var moneyRuleSetting model.Setting
	if err := database.DB.Where("key = ?", "risk_money_profit_probability").First(&moneyRuleSetting).Error; err == nil {
		config.MoneyRules = s.parseMoneyRules(moneyRuleSetting.Value)
	}

	// 获取概率控制配置
	var profitProbSetting model.Setting
	if err := database.DB.Where("key = ?", "risk_profit_probability").First(&profitProbSetting).Error; err == nil {
		fmt.Sscanf(profitProbSetting.Value, "%d", &config.ProbabilityProfitPercent)
	}

	var balanceProbSetting model.Setting
	if err := database.DB.Where("key = ?", "risk_probability_balance").First(&balanceProbSetting).Error; err == nil {
		fmt.Sscanf(balanceProbSetting.Value, "%d", &config.ProbabilityBalancePercent)
	}

	// 缓存配置（5分钟）
	configJSON, _ := json.Marshal(config)
	cache.SetString(cacheKey, string(configJSON), 5*time.Minute)

	return config, nil
}

// UpdateRiskConfig 更新风控配置
func (s *RiskControlService) UpdateRiskConfig(config *RiskConfig) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新风控模式
	s.updateOrCreateSetting(tx, "risk_mode", fmt.Sprintf("%d", config.Mode))
	s.updateOrCreateSetting(tx, "risk_group_result", fmt.Sprintf("%d", config.GroupResult))
	s.updateOrCreateSetting(tx, "risk_advance_seconds", "5") // 强制固定为5
	s.updateOrCreateSetting(tx, "risk_profit_probability", fmt.Sprintf("%d", config.ProbabilityProfitPercent))
	s.updateOrCreateSetting(tx, "risk_probability_balance", fmt.Sprintf("%d", config.ProbabilityBalancePercent))

	// 更新金额风控规则
	if len(config.MoneyRules) > 0 {
		ruleStr := s.formatMoneyRules(config.MoneyRules)
		s.updateOrCreateSetting(tx, "risk_money_profit_probability", ruleStr)
	}

	tx.Commit()

	// 清除缓存（如果使用了Redis缓存）
	// cache.Delete("risk:config")

	logger.Infof("风控配置已更新: mode=%d", config.Mode)
	return nil
}

// parseMoneyRules 解析金额风控规则
// 格式: "0-100:30|100-1000:20|1000-0:10"
func (s *RiskControlService) parseMoneyRules(ruleStr string) []MoneyRiskRule {
	if ruleStr == "" {
		return []MoneyRiskRule{}
	}

	rules := []MoneyRiskRule{}
	// 按 | 分割规则
	parts := splitString(ruleStr, "|")
	for _, part := range parts {
		// 按 : 分割范围和概率
		rangeParts := splitString(part, ":")
		if len(rangeParts) != 2 {
			continue
		}

		// 解析金额范围
		rangeStr := rangeParts[0]
		probability := 0
		fmt.Sscanf(rangeParts[1], "%d", &probability)

		minMax := splitString(rangeStr, "-")
		if len(minMax) != 2 {
			continue
		}

		min, max := 0.0, 0.0
		fmt.Sscanf(minMax[0], "%f", &min)
		fmt.Sscanf(minMax[1], "%f", &max)

		rule := MoneyRiskRule{
			Min:               min,
			Max:               max,
			ProfitProbability: probability,
			LossProbability:   100 - probability,
		}
		rules = append(rules, rule)
	}

	return rules
}

// formatMoneyRules 格式化金额风控规则
func (s *RiskControlService) formatMoneyRules(rules []MoneyRiskRule) string {
	if len(rules) == 0 {
		return ""
	}

	parts := []string{}
	for _, rule := range rules {
		part := fmt.Sprintf("%.0f-%.0f:%d", rule.Min, rule.Max, rule.ProfitProbability)
		parts = append(parts, part)
	}

	return joinString(parts, "|")
}

// updateOrCreateSetting 更新或创建设置
func (s *RiskControlService) updateOrCreateSetting(tx *gorm.DB, key, value string) error {
	var setting model.Setting
	err := tx.Where("key = ?", key).First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新设置
			setting = model.Setting{
				Key:   key,
				Value: value,
			}
			return tx.Create(&setting).Error
		}
		return err
	}

	// 更新现有设置
	setting.Value = value
	return tx.Save(&setting).Error
}

// SetUserRisk 设置用户风控标记
func (s *RiskControlService) SetUserRisk(userID uint, risk types.RiskResult) error {
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	user.Risk = int8(risk)
	if err := database.DB.Save(&user).Error; err != nil {
		return fmt.Errorf("设置用户风控失败: %w", err)
	}

	logger.Infof("设置用户风控: userID=%d, risk=%d", userID, risk)
	return nil
}

// BatchSetUserRisk 批量设置用户风控
func (s *RiskControlService) BatchSetUserRisk(userIDs []uint, risk types.RiskResult) error {
	if len(userIDs) == 0 {
		return errors.New("用户ID列表为空")
	}

	result := database.DB.Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("risk", int8(risk))

	if result.Error != nil {
		return fmt.Errorf("批量设置用户风控失败: %w", result.Error)
	}

	logger.Infof("批量设置用户风控: count=%d, risk=%d", result.RowsAffected, risk)
	return nil
}

// SetCurrencyMatchRisk 设置交易对风控
func (s *RiskControlService) SetCurrencyMatchRisk(matchID uint, risk types.RiskResult) error {
	var match model.CurrencyMatch
	if err := database.DB.First(&match, matchID).Error; err != nil {
		return errors.New("交易对不存在")
	}

	match.RiskGroupResult = int8(risk)
	if err := database.DB.Save(&match).Error; err != nil {
		return fmt.Errorf("设置交易对风控失败: %w", err)
	}

	logger.Infof("设置交易对风控: matchID=%d, risk=%d", matchID, risk)
	return nil
}

// BatchSetCurrencyMatchRisk 批量设置交易对风控
func (s *RiskControlService) BatchSetCurrencyMatchRisk(matchIDs []uint, risk types.RiskResult) error {
	if len(matchIDs) == 0 {
		return errors.New("交易对ID列表为空")
	}

	result := database.DB.Model(&model.CurrencyMatch{}).
		Where("id IN ?", matchIDs).
		Update("risk_group_result", int8(risk))

	if result.Error != nil {
		return fmt.Errorf("批量设置交易对风控失败: %w", result.Error)
	}

	logger.Infof("批量设置交易对风控: count=%d, risk=%d", result.RowsAffected, risk)
	return nil
}

// SetOrderRisk 设置订单风控
func (s *RiskControlService) SetOrderRisk(orderID uint, risk types.RiskResult) error {
	var order model.MicroOrder
	if err := database.DB.First(&order, orderID).Error; err != nil {
		return errors.New("订单不存在")
	}

	if order.Status != 0 && order.Status != 1 {
		return errors.New("订单已结算，无法设置风控")
	}

	order.PreProfitResult = int8(risk)
	if err := database.DB.Save(&order).Error; err != nil {
		return fmt.Errorf("设置订单风控失败: %w", err)
	}

	logger.Infof("设置订单风控: orderID=%d, risk=%d", orderID, risk)
	return nil
}

// BatchSetOrderRisk 批量设置订单风控
func (s *RiskControlService) BatchSetOrderRisk(orderIDs []uint, risk types.RiskResult) error {
	if len(orderIDs) == 0 {
		return errors.New("订单ID列表为空")
	}

	result := database.DB.Model(&model.MicroOrder{}).
		Where("id IN ? AND status IN (0, 1)", orderIDs).
		Update("pre_profit_result", int8(risk))

	if result.Error != nil {
		return fmt.Errorf("批量设置订单风控失败: %w", result.Error)
	}

	logger.Infof("批量设置订单风控: count=%d, risk=%d", result.RowsAffected, risk)
	return nil
}

// GetRandomResult 根据概率获取随机结果
func (s *RiskControlService) GetRandomResult(profitProb, balanceProb int) types.RiskResult {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(100) + 1 // 1-100

	if randNum <= profitProb {
		return types.RiskResultProfit
	} else if randNum <= profitProb+balanceProb {
		return types.RiskResultNone // 平局
	}
	return types.RiskResultLoss
}

// 工具函数
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func joinString(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
