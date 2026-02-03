package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
)

func main() {
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("=== 模拟用户列表API逻辑 ===")

	// 1. 查询用户列表
	rows, err := db.Query("SELECT id, account_number FROM users ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var accountNumber string
		rows.Scan(&id, &accountNumber)

		// 2. 获取资金钱包余额
		var availableBalance sql.NullString
		err = db.QueryRow(`
			SELECT available_balance FROM fund_wallets WHERE user_id = ?`, id).Scan(&availableBalance)

		if err == sql.ErrNoRows {
			fmt.Printf("用户 %d (%s): 没有资金钱包记录, FundBalance = 0\n", id, accountNumber)
		} else if err != nil {
			fmt.Printf("用户 %d (%s): 查询错误: %v\n", id, accountNumber, err)
		} else {
			// 转换为float64
			balanceStr := availableBalance.String
			availableDecimal, err := decimal.NewFromString(balanceStr)
			if err != nil {
				fmt.Printf("用户 %d (%s): 解析余额失败: %v\n", id, accountNumber, err)
				continue
			}
			fundBalance, _ := availableDecimal.Float64()
			fmt.Printf("用户 %d (%s): AvailableBalance = %s, FundBalance(float64) = %.4f\n", 
				id, accountNumber, balanceStr, fundBalance)
		}
	}
}
