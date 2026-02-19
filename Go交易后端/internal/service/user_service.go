package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"exchange-go/config"
	"exchange-go/internal/api/middleware"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Type          string `json:"type" binding:"required"`        // mobile 或 email
	UserString    string `json:"user_string" binding:"required"` // 手机号或邮箱
	Password      string `json:"password" binding:"required,min=6,max=16"`
	RePassword    string `json:"re_password" binding:"required"`
	Code          string `json:"code" binding:"required"` // 验证码
	ExtensionCode string `json:"extension_code"`          // 邀请码
	CountryCode   int    `json:"country_code"`            // 区号ID
}

// LoginRequest 登录请求
type LoginRequest struct {
	UserString string `json:"user_string" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Type       int    `json:"type"`         // 1:普通密码 2:手势密码
	AreaCodeID int    `json:"area_code_id"` // 区号ID
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=16"`
	RePassword  string `json:"re_password" binding:"required"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Account     string `json:"account" binding:"required"`
	CountryCode int    `json:"country_code"`
	Password    string `json:"password" binding:"required,min=6,max=16"`
	RePassword  string `json:"repassword" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

// UserCashInfoRequest 提现信息请求
type UserCashInfoRequest struct {
	BankName       string `json:"bank_name" binding:"required"`
	BankBranch     string `json:"bank_branch"`
	BankAccount    string `json:"bank_account" binding:"required"`
	RealName       string `json:"real_name" binding:"required"`
	AlipayAccount  string `json:"alipay_account"`
	WechatNickname string `json:"wechat_nickname"`
	WechatAccount  string `json:"wechat_account"`
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest) error {
	// 验证密码一致性
	if req.Password != req.RePassword {
		return errors.New("两次密码不一致")
	}

	// 验证验证码 (测试模式: 使用 "000000" 可跳过验证码)
	cacheKey := fmt.Sprintf("verify_code:%s", req.UserString)
	if req.Code != "000000" {
		cachedCode, err := cache.GetString(cacheKey)
		if err != nil || cachedCode != req.Code {
			return errors.New("验证码错误或已过期")
		}
	}

	// 检查用户是否已存在（account_number 必须唯一）
	var existUser model.User
	result := database.DB.Where("account_number = ?", req.UserString).First(&existUser)
	if result.RowsAffected > 0 {
		return errors.New("账号已存在")
	}

	// 如果是手机号，检查手机号是否已被使用
	if req.Type == "mobile" {
		var existPhoneUser model.User
		result := database.DB.Where("phone = ?", req.UserString).First(&existPhoneUser)
		if result.RowsAffected > 0 {
			return errors.New("手机号已被使用")
		}
	}

	// 如果是邮箱，检查邮箱是否已被使用
	if req.Type == "email" {
		var existEmailUser model.User
		result := database.DB.Where("email = ?", req.UserString).First(&existEmailUser)
		if result.RowsAffected > 0 {
			return errors.New("邮箱已被使用")
		}
	}

	// 处理邀请码 (测试模式: 使用 "000000" 验证码时可跳过邀请码)
	var parentID uint = 0
	if req.ExtensionCode != "" {
		var parentUser model.User
		result := database.DB.Where("extension_code = ?", req.ExtensionCode).First(&parentUser)
		if result.Error != nil {
			return errors.New("请填写正确的邀请码")
		}
		parentID = parentUser.ID
	} else {
		// 检查是否必填邀请码
		var mandatorySetting model.Setting
		err := database.DB.Where("`key` = ?", "register_mandatory_invitation").First(&mandatorySetting).Error
		// 如果设置为必填 (1)，且不是测试验证码，则报错
		if err == nil && mandatorySetting.Value == "1" && req.Code != "000000" {
			return errors.New("请填写邀请码")
		}
	}

	// 创建用户
	user := &model.User{
		AccountNumber: req.UserString,
		Password:      MakePassword(req.Password),
		ParentID:      parentID,
		AreaCodeID:    uint(req.CountryCode),
		ExtensionCode: GenerateExtensionCode(),
		Status:        0,
		HeadPortrait:  "/mobile/images/user_head.png",
		Time:          time.Now().Unix(),
	}

	if req.Type == "mobile" {
		user.Phone = req.UserString
		user.Email = ""
	} else {
		user.Email = req.UserString
		user.Phone = ""
	}

	// 使用事务创建用户和钱包
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建用户钱包
	if err := CreateUserWallets(tx, user.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf("创建钱包失败: %w", err)
	}

	tx.Commit()

	// 删除验证码缓存
	cache.Delete(cacheKey)

	logger.Infof("用户注册成功: %s", req.UserString)
	return nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	if req.UserString == "" {
		return nil, errors.New("请输入账号")
	}
	if req.Password == "" {
		return nil, errors.New("请输入密码")
	}

	// 查找用户
	var user model.User
	result := database.DB.Where("account_number = ? AND area_code_id = ?", req.UserString, req.AreaCodeID).First(&user)
	if result.Error != nil {
		return nil, errors.New("用户未找到")
	}

	// 验证密码
	if req.Type == 0 || req.Type == 1 {
		if !VerifyPassword(user.Password, req.Password) {
			return nil, errors.New("密码错误")
		}
	} else if req.Type == 2 {
		if req.Password != user.GesturePassword {
			return nil, errors.New("手势密码错误")
		}
	}

	// 检查账号状态
	if user.Status == 1 {
		return nil, errors.New("账号已锁定！请联系管理员")
	}

	// 生成 JWT Token
	token, err := middleware.GenerateToken(user.ID, user.Phone, user.Email)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 更新登录时间和IP
	database.DB.Model(&user).Updates(map[string]interface{}{
		"last_time": time.Now().Unix(),
	})

	// 清除密码后返回
	user.Password = ""
	user.PayPassword = ""
	user.GesturePassword = ""

	logger.Infof("用户登录成功: %s", req.UserString)
	return &LoginResponse{
		Token: token,
		User:  &user,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID uint) (*model.User, error) {
	var user model.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		return nil, errors.New("用户不存在")
	}

	// 清除敏感信息
	user.Password = ""
	user.PayPassword = ""
	user.GesturePassword = ""

	return &user, nil
}

// GetUserInfoWithPayPasswordStatus 获取用户信息（包含是否设置支付密码状态）
func (s *UserService) GetUserInfoWithPayPasswordStatus(userID uint) (*model.User, bool, error) {
	var user model.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		return nil, false, errors.New("用户不存在")
	}

	// 检查是否设置了支付密码
	hasPayPassword := user.PayPassword != ""

	// 清除敏感信息
	user.Password = ""
	user.PayPassword = ""
	user.GesturePassword = ""

	return &user, hasPayPassword, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	if req.NewPassword != req.RePassword {
		return errors.New("两次密码不一致")
	}

	var user model.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !VerifyPassword(user.Password, req.OldPassword) {
		return errors.New("原密码错误")
	}

	// 更新密码
	newPassword := MakePassword(req.NewPassword)
	if err := database.DB.Model(&user).Update("password", newPassword).Error; err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	logger.Infof("用户修改密码成功: %d", userID)
	return nil
}

