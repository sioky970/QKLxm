package database

import (
	"fmt"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// ============================================================
// 迁移配置
// ============================================================

// MigrationConfig 迁移配置
type MigrationConfig struct {
	AutoMigrate     bool // 是否自动迁移表结构
	InitDefaultData bool // 是否初始化默认数据
	Verbose         bool // 是否输出详细日志
}

// DefaultMigrationConfig 默认迁移配置
var DefaultMigrationConfig = MigrationConfig{
	AutoMigrate:     true,
	InitDefaultData: true,
	Verbose:         true,
}

// ============================================================
// 需要迁移的模型列表（按依赖顺序排列）
// ============================================================

// CoreModels 核心配置表（优先迁移）
var CoreModels = []interface{}{
	&model.Setting{},             // 系统配置表
	&model.AdminRole{},           // 管理员角色表
	&model.AdminRolePermission{}, // 管理员角色权限关联表
	&model.AdminModule{},         // 后台功能模块表
	&model.AdminModuleAction{},   // 模块操作权限表
	&model.Admin{},               // 管理员表
	&model.AreaCode{},            // 地区代码表
	&model.Bank{},                // 银行列表表
}

// UserModels 用户相关表
var UserModels = []interface{}{
	&model.User{},          // 用户表
	&model.UserAssets{},    // 用户资产表
	&model.UsersWallet{},   // 用户钱包表（兼容旧数据）
	&model.UserReal{},      // 用户实名认证
	&model.UserProfiles{},  // 用户详情
	&model.UserCashInfo{},  // 用户提现信息
	&model.UserLoginInfo{}, // 用户登录信息
	&model.Token{},         // 用户登录Token表
	&model.Level{},         // 用户等级表
}

// CurrencyModels 币种相关表
var CurrencyModels = []interface{}{
	&model.Currency{},          // 币种表
	&model.CurrencyMatch{},     // 交易对配置表
	&model.CurrencyQuotation{}, // 行情数据表
	&model.LeverMultiple{},     // 杠杆倍数配置表
	&model.Market{},            // 市场信息表
	&model.MarketDay{},         // 日K线数据表
	&model.MarketHour{},        // 小时K线数据表
}

// TradingModels 交易相关表
var TradingModels = []interface{}{
	&model.Transaction{},         // 现货交易订单表
	&model.TransactionComplete{}, // 成交记录表
	&model.LeverTransaction{},    // 合约交易表
	&model.MicroOrder{},          // 秒合约订单表
	&model.MicroSeconds{},        // 秒合约周期配置表
}

// DepositModels 充值提现相关表
var DepositModels = []interface{}{
	&model.DepositAddress{},   // 充值地址配置表
	&model.DepositOrder{},     // 充值订单表
	&model.UsersWalletOut{},   // 提现订单表
	&model.Address{},          // 用户提币地址表
	&model.WalletAddressLog{}, // 钱包地址变更日志表
}

// LogModels 日志相关表
var LogModels = []interface{}{
	&model.AccountLog{}, // 账户日志
	&model.WalletLog{},  // 钱包日志
}

// MessageModels 消息相关表
var MessageModels = []interface{}{
	&model.Message{},         // 消息表
	&model.UserMessage{},     // 用户消息关联表
	&model.MessageTemplate{}, // 消息模板表
}

// RiskModels 风控相关表
var RiskModels = []interface{}{
	&model.UserRiskControl{}, // 用户风控配置
	&model.RiskKline{},       // 风控K线数据
}

// OtherModels 其他表
var OtherModels = []interface{}{
	&model.News{},         // 新闻表
	&model.NewsCategory{}, // 新闻分类表
	&model.Feedback{},     // 用户反馈表
	&model.Robot{},        // 机器人配置表
	&model.Algebra{},      // 推广层级配置表
}

// ============================================================
// 默认数据定义
// ============================================================

// DefaultSettings 默认系统配置
var DefaultSettings = []model.Setting{
	{Key: "register_mandatory_invitation", Value: "0", Notes: "注册是否必须使用邀请码：0-选填，1-必填"},
	{Key: "site_name", Value: "Exchange", Notes: "网站名称"},
	{Key: "site_email", Value: "support@exchange.com", Notes: "官方邮箱"},
	{Key: "withdraw_fee_rate", Value: "0.001", Notes: "提现手续费率"},
	{Key: "withdraw_min", Value: "10", Notes: "最小提现金额"},
	{Key: "deposit_min", Value: "10", Notes: "最小充值金额"},
	{Key: "trade_fee_rate", Value: "0.001", Notes: "交易手续费率"},
	{Key: "contract_fee_rate", Value: "0.0006", Notes: "合约手续费率"},
}

// DefaultMicroSeconds 默认秒合约周期配置
var DefaultMicroSeconds = []model.MicroSeconds{
	{Seconds: 30, ProfitRatio: 0.40, Status: 1},
	{Seconds: 60, ProfitRatio: 0.50, Status: 1},
	{Seconds: 120, ProfitRatio: 0.60, Status: 1},
	{Seconds: 180, ProfitRatio: 0.80, Status: 1},
	{Seconds: 300, ProfitRatio: 1.00, Status: 1},
}

// DefaultLeverMultiples 默认杠杆倍数配置
var DefaultLeverMultiples = []model.LeverMultiple{
	{Type: 1, Value: "5"},
	{Type: 1, Value: "10"},
	{Type: 1, Value: "20"},
	{Type: 1, Value: "50"},
	{Type: 1, Value: "100"},
	{Type: 1, Value: "125"},
	{Type: 1, Value: "150"},
	{Type: 1, Value: "200"},
}

// DefaultAdminRole 默认管理员角色
var DefaultAdminRole = model.AdminRole{
	ID:      1,
	Name:    "超级管理员",
	IsSuper: 1,
}

// DefaultAdminModules 默认后台功能模块
var DefaultAdminModules = []model.AdminModule{
	{ID: 1, Name: "基础", Module: "base"},
	{ID: 2, Name: "设置", Module: "setting"},
	{ID: 3, Name: "用户", Module: "user"},
	{ID: 4, Name: "钱包", Module: "wallet"},
	{ID: 5, Name: "管理员", Module: "manager"},
	{ID: 6, Name: "新闻", Module: "news"},
	{ID: 7, Name: "反馈建议", Module: "feedback"},
	{ID: 8, Name: "币种管理", Module: "currency"},
	{ID: 9, Name: "法币交易", Module: "legal"},
	{ID: 10, Name: "C2C交易", Module: "c2c"},
	{ID: 11, Name: "合约交易", Module: "lever"},
	{ID: 12, Name: "记录查看", Module: "record"},
}

// DefaultBanks 默认银行列表
var DefaultBanks = []model.Bank{
	{ID: 1, Name: "中国银行"},
	{ID: 2, Name: "中国工商银行"},
	{ID: 3, Name: "中国建设银行"},
	{ID: 4, Name: "中国农业银行"},
	{ID: 5, Name: "招商银行"},
	{ID: 6, Name: "交通银行"},
	{ID: 7, Name: "中国邮政储蓄银行"},
	{ID: 8, Name: "兴业银行"},
	{ID: 9, Name: "浦发银行"},
	{ID: 10, Name: "中信银行"},
	{ID: 11, Name: "光大银行"},
	{ID: 12, Name: "广发银行"},
	{ID: 13, Name: "华夏银行"},
	{ID: 14, Name: "民生银行"},
	{ID: 15, Name: "平安银行"},
}

// DefaultAreaCodes 默认地区代码
var DefaultAreaCodes = []model.AreaCode{
	{Name: "中国", AreaCode: "+86"},
	{Name: "香港", AreaCode: "+852"},
	{Name: "台湾", AreaCode: "+886"},
	{Name: "美国", AreaCode: "+1"},
	{Name: "日本", AreaCode: "+81"},
	{Name: "韩国", AreaCode: "+82"},
	{Name: "新加坡", AreaCode: "+65"},
	{Name: "英国", AreaCode: "+44"},
}

// DefaultCurrencies 默认币种数据（48种主流币种）
var DefaultCurrencies = []model.Currency{
	// 主流币种 - 高优先级
	{Name: "BTC", Logo: "/coins/btc.png", Type: "coin", DecimalScale: 8, Rate: 88737.18, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 1000},
	{Name: "ETH", Logo: "/coins/eth.png", Type: "coin", DecimalScale: 8, Rate: 2930.35, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 999},
	{Name: "USDT", Logo: "/coins/usdt.png", Type: "coin", DecimalScale: 8, Rate: 1, IsLegal: 1, IsDisplay: 1, IsChange: 1, Sort: 998}, // 法币
	{Name: "BNB", Logo: "/coins/bnb.png", Type: "coin", DecimalScale: 8, Rate: 878.56, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 997},
	{Name: "XRP", Logo: "/coins/xrp.png", Type: "coin", DecimalScale: 8, Rate: 1.88735, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 996},
	{Name: "SOL", Logo: "/coins/sol.png", Type: "coin", DecimalScale: 8, Rate: 126.0295, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 995},
	{Name: "TRX", Logo: "/coins/trx.png", Type: "coin", DecimalScale: 8, Rate: 0.298202, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 994},
	{Name: "DOGE", Logo: "/coins/doge.png", Type: "coin", DecimalScale: 8, Rate: 0.122078, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 993},
	{Name: "ADA", Logo: "/coins/ada.png", Type: "coin", DecimalScale: 8, Rate: 0.352364, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 992},
	{Name: "BCH", Logo: "/coins/bch.png", Type: "coin", DecimalScale: 8, Rate: 582.73, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 991},
	{Name: "XMR", Logo: "/coins/xmr.png", Type: "coin", DecimalScale: 8, Rate: 467.32, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 990},
	{Name: "LINK", Logo: "/coins/link.png", Type: "coin", DecimalScale: 8, Rate: 11.9891, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 989},
	{Name: "XLM", Logo: "/coins/xlm.png", Type: "coin", DecimalScale: 8, Rate: 0.208438, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 988},
	{Name: "ZEC", Logo: "/coins/zec.png", Type: "coin", DecimalScale: 8, Rate: 358.15, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 987},
	{Name: "LTC", Logo: "/coins/ltc.png", Type: "coin", DecimalScale: 8, Rate: 68.91, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 986},
	{Name: "AVAX", Logo: "/coins/avax.png", Type: "coin", DecimalScale: 8, Rate: 11.8746, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 985},
	{Name: "DOT", Logo: "/coins/dot.png", Type: "coin", DecimalScale: 8, Rate: 1.8884, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 984},
	{Name: "UNI", Logo: "/coins/uni.png", Type: "coin", DecimalScale: 8, Rate: 4.8278, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 983},
	{Name: "AAVE", Logo: "/coins/aave.png", Type: "coin", DecimalScale: 8, Rate: 153.7133, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 982},
	{Name: "ETC", Logo: "/coins/etc.png", Type: "coin", DecimalScale: 8, Rate: 11.5, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 981},
	{Name: "ATOM", Logo: "/coins/atom.png", Type: "coin", DecimalScale: 8, Rate: 2.267, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 980},
	{Name: "ALGO", Logo: "/coins/algo.png", Type: "coin", DecimalScale: 8, Rate: 0.116, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 979},
	{Name: "FIL", Logo: "/coins/fil.png", Type: "coin", DecimalScale: 8, Rate: 1.2836, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 978},
	{Name: "VET", Logo: "/coins/vet.png", Type: "coin", DecimalScale: 8, Rate: 0.010212, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 977},
	{Name: "DASH", Logo: "/coins/dash.png", Type: "coin", DecimalScale: 8, Rate: 62.25, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 976},
	{Name: "XTZ", Logo: "/coins/xtz.png", Type: "coin", DecimalScale: 8, Rate: 0.5785, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 975},
	{Name: "CRV", Logo: "/coins/crv.png", Type: "coin", DecimalScale: 8, Rate: 0.3569, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 974},
	{Name: "GRT", Logo: "/coins/grt.png", Type: "coin", DecimalScale: 8, Rate: 0.036025, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 973},
	{Name: "IOTA", Logo: "/coins/iota.png", Type: "coin", DecimalScale: 8, Rate: 0.0855, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 972},
	{Name: "SAND", Logo: "/coins/sand.png", Type: "coin", DecimalScale: 8, Rate: 0.13626, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 971},
	{Name: "THETA", Logo: "/coins/theta.png", Type: "coin", DecimalScale: 8, Rate: 0.2868, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 970},
	{Name: "MANA", Logo: "/coins/mana.png", Type: "coin", DecimalScale: 8, Rate: 0.1471, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 969},
	{Name: "BAT", Logo: "/coins/bat.png", Type: "coin", DecimalScale: 8, Rate: 0.1744, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 968},
	{Name: "NEO", Logo: "/coins/neo.png", Type: "coin", DecimalScale: 8, Rate: 3.57, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 967},
	{Name: "COMP", Logo: "/coins/comp.png", Type: "coin", DecimalScale: 8, Rate: 24.61, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 966},
	{Name: "APE", Logo: "/coins/ape.png", Type: "coin", DecimalScale: 8, Rate: 0.1815, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 965},
	{Name: "SNX", Logo: "/coins/snx.png", Type: "coin", DecimalScale: 8, Rate: 0.4132, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 964},
	{Name: "QTUM", Logo: "/coins/qtum.png", Type: "coin", DecimalScale: 8, Rate: 1.2723, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 963},
	{Name: "KSM", Logo: "/coins/ksm.png", Type: "coin", DecimalScale: 8, Rate: 8.4936, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 962},
	{Name: "YFI", Logo: "/coins/yfi.png", Type: "coin", DecimalScale: 8, Rate: 3309.27, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 961},
	{Name: "ZRX", Logo: "/coins/zrx.png", Type: "coin", DecimalScale: 8, Rate: 0.1258, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 960},
	{Name: "ZIL", Logo: "/coins/zil.png", Type: "coin", DecimalScale: 8, Rate: 0.004908, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 959},
	{Name: "SUSHI", Logo: "/coins/sushi.png", Type: "coin", DecimalScale: 8, Rate: 0.2992, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 958},
	{Name: "ICX", Logo: "/coins/icx.png", Type: "coin", DecimalScale: 8, Rate: 0.0553, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 957},
	{Name: "ONT", Logo: "/coins/ont.png", Type: "coin", DecimalScale: 8, Rate: 0.0577, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 956},
	{Name: "KNC", Logo: "/coins/knc.png", Type: "coin", DecimalScale: 8, Rate: 0.2248, IsLegal: 0, IsDisplay: 1, IsChange: 1, Sort: 955},
}

// ============================================================
// 迁移执行函数
// ============================================================

// RunMigration 执行数据库迁移（使用默认配置）
func RunMigration() error {
	return RunMigrationWithConfig(DefaultMigrationConfig)
}

// RunMigrationWithConfig 执行数据库迁移（自定义配置）
func RunMigrationWithConfig(config MigrationConfig) error {
	startTime := time.Now()
	logger.Info("========================================")
	logger.Info("开始数据库迁移...")
	logger.Info("========================================")

	// 1. 自动迁移表结构
	if config.AutoMigrate {
		if err := migrateAllTables(config.Verbose); err != nil {
			return fmt.Errorf("表结构迁移失败: %w", err)
		}
	}

	// 2. 初始化默认数据
	if config.InitDefaultData {
		if err := initAllDefaultData(config.Verbose); err != nil {
			return fmt.Errorf("默认数据初始化失败: %w", err)
		}
	}

	duration := time.Since(startTime)
	logger.Info("========================================")
	logger.Info("数据库迁移完成，耗时: %v", duration)
	logger.Info("========================================")

	return nil
}

// migrateAllTables 迁移所有表结构
func migrateAllTables(verbose bool) error {
	// 按模块分组迁移，便于追踪问题
	moduleGroups := []struct {
		name   string
		models []interface{}
	}{
		{"核心配置表", CoreModels},
		{"用户相关表", UserModels},
		{"币种相关表", CurrencyModels},
		{"交易相关表", TradingModels},
		{"充值提现表", DepositModels},
		{"日志相关表", LogModels},
		{"消息相关表", MessageModels},
		{"风控相关表", RiskModels},
		{"其他业务表", OtherModels},
	}

	for _, group := range moduleGroups {
		if verbose {
			logger.Info("[迁移] %s...", group.name)
		}

		if err := DB.AutoMigrate(group.models...); err != nil {
			logger.Error("[迁移] %s 失败: %v", group.name, err)
			return err
		}

		if verbose {
			logger.Info("[迁移] %s 完成 (%d个表)", group.name, len(group.models))
		}
	}

	return nil
}

// initAllDefaultData 初始化所有默认数据
func initAllDefaultData(verbose bool) error {
	// 使用事务确保数据一致性
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. 初始化系统配置
		if verbose {
			logger.Info("[初始化] 系统配置...")
		}
		if err := initSettings(tx, verbose); err != nil {
			return err
		}

		// 2. 初始化管理员角色
		if verbose {
			logger.Info("[初始化] 管理员角色...")
		}
		if err := initAdminRole(tx, verbose); err != nil {
			return err
		}

		// 3. 初始化后台功能模块
		if verbose {
			logger.Info("[初始化] 后台功能模块...")
		}
		if err := initAdminModules(tx, verbose); err != nil {
			return err
		}

		// 4. 初始化地区代码
		if verbose {
			logger.Info("[初始化] 地区代码...")
		}
		if err := initAreaCodes(tx, verbose); err != nil {
			return err
		}

		// 5. 初始化银行列表
		if verbose {
			logger.Info("[初始化] 银行列表...")
		}
		if err := initBanks(tx, verbose); err != nil {
			return err
		}

		// 6. 初始化秒合约周期配置
		if verbose {
			logger.Info("[初始化] 秒合约周期...")
		}
		if err := initMicroSeconds(tx, verbose); err != nil {
			return err
		}

		// 7. 初始化杠杆倍数配置
		if verbose {
			logger.Info("[初始化] 杠杆倍数...")
		}
		if err := initLeverMultiples(tx, verbose); err != nil {
			return err
		}

		// 8. 初始化基础币种（48种主流币种）
		if verbose {
			logger.Info("[初始化] 基础币种 (48种)...")
		}
		if err := initCurrencies(tx, verbose); err != nil {
			return err
		}

		return nil
	})
}

