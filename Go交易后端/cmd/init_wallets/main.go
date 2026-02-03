package main

import (
	"fmt"
	"log"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Filename, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge, cfg.Log.Compress)

	// 连接数据库
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("开始为现有用户初始化钱包...")

	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	fmt.Printf("找到 %d 个用户\n", len(users))

	walletService := service.GetWalletService()
	createdCount := 0

	for _, user := range users {
		// 检查现货钱包
		spotAssets, err := service.GetUserAssetsService().GetUserAssetsByType(user.ID, model.WalletTypeSpot)
		if err != nil {
			// 创建现货钱包
			if _, err := service.GetUserAssetsService().CreateUserAssetsByType(user.ID, model.WalletTypeSpot); err != nil {
				log.Printf("为用户 %d 创建现货钱包失败: %v", user.ID, err)
			} else {
				fmt.Printf("用户 %d 现货钱包创建成功\n", user.ID)
				createdCount++
			}
		} else {
			fmt.Printf("用户 %d 现货钱包已存在，余额: %.8f\n", user.ID, spotAssets.UsdtBalance)
		}

		// 检查合约钱包
		contractAssets, err := service.GetUserAssetsService().GetUserAssetsByType(user.ID, model.WalletTypeContract)
		if err != nil {
			// 创建合约钱包
			if _, err := service.GetUserAssetsService().CreateUserAssetsByType(user.ID, model.WalletTypeContract); err != nil {
				log.Printf("为用户 %d 创建合约钱包失败: %v", user.ID, err)
			} else {
				fmt.Printf("用户 %d 合约钱包创建成功\n", user.ID)
				createdCount++
			}
		} else {
			fmt.Printf("用户 %d 合约钱包已存在，余额: %.8f\n", user.ID, contractAssets.UsdtBalance)
		}

		// 调用旧的CreateUserWallets以确保一致性（如果需要）
		if err := walletService.CreateUserWallets(user.ID); err != nil {
			// 忽略重复创建的错误
		}
	}

	fmt.Printf("\n处理完成！共创建/验证了 %d 个钱包记录\n", createdCount)

	// 验证所有用户都有两个钱包
	var spotCount, contractCount int64
	database.DB.Model(&model.UserAssets{}).Where("wallet_type = ?", model.WalletTypeSpot).Count(&spotCount)
	database.DB.Model(&model.UserAssets{}).Where("wallet_type = ?", model.WalletTypeContract).Count(&contractCount)
	fmt.Printf("数据库中现货钱包记录数: %d\n", spotCount)
	fmt.Printf("数据库中合约钱包记录数: %d\n", contractCount)
}
