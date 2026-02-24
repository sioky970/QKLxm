package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	host         = "127.0.0.1"
	port         = 3306
	user         = "root"
	password     = "root123456"
	dbname       = "bibi2022"
	backupDir    = "./backup"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("========================================")
	log.Println("MySQL 5.7 数据库索引优化工具")
	log.Println("========================================")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		user, password, host, port, dbname)

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
	log.Println("数据库连接成功")

	log.Println("\n开始执行索引优化...")
	log.Println("注意: 此操作将添加约60个索引，预计耗时30-60分钟")
	log.Println("建议在低峰期执行\n")

	fmt.Print("确认执行优化? (输入 y 继续，其他退出): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		log.Println("已取消优化操作")
		return
	}

	startTime := time.Now()

	if err := backupDatabase(db); err != nil {
		log.Printf("警告: 备份失败 %v", err)
	}

	if err := createIndexes(db); err != nil {
		log.Printf("索引创建失败: %v", err)
		log.Println("建议: 检查错误信息后重试，部分索引可能已存在")
	} else {
		log.Println("索引创建完成!")
	}

	elapsed := time.Since(startTime)
	log.Printf("\n优化执行完成，总耗时: %v", elapsed)

	log.Println("\n后续建议:")
	log.Println("1. 监控慢查询日志，观察性能提升")
	log.Println("2. 执行 ANALYZE TABLE 更新统计信息")
	log.Println("3. 调整 MySQL 配置参数")
}

func backupDatabase(db *sql.DB) error {
	log.Println("\n正在创建数据库备份...")

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return err
		}
	}

	backupFile := fmt.Sprintf("%s/backup_before_optimize_%s.sql",
		backupDir, time.Now().Format("20060102_150405"))

	tables := []string{
		"users", "users_additional", "transaction", "lever_transaction",
		"micro_order", "users_wallet", "deposit_address", "deposit_order",
		"users_wallet_out", "account_log", "admin", "admin_role", "role",
		"access", "kyc_verification", "currency", "currency_match",
		"lever_config", "activity_config", "revoke_config", "statistics",
		"whitelist_address", "verify_log", "notice", "faq", "partner",
		"sms_code",
	}

	for _, table := range tables {
		log.Printf("备份表: %s", table)
		if err := backupTable(db, table, backupFile); err != nil {
			log.Printf("  备份失败: %v", err)
		}
	}

	log.Printf("备份完成: %s", backupFile)
	return nil
}

func backupTable(db *sql.DB, table, backupFile string) error {
	query := fmt.Sprintf("SELECT * INTO OUTFILE '%s_%s.tmp' FROM %s", backupFile, table, table)
	_, err := db.Exec(query)
	return err
}