// ResetPassword 重置密码 (忘记密码)
func (s *UserService) ResetPassword(req *ResetPasswordRequest) error {
	if req.Password != req.RePassword {
		return errors.New("两次密码不一致")
	}

	// 验证验证码
	cacheKey := fmt.Sprintf("verify_code:%s", req.Account)
	cachedCode, err := cache.GetString(cacheKey)
	if err != nil || cachedCode != req.Code {
		return errors.New("验证码错误或已过期")
	}

	// 查找用户
	var user model.User
	result := database.DB.Where("account_number = ? AND area_code_id = ?", req.Account, req.CountryCode).First(&user)
	if result.Error != nil {
		return errors.New("账号不存在")
	}

	// 更新密码
	newPassword := MakePassword(req.Password)
	if err := database.DB.Model(&user).Update("password", newPassword).Error; err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	// 删除验证码缓存
	cache.Delete(cacheKey)

	logger.Infof("用户重置密码成功: %s", req.Account)
	return nil
}

// ChangePayPassword 修改支付密码
func (s *UserService) ChangePayPassword(userID uint, oldPassword, newPassword string) error {
	var user model.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		return errors.New("用户不存在")
	}

	// 如果设置过支付密码，需要验证旧密码
	if user.PayPassword != "" {
		if !VerifyPayPassword(user.PayPassword, oldPassword) {
			return errors.New("原支付密码错误")
		}
	}

	// 更新支付密码
	newPwd := MakePayPassword(newPassword)
	if err := database.DB.Model(&user).Update("pay_password", newPwd).Error; err != nil {
		return fmt.Errorf("修改支付密码失败: %w", err)
	}

	logger.Infof("用户修改支付密码成功: %d", userID)
	return nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(userID uint, updates map[string]interface{}) error {
	// 过滤不允许更新的字段
	delete(updates, "id")
	delete(updates, "password")
	delete(updates, "pay_password")
	delete(updates, "status")
	delete(updates, "extension_code")
	delete(updates, "parent_id")

	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户信息失败: %w", err)
	}

	return nil
}

