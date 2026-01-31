package main

import (
	"fmt"
	"log"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

func main() {
	// 1. 加载配置
	if err := config.Init("config/config.yaml"); err != nil {
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

	// 4. 删除火币不支持的币种
	invalidCurrencies := []string{"HT", "OMG", "MKR", "EOS"}

	fmt.Println("======================================")
	fmt.Println("删除火币不支持的币种")
	fmt.Println("======================================\n")

	// 先查询这些币种的信息
	var currencies []model.Currency
	if err := database.DB.Where("name IN ?", invalidCurrencies).Find(&currencies).Error; err != nil {
		log.Fatalf("查询币种失败: %v", err)
	}

	if len(currencies) == 0 {
		fmt.Println("✓ 数据库中没有这些币种，无需删除")
		return
	}

	fmt.Println("找到以下币种：")
	for _, c := range currencies {
		fmt.Printf("  - ID: %d, 名称: %s, 显示状态: %d, 排序: %d\n", c.ID, c.Name, c.IsDisplay, c.Sort)
	}
	fmt.Println()

	// 执行删除
	result := database.DB.Where("name IN ?", invalidCurrencies).Delete(&model.Currency{})
	if result.Error != nil {
		log.Fatalf("❌ 删除失败: %v", result.Error)
	}

	fmt.Printf("✅ 成功删除 %d 个币种\n", result.RowsAffected)
	fmt.Println("\n删除的币种：")
	for _, name := range invalidCurrencies {
		fmt.Printf("  - %s\n", name)
	}
}
