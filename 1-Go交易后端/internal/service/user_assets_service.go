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

// GetUserAssets 获取用户资产
func (s *UserAssetsService) GetUserAssets(userID uint) (*model.UserAssets, error) {
	var assets model.UserAssets
	err := database.DB.Where("user_id = ?", userID).First(&assets).Error

	if err == gorm.ErrRecordNotFound {
		// 用户资产不存在,创建默认记录
		return s.CreateUserAssets(userID)
	}

	if err != nil {
		logger.Errorf("[UserAssets] 查询用户资产失败: userID=%d, err=%v", userID, err)
		return nil, err
	}

	return &assets, nil
}

// CreateUserAssets 创建用户资产记录
func (s *UserAssetsService) CreateUserAssets(userID uint) (*model.UserAssets, error) {
	now := time.Now().Unix()
	assets := &model.UserAssets{
		UserID:           userID,
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
		logger.Errorf("[UserAssets] 创建用户资产失败: userID=%d, err=%v", userID, err)
		return nil, err
	}

	logger.Infof("[UserAssets] 创建用户资产成功: userID=%d", userID)
	return assets, nil
}

// GetOrCreateUserAssets 获取或创建用户资产
func (s *UserAssetsService) GetOrCreateUserAssets(userID uint) (*model.UserAssets, error) {
	assets, err := s.GetUserAssets(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return assets, nil
}

// ============================================================
// USDT余额操作
// ============================================================

// UpdateUsdtBalance 更新USDT余额
func (s *UserAssetsService) UpdateUsdtBalance(userID uint, amount float64, isLocked bool) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		// 使用乐观锁
		err := tx.Where("user_id = ?", userID).First(&assets).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 首次操作,创建记录
				assets := &model.UserAssets{
					UserID:           userID,
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
			Where("user_id = ? AND version = ?", userID, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("concurrent update detected, please retry")
		}

		logger.Infof("[UserAssets] USDT余额更新: userID=%d, amount=%.8f, isLocked=%v",
			userID, amount, isLocked)
		return nil
	})
}

// LockUsdt 锁定USDT
func (s *UserAssetsService) LockUsdt(userID uint, amount float64) error {
	if amount <= 0 {
		return errors.New("lock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ?", userID).First(&assets).Error
		if err != nil {
			return err
		}

		if assets.UsdtBalance < amount {
			return fmt.Errorf("insufficient USDT balance: have=%.8f, need=%.8f",
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
			Where("user_id = ? AND version = ? AND usdt_balance >= ?", userID, assets.Version, amount).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("lock USDT failed: concurrent update or insufficient balance")
		}

		logger.Infof("[UserAssets] 锁定USDT: userID=%d, amount=%.8f", userID, amount)
		return nil
	})
}

// UnlockUsdt 解锁USDT
func (s *UserAssetsService) UnlockUsdt(userID uint, amount float64) error {
	if amount <= 0 {
		return errors.New("unlock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ?", userID).First(&assets).Error
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
			Where("user_id = ? AND version = ? AND usdt_locked >= ?", userID, assets.Version, amount).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("unlock USDT failed: concurrent update or insufficient locked balance")
		}

		logger.Infof("[UserAssets] 解锁USDT: userID=%d, amount=%.8f", userID, amount)
		return nil
	})
}

// ============================================================
// 币种余额操作
// ============================================================

// UpdateCurrencyBalance 更新币种余额
func (s *UserAssetsService) UpdateCurrencyBalance(userID uint, currencyName string, amount float64, isLocked bool) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ?", userID).First(&assets).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 首次操作,创建记录
				assets := &model.UserAssets{
					UserID:           userID,
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
			"update_time":      time.Now().Unix(),
			"last_update_time": time.Now().Unix(),
			"version":          gorm.Expr("version + 1"),
		}

		if isLocked {
			updates["currency_locked"] = assets.CurrencyLocked
		} else {
			updates["currency_balances"] = assets.CurrencyBalances
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND version = ?", userID, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("concurrent update detected, please retry")
		}

		logger.Infof("[UserAssets] 币种余额更新: userID=%d, currency=%s, amount=%.8f, isLocked=%v",
			userID, currencyName, amount, isLocked)
		return nil
	})
}

