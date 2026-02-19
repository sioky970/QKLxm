package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func main() {
	password := "admin123"
	
	// 使用Go标准MD5加密
	hasher := md5.New()
	hasher.Write([]byte(password))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	
	fmt.Printf("密码: %s\n", password)
	fmt.Printf("MD5哈希: %s\n", hashString)
	fmt.Printf("\nSQL更新语句:\n")
	fmt.Printf("UPDATE admin SET password = '%s' WHERE username = 'admin';\n", hashString)
}
