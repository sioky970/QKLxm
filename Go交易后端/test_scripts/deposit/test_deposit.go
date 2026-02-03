package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	baseURL       = "http://localhost:8080"
	testUsername  = "deposit_test@test.com"
	testPassword  = "123456"
	adminUsername = "admin"
	adminPassword = "admin123"
)

// 全局测试状态
var (
	userToken      string
	adminToken     string
	testOrderID    uint
	testOrderNo    string
	testAddressID  uint
	initialBalance float64
)

// Response 通用响应结构
type Response struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// LayuiResponse Layui格式响应
type LayuiResponse struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Count int64       `json:"count"`
	Data  interface{} `json:"data"`
}

// DepositAddress 充值地址
type DepositAddress struct {
	ID         uint   `json:"id"`
	Network    string `json:"network"`
	Address    string `json:"address"`
	QrCode     string `json:"qr_code"`
	Status     int8   `json:"status"`
	Sort       int    `json:"sort"`
	CreateTime int64  `json:"create_time"`
}

// DepositOrder 充值订单
type DepositOrder struct {
	ID          uint    `json:"id"`
	OrderNo     string  `json:"order_no"`
	UserID      uint    `json:"user_id"`
	Network     string  `json:"network"`
	Address     string  `json:"address"`
	Amount      float64 `json:"amount"`
	Screenshot  string  `json:"screenshot"`
	Status      int8    `json:"status"`
	AdminID     uint    `json:"admin_id"`
	AdminRemark string  `json:"admin_remark"`
	CreateTime  int64   `json:"create_time"`
	UpdateTime  int64   `json:"update_time"`
	ReviewTime  int64   `json:"review_time"`
	ExpireTime  int64   `json:"expire_time"`
}

