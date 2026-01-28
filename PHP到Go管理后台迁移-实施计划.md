# PHP到Go管理后台迁移 - 详细实施计划

## 实施概览

### 迁移原则
1. **渐进式迁移**: 按功能模块逐步迁移，每个模块独立可测试
2. **API兼容**: 新API设计保持与前端的兼容性
3. **数据库共享**: Go后端直接使用现有MySQL数据库
4. **权限复用**: 适配现有权限表结构

---

## 阶段一：基础架构完善 (P0 - 核心优先)

### 1.1 管理员认证系统

**目标文件:** `admin-go/api/v1/system/auth.go`

```go
// 需要实现的API
POST /api/v1/admin/login          // 管理员登录
POST /api/v1/admin/logout         // 退出登录
GET  /api/v1/admin/info           // 获取当前管理员信息
POST /api/v1/admin/change-password // 修改密码
```

**数据库表适配:**
```sql
-- 使用现有admin表
SELECT id, username, password, role_id, is_super, created_at, updated_at 
FROM admin;
```

**实现步骤:**
1. 创建Admin模型 (`model/admin.go`)
2. 实现JWT生成和验证
3. 实现登录/登出接口
4. 添加中间件保护路由

---

### 1.2 RBAC权限管理

**目标文件:** 
- `admin-go/api/v1/system/role.go`
- `admin-go/api/v1/system/permission.go`

```go
// 角色管理
GET    /api/v1/admin/role/list              // 角色列表
POST   /api/v1/admin/role/create            // 创建角色
PUT    /api/v1/admin/role/:id               // 更新角色
DELETE /api/v1/admin/role/:id               // 删除角色

// 权限管理
GET    /api/v1/admin/permission/list        // 权限列表
POST   /api/v1/admin/role/assign-permission // 分配权限
```

**数据库表:**
```sql
-- admin_role 角色表
CREATE TABLE admin_role (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    created_at DATETIME,
    updated_at DATETIME
);

-- admin_role_permission 权限关联表
CREATE TABLE admin_role_permission (
    id INT PRIMARY KEY AUTO_INCREMENT,
    role_id INT NOT NULL,
    module_id INT NOT NULL,
    action_id INT
);

-- admin_module 模块表
CREATE TABLE admin_module (
    id INT PRIMARY KEY AUTO_INCREMENT,
    pid INT DEFAULT 0,
    name VARCHAR(50),
    route VARCHAR(100),
    icon VARCHAR(50),
    sort INT DEFAULT 0
);

-- admin_module_action 操作表
CREATE TABLE admin_module_action (
    id INT PRIMARY KEY AUTO_INCREMENT,
    module_id INT NOT NULL,
    name VARCHAR(50),
    route VARCHAR(100)
);
```

---

### 1.3 用户管理完善

**目标文件:** `admin-go/api/v1/exchange/user.go`

**需要新增的API:**
```go
// 余额调节 - 核心功能
POST /api/v1/admin/user/adjust-balance
Request: {
    "user_id": 1,
    "currency_id": 3,
    "balance_type": "legal_balance", // legal_balance/change_balance/lever_balance/micro_balance
    "amount": 100.5,                  // 正数加款，负数扣款
    "reason": "后台调节"
}

// 钱包锁定
POST /api/v1/admin/user/wallet/lock
Request: {
    "wallet_id": 1,
    "lock_type": "all" // all/legal/change/lever
}

// 批量风控设置
POST /api/v1/admin/user/batch-risk
Request: {
    "user_ids": [1, 2, 3],
    "risk_level": 1  // 0-正常, 1-风控
}

// 用户导出
GET /api/v1/admin/user/export?account=&status=
```

**服务层实现:** `admin-go/service/user_service.go`

