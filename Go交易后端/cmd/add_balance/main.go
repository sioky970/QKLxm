package main

import (
	"fmt"
	"log"
	"time"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	if err := config.Init("../../config/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 初始化日志
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 3. 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer func() {
		sqlDB, _ := database.DB.DB()
		sqlDB.Close()
	}()

	// 4. 执行余额增加操作
	accountNumber := "test@test.com"
	amount := 1000000.0
	reason := "后台充值 - 管理员手动添加"

	fmt.Printf("======================================\n")
	fmt.Printf("为用户 %s 增加 %.2f USDT\n", accountNumber, amount)
	fmt.Printf("======================================\n\n")

	err := addUserBalance(accountNumber, amount, reason)
	if err != nil {
		log.Fatalf("❌ 操作失败: %v", err)
	}

	fmt.Printf("\n✅ 操作成功！\n")
}

func addUserBalance(accountNumber string, amount float64, reason string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查询用户信息
		fmt.Println("步骤1：查询用户信息...")
		var user model.User
		if err := tx.Where("account_number = ? OR phone = ?", accountNumber, accountNumber).First(&user).Error; err != nil {
			return fmt.Errorf("用户不存在: %v", err)
		}
		fmt.Printf("  ✓ 找到用户: ID=%d, 账号=%s, 状态=%d\n", user.ID, user.AccountNumber, user.Status)

		// 2. 查询USDT钱包（currency=3）
		fmt.Println("\n步骤2：查询USDT钱包...")
		var wallet model.UsersWallet
		err := tx.Where("user_id = ? AND currency = ?", user.ID, 3).First(&wallet).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 钱包不存在，自动创建
				fmt.Println("  ⚠ USDT钱包不存在，正在创建...")
				wallet = model.UsersWallet{
					UserID:          user.ID,
					CurrencyID:      3, // USDT
					UsdtBalance:     0,
					LockUsdtBalance: 0,
					Status:          1,
				}
				if err := tx.Create(&wallet).Error; err != nil {
					return fmt.Errorf("创建USDT钱包失败: %v", err)
				}
				fmt.Printf("  ✓ USDT钱包已创建: ID=%d\n", wallet.ID)
			} else {
				return fmt.Errorf("USDT钱包查询失败: %v", err)
			}
		}
		fmt.Printf("  ✓ 找到钱包: ID=%d, 当前余额=%.8f USDT\n", wallet.ID, wallet.UsdtBalance)

		// 3. 更新余额
		fmt.Println("\n步骤3：更新余额...")
		oldBalance := wallet.UsdtBalance
		newBalance := oldBalance + amount

		if err := tx.Model(&wallet).Update("usdt_balance", newBalance).Error; err != nil {
			return fmt.Errorf("更新余额失败: %v", err)
		}
		fmt.Printf("  ✓ 余额已更新:\n")
		fmt.Printf("    - 原余额: %.8f USDT\n", oldBalance)
		fmt.Printf("    - 增加: +%.8f USDT\n", amount)
		fmt.Printf("    - 新余额: %.8f USDT\n", newBalance)

		// 4. 记录账户日志
		fmt.Println("\n步骤4：记录账户日志...")
		accountLog := model.AccountLog{
			UserID:      user.ID,
			Value:       amount,
			CurrencyID:  3,   // USDT
			Type:        101, // 类型101表示余额增加
			Info:        reason,
			CreatedTime: time.Now().Unix(),
		}
		if err := tx.Create(&accountLog).Error; err != nil {
			return fmt.Errorf("创建账户日志失败: %v", err)
		}
		fmt.Printf("  ✓ 账户日志已创建: ID=%d, 类型=101(余额增加)\n", accountLog.ID)

		// 5. 验证最终余额
		fmt.Println("\n步骤5：验证最终余额...")
		var verifyWallet model.UsersWallet
		if err := tx.Where("user_id = ? AND currency = ?", user.ID, 3).First(&verifyWallet).Error; err != nil {
			return fmt.Errorf("验证失败: %v", err)
		}
		fmt.Printf("  ✓ 验证通过: 最终余额=%.8f USDT\n", verifyWallet.UsdtBalance)

		return nil
	})
}
