package service

import (
	"crypto/md5"
	"encoding/hex"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

// InitService 初始化服务
type InitService struct{}

// NewInitService 创建初始化服务实例
func NewInitService() *InitService {
	return &InitService{}
}

// InitTestData 初始化测试数据
func (s *InitService) InitTestData() error {
	// 执行数据库迁移，添加缺失的列
	s.migrateTransactionTable()

	// 修复admin密码（确保与PHP算法一致）
	s.fixAdminPassword()

	// 检查是否有管理员账号
	var count int64
	database.DB.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		logger.Info("管理员账号已存在，跳过初始化")
	}

	// 为测试用户添加初始余额（仅在开发模式下）
	s.initTestUserBalance()

	logger.Info("测试数据初始化完成")
	return nil
}

// hashPasswordPHP 与PHP的Users::MakePassword($password, 0)保持一致
func hashPasswordPHP(password string) string {
	// PHP算法：
	// $salt = 'ABCDEFG';
	// foreach (str_split($password) as $char) {
	//     $salt .= md5($char);
	// }
	// return md5($salt);
	salt := "ABCDEFG"

	for _, char := range password {
		charHash := md5.Sum([]byte(string(char)))
		salt += hex.EncodeToString(charHash[:])
	}

	finalHash := md5.Sum([]byte(salt))
	return hex.EncodeToString(finalHash[:])
}

// fixAdminPassword 修复admin密码使其使用标准MD5加密
func (s *InitService) fixAdminPassword() {
	// 检查admin账号是否存在
	var admin model.Admin
	if err := database.DB.Where("username = ?", "admin").First(&admin).Error; err != nil {
		logger.Info("admin账号不存在，跳过密码修复")
		return
	}

	// 计算正确的密码哈希（使用标准MD5加密）
	password := "admin123"
	hasher := md5.New()
	hasher.Write([]byte(password))
	correctPassword := hex.EncodeToString(hasher.Sum(nil))

	// 如果密码不匹配，则更新
	if admin.Password != correctPassword {
		if err := database.DB.Model(&admin).Update("password", correctPassword).Error; err != nil {
			logger.Errorf("修复admin密码失败: %v", err)
		} else {
			logger.Info("✓ admin密码已修复为标准MD5格式 (admin/admin123)")
		}
	} else {
		logger.Info("admin密码已是正确格式")
	}
}

// migrateTransactionTable 迁移transaction表，添加缺失的列
func (s *InitService) migrateTransactionTable() {
	// 检查并添加 legal 列
	if !s.columnExists("transaction", "legal") {
		if err := database.DB.Exec("ALTER TABLE `transaction` ADD COLUMN `legal` int(11) NOT NULL DEFAULT 0 AFTER `currency`").Error; err != nil {
			logger.Errorf("添加 transaction.legal 列失败: %v", err)
		} else {
			logger.Info("成功添加 transaction.legal 列")
		}
	}

	// 检查并添加 deal 列
	if !s.columnExists("transaction", "deal") {
		if err := database.DB.Exec("ALTER TABLE `transaction` ADD COLUMN `deal` decimal(20,5) NOT NULL DEFAULT 0.00000 AFTER `number`").Error; err != nil {
			logger.Errorf("添加 transaction.deal 列失败: %v", err)
		} else {
			logger.Info("成功添加 transaction.deal 列")
		}
	}

	// 检查并添加 fee 列
	if !s.columnExists("transaction", "fee") {
		if err := database.DB.Exec("ALTER TABLE `transaction` ADD COLUMN `fee` decimal(20,8) NOT NULL DEFAULT 0.00000000 AFTER `status`").Error; err != nil {
			logger.Errorf("添加 transaction.fee 列失败: %v", err)
		} else {
			logger.Info("成功添加 transaction.fee 列")
		}
	}
}

// columnExists 检查表中是否存在指定列
func (s *InitService) columnExists(tableName, columnName string) bool {
	var count int64
	database.DB.Raw(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName, columnName,
	).Scan(&count)
	return count > 0
}

// initTestUserBalance 为测试用户初始化余额
// 统一使用 UserAssets 表存储用户资产
func (s *InitService) initTestUserBalance() {
	// 获取所有用户
	var users []model.User
	database.DB.Find(&users)

	for _, user := range users {
		// 获取或创建 UserAssets 记录
		userAssets, err := GetUserAssetsService().GetOrCreateUserAssets(user.ID)
		if err != nil {
			logger.Errorf("获取用户 %d 资产记录失败: %v", user.ID, err)
			continue
		}

		// 如果USDT余额为0，则初始化测试余额
		if userAssets.UsdtBalance == 0 {
			// 初始化USDT余额
			if err := GetUserAssetsService().UpdateUsdtBalance(user.ID, 100000.0, false); err != nil {
				logger.Errorf("初始化用户 %d USDT余额失败: %v", user.ID, err)
			} else {
				logger.Infof("已为用户 %d 初始化USDT测试余额: 100000", user.ID)
			}

			// 初始化BTC余额
			if err := GetUserAssetsService().UpdateCurrencyBalance(user.ID, "BTC", 10.0, false); err != nil {
				logger.Errorf("初始化用户 %d BTC余额失败: %v", user.ID, err)
			} else {
				logger.Infof("已为用户 %d 初始化BTC测试余额: 10", user.ID)
			}

			// 初始化ETH余额
			if err := GetUserAssetsService().UpdateCurrencyBalance(user.ID, "ETH", 100.0, false); err != nil {
				logger.Errorf("初始化用户 %d ETH余额失败: %v", user.ID, err)
			} else {
				logger.Infof("已为用户 %d 初始化ETH测试余额: 100", user.ID)
			}
		}
	}
}
