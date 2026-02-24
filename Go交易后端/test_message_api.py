#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
站内信系统API测试脚本
测试前台用户API和后台管理API的完整功能
"""

import requests
import json
import time
from typing import Dict, Any, Optional

# ========== 配置 ==========
BASE_URL = "http://localhost:8080/api"
ADMIN_BASE_URL = "http://localhost:8080/api/admin"

# 测试用户Token（需要先登录获取，或使用已有token）
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

# ========== 后台管理API测试 ==========

def test_admin_create_message():
    """测试：创建消息并发送"""
    print_section("后台管理 - 创建消息")
    
    url = f"{ADMIN_BASE_URL}/messages/send"
    data = {
        "title": "系统升级通知",
        "content": "<p>尊敬的用户，系统将于今晚22:00-24:00进行升级维护，期间暂停服务。</p><p>给您带来不便敬请谅解！</p>",
        "type": 1,  # 系统消息
        "priority": 1,  # 重要
        "receiver_ids": [1, 2, 3],  # 测试用户ID
        "is_batch": 0
    }
    
    response = make_request("POST", url, ADMIN_TOKEN, data)
    result = print_result(response, "创建消息")
    
    if result and result.get("code") == 200:
        return result.get("data", {}).get("id")
    return None

def test_admin_send_to_all():
    """测试：全员发送消息"""
    print_section("后台管理 - 全员发送")
    
    url = f"{ADMIN_BASE_URL}/messages/send-all"
    data = {
        "title": "新功能上线通知",
        "content": "<p>平台新增站内信功能，您可以实时接收系统通知和交易提醒。</p>",
        "type": 1,
        "priority": 0
    }
    
    response = make_request("POST", url, ADMIN_TOKEN, data)
    print_result(response, "全员发送")

def test_admin_message_list():
    """测试：获取消息列表"""
    print_section("后台管理 - 消息列表")
    
    url = f"{ADMIN_BASE_URL}/messages/list"
    data = {
        "page": 1,
        "page_size": 10,
        "type": 0,  # 0=全部
        "status": 0  # 0=全部
    }
    
    response = make_request("POST", url, ADMIN_TOKEN, data)
    print_result(response, "消息列表")

def test_admin_message_detail(message_id: int):
    """测试：获取消息详情"""
    print_section("后台管理 - 消息详情")
    
    url = f"{ADMIN_BASE_URL}/messages/{message_id}"
    response = make_request("GET", url, ADMIN_TOKEN)
    print_result(response, f"消息详情 (ID: {message_id})")

def test_admin_message_statistics():
    """测试：获取消息统计"""
    print_section("后台管理 - 消息统计")
    
    url = f"{ADMIN_BASE_URL}/messages/statistics"
    response = make_request("GET", url, ADMIN_TOKEN)
    print_result(response, "全局统计")

def test_admin_receiver_list(message_id: int):
    """测试：获取接收者列表"""
    print_section("后台管理 - 接收者列表")
    
    url = f"{ADMIN_BASE_URL}/messages/{message_id}/receivers"
    params = {"page": 1, "page_size": 10}
    response = make_request("GET", url, ADMIN_TOKEN, params=params)
    print_result(response, f"接收者列表 (消息ID: {message_id})")

def test_admin_withdraw_message(message_id: int):
    """测试：撤回消息"""
    print_section("后台管理 - 撤回消息")
    
    url = f"{ADMIN_BASE_URL}/messages/withdraw"
    data = {"id": message_id}
    response = make_request("POST", url, ADMIN_TOKEN, data)
    print_result(response, f"撤回消息 (ID: {message_id})")

def test_admin_delete_message(message_id: int):
    """测试：删除消息"""
    print_section("后台管理 - 删除消息")
    
    url = f"{ADMIN_BASE_URL}/messages/{message_id}"
    response = make_request("DELETE", url, ADMIN_TOKEN)
    print_result(response, f"删除消息 (ID: {message_id})")

# ========== 消息模板管理测试 ==========

def test_admin_template_list():
    """测试：获取模板列表"""
    print_section("后台管理 - 模板列表")
    
    url = f"{ADMIN_BASE_URL}/message-templates/list"
    params = {"page": 1, "page_size": 20}
    response = make_request("GET", url, ADMIN_TOKEN, params=params)
    result = print_result(response, "模板列表")
    
    if result and result.get("code") == 200:
        templates = result.get("data", {}).get("list", [])
        if templates:
            return templates[0].get("id")
    return None

def test_admin_create_template():
    """测试：创建消息模板"""
    print_section("后台管理 - 创建模板")
    
    url = f"{ADMIN_BASE_URL}/message-templates/create"
    data = {
        "code": "test_template",
        "title": "测试模板",
        "content": "<p>尊敬的 {{username}}，这是一条测试消息。</p><p>金额：{{amount}}</p>",
        "type": 1,
        "description": "仅供测试使用",
        "variables": '["username", "amount"]',
        "status": 1
    }
    
    response = make_request("POST", url, ADMIN_TOKEN, data)
    result = print_result(response, "创建模板")
    
    if result and result.get("code") == 200:
        return result.get("data", {}).get("id")
    return None

def test_admin_update_template(template_id: int):
    """测试：更新消息模板"""
    print_section("后台管理 - 更新模板")
    
    url = f"{ADMIN_BASE_URL}/message-templates/{template_id}"
    data = {
        "title": "测试模板（已更新）",
        "content": "<p>这是更新后的模板内容</p>",
        "status": 1
    }
    
    response = make_request("PUT", url, ADMIN_TOKEN, data)
    print_result(response, f"更新模板 (ID: {template_id})")

def test_admin_delete_template(template_id: int):
    """测试：删除消息模板"""
    print_section("后台管理 - 删除模板")
    
    url = f"{ADMIN_BASE_URL}/message-templates/{template_id}"
    response = make_request("DELETE", url, ADMIN_TOKEN)
    print_result(response, f"删除模板 (ID: {template_id})")

# ========== 前台用户API测试 ==========

def test_user_message_list():
    """测试：获取用户消息列表"""
    print_section("前台用户 - 消息列表")
    
    url = f"{BASE_URL}/messages/list"
    params = {
        "page": 1,
        "page_size": 10,
        "type": 0,  # 0=全部
        "is_read": -1  # -1=全部, 0=未读, 1=已读
    }
    
    response = make_request("GET", url, USER_TOKEN, params=params)
    result = print_result(response, "用户消息列表")
    
    if result and result.get("code") == 200:
        messages = result.get("data", {}).get("list", [])
        if messages:
            return messages[0].get("message_id")
    return None

def test_user_message_detail(message_id: int):
    """测试：获取消息详情（自动标记已读）"""
    print_section("前台用户 - 消息详情")
    
    url = f"{BASE_URL}/messages/detail"
    params = {"id": message_id}
    response = make_request("GET", url, USER_TOKEN, params=params)
    print_result(response, f"消息详情 (ID: {message_id})")

def test_user_unread_count():
    """测试：获取未读消息数量"""
    print_section("前台用户 - 未读数量")
    
    url = f"{BASE_URL}/messages/unread-count"
    response = make_request("GET", url, USER_TOKEN)
    print_result(response, "未读消息数量")

def test_user_mark_as_read(message_id: int):
    """测试：标记消息为已读"""
    print_section("前台用户 - 标记已读")
    
    url = f"{BASE_URL}/messages/read"
    data = {"id": message_id}
    response = make_request("POST", url, USER_TOKEN, data)
    print_result(response, f"标记已读 (ID: {message_id})")

def test_user_mark_as_unread(message_id: int):
    """测试：标记消息为未读"""
    print_section("前台用户 - 标记未读")
    
    url = f"{BASE_URL}/messages/unread"
    data = {"id": message_id}
    response = make_request("POST", url, USER_TOKEN, data)
    print_result(response, f"标记未读 (ID: {message_id})")

def test_user_batch_mark_as_read():
    """测试：批量标记已读"""
    print_section("前台用户 - 批量标记已读")
    
    url = f"{BASE_URL}/messages/batch-read"
    data = {"ids": [1, 2, 3]}  # 替换为实际消息ID
    response = make_request("POST", url, USER_TOKEN, data)
    print_result(response, "批量标记已读")

def test_user_mark_all_as_read():
    """测试：全部标记已读"""
    print_section("前台用户 - 全部标记已读")
    
    url = f"{BASE_URL}/messages/read-all"
    response = make_request("POST", url, USER_TOKEN)
    print_result(response, "全部标记已读")

def test_user_delete_message(message_id: int):
    """测试：删除消息"""
    print_section("前台用户 - 删除消息")
    
    url = f"{BASE_URL}/messages/delete"
    data = {"id": message_id}
    response = make_request("POST", url, USER_TOKEN, data)
    print_result(response, f"删除消息 (ID: {message_id})")

def test_user_statistics():
    """测试：获取用户消息统计"""
    print_section("前台用户 - 消息统计")
    
    url = f"{BASE_URL}/messages/statistics"
    response = make_request("GET", url, USER_TOKEN)
    print_result(response, "用户消息统计")

# ========== 主测试流程 ==========

def run_admin_tests():
    """运行后台管理API测试"""
    print("\n" + "="*60)
    print("  开始测试后台管理API")
    print("="*60)
    
    if not ADMIN_TOKEN:
        print("\n⚠️  警告: 未设置ADMIN_TOKEN，跳过后台测试")
        return
    
    # 1. 创建消息
    message_id = test_admin_create_message()
    time.sleep(1)
    
    # 2. 获取消息列表
    test_admin_message_list()
    time.sleep(1)
    
    # 3. 获取消息详情
    if message_id:
        test_admin_message_detail(message_id)
        time.sleep(1)
    
    # 4. 获取接收者列表
    if message_id:
        test_admin_receiver_list(message_id)
        time.sleep(1)
    
    # 5. 获取统计信息
    test_admin_message_statistics()
    time.sleep(1)
    
    # 6. 全员发送
    test_admin_send_to_all()
    time.sleep(1)
    
    # 7. 模板管理
    template_id = test_admin_template_list()
    time.sleep(1)
    
    new_template_id = test_admin_create_template()
    time.sleep(1)
    
    if new_template_id:
        test_admin_update_template(new_template_id)
        time.sleep(1)
    
    # 8. 撤回消息（可选）
    # if message_id:
    #     test_admin_withdraw_message(message_id)
    #     time.sleep(1)
    
    # 9. 删除测试数据（可选）
    # if new_template_id:
    #     test_admin_delete_template(new_template_id)

def run_user_tests():
    """运行前台用户API测试"""
    print("\n" + "="*60)
    print("  开始测试前台用户API")
    print("="*60)
    
    if not USER_TOKEN:
        print("\n⚠️  警告: 未设置USER_TOKEN，跳过用户测试")
        return
    
    # 1. 获取消息列表
    message_id = test_user_message_list()
    time.sleep(1)
    
    # 2. 获取未读数量
    test_user_unread_count()
    time.sleep(1)
    
    # 3. 获取消息详情
    if message_id:
        test_user_message_detail(message_id)
        time.sleep(1)
    
    # 4. 标记已读/未读
    if message_id:
        test_user_mark_as_read(message_id)
        time.sleep(1)
        test_user_mark_as_unread(message_id)
        time.sleep(1)
    
    # 5. 批量操作
    # test_user_batch_mark_as_read()
    # time.sleep(1)
    
    # 6. 全部标记已读
    # test_user_mark_all_as_read()
    # time.sleep(1)
    
    # 7. 获取统计信息
    test_user_statistics()
    time.sleep(1)
    
    # 8. 删除消息（可选）
    # if message_id:
    #     test_user_delete_message(message_id)

def main():
    """主函数"""
    print("\n" + "="*60)
    print("  站内信系统API测试工具")
    print("="*60)
    print(f"\n基础URL: {BASE_URL}")
    print(f"管理后台URL: {ADMIN_BASE_URL}")
    print(f"用户Token: {'已设置' if USER_TOKEN else '未设置'}")
    print(f"管理员Token: {'已设置' if ADMIN_TOKEN else '未设置'}")
    
    if not USER_TOKEN and not ADMIN_TOKEN:
        print("\n❌ 错误: 请先设置USER_TOKEN或ADMIN_TOKEN")
        print("\n使用说明:")
        print("1. 先启动Go后端服务: cd 1-Go交易后端 && go run cmd/api/main.go")
        print("2. 通过登录接口获取Token")
        print("3. 在脚本顶部设置USER_TOKEN和ADMIN_TOKEN")
        print("4. 重新运行测试脚本")
        return
    
    try:
        # 运行测试
        run_admin_tests()
        run_user_tests()
        
        print("\n" + "="*60)
        print("  测试完成！")
        print("="*60)
        
    except Exception as e:
        print(f"\n❌ 测试出错: {str(e)}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()
