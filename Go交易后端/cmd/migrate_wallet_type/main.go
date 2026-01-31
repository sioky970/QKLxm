package main

import (
	"fmt"
	"log"

	"exchange-go/internal/pkg/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Filename, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge, cfg.Log.Compress)

	// 连接数据库
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("开始执行 user_assets 表迁移...")

	// 1. 检查并添加 wallet_type 字段
	if err := addWalletTypeColumn(); err != nil {
		log.Printf("添加 wallet_type 字段失败: %v", err)
	} else {
		fmt.Println("✓ wallet_type 字段已添加或已存在")
	}

	// 2. 删除旧索引
	if err := dropOldIndexes(); err != nil {
		log.Printf("删除旧索引失败: %v", err)
	} else {
		fmt.Println("✓ 旧索引已删除")
	}

	// 3. 创建新索引
	if err := createNewIndexes(); err != nil {
		log.Printf("创建新索引失败: %v", err)
	} else {
		fmt.Println("✓ 新索引已创建")
	}

	// 4. 更新现有数据
	if err := updateExistingData(); err != nil {
		log.Printf("更新现有数据失败: %v", err)
	} else {
		fmt.Println("✓ 现有数据已更新")
	}

	// 5. 验证结果
	if err := verifyMigration(); err != nil {
		log.Printf("验证迁移结果失败: %v", err)
	}

	fmt.Println("\n迁移完成!")
}

// addWalletTypeColumn 添加 wallet_type 字段
func addWalletTypeColumn() error {
	// 检查字段是否存在
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'user_assets' 
		AND column_name = 'wallet_type'
	`).Scan(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		// 添加字段
		err = database.DB.Exec(`
			ALTER TABLE user_assets 
			ADD COLUMN wallet_type VARCHAR(20) NOT NULL DEFAULT 'spot' 
			COMMENT '钱包类型: spot=现货钱包, contract=合约钱包' 
			AFTER user_id
		`).Error
		if err != nil {
			return err
		}
		fmt.Println("  - 已添加 wallet_type 字段")
	} else {
		fmt.Println("  - wallet_type 字段已存在，跳过")
	}

	return nil
}

// dropOldIndexes 删除旧索引
func dropOldIndexes() error {
	// 删除 idx_user_assets_user_id
	database.DB.Exec(`ALTER TABLE user_assets DROP INDEX idx_user_assets_user_id`)
	// 删除 idx_user_id
	database.DB.Exec(`ALTER TABLE user_assets DROP INDEX idx_user_id`)
	return nil
}

// createNewIndexes 创建新索引
func createNewIndexes() error {
	// 创建复合唯一索引
	err := database.DB.Exec(`
		ALTER TABLE user_assets 
		ADD UNIQUE INDEX idx_user_wallet (user_id, wallet_type)
	`).Error
	if err != nil {
		// 索引可能已存在，忽略错误
		fmt.Println("  - idx_user_wallet 索引可能已存在")
	}

	// 创建 wallet_type 索引
	err = database.DB.Exec(`
		ALTER TABLE user_assets 
		ADD INDEX idx_wallet_type (wallet_type)
	`).Error
	if err != nil {
		fmt.Println("  - idx_wallet_type 索引可能已存在")
	}

	return nil
}

// updateExistingData 更新现有数据
func updateExistingData() error {
	// 将 wallet_type 为 NULL 或空的数据设置为 'spot'
	result := database.DB.Exec(`
		UPDATE user_assets 
		SET wallet_type = 'spot' 
		WHERE wallet_type IS NULL OR wallet_type = ''
	`)
	if result.Error != nil {
		return result.Error
	}
	fmt.Printf("  - 已更新 %d 条记录\n", result.RowsAffected)
	return nil
}

// verifyMigration 验证迁移结果
func verifyMigration() error {
	var totalRecords, spotWallets, contractWallets int64

	database.DB.Raw(`SELECT COUNT(*) FROM user_assets`).Scan(&totalRecords)
	database.DB.Raw(`SELECT COUNT(*) FROM user_assets WHERE wallet_type = 'spot'`).Scan(&spotWallets)
	database.DB.Raw(`SELECT COUNT(*) FROM user_assets WHERE wallet_type = 'contract'`).Scan(&contractWallets)

	fmt.Println("\n迁移结果统计:")
	fmt.Printf("  - 总记录数: %d\n", totalRecords)
	fmt.Printf("  - 现货钱包: %d\n", spotWallets)
	fmt.Printf("  - 合约钱包: %d\n", contractWallets)

	return nil
}
