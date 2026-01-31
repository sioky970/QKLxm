package service

import (
	"errors"
	"fmt"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// UserAssetsService 用户资产服务(JSON聚合存储)
// 支持现货钱包(spot)和合约钱包(contract)分离
type UserAssetsService struct{}

var userAssetsServiceInstance *UserAssetsService

// GetUserAssetsService 获取用户资产服务实例
func GetUserAssetsService() *UserAssetsService {
	if userAssetsServiceInstance == nil {
		userAssetsServiceInstance = &UserAssetsService{}
	}
	return userAssetsServiceInstance
}

// ============================================================
// 基础查询操作
// ============================================================

// GetUserAssets 获取用户资产(默认现货钱包)
func (s *UserAssetsService) GetUserAssets(userID uint) (*model.UserAssets, error) {
	return s.GetUserAssetsByType(userID, model.WalletTypeSpot)
}

// GetUserAssetsByType 获取指定类型的用户资产
func (s *UserAssetsService) GetUserAssetsByType(userID uint, walletType model.WalletType) (*model.UserAssets, error) {
	var assets model.UserAssets
	err := database.DB.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error

	if err == gorm.ErrRecordNotFound {
		// 用户资产不存在,创建默认记录
		logger.Infof("[UserAssets] 钱包不存在,开始创建: userID=%d, walletType=%s", userID, walletType)
		newAssets, createErr := s.CreateUserAssetsByType(userID, walletType)
		if createErr != nil {
			logger.Errorf("[UserAssets] 创建钱包失败: userID=%d, walletType=%s, err=%v", userID, walletType, createErr)
			return nil, createErr
		}
		logger.Infof("[UserAssets] 钱包创建成功: userID=%d, walletType=%s, usdtBalance=%.8f", 
			userID, walletType, newAssets.UsdtBalance)
		return newAssets, nil
	}

	if err != nil {
		logger.Errorf("[UserAssets] 查询用户资产失败: userID=%d, walletType=%s, err=%v", userID, walletType, err)
		return nil, err
	}

	logger.Infof("[UserAssets] 获取钱包成功: userID=%d, walletType=%s, usdtBalance=%.8f", 
		userID, walletType, assets.UsdtBalance)
	return &assets, nil
}

// CreateUserAssets 创建用户资产记录(默认现货钱包)
func (s *UserAssetsService) CreateUserAssets(userID uint) (*model.UserAssets, error) {
	return s.CreateUserAssetsByType(userID, model.WalletTypeSpot)
}

// CreateUserAssetsByType 创建指定类型的用户资产记录
func (s *UserAssetsService) CreateUserAssetsByType(userID uint, walletType model.WalletType) (*model.UserAssets, error) {
	now := time.Now().Unix()
	assets := &model.UserAssets{
		UserID:           userID,
		WalletType:       walletType,
		UsdtBalance:      0,
		UsdtLocked:       0,
		CurrencyBalances: make(model.CurrencyBalance),
		CurrencyLocked:   make(model.CurrencyBalance),
		TotalValueUsdt:   0,
		LastTradeTime:    0,
		LastUpdateTime:   now,
		Version:          0,
		CreateTime:       now,
		UpdateTime:       now,
	}

	err := database.DB.Create(assets).Error
	if err != nil {
		logger.Errorf("[UserAssets] 创建用户资产失败: userID=%d, walletType=%s, err=%v", userID, walletType, err)
		return nil, err
	}

	logger.Infof("[UserAssets] 创建用户资产成功: userID=%d, walletType=%s", userID, walletType)
	return assets, nil
}

// GetOrCreateUserAssets 获取或创建用户资产(默认现货钱包)
func (s *UserAssetsService) GetOrCreateUserAssets(userID uint) (*model.UserAssets, error) {
	return s.GetOrCreateUserAssetsByType(userID, model.WalletTypeSpot)
}

// GetOrCreateUserAssetsByType 获取或创建指定类型的用户资产
func (s *UserAssetsService) GetOrCreateUserAssetsByType(userID uint, walletType model.WalletType) (*model.UserAssets, error) {
	assets, err := s.GetUserAssetsByType(userID, walletType)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return assets, nil
}

// ============================================================
// USDT余额操作 - 支持钱包类型
// ============================================================

// UpdateUsdtBalance 更新USDT余额(默认现货钱包)
func (s *UserAssetsService) UpdateUsdtBalance(userID uint, amount float64, isLocked bool) error {
	return s.UpdateUsdtBalanceByType(userID, amount, isLocked, model.WalletTypeSpot)
}

// UpdateUsdtBalanceByType 更新指定钱包类型的USDT余额
func (s *UserAssetsService) UpdateUsdtBalanceByType(userID uint, amount float64, isLocked bool, walletType model.WalletType) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		// 使用乐观锁
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 首次操作,创建记录
				assets := &model.UserAssets{
					UserID:           userID,
					WalletType:       walletType,
					UsdtBalance:      0,
					UsdtLocked:       0,
					CurrencyBalances: make(model.CurrencyBalance),
					CurrencyLocked:   make(model.CurrencyBalance),
					CreateTime:       time.Now().Unix(),
					UpdateTime:       time.Now().Unix(),
				}
				if isLocked {
					assets.UsdtLocked = amount
				} else {
					assets.UsdtBalance = amount
				}
				return tx.Create(assets).Error
			}
			return err
		}

		// 更新余额
		updates := map[string]interface{}{
			"update_time":      time.Now().Unix(),
			"last_update_time": time.Now().Unix(),
			"version":          gorm.Expr("version + 1"),
		}

		if isLocked {
			newLocked := assets.UsdtLocked + amount
			if newLocked < 0 {
				return errors.New("insufficient locked USDT balance")
			}
			updates["usdt_locked"] = newLocked
		} else {
			newBalance := assets.UsdtBalance + amount
			if newBalance < 0 {
				return errors.New("insufficient USDT balance")
			}
			updates["usdt_balance"] = newBalance
		}

		// 使用版本号防止并发问题
		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND wallet_type = ? AND version = ?", userID, walletType, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("concurrent update detected, please retry")
		}

		logger.Infof("[UserAssets] USDT余额更新: userID=%d, walletType=%s, amount=%.8f, isLocked=%v",
			userID, walletType, amount, isLocked)
		return nil
	})
}

