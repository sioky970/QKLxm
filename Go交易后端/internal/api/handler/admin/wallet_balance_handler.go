package admin

import (
	"net/http"
	"strconv"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type WalletBalanceHandler struct{}

var walletBalanceHandler *WalletBalanceHandler

func GetWalletBalanceHandler() *WalletBalanceHandler {
	if walletBalanceHandler == nil {
		walletBalanceHandler = &WalletBalanceHandler{}
	}
	return walletBalanceHandler
}

type UserWalletBalanceResponse struct {
	UserID          uint64 `json:"user_id"`
	Username        string `json:"username"`
	Phone           string `json:"phone"`
	SpotBalance     string `json:"spot_balance"`
	ContractBalance string `json:"contract_balance"`
	TotalBalance    string `json:"total_balance"`
	SpotLocked      string `json:"spot_locked"`
	ContractLocked  string `json:"contract_locked"`
	LastTradeTime   int64  `json:"last_trade_time"`
	UpdateTime      int64  `json:"update_time"`
}

type AdjustmentRequest struct {
	UserID         uint64 `json:"user_id" binding:"required"`
	WalletType     string `json:"wallet_type" binding:"required"`
	Amount         string `json:"amount" binding:"required"`
	AdjustmentType int8   `json:"adjustment_type" binding:"required,oneof=1 2"`
	Reason         string `json:"reason" binding:"required,min=2,max=500"`
	Remark         string `json:"remark"`
}

type AdjustmentResponse struct {
	Success       bool   `json:"success"`
	AdjustmentNo  string `json:"adjustment_no,omitempty"`
	BeforeBalance string `json:"before_balance,omitempty"`
	AfterBalance  string `json:"after_balance,omitempty"`
	Amount        string `json:"amount,omitempty"`
	Message       string `json:"message,omitempty"`
	ErrorMsg      string `json:"error_msg,omitempty"`
}

type AdjustmentRecordResponse struct {
	ID             uint64 `json:"id"`
	AdjustmentNo   string `json:"adjustment_no"`
	UserID         uint64 `json:"user_id"`
	WalletType     string `json:"wallet_type"`
	AdjustmentType int8   `json:"adjustment_type"`
	Amount         string `json:"amount"`
	BalanceBefore  string `json:"balance_before"`
	BalanceAfter   string `json:"balance_after"`
	OperatorID     uint64 `json:"operator_id"`
	OperatorName   string `json:"operator_name"`
	Reason         string `json:"reason"`
	Status         int8   `json:"status"`
	CreatedTime    int64  `json:"created_time"`
	CompletedTime  *int64 `json:"completed_time,omitempty"`
}

func (h *WalletBalanceHandler) GetUserWalletBalance(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.FailWithMessage(c, "用户ID无效")
		return
	}

	walletService := admin.GetWalletBalanceService()
	balance, err := walletService.GetUserWalletBalances(userID)
	if err != nil {
		response.FailWithMessage(c, "获取钱包余额失败: "+err.Error())
		return
	}

	response.SuccessWithData(c, UserWalletBalanceResponse{
		UserID:          balance.UserID,
		Username:        balance.Username,
		Phone:           balance.Phone,
		SpotBalance:     balance.SpotBalance.StringFixed(4),
		ContractBalance: balance.ContractBalance.StringFixed(4),
		TotalBalance:    balance.TotalBalance.StringFixed(4),
		SpotLocked:      balance.SpotLocked.StringFixed(4),
		ContractLocked:  balance.ContractLocked.StringFixed(4),
		LastTradeTime:   balance.LastTradeTime,
		UpdateTime:      balance.UpdateTime,
	})
}

func (h *WalletBalanceHandler) AdjustBalance(c *gin.Context) {
	operatorID, exists := c.Get("admin_id")
	if !exists {
		response.FailWithCode(c, http.StatusUnauthorized, 401, "未授权访问")
		return
	}
	oid, ok := operatorID.(uint64)
	if !ok {
		response.FailWithCode(c, http.StatusBadRequest, 400, "操作员ID无效")
		return
	}

	operatorName, _ := c.Get("username")
	oname, _ := operatorName.(string)
	if oname == "" {
		oname = "管理员"
	}

	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, "请求参数错误: "+err.Error())
		return
	}

	walletType := model.WalletType(req.WalletType)
	if walletType != model.WalletTypeSpot && walletType != model.WalletTypeContract {
		response.FailWithMessage(c, "钱包类型无效，仅支持 spot(现货) 或 contract(合约)")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.FailWithMessage(c, "金额格式无效，必须为大于0的数字")
		return
	}

	if req.AdjustmentType != 1 && req.AdjustmentType != 2 {
		response.FailWithMessage(c, "调整类型无效，仅支持 1(增加) 或 2(减少)")
		return
	}

	if len(req.Reason) < 2 || len(req.Reason) > 500 {
		response.FailWithMessage(c, "调整原因长度必须在2-500字符之间")
		return
	}

	serviceReq := &admin.AdjustmentRequest{
		UserID:         req.UserID,
		WalletType:     walletType,
		Amount:         amount,
		AdjustmentType: req.AdjustmentType,
		Reason:         req.Reason,
		Remark:         req.Remark,
		OperatorID:     oid,
		OperatorName:   oname,
	}

	walletService := admin.GetWalletBalanceService()
	result := walletService.AdjustBalance(serviceReq)

	if result.Success {
		response.SuccessWithData(c, AdjustmentResponse{
			Success:       true,
			AdjustmentNo:  result.AdjustmentNo,
			BeforeBalance: result.BeforeBalance,
			AfterBalance:  result.AfterBalance,
			Amount:        result.Amount,
			Message:       "余额调整成功",
		})
	} else {
		response.FailWithData(c, AdjustmentResponse{
			Success:      false,
			AdjustmentNo: result.AdjustmentNo,
			ErrorMsg:     result.ErrorMsg,
		})
	}
}

func (h *WalletBalanceHandler) GetAdjustmentRecords(c *gin.Context) {
	userIDStr := c.Query("user_id")
	var userID uint64
	var err error
	if userIDStr != "" {
		userID, err = strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			response.FailWithMessage(c, "用户ID无效")
			return
		}
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	walletService := admin.GetWalletBalanceService()
	var records []model.WalletAdjustmentRecord
	var total int64

	if userID > 0 {
		records, total, err = walletService.GetAdjustmentRecords(userID, page, pageSize)
	} else {
		records, total, err = walletService.GetAllAdjustmentRecords(page, pageSize, 0, "", 0, 0)
	}

	if err != nil {
		response.FailWithMessage(c, "获取调整记录失败: "+err.Error())
		return
	}

	responseData := make([]AdjustmentRecordResponse, len(records))
	for i, record := range records {
		responseData[i] = AdjustmentRecordResponse{
			ID:             record.ID,
			AdjustmentNo:   record.AdjustmentNo,
			UserID:         record.UserID,
			WalletType:     string(record.WalletType),
			AdjustmentType: record.AdjustmentType,
			Amount:         record.Amount.StringFixed(4),
			BalanceBefore:  record.BalanceBefore.StringFixed(4),
			BalanceAfter:   record.BalanceAfter.StringFixed(4),
			OperatorID:     record.OperatorID,
			OperatorName:   record.OperatorName,
			Reason:         record.Reason,
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
