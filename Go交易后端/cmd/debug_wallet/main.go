package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询fund_wallets表
	fmt.Println("=== fund_wallets 表数据 ===")
	rows, err := db.Query("SELECT id, user_id, currency_id, available_balance FROM fund_wallets")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("%-5s %-10s %-12s %-20s\n", "ID", "UserID", "CurrencyID", "AvailableBalance")
	for rows.Next() {
		var id, userID, currencyID int
		var availableBalance string
		rows.Scan(&id, &userID, &currencyID, &availableBalance)
		fmt.Printf("%-5d %-10d %-12d %-20s\n", id, userID, currencyID, availableBalance)
	}

	// 查询users表
	fmt.Println("\n=== users 表数据 ===")
	rows2, err := db.Query("SELECT id, account_number FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	fmt.Printf("%-5s %-20s\n", "ID", "AccountNumber")
	for rows2.Next() {
		var id int
		var accountNumber string
		rows2.Scan(&id, &accountNumber)
		fmt.Printf("%-5d %-20s\n", id, accountNumber)
	}

	// 检查用户ID 1和3的资金钱包
	fmt.Println("\n=== 检查特定用户的资金钱包 ===")
	for _, userID := range []int{1, 3} {
		var availableBalance string
		err := db.QueryRow("SELECT available_balance FROM fund_wallets WHERE user_id = ?", userID).Scan(&availableBalance)
		if err == sql.ErrNoRows {
			fmt.Printf("UserID %d: 没有找到资金钱包记录\n", userID)
		} else if err != nil {
			fmt.Printf("UserID %d: 查询错误: %v\n", userID, err)
		} else {
			fmt.Printf("UserID %d: available_balance = %s\n", userID, availableBalance)
		}
	}
}