// ============================================================
// 各模块初始化函数
// ============================================================

// initSettings 初始化系统配置
func initSettings(tx *gorm.DB, verbose bool) error {
	for _, setting := range DefaultSettings {
		var existing model.Setting
		err := tx.Where("`key` = ?", setting.Key).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&setting).Error; err != nil {
				return fmt.Errorf("创建配置 %s 失败: %w", setting.Key, err)
			}
			if verbose {
				logger.Info("  + 创建配置: %s = %s", setting.Key, setting.Value)
			}
		}
	}
	return nil
}

// initAdminRole 初始化管理员角色
func initAdminRole(tx *gorm.DB, verbose bool) error {
	var existing model.AdminRole
	err := tx.Where("id = ?", DefaultAdminRole.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := tx.Create(&DefaultAdminRole).Error; err != nil {
			return fmt.Errorf("创建管理员角色失败: %w", err)
		}
		if verbose {
			logger.Info("  + 创建角色: %s (ID=%d)", DefaultAdminRole.Name, DefaultAdminRole.ID)
		}
	}
	return nil
}

// initAdminModules 初始化后台功能模块
func initAdminModules(tx *gorm.DB, verbose bool) error {
	createdCount := 0
	for _, module := range DefaultAdminModules {
		var existing model.AdminModule
		err := tx.Where("id = ?", module.ID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&module).Error; err != nil {
				return fmt.Errorf("创建后台模块 %s 失败: %w", module.Name, err)
			}
			createdCount++
			if verbose {
				logger.Info("  + 创建模块: %s (%s)", module.Name, module.Module)
			}
		}
	}
	if verbose && createdCount > 0 {
		logger.Info("  后台模块初始化完成: 创建 %d 个", createdCount)
	}
	return nil
}

