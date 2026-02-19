package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type IndexInfo struct {
	TableName string `gorm:"column:Table"`
	NonUnique int    `gorm:"column:Non_unique"`
	KeyName   string `gorm:"column:Key_name"`
	SeqInIndex int `gorm:"column:Seq_in_index"`
	ColumnName string `gorm:"column:Column_name"`
}

type TableInfo struct {
	TableName string `gorm:"column:Table_name"`
	Engine    string `gorm:"column:Engine"`
}

func main() {
	// 数据库连接配置
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 1. 检查users表结构
	fmt.Println("========================================")
	fmt.Println("1. users 表结构")
	fmt.Println("========================================")
	
	var tables []TableInfo
	db.Raw("SHOW TABLE STATUS LIKE 'users'").Find(&tables)
	if len(tables) > 0 {
		fmt.Printf("表名: %s\n引擎: %s\n", tables[0].TableName, tables[0].Engine)
	}
	
	// 查看users表的索引
	var indexes []IndexInfo
	db.Raw("SHOW INDEX FROM users").Find(&indexes)
	
	fmt.Println("\n索引信息:")
	for _, idx := range indexes {
		unique := "普通索引"
		if idx.NonUnique == 0 {
			unique = "唯一索引"
		}
		fmt.Printf("  %s: %s (%s)\n", idx.KeyName, idx.ColumnName, unique)
	}
	
	// 2. 检查是否有重复账号
	fmt.Println("\n========================================")
	fmt.Println("2. 检查重复账号")
	fmt.Println("========================================")
	
	type DuplicateAccount struct {
		AccountNumber string `gorm:"column:account_number"`
		Count        int    `gorm:"column:count"`
	}
	
	var duplicates []DuplicateAccount
	db.Raw(`
		SELECT account_number, COUNT(*) as count 
		FROM users 
		GROUP BY account_number 
		HAVING count > 1
		ORDER BY count DESC
	`).Find(&duplicates)
	
	if len(duplicates) > 0 {
		fmt.Println("发现重复账号:")
		for _, dup := range duplicates {
			fmt.Printf("  账号: %s, 重复次数: %d\n", dup.AccountNumber, dup.Count)
			
			// 显示该账号的所有用户
			type UserInfo struct {
				ID           uint   `gorm:"column:id"`
				AccountNumber string `gorm:"column:account_number"`
				Phone        string `gorm:"column:phone"`
				Email        string `gorm:"column:email"`
				AreaCodeID   uint   `gorm:"column:area_code_id"`
			}
			var users []UserInfo
			db.Raw("SELECT id, account_number, phone, email, area_code_id FROM users WHERE account_number = ?", dup.AccountNumber).Find(&users)
			for _, u := range users {
				fmt.Printf("    ID=%d, Phone=%s, Email=%s, AreaCodeID=%d\n", u.ID, u.Phone, u.Email, u.AreaCodeID)
			}
		}
	} else {
		fmt.Println("✓ 未发现重复账号")
	}
	
	// 3. 检查注册代码逻辑
	fmt.Println("\n========================================")
	fmt.Println("3. 注册逻辑分析")
	fmt.Println("========================================")
	fmt.Println("问题分析:")
	fmt.Println("  1. 注册时检查: account_number + area_code_id 组合唯一")
	fmt.Println("  2. 但实际存储时: phone 和 email 都被设置为相同的值")
	fmt.Println("  3. 这可能导致同一手机号/邮箱被多次注册")
	fmt.Println()
	fmt.Println("建议修复:")
	fmt.Println("  1. 在数据库层面添加 account_number 的唯一索引")
	fmt.Println("  2. 修改注册逻辑，根据类型正确设置 phone 和 email")
	fmt.Println("  3. 添加数据库约束防止重复")
}
