package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Type    string      `json:"type"`           // success 或 error
	Message string      `json:"message"`        // 提示消息
	Error   string      `json:"error"`          // 错误信息（成功时为空字符串）
	Data    interface{} `json:"data,omitempty"` // 响应数据
}

// PageData 分页数据结构
type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// PageDataWithExtra 分页数据结构（附加额外信息）
type PageDataWithExtra struct {
	List      interface{} `json:"list"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	Size      int         `json:"size"`
	ExtraData interface{} `json:"extra_data,omitempty"`
}

// LayuiResponse Layui表格响应格式 (兼容PHP后台)
type LayuiResponse struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Count     int64       `json:"count"`
	Data      interface{} `json:"data"`
	ExtraData interface{} `json:"extra_data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, message string, data ...interface{}) {
	resp := Response{
		Type:    "success",
		Message: message,
		Error:   "", // 成功时错误信息为空
	}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Type:    "error",
		Message: "操作失败",
		Error:   message, // 错误详情
	})
}

// ErrorWithCode 带状态码的错误响应
func ErrorWithCode(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Type:    "error",
		Message: "操作失败",
		Error:   message,
	})
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "请先登录"
	}
	c.JSON(http.StatusUnauthorized, Response{
		Type:    "error",
		Message: "未授权",
		Error:   message,
	})
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "没有权限"
	}
	c.JSON(http.StatusForbidden, Response{
		Type:    "error",
		Message: "禁止访问",
		Error:   message,
	})
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	c.JSON(http.StatusNotFound, Response{
		Type:    "error",
		Message: "资源不存在",
		Error:   message,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "请求参数错误"
	}
	c.JSON(http.StatusBadRequest, Response{
		Type:    "error",
		Message: "参数错误",
		Error:   message,
	})
}

// ServerError 服务器错误
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	c.JSON(http.StatusInternalServerError, Response{
		Type:    "error",
		Message: "服务器错误",
		Error:   message,
	})
}

// SuccessWithPage 分页数据成功响应
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, Response{
		Type:    "success",
		Message: "获取成功",
		Error:   "", // 成功时错误信息为空
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	})
}

// SuccessWithPageExtra 分页数据成功响应（附加额外信息）
func SuccessWithPageExtra(c *gin.Context, list interface{}, total int64, page, size int, extra interface{}) {
	c.JSON(http.StatusOK, Response{
		Type:    "success",
		Message: "获取成功",
		Error:   "", // 成功时错误信息为空
		Data: PageDataWithExtra{
			List:      list,
			Total:     total,
			Page:      page,
			Size:      size,
			ExtraData: extra,
		},
	})
}

// LayuiSuccess Layui表格成功响应 (兼容PHP后台)
func LayuiSuccess(c *gin.Context, data interface{}, count int64, extraData ...interface{}) {
	resp := LayuiResponse{
		Code:  0,
		Msg:   "",
		Count: count,
		Data:  data,
	}
	if len(extraData) > 0 {
		resp.ExtraData = extraData[0]
	}
	c.JSON(http.StatusOK, resp)
}

// LayuiError Layui表格错误响应
func LayuiError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, LayuiResponse{
		Code: 1,
		Msg:  message,
	})
}

// JSON 自定义JSON响应
func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}
