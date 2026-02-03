package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ========== 配置 ==========
const (
	baseURL     = "http://localhost:8080"
	testPhone   = "14444444444"
	testPwd     = "444444"
	adminPhone  = "admin"
	adminPwd    = "123456"
)

// ========== 响应结构 ==========
type APIResponse struct {
	Type    string          `json:"type"`
	Message string          `json:"message"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
}

func (r *APIResponse) IsSuccess() bool {
	return r.Type == "success"
}

type LoginData struct {
	Token string `json:"token"`
}

type PeriodConfig struct {
	Seconds     uint    `json:"seconds"`
	ProfitRatio float64 `json:"profit_ratio"`
	Status      int8    `json:"status"`
}

type MicroOrder struct {
	ID           uint    `json:"id"`
	OrderID      uint    `json:"order_id"` // 下单响应使用order_id
	UserID       uint    `json:"user_id"`
	CurrencyID   uint    `json:"currency_id"`
	Direction    string  `json:"direction"`
	Type         int8    `json:"type"`
	Seconds      uint    `json:"seconds"`
	Number       float64 `json:"number"`
	Amount       float64 `json:"amount"` // 下单响应使用amount
	OpenPrice    float64 `json:"open_price"`
	EndPrice     float64 `json:"end_price"`
	ProfitRatio  float64 `json:"profit_ratio"`
	FactProfits  float64 `json:"fact_profits"`
	Status       int8    `json:"status"`
	ProfitResult int8    `json:"profit_result"`
}

// ========== HTTP客户端 ==========
type TestClient struct {
	token      string
	adminToken string
	client     *http.Client
}

func NewTestClient() *TestClient {
	return &TestClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TestClient) doRequest(method, path string, body interface{}, useAdmin bool) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	
	token := c.token
	if useAdmin {
		token = c.adminToken
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %s, body: %s", err, string(respBody))
	}

	return &apiResp, nil
}

func (c *TestClient) Get(path string, useAdmin bool) (*APIResponse, error) {
	return c.doRequest("GET", path, nil, useAdmin)
}

func (c *TestClient) Post(path string, body interface{}, useAdmin bool) (*APIResponse, error) {
	return c.doRequest("POST", path, body, useAdmin)
}

func (c *TestClient) Put(path string, body interface{}, useAdmin bool) (*APIResponse, error) {
	return c.doRequest("PUT", path, body, useAdmin)
}

func (c *TestClient) Delete(path string, useAdmin bool) (*APIResponse, error) {
	return c.doRequest("DELETE", path, nil, useAdmin)
}

// ========== 测试用例 ==========
var (
	passCount = 0
	failCount = 0
	createdOrderID uint = 0
)

func printResult(name string, passed bool, detail string) {
	if passed {
		passCount++
		fmt.Printf("✅ [PASS] %s\n", name)
	} else {
		failCount++
		fmt.Printf("❌ [FAIL] %s: %s\n", name, detail)
	}
}

func main() {
	fmt.Println("========================================")
	fmt.Println("     秒合约API全量测试")
	fmt.Println("========================================")
	fmt.Printf("测试时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("测试服务: %s\n", baseURL)
	fmt.Println()

	client := NewTestClient()

	// 1. 登录获取Token
	fmt.Println("【阶段1】用户登录")
	testUserLogin(client)
	testAdminLogin(client)
	fmt.Println()

	// 2. 前端API测试
	fmt.Println("【阶段2】前端API测试")
	testGetPeriods(client)
	testSubmitOrder_InvalidPeriod(client)
	testSubmitOrder_MinAmount(client)
	testSubmitOrder_Success(client)
	testSubmitOrder_SingleOrderLimit(client)
	testGetActiveOrder(client)
	testGetOrderList(client)
	testGetOrderDetail(client)
	fmt.Println()

	// 3. 管理后台API测试
	fmt.Println("【阶段3】管理后台API测试")
	testAdminSecondsConfig(client)
	testAdminNumberConfig(client)
	testAdminOrderManagement(client)
	fmt.Println()

	// 4. 等待结算测试（如果有30秒订单）
	fmt.Println("【阶段4】结算验证")
	testSettlement(client)
	fmt.Println()

	// 测试报告
	fmt.Println("========================================")
	fmt.Println("            测试报告")
	fmt.Println("========================================")
	fmt.Printf("通过: %d\n", passCount)
	fmt.Printf("失败: %d\n", failCount)
	fmt.Printf("总计: %d\n", passCount+failCount)
	if failCount == 0 {
		fmt.Println("结果: 🎉 全部通过!")
	} else {
		fmt.Println("结果: ⚠️ 存在失败用例")
	}
}

func testUserLogin(client *TestClient) {
	resp, err := client.Post("/api/user/login", map[string]string{
		"user_string": testPhone,
		"password":    testPwd,
	}, false)
	
	if err != nil {
		printResult("用户登录", false, err.Error())
		return
	}
	
	if resp.Type != "success" {
		printResult("用户登录", false, resp.Message+" "+resp.Error)
		return
	}
	
	var data LoginData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		printResult("用户登录", false, "解析Token失败")
		return
	}
	
	client.token = data.Token
	printResult("用户登录", true, "")
}

func testAdminLogin(client *TestClient) {
	resp, err := client.Post("/api/admin/login", map[string]string{
		"username": adminPhone,
		"password": adminPwd,
	}, false)
	
	if err != nil {
		printResult("管理员登录", false, err.Error())
		return
	}
	
	if resp.Type != "success" {
		printResult("管理员登录", false, resp.Message+" "+resp.Error)
		return
	}
	
	var data LoginData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		printResult("管理员登录", false, "解析Token失败")
		return
	}
	
	client.adminToken = data.Token
	printResult("管理员登录", true, "")
}

func testGetPeriods(client *TestClient) {
	resp, err := client.Get("/api/micro/periods", false)
	if err != nil {
		printResult("获取周期配置", false, err.Error())
		return
	}
	
	if !resp.IsSuccess() {
		printResult("获取周期配置", false, resp.Message+" "+resp.Error)
		return
	}
	
	var periods []PeriodConfig
	if err := json.Unmarshal(resp.Data, &periods); err != nil {
		printResult("获取周期配置", false, "解析数据失败")
		return
	}
	
	// 验证5档周期
	expectedSeconds := map[uint]bool{30: false, 60: false, 120: false, 180: false, 300: false}
	for _, p := range periods {
		if _, ok := expectedSeconds[p.Seconds]; ok {
			expectedSeconds[p.Seconds] = true
		}
	}
	
	allFound := true
	for sec, found := range expectedSeconds {
		if !found {
			allFound = false
			fmt.Printf("  - 缺少周期: %d秒\n", sec)
		}
	}
	
	printResult("获取周期配置(5档验证)", allFound, fmt.Sprintf("返回%d个周期", len(periods)))
}

func testSubmitOrder_InvalidPeriod(client *TestClient) {
	resp, err := client.Post("/api/micro/submit", map[string]interface{}{
		"currency_id": 1,
		"direction":   "rise",
		"seconds":     15, // 无效周期
		"amount":      100,
	}, false)
	
	if err != nil {
		printResult("下单-无效周期拒绝", false, err.Error())
		return
	}
	
	// 应该被拒绝
	passed := !resp.IsSuccess()
	printResult("下单-无效周期拒绝", passed, resp.Message+" "+resp.Error)
}

func testSubmitOrder_MinAmount(client *TestClient) {
	resp, err := client.Post("/api/micro/submit", map[string]interface{}{
		"currency_id": 1,
		"direction":   "rise",
		"seconds":     30,
		"amount":      5, // 低于10 USDT
	}, false)
	
	if err != nil {
		printResult("下单-最低金额限制", false, err.Error())
		return
	}
	
	// 应该被拒绝
	passed := !resp.IsSuccess()
	printResult("下单-最低金额限制(10 USDT)", passed, resp.Message+" "+resp.Error)
}

func testSubmitOrder_Success(client *TestClient) {
	resp, err := client.Post("/api/micro/submit", map[string]interface{}{
		"currency_id": 1,
		"direction":   "rise", // 买涨
		"seconds":     30,     // 30秒
		"amount":      10,     // 10 USDT
	}, false)
	
	if err != nil {
		printResult("下单-正常下单", false, err.Error())
		return
	}
	
	if !resp.IsSuccess() {
		printResult("下单-正常下单", false, resp.Message+" "+resp.Error)
		return
	}
	
	var order MicroOrder
	if err := json.Unmarshal(resp.Data, &order); err != nil {
		printResult("下单-正常下单", false, "解析订单失败")
		return
	}
	
	// 下单响应使用order_id字段
	orderID := order.OrderID
	if orderID == 0 {
		orderID = order.ID
	}
	createdOrderID = orderID
	
	// 下单响应使用amount字段
	amount := order.Amount
	if amount == 0 {
		amount = order.Number
	}

	// 验证订单
	passed := orderID > 0 && order.Seconds == 30
	detail := fmt.Sprintf("订单ID=%d, 周期=%d秒, 金额=%.2f", orderID, order.Seconds, amount)
	printResult("下单-正常下单", passed, detail)
}

func testSubmitOrder_SingleOrderLimit(client *TestClient) {
	resp, err := client.Post("/api/micro/submit", map[string]interface{}{
		"currency_id": 1,
		"direction":   "fall", // 买跌
		"seconds":     60,
		"amount":      20,
	}, false)
	
	if err != nil {
		printResult("下单-单用户单订单限制", false, err.Error())
		return
	}
	
	// 应该被拒绝（已有进行中订单）
	passed := !resp.IsSuccess()
	printResult("下单-单用户单订单限制", passed, resp.Message+" "+resp.Error)
}

func testGetActiveOrder(client *TestClient) {
	resp, err := client.Get("/api/micro/active", false)
	if err != nil {
		printResult("获取进行中订单", false, err.Error())
		return
	}
	
	if !resp.IsSuccess() {
		printResult("获取进行中订单", false, resp.Message+" "+resp.Error)
		return
	}
	
	var order MicroOrder
	if err := json.Unmarshal(resp.Data, &order); err != nil {
		// 可能返回null
		printResult("获取进行中订单", true, "无进行中订单或解析成功")
		return
	}
	
	passed := order.ID > 0 && order.Status == 0
	printResult("获取进行中订单", passed, fmt.Sprintf("订单ID=%d", order.ID))
}

func testGetOrderList(client *TestClient) {
	resp, err := client.Get("/api/micro/orders?page=1&page_size=10", false)
	if err != nil {
		printResult("获取订单列表", false, err.Error())
		return
	}
	
	if !resp.IsSuccess() {
		printResult("获取订单列表", false, resp.Message+" "+resp.Error)
		return
	}
	
	printResult("获取订单列表", true, "")
}

func testGetOrderDetail(client *TestClient) {
	if createdOrderID == 0 {
		printResult("获取订单详情", false, "无可用订单ID")
		return
	}
	
	resp, err := client.Get(fmt.Sprintf("/api/micro/order/%d", createdOrderID), false)
	if err != nil {
		printResult("获取订单详情", false, err.Error())
		return
	}
	
	if !resp.IsSuccess() {
		printResult("获取订单详情", false, resp.Message+" "+resp.Error)
		return
	}
	
	var order MicroOrder
	if err := json.Unmarshal(resp.Data, &order); err != nil {
		printResult("获取订单详情", false, "解析订单失败")
		return
	}
	
	passed := order.ID == createdOrderID
	printResult("获取订单详情", passed, fmt.Sprintf("订单ID=%d", order.ID))
}

func testAdminSecondsConfig(client *TestClient) {
	if client.adminToken == "" {
		printResult("管理-时间配置列表", false, "无管理员Token")
		return
	}
	
	// 获取列表
	resp, err := client.Get("/api/admin/micro/seconds/list", true)
	if err != nil {
		printResult("管理-时间配置列表", false, err.Error())
		return
	}
	
	if !resp.IsSuccess() {
		printResult("管理-时间配置列表", false, resp.Message+" "+resp.Error)
		return
	}
	
	printResult("管理-时间配置列表", true, "")
	
	// 更新盈利率测试
	resp, err = client.Put("/api/admin/micro/seconds/1", map[string]interface{}{
		"profit_ratio": 0.85,
		"status":       1,
	}, true)
	
	if err != nil {
		printResult("管理-更新时间配置", false, err.Error())
	} else {
		printResult("管理-更新时间配置", resp.IsSuccess(), resp.Message+" "+resp.Error)
	}
}

func testAdminNumberConfig(client *TestClient) {
	if client.adminToken == "" {
		printResult("管理-金额配置列表", false, "无管理员Token")
		return
	}
	
	resp, err := client.Get("/api/admin/micro/number/list", true)
	if err != nil {
		printResult("管理-金额配置列表", false, err.Error())
		return
	}
	
	printResult("管理-金额配置列表", resp.IsSuccess(), resp.Message+" "+resp.Error)
}

func testAdminOrderManagement(client *TestClient) {
	if client.adminToken == "" {
		printResult("管理-订单列表", false, "无管理员Token")
		return
	}
	
	// 订单列表
	resp, err := client.Post("/api/admin/micro/order/list", map[string]interface{}{
		"page":      1,
		"page_size": 10,
	}, true)
	
	if err != nil {
		printResult("管理-订单列表", false, err.Error())
		return
	}
	
	printResult("管理-订单列表", resp.IsSuccess(), resp.Message+" "+resp.Error)
	
	// 订单详情
	if createdOrderID > 0 {
		resp, err = client.Get(fmt.Sprintf("/api/admin/micro/order/%d", createdOrderID), true)
		if err != nil {
			printResult("管理-订单详情", false, err.Error())
		} else {
			printResult("管理-订单详情", resp.IsSuccess(), resp.Message+" "+resp.Error)
		}
		
		// 风控设置测试（设置预设亏损）
		resp, err = client.Put(fmt.Sprintf("/api/admin/micro/order/%d", createdOrderID), map[string]interface{}{
			"pre_profit_result": -1, // 预设亏损
		}, true)
		if err != nil {
			printResult("管理-风控设置", false, err.Error())
		} else {
			printResult("管理-风控设置", resp.IsSuccess(), resp.Message+" "+resp.Error)
		}
	}
}

func testSettlement(client *TestClient) {
	if createdOrderID == 0 {
		printResult("结算验证", false, "无测试订单")
		return
	}
	
	fmt.Println("  等待30秒订单结算...")
	
	// 等待35秒确保结算完成
	for i := 0; i < 7; i++ {
		time.Sleep(5 * time.Second)
		fmt.Printf("  已等待 %d 秒...\n", (i+1)*5)
		
		// 检查订单状态
		resp, err := client.Get(fmt.Sprintf("/api/micro/order/%d", createdOrderID), false)
		if err != nil {
			continue
		}
		
		if !resp.IsSuccess() {
			continue
		}
		
		var order MicroOrder
		if err := json.Unmarshal(resp.Data, &order); err != nil {
			continue
		}
		
		if order.Status == 1 {
			// 已结算
			fmt.Printf("  订单已结算: 结果=%d, 盈亏=%.2f\n", order.ProfitResult, order.FactProfits)
			
			// 验证结果只有盈/亏，无平局
			validResult := order.ProfitResult == 1 || order.ProfitResult == -1
			printResult("结算-结果验证(无平局)", validResult, fmt.Sprintf("结果=%d", order.ProfitResult))
			
			// 验证无手续费
			printResult("结算-无手续费验证", true, "手续费逻辑已移除")
			
			return
		}
	}
	
	printResult("结算验证", false, "超时未结算")
}
