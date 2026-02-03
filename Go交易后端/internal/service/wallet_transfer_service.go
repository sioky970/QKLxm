package service

import (
	"fmt"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/websocket"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletTransferService struct{}

var walletTransferInstance *WalletTransferService

func GetWalletTransferService() *WalletTransferService {
	if walletTransferInstance == nil {
		walletTransferInstance = &WalletTransferService{}
	}
	return walletTransferInstance
}

type WalletBalanceResponse struct {
	SpotBalance     decimal.Decimal `json:"spot_balance"`
	ContractBalance decimal.Decimal `json:"contract_balance"`
	DeliveryBalance decimal.Decimal `json:"delivery_balance"`
	TotalBalance    decimal.Decimal `json:"total_balance"`
	CurrencyName    string          `json:"currency_name"`
	CurrencyID      uint64          `json:"currency_id"`
}

type TransferResult struct {
	Success        bool   `json:"success"`
	TransferNo     string `json:"transfer_no,omitempty"`
	FromWallet     string `json:"from_wallet"`
	ToWallet       string `json:"to_wallet"`
	Amount         string `json:"amount"`
	FromBalance    string `json:"from_balance"`
	ToBalance      string `json:"to_balance"`
	ErrorMsg       string `json:"error_msg,omitempty"`
}

func (s *WalletTransferService) GenerateTransferNo() string {
	now := time.Now()
	return fmt.Sprintf("TF%s%s%s", now.Format("20060102150405"), now.Format("05"), strings.ReplaceAll(uuid.New().String()[:8], "-", ""))
}

func (s *WalletTransferService) GetUserWalletBalance(userID uint64, walletType model.WalletType) (*model.UserAssets, error) {
	return GetUserAssetsService().GetUserAssetsByType(uint(userID), walletType)
}

func (s *WalletTransferService) CreateUserWallet(userID uint64, walletType model.WalletType) (*model.UserAssets, error) {
	return GetUserAssetsService().CreateUserAssetsByType(uint(userID), walletType)
}

func (s *WalletTransferService) GetUserAllWallets(userID uint64) (*WalletBalanceResponse, error) {
	spotAssets, err := s.GetUserWalletBalance(userID, model.WalletTypeSpot)
	if err != nil {
		return nil, err
	}

	contractAssets, err := s.GetUserWalletBalance(userID, model.WalletTypeContract)
	if err != nil {
		return nil, err
	}

	deliveryAssets, err := s.GetUserWalletBalance(userID, model.WalletTypeDelivery)
	if err != nil {
		return nil, err
	}

	spotBalance := decimal.NewFromFloat(spotAssets.UsdtBalance)
	contractBalance := decimal.NewFromFloat(contractAssets.UsdtBalance)
	deliveryBalance := decimal.NewFromFloat(deliveryAssets.UsdtBalance)
	totalBalance := spotBalance.Add(contractBalance).Add(deliveryBalance)

	return &WalletBalanceResponse{
		SpotBalance:     spotBalance,
		ContractBalance: contractBalance,
		DeliveryBalance: deliveryBalance,
		TotalBalance:    totalBalance,
		CurrencyName:    "USDT",
		CurrencyID:      1,
	}, nil
}

func (s *WalletTransferService) Transfer(userID uint64, fromWallet, toWallet model.WalletType, amount decimal.Decimal, clientIP string) *TransferResult {
	transferNo := s.GenerateTransferNo()
	now := time.Now().Unix()

	// 检查是否涉及资金钱包，如果是则使用 FundWalletService
	if fromWallet == model.WalletTypeFund || toWallet == model.WalletTypeFund {
		logger.Infof("[WalletTransfer] 检测到资金钱包划转，委托 FundWalletService 处理: from=%s, to=%s", fromWallet, toWallet)
		return GetFundWalletService().Transfer(userID, fromWallet, toWallet, amount, clientIP)
	}

	// 以下是原有逻辑，仅处理现货、合约、交割之间的划转
	if fromWallet == toWallet {
		return &TransferResult{
			Success:    false,
			TransferNo: transferNo,
			ErrorMsg:   "转出钱包和转入钱包不能相同",
		}
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return &TransferResult{
			Success:    false,
			TransferNo: transferNo,
			ErrorMsg:   "划转金额必须大于0",
		}
	}

	var fromWalletName, toWalletName string
	switch fromWallet {
	case model.WalletTypeSpot:
		fromWalletName = "现货账户"
	case model.WalletTypeContract:
		fromWalletName = "合约账户"
	case model.WalletTypeDelivery:
		fromWalletName = "交割账户"
	case model.WalletTypeFund:
		fromWalletName = "资金钱包"
	}
	switch toWallet {
	case model.WalletTypeSpot:
		toWalletName = "现货账户"
	case model.WalletTypeContract:
		toWalletName = "合约账户"
	case model.WalletTypeDelivery:
		toWalletName = "交割账户"
	case model.WalletTypeFund:
		toWalletName = "资金钱包"
	}

	record := &model.WalletTransferRecord{
		TransferNo:     transferNo,
		UserID:         userID,
		FromWalletType: fromWallet,
		ToWalletType:   toWallet,
		CurrencyID:     1,
		CurrencyName:   "USDT",
		Amount:         amount,
		Fee:            decimal.Zero,
		Status:         model.TransferStatusPending,
		ClientIP:       clientIP,
		CreatedTime:    now,
	}

	amountFloat, _ := amount.Float64()

	logger.Infof("[WalletTransfer] 开始划转: userID=%d, from=%s, to=%s, amount=%s", userID, fromWallet, toWallet, amount.String())

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 获取转出账户资产
		var fromAssets model.UserAssets
		if err := tx.Where("user_id = ? AND wallet_type = ?", userID, fromWallet).First(&fromAssets).Error; err != nil {
			return fmt.Errorf("%s不存在", fromWalletName)
		}

		tolerance := decimal.NewFromFloat(0.0001)
		fromBalance := decimal.NewFromFloat(fromAssets.UsdtBalance)
		if fromBalance.Add(tolerance).LessThan(amount) {
			return fmt.Errorf("%s余额不足，当前可用余额: %s", fromWalletName, fromBalance.StringFixed(4))
		}

		// 扣除转出账户余额
		if amount.GreaterThan(fromBalance) && amount.Sub(fromBalance).LessThanOrEqual(tolerance) {
			amountFloat = fromAssets.UsdtBalance
		}
		
		if err := tx.Model(&model.UserAssets{}).Where("id = ?", fromAssets.ID).Update("usdt_balance", gorm.Expr("usdt_balance - ?", amountFloat)).Error; err != nil {
			return err
		}

		// 增加转入账户余额
		var toAssets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, toWallet).First(&toAssets).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建转入账户
				toAssets = model.UserAssets{
					UserID:           uint(userID),
					WalletType:       toWallet,
					UsdtBalance:      amountFloat,
					CreateTime:       now,
					UpdateTime:       now,
					CurrencyBalances: make(model.CurrencyBalance),
					CurrencyLocked:   make(model.CurrencyBalance),
				}
				if err := tx.Create(&toAssets).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err := tx.Model(&model.UserAssets{}).Where("id = ?", toAssets.ID).Update("usdt_balance", gorm.Expr("usdt_balance + ?", amountFloat)).Error; err != nil {
				return err
			}
		}

		record.Status = model.TransferStatusSuccess
		completedTime := now
		record.CompletedTime = &completedTime
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		record.Status = model.TransferStatusFailed
		record.ErrorMsg = err.Error()
		completedTime := now
		record.CompletedTime = &completedTime
		database.DB.Create(record)

		logger.Errorf("[WalletTransfer] 划转失败: userID=%d, from=%s, to=%s, amount=%s, err=%v",
			userID, fromWallet, toWallet, amount.String(), err)

		return &TransferResult{
			Success:    false,
			TransferNo: transferNo,
			FromWallet: fromWalletName,
			ToWallet:   toWalletName,
			Amount:     amount.String(),
			ErrorMsg:   err.Error(),
		}
	}

	s.broadcastBalanceUpdate(userID)

	logger.Infof("[WalletTransfer] 划转成功: userID=%d, from=%s, to=%s, amount=%s, transferNo=%s",
		userID, fromWallet, toWallet, amount.String(), transferNo)

	return &TransferResult{
		Success:     true,
		TransferNo:  transferNo,
		FromWallet:  fromWalletName,
		ToWallet:    toWalletName,
		Amount:      amount.StringFixed(4),
		FromBalance: s.GetWalletBalanceDisplay(userID, fromWallet),
		ToBalance:   s.GetWalletBalanceDisplay(userID, toWallet),
	}
}

