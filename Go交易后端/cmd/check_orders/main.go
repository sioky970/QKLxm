package main

import (
	"exchange-go/config"
	"exchange-go/internal/model"
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

	// 查询最新的5条订单
	var orders []model.Transaction
	result := database.DB.Order("id DESC").Limit(5).Find(&orders)
	if result.Error != nil {
		fmt.Printf("❌ 查询订单失败: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("最新的5条订单：")
	fmt.Println("================================================================================")
	for _, order := range orders {
		fmt.Printf("ID: %d\n", order.ID)
		fmt.Printf("  FromUserID: %d\n", order.FromUserID)
		fmt.Printf("  Currency (int): %d\n", order.Currency)
		fmt.Printf("  CurrencyID (uint): %d\n", order.CurrencyID)
		fmt.Printf("  LegalID: %d\n", order.LegalID)
		fmt.Printf("  Type: %d (1=买入, 2=卖出)\n", order.Type)
		fmt.Printf("  Price: %.4f\n", order.Price)
		fmt.Printf("  Number: %.8f\n", order.Number)
		fmt.Printf("  DealNumber: %.8f\n", order.DealNumber)
		fmt.Printf("  Status: %d (0=未成交, 1=部分成交, 2=已成交, 3=已撤销)\n", order.Status)
		fmt.Println("--------------------------------------------------------------------------------")
	}

	// 直接查询数据库的原始值
	type RawOrder struct {
		ID       uint
		Currency int
	}
	var rawOrders []RawOrder
	database.DB.Raw("SELECT id, currency FROM transaction ORDER BY id DESC LIMIT 5").Scan(&rawOrders)
	fmt.Println("\n数据库原始currency字段值：")
	fmt.Println("================================================================================")
	for _, raw := range rawOrders {
		fmt.Printf("ID: %d, Currency: %d\n", raw.ID, raw.Currency)
	}

	// 查询币种列表
	var currencies []model.Currency
	database.DB.Find(&currencies)
	fmt.Println("\n所有币种：")
	fmt.Println("================================================================================")
	for _, c := range currencies {
		fmt.Printf("ID: %d, Name: %s\n", c.ID, c.Name)
	}
}
