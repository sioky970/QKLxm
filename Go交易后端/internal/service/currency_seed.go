package service

import (
	"errors"
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"

	"gorm.io/gorm"
)

func EnsureDefaultCurrencies() error {
	now := time.Now().Unix()
	if err := ensureCurrencyRiskColumns(); err != nil {
		return err
	}
	hasIsChange := columnExists("currency", "is_change")
	hasUpdateTime := columnExists("currency", "update_time")
	hasCreateTime := columnExists("currency", "create_time")
	// 币种精度配置说明：
	// - DecimalScale: 价格显示小数位数
	// - 高价币(>$1000): 2位 | 中价币($10-$1000): 2-4位 | 低价币($0.1-$10): 4位 | 超低价(<$0.1): 6-8位
	defaults := []model.Currency{
		{Name: "BTC", Logo: "/coins/btc.png", Type: "coin", DecimalScale: 2, Rate: 88737.18, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 1000},
		{Name: "ETH", Logo: "/coins/eth.png", Type: "coin", DecimalScale: 2, Rate: 2930.35, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 999},
		{Name: "USDT", Logo: "/coins/usdt.png", Type: "coin", DecimalScale: 4, Rate: 1, IsLegal: 1, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 998},
		// HT - 已删除（火币不支持）
		{Name: "XRP", Logo: "/coins/xrp.png", Type: "coin", DecimalScale: 4, Rate: 1.88735, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 996},
		{Name: "LTC", Logo: "/coins/ltc.png", Type: "coin", DecimalScale: 2, Rate: 68.91, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 986},
		{Name: "BCH", Logo: "/coins/bch.png", Type: "coin", DecimalScale: 2, Rate: 582.73, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 991},
		// EOS - 已删除（火币不支持）
		{Name: "ETC", Logo: "/coins/etc.png", Type: "coin", DecimalScale: 2, Rate: 11.5, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 981},
		{Name: "TRX", Logo: "/coins/trx.png", Type: "coin", DecimalScale: 4, Rate: 0.298202, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 994},
		{Name: "ADA", Logo: "/coins/ada.png", Type: "coin", DecimalScale: 4, Rate: 0.352364, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 992},
		{Name: "DOT", Logo: "/coins/dot.png", Type: "coin", DecimalScale: 4, Rate: 1.8884, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 984},
		{Name: "LINK", Logo: "/coins/link.png", Type: "coin", DecimalScale: 2, Rate: 11.9891, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 989},
		{Name: "UNI", Logo: "/coins/uni.png", Type: "coin", DecimalScale: 4, Rate: 4.8278, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 983},
		{Name: "BNB", Logo: "/coins/bnb.png", Type: "coin", DecimalScale: 2, Rate: 878.56, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 997},
		{Name: "XLM", Logo: "/coins/xlm.png", Type: "coin", DecimalScale: 4, Rate: 0.208438, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 988},
		{Name: "ATOM", Logo: "/coins/atom.png", Type: "coin", DecimalScale: 4, Rate: 2.267, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 980},
		{Name: "FIL", Logo: "/coins/fil.png", Type: "coin", DecimalScale: 4, Rate: 1.2836, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 978},
		{Name: "XTZ", Logo: "/coins/xtz.png", Type: "coin", DecimalScale: 4, Rate: 0.5785, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 975},
		{Name: "NEO", Logo: "/coins/neo.png", Type: "coin", DecimalScale: 4, Rate: 3.57, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 967},
		{Name: "IOTA", Logo: "/coins/iota.png", Type: "coin", DecimalScale: 6, Rate: 0.0855, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 972},
		{Name: "DASH", Logo: "/coins/dash.png", Type: "coin", DecimalScale: 2, Rate: 62.25, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 976},
		{Name: "ZEC", Logo: "/coins/zec.png", Type: "coin", DecimalScale: 2, Rate: 358.15, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 987},
		{Name: "XMR", Logo: "/coins/xmr.png", Type: "coin", DecimalScale: 2, Rate: 467.32, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 990},
		{Name: "DOGE", Logo: "/coins/doge.png", Type: "coin", DecimalScale: 4, Rate: 0.122078, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 993},
		{Name: "AVAX", Logo: "/coins/avax.png", Type: "coin", DecimalScale: 2, Rate: 11.8746, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 985},
		{Name: "SOL", Logo: "/coins/sol.png", Type: "coin", DecimalScale: 2, Rate: 126.0295, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 995},
		{Name: "VET", Logo: "/coins/vet.png", Type: "coin", DecimalScale: 6, Rate: 0.010212, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 977},
		{Name: "THETA", Logo: "/coins/theta.png", Type: "coin", DecimalScale: 4, Rate: 0.2868, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 970},
		{Name: "AAVE", Logo: "/coins/aave.png", Type: "coin", DecimalScale: 2, Rate: 153.7133, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 982},
		{Name: "COMP", Logo: "/coins/comp.png", Type: "coin", DecimalScale: 2, Rate: 24.61, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 966},
		{Name: "SNX", Logo: "/coins/snx.png", Type: "coin", DecimalScale: 4, Rate: 0.4132, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 964},
		{Name: "YFI", Logo: "/coins/yfi.png", Type: "coin", DecimalScale: 2, Rate: 3309.27, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 961},
		// MKR - 已删除（火币不支持）
		{Name: "BAT", Logo: "/coins/bat.png", Type: "coin", DecimalScale: 4, Rate: 0.1744, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 968},
		{Name: "ZRX", Logo: "/coins/zrx.png", Type: "coin", DecimalScale: 4, Rate: 0.1258, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 960},
		{Name: "QTUM", Logo: "/coins/qtum.png", Type: "coin", DecimalScale: 4, Rate: 1.2723, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 963},
		{Name: "ICX", Logo: "/coins/icx.png", Type: "coin", DecimalScale: 6, Rate: 0.0553, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 957},
		{Name: "ONT", Logo: "/coins/ont.png", Type: "coin", DecimalScale: 6, Rate: 0.0577, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 956},
		{Name: "ZIL", Logo: "/coins/zil.png", Type: "coin", DecimalScale: 6, Rate: 0.004908, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 959},
		{Name: "KNC", Logo: "/coins/knc.png", Type: "coin", DecimalScale: 4, Rate: 0.2248, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 955},
		// OMG - 已删除（火币不支持）
		{Name: "SUSHI", Logo: "/coins/sushi.png", Type: "coin", DecimalScale: 4, Rate: 0.2992, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 958},
		{Name: "CRV", Logo: "/coins/crv.png", Type: "coin", DecimalScale: 4, Rate: 0.3569, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 974},
		{Name: "GRT", Logo: "/coins/grt.png", Type: "coin", DecimalScale: 6, Rate: 0.036025, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 973},
		{Name: "SAND", Logo: "/coins/sand.png", Type: "coin", DecimalScale: 4, Rate: 0.13626, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 971},
		{Name: "MANA", Logo: "/coins/mana.png", Type: "coin", DecimalScale: 4, Rate: 0.1471, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 969},
		{Name: "ALGO", Logo: "/coins/algo.png", Type: "coin", DecimalScale: 4, Rate: 0.116, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 979},
		{Name: "KSM", Logo: "/coins/ksm.png", Type: "coin", DecimalScale: 4, Rate: 8.4936, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 962},
		{Name: "APE", Logo: "/coins/ape.png", Type: "coin", DecimalScale: 4, Rate: 0.1815, IsLegal: 0, IsLever: 0, IsMicro: 0, IsDisplay: 1, IsChange: 1, Sort: 965},
	}

	names := make([]string, 0, len(defaults))
	for _, currency := range defaults {
		names = append(names, currency.Name)
		currency.CreateTime = now
		currency.UpdateTime = now

		var existing model.Currency
		err := database.DB.Where("name = ?", currency.Name).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				createData := map[string]interface{}{
					"name":          currency.Name,
					"logo":          currency.Logo,
					"type":          currency.Type,
					"decimal_scale": currency.DecimalScale,
					"rate":          currency.Rate,
					"is_legal":      currency.IsLegal,
					"is_lever":      currency.IsLever,
					"is_micro":      currency.IsMicro,
					"is_display":    currency.IsDisplay,
					"sort":          currency.Sort,
					"get_address":   currency.GetAddress,
				}
				if hasCreateTime {
					createData["create_time"] = now
				}
				if hasUpdateTime {
					createData["update_time"] = now
				}
				if hasIsChange {
					createData["is_change"] = currency.IsChange
				}
				if err := database.DB.Table("currency").Create(createData).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}

		updates := map[string]interface{}{
			"logo":          currency.Logo,
			"type":          currency.Type,
			"decimal_scale": currency.DecimalScale,
			"rate":          currency.Rate,
			"is_legal":      currency.IsLegal,
			"is_lever":      currency.IsLever,
			"is_micro":      currency.IsMicro,
			// is_display 不更新，保留用户设置的启用/禁用状态
			"is_change":   currency.IsChange,
			"sort":        currency.Sort,
			"get_address": currency.GetAddress,
		}
		if hasUpdateTime {
			updates["update_time"] = now
		}
		if !hasIsChange {
			delete(updates, "is_change")
		}
		if err := database.DB.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
	}

	if len(names) > 0 {
		if err := database.DB.Where("name NOT IN ?", names).
			Delete(&model.Currency{}).Error; err != nil {
			return err
		}
	}

	logger.Infof("initialized default currencies: %d", len(defaults))
	return nil
}

