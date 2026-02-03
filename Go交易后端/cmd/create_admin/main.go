package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	dsn = "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8mb4&parseTime=True&loc=Local"
)

type Admin struct {
	ID       uint   `gorm:"primaryKey;column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
	RoleID   uint   `gorm:"column:role_id"`
}

type AdminRole struct {
	ID      uint   `gorm:"primaryKey;column:id"`
	Name    string `gorm:"column:name"`
	IsSuper int    `gorm:"column:is_super"`
}

func (Admin) TableName() string {
	return "admin"
}

func (AdminRole) TableName() string {
	return "admin_role"
}

func main() {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库连接成功")

	// 创建管理员角色
	fmt.Println("\n[1] 创建管理员角色...")
	role := AdminRole{ID: 1, Name: "超级管理员", IsSuper: 1}
	if err := db.Table("admin_role").Save(&role).Error; err != nil {
		fmt.Printf("⚠️  角色创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 管理员角色创建成功")
	}

	// 创建管理员账户
	fmt.Println("\n[2] 创建管理员账户...")
	password := hashPassword("admin123")
	admin := Admin{ID: 1, Username: "admin", Password: password, RoleID: 1}
	if err := db.Table("admin").Save(&admin).Error; err != nil {
		fmt.Printf("⚠️  管理员创建失败: %v\n", err)
	} else {
		fmt.Printf("✅ 管理员账户创建成功\n")
		fmt.Printf("   用户名: admin\n")
		fmt.Printf("   密码: admin123\n")
		fmt.Printf("   密码哈希: %s\n", password)
	}

	// 验证
	var count int64
	db.Table("admin").Where("username = ?", "admin").Count(&count)
	fmt.Printf("\n数据库中管理员数量: %d\n", count)
}

func hashPassword(password string) string {
	salt := "ABCDEFG"
	for _, char := range password {
		charHash := md5.Sum([]byte(string(char)))
		salt += hex.EncodeToString(charHash[:])
	}
	finalHash := md5.Sum([]byte(salt))
	return hex.EncodeToString(finalHash[:])
}
