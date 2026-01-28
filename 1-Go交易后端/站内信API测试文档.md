# 站内信系统API测试文档

## 环境准备

### 1. 启动Go后端服务
```bash
cd "d:\工程\交易所源码\OKCoinsgp交易所源码带教程\混合架构版本\1-Go交易后端"
go run cmd/api/main.go
```

服务将在 `http://localhost:8080` 启动

### 2. 获取测试Token

#### 用户登录获取Token
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
```

#### 管理员登录获取Token
```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

将返回的 `token` 字段保存，用于后续请求。

---

## 一、前台用户API测试

### 1.1 获取用户消息列表

**请求**：
```bash
curl -X GET "http://localhost:8080/api/messages/list?page=1&page_size=10&type=0&is_read=-1" \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

**参数说明**：
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认10）
- `type`: 消息类型（0=全部，1=系统消息，2=交易通知，3=活动通知，4=个人消息）
- `is_read`: 已读状态（-1=全部，0=未读，1=已读）

**预期响应**：
```json
{
  "code": 200,
  "msg": "获取成功",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 1,
        "message_id": 1,
        "is_read": 0,
        "is_deleted": 0,
        "read_time": 0,
        "create_time": 1737763200,
        "message": {
          "id": 1,
          "title": "欢迎注册",
          "content": "<p>欢迎您注册本平台！</p>",
          "type": 1,
          "priority": 0
        }
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 10
  }
}
```

### 1.2 获取消息详情

**请求**：
```bash
curl -X GET "http://localhost:8080/api/messages/detail?id=1" \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

**说明**：获取详情时会自动标记消息为已读

### 1.3 获取未读消息数量

**请求**：
```bash
curl -X GET "http://localhost:8080/api/messages/unread-count" \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "msg": "获取成功",
  "data": {
    "unread_count": 3
  }
}
```

### 1.4 标记消息为已读

**请求**：
```bash
curl -X POST http://localhost:8080/api/messages/read \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

### 1.5 标记消息为未读

**请求**：
```bash
curl -X POST http://localhost:8080/api/messages/unread \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

### 1.6 批量标记已读

**请求**：
```bash
curl -X POST http://localhost:8080/api/messages/batch-read \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": [1, 2, 3]}'
```

### 1.7 全部标记已读

**请求**：
```bash
curl -X POST http://localhost:8080/api/messages/read-all \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

### 1.8 删除消息（软删除）

**请求**：
```bash
curl -X POST http://localhost:8080/api/messages/delete \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

### 1.9 批量删除消息

**请求**：
```bash
curl -X POST http://localhost:8080/api/messages/batch-delete \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": [1, 2, 3]}'
```

### 1.10 获取用户消息统计

**请求**：
```bash
curl -X GET http://localhost:8080/api/messages/statistics \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "msg": "获取成功",
  "data": {
    "total_count": 10,
    "unread_count": 3,
    "read_count": 7,
    "system_count": 5,
    "trade_count": 3,
    "activity_count": 2
  }
}
```

---

## 二、后台管理API测试

### 2.1 发送消息（指定用户）

**请求**：
```bash
curl -X POST http://localhost:8080/api/admin/messages/send \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "系统升级通知",
    "content": "<p>系统将于今晚22:00进行升级维护。</p>",
    "type": 1,
    "priority": 1,
    "receiver_ids": [1, 2, 3],
    "is_batch": 0
  }'
```

**参数说明**：
- `title`: 消息标题（必填）
- `content`: 消息内容HTML格式（必填）
- `type`: 消息类型（1=系统消息，2=交易通知，3=活动通知，4=个人消息）
- `priority`: 优先级（0=普通，1=重要，2=紧急）
- `receiver_ids`: 接收者用户ID数组（必填）
- `is_batch`: 是否群发（0=否，1=是）
- `attachment`: 附件链接（可选）

### 2.2 全员发送消息

