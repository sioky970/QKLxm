package main

import (
	"crypto/md5"
	"encoding/hex"
	"exchange-go/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"fmt"
	"log"
)

// HashPassword 与PHP的Users::MakePassword($password, 0)保持一致
func HashPassword(password string) string {
	salt := "ABCDEFG"
	
	for _, char := range password {
		charHash := md5.Sum([]byte(string(char)))
		salt += hex.EncodeToString(charHash[:])
	}
	
	finalHash := md5.Sum([]byte(salt))
	return hex.EncodeToString(finalHash[:])
}

func main() {
	// 加载配置
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 初始化日志 (database.Init需要)
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()
	
	// 计算admin的正确密码哈希
	password := "admin"
	hashedPassword := HashPassword(password)
	
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("加密后: %s\n", hashedPassword)
	
	// 更新数据库
	result := database.DB.Exec("UPDATE admin SET password = ? WHERE username = ?", hashedPassword, "admin")
	if result.Error != nil {
		log.Fatalf("更新密码失败: %v", result.Error)
	}
	
	fmt.Printf("成功更新 %d 条记录\n", result.RowsAffected)
	fmt.Println("\n现在可以使用 admin/admin 登录了！")
}