// LockCurrency 锁定币种余额
func (s *UserAssetsService) LockCurrency(userID uint, currencyName string, amount float64) error {
	if amount <= 0 {
		return errors.New("lock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ?", userID).First(&assets).Error
		if err != nil {
			return err
		}

		// 检查余额
		currentBalance := assets.GetCurrencyBalance(currencyName)
		if currentBalance < amount {
			return fmt.Errorf("insufficient %s balance: have=%.8f, need=%.8f",
				currencyName, currentBalance, amount)
		}

		// 执行锁定
		err = assets.LockCurrency(currencyName, amount)
		if err != nil {
			return err
		}

		// 更新到数据库
		updates := map[string]interface{}{
			"currency_balances": assets.CurrencyBalances,
			"currency_locked":   assets.CurrencyLocked,
			"update_time":       time.Now().Unix(),
			"last_update_time":  time.Now().Unix(),
			"version":           gorm.Expr("version + 1"),
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND version = ?", userID, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("lock currency failed: concurrent update")
		}

		logger.Infof("[UserAssets] 锁定币种: userID=%d, currency=%s, amount=%.8f",
			userID, currencyName, amount)
		return nil
	})
}

// UnlockCurrency 解锁币种余额
func (s *UserAssetsService) UnlockCurrency(userID uint, currencyName string, amount float64) error {
	if amount <= 0 {
		return errors.New("unlock amount must be positive")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var assets model.UserAssets
		err := tx.Where("user_id = ?", userID).First(&assets).Error
		if err != nil {
			return err
		}

		// 执行解锁
		err = assets.UnlockCurrency(currencyName, amount)
		if err != nil {
			return err
		}

		// 更新到数据库
		updates := map[string]interface{}{
			"currency_balances": assets.CurrencyBalances,
			"currency_locked":   assets.CurrencyLocked,
			"update_time":       time.Now().Unix(),
			"last_update_time":  time.Now().Unix(),
			"version":           gorm.Expr("version + 1"),
		}

		result := tx.Model(&model.UserAssets{}).
			Where("user_id = ? AND version = ?", userID, assets.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("unlock currency failed: concurrent update")
		}

		logger.Infof("[UserAssets] 解锁币种: userID=%d, currency=%s, amount=%.8f",
			userID, currencyName, amount)
		return nil
	})
}

// ============================================================
// 资产概览查询
// ============================================================

// GetAssetOverviewFromUserAssets 从 user_assets表获取资产概览
func (s *UserAssetsService) GetAssetOverviewFromUserAssets(userID uint) (*AssetOverviewResponse, error) {
	// 1. 查询用户资产
	assets, err := s.GetOrCreateUserAssets(userID)
	if err != nil {
		return nil, err
	}

	// 2. 获取所有启用的币种
	var currencies []model.Currency
	database.DB.Where("is_display = 1").Order("sort DESC").Find(&currencies)

	// 3. 构建资产列表
	resp := &AssetOverviewResponse{
		TotalBalance:  0, // 从0开始累加，避免重复计算
		TotalUsdValue: 0,
		Assets:        make([]AssetItemInfo, 0),
	}

	// 遍历币种,计算持仓价值
	for _, currency := range currencies {
		var balance float64

		// 特殊处理USDT: 从usdt_balance字段读取,而不是从JSON
		if currency.Name == "USDT" {
			balance = assets.UsdtBalance
		} else {
			// 其他币种从JSON字段读取
			balance = assets.GetCurrencyBalance(currency.Name)
		}

		// 计算USDT价值
		usdtValue := balance * currency.Rate
		usdValue := usdtValue * 6.5

		// 累加总资产(USDT不需要再乘以rate)
		if currency.Name == "USDT" {
			resp.TotalBalance += balance
			resp.TotalUsdValue += balance * 6.5
		} else {
			resp.TotalBalance += usdtValue
			resp.TotalUsdValue += usdValue
		}

		// 添加到资产列表(包括余额为0的币种)
		assetItem := AssetItemInfo{
			CurrencyID:   currency.ID,
			CurrencyName: currency.Name,
			Symbol:       currency.Name,
			Logo:         currency.Logo,
			Balance:      balance,
			UsdtValue:    usdtValue,
			UsdValue:     usdValue,
			Price:        currency.Rate,
			Sort:         currency.Sort,
		}
		resp.Assets = append(resp.Assets, assetItem)
	}

	// 4. 获取今日盈亏(复用原有逻辑)
	walletService := GetWalletService()
	profitLoss, err := walletService.GetTodayProfitLoss(userID)
	if err == nil {
		resp.TodayProfit = profitLoss.NetProfitLoss

		// 计算盈亏率
		if resp.TotalBalance > 0 && profitLoss.NetProfitLoss != 0 {
			rate := (profitLoss.NetProfitLoss / resp.TotalBalance) * 100
			if rate >= 0 {
				resp.TodayProfitRate = fmt.Sprintf("+%.2f%%", rate)
			} else {
				resp.TodayProfitRate = fmt.Sprintf("%.2f%%", rate)
			}
		} else {
			resp.TodayProfitRate = "0.00%"
		}
	} else {
		resp.TodayProfit = 0
		resp.TodayProfitRate = "0.00%"
	}

	return resp, nil
}

