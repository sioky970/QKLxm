package service

import (
	"errors"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
	"gorm.io/gorm"
)

// CurrencyService 币种服务
type CurrencyService struct{}

// NewCurrencyService 创建币种服务实例
func NewCurrencyService() *CurrencyService {
	return &CurrencyService{}
}

// CurrencyListResponse 币种列表响应
type CurrencyListResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Logo         string  `json:"logo"`
	IsLegal      int8    `json:"is_legal"`
	IsLever      int8    `json:"is_lever"`
	IsMicro      int8    `json:"is_micro"`
	DecimalScale int     `json:"decimal_scale"`
	Type         string  `json:"type"`
	Rate         float64 `json:"rate"`
}

// GetCurrencyList 获取币种列表
func (s *CurrencyService) GetCurrencyList() ([]CurrencyListResponse, error) {
	var currencies []model.Currency
	result := database.DB.Where("is_display = 1").Order("sort ASC").Find(&currencies)
	if result.Error != nil {
		return nil, result.Error
	}

	var list []CurrencyListResponse
	for _, c := range currencies {
		list = append(list, CurrencyListResponse{
			ID:           c.ID,
			Name:         c.Name,
			Logo:         c.Logo,
			IsLegal:      c.IsLegal,
			IsLever:      c.IsLever,
			IsMicro:      c.IsMicro,
			DecimalScale: c.DecimalScale,
			Type:         c.Type,
			Rate:         c.Rate,
		})
	}

	return list, nil
}

// GetCurrencyByID 根据ID获取币种
func (s *CurrencyService) GetCurrencyByID(id uint) (*model.Currency, error) {
	var currency model.Currency
	result := database.DB.First(&currency, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("币种不存在")
		}
		return nil, result.Error
	}
	return &currency, nil
}

// GetCurrencyByName 根据名称获取币种
func (s *CurrencyService) GetCurrencyByName(name string) (*model.Currency, error) {
	var currency model.Currency
	result := database.DB.Where("name = ?", name).First(&currency)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("币种不存在")
		}
		return nil, result.Error
	}
	return &currency, nil
}

// GetTradePairs 获取交易对列表
func (s *CurrencyService) GetTradePairs() ([]model.CurrencyMatch, error) {
	var pairs []model.CurrencyMatch
	result := database.DB.Where("is_display = 1").Order("sort ASC").Find(&pairs)
	if result.Error != nil {
		return nil, result.Error
	}
	return pairs, nil
}

// GetTradePairByID 根据ID获取交易对
func (s *CurrencyService) GetTradePairByID(id uint) (*model.CurrencyMatch, error) {
	var pair model.CurrencyMatch
	result := database.DB.First(&pair, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("交易对不存在")
		}
		return nil, result.Error
	}
	return &pair, nil
}

// GetTradePairByCurrency 根据币种获取交易对
func (s *CurrencyService) GetTradePairByCurrency(currencyID, legalID uint) (*model.CurrencyMatch, error) {
	var pair model.CurrencyMatch
	result := database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).First(&pair)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("交易对不存在")
		}
		return nil, result.Error
	}
	return &pair, nil
}

// GetLeverPairs 获取合约交易对
func (s *CurrencyService) GetLeverPairs() ([]model.CurrencyMatch, error) {
	var pairs []model.CurrencyMatch
	result := database.DB.Where("is_display = 1 AND is_lever = 1").Order("sort ASC").Find(&pairs)
	if result.Error != nil {
		return nil, result.Error
	}
	return pairs, nil
}

// GetMicroPairs 获取秒合约交易对
func (s *CurrencyService) GetMicroPairs() ([]model.CurrencyMatch, error) {
	var pairs []model.CurrencyMatch
	result := database.DB.Where("is_display = 1 AND is_micro = 1").Order("sort ASC").Find(&pairs)
	if result.Error != nil {
		return nil, result.Error
	}
	return pairs, nil
}

// GetLegalCurrencies 获取法币列表
func (s *CurrencyService) GetLegalCurrencies() ([]model.Currency, error) {
	var currencies []model.Currency
	result := database.DB.Where("is_display = 1 AND is_legal = 1").Order("sort ASC").Find(&currencies)
	if result.Error != nil {
		return nil, result.Error
	}
	return currencies, nil
}

// GetDefaultLegalCurrencyID returns the USDT currency id as default quote.
func (s *CurrencyService) GetDefaultLegalCurrencyID() (uint, error) {
	var currency model.Currency
	if err := database.DB.Where("name = ?", "USDT").First(&currency).Error; err != nil {
		return 0, err
	}
	return currency.ID, nil
}

// UpdateCurrencyRate 更新币种汇率
func (s *CurrencyService) UpdateCurrencyRate(id uint, rate float64) error {
	result := database.DB.Model(&model.Currency{}).Where("id = ?", id).Update("rate", rate)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("币种不存在")
	}
	logger.Infof("更新币种汇率: id=%d, rate=%f", id, rate)
	return nil
}
