package main

import (
	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func main() {
	// 1. 初始化配置
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志 (database.Init 内部会用到 logger)
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	// 2. 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	db := database.DB

	// 3. 定义要插入的倍数（全局统一配置）
	multiples := []string{
		"10", "20", "30", "40", "50", "60", "70", "80", "90", "100", "120", "150", "180",
	}

	// 4. 执行更新
	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除所有旧的倍数设置（包括按币种配置的）
		if err := tx.Where("type = ?", 1).Delete(&model.LeverMultiple{}).Error; err != nil {
			return err
		}

		// 插入新的全局倍数（currency_id为nil）
		for _, v := range multiples {
			item := model.LeverMultiple{
				Type:       1, // 1=倍数
				Value:      v,
				CurrencyID: nil, // 全局配置
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("更新杠杆倍数失败: %v", err)
	}

	fmt.Println("成功更新杠杆倍数至数据库！")
	fmt.Printf("已插入全局配置: %v\n", multiples)
	fmt.Println("注意：所有币种现在使用相同的杠杆倍数")
}
