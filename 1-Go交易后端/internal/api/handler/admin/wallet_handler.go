package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/pkg/response"
)

// ============ 钱包管理 ============

// GetWalletList 获取钱包列表
func GetWalletList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page" form:"page"`
		PageSize      int    `json:"page_size" form:"page_size"`
		AccountNumber string `json:"account_number" form:"account_number"`
		CurrencyID    uint   `json:"currency_id" form:"currency_id"`
		Status        *int   `json:"status" form:"status"`
		Address       string `json:"address" form:"address"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Wallet.GetWalletListAdvanced(req.AccountNumber, req.CurrencyID, req.Status, req.Address, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	totals, err := adminService.Wallet.GetWalletTotals(req.AccountNumber, req.CurrencyID, req.Status, req.Address)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPageExtra(c, list, total, req.Page, req.PageSize, gin.H{"total": totals})
}

// GetWalletDetail 获取钱包详情
func GetWalletDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	wallet, err := adminService.Wallet.GetWalletInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", wallet)
}

// UpdateWalletBalance 更新钱包余额
func UpdateWalletBalance(c *gin.Context) {
	var req struct {
		ID          uint    `json:"id" binding:"required"`
		BalanceType string  `json:"balance_type" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		Reason      string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Wallet.UpdateBalance(req.ID, req.BalanceType, req.Amount, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// FreezeWallet 冻结钱包
func FreezeWallet(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Wallet.FreezeWallet(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "冻结成功")
}

// ActivateWallet 激活钱包
func ActivateWallet(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Wallet.ActivateWallet(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "激活成功")
}

// GetWithdrawalsList 获取提币列表
func GetWithdrawalsList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		Status        *int   `json:"status"`
		CurrencyID    uint   `json:"currency_id"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	list, total, err := adminService.Wallet.GetWithdrawalsListAdvanced(req.AccountNumber, req.CurrencyID, req.Status, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetWithdrawalDetail 获取提币详情
func GetWithdrawalDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	detail, err := adminService.Wallet.GetWithdrawalDetail(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", detail)
}

// ApproveWithdrawal 审核通过提币
func ApproveWithdrawal(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Method string `json:"method"` // manual-手动, auto-自动链上转账
		Txid   string `json:"txid"`   // 交易哈希（手动时填写）
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Method == "" {
		req.Method = "manual"
	}

	if err := adminService.Wallet.ApproveWithdrawalAdvanced(req.ID, req.Method, req.Txid, req.Notes); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "审核通过")
}

// RejectWithdrawal 拒绝提币
func RejectWithdrawal(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Wallet.RejectWithdrawal(req.ID, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "已拒绝")
}

// GetChargeList 获取充值列表
func GetChargeList(c *gin.Context) {
	var req struct {
		Page          int    `json:"page" form:"page"`
		PageSize      int    `json:"page_size" form:"page_size"`
		AccountNumber string `json:"account_number" form:"account_number"`
		CurrencyID    uint   `json:"currency_id" form:"currency_id"`
		Status        *int   `json:"status" form:"status"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.Wallet.GetChargeListAdvanced(req.AccountNumber, req.CurrencyID, req.Status, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// ApproveCharge 审核通过充值
func ApproveCharge(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Wallet.ApproveCharge(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "审核通过")
}

// RejectCharge 拒绝充值
func RejectCharge(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.Wallet.RejectCharge(req.ID, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "已拒绝")
}
