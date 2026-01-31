package service

import (
	"testing"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) {
	dsn := "root:@tcp(127.0.0.1:3306)/bibi2022_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("连接测试数据库失败: %v", err)
	}
	database.DB = db

	// 自动迁移测试表
	db.AutoMigrate(&model.UserAssets{}, &model.User{})
}

// cleanupTestDB 清理测试数据
func cleanupTestDB(t *testing.T) {
	database.DB.Exec("TRUNCATE TABLE user_assets")
}

func TestUserAssetsService_CreateUserAssets(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9999)

	// 测试创建用户资产
	assets, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)
	assert.NotNil(t, assets)
	assert.Equal(t, userID, assets.UserID)
	assert.Equal(t, float64(0), assets.UsdtBalance)
	assert.Equal(t, float64(0), assets.UsdtLocked)
	assert.NotNil(t, assets.CurrencyBalances)
	assert.NotNil(t, assets.CurrencyLocked)
}

func TestUserAssetsService_GetUserAssets(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9998)

	// 创建测试数据
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)

	// 测试获取用户资产
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.NotNil(t, assets)
	assert.Equal(t, userID, assets.UserID)
}

func TestUserAssetsService_UpdateUsdtBalance(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9997)

	// 创建测试数据
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)

	// 测试增加USDT余额
	err = service.UpdateUsdtBalance(userID, 1000.0, false)
	assert.NoError(t, err)

	// 验证余额
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(1000.0), assets.UsdtBalance)

	// 测试扣除USDT余额
	err = service.UpdateUsdtBalance(userID, -500.0, false)
	assert.NoError(t, err)

	// 验证余额
	assets, err = service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(500.0), assets.UsdtBalance)

	// 测试余额不足
	err = service.UpdateUsdtBalance(userID, -1000.0, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestUserAssetsService_LockUsdt(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9996)

	// 创建测试数据并充值
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)
	err = service.UpdateUsdtBalance(userID, 1000.0, false)
	assert.NoError(t, err)

	// 测试锁定USDT
	err = service.LockUsdt(userID, 300.0)
	assert.NoError(t, err)

	// 验证余额和锁定余额
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(700.0), assets.UsdtBalance)
	assert.Equal(t, float64(300.0), assets.UsdtLocked)

	// 测试锁定余额不足
	err = service.LockUsdt(userID, 1000.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestUserAssetsService_UnlockUsdt(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9995)

	// 创建测试数据并充值、锁定
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)
	err = service.UpdateUsdtBalance(userID, 1000.0, false)
	assert.NoError(t, err)
	err = service.LockUsdt(userID, 300.0)
	assert.NoError(t, err)

	// 测试解锁USDT
	err = service.UnlockUsdt(userID, 100.0)
	assert.NoError(t, err)

	// 验证余额和锁定余额
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(800.0), assets.UsdtBalance)
	assert.Equal(t, float64(200.0), assets.UsdtLocked)

	// 测试解锁余额不足
	err = service.UnlockUsdt(userID, 300.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestUserAssetsService_UpdateCurrencyBalance(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9994)

	// 创建测试数据
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)

	// 测试增加BTC余额
	err = service.UpdateCurrencyBalance(userID, "BTC", 0.5, false)
	assert.NoError(t, err)

	// 验证余额
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(0.5), assets.GetCurrencyBalance("BTC"))

	// 测试增加ETH余额
	err = service.UpdateCurrencyBalance(userID, "ETH", 2.0, false)
	assert.NoError(t, err)

	// 验证余额
	assets, err = service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(2.0), assets.GetCurrencyBalance("ETH"))

	// 测试扣除BTC余额
	err = service.UpdateCurrencyBalance(userID, "BTC", -0.2, false)
	assert.NoError(t, err)

	// 验证余额
	assets, err = service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(0.3), assets.GetCurrencyBalance("BTC"))
}

func TestUserAssetsService_LockCurrency(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9993)

	// 创建测试数据并增加币种余额
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)
	err = service.UpdateCurrencyBalance(userID, "BTC", 1.0, false)
	assert.NoError(t, err)

	// 测试锁定BTC
	err = service.LockCurrency(userID, "BTC", 0.3)
	assert.NoError(t, err)

	// 验证余额和锁定余额
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(0.7), assets.GetCurrencyBalance("BTC"))
	assert.Equal(t, float64(0.3), assets.GetCurrencyLocked("BTC"))

	// 测试锁定余额不足
	err = service.LockCurrency(userID, "BTC", 1.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestUserAssetsService_UnlockCurrency(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9992)

	// 创建测试数据并增加币种余额、锁定
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)
	err = service.UpdateCurrencyBalance(userID, "ETH", 5.0, false)
	assert.NoError(t, err)
	err = service.LockCurrency(userID, "ETH", 2.0)
	assert.NoError(t, err)

	// 测试解锁ETH
	err = service.UnlockCurrency(userID, "ETH", 1.0)
	assert.NoError(t, err)

	// 验证余额和锁定余额
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(4.0), assets.GetCurrencyBalance("ETH"))
	assert.Equal(t, float64(1.0), assets.GetCurrencyLocked("ETH"))
}

func TestUserAssetsService_ConcurrentUpdate(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9991)

	// 创建测试数据并充值
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)
	err = service.UpdateUsdtBalance(userID, 10000.0, false)
	assert.NoError(t, err)

	// 测试并发更新(乐观锁)
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			err := service.UpdateUsdtBalance(userID, -100.0, false)
			// 可能部分失败(并发冲突),但不应该数据错误
			done <- (err == nil)
		}()
	}

	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-done {
			successCount++
		}
	}

	// 验证最终余额
	time.Sleep(100 * time.Millisecond) // 等待所有事务完成
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)

	// 余额应该是: 10000 - (成功次数 * 100)
	expectedBalance := 10000.0 - float64(successCount)*100.0
	assert.Equal(t, expectedBalance, assets.UsdtBalance)

	t.Logf("并发测试: 成功=%d/%d, 最终余额=%.2f", successCount, concurrency, assets.UsdtBalance)
}

func TestUserAssetsService_JSONSerialization(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9990)

	// 创建测试数据
	_, err := service.CreateUserAssets(userID)
	assert.NoError(t, err)

	// 添加多个币种余额
	currencies := map[string]float64{
		"BTC":  1.23456789,
		"ETH":  10.5,
		"USDT": 1000.88888888,
		"BNB":  5.0,
		"ADA":  100.0,
	}

	for currency, balance := range currencies {
		err = service.UpdateCurrencyBalance(userID, currency, balance, false)
		assert.NoError(t, err)
	}

	// 验证JSON序列化/反序列化
	assets, err := service.GetUserAssets(userID)
	assert.NoError(t, err)

	for currency, expectedBalance := range currencies {
		actualBalance := assets.GetCurrencyBalance(currency)
		assert.Equal(t, expectedBalance, actualBalance, "币种 %s 余额不匹配", currency)
	}
}

func TestUserAssetsService_GetOrCreate(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	service := GetUserAssetsService()
	userID := uint(9989)

	// 测试首次获取(自动创建)
	assets1, err := service.GetOrCreateUserAssets(userID)
	assert.NoError(t, err)
	assert.NotNil(t, assets1)
	assert.Equal(t, userID, assets1.UserID)

	// 测试再次获取(不重复创建)
	assets2, err := service.GetOrCreateUserAssets(userID)
	assert.NoError(t, err)
	assert.NotNil(t, assets2)
	assert.Equal(t, assets1.ID, assets2.ID)
}
