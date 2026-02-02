package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"
	"exchange-go/internal/service/admin"
)

// AdminHandler 管理端处理器入口
var adminService = admin.NewAdminService()

// UserListItem 用户列表项（包含余额信息）
type UserListItem struct {
	model.User
	SpotBalance     float64 `json:"spot_balance"`
	ContractBalance float64 `json:"contract_balance"`
	DeliveryBalance float64 `json:"delivery_balance"`
}

// ============ 用户管理 ============

// GetUserList 获取用户列表
func GetUserList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.User.GetUserList(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	// 查询每个用户的现货、永续合约和交割合约余额
	walletService := service.GetWalletTransferService()
	userList := make([]UserListItem, 0, len(list))
	for _, user := range list {
		item := UserListItem{
			User:            user,
			SpotBalance:     0,
			ContractBalance: 0,
			DeliveryBalance: 0,
		}

		// 获取用户钱包余额
		balance, err := walletService.GetUserAllWallets(uint64(user.ID))
		if err == nil && balance != nil {
			item.SpotBalance, _ = balance.SpotBalance.Float64()
			item.ContractBalance, _ = balance.ContractBalance.Float64()
			item.DeliveryBalance, _ = balance.DeliveryBalance.Float64()
		}

		userList = append(userList, item)
	}

	response.SuccessWithPage(c, userList, total, info.Page, info.PageSize)
}

// GetUserDetail 获取用户详情
func GetUserDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	user, err := adminService.User.GetUserInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", user)
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		AccountNumber *string `json:"account_number"`
		Phone         *string `json:"phone"`
		Email         *string `json:"email"`
		Nickname      *string `json:"nickname"`
		Risk          *int8   `json:"risk"`
		UserRemark    *string `json:"user_remark"`
		Password      *string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.AccountNumber != nil {
		updates["account_number"] = *req.AccountNumber
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.UserRemark != nil {
		updates["user_remark"] = *req.UserRemark
	}
	if req.Risk != nil {
		updates["risk"] = *req.Risk
	}
	if req.Password != nil && *req.Password != "" {
		updates["password"] = *req.Password
	}

	if err := adminService.User.UpdateUserFields(uint(id), updates); err != nil {
		response.Error(c, err.Error())
		return
	}

	user, err := adminService.User.GetUserInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功", user)
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := adminService.User.DeleteUserInfo(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// FreezeUser 冻结用户
func FreezeUser(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.User.FreezeUser(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "冻结成功")
}

// ActivateUser 激活用户
func ActivateUser(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.User.ActivateUser(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "激活成功")
}

// ResetPassword 重置密码
func ResetPassword(c *gin.Context) {
	var req struct {
		ID       uint   `json:"id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.User.ResetPassword(req.ID, req.Password); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "重置成功")
}

// AdjustBalance 余额调节（支持现货、永续合约和交割合约账户）
func AdjustBalance(c *gin.Context) {
	var req struct {
		UserID      uint    `json:"user_id" binding:"required"`
		CurrencyID  uint    `json:"currency_id" binding:"required"`
		WalletType  string  `json:"wallet_type"` // spot: 现货账户, contract: 永续合约账户, delivery: 交割合约账户
		Amount      float64 `json:"amount" binding:"required"`
		Reason      string  `json:"reason"` // 可选字段，默认为空
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 验证金额不能为0
	if req.Amount == 0 {
		response.BadRequest(c, "调整金额不能为0")
		return
	}

	// 默认现货账户
	if req.WalletType == "" {
		req.WalletType = "spot"
	}

	// 验证钱包类型
	if req.WalletType != "spot" && req.WalletType != "contract" && req.WalletType != "delivery" {
		response.BadRequest(c, "钱包类型无效，只能是 spot、contract 或 delivery")
		return
	}

	// 如果未填写原因，使用默认值
	if req.Reason == "" {
		if req.Amount > 0 {
			req.Reason = "后台充值"
		} else {
			req.Reason = "后台扣除"
		}
	}

	if err := adminService.User.AdjustBalance(req.UserID, req.CurrencyID, req.WalletType, req.Amount, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "调节成功")
}

// BatchSetRisk 批量设置风控
func BatchSetRisk(c *gin.Context) {
	var req struct {
		UserIDs   []uint `json:"user_ids" binding:"required"`
		RiskLevel int8   `json:"risk_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.User.BatchSetRisk(req.UserIDs, req.RiskLevel); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功")
}

// GetUserWallets 获取用户钱包
func GetUserWallets(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	wallets, err := adminService.User.GetUserWallets(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", wallets)
}

// SearchUsers 搜索用户
func SearchUsers(c *gin.Context) {
	var req struct {
		Keyword  string `json:"keyword"`
		RealName string `json:"real_name"`
		Status   *int   `json:"status"`
		IsReal   *int   `json:"is_real"`
		Risk     *int8  `json:"risk"`
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	list, total, err := adminService.User.SearchUsers(req.Keyword, req.RealName, req.Status, req.IsReal, req.Risk, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}