func columnExists(tableName, columnName string) bool {
	var count int64
	database.DB.Raw(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName, columnName,
	).Scan(&count)
	return count > 0
}

func ensureCurrencyRiskColumns() error {
	columns := []struct {
		name string
		def  string
	}{
		{"risk_prob_enabled", "TINYINT NOT NULL DEFAULT 0"},
		{"risk_profit_probability", "INT NOT NULL DEFAULT 50"},
		{"risk_money_enabled", "TINYINT NOT NULL DEFAULT 0"},
		{"risk_money_min", "DECIMAL(20,8) NOT NULL DEFAULT 0"},
		{"risk_money_max", "DECIMAL(20,8) NOT NULL DEFAULT 0"},
		{"risk_money_result", "TINYINT NOT NULL DEFAULT 0"},
		{"risk_time_enabled", "TINYINT NOT NULL DEFAULT 0"},
		{"risk_time_start", "VARCHAR(10) NOT NULL DEFAULT ''"},
		{"risk_time_end", "VARCHAR(10) NOT NULL DEFAULT ''"},
		{"risk_time_result", "TINYINT NOT NULL DEFAULT 0"},
	}

	for _, col := range columns {
		if columnExists("currency", col.name) {
			continue
		}
		sql := "ALTER TABLE `currency` ADD COLUMN `" + col.name + "` " + col.def
		if err := database.DB.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
