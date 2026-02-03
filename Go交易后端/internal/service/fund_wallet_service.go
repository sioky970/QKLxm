package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FundWalletService 资金钱包服务
type FundWalletService struct{}

var fundWalletInstance *FundWalletService

func GetFundWalletService() *FundWalletService {
	if fundWalletInstance == nil {
		fundWalletInstance = &FundWalletService{}
	}
	return fundWalletInstance
}

// ============================================================
// 基础查询操作
// ============================================================

// GetOrCreateFundWallet 获取或创建用户的资金钱包
func (s *FundWalletService) GetOrCreateFundWallet(userID uint64, currencyID uint64) (*model.FundWallet, error) {
	var wallet model.FundWallet
	err := database.DB.Where("user_id = ? AND currency_id = ?", userID, currencyID).First(&wallet).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新的资金钱包
		now := time.Now().Unix()
		wallet = model.FundWallet{
			UserID:           userID,
			CurrencyID:       currencyID,
			CurrencyName:     "USDT",
			AvailableBalance: decimal.Zero,
			LockedBalance:    decimal.Zero,
			TotalDeposit:     decimal.Zero,
			TotalWithdraw:    decimal.Zero,
			Version:          0,
			LastTradeTime:    0,
			CreateTime:       now,
			UpdateTime:       now,
		}

		if err := database.DB.Create(&wallet).Error; err != nil {
			logger.Errorf("[FundWallet] 创建资金钱包失败: userID=%d, currencyID=%d, err=%v", userID, currencyID, err)
			return nil, err
		}

		logger.Infof("[FundWallet] 创建资金钱包成功: userID=%d, currencyID=%d, walletID=%d", userID, currencyID, wallet.ID)
		return &wallet, nil
	}

	if err != nil {
		logger.Errorf("[FundWallet] 查询资金钱包失败: userID=%d, currencyID=%d, err=%v", userID, currencyID, err)
		return nil, err
	}

	return &wallet, nil
}

// GetFundWalletByUserID 获取用户的所有资金钱包
func (s *FundWalletService) GetFundWalletByUserID(userID uint64) ([]model.FundWallet, error) {
	var wallets []model.FundWallet
	err := database.DB.Where("user_id = ?", userID).Find(&wallets).Error
	if err != nil {
		logger.Errorf("[FundWallet] 查询用户资金钱包列表失败: userID=%d, err=%v", userID, err)
		return nil, err
	}
	return wallets, nil
}

// GetFundWalletBalance 获取资金钱包余额
func (s *FundWalletService) GetFundWalletBalance(userID uint64, currencyID uint64) (decimal.Decimal, error) {
	wallet, err := s.GetOrCreateFundWallet(userID, currencyID)
	if err != nil {
		return decimal.Zero, err
	}
	return wallet.AvailableBalance, nil
}

// ============================================================
// 划转操作
// ============================================================

// GenerateTransferNo 生成划转单号
func (s *FundWalletService) GenerateTransferNo() string {
	now := time.Now()
	return fmt.Sprintf("FW%s%s%s", now.Format("20060102150405"), now.Format("05"), strings.ReplaceAll(uuid.New().String()[:8], "-", ""))
}

