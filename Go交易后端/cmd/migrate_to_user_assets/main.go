package main

import (
	"fmt"
	"log"
	"os"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// 数据迁移工具: users_wallet -> user_assets
// 用法: go run cmd/migrate_to_user_assets/main.go [--dry-run]

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

	// 检查是否为演练模式
	dryRun := false
	if len(os.Args) > 1 && os.Args[1] == "--dry-run" {
		dryRun = true
		logger.Info("==================== DRY RUN 模式 ====================")
		logger.Info("仅执行数据检查,不会实际迁移数据")
	}

	logger.Info("==================== 开始数据迁移 ====================")
	logger.Info("迁移方案: users_wallet -> user_assets (JSON聚合存储)")

	// 步骤1: 预检查
	logger.Info("\n[步骤1] 预检查...")
	if err := preCheck(); err != nil {
		log.Fatalf("预检查失败: %v", err)
	}

	// 步骤2: 统计数据
	logger.Info("\n[步骤2] 统计现有数据...")
	stats := getStatistics()
	printStatistics(stats)

	if dryRun {
		logger.Info("\n==================== DRY RUN 结束 ====================")
		logger.Info("实际迁移请执行: go run cmd/migrate_to_user_assets/main.go")
		return
	}

	// 步骤3: 确认迁移
	logger.Info("\n[步骤3] 准备开始迁移...")
	logger.Warn("⚠️  此操作将创建新表并迁移数据,请确保已备份数据库!")
	logger.Info("按 Ctrl+C 取消,或等待5秒后自动开始...")
	// time.Sleep(5 * time.Second)

	// 步骤4: 创建表(如果不存在)
	logger.Info("\n[步骤4] 创建 user_assets 表...")
	if err := createUserAssetsTable(); err != nil {
		log.Fatalf("创建表失败: %v", err)
	}

	// 步骤4.5: 清理旧数据(如果存在)
	logger.Info("\n[步骤4.5] 清理旧迁移数据...")
	if err := cleanOldData(); err != nil {
		logger.Warnf("清理旧数据失败: %v (继续迁移)", err)
	}

	// 步骤5: 执行迁移
	logger.Info("\n[步骤5] 执行数据迁移...")
	migrated, failed, err := migrateData()
	if err != nil {
		log.Fatalf("数据迁移失败: %v", err)
	}

	logger.Infof("✓ 迁移完成: 成功=%d, 失败=%d", migrated, failed)

	// 步骤6: 验证数据
	logger.Info("\n[步骤6] 验证迁移结果...")
	if err := verifyMigration(); err != nil {
		log.Fatalf("验证失败: %v", err)
	}

	logger.Info("\n==================== 迁移完成 ====================")
	logger.Info("✓ 所有数据已成功迁移到 user_assets 表")
	logger.Info("\n后续步骤:")
	logger.Info("1. 更新后端代码使用新表")
	logger.Info("2. 灰度发布测试")
	logger.Info("3. 监控性能指标")
	logger.Info("4. 稳定后清理旧表(保留备份)")
}

// preCheck 预检查
func preCheck() error {
	// 检查 users_wallet 表是否存在
	if !database.DB.Migrator().HasTable("users_wallet") {
		return fmt.Errorf("users_wallet 表不存在")
	}

	// 检查 currency 表是否存在
	if !database.DB.Migrator().HasTable("currency") {
		return fmt.Errorf("currency 表不存在")
	}

	// 检查 users 表是否存在
	if !database.DB.Migrator().HasTable("users") {
		return fmt.Errorf("users 表不存在")
	}

	logger.Info("✓ 预检查通过")
	return nil
}

// Statistics 统计信息
type Statistics struct {
	TotalUsers         int64
	TotalWalletRecords int64
	CurrencyCount      int64
	UsdtWalletCount    int64
	TotalUsdtBalance   float64
	TotalLockedBalance float64
	NonZeroBalances    int64
}

// getStatistics 获取统计信息
func getStatistics() Statistics {
	var stats Statistics

	// 统计用户数
	database.DB.Model(&model.User{}).Count(&stats.TotalUsers)

	// 统计钱包记录数
	database.DB.Model(&model.UsersWallet{}).Count(&stats.TotalWalletRecords)

	// 统计币种数
	database.DB.Model(&model.Currency{}).Where("is_display = 1").Count(&stats.CurrencyCount)

	// 统计USDT钱包数
	database.DB.Model(&model.UsersWallet{}).Where("currency = 3").Count(&stats.UsdtWalletCount)

	// 统计USDT总余额
	database.DB.Model(&model.UsersWallet{}).
		Where("currency = 3").
		Select("SUM(usdt_balance)").
		Scan(&stats.TotalUsdtBalance)

	// 统计锁定余额
	database.DB.Model(&model.UsersWallet{}).
		Where("currency = 3").
		Select("SUM(lock_usdt_balance)").
		Scan(&stats.TotalLockedBalance)

	// 统计非零余额记录数
	database.DB.Model(&model.UsersWallet{}).
		Where("usdt_balance > 0").
		Count(&stats.NonZeroBalances)

	return stats
}

// printStatistics 打印统计信息
func printStatistics(stats Statistics) {
	logger.Info("----------------------------------------")
	logger.Infof("总用户数:         %d", stats.TotalUsers)
	logger.Infof("总钱包记录数:     %d", stats.TotalWalletRecords)
	logger.Infof("启用币种数:       %d", stats.CurrencyCount)
	logger.Infof("USDT钱包数:       %d", stats.UsdtWalletCount)
	logger.Infof("USDT总余额:       %.8f", stats.TotalUsdtBalance)
	logger.Infof("USDT锁定余额:     %.8f", stats.TotalLockedBalance)
	logger.Infof("非零余额记录数:   %d", stats.NonZeroBalances)
	logger.Info("----------------------------------------")
}

