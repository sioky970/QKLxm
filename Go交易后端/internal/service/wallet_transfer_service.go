package service

import (
	"errors"
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
		TotalBalance:    totalBalance,
		CurrencyName:    "USDT",
		CurrencyID:      1,
	}, nil
}

func (s *WalletTransferService) Transfer(userID uint64, fromWallet, toWallet model.WalletType, amount decimal.Decimal, clientIP string) *TransferResult {
	transferNo := s.GenerateTransferNo()
	now := time.Now().Unix()

	if fromWallet == toWallet {
		return &TransferResult{
			Success:    false,
			TransferNo: transferNo,
			ErrorMsg:  "转出钱包和转入钱包不能相同",
		}
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return &TransferResult{
			Success:    false,
			TransferNo: transferNo,
			ErrorMsg:  "划转金额必须大于0",
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
	}
	switch toWallet {
	case model.WalletTypeSpot:
		toWalletName = "现货账户"
	case model.WalletTypeContract:
		toWalletName = "合约账户"
	case model.WalletTypeDelivery:
		toWalletName = "交割账户"
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

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		fromAssets, err := GetUserAssetsService().GetUserAssetsByType(uint(userID), fromWallet)
		if err != nil {
			return errors.New(fmt.Sprintf("%s余额不足", fromWalletName))
		}

		tolerance := decimal.NewFromFloat(0.0001)
		fromBalance := decimal.NewFromFloat(fromAssets.UsdtBalance)
		if fromBalance.Add(tolerance).LessThan(amount) {
			return errors.New(fmt.Sprintf("%s余额不足，当前可用余额: %s", fromWalletName, fromBalance.StringFixed(4)))
		}

		if amount.GreaterThan(fromBalance) && amount.Sub(fromBalance).LessThanOrEqual(tolerance) {
			amountFloat = fromAssets.UsdtBalance
		}

		if err := GetUserAssetsService().UpdateUsdtBalanceByType(uint(userID), -amountFloat, false, fromWallet); err != nil {
			return err
		}

		_, err = GetUserAssetsService().GetUserAssetsByType(uint(userID), toWallet)
		if err != nil {
			_, err = GetUserAssetsService().CreateUserAssetsByType(uint(userID), toWallet)
			if err != nil {
				return err
			}
		}

		if err := GetUserAssetsService().UpdateUsdtBalanceByType(uint(userID), amountFloat, false, toWallet); err != nil {
			return err
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
		FromBalance: s.GetWalletBalanceDisplay(fromWallet),
		ToBalance:   s.GetWalletBalanceDisplay(toWallet),
	}
}

func (s *WalletTransferService) GetWalletBalanceDisplay(walletType model.WalletType) string {
	return fmt.Sprintf("可用余额: %s USDT", s.GetUserWalletBalanceDisplay(0, walletType))
}

func (s *WalletTransferService) GetUserWalletBalanceDisplay(userID uint64, walletType model.WalletType) string {
	assets, err := s.GetUserWalletBalance(userID, walletType)
	if err != nil {
		return "0.0000"
	}
	return decimal.NewFromFloat(assets.UsdtBalance).StringFixed(4)
}

func (s *WalletTransferService) broadcastBalanceUpdate(userID uint64) {
	balance, err := s.GetUserAllWallets(userID)
	if err != nil {
		logger.Errorf("[WalletTransfer] 广播余额更新失败: userID=%d, err=%v", userID, err)
		return
	}

	msg := map[string]interface{}{
		"type": "balance_update",
		"data": map[string]interface{}{
			"user_id":          userID,
			"spot_balance":     balance.SpotBalance.StringFixed(4),
			"contract_balance": balance.ContractBalance.StringFixed(4),
			"total_balance":    balance.TotalBalance.StringFixed(4),
			"currency_name":    balance.CurrencyName,
			"timestamp":        time.Now().Unix(),
		},
	}

	websocket.BroadcastToUser(userID, msg)

	logger.Debugf("[WalletTransfer] 广播余额更新: userID=%d, spot=%s, contract=%s, total=%s",
		userID, balance.SpotBalance.String(), balance.ContractBalance.String(), balance.TotalBalance.String())
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
