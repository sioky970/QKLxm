# 新闻管理 API 文档

## 概述

新闻管理模块提供了完整的新闻和新闻分类管理功能，包括前台展示API和后台管理API。

## 一、前台展示API（公开访问）

### 1.1 获取新闻列表

**接口地址**: `GET /api/news/list`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| category_id | uint | 否 | 新闻分类ID，不传或传0表示获取全部 |
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10，最大100 |

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "新闻标题",
        "content": "新闻内容",
        "category_id": 1,
        "author": "作者",
        "status": 1,
        "view_count": 100,
        "publish_time": 1640000000,
        "create_time": 1640000000,
        "update_time": 1640000000
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 1.2 获取新闻详情

**接口地址**: `GET /api/news/detail`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint | 是 | 新闻ID |

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": {
    "id": 1,
    "title": "新闻标题",
    "content": "新闻内容HTML",
    "category_id": 1,
    "author": "作者",
    "status": 1,
    "view_count": 101,
    "publish_time": 1640000000,
    "create_time": 1640000000,
    "update_time": 1640000000
  }
}
```

**说明**: 每次访问详情会自动增加浏览次数

### 1.3 获取新闻分类列表

**接口地址**: `GET /api/news/category`

**请求参数**: 无

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "name": "公告",
      "description": "平台公告",
      "status": 1,
      "create_time": 1640000000,
      "update_time": 1640000000
    }
  ]
}
```

### 1.4 获取推荐新闻

**接口地址**: `GET /api/news/recommended`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| limit | int | 否 | 返回数量，默认5，最大20 |

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "title": "最新新闻标题",
      "content": "内容摘要...",
      "publish_time": 1640000000
    }
  ]
}
```

**说明**: 按发布时间倒序返回最新的N条新闻

### 1.5 获取热门新闻

**接口地址**: `GET /api/news/hot`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| limit | int | 否 | 返回数量，默认5，最大20 |

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": [
    {
      "id": 2,
      "title": "热门新闻标题",
      "view_count": 1000,
      "publish_time": 1640000000
    }
  ]
}
```

**说明**: 按浏览量倒序返回最热门的N条新闻

---

## 二、后台管理API（需要管理员权限）

### 2.1 获取新闻列表（后台）

**接口地址**: `POST /api/admin/news/list`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "page": 1,
  "page_size": 10
}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "新闻标题",
        "status": 0,
        "view_count": 100,
        "create_time": 1640000000
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

**说明**: 包含所有状态的新闻（草稿、已发布等）

### 2.2 获取新闻详情（后台）

**接口地址**: `GET /api/admin/news/:id`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": {
    "id": 1,
    "title": "新闻标题",
    "content": "新闻完整内容",
    "category_id": 1,
    "author": "管理员",
    "status": 0,
    "view_count": 100,
    "publish_time": 0,
    "create_time": 1640000000,
    "update_time": 1640000000
  }
}
```

### 2.3 创建新闻

**接口地址**: `POST /api/admin/news/create`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "title": "新闻标题",
  "content": "新闻内容HTML",
  "category_id": 1,
  "author": "作者名称",
  "status": 0
}
```

**字段说明**:
- title: 新闻标题（必填）
- content: 新闻内容（必填，支持HTML）
- category_id: 分类ID
- author: 作者
- status: 状态（0=草稿，1=已发布）

**响应示例**:
```json
{
  "type": "ok",
  "message": "创建成功"
}
```

### 2.4 更新新闻

**接口地址**: `PUT /api/admin/news/:id`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "title": "更新后的标题",
  "content": "更新后的内容",
  "category_id": 2,
  "author": "作者"
}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "更新成功"
}
```

### 2.5 删除新闻

**接口地址**: `DELETE /api/admin/news/:id`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "删除成功"
}
```

### 2.6 发布新闻

**接口地址**: `POST /api/admin/news/publish`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "id": 1
}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "发布成功"
}
```

**说明**: 
- 将草稿状态的新闻发布到前台
- 自动设置发布时间为当前时间
- 状态更新为已发布（status=1）

### 2.7 设为草稿

**接口地址**: `POST /api/admin/news/draft`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "id": 1
}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "已设为草稿"
}
```

**说明**: 将已发布的新闻撤回为草稿状态

---

## 三、新闻分类管理API（后台）

### 3.1 获取分类列表

