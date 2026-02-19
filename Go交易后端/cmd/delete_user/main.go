package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserAssets struct {
	ID uint `gorm:"column:id"`
}

type User struct {
	ID uint `gorm:"column:id"`
}

func main() {
	// 数据库连接配置
	dsn := "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 删除用户ID: 4（余额较少的用户）
	userID := uint(4)
	
	// 开始事务
	tx := db.Begin()
	
	// 删除用户资产
	if err := tx.Table("user_assets").Where("user_id = ?", userID).Delete(&UserAssets{}).Error; err != nil {
		tx.Rollback()
		log.Fatalf("删除用户资产失败: %v", err)
	}
	
	// 删除用户
	if err := tx.Table("users").Delete(&User{}, userID).Error; err != nil {
		tx.Rollback()
		log.Fatalf("删除用户失败: %v", err)
	}
	
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}
	
	fmt.Printf("成功删除用户ID: %d 及其相关数据\n", userID)
}
