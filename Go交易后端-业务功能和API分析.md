# Go 交易后端 - 业务功能和API完整分析

## 📋 目录
- [系统概述](#系统概述)
- [技术架构](#技术架构)
- [核心业务模块](#核心业务模块)
- [API接口清单](#api接口清单)
- [数据模型](#数据模型)
- [后台任务](#后台任务)

---

## 系统概述

### 项目信息
- **项目名称**: OKCoinsgp数字货币交易所 Go后端
- **版本**: 1.0
- **端口配置**:
  - API服务: 8080
  - WebSocket服务: 8081
- **技术栈**: Go + Gin + GORM + Redis + MySQL + WebSocket
- **API文档**: Swagger UI (http://localhost:8080/swagger/index.html)

### 系统特点
1. **混合架构**: 前端Go，后台PHP并存
2. **JWT认证**: 基于Token的用户认证
3. **实时推送**: WebSocket支持市场行情实时推送
4. **自动结算**: 秒合约定时自动结算任务
5. **风控系统**: 完整的风险控制机制

---

## 技术架构

### 目录结构
```
1-Go交易后端/
├── cmd/api/                    # 主程序入口
│   └── main.go                 # 应用启动文件
├── config/                     # 配置文件
│   ├── config.go              # 配置加载
│   └── config.yaml            # 配置文件
├── internal/
│   ├── api/                   # API层
│   │   ├── handler/           # 处理器
│   │   │   ├── admin/         # 管理端处理器
│   │   │   ├── handler.go     # 用户端处理器
│   │   │   └── kyc_handler.go # KYC处理器
│   │   ├── middleware/        # 中间件
│   │   └── router/            # 路由配置
│   ├── model/                 # 数据模型
│   ├── pkg/                   # 公共包
│   │   ├── cache/             # Redis缓存
│   │   ├── database/          # 数据库连接
│   │   ├── logger/            # 日志系统
│   │   └── response/          # 响应封装
│   ├── scheduler/             # 定时任务
│   ├── service/               # 业务逻辑层
│   │   ├── admin/             # 管理端服务
│   │   ├── *_service.go       # 各业务服务
│   │   └── types.go           # 类型定义
│   └── websocket/             # WebSocket服务
└── docs/                      # Swagger文档
```

### 核心组件

#### 1. 数据库 (MySQL)
- 数据库名: `bibi2022`
- 连接池: 最大100连接，空闲10连接
- 字符集: utf8mb4

#### 2. 缓存 (Redis)
- 端口: 6379
- 用途: 验证码、Token、行情缓存
- 连接池: 100

#### 3. 日志系统
- 级别: debug/info/warn/error
- 文件: logs/app.log
- 自动轮转: 100MB, 保留30天

---

## 核心业务模块

### 1. 用户模块 (User Module)

#### 功能特性
- ✅ 用户注册（手机号/邮箱）
- ✅ 用户登录（支持普通密码、手势密码）
- ✅ 密码找回
- ✅ 修改登录密码
- ✅ 修改支付密码
- ✅ 用户信息管理
- ✅ 邀请码系统
- ✅ 验证码发送

#### 核心流程
**注册流程:**
1. 发送验证码 → Redis缓存（5分钟有效）
2. 验证验证码 → 检查账号是否存在
3. 验证邀请码 → 创建用户
4. 自动创建各币种钱包

**登录流程:**
1. 验证账号密码/手势密码
2. 检查账号状态（是否锁定）
3. 生成JWT Token
4. 更新登录时间和IP

---

### 2. 钱包模块 (Wallet Module)

#### 功能特性
- ✅ 钱包列表查询（各币种余额）
- ✅ 钱包详情
- ✅ 充值申请
- ✅ 提现申请
- ✅ 钱包流水记录

#### 钱包类型
- **法币钱包** (legal_balance): 用于法币交易
- **币币钱包** (change_balance): 用于币币交易
- **合约钱包** (lever_balance): 用于合约交易
- **秒合约钱包** (micro_balance): 用于秒合约交易
- **锁定余额**: 挂单锁定、合约保证金锁定

#### 充值/提现流程
**充值流程:**
1. 提交充值申请（金额、凭证、Hash等）
2. 管理员审核
3. 审核通过 → 增加余额

**提现流程:**
1. 验证支付密码
2. 扣除余额 → 锁定
3. 管理员审核
4. 审核通过 → 转账，拒绝 → 解锁余额

---

### 3. 币币交易模块 (Spot Trading)

#### 功能特性
- ✅ 限价单交易
- ✅ 市价单交易
- ✅ 条件单（触发单）
- ✅ 订单撤销
- ✅ 委托订单列表
- ✅ 成交历史记录

#### 交易流程
**下单流程:**
1. 验证交易对、价格、数量
2. 检查余额是否足够
3. 锁定对应资产
4. 创建订单记录
5. 匹配引擎撮合
6. 成交 → 解锁并转账，未成交 → 保持锁定

**撮合机制:**
- 价格优先、时间优先
- 买单匹配最低卖单
- 卖单匹配最高买单

#### 订单状态
- 0: 待成交
- 1: 部分成交
- 2: 完全成交
- 3: 已撤销

---

### 4. 合约交易模块 (Futures Trading)

#### 功能特性
- ✅ 开多/开空
- ✅ 杠杆交易（5倍、10倍、20倍、50倍）
- ✅ 限价单/市价单
- ✅ 止盈止损设置
- ✅ 持仓管理
- ✅ 平仓操作
- ✅ 保证金管理
- ✅ 强制平仓（爆仓）

#### 交易类型
- **多头** (long): 做多，预期价格上涨
- **空头** (short): 做空，预期价格下跌

#### 计算公式
```
保证金 = 开仓价格 × 数量 ÷ 杠杆倍数
手续费 = 开仓价格 × 数量 × 费率
盈亏计算:
  - 多头: (当前价 - 开仓价) × 数量 × 杠杆倍数
  - 空头: (开仓价 - 当前价) × 数量 × 杠杆倍数
强平价格:
  - 多头: 开仓价 × (1 - 1/杠杆)
  - 空头: 开仓价 × (1 + 1/杠杆)
```

#### 风控机制
- 实时监控持仓盈亏
- 保证金率低于阈值 → 预警
- 保证金率低于维持保证金率 → 强制平仓

---

### 5. 秒合约模块 (Micro Trading)

#### 功能特性
- ✅ 预测涨跌
- ✅ 多时长选择（5秒、10秒、15秒、30秒、60秒）
- ✅ 固定盈利率
- ✅ 自动结算
- ✅ 订单列表查询

#### 交易流程
1. 选择币种交易对
2. 选择时长和方向（买涨/买跌）
3. 输入投资金额和盈利率
4. 记录开仓价格
5. 定时任务自动结算
6. 比较开盘价和收盘价
7. 判断盈亏并结算

#### 盈亏计算
```
预测正确: 本金 + 本金 × 盈利率
预测错误: 损失本金
```

#### 结算状态
- 0: 待结算
- 1: 已结算（盈利）
- 2: 已结算（亏损）
- 3: 已撤销

---

### 6. 行情模块 (Market Module)

#### 功能特性
- ✅ 实时行情获取
- ✅ K线数据查询
- ✅ 币种报价查询
- ✅ 市场深度数据
- ✅ 24小时成交数据

#### 数据源
- **火币API**: 主要行情数据来源
- **自定义行情**: 管理员可手动设置行情

#### K线周期
- 1分钟 (1min)
- 5分钟 (5min)
- 15分钟 (15min)
- 30分钟 (30min)
- 1小时 (1hour)
- 4小时 (4hour)
- 1天 (1day)
- 1周 (1week)
- 1月 (1mon)

#### 行情数据结构
```json
{
  "currency_id": 1,
  "legal_id": 2,
  "price": 50000.00,
  "change": 2.5,
  "volume": 1234.56,
  "high": 51000.00,
  "low": 49000.00
}
```

---

### 7. KYC实名认证模块

#### 功能特性
- ✅ 提交实名认证
- ✅ 上传身份证照片（正面、反面、手持）
- ✅ 管理员审核
- ✅ 审核状态查询

#### 认证状态
- 0: 未认证
- 1: 审核中
- 2: 已通过
- 3: 已拒绝

---

### 8. 管理后台模块 (Admin Module)

#### 8.1 用户管理
- ✅ 用户列表查询
- ✅ 用户详情查看
- ✅ 用户信息修改
- ✅ 冻结/激活用户
- ✅ 重置密码
- ✅ 余额调整
- ✅ 批量设置风控
- ✅ 查看用户钱包
- ✅ 用户搜索

#### 8.2 币种管理
- ✅ 币种列表
- ✅ 添加币种
- ✅ 编辑币种
- ✅ 删除币种
- ✅ 币种显示开关
- ✅ 更新汇率
- ✅ 交易对管理
- ✅ 交易对风控设置

#### 8.3 钱包管理
- ✅ 钱包列表
- ✅ 钱包详情
- ✅ 余额修改
- ✅ 冻结/激活钱包
- ✅ 提现审核（通过/拒绝）
- ✅ 充值审核（通过/拒绝）

#### 8.4 交易管理
- ✅ 币币交易列表
- ✅ 合约交易列表
- ✅ 秒合约交易列表
- ✅ 法币交易列表
- ✅ C2C交易列表
- ✅ 订单撤销
- ✅ 账户日志查询
- ✅ 商家申请审核

#### 8.5 统计功能
- ✅ 数据看板（Dashboard）
- ✅ 用户统计
- ✅ 交易统计
- ✅ 财务统计
- ✅ Top交易者排行
- ✅ 币种统计

#### 8.6 风控管理
- ✅ 风控配置
- ✅ 用户风控列表
- ✅ 设置用户风控等级
- ✅ 交易对风控管理
- ✅ 秒合约订单风控

#### 8.7 系统配置
- ✅ 配置列表
- ✅ 获取配置
- ✅ 更新配置
- ✅ 批量更新配置

#### 8.8 管理员管理
- ✅ 管理员列表
- ✅ 添加管理员
- ✅ 编辑管理员
- ✅ 删除管理员

#### 8.9 角色管理
- ✅ 角色列表
- ✅ 角色详情
- ✅ 创建角色
- ✅ 更新角色
- ✅ 删除角色
- ✅ 分配权限

#### 8.10 等级管理
- ✅ 等级列表
- ✅ 创建等级
- ✅ 更新等级
- ✅ 删除等级

---

### 9. WebSocket模块

#### 功能特性
- ✅ 实时连接管理
- ✅ 频道订阅机制
- ✅ 心跳检测
- ✅ 断线重连

#### 支持频道
- **market**: 市场行情推送
- **kline**: K线数据推送
- **orderbook**: 订单簿推送
- **trade**: 成交记录推送

#### 消息格式
```json
{
  "type": "channel_message",
  "channel": "market",
  "data": { ... },
  "timestamp": 1706000000
}
```

#### 订阅示例
```json
// 订阅
{
  "type": "subscribe",
  "channel": "market"
}

// 取消订阅
{
  "type": "unsubscribe",
  "channel": "market"
}
```

---

## API接口清单

### 公开接口（无需认证）

#### 用户相关
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/user/register | 用户注册 |
| POST | /api/user/login | 用户登录 |
| POST | /api/user/send_code | 发送验证码 |
| POST | /api/user/reset_password | 重置密码 |

#### 行情相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/quotation/new | 获取最新行情 |
| POST | /api/market/market | 获取市场数据 |
| GET | /api/kline | 获取K线数据 |
| GET | /api/currency/quotation_new | 获取币种行情 |

---

### 需认证接口（需JWT Token）

#### 用户模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/info | 获取用户信息 |
| POST | /api/user/update | 更新用户信息 |
| POST | /api/user/change_password | 修改登录密码 |
| POST | /api/user/change_pay_password | 修改支付密码 |

#### 钱包模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/wallet/list | 钱包列表 |
| GET | /api/wallet/info | 钱包详情 |
| POST | /api/wallet/recharge | 充值申请 |
| POST | /api/wallet/withdraw | 提现申请 |
| GET | /api/wallet/logs | 钱包流水 |

#### 币币交易模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/transaction/submit | 提交订单 |
| POST | /api/transaction/cancel | 撤销订单 |
| GET | /api/transaction/list | 委托列表 |
| GET | /api/transaction/history | 成交历史 |

#### 合约交易模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/lever/submit | 开仓/加仓 |
| POST | /api/lever/close | 平仓 |
| GET | /api/lever/position | 持仓列表 |
| GET | /api/lever/history | 历史订单 |

#### 秒合约模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/micro/submit | 提交订单 |
| GET | /api/micro/list | 订单列表 |

#### KYC模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/kyc/submit | 提交认证 |
| GET | /api/kyc/status | 认证状态 |

---

### 管理后台接口（需管理员权限）

#### 用户管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/user/list | 用户列表 |
| GET | /api/admin/user/:id | 用户详情 |
| PUT | /api/admin/user/:id | 更新用户 |
| DELETE | /api/admin/user/:id | 删除用户 |
| POST | /api/admin/user/freeze | 冻结用户 |
| POST | /api/admin/user/activate | 激活用户 |
| POST | /api/admin/user/reset-password | 重置密码 |
| POST | /api/admin/user/adjust-balance | 调整余额 |
| POST | /api/admin/user/batch-risk | 批量设置风控 |
| GET | /api/admin/user/:id/wallets | 用户钱包 |
| POST | /api/admin/user/search | 搜索用户 |

#### 币种管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/currency/list | 币种列表 |
| GET | /api/admin/currency/:id | 币种详情 |
| POST | /api/admin/currency/create | 创建币种 |
| PUT | /api/admin/currency/:id | 更新币种 |
| DELETE | /api/admin/currency/:id | 删除币种 |
| POST | /api/admin/currency/toggle-display | 显示开关 |
| POST | /api/admin/currency/update-rate | 更新汇率 |
| POST | /api/admin/currency/match/list | 交易对列表 |
| POST | /api/admin/currency/match/risk | 交易对风控 |

#### 钱包管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/wallet/list | 钱包列表 |
| GET | /api/admin/wallet/:id | 钱包详情 |
| POST | /api/admin/wallet/update-balance | 更新余额 |
| POST | /api/admin/wallet/freeze | 冻结钱包 |
| POST | /api/admin/wallet/activate | 激活钱包 |
| POST | /api/admin/wallet/withdrawals | 提现列表 |
| POST | /api/admin/wallet/withdrawals/approve | 批准提现 |
| POST | /api/admin/wallet/withdrawals/reject | 拒绝提现 |
| POST | /api/admin/wallet/charge/list | 充值列表 |
| POST | /api/admin/wallet/charge/approve | 批准充值 |
| POST | /api/admin/wallet/charge/reject | 拒绝充值 |

#### 交易管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/transaction/spot/list | 币币交易列表 |
| POST | /api/admin/transaction/lever/list | 合约交易列表 |
| POST | /api/admin/transaction/micro/list | 秒合约列表 |
| POST | /api/admin/transaction/cancel | 撤销订单 |
| POST | /api/admin/transaction/legal/list | 法币交易列表 |
| POST | /api/admin/transaction/c2c/list | C2C交易列表 |
| POST | /api/admin/transaction/account-log/list | 账户日志 |
| POST | /api/admin/transaction/seller/list | 商家列表 |
| POST | /api/admin/transaction/seller/approve | 批准商家 |
| POST | /api/admin/transaction/seller/reject | 拒绝商家 |
| GET | /api/admin/transaction/micro/config | 秒合约配置 |
| GET | /api/admin/transaction/lever/multiple | 杠杆倍数 |

#### 统计功能
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/statistics/dashboard | 数据看板 |
| GET | /api/admin/statistics/user | 用户统计 |
| GET | /api/admin/statistics/trade | 交易统计 |
| GET | /api/admin/statistics/finance | 财务统计 |
| GET | /api/admin/statistics/top-traders | Top交易者 |
| GET | /api/admin/statistics/currency | 币种统计 |

#### 风控管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/risk/config | 获取风控配置 |
| POST | /api/admin/risk/config | 更新风控配置 |
| POST | /api/admin/risk/users | 用户风控列表 |
| POST | /api/admin/risk/users/set | 设置用户风控 |
| GET | /api/admin/risk/matches | 交易对风控列表 |
| POST | /api/admin/risk/matches/set | 设置交易对风控 |
| POST | /api/admin/risk/orders | 订单风控列表 |
| POST | /api/admin/risk/orders/set | 设置订单风控 |

#### 实名认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/kyc/list | KYC列表 |
| GET | /api/admin/kyc/:id | KYC详情 |
| POST | /api/admin/kyc/review | 审核KYC |

#### 系统配置
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/config/list | 配置列表 |
| GET | /api/admin/config/:key | 获取配置 |
| POST | /api/admin/config/update | 更新配置 |
| POST | /api/admin/config/batch-update | 批量更新 |

#### 管理员管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/admin/list | 管理员列表 |
| GET | /api/admin/admin/:id | 管理员详情 |
| POST | /api/admin/admin/create | 创建管理员 |
| PUT | /api/admin/admin/update | 更新管理员 |
| DELETE | /api/admin/admin/:id | 删除管理员 |

#### 角色管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/role/list | 角色列表 |
| GET | /api/admin/role/:id | 角色详情 |
| POST | /api/admin/role/create | 创建角色 |
| PUT | /api/admin/role/update | 更新角色 |
| DELETE | /api/admin/role/:id | 删除角色 |
| POST | /api/admin/role/assign-permissions | 分配权限 |

#### 等级管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/level/list | 等级列表 |
| POST | /api/admin/level/create | 创建等级 |
| PUT | /api/admin/level/update | 更新等级 |
| DELETE | /api/admin/level/:id | 删除等级 |

---

## 数据模型

### 核心表结构

#### 1. 用户表 (users)
```go
type User struct {
    ID                uint      // 用户ID
    AccountNumber     string    // 账号
    Phone             string    // 手机号
    Email             string    // 邮箱
    Password          string    // 登录密码
    PayPassword       string    // 支付密码
    GesturePassword   string    // 手势密码
    ExtensionCode     string    // 邀请码
    ParentID          uint      // 上级ID
    Status            int8      // 状态: 0正常 1锁定
    IsReal            int8      // 是否实名: 0否 1是
    Risk              int8      // 风控等级
    // ... 更多字段
}
```

#### 2. 钱包表 (users_wallet)
```go
type UsersWallet struct {
    ID            uint      // 钱包ID
    UserID        uint      // 用户ID
    CurrencyID    uint      // 币种ID
    LegalBalance  float64   // 法币余额
    ChangeBalance float64   // 币币余额
    LeverBalance  float64   // 合约余额
    MicroBalance  float64   // 秒合约余额
    LockBalance   float64   // 锁定余额
    Status        int8      // 状态: 0冻结 1正常
}
```

#### 3. 币种表 (currency)
```go
type Currency struct {
    ID           uint      // 币种ID
    Name         string    // 币种名称
    Logo         string    // 币种图标
    Type         string    // 币种类型
    IsLegal      int8      // 是否法币
    IsLever      int8      // 是否支持合约
    IsMicro      int8      // 是否支持秒合约
    IsDisplay    int8      // 是否显示
    Rate         float64   // 汇率
}
```

#### 4. 交易对表 (currency_matches)
```go
type CurrencyMatch struct {
    ID              uint      // 交易对ID
    CurrencyID      uint      // 币种ID
    LegalID         uint      // 法币ID
    Symbol          string    // 交易对符号
    MinPrice        float64   // 最小价格
    MaxPrice        float64   // 最大价格
    MinNumber       float64   // 最小数量
    MaxNumber       float64   // 最大数量
    Fee             float64   // 手续费率
    IsDisplay       int8      // 是否显示
    RiskGroupResult int8      // 风控组结果
}
```

#### 5. 币币交易表 (transaction)
```go
type Transaction struct {
    ID         uint      // 订单ID
    UserID     uint      // 用户ID
    CurrencyID uint      // 币种ID
    LegalID    uint      // 法币ID
    Type       int8      // 类型: 1买入 2卖出
    Price      float64   // 价格
    Number     float64   // 数量
    DealNumber float64   // 成交数量
    Status     int8      // 状态
    Fee        float64   // 手续费
}
```

#### 6. 合约交易表 (lever_transaction)
```go
type LeverTransaction struct {
    ID              uint      // 订单ID
    UserID          uint      // 用户ID
    CurrencyID      uint      // 币种ID
    LegalID         uint      // 法币ID
    Type            int8      // 类型: 1多头 2空头
    Leverage        int       // 杠杆倍数
    Price           float64   // 开仓价格
    Number          float64   // 数量
    InitialMargin   float64   // 保证金
    StopLossPrice   float64   // 止损价
    TakeProfitPrice float64   // 止盈价
    Profits         float64   // 盈亏
    Status          int8      // 状态
}
```

#### 7. 秒合约表 (micro_orders)
```go
type MicroOrder struct {
    ID              uint      // 订单ID
    UserID          uint      // 用户ID
    CurrencyID      uint      // 币种ID
    LegalID         uint      // 法币ID
    Type            int8      // 类型: 1买涨 2买跌
    Duration        int       // 持续时间(秒)
    Number          float64   // 投资金额
    ProfitRatio     float64   // 盈利率
    OpenPrice       float64   // 开盘价
    EndPrice        float64   // 收盘价
    ProfitResult    int8      // 盈亏结果
    Profits         float64   // 盈亏金额
    Status          int8      // 状态
}
```

---

## 后台任务

### 1. 秒合约自动结算任务

#### 功能说明
- 每秒执行一次
- 查询到期的秒合约订单
- 获取当前市场价格
- 判断涨跌结果
- 结算盈亏
- 更新钱包余额

#### 执行流程
```
1. 查询 status=0 且到期的订单
2. 获取币种当前价格
3. 比较开盘价和收盘价
   - 买涨 + 价格上涨 = 盈利
   - 买跌 + 价格下跌 = 盈利
   - 其他情况 = 亏损
4. 计算盈亏金额
5. 更新钱包余额
6. 更新订单状态
7. 记录账户日志
```

#### 代码位置
- `internal/scheduler/micro_order_scheduler.go`

---

## 安全机制

### 1. 认证授权
- **JWT Token**: 有效期24小时
- **Bearer Token**: 请求头携带
- **管理员权限**: 双重验证（JWT + AdminAuth中间件）

### 2. 密码安全
- **Bcrypt加密**: 密码和支付密码均使用Bcrypt加密
- **验证码**: 5分钟有效期，Redis存储
- **支付密码**: 敏感操作二次验证

### 3. 风控系统
- **用户风控**: 针对单个用户的风险等级设置
- **交易对风控**: 针对交易对的风险控制
- **订单风控**: 秒合约订单风险预判

### 4. 日志审计
- **操作日志**: 记录所有关键操作
- **账户日志**: 记录所有资金变动
- **登录日志**: 记录登录IP和时间

---

## 性能优化

### 1. 缓存策略
- **验证码缓存**: Redis，5分钟过期
- **Token缓存**: Redis，24小时过期
- **行情缓存**: Redis，实时更新

### 2. 数据库优化
- **连接池**: 最大100连接
- **索引优化**: 关键字段建立索引
- **分表策略**: 大表考虑分表

### 3. 并发控制
- **锁机制**: 防止超卖、重复成交
- **事务处理**: 保证数据一致性
- **队列机制**: 异步处理耗时任务

---

## 错误码说明

### HTTP状态码
- 200: 成功
- 400: 请求参数错误
- 401: 未授权（未登录或Token失效）
- 403: 无权限
- 404: 资源不存在
- 500: 服务器内部错误

### 业务错误码
```json
{
  "type": "error",
  "message": "具体错误信息",
  "data": null
}
```

---

## 部署说明

### 环境要求
- Go 1.18+
- MySQL 5.7+
- Redis 5.0+

### 启动命令
```bash
# 开发环境
go run cmd/api/main.go

# 生产环境
go build -o exchange-api cmd/api/main.go
./exchange-api -c config/config.yaml
```

### 配置文件
编辑 `config/config.yaml`，修改数据库、Redis等配置

---

## 开发指南

### 添加新接口
1. 在 `internal/service/` 添加业务逻辑
2. 在 `internal/api/handler/` 添加处理器
3. 在 `internal/api/router/` 注册路由
4. 添加Swagger注释
5. 运行 `swag init` 生成文档

### 添加新模型
1. 在 `internal/model/model.go` 定义结构体
2. 实现 `TableName()` 方法
3. 数据库中创建对应表

---

## 总结

### 系统优势
✅ **架构清晰**: 分层设计，职责明确
✅ **功能完整**: 覆盖交易所核心业务
✅ **安全可靠**: 多重安全机制
✅ **性能优良**: 缓存、连接池优化
✅ **易于扩展**: 模块化设计
✅ **文档完善**: Swagger自动生成API文档

### 技术亮点
1. JWT认证 + 中间件鉴权
2. WebSocket实时推送
3. 定时任务自动结算
4. Redis缓存加速
5. GORM ORM简化数据库操作
6. Swagger API文档

### 改进建议
1. 增加单元测试覆盖率
2. 完善错误处理和日志
3. 优化数据库查询性能
4. 增加API限流和防刷机制
5. 完善监控和告警系统

---

## 附录

### API请求示例

#### 1. 用户注册
```bash
POST /api/user/register
Content-Type: application/json

{
  "type": "mobile",
  "user_string": "13800138000",
  "password": "123456",
  "re_password": "123456",
  "code": "123456",
  "extension_code": "abc123",
  "country_code": 86
}
```

#### 2. 用户登录
```bash
POST /api/user/login
Content-Type: application/json

{
  "user_string": "13800138000",
  "password": "123456",
  "area_code_id": 86
}
```

#### 3. 币币交易下单
```bash
POST /api/transaction/submit
Authorization: Bearer {token}
Content-Type: application/json

{
  "currency_id": 1,
  "legal_id": 2,
  "type": "limit",
  "side": "buy",
  "price": 50000,
  "quantity": 0.1
}
```

#### 4. 秒合约下单
```bash
POST /api/micro/submit
Authorization: Bearer {token}
Content-Type: application/json

{
  "currency_id": 1,
  "legal_id": 2,
  "direction": "rise",
  "duration": 60,
  "amount": 100,
  "profit_ratio": 0.85
}
```

---

**文档版本**: v1.0  
**更新时间**: 2024-01-24  
**作者**: 系统架构分析