```go
// AdjustBalance 余额调节
func (s *UserSrv) AdjustBalance(userId uint, currencyId int, balanceType string, amount float64, reason string) error {
    return global.DB.Transaction(func(tx *gorm.DB) error {
        // 1. 查找用户钱包
        var wallet model.UsersWallet
        if err := tx.Where("user_id = ? AND currency = ?", userId, currencyId).First(&wallet).Error; err != nil {
            return err
        }
        
        // 2. 更新余额
        updates := map[string]interface{}{}
        switch balanceType {
        case "legal_balance":
            updates["legal_balance"] = gorm.Expr("legal_balance + ?", amount)
        case "change_balance":
            updates["change_balance"] = gorm.Expr("change_balance + ?", amount)
        case "lever_balance":
            updates["lever_balance"] = gorm.Expr("lever_balance + ?", amount)
        case "micro_balance":
            updates["micro_balance"] = gorm.Expr("micro_balance + ?", amount)
        default:
            return errors.New("无效的余额类型")
        }
        
        if err := tx.Model(&wallet).Updates(updates).Error; err != nil {
            return err
        }
        
        // 3. 记录日志
        logType := getAccountLogType(balanceType, amount > 0)
        accountLog := model.AccountLog{
            UserId:      userId,
            Value:       amount,
            Currency:    currencyId,
            Type:        logType,
            Info:        reason,
            CreatedTime: time.Now().Unix(),
        }
        return tx.Create(&accountLog).Error
    })
}
```

---

### 1.4 提币审核完善

**目标文件:** `admin-go/api/v1/exchange/withdrawal.go`

```go
// 提币列表
POST /api/v1/admin/withdrawal/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "status": 0,  // 0-待审核, 1-已通过, 2-已拒绝
    "currency_id": 0
}

// 提币详情
GET /api/v1/admin/withdrawal/:id

// 审核通过
POST /api/v1/admin/withdrawal/approve
Request: {
    "id": 1,
    "method": "manual", // manual-手动, auto-自动链上转账
    "txid": "",         // 手动时填写交易哈希
    "notes": ""
}

// 审核拒绝
POST /api/v1/admin/withdrawal/reject
Request: {
    "id": 1,
    "reason": "拒绝原因"
}
```

**服务层实现:** `admin-go/service/withdrawal_service.go`

```go
// ApproveWithdrawal 审核通过提币
func (s *WithdrawalSrv) ApproveWithdrawal(id uint, method string, txid string, notes string) error {
    return global.DB.Transaction(func(tx *gorm.DB) error {
        var walletOut model.UsersWalletOut
        if err := tx.First(&walletOut, id).Error; err != nil {
            return err
        }
        
        if walletOut.Status != 0 {
            return errors.New("该申请已处理")
        }
        
        // 更新状态
        updates := map[string]interface{}{
            "status":     1,
            "updated_at": time.Now(),
        }
        
        if method == "manual" {
            updates["real_number"] = walletOut.Number
            updates["notes"] = notes
            if txid != "" {
                updates["info"] = txid
            }
        }
        
        if err := tx.Model(&walletOut).Updates(updates).Error; err != nil {
            return err
        }
        
        // 记录账户日志
        return s.recordAccountLog(tx, walletOut.UserId, walletOut.Currency, -walletOut.Number, "提币成功")
    })
}

// RejectWithdrawal 拒绝提币（退还余额）
func (s *WithdrawalSrv) RejectWithdrawal(id uint, reason string) error {
    return global.DB.Transaction(func(tx *gorm.DB) error {
        var walletOut model.UsersWalletOut
        if err := tx.First(&walletOut, id).Error; err != nil {
            return err
        }
        
        if walletOut.Status != 0 {
            return errors.New("该申请已处理")
        }
        
        // 退还余额到法币账户
        if err := tx.Model(&model.UsersWallet{}).
            Where("user_id = ? AND currency = ?", walletOut.UserId, walletOut.Currency).
            Update("legal_balance", gorm.Expr("legal_balance + ?", walletOut.Number)).Error; err != nil {
            return err
        }
        
        // 更新提币状态
        return tx.Model(&walletOut).Updates(map[string]interface{}{
            "status": 2,
            "notes":  reason,
        }).Error
    })
}
```

---

### 1.5 实名认证审核

**目标文件:** `admin-go/api/v1/exchange/kyc.go`

