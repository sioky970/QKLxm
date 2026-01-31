package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/logger"

	"github.com/go-redis/redis/v8"
)

const (
	// 用户资产缓存相关
	UserAssetsCachePrefix     = "user:assets:"      // 用户资产前缀
	UserAssetsCacheExpire     = 30 * time.Second    // 用户资产缓存时间（短时间，因为涉及实时交易）

	// 用户余额缓存相关
	UserBalanceCachePrefix    = "user:balance:"     // 用户余额前缀
	UserBalanceCacheExpire    = 10 * time.Second    // 用户余额缓存时间

	// 交易对配置缓存
	CurrencyMatchCachePrefix  = "currency:match:"   // 交易对配置前缀
	CurrencyMatchCacheExpire  = 5 * time.Minute     // 交易对配置缓存时间

	// 币种配置缓存
	CurrencyCachePrefix       = "currency:"         // 币种配置前缀
	CurrencyCacheExpire       = 10 * time.Minute    // 币种配置缓存时间

	// KYC状态缓存
	KYCStatusCachePrefix      = "kyc:status:"       // KYC状态前缀
	KYCStatusCacheExpire      = 5 * time.Minute     // KYC状态缓存时间

	// 统计数据缓存
	StatisticsCachePrefix     = "statistics:"       // 统计数据前缀
	StatisticsCacheExpire     = 1 * time.Minute     // 统计数据缓存时间（需要实时性）

	// 行情数据缓存
	MarketPriceCachePrefix    = "market:price:"     // 行情价格前缀
	MarketPriceCacheExpire    = 5 * time.Second     // 行情价格缓存时间

	// 用户会话缓存
	UserSessionCachePrefix    = "session:"          // 用户会话前缀
	UserSessionCacheExpire    = 24 * time.Hour      // 用户会话缓存时间

	// 限流缓存
	RateLimitCachePrefix      = "ratelimit:"        // 限流前缀
	RateLimitWindow           = 1 * time.Minute     // 限流时间窗口

	// 排行榜缓存
	LeaderboardCachePrefix    = "leaderboard:"      // 排行榜前缀
	LeaderboardCacheExpire    = 5 * time.Minute     // 排行榜缓存时间

	// 订单统计缓存
	OrderStatsCachePrefix     = "order:stats:"      // 订单统计前缀
	OrderStatsCacheExpire     = 1 * time.Minute     // 订单统计缓存时间
)

// CacheConfig 缓存配置
type CacheConfig struct {
	EnableUserAssetsCache bool // 是否启用用户资产缓存
	EnableBalanceCache    bool // 是否启用余额缓存
	EnableMatchCache      bool // 是否启用交易对配置缓存
	EnableMarketCache     bool // 是否启用行情缓存
	EnableStatisticsCache bool // 是否启用统计缓存
}

// InitWithConfig 初始化Redis连接（带配置）
func InitWithConfig(cfg *config.RedisConfig, cacheCfg *CacheConfig) error {
	Init(cfg)
	return nil
}

// ============================================================
// 用户资产缓存操作
// ============================================================

