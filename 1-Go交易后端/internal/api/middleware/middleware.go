package middleware

import (
	"strings"
	"time"

	"exchange-go/config"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Claims JWT claims 结构
type Claims struct {
	UserID uint   `json:"user_id"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		logger.Infof("%s %s %s %d %v",
			clientIP,
			method,
			path,
			statusCode,
			latency,
		)
	}
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Token
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.GetHeader("Token")
		}
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 解析 Token
		claims, err := ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}

// GenerateToken 生成JWT Token
func GenerateToken(userID uint, phone, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Phone:  phone,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.GlobalConfig.App.JwtExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.GlobalConfig.App.Name,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.App.JwtSecret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.App.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetClaims 从上下文获取完整Claims
func GetClaims(c *gin.Context) *Claims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}
	return claims.(*Claims)
}

// AdminAuth 管理员权限中间件 (检查用户是否是管理员)
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

// AdminClaims 管理员专用JWT Claims
type AdminClaims struct {
	AdminID  uint   `json:"admin_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
	IsSuper  int8   `json:"is_super"`
	jwt.RegisteredClaims
}

// AdminJWTAuth 管理员专用JWT认证中间件
func AdminJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Token
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.GetHeader("Token")
		}
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 解析管理员 Token
		claims, err := ParseAdminToken(token)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		// 将管理员信息存入上下文
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_role_id", claims.RoleID)
		c.Set("admin_is_super", claims.IsSuper)
		c.Set("admin_claims", claims)

		c.Next()
	}
}

// ParseAdminToken 解析管理员JWT Token
func ParseAdminToken(tokenString string) (*AdminClaims, error) {
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

// GetAdminID 从上下文获取管理员ID
func GetAdminID(c *gin.Context) uint {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return 0
	}
	return adminID.(uint)
}

// GetAdminClaims 从上下文获取管理员Claims
func GetAdminClaims(c *gin.Context) *AdminClaims {
	claims, exists := c.Get("admin_claims")
	if !exists {
		return nil
	}
	return claims.(*AdminClaims)
}

// IsAdminSuper 检查是否是超级管理员
func IsAdminSuper(c *gin.Context) bool {
	isSuper, exists := c.Get("admin_is_super")
	if !exists {
		return false
	}
	return isSuper.(int8) == 1
}