```go
// KYC列表
GET /api/v1/admin/kyc/list?page=1&page_size=10&account=&status=

// KYC详情
GET /api/v1/admin/kyc/:id

// 审核通过
POST /api/v1/admin/kyc/approve
Request: {
    "id": 1
}

// 审核拒绝
POST /api/v1/admin/kyc/reject
Request: {
    "id": 1,
    "reason": "资料不清晰"
}

// 删除记录
DELETE /api/v1/admin/kyc/:id
```

---

### 1.6 系统设置

**目标文件:** `admin-go/api/v1/system/setting.go`

```go
// 配置列表
GET /api/v1/admin/setting/list

// 更新配置
POST /api/v1/admin/setting/update
Request: {
    "settings": [
        {"key": "site_name", "value": "交易所"},
        {"key": "site_logo", "value": "/logo.png"}
    ]
}

// 获取单个配置
GET /api/v1/admin/setting/:key

// 基础设置页面数据
GET /api/v1/admin/setting/base

// 保存基础设置
POST /api/v1/admin/setting/base
```

---

## 阶段二：交易管理完善 (P1)

### 2.1 币币交易详细管理

**目标文件:** `admin-go/api/v1/exchange/spot_trade.go`

```go
// 买入订单列表
POST /api/v1/admin/transaction/in/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "currency_id": 0,
    "status": 0
}

// 卖出订单列表
POST /api/v1/admin/transaction/out/list

// 成交记录列表
POST /api/v1/admin/transaction/complete/list

// 订单详情
GET /api/v1/admin/transaction/:type/:id
```

---

### 2.2 杠杆交易管理

**目标文件:** `admin-go/api/v1/exchange/lever.go`

```go
// 杠杆订单列表
POST /api/v1/admin/lever/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "currency_id": 0,
    "legal_id": 0,
    "status": 0,  // 0-持仓中, 1-已平仓
    "type": ""    // buy/sell
}

// 杠杆倍数配置列表
GET /api/v1/admin/lever/multiple/list

// 添加杠杆倍数
POST /api/v1/admin/lever/multiple/create
Request: {
    "type": "multiple", // multiple-倍数, hand-手数
    "value": "100"
}

// 删除杠杆倍数
DELETE /api/v1/admin/lever/multiple/:id

// 风险率列表
POST /api/v1/admin/lever/hazard/list
Request: {
    "legal_id": 3,
    "threshold": 0.5  // 风险率阈值
}

// 价格干预
POST /api/v1/admin/lever/hazard/handle
Request: {
    "trade_id": 1,
    "update_price": 50000.5,
    "write_market": 1  // 是否写入行情
}

// 导出杠杆订单
GET /api/v1/admin/lever/export?account_number=&currency_id=&status=
```

---

### 2.3 秒合约管理完善

**目标文件:** `admin-go/api/v1/exchange/micro.go`

```go
// 金额配置列表
GET /api/v1/admin/micro/number/list

// 添加金额配置
POST /api/v1/admin/micro/number/create
Request: {
    "currency_id": 3,
    "number": "100,200,500,1000"
}

// 删除金额配置
DELETE /api/v1/admin/micro/number/:id

// 时间配置列表
GET /api/v1/admin/micro/seconds/list

// 添加时间配置
POST /api/v1/admin/micro/seconds/create
Request: {
    "seconds": 30,
    "profit_ratio": 0.85,  // 盈利比例
    "status": 1
}

// 时间配置状态切换
POST /api/v1/admin/micro/seconds/status
Request: {
    "id": 1,
    "status": 1
}

// 删除时间配置
DELETE /api/v1/admin/micro/seconds/:id

// 秒合约订单列表
POST /api/v1/admin/micro/order/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "match_id": 0,
    "status": 0,
    "result": ""  // win/lose/draw
}

// 编辑订单(风控)
PUT /api/v1/admin/micro/order/:id
Request: {
    "result_profit": -100,  // 强制设置结果
    "end_price": 50000
}

// 批量风控
POST /api/v1/admin/micro/order/batch-risk
Request: {
    "order_ids": [1, 2, 3],
    "result": "lose"  // win/lose
}
```

---

### 2.4 账户流水查询

**目标文件:** `admin-go/api/v1/exchange/account_log.go`

