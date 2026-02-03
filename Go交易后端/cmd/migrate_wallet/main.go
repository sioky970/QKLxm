package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:root123456@tcp(127.0.0.1:3306)/bibi2022?parseTime=true&charset=utf8mb4&loc=Local"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	statements := []string{
		`CREATE TABLE IF NOT EXISTS user_wallets (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
			user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
			wallet_type VARCHAR(20) NOT NULL COMMENT '钱包类型: spot-现货, contract-合约',
			currency_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '货币ID',
			currency_name VARCHAR(50) DEFAULT 'USDT' COMMENT '货币名称',
			available_balance DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '可用余额',
			locked_balance DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '锁定余额',
			version INT NOT NULL DEFAULT 0 COMMENT '版本号',
			last_trade_time BIGINT NOT NULL DEFAULT 0 COMMENT '最后交易时间',
			create_time BIGINT NOT NULL COMMENT '创建时间',
			update_time BIGINT NOT NULL COMMENT '更新时间',
			PRIMARY KEY (id),
			UNIQUE KEY uk_user_wallet (user_id, wallet_type, currency_id),
			KEY idx_user_id (user_id),
			KEY idx_wallet_type (wallet_type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS wallet_transfer_records (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
			transfer_no VARCHAR(64) NOT NULL COMMENT '划转流水号',
			user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
			from_wallet_type VARCHAR(20) NOT NULL COMMENT '转出钱包类型',
			to_wallet_type VARCHAR(20) NOT NULL COMMENT '转入钱包类型',
			currency_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '货币ID',
			currency_name VARCHAR(50) DEFAULT 'USDT' COMMENT '货币名称',
			amount DECIMAL(20,8) NOT NULL COMMENT '划转金额',
			fee DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '手续费',
			status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-处理中, 2-成功, 3-失败',
			remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
			client_ip VARCHAR(45) DEFAULT NULL COMMENT '客户端IP',
			created_time BIGINT NOT NULL COMMENT '创建时间',
			completed_time BIGINT DEFAULT NULL COMMENT '完成时间',
			error_msg VARCHAR(500) DEFAULT NULL COMMENT '错误信息',
			PRIMARY KEY (id),
			UNIQUE KEY uk_transfer_no (transfer_no),
			KEY idx_user_id (user_id),
			KEY idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS wallet_adjustment_records (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
			adjustment_no VARCHAR(64) NOT NULL COMMENT '调整流水号',
			user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
			wallet_type VARCHAR(20) NOT NULL COMMENT '钱包类型',
			currency_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '货币ID',
			currency_name VARCHAR(50) DEFAULT 'USDT' COMMENT '货币名称',
			adjustment_type TINYINT NOT NULL COMMENT '调整类型: 1-增加, 2-减少',
			amount DECIMAL(20,8) NOT NULL COMMENT '调整金额',
			balance_before DECIMAL(20,8) NOT NULL COMMENT '调整前余额',
			balance_after DECIMAL(20,8) NOT NULL COMMENT '调整后余额',
			operator_id BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
			operator_name VARCHAR(100) NOT NULL COMMENT '操作员名称',
			reason VARCHAR(500) NOT NULL COMMENT '调整原因',
			remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
			status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-成功, 2-失败',
			created_time BIGINT NOT NULL COMMENT '创建时间',
			completed_time BIGINT DEFAULT NULL COMMENT '完成时间',
			PRIMARY KEY (id),
			UNIQUE KEY uk_adjustment_no (adjustment_no),
			KEY idx_user_id (user_id),
			KEY idx_operator_id (operator_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	successCount := 0
	for i, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Printf("  ❌ 表%d创建失败: %v", i+1, err)
		} else {
			log.Printf("  ✅ 表%d创建成功", i+1)
			successCount++
		}
	}

	log.Println("\n==================================================")
	log.Println("📊 数据库迁移结果:")
	log.Printf("  成功: %d/%d", successCount, len(statements))
	log.Println("==================================================")

	fmt.Println("\n✅ 数据库迁移完成!")
	fmt.Println("\n新建表结构:")
	fmt.Println("  1. user_wallets (用户钱包资产表)")
	fmt.Println("  2. wallet_transfer_records (划转记录表)")
	fmt.Println("  3. wallet_adjustment_records (余额调整记录表)")
}
