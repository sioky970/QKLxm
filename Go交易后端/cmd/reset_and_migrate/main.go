package main

import (
	"fmt"
	"log"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

func main() {
	// 初始化配置
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化日志
	logger.Init(&config.GlobalConfig.Log)
	defer logger.Sync()

	// 连接数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("=== 开始重置并迁移数据 ===")

	// 步骤1: 清空user_assets表
	fmt.Println("\n[步骤1] 清空 user_assets 表...")
	if err := database.DB.Exec("TRUNCATE TABLE user_assets").Error; err != nil {
		log.Fatalf("清空表失败: %v", err)
	}
	fmt.Println("✓ user_assets 表已清空")

	// 步骤2: 获取所有用户
	fmt.Println("\n[步骤2] 获取用户列表...")
	var users []model.User
	database.DB.Find(&users)
	fmt.Printf("✓ 找到 %d 个用户\n", len(users))

	// 步骤3: 为每个用户迁移数据
	fmt.Println("\n[步骤3] 开始迁移数据...")

	// 构建币种映射
	var currencies []model.Currency
	database.DB.Find(&currencies)
	currencyMap := make(map[uint]model.Currency)
	for _, c := range currencies {
		currencyMap[c.ID] = c
	}

	successCount := 0
	failCount := 0

	for _, user := range users {
		if err := migrateUserData(user.ID, currencyMap); err != nil {
			fmt.Printf("✗ 用户 %d 迁移失败: %v\n", user.ID, err)
			failCount++
		} else {
			fmt.Printf("✓ 用户 %d 迁移成功\n", user.ID)
			successCount++
		}
	}

	fmt.Printf("\n=== 迁移完成 ===\n")
	fmt.Printf("成功: %d, 失败: %d\n", successCount, failCount)
}

func migrateUserData(userID uint, currencyMap map[uint]model.Currency) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查询用户所有钱包
		var wallets []model.UsersWallet
		tx.Where("user_id = ?", userID).Find(&wallets)

		if len(wallets) == 0 {
			// 用户没有钱包记录,创建空资产
			userAssets := &model.UserAssets{
				UserID:           userID,
				UsdtBalance:      0,
				UsdtLocked:       0,
				CurrencyBalances: make(model.CurrencyBalance),
				CurrencyLocked:   make(model.CurrencyBalance),
				TotalValueUsdt:   0,
			}
			return tx.Create(userAssets).Error
		}

		// 2. 聚合数据
		var usdtBalance, usdtLocked float64
		currencyBalances := make(model.CurrencyBalance)
		currencyLocked := make(model.CurrencyBalance)
		totalValue := float64(0)

		for _, wallet := range wallets {
			if wallet.CurrencyID == 3 {
				// USDT (currency_id=3)
				usdtBalance = wallet.UsdtBalance
				usdtLocked = wallet.LockUsdtBalance
				totalValue += usdtBalance
			} else {
				// 其他币种
				currency, exists := currencyMap[wallet.CurrencyID]
				if !exists {
					continue
				}

				if wallet.UsdtBalance > 0 {
					currencyBalances[currency.Name] = wallet.UsdtBalance
					totalValue += wallet.UsdtBalance * currency.Rate
				}

				if wallet.LockUsdtBalance > 0 {
					currencyLocked[currency.Name] = wallet.LockUsdtBalance
				}
			}
		}

		// 3. 创建 user_assets 记录
		userAssets := &model.UserAssets{
			UserID:           userID,
			UsdtBalance:      usdtBalance,
			UsdtLocked:       usdtLocked,
			CurrencyBalances: currencyBalances,
			CurrencyLocked:   currencyLocked,
			TotalValueUsdt:   totalValue,
		}

		return tx.Create(userAssets).Error
	})
}
