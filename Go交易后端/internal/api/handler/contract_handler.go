package handler

import (
	"strconv"

	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ====================================
// 合约交易Handler
// 处理永续合约交易相关的API请求
// ====================================

// ContractOpenPosition 合约开仓
// @Summary 合约开仓
// @Description 市价单/限价单开仓
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body service.OpenPositionRequest true "开仓请求"
// @Success 200 {object} response.Response{data=service.OpenPositionResponse}
// @Router /api/contract/open [post]
func ContractOpenPosition(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	var req service.OpenPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	svc := service.GetFuturesTradingService()
	resp, err := svc.OpenPosition(userID.(uint), &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "开仓成功", resp)
}

// ContractClosePosition 合约平仓
// @Summary 合约平仓
// @Description 手动平仓
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body service.ClosePositionRequest true "平仓请求"
// @Success 200 {object} response.Response{data=service.ClosePositionResponse}
// @Router /api/contract/close [post]
func ContractClosePosition(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	var req service.ClosePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	svc := service.GetFuturesTradingService()
	resp, err := svc.ClosePosition(userID.(uint), &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "平仓成功", resp)
}

// ContractChaseOrder 合约追单
// @Summary 合约追单
// @Description 将限价单转为市价单立即成交
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body service.ContractChaseOrderRequest true "追单请求"
// @Success 200 {object} response.Response{data=service.OpenPositionResponse}
// @Router /api/contract/chase [post]
func ContractChaseOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	var req service.ContractChaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	svc := service.GetFuturesTradingService()
	resp, err := svc.ChaseOrder(userID.(uint), &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "追单成功", resp)
}

// ContractSetTPSL 设置止盈止损
// @Summary 设置止盈止损
// @Description 为持仓中的仓位设置止盈止损金额
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body service.SetTPSLRequest true "止盈止损请求"
// @Success 200 {object} response.Response
// @Router /api/contract/tpsl [post]
func ContractSetTPSL(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	var req service.SetTPSLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	svc := service.GetFuturesTradingService()
	if err := svc.SetTPSL(userID.(uint), &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功", nil)
}

// ContractCancelOrder 撤销限价单
// @Summary 撤销限价单
// @Description 撤销挂单中的限价单，退还保证金
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param order_id path int true "订单ID"
// @Success 200 {object} response.Response
// @Router /api/contract/cancel/{order_id} [post]
func ContractCancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Error(c, "订单ID无效")
		return
	}

	svc := service.GetFuturesTradingService()
	if err := svc.CancelOrder(userID.(uint), uint(orderID)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "撤单成功", nil)
}

// ContractGetPositions 获取持仓列表
// @Summary 获取持仓列表
// @Description 获取用户的合约持仓列表
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param currency_id query int false "币种ID"
// @Param legal_id query int false "法币ID"
// @Param status query int false "状态(-1=全部,0=挂单中,1=持仓中)"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=[]service.PositionDetailResponse}
// @Router /api/contract/positions [get]
func ContractGetPositions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	var req service.PositionListRequest
	req.Status = -1 // 默认全部

	if currencyID := c.Query("currency_id"); currencyID != "" {
		if id, err := strconv.ParseUint(currencyID, 10, 64); err == nil {
			req.CurrencyID = uint(id)
		}
	}
	if legalID := c.Query("legal_id"); legalID != "" {
		if id, err := strconv.ParseUint(legalID, 10, 64); err == nil {
			req.LegalID = uint(id)
		}
	}
	if status := c.Query("status"); status != "" {
		if s, err := strconv.ParseInt(status, 10, 8); err == nil {
			req.Status = int8(s)
		}
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			req.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			req.PageSize = ps
		}
	}

	svc := service.GetFuturesTradingService()
	positions, total, err := svc.GetPositionList(userID.(uint), &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", gin.H{
		"list":  positions,
		"total": total,
		"page":  req.Page,
		"size":  req.PageSize,
	})
}

// ContractGetPendingOrders 获取挂单列表
// @Summary 获取挂单列表
// @Description 获取用户的限价单挂单列表
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param currency_id query int false "币种ID"
// @Param legal_id query int false "法币ID"
// @Success 200 {object} response.Response{data=[]service.PositionDetailResponse}
// @Router /api/contract/pending [get]
func ContractGetPendingOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	var currencyID, legalID uint
	if cid := c.Query("currency_id"); cid != "" {
		if id, err := strconv.ParseUint(cid, 10, 64); err == nil {
			currencyID = uint(id)
		}
	}
	if lid := c.Query("legal_id"); lid != "" {
		if id, err := strconv.ParseUint(lid, 10, 64); err == nil {
			legalID = uint(id)
		}
	}

	svc := service.GetFuturesTradingService()
	orders, err := svc.GetPendingOrders(userID.(uint), currencyID, legalID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", orders)
}

// ContractGetPositionDetail 获取持仓详情
// @Summary 获取持仓详情
// @Description 获取单个仓位的详细信息
// @Tags 合约交易
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param position_id path int true "仓位ID"
// @Success 200 {object} response.Response{data=service.PositionDetailResponse}
// @Router /api/contract/position/{position_id} [get]
func ContractGetPositionDetail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "用户未登录")
		return
	}

	positionIDStr := c.Param("position_id")
	positionID, err := strconv.ParseUint(positionIDStr, 10, 64)
	if err != nil {
		response.Error(c, "仓位ID无效")
		return
	}

	// 构造请求获取单个仓位
	svc := service.GetFuturesTradingService()
	positions, _, err := svc.GetPositionList(userID.(uint), &service.PositionListRequest{
		Status:   -1,
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	// 查找指定仓位
	for _, pos := range positions {
		if pos.ID == uint(positionID) {
			response.Success(c, "获取成功", pos)
			return
		}
	}

	response.Error(c, "仓位不存在")
}
