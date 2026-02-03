package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func main() {
	password := "admin123"
	hash := hashPassword(password)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", hash)
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