// Transfer 资金钱包划转（与现货、合约账户互转）
func (s *FundWalletService) Transfer(userID uint64, fromWallet, toWallet model.WalletType, amount decimal.Decimal, clientIP string) *TransferResult {
	transferNo := s.GenerateTransferNo()
	now := time.Now().Unix()

	// 验证钱包类型
	if !s.isValidWalletType(fromWallet) || !s.isValidWalletType(toWallet) {
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  "无效的钱包类型",
		}
	}

	// 不能自己转给自己
	if fromWallet == toWallet {
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  "转出钱包和转入钱包不能相同",
		}
	}

	// 金额必须大于0
	if amount.LessThanOrEqual(decimal.Zero) {
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  "划转金额必须大于0",
		}
	}

	// 确定划转类型
	transferType := s.getTransferType(fromWallet, toWallet)

	logger.Infof("[FundWallet] 开始划转: userID=%d, from=%s, to=%s, amount=%s, transferNo=%s",
		userID, fromWallet, toWallet, amount.String(), transferNo)

	// 开始事务
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var fundWallet *model.FundWallet
		var err error
		var fromBalanceBefore, fromBalanceAfter, toBalanceBefore, toBalanceAfter decimal.Decimal

		// 处理从资金钱包转出
		if fromWallet == model.WalletTypeFund {
			// 从资金钱包转出
			fundWallet, err = s.getFundWalletForUpdate(tx, userID, 1)
			if err != nil {
				return errors.New("查询资金钱包失败")
			}

			if fundWallet.AvailableBalance.LessThan(amount) {
				return errors.New("资金钱包余额不足")
			}

			fromBalanceBefore = fundWallet.AvailableBalance
			fundWallet.AvailableBalance = fundWallet.AvailableBalance.Sub(amount)
			fundWallet.UpdateTime = now

			if err := tx.Save(fundWallet).Error; err != nil {
				return errors.New("更新资金钱包失败")
			}
			fromBalanceAfter = fundWallet.AvailableBalance

			// 更新目标钱包
			if toWallet == model.WalletTypeSpot {
				_, err := s.updateUserAssets(tx, userID, model.WalletTypeSpot, amount, false)
				if err != nil {
					return err
				}
			} else if toWallet == model.WalletTypeContract || toWallet == model.WalletTypeDelivery {
				_, err := s.updateUserAssets(tx, userID, toWallet, amount, false)
				if err != nil {
					return err
				}
			}

			// 记录划转日志
			transfer := &model.FundWalletTransfer{
				TransferNo:      transferNo,
				UserID:          userID,
				TransferType:    transferType,
				FromWalletType:  fromWallet,
				ToWalletType:    toWallet,
				CurrencyID:      1,
				CurrencyName:    "USDT",
				Amount:          amount,
				BalanceBefore:   fromBalanceBefore,
				BalanceAfter:    fromBalanceAfter,
				Status:          model.FundTransferSuccess,
				ClientIP:        clientIP,
				CreatedTime:     now,
				CompletedTime:   now,
			}

			if err := tx.Create(transfer).Error; err != nil {
				return errors.New("创建划转记录失败")
			}

			return nil
		}

		// 从其他钱包转入资金钱包
		// 先扣减转出钱包
		if fromWallet == model.WalletTypeSpot {
			_, err := s.updateUserAssets(tx, userID, model.WalletTypeSpot, amount, true)
			if err != nil {
				return err
			}
		} else if fromWallet == model.WalletTypeContract || fromWallet == model.WalletTypeDelivery {
			_, err := s.updateUserAssets(tx, userID, fromWallet, amount, true)
			if err != nil {
				return err
			}
		}

		// 再增加资金钱包余额
		fundWallet, err = s.getFundWalletForUpdate(tx, userID, 1)
		if err != nil {
			return errors.New("查询资金钱包失败")
		}

		fromBalanceBefore = decimal.Zero // 不记录转出钱包的前余额
		toBalanceBefore = fundWallet.AvailableBalance
		fundWallet.AvailableBalance = fundWallet.AvailableBalance.Add(amount)
		fundWallet.TotalDeposit = fundWallet.TotalDeposit.Add(amount)
		fundWallet.UpdateTime = now

		if err := tx.Save(fundWallet).Error; err != nil {
			return errors.New("更新资金钱包失败")
		}
		toBalanceAfter = fundWallet.AvailableBalance

		// 记录划转日志
		transfer := &model.FundWalletTransfer{
			TransferNo:     transferNo,
			UserID:         userID,
			TransferType:   transferType,
			FromWalletType: fromWallet,
			ToWalletType:   toWallet,
			CurrencyID:     1,
			CurrencyName:   "USDT",
			Amount:         amount,
			BalanceBefore:  toBalanceBefore,
			BalanceAfter:   toBalanceAfter,
			Status:         model.FundTransferSuccess,
			ClientIP:       clientIP,
			CreatedTime:    now,
			CompletedTime:  now,
		}

		if err := tx.Create(transfer).Error; err != nil {
			return errors.New("创建划转记录失败")
		}

		return nil
	})

	if err != nil {
		logger.Errorf("[FundWallet] 划转失败: userID=%d, err=%v", userID, err)
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  err.Error(),
		}
	}

	// 获取最新的资金钱包余额
	fundWallet, _ := s.GetOrCreateFundWallet(userID, 1)

	logger.Infof("[FundWallet] 划转成功: userID=%d, from=%s, to=%s, amount=%s",
		userID, fromWallet, toWallet, amount.String())

	return &TransferResult{
		Success:     true,
		TransferNo:  transferNo,
		FromWallet:  string(fromWallet),
		ToWallet:    string(toWallet),
		Amount:      amount.String(),
		FromBalance: decimal.Zero.String(),
		ToBalance:   fundWallet.AvailableBalance.String(),
	}
}

