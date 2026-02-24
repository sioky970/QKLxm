package admin

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

const (
	cacheKeyUSDTID            = "admin:currency:usdt_id"
	cacheKeyCurrencyList      = "admin:currency:list"
	cacheKeyCurrencyQuotation = "admin:currency:quotation:"
	cacheKeyMarketHourCount   = "admin:market:hour:count:"
	cacheExpireUSDTID         = 24 * time.Hour
	cacheExpireCurrencyList   = 5 * time.Minute
	cacheExpireQuotation      = 10 * time.Second
	cacheExpireKlineCount     = 1 * time.Minute
)

type CurrencyAdminService struct{}

func NewCurrencyAdminService() *CurrencyAdminService {
	return &CurrencyAdminService{}
}

type CurrencyListItem struct {
	model.Currency
	LatestPrice   float64 `json:"latest_price"`
	LastQuoteTime int64   `json:"last_quote_time"`
	KlineCount    int64   `json:"kline_count"`
}

func (s *CurrencyAdminService) GetCurrencyList(info PageInfo) (list []model.Currency, total int64, err error) {
	db := database.DB.Model(&model.Currency{}).Where("name <> ?", "USDT")

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("sort DESC, id DESC").Find(&list).Error
	return list, total, err
}

func (s *CurrencyAdminService) GetCurrencyListWithMarket(info PageInfo) (list []CurrencyListItem, total int64, err error) {
	cacheKey := fmt.Sprintf("%s:%d:%d", cacheKeyCurrencyList, info.Page, info.PageSize)

	if cache.RDB != nil {
		var cachedResult struct {
			List  []CurrencyListItem
			Total int64
		}
		if err := cache.Get(cacheKey, &cachedResult); err == nil {
			return cachedResult.List, cachedResult.Total, nil
		}
	}

	legalID, err := s.getUSDTID()
	if err != nil {
		return nil, 0, err
	}

	result, total, err := s.queryCurrencyWithMarket(legalID, info)
	if err != nil {
		return nil, 0, err
	}

	if cache.RDB != nil && len(result) > 0 {
		cachedData := struct {
			List  []CurrencyListItem
			Total int64
		}{
			List:  result,
			Total: total,
		}
		cache.Set(cacheKey, cachedData, cacheExpireCurrencyList)
	}

	return result, total, nil
}

