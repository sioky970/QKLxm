package admin

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/service"
)

// WalletAdminService 钱包管理服务
type WalletAdminService struct{}

// NewWalletAdminService 创建钱包管理服务实例
func NewWalletAdminService() *WalletAdminService {
	return &WalletAdminService{}
}

// GetWalletList 获取钱包列表
func (s *WalletAdminService) GetWalletList(info PageInfo) (list []model.UsersWallet, total int64, err error) {
	db := database.DB.Model(&model.UsersWallet{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Find(&list).Error
	return list, total, err
}

// WalletListItem 钱包列表项（带用户与币种信息）
type WalletListItem struct {
	model.UsersWallet
	AccountNumber string `json:"account_number"`
	CurrencyName  string `json:"currency_name"`
}

// WalletTotals 钱包余额汇总
type WalletTotals struct {
	UsdtBalance     float64 `json:"usdt_balance"`
	LockUsdtBalance float64 `json:"lock_usdt_balance"`
}

// GetWalletListAdvanced 获取钱包列表（支持筛选）
func (s *WalletAdminService) GetWalletListAdvanced(accountNumber string, currencyID uint, status *int, address string, page, pageSize int) (list []WalletListItem, total int64, err error) {
	db := database.DB.Table("users_wallet AS w").
		Select("w.*, u.account_number, c.name AS currency_name").
		Joins("LEFT JOIN users u ON w.user_id = u.id").
		Joins("LEFT JOIN currency c ON w.currency = c.id")
	db = db.Where("c.is_display = 1")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID > 0 {
		db = db.Where("w.currency = ?", currencyID)
	}
	if status != nil {
		db = db.Where("w.status = ?", *status)
	}
	if address != "" {
		db = db.Where("w.address LIKE ?", "%"+address+"%")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	if err = db.Count(&total).Error; err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("w.id DESC").Find(&list).Error
	return
}

// GetWalletTotals 获取钱包余额汇总（支持筛选）
func (s *WalletAdminService) GetWalletTotals(accountNumber string, currencyID uint, status *int, address string) (totals WalletTotals, err error) {
	db := database.DB.Table("users_wallet AS w").
		Select("COALESCE(SUM(w.legal_balance), 0) AS legal_balance, COALESCE(SUM(w.change_balance), 0) AS change_balance, COALESCE(SUM(w.lever_balance), 0) AS lever_balance").
		Joins("LEFT JOIN users u ON w.user_id = u.id").
		Joins("LEFT JOIN currency c ON w.currency = c.id")
	db = db.Where("c.is_display = 1")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID > 0 {
		db = db.Where("w.currency = ?", currencyID)
	}
	if status != nil {
		db = db.Where("w.status = ?", *status)
	}
	if address != "" {
		db = db.Where("w.address LIKE ?", "%"+address+"%")
	}

	err = db.Scan(&totals).Error
	return
}

// GetWalletInfo 获取钱包详情
func (s *WalletAdminService) GetWalletInfo(id uint) (wallet *model.UsersWallet, err error) {
	var w model.UsersWallet
	err = database.DB.Where("id = ?", id).First(&w).Error
	return &w, err
}

// UpdateBalance 更新钱包余额
func (s *WalletAdminService) UpdateBalance(id uint, balanceType string, amount float64, reason string) (err error) {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var wallet model.UsersWallet
		if err := tx.Where("id = ?", id).First(&wallet).Error; err != nil {
			return err
		}

		// 根据类型更新相应余额
		updates := map[string]interface{}{}
		switch balanceType {
		case "legal":
			updates["legal_balance"] = gorm.Expr("legal_balance + ?", amount)
		case "change":
			updates["change_balance"] = gorm.Expr("change_balance + ?", amount)
		case "lever":
			updates["lever_balance"] = gorm.Expr("lever_balance + ?", amount)
		case "micro":
			updates["micro_balance"] = gorm.Expr("micro_balance + ?", amount)
		default:
			return errors.New("无效的余额类型")
		}

		if err := tx.Model(&wallet).Updates(updates).Error; err != nil {
			return err
		}

		// 记录余额变动日志
		accountLog := model.AccountLog{
			UserID:      wallet.UserID,
			Value:       amount,
			CurrencyID:  wallet.CurrencyID, // 使用CurrencyID
			Type:        1,                 // 后台调节
			Info:        reason,
			CreatedTime: time.Now().Unix(),
		}
		if err := tx.Create(&accountLog).Error; err != nil {
			return err
		}

		// WebSocket实时推送余额变更
		go service.GetWalletService().BroadcastBalanceUpdate(wallet.UserID)

		return nil
	})
}

// FreezeWallet 冻结钱包
func (s *WalletAdminService) FreezeWallet(id uint) (err error) {
	return database.DB.Model(&model.UsersWallet{}).Where("id = ?", id).Update("status", 0).Error
}

// ActivateWallet 激活钱包
func (s *WalletAdminService) ActivateWallet(id uint) (err error) {
	return database.DB.Model(&model.UsersWallet{}).Where("id = ?", id).Update("status", 1).Error
}

// GetWithdrawalsList 获取提币列表
func (s *WalletAdminService) GetWithdrawalsList(info PageInfo) (list []model.UsersWalletOut, total int64, err error) {
	db := database.DB.Model(&model.UsersWalletOut{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetWithdrawalsListWithFilter 带筛选条件的提币列表
func (s *WalletAdminService) GetWithdrawalsListWithFilter(userID uint, currencyID uint, status int, page, pageSize int) (list []model.UsersWalletOut, total int64, err error) {
	db := database.DB.Model(&model.UsersWalletOut{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if currencyID > 0 {
		db = db.Where("currency = ?", currencyID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("id DESC").Find(&list).Error
	return
}

// ApproveWithdrawal 审核通过提币
func (s *WalletAdminService) ApproveWithdrawal(id uint) (err error) {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var walletOut model.UsersWalletOut
		if err := tx.First(&walletOut, id).Error; err != nil {
			return err
		}

		if walletOut.Status != 1 {
			return errors.New("该申请已处理")
		}

		// 更新状态为已通过
		if err := tx.Model(&walletOut).Updates(map[string]interface{}{
			"status":      2,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
			return err
		}

		// 记录账户日志
		accountLog := model.AccountLog{
			UserID:      walletOut.UserID,
			Value:       -walletOut.Number,
			CurrencyID:  walletOut.CurrencyID, // 使用CurrencyID
			Type:        34,                   // 提币成功
			Info:        "提币审核通过",
			CreatedTime: time.Now().Unix(),
		}
		return tx.Create(&accountLog).Error
	})
}

// RejectWithdrawal 拒绝提币（退还余额）
func (s *WalletAdminService) RejectWithdrawal(id uint, reason string) (err error) {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var walletOut model.UsersWalletOut
		if err := tx.First(&walletOut, id).Error; err != nil {
			return err
		}

		if walletOut.Status != 1 {
			return errors.New("该申请已处理")
		}

		// 退还余额到法币账户
		if err := tx.Model(&model.UsersWallet{}).
			Where("user_id = ? AND currency = ?", walletOut.UserID, walletOut.CurrencyID).
			Update("usdt_balance", gorm.Expr("usdt_balance + ?", walletOut.Number)).Error; err != nil {
			return err
		}

		// 更新提币状态为已拒绝
		if err := tx.Model(&walletOut).Updates(map[string]interface{}{
			"status":      3,
			"notes":       reason,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
			return err
		}

		// 记录账户日志
		accountLog := model.AccountLog{
			UserID:      walletOut.UserID,
			Value:       walletOut.Number,
			CurrencyID:  walletOut.CurrencyID, // 使用CurrencyID
			Type:        35,                   // 提币退还
			Info:        "提币被拒绝，余额已退还: " + reason,
			CreatedTime: time.Now().Unix(),
		}
		if err := tx.Create(&accountLog).Error; err != nil {
			return err
		}

		// WebSocket实时推送余额变更
		go service.GetWalletService().BroadcastBalanceUpdate(walletOut.UserID)

		return nil
	})
}

// GetWithdrawalDetail 获取提币详情
func (s *WalletAdminService) GetWithdrawalDetail(id uint) (walletOut *model.UsersWalletOut, err error) {
	var w model.UsersWalletOut
	err = database.DB.Where("id = ?", id).First(&w).Error
	return &w, err
}

// WithdrawalListItem 提币列表项（带用户信息）
type WithdrawalListItem struct {
	model.UsersWalletOut
	AccountNumber string `json:"account_number"`
	CurrencyName  string `json:"currency_name"`
}

// GetWithdrawalsListAdvanced 带高级筛选的提币列表
func (s *WalletAdminService) GetWithdrawalsListAdvanced(accountNumber string, currencyID uint, status *int, page, pageSize int) (list []WithdrawalListItem, total int64, err error) {
	db := database.DB.Table("users_wallet_out AS wo").
		Select("wo.*, u.account_number, c.name AS currency_name").
		Joins("LEFT JOIN users u ON wo.user_id = u.id").
		Joins("LEFT JOIN currency c ON wo.currency = c.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID > 0 {
		db = db.Where("wo.currency = ?", currencyID)
	}
	if status != nil {
		db = db.Where("wo.status = ?", *status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("wo.id DESC").Find(&list).Error
	return
}

// ApproveWithdrawalAdvanced 审核通过提币（增强版）
func (s *WalletAdminService) ApproveWithdrawalAdvanced(id uint, method, txid, notes string) (err error) {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var walletOut model.UsersWalletOut
		if err := tx.First(&walletOut, id).Error; err != nil {
			return err
		}

		if walletOut.Status != 1 {
			return errors.New("该申请已处理")
		}

		// 更新状态为已通过
		updates := map[string]interface{}{
			"status":      2,
			"update_time": time.Now().Unix(),
		}

		if method == "manual" && txid != "" {
			updates["hash"] = txid
		}
		if notes != "" {
			updates["notes"] = notes
		}

		if err := tx.Model(&walletOut).Updates(updates).Error; err != nil {
			return err
		}

		// 记录账户日志
		accountLog := model.AccountLog{
			UserID:      walletOut.UserID,
			Value:       -walletOut.Number,
			CurrencyID:  walletOut.CurrencyID, // 使用CurrencyID替代Currency
			Type:        34,                   // 提币成功
			Info:        "提币审核通过",
			CreatedTime: time.Now().Unix(),
		}
		return tx.Create(&accountLog).Error
	})
}

// GetChargeList 获取充值列表
func (s *WalletAdminService) GetChargeList(info PageInfo) (list []model.ChargeReq, total int64, err error) {
	db := database.DB.Model(&model.ChargeReq{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ChargeListItem 充值列表项（带币种信息）
type ChargeListItem struct {
	model.ChargeReq
	CurrencyName string `json:"currency_name"`
}

// GetChargeListAdvanced 获取充值列表（支持筛选）
func (s *WalletAdminService) GetChargeListAdvanced(accountNumber string, currencyID uint, status *int, page, pageSize int) (list []ChargeListItem, total int64, err error) {
	db := database.DB.Table("charge_req AS cr").
		Select("cr.*, c.name AS currency_name").
		Joins("LEFT JOIN currency c ON cr.currency_id = c.id")

	if accountNumber != "" {
		db = db.Where("cr.user_account LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID > 0 {
		db = db.Where("cr.currency_id = ?", currencyID)
	}
	if status != nil {
		db = db.Where("cr.status = ?", *status)
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	if err = db.Count(&total).Error; err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("cr.id DESC").Find(&list).Error
	return
}

// ApproveCharge 审核通过充值
func (s *WalletAdminService) ApproveCharge(id uint) (err error) {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var charge model.ChargeReq
		if err := tx.First(&charge, id).Error; err != nil {
			return err
		}

		if charge.Status != 1 {
			return errors.New("该申请已处理")
		}

		// 增加用户余额
		if err := tx.Model(&model.UsersWallet{}).
			Where("user_id = ? AND currency = ?", charge.UID, charge.CurrencyID).
			Update("legal_balance", gorm.Expr("legal_balance + ?", charge.Amount)).Error; err != nil {
			return err
		}

		// 更新状态为已通过
		if err := tx.Model(&charge).Update("status", 2).Error; err != nil {
			return err
		}

		// 记录账户日志
		accountLog := model.AccountLog{
			UserID:      uint(charge.UID),
			Value:       charge.Amount,
			CurrencyID:  charge.CurrencyID, // 使用CurrencyID替代Currency
			Type:        31,                // 充值成功
			Info:        "充值审核通过",
			CreatedTime: time.Now().Unix(),
		}
		if err := tx.Create(&accountLog).Error; err != nil {
			return err
		}

		// WebSocket实时推送余额变更
		go service.GetWalletService().BroadcastBalanceUpdate(uint(charge.UID))

		return nil
	})
}

// RejectCharge 拒绝充值
func (s *WalletAdminService) RejectCharge(id uint, reason string) (err error) {
	return database.DB.Model(&model.ChargeReq{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": 3,
		"remark": reason,
	}).Error
}
