package admin

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/service"
)

// UserAdminService 用户管理服务
type UserAdminService struct{}

// NewUserAdminService 创建用户管理服务实例
func NewUserAdminService() *UserAdminService {
	return &UserAdminService{}
}

// GetUserList 获取用户列表
func (s *UserAdminService) GetUserList(info PageInfo) (list []model.User, total int64, err error) {
	db := database.DB.Model(&model.User{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetUserInfo 获取用户详情
func (s *UserAdminService) GetUserInfo(id uint) (user *model.User, err error) {
	var u model.User
	err = database.DB.Where("id = ?", id).First(&u).Error
	return &u, err
}

// UpdateUserInfo 更新用户信息
func (s *UserAdminService) UpdateUserInfo(user model.User) (err error) {
	var existingUser model.User
	err = database.DB.Where("id = ?", user.ID).First(&existingUser).Error
	if err != nil {
		return err
	}

	return database.DB.Save(&user).Error
}

// UpdateUserFields 更新用户指定字段
func (s *UserAdminService) UpdateUserFields(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("未提交任何可更新字段")
	}

	var existing model.User
	if err := database.DB.Select("id").Where("id = ?", id).First(&existing).Error; err != nil {
		return err
	}

	if password, ok := updates["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password"] = string(hashedPassword)
	} else {
		delete(updates, "password")
	}

	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")

	result := database.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户信息未发生变化")
	}
	return nil
}

// DeleteUserInfo 删除用户
func (s *UserAdminService) DeleteUserInfo(id uint) (err error) {
	return database.DB.Delete(&model.User{}, id).Error
}

// FreezeUser 冻结用户
func (s *UserAdminService) FreezeUser(id uint) (err error) {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("status", 1).Error
}

// ActivateUser 激活用户
func (s *UserAdminService) ActivateUser(id uint) (err error) {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("status", 0).Error
}

// ResetPassword 重置密码
func (s *UserAdminService) ResetPassword(id uint, password string) (err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("password", string(hashedPassword)).Error
}

// AdjustBalance 余额调节（使用UserAssets表，支持现货和合约账户）
func (s *UserAdminService) AdjustBalance(userID uint, currencyID uint, walletType string, amount float64, reason string) error {
	logger.Infof("[AdjustBalance] 开始调整余额: userID=%d, currencyID=%d, walletType=%s, amount=%.8f", userID, currencyID, walletType, amount)

	walletTypeEnum := model.WalletTypeSpot
	if walletType == "contract" {
		walletTypeEnum = model.WalletTypeContract
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查找用户钱包（user_assets表）
		var assets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletTypeEnum).First(&assets).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Infof("[AdjustBalance] 钱包不存在，自动创建: userID=%d, walletType=%s", userID, walletType)
				now := time.Now().Unix()
				assets = model.UserAssets{
					UserID:           userID,
					WalletType:       walletTypeEnum,
					UsdtBalance:      0,
					UsdtLocked:       0,
					CurrencyBalances: make(model.CurrencyBalance),
					CurrencyLocked:   make(model.CurrencyBalance),
					CreateTime:       now,
					UpdateTime:       now,
				}
				if err := tx.Create(&assets).Error; err != nil {
					logger.Errorf("[AdjustBalance] 创建钱包失败: %v", err)
					return errors.New("创建用户钱包失败: " + err.Error())
				}
				logger.Infof("[AdjustBalance] 钱包创建成功: id=%d", assets.ID)
			} else {
				logger.Errorf("[AdjustBalance] 查询钱包出错: %v", err)
				return err
			}
		} else {
			logger.Infof("[AdjustBalance] 找到现有钱包: id=%d, 当前余额=%.8f", assets.ID, assets.UsdtBalance)
		}

		// 2. 检查余额是否足够（扣款时检查）
		if amount < 0 && assets.UsdtBalance+amount < 0 {
			return errors.New("USDT余额不足")
		}

		// 3. 更新余额
		newBalance := assets.UsdtBalance + amount
		if err := tx.Model(&model.UserAssets{}).
			Where("id = ?", assets.ID).
			Updates(map[string]interface{}{
				"usdt_balance": newBalance,
				"update_time":  time.Now().Unix(),
				"version":      gorm.Expr("version + 1"),
			}).Error; err != nil {
			return err
		}

		// 4. 记录日志
		logType := 100
		if walletType == "spot" {
			if amount > 0 {
				logType = 101
			} else {
				logType = 102
			}
		} else {
			if amount > 0 {
				logType = 103
			} else {
				logType = 104
			}
		}

		accountLog := model.AccountLog{
			UserID:      userID,
			Value:       amount,
			CurrencyID:  currencyID,
			Type:        logType,
			Info:        reason + " [" + walletType + "]",
			CreatedTime: time.Now().Unix(),
		}
		if err := tx.Create(&accountLog).Error; err != nil {
			return err
		}

		logger.Infof("[AdjustBalance] 用户 %d %s账户余额调整成功: 变动=%.8f, 新余额=%.8f, 原因=%s",
			userID, walletType, amount, newBalance, reason)

		return nil
	})
}

