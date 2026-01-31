package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// ============ 系统管理 ============

// GetSettingList 获取系统设置列表
func GetSettingList(c *gin.Context) {
	list, err := adminService.System.GetSettingList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// GetSetting 获取单个设置
func GetSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := adminService.System.GetSetting(key)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", setting)
}

// UpdateSetting 更新设置
func UpdateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: key是必填项")
		return
	}

	if err := adminService.System.UpdateSetting(req.Key, req.Value); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// BatchUpdateSettings 批量更新设置
func BatchUpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.System.BatchUpdateSettings(req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// GetAdminList 获取管理员列表
func GetAdminList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.System.GetAdminList(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetAdminDetail 获取管理员详情
func GetAdminDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	admin, err := adminService.System.GetAdminInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", admin)
}

// CreateAdmin 创建管理员
func CreateAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		RoleID   uint   `json:"role_id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	adminUser := model.Admin{
		Username: req.Username,
		RoleID:   req.RoleID,
		Password: authService.HashPassword(req.Password),
	}

	if err := adminService.System.CreateAdmin(adminUser); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功")
}

// UpdateAdmin 更新管理员
func UpdateAdmin(c *gin.Context) {
	var req struct {
		ID       uint   `json:"id" binding:"required"`
		Username string `json:"username" binding:"required"`
		RoleID   uint   `json:"role_id" binding:"required"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	adminUser := model.Admin{
		ID:       req.ID,
		Username: req.Username,
		RoleID:   req.RoleID,
	}

	if req.Password != "" {
		adminUser.Password = authService.HashPassword(req.Password)
	} else {
		existing, err := adminService.System.GetAdminInfo(req.ID)
		if err != nil {
			response.Error(c, err.Error())
			return
		}
		adminUser.Password = existing.Password
	}

	if err := adminService.System.UpdateAdmin(adminUser); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteAdmin 删除管理员
func DeleteAdmin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := adminService.System.DeleteAdmin(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// GetRoleList 获取角色列表
func GetRoleList(c *gin.Context) {
	list, err := adminService.System.GetRoleList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// GetRoleDetail 获取角色详情
func GetRoleDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	role, err := adminService.System.GetRoleInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", role)
}

// CreateRole 创建角色
func CreateRole(c *gin.Context) {
	var role model.AdminRole
	if err := c.ShouldBindJSON(&role); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.System.CreateRole(role); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功")
}

// UpdateRole 更新角色
func UpdateRole(c *gin.Context) {
	var role model.AdminRole
	if err := c.ShouldBindJSON(&role); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.System.UpdateRole(role); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteRole 删除角色
func DeleteRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := adminService.System.DeleteRole(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// AssignRolePermissions 分配角色权限
func AssignRolePermissions(c *gin.Context) {
	var req struct {
		RoleID    uint   `json:"role_id" binding:"required"`
		ModuleIDs []uint `json:"module_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.System.AssignRolePermissions(req.RoleID, req.ModuleIDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "分配成功")
}

// GetLevelList 获取等级列表
func GetLevelList(c *gin.Context) {
	list, err := adminService.System.GetLevelList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// CreateLevel 创建等级
func CreateLevel(c *gin.Context) {
	var level model.Level
	if err := c.ShouldBindJSON(&level); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.System.CreateLevel(level); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功")
}

// UpdateLevel 更新等级
func UpdateLevel(c *gin.Context) {
	var level model.Level
	if err := c.ShouldBindJSON(&level); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.System.UpdateLevel(level); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteLevel 删除等级
func DeleteLevel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := adminService.System.DeleteLevel(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}
