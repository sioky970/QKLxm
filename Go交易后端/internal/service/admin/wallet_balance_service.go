package admin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletBalanceService struct{}

var walletBalanceInstance *WalletBalanceService

func GetWalletBalanceService() *WalletBalanceService {
	if walletBalanceInstance == nil {
		walletBalanceInstance = &WalletBalanceService{}
	}
	return walletBalanceInstance
}

type AdjustmentRequest struct {
	UserID         uint64           `json:"user_id" binding:"required"`
	WalletType     model.WalletType `json:"wallet_type" binding:"required"`
	CurrencyID     uint64           `json:"currency_id"`
	Amount         decimal.Decimal  `json:"amount" binding:"required"`
	AdjustmentType int8             `json:"adjustment_type" binding:"required,oneof=1 2"`
	Reason         string           `json:"reason" binding:"required,min=2,max=500"`
	Remark         string           `json:"remark"`
	OperatorID     uint64           `json:"operator_id"`
	OperatorName   string           `json:"operator_name"`
}

type AdjustmentResult struct {
	Success       bool   `json:"success"`
	AdjustmentNo  string `json:"adjustment_no,omitempty"`
	BeforeBalance string `json:"before_balance"`
	AfterBalance  string `json:"after_balance"`
	Amount        string `json:"amount"`
	ErrorMsg      string `json:"error_msg,omitempty"`
}

type UserWalletBalance struct {
	UserID          uint64          `json:"user_id"`
	Username        string          `json:"username"`
	Phone           string          `json:"phone"`
	SpotBalance     decimal.Decimal `json:"spot_balance"`
	ContractBalance decimal.Decimal `json:"contract_balance"`
	TotalBalance    decimal.Decimal `json:"total_balance"`
	SpotLocked      decimal.Decimal `json:"spot_locked"`
	ContractLocked  decimal.Decimal `json:"contract_locked"`
	LastTradeTime   int64           `json:"last_trade_time"`
	UpdateTime      int64           `json:"update_time"`
}

func (s *WalletBalanceService) GenerateAdjustmentNo() string {
	now := time.Now()
	return fmt.Sprintf("ADJ%s%s%s", now.Format("20060102150405"), now.Format("05"), strings.ReplaceAll(uuid.New().String()[:8], "-", ""))
}

func (s *WalletBalanceService) GetUserWalletBalances(userID uint64) (*UserWalletBalance, error) {
	var user model.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	spotAssets, _ := service.GetUserAssetsService().GetOrCreateUserAssetsByType(uint(userID), model.WalletTypeSpot)
	contractAssets, _ := service.GetUserAssetsService().GetOrCreateUserAssetsByType(uint(userID), model.WalletTypeContract)

	spotBalance := decimal.NewFromFloat(spotAssets.UsdtBalance)
	contractBalance := decimal.NewFromFloat(contractAssets.UsdtBalance)
	spotLocked := decimal.NewFromFloat(spotAssets.UsdtLocked)
	contractLocked := decimal.NewFromFloat(contractAssets.UsdtLocked)
	totalBalance := spotBalance.Add(contractBalance)

	return &UserWalletBalance{
		UserID:          userID,
		Username:        user.Phone,
		Phone:           user.Phone,
		SpotBalance:     spotBalance,
		ContractBalance: contractBalance,
		TotalBalance:    totalBalance,
		SpotLocked:      spotLocked,
		ContractLocked:  contractLocked,
		LastTradeTime:   spotAssets.LastUpdateTime,
		UpdateTime:      spotAssets.UpdateTime,
	}, nil
}

func (s *WalletBalanceService) GetUserWallet(userID uint64, walletType model.WalletType) (*model.UserAssets, error) {
	return service.GetUserAssetsService().GetUserAssetsByType(uint(userID), walletType)
}

func (s *WalletBalanceService) CreateWallet(userID uint64, walletType model.WalletType) (*model.UserAssets, error) {
	return service.GetUserAssetsService().CreateUserAssetsByType(uint(userID), walletType)
}

