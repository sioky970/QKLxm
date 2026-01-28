package handler

import (
	"net/http"

	"exchange-go/internal/service"
	"github.com/gin-gonic/gin"
)

var kycService = service.NewKYCService()

// SubmitKYC 提交实名认证
func SubmitKYC(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type": "error",
			"message": "未授权",
		})
		return
	}

	var req service.KYCSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := kycService.SubmitKYC(userID.(uint), &req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"type": "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type": "ok",
		"message": "提交成功，请等待审核",
	})
}

// GetKYCStatus 获取实名认证状态
func GetKYCStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type": "error",
			"message": "未授权",
		})
		return
	}

	status, err := kycService.GetKYCStatus(userID.(uint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"type": "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type": "ok",
		"message": "获取成功",
		"data": status,
	})
}