// ============================================================
// 批量操作
// ============================================================

// BatchUpdateBalances 批量更新余额(事务中)
func (s *UserAssetsService) BatchUpdateBalances(tx *gorm.DB, updates []BalanceUpdate) error {
	for _, update := range updates {
		var assets model.UserAssets
		err := tx.Where("user_id = ?", update.UserID).First(&assets).Error
		if err != nil {
			return err
		}

		// 根据类型更新
		switch update.BalanceType {
		case "usdt":
			if update.IsLocked {
				assets.UsdtLocked += update.Amount
			} else {
				assets.UsdtBalance += update.Amount
			}
		default:
			// 币种余额
			if update.IsLocked {
				locked := assets.GetCurrencyLocked(update.BalanceType)
				assets.SetCurrencyLocked(update.BalanceType, locked+update.Amount)
			} else {
				balance := assets.GetCurrencyBalance(update.BalanceType)
				assets.SetCurrencyBalance(update.BalanceType, balance+update.Amount)
			}
		}

		// 保存
		err = tx.Save(&assets).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// BalanceUpdate 余额更新结构
type BalanceUpdate struct {
	UserID      uint
	BalanceType string // "usdt" or currency name
	Amount      float64
	IsLocked    bool
}

// ============================================================
// 统计查询
// ============================================================

// GetTopUsers 获取资产排行榜
func (s *UserAssetsService) GetTopUsers(limit int) ([]model.UserAssets, error) {
	var assets []model.UserAssets
	err := database.DB.Order("total_value_usdt DESC").Limit(limit).Find(&assets).Error
	return assets, err
}

// GetActiveUsers 获取活跃用户(最近交易)
func (s *UserAssetsService) GetActiveUsers(limit int) ([]model.UserAssets, error) {
	var assets []model.UserAssets
	err := database.DB.Where("last_trade_time > ?", time.Now().Unix()-86400*7).
		Order("last_trade_time DESC").
		Limit(limit).
		Find(&assets).Error
	return assets, err
}

// GetTotalAssets 获取平台总资产
func (s *UserAssetsService) GetTotalAssets() (map[string]interface{}, error) {
	var result struct {
		TotalUsers      int64
		TotalUsdt       float64
		TotalValue      float64
		AvgAssetPerUser float64
	}

	database.DB.Model(&model.UserAssets{}).Count(&result.TotalUsers)
	database.DB.Model(&model.UserAssets{}).Select("SUM(usdt_balance)").Scan(&result.TotalUsdt)
	database.DB.Model(&model.UserAssets{}).Select("SUM(total_value_usdt)").Scan(&result.TotalValue)

	if result.TotalUsers > 0 {
		result.AvgAssetPerUser = result.TotalValue / float64(result.TotalUsers)
	}

	return map[string]interface{}{
		"total_users":        result.TotalUsers,
		"total_usdt":         result.TotalUsdt,
		"total_value":        result.TotalValue,
		"avg_asset_per_user": result.AvgAssetPerUser,
	}, nil
}
