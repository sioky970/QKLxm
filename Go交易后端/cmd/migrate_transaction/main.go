package main

import (
	"exchange-go/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"flag"
	"fmt"
	"os"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./config/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 初始化配置
	if err := config.Init(configPath); err != nil {
		fmt.Printf("❌ 初始化配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 配置加载成功")

	// 初始化日志
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		fmt.Printf("❌ 初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	fmt.Println("✓ 日志系统初始化成功")

	// 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		fmt.Printf("❌ 初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("✓ 数据库连接成功")

	// 执行迁移
	fmt.Println("\n开始迁移transaction表...")

	// 1. 添加 legal_id 字段
	sql1 := `
		ALTER TABLE transaction 
		ADD COLUMN legal_id INT(10) UNSIGNED NOT NULL DEFAULT 3 COMMENT '法币ID(3=USDT)' 
		AFTER currency
	`

	if err := database.DB.Exec(sql1).Error; err != nil {
		// 如果字段已存在，忽略错误
		if fmt.Sprintf("%v", err) == "Error 1060 (42S21): Duplicate column name 'legal_id'" {
			fmt.Println("  legal_id 字段已存在，跳过")
		} else {
			fmt.Printf("❌ 添加 legal_id 字段失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("✓ 添加 legal_id 字段成功")
	}

	// 2. 添加 deal_number 字段
	sql2 := `
		ALTER TABLE transaction 
		ADD COLUMN deal_number DECIMAL(20,5) NOT NULL DEFAULT 0.00000 COMMENT '成交数量' 
		AFTER number
	`

	if err := database.DB.Exec(sql2).Error; err != nil {
		// 如果字段已存在，忽略错误
		if fmt.Sprintf("%v", err) == "Error 1060 (42S21): Duplicate column name 'deal_number'" {
			fmt.Println("  deal_number 字段已存在，跳过")
		} else {
			fmt.Printf("❌ 添加 deal_number 字段失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("✓ 添加 deal_number 字段成功")
	}

	fmt.Println("\n✅ transaction表迁移完成！")
	fmt.Println("\n表结构更新:")
	fmt.Println("  - legal_id: INT(10) UNSIGNED DEFAULT 3 (法币ID)")
	fmt.Println("  - deal_number: DECIMAL(20,5) DEFAULT 0 (成交数量)")
}
