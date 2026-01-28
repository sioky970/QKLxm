# 新闻管理API测试脚本

## Postman测试集合

### 环境变量设置

```json
{
  "base_url": "http://localhost:8080",
  "admin_token": "YOUR_ADMIN_TOKEN_HERE"
}
```

---

## 一、前台API测试

### 1. 获取新闻列表

**请求**
```http
GET {{base_url}}/api/news/list?page=1&page_size=10
```

**预期响应**
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 2. 按分类获取新闻

**请求**
```http
GET {{base_url}}/api/news/list?category_id=1&page=1&page_size=5
```

### 3. 获取新闻详情

**请求**
```http
GET {{base_url}}/api/news/detail?id=1
```

**预期响应**
```json
{
  "type": "ok",
  "message": "获取成功",
  "data": {
    "id": 1,
    "title": "新闻标题",
    "content": "新闻内容",
    "view_count": 101
  }
}
```

### 4. 获取分类列表

**请求**
```http
GET {{base_url}}/api/news/category
```

### 5. 获取推荐新闻

**请求**
```http
GET {{base_url}}/api/news/recommended?limit=5
```

### 6. 获取热门新闻

**请求**
```http
GET {{base_url}}/api/news/hot?limit=5
```

---

## 二、后台API测试

### 1. 获取新闻列表（后台）

**请求**
```http
POST {{base_url}}/api/admin/news/list
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "page": 1,
  "page_size": 10
}
```

### 2. 创建新闻

**请求**
```http
POST {{base_url}}/api/admin/news/create
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "title": "测试新闻标题",
  "content": "<p>这是测试新闻的内容</p>",
  "category_id": 1,
  "author": "测试作者",
  "status": 0
}
```

**预期响应**
```json
{
  "type": "ok",
  "message": "创建成功"
}
```

### 3. 获取新闻详情（后台）

**请求**
```http
GET {{base_url}}/api/admin/news/1
Authorization: Bearer {{admin_token}}
```

### 4. 更新新闻

**请求**
```http
PUT {{base_url}}/api/admin/news/1
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "title": "更新后的标题",
  "content": "<p>更新后的内容</p>",
  "category_id": 1,
  "author": "测试作者"
}
```

### 5. 发布新闻

**请求**
```http
POST {{base_url}}/api/admin/news/publish
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "id": 1
}
```

**预期响应**
```json
{
  "type": "ok",
  "message": "发布成功"
}
```

### 6. 设为草稿

**请求**
```http
POST {{base_url}}/api/admin/news/draft
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "id": 1
}
```

### 7. 删除新闻

**请求**
```http
DELETE {{base_url}}/api/admin/news/1
Authorization: Bearer {{admin_token}}
```

**预期响应**
```json
{
  "type": "ok",
  "message": "删除成功"
}
```

---

## 三、分类管理测试

### 1. 获取分类列表

**请求**
```http
GET {{base_url}}/api/admin/news-category/list
Authorization: Bearer {{admin_token}}
```

### 2. 创建分类

**请求**
```http
POST {{base_url}}/api/admin/news-category/create
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "测试分类",
  "description": "这是测试分类",
  "status": 1
}
```

### 3. 更新分类

**请求**
```http
PUT {{base_url}}/api/admin/news-category/1
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "更新后的分类名",
  "description": "更新后的描述",
  "status": 1
}
```

### 4. 删除分类

**请求**
```http
DELETE {{base_url}}/api/admin/news-category/1
Authorization: Bearer {{admin_token}}
```

---

## 四、cURL测试命令

### 前台API

```bash
# 获取新闻列表
curl "http://localhost:8080/api/news/list?page=1&page_size=10"

# 获取新闻详情
curl "http://localhost:8080/api/news/detail?id=1"

# 获取分类列表
curl "http://localhost:8080/api/news/category"

# 获取推荐新闻
curl "http://localhost:8080/api/news/recommended?limit=5"

# 获取热门新闻
curl "http://localhost:8080/api/news/hot?limit=5"
```

### 后台API

```bash
# 设置Token变量
TOKEN="your_admin_token_here"

# 获取新闻列表
curl -X POST "http://localhost:8080/api/admin/news/list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"page":1,"page_size":10}'

# 创建新闻
curl -X POST "http://localhost:8080/api/admin/news/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试新闻",
    "content": "<p>测试内容</p>",
    "category_id": 1,
    "author": "测试作者",
    "status": 0
  }'

# 发布新闻
curl -X POST "http://localhost:8080/api/admin/news/publish" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":1}'

# 删除新闻
curl -X DELETE "http://localhost:8080/api/admin/news/1" \
  -H "Authorization: Bearer $TOKEN"

# 创建分类
curl -X POST "http://localhost:8080/api/admin/news-category/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试分类",
    "description": "测试描述",
    "status": 1
  }'
```

