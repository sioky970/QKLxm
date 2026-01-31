package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/response"
)

// ============ 账户流水管理 ============

// 流水类型定义
const (
	LOG_ADMIN_LEGAL_BALANCE       = 1  // 后台调节法币账户余额
	LOG_ADMIN_LOCK_LEGAL_BALANCE  = 2  // 后台调节法币账户锁定余额
	LOG_ADMIN_CHANGE_BALANCE      = 3  // 后台调节币币账户余额
	LOG_ADMIN_LOCK_CHANGE_BALANCE = 4  // 后台调节币币账户锁定余额
	LOG_ADMIN_LEVER_BALANCE       = 5  // 后台调节合约账户余额
	LOG_ADMIN_LOCK_LEVER_BALANCE  = 6  // 后台调节合约账户锁定余额
	LOG_ADMIN_MICRO_BALANCE       = 7  // 后台调节秒合约账户余额
	LOG_ADMIN_LOCK_MICRO_BALANCE  = 8  // 后台调节秒合约账户锁定余额
	LOG_TRANSFER_LEGAL_TO_CHANGE  = 9  // 法币划转币币
	LOG_TRANSFER_CHANGE_TO_LEGAL  = 10 // 币币划转法币
	LOG_TRANSFER_CHANGE_TO_LEVER  = 11 // 币币划转合约
	LOG_TRANSFER_LEVER_TO_CHANGE  = 12 // 合约划转币币
	LOG_TRANSFER_CHANGE_TO_MICRO  = 13 // 币币划转秒合约
	LOG_TRANSFER_MICRO_TO_CHANGE  = 14 // 秒合约划转币币
	LOG_TRANSFER_LEGAL_TO_LEVER   = 15 // 法币划转合约
	LOG_TRANSFER_LEVER_TO_LEGAL   = 16 // 合约划转法币
	LOG_TRANSFER_LEGAL_TO_MICRO   = 17 // 法币划转秒合约
	LOG_TRANSFER_MICRO_TO_LEGAL   = 18 // 秒合约划转法币
	LOG_SPOT_BUY                  = 22 // 币币买入扣款
	LOG_SPOT_SELL                 = 23 // 币币卖出扣币
	LOG_SPOT_BUY_GET              = 25 // 币币买入到账
	LOG_SPOT_SELL_GET             = 26 // 币币卖出到账
	LOG_CHARGE_SUCCESS            = 31 // 充值成功
	LOG_WITHDRAW_APPLY            = 32 // 提币申请
	LOG_WITHDRAW_SUCCESS          = 34 // 提币成功
	LOG_WITHDRAW_REJECT           = 35 // 提币被拒绝
	LOG_LEVER_OPEN                = 40 // 合约开仓
	LOG_LEVER_CLOSE               = 41 // 合约平仓
	LOG_LEVER_FEE                 = 42 // 合约手续费
	LOG_MICRO_BET                 = 45 // 秒合约下单
	LOG_MICRO_WIN                 = 46 // 秒合约盈利
	LOG_MICRO_LOSE                = 47 // 秒合约亏损
	LOG_LEVER_CLOSE_ADMIN         = 50 // 管理员强平
	LOG_LEGAL_BUY                 = 60 // 法币买入
	LOG_LEGAL_SELL                = 61 // 法币卖出
	LOG_C2C_BUY                   = 62 // C2C买入
	LOG_C2C_SELL                  = 63 // C2C卖出
)