// ============================================================
// 充值和提现（与资金钱包交互）
// ============================================================

// DepositToFundWallet 充值到资金钱包
func (s *FundWalletService) DepositToFundWallet(userID uint64, amount decimal.Decimal, clientIP string) *TransferResult {
	transferNo := s.GenerateTransferNo()
	now := time.Now().Unix()

	if amount.LessThanOrEqual(decimal.Zero) {
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  "充值金额必须大于0",
		}
	}

	logger.Infof("[FundWallet] 开始充值: userID=%d, amount=%s, transferNo=%s", userID, amount.String(), transferNo)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 从现货账户扣减
		_, err := s.updateUserAssets(tx, userID, model.WalletTypeSpot, amount, true)
		if err != nil {
			return err
		}

		// 增加资金钱包余额
		fundWallet, err := s.getFundWalletForUpdate(tx, userID, 1)
		if err != nil {
			return errors.New("查询资金钱包失败")
		}

		balanceBefore := fundWallet.AvailableBalance
		fundWallet.AvailableBalance = fundWallet.AvailableBalance.Add(amount)
		fundWallet.TotalDeposit = fundWallet.TotalDeposit.Add(amount)
		fundWallet.UpdateTime = now

		if err := tx.Save(fundWallet).Error; err != nil {
			return errors.New("更新资金钱包失败")
		}

		// 记录划转日志
		transfer := &model.FundWalletTransfer{
			TransferNo:     transferNo,
			UserID:         userID,
			TransferType:   model.FundTransferToSpot, // 从现货转入资金钱包
			FromWalletType: model.WalletTypeSpot,
			ToWalletType:   model.WalletTypeFund,
			CurrencyID:     1,
			CurrencyName:   "USDT",
			Amount:         amount,
			Fee:            decimal.Zero,
			BalanceBefore:  balanceBefore,
			BalanceAfter:   fundWallet.AvailableBalance,
			Status:         model.FundTransferSuccess,
			ClientIP:       clientIP,
			CreatedTime:    now,
			CompletedTime:  now,
		}

		return tx.Create(transfer).Error
	})

	if err != nil {
		logger.Errorf("[FundWallet] 充值失败: userID=%d, err=%v", userID, err)
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  err.Error(),
		}
	}

	fundWallet, _ := s.GetOrCreateFundWallet(userID, 1)

	return &TransferResult{
		Success:     true,
		TransferNo:  transferNo,
		FromWallet:  string(model.WalletTypeSpot),
		ToWallet:    string(model.WalletTypeFund),
		Amount:      amount.String(),
		FromBalance: decimal.Zero.String(),
		ToBalance:   fundWallet.AvailableBalance.String(),
	}
}