```go
// 流水类型定义
const (
    ADMIN_LEGAL_BALANCE       = 1   // 后台调节法币账户余额
    ADMIN_LOCK_LEGAL_BALANCE  = 2   // 后台调节法币账户锁定余额
    ADMIN_CHANGE_BALANCE      = 3   // 后台调节币币账户余额
    ADMIN_LOCK_CHANGE_BALANCE = 4   // 后台调节币币账户锁定余额
    // ... 更多类型
)

// 流水列表
POST /api/v1/admin/account/log/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "type": 0,
    "currency_id": 0,
    "start_time": "",
    "end_time": ""
}

// 流水详情
GET /api/v1/admin/account/log/:id

// 盈亏统计
GET /api/v1/admin/account/profits
Response: {
    "total_profit": 10000,
    "total_loss": 5000,
    "net_profit": 5000
}
```

---

## 阶段三：法币/C2C模块 (P1)

### 3.1 法币交易管理

**目标文件:** `admin-go/api/v1/exchange/legal.go`

```go
// 法币订单列表
POST /api/v1/admin/legal/order/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "seller_number": "",
    "type": "",  // buy/sell
    "status": 0,
    "currency_id": 0
}

// 订单统计
GET /api/v1/admin/legal/statistics
Response: {
    "today_buy_usdt": 10000,
    "today_sell_usdt": 8000,
    "total_buy_usdt": 100000,
    "total_sell_usdt": 80000,
    "lock_balance": 5000,
    "available_balance": 95000
}

// 管理员取消订单
POST /api/v1/admin/legal/cancel
Request: {
    "deal_id": 1
}

// 管理员确认付款
POST /api/v1/admin/legal/confirm-pay
Request: {
    "deal_id": 1
}

// 管理员确认收款（完成交易）
POST /api/v1/admin/legal/confirm-receive
Request: {
    "deal_id": 1
}

// 法币发布列表
GET /api/v1/admin/legal/send/list
```

**模型定义:** `admin-go/model/legal.go`

```go
// LegalDeal 法币交易订单
type LegalDeal struct {
    ID              uint    `gorm:"primaryKey" json:"id"`
    UserID          uint    `json:"user_id"`
    SellerID        uint    `json:"seller_id"`
    LegalDealSendID uint    `json:"legal_deal_send_id"`
    Number          float64 `json:"number"`        // 交易数量
    Price           float64 `json:"price"`         // 单价
    TotalPrice      float64 `json:"total_price"`   // 总价
    IsPay           int8    `json:"is_pay"`        // 是否已付款
    IsSure          int8    `json:"is_sure"`       // 是否已确认
    Status          int8    `json:"status"`        // 状态
    CreateTime      int64   `json:"create_time"`
    UpdateTime      int64   `json:"update_time"`
    
    // 关联
    User          User          `gorm:"foreignKey:UserID" json:"user"`
    Seller        User          `gorm:"foreignKey:SellerID" json:"seller"`
    LegalDealSend LegalDealSend `gorm:"foreignKey:LegalDealSendID" json:"legal_deal_send"`
}

// LegalDealSend 法币发布
type LegalDealSend struct {
    ID         uint    `gorm:"primaryKey" json:"id"`
    UserID     uint    `json:"user_id"`
    CurrencyID uint    `json:"currency_id"`
    Type       string  `json:"type"`        // buy/sell
    WayType    int8    `json:"way_type"`    // 支付方式
    Price      float64 `json:"price"`       // 单价
    TotalNum   float64 `json:"total_num"`   // 总数量
    SurplusNum float64 `json:"surplus_num"` // 剩余数量
    MinLimit   float64 `json:"min_limit"`   // 最小限额
    MaxLimit   float64 `json:"max_limit"`   // 最大限额
    IsShelves  int8    `json:"is_shelves"`  // 是否上架
    CreateTime int64   `json:"create_time"`
}
```

---

### 3.2 C2C交易管理

**目标文件:** `admin-go/api/v1/exchange/c2c.go`