// AccountLogTypeMap 流水类型映射
var AccountLogTypeMap = map[int]string{
	1:  "后台调节法币账户余额",
	2:  "后台调节法币账户锁定余额",
	3:  "后台调节币币账户余额",
	4:  "后台调节币币账户锁定余额",
	5:  "后台调节合约账户余额",
	6:  "后台调节合约账户锁定余额",
	7:  "后台调节秒合约账户余额",
	8:  "后台调节秒合约账户锁定余额",
	9:  "法币划转币币",
	10: "币币划转法币",
	11: "币币划转合约",
	12: "合约划转币币",
	13: "币币划转秒合约",
	14: "秒合约划转币币",
	15: "法币划转合约",
	16: "合约划转法币",
	17: "法币划转秒合约",
	18: "秒合约划转法币",
	22: "币币买入扣款",
	23: "币币卖出扣币",
	25: "币币买入到账",
	26: "币币卖出到账",
	31: "充值成功",
	32: "提币申请",
	34: "提币成功",
	35: "提币被拒绝",
	40: "合约开仓",
	41: "合约平仓",
	42: "合约手续费",
	45: "秒合约下单",
	46: "秒合约盈利",
	47: "秒合约亏损",
	50: "管理员强平",
	60: "法币买入",
	61: "法币卖出",
	62: "C2C买入",
	63: "C2C卖出",
}

// AccountLogListItem 流水列表项
type AccountLogListItem struct {
	model.AccountLog
	AccountNumber string `json:"account_number"`
	CurrencyName  string `json:"currency_name"`
	TypeName      string `json:"type_name"`
}