**接口地址**: `GET /api/admin/news-category/list`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "name": "公告",
      "description": "平台公告分类",
      "status": 1,
      "create_time": 1640000000,
      "update_time": 1640000000
    }
  ]
}
```

### 3.2 创建分类

**接口地址**: `POST /api/admin/news-category/create`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "name": "分类名称",
  "description": "分类描述",
  "status": 1
}
```

**字段说明**:
- name: 分类名称（必填）
- description: 分类描述
- status: 状态（0=禁用，1=启用）

**响应示例**:
```json
{
  "type": "ok",
  "message": "创建成功"
}
```

### 3.3 更新分类

**接口地址**: `PUT /api/admin/news-category/:id`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**请求参数**:
```json
{
  "name": "更新后的分类名称",
  "description": "更新后的描述",
  "status": 1
}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "更新成功"
}
```

### 3.4 删除分类

**接口地址**: `DELETE /api/admin/news-category/:id`

**请求头**:
```
Authorization: Bearer {admin_token}
```

**响应示例**:
```json
{
  "type": "ok",
  "message": "删除成功"
}
```

**注意**: 删除分类前请确保该分类下没有新闻

---

## 四、数据模型

### News（新闻表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键ID |
| title | string | 标题 |
| content | string | 内容（HTML） |
| category_id | uint | 分类ID |
| author | string | 作者 |
| status | int8 | 状态（0=草稿，1=已发布） |
| view_count | int | 浏览次数 |
| publish_time | int64 | 发布时间（Unix时间戳） |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |

### NewsCategory（新闻分类表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键ID |
| name | string | 分类名称 |
| description | string | 分类描述 |
| status | int8 | 状态（0=禁用，1=启用） |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |

---

## 五、错误码说明

| 错误类型 | HTTP状态码 | 说明 |
|---------|-----------|------|
| BadRequest | 400 | 请求参数错误 |
| Unauthorized | 401 | 未授权（需要登录或管理员权限） |
| Error | 500 | 服务器内部错误 |

**错误响应示例**:
```json
{
  "type": "error",
  "message": "新闻不存在或已下架"
}
```

---

## 六、使用说明

### 前台调用示例

```javascript
// 获取新闻列表
fetch('/api/news/list?category_id=1&page=1&page_size=10')
  .then(res => res.json())
  .then(data => console.log(data));

// 获取新闻详情
fetch('/api/news/detail?id=1')
  .then(res => res.json())
  .then(data => console.log(data));

// 获取推荐新闻
fetch('/api/news/recommended?limit=5')
  .then(res => res.json())
  .then(data => console.log(data));
```

### 后台调用示例

```javascript
// 创建新闻
fetch('/api/admin/news/create', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + adminToken
  },
  body: JSON.stringify({
    title: '新闻标题',
    content: '<p>新闻内容</p>',
    category_id: 1,
    author: '管理员',
    status: 0
  })
})
.then(res => res.json())
.then(data => console.log(data));

// 发布新闻
fetch('/api/admin/news/publish', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + adminToken
  },
  body: JSON.stringify({ id: 1 })
})
.then(res => res.json())
.then(data => console.log(data));
```

---

## 七、注意事项

1. **权限验证**: 所有后台管理API需要管理员权限，必须在请求头中携带有效的JWT Token
2. **输入验证**: 创建和更新时会自动验证必填字段，标题和内容不能为空
3. **时间戳**: 所有时间字段均为Unix时间戳（秒级）
4. **浏览统计**: 前台访问新闻详情时会自动增加浏览次数
5. **状态管理**: 只有status=1的新闻才会在前台展示
6. **分页限制**: 前台API的page_size最大为100，后台API的page_size最大为100
7. **HTML内容**: 新闻内容支持富文本HTML格式

---

## 八、实现文件说明

### 服务层
- **internal/service/news_service.go**: 前台新闻服务
- **internal/service/admin/news_admin_service.go**: 后台新闻管理服务

### Handler层
- **internal/api/handler/handler.go**: 前台新闻API Handler
- **internal/api/handler/admin/news_handler.go**: 后台新闻管理API Handler

### 路由配置
- **internal/api/router/router.go**: 前台路由配置
- **internal/api/router/admin_router.go**: 后台路由配置

### 数据模型
- **internal/model/model.go**: News和NewsCategory模型定义