func (s *CurrencyAdminService) queryCurrencyWithMarket(legalID uint, info PageInfo) ([]CurrencyListItem, int64, error) {
	var currencies []model.Currency
	var total int64

	db := database.DB.Model(&model.Currency{}).Where("name <> ?", "USDT")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("sort DESC, id DESC").Find(&currencies).Error; err != nil {
		return nil, 0, err
	}

	if len(currencies) == 0 {
		return []CurrencyListItem{}, total, nil
	}

	currencyIDs := make([]uint, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	quotationMap := make(map[uint]model.CurrencyQuotation)
	if legalID > 0 {
		var quotations []model.CurrencyQuotation
		query := database.DB.Where("currency_id IN ? AND legal_id = ?", currencyIDs, legalID)
		if err := query.Find(&quotations).Error; err != nil {
			return nil, 0, err
		}
		for _, q := range quotations {
			quotationMap[q.CurrencyID] = q
		}
	}

	klineCountMap := make(map[uint]int64)
	if legalID > 0 {
		var klineCounts []struct {
			CurrencyID uint  `gorm:"column:currency_id"`
			Count      int64 `gorm:"column:cnt"`
		}
		query := database.DB.Model(&model.MarketHour{}).
			Select("currency_id, COUNT(*) as cnt").
			Where("currency_id IN ? AND legal_id = ? AND period = ?", currencyIDs, legalID, "1min").
			Group("currency_id")
		if err := query.Scan(&klineCounts).Error; err != nil {
			return nil, 0, err
		}
		for _, kc := range klineCounts {
			klineCountMap[kc.CurrencyID] = kc.Count
		}
	}

	var list []CurrencyListItem
	for _, c := range currencies {
		item := CurrencyListItem{Currency: c}
		if q, ok := quotationMap[c.ID]; ok {
			item.LatestPrice = q.Price
			item.LastQuoteTime = q.AddTime
		}
		if c.ID != legalID {
			item.KlineCount = klineCountMap[c.ID]
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (s *CurrencyAdminService) getUSDTID() (uint, error) {
	if cache.RDB != nil {
		var usdtID uint
		if err := cache.Get(cacheKeyUSDTID, &usdtID); err == nil && usdtID > 0 {
			return usdtID, nil
		}
	}

	var usdt model.Currency
	if err := database.DB.Where("name = ?", "USDT").First(&usdt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	if cache.RDB != nil && usdt.ID > 0 {
		cache.Set(cacheKeyUSDTID, usdt.ID, cacheExpireUSDTID)
	}

	return usdt.ID, nil
}

func (s *CurrencyAdminService) GetCurrencyInfo(id uint) (currency *model.Currency, err error) {
	var c model.Currency
	err = database.DB.Where("id = ?", id).First(&c).Error
	return &c, err
}

func (s *CurrencyAdminService) CreateCurrency(currency model.Currency) (err error) {
	var existing model.Currency
	err = database.DB.Where("name = ?", currency.Name).First(&existing).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("币种名称已存在")
	}

	if err := database.DB.Create(&currency).Error; err != nil {
		return err
	}

	s.invalidateCurrencyCache()

	return nil
}

func (s *CurrencyAdminService) UpdateCurrency(currency model.Currency) (err error) {
	var existing model.Currency
	err = database.DB.Where("id = ?", currency.ID).First(&existing).Error
	if err != nil {
		return err
	}

	var other model.Currency
	err = database.DB.Where("name = ? AND id != ?", currency.Name, currency.ID).First(&other).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("币种名称已被其他币种使用")
	}

	if err := database.DB.Save(&currency).Error; err != nil {
		return err
	}

	s.invalidateCurrencyCache()

	return nil
}

func (s *CurrencyAdminService) DeleteCurrency(id uint) (err error) {
	if err := database.DB.Delete(&model.Currency{}, id).Error; err != nil {
		return err
	}

	s.invalidateCurrencyCache()

	return nil
}

func (s *CurrencyAdminService) invalidateCurrencyCache() {
	if cache.RDB != nil {
		pattern := cacheKeyCurrencyList + ":*"
		keys, err := cache.RDB.Keys(cache.Ctx, pattern).Result()
		if err != nil {
			logger.Warnf("获取缓存键失败: %v", err)
			return
		}
		if len(keys) > 0 {
			if err := cache.Delete(keys...); err != nil {
				logger.Warnf("删除缓存失败: %v", err)
			}
		}
		cache.Delete(cacheKeyUSDTID)
	}
}

func (s *CurrencyAdminService) ToggleCurrencyDisplay(id uint, isDisplay int8) (err error) {
	if err := database.DB.Model(&model.Currency{}).Where("id = ?", id).Update("is_display", isDisplay).Error; err != nil {
		return err
	}
	s.invalidateCurrencyCache()
	return nil
}

func (s *CurrencyAdminService) UpdateCurrencyRate(id uint, rate float64) (err error) {
	if err := database.DB.Model(&model.Currency{}).Where("id = ?", id).Update("rate", rate).Error; err != nil {
		return err
	}
	s.invalidateCurrencyCache()
	return nil
}

func (s *CurrencyAdminService) GetCurrencyMatchList(info PageInfo) (list []model.CurrencyMatch, total int64, err error) {
	db := database.DB.Model(&model.CurrencyMatch{})

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(info.GetLimit()).Offset(info.GetOffset()).Order("sort ASC, id DESC").Find(&list).Error
	return list, total, err
}

func (s *CurrencyAdminService) GetCurrencyMatchInfo(id uint) (match *model.CurrencyMatch, err error) {
	var m model.CurrencyMatch
	err = database.DB.Where("id = ?", id).First(&m).Error
	return &m, err
}

func (s *CurrencyAdminService) UpdateCurrencyMatch(match model.CurrencyMatch) (err error) {
	return database.DB.Save(&match).Error
}

func (s *CurrencyAdminService) ToggleCurrencyMatchDisplay(id uint, isDisplay int8) (err error) {
	return database.DB.Model(&model.CurrencyMatch{}).Where("id = ?", id).Update("is_display", isDisplay).Error
}

func (s *CurrencyAdminService) SetCurrencyMatchRisk(id uint, riskResult int8) (err error) {
	return database.DB.Model(&model.CurrencyMatch{}).Where("id = ?", id).Update("risk_group_result", riskResult).Error
}

func (s *CurrencyAdminService) UpdateCurrencyRisk(id uint, updates map[string]interface{}) (err error) {
	return database.DB.Model(&model.Currency{}).Where("id = ?", id).Updates(updates).Error
}
