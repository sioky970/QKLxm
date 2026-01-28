package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// DepositService 充值服务
type DepositService struct{}

var depositServiceInstance *DepositService

// GetDepositService 获取充值服务实例
func GetDepositService() *DepositService {
	if depositServiceInstance == nil {
		depositServiceInstance = &DepositService{}
	}
	return depositServiceInstance
}

// 业务常量
const (
	MinDepositAmount   = 10.0                // 最小充值金额 10 USDT
	MaxScreenshotSize  = 20 * 1024 * 1024    // 最大截图大小 20MB
	DepositExpireHours = 24                  // 订单过期时间 24小时
	UploadDir          = "./uploads/deposit" // 截图上传目录
)

// 允许的图片格式
var AllowedImageTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// ============================================================
// 用户端服务
// ============================================================

// GetDepositAddresses 获取启用的充值地址列表
func (s *DepositService) GetDepositAddresses() ([]model.DepositAddress, error) {
	var addresses []model.DepositAddress
	err := database.DB.Where("status = ? AND address != ''", model.DepositAddressEnabled).
		Order("sort DESC, id ASC").
		Find(&addresses).Error
	if err != nil {
		logger.Errorf("[Deposit] 获取充值地址失败: %v", err)
		return nil, err
	}
	return addresses, nil
}

