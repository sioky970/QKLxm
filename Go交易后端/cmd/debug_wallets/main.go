package main

import (
	"fmt"
	"log"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
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

	fmt.Println("检查 UserAssets 表中的记录...")

	// 查询所有 UserAssets 记录
	var assets []model.UserAssets
	result := database.DB.Find(&assets)
	if result.Error != nil {
		log.Fatalf("查询失败: %v", result.Error)
	}

	fmt.Printf("共找到 %d 条 UserAssets 记录\n\n", len(assets))

	// 按用户分组显示
	userMap := make(map[uint][]model.UserAssets)
	for _, asset := range assets {
		userMap[asset.UserID] = append(userMap[asset.UserID], asset)
	}

	for userID, wallets := range userMap {
		fmt.Printf("用户 %d:\n", userID)
		for _, wallet := range wallets {
			fmt.Printf("  - WalletType: %s, USDT余额: %.8f, 锁定: %.8f\n", 
				wallet.WalletType, wallet.UsdtBalance, wallet.UsdtLocked)
		}
	}

	// 检查是否有缺失钱包类型的记录
	var oldAssets []model.UserAssets
	database.DB.Where("wallet_type = ? OR wallet_type = ?", "", nil).Find(&oldAssets)
	if len(oldAssets) > 0 {
		fmt.Printf("\n警告: 找到 %d 条没有设置 wallet_type 的旧记录\n", len(oldAssets))
		for _, asset := range oldAssets {
			fmt.Printf("  - UserID: %d, USDT余额: %.8f\n", asset.UserID, asset.UsdtBalance)
		}
	}
}
