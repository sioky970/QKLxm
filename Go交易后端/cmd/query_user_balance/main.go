package main

import (
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"column:id"`
	AccountNumber string `gorm:"column:account_number"`
	Phone        string `gorm:"column:phone"`
	Email        string `gorm:"column:email"`
}

type UserAssets struct {
	ID               uint   `gorm:"column:id"`
	UserID           uint   `gorm:"column:user_id"`
	WalletType       string `gorm:"column:wallet_type"`
	UsdtBalance      float64 `gorm:"column:usdt_balance"`
	UsdtLocked       float64 `gorm:"column:usdt_locked"`
	CurrencyBalances string `gorm:"column:currency_balances"`
	CurrencyLocked   string `gorm:"column:currency_locked"`
	DeliveryMargin   float64 `gorm:"column:delivery_margin"`
	DeliveryPnL      float64 `gorm:"column:delivery_pnl"`
	TotalValueUsdt   float64 `gorm:"column:total_value_usdt"`
}

type UserWithAssets struct {
	User
	UserAssets
}

func main() {
	// 数据库连接配置
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 查询用户
	var users []User
	if err := db.Where("account_number = ? OR phone = ?", "13333333333", "13333333333").Find(&users).Error; err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("未找到用户: 13333333333")
		return
	}

	fmt.Println("========================================")
	fmt.Println("用户信息")
	fmt.Println("========================================")
	for _, user := range users {
		fmt.Printf("ID: %d\n", user.ID)
		fmt.Printf("账号: %s\n", user.AccountNumber)
		fmt.Printf("手机: %s\n", user.Phone)
		fmt.Printf("邮箱: %s\n", user.Email)
		fmt.Println()
	}

	// 查询用户资产
	var assets []UserAssets
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	if err := db.Where("user_id IN ?", userIDs).Find(&assets).Error; err != nil {
		log.Fatalf("查询用户资产失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("账户余额")
	fmt.Println("========================================")
	
	for _, asset := range assets {
		fmt.Printf("用户ID: %d\n", asset.UserID)
		fmt.Printf("钱包类型: %s\n", asset.WalletType)
		fmt.Printf("USDT可用余额: %.8f\n", asset.UsdtBalance)
		fmt.Printf("USDT冻结余额: %.8f\n", asset.UsdtLocked)
		fmt.Printf("USDT总余额: %.8f\n", asset.UsdtBalance+asset.UsdtLocked)
		
		// 解析JSON格式的币种余额
		if asset.CurrencyBalances != "" {
			var balances map[string]float64
			if err := json.Unmarshal([]byte(asset.CurrencyBalances), &balances); err == nil {
				fmt.Println("其他币种余额:")
				for currency, balance := range balances {
					if balance > 0 {
						fmt.Printf("  %s: %.8f\n", currency, balance)
					}
				}
			}
		}
		
		if asset.DeliveryMargin > 0 {
			fmt.Printf("交割保证金: %.8f\n", asset.DeliveryMargin)
		}
		if asset.DeliveryPnL != 0 {
			fmt.Printf("交割盈亏: %.8f\n", asset.DeliveryPnL)
		}
		fmt.Printf("总价值(USDT): %.8f\n", asset.TotalValueUsdt)
		fmt.Println("----------------------------------------")
	}
}