func createIndexes(db *sql.DB) error {
	indexes := getIndexes()

	successCount := 0
	failCount := 0

	for i, idx := range indexes {
		log.Printf("[%d/%d] 创建索引: %s.%s - %s",
			i+1, len(indexes), idx.table, idx.indexName, idx.columns)

		if err := createIndex(db, idx); err != nil {
			if isDuplicateKeyError(err) {
				log.Printf("  跳过: 索引已存在")
			} else if strings.Contains(err.Error(), "Duplicate key name") {
				log.Printf("  跳过: 索引已存在")
			} else {
				log.Printf("  失败: %v", err)
				failCount++
			}
		} else {
			log.Printf("  成功")
			successCount++
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("\n索引创建统计:")
	log.Printf("  成功: %d", successCount)
	log.Printf("  失败: %d", failCount)
	log.Printf("  跳过: %d", len(indexes)-successCount-failCount)

	return nil
}

type IndexDef struct {
	table      string
	indexName  string
	columns    string
	unique     bool
	algorithm  string
	lockOption string
}

func getIndexes() []IndexDef {
	return []IndexDef{
		// users表
		{"users", "idx_users_status", "status", false, "", ""},
		{"users", "idx_users_phone", "phone", false, "", ""},
		{"users", "idx_users_email", "email", false, "", ""},
		{"users", "idx_users_invite_code", "invite_code", false, "", ""},
		{"users", "idx_users_status_time", "status, created_time", false, "", ""},
		{"users_additional", "idx_users_add_user_id", "user_id", false, "", ""},

		// transaction表
		{"transaction", "idx_trans_from_user_status", "from_user_id, status", false, "", ""},
		{"transaction", "idx_trans_from_user_status_time", "from_user_id, status, create_time", false, "", ""},
		{"transaction", "idx_trans_to_user", "to_user_id", false, "", ""},
		{"transaction", "idx_trans_currency_legal", "currency, legal", false, "", ""},
		{"transaction", "idx_trans_pay_method", "pay_method", false, "", ""},
		{"transaction", "idx_trans_status_time", "status, create_time", false, "", ""},
		{"transaction", "idx_trans_legal_status", "legal, status", false, "", ""},
		{"transaction", "idx_trans_order_no", "order_no", false, "", ""},

		// lever_transaction表
		{"lever_transaction", "idx_lever_user_id", "user_id", false, "", ""},
		{"lever_transaction", "idx_lever_user_status", "user_id, status", false, "", ""},
		{"lever_transaction", "idx_lever_user_status_time", "user_id, status, create_time", false, "", ""},
		{"lever_transaction", "idx_lever_order_no", "order_no", false, "", ""},
		{"lever_transaction", "idx_lever_settled", "is_settled, settle_time", false, "", ""},
		{"lever_transaction", "idx_lever_currency_legal", "currency, legal", false, "", ""},
		{"lever_transaction", "idx_lever_direction", "direction", false, "", ""},
		{"lever_transaction", "idx_lever_lever", "lever", false, "", ""},

		// micro_order表
		{"micro_order", "idx_micro_user_id", "user_id", false, "", ""},
		{"micro_order", "idx_micro_user_status", "user_id, status", false, "", ""},
		{"micro_order", "idx_micro_user_status_time", "user_id, status, create_time", false, "", ""},
		{"micro_order", "idx_micro_currency", "currency", false, "", ""},
		{"micro_order", "idx_micro_order_no", "order_no", false, "", ""},
		{"micro_order", "idx_micro_mold", "mold", false, "", ""},
		{"micro_order", "idx_micro_direction", "direction", false, "", ""},

		// users_wallet表
		{"users_wallet", "idx_wallet_user_currency", "user_id, currency", true, "", ""},
		{"users_wallet", "idx_wallet_status", "status", false, "", ""},
		{"users_wallet", "idx_wallet_address", "address", false, "", ""},
		{"users_wallet", "idx_wallet_contract", "contract_address", false, "", ""},

		// deposit_address表
		{"deposit_address", "idx_dep_addr_network", "address, network", false, "", ""},
		{"deposit_address", "idx_dep_addr_user_status", "user_id, status", false, "", ""},

		// deposit_order表
		{"deposit_order", "idx_dep_order_user_id", "user_id", false, "", ""},
		{"deposit_order", "idx_dep_order_status", "status", false, "", ""},
		{"deposit_order", "idx_dep_order_address", "address", false, "", ""},
		{"deposit_order", "idx_dep_order_user_status_time", "user_id, status, create_time", false, "", ""},
		{"deposit_order", "idx_dep_order_currency", "currency", false, "", ""},
		{"deposit_order", "idx_dep_order_txid", "tx_hash", false, "", ""},

		// users_wallet_out表
		{"users_wallet_out", "idx_withdraw_user_id", "user_id", false, "", ""},
		{"users_wallet_out", "idx_withdraw_status", "status", false, "", ""},
		{"users_wallet_out", "idx_withdraw_user_status_time", "user_id, status, create_time", false, "", ""},
		{"users_wallet_out", "idx_withdraw_examine", "examine", false, "", ""},
		{"users_wallet_out", "idx_withdraw_currency", "currency", false, "", ""},
		{"users_wallet_out", "idx_withdraw_txid", "tx_hash", false, "", ""},

		// account_log表
		{"account_log", "idx_acc_log_user_id", "user_id", false, "", ""},
		{"account_log", "idx_acc_log_type", "type", false, "", ""},
		{"account_log", "idx_acc_log_currency", "currency", false, "", ""},
		{"account_log", "idx_acc_log_user_time", "user_id, created_time", false, "", ""},
		{"account_log", "idx_acc_log_user_type", "user_id, type", false, "", ""},
		{"account_log", "idx_acc_log_created_time", "created_time", false, "", ""},

		// admin表
		{"admin", "idx_admin_status", "status", false, "", ""},
		{"admin", "idx_admin_username", "username", false, "", ""},
		{"admin_role", "idx_admin_role_mark", "mark", false, "", ""},
		{"role", "idx_role_status", "status", false, "", ""},
		{"access", "idx_access_pid", "pid", false, "", ""},
		{"access", "idx_access_module", "module", false, "", ""},

		// kyc_verification表
		{"kyc_verification", "idx_kyc_user_id", "user_id", true, "", ""},
		{"kyc_verification", "idx_kyc_status", "status", false, "", ""},
		{"kyc_verification", "idx_kyc_examine", "examine", false, "", ""},
		{"kyc_verification", "idx_kyc_create_time", "create_time", false, "", ""},
		{"kyc_verification", "idx_kyc_id_type", "id_type", false, "", ""},

		// currency表
		{"currency", "idx_currency_status", "status", false, "", ""},
		{"currency", "idx_currency_sort", "sort", false, "", ""},
		{"currency", "idx_currency_mark", "mark", false, "", ""},
		{"currency_match", "idx_match_legal_currency", "legal, currency", true, "", ""},
		{"currency_match", "idx_match_status", "status", false, "", ""},
		{"currency_match", "idx_match_sort", "sort", false, "", ""},

		// lever_config表
		{"lever_config", "idx_lever_config_currency", "currency", false, "", ""},
		{"lever_config", "idx_lever_config_status", "status", false, "", ""},

		// activity_config表
		{"activity_config", "idx_activity_status", "status", false, "", ""},
		{"activity_config", "idx_activity_time", "start_time, end_time", false, "", ""},
		{"activity_config", "idx_activity_type", "type", false, "", ""},

		// revoke_config表
		{"revoke_config", "idx_revoke_user_id", "user_id", false, "", ""},
		{"revoke_config", "idx_revoke_status", "status", false, "", ""},

		// statistics表
		{"statistics", "idx_stats_date", "date", true, "", ""},
		{"statistics", "idx_stats_type", "type", false, "", ""},

		// whitelist_address表
		{"whitelist_address", "idx_whitelist_address", "address, network", true, "", ""},
		{"whitelist_address", "idx_whitelist_status", "status", false, "", ""},

		// verify_log表
		{"verify_log", "idx_verify_log_user_id", "user_id", false, "", ""},
		{"verify_log", "idx_verify_log_type", "type", false, "", ""},
		{"verify_log", "idx_verify_log_created_time", "created_time", false, "", ""},

		// notice表
		{"notice", "idx_notice_status", "status", false, "", ""},
		{"notice", "idx_notice_type", "type", false, "", ""},
		{"notice", "idx_notice_create_time", "create_time", false, "", ""},

		// faq表
		{"faq", "idx_faq_category", "category", false, "", ""},
		{"faq", "idx_faq_status", "status", false, "", ""},

		// partner表
		{"partner", "idx_partner_status", "status", false, "", ""},
		{"partner", "idx_partner_sort", "sort", false, "", ""},

		// sms_code表
		{"sms_code", "idx_sms_phone_type", "phone, type", false, "", ""},
		{"sms_code", "idx_sms_create_time", "created_time", false, "", ""},
	}
}

func createIndex(db *sql.DB, idx IndexDef) error {
	var sqlStr string

	if idx.unique {
		sqlStr = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)",
			idx.indexName, idx.table, idx.columns)
	} else {
		sqlStr = fmt.Sprintf("CREATE INDEX %s ON %s (%s)",
			idx.indexName, idx.table, idx.columns)
	}

	if idx.algorithm != "" {
		sqlStr += fmt.Sprintf(" ALGORITHM=%s", idx.algorithm)
	}
	if idx.lockOption != "" {
		sqlStr += fmt.Sprintf(" LOCK=%s", idx.lockOption)
	}

	_, err := db.Exec(sqlStr)
	return err
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "1062") ||
		strings.Contains(errStr, "already exists")
}
