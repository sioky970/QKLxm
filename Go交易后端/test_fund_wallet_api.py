#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
资金钱包API测试脚本
测试资金钱包的完整功能：充值、转账、查询等
"""

import requests
import json
import time
from typing import Dict, Any, Optional

# ========== 配置 ==========
BASE_URL = "http://localhost:8080/api"
ADMIN_BASE_URL = "http://localhost:8080/api/admin"

# 测试用户Token（需要先登录获取）
USER_TOKEN = ""  # 填入真实的用户Token
ADMIN_TOKEN = ""  # 填入真实的管理员Token

# ========== 辅助函数 ==========

def print_section(title: str):
    """打印分隔线和标题"""
    print("\n" + "="*60)
    print(f"  {title}")
    print("="*60)

def print_result(response: requests.Response, title: str = ""):
    """格式化打印API响应"""
    if title:
        print(f"\n>>> {title}")
    print(f"状态码: {response.status_code}")
    try:
        data = response.json()
        print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")
        return data
    except:
        print(f"响应: {response.text}")
        return None

def make_request(method: str, url: str, token: Optional[str] = None, 
                 json_data: Optional[Dict] = None, params: Optional[Dict] = None) -> requests.Response:
    """统一的请求方法"""
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    
    if method.upper() == "GET":
        return requests.get(url, headers=headers, params=params)
    elif method.upper() == "POST":
        return requests.post(url, headers=headers, json=json_data)
    elif method.upper() == "PUT":
        return requests.put(url, headers=headers, json=json_data)
    elif method.upper() == "DELETE":
        return requests.delete(url, headers=headers, json=json_data)

# ========== 用户前台API测试 ==========

def test_get_balance():
    """测试：获取资金钱包余额"""
    print_section("获取资金钱包余额")
    
    url = f"{BASE_URL}/wallet/fund/balance"
    response = make_request("GET", url, USER_TOKEN)
    return print_result(response, "查询余额")

def test_get_overview():
    """测试：获取资金钱包总览"""
    print_section("获取资金钱包总览")
    
    url = f"{BASE_URL}/wallet/fund/overview"
    response = make_request("GET", url, USER_TOKEN)
    return print_result(response, "总览信息")

def test_get_transfer_list():
    """测试：获取划转记录列表"""
    print_section("获取划转记录列表")
    
    url = f"{BASE_URL}/wallet/fund/transfer/list"
    params = {"page": 1, "page_size": 10, "transfer_type": 0}
    response = make_request("GET", url, USER_TOKEN, params=params)
    return print_result(response, "划转记录")

def test_deposit_to_fund():
    """测试：从现货账户充值到资金钱包"""
    print_section("充值到资金钱包")
    
    url = f"{BASE_URL}/wallet/fund/deposit"
    data = {
        "amount": "100",
        "from_wallet": "spot"
    }
    response = make_request("POST", url, USER_TOKEN, data)
    return print_result(response, "充值请求")

def test_withdraw_from_fund():
    """测试：从资金钱包提现到现货账户"""
    print_section("从资金钱包提现")
    
    url = f"{BASE_URL}/wallet/fund/withdraw"
    data = {
        "amount": "50",
        "to_wallet": "spot"
    }
    response = make_request("POST", url, USER_TOKEN, data)
    return print_result(response, "提现请求")

def test_transfer_spot_to_contract():
    """测试：从现货转到合约"""
    print_section("现货转到合约")
    
    url = f"{BASE_URL}/wallet/fund/transfer-in"
    data = {
        "amount": "25",
        "from_wallet": "spot",
        "to_wallet": "contract"
    }
    response = make_request("POST", url, USER_TOKEN, data)
    return print_result(response, "转账请求")

def test_transfer_contract_to_fund():
    """测试：从合约转到资金钱包"""
    print_section("合约转到资金钱包")
    
    url = f"{BASE_URL}/wallet/fund/transfer-in"
    data = {
        "amount": "10",
        "from_wallet": "contract",
        "to_wallet": "fund"
    }
    response = make_request("POST", url, USER_TOKEN, data)
    return print_result(response, "转账请求")

def test_get_all_wallets():
    """测试：获取所有钱包余额"""
    print_section("获取所有钱包余额")
    
    url = f"{BASE_URL}/wallet/fund/all"
    response = make_request("GET", url, USER_TOKEN)
    return print_result(response, "所有钱包")

# ========== 后台管理API测试 ==========

def test_admin_get_wallet_list():
    """测试：后台获取钱包列表"""
    print_section("后台 - 获取钱包列表")
    
    url = f"{ADMIN_BASE_URL}/fund-wallets"
    params = {"page": 1, "page_size": 10}
    response = make_request("GET", url, ADMIN_TOKEN, params=params)
    return print_result(response, "钱包列表")

def test_admin_get_wallet_detail(wallet_id: int):
    """测试：后台获取钱包详情"""
    print_section(f"后台 - 获取钱包详情 (ID: {wallet_id})")
    
    url = f"{ADMIN_BASE_URL}/fund-wallets/{wallet_id}"
    response = make_request("GET", url, ADMIN_TOKEN)
    return print_result(response, "钱包详情")

def test_admin_get_wallet_statistics():
    """测试：后台获取钱包统计"""
    print_section("后台 - 获取钱包统计")
    
    url = f"{ADMIN_BASE_URL}/fund-wallets/statistics"
    response = make_request("GET", url, ADMIN_TOKEN)
    return print_result(response, "统计信息")

def test_admin_get_transfer_list():
    """测试：后台获取划转记录"""
    print_section("后台 - 获取划转记录列表")
    
    url = f"{ADMIN_BASE_URL}/fund-wallets/transfer/list"
    params = {"page": 1, "page_size": 10}
    response = make_request("GET", url, ADMIN_TOKEN, params=params)
    return print_result(response, "划转记录")

# ========== 测试流程 ==========

def test_user_apis():
    """测试用户前台API"""
    print_section("开始测试用户前台API")
    
    # 1. 获取余额
    test_get_balance()
    time.sleep(0.5)
    
    # 2. 获取总览
    test_get_overview()
    time.sleep(0.5)
    
    # 3. 获取所有钱包
    test_get_all_wallets()
    time.sleep(0.5)
    
    # 4. 获取划转记录
    test_get_transfer_list()
    time.sleep(0.5)
    
    # 5. 测试充值（如果余额充足）
    print_section("测试充值到资金钱包")
    test_deposit_to_fund()
    time.sleep(0.5)
    
    # 6. 测试提现
    print_section("测试从资金钱包提现")
    test_withdraw_from_fund()
    time.sleep(0.5)
    
    # 7. 测试转账
    print_section("测试现货转到合约")
    test_transfer_spot_to_contract()
    time.sleep(0.5)
    
    print_section("测试合约转到资金钱包")
    test_transfer_contract_to_fund()
    time.sleep(0.5)

def test_admin_apis():
    """测试后台管理API"""
    print_section("开始测试后台管理API")
    
    # 1. 获取统计
    test_admin_get_wallet_statistics()
    time.sleep(0.5)
    
    # 2. 获取列表
    test_admin_get_wallet_list()
    time.sleep(0.5)
    
    # 3. 获取划转记录
    test_admin_get_transfer_list()
    time.sleep(0.5)
    
    # 4. 获取详情
    print_section("尝试获取钱包详情")
    # 先获取列表中的第一个
    url = f"{ADMIN_BASE_URL}/fund-wallets"
    params = {"page": 1, "page_size": 1}
    response = requests.get(url, headers={"Authorization": f"Bearer {ADMIN_TOKEN}"}, params=params)
    if response.status_code == 200:
        data = response.json()
        if data.get("code") == 200 and data.get("data", {}).get("list"):
            wallet_id = data["data"]["list"][0]["id"]
            test_admin_get_wallet_detail(wallet_id)
        else:
            print("暂无钱包数据")
    else:
        print(f"获取列表失败: {response.status_code}")

def main():
    """主测试函数"""
    print_section("资金钱包API测试")
    print("注意：请先在脚本中配置 USER_TOKEN 和 ADMIN_TOKEN")
    print(f"API服务地址: {BASE_URL}")
    print(f"管理API地址: {ADMIN_BASE_URL}")
    
    if not USER_TOKEN:
        print("\n⚠️ 警告: 未配置 USER_TOKEN，将跳过需要认证的测试")
        print("请先获取token并配置到脚本中")
    else:
        test_user_apis()
    
    if not ADMIN_TOKEN:
        print("\n⚠️ 警告: 未配置 ADMIN_TOKEN，将跳过后台API测试")
        print("请先获取管理员token并配置到脚本中")
    else:
        test_admin_apis()
    
    print_section("测试完成")

if __name__ == "__main__":
    main()
