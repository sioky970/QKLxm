package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"exchange-go/internal/model"
	"exchange-go/internal/pkg/database"
)

// ============ 导出功能 ============

// ExportUsers 导出用户列表
func ExportUsers(c *gin.Context) {
	// 获取筛选参数
	account := c.Query("account")
	status := c.Query("status")
	isReal := c.Query("is_real")

	db := database.DB.Model(&model.User{})

	if account != "" {
		db = db.Where("account_number LIKE ? OR phone LIKE ? OR email LIKE ?",
			"%"+account+"%", "%"+account+"%", "%"+account+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if isReal != "" {
		db = db.Where("is_realname = ?", isReal)
	}

	var users []model.User
	if err := db.Order("id DESC").Find(&users).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 创建CSV
	buf := new(bytes.Buffer)
	// 添加BOM头，支持Excel中文
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)

	// 写入表头
	header := []string{"ID", "账号", "手机号", "邮箱", "实名状态", "账户状态", "风控等级", "注册时间", "最后登录"}
	writer.Write(header)

	// 写入数据
	for _, user := range users {
		realStatus := "未认证"
		if user.IsRealname == 2 {
			realStatus = "已认证"
		}

		userStatus := "正常"
		if user.Status == 1 {
			userStatus = "冻结"
		}

		riskLevel := "正常"
		if user.Risk == 1 {
			riskLevel = "风控"
		}

		regTime := ""
		if user.Time > 0 {
			regTime = time.Unix(user.Time, 0).Format("2006-01-02 15:04:05")
		}

		lastTime := ""
		if user.LastTime > 0 {
			lastTime = time.Unix(user.LastTime, 0).Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", user.ID),
			user.AccountNumber,
			user.Phone,
			user.Email,
			realStatus,
			userStatus,
			riskLevel,
			regTime,
			lastTime,
		}
		writer.Write(row)
	}

	writer.Flush()

	// 设置响应头
	filename := fmt.Sprintf("users_%s.csv", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

// ExportLeverOrders 导出杠杆订单
func ExportLeverOrdersCSV(c *gin.Context) {
	// 获取筛选参数
	accountNumber := c.Query("account_number")
	currencyID := c.Query("currency_id")
	status := c.Query("status")

	db := database.DB.Table("lever_transaction AS lt").
		Select("lt.*, u.account_number, c1.name AS currency_name, c2.name AS legal_name").
		Joins("LEFT JOIN users u ON lt.user_id = u.id").
		Joins("LEFT JOIN currency c1 ON lt.currency = c1.id").
		Joins("LEFT JOIN currency c2 ON lt.legal = c2.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if currencyID != "" {
		db = db.Where("lt.currency = ?", currencyID)
	}
	if status != "" {
		db = db.Where("lt.status = ?", status)
	}

	type LeverExportItem struct {
		model.LeverTransaction
		AccountNumber string `json:"account_number"`
		CurrencyName  string `json:"currency_name"`
		LegalName     string `json:"legal_name"`
	}

	var orders []LeverExportItem
	if err := db.Order("lt.id DESC").Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 创建CSV
	buf := new(bytes.Buffer)
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)

	// 写入表头
	header := []string{"ID", "用户账号", "交易对", "方向", "杠杆倍数", "开仓价", "当前价", "数量", "保证金", "盈亏", "状态", "开仓时间", "平仓时间"}
	writer.Write(header)

	// 写入数据
	for _, order := range orders {
		direction := "做多"
		if order.Type == 2 {
			direction = "做空"
		}

		orderStatus := "持仓中"
		if order.Status == 1 {
			orderStatus = "已平仓"
		} else if order.Status == 2 {
			orderStatus = "已取消"
		}

		createTime := ""
		if order.CreateTime > 0 {
			createTime = time.Unix(order.CreateTime, 0).Format("2006-01-02 15:04:05")
		}

		completeTime := ""
		if order.CompleteTime > 0 {
			completeTime = time.Unix(int64(order.CompleteTime), 0).Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", order.ID),
			order.AccountNumber,
			fmt.Sprintf("%s/%s", order.CurrencyName, order.LegalName),
			direction,
			fmt.Sprintf("%d", order.Multiple),
			fmt.Sprintf("%.8f", order.Price),
			fmt.Sprintf("%.8f", order.UpdatePrice), // 使用UpdatePrice代替CurrentPrice
			fmt.Sprintf("%.8f", order.Number),
			fmt.Sprintf("%.8f", order.CautionMoney),
			fmt.Sprintf("%.8f", order.FactProfits),
			orderStatus,
			createTime,
			completeTime,
		}
		writer.Write(row)
	}

	writer.Flush()

	filename := fmt.Sprintf("lever_orders_%s.csv", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

// ExportC2COrders 导出C2C交易
func ExportC2COrders(c *gin.Context) {
	// 获取筛选参数
	accountNumber := c.Query("account_number")
	sellerNumber := c.Query("seller_number")
	orderType := c.Query("type")

	db := database.DB.Table("c2c_deal AS cd").
		Select("cd.*, u1.account_number AS buyer_account, u2.account_number AS seller_account, c.name AS currency_name, cds.type AS deal_type").
		Joins("LEFT JOIN users u1 ON cd.user_id = u1.id").
		Joins("LEFT JOIN users u2 ON cd.seller_id = u2.id").
		Joins("LEFT JOIN c2c_deal_send cds ON cd.legal_deal_send_id = cds.id").
		Joins("LEFT JOIN currency c ON cds.currency_id = c.id")

	if accountNumber != "" {
		db = db.Where("u1.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if sellerNumber != "" {
		db = db.Where("u2.account_number LIKE ?", "%"+sellerNumber+"%")
	}
	if orderType != "" {
		db = db.Where("cds.type = ?", orderType)
	}

	type C2CExportItem struct {
		model.C2cDeal
		BuyerAccount  string `json:"buyer_account"`
		SellerAccount string `json:"seller_account"`
		CurrencyName  string `json:"currency_name"`
		DealType      string `json:"deal_type"`
	}

	var orders []C2CExportItem
	if err := db.Order("cd.id DESC").Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 创建CSV
	buf := new(bytes.Buffer)
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)

	// 写入表头
	header := []string{"ID", "买家账号", "卖家账号", "币种", "类型", "数量", "状态", "创建时间"}
	writer.Write(header)

	// 写入数据
	for _, order := range orders {
		dealType := "买入"
		if order.DealType == "sell" {
			dealType = "卖出"
		}

		// is_sure: 0未完成 1已完成 2取消 3已付款
		orderStatus := "未完成"
		switch order.IsSure {
		case 1:
			orderStatus = "已完成"
		case 2:
			orderStatus = "已取消"
		case 3:
			orderStatus = "已付款"
		}

		createTime := ""
		if order.CreateTime > 0 {
			createTime = time.Unix(order.CreateTime, 0).Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", order.ID),
			order.BuyerAccount,
			order.SellerAccount,
			order.CurrencyName,
			dealType,
			fmt.Sprintf("%.5f", order.Number),
			orderStatus,
			createTime,
		}
		writer.Write(row)
	}

	writer.Flush()

	filename := fmt.Sprintf("c2c_orders_%s.csv", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

// ExportLegalOrders 导出法币交易
func ExportLegalOrders(c *gin.Context) {
	// 获取筛选参数
	accountNumber := c.Query("account_number")
	sellerNumber := c.Query("seller_number")
	orderType := c.Query("type")

	db := database.DB.Table("legal_deal AS ld").
		Select("ld.*, u1.account_number AS buyer_account, u2.account_number AS seller_account, c.name AS currency_name, lds.type AS deal_type").
		Joins("LEFT JOIN users u1 ON ld.user_id = u1.id").
		Joins("LEFT JOIN users u2 ON ld.seller_id = u2.id").
		Joins("LEFT JOIN legal_deal_send lds ON ld.legal_deal_send_id = lds.id").
		Joins("LEFT JOIN currency c ON lds.currency_id = c.id")

	if accountNumber != "" {
		db = db.Where("u1.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if sellerNumber != "" {
		db = db.Where("u2.account_number LIKE ?", "%"+sellerNumber+"%")
	}
	if orderType != "" {
		db = db.Where("lds.type = ?", orderType)
	}

	type LegalExportItem struct {
		model.LegalDeal
		BuyerAccount  string  `json:"buyer_account"`
		SellerAccount string  `json:"seller_account"`
		CurrencyName  string  `json:"currency_name"`
		DealType      string  `json:"deal_type"`
		Price         float64 `json:"price"`
		TotalPrice    float64 `json:"total_price"`
	}

	var orders []LegalExportItem
	if err := db.Order("ld.id DESC").Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 创建CSV
	buf := new(bytes.Buffer)
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)

	// 写入表头
	header := []string{"ID", "买家账号", "卖家账号", "币种", "类型", "数量", "单价", "总价", "是否付款", "是否确认", "状态", "创建时间"}
	writer.Write(header)

	// 写入数据
	for _, order := range orders {
		dealType := "买入"
		if order.DealType == "sell" {
			dealType = "卖出"
		}

		// IsSure: 0未完成 1已完成 2取消
		isPay := "否"
		if order.IsSure == 1 || order.IsSure == 3 {
			isPay = "是"
		}

		isSure := "否"
		if order.IsSure == 1 {
			isSure = "是"
		}

		orderStatus := "进行中"
		if order.IsSure == 1 {
			orderStatus = "已完成"
		} else if order.IsSure == 2 {
			orderStatus = "已取消"
		}

		createTime := ""
		if order.CreateTime > 0 {
			createTime = time.Unix(order.CreateTime, 0).Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", order.ID),
			order.BuyerAccount,
			order.SellerAccount,
			order.CurrencyName,
			dealType,
			fmt.Sprintf("%.8f", order.Number),
			fmt.Sprintf("%.8f", order.Price),
			fmt.Sprintf("%.8f", order.TotalPrice),
			isPay,
			isSure,
			orderStatus,
			createTime,
		}
		writer.Write(row)
	}

	writer.Flush()

	filename := fmt.Sprintf("legal_orders_%s.csv", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

// ExportAccountLogs 导出账户流水
func ExportAccountLogs(c *gin.Context) {
	accountNumber := c.Query("account_number")
	logType := c.Query("type")
	currencyID := c.Query("currency_id")

	db := database.DB.Table("account_log AS al").
		Select("al.*, u.account_number, c.name AS currency_name").
		Joins("LEFT JOIN users u ON al.user_id = u.id").
		Joins("LEFT JOIN currency c ON al.currency = c.id")

	if accountNumber != "" {
		db = db.Where("u.account_number LIKE ?", "%"+accountNumber+"%")
	}
	if logType != "" {
		db = db.Where("al.type = ?", logType)
	}
	if currencyID != "" {
		db = db.Where("al.currency = ?", currencyID)
	}

	type LogExportItem struct {
		model.AccountLog
		AccountNumber string `json:"account_number"`
		CurrencyName  string `json:"currency_name"`
	}

	var logs []LogExportItem
	if err := db.Order("al.id DESC").Limit(10000).Find(&logs).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 创建CSV
	buf := new(bytes.Buffer)
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)

	// 写入表头
	header := []string{"ID", "用户账号", "币种", "变动金额", "类型", "备注", "时间"}
	writer.Write(header)

	// 写入数据
	for _, log := range logs {
		typeName := "未知"
		if name, ok := AccountLogTypeMap[log.Type]; ok {
			typeName = name
		}

		createTime := ""
		if log.CreatedTime > 0 {
			createTime = time.Unix(log.CreatedTime, 0).Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", log.ID),
			log.AccountNumber,
			log.CurrencyName,
			fmt.Sprintf("%.8f", log.Value),
			typeName,
			log.Info,
			createTime,
		}
		writer.Write(row)
	}

	writer.Flush()

	filename := fmt.Sprintf("account_logs_%s.csv", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}