// AssetOverview 资产概览
type AssetOverview struct {
	TotalBalance float64 `json:"total_balance"`
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("                    USDT充值功能完整测试")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("测试开始时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 70))

	testResults := make(map[string]bool)

	// 1. 用户认证流程测试
	fmt.Println("\n[测试1] 用户认证流程测试")
	fmt.Println(strings.Repeat("-", 50))
	if err := testUserAuth(); err != nil {
		fmt.Printf("❌ 用户认证测试失败: %v\n", err)
		testResults["用户认证"] = false
	} else {
		fmt.Println("✅ 用户认证测试通过")
		testResults["用户认证"] = true
	}

	// 2. 管理员登录（用于后续审核测试）
	fmt.Println("\n[测试2] 管理员认证测试")
	fmt.Println(strings.Repeat("-", 50))
	if err := testAdminAuth(); err != nil {
		fmt.Printf("⚠️  管理员认证失败（将跳过管理员相关测试）: %v\n", err)
		testResults["管理员认证"] = false
	} else {
		fmt.Println("✅ 管理员认证测试通过")
		testResults["管理员认证"] = true
	}

	// 3. 管理员创建充值地址（准备测试数据）
	fmt.Println("\n[测试3] 管理员创建充值地址")
	fmt.Println(strings.Repeat("-", 50))
	if adminToken != "" {
		if err := testAdminCreateAddress(); err != nil {
			fmt.Printf("⚠️  创建充值地址失败: %v\n", err)
			testResults["创建充值地址"] = false
		} else {
			fmt.Println("✅ 创建充值地址测试通过")
			testResults["创建充值地址"] = true
		}
	} else {
		fmt.Println("⏭️  跳过（管理员未登录）")
	}

	// 4. 充值地址获取测试
	fmt.Println("\n[测试4] 充值地址获取测试")
	fmt.Println(strings.Repeat("-", 50))
	if err := testGetDepositAddresses(); err != nil {
		fmt.Printf("❌ 获取充值地址失败: %v\n", err)
		testResults["获取充值地址"] = false
	} else {
		fmt.Println("✅ 获取充值地址测试通过")
		testResults["获取充值地址"] = true
	}

	// 5. 充值订单创建测试
	fmt.Println("\n[测试5] 充值订单创建测试")
	fmt.Println(strings.Repeat("-", 50))
	if err := testCreateDepositOrder(); err != nil {
		fmt.Printf("❌ 创建充值订单失败: %v\n", err)
		testResults["创建充值订单"] = false
	} else {
		fmt.Println("✅ 创建充值订单测试通过")
		testResults["创建充值订单"] = true
	}

	// 6. 金额验证测试（最小10 USDT）
	fmt.Println("\n[测试6] 金额验证测试（最小10 USDT）")
	fmt.Println(strings.Repeat("-", 50))
	if err := testMinAmountValidation(); err != nil {
		fmt.Printf("❌ 金额验证测试失败: %v\n", err)
		testResults["金额验证"] = false
	} else {
		fmt.Println("✅ 金额验证测试通过")
		testResults["金额验证"] = true
	}

	// 7. 转账截图上传测试
	fmt.Println("\n[测试7] 转账截图上传测试")
	fmt.Println(strings.Repeat("-", 50))
	if testOrderID > 0 {
		if err := testUploadScreenshot(); err != nil {
			fmt.Printf("❌ 截图上传测试失败: %v\n", err)
			testResults["截图上传"] = false
		} else {
			fmt.Println("✅ 截图上传测试通过")
			testResults["截图上传"] = true
		}
	} else {
		fmt.Println("⏭️  跳过（无有效订单）")
	}

	// 8. 充值记录查询测试
	fmt.Println("\n[测试8] 充值记录查询测试")
	fmt.Println(strings.Repeat("-", 50))
	if err := testGetDepositOrders(); err != nil {
		fmt.Printf("❌ 充值记录查询失败: %v\n", err)
		testResults["充值记录查询"] = false
	} else {
		fmt.Println("✅ 充值记录查询测试通过")
		testResults["充值记录查询"] = true
	}

	// 9. 获取当前余额（用于后续对比）
	fmt.Println("\n[测试9] 获取当前用户余额")
	fmt.Println(strings.Repeat("-", 50))
	if balance, err := getUserBalance(); err != nil {
		fmt.Printf("⚠️  获取余额失败: %v\n", err)
	} else {
		initialBalance = balance
		fmt.Printf("✅ 当前USDT余额: %.8f\n", initialBalance)
	}

	// 10. 管理员审核通过测试
	fmt.Println("\n[测试10] 管理员审核通过测试")
	fmt.Println(strings.Repeat("-", 50))
	if adminToken != "" && testOrderID > 0 {
		if err := testAdminApprove(); err != nil {
			fmt.Printf("❌ 管理员审核失败: %v\n", err)
			testResults["管理员审核通过"] = false
		} else {
			fmt.Println("✅ 管理员审核通过测试完成")
			testResults["管理员审核通过"] = true
		}
	} else {
		fmt.Println("⏭️  跳过（管理员未登录或无订单）")
	}

	// 11. 余额更新验证
	fmt.Println("\n[测试11] 余额更新验证")
	fmt.Println(strings.Repeat("-", 50))
	if adminToken != "" && testResults["管理员审核通过"] {
		if err := testBalanceUpdate(); err != nil {
			fmt.Printf("❌ 余额更新验证失败: %v\n", err)
			testResults["余额更新验证"] = false
		} else {
			fmt.Println("✅ 余额更新验证通过")
			testResults["余额更新验证"] = true
		}
	} else {
		fmt.Println("⏭️  跳过（审核未完成）")
	}

	// 12. 创建新订单用于取消测试
	fmt.Println("\n[测试12] 订单取消测试")
	fmt.Println(strings.Repeat("-", 50))
	if err := testCancelOrder(); err != nil {
		fmt.Printf("❌ 订单取消测试失败: %v\n", err)
		testResults["订单取消"] = false
	} else {
		fmt.Println("✅ 订单取消测试通过")
		testResults["订单取消"] = true
	}

	// 13. 管理员拒绝订单测试
	fmt.Println("\n[测试13] 管理员拒绝订单测试")
	fmt.Println(strings.Repeat("-", 50))
	if adminToken != "" {
		if err := testAdminReject(); err != nil {
			fmt.Printf("❌ 管理员拒绝测试失败: %v\n", err)
			testResults["管理员拒绝"] = false
		} else {
			fmt.Println("✅ 管理员拒绝测试通过")
			testResults["管理员拒绝"] = true
		}
	} else {
		fmt.Println("⏭️  跳过（管理员未登录）")
	}

	// 清理测试数据
	fmt.Println("\n[清理] 删除测试充值地址")
	fmt.Println(strings.Repeat("-", 50))
	if adminToken != "" && testAddressID > 0 {
		if err := testAdminDeleteAddress(); err != nil {
			fmt.Printf("⚠️  清理失败: %v\n", err)
		} else {
			fmt.Println("✅ 测试数据清理完成")
		}
	}

	// 输出测试报告
	printTestReport(testResults)
}

// ============================================================
// 测试函数实现
// ============================================================

// testUserAuth 用户认证测试
func testUserAuth() error {
	// 1. 注册
	fmt.Println("  [1.1] 注册测试账号...")
	err := doRegister(testUsername, testPassword)
	if err != nil {
		fmt.Printf("  ⚠️  注册失败（可能已存在）: %v\n", err)
	} else {
		fmt.Println("  ✓ 注册成功")
	}

	// 2. 登录
	fmt.Println("  [1.2] 登录获取token...")
	token, err := doLogin(testUsername, testPassword)
	if err != nil {
		return fmt.Errorf("登录失败: %v", err)
	}
	userToken = token
	fmt.Printf("  ✓ 登录成功，token: %s...\n", token[:minInt(20, len(token))])

	return nil
}

// testAdminAuth 管理员认证测试
func testAdminAuth() error {
	fmt.Println("  [2.1] 管理员登录...")
	token, err := doAdminLogin(adminUsername, adminPassword)
	if err != nil {
		return fmt.Errorf("管理员登录失败: %v", err)
	}
	adminToken = token
	fmt.Printf("  ✓ 管理员登录成功，token: %s...\n", token[:minInt(20, len(token))])
	return nil
}

// testAdminCreateAddress 管理员创建充值地址
func testAdminCreateAddress() error {
	fmt.Println("  [3.1] 创建TRC20充值地址...")

	reqData := map[string]interface{}{
		"network": "TRC20",
		"address": "TTestAddress123456789012345678901234",
		"qr_code": "/uploads/qrcode/trc20_test.png",
		"status":  1,
		"sort":    100,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", baseURL+"/api/admin/deposit/address/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		// 如果已存在，尝试获取现有地址
		fmt.Printf("  ⚠️  创建失败: %s，尝试获取现有地址\n", result.Message)
		return getExistingAddress()
	}

	// 解析返回的地址ID
	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(float64); ok {
			testAddressID = uint(id)
			fmt.Printf("  ✓ 创建成功，地址ID: %d\n", testAddressID)
		}
	}

	return nil
}

