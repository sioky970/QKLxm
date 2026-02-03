package huobi

import (
	"fmt"
	"strings"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

func loadEnabledCurrencies() ([]model.Currency, model.Currency, error) {
	var currencies []model.Currency
	if err := database.DB.Where("is_display = 1").Order("sort DESC").Find(&currencies).Error; err != nil {
		return nil, model.Currency{}, err
	}

	var usdt model.Currency
	for _, c := range currencies {
		if strings.EqualFold(strings.TrimSpace(c.Name), "USDT") {
			usdt = c
			break
		}
	}

	if usdt.ID == 0 {
		logger.Warn("USDT currency not found in enabled list")
		return currencies, model.Currency{}, fmt.Errorf("USDT currency not found")
	}

	return currencies, usdt, nil
}

func buildSymbolsFromCurrencies(currencies []model.Currency, usdt model.Currency) ([]string, map[string]SymbolInfo) {
	symbols := make([]string, 0, len(currencies))
	symbolMap := make(map[string]SymbolInfo)

	for _, c := range currencies {
		name := strings.TrimSpace(c.Name)
		if name == "" || strings.EqualFold(name, "USDT") {
			continue
		}

		symbol := strings.ToLower(name + "usdt")
		if _, exists := symbolMap[symbol]; exists {
			continue
		}

		// 获取精度配置，默认为4
		decimalScale := c.DecimalScale
		if decimalScale <= 0 {
			decimalScale = 4
		}

		symbols = append(symbols, symbol)
		symbolMap[symbol] = SymbolInfo{
			HuobiSymbol:  symbol,
			LocalSymbol:  fmt.Sprintf("%s/USDT", strings.ToUpper(name)),
			CurrencyID:   c.ID,
			LegalID:      usdt.ID,
			CurrencyName: c.Name,
			LegalName:    "USDT",
			DecimalScale: decimalScale, // 价格精度
		}
	}

	return symbols, symbolMap
}

func loadEnabledSymbolMap() ([]string, map[string]SymbolInfo, error) {
	currencies, usdt, err := loadEnabledCurrencies()
	if err != nil {
		return nil, nil, err
	}

	symbols, symbolMap := buildSymbolsFromCurrencies(currencies, usdt)
	return symbols, symbolMap, nil
}
