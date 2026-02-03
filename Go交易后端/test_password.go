package main
import (
"crypto/md5"
"encoding/hex"
"fmt"
)

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
fmt.Println("admin ->", HashPassword("admin"))
}