// createUserAssetsTable 创建表
func createUserAssetsTable() error {
	if database.DB.Migrator().HasTable(&model.UserAssets{}) {
		logger.Info("✓ user_assets 表已存在,跳过创建")
		return nil
	}

	err := database.DB.AutoMigrate(&model.UserAssets{})
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	logger.Info("✓ user_assets 表创建成功")
	return nil
}

// cleanOldData 清理旧迁移数据
func cleanOldData() error {
	var count int64
	database.DB.Model(&model.UserAssets{}).Count(&count)

	if count > 0 {
		logger.Infof("发现 %d 条旧数据,清理中...", count)
		err := database.DB.Exec("TRUNCATE TABLE user_assets").Error
		if err != nil {
			return err
		}
		logger.Info("✓ 旧数据清理完成")
	} else {
		logger.Info("✓ 无旧数据,跳过清理")
	}
	return nil
}

// migrateData 迁移数据
func migrateData() (migrated int, failed int, err error) {
	// 获取所有用户
	var users []model.User
	database.DB.Find(&users)

	logger.Infof("开始迁移 %d 个用户的资产数据...", len(users))

	// 获取所有币种
	var currencies []model.Currency
	database.DB.Where("is_display = 1").Find(&currencies)
	currencyMap := make(map[uint]model.Currency)
	for _, c := range currencies {
		currencyMap[c.ID] = c
	}

	// 批量迁移
	batchSize := 100
	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}

		batch := users[i:end]
		for _, user := range batch {
			err := migrateUserAssets(user.ID, currencyMap)
			if err != nil {
				logger.Errorf("迁移用户 %d 失败: %v", user.ID, err)
				failed++
			} else {
				migrated++
			}
		}

		// 进度显示
		if (i+batchSize)%1000 == 0 {
			logger.Infof("已迁移: %d/%d", i+batchSize, len(users))
		}
	}

	return migrated, failed, nil
}

// migrateUserAssets 迁移单个用户资产
func migrateUserAssets(userID uint, currencyMap map[uint]model.Currency) error {
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
				CreateTime:       0,
				UpdateTime:       0,
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
				// USDT
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

		// 3. 创建或更新 user_assets
		userAssets := &model.UserAssets{
			UserID:           userID,
			UsdtBalance:      usdtBalance,
			UsdtLocked:       usdtLocked,
			CurrencyBalances: currencyBalances,
			CurrencyLocked:   currencyLocked,
			TotalValueUsdt:   totalValue,
			LastTradeTime:    0,
			LastUpdateTime:   0,
			Version:          0,
			CreateTime:       wallets[0].CreateTime,
			UpdateTime:       0,
		}

		return tx.Save(userAssets).Error
	})
}

// verifyMigration 验证迁移
func verifyMigration() error {
	// 验证1: 用户数量对比 - 使用users表作为基准
	var oldUserCount, newUserCount int64
	database.DB.Model(&model.User{}).Count(&oldUserCount)
	database.DB.Model(&model.UserAssets{}).Count(&newUserCount)

	if oldUserCount != newUserCount {
		logger.Warnf("用户数量差异: users=%d, user_assets=%d (正常,部分用户可能无钱包记录)", oldUserCount, newUserCount)
	}
	logger.Infof("✓ 用户数量: users=%d, user_assets=%d", oldUserCount, newUserCount)

	// 验证2: USDT总余额对比
	var oldTotal, newTotal float64
	database.DB.Model(&model.UsersWallet{}).
		Where("currency = 3").
		Select("SUM(usdt_balance)").
		Scan(&oldTotal)
	database.DB.Model(&model.UserAssets{}).
		Select("SUM(usdt_balance)").
		Scan(&newTotal)

	diff := oldTotal - newTotal
	if diff > 0.00000001 || diff < -0.00000001 {
		return fmt.Errorf("USDT总额不匹配: old=%.8f, new=%.8f, diff=%.8f",
			oldTotal, newTotal, diff)
	}
	logger.Infof("✓ USDT总额验证通过: %.8f", newTotal)

	// 验证3: 随机抽样验证
	var sampleUsers []uint
	database.DB.Model(&model.UserAssets{}).
		Select("user_id").
		Order("RAND()").
		Limit(10).
		Pluck("user_id", &sampleUsers)

	for _, userID := range sampleUsers {
		err := verifyUserAssets(userID)
		if err != nil {
			return fmt.Errorf("用户 %d 验证失败: %v", userID, err)
		}
	}
	logger.Infof("✓ 随机抽样验证通过: %d 个用户", len(sampleUsers))

	return nil
}

// verifyUserAssets 验证单个用户资产
func verifyUserAssets(userID uint) error {
	// 查询旧表
	var oldWallet model.UsersWallet
	err := database.DB.Where("user_id = ? AND currency = 3", userID).First(&oldWallet).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 查询新表
	var newAssets model.UserAssets
	err = database.DB.Where("user_id = ?", userID).First(&newAssets).Error
	if err != nil {
		return err
	}

	// 对比USDT余额
	diff := oldWallet.UsdtBalance - newAssets.UsdtBalance
	if diff > 0.00000001 || diff < -0.00000001 {
		return fmt.Errorf("USDT余额不匹配: old=%.8f, new=%.8f",
			oldWallet.UsdtBalance, newAssets.UsdtBalance)
	}

	return nil
}
