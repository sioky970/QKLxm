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

	// 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		fmt.Printf("❌ 初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("✓ 数据库连接成功\n")

	// 1. 添加 order_value 字段
	sql1 := `
		ALTER TABLE transaction 
		ADD COLUMN order_value DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '订单价值(USDT)' 
		AFTER deal_number
	`

	fmt.Println("正在添加 order_value 字段...")
	if err := database.DB.Exec(sql1).Error; err != nil {
		// 检查是否是字段已存在的错误
		if err.Error() != "Error 1060 (42S21): Duplicate column name 'order_value'" {
			fmt.Printf("❌ 添加字段失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("⚠ order_value 字段已存在，跳过")
	} else {
		fmt.Println("✅ order_value 字段添加成功")
	}

	// 2. 更新现有订单的order_value（根据价格和数量计算）
	sql2 := `
		UPDATE transaction 
		SET order_value = price * number 
		WHERE order_value = 0 AND price > 0
	`

	fmt.Println("\n正在更新现有订单的order_value...")
	result := database.DB.Exec(sql2)
	if result.Error != nil {
		fmt.Printf("❌ 更新失败: %v\n", result.Error)
		os.Exit(1)
	}
	fmt.Printf("✅ 已更新 %d 条订单记录\n", result.RowsAffected)

	// 3. 验证结果
	var count int64
	database.DB.Raw("SELECT COUNT(*) FROM transaction WHERE order_value > 0").Scan(&count)
	fmt.Printf("\n✅ transaction表迁移完成！当前有 %d 条订单有价值数据\n", count)
}
