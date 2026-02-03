package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "dry run mode, no actual changes")
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	fmt.Println("=== 余额迁移工具启动 ===")
	fmt.Printf("模式: %s\n", map[bool]string{true: "预览", false: "执行"}[*dryRun])
	fmt.Printf("配置文件: %s\n", *configPath)

	if err := loadConfig(*configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	if err := initDatabase(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer database.Close()

	fmt.Println("✓ 数据库连接成功")
	fmt.Println("")

	if *dryRun {
		previewMigration()
	} else {
		runMigration()
	}
}

func loadConfig(path string) error {
	return config.Init(path)
}

func initDatabase() error {
	return database.Init(&config.GlobalConfig.Database)
}

func previewMigration() {
	var assets []model.UserAssets
	if err := database.DB.Find(&assets).Error; err != nil {
		log.Fatalf("查询用户资产失败: %v", err)
	}

	fmt.Printf("发现 %d 个用户资产记录\n", len(assets))
	fmt.Printf("%-10s %-20s %-20s %-20s\n", "用户ID", "可用余额(USDT)", "锁定余额(USDT)", "总余额(USDT)")
	fmt.Printf("%-10s %-20s %-20s %-20s\n", "------", "--------------", "--------------", "--------------")

	totalUSDT := decimal.Zero
	for _, asset := range assets {
		balance := decimal.NewFromFloat(asset.UsdtBalance)
		locked := decimal.NewFromFloat(asset.UsdtLocked)
		total := balance.Add(locked)
		totalUSDT = totalUSDT.Add(total)

		fmt.Printf("%-10d %-20s %-20s %-20s\n",
			asset.UserID,
			balance.StringFixed(8),
			locked.StringFixed(8),
			total.StringFixed(8))
	}

	fmt.Println("")
	fmt.Printf("总迁移余额: %s USDT\n", totalUSDT.StringFixed(8))
	fmt.Printf("将创建 %d 个现货钱包记录\n", len(assets))
	fmt.Println("")
	fmt.Println("预览模式 - 无实际数据变更")
}

func runMigration() {
	var assets []model.UserAssets
	if err := database.DB.Find(&assets).Error; err != nil {
		log.Fatalf("查询用户资产失败: %v", err)
	}

	fmt.Printf("发现 %d 个用户资产记录\n", len(assets))

	successCount := 0
	failCount := 0
	totalMigrated := decimal.Zero
	now := time.Now().Unix()

	for _, asset := range assets {
		availableBalance := decimal.NewFromFloat(asset.UsdtBalance)
		lockedBalance := decimal.NewFromFloat(asset.UsdtLocked)
		totalBalance := availableBalance.Add(lockedBalance)

		if totalBalance.IsZero() {
			fmt.Printf("用户ID %d 余额为0，跳过\n", asset.UserID)
			continue
		}

		err := database.DB.Transaction(func(tx *gorm.DB) error {
			var existingWallet model.UserWallet
			err := tx.Where("user_id = ? AND wallet_type = ? AND currency_id = ?",
				uint64(asset.UserID), model.WalletTypeSpot, 1).First(&existingWallet).Error

			if err == nil {
				fmt.Printf("用户ID %d 已存在现货钱包，更新余额: %s -> %s\n",
					asset.UserID, existingWallet.AvailableBalance.String(), totalBalance.String())

				result := tx.Model(&existingWallet).Updates(map[string]interface{}{
					"available_balance": totalBalance,
					"update_time":       now,
				})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("更新钱包失败")
				}
			} else if err == gorm.ErrRecordNotFound {
				wallet := &model.UserWallet{
					UserID:           uint64(asset.UserID),
					WalletType:       model.WalletTypeSpot,
					CurrencyID:       1,
					CurrencyName:     "USDT",
					AvailableBalance: totalBalance,
					LockedBalance:    decimal.Zero,
					Version:          0,
					LastTradeTime:    now,
					CreateTime:       now,
					UpdateTime:       now,
				}

				fmt.Printf("用户ID %d 创建现货钱包，余额: %s USDT\n", asset.UserID, totalBalance.String())

				if err := tx.Create(wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}

			return nil
		})

		if err != nil {
			fmt.Printf("用户ID %d 迁移失败: %v\n", asset.UserID, err)
			failCount++
		} else {
			successCount++
			totalMigrated = totalMigrated.Add(totalBalance)
		}
	}

	fmt.Println("")
	fmt.Println("=== 迁移完成 ===")
	fmt.Printf("成功: %d, 失败: %d\n", successCount, failCount)
	fmt.Printf("总迁移余额: %s USDT\n", totalMigrated.StringFixed(8))
	fmt.Println("")

	verifyMigration()
}

func verifyMigration() {
	var walletCount int64
	var totalBalance decimal.Decimal

	database.DB.Model(&model.UserWallet{}).Where("wallet_type = ?", model.WalletTypeSpot).Count(&walletCount)

	var wallets []model.UserWallet
	database.DB.Where("wallet_type = ?", model.WalletTypeSpot).Find(&wallets)

	for _, wallet := range wallets {
		totalBalance = totalBalance.Add(wallet.AvailableBalance).Add(wallet.LockedBalance)
	}

	fmt.Println("=== 验证结果 ===")
	fmt.Printf("现货钱包数量: %d\n", walletCount)
	fmt.Printf("现货钱包总余额: %s USDT\n", totalBalance.StringFixed(8))

	if walletCount > 0 {
		fmt.Println("")
		fmt.Println("✓ 余额迁移成功完成！")
	}
}