// initAreaCodes 初始化地区代码
func initAreaCodes(tx *gorm.DB, verbose bool) error {
	createdCount := 0
	for _, areaCode := range DefaultAreaCodes {
		var existing model.AreaCode
		err := tx.Where("area_code = ?", areaCode.AreaCode).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&areaCode).Error; err != nil {
				return fmt.Errorf("创建地区代码 %s 失败: %w", areaCode.Name, err)
			}
			createdCount++
			if verbose {
				logger.Info("  + 创建地区: %s (%s)", areaCode.Name, areaCode.AreaCode)
			}
		}
	}
	if verbose && createdCount > 0 {
		logger.Info("  地区代码初始化完成: 创建 %d 个", createdCount)
	}
	return nil
}

// initBanks 初始化银行列表
func initBanks(tx *gorm.DB, verbose bool) error {
	createdCount := 0
	for _, bank := range DefaultBanks {
		var existing model.Bank
		err := tx.Where("id = ?", bank.ID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&bank).Error; err != nil {
				return fmt.Errorf("创建银行 %s 失败: %w", bank.Name, err)
			}
			createdCount++
			if verbose {
				logger.Info("  + 创建银行: %s", bank.Name)
			}
		}
	}
	if verbose && createdCount > 0 {
		logger.Info("  银行列表初始化完成: 创建 %d 个", createdCount)
	}
	return nil
}

