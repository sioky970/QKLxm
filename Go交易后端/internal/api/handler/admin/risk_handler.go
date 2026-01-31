package admin

import (
	"github.com/gin-gonic/gin"

	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// ============ 风控管理 ============

// GetRiskConfig 获取风控配置
func GetRiskConfig(c *gin.Context) {
	config, err := adminService.Risk.GetRiskConfig()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", config)
}

// UpdateRiskConfig 更新风控配置
func UpdateRiskConfig(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"` // 去掉 binding:"required"，允许空值
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Risk.UpdateRiskConfig(req.Key, req.Value); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// GetUserRiskList 获取用户风控列表
func GetUserRiskList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Risk.GetUserRiskList(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// SetUserRisk 设置用户风控
func SetUserRisk(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
		Risk   int8 `json:"risk"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Risk.SetUserRisk(req.UserID, req.Risk); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功")
}

// GetCurrencyMatchRiskList 获取交易对风控列表
func GetCurrencyMatchRiskList(c *gin.Context) {
	list, err := adminService.Risk.GetCurrencyMatchRiskList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// SetMatchRisk 设置交易对风控（风控模块专用）
func SetMatchRisk(c *gin.Context) {
	var req struct {
		MatchID    uint `json:"match_id" binding:"required"`
		RiskResult int8 `json:"risk_result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Risk.SetCurrencyMatchRisk(req.MatchID, req.RiskResult); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功")
}

// GetMicroOrderRiskList 获取秒合约订单风控列表
func GetMicroOrderRiskList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Risk.GetMicroOrderRiskList(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// SetMicroOrderRisk 设置秒合约订单风控
func SetMicroOrderRisk(c *gin.Context) {
	var req struct {
		OrderID         uint `json:"order_id" binding:"required"`
		PreProfitResult int8 `json:"pre_profit_result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Risk.SetMicroOrderRisk(req.OrderID, req.PreProfitResult); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功")
}
