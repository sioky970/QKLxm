package handler

import (
	"net/http"
	"strconv"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type WalletTransferHandler struct{}

var walletTransferHandler *WalletTransferHandler

func GetWalletTransferHandler() *WalletTransferHandler {
	if walletTransferHandler == nil {
		walletTransferHandler = &WalletTransferHandler{}
	}
	return walletTransferHandler
}

func getUserIDFromContext(c *gin.Context) (uint64, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// 尝试 uint64
	if uid, ok := userIDVal.(uint64); ok {
		return uid, true
	}
	// 尝试 uint
	if uid, ok := userIDVal.(uint); ok {
		return uint64(uid), true
	}
	// 尝试 int (某些场景可能使用)
	if uid, ok := userIDVal.(int); ok {
		return uint64(uid), true
	}
	// 尝试 string (从token解析可能是字符串)
	if uidStr, ok := userIDVal.(string); ok {
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			return uid, true
		}
	}
	return 0, false
}

type WalletBalanceRequest struct {
	UserID uint64 `json:"user_id"`
}

type WalletBalanceResponse struct {
	SpotBalance     string `json:"spot_balance"`
	ContractBalance string `json:"contract_balance"`
	DeliveryBalance string `json:"delivery_balance"`
	TotalBalance    string `json:"total_balance"`
	CurrencyName    string `json:"currency_name"`
	CurrencyID      uint64 `json:"currency_id"`
}

type TransferRequest struct {
	FromWallet string  `json:"from_wallet" binding:"required"`
	ToWallet   string  `json:"to_wallet" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

type TransferResponse struct {
	Success     bool   `json:"success"`
	TransferNo  string `json:"transfer_no"`
	FromWallet  string `json:"from_wallet"`
	ToWallet    string `json:"to_wallet"`
	Amount      string `json:"amount"`
	FromBalance string `json:"from_balance"`
	ToBalance   string `json:"to_balance"`
	Message     string `json:"message,omitempty"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}

type TransferRecordResponse struct {
	ID             uint64 `json:"id"`
	TransferNo     string `json:"transfer_no"`
	FromWalletType string `json:"from_wallet_type"`
	ToWalletType   string `json:"to_wallet_type"`
	Amount         string `json:"amount"`
	Fee            string `json:"fee"`
	Status         int8   `json:"status"`
	CreatedTime    int64  `json:"created_time"`
	CompletedTime  *int64 `json:"completed_time,omitempty"`
}

func (h *WalletTransferHandler) GetBalance(c *gin.Context) {
	uid, ok := getUserIDFromContext(c)
	if !ok {
		response.FailWithCode(c, http.StatusBadRequest, 400, "用户ID无效")
		return
	}

	walletService := service.GetWalletTransferService()
	balance, err := walletService.GetUserAllWallets(uid)
	if err != nil {
		response.FailWithMessage(c, "获取钱包余额失败: "+err.Error())
		return
	}

	response.SuccessWithData(c, WalletBalanceResponse{
		SpotBalance:     balance.SpotBalance.StringFixed(4),
		ContractBalance: balance.ContractBalance.StringFixed(4),
		DeliveryBalance: balance.DeliveryBalance.StringFixed(4),
		TotalBalance:    balance.TotalBalance.StringFixed(4),
		CurrencyName:    balance.CurrencyName,
		CurrencyID:      balance.CurrencyID,
	})
}

func (h *WalletTransferHandler) Transfer(c *gin.Context) {
	uid, ok := getUserIDFromContext(c)
	if !ok {
		response.FailWithCode(c, http.StatusBadRequest, 400, "用户ID无效")
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, "请求参数错误: "+err.Error())
		return
	}

	fromWallet := model.WalletType(req.FromWallet)
	toWallet := model.WalletType(req.ToWallet)

	if fromWallet != model.WalletTypeSpot && fromWallet != model.WalletTypeContract && fromWallet != model.WalletTypeDelivery && fromWallet != model.WalletTypeFund {
		response.FailWithMessage(c, "转出钱包类型无效，仅支持 spot(现货)、contract(合约)、delivery(交割) 或 fund(资金钱包)")
		return
	}

	if toWallet != model.WalletTypeSpot && toWallet != model.WalletTypeContract && toWallet != model.WalletTypeDelivery && toWallet != model.WalletTypeFund {
		response.FailWithMessage(c, "转入钱包类型无效，仅支持 spot(现货)、contract(合约)、delivery(交割) 或 fund(资金钱包)")
		return
	}

	if fromWallet == toWallet {
		response.FailWithMessage(c, "转出钱包和转入钱包不能相同")
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		response.FailWithMessage(c, "划转金额必须大于0")
		return
	}

	clientIP := c.ClientIP()
	walletService := service.GetWalletTransferService()
	result := walletService.Transfer(uid, fromWallet, toWallet, amount, clientIP)

	if result.Success {
		response.SuccessWithData(c, TransferResponse{
			Success:     true,
			TransferNo:  result.TransferNo,
			FromWallet:  result.FromWallet,
			ToWallet:    result.ToWallet,
			Amount:      result.Amount,
			FromBalance: result.FromBalance,
			ToBalance:   result.ToBalance,
			Message:     "划转成功",
		})
	} else {
		response.FailWithData(c, TransferResponse{
			Success:    false,
			TransferNo: result.TransferNo,
			ErrorMsg:   result.ErrorMsg,
		})
	}
}

