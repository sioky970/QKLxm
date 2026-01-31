package admin

import (
	"errors"

	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// CurrencyAdminService 币种管理服务
type CurrencyAdminService struct{}

// NewCurrencyAdminService 创建币种管理服务实例
func NewCurrencyAdminService() *CurrencyAdminService {
	return &CurrencyAdminService{}
}

// CurrencyListItem includes market data for admin list.
type CurrencyListItem struct {
	model.Currency
	LatestPrice   float64 `json:"latest_price"`
	LastQuoteTime int64   `json:"last_quote_time"`
	KlineCount    int64   `json:"kline_count"`
}

// GetCurrencyList 获取币种列表
func (s *CurrencyAdminService) GetCurrencyList(info PageInfo) (list []model.Currency, total int64, err error) {
	db := database.DB.Model(&model.Currency{}).Where("name <> ?", "USDT")

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("sort DESC, id DESC").Find(&list).Error
	return list, total, err
}

// GetCurrencyListWithMarket gets currency list with market data.
func (s *CurrencyAdminService) GetCurrencyListWithMarket(info PageInfo) (list []CurrencyListItem, total int64, err error) {
	db := database.DB.Model(&model.Currency{}).Where("name <> ?", "USDT")

	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var currencies []model.Currency
	if err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("sort DESC, id DESC").Find(&currencies).Error; err != nil {
		return nil, 0, err
	}

	var usdt model.Currency
	if err = database.DB.Where("name = ?", "USDT").First(&usdt).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, err
	}
	legalID := usdt.ID

	for _, c := range currencies {
		item := CurrencyListItem{
			Currency: c,
		}

		if legalID > 0 {
			var quotation model.CurrencyQuotation
			q := database.DB.Where("currency_id = ? AND legal_id = ?", c.ID, legalID).First(&quotation)
			if q.Error == nil {
				item.LatestPrice = quotation.Price
				item.LastQuoteTime = quotation.AddTime
			} else if q.Error != nil && !errors.Is(q.Error, gorm.ErrRecordNotFound) {
				return nil, 0, q.Error
			}

			if c.ID != legalID {
				var count int64
				if err = database.DB.Model(&model.MarketHour{}).
					Where("currency_id = ? AND legal_id = ? AND period = ?", c.ID, legalID, "1min").
					Count(&count).Error; err != nil {
					return nil, 0, err
				}
				item.KlineCount = count
			}
		}

		list = append(list, item)
	}

	return list, total, nil
}

// GetCurrencyInfo 获取币种详情
func (s *CurrencyAdminService) GetCurrencyInfo(id uint) (currency *model.Currency, err error) {
	var c model.Currency
	err = database.DB.Where("id = ?", id).First(&c).Error
	return &c, err
}

// CreateCurrency 创建币种
func (s *CurrencyAdminService) CreateCurrency(currency model.Currency) (err error) {
	var existing model.Currency
	err = database.DB.Where("name = ?", currency.Name).First(&existing).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("币种名称已存在")
	}

	return database.DB.Create(&currency).Error
}

// UpdateCurrency 更新币种
func (s *CurrencyAdminService) UpdateCurrency(currency model.Currency) (err error) {
	var existing model.Currency
	err = database.DB.Where("id = ?", currency.ID).First(&existing).Error
	if err != nil {
		return err
	}

	// 检查币种名称是否被其他记录占用
	var other model.Currency
	err = database.DB.Where("name = ? AND id != ?", currency.Name, currency.ID).First(&other).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("币种名称已被其他币种使用")
	}

	return database.DB.Save(&currency).Error
}

// DeleteCurrency 删除币种
func (s *CurrencyAdminService) DeleteCurrency(id uint) (err error) {
	return database.DB.Delete(&model.Currency{}, id).Error
}

// ToggleCurrencyDisplay 切换币种显示状态
func (s *CurrencyAdminService) ToggleCurrencyDisplay(id uint, isDisplay int8) (err error) {
	return database.DB.Model(&model.Currency{}).Where("id = ?", id).Update("is_display", isDisplay).Error
}

// UpdateCurrencyRate 更新币种汇率
func (s *CurrencyAdminService) UpdateCurrencyRate(id uint, rate float64) (err error) {
	return database.DB.Model(&model.Currency{}).Where("id = ?", id).Update("rate", rate).Error
}

// GetCurrencyMatchList 获取交易对列表
func (s *CurrencyAdminService) GetCurrencyMatchList(info PageInfo) (list []model.CurrencyMatch, total int64, err error) {
	db := database.DB.Model(&model.CurrencyMatch{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("sort ASC, id DESC").Find(&list).Error
	return list, total, err
}

// GetCurrencyMatchInfo 获取交易对详情
func (s *CurrencyAdminService) GetCurrencyMatchInfo(id uint) (match *model.CurrencyMatch, err error) {
	var m model.CurrencyMatch
	err = database.DB.Where("id = ?", id).First(&m).Error
	return &m, err
}

// UpdateCurrencyMatch 更新交易对
func (s *CurrencyAdminService) UpdateCurrencyMatch(match model.CurrencyMatch) (err error) {
	return database.DB.Save(&match).Error
}

// ToggleCurrencyMatchDisplay 切换交易对显示状态
func (s *CurrencyAdminService) ToggleCurrencyMatchDisplay(id uint, isDisplay int8) (err error) {
	return database.DB.Model(&model.CurrencyMatch{}).Where("id = ?", id).Update("is_display", isDisplay).Error
}

// SetCurrencyMatchRisk 设置交易对风控
func (s *CurrencyAdminService) SetCurrencyMatchRisk(id uint, riskResult int8) (err error) {
	return database.DB.Model(&model.CurrencyMatch{}).Where("id = ?", id).Update("risk_group_result", riskResult).Error
}

// UpdateCurrencyRisk 更新币种风控参数
func (s *CurrencyAdminService) UpdateCurrencyRisk(id uint, updates map[string]interface{}) (err error) {
	return database.DB.Model(&model.Currency{}).Where("id = ?", id).Updates(updates).Error
}
