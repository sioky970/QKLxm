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

	tables := []string{"user_wallets", "wallet_transfer_records", "wallet_adjustment_records"}
	
	log.Println("\n📋 验证数据库表:")
	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			log.Printf("  ❌ %s: 查询失败 - %v", table, err)
		} else {
			log.Printf("  ✅ %s: %d 条记录", table, count)
		}
	}

	log.Println("\n✅ 数据库验证完成!")
}