**请求**：
```bash
curl -X POST http://localhost:8080/api/admin/messages/send-all \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "新功能上线通知",
    "content": "<p>平台新增站内信功能！</p>",
    "type": 1,
    "priority": 0
  }'
```

### 2.3 按用户等级发送

**请求**：
```bash
curl -X POST http://localhost:8080/api/admin/messages/send-by-level \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "VIP专享活动",
    "content": "<p>尊敬的VIP用户，您有新的专享活动。</p>",
    "type": 3,
    "priority": 1,
    "level_id": 2
  }'
```

### 2.4 获取消息列表

**请求**：
```bash
curl -X POST http://localhost:8080/api/admin/messages/list \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "page": 1,
    "page_size": 10,
    "type": 0,
    "status": 0,
    "priority": -1,
    "keyword": ""
  }'
```

**参数说明**：
- `type`: 消息类型（0=全部）
- `status`: 消息状态（0=全部，1=已发送，2=撤回，3=删除）
- `priority`: 优先级（-1=全部，0=普通，1=重要，2=紧急）
- `keyword`: 搜索关键词（标题或内容）

### 2.5 获取消息详情

**请求**：
```bash
curl -X GET http://localhost:8080/api/admin/messages/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 2.6 更新消息

**请求**：
```bash
curl -X PUT http://localhost:8080/api/admin/messages/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "系统升级通知（更新）",
    "content": "<p>升级时间调整为23:00。</p>",
    "priority": 2
  }'
```

### 2.7 撤回消息

**请求**：
```bash
curl -X POST http://localhost:8080/api/admin/messages/withdraw \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

**说明**：撤回后，用户将看不到该消息

### 2.8 删除消息（硬删除）

