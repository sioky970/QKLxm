package admin

import (
	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
)

// RiskAdminService 风控管理服务
type RiskAdminService struct{}

// NewRiskAdminService 创建风控管理服务实例
func NewRiskAdminService() *RiskAdminService {
	return &RiskAdminService{}
}

// RiskConfig 风控配置
type RiskConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

// GetRiskConfig 获取风控配置
func (s *RiskAdminService) GetRiskConfig() ([]RiskConfig, error) {
	var settings []model.Setting
	err := database.DB.Where("`key` LIKE ?", "risk_%").Find(&settings).Error
	if err != nil {
		return nil, err
	}

	configs := make([]RiskConfig, len(settings))
	for i, setting := range settings {
		configs[i] = RiskConfig{
			Key:   setting.Key,
			Value: setting.Value,
			Name:  setting.Name,
		}
	}
	return configs, nil
}

// UpdateRiskConfig 更新风控配置 (支持新增和修改)
func (s *RiskAdminService) UpdateRiskConfig(key, value string) error {
	var setting model.Setting
	// 尝试查找现有配置
	err := database.DB.Where("`key` = ?", key).First(&setting).Error

	if err != nil {
		// 如果不存在则创建
		setting = model.Setting{
			Key:   key,
			Value: value,
		}
		err = database.DB.Create(&setting).Error
	} else {
		// 如果存在则更新
		err = database.DB.Model(&setting).Update("value", value).Error
	}

	if err == nil {
		// 清除风控配置缓存，确保立即生效
		cache.Delete("risk:config")
	}
	return err
}

// UserRiskInfo 用户风控信息
type UserRiskInfo struct {
	ID      uint   `json:"id"`
	Account string `json:"account"`
	Risk    int8   `json:"risk"`
}

// GetUserRiskList 获取用户风控列表
func (s *RiskAdminService) GetUserRiskList(info PageInfo) (list []UserRiskInfo, total int64, err error) {
	db := database.DB.Model(&model.User{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var users []model.User
	err = db.Select("id, account_number, risk").
		Limit(info.GetLimit()).Offset(info.GetOffset()).
		Order("id DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	list = make([]UserRiskInfo, len(users))
	for i, user := range users {
		list[i] = UserRiskInfo{
			ID:      user.ID,
			Account: user.AccountNumber,
			Risk:    user.Risk,
		}
	}
	return list, total, nil
}

// SetUserRisk 设置用户风控
func (s *RiskAdminService) SetUserRisk(userID uint, risk int8) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("risk", risk).Error
}

// BatchSetUserRisk 批量设置用户风控
func (s *RiskAdminService) BatchSetUserRisk(userIDs []uint, risk int8) error {
	return database.DB.Model(&model.User{}).Where("id IN ?", userIDs).Update("risk", risk).Error
}

// CurrencyMatchRiskInfo 交易对风控信息
type CurrencyMatchRiskInfo struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	RiskGroupResult int8   `json:"risk_group_result"`
}

// GetCurrencyMatchRiskList 获取交易对风控列表
func (s *RiskAdminService) GetCurrencyMatchRiskList() (list []CurrencyMatchRiskInfo, err error) {
	// 通过JOIN获取交易对名称
	err = database.DB.Table("currency_matches AS cm").
		Select("cm.id, CONCAT(c1.name, '/', c2.name) AS name, cm.risk_group_result").
		Joins("LEFT JOIN currency c1 ON cm.currency_id = c1.id").
		Joins("LEFT JOIN currency c2 ON cm.legal_id = c2.id").
		Find(&list).Error
	return list, err
}

// SetCurrencyMatchRisk 设置交易对风控
func (s *RiskAdminService) SetCurrencyMatchRisk(matchID uint, riskResult int8) error {
	return database.DB.Model(&model.CurrencyMatch{}).Where("id = ?", matchID).Update("risk_group_result", riskResult).Error
}

// OrderRiskInfo 订单风控信息
type OrderRiskInfo struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	Type            string `json:"type"`
	PreProfitResult int8   `json:"pre_profit_result"`
}

// GetMicroOrderRiskList 获取秒合约订单风控列表
func (s *RiskAdminService) GetMicroOrderRiskList(info PageInfo) (list []model.MicroOrder, total int64, err error) {
	db := database.DB.Model(&model.MicroOrder{}).Where("status = ?", 0)

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// SetMicroOrderRisk 设置秒合约订单风控（预设盈亏）
func (s *RiskAdminService) SetMicroOrderRisk(orderID uint, preProfitResult int8) error {
	return database.DB.Model(&model.MicroOrder{}).Where("id = ?", orderID).Update("pre_profit_result", preProfitResult).Error
}

// GetKYCList 获取实名认证列表
func (s *RiskAdminService) GetKYCList(info PageInfo, status *int) (list []model.UserReal, total int64, err error) {
	db := database.DB.Model(&model.UserReal{})

	if status != nil {
		db = db.Where("status = ?", *status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetKYCDetail 获取实名认证详情
func (s *RiskAdminService) GetKYCDetail(id uint) (kyc *model.UserReal, err error) {
	var k model.UserReal
	err = database.DB.Where("id = ?", id).First(&k).Error
	return &k, err
}

// ApproveKYC 审核通过实名认证
func (s *RiskAdminService) ApproveKYC(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var kyc model.UserReal
		if err := tx.Where("id = ?", id).First(&kyc).Error; err != nil {
			return err
		}

		// 更新实名认证状态
		if err := tx.Model(&kyc).Update("status", 1).Error; err != nil {
			return err
		}

		// 更新用户实名状态
		return tx.Model(&model.User{}).Where("id = ?", kyc.UserID).Updates(map[string]interface{}{
			"is_realname": 2,
		}).Error
	})
}

// RejectKYC 拒绝实名认证
func (s *RiskAdminService) RejectKYC(id uint, reason string) error {
	return database.DB.Model(&model.UserReal{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        2,
		"refuse_reason": reason,
	}).Error
}