// GetUserAssets 获取用户资产缓存
func GetUserAssets(userID uint) (*model.UserAssets, error) {
	key := fmt.Sprintf("%s%d", UserAssetsCachePrefix, userID)
	var assets model.UserAssets
	err := Get(key, &assets)
	if err == nil {
		logger.Debug(fmt.Sprintf("缓存命中: 用户资产 %d", userID))
		return &assets, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetUserAssets 设置用户资产缓存
func SetUserAssets(userID uint, assets *model.UserAssets) error {
	if assets == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d", UserAssetsCachePrefix, userID)
	return Set(key, assets, UserAssetsCacheExpire)
}

// DeleteUserAssets 删除用户资产缓存
func DeleteUserAssets(userID uint) error {
	key := fmt.Sprintf("%s%d", UserAssetsCachePrefix, userID)
	return Delete(key)
}

// InvalidateUserAssetsBatch 批量失效用户资产缓存
func InvalidateUserAssetsBatch(userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = fmt.Sprintf("%s%d", UserAssetsCachePrefix, userID)
	}
	return Delete(keys...)
}

// ============================================================
// 用户余额缓存操作
// ============================================================

// UserBalance 用户余额结构
type UserBalance struct {
	UserID          uint    `json:"user_id"`
	UsdtBalance     float64 `json:"usdt_balance"`
	UsdtLocked      float64 `json:"usdt_locked"`
	CurrencyBalances string  `json:"currency_balances"`
	CurrencyLocked  string  `json:"currency_locked"`
	TotalValueUsdt  float64 `json:"total_value_usdt"`
	UpdateTime      int64   `json:"update_time"`
}

// GetUserBalance 获取用户余额缓存
func GetUserBalance(userID uint) (*UserBalance, error) {
	key := fmt.Sprintf("%s%d", UserBalanceCachePrefix, userID)
	var balance UserBalance
	err := Get(key, &balance)
	if err == nil {
		return &balance, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetUserBalance 设置用户余额缓存
func SetUserBalance(userID uint, balance *UserBalance) error {
	if balance == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d", UserBalanceCachePrefix, userID)
	return Set(key, balance, UserBalanceCacheExpire)
}

// DeleteUserBalance 删除用户余额缓存
func DeleteUserBalance(userID uint) error {
	key := fmt.Sprintf("%s%d", UserBalanceCachePrefix, userID)
	return Delete(key)
}

// ============================================================
// 交易对配置缓存操作
// ============================================================

// GetCurrencyMatch 获取交易对配置缓存
func GetCurrencyMatch(legalID, currencyID uint) (*model.CurrencyMatch, error) {
	key := fmt.Sprintf("%s%d:%d", CurrencyMatchCachePrefix, legalID, currencyID)
	var match model.CurrencyMatch
	err := Get(key, &match)
	if err == nil {
		return &match, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetCurrencyMatch 设置交易对配置缓存
func SetCurrencyMatch(legalID, currencyID uint, match *model.CurrencyMatch) error {
	if match == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d:%d", CurrencyMatchCachePrefix, legalID, currencyID)
	return Set(key, match, CurrencyMatchCacheExpire)
}

// DeleteCurrencyMatch 删除交易对配置缓存
func DeleteCurrencyMatch(legalID, currencyID uint) error {
	key := fmt.Sprintf("%s%d:%d", CurrencyMatchCachePrefix, legalID, currencyID)
	return Delete(key)
}

// InvalidateAllMatches 失效所有交易对配置缓存
func InvalidateAllMatches() error {
	pattern := fmt.Sprintf("%s*", CurrencyMatchCachePrefix)
	keys, err := RDB.Keys(Ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return Delete(keys...)
	}
	return nil
}

// ============================================================
// 币种配置缓存操作
// ============================================================

// GetCurrency 获取币种配置缓存
func GetCurrency(currencyID uint) (*model.Currency, error) {
	key := fmt.Sprintf("%s%d", CurrencyCachePrefix, currencyID)
	var currency model.Currency
	err := Get(key, &currency)
	if err == nil {
		return &currency, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetCurrency 设置币种配置缓存
func SetCurrency(currencyID uint, currency *model.Currency) error {
	if currency == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d", CurrencyCachePrefix, currencyID)
	return Set(key, currency, CurrencyCacheExpire)
}

// DeleteCurrency 删除币种配置缓存
func DeleteCurrency(currencyID uint) error {
	key := fmt.Sprintf("%s%d", CurrencyCachePrefix, currencyID)
	return Delete(key)
}

// ============================================================
// KYC状态缓存操作
// ============================================================

// GetKYCStatus 获取用户KYC状态缓存
func GetKYCStatus(userID uint) (int, error) {
	key := fmt.Sprintf("%s%d", KYCStatusCachePrefix, userID)
	result, err := RDB.Get(Ctx, key).Int()
	if err == nil {
		return result, nil
	}
	if err == redis.Nil {
		return -1, nil // -1表示未认证
	}
	return -1, err
}

// SetKYCStatus 设置用户KYC状态缓存
func SetKYCStatus(userID uint, status int) error {
	key := fmt.Sprintf("%s%d", KYCStatusCachePrefix, userID)
	return RDB.Set(Ctx, key, status, KYCStatusCacheExpire).Err()
}

// DeleteKYCStatus 删除用户KYC状态缓存
func DeleteKYCStatus(userID uint) error {
	key := fmt.Sprintf("%s%d", KYCStatusCachePrefix, userID)
	return Delete(key)
}

// ============================================================
// 行情数据缓存操作
// ============================================================

// MarketPrice 行情价格结构
type MarketPrice struct {
	Symbol       string  `json:"symbol"`
	Price        float64 `json:"price"`
	Change24h    float64 `json:"change_24h"`
	Volume24h    float64 `json:"volume_24h"`
	High24h      float64 `json:"high_24h"`
	Low24h       float64 `json:"low_24h"`
	UpdateTime   int64   `json:"update_time"`
}

// GetMarketPrice 获取行情价格缓存
func GetMarketPrice(symbol string) (*MarketPrice, error) {
	key := fmt.Sprintf("%s%s", MarketPriceCachePrefix, symbol)
	var price MarketPrice
	err := Get(key, &price)
	if err == nil {
		return &price, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetMarketPrice 设置行情价格缓存
func SetMarketPrice(symbol string, price *MarketPrice) error {
	if price == nil {
		return nil
	}
	key := fmt.Sprintf("%s%s", MarketPriceCachePrefix, symbol)
	return Set(key, price, MarketPriceCacheExpire)
}

// SetMarketPriceBatch 批量设置行情价格
func SetMarketPriceBatch(prices map[string]*MarketPrice) error {
	if len(prices) == 0 {
		return nil
	}
	pipe := RDB.Pipeline()
	for symbol, price := range prices {
		key := fmt.Sprintf("%s%s", MarketPriceCachePrefix, symbol)
		data, _ := json.Marshal(price)
		pipe.Set(Ctx, key, data, MarketPriceCacheExpire)
	}
	_, err := pipe.Exec(Ctx)
	return err
}

// ============================================================
// 统计数据缓存操作
// ============================================================

// StatisticsData 统计数据结构
type StatisticsData struct {
	TotalUsers       int64   `json:"total_users"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalVolume      float64 `json:"total_volume"`
	TotalMarketCap   float64 `json:"total_market_cap"`
	UpdateTime       int64   `json:"update_time"`
}

// GetStatistics 获取统计数据缓存
func GetStatistics(statType string) (*StatisticsData, error) {
	key := fmt.Sprintf("%s%s", StatisticsCachePrefix, statType)
	var stats StatisticsData
	err := Get(key, &stats)
	if err == nil {
		return &stats, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetStatistics 设置统计数据缓存
func SetStatistics(statType string, stats *StatisticsData) error {
	if stats == nil {
		return nil
	}
	key := fmt.Sprintf("%s%s", StatisticsCachePrefix, statType)
	return Set(key, stats, StatisticsCacheExpire)
}

// ============================================================
// 用户会话缓存操作
// ============================================================

// UserSession 用户会话结构
type UserSession struct {
	UserID        uint   `json:"user_id"`
	SessionToken  string `json:"session_token"`
	LoginTime     int64  `json:"login_time"`
	ExpireTime    int64  `json:"expire_time"`
	ClientIP      string `json:"client_ip"`
	ClientType    string `json:"client_type"`
}

// SetUserSession 设置用户会话
func SetUserSession(userID uint, session *UserSession) error {
	key := fmt.Sprintf("%s%d", UserSessionCachePrefix, userID)
	return Set(key, session, UserSessionCacheExpire)
}

// GetUserSession 获取用户会话
func GetUserSession(userID uint) (*UserSession, error) {
	key := fmt.Sprintf("%s%d", UserSessionCachePrefix, userID)
	var session UserSession
	err := Get(key, &session)
	if err == nil {
		return &session, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// DeleteUserSession 删除用户会话
func DeleteUserSession(userID uint) error {
	key := fmt.Sprintf("%s%d", UserSessionCachePrefix, userID)
	return Delete(key)
}

// ValidateUserSession 验证用户会话是否有效
func ValidateUserSession(userID uint, sessionToken string) (bool, error) {
	session, err := GetUserSession(userID)
	if err != nil {
		return false, err
	}
	if session == nil {
		return false, nil
	}
	if session.SessionToken != sessionToken {
		return false, nil
	}
	return true, nil
}

// ============================================================
// 限流操作
// ============================================================

// RateLimitInfo 限流信息结构
type RateLimitInfo struct {
	Count     int64 `json:"count"`
	Limit     int64 `json:"limit"`
	WindowSec int64 `json:"window_sec"`
	ResetTime int64 `json:"reset_time"`
}

// CheckRateLimit 检查是否超过限流阈值
func CheckRateLimit(identifier string, limit int64) (*RateLimitInfo, error) {
	key := fmt.Sprintf("%s%s", RateLimitCachePrefix, identifier)
	
	pipe := RDB.Pipeline()
	incr := pipe.Incr(Ctx, key)
	pipe.Expire(Ctx, key, RateLimitWindow)
	_, _ = pipe.Exec(Ctx)
	
	count := incr.Val()
	
	info := &RateLimitInfo{
		Count:     count,
		Limit:     limit,
		WindowSec: int64(RateLimitWindow.Seconds()),
		ResetTime: time.Now().Add(RateLimitWindow).Unix(),
	}
	
	return info, nil
}

// IsRateLimited 检查是否被限流
func IsRateLimited(identifier string, limit int64) (bool, error) {
	info, err := CheckRateLimit(identifier, limit)
	if err != nil {
		return false, err
	}
	return info.Count > info.Limit, nil
}

// ============================================================
// 排行榜操作
// ============================================================

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	UserID    uint    `json:"user_id"`
	Username  string  `json:"username"`
	Score     float64 `json:"score"`
	Rank      int     `json:"rank"`
}

// SetLeaderboard 设置排行榜（使用Sorted Set）
func SetLeaderboard(leaderboardType string, entries []LeaderboardEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	key := fmt.Sprintf("%s%s", LeaderboardCachePrefix, leaderboardType)
	pipe := RDB.Pipeline()
	
	for _, entry := range entries {
		member := fmt.Sprintf("%d:%s", entry.UserID, entry.Username)
		pipe.ZAdd(Ctx, key, &redis.Z{
			Score:  entry.Score,
			Member: member,
		})
	}
	
	pipe.Expire(Ctx, key, LeaderboardCacheExpire)
	_, err := pipe.Exec(Ctx)
	return err
}

// GetLeaderboard 获取排行榜
func GetLeaderboard(leaderboardType string, start, stop int64) ([]LeaderboardEntry, error) {
	key := fmt.Sprintf("%s%s", LeaderboardCachePrefix, leaderboardType)
	
	results, err := RDB.ZRevRangeWithScores(Ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	
	entries := make([]LeaderboardEntry, 0, len(results))
	for rank, z := range results {
		member := z.Member.(string)
		// 解析 user_id:username 格式
		var entry LeaderboardEntry
		fmt.Sscanf(member, "%d:%s", &entry.UserID, &entry.Username)
		entry.Score = z.Score
		entry.Rank = int(rank) + int(start) + 1
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// ============================================================
// 订单统计缓存操作
// ============================================================

// OrderStats 订单统计数据
type OrderStats struct {
	UserID            uint   `json:"user_id"`
	TotalOrders       int64  `json:"total_orders"`
	PendingOrders     int64  `json:"pending_orders"`
	CompletedOrders   int64  `json:"completed_orders"`
	CancelledOrders   int64  `json:"cancelled_orders"`
	TotalVolume       float64 `json:"total_volume"`
	UpdateTime        int64  `json:"update_time"`
}

// GetOrderStats 获取用户订单统计缓存
func GetOrderStats(userID uint) (*OrderStats, error) {
	key := fmt.Sprintf("%s%d", OrderStatsCachePrefix, userID)
	var stats OrderStats
	err := Get(key, &stats)
	if err == nil {
		return &stats, nil
	}
	if err == redis.Nil {
		return nil, nil
	}
	return nil, err
}

// SetOrderStats 设置用户订单统计缓存
func SetOrderStats(userID uint, stats *OrderStats) error {
	if stats == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d", OrderStatsCachePrefix, userID)
	return Set(key, stats, OrderStatsCacheExpire)
}

// DeleteOrderStats 删除用户订单统计缓存
func DeleteOrderStats(userID uint) error {
	key := fmt.Sprintf("%s%d", OrderStatsCachePrefix, userID)
	return Delete(key)
}

// ============================================================
// 缓存失效策略
// ============================================================

// InvalidateAll 失效所有缓存（谨慎使用）
func InvalidateAll() error {
	patterns := []string{
		UserAssetsCachePrefix + "*",
		UserBalanceCachePrefix + "*",
		CurrencyMatchCachePrefix + "*",
		CurrencyCachePrefix + "*",
		KYCStatusCachePrefix + "*",
		StatisticsCachePrefix + "*",
		MarketPriceCachePrefix + "*",
		UserSessionCachePrefix + "*",
		OrderStatsCachePrefix + "*",
	}
	
	for _, pattern := range patterns {
		keys, err := RDB.Keys(Ctx, pattern).Result()
		if err != nil {
			logger.Warn(fmt.Sprintf("获取缓存键失败: %s", pattern))
			continue
		}
		if len(keys) > 0 {
			if err := Delete(keys...); err != nil {
				logger.Warn(fmt.Sprintf("删除缓存失败: %s", pattern))
			}
		}
	}
	
	return nil
}

// CacheStats 缓存统计信息
type CacheStats struct {
	KeysCount     int64  `json:"keys_count"`
	MemoryUsage   string `json:"memory_usage"`
	Connected     bool   `json:"connected"`
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() (*CacheStats, error) {
	stats := &CacheStats{
		Connected: true,
	}
	
	// 获取键总数
	count, err := RDB.DBSize(Ctx).Result()
	if err != nil {
		return nil, err
	}
	stats.KeysCount = count
	
	// 获取内存使用信息
	info, err := RDB.Info(Ctx, "memory").Result()
	if err == nil {
		// 解析内存信息
		logger.Debug(fmt.Sprintf("Redis内存信息: %s", info))
	}
	
	return stats, nil
}
