package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/response"
	"exchange-go/internal/service/admin"
)

// ============ 新闻管理 ============

// GetNewsList 获取新闻列表
func GetNewsList(c *gin.Context) {
	var info admin.PageInfo
	if err := c.ShouldBind(&info); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := adminService.News.GetNewsList(info)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, info.Page, info.PageSize)
}

// GetNewsDetail 获取新闻详情
func GetNewsDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	news, err := adminService.News.GetNewsInfo(uint(id))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", news)
}

// CreateNews 创建新闻
func CreateNews(c *gin.Context) {
	var req model.News
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.Title == "" {
		response.BadRequest(c, "标题不能为空")
		return
	}
	if req.Content == "" {
		response.BadRequest(c, "内容不能为空")
		return
	}

	if err := adminService.News.CreateNews(req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功")
}

// UpdateNews 更新新闻
func UpdateNews(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	var req model.News
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	req.ID = uint(id)

	if err := adminService.News.UpdateNews(req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteNews 删除新闻
func DeleteNews(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	if err := adminService.News.DeleteNews(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// PublishNews 发布新闻
func PublishNews(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.News.PublishNews(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "发布成功")
}

// DraftNews 设为草稿
func DraftNews(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := adminService.News.DraftNews(req.ID); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "已设为草稿")
}

// ============ 新闻分类管理 ============

// GetNewsCategoryList 获取新闻分类列表
func GetNewsCategoryList(c *gin.Context) {
	list, err := adminService.News.GetNewsCategoryList()
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "获取成功", list)
}

// CreateNewsCategory 创建新闻分类
func CreateNewsCategory(c *gin.Context) {
	var req model.NewsCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.Name == "" {
		response.BadRequest(c, "分类名称不能为空")
		return
	}

	if err := adminService.News.CreateNewsCategory(req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "创建成功")
}

// UpdateNewsCategory 更新新闻分类
func UpdateNewsCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	var req model.NewsCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	req.ID = uint(id)

	if err := adminService.News.UpdateNewsCategory(req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// DeleteNewsCategory 删除新闻分类
func DeleteNewsCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	if err := adminService.News.DeleteNewsCategory(uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "删除成功")
}