// getExistingAddress 获取已存在的充值地址
func getExistingAddress() error {
	req, _ := http.NewRequest("GET", baseURL+"/api/admin/deposit/address/list", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type == "success" {
		if dataList, ok := result.Data.([]interface{}); ok && len(dataList) > 0 {
			if addr, ok := dataList[0].(map[string]interface{}); ok {
				testAddressID = uint(addr["id"].(float64))
				fmt.Printf("  ✓ 获取现有地址ID: %d\n", testAddressID)
			}
		}
	}
	return nil
}

// testGetDepositAddresses 获取充值地址测试
func testGetDepositAddresses() error {
	fmt.Println("  [4.1] 获取充值地址列表...")

	req, _ := http.NewRequest("GET", baseURL+"/api/deposit/addresses", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  响应: %s\n", string(body))

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		return fmt.Errorf("获取充值地址失败: %s", result.Message)
	}

	// 解析地址列表
	if dataList, ok := result.Data.([]interface{}); ok {
		fmt.Printf("  ✓ 获取到 %d 个充值地址\n", len(dataList))
		for i, item := range dataList {
			if addr, ok := item.(map[string]interface{}); ok {
				fmt.Printf("    [%d] 网络: %s, 地址: %s\n", i+1, addr["network"], addr["address"])
			}
		}
	}

	return nil
}

// testCreateDepositOrder 创建充值订单测试
func testCreateDepositOrder() error {
	fmt.Println("  [5.1] 创建充值订单（100 USDT, TRC20）...")

	reqData := map[string]interface{}{
		"network": "TRC20",
		"amount":  100.0,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", baseURL+"/api/deposit/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  响应: %s\n", string(body))

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		return fmt.Errorf("创建订单失败: %s", result.Message)
	}

	// 解析订单信息
	if orderData, ok := result.Data.(map[string]interface{}); ok {
		testOrderID = uint(orderData["id"].(float64))
		testOrderNo = orderData["order_no"].(string)
		fmt.Printf("  ✓ 订单创建成功\n")
		fmt.Printf("    订单ID: %d\n", testOrderID)
		fmt.Printf("    订单号: %s\n", testOrderNo)
		fmt.Printf("    金额: %.2f USDT\n", orderData["amount"].(float64))
		fmt.Printf("    网络: %s\n", orderData["network"].(string))
		fmt.Printf("    充值地址: %s\n", orderData["address"].(string))
	}

	return nil
}

// testMinAmountValidation 最小金额验证测试
func testMinAmountValidation() error {
	fmt.Println("  [6.1] 尝试创建小于10 USDT的订单...")

	reqData := map[string]interface{}{
		"network": "TRC20",
		"amount":  5.0, // 小于最小金额10 USDT
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", baseURL+"/api/deposit/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type == "success" {
		return fmt.Errorf("应该拒绝小于10 USDT的订单，但却成功了")
	}

	fmt.Printf("  ✓ 正确拒绝了小于最小金额的订单: %s\n", result.Message)
	return nil
}

// testUploadScreenshot 上传截图测试
func testUploadScreenshot() error {
	fmt.Println("  [7.1] 创建测试截图文件...")

	// 创建临时测试图片
	tmpDir := os.TempDir()
	testImagePath := filepath.Join(tmpDir, "test_screenshot.png")

	// 创建一个简单的PNG文件（最小有效PNG）
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG签名
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	if err := os.WriteFile(testImagePath, pngData, 0644); err != nil {
		return fmt.Errorf("创建测试图片失败: %v", err)
	}
	defer os.Remove(testImagePath)

	fmt.Println("  [7.2] 上传转账截图...")

	// 创建multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加order_id字段
	writer.WriteField("order_id", fmt.Sprintf("%d", testOrderID))

	// 添加文件
	part, err := writer.CreateFormFile("file", "screenshot.png")
	if err != nil {
		return err
	}

	fileData, _ := os.ReadFile(testImagePath)
	part.Write(fileData)
	writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/deposit/upload", &buf)
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  响应: %s\n", string(body))

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		return fmt.Errorf("上传截图失败: %s", result.Message)
	}

	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		fmt.Printf("  ✓ 截图上传成功，路径: %s\n", dataMap["screenshot"])
	}

	return nil
}