// getAccountLogType 获取账户日志类型
func getAccountLogType(balanceType string, isAdd bool) int {
	typeMap := map[string]map[bool]int{
		"legal_balance":  {true: 1, false: 2},
		"change_balance": {true: 3, false: 4},
		"lever_balance":  {true: 5, false: 6},
		"micro_balance":  {true: 7, false: 8},
	}
	if types, ok := typeMap[balanceType]; ok {
		return types[isAdd]
	}
	return 0
}

// BatchSetRisk 批量设置风控
func (s *UserAdminService) BatchSetRisk(userIDs []uint, riskLevel int8) error {
	if len(userIDs) == 0 {
		return errors.New("用户ID列表不能为空")
	}
	return database.DB.Model(&model.User{}).Where("id IN ?", userIDs).Update("risk", riskLevel).Error
}

// UserWalletItem 用户钱包项（带币种信息）
// 注意：已迁移到UserAssets表
type UserWalletItem struct {
	UserID       uint    `json:"user_id"`
	CurrencyID   uint    `json:"currency_id"`
	Balance      float64 `json:"balance"`
	Locked       float64 `json:"locked"`
	CurrencyName string  `json:"currency_name"`
}

// UserAssetsItem 用户资产项(基于user_assets表)
type UserAssetsItem struct {
	CurrencyName string  `json:"currency_name"`
	CurrencyID   uint    `json:"currency_id"`
	Balance      float64 `json:"balance"`
	Locked       float64 `json:"locked"`
	Total        float64 `json:"total"`
	UsdtValue    float64 `json:"usdt_value"` // 折合USDT价值
}