func (s *WalletBalanceService) AdjustBalance(req *AdjustmentRequest) *AdjustmentResult {
	adjustmentNo := s.GenerateAdjustmentNo()
	now := time.Now().Unix()

	if req.WalletType != model.WalletTypeSpot && req.WalletType != model.WalletTypeContract {
		return &AdjustmentResult{
			Success:      false,
			AdjustmentNo: adjustmentNo,
			ErrorMsg:     "钱包类型无效，仅支持 spot(现货) 或 contract(合约)",
		}
	}

	if req.AdjustmentType != model.AdjustmentTypeIncrease && req.AdjustmentType != model.AdjustmentTypeDecrease {
		return &AdjustmentResult{
			Success:      false,
			AdjustmentNo: adjustmentNo,
			ErrorMsg:     "调整类型无效，仅支持 1(增加) 或 2(减少)",
		}
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return &AdjustmentResult{
			Success:      false,
			AdjustmentNo: adjustmentNo,
			ErrorMsg:     "调整金额必须大于0",
		}
	}

	var walletTypeName string
	switch req.WalletType {
	case model.WalletTypeSpot:
		walletTypeName = "现货账户"
	case model.WalletTypeContract:
		walletTypeName = "合约账户"
	}

	record := &model.WalletAdjustmentRecord{
		AdjustmentNo:   adjustmentNo,
		UserID:         req.UserID,
		WalletType:     req.WalletType,
		CurrencyID:     1,
		CurrencyName:   "USDT",
		AdjustmentType: req.AdjustmentType,
		Amount:         req.Amount,
		OperatorID:     req.OperatorID,
		OperatorName:   req.OperatorName,
		Reason:         req.Reason,
		Remark:         req.Remark,
		Status:         model.AdjustmentStatusPending,
		CreatedTime:    now,
	}

	var afterBalance decimal.Decimal
	var beforeBalance decimal.Decimal

	amountFloat, _ := req.Amount.Float64()

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		assets, err := service.GetUserAssetsService().GetUserAssetsByType(uint(req.UserID), req.WalletType)
		if err != nil {
			if req.AdjustmentType == model.AdjustmentTypeDecrease {
				return errors.New(fmt.Sprintf("%s余额为0，无法减少", walletTypeName))
			}
			_, err = service.GetUserAssetsService().CreateUserAssetsByType(uint(req.UserID), req.WalletType)
			if err != nil {
				return err
			}
			beforeBalance = decimal.Zero
		} else {
			beforeBalance = decimal.NewFromFloat(assets.UsdtBalance)
		}

		if req.AdjustmentType == model.AdjustmentTypeIncrease {
			if err := service.GetUserAssetsService().UpdateUsdtBalanceByType(uint(req.UserID), amountFloat, false, req.WalletType); err != nil {
				return err
			}
		} else {
			currentBalance := assets.UsdtBalance
			if currentBalance < amountFloat {
				return errors.New(fmt.Sprintf("%s余额不足，当前余额: %s", walletTypeName, decimal.NewFromFloat(currentBalance).String()))
			}
			if err := service.GetUserAssetsService().UpdateUsdtBalanceByType(uint(req.UserID), -amountFloat, false, req.WalletType); err != nil {
				return err
			}
		}

		newAssets, _ := service.GetUserAssetsService().GetUserAssetsByType(uint(req.UserID), req.WalletType)
		afterBalance = decimal.NewFromFloat(newAssets.UsdtBalance)

		record.BalanceBefore = beforeBalance
		record.BalanceAfter = afterBalance
		record.Status = model.AdjustmentStatusSuccess
		completedTime := now
		record.CompletedTime = &completedTime

		if err := tx.Create(record).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		record.Status = model.AdjustmentStatusFailed
		completedTime := now
		record.CompletedTime = &completedTime
		database.DB.Create(record)

		logger.Errorf("[WalletBalance] 余额调整失败: userID=%d, walletType=%s, amount=%s, err=%v",
			req.UserID, req.WalletType, req.Amount.String(), err)

		return &AdjustmentResult{
			Success:       false,
			AdjustmentNo:  adjustmentNo,
			BeforeBalance: beforeBalance.StringFixed(4),
			ErrorMsg:      err.Error(),
		}
	}

	logger.Infof("[WalletBalance] 余额调整成功: userID=%d, walletType=%s, type=%d, amount=%s, before=%s, after=%s, operator=%s",
		req.UserID, req.WalletType, req.AdjustmentType, req.Amount.String(), beforeBalance.String(), afterBalance.String(), req.OperatorName)

	return &AdjustmentResult{
		Success:       true,
		AdjustmentNo:  adjustmentNo,
		BeforeBalance: beforeBalance.StringFixed(4),
		AfterBalance:  afterBalance.StringFixed(4),
		Amount:        req.Amount.StringFixed(4),
	}
}

func (s *WalletBalanceService) GetAdjustmentRecords(userID uint64, page, pageSize int) ([]model.WalletAdjustmentRecord, int64, error) {
	var records []model.WalletAdjustmentRecord
	var total int64

	query := database.DB.Where("user_id = ?", userID)

	if err := query.Model(&model.WalletAdjustmentRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (s *WalletBalanceService) GetAllAdjustmentRecords(page, pageSize int, operatorID uint64, walletType model.WalletType, startTime, endTime int64) ([]model.WalletAdjustmentRecord, int64, error) {
	var records []model.WalletAdjustmentRecord
	var total int64

	query := database.DB.Model(&model.WalletAdjustmentRecord{})

	if operatorID > 0 {
		query = query.Where("operator_id = ?", operatorID)
	}
	if walletType != "" {
		query = query.Where("wallet_type = ?", walletType)
	}
	if startTime > 0 {
		query = query.Where("created_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_time <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