// initMicroSeconds 初始化秒合约周期配置
func initMicroSeconds(tx *gorm.DB, verbose bool) error {
	for _, sec := range DefaultMicroSeconds {
		var existing model.MicroSeconds
		err := tx.Where("seconds = ?", sec.Seconds).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			now := time.Now()
			sec.CreatedAt = now
			sec.UpdatedAt = now
			if err := tx.Create(&sec).Error; err != nil {
				return fmt.Errorf("创建秒合约周期 %d秒 失败: %w", sec.Seconds, err)
			}
			if verbose {
				logger.Info("  + 创建周期: %d秒, 盈利率=%.0f%%", sec.Seconds, sec.ProfitRatio*100)
			}
		}
	}
	return nil
}

// initLeverMultiples 初始化杠杆倍数配置
func initLeverMultiples(tx *gorm.DB, verbose bool) error {
	for _, lm := range DefaultLeverMultiples {
		var existing model.LeverMultiple
		err := tx.Where("type = ? AND value = ? AND currency_id IS NULL", lm.Type, lm.Value).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&lm).Error; err != nil {
				return fmt.Errorf("创建杠杆倍数 %sx 失败: %w", lm.Value, err)
			}
			if verbose {
				logger.Info("  + 创建杠杆: %sx (全局)", lm.Value)
			}
		}
	}
	return nil
}