// GetUserWallets 获取用户钱包列表(从user_assets表)
func (s *UserAdminService) GetUserWallets(userID uint) ([]UserAssetsItem, error) {
	// 1. 获取用户资产
	userAssetsService := service.GetUserAssetsService()
	assets, err := userAssetsService.GetUserAssets(userID)
	if err != nil {
		return nil, err
	}

	// 2. 获取所有币种信息
	var currencies []model.Currency
	if err := database.DB.Where("is_display = 1").Find(&currencies).Error; err != nil {
		return nil, err
	}

	// 创建币种映射
	currencyMap := make(map[string]model.Currency)
	for _, c := range currencies {
		currencyMap[c.Name] = c
	}

	// 3. 构建返回数据
	var result []UserAssetsItem

	// 首先添加USDT
	if assets.UsdtBalance > 0 || assets.UsdtLocked > 0 {
		usdtCurrency, exists := currencyMap["USDT"]
		usdtID := uint(0)
		if exists {
			usdtID = usdtCurrency.ID
		}
		result = append(result, UserAssetsItem{
			CurrencyName: "USDT",
			CurrencyID:   usdtID,
			Balance:      assets.UsdtBalance,
			Locked:       assets.UsdtLocked,
			Total:        assets.UsdtBalance + assets.UsdtLocked,
			UsdtValue:    assets.UsdtBalance + assets.UsdtLocked,
		})
	}

	// 4. 添加其他币种
	for currencyName, balance := range assets.CurrencyBalances {
		if balance == 0 && assets.CurrencyLocked[currencyName] == 0 {
			continue // 跳过余额为0的币种
		}

		currency, exists := currencyMap[currencyName]
		currencyID := uint(0)
		if exists {
			currencyID = currency.ID
		}

		locked := assets.CurrencyLocked[currencyName]
		total := balance + locked

		// 计算USDT价值(简化版本,实际应该查询实时价格)
		usdtValue := total // TODO: 乘以当前币种价格

		result = append(result, UserAssetsItem{
			CurrencyName: currencyName,
			CurrencyID:   currencyID,
			Balance:      balance,
			Locked:       locked,
			Total:        total,
			UsdtValue:    usdtValue,
		})
	}

	// 5. 也添加余额为0的币种(便于管理后台显示完整列表)
	for _, currency := range currencies {
		// 检查是否已存在
		exists := false
		for _, item := range result {
			if item.CurrencyName == currency.Name {
				exists = true
				break
			}
		}

		if !exists {
			result = append(result, UserAssetsItem{
				CurrencyName: currency.Name,
				CurrencyID:   currency.ID,
				Balance:      0,
				Locked:       0,
				Total:        0,
				UsdtValue:    0,
			})
		}
	}

	return result, nil
}

// SearchUsers 搜索用户
// UserListItem 用户列表项（带实名、上级账号与USDT余额）
type UserListItem struct {
	model.User
	RealName        string  `gorm:"column:real_name" json:"real_name"`
	CardID          string  `gorm:"column:card_id" json:"card_id"`
	ParentAccount   string  `gorm:"column:parent_account" json:"parent_account"`
	UsdtBalance     float64 `gorm:"column:usdt_balance" json:"usdt_balance"`
	LockUsdtBalance float64 `gorm:"column:lock_usdt_balance" json:"lock_usdt_balance"`
}

// SearchUsers 搜索用户
func (s *UserAdminService) SearchUsers(keyword, realName string, status *int, isReal *int, risk *int8, page, pageSize int) (list []UserListItem, total int64, err error) {
	// 已迁移到user_assets表，使用JSON字段存储USDT余额
	db := database.DB.Table("users AS u").
		Select("u.*, ur.name AS real_name, ur.card_id, p.account_number AS parent_account, a.usdt_balance, a.usdt_locked AS lock_usdt_balance").
		Joins("LEFT JOIN user_real ur ON ur.user_id = u.id").
		Joins("LEFT JOIN users p ON u.parent_id = p.id").
		Joins("LEFT JOIN user_assets a ON a.user_id = u.id") // 使用user_assets表

	if keyword != "" {
		db = db.Where("u.account_number LIKE ? OR u.phone LIKE ? OR u.email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if realName != "" {
		db = db.Where("ur.name LIKE ?", "%"+realName+"%")
	}
	if status != nil {
		db = db.Where("u.status = ?", *status)
	}
	if isReal != nil {
		db = db.Where("u.is_realname = ?", *isReal)
	}
	if risk != nil {
		db = db.Where("u.risk = ?", *risk)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("u.id DESC").Find(&list).Error
	return
}

// SetUserRisk 设置单个用户风控
func (s *UserAdminService) SetUserRisk(userID uint, riskLevel int8) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("risk", riskLevel).Error
}

// LockUser 锁定用户
func (s *UserAdminService) LockUser(id uint) (err error) {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("is_lock", 1).Error
}

// UnlockUser 解锁用户
func (s *UserAdminService) UnlockUser(id uint) (err error) {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("is_lock", 0).Error
}
