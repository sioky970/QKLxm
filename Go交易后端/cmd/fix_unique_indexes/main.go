package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接配置
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("修复数据库唯一索引")
	fmt.Println("========================================")

	// 1. 删除旧的普通索引
	fmt.Println("\n1. 删除旧索引...")
	
	indexes := []string{
		"idx_users_account_number",
		"idx_users_phone",
		"idx_users_email",
	}
	
	for _, idx := range indexes {
		if err := db.Exec(fmt.Sprintf("DROP INDEX %s ON users", idx)).Error; err != nil {
			fmt.Printf("  ⚠️  删除索引 %s 失败: %v (可能不存在)\n", idx, err)
		} else {
			fmt.Printf("  ✓ 删除索引 %s 成功\n", idx)
		}
	}

	// 2. 创建唯一索引
	fmt.Println("\n2. 创建唯一索引...")
	
	uniqueIndexes := []struct {
		name   string
		column string
	}{
		{"idx_users_account_number", "account_number"},
		{"idx_users_phone", "phone"},
		{"idx_users_email", "email"},
	}
	
	for _, idx := range uniqueIndexes {
		sql := fmt.Sprintf("CREATE UNIQUE INDEX %s ON users(%s)", idx.name, idx.column)
		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("  ❌ 创建唯一索引 %s 失败: %v\n", idx.name, err)
			fmt.Printf("     SQL: %s\n", sql)
		} else {
			fmt.Printf("  ✓ 创建唯一索引 %s 成功\n", idx.name)
		}
	}

	// 3. 验证索引
	fmt.Println("\n3. 验证索引...")
	
	type IndexInfo struct {
		KeyName   string `gorm:"column:Key_name"`
		NonUnique int    `gorm:"column:Non_unique"`
	}
	
	var indexInfos []IndexInfo
	db.Raw("SHOW INDEX FROM users").Find(&indexInfos)
	
	fmt.Println("\n当前索引状态:")
	for _, idx := range indexInfos {
		if idx.KeyName == "PRIMARY" {
			continue
		}
		status := "❌ 普通索引"
		if idx.NonUnique == 0 {
			status = "✓ 唯一索引"
		}
		fmt.Printf("  %s: %s\n", idx.KeyName, status)
	}

	fmt.Println("\n========================================")
	fmt.Println("修复完成！")
	fmt.Println("========================================")
}