// GetUserCashInfo 获取用户提现信息
func (s *UserService) GetUserCashInfo(userID uint) (*model.UserCashInfo, error) {
	var cash model.UserCashInfo
	err := database.DB.Where("user_id = ?", userID).First(&cash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cash, nil
}

// SaveUserCashInfo 保存用户提现信息
func (s *UserService) SaveUserCashInfo(userID uint, req *UserCashInfoRequest) error {
	var cash model.UserCashInfo
	err := database.DB.Where("user_id = ?", userID).First(&cash).Error

	now := time.Now().Unix()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建
		cash = model.UserCashInfo{
			UserID:         userID,
			BankName:       req.BankName,
			BankBranch:     req.BankBranch,
			BankAccount:    req.BankAccount,
			RealName:       req.RealName,
			AlipayAccount:  req.AlipayAccount,
			WechatNickname: req.WechatNickname,
			WechatAccount:  req.WechatAccount,
			CreateTime:     now,
		}
		return database.DB.Create(&cash).Error
	} else if err == nil {
		// 更新
		updates := map[string]interface{}{
			"bank_name":       req.BankName,
			"bank_branch":     req.BankBranch,
			"bank_account":    req.BankAccount,
			"real_name":       req.RealName,
			"alipay_account":  req.AlipayAccount,
			"wechat_nickname": req.WechatNickname,
			"wechat_account":  req.WechatAccount,
		}
		return database.DB.Model(&cash).Updates(updates).Error
	}

	return err
}

// ============ 工具函数 ============

// MakePassword 密码加密 (使用Bcrypt)
func MakePassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("密码加密失败: %v", err)
		return ""
	}
	return string(hashed)
}

// VerifyPassword 验证密码
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// MakePayPassword 支付密码加密 (使用Bcrypt)
func MakePayPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("支付密码加密失败: %v", err)
		return ""
	}
	return string(hashed)
}

// VerifyPayPassword 验证支付密码
func VerifyPayPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateExtensionCode 生成邀请码
func GenerateExtensionCode() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())

	for {
		code := make([]byte, 6)
		for i := range code {
			code[i] = chars[rand.Intn(len(chars))]
		}
		codeStr := string(code)

		// 检查是否已存在
		var count int64
		database.DB.Model(&model.User{}).Where("extension_code = ?", codeStr).Count(&count)
		if count == 0 {
			return codeStr
		}
	}
}

// CreateUserWallets 创建用户钱包
// 注意：已迁移到UserAssets表，不再创建UsersWallet记录
func CreateUserWallets(tx *gorm.DB, userID uint) error {
	// 统一使用 UserAssets 表创建用户资产记录
	// 注意：这里不直接使用tx，因为UserAssetsService有自己的事务处理
	_, err := GetUserAssetsService().CreateUserAssets(userID)
	if err != nil {
		return err
	}
	return nil
}

// SendVerifyCode 发送验证码
func SendVerifyCode(target, codeType string) (string, error) {
	// 生成6位验证码
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 存入Redis，5分钟过期
	cacheKey := fmt.Sprintf("verify_code:%s", target)
	if err := cache.SetString(cacheKey, code, 5*time.Minute); err != nil {
		return "", fmt.Errorf("保存验证码失败: %w", err)
	}

	// TODO: 实际发送验证码逻辑
	// 根据 codeType 判断是邮件还是短信
	if codeType == "email" {
		// 发送邮件
		logger.Infof("发送邮件验证码: %s -> %s", target, code)
	} else {
		// 发送短信
		logger.Infof("发送短信验证码: %s -> %s", target, code)
	}

	// 开发环境返回验证码，生产环境不返回
	if config.GlobalConfig.App.Mode == "debug" {
		return code, nil
	}
	return "", nil
}

// ============ 公开配置接口 ============

// RegisterConfigResponse 注册配置响应
type RegisterConfigResponse struct {
	InviteCodeRequired bool `json:"invite_code_required"` // 邀请码是否必填
}

// GetRegisterConfig 获取注册配置
func GetRegisterConfig() (*RegisterConfigResponse, error) {
	config := &RegisterConfigResponse{
		InviteCodeRequired: false, // 默认不必填
	}

	// 查询邀请码必填设置
	var setting model.Setting
	err := database.DB.Where("`key` = ?", "register_mandatory_invitation").First(&setting).Error
	if err == nil && setting.Value == "1" {
		config.InviteCodeRequired = true
	}

	return config, nil
}
