package main

import (
	"exchange-go/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/model"
	"fmt"
	"log"
)

func main() {
	// 1. 初始化配置
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 1.5 初始化日志 (防止 database.Init 里的 logger.Info 崩溃)
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 2. 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	fmt.Println("正在检查并同步 currency 表结构...")

	// 3. 使用 GORM 的 AutoMigrate 自动同步表结构
	// 这将自动添加 Currency 模型中定义但数据库表中缺少的字段
	err := database.DB.AutoMigrate(&model.Currency{})
	if err != nil {
		log.Fatalf("同步表结构失败: %v", err)
	}

	fmt.Println("币种表结构同步成功！")

	// 4. 验证字段是否存在
	fmt.Println("验证风控字段...")
	var currency model.Currency
	err = database.DB.First(&currency).Error
	if err != nil {
		fmt.Printf("查询币种数据失败（可能是表为空）: %v\n", err)
	} else {
		fmt.Printf("成功查询到币种: %s (ID: %d)\n", currency.Name, currency.ID)
		fmt.Println("风控字段已在模型中定义，AutoMigrate 已尝试添加。")
	}

	fmt.Println("=== 修复完成 ===")
}
