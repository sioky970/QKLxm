package admin

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
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
// 注意：已迁移到UserAssets表，此方法返回基于UserAssets的数据
func (s *WalletAdminService) GetWalletList(info PageInfo) (list []model.UserAssets, total int64, err error) {
	db := database.DB.Model(&model.UserAssets{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Find(&list).Error
	return list, total, err
}

// WalletListItem 钱包列表项（带用户与币种信息）
// 注意：已迁移到UserAssets表
type WalletListItem struct {
	UserID          uint    `json:"user_id"`
	UsdtBalance     float64 `json:"usdt_balance"`
	LockUsdtBalance float64 `json:"lock_usdt_balance"`
	AccountNumber   string  `json:"account_number"`
	CurrencyName    string  `json:"currency_name"`
}

// WalletTotals 钱包余额汇总
type WalletTotals struct {
	UsdtBalance     float64 `json:"usdt_balance"`
	LockUsdtBalance float64 `json:"lock_usdt_balance"`
}

// GetWalletListAdvanced 获取钱包列表（支持筛选）
// 注意：已迁移到UserAssets表
func (s *WalletAdminService) GetWalletListAdvanced(accountNumber string, currencyID uint, status *int, address string, page, pageSize int) (list []WalletListItem, total int64, err error) {
	// 从UserAssets表查询
	db := database.DB.Table("user_assets AS a").
		Select("a.user_id, a.usdt_balance, a.usdt_locked AS lock_usdt_balance, u.account_number, 'USDT' AS currency_name").
		Joins("LEFT JOIN users u ON a.user_id = u.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Limit(pageSize).Offset(offset).Order("a.user_id DESC").Find(&list).Error
	return
}

// GetWalletTotals 获取钱包余额汇总（支持筛选）
// 注意：已迁移到UserAssets表
func (s *WalletAdminService) GetWalletTotals(accountNumber string, currencyID uint, status *int, address string) (totals WalletTotals, err error) {
	db := database.DB.Table("user_assets AS a").
		Select("COALESCE(SUM(a.usdt_balance), 0) AS usdt_balance, COALESCE(SUM(a.usdt_locked), 0) AS lock_usdt_balance").
		Joins("LEFT JOIN users u ON a.user_id = u.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}

	err = db.Scan(&totals).Error
	return
}

// GetWalletInfo 获取钱包详情
// 注意：已迁移到UserAssets表
func (s *WalletAdminService) GetWalletInfo(id uint) (wallet *model.UserAssets, err error) {
	var w model.UserAssets
	err = database.DB.Where("user_id = ?", id).First(&w).Error
	return &w, err
}

// UpdateBalance 更新钱包余额
// 注意：已迁移到UserAssets表，支持指定钱包类型（现货/合约）
func (s *WalletAdminService) UpdateBalance(id uint, balanceType string, amount float64, reason string, walletType string) (err error) {
	wt := model.WalletTypeSpot
	if walletType == "contract" {
		wt = model.WalletTypeContract
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		if err := tx.Where("user_id = ? AND wallet_type = ?", id, wt).First(&assets).Error; err != nil {
			return errors.New("找不到指定的钱包记录，请确认用户的钱包是否存在")
		}

		// 根据类型更新相应余额
		switch balanceType {
		case "usdt":
			// 更新USDT余额
			newBalance := assets.UsdtBalance + amount
			if newBalance < 0 {
				return errors.New("余额不足")
			}
			if err := tx.Model(&assets).Update("usdt_balance", newBalance).Error; err != nil {
				return err
			}
		case "lock_usdt":
			// 更新锁定USDT余额
			newLocked := assets.UsdtLocked + amount
			if newLocked < 0 {
				return errors.New("锁定余额不足")
			}
			if err := tx.Model(&assets).Update("usdt_locked", newLocked).Error; err != nil {
				return err
			}
		default:
			return errors.New("未知的余额类型")
		}

		// 记录操作日志
		log := &model.WalletLog{
			UserID:      id,
			BalanceType: 1,
			Change:      amount,
			Memo:        fmt.Sprintf("%s | 钱包类型: %s | 原因: %s", 
				map[bool]string{true: "增加", false: "减少"}[amount > 0], walletType, reason),
			CreateTime:  time.Now().Unix(),
		}
		return tx.Create(log).Error
	})
}

// FreezeWallet 冻结钱包
// 注意：UserAssets表没有status字段，此方法仅作兼容
func (s *WalletAdminService) FreezeWallet(id uint) (err error) {
	// UserAssets表没有status字段，无法冻结
	// 可以在这里添加其他冻结逻辑，如设置标记
	return nil
}

// ActivateWallet 激活钱包
// 注意：UserAssets表没有status字段，此方法仅作兼容
func (s *WalletAdminService) ActivateWallet(id uint) (err error) {
	// UserAssets表没有status字段，无法激活
	// 可以在这里添加其他激活逻辑
	return nil
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
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Limit(pageSize).Offset(offset).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetWithdrawalsListWithExtra 带额外字段的提币列表（包含用户账号和币种名称）
func (s *WalletAdminService) GetWithdrawalsListWithExtra(userID uint, currencyID uint, status int, page, pageSize int) (interface{}, int64, error) {
	type WithdrawalWithExtra struct {
		model.UsersWalletOut
		AccountNumber string `json:"account_number"`
		CurrencyName  string `json:"currency_name"`
	}

	var results []WithdrawalWithExtra

	// 构建查询，关联用户表和币种表
	query := database.DB.Table("users_wallet_out AS w").
		Select(`w.*,
			u.account_number,
			c.name AS currency_name`).
		Joins("LEFT JOIN users u ON w.user_id = u.id").
		Joins("LEFT JOIN currency c ON w.currency = c.id")

	if userID > 0 {
		query = query.Where("w.user_id = ?", userID)
	}
	if currencyID > 0 {
		query = query.Where("w.currency = ?", currencyID)
	}
	if status >= 0 {
		query = query.Where("w.status = ?", status)
	}

	// 先查询总数
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Limit(pageSize).Offset(offset).Order("w.id DESC").Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetWithdrawalDetail 获取提币详情
func (s *WalletAdminService) GetWithdrawalDetail(id uint) (*model.UsersWalletOut, error) {
	var withdrawal model.UsersWalletOut
	err := database.DB.First(&withdrawal, id).Error
	return &withdrawal, err
}

// ApproveWithdrawal 审核通过提币
func (s *WalletAdminService) ApproveWithdrawal(id uint, operatorID uint, operatorName string, remark string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var withdrawal model.UsersWalletOut
		if err := tx.First(&withdrawal, id).Error; err != nil {
			return err
		}

		if withdrawal.Status != 1 {
			return errors.New("只能审核待处理的提币申请")
		}

		// 更新状态为已通过
		updates := map[string]interface{}{
			"status":      2, // 已通过
			"update_time": time.Now().Unix(),
		}
		if remark != "" {
			updates["notes"] = withdrawal.Notes + " | 审核通过: " + remark
		}

		return tx.Model(&withdrawal).Updates(updates).Error
	})
}

// RejectWithdrawal 拒绝提币
func (s *WalletAdminService) RejectWithdrawal(id uint, operatorID uint, operatorName string, reason string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var withdrawal model.UsersWalletOut
		if err := tx.First(&withdrawal, id).Error; err != nil {
			return err
		}

		if withdrawal.Status != 1 {
			return errors.New("只能拒绝待处理的提币申请")
		}

		// 返还资金钱包的锁定余额到可用余额
		if withdrawal.CurrencyID > 0 {
			fundSvc := service.GetFundWalletService()
			amount := decimal.NewFromFloat(withdrawal.Number)
			
			// 获取资金钱包
			fundWallet, err := fundSvc.GetOrCreateFundWallet(uint64(withdrawal.UserID), 1)
			if err != nil {
				return fmt.Errorf("获取资金钱包失败: %w", err)
			}

			// 从锁定余额返还到可用余额
			fundWallet.LockedBalance = fundWallet.LockedBalance.Sub(amount)
			fundWallet.AvailableBalance = fundWallet.AvailableBalance.Add(amount)
			fundWallet.TotalWithdraw = fundWallet.TotalWithdraw.Sub(amount)
			fundWallet.UpdateTime = time.Now().Unix()

			if err := tx.Save(fundWallet).Error; err != nil {
				return fmt.Errorf("返还资金钱包余额失败: %w", err)
			}
		}

		// 更新状态为已拒绝
		updates := map[string]interface{}{
			"status":      3, // 已拒绝
			"update_time": time.Now().Unix(),
		}
		if reason != "" {
			updates["notes"] = withdrawal.Notes + " | 拒绝原因: " + reason
		}

		return tx.Model(&withdrawal).Updates(updates).Error
	})
}

// CompleteWithdrawal 完成提币（打款完成）
func (s *WalletAdminService) CompleteWithdrawal(id uint, txHash string, operatorID uint, operatorName string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var withdrawal model.UsersWalletOut
		if err := tx.First(&withdrawal, id).Error; err != nil {
			return err
		}

		if withdrawal.Status != 2 {
			return errors.New("只能完成已通过的提币申请")
		}

		// 从资金钱包扣除锁定余额（实际打款已完成）
		if withdrawal.CurrencyID > 0 {
			fundSvc := service.GetFundWalletService()
			amount := decimal.NewFromFloat(withdrawal.Number)
			
			// 获取资金钱包
			fundWallet, err := fundSvc.GetOrCreateFundWallet(uint64(withdrawal.UserID), 1)
			if err != nil {
				return fmt.Errorf("获取资金钱包失败: %w", err)
			}

			// 检查锁定余额是否足够
			if fundWallet.LockedBalance.LessThan(amount) {
				return errors.New("资金钱包锁定余额不足")
			}

			// 扣除锁定余额（资金已实际转出）
			fundWallet.LockedBalance = fundWallet.LockedBalance.Sub(amount)
			fundWallet.UpdateTime = time.Now().Unix()

			if err := tx.Save(fundWallet).Error; err != nil {
				return fmt.Errorf("扣除资金钱包锁定余额失败: %w", err)
			}
		}

		// 更新状态为已完成
		updates := map[string]interface{}{
			"status":      4, // 已完成
			"update_time": time.Now().Unix(),
		}
		if txHash != "" {
			updates["tx_hash"] = txHash
		}

		return tx.Model(&withdrawal).Updates(updates).Error
	})
}

