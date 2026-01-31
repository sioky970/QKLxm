package admin

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"exchange-go/config"
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"

	"github.com/golang-jwt/jwt/v4"
)

// AuthService ?
type AuthService struct{}

// NewAuthService 
func NewAuthService() *AuthService {
	return &AuthService{}
}

// AdminClaims JWT Claims
type AdminClaims struct {
	AdminID  uint   `json:"admin_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
	IsSuper  int8   `json:"is_super"`
	jwt.RegisteredClaims
}

// LoginRequest 
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt int64        `json:"expires_at"`
	Admin     *model.Admin `json:"admin"`
}

// AdminInfoResponse ?
type AdminInfoResponse struct {
	Admin       *model.Admin           `json:"admin"`
	Role        *model.AdminRole       `json:"role"`
	Permissions []model.AdminModule    `json:"permissions"`
	Modules     []model.AdminModule    `json:"modules"`
}

// Login ?
func (s *AuthService) Login(username, password, ip string) (*LoginResponse, error) {
	// ?
	var admin model.Admin
	if err := database.DB.Where("username = ?", username).First(&admin).Error; err != nil {
		println("??", username)
		return nil, errors.New("")
	}

	//  (MD5PHP?
	hashedPassword := s.HashPassword(password)
	println(" :")
	println("   ?", username)
	println("   :", password)
	println("   MD5?", hashedPassword)
	println("   ?", admin.Password)
	println("   :", admin.Password == hashedPassword)
	
	if admin.Password != hashedPassword {
		return nil, errors.New("")
	}

	// Token
	token, expiresAt, err := s.GenerateAdminToken(&admin)
	if err != nil {
		return nil, errors.New("Token")
	}

	// 
	admin.Password = ""

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Admin:     &admin,
	}, nil
}

// GenerateAdminToken Token
func (s *AuthService) GenerateAdminToken(admin *model.Admin) (string, int64, error) {
	expiresAt := time.Now().Add(time.Duration(config.GlobalConfig.App.JwtExpire) * time.Second)
	
	// ?
	var isSuper int8 = 0
	var role model.AdminRole
	if err := database.DB.Where("id = ?", admin.RoleID).First(&role).Error; err == nil {
		isSuper = role.IsSuper
	}
	
	claims := AdminClaims{
		AdminID:  admin.ID,
		Username: admin.Username,
		RoleID:   admin.RoleID,
		IsSuper:  isSuper,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.GlobalConfig.App.Name + "_admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GlobalConfig.App.JwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ParseAdminToken Token
func (s *AuthService) ParseAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.App.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// GetAdminInfo ?
func (s *AuthService) GetAdminInfo(adminID uint) (*AdminInfoResponse, error) {
	var admin model.Admin
	if err := database.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
		return nil, errors.New("")
	}
	admin.Password = ""

	// 
	var role model.AdminRole
	database.DB.Where("id = ?", admin.RoleID).First(&role)

	// 
	var permissions []model.AdminModule
	var modules []model.AdminModule
	
	// ?
	if role.IsSuper == 1 {
		database.DB.Order("id ASC").Find(&modules)
		permissions = modules
	} else {
		// ?
		var rolePermissions []model.AdminRolePermission
		database.DB.Where("role_id = ?", admin.RoleID).Find(&rolePermissions)
		
		// 
		var moduleIDs []uint
		moduleIDSet := make(map[uint]struct{})
		var moduleList []model.AdminModule
		database.DB.Find(&moduleList)
		moduleNameToID := make(map[string]uint, len(moduleList))
		for _, m := range moduleList {
			moduleNameToID[m.Module] = m.ID
		}
		for _, rp := range rolePermissions {
			if rp.ModuleID == 0 && rp.Module != "" {
				if id, ok := moduleNameToID[rp.Module]; ok {
					rp.ModuleID = id
				}
			}
			if rp.ModuleID == 0 {
				continue
			}
			if _, exists := moduleIDSet[rp.ModuleID]; exists {
				continue
			}
			moduleIDSet[rp.ModuleID] = struct{}{}
			moduleIDs = append(moduleIDs, rp.ModuleID)
		}
		
		if len(moduleIDs) > 0 {
			database.DB.Where("id IN ?", moduleIDs).Order("id ASC").Find(&permissions)
		}
		
		// 
		database.DB.Order("id ASC").Find(&modules)
	}

	return &AdminInfoResponse{
		Admin:       &admin,
		Role:        &role,
		Permissions: permissions,
		Modules:     modules,
	}, nil
}

// ChangePassword 
func (s *AuthService) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	var admin model.Admin
	if err := database.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
		return errors.New("admin not found")
	}

	// Verify old password
	if admin.Password != s.HashPassword(oldPassword) {
		return errors.New("invalid password")
	}

	// Update password
	return database.DB.Model(&admin).Update("password", s.HashPassword(newPassword)).Error
}

// HashPassword  (PHPUsers::MakePassword?
func (s *AuthService) HashPassword(password string) string {
	// PHP?
	// $salt = 'ABCDEFG';
	// $passwordChars = str_split($password);
	// foreach ($passwordChars as $char) {
	//     $salt .= md5($char);
	// }
	// return md5($salt);
	
	salt := "ABCDEFG"
	
	// MD5?
	for _, char := range password {
		charHash := md5.Sum([]byte(string(char)))
		salt += hex.EncodeToString(charHash[:])
	}
	
	// saltMD5
	finalHash := md5.Sum([]byte(salt))
	return hex.EncodeToString(finalHash[:])
}

// Logout ?()
func (s *AuthService) Logout(adminID uint) error {
	// ?
	// 1. ?
	// 2. RedisToken?
	// 3. 
	return nil
}

// CheckPermission ?
func (s *AuthService) CheckPermission(adminID uint, module, action string) bool {
	var admin model.Admin
	if err := database.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
		return false
	}

	// 
	var role model.AdminRole
	if err := database.DB.Where("id = ?", admin.RoleID).First(&role).Error; err != nil {
		return false
	}
	if role.IsSuper == 1 {
		return true
	}

	// ?
	var count int64
	database.DB.Model(&model.AdminRolePermission{}).
		Where("role_id = ? AND module = ? AND action = ?", admin.RoleID, module, action).
		Count(&count)

	return count > 0
}

