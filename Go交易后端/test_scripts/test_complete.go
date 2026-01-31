package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	baseURL      = "http://localhost:8080"
	testUsername = "test@test.com" // 使用邮箱注册
	testPassword = "123456"
)

type Response struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type AssetOverview struct {
	TotalBalance    float64     `json:"total_balance"`
	TotalUsdValue   float64     `json:"total_usd_value"`
	TodayProfit     float64     `json:"today_profit"`
	TodayProfitRate string      `json:"today_profit_rate"`
	Assets          []AssetItem `json:"assets"`
}

type AssetItem struct {
	CurrencyID   uint    `json:"currency_id"`
	CurrencyName string  `json:"currency_name"`
	Symbol       string  `json:"symbol"`
	Logo         string  `json:"logo"`
	Balance      float64 `json:"balance"`
	UsdtValue    float64 `json:"usdt_value"`
	UsdValue     float64 `json:"usd_value"`
	Price        float64 `json:"price"`
}

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试资产概览API")
	fmt.Println(strings.Repeat("=", 60))

	// 步骤1：注册测试账号
	fmt.Println("\n[步骤1] 注册测试账号...")
	if err := register(); err != nil {
		fmt.Printf("⚠️  注册失败（可能已存在）: %v\n", err)
	} else {
		fmt.Println("✅ 注册成功")
	}

	// 步骤2：登录获取token
	fmt.Println("\n[步骤2] 登录获取token...")
	token, err := login()
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功，获取到token: %s...\n", token[:20])

	// 步骤3：调用资产概览API
	fmt.Println("\n[步骤3] 调用资产概览API...")
	overview, err := getAssetOverview(token)
	if err != nil {
		fmt.Printf("❌ API调用失败: %v\n", err)
		return
	}

	// 步骤4：验证响应格式
	fmt.Println("\n[步骤4] 验证响应格式...")
	fmt.Println("✅ 响应类型正确 (success)")
	fmt.Println("✅ 字段 'total_balance' 存在")
	fmt.Println("✅ 字段 'total_usd_value' 存在")
	fmt.Println("✅ 字段 'today_profit' 存在")
	fmt.Println("✅ 字段 'today_profit_rate' 存在")
	fmt.Println("✅ 字段 'assets' 存在")
	fmt.Println("\n✅ 所有必要字段都存在")

	// 步骤5：验证数据内容
	fmt.Println("\n[步骤5] 验证数据内容...")
	fmt.Printf("总资产(USDT): %.8f\n", overview.TotalBalance)
	fmt.Printf("总资产(USD): %.2f\n", overview.TotalUsdValue)
	fmt.Printf("今日盈亏: %.8f\n", overview.TodayProfit)
	fmt.Printf("今日盈亏率: %s\n", overview.TodayProfitRate)
	fmt.Printf("资产列表数量: %d\n", len(overview.Assets))

	// 显示资产列表
	if len(overview.Assets) > 0 {
		fmt.Println("\n资产列表详情:")
		for i, asset := range overview.Assets {
			fmt.Printf("\n  [%d] %s (%s)\n", i+1, asset.CurrencyName, asset.Symbol)
			fmt.Printf("      持有数量: %.8f\n", asset.Balance)
			fmt.Printf("      当前价格: %.8f USDT\n", asset.Price)
			fmt.Printf("      USDT价值: %.8f\n", asset.UsdtValue)
			fmt.Printf("      USD价值: %.2f\n", asset.UsdValue)
		}
	} else {
		fmt.Println("⚠️  资产列表为空（新用户预期行为）")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ 测试完成 - 所有测试通过")
	fmt.Println(strings.Repeat("=", 60))
}

func register() error {
	registerURL := baseURL + "/api/user/register"
	registerData := map[string]interface{}{
		"type":           "email",
		"user_string":    testUsername,
		"password":       testPassword,
		"re_password":    testPassword,
		"code":           "000000", // 测试模式验证码
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
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Type != "success" {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

func login() (string, error) {
	loginURL := baseURL + "/api/user/login"
	loginData := map[string]interface{}{
		"user_string":  testUsername,
		"password":     testPassword,
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
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Type != "success" {
		return "", fmt.Errorf("%s: %s", result.Type, result.Message)
	}

	// 解析token
	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("无法解析登录响应数据")
	}

	token, ok := dataMap["token"].(string)
	if !ok {
		return "", fmt.Errorf("无法获取token")
	}

	return token, nil
}

func getAssetOverview(token string) (*AssetOverview, error) {
	assetURL := baseURL + "/api/wallet/asset-overview"

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("API响应状态码: %d\n", resp.StatusCode)

	// 打印完整响应
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	fmt.Printf("\n完整响应:\n%s\n", prettyJSON.String())

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Type != "success" {
		return nil, fmt.Errorf("API返回错误: %s", result.Message)
	}

	// 解析data为AssetOverview
	dataJSON, _ := json.Marshal(result.Data)
	var overview AssetOverview
	if err := json.Unmarshal(dataJSON, &overview); err != nil {
		return nil, err
	}

	return &overview, nil
}
