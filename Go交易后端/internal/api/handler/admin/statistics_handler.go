package admin

import (
	"github.com/gin-gonic/gin"

	"exchange-go/internal/pkg/response"
)

// ============ 统计功能 ============

// GetDashboard 获取仪表盘数据
func GetDashboard(c *gin.Context) {
	data, err := adminService.Statistics.GetDashboard()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", data)
}

// GetUserStatistics 获取用户统计
func GetUserStatistics(c *gin.Context) {
	data, err := adminService.Statistics.GetUserStatistics()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", data)
}

// GetTradeStatistics 获取交易统计
func GetTradeStatistics(c *gin.Context) {
	data, err := adminService.Statistics.GetTradeStatistics()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", data)
}

// GetFinanceStatistics 获取财务统计
func GetFinanceStatistics(c *gin.Context) {
	data, err := adminService.Statistics.GetFinanceStatistics()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", data)
}

// GetTopTraders 获取顶级交易者
func GetTopTraders(c *gin.Context) {
	traders, err := adminService.Statistics.GetTopTraders(10)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", traders)
}

// GetCurrencyStatistics 获取币种统计
func GetCurrencyStatistics(c *gin.Context) {
	data, err := adminService.Statistics.GetCurrencyStatistics()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", data)
}
