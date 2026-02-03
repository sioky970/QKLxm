package admin

import (
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// SystemService
type SystemService struct{}

// NewSystemService
func NewSystemService() *SystemService {
	return &SystemService{}
}

// GetSettingList
func (s *SystemService) GetSettingList() (list []model.Setting, err error) {
	err = database.DB.Find(&list).Error
	return
}

// GetSetting
func (s *SystemService) GetSetting(key string) (setting *model.Setting, err error) {
	var st model.Setting
	err = database.DB.Where("`key` = ?", key).First(&st).Error
	return &st, err
}

// UpdateSetting
func (s *SystemService) UpdateSetting(key, value string) error {
	var setting model.Setting
	err := database.DB.Where("`key` = ?", key).First(&setting).Error
	if err != nil {
		// 如果不存在则创建
		setting = model.Setting{
			Key:   key,
			Value: value,
		}
		return database.DB.Create(&setting).Error
	}
	// 如果存在则更新
	return database.DB.Model(&setting).Update("value", value).Error
}

// BatchUpdateSettings
func (s *SystemService) BatchUpdateSettings(settings map[string]string) error {
	for key, value := range settings {
		if err := s.UpdateSetting(key, value); err != nil {
			return err
		}
	}
	return nil
}

// CreateSetting
func (s *SystemService) CreateSetting(setting model.Setting) error {
	return database.DB.Create(&setting).Error
}

// DeleteSetting
func (s *SystemService) DeleteSetting(id uint) error {
	return database.DB.Delete(&model.Setting{}, id).Error
}

// GetAdminList ?
func (s *SystemService) GetAdminList(info PageInfo) (list []model.Admin, total int64, err error) {
	db := database.DB.Model(&model.Admin{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetAdminInfo ?
func (s *SystemService) GetAdminInfo(id uint) (admin *model.Admin, err error) {
	var a model.Admin
	err = database.DB.Where("id = ?", id).First(&a).Error
	return &a, err
}

// CreateAdmin ?
func (s *SystemService) CreateAdmin(admin model.Admin) error {
	return database.DB.Create(&admin).Error
}

// UpdateAdmin ?
func (s *SystemService) UpdateAdmin(admin model.Admin) error {
	return database.DB.Save(&admin).Error
}

// DeleteAdmin ?
func (s *SystemService) DeleteAdmin(id uint) error {
	return database.DB.Delete(&model.Admin{}, id).Error
}

// GetRoleList
func (s *SystemService) GetRoleList() (list []model.AdminRole, err error) {
	err = database.DB.Find(&list).Error
	return
}

// GetRoleInfo
func (s *SystemService) GetRoleInfo(id uint) (role *model.AdminRole, err error) {
	var r model.AdminRole
	err = database.DB.Where("id = ?", id).First(&r).Error
	return &r, err
}

// CreateRole
func (s *SystemService) CreateRole(role model.AdminRole) error {
	return database.DB.Create(&role).Error
}

// UpdateRole
func (s *SystemService) UpdateRole(role model.AdminRole) error {
	return database.DB.Save(&role).Error
}

// DeleteRole
func (s *SystemService) DeleteRole(id uint) error {
	return database.DB.Delete(&model.AdminRole{}, id).Error
}

// GetModuleList
func (s *SystemService) GetModuleList() (list []model.AdminModule, err error) {
	err = database.DB.Order("sort ASC").Find(&list).Error
	return
}

// GetRolePermissions
func (s *SystemService) GetRolePermissions(roleID uint) (list []model.AdminRolePermission, err error) {
	err = database.DB.Where("role_id = ?", roleID).Find(&list).Error
	return
}

// AssignRolePermissions
func (s *SystemService) AssignRolePermissions(roleID uint, moduleIDs []uint) error {
	// ?
	if err := database.DB.Where("role_id = ?", roleID).Delete(&model.AdminRolePermission{}).Error; err != nil {
		return err
	}

	// ?
	var moduleList []model.AdminModule
	database.DB.Find(&moduleList)
	moduleIDToName := make(map[uint]string, len(moduleList))
	for _, m := range moduleList {
		moduleIDToName[m.ID] = m.Module
	}
	for _, moduleID := range moduleIDs {
		permission := model.AdminRolePermission{
			RoleID:   roleID,
			ModuleID: moduleID,
			Module:   moduleIDToName[moduleID],
		}
		if err := database.DB.Create(&permission).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetLevelList
func (s *SystemService) GetLevelList() (list []model.Level, err error) {
	err = database.DB.Order("level ASC").Find(&list).Error
	return
}

// CreateLevel
func (s *SystemService) CreateLevel(level model.Level) error {
	return database.DB.Create(&level).Error
}

// UpdateLevel
func (s *SystemService) UpdateLevel(level model.Level) error {
	return database.DB.Save(&level).Error
}

// DeleteLevel
func (s *SystemService) DeleteLevel(id uint) error {
	return database.DB.Delete(&model.Level{}, id).Error
}
