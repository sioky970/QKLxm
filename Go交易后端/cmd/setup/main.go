package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 使用与项目相同的配置
const (
	dsn = "root:root123456@tcp(127.0.0.1:3306)/bibi2022?charset=utf8mb4&parseTime=True&loc=Local"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("初始化测试数据")
	fmt.Println("========================================")

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库连接成功")

	// 1. 创建测试管理员角色
	fmt.Println("\n[1] 创建管理员角色...")
	roleSQL := `INSERT INTO admin_role (id, name, is_super) VALUES (1, '超级管理员', 1) 
		ON DUPLICATE KEY UPDATE name='超级管理员', is_super=1`
	if err := db.Exec(roleSQL).Error; err != nil {
		fmt.Printf("⚠️  角色创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 管理员角色创建成功")
	}

	// 2. 创建测试管理员账户
	fmt.Println("\n[2] 创建测试管理员账户...")
	password := hashPassword("admin123")
	adminSQL := `INSERT INTO admin (id, username, password, role_id) VALUES (1, 'admin', ?, 1)
		ON DUPLICATE KEY UPDATE password=?, role_id=1`
	if err := db.Exec(adminSQL, password, password).Error; err != nil {
		fmt.Printf("⚠️  管理员创建失败: %v\n", err)
	} else {
		fmt.Printf("✅ 管理员账户创建成功 (username: admin, password: admin123)\n")
	}

	// 3. 创建充值地址
	fmt.Println("\n[3] 创建充值地址...")
	now := time.Now().Unix()

	addresses := []struct {
		Network string
		Address string
		QrCode  string
		Sort    int
	}{
		{"TRC20", "TTestTRC20Address1234567890123456789", "/uploads/qrcode/trc20.png", 100},
		{"ERC20", "0xTestERC20Address1234567890123456789012", "/uploads/qrcode/erc20.png", 90},
		{"BEP20", "0xTestBEP20Address1234567890123456789012", "/uploads/qrcode/bep20.png", 80},
	}

	for _, addr := range addresses {
		sql := `INSERT INTO deposit_address (network, address, qr_code, status, sort, create_time, update_time) 
			VALUES (?, ?, ?, 1, ?, ?, ?)
			ON DUPLICATE KEY UPDATE address=?, qr_code=?, status=1, sort=?, update_time=?`
		if err := db.Exec(sql, addr.Network, addr.Address, addr.QrCode, addr.Sort, now, now,
			addr.Address, addr.QrCode, addr.Sort, now).Error; err != nil {
			fmt.Printf("⚠️  %s地址创建失败: %v\n", addr.Network, err)
		} else {
			fmt.Printf("✅ %s充值地址创建成功: %s\n", addr.Network, addr.Address)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("✅ 测试数据初始化完成")
	fmt.Println("========================================")
	fmt.Println("\n现在可以运行测试脚本：")
	fmt.Println("  cd test_scripts/deposit && go run test_deposit.go")
}

// hashPassword 与系统使用相同的密码哈希算法
func hashPassword(password string) string {
	salt := "ABCDEFG"
	for _, char := range password {
		charHash := md5.Sum([]byte(string(char)))
		salt += hex.EncodeToString(charHash[:])
	}
	finalHash := md5.Sum([]byte(salt))
	return hex.EncodeToString(finalHash[:])
}
