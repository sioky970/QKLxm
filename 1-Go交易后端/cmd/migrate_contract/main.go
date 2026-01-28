package main

import (
	"fmt"
	"os"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("====================================")
	fmt.Println("永续合约交易数据库迁移脚本")
	fmt.Println("====================================")

	// 1. 加载配置
	if err := config.Init("config/config.yaml"); err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 配置加载成功")

	// 2. 初始化日志
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		fmt.Printf("❌ 初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	fmt.Println("✓ 日志系统初始化成功")

	// 3. 连接数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		fmt.Printf("❌ 连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("✓ 数据库连接成功")

	db := database.DB

	// 3. 执行迁移
	fmt.Println("\n开始执行数据库迁移...")

	// 3.1 创建风控K线表
	if err := createRiskKlineTable(db); err != nil {
		fmt.Printf("❌ 创建风控K线表失败: %v\n", err)
		os.Exit(1)
	}

	// 3.2 创建用户风控配置表
	if err := createUserRiskControlTable(db); err != nil {
		fmt.Printf("❌ 创建用户风控配置表失败: %v\n", err)
		os.Exit(1)
	}

	// 3.3 补充lever_transaction表字段
	if err := alterLeverTransactionTable(db); err != nil {
		fmt.Printf("❌ 补充合约交易表字段失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n====================================")
	fmt.Println("✅ 数据库迁移完成!")
	fmt.Println("====================================")
}

// createRiskKlineTable 创建风控K线表
func createRiskKlineTable(db *gorm.DB) error {
	fmt.Println("\n[1/3] 创建风控K线表 (risk_kline)...")

	// 检查表是否存在
	if db.Migrator().HasTable(&model.RiskKline{}) {
		fmt.Println("  ⚠ 表已存在，跳过创建")
		return nil
	}

	// 自动迁移创建表
	if err := db.AutoMigrate(&model.RiskKline{}); err != nil {
		return err
	}

	fmt.Println("  ✓ 风控K线表创建成功")
	return nil
}

// createUserRiskControlTable 创建用户风控配置表
func createUserRiskControlTable(db *gorm.DB) error {
	fmt.Println("\n[2/3] 创建用户风控配置表 (user_risk_control)...")

	// 检查表是否存在
	if db.Migrator().HasTable(&model.UserRiskControl{}) {
		fmt.Println("  ⚠ 表已存在，跳过创建")
		return nil
	}

	// 自动迁移创建表
	if err := db.AutoMigrate(&model.UserRiskControl{}); err != nil {
		return err
	}

	fmt.Println("  ✓ 用户风控配置表创建成功")
	return nil
}

// alterLeverTransactionTable 补充lever_transaction表字段
func alterLeverTransactionTable(db *gorm.DB) error {
	fmt.Println("\n[3/3] 补充合约交易表字段 (lever_transaction)...")

	// 检查表是否存在
	if !db.Migrator().HasTable("lever_transaction") {
		fmt.Println("  ⚠ lever_transaction表不存在，跳过")
		return nil
	}

	// 需要添加的字段列表
	alterColumns := []struct {
		column   string
		dataType string
		comment  string
	}{
		{"order_type", "TINYINT(1) NOT NULL DEFAULT 1", "订单类型:1=市价单,2=限价单"},
		{"limit_price", "DECIMAL(20,8) DEFAULT NULL", "限价价格"},
		{"take_profit_amount", "DECIMAL(20,8) DEFAULT NULL", "止盈金额(USDT)"},
		{"stop_loss_amount", "DECIMAL(20,8) DEFAULT NULL", "止损金额(USDT)"},
		{"actual_margin", "DECIMAL(20,8) DEFAULT NULL", "实际保证金(扣除手续费后)"},
		{"close_type", "TINYINT(1) DEFAULT NULL", "平仓类型:1=手动,2=爆仓,3=止盈,4=止损"},
		{"close_price", "DECIMAL(20,8) DEFAULT NULL", "平仓价格"},
	}

	addedCount := 0
	for _, col := range alterColumns {
		// 检查字段是否存在
		if db.Migrator().HasColumn(&model.LeverTransaction{}, col.column) {
			fmt.Printf("  ⚠ 字段 %s 已存在，跳过\n", col.column)
			continue
		}

		// 添加字段
		sql := fmt.Sprintf("ALTER TABLE `lever_transaction` ADD COLUMN `%s` %s COMMENT '%s'",
			col.column, col.dataType, col.comment)

		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("  ⚠ 添加字段 %s 失败: %v\n", col.column, err)
			// 继续处理其他字段，不中断
		} else {
			fmt.Printf("  ✓ 添加字段: %s\n", col.column)
			addedCount++
		}
	}

	if addedCount == 0 {
		fmt.Println("  ⚠ 所有字段已存在，无需添加")
	} else {
		fmt.Printf("  ✓ 成功添加 %d 个字段\n", addedCount)
	}

	return nil
}
