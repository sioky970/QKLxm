package main

import (
	"crypto/md5"
	"encoding/hex"
	"exchange-go/config"
	"exchange-go/internal/database"
	"exchange-go/internal/model"
	"fmt"
	"log"
	"time"
)

func md5Hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	db := database.DB

	// 创建测试用户
	phone := "13800138001"
	password := md5Hash("Test123456")

	var existingUser model.User
	result := db.Where("phone = ?", phone).First(&existingUser)

	var userID uint
	if result.Error != nil {
		// 用户不存在，创建新用户
		user := model.User{
			Phone:       phone,
			Password:    password,
			PayPassword: password,
			Status:      1,
			AccountType: 1,
			ParentPath:  "0",
		}
		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("创建用户失败: %v", err)
		}
		userID = user.ID
		fmt.Printf("创建测试用户成功: ID=%d, Phone=%s\n", userID, phone)
	} else {
		userID = existingUser.ID
		fmt.Printf("测试用户已存在: ID=%d, Phone=%s\n", userID, phone)
	}

	// 获取可用币种
	var currencies []model.Currency
	db.Where("is_display = 1").Find(&currencies)
	fmt.Printf("找到 %d 个可用币种\n", len(currencies))

	// 为用户创建钱包并充值
	for _, currency := range currencies {
		var wallet model.UsersWallet
		result := db.Where("user_id = ? AND currency = ?", userID, currency.ID).First(&wallet)

		if result.Error != nil {
			// 钱包不存在，创建新钱包
			wallet = model.UsersWallet{
				UserID:          userID,      // uint
				CurrencyID:      currency.ID, // 使用CurrencyID
				Address:         fmt.Sprintf("test_%d_%d", userID, currency.ID),
				UsdtBalance:     200000.0, // 统一USDT账户余额
				LockUsdtBalance: 0,
				Status:          1,
				CreateTime:      time.Now().Unix(), // int64
			}
			if err := db.Create(&wallet).Error; err != nil {
				log.Printf("创建钱包失败 [%s]: %v", currency.Name, err)
				continue
			}
			fmt.Printf("创建钱包: %s (ID=%d), USDT余额=%.2f\n", currency.Name, wallet.ID, wallet.UsdtBalance)
		} else {
			// 钱包已存在，更新余额
			db.Model(&wallet).Updates(map[string]interface{}{
				"usdt_balance": 200000.0,
			})
			fmt.Printf("更新钱包: %s (ID=%d), USDT余额=200000.00\n", currency.Name, wallet.ID)
		}
	}

	// 查询法币(USDT)的ID
	var usdt model.Currency
	if err := db.Where("name = ?", "USDT").First(&usdt).Error; err != nil {
		log.Printf("警告: 找不到USDT币种")
	} else {
		fmt.Printf("\nUSDT币种ID: %d\n", usdt.ID)
	}

	// 查询BTC的ID
	var btc model.Currency
	if err := db.Where("name = ?", "BTC").First(&btc).Error; err != nil {
		log.Printf("警告: 找不到BTC币种")
	} else {
		fmt.Printf("BTC币种ID: %d\n", btc.ID)
	}

	fmt.Println("\n=== 测试账户设置完成 ===")
	fmt.Println("登录信息:")
	fmt.Printf("  手机号: %s\n", phone)
	fmt.Printf("  密码: Test123456\n")
	fmt.Printf("  用户ID: %d\n", userID)
}
