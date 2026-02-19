package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
)

func main() {
	// 数据库连接配置
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8mb4&parseTime=True&loc=Local"
	
	db, err := sql.Open(dsn)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
		defer db.Close()
	
	// admin123 的MD5哈希
	hasher := md5.New()
	hasher.Write([]byte("admin123"))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	
	fmt.Printf("✅ admin123 的MD5哈希: %s\n", hashString)
	
	// 更新admin密码
	query := "UPDATE admin SET password = ? WHERE username = 'admin'"
	result, err := db.Exec(query, hashString)
	if err != nil {
		fmt.Printf("❌ 更新失败: %v\n", err)
		return
	}
	
	affected, _ := result.RowsAffected()
	fmt.Printf("✅ 成功更新 %d 行\n", affected)
	fmt.Printf("\nSQL: %s\n", query)
}