// CreateDepositOrderRequest 创建充值订单请求
type CreateDepositOrderRequest struct {
	Network string  `json:"network" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
}

// CreateDepositOrder 创建充值订单
func (s *DepositService) CreateDepositOrder(userID uint, req *CreateDepositOrderRequest) (*model.DepositOrder, error) {
	// 验证金额
	if req.Amount < MinDepositAmount {
		return nil, fmt.Errorf("充值金额不能小于 %.2f USDT", MinDepositAmount)
	}

	// 获取对应网络的充值地址
	var address model.DepositAddress
	err := database.DB.Where("network = ? AND status = ?", req.Network, model.DepositAddressEnabled).First(&address).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("该网络暂不支持充值")
		}
		return nil, err
	}

	// 检查是否有未完成的相同金额订单（防重复）
	var existingOrder model.DepositOrder
	err = database.DB.Where("user_id = ? AND amount = ? AND status = ?",
		userID, req.Amount, model.DepositStatusPending).First(&existingOrder).Error
	if err == nil {
		return nil, errors.New("您有一笔相同金额的充值订单正在处理中")
	}

	now := time.Now().Unix()
	order := &model.DepositOrder{
		OrderNo:    s.generateOrderNo(),
		UserID:     userID,
		Network:    req.Network,
		Address:    address.Address,
		Amount:     req.Amount,
		Status:     model.DepositStatusPending,
		CreateTime: now,
		UpdateTime: now,
		ExpireTime: now + DepositExpireHours*3600,
	}

	if err := database.DB.Create(order).Error; err != nil {
		logger.Errorf("[Deposit] 创建充值订单失败: userID=%d, err=%v", userID, err)
		return nil, err
	}

	logger.Infof("[Deposit] 创建充值订单成功: orderNo=%s, userID=%d, amount=%.2f, network=%s",
		order.OrderNo, userID, req.Amount, req.Network)
	return order, nil
}

// UploadScreenshot 上传转账截图
func (s *DepositService) UploadScreenshot(userID uint, orderID uint, file *multipart.FileHeader) (string, error) {
	// 查找订单
	var order model.DepositOrder
	err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("订单不存在")
		}
		return "", err
	}

	// 检查订单状态
	if order.Status != model.DepositStatusPending {
		return "", errors.New("订单状态不允许上传截图")
	}

	// 检查是否过期
	if time.Now().Unix() > order.ExpireTime {
		// 更新订单状态为已过期
		database.DB.Model(&order).Updates(map[string]interface{}{
			"status":      model.DepositStatusExpired,
			"update_time": time.Now().Unix(),
		})
		return "", errors.New("订单已过期")
	}

	// 验证文件大小
	if file.Size > MaxScreenshotSize {
		return "", fmt.Errorf("文件大小不能超过 %dMB", MaxScreenshotSize/1024/1024)
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !s.isAllowedImageType(ext) {
		return "", errors.New("不支持的图片格式，请上传 jpg/jpeg/png/gif/webp 格式")
	}

	// 确保上传目录存在
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		logger.Errorf("[Deposit] 创建上传目录失败: %v", err)
		return "", errors.New("服务器错误")
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%s_%d%s", userID, order.OrderNo, time.Now().UnixNano(), ext)
	filePath := filepath.Join(UploadDir, filename)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		logger.Errorf("[Deposit] 创建文件失败: %v", err)
		return "", errors.New("服务器错误")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		logger.Errorf("[Deposit] 保存文件失败: %v", err)
		return "", errors.New("服务器错误")
	}

	// 更新订单截图路径
	relativePath := "/uploads/deposit/" + filename
	err = database.DB.Model(&order).Updates(map[string]interface{}{
		"screenshot":  relativePath,
		"update_time": time.Now().Unix(),
	}).Error
	if err != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		return "", err
	}

	logger.Infof("[Deposit] 上传截图成功: orderNo=%s, path=%s", order.OrderNo, relativePath)
	return relativePath, nil
}

// GetUserDepositOrdersRequest 获取用户充值记录请求
type GetUserDepositOrdersRequest struct {
	Status   int `form:"status" json:"status"` // -1全部 0待审核 1已通过 2已拒绝 3已取消 4已过期
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// GetUserDepositOrders 获取用户充值记录
func (s *DepositService) GetUserDepositOrders(userID uint, req *GetUserDepositOrdersRequest) ([]model.DepositOrder, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := database.DB.Model(&model.DepositOrder{}).Where("user_id = ?", userID)

	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var orders []model.DepositOrder
	err := query.Order("create_time DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// CancelDepositOrder 用户取消充值订单
func (s *DepositService) CancelDepositOrder(userID uint, orderID uint) error {
	var order model.DepositOrder
	err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("订单不存在")
		}
		return err
	}

	if order.Status != model.DepositStatusPending {
		return errors.New("只能取消待审核的订单")
	}

	err = database.DB.Model(&order).Updates(map[string]interface{}{
		"status":      model.DepositStatusCanceled,
		"update_time": time.Now().Unix(),
	}).Error
	if err != nil {
		return err
	}

	logger.Infof("[Deposit] 用户取消订单: orderNo=%s, userID=%d", order.OrderNo, userID)
	return nil
}

// ============================================================
// 管理端服务
// ============================================================

// AdminGetDepositOrdersRequest 管理员获取充值订单请求
type AdminGetDepositOrdersRequest struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	OrderNo  string `form:"order_no" json:"order_no"`
	Network  string `form:"network" json:"network"`
	Status   int    `form:"status" json:"status"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

// AdminGetDepositOrders 管理员获取充值订单列表
func (s *DepositService) AdminGetDepositOrders(req *AdminGetDepositOrdersRequest) ([]map[string]interface{}, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := database.DB.Model(&model.DepositOrder{})

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+req.OrderNo+"%")
	}
	if req.Network != "" {
		query = query.Where("network = ?", req.Network)
	}
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var orders []model.DepositOrder
	err := query.Order("create_time DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取用户信息
	userIDs := make([]uint, 0)
	for _, order := range orders {
		userIDs = append(userIDs, order.UserID)
	}

	var users []model.User
	if len(userIDs) > 0 {
		database.DB.Where("id IN ?", userIDs).Find(&users)
	}
	userMap := make(map[uint]model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 组装返回数据
	result := make([]map[string]interface{}, 0)
	for _, order := range orders {
		item := map[string]interface{}{
			"id":           order.ID,
			"order_no":     order.OrderNo,
			"user_id":      order.UserID,
			"network":      order.Network,
			"address":      order.Address,
			"amount":       order.Amount,
			"screenshot":   order.Screenshot,
			"status":       order.Status,
			"admin_id":     order.AdminID,
			"admin_remark": order.AdminRemark,
			"create_time":  order.CreateTime,
			"update_time":  order.UpdateTime,
			"review_time":  order.ReviewTime,
			"expire_time":  order.ExpireTime,
		}
		if user, ok := userMap[order.UserID]; ok {
			item["username"] = user.Phone
			item["email"] = user.Email
		}
		result = append(result, item)
	}

	return result, total, nil
}

// AdminGetDepositOrderDetail 管理员获取充值订单详情
func (s *DepositService) AdminGetDepositOrderDetail(orderID uint) (*model.DepositOrder, error) {
	var order model.DepositOrder
	err := database.DB.First(&order, orderID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}
	return &order, nil
}

// AdminApproveDepositRequest 审核通过请求
type AdminApproveDepositRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Remark  string `json:"remark"`
}