---

## 五、JavaScript测试脚本

### Node.js测试脚本

```javascript
// test-news-api.js
const axios = require('axios');

const BASE_URL = 'http://localhost:8080';
const ADMIN_TOKEN = 'YOUR_ADMIN_TOKEN';

const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

const adminApi = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${ADMIN_TOKEN}`
  }
});

// 测试前台API
async function testPublicAPI() {
  console.log('=== 测试前台API ===');
  
  try {
    // 1. 获取新闻列表
    console.log('1. 获取新闻列表...');
    const listRes = await api.get('/api/news/list', {
      params: { page: 1, page_size: 10 }
    });
    console.log('✓ 新闻列表获取成功，总数:', listRes.data.data.total);
    
    // 2. 获取新闻详情
    if (listRes.data.data.list.length > 0) {
      const newsId = listRes.data.data.list[0].id;
      console.log('2. 获取新闻详情 ID:', newsId);
      const detailRes = await api.get('/api/news/detail', {
        params: { id: newsId }
      });
      console.log('✓ 新闻详情获取成功:', detailRes.data.data.title);
    }
    
    // 3. 获取分类列表
    console.log('3. 获取分类列表...');
    const categoryRes = await api.get('/api/news/category');
    console.log('✓ 分类列表获取成功，数量:', categoryRes.data.data.length);
    
    // 4. 获取推荐新闻
    console.log('4. 获取推荐新闻...');
    const recommendedRes = await api.get('/api/news/recommended', {
      params: { limit: 5 }
    });
    console.log('✓ 推荐新闻获取成功，数量:', recommendedRes.data.data.length);
    
    // 5. 获取热门新闻
    console.log('5. 获取热门新闻...');
    const hotRes = await api.get('/api/news/hot', {
      params: { limit: 5 }
    });
    console.log('✓ 热门新闻获取成功，数量:', hotRes.data.data.length);
    
  } catch (error) {
    console.error('✗ 测试失败:', error.message);
  }
}

// 测试后台API
async function testAdminAPI() {
  console.log('\n=== 测试后台API ===');
  
  try {
    // 1. 创建新闻
    console.log('1. 创建新闻...');
    const createRes = await adminApi.post('/api/admin/news/create', {
      title: '测试新闻 ' + Date.now(),
      content: '<p>这是自动化测试创建的新闻内容</p>',
      category_id: 1,
      author: '自动化测试',
      status: 0
    });
    console.log('✓ 新闻创建成功');
    
    // 2. 获取新闻列表
    console.log('2. 获取后台新闻列表...');
    const listRes = await adminApi.post('/api/admin/news/list', {
      page: 1,
      page_size: 10
    });
    console.log('✓ 后台新闻列表获取成功，总数:', listRes.data.data.total);
    
    if (listRes.data.data.list.length > 0) {
      const newsId = listRes.data.data.list[0].id;
      
      // 3. 获取新闻详情
      console.log('3. 获取新闻详情 ID:', newsId);
      const detailRes = await adminApi.get(`/api/admin/news/${newsId}`);
      console.log('✓ 新闻详情获取成功:', detailRes.data.data.title);
      
      // 4. 发布新闻
      console.log('4. 发布新闻...');
      await adminApi.post('/api/admin/news/publish', { id: newsId });
      console.log('✓ 新闻发布成功');
      
      // 5. 撤回为草稿
      console.log('5. 撤回为草稿...');
      await adminApi.post('/api/admin/news/draft', { id: newsId });
      console.log('✓ 撤回草稿成功');
    }
    
    // 6. 测试分类管理
    console.log('6. 测试分类管理...');
    const categoryListRes = await adminApi.get('/api/admin/news-category/list');
    console.log('✓ 分类列表获取成功，数量:', categoryListRes.data.data.length);
    
  } catch (error) {
    console.error('✗ 测试失败:', error.response?.data || error.message);
  }
}

// 运行测试
async function runTests() {
  await testPublicAPI();
  await testAdminAPI();
  console.log('\n=== 测试完成 ===');
}

runTests();
```

### 运行测试

```bash
# 安装依赖
npm install axios

# 运行测试
node test-news-api.js
```

---

## 六、性能测试

### 使用Apache Bench (ab)

```bash
# 测试新闻列表接口性能
ab -n 1000 -c 10 "http://localhost:8080/api/news/list?page=1&page_size=10"

# 测试新闻详情接口性能
ab -n 1000 -c 10 "http://localhost:8080/api/news/detail?id=1"
```

### 使用wrk

```bash
# 安装wrk (Linux)
sudo apt-get install wrk

