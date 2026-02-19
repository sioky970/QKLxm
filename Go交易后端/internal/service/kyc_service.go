package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// KYCService 实名认证服务
type KYCService struct{}

// NewKYCService 创建实名认证服务实例
func NewKYCService() *KYCService {
	return &KYCService{}
}

// KYCSubmitRequest 实名认证提交请求
type KYCSubmitRequest struct {
	Name       string `json:"name" binding:"required"`
	CardID     string `json:"card_id" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	BankCard   string `json:"bank_card" binding:"required"`
	FrontPic   string `json:"front_pic" binding:"required"`
	ReversePic string `json:"reverse_pic" binding:"required"`
	HandPic    string `json:"hand_pic"`
	BankPic    string `json:"bank_pic"`
}

// KYCStatusResponse 实名认证状态响应
type KYCStatusResponse struct {
	ID           uint   `json:"id"`
	UserID       int    `json:"user_id"`
	Name         string `json:"name"`
	CardID       string `json:"card_id_masked"`
	Phone        string `json:"phone"`
	BankCard     string `json:"bank_card"`     // 银行卡号（脱敏）
	ReviewStatus int8   `json:"review_status"` // 0:初始 1:待审核 2:已通过 3:已拒绝
	RejectReason string `json:"reject_reason"` // 拒绝原因
	SubmitTime   int64  `json:"submit_time"`
	ReviewTime   int64  `json:"review_time"`
}

// SubmitKYC 提交实名认证
func (s *KYCService) SubmitKYC(userID uint, req *KYCSubmitRequest) error {
	if !s.validateIDCard(req.CardID) {
		return errors.New("身份证号格式不正确")
	}

	if !s.validateName(req.Name) {
		return errors.New("姓名格式不正确")
	}

	if !s.validatePhone(req.Phone) {
		return errors.New("手机号码格式不正确")
	}

	// 银行卡号只检查是否为空，不做格式校验
	if req.BankCard == "" {
		return errors.New("请输入银行卡号")
	}

	var existingKYC model.UserReal
	result := database.DB.Where("user_id = ?", userID).First(&existingKYC)

	if result.Error == nil {
		if existingKYC.ReviewStatus == 2 {
			return errors.New("您已通过实名认证")
		}
		if existingKYC.ReviewStatus == 1 {
			return errors.New("您的实名认证正在审核中，请勿重复提交")
		}
		// 防重复提交：被拒绝后需等待5分钟才能重新提交
		if existingKYC.ReviewStatus == 3 {
			cooldownSeconds := int64(300) // 5分钟冷却时间
			if time.Now().Unix()-existingKYC.CreateTime < cooldownSeconds {
				remaining := cooldownSeconds - (time.Now().Unix() - existingKYC.CreateTime)
				return fmt.Errorf("请等待%d秒后再重新提交", remaining)
			}
		}
	}

	var count int64
	database.DB.Model(&model.UserReal{}).
		Where("card_id = ? AND user_id != ? AND review_status = 2", req.CardID, userID).
		Count(&count)
	if count > 0 {
		return errors.New("该身份证号已被其他用户使用")
	}

	kyc := &model.UserReal{
		UserID:       userID,
		Name:         req.Name,
		CardID:       req.CardID,
		Phone:        req.Phone,
		BankCard:     req.BankCard,
		FrontPic:     req.FrontPic,
		ReversePic:   req.ReversePic,
		HandPic:      req.HandPic,
		BankPic:      req.BankPic,
		ReviewStatus: 1,
		CreateTime:   time.Now().Unix(),
	}

	if result.Error == nil {
		kyc.ID = existingKYC.ID
		if err := database.DB.Save(kyc).Error; err != nil {
			return fmt.Errorf("更新实名认证失败: %w", err)
		}
	} else {
		if err := database.DB.Create(kyc).Error; err != nil {
			return fmt.Errorf("提交实名认证失败: %w", err)
		}
	}

	logger.Infof("用户提交实名认证: userID=%d, name=%s", userID, req.Name)
	return nil
}

// GetKYCStatus 获取实名认证状态
func (s *KYCService) GetKYCStatus(userID uint) (*KYCStatusResponse, error) {
	var kyc model.UserReal
	result := database.DB.Where("user_id = ?", userID).First(&kyc)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("未找到实名认证记录")
		}
		return nil, result.Error
	}

	maskedCardID := s.maskIDCard(kyc.CardID)

	response := &KYCStatusResponse{
		ID:           kyc.ID,
		UserID:       int(kyc.UserID),
		Name:         kyc.Name,
		CardID:       maskedCardID,
		Phone:        kyc.Phone,
		BankCard:     kyc.BankCard, // 返回完整银行卡号用于提现
		ReviewStatus: kyc.ReviewStatus,
		RejectReason: kyc.RejectReason,
		SubmitTime:   kyc.CreateTime,
		ReviewTime:   kyc.ReviewTime,
	}

	return response, nil
}

// ReviewKYC 审核实名认证 (管理员功能)
func (s *KYCService) ReviewKYC(kycID uint, approved bool, rejectReason string) error {
	var kyc model.UserReal
	result := database.DB.First(&kyc, kycID)
	if result.Error != nil {
		return errors.New("实名认证记录不存在")
	}

	if kyc.ReviewStatus == 2 {
		return errors.New("该认证已审核通过，无法重复审核")
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新审核状态：1=待审核, 2=通过, 3=拒绝
	if approved {
		kyc.ReviewStatus = 2
		kyc.RejectReason = "" // 清空拒绝原因
	} else {
		kyc.ReviewStatus = 3 // 拒绝状态为3
		kyc.RejectReason = rejectReason
	}
	kyc.ReviewTime = time.Now().Unix()

	if err := tx.Save(&kyc).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新审核状态失败: %w", err)
	}

	// 审核通过后同步更新User表的is_realname字段
	if approved {
		if err := tx.Model(&model.User{}).Where("id = ?", kyc.UserID).
			Update("is_realname", 2).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新用户实名状态失败: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	status := "通过"
	if !approved {
		status = "拒绝"
	}
	logger.Infof("实名认证审核完成: kycID=%d, userID=%d, status=%s", kycID, kyc.UserID, status)

	return nil
}

// GetKYCList 获取实名认证列表 (管理员功能)
func (s *KYCService) GetKYCList(status int8, page, pageSize int) ([]model.UserReal, int64, error) {
	var kycList []model.UserReal
	var total int64

	query := database.DB.Model(&model.UserReal{})
	if status >= 0 {
		query = query.Where("review_status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&kycList)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return kycList, total, nil
}

// validateIDCard 验证身份证号（基本格式检查）
func (s *KYCService) validateIDCard(cardID string) bool {
	// 只检查基本格式：15位或18位，18位最后一位可以是X
	pattern := `^(\d{15}|\d{17}[\dXx])$`
	matched, err := regexp.MatchString(pattern, cardID)
	return err == nil && matched
}

// validateIDCardChecksum 验证身份证校验位
func (s *KYCService) validateIDCardChecksum(cardID string) bool {
	if len(cardID) != 18 {
		return false
	}

	weight := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCode := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		digit := int(cardID[i] - '0')
		sum += digit * weight[i]
	}

	mod := sum % 11
	expectedCheck := checkCode[mod]
	actualCheck := cardID[17]

	if actualCheck == 'x' {
		actualCheck = 'X'
	}

	return actualCheck == expectedCheck
}

// validateName 验证姓名
func (s *KYCService) validateName(name string) bool {
	pattern := `^[\p{Han}]{2,20}$`
	matched, err := regexp.MatchString(pattern, name)
	return err == nil && matched
}

// validatePhone 验证手机号格式
func (s *KYCService) validatePhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, err := regexp.MatchString(pattern, phone)
	return err == nil && matched
}

// validateBankCard 验证银行卡号格式和Luhn校验
func (s *KYCService) validateBankCard(card string) bool {
	// 长度检查: 16-19位
	if len(card) < 16 || len(card) > 19 {
		return false
	}

	// 纯数字检查
	for _, c := range card {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Luhn算法校验
	return s.luhnCheck(card)
}

// luhnCheck Luhn算法校验银行卡号
func (s *KYCService) luhnCheck(cardNumber string) bool {
	n := len(cardNumber)
	sum := 0
	isSecond := false

	for i := n - 1; i >= 0; i-- {
		d := int(cardNumber[i] - '0')

		if isSecond {
			d = d * 2
			if d > 9 {
				d = d - 9
			}
		}

		sum += d
		isSecond = !isSecond
	}

	return sum%10 == 0
}

// maskIDCard 脱敏身份证号
func (s *KYCService) maskIDCard(cardID string) string {
	if len(cardID) < 10 {
		return cardID
	}
	return cardID[:6] + "********" + cardID[len(cardID)-4:]
}

// CheckUserKYCStatus 检查用户是否已完成实名认证（已通过审核）
func (s *KYCService) CheckUserKYCStatus(userID uint) (bool, error) {
	var kyc model.UserReal
	result := database.DB.Where("user_id = ? AND review_status = 2", userID).First(&kyc)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}