// testGetDepositOrders 获取充值记录测试
func testGetDepositOrders() error {
	fmt.Println("  [8.1] 获取全部充值记录...")

	req, _ := http.NewRequest("GET", baseURL+"/api/deposit/orders?status=-1&page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  响应: %s\n", string(body))

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		return fmt.Errorf("获取充值记录失败: %s", result.Message)
	}

	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		total := int64(dataMap["total"].(float64))
		fmt.Printf("  ✓ 获取成功，总记录数: %d\n", total)

		if list, ok := dataMap["list"].([]interface{}); ok {
			for i, item := range list {
				if order, ok := item.(map[string]interface{}); ok {
					statusText := getStatusText(int8(order["status"].(float64)))
					fmt.Printf("    [%d] 订单号: %s, 金额: %.2f, 状态: %s\n",
						i+1, order["order_no"], order["amount"], statusText)
				}
			}
		}
	}

	// 测试状态筛选
	fmt.Println("  [8.2] 筛选待审核订单...")
	req2, _ := http.NewRequest("GET", baseURL+"/api/deposit/orders?status=0", nil)
	req2.Header.Set("Authorization", "Bearer "+userToken)
	resp2, _ := http.DefaultClient.Do(req2)
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	var result2 Response
	json.Unmarshal(body2, &result2)
	if result2.Type == "success" {
		if dataMap, ok := result2.Data.(map[string]interface{}); ok {
			total := int64(dataMap["total"].(float64))
			fmt.Printf("  ✓ 待审核订单数: %d\n", total)
		}
	}

	return nil
}

// getUserBalance 获取用户余额
func getUserBalance() (float64, error) {
	req, _ := http.NewRequest("GET", baseURL+"/api/wallet/asset-overview", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if result.Type != "success" {
		return 0, fmt.Errorf("获取余额失败")
	}

	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		return dataMap["total_balance"].(float64), nil
	}

	return 0, nil
}

// testAdminApprove 管理员审核通过测试
func testAdminApprove() error {
	fmt.Println("  [10.1] 管理员审核通过订单...")

	reqData := map[string]interface{}{
		"order_id": testOrderID,
		"remark":   "测试审核通过",
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", baseURL+"/api/admin/deposit/approve", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  响应: %s\n", string(body))

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		return fmt.Errorf("审核通过失败: %s", result.Message)
	}

	fmt.Printf("  ✓ 审核通过成功\n")
	return nil
}