// AdminApproveDeposit 管理员审核通过充值订单
func (s *DepositService) AdminApproveDeposit(adminID uint, req *AdminApproveDepositRequest) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var order model.DepositOrder
		err := tx.First(&order, req.OrderID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("订单不存在")
			}
			return err
		}

		if order.Status != model.DepositStatusPending {
			return errors.New("只能审核待审核的订单")
		}

		// 检查是否已上传截图
		if order.Screenshot == "" {
			return errors.New("用户尚未上传转账截图")
		}

		now := time.Now().Unix()

		// 更新订单状态
		err = tx.Model(&order).Updates(map[string]interface{}{
			"status":       model.DepositStatusApproved,
			"admin_id":     adminID,
			"admin_remark": req.Remark,
			"update_time":  now,
			"review_time":  now,
		}).Error
		if err != nil {
			return err
		}

		// 增加用户USDT余额
		err = s.addUserBalance(tx, order.UserID, order.Amount)
		if err != nil {
			return fmt.Errorf("增加用户余额失败: %v", err)
		}

		logger.Infof("[Deposit] 审核通过: orderNo=%s, userID=%d, amount=%.2f, adminID=%d",
			order.OrderNo, order.UserID, order.Amount, adminID)
		return nil
	})
}

// addUserBalance 增加用户USDT余额（在事务中）
func (s *DepositService) addUserBalance(tx *gorm.DB, userID uint, amount float64) error {
	var assets model.UserAssets
	err := tx.Where("user_id = ?", userID).First(&assets).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户资产不存在，创建
			assets = model.UserAssets{
				UserID:           userID,
				UsdtBalance:      amount,
				UsdtLocked:       0,
				CurrencyBalances: make(model.CurrencyBalance),
				CurrencyLocked:   make(model.CurrencyBalance),
				CreateTime:       time.Now().Unix(),
				UpdateTime:       time.Now().Unix(),
			}
			return tx.Create(&assets).Error
		}
		return err
	}

	// 更新余额
	return tx.Model(&assets).Updates(map[string]interface{}{
		"usdt_balance":     gorm.Expr("usdt_balance + ?", amount),
		"update_time":      time.Now().Unix(),
		"last_update_time": time.Now().Unix(),
		"version":          gorm.Expr("version + 1"),
	}).Error
}

// AdminRejectDepositRequest 审核拒绝请求
type AdminRejectDepositRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Remark  string `json:"remark" binding:"required"`
}

// AdminRejectDeposit 管理员拒绝充值订单
func (s *DepositService) AdminRejectDeposit(adminID uint, req *AdminRejectDepositRequest) error {
	var order model.DepositOrder
	err := database.DB.First(&order, req.OrderID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("订单不存在")
		}
		return err
	}

	if order.Status != model.DepositStatusPending {
		return errors.New("只能审核待审核的订单")
	}

	now := time.Now().Unix()
	err = database.DB.Model(&order).Updates(map[string]interface{}{
		"status":       model.DepositStatusRejected,
		"admin_id":     adminID,
		"admin_remark": req.Remark,
		"update_time":  now,
		"review_time":  now,
	}).Error
	if err != nil {
		return err
	}

	logger.Infof("[Deposit] 审核拒绝: orderNo=%s, userID=%d, adminID=%d, remark=%s",
		order.OrderNo, order.UserID, adminID, req.Remark)
	return nil
}

