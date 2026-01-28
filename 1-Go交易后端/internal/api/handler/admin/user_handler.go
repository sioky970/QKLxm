package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// AdminHandler 管理端处理器入口
var adminService = admin.NewAdminService()

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

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
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

// AdjustBalance 余额调节（统一USDT余额）
func AdjustBalance(c *gin.Context) {
	var req struct {
		UserID      uint    `json:"user_id" binding:"required"`
		CurrencyID  uint    `json:"currency_id" binding:"required"`
		BalanceType string  `json:"balance_type"` // 可选，兼容前端（实际不使用）
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

	// 如果未填写原因，使用默认值
	if req.Reason == "" {
		if req.Amount > 0 {
			req.Reason = "后台充值"
		} else {
			req.Reason = "后台扣除"
		}
	}

	if err := adminService.User.AdjustBalance(req.UserID, req.CurrencyID, req.BalanceType, req.Amount, req.Reason); err != nil {
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
