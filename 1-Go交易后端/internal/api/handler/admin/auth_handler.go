package admin

import (
	"strings"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/api/middleware"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// 认证服务实例
var authService = admin.NewAuthService()

// ============ 管理员认证 ============

// AdminLogin 管理员登录
// @Summary 管理员登录
// @Tags 管理端-认证
// @Accept json
// @Produce json
// @Param data body admin.LoginRequest true "登录参数"
// @Success 200 {object} response.Response{data=admin.LoginResponse}
// @Router /api/admin/login [post]
func AdminLogin(c *gin.Context) {
	var req admin.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 获取客户端IP
	ip := c.ClientIP()

	result, err := authService.Login(req.Username, req.Password, ip)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "登录成功", result)
}

// AdminLogout 管理员退出
// @Summary 管理员退出登录
// @Tags 管理端-认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Router /api/admin/logout [post]
func AdminLogout(c *gin.Context) {
	adminID := GetAdminID(c)
	if adminID > 0 {
		authService.Logout(adminID)
	}

	response.Success(c, "退出成功")
}

// AdminInfo 获取当前管理员信息
// @Summary 获取当前管理员信息
// @Tags 管理端-认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=admin.AdminInfoResponse}
// @Router /api/admin/info [get]
func AdminInfo(c *gin.Context) {
	adminID := GetAdminID(c)
	if adminID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	info, err := authService.GetAdminInfo(adminID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", info)
}

// AdminChangePassword 修改密码
// @Summary 管理员修改密码
// @Tags 管理端-认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body object{old_password=string,new_password=string} true "密码参数"
// @Success 200 {object} response.Response
// @Router /api/admin/change-password [post]
func AdminChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误，密码长度至少6位")
		return
	}

	adminID := GetAdminID(c)
	if adminID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	if err := authService.ChangePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "密码修改成功")
}

// ============ 辅助函数 ============

// GetAdminID 从上下文获取管理员ID
func GetAdminID(c *gin.Context) uint {
	return middleware.GetAdminID(c)
}

// GetAdminClaims 从上下文获取管理员Claims
func GetAdminClaims(c *gin.Context) *middleware.AdminClaims {
	return middleware.GetAdminClaims(c)
}

// IsSuper 检查是否是超级管理员
func IsSuper(c *gin.Context) bool {
	return middleware.IsAdminSuper(c)
}

// ParseAdminToken 解析管理员Token (供中间件使用)
func ParseAdminToken(tokenString string) (*admin.AdminClaims, error) {
	// 去掉 Bearer 前缀
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}
	return authService.ParseAdminToken(tokenString)
}