// GetAccountLogListAdvanced 获取账户流水列表
func GetAccountLogListAdvanced(c *gin.Context) {
	var req struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		AccountNumber string `json:"account_number"`
		Type          *int   `json:"type"`
		CurrencyID    uint   `json:"currency_id"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	db := database.DB.Table("account_log AS al").
		Select("al.*, u.account_number, c.name AS currency_name").
		Joins("LEFT JOIN users u ON al.user_id = u.id").
		Joins("LEFT JOIN currency c ON al.currency = c.id")

	if req.AccountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+req.AccountNumber+"%")
	}
	if req.Type != nil {
		db = db.Where("al.type = ?", *req.Type)
	}
	if req.CurrencyID > 0 {
		db = db.Where("al.currency = ?", req.CurrencyID)
	}
	if req.StartTime != "" {
		t, _ := time.Parse("2006-01-02", req.StartTime)
		db = db.Where("al.created_time >= ?", t.Unix())
	}
	if req.EndTime != "" {
		t, _ := time.Parse("2006-01-02", req.EndTime)
		db = db.Where("al.created_time <= ?", t.Add(24*time.Hour).Unix())
	}

	var total int64
	db.Count(&total)

	var list []AccountLogListItem
	offset := req.PageSize * (req.Page - 1)
	if err := db.Limit(req.PageSize).Offset(offset).Order("al.id DESC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	// 填充类型名称
	for i := range list {
		if name, ok := AccountLogTypeMap[list[i].Type]; ok {
			list[i].TypeName = name
		} else {
			list[i].TypeName = "未知类型"
		}
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

// GetAccountLogDetail 获取流水详情
func GetAccountLogDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var log model.AccountLog
	if err := database.DB.Where("id = ?", id).First(&log).Error; err != nil {
		response.Error(c, "记录不存在")
		return
	}

	// 获取用户信息
	var user model.User
	database.DB.Select("id, account_number, phone, email").Where("id = ?", log.UserID).First(&user)

	// 获取币种信息
	var currency model.Currency
	database.DB.Select("id, name").Where("id = ?", log.CurrencyID).First(&currency)

	typeName := "未知类型"
	if name, ok := AccountLogTypeMap[log.Type]; ok {
		typeName = name
	}

	result := map[string]interface{}{
		"log":           log,
		"user":          user,
		"currency":      currency,
		"type_name":     typeName,
	}

	response.Success(c, "获取成功", result)
}

// GetAccountLogTypes 获取流水类型列表
func GetAccountLogTypes(c *gin.Context) {
	var types []map[string]interface{}
	for k, v := range AccountLogTypeMap {
		types = append(types, map[string]interface{}{
			"type": k,
			"name": v,
		})
	}

	response.Success(c, "获取成功", types)
}

// GetProfitStatistics 盈亏统计
func GetProfitStatistics(c *gin.Context) {
	var result struct {
		TotalProfit   float64 `json:"total_profit"`   // 总盈利
		TotalLoss     float64 `json:"total_loss"`     // 总亏损
		NetProfit     float64 `json:"net_profit"`     // 净盈利
		LeverProfit   float64 `json:"lever_profit"`   // 合约盈利
		LeverLoss     float64 `json:"lever_loss"`     // 合约亏损
		MicroProfit   float64 `json:"micro_profit"`   // 秒合约盈利
		MicroLoss     float64 `json:"micro_loss"`     // 秒合约亏损
		TodayProfit   float64 `json:"today_profit"`   // 今日盈利
		TodayLoss     float64 `json:"today_loss"`     // 今日亏损
		WeekProfit    float64 `json:"week_profit"`    // 本周盈利
		WeekLoss      float64 `json:"week_loss"`      // 本周亏损
		MonthProfit   float64 `json:"month_profit"`   // 本月盈利
		MonthLoss     float64 `json:"month_loss"`     // 本月亏损
	}

	now := time.Now()
	todayStart := now.Truncate(24 * time.Hour).Unix()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())).Truncate(24 * time.Hour).Unix()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	// 合约盈亏
	database.DB.Model(&model.LeverTransaction{}).
		Where("status = 1 AND fact_profits > 0").
		Select("COALESCE(SUM(fact_profits), 0)").
		Scan(&result.LeverProfit)
	database.DB.Model(&model.LeverTransaction{}).
		Where("status = 1 AND fact_profits < 0").
		Select("COALESCE(SUM(ABS(fact_profits)), 0)").
		Scan(&result.LeverLoss)

	// 秒合约盈亏
	database.DB.Model(&model.MicroOrder{}).
		Where("status = 1 AND profit > 0").
		Select("COALESCE(SUM(profit), 0)").
		Scan(&result.MicroProfit)
	database.DB.Model(&model.MicroOrder{}).
		Where("status = 1 AND profit < 0").
		Select("COALESCE(SUM(ABS(profit)), 0)").
		Scan(&result.MicroLoss)

	// 今日盈亏
	database.DB.Model(&model.AccountLog{}).
		Where("type IN (46) AND created_time >= ?", todayStart).
		Select("COALESCE(SUM(value), 0)").
		Scan(&result.TodayProfit)
	database.DB.Model(&model.AccountLog{}).
		Where("type IN (47) AND created_time >= ?", todayStart).
		Select("COALESCE(SUM(ABS(value)), 0)").
		Scan(&result.TodayLoss)

	// 本周盈亏
	database.DB.Model(&model.AccountLog{}).
		Where("type IN (46) AND created_time >= ?", weekStart).
		Select("COALESCE(SUM(value), 0)").
		Scan(&result.WeekProfit)
	database.DB.Model(&model.AccountLog{}).
		Where("type IN (47) AND created_time >= ?", weekStart).
		Select("COALESCE(SUM(ABS(value)), 0)").
		Scan(&result.WeekLoss)

	// 本月盈亏
	database.DB.Model(&model.AccountLog{}).
		Where("type IN (46) AND created_time >= ?", monthStart).
		Select("COALESCE(SUM(value), 0)").
		Scan(&result.MonthProfit)
	database.DB.Model(&model.AccountLog{}).
		Where("type IN (47) AND created_time >= ?", monthStart).
		Select("COALESCE(SUM(ABS(value)), 0)").
		Scan(&result.MonthLoss)

	// 汇总
	result.TotalProfit = result.LeverProfit + result.MicroProfit
	result.TotalLoss = result.LeverLoss + result.MicroLoss
	result.NetProfit = result.TotalProfit - result.TotalLoss

	response.Success(c, "获取成功", result)
}

// GetUserAccountLogs 获取用户流水
func GetUserAccountLogs(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("user_id"), 10, 32)

	var req struct {
		Page     int  `json:"page"`
		PageSize int  `json:"page_size"`
		Type     *int `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	db := database.DB.Model(&model.AccountLog{}).Where("user_id = ?", userID)
	if req.Type != nil {
		db = db.Where("type = ?", *req.Type)
	}

	var total int64
	db.Count(&total)

	var list []model.AccountLog
	offset := req.PageSize * (req.Page - 1)
	if err := db.Limit(req.PageSize).Offset(offset).Order("id DESC").Find(&list).Error; err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}