// testBalanceUpdate 余额更新验证
func testBalanceUpdate() error {
	fmt.Println("  [11.1] 验证余额是否增加...")

	// 等待数据库更新
	time.Sleep(500 * time.Millisecond)

	newBalance, err := getUserBalance()
	if err != nil {
		return err
	}

	expectedBalance := initialBalance + 100.0 // 充值了100 USDT
	fmt.Printf("  初始余额: %.8f\n", initialBalance)
	fmt.Printf("  当前余额: %.8f\n", newBalance)
	fmt.Printf("  预期余额: %.8f\n", expectedBalance)

	// 允许0.01的误差
	if newBalance < expectedBalance-0.01 || newBalance > expectedBalance+0.01 {
		return fmt.Errorf("余额更新不正确，预期 %.2f，实际 %.2f", expectedBalance, newBalance)
	}

	fmt.Printf("  ✓ 余额正确增加 100 USDT\n")
	return nil
}

// testCancelOrder 订单取消测试
func testCancelOrder() error {
	// 先创建一个新订单
	fmt.Println("  [12.1] 创建新订单用于取消测试...")

	reqData := map[string]interface{}{
		"network": "TRC20",
		"amount":  50.0,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", baseURL+"/api/deposit/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type != "success" {
		return fmt.Errorf("创建订单失败: %s", result.Message)
	}

	var cancelOrderID uint
	if orderData, ok := result.Data.(map[string]interface{}); ok {
		cancelOrderID = uint(orderData["id"].(float64))
		fmt.Printf("  ✓ 订单创建成功，ID: %d\n", cancelOrderID)
	}

	// 取消订单
	fmt.Println("  [12.2] 取消订单...")
	cancelData := map[string]interface{}{
		"order_id": cancelOrderID,
	}
	cancelJSON, _ := json.Marshal(cancelData)
	cancelReq, _ := http.NewRequest("POST", baseURL+"/api/deposit/cancel", bytes.NewBuffer(cancelJSON))
	cancelReq.Header.Set("Authorization", "Bearer "+userToken)
	cancelReq.Header.Set("Content-Type", "application/json")

	cancelResp, err := http.DefaultClient.Do(cancelReq)
	if err != nil {
		return err
	}
	defer cancelResp.Body.Close()

	cancelBody, _ := io.ReadAll(cancelResp.Body)
	fmt.Printf("  响应: %s\n", string(cancelBody))

	var cancelResult Response
	json.Unmarshal(cancelBody, &cancelResult)

	if cancelResult.Type != "success" {
		return fmt.Errorf("取消订单失败: %s", cancelResult.Message)
	}

	fmt.Println("  ✓ 订单取消成功")

	// 验证：尝试再次取消应该失败
	fmt.Println("  [12.3] 验证重复取消被拒绝...")
	cancelReq2, _ := http.NewRequest("POST", baseURL+"/api/deposit/cancel", bytes.NewBuffer(cancelJSON))
	cancelReq2.Header.Set("Authorization", "Bearer "+userToken)
	cancelReq2.Header.Set("Content-Type", "application/json")
	cancelResp2, _ := http.DefaultClient.Do(cancelReq2)
	defer cancelResp2.Body.Close()
	cancelBody2, _ := io.ReadAll(cancelResp2.Body)
	var cancelResult2 Response
	json.Unmarshal(cancelBody2, &cancelResult2)

	if cancelResult2.Type == "success" {
		return fmt.Errorf("已取消的订单不应该能再次取消")
	}
	fmt.Printf("  ✓ 正确拒绝了重复取消: %s\n", cancelResult2.Message)

	return nil
}

// testAdminReject 管理员拒绝订单测试
func testAdminReject() error {
	// 先创建一个新订单
	fmt.Println("  [13.1] 创建新订单用于拒绝测试...")

	reqData := map[string]interface{}{
		"network": "TRC20",
		"amount":  30.0,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", baseURL+"/api/deposit/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type != "success" {
		return fmt.Errorf("创建订单失败: %s", result.Message)
	}

	var rejectOrderID uint
	if orderData, ok := result.Data.(map[string]interface{}); ok {
		rejectOrderID = uint(orderData["id"].(float64))
		fmt.Printf("  ✓ 订单创建成功，ID: %d\n", rejectOrderID)
	}

	// 管理员拒绝
	fmt.Println("  [13.2] 管理员拒绝订单...")
	rejectData := map[string]interface{}{
		"order_id": rejectOrderID,
		"remark":   "测试拒绝原因：截图不清晰",
	}
	rejectJSON, _ := json.Marshal(rejectData)
	rejectReq, _ := http.NewRequest("POST", baseURL+"/api/admin/deposit/reject", bytes.NewBuffer(rejectJSON))
	rejectReq.Header.Set("Authorization", "Bearer "+adminToken)
	rejectReq.Header.Set("Content-Type", "application/json")

	rejectResp, err := http.DefaultClient.Do(rejectReq)
	if err != nil {
		return err
	}
	defer rejectResp.Body.Close()

	rejectBody, _ := io.ReadAll(rejectResp.Body)
	fmt.Printf("  响应: %s\n", string(rejectBody))

	var rejectResult Response
	json.Unmarshal(rejectBody, &rejectResult)

	if rejectResult.Type != "success" {
		return fmt.Errorf("拒绝订单失败: %s", rejectResult.Message)
	}

	fmt.Println("  ✓ 订单拒绝成功")
	return nil
}

// testAdminDeleteAddress 删除测试充值地址
func testAdminDeleteAddress() error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/admin/deposit/address/%d", baseURL, testAddressID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type != "success" {
		return fmt.Errorf("删除地址失败: %s", result.Message)
	}

	return nil
}

