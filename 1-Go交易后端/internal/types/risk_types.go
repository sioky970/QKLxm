package types

// RiskResult 风控结果
type RiskResult int8

const (
	RiskResultLoss   RiskResult = -1 // 亏损
	RiskResultNone   RiskResult = 0  // 无预设
	RiskResultProfit RiskResult = 1  // 盈利
)