// LockUsdt 锁定USDT(默认现货钱包)
func (s *UserAssetsService) LockUsdt(userID uint, amount float64) error {
	return s.LockUsdtByType(userID, amount, model.WalletTypeSpot)
}

// LockUsdtByType 锁定指定钱包类型的USDT
func (s *UserAssetsService) LockUsdtByType(userID uint, amount float64, walletType model.WalletType) error {
	if amount <= 0 {
		return errors.New("lock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 用户资产记录不存在，说明余额为0
				return fmt.Errorf("USDT余额不足: 当前余额 0, 需要 %.2f", amount)
			}
			return err
		}

		if assets.UsdtBalance < amount {
			return fmt.Errorf("USDT余额不足: 当前余额 %.2f, 需要 %.2f",
				assets.UsdtBalance, amount)
		}

		updates := map[string]interface{}{
			"usdt_balance":     gorm.Expr("usdt_balance - ?", amount),
			"usdt_locked":      gorm.Expr("usdt_locked + ?", amount),
			"update_time":      time.Now().Unix(),
			"last_update_time": time.Now().Unix(),
			"version":          gorm.Expr("version + 1"),
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND wallet_type = ? AND version = ? AND usdt_balance >= ?", userID, walletType, assets.Version, amount).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("锁定失败: 并发更新或余额不足，请重试")
		}

		logger.Infof("[UserAssets] 锁定USDT: userID=%d, walletType=%s, amount=%.8f", userID, walletType, amount)
		return nil
	})
}

// UnlockUsdt 解锁USDT(默认现货钱包)
func (s *UserAssetsService) UnlockUsdt(userID uint, amount float64) error {
	return s.UnlockUsdtByType(userID, amount, model.WalletTypeSpot)
}

// UnlockUsdtByType 解锁指定钱包类型的USDT
func (s *UserAssetsService) UnlockUsdtByType(userID uint, amount float64, walletType model.WalletType) error {
	if amount <= 0 {
		return errors.New("unlock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error
		if err != nil {
			return err
		}

		if assets.UsdtLocked < amount {
			return fmt.Errorf("insufficient locked USDT: have=%.8f, need=%.8f",
				assets.UsdtLocked, amount)
		}

		updates := map[string]interface{}{
			"usdt_balance":     gorm.Expr("usdt_balance + ?", amount),
			"usdt_locked":      gorm.Expr("usdt_locked - ?", amount),
			"update_time":      time.Now().Unix(),
			"last_update_time": time.Now().Unix(),
			"version":          gorm.Expr("version + 1"),
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND wallet_type = ? AND version = ? AND usdt_locked >= ?", userID, walletType, assets.Version, amount).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("unlock USDT failed: concurrent update or insufficient locked balance")
		}

		logger.Infof("[UserAssets] 解锁USDT: userID=%d, walletType=%s, amount=%.8f", userID, walletType, amount)
		return nil
	})
}

// ============================================================
// 币种余额操作 - 支持钱包类型
// ============================================================

// UpdateCurrencyBalance 更新币种余额(默认现货钱包)
func (s *UserAssetsService) UpdateCurrencyBalance(userID uint, currencyName string, amount float64, isLocked bool) error {
	return s.UpdateCurrencyBalanceByType(userID, currencyName, amount, isLocked, model.WalletTypeSpot)
}

// UpdateCurrencyBalanceByType 更新指定钱包类型的币种余额
func (s *UserAssetsService) UpdateCurrencyBalanceByType(userID uint, currencyName string, amount float64, isLocked bool, walletType model.WalletType) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 首次操作,创建记录
				assets := &model.UserAssets{
					UserID:           userID,
					WalletType:       walletType,
					UsdtBalance:      0,
					UsdtLocked:       0,
					CurrencyBalances: make(model.CurrencyBalance),
					CurrencyLocked:   make(model.CurrencyBalance),
					CreateTime:       time.Now().Unix(),
					UpdateTime:       time.Now().Unix(),
				}
				if isLocked {
					assets.SetCurrencyLocked(currencyName, amount)
				} else {
					assets.SetCurrencyBalance(currencyName, amount)
				}
				return tx.Create(assets).Error
			}
			return err
		}

		// 更新币种余额
		if isLocked {
			currentLocked := assets.GetCurrencyLocked(currencyName)
			newLocked := currentLocked + amount
			if newLocked < 0 {
				return fmt.Errorf("insufficient locked %s balance", currencyName)
			}
			assets.SetCurrencyLocked(currencyName, newLocked)
		} else {
			currentBalance := assets.GetCurrencyBalance(currencyName)
			newBalance := currentBalance + amount
			if newBalance < 0 {
				return fmt.Errorf("insufficient %s balance", currencyName)
			}
			assets.SetCurrencyBalance(currencyName, newBalance)
		}

		// 更新到数据库
		updates := map[string]interface{}{
			"update_time":       time.Now().Unix(),
			"last_update_time":  time.Now().Unix(),
			"version":           gorm.Expr("version + 1"),
			"currency_balances": assets.CurrencyBalances,
			"currency_locked":   assets.CurrencyLocked,
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND wallet_type = ? AND version = ?", userID, walletType, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("concurrent update detected, please retry")
		}

		logger.Infof("[UserAssets] 币种余额更新: userID=%d, walletType=%s, currency=%s, amount=%.8f, isLocked=%v",
			userID, walletType, currencyName, amount, isLocked)
		return nil
	})
}

// LockCurrency 锁定币种(默认现货钱包)
func (s *UserAssetsService) LockCurrency(userID uint, currencyName string, amount float64) error {
	return s.LockCurrencyByType(userID, currencyName, amount, model.WalletTypeSpot)
}

// LockCurrencyByType 锁定指定钱包类型的币种
func (s *UserAssetsService) LockCurrencyByType(userID uint, currencyName string, amount float64, walletType model.WalletType) error {
	if amount <= 0 {
		return errors.New("lock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error
		if err != nil {
			return err
		}

		currentBalance := assets.GetCurrencyBalance(currencyName)
		if currentBalance < amount {
			return fmt.Errorf("%s余额不足: 当前余额 %.8f, 需要 %.8f", currencyName, currentBalance, amount)
		}

		newBalance := currentBalance - amount
		currentLocked := assets.GetCurrencyLocked(currencyName)
		newLocked := currentLocked + amount

		assets.SetCurrencyBalance(currencyName, newBalance)
		assets.SetCurrencyLocked(currencyName, newLocked)

		updates := map[string]interface{}{
			"currency_balances": assets.CurrencyBalances,
			"currency_locked":   assets.CurrencyLocked,
			"update_time":       time.Now().Unix(),
			"last_update_time":  time.Now().Unix(),
			"version":           gorm.Expr("version + 1"),
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND wallet_type = ? AND version = ?", userID, walletType, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("锁定失败: 并发更新，请重试")
		}

		logger.Infof("[UserAssets] 锁定币种: userID=%d, walletType=%s, currency=%s, amount=%.8f",
			userID, walletType, currencyName, amount)
		return nil
	})
}

// UnlockCurrency 解锁币种(默认现货钱包)
func (s *UserAssetsService) UnlockCurrency(userID uint, currencyName string, amount float64) error {
	return s.UnlockCurrencyByType(userID, currencyName, amount, model.WalletTypeSpot)
}

// UnlockCurrencyByType 解锁指定钱包类型的币种
func (s *UserAssetsService) UnlockCurrencyByType(userID uint, currencyName string, amount float64, walletType model.WalletType) error {
	if amount <= 0 {
		return errors.New("unlock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ? AND wallet_type = ?", userID, walletType).First(&assets).Error
		if err != nil {
			return err
		}

		currentLocked := assets.GetCurrencyLocked(currencyName)
		if currentLocked < amount {
			return fmt.Errorf("锁定的%s余额不足: 当前锁定 %.8f, 需要 %.8f", currencyName, currentLocked, amount)
		}

		newLocked := currentLocked - amount
		currentBalance := assets.GetCurrencyBalance(currencyName)
		newBalance := currentBalance + amount

		assets.SetCurrencyLocked(currencyName, newLocked)
		assets.SetCurrencyBalance(currencyName, newBalance)

		updates := map[string]interface{}{
			"currency_balances": assets.CurrencyBalances,
			"currency_locked":   assets.CurrencyLocked,
			"update_time":       time.Now().Unix(),
			"last_update_time":  time.Now().Unix(),
			"version":           gorm.Expr("version + 1"),
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND wallet_type = ? AND version = ?", userID, walletType, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("解锁失败: 并发更新，请重试")
		}

		logger.Infof("[UserAssets] 解锁币种: userID=%d, walletType=%s, currency=%s, amount=%.8f",
			userID, walletType, currencyName, amount)
		return nil
	})
}

// ============================================================
// 便捷方法 - 合约钱包专用
// ============================================================

// GetContractAssets 获取合约钱包资产
func (s *UserAssetsService) GetContractAssets(userID uint) (*model.UserAssets, error) {
	return s.GetUserAssetsByType(userID, model.WalletTypeContract)
}

// UpdateContractUsdtBalance 更新合约钱包USDT余额
func (s *UserAssetsService) UpdateContractUsdtBalance(userID uint, amount float64, isLocked bool) error {
	return s.UpdateUsdtBalanceByType(userID, amount, isLocked, model.WalletTypeContract)
}

// LockContractUsdt 锁定合约钱包USDT
func (s *UserAssetsService) LockContractUsdt(userID uint, amount float64) error {
	return s.LockUsdtByType(userID, amount, model.WalletTypeContract)
}

// UnlockContractUsdt 解锁合约钱包USDT
func (s *UserAssetsService) UnlockContractUsdt(userID uint, amount float64) error {
	return s.UnlockUsdtByType(userID, amount, model.WalletTypeContract)
}

// UpdateContractCurrencyBalance 更新合约钱包币种余额
func (s *UserAssetsService) UpdateContractCurrencyBalance(userID uint, currencyName string, amount float64, isLocked bool) error {
	return s.UpdateCurrencyBalanceByType(userID, currencyName, amount, isLocked, model.WalletTypeContract)
}

// ============================================================
// 便捷方法 - 现货钱包专用
// ============================================================

// GetSpotAssets 获取现货钱包资产
func (s *UserAssetsService) GetSpotAssets(userID uint) (*model.UserAssets, error) {
	return s.GetUserAssetsByType(userID, model.WalletTypeSpot)
}

// UpdateSpotUsdtBalance 更新现货钱包USDT余额
func (s *UserAssetsService) UpdateSpotUsdtBalance(userID uint, amount float64, isLocked bool) error {
	return s.UpdateUsdtBalanceByType(userID, amount, isLocked, model.WalletTypeSpot)
}