```go
// C2C订单列表
POST /api/v1/admin/c2c/order/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "seller_number": "",
    "type": ""
}

// 统计数据
GET /api/v1/admin/c2c/statistics
Response: {
    "buy_total": 50000,
    "sell_total": 48000
}

// 撤回发布
POST /api/v1/admin/c2c/send/back
Request: {
    "send_id": 1
}

// 删除发布
DELETE /api/v1/admin/c2c/send/:id

// 导出C2C交易
GET /api/v1/admin/c2c/export
```

---

### 3.3 商家管理

**目标文件:** `admin-go/api/v1/exchange/seller.go`

```go
// 商家列表
POST /api/v1/admin/seller/list
Request: {
    "page": 1,
    "page_size": 10,
    "account_number": "",
    "status": 0
}

// 添加商家
POST /api/v1/admin/seller/create
Request: {
    "user_id": 1,
    "status": 1
}

// 审核商家
POST /api/v1/admin/seller/review
Request: {
    "id": 1,
    "status": 1,  // 1-通过, 2-拒绝
    "reason": ""
}

// 删除商家
DELETE /api/v1/admin/seller/:id

// 商家发布管理
GET /api/v1/admin/seller/:id/sends

// 撤回发布
POST /api/v1/admin/seller/send/back
Request: {
    "send_id": 1
}

// 下架发布
POST /api/v1/admin/seller/send/shelves
Request: {
    "send_id": 1,
    "is_shelves": 0
}
```

---

## 阶段四：扩展功能 (P2)

### 4.1 邀请返佣管理

```go
// api/v1/exchange/invite.go

// 返佣记录列表
POST /api/v1/admin/invite/return/list

// 邀请关系图
GET /api/v1/admin/invite/tree/:user_id

// 分享设置
GET /api/v1/admin/invite/share/setting
POST /api/v1/admin/invite/share/setting

// 编辑返佣比例
PUT /api/v1/admin/invite/return/:id
```

---

### 4.2 质押/锁仓管理

```go
// api/v1/exchange/zhiya.go

// 质押产品列表
GET /api/v1/admin/zhiya/list

// 添加质押产品
POST /api/v1/admin/zhiya/create
Request: {
    "currency_id": 3,
    "min_amount": 100,
    "max_amount": 10000,
    "rate": 0.05,  // 日利率
    "days": 30,    // 锁仓天数
    "status": 1
}

// 编辑质押产品
PUT /api/v1/admin/zhiya/:id

// 删除质押产品
DELETE /api/v1/admin/zhiya/:id

// 质押订单列表
POST /api/v1/admin/zhiya/order/list

// 取消质押订单
POST /api/v1/admin/zhiya/order/cancel
```

---

### 4.3 闪兑管理

```go
// api/v1/exchange/flash.go

// 闪兑订单列表
POST /api/v1/admin/flash/list
Request: {
    "page": 1,
    "page_size": 10,
    "status": 0  // 0-待审核
}

// 审核通过
POST /api/v1/admin/flash/approve
Request: {
    "id": 1
}

// 审核拒绝
POST /api/v1/admin/flash/reject
Request: {
    "id": 1,
    "reason": ""
}
```

---

### 4.4 机器人管理

```go
// api/v1/exchange/robot.go

// 机器人列表
GET /api/v1/admin/robot/list

// 添加机器人
POST /api/v1/admin/robot/create
Request: {
    "match_id": 1,
    "min_amount": 0.01,
    "max_amount": 1,
    "frequency": 10,  // 频率(秒)
    "status": 1
}

// 开启/关闭机器人
POST /api/v1/admin/robot/toggle
Request: {
    "id": 1,
    "status": 1
}

// 删除机器人
DELETE /api/v1/admin/robot/:id

// 机器人计划任务
GET /api/v1/admin/robot/schedule/list
POST /api/v1/admin/robot/schedule/create
DELETE /api/v1/admin/robot/schedule/:id
```

---

### 4.5 保险管理

```go
// api/v1/exchange/insurance.go

// 险种列表
GET /api/v1/admin/insurance/type/list

// 添加险种
POST /api/v1/admin/insurance/type/create

// 保险规则列表
GET /api/v1/admin/insurance/rule/list

// 添加保险规则
POST /api/v1/admin/insurance/rule/create

// 理赔申请列表
POST /api/v1/admin/insurance/claim/list

// 理赔审核
POST /api/v1/admin/insurance/claim/review
Request: {
    "id": 1,
    "status": 1,  // 1-通过, 2-拒绝
    "amount": 100
}

// 用户保险订单
POST /api/v1/admin/insurance/order/list
```

