package admin

import (
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// NewsAdminService 新闻管理服务
type NewsAdminService struct{}

// NewNewsAdminService 创建新闻管理服务实例
func NewNewsAdminService() *NewsAdminService {
	return &NewsAdminService{}
}

// GetNewsList 获取新闻列表
func (s *NewsAdminService) GetNewsList(info PageInfo) (list []model.News, total int64, err error) {
	db := database.DB.Model(&model.News{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetNewsInfo 获取新闻详情
func (s *NewsAdminService) GetNewsInfo(id uint) (news *model.News, err error) {
	var n model.News
	err = database.DB.Where("id = ?", id).First(&n).Error
	return &n, err
}

// CreateNews 创建新闻
func (s *NewsAdminService) CreateNews(news model.News) (err error) {
	news.CreateTime = time.Now().Unix()
	news.UpdateTime = time.Now().Unix()
	return database.DB.Create(&news).Error
}

// UpdateNews 更新新闻
func (s *NewsAdminService) UpdateNews(news model.News) (err error) {
	news.UpdateTime = time.Now().Unix()
	return database.DB.Save(&news).Error
}

// DeleteNews 删除新闻
func (s *NewsAdminService) DeleteNews(id uint) (err error) {
	return database.DB.Delete(&model.News{}, id).Error
}

// PublishNews 发布新闻
func (s *NewsAdminService) PublishNews(id uint) (err error) {
	return database.DB.Model(&model.News{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       1,
		"publish_time": time.Now().Unix(),
		"update_time":  time.Now().Unix(),
	}).Error
}

// DraftNews 将新闻设为草稿
func (s *NewsAdminService) DraftNews(id uint) (err error) {
	return database.DB.Model(&model.News{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      0,
		"update_time": time.Now().Unix(),
	}).Error
}

// GetNewsCategoryList 获取新闻分类列表
func (s *NewsAdminService) GetNewsCategoryList() (list []model.NewsCategory, err error) {
	err = database.DB.Find(&list).Error
	return
}

// CreateNewsCategory 创建新闻分类
func (s *NewsAdminService) CreateNewsCategory(category model.NewsCategory) (err error) {
	category.CreateTime = time.Now().Unix()
	category.UpdateTime = time.Now().Unix()
	return database.DB.Create(&category).Error
}

// UpdateNewsCategory 更新新闻分类
func (s *NewsAdminService) UpdateNewsCategory(category model.NewsCategory) (err error) {
	category.UpdateTime = time.Now().Unix()
	return database.DB.Save(&category).Error
}

// DeleteNewsCategory 删除新闻分类
func (s *NewsAdminService) DeleteNewsCategory(id uint) (err error) {
	return database.DB.Delete(&model.NewsCategory{}, id).Error
}

// GetFeedbackList 获取反馈列表
func (s *NewsAdminService) GetFeedbackList(info PageInfo) (list []model.FeedBack, total int64, err error) {
	db := database.DB.Model(&model.FeedBack{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ReplyFeedback 回复反馈
func (s *NewsAdminService) ReplyFeedback(id uint, reply string) (err error) {
	return database.DB.Model(&model.FeedBack{}).Where("id = ?", id).Updates(map[string]interface{}{
		"reply":       reply,
		"status":      1,
		"update_time": time.Now().Unix(),
	}).Error
}

// CloseFeedback 关闭反馈
func (s *NewsAdminService) CloseFeedback(id uint) (err error) {
	return database.DB.Model(&model.FeedBack{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      2,
		"update_time": time.Now().Unix(),
	}).Error
}