**请求**：
```bash
curl -X DELETE http://localhost:8080/api/admin/messages/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

**警告**：硬删除会同时删除消息主表和所有用户关联记录，请谨慎操作！

### 2.9 获取接收者列表

**请求**：
```bash
curl -X GET "http://localhost:8080/api/admin/messages/1/receivers?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "msg": "获取成功",
  "data": {
    "list": [
      {
        "user_id": 1,
        "username": "user001",
        "email": "user001@example.com",
        "is_read": 1,
        "read_time": 1737763800,
        "create_time": 1737763200
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

### 2.10 获取单条消息统计

**请求**：
```bash
curl -X GET http://localhost:8080/api/admin/messages/1/statistics \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "msg": "获取成功",
  "data": {
    "total_receivers": 10,
    "read_count": 7,
    "unread_count": 3,
    "read_rate": 70.0
  }
}
```

### 2.11 获取全局统计

**请求**：
```bash
curl -X GET http://localhost:8080/api/admin/messages/statistics \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

---

## 三、消息模板管理API测试

### 3.1 获取模板列表

**请求**：
```bash
curl -X GET "http://localhost:8080/api/admin/message-templates/list?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 3.2 创建消息模板

**请求**：
```bash
curl -X POST http://localhost:8080/api/admin/message-templates/create \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "custom_template",
    "title": "自定义模板",
    "content": "<p>尊敬的 {{username}}，您的账户余额为 {{balance}}。</p>",
    "type": 1,
    "description": "账户余额通知模板",
    "variables": "[\"username\", \"balance\"]",
    "status": 1
  }'
```

**参数说明**：
- `code`: 模板唯一代码（必填）
- `title`: 模板标题（必填）
- `content`: 模板内容，支持 `{{variable}}` 占位符（必填）
- `variables`: 变量列表JSON字符串（可选）
- `status`: 状态（0=禁用，1=启用）

### 3.3 更新消息模板

**请求**：
```bash
curl -X PUT http://localhost:8080/api/admin/message-templates/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "自定义模板（已更新）",
    "content": "<p>更新后的模板内容</p>",
    "status": 1
  }'
```

### 3.4 删除消息模板

**请求**：
```bash
curl -X DELETE http://localhost:8080/api/admin/message-templates/1 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

---

## 四、Python测试脚本使用

### 4.1 安装依赖
```bash
pip install requests
```

### 4.2 配置Token
编辑 `test_message_api.py`，设置：
```python
USER_TOKEN = "your_user_token_here"
ADMIN_TOKEN = "your_admin_token_here"
```

### 4.3 运行测试
```bash
cd "d:\工程\交易所源码\OKCoinsgp交易所源码带教程\混合架构版本\1-Go交易后端"
python test_message_api.py
```

---

## 五、Postman导入

### 5.1 创建Postman Collection

1. 打开Postman
2. 点击 `Import` -> `Raw text`
3. 复制以下JSON（简化示例）：

```json
{
  "info": {
    "name": "站内信系统API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "前台用户API",
      "item": [
        {
          "name": "获取消息列表",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{user_token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/api/messages/list?page=1&page_size=10",
              "host": ["{{base_url}}"],
              "path": ["api", "messages", "list"],
              "query": [
                {"key": "page", "value": "1"},
                {"key": "page_size", "value": "10"}
              ]
            }
          }
        }
      ]
    },
    {
      "name": "后台管理API",
      "item": [
        {
          "name": "发送消息",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{admin_token}}"
              },
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"title\": \"测试消息\",\n  \"content\": \"<p>测试内容</p>\",\n  \"type\": 1,\n  \"priority\": 0,\n  \"receiver_ids\": [1]\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/admin/messages/send",
              "host": ["{{base_url}}"],
              "path": ["api", "admin", "messages", "send"]
            }
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    },
    {
      "key": "user_token",
      "value": ""
    },
    {
      "key": "admin_token",
      "value": ""
    }
  ]
}
```

### 5.2 设置环境变量
在Postman中设置：
- `base_url`: http://localhost:8080
- `user_token`: 用户登录后获取的Token
- `admin_token`: 管理员登录后获取的Token

---

## 六、常见问题排查

### 6.1 401 Unauthorized
- 检查Token是否正确
- Token是否过期（默认24小时）
- 请求头格式：`Authorization: Bearer YOUR_TOKEN`

### 6.2 数据库连接失败
- 检查MySQL是否启动
- 检查config.yaml中的数据库配置
- 确认数据库名称为 `bibi2022`

### 6.3 外键约束错误
- 确保users表已存在
- 可临时移除外键约束进行测试

### 6.4 消息发送失败
- 检查receiver_ids是否为有效的用户ID
- 确认用户在users表中存在
- 查看后端日志获取详细错误信息

---

## 七、测试清单

### 前台用户功能
- [ ] 获取消息列表（全部/未读/已读）
- [ ] 获取消息详情（自动标记已读）
- [ ] 获取未读数量
- [ ] 标记单条已读/未读
- [ ] 批量标记已读
- [ ] 全部标记已读
- [ ] 删除单条消息
- [ ] 批量删除消息
- [ ] 获取统计信息

### 后台管理功能
- [ ] 发送消息（指定用户）
- [ ] 全员发送
- [ ] 按等级发送
- [ ] 获取消息列表（筛选/搜索/分页）
- [ ] 获取消息详情
- [ ] 更新消息
- [ ] 撤回消息
- [ ] 删除消息
- [ ] 查看接收者列表
- [ ] 查看单条消息统计
- [ ] 查看全局统计

### 模板管理功能
- [ ] 获取模板列表
- [ ] 创建模板
- [ ] 更新模板
- [ ] 删除模板
- [ ] 使用模板发送（如果实现）

---

**测试提示**：
1. 建议按顺序测试：先后台创建消息 -> 前台查看接收
2. 关注响应时间和性能
3. 测试边界情况（空列表、大数据量等）
4. 验证权限控制（用户不能访问他人消息）
5. 测试并发场景（批量发送时）