// initCurrencies 初始化基础币种（48种主流币种）
func initCurrencies(tx *gorm.DB, verbose bool) error {
	createdCount := 0
	skippedCount := 0
	now := time.Now().Unix()

	for _, currency := range DefaultCurrencies {
		var existing model.Currency
		err := tx.Where("name = ?", currency.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			// 创建新币种
			currency.CreateTime = now
			currency.UpdateTime = now
			if err := tx.Create(&currency).Error; err != nil {
				return fmt.Errorf("创建币种 %s 失败: %w", currency.Name, err)
			}
			createdCount++
			if verbose {
				logger.Info("  + 创建币种: %s (Sort=%d)", currency.Name, currency.Sort)
			}
		} else if err == nil {
			// 币种已存在，更新基本信息（不覆盖用户修改的显示状态）
			updates := map[string]interface{}{
				"logo":          currency.Logo,
				"type":          currency.Type,
				"decimal_scale": currency.DecimalScale,
				"rate":          currency.Rate,
				"is_legal":      currency.IsLegal,
				"is_change":     currency.IsChange,
				"sort":          currency.Sort,
				"update_time":   now,
			}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新币种 %s 失败: %w", currency.Name, err)
			}
			skippedCount++
		} else {
			return fmt.Errorf("查询币种 %s 失败: %w", currency.Name, err)
		}
	}

	if verbose {
		logger.Info("  币种初始化完成: 新建 %d 个, 更新 %d 个", createdCount, skippedCount)
	}
	return nil
}

