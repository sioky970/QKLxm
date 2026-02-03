package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type IndexDef struct {
	Table     string
	IndexName string
	Columns   string
	Unique    bool
}

func getIndexes() []IndexDef {
	return []IndexDef{
		// currency_quotation 报价表索引
		{"currency_quotation", "idx_q_currency_legal", "currency_id, legal_id", false},
		{"currency_quotation", "idx_q_add_time", "add_time DESC", false},
		{"currency_quotation", "idx_q_currency_legal_time", "currency_id, legal_id, add_time DESC", false},

		// market_hour K线表索引
		{"market_hour", "idx_mh_currency_legal_period", "currency_id, legal_id, period", false},
		{"market_hour", "idx_mh_currency_period", "currency_id, period", false},

		// transaction 交易订单表（补充）
		{"transaction", "idx_trans_from_user_status", "from_user_id, status", false},
		{"transaction", "idx_trans_from_user_status_time", "from_user_id, status, create_time", false},
		{"transaction", "idx_trans_status_time", "status, create_time", false},
		{"transaction", "idx_trans_legal_status", "legal, status", false},
		{"transaction", "idx_trans_order_no", "order_no", false},

		// lever_transaction 合约交易表
		{"lever_transaction", "idx_lever_user_status", "user_id, status", false},
		{"lever_transaction", "idx_lever_user_status_time", "user_id, status, create_time DESC", false},
		{"lever_transaction", "idx_lever_order_no", "order_no", false},
		{"lever_transaction", "idx_lever_settled", "is_settled, settle_time", false},

		// micro_order 秒合约订单表
		{"micro_order", "idx_micro_user_status", "user_id, status", false},
		{"micro_order", "idx_micro_user_status_time", "user_id, status, create_time DESC", false},
		{"micro_order", "idx_micro_order_no", "order_no", false},

		// users 用户表
		{"users", "idx_users_status", "status", false},
		{"users", "idx_users_status_time", "status, created_time", false},

		// account_log 账户流水表
		{"account_log", "idx_acc_log_user_time", "user_id, created_time DESC", false},
		{"account_log", "idx_acc_log_user_type", "user_id, type", false},

		// users_wallet 用户钱包表
		{"users_wallet", "idx_wallet_user_currency", "user_id, currency", true},

		// deposit_order 充值订单表
		{"deposit_order", "idx_dep_order_user_status_time", "user_id, status, create_time DESC", false},
		{"deposit_order", "idx_dep_order_txid", "tx_hash", false},

		// users_wallet_out 提现订单表
		{"users_wallet_out", "idx_withdraw_user_status_time", "user_id, status, create_time DESC", false},
		{"users_wallet_out", "idx_withdraw_txid", "tx_hash", false},
	}
}

func createIndex(db *sql.DB, idx IndexDef) error {
	var sql string
	if idx.Unique {
		sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", idx.IndexName, idx.Table, idx.Columns)
	} else {
		sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.IndexName, idx.Table, idx.Columns)
	}

	_, err := db.Exec(sql)
	return err
}

func indexExists(db *sql.DB, tableName, indexName string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		AND table_name = ?
		AND index_name = ?
	`
	var count int
	err := db.QueryRow(query, tableName, indexName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func main() {
	dsn := "root:root123456@tcp(127.0.0.1:3306)/bibi2022?parseTime=true&charset=utf8mb4&loc=Local"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	indexes := getIndexes()
	successCount := 0
	failCount := 0
	skipCount := 0

	for i, idx := range indexes {
		log.Printf("[%d/%d] 检查索引: %s.%s - %s",
			i+1, len(indexes), idx.Table, idx.IndexName, idx.Columns)

		exists, err := indexExists(db, idx.Table, idx.IndexName)
		if err != nil {
			log.Printf("  ❌ 检查失败: %v", err)
			failCount++
			continue
		}

		if exists {
			log.Printf("  ⏭️  跳过: 索引已存在")
			skipCount++
			continue
		}

		if err := createIndex(db, idx); err != nil {
			log.Printf("  ❌ 创建失败: %v", err)
			failCount++
		} else {
			log.Printf("  ✅ 成功")
			successCount++
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Println("\n" + strings.Repeat("=", 50))
	log.Println("📊 索引创建统计:")
	log.Printf("  ✅ 成功: %d", successCount)
	log.Printf("  ❌ 失败: %d", failCount)
	log.Printf("  ⏭️  跳过: %d", skipCount)
	log.Printf("  📦 总计: %d", len(indexes))
	log.Println(strings.Repeat("=", 50))
}