---

## 路由配置示例

**文件:** `admin-go/router/admin_router.go`

```go
package router

import (
    v1 "admin-go/api/v1"
    "admin-go/middleware"
    "github.com/gin-gonic/gin"
)

func InitAdminRouter(Router *gin.RouterGroup) {
    // 无需认证的路由
    publicRouter := Router.Group("admin")
    {
        publicRouter.POST("login", v1.Login)
    }
    
    // 需要认证的路由
    privateRouter := Router.Group("admin")
    privateRouter.Use(middleware.JWTAuth(), middleware.CasbinHandler())
    {
        // 用户管理
        userRouter := privateRouter.Group("user")
        {
            userRouter.POST("list", v1.GetUserList)
            userRouter.GET(":id", v1.GetUserDetail)
            userRouter.PUT(":id", v1.UpdateUser)
            userRouter.DELETE(":id", v1.DeleteUser)
            userRouter.POST("freeze", v1.FreezeUser)
            userRouter.POST("activate", v1.ActivateUser)
            userRouter.POST("adjust-balance", v1.AdjustBalance)
            userRouter.POST("batch-risk", v1.BatchRisk)
            userRouter.GET("export", v1.ExportUsers)
        }
        
        // 提币管理
        withdrawalRouter := privateRouter.Group("withdrawal")
        {
            withdrawalRouter.POST("list", v1.GetWithdrawalList)
            withdrawalRouter.GET(":id", v1.GetWithdrawalDetail)
            withdrawalRouter.POST("approve", v1.ApproveWithdrawal)
            withdrawalRouter.POST("reject", v1.RejectWithdrawal)
        }
        
        // KYC管理
        kycRouter := privateRouter.Group("kyc")
        {
            kycRouter.GET("list", v1.GetKYCList)
            kycRouter.GET(":id", v1.GetKYCDetail)
            kycRouter.POST("approve", v1.ApproveKYC)
            kycRouter.POST("reject", v1.RejectKYC)
            kycRouter.DELETE(":id", v1.DeleteKYC)
        }
        
        // 币种管理
        currencyRouter := privateRouter.Group("currency")
        {
            currencyRouter.POST("list", v1.GetCurrencyList)
            currencyRouter.POST("create", v1.CreateCurrency)
            currencyRouter.PUT(":id", v1.UpdateCurrency)
            currencyRouter.DELETE(":id", v1.DeleteCurrency)
            currencyRouter.POST("toggle-display", v1.ToggleCurrencyDisplay)
        }
        
        // 交易管理
        transactionRouter := privateRouter.Group("transaction")
        {
            transactionRouter.POST("spot/list", v1.GetSpotList)
            transactionRouter.POST("in/list", v1.GetInList)
            transactionRouter.POST("out/list", v1.GetOutList)
            transactionRouter.POST("complete/list", v1.GetCompleteList)
        }
        
        // 杠杆管理
        leverRouter := privateRouter.Group("lever")
        {
            leverRouter.POST("list", v1.GetLeverList)
            leverRouter.GET("multiple/list", v1.GetMultipleList)
            leverRouter.POST("multiple/create", v1.CreateMultiple)
            leverRouter.DELETE("multiple/:id", v1.DeleteMultiple)
            leverRouter.POST("hazard/list", v1.GetHazardList)
            leverRouter.POST("hazard/handle", v1.HandleHazard)
        }
        
        // 秒合约管理
        microRouter := privateRouter.Group("micro")
        {
            microRouter.GET("number/list", v1.GetMicroNumberList)
            microRouter.POST("number/create", v1.CreateMicroNumber)
            microRouter.DELETE("number/:id", v1.DeleteMicroNumber)
            microRouter.GET("seconds/list", v1.GetMicroSecondsList)
            microRouter.POST("seconds/create", v1.CreateMicroSeconds)
            microRouter.POST("seconds/status", v1.ToggleMicroSecondsStatus)
            microRouter.DELETE("seconds/:id", v1.DeleteMicroSeconds)
            microRouter.POST("order/list", v1.GetMicroOrderList)
            microRouter.PUT("order/:id", v1.UpdateMicroOrder)
            microRouter.POST("order/batch-risk", v1.BatchMicroRisk)
        }
        
        // 法币管理
        legalRouter := privateRouter.Group("legal")
        {
            legalRouter.POST("order/list", v1.GetLegalOrderList)
            legalRouter.GET("statistics", v1.GetLegalStatistics)
            legalRouter.POST("cancel", v1.CancelLegalOrder)
            legalRouter.POST("confirm-pay", v1.ConfirmLegalPay)
            legalRouter.POST("confirm-receive", v1.ConfirmLegalReceive)
        }
        
        // C2C管理
        c2cRouter := privateRouter.Group("c2c")
        {
            c2cRouter.POST("order/list", v1.GetC2COrderList)
            c2cRouter.GET("statistics", v1.GetC2CStatistics)
            c2cRouter.POST("send/back", v1.BackC2CSend)
            c2cRouter.DELETE("send/:id", v1.DeleteC2CSend)
        }
        
        // 商家管理
        sellerRouter := privateRouter.Group("seller")
        {
            sellerRouter.POST("list", v1.GetSellerList)
            sellerRouter.POST("create", v1.CreateSeller)
            sellerRouter.POST("review", v1.ReviewSeller)
            sellerRouter.DELETE(":id", v1.DeleteSeller)
        }
        
        // 系统设置
        settingRouter := privateRouter.Group("setting")
        {
            settingRouter.GET("list", v1.GetSettingList)
            settingRouter.POST("update", v1.UpdateSettings)
            settingRouter.GET("base", v1.GetBaseSetting)
            settingRouter.POST("base", v1.SaveBaseSetting)
        }
        
        // 角色权限
        roleRouter := privateRouter.Group("role")
        {
            roleRouter.GET("list", v1.GetRoleList)
            roleRouter.POST("create", v1.CreateRole)
            roleRouter.PUT(":id", v1.UpdateRole)
            roleRouter.DELETE(":id", v1.DeleteRole)
            roleRouter.POST("assign-permission", v1.AssignPermission)
        }
        
        // 账户流水
        accountRouter := privateRouter.Group("account")
        {
            accountRouter.POST("log/list", v1.GetAccountLogList)
            accountRouter.GET("log/:id", v1.GetAccountLogDetail)
            accountRouter.GET("profits", v1.GetProfits)
        }
        
        // 统计报表
        statisticsRouter := privateRouter.Group("statistics")
        {
            statisticsRouter.GET("dashboard", v1.GetDashboard)
            statisticsRouter.GET("user", v1.GetUserStatistics)
            statisticsRouter.GET("trade", v1.GetTradeStatistics)
            statisticsRouter.GET("finance", v1.GetFinanceStatistics)
        }
    }
}
```

---

## 测试清单

### 单元测试

```go
// service/user_service_test.go
func TestAdjustBalance(t *testing.T) {
    // 测试加款
    err := userSrv.AdjustBalance(1, 3, "legal_balance", 100, "测试加款")
    assert.NoError(t, err)
    
    // 测试扣款
    err = userSrv.AdjustBalance(1, 3, "legal_balance", -50, "测试扣款")
    assert.NoError(t, err)
    
    // 测试余额不足
    err = userSrv.AdjustBalance(1, 3, "legal_balance", -99999, "测试余额不足")
    assert.Error(t, err)
}
```

### API测试

```bash
# 登录获取token
curl -X POST http://localhost:8000/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'

# 获取用户列表
curl -X POST http://localhost:8000/api/v1/admin/user/list \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"page":1,"page_size":10}'

# 余额调节
curl -X POST http://localhost:8000/api/v1/admin/user/adjust-balance \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"currency_id":3,"balance_type":"legal_balance","amount":100,"reason":"测试加款"}'
```

---

**文档版本**: v1.0  
**创建时间**: 2026-01-24  