// WithdrawFromFundWallet 从资金钱包提现到现货账户
func (s *FundWalletService) WithdrawFromFundWallet(userID uint64, amount decimal.Decimal, clientIP string) *TransferResult {
	transferNo := s.GenerateTransferNo()
	now := time.Now().Unix()

	if amount.LessThanOrEqual(decimal.Zero) {
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  "提现金额必须大于0",
		}
	}

	logger.Infof("[FundWallet] 开始提现: userID=%d, amount=%s, transferNo=%s", userID, amount.String(), transferNo)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查并扣减资金钱包
		fundWallet, err := s.getFundWalletForUpdate(tx, userID, 1)
		if err != nil {
			return errors.New("查询资金钱包失败")
		}

		if fundWallet.AvailableBalance.LessThan(amount) {
			return errors.New("资金钱包余额不足")
		}

		balanceBefore := fundWallet.AvailableBalance
		fundWallet.AvailableBalance = fundWallet.AvailableBalance.Sub(amount)
		fundWallet.TotalWithdraw = fundWallet.TotalWithdraw.Add(amount)
		fundWallet.UpdateTime = now

		if err := tx.Save(fundWallet).Error; err != nil {
			return errors.New("更新资金钱包失败")
		}

		// 增加现货账户余额
		_, err = s.updateUserAssets(tx, userID, model.WalletTypeSpot, amount, false)
		if err != nil {
			return err
		}

		// 记录划转日志
		transfer := &model.FundWalletTransfer{
			TransferNo:     transferNo,
			UserID:         userID,
			TransferType:   model.FundTransferFromSpot, // 从资金钱包转到现货
			FromWalletType: model.WalletTypeFund,
			ToWalletType:   model.WalletTypeSpot,
			CurrencyID:     1,
			CurrencyName:   "USDT",
			Amount:         amount,
			Fee:            decimal.Zero,
			BalanceBefore:  balanceBefore,
			BalanceAfter:   fundWallet.AvailableBalance,
			Status:         model.FundTransferSuccess,
			ClientIP:       clientIP,
			CreatedTime:    now,
			CompletedTime:  now,
		}

		return tx.Create(transfer).Error
	})

	if err != nil {
		logger.Errorf("[FundWallet] 提现失败: userID=%d, err=%v", userID, err)
		return &TransferResult{
			Success:   false,
			TransferNo: transferNo,
			ErrorMsg:  err.Error(),
		}
	}

	fundWallet, _ := s.GetOrCreateFundWallet(userID, 1)

	return &TransferResult{
		Success:     true,
		TransferNo:  transferNo,
		FromWallet:  string(model.WalletTypeFund),
		ToWallet:    string(model.WalletTypeSpot),
		Amount:      amount.String(),
		FromBalance: fundWallet.AvailableBalance.String(),
		ToBalance:   decimal.Zero.String(),
	}
}

// ============================================================
// 辅助方法
// ============================================================

// isValidWalletType 验证钱包类型是否有效
func (s *FundWalletService) isValidWalletType(walletType model.WalletType) bool {
	switch walletType {
	case model.WalletTypeSpot, model.WalletTypeContract, model.WalletTypeDelivery, model.WalletTypeFund:
		return true
	default:
		return false
	}
}

// getTransferType 获取划转类型
func (s *FundWalletService) getTransferType(fromWallet, toWallet model.WalletType) int8 {
	if fromWallet == model.WalletTypeFund && toWallet == model.WalletTypeSpot {
		return model.FundTransferFromSpot
	}
	if fromWallet == model.WalletTypeSpot && toWallet == model.WalletTypeFund {
		return model.FundTransferToSpot
	}
	if fromWallet == model.WalletTypeFund && toWallet == model.WalletTypeContract {
		return model.FundTransferToContract
	}
	if fromWallet == model.WalletTypeContract && toWallet == model.WalletTypeFund {
		return model.FundTransferFromContract
	}
	if fromWallet == model.WalletTypeFund && toWallet == model.WalletTypeDelivery {
		return model.FundTransferToDelivery
	}
	if fromWallet == model.WalletTypeDelivery && toWallet == model.WalletTypeFund {
		return model.FundTransferFromDelivery
	}
	return 0
}

