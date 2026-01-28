package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service"
)

// ============ 实名认证管理 ============

var kycSrv = service.NewKYCService()

// GetKYCList 获取实名认证列表
// @Summary 获取实名认证列表
// @Tags 管理端-KYC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body object{page=int,page_size=int,account=string,status=int} true "查询参数"
// @Success 200 {object} response.Response
// @Router /api/admin/kyc/list [post]
func GetKYCList(c *gin.Context) {
	var req struct {
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
		Account  string `json:"account"`
		Status   *int   `json:"status"` // nil: 全部, 0: 待审核, 1: 已通过, 2: 已拒绝
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

	if list, total, err := GetKYCListAdvanced(req.Account, req.Status, req.Page, req.PageSize); err != nil {
		response.Error(c, err.Error())
		return
	} else {
		response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
	}
}

// KYCListItem 实名认证列表项（带用户信息）
type KYCListItem struct {
	model.UserReal
	AccountNumber string `json:"account_number"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
}

// GetKYCListAdvanced 获取实名认证列表（带高级筛选）
func GetKYCListAdvanced(account string, status *int, page, pageSize int) (list []KYCListItem, total int64, err error) {
	db := database.DB.Table("user_real AS ur").
		Select("ur.*, u.account_number, u.phone, u.email").
		Joins("LEFT JOIN users u ON ur.user_id = u.id")

	if account != "" {
		db = db.Where("u.account_number LIKE ? OR u.phone LIKE ? OR u.email LIKE ?",
			"%"+account+"%", "%"+account+"%", "%"+account+"%")
	}
	if status != nil {
		db = db.Where("ur.review_status = ?", *status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	offset := pageSize * (page - 1)
	err = db.Limit(pageSize).Offset(offset).Order("ur.id DESC").Find(&list).Error
	return
}

// GetKYCDetail 获取实名认证详情
// @Summary 获取实名认证详情
// @Tags 管理端-KYC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "KYC ID"
// @Success 200 {object} response.Response
// @Router /api/admin/kyc/{id} [get]
func GetKYCDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var kyc model.UserReal
	if err := database.DB.Where("id = ?", id).First(&kyc).Error; err != nil {
		response.Error(c, "记录不存在")
		return
	}

	// 获取用户信息
	var user model.User
	database.DB.Select("id, account_number, phone, email").Where("id = ?", kyc.UserID).First(&user)

	result := map[string]interface{}{
		"kyc":  kyc,
		"user": user,
	}

	response.Success(c, "获取成功", result)
}

// ReviewKYC 审核实名认证
// @Summary 审核实名认证
// @Tags 管理端-KYC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body object{id=int,approved=bool,reason=string} true "审核参数"
// @Success 200 {object} response.Response
// @Router /api/admin/kyc/review [post]
func ReviewKYC(c *gin.Context) {
	var req struct {
		ID       uint   `json:"id" binding:"required"`
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := kycSrv.ReviewKYC(req.ID, req.Approved, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	msg := "审核通过"
	if !req.Approved {
		msg = "已拒绝"
	}
	response.Success(c, msg)
}

// ApproveKYC 通过实名认证
// @Summary 通过实名认证
// @Tags 管理端-KYC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body object{id=int} true "审核参数"
// @Success 200 {object} response.Response
// @Router /api/admin/kyc/approve [post]
func ApproveKYC(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := kycSrv.ReviewKYC(req.ID, true, ""); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "审核通过")
}

// RejectKYC 拒绝实名认证
// @Summary 拒绝实名认证
// @Tags 管理端-KYC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body object{id=int,reason=string} true "审核参数"
// @Success 200 {object} response.Response
// @Router /api/admin/kyc/reject [post]
func RejectKYC(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写拒绝原因")
		return
	}

	if err := kycSrv.ReviewKYC(req.ID, false, req.Reason); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "已拒绝")
}

// DeleteKYC 删除实名认证记录
// @Summary 删除实名认证记录
// @Tags 管理端-KYC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "KYC ID"
// @Success 200 {object} response.Response
// @Router /api/admin/kyc/{id} [delete]
func DeleteKYC(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := database.DB.Delete(&model.UserReal{}, id).Error; err != nil {
		response.Error(c, "删除失败")
		return
	}

	response.Success(c, "删除成功")
}