// ============================================================
// 工具函数
// ============================================================

// CheckTablesExist 检查指定表是否存在
func CheckTablesExist(tableNames ...string) map[string]bool {
	result := make(map[string]bool)
	for _, name := range tableNames {
		result[name] = DB.Migrator().HasTable(name)
	}
	return result
}

// GetMigrationStatus 获取迁移状态报告
func GetMigrationStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// 检查核心表
	coreTables := []string{
		"settings", "admin", "admin_role", "users", "user_assets",
		"currency", "currency_matches", "transaction", "lever_transaction",
		"micro_orders", "micro_seconds", "deposit_address", "deposit_order",
	}

	tableStatus := make(map[string]bool)
	for _, table := range coreTables {
		tableStatus[table] = DB.Migrator().HasTable(table)
	}
	status["tables"] = tableStatus

	// 统计数据
	var settingCount, adminCount, currencyCount int64
	DB.Model(&model.Setting{}).Count(&settingCount)
	DB.Model(&model.Admin{}).Count(&adminCount)
	DB.Model(&model.Currency{}).Count(&currencyCount)

	status["counts"] = map[string]int64{
		"settings":   settingCount,
		"admins":     adminCount,
		"currencies": currencyCount,
	}

	return status
}

// ResetMigration 重置迁移（危险操作，仅用于开发环境）
func ResetMigration(confirm bool) error {
	if !confirm {
		return fmt.Errorf("请确认执行重置操作")
	}

	logger.Warn("警告: 正在执行数据库重置...")

	// 删除所有表并重新迁移
	allModels := append(append(append(append(append(append(append(append(
		CoreModels,
		UserModels...),
		CurrencyModels...),
		TradingModels...),
		DepositModels...),
		LogModels...),
		MessageModels...),
		RiskModels...),
		OtherModels...)

	// 删除表
	for _, m := range allModels {
		if err := DB.Migrator().DropTable(m); err != nil {
			logger.Warn("删除表失败: %v", err)
		}
	}

	// 重新迁移
	return RunMigration()
}
