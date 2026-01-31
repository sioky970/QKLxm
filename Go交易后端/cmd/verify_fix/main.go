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

	// 2. 初始化日志
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 3. 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 4. 测试更新 ID 为 1 的币种 (BTC)
	fmt.Println("正在测试更新 BTC (ID: 1) 的风控参数...")
	
	updates := map[string]interface{}{
		"risk_prob_enabled":       1,
		"risk_profit_probability": 60,
		"risk_money_enabled":      1,
		"risk_money_min":          100.0,
		"risk_money_max":          1000.0,
		"risk_money_result":       1,
		"risk_time_enabled":       0,
	}

	err := database.DB.Model(&model.Currency{}).Where("id = ?", 1).Updates(updates).Error
	if err != nil {
		log.Fatalf("更新失败: %v", err)
	}

	fmt.Println("更新成功！")

	// 5. 查询并验证
	var currency model.Currency
	err = database.DB.First(&currency, 1).Error
	if err != nil {
		log.Fatalf("查询验证失败: %v", err)
	}

	fmt.Printf("验证结果:\n")
	fmt.Printf("  RiskProbEnabled: %d\n", currency.RiskProbEnabled)
	fmt.Printf("  RiskProfitProbability: %d\n", currency.RiskProfitProbability)
	fmt.Printf("  RiskMoneyMin: %f\n", currency.RiskMoneyMin)

	if currency.RiskProbEnabled == 1 && currency.RiskProfitProbability == 60 {
		fmt.Println("数据一致，风控参数保存功能修复成功！")
	} else {
		fmt.Println("数据不一致，请检查原因。")
	}

	fmt.Println("=== 验证完成 ===")
}