// ============================================================
// 充值地址管理
// ============================================================

// GetAllDepositAddresses 获取所有充值地址（管理员）
func (s *DepositService) GetAllDepositAddresses() ([]model.DepositAddress, error) {
	var addresses []model.DepositAddress
	err := database.DB.Order("sort DESC, id ASC").Find(&addresses).Error
	return addresses, err
}

// CreateDepositAddressRequest 创建充值地址请求
type CreateDepositAddressRequest struct {
	Network string `json:"network" binding:"required"`
	Address string `json:"address" binding:"required"`
	QrCode  string `json:"qr_code"`
	Status  int8   `json:"status"`
	Sort    int    `json:"sort"`
}

// CreateDepositAddress 创建充值地址
func (s *DepositService) CreateDepositAddress(req *CreateDepositAddressRequest) (*model.DepositAddress, error) {
	// 检查是否已存在相同网络的地址
	var existing model.DepositAddress
	err := database.DB.Where("network = ?", req.Network).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("网络 %s 的充值地址已存在", req.Network)
	}

	now := time.Now().Unix()
	address := &model.DepositAddress{
		Network:    req.Network,
		Address:    req.Address,
		QrCode:     req.QrCode,
		Status:     req.Status,
		Sort:       req.Sort,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := database.DB.Create(address).Error; err != nil {
		return nil, err
	}

	logger.Infof("[Deposit] 创建充值地址: network=%s, address=%s", req.Network, req.Address)
	return address, nil
}

// UpdateDepositAddressRequest 更新充值地址请求
type UpdateDepositAddressRequest struct {
	Address string `json:"address"`
	QrCode  string `json:"qr_code"`
	Status  *int8  `json:"status"`
	Sort    *int   `json:"sort"`
}

// UpdateDepositAddress 更新充值地址
func (s *DepositService) UpdateDepositAddress(id uint, req *UpdateDepositAddressRequest) error {
	updates := map[string]interface{}{
		"update_time": time.Now().Unix(),
	}

	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.QrCode != "" {
		updates["qr_code"] = req.QrCode
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}

	result := database.DB.Model(&model.DepositAddress{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("充值地址不存在")
	}

	logger.Infof("[Deposit] 更新充值地址: id=%d", id)
	return nil
}

// DeleteDepositAddress 删除充值地址
func (s *DepositService) DeleteDepositAddress(id uint) error {
	result := database.DB.Delete(&model.DepositAddress{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("充值地址不存在")
	}

	logger.Infof("[Deposit] 删除充值地址: id=%d", id)
	return nil
}

// ============================================================
// 辅助方法
// ============================================================

// generateOrderNo 生成订单号
func (s *DepositService) generateOrderNo() string {
	return fmt.Sprintf("D%s%06d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/1000)
}

// isAllowedImageType 检查是否是允许的图片类型
func (s *DepositService) isAllowedImageType(ext string) bool {
	for _, t := range AllowedImageTypes {
		if t == ext {
			return true
		}
	}
	return false
}

// CheckExpiredOrders 检查并更新过期订单（可用于定时任务）
func (s *DepositService) CheckExpiredOrders() error {
	now := time.Now().Unix()
	result := database.DB.Model(&model.DepositOrder{}).
		Where("status = ? AND expire_time < ?", model.DepositStatusPending, now).
		Updates(map[string]interface{}{
			"status":      model.DepositStatusExpired,
			"update_time": now,
		})
	if result.Error != nil {
		logger.Errorf("[Deposit] 更新过期订单失败: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected > 0 {
		logger.Infof("[Deposit] 更新 %d 个过期订单", result.RowsAffected)
	}
	return nil
}