func (h *WalletTransferHandler) GetTransferRecords(c *gin.Context) {
	uid, ok := getUserIDFromContext(c)
	if !ok {
		response.FailWithCode(c, http.StatusBadRequest, 400, "用户ID无效")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	walletService := service.GetWalletTransferService()
	records, total, err := walletService.GetTransferRecords(uid, page, pageSize)
	if err != nil {
		response.FailWithMessage(c, "获取划转记录失败: "+err.Error())
		return
	}

	responseData := make([]TransferRecordResponse, len(records))
	for i, record := range records {
		responseData[i] = TransferRecordResponse{
			ID:             record.ID,
			TransferNo:     record.TransferNo,
			FromWalletType: string(record.FromWalletType),
			ToWalletType:   string(record.ToWalletType),
			Amount:         record.Amount.StringFixed(4),
			Fee:            record.Fee.StringFixed(4),
			Status:         record.Status,
			CreatedTime:    record.CreatedTime,
			CompletedTime:  record.CompletedTime,
		}
	}

	response.SuccessWithData(c, gin.H{
		"list":  responseData,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *WalletTransferHandler) GetSpotBalance(c *gin.Context) {
	uid, ok := getUserIDFromContext(c)
	if !ok {
		response.FailWithCode(c, http.StatusBadRequest, 400, "用户ID无效")
		return
	}

	walletService := service.GetWalletTransferService()
	assets, err := walletService.GetUserWalletBalance(uid, model.WalletTypeSpot)
	if err != nil {
		response.FailWithMessage(c, "获取现货钱包余额失败: "+err.Error())
		return
	}

	response.SuccessWithData(c, gin.H{
		"wallet_type":       "spot",
		"wallet_type_name":  "现货账户",
		"available_balance": decimal.NewFromFloat(assets.UsdtBalance).StringFixed(4),
		"locked_balance":    decimal.NewFromFloat(assets.UsdtLocked).StringFixed(4),
		"total_balance":     decimal.NewFromFloat(assets.UsdtBalance + assets.UsdtLocked).StringFixed(4),
		"currency_name":     "USDT",
		"currency_id":       1,
		"last_trade_time":   assets.LastUpdateTime,
		"update_time":       assets.UpdateTime,
	})
}

func (h *WalletTransferHandler) GetContractBalance(c *gin.Context) {
	uid, ok := getUserIDFromContext(c)
	if !ok {
		response.FailWithCode(c, http.StatusBadRequest, 400, "用户ID无效")
		return
	}

	walletService := service.GetWalletTransferService()
	assets, err := walletService.GetUserWalletBalance(uid, model.WalletTypeContract)
	if err != nil {
		response.FailWithMessage(c, "获取合约钱包余额失败: "+err.Error())
		return
	}

	response.SuccessWithData(c, gin.H{
		"wallet_type":       "contract",
		"wallet_type_name":  "合约账户",
		"available_balance": decimal.NewFromFloat(assets.UsdtBalance).StringFixed(4),
		"locked_balance":    decimal.NewFromFloat(assets.UsdtLocked).StringFixed(4),
		"total_balance":     decimal.NewFromFloat(assets.UsdtBalance + assets.UsdtLocked).StringFixed(4),
		"currency_name":     "USDT",
		"currency_id":       1,
		"last_trade_time":   assets.LastUpdateTime,
		"update_time":       assets.UpdateTime,
	})
}
