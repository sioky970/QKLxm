package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/huobi"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

// MarketService 行情服务
type MarketService struct{}

// NewMarketService 创建行情服务实例
func NewMarketService() *MarketService {
	return &MarketService{}
}

// MarketDataRequest 行情数据请求
type MarketDataRequest struct {
	LegalID    uint `json:"legal_id" form:"legal_id"`
	CurrencyID uint `json:"currency_id" form:"currency_id"`
}

// MarketDataResponse 行情数据响应
type MarketDataResponse struct {
	LegalID   uint         `json:"legal_id"`
	LegalName string       `json:"legal_name"`
	Markets   []MarketItem `json:"markets"`
}

// MarketItem 单个市场数据项
type MarketItem struct {
	CurrencyID   uint    `json:"currency_id"`
	CurrencyName string  `json:"currency_name"`
	LegalID      uint    `json:"legal_id"`
	LegalName    string  `json:"legal_name"`
	Change       string  `json:"change"`
	NowPrice     float64 `json:"now_price"`
	Volume       float64 `json:"volume"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
}

// KlineRequest K线请求
type KlineRequest struct {
	Symbol string `json:"symbol" form:"symbol" binding:"required"`
	Period string `json:"period" form:"period" default:"5min"`
	From   int64  `json:"from" form:"from"`
	To     int64  `json:"to" form:"to"`
}

// KlineResponse K线响应
type KlineResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []KlineItem `json:"data"`
}

// KlineItem K线数据项
type KlineItem struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}

// GetMarketData 获取行情数据
func (s *MarketService) GetMarketData(req *MarketDataRequest) ([]MarketDataResponse, error) {
	// 获取所有法币
	var legalCurrencies []model.Currency
	result := database.DB.Where("is_display = 1 AND is_legal = 1").Order("sort ASC").Find(&legalCurrencies)
	if result.Error != nil {
		return nil, result.Error
	}

	var responses []MarketDataResponse

	for _, legal := range legalCurrencies {
		// 获取该法币下的所有交易对行情
		var quotations []model.CurrencyQuotation
		query := database.DB.Where("legal_id = ?", legal.ID)
		if req.LegalID > 0 {
			query = query.Where("legal_id = ?", req.LegalID)
		}
		query = query.Where("is_display = 1").Order("sort DESC").Find(&quotations)

		var marketData []MarketItem
		for _, quotation := range quotations {
			var currency model.Currency
			result := database.DB.First(&currency, quotation.CurrencyID)
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, result.Error
			}

			if currency.IsDisplay != 1 {
				continue
			}

			marketData = append(marketData, MarketItem{
				CurrencyID:   quotation.CurrencyID,
				CurrencyName: currency.Name,
				LegalID:      legal.ID,
				LegalName:    legal.Name,
				Change:       quotation.Change,
				NowPrice:     quotation.Price,
				Volume:       quotation.Volume,
				High:         quotation.Price,
				Low:          quotation.Price,
			})
		}

		responses = append(responses, MarketDataResponse{
			LegalID:   legal.ID,
			LegalName: legal.Name,
			Markets:   marketData,
		})
	}

	return responses, nil
}

// GetNewQuotation 获取最新行情
func (s *MarketService) GetNewQuotation() ([]model.CurrencyQuotation, error) {
	var quotations []model.CurrencyQuotation
	result := database.DB.Order("id DESC").Find(&quotations)
	if result.Error != nil {
		return nil, result.Error
	}
	return quotations, nil
}

// GetKline 获取K线数据
func (s *MarketService) GetKline(req *KlineRequest) (*KlineResponse, error) {
	// 验证周期参数 - 移除1min支持
	validPeriods := map[string]bool{
		"5min": true, "15min": true, "30min": true,
		"60min": true, "1H": true, "4H": true, "1D": true, "1W": true, "1M": true,
		"1day": true, "1week": true, "1mon": true, "4hour": true,
	}

	if !validPeriods[req.Period] {
		return &KlineResponse{
			Code: -1,
			Msg:  "error: period invalid",
			Data: []KlineItem{},
		}, nil
	}

	// 解析交易对
	symbolParts := []string{}
	if req.Symbol != "" {
		symbolParts = []string{}
		for i, char := range req.Symbol {
			if char == '/' {
				symbolParts = append(symbolParts, req.Symbol[:i])
				symbolParts = append(symbolParts, req.Symbol[i+1:])
				break
			}
		}
	}

	if len(symbolParts) != 2 {
		return &KlineResponse{
			Code: -1,
			Msg:  "error: symbol format invalid, should be like BTC/USDT",
			Data: []KlineItem{},
		}, nil
	}

	baseCurrencyName := symbolParts[0]
	quoteCurrencyName := symbolParts[1]

	// 映射周期到火币格式 - 移除1min支持
	periodMap := map[string]string{
		"5min":  "5min",
		"15min": "15min",
		"30min": "30min",
		"60min": "60min",
		"1H":    "60min",
		"4H":    "4hour",
		"4hour": "4hour",
		"1D":    "1day",
		"1W":    "1week",
		"1M":    "1mon",
		"1day":  "1day",
		"1week": "1week",
		"1mon":  "1mon",
	}
	period := periodMap[req.Period]

	// 构建火币格式的symbol (btcusdt)
	huobiSymbol := strings.ToLower(baseCurrencyName + quoteCurrencyName)

	// 首先尝试从Redis获取K线数据(火币WebSocket数据)
	klines, err := huobi.GetKlineFromRedis(huobiSymbol, period, 1000)
	if err == nil && len(klines) > 0 {
		logger.Debugf("从Redis获取到%d条%s-%s的K线数据", len(klines), baseCurrencyName, quoteCurrencyName)
		var klineItems []KlineItem
		for _, kline := range klines {
			klineItems = append(klineItems, KlineItem{
				Time:   kline.ID * 1000, // 转换为毫秒
				Open:   kline.Open,
				Close:  kline.Close,
				High:   kline.High,
				Low:    kline.Low,
				Volume: kline.Vol,
			})
		}
		return &KlineResponse{
			Code: 1,
			Msg:  "success",
			Data: klineItems,
		}, nil
	}

	// 如果Redis没有数据，尝试从历史K线调度器获取（会从API获取并缓存）
	historyScheduler := huobi.GetHistoryKlineScheduler()
	if historyScheduler != nil {
		historyKlines, err := historyScheduler.GetHistoryKline(huobiSymbol, period, 300)
		if err == nil && len(historyKlines) > 0 {
			logger.Debugf("从历史调度器获取到%d条%s-%s的K线数据", len(historyKlines), baseCurrencyName, quoteCurrencyName)
			var klineItems []KlineItem
			for _, kline := range historyKlines {
				klineItems = append(klineItems, KlineItem{
					Time:   kline.ID * 1000, // 转换为毫秒
					Open:   kline.Open,
					Close:  kline.Close,
					High:   kline.High,
					Low:    kline.Low,
					Volume: kline.Vol,
				})
			}
			return &KlineResponse{
				Code: 1,
				Msg:  "success",
				Data: klineItems,
			}, nil
		} else if err != nil {
			logger.Errorf("从历史调度器获取%s-%s的K线数据失败: %v", baseCurrencyName, quoteCurrencyName, err)
		}
	}

	// 如果Redis没有数据，尝试从数据库获取
	var baseCurrency, quoteCurrency model.Currency
	result := database.DB.Where("name = ? AND is_display = 1", baseCurrencyName).First(&baseCurrency)
	if result.Error != nil {
		return &KlineResponse{
			Code: -1,
			Msg:  "error: symbol not exist",
			Data: []KlineItem{},
		}, nil
	}

	result = database.DB.Where("name = ? AND is_display = 1 AND is_legal = 1", quoteCurrencyName).First(&quoteCurrency)
	if result.Error != nil {
		return &KlineResponse{
			Code: -1,
			Msg:  "error: symbol not exist",
			Data: []KlineItem{},
		}, nil
	}

	// 设置默认时间范围
	from := req.From
	to := req.To
	if from == 0 {
		from = time.Now().AddDate(0, 0, -7).Unix()
	}
	if to == 0 {
		to = time.Now().Unix()
	}

	// 从数据库获取K线数据
	var dbKlines []model.MarketHour
	query := database.DB.Where("currency_id = ? AND legal_id = ?", baseCurrency.ID, quoteCurrency.ID)
	query = query.Where("period = ?", period)
	result = query.Order("id DESC").Limit(1000).Find(&dbKlines)

	if result.Error != nil {
		logger.Warnf("从数据库获取K线数据失败: %v", result.Error)
		// 返回空数据而不是错误
		return &KlineResponse{
			Code: 1,
			Msg:  "success",
			Data: []KlineItem{},
		}, nil
	}

	// 转换数据格式
	var klineItems []KlineItem
	for _, kline := range dbKlines {
		klineItems = append(klineItems, KlineItem{
			Time:   kline.Timestamp * 1000, // 转换为毫秒
			Open:   kline.Open,
			Close:  kline.Close,
			High:   kline.High,
			Low:    kline.Low,
			Volume: kline.Volume,
		})
	}

	return &KlineResponse{
		Code: 1,
		Msg:  "success",
		Data: klineItems,
	}, nil
}

// GetCurrencyQuotation 获取币种行情
func (s *MarketService) GetCurrencyQuotation(currencyID, legalID uint) (*model.CurrencyQuotation, error) {
	var quotation model.CurrencyQuotation
	var conditions []string
	var values []interface{}

	if currencyID > 0 {
		conditions = append(conditions, "currency_id = ?")
		values = append(values, currencyID)
	}
	if legalID > 0 {
		conditions = append(conditions, "legal_id = ?")
		values = append(values, legalID)
	}

	query := database.DB
	if len(conditions) > 0 {
		whereClause := ""
		for i, condition := range conditions {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += condition
		}
		query = query.Where(whereClause, values...)
	}

	result := query.First(&quotation)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("行情数据不存在")
		}
		return nil, result.Error
	}

	return &quotation, nil
}

// UpdateQuotation 更新行情数据
func (s *MarketService) UpdateQuotation(currencyID, legalID uint, data map[string]interface{}) error {
	var quotation model.CurrencyQuotation
	result := database.DB.Where("currency_id = ? AND legal_id = ?", currencyID, legalID).First(&quotation)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 创建新的行情记录
			quotation = model.CurrencyQuotation{
				CurrencyID: currencyID,
				LegalID:    legalID,
			}
			for key, value := range data {
				switch key {
				case "price":
					if v, ok := value.(float64); ok {
						quotation.Price = v
					}
				case "volume":
					if v, ok := value.(float64); ok {
						quotation.Volume = v
					}
				case "change":
					if v, ok := value.(string); ok {
						quotation.Change = v
					} else if v, ok := value.(float64); ok {
						quotation.Change = fmt.Sprintf("%.2f", v)
					}
				}
			}
			return database.DB.Create(&quotation).Error
		}
		return result.Error
	}

	// 更新现有行情记录
	return database.DB.Model(&quotation).Updates(data).Error
}

// GetMarketDepth 获取市场深度
func (s *MarketService) GetMarketDepth(currencyID, legalID uint) (map[string]interface{}, error) {
	// 这里应该从订单薄数据中获取深度信息
	// 暂时返回模拟数据
	return map[string]interface{}{
		"bids": []interface{}{ // 买单
			[]float64{45000.0, 1.2}, // [价格, 数量]
			[]float64{44999.5, 0.8},
			[]float64{44999.0, 2.5},
		},
		"asks": []interface{}{ // 卖单
			[]float64{45001.0, 0.5}, // [价格, 数量]
			[]float64{45001.5, 1.0},
			[]float64{45002.0, 3.2},
		},
		"timestamp": time.Now().Unix(),
	}, nil
}
