package handler

import (
	"strconv"

	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 用户端充值API
// ============================================================

// GetDepositAddresses 获取充值地址列表
// @Summary 获取充值地址列表
// @Description 获取所有启用的充值地址（按网络类型）
// @Tags 充值
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]model.DepositAddress}
// @Router /api/deposit/addresses [get]
func GetDepositAddresses(c *gin.Context) {
	addresses, err := service.GetDepositService().GetDepositAddresses()
	if err != nil {
		response.ServerError(c, "获取充值地址失败")
		return
	}
	response.Success(c, "获取成功", addresses)
}

// CreateDepositOrder 创建充值订单
// @Summary 创建充值订单
// @Description 创建一个新的充值订单
// @Tags 充值
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body service.CreateDepositOrderRequest true "充值信息"
// @Success 200 {object} response.Response{data=model.DepositOrder}
// @Router /api/deposit/create [post]
func CreateDepositOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.CreateDepositOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	order, err := service.GetDepositService().CreateDepositOrder(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功", order)
}

// UploadDepositScreenshot 上传转账截图
// @Summary 上传转账截图
// @Description 上传充值订单的转账截图
// @Tags 充值
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param order_id formData int true "订单ID"
// @Param file formData file true "截图文件"
// @Success 200 {object} response.Response{data=object{screenshot=string}}
// @Router /api/deposit/upload [post]
func UploadDepositScreenshot(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	orderIDStr := c.PostForm("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil || orderID == 0 {
		response.BadRequest(c, "订单ID无效")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的图片")
		return
	}

	path, err := service.GetDepositService().UploadScreenshot(userID, uint(orderID), file)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "上传成功", gin.H{"screenshot": path})
}

// GetDepositOrders 获取用户充值记录
// @Summary 获取用户充值记录
// @Description 获取当前用户的充值订单列表
// @Tags 充值
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query int false "状态筛选: -1全部 0待审核 1已通过 2已拒绝 3已取消 4已过期"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.DepositOrder,total=int}}
// @Router /api/deposit/orders [get]
func GetDepositOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.GetUserDepositOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 默认状态为-1（全部）
	if c.Query("status") == "" {
		req.Status = -1
	}

	orders, total, err := service.GetDepositService().GetUserDepositOrders(userID, &req)
	if err != nil {
		response.ServerError(c, "获取充值记录失败")
		return
	}

	response.Success(c, "获取成功", gin.H{
		"list":  orders,
		"total": total,
	})
}

// CancelDepositOrder 取消充值订单
// @Summary 取消充值订单
// @Description 取消一个待审核的充值订单
// @Tags 充值
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body object{order_id=int} true "订单ID"
// @Success 200 {object} response.Response
// @Router /api/deposit/cancel [post]
func CancelDepositOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		OrderID uint `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := service.GetDepositService().CancelDepositOrder(userID, req.OrderID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "取消成功")
}
