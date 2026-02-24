package admin

import (
	"strconv"

	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 管理端充值API
// ============================================================

// GetDepositAddressList 获取充值地址配置列表
func GetDepositAddressList(c *gin.Context) {
	addresses, err := service.GetDepositService().GetAllDepositAddresses()
	if err != nil {
		response.ServerError(c, "获取充值地址失败")
		return
	}
	response.Success(c, "获取成功", addresses)
}

// CreateDepositAddress 创建充值地址
func CreateDepositAddress(c *gin.Context) {
	var req service.CreateDepositAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	address, err := service.GetDepositService().CreateDepositAddress(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功", address)
}

// UpdateDepositAddress 更新充值地址
func UpdateDepositAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "ID无效")
		return
	}

	var req service.UpdateDepositAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	err = service.GetDepositService().UpdateDepositAddress(uint(id), &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteDepositAddress 删除充值地址
func DeleteDepositAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "ID无效")
		return
	}

	err = service.GetDepositService().DeleteDepositAddress(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// GetDepositOrderList 获取充值订单列表
func GetDepositOrderList(c *gin.Context) {
	var req service.AdminGetDepositOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 默认状态为-1（全部）
	if req.Status == 0 && c.Request.ContentLength > 0 {
		// 检查请求体中是否明确设置了status
		// 如果没有设置，则默认为-1
	}

	orders, total, err := service.GetDepositService().AdminGetDepositOrders(&req)
	if err != nil {
		response.ServerError(c, "获取订单列表失败")
		return
	}

	response.LayuiSuccess(c, orders, total)
}

// GetDepositOrderDetail 获取充值订单详情
func GetDepositOrderDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "订单ID无效")
		return
	}

	order, err := service.GetDepositService().AdminGetDepositOrderDetail(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", order)
}

// ApproveDepositOrder 审核通过充值订单
func ApproveDepositOrder(c *gin.Context) {
	adminID := c.GetUint("admin_id")
	if adminID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.AdminApproveDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	err := service.GetDepositService().AdminApproveDeposit(adminID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "审核通过")
}

// RejectDepositOrder 审核拒绝充值订单
func RejectDepositOrder(c *gin.Context) {
	adminID := c.GetUint("admin_id")
	if adminID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.AdminRejectDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	err := service.GetDepositService().AdminRejectDeposit(adminID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "审核拒绝")
}
