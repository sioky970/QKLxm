package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// ============ 交易管理 ============

// GetSpotTransactionList 获取币币交易列表
func GetSpotTransactionList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := adminService.Transaction.GetSpotTransactionList(info, uint(userID), status)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetLeverTransactionList 获取合约交易列表
func GetLeverTransactionList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := adminService.Transaction.GetLeverTransactionList(info, uint(userID), status)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetMicroTransactionList 获取秒合约交易列表
func GetMicroTransactionList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := adminService.Transaction.GetMicroTransactionList(info, uint(userID), status)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// CancelOrder 取消订单
func CancelOrder(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Transaction.CancelSpotOrder(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "取消成功")
}

// GetLegalDealList 获取法币交易列表
func GetLegalDealList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := adminService.Transaction.GetLegalDealList(info, status)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetC2cDealList 获取C2C交易列表
func GetC2cDealList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := adminService.Transaction.GetC2cDealList(info, status)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetAccountLogList 获取账户流水列表
func GetAccountLogList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	logType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))

	list, total, err := adminService.Transaction.GetAccountLogList(info, uint(userID), logType)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetSellerList 获取商家列表
func GetSellerList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := adminService.Transaction.GetSellerList(info, status)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// ApproveSeller 审批商家
func ApproveSeller(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Transaction.ApproveSeller(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "审批通过")
}

// RejectSeller 拒绝商家申请
func RejectSeller(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Transaction.RejectSeller(req.ID, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "已拒绝")
}

// GetMicroConfig 获取秒合约配置
func GetMicroConfig(c *gin.Context) {
	numbers, seconds, err := adminService.Transaction.GetMicroConfigList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"numbers": numbers,
		"seconds": seconds,
	})
}

// GetLeverMultipleList 获取杠杆倍数配置
func GetLeverMultipleList(c *gin.Context) {
	list, err := adminService.Transaction.GetLeverMultipleList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}