func (s *WalletTransferService) GetWalletBalanceDisplay(userID uint64, walletType model.WalletType) string {
	return fmt.Sprintf("可用余额: %s USDT", s.GetUserWalletBalanceDisplay(userID, walletType))
}

func (s *WalletTransferService) GetUserWalletBalanceDisplay(userID uint64, walletType model.WalletType) string {
	assets, err := s.GetUserWalletBalance(userID, walletType)
	if err != nil {
		return "0.0000"
	}
	return decimal.NewFromFloat(assets.UsdtBalance).StringFixed(4)
}

func (s *WalletTransferService) broadcastBalanceUpdate(userID uint64) {
	// 获取现货、合约、交割余额
	balance, err := s.GetUserAllWallets(userID)
	if err != nil {
		logger.Errorf("[WalletTransfer] 获取钱包余额失败: userID=%d, err=%v", userID, err)
		return
	}

	// 获取资金钱包余额
	fundBalance, err := GetFundWalletService().GetFundWalletBalance(userID, 1)
	if err != nil {
		logger.Warnf("[WalletTransfer] 获取资金钱包余额失败: userID=%d, err=%v", userID, err)
		fundBalance = decimal.Zero
	}

	// 计算包含资金钱包的总余额
	totalWithFund := balance.TotalBalance.Add(fundBalance)

	msg := map[string]interface{}{
		"type": "balance_update",
		"data": map[string]interface{}{
			"user_id":           userID,
			"spot_balance":      balance.SpotBalance.StringFixed(4),
			"contract_balance":  balance.ContractBalance.StringFixed(4),
			"delivery_balance":  balance.DeliveryBalance.StringFixed(4),
			"fund_balance":      fundBalance.StringFixed(4),
			"total_balance":     totalWithFund.StringFixed(4),
			"currency_name":     balance.CurrencyName,
			"timestamp":         time.Now().Unix(),
		},
	}

	websocket.BroadcastToUser(userID, msg)

	logger.Debugf("[WalletTransfer] 广播余额更新: userID=%d, spot=%s, contract=%s, delivery=%s, fund=%s, total=%s",
		userID, balance.SpotBalance.String(), balance.ContractBalance.String(), balance.DeliveryBalance.String(), fundBalance.String(), totalWithFund.String())
}

func (s *WalletTransferService) GetTransferRecords(userID uint64, page, pageSize int) ([]model.WalletTransferRecord, int64, error) {
	var records []model.WalletTransferRecord
	var total int64

	query := database.DB.Where("user_id = ?", userID)

	if err := query.Model(&model.WalletTransferRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
