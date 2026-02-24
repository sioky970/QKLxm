package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"exchange-go/config"
	"exchange-go/internal/pkg/logger"

	"github.com/go-redis/redis/v8"
)

// RDB 全局Redis客户端
var RDB *redis.Client

// Ctx 全局上下文
var Ctx = context.Background()

// Init 初始化Redis连接
func Init(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.PoolSize / 4, // 最小空闲连接数
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读超时
		WriteTimeout: 3 * time.Second,  // 写超时
		PoolTimeout:  4 * time.Second,  // 获取连接超时
		IdleTimeout:  5 * time.Minute,  // 空闲连接超时
		MaxConnAge:   0,                // 连接存活时间（0表示不限制）
	})

	// 测试连接
	if _, err := RDB.Ping(Ctx).Result(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	logger.Info("Redis连接成功")
	return nil
}

// Close 关闭Redis连接
func Close() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}

// Set 设置缓存
func Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.Set(Ctx, key, data, expiration).Err()
}

// Get 获取缓存
func Get(key string, dest interface{}) error {
	data, err := RDB.Get(Ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// GetString 获取字符串缓存
func GetString(key string) (string, error) {
	return RDB.Get(Ctx, key).Result()
}

// SetString 设置字符串缓存
func SetString(key string, value string, expiration time.Duration) error {
	return RDB.Set(Ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func Delete(keys ...string) error {
	return RDB.Del(Ctx, keys...).Err()
}

// Exists 检查key是否存在
func Exists(key string) (bool, error) {
	n, err := RDB.Exists(Ctx, key).Result()
	return n > 0, err
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return RDB.Expire(Ctx, key, expiration).Err()
}

// Incr 自增
func Incr(key string) (int64, error) {
	return RDB.Incr(Ctx, key).Result()
}

// IncrBy 自增指定值
func IncrBy(key string, value int64) (int64, error) {
	return RDB.IncrBy(Ctx, key, value).Result()
}

// Decr 自减
func Decr(key string) (int64, error) {
	return RDB.Decr(Ctx, key).Result()
}

// HSet Hash设置
func HSet(key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.HSet(Ctx, key, field, data).Err()
}

// HGet Hash获取
func HGet(key string, field string, dest interface{}) error {
	data, err := RDB.HGet(Ctx, key, field).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// HGetAll Hash获取所有
func HGetAll(key string) (map[string]string, error) {
	return RDB.HGetAll(Ctx, key).Result()
}

// HDel Hash删除
func HDel(key string, fields ...string) error {
	return RDB.HDel(Ctx, key, fields...).Err()
}

// LPush 列表左推入
func LPush(key string, values ...interface{}) error {
	return RDB.LPush(Ctx, key, values...).Err()
}

// RPush 列表右推入
func RPush(key string, values ...interface{}) error {
	return RDB.RPush(Ctx, key, values...).Err()
}

// LPop 列表左弹出
func LPop(key string) (string, error) {
	return RDB.LPop(Ctx, key).Result()
}

// RPop 列表右弹出
func RPop(key string) (string, error) {
	return RDB.RPop(Ctx, key).Result()
}

// LRange 获取列表范围
func LRange(key string, start, stop int64) ([]string, error) {
	return RDB.LRange(Ctx, key, start, stop).Result()
}

// SetNX 设置缓存(不存在时才设置) - 用于分布式锁
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return RDB.SetNX(Ctx, key, value, expiration).Result()
}

// Publish 发布消息
func Publish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return RDB.Publish(Ctx, channel, data).Err()
}

// Subscribe 订阅消息
func Subscribe(channels ...string) *redis.PubSub {
	return RDB.Subscribe(Ctx, channels...)
}
