package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"exchange-go/internal/service"

	"github.com/gin-gonic/gin"
)

var kycService = service.NewKYCService()

// UploadFile 通用文件上传
func UploadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type":    "error",
			"message": "未授权",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":    "error",
			"message": "请选择要上传的文件",
		})
		return
	}

	// 检查文件大小（最大5MB）
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":    "error",
			"message": "文件大小不能超过5MB",
		})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":    "error",
			"message": "不支持的文件格式，请上传 jpg/jpeg/png/gif/webp 格式",
		})
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/kyc"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":    "error",
			"message": "创建上传目录失败",
		})
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%d%s", userID.(uint), time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":    "error",
			"message": "保存文件失败",
		})
		return
	}

	// 返回文件URL
	fileURL := "/uploads/kyc/" + filename
	c.JSON(http.StatusOK, gin.H{
		"type":    "ok",
		"message": "上传成功",
		"data": gin.H{
			"url": fileURL,
		},
	})
}

// SubmitKYC 提交实名认证
func SubmitKYC(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type":    "error",
			"message": "未授权",
		})
		return
	}

	var req service.KYCSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type":    "error",
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := kycService.SubmitKYC(userID.(uint), &req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"type":    "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":    "ok",
		"message": "提交成功，请等待审核",
	})
}

// GetKYCStatus 获取实名认证状态
func GetKYCStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type":    "error",
			"message": "未授权",
		})
		return
	}

	status, err := kycService.GetKYCStatus(userID.(uint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"type":    "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":    "ok",
		"message": "获取成功",
		"data":    status,
	})
}