// ============================================================
// 辅助函数
// ============================================================

func doRegister(username, password string) error {
	registerURL := baseURL + "/api/user/register"
	registerData := map[string]interface{}{
		"type":           "email",
		"user_string":    username,
		"password":       password,
		"re_password":    password,
		"code":           "000000",
		"extension_code": "",
		"country_code":   0,
	}

	jsonData, _ := json.Marshal(registerData)
	resp, err := http.Post(registerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type != "success" {
		return fmt.Errorf("%s", result.Message)
	}
	return nil
}

func doLogin(username, password string) (string, error) {
	loginURL := baseURL + "/api/user/login"
	loginData := map[string]interface{}{
		"user_string":  username,
		"password":     password,
		"type":         1,
		"area_code_id": 0,
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type != "success" {
		return "", fmt.Errorf("%s", result.Message)
	}

	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("无法解析登录响应")
	}

	token, ok := dataMap["token"].(string)
	if !ok {
		return "", fmt.Errorf("无法获取token")
	}

	return token, nil
}

func doAdminLogin(username, password string) (string, error) {
	loginURL := baseURL + "/api/admin/login"
	loginData := map[string]interface{}{
		"username": username,
		"password": password,
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Type != "success" {
		return "", fmt.Errorf("%s", result.Message)
	}

	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("无法解析登录响应")
	}

	token, ok := dataMap["token"].(string)
	if !ok {
		return "", fmt.Errorf("无法获取token")
	}

	return token, nil
}

func getStatusText(status int8) string {
	switch status {
	case 0:
		return "待审核"
	case 1:
		return "已通过"
	case 2:
		return "已拒绝"
	case 3:
		return "已取消"
	case 4:
		return "已过期"
	default:
		return "未知"
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// printTestReport 输出测试报告
func printTestReport(results map[string]bool) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("                         测试报告")
	fmt.Println(strings.Repeat("=", 70))

	passed := 0
	failed := 0
	skipped := 0

	testOrder := []string{
		"用户认证", "管理员认证", "创建充值地址", "获取充值地址",
		"创建充值订单", "金额验证", "截图上传", "充值记录查询",
		"管理员审核通过", "余额更新验证", "订单取消", "管理员拒绝",
	}

	fmt.Println("\n测试项目                                    结果")
	fmt.Println(strings.Repeat("-", 50))

	for _, name := range testOrder {
		result, exists := results[name]
		status := "⏭️  跳过"
		if exists {
			if result {
				status = "✅ 通过"
				passed++
			} else {
				status = "❌ 失败"
				failed++
			}
		} else {
			skipped++
		}
		fmt.Printf("%-40s %s\n", name, status)
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("\n总计: %d 项测试\n", passed+failed+skipped)
	fmt.Printf("  ✅ 通过: %d\n", passed)
	fmt.Printf("  ❌ 失败: %d\n", failed)
	fmt.Printf("  ⏭️  跳过: %d\n", skipped)

	fmt.Println(strings.Repeat("=", 70))
	if failed == 0 {
		fmt.Println("🎉 所有测试通过！USDT充值功能正常工作")
	} else {
		fmt.Printf("⚠️  有 %d 项测试失败，请检查相关功能\n", failed)
	}
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("测试完成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
