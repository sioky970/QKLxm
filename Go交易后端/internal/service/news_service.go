package service

import (
	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// NewsService 新闻服务（前台）
type NewsService struct{}

// NewNewsService 创建新闻服务实例
func NewNewsService() *NewsService {
	return &NewsService{}
}

// GetPublishedNewsList 获取已发布的新闻列表
func (s *NewsService) GetPublishedNewsList(categoryID uint, page, pageSize int) (list []model.News, total int64, err error) {
	db := database.DB.Model(&model.News{}).Where("status = ?", 1)

	// 按分类筛选
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	err = db.Limit(pageSize).Offset(offset).Order("publish_time DESC").Find(&list).Error
	return list, total, err
}

// GetNewsDetail 获取新闻详情并增加浏览次数
func (s *NewsService) GetNewsDetail(id uint) (news *model.News, err error) {
	var n model.News
	err = database.DB.Where("id = ? AND status = ?", id, 1).First(&n).Error
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	database.DB.Model(&model.News{}).Where("id = ?", id).Update("views", n.Views+1)
	n.Views++

	return &n, nil
}

// GetNewsCategoryList 获取新闻分类列表
func (s *NewsService) GetNewsCategoryList() (list []model.NewsCategory, err error) {
	err = database.DB.Where("status = ?", 1).Find(&list).Error
	return
}

// GetRecommendedNews 获取推荐新闻（取最新发布的前N条）
func (s *NewsService) GetRecommendedNews(limit int) (list []model.News, err error) {
	if limit <= 0 {
		limit = 5
	}
	err = database.DB.Model(&model.News{}).
		Where("status = ?", 1).
		Order("publish_time DESC").
		Limit(limit).
		Find(&list).Error
	return
}

// GetHotNews 获取热门新闻（按浏览量排序）
func (s *NewsService) GetHotNews(limit int) (list []model.News, err error) {
	if limit <= 0 {
		limit = 5
	}
	err = database.DB.Model(&model.News{}).
		Where("status = ?", 1).
		Order("views DESC").
		Limit(limit).
		Find(&list).Error
	return
}
