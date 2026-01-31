package admin

import (
	"errors"

	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// TransactionAdminService 交易管理服务
type TransactionAdminService struct{}

// NewTransactionAdminService 创建交易管理服务实例
func NewTransactionAdminService() *TransactionAdminService {
	return &TransactionAdminService{}
}

// GetSpotTransactionList 获取币币交易列表
func (s *TransactionAdminService) GetSpotTransactionList(info PageInfo, userID uint, status int) (list []model.Transaction, total int64, err error) {
	db := database.DB.Model(&model.Transaction{})

	if userID > 0 {
		db = db.Where("from_user_id = ? OR to_user_id = ?", userID, userID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetLeverTransactionList 获取合约交易列表
func (s *TransactionAdminService) GetLeverTransactionList(info PageInfo, userID uint, status int) (list []model.LeverTransaction, total int64, err error) {
	db := database.DB.Model(&model.LeverTransaction{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetMicroTransactionList 获取秒合约交易列表
func (s *TransactionAdminService) GetMicroTransactionList(info PageInfo, userID uint, status int) (list []model.MicroOrder, total int64, err error) {
	db := database.DB.Model(&model.MicroOrder{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// CancelSpotOrder 取消币币订单
// TODO: 需要根据实际的订单表结构重新实现
func (s *TransactionAdminService) CancelSpotOrder(orderID uint) error {
	return errors.New("功能暂未实现，需要完善订单表结构")
}

// GetLegalDealList 获取法币交易列表
func (s *TransactionAdminService) GetLegalDealList(info PageInfo, status int) (list []model.LegalDeal, total int64, err error) {
	db := database.DB.Model(&model.LegalDeal{})

	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetC2cDealList 获取C2C交易列表
func (s *TransactionAdminService) GetC2cDealList(info PageInfo, status int) (list []model.C2cDeal, total int64, err error) {
	db := database.DB.Model(&model.C2cDeal{})

	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetAccountLogList 获取账户流水列表
func (s *TransactionAdminService) GetAccountLogList(info PageInfo, userID uint, logType int) (list []model.AccountLog, total int64, err error) {
	db := database.DB.Model(&model.AccountLog{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if logType > 0 {
		db = db.Where("type = ?", logType)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetSellerList 获取商家列表
func (s *TransactionAdminService) GetSellerList(info PageInfo, status int) (list []model.Seller, total int64, err error) {
	db := database.DB.Model(&model.Seller{})

	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ApproveSeller 审批商家
func (s *TransactionAdminService) ApproveSeller(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var seller model.Seller
		if err := tx.Where("id = ?", id).First(&seller).Error; err != nil {
			return err
		}

		// 更新商家状态
		if err := tx.Model(&seller).Update("status", 1).Error; err != nil {
			return err
		}

		// 更新用户商家标识
		return tx.Model(&model.User{}).Where("id = ?", seller.UserID).Update("is_seller", 1).Error
	})
}

// RejectSeller 拒绝商家申请
func (s *TransactionAdminService) RejectSeller(id uint, reason string) error {
	return database.DB.Model(&model.Seller{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": 2,
		"reason": reason,
	}).Error
}

// GetMicroConfigList 获取秒合约配置列表
func (s *TransactionAdminService) GetMicroConfigList() (numbers []model.MicroNumber, seconds []model.MicroSeconds, err error) {
	err = database.DB.Find(&numbers).Error
	if err != nil {
		return
	}
	err = database.DB.Find(&seconds).Error
	return
}

// UpdateMicroSeconds 更新秒合约时间配置
func (s *TransactionAdminService) UpdateMicroSeconds(item model.MicroSeconds) error {
	return database.DB.Save(&item).Error
}

// GetLeverMultipleList 获取杠杆倍数配置列表（全局统一）
func (s *TransactionAdminService) GetLeverMultipleList() (list []model.LeverMultiple, err error) {
	// 只查询currency_id为null的全局配置
	err = database.DB.Where("currency_id IS NULL").Find(&list).Error
	return
}

// UpdateLeverMultiple 更新杠杆倍数配置
func (s *TransactionAdminService) UpdateLeverMultiple(item model.LeverMultiple) error {
	return database.DB.Save(&item).Error
}