// LockSpotUsdt 锁定现货钱包USDT
func (s *UserAssetsService) LockSpotUsdt(userID uint, amount float64) error {
	return s.LockUsdtByType(userID, amount, model.WalletTypeSpot)
}

// UnlockSpotUsdt 解锁现货钱包USDT
func (s *UserAssetsService) UnlockSpotUsdt(userID uint, amount float64) error {
	return s.UnlockUsdtByType(userID, amount, model.WalletTypeSpot)
}

// UpdateSpotCurrencyBalance 更新现货钱包币种余额
func (s *UserAssetsService) UpdateSpotCurrencyBalance(userID uint, currencyName string, amount float64, isLocked bool) error {
	return s.UpdateCurrencyBalanceByType(userID, currencyName, amount, isLocked, model.WalletTypeSpot)
}

// ============================================================
// 资产概览方法
// ============================================================

// GetAssetOverviewFromUserAssets 获取资产概览（用于首页和资产页）
// 返回 wallet_service.go 中定义的 AssetOverviewResponse 类型
func (s *UserAssetsService) GetAssetOverviewFromUserAssets(userID uint) (*AssetOverviewResponse, error) {
	// 获取现货钱包资产
	spotAssets, err := s.GetSpotAssets(userID)
	if err != nil {
		logger.Errorf("[UserAssets] 获取现货资产失败: userID=%d, err=%v", userID, err)
		return nil, err
	}

	// 获取合约钱包资产
	contractAssets, err := s.GetContractAssets(userID)
	if err != nil {
		logger.Errorf("[UserAssets] 获取合约资产失败: userID=%d, err=%v", userID, err)
		contractAssets = &model.UserAssets{
			UsdtBalance: 0,
			UsdtLocked:  0,
		}
	}

	// 计算总资产（现货 + 合约）
	totalUsdt := spotAssets.UsdtBalance + spotAssets.UsdtLocked +
		contractAssets.UsdtBalance + contractAssets.UsdtLocked

	// 获取所有启用的币种列表
	var currencies []model.Currency
	if err := database.DB.Where("is_display = ?", 1).Order("sort DESC").Find(&currencies).Error; err != nil {
		logger.Errorf("[UserAssets] 获取币种列表失败: err=%v", err)
		currencies = []model.Currency{}
	}

	// 构建币种资产列表
	var assetItems []AssetItemInfo

	// 添加USDT资产（始终显示）
	usdtPrice := 1.0
	usdtItem := AssetItemInfo{
		CurrencyID:   1,
		CurrencyName: "USDT",
		Symbol:       "USDT",
		Logo:         "",
		Balance:      spotAssets.UsdtBalance + spotAssets.UsdtLocked,
		UsdtValue:    spotAssets.UsdtBalance + spotAssets.UsdtLocked,
		UsdValue:     (spotAssets.UsdtBalance + spotAssets.UsdtLocked) * usdtPrice,
		Price:        usdtPrice,
		Sort:         100,
	}
	assetItems = append(assetItems, usdtItem)

	// 构建币种名称到价格的映射
	currencyPriceMap := make(map[string]float64)
	for _, c := range currencies {
		currencyPriceMap[c.Name] = c.Price
	}

	// 添加其他币种资产
	for _, currency := range currencies {
		if currency.Name == "USDT" {
			continue
		}

		// 获取用户在该币种的余额
		balance := spotAssets.GetCurrencyBalance(currency.Name)
		locked := spotAssets.GetCurrencyLocked(currency.Name)
		totalBalance := balance + locked

		if totalBalance <= 0 {
			continue
		}

		// 获取币种价格
		price := currencyPriceMap[currency.Name]
		if price <= 0 {
			price = 0
		}

		// 计算USDT价值
		usdtValue := totalBalance * price

		item := AssetItemInfo{
			CurrencyID:   currency.ID,
			CurrencyName: currency.Name,
			Symbol:       currency.Name,
			Logo:         currency.Logo,
			Balance:      totalBalance,
			UsdtValue:    usdtValue,
			UsdValue:     usdtValue,
			Price:        price,
			Sort:         currency.Sort,
		}
		assetItems = append(assetItems, item)
	}

	// 构建响应
	response := &AssetOverviewResponse{
		TotalBalance:    totalUsdt,
		TotalUsdValue:   totalUsdt,
		TodayProfit:     0,
		TodayProfitRate: "0%",
		Assets:          assetItems,
	}

	return response, nil
}