# 测试接口
wrk -t4 -c100 -d30s "http://localhost:8080/api/news/list?page=1&page_size=10"
```

---

## 七、自动化测试脚本

### Python测试脚本

```python
# test_news_api.py
import requests
import json

BASE_URL = "http://localhost:8080"
ADMIN_TOKEN = "YOUR_ADMIN_TOKEN"

headers = {
    "Content-Type": "application/json"
}

admin_headers = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {ADMIN_TOKEN}"
}

def test_public_api():
    print("=== 测试前台API ===")
    
    # 获取新闻列表
    response = requests.get(f"{BASE_URL}/api/news/list", params={"page": 1, "page_size": 10})
    assert response.status_code == 200
    data = response.json()
    assert data["type"] == "ok"
    print(f"✓ 新闻列表获取成功，总数: {data['data']['total']}")
    
    # 获取分类列表
    response = requests.get(f"{BASE_URL}/api/news/category")
    assert response.status_code == 200
    print(f"✓ 分类列表获取成功")
    
    # 获取推荐新闻
    response = requests.get(f"{BASE_URL}/api/news/recommended", params={"limit": 5})
    assert response.status_code == 200
    print(f"✓ 推荐新闻获取成功")

def test_admin_api():
    print("\n=== 测试后台API ===")
    
    # 创建新闻
    news_data = {
        "title": f"测试新闻 {int(time.time())}",
        "content": "<p>测试内容</p>",
        "category_id": 1,
        "author": "测试",
        "status": 0
    }
    response = requests.post(f"{BASE_URL}/api/admin/news/create", 
                            headers=admin_headers, 
                            json=news_data)
    assert response.status_code == 200
    print("✓ 新闻创建成功")
    
    # 获取新闻列表
    response = requests.post(f"{BASE_URL}/api/admin/news/list", 
                            headers=admin_headers,
                            json={"page": 1, "page_size": 10})
    assert response.status_code == 200
    data = response.json()
    print(f"✓ 后台新闻列表获取成功，总数: {data['data']['total']}")

if __name__ == "__main__":
    import time
    test_public_api()
    test_admin_api()
    print("\n=== 所有测试通过 ===")
```

---

## 八、测试检查清单

### 功能测试
- [ ] 前台获取新闻列表（无分类）
- [ ] 前台获取新闻列表（指定分类）
- [ ] 前台获取新闻列表（分页）
- [ ] 前台获取新闻详情
- [ ] 验证浏览次数自动增加
- [ ] 前台获取分类列表
- [ ] 前台获取推荐新闻
- [ ] 前台获取热门新闻
- [ ] 后台创建新闻
- [ ] 后台更新新闻
- [ ] 后台删除新闻
- [ ] 后台发布新闻
- [ ] 后台设为草稿
- [ ] 后台创建分类
- [ ] 后台更新分类
- [ ] 后台删除分类

### 安全测试
- [ ] 未授权访问后台API返回401
- [ ] 草稿状态新闻不在前台显示
- [ ] 必填字段验证
- [ ] SQL注入测试
- [ ] XSS攻击测试

### 性能测试
- [ ] 并发请求测试
- [ ] 大数据量分页测试
- [ ] 响应时间测试

### 边界测试
- [ ] 空标题/内容创建
- [ ] 超长标题/内容
- [ ] 不存在的新闻ID
- [ ] 不存在的分类ID
- [ ] 无效的page/page_size参数

---

## 九、测试报告模板

```markdown
# 新闻管理API测试报告

## 测试环境
- 服务器: http://localhost:8080
- 测试时间: 2024-01-01
- 测试人员: XXX

## 测试结果汇总
- 总测试用例数: 20
- 通过: 18
- 失败: 2
- 通过率: 90%

## 详细测试结果

### 前台API测试 (8/8 通过)
✓ 获取新闻列表
✓ 按分类筛选
✓ 分页功能
✓ 获取详情
✓ 浏览次数增加
✓ 获取分类
✓ 推荐新闻
✓ 热门新闻

### 后台API测试 (10/12 通过)
✓ 创建新闻
✓ 更新新闻
✓ 删除新闻
✓ 发布新闻
✓ 设为草稿
✓ 获取列表
✓ 获取详情
✓ 创建分类
✓ 更新分类
✓ 删除分类
✗ 批量操作 (未实现)
✗ 图片上传 (未实现)

## 问题清单
1. 批量操作功能未实现
2. 图片上传功能未实现

## 建议
1. 添加批量发布/删除功能
2. 集成图片上传接口
3. 添加新闻搜索功能
```
