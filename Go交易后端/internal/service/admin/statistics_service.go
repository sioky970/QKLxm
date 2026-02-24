package admin

import (
	"time"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// StatisticsService 统计服务
type StatisticsService struct{}

// NewStatisticsService 创建统计服务实例
func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

// DashboardData 仪表盘数据
type DashboardData struct {
	TotalUsers       int64   `json:"total_users"`
	TodayNewUsers    int64   `json:"today_new_users"`
	TotalDeposit     float64 `json:"total_deposit"`
	TotalWithdraw    float64 `json:"total_withdraw"`
	PendingWithdraw  int64   `json:"pending_withdraw"`
	PendingKYC       int64   `json:"pending_kyc"`
	TodayTrades      int64   `json:"today_trades"`
	TodayTradeAmount float64 `json:"today_trade_amount"`
}

// GetDashboard 获取仪表盘数据
func (s *StatisticsService) GetDashboard() (*DashboardData, error) {
	data := &DashboardData{}

	// 总用户数
	database.DB.Model(&model.User{}).Count(&data.TotalUsers)

	// 今日新增用户
	todayStart := time.Now().Truncate(24 * time.Hour).Unix()
	database.DB.Model(&model.User{}).Where("create_time >= ?", todayStart).Count(&data.TodayNewUsers)

	// 待审核提现
	database.DB.Model(&model.UsersWalletOut{}).Where("status = ?", 1).Count(&data.PendingWithdraw)

	// 待审核KYC
	database.DB.Model(&model.UserReal{}).Where("status = ?", 0).Count(&data.PendingKYC)

	// 今日交易数
	database.DB.Model(&model.Transaction{}).Where("create_time >= ?", time.Unix(todayStart, 0)).Count(&data.TodayTrades)

	return data, nil
}

// UserStatistics 用户统计数据
type UserStatistics struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Frozen     int64 `json:"frozen"`
	Verified   int64 `json:"verified"`
	Unverified int64 `json:"unverified"`
}

// GetUserStatistics 获取用户统计
func (s *StatisticsService) GetUserStatistics() (*UserStatistics, error) {
	stats := &UserStatistics{}

	database.DB.Model(&model.User{}).Count(&stats.Total)
	database.DB.Model(&model.User{}).Where("status = ?", 0).Count(&stats.Active)
	database.DB.Model(&model.User{}).Where("status = ?", 1).Count(&stats.Frozen)
	database.DB.Model(&model.User{}).Where("is_realname = ?", 2).Count(&stats.Verified)
	database.DB.Model(&model.User{}).Where("is_realname = ?", 1).Count(&stats.Unverified)

	return stats, nil
}

// TradeStatistics 交易统计数据
type TradeStatistics struct {
	SpotTotal    int64   `json:"spot_total"`
	SpotAmount   float64 `json:"spot_amount"`
	LeverTotal   int64   `json:"lever_total"`
	LeverAmount  float64 `json:"lever_amount"`
	MicroTotal   int64   `json:"micro_total"`
	MicroAmount  float64 `json:"micro_amount"`
}

// GetTradeStatistics 获取交易统计
func (s *StatisticsService) GetTradeStatistics() (*TradeStatistics, error) {
	stats := &TradeStatistics{}

	// 币币交易统计
	database.DB.Model(&model.Transaction{}).Count(&stats.SpotTotal)

	// 合约交易统计
	database.DB.Model(&model.LeverTransaction{}).Count(&stats.LeverTotal)

	// 秒合约统计
	database.DB.Model(&model.MicroOrder{}).Count(&stats.MicroTotal)

	return stats, nil
}

// FinanceStatistics 财务统计数据
type FinanceStatistics struct {
	TotalDeposit        float64 `json:"total_deposit"`
	TodayDeposit        float64 `json:"today_deposit"`
	TotalWithdraw       float64 `json:"total_withdraw"`
	TodayWithdraw       float64 `json:"today_withdraw"`
	PendingWithdraw     float64 `json:"pending_withdraw"`
	PendingWithdrawNum  int64   `json:"pending_withdraw_num"`
}

// GetFinanceStatistics 获取财务统计
func (s *StatisticsService) GetFinanceStatistics() (*FinanceStatistics, error) {
	stats := &FinanceStatistics{}
	todayStart := time.Now().Truncate(24 * time.Hour).Unix()

	// 待审核提现金额和数量
	database.DB.Model(&model.UsersWalletOut{}).Where("status = ?", 1).Count(&stats.PendingWithdrawNum)

	var pendingSum struct {
		Total float64
	}
	database.DB.Model(&model.UsersWalletOut{}).
		Select("COALESCE(SUM(number), 0) as total").
		Where("status = ?", 1).
		Scan(&pendingSum)
	stats.PendingWithdraw = pendingSum.Total

	// 今日提现金额
	var todayWithdrawSum struct {
		Total float64
	}
	database.DB.Model(&model.UsersWalletOut{}).
		Select("COALESCE(SUM(number), 0) as total").
		Where("status = ? AND create_time >= ?", 2, todayStart).
		Scan(&todayWithdrawSum)
	stats.TodayWithdraw = todayWithdrawSum.Total

	return stats, nil
}

// TopTrader 顶级交易者
type TopTrader struct {
	UserID      uint    `json:"user_id"`
	Account     string  `json:"account"`
	TradeCount  int64   `json:"trade_count"`
	TradeAmount float64 `json:"trade_amount"`
}

// GetTopTraders 获取顶级交易者
func (s *StatisticsService) GetTopTraders(limit int) ([]TopTrader, error) {
	var traders []TopTrader

	database.DB.Raw(`
		SELECT t.user_id, u.account_number as account, 
			   COUNT(*) as trade_count, 
			   COALESCE(SUM(t.number * t.price), 0) as trade_amount
		FROM transaction t
		LEFT JOIN users u ON t.user_id = u.id
		GROUP BY t.user_id
		ORDER BY trade_amount DESC
		LIMIT ?
	`, limit).Scan(&traders)

	return traders, nil
}

// CurrencyStatistics 币种统计
type CurrencyStatistics struct {
	CurrencyID   uint    `json:"currency_id"`
	CurrencyName string  `json:"currency_name"`
	TotalBalance float64 `json:"total_balance"`
	UserCount    int64   `json:"user_count"`
}

// GetCurrencyStatistics 获取币种统计
func (s *StatisticsService) GetCurrencyStatistics() ([]CurrencyStatistics, error) {
	var stats []CurrencyStatistics

	database.DB.Raw(`
		SELECT w.currency as currency_id, c.name as currency_name,
			   COALESCE(SUM(w.legal_balance + w.change_balance + w.lever_balance + w.micro_balance), 0) as total_balance,
			   COUNT(DISTINCT w.user_id) as user_count
		FROM users_wallet w
		LEFT JOIN currency c ON w.currency = c.id
		GROUP BY w.currency
		ORDER BY total_balance DESC
	`).Scan(&stats)

	return stats, nil
}