// GetRechargeList 获取充值列表
func (s *WalletAdminService) GetRechargeList(info PageInfo) (list []model.ChargeReq, total int64, err error) {
	db := database.DB.Model(&model.ChargeReq{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetRechargeListWithFilter 带筛选条件的充值列表
func (s *WalletAdminService) GetRechargeListWithFilter(userID uint, currencyID uint, status int, page, pageSize int) (list []model.ChargeReq, total int64, err error) {
	db := database.DB.Model(&model.ChargeReq{})

	if userID > 0 {
		db = db.Where("uid = ?", userID)
	}
	if currencyID > 0 {
		db = db.Where("currency_id = ?", currencyID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Limit(pageSize).Offset(offset).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ApproveRecharge 审核通过充值
func (s *WalletAdminService) ApproveRecharge(id uint, operatorID uint, operatorName string, remark string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var recharge model.ChargeReq
		if err := tx.First(&recharge, id).Error; err != nil {
			return err
		}

		if recharge.Status != 1 {
			return errors.New("只能审核待处理的充值申请")
		}

		// 增加用户余额
		assetsSvc := service.GetUserAssetsService()
		if err := assetsSvc.UpdateUsdtBalance(uint(recharge.UID), recharge.Amount, false); err != nil {
			return err
		}

		// 更新状态为已通过
		updates := map[string]interface{}{
			"status":     2, // 已通过
			"updated_at": time.Now().Unix(),
		}
		if remark != "" {
			updates["remark"] = remark
		}

		return tx.Model(&recharge).Updates(updates).Error
	})
}

// RejectRecharge 拒绝充值
func (s *WalletAdminService) RejectRecharge(id uint, operatorID uint, operatorName string, reason string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var recharge model.ChargeReq
		if err := tx.First(&recharge, id).Error; err != nil {
			return err
		}

		if recharge.Status != 1 {
			return errors.New("只能拒绝待处理的充值申请")
		}

		// 更新状态为已拒绝
		updates := map[string]interface{}{
			"status":     3, // 已拒绝
			"updated_at": time.Now().Unix(),
		}
		if reason != "" {
			updates["remark"] = reason
		}

		return tx.Model(&recharge).Updates(updates).Error
	})
}
