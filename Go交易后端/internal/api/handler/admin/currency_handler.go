package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/pkg/huobi"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// ============ 币种管理 ============

// GetCurrencyList 获取币种列表
func GetCurrencyList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Currency.GetCurrencyListWithMarket(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetCurrencyDetail 获取币种详情
func GetCurrencyDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	currency, err := adminService.Currency.GetCurrencyInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", currency)
}

// CreateCurrency 创建币种
func CreateCurrency(c *gin.Context) {
	response.BadRequest(c, "已禁用新增币种")
}

// UpdateCurrency 更新币种
func UpdateCurrency(c *gin.Context) {
	response.BadRequest(c, "仅允许启用/停用币种")
}

// DeleteCurrency 删除币种
func DeleteCurrency(c *gin.Context) {
	response.BadRequest(c, "已禁用删除币种")
}

// ToggleCurrencyDisplay 切换币种显示
func ToggleCurrencyDisplay(c *gin.Context) {
	var req struct {
		ID        uint `json:"id" binding:"required"`
		IsDisplay int8 `json:"is_display"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Currency.ToggleCurrencyDisplay(req.ID, req.IsDisplay); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// UpdateCurrencyRate 更新币种汇率
func UpdateCurrencyRate(c *gin.Context) {
	response.BadRequest(c, "已禁用修改汇率")
}

// UpdateCurrencyRisk 更新币种风控参数
func UpdateCurrencyRisk(c *gin.Context) {
	var req struct {
		ID                    uint    `json:"id" binding:"required"`
		RiskProbEnabled       int8    `json:"risk_prob_enabled"`
		RiskProfitProbability int     `json:"risk_profit_probability"`
		RiskMoneyEnabled      int8    `json:"risk_money_enabled"`
		RiskMoneyMin          float64 `json:"risk_money_min"`
		RiskMoneyMax          float64 `json:"risk_money_max"`
		RiskMoneyResult       int8    `json:"risk_money_result"`
		RiskTimeEnabled       int8    `json:"risk_time_enabled"`
		RiskTimeStart         string  `json:"risk_time_start"`
		RiskTimeEnd           string  `json:"risk_time_end"`
		RiskTimeResult        int8    `json:"risk_time_result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"risk_prob_enabled":       req.RiskProbEnabled,
		"risk_profit_probability": req.RiskProfitProbability,
		"risk_money_enabled":      req.RiskMoneyEnabled,
		"risk_money_min":          req.RiskMoneyMin,
		"risk_money_max":          req.RiskMoneyMax,
		"risk_money_result":       req.RiskMoneyResult,
		"risk_time_enabled":       req.RiskTimeEnabled,
		"risk_time_start":         req.RiskTimeStart,
		"risk_time_end":           req.RiskTimeEnd,
		"risk_time_result":        req.RiskTimeResult,
	}

	if err := adminService.Currency.UpdateCurrencyRisk(req.ID, updates); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// RefreshCurrencyMarket refreshes Huobi subscriptions for enabled currencies.
func RefreshCurrencyMarket(c *gin.Context) {
	manager := huobi.GetManager()
	if err := manager.RefreshSymbols(); err != nil {
		response.Error(c, err.Error())
		return
	}

	historyScheduler := huobi.GetHistoryKlineScheduler()
	if err := historyScheduler.RefreshSymbols(); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "刷新成功")
}

// GetCurrencyMatchList 获取交易对列表
func GetCurrencyMatchList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Currency.GetCurrencyMatchList(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// SetCurrencyMatchRisk 设置交易对风控
func SetCurrencyMatchRisk(c *gin.Context) {
	var req struct {
		ID         uint `json:"id" binding:"required"`
		RiskResult int8 `json:"risk_result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Currency.SetCurrencyMatchRisk(req.ID, req.RiskResult); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功")
}