// getFundWalletForUpdate 获取资金钱包（带行锁）
func (s *FundWalletService) getFundWalletForUpdate(tx *gorm.DB, userID uint64, currencyID uint64) (*model.FundWallet, error) {
	var wallet model.FundWallet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND currency_id = ?", userID, currencyID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// updateUserAssets 更新用户资产（现货/合约/交割钱包）
func (s *FundWalletService) updateUserAssets(tx *gorm.DB, userID uint64, walletType model.WalletType, amount decimal.Decimal, isDeduct bool) (*model.UserAssets, error) {
	var assets model.UserAssets
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		now := time.Now().Unix()
		assets = model.UserAssets{
			UserID:           uint(userID),
			WalletType:       walletType,
			UsdtBalance:      0,
			UsdtLocked:       0,
			CurrencyBalances: make(model.CurrencyBalance),
			CurrencyLocked:   make(model.CurrencyBalance),
			CreateTime:       now,
			UpdateTime:       now,
		}
		if !isDeduct {
			assets.UsdtBalance, _ = amount.Float64()
		}
		if err := tx.Create(&assets).Error; err != nil {
			return nil, err
		}
		return &assets, nil
	}

	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	updates := map[string]interface{}{
		"update_time": now,
		"version":    gorm.Expr("version + 1"),
	}

	if isDeduct {
		newBalance := decimal.NewFromFloat(assets.UsdtBalance).Sub(amount)
		if newBalance.LessThan(decimal.Zero) {
			return nil, errors.New("余额不足")
		}
		updates["usdt_balance"] = newBalance
	} else {
		newBalance := decimal.NewFromFloat(assets.UsdtBalance).Add(amount)
		updates["usdt_balance"] = newBalance
	}

	if err := tx.Model(&assets).Updates(updates).Error; err != nil {
		return nil, err
	}

	assets.UsdtBalance, _ = updates["usdt_balance"].(decimal.Decimal).Float64()
	return &assets, nil
}

// ============================================================
// 查询方法
// ============================================================

// GetTransferList 获取划转记录列表
func (s *FundWalletService) GetTransferList(userID uint64, page, pageSize int, transferType int8) ([]model.FundWalletTransfer, int64, error) {
	var transfers []model.FundWalletTransfer
	var total int64

	query := database.DB.Model(&model.FundWalletTransfer{}).Where("user_id = ?", userID)

	if transferType > 0 {
		query = query.Where("transfer_type = ?", transferType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&transfers).Error
	if err != nil {
		return nil, 0, err
	}

	return transfers, total, nil
}

// GetTransferDetail 获取划转记录详情
func (s *FundWalletService) GetTransferDetail(transferNo string) (*model.FundWalletTransfer, error) {
	var transfer model.FundWalletTransfer
	err := database.DB.Where("transfer_no = ?", transferNo).First(&transfer).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

// FundWalletOverview 资金钱包总览
type FundWalletOverview struct {
	TotalBalance    decimal.Decimal `json:"total_balance"`    // 总余额
	TotalDeposit    decimal.Decimal `json:"total_deposit"`    // 累计充值
	TotalWithdraw   decimal.Decimal `json:"total_withdraw"`   // 累计提现
	TransferIn24h   decimal.Decimal `json:"transfer_in_24h"`  // 24小时转入
	TransferOut24h  decimal.Decimal `json:"transfer_out_24h"` // 24小时转出
}

// GetOverview 获取资金钱包总览
func (s *FundWalletService) GetOverview(userID uint64) (*FundWalletOverview, error) {
	wallet, err := s.GetOrCreateFundWallet(userID, 1)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	oneDayAgo := now - 86400

	// 24小时转入转出统计
	var in24h, out24h decimal.Decimal
	database.DB.Model(&model.FundWalletTransfer{}).
		Where("user_id = ? AND status = ? AND created_time > ?", userID, model.FundTransferSuccess, oneDayAgo).
		Where("to_wallet_type = ?", model.WalletTypeFund).
		Select("COALESCE(SUM(amount), 0)").Scan(&in24h)

	database.DB.Model(&model.FundWalletTransfer{}).
		Where("user_id = ? AND status = ? AND created_time > ?", userID, model.FundTransferSuccess, oneDayAgo).
		Where("from_wallet_type = ?", model.WalletTypeFund).
		Select("COALESCE(SUM(amount), 0)").Scan(&out24h)

	return &FundWalletOverview{
		TotalBalance:   wallet.AvailableBalance.Add(wallet.LockedBalance),
		TotalDeposit:   wallet.TotalDeposit,
		TotalWithdraw:  wallet.TotalWithdraw,
		TransferIn24h:  in24h,
		TransferOut24h: out24h,
	}, nil
}
