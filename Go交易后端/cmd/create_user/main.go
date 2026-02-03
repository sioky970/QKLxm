package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	dsn = "root:root123@tcp(127.0.0.1:3306)/bibi2022?charset=utf8mb4&parseTime=True&loc=Local"
)

type User struct {
	ID                           uint    `gorm:"primaryKey;column:id"`
	AreaCodeID                   uint    `gorm:"column:area_code_id"`
	AccountNumber                string  `gorm:"column:account_number;type:varchar(255);index"`
	Type                         int8    `gorm:"column:type"`
	Phone                        string  `gorm:"column:phone;type:varchar(50);index"`
	AgentID                      uint    `gorm:"column:agent_id"`
	AgentNoteID                  uint    `gorm:"column:agent_note_id"`
	ParentID                     uint    `gorm:"column:parent_id"`
	Email                        string  `gorm:"column:email;type:varchar(255);index"`
	Password                     string  `gorm:"column:password"`
	PayPassword                  string  `gorm:"column:pay_password"`
	Time                         int64   `gorm:"column:time"`
	HeadPortrait                 string  `gorm:"column:head_portrait"`
	ExtensionCode                string  `gorm:"column:extension_code"`
	Status                       int     `gorm:"column:status"`
	GesturePassword              string  `gorm:"column:gesture_password"`
	IsAuth                       string  `gorm:"column:is_auth"`
	Nickname                     string  `gorm:"column:nickname"`
	WalletAddress                string  `gorm:"column:wallet_address"`
	IsBlacklist                  int8    `gorm:"column:is_blacklist"`
	ParentsPath                  string  `gorm:"column:parents_path"`
	PushStatus                   int     `gorm:"column:push_status"`
	CandyNumber                  float64 `gorm:"column:candy_number"`
	ZhituiRealNumber             int     `gorm:"column:zhitui_real_number"`
	RealTeamnumber               uint    `gorm:"column:real_teamnumber"`
	TopUpnumber                  float64 `gorm:"column:top_upnumber"`
	IsRealname                   int     `gorm:"column:is_realname"`
	IsAtelier                    int     `gorm:"column:is_atelier"`
	NewIsrealTime                int64   `gorm:"column:new_isreal_time"`
	TodayRealTeamnumber          int     `gorm:"column:today_real_teamnumber"`
	TodayLegalDealCancelNum      int     `gorm:"column:today_LegalDealCancel_num"`
	LegalDealCancelNumUpdateTime int64   `gorm:"column:LegalDealCancel_num__update_time"`
	Risk                         int8    `gorm:"column:risk"`
	LockTime                     int64   `gorm:"column:lock_time"`
	Level                        int     `gorm:"column:level"`
	Fund                         float64 `gorm:"column:fund"`
	IsService                    uint8   `gorm:"column:is_service"`
	AgentPath                    string  `gorm:"column:agent_path"`
	StoreID                      uint    `gorm:"column:store_id"`
	LastTime                     int64   `gorm:"column:last_time"`
	LastLoginIP                  string  `gorm:"column:last_login_ip"`
	UserRemark                   string  `gorm:"column:user_remark"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库连接成功")

	// 检查用户是否已存在
	var existUser User
	result := db.Where("account_number = ?", "13333333333").First(&existUser)
	if result.RowsAffected > 0 {
		fmt.Printf("⚠️  用户已存在: %s\n", "13333333333")
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("333333"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ 密码加密失败: %v\n", err)
		return
	}

	// 创建用户
	user := User{
		AreaCodeID:      1,
		AccountNumber:   "13333333333",
		Type:            0,
		Phone:           "13333333333",
		Email:           "13333333333",
		Password:        string(hashedPassword),
		Time:            time.Now().Unix(),
		HeadPortrait:    "/mobile/images/user_head.png",
		ExtensionCode:   generateExtensionCode(),
		Status:          0,
		IsAuth:          "0",
		Nickname:        "用户" + generateNickname(),
		WalletAddress:   "",
		IsBlacklist:     0,
		ParentsPath:     "",
		PushStatus:      1,
		CandyNumber:     0,
		ZhituiRealNumber: 0,
		RealTeamnumber:  0,
		TopUpnumber:     0,
		IsRealname:      0,
		IsAtelier:       0,
		NewIsrealTime:   0,
		TodayRealTeamnumber: 0,
		TodayLegalDealCancelNum: 0,
		LegalDealCancelNumUpdateTime: 0,
		Risk:           0,
		LockTime:       0,
		Level:          0,
		Fund:           0,
		IsService:      0,
		AgentPath:      "",
		LastTime:       0,
		LastLoginIP:    "",
		UserRemark:     "",
	}

	// 事务创建用户和钱包
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Table("users").Create(&user).Error; err != nil {
		tx.Rollback()
		fmt.Printf("❌ 创建用户失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 用户创建成功: ID=%d, Account=%s\n", user.ID, user.AccountNumber)

	// 创建钱包
	if err := createUserWallets(tx, user.ID); err != nil {
		tx.Rollback()
		fmt.Printf("❌ 创建钱包失败: %v\n", err)
		return
	}

	tx.Commit()
	fmt.Println("✅ 钱包创建成功")

	// 验证
	var checkUser User
	db.Where("account_number = ?", "13333333333").First(&checkUser)
	fmt.Printf("\n✅ 注册完成！\n")
	fmt.Printf("   用户名: 13333333333\n")
	fmt.Printf("   密码: 333333\n")
	fmt.Printf("   用户ID: %d\n", checkUser.ID)
}

func generateExtensionCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateNickname() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)[:6]
}

func createUserWallets(tx *gorm.DB, userID uint) error {
	now := time.Now().Unix()

	// 创建现货钱包
	spotWallet := fmt.Sprintf(`
		INSERT INTO user_wallets (user_id, wallet_type, currency_id, currency_name, available_balance, locked_balance, version, last_trade_time, create_time, update_time)
		VALUES (%d, 'spot', 1, 'USDT', 0, 0, 0, 0, %d, %d)
	`, userID, now, now)
	if err := tx.Exec(spotWallet).Error; err != nil {
		return fmt.Errorf("创建现货钱包失败: %w", err)
	}

	// 创建合约钱包
	contractWallet := fmt.Sprintf(`
		INSERT INTO user_wallets (user_id, wallet_type, currency_id, currency_name, available_balance, locked_balance, version, last_trade_time, create_time, update_time)
		VALUES (%d, 'contract', 1, 'USDT', 0, 0, 0, 0, %d, %d)
	`, userID, now, now)
	if err := tx.Exec(contractWallet).Error; err != nil {
		return fmt.Errorf("创建合约钱包失败: %w", err)
	}

	// 创建交割钱包
	deliveryWallet := fmt.Sprintf(`
		INSERT INTO user_wallets (user_id, wallet_type, currency_id, currency_name, available_balance, locked_balance, version, last_trade_time, create_time, update_time)
		VALUES (%d, 'delivery', 1, 'USDT', 0, 0, 0, 0, %d, %d)
	`, userID, now, now)
	if err := tx.Exec(deliveryWallet).Error; err != nil {
		return fmt.Errorf("创建交割钱包失败: %w", err)
	}

	return nil
}
