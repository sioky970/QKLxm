# 用户 USDT 余额调整功能实现说明

## 功能概述

在管理后台用户列表页面 (`/admin/user/list`) 新增了 USDT 余额调整功能，允许管理员对用户的各个 USDT 账户进行充值或扣除操作，每次操作都会自动生成账户变动日志。

## 一、前端实现

### 1.1 入口位置

在用户列表页面的操作列下拉菜单中，新增了"调整余额"选项：

```
操作列 → 更多 → 调整余额
```

### 1.2 功能界面

余额调整弹窗包含以下功能：

1. **用户信息显示**
   - 用户账号
   - 用户ID

2. **当前余额展示**
   - 法币账户余额
   - 币币账户余额
   - 合约账户余额
   - 秒合约账户余额

3. **调整参数**
   - 余额类型选择（法币/币币/合约/秒合约）
   - 操作类型选择（充值/扣除）
   - 调整金额输入（精确到8位小数）
   - 操作原因说明（必填，记录到日志）

4. **安全提醒**
   - 显示操作影响说明
   - 二次确认弹窗

### 1.3 文件位置

```
4-管理后台前端/arco-design-pro-vite/src/views/admin/user/list.vue
```

### 1.4 关键代码

**API 调用：**
```typescript
await axios.post('/admin/user/adjust-balance', {
  user_id: adjustBalanceForm.user_id,
  currency_id: 3, // USDT 的 currency_id
  balance_type: adjustBalanceForm.balance_type,
  amount: actualAmount, // 充值为正数，扣除为负数
  reason: adjustBalanceForm.reason.trim(),
});
```

**余额类型：**
- `legal_balance` - 法币账户
- `change_balance` - 币币账户
- `lever_balance` - 合约账户
- `micro_balance` - 秒合约账户

## 二、后端实现

### 2.1 API 路由

```
POST /admin/user/adjust-balance
```

**路由定义位置：**
```
1-Go交易后端/internal/api/router/admin_router.go (第37行)
```

### 2.2 请求参数

```json
{
  "user_id": 123,           // 用户ID（必填）
  "currency_id": 3,         // 币种ID（USDT=3，必填）
  "balance_type": "legal_balance",  // 余额类型（必填）
  "amount": 100.50,         // 调整金额，正数为充值，负数为扣除（必填）
  "reason": "管理员充值"     // 操作原因（选填）
}
```

### 2.3 核心服务

**文件位置：**
```
1-Go交易后端/internal/service/admin/user_admin_service.go (第112-167行)
```

**核心功能：**
```go
func (s *UserAdminService) AdjustBalance(
    userID uint, 
    currencyID uint, 
    balanceType string, 
    amount float64, 
    reason string
) error
```

**实现特点：**
1. ✅ 使用数据库事务保证原子性
2. ✅ 扣除时检查余额是否充足
3. ✅ 自动生成账户变动日志
4. ✅ 支持四种账户类型调整

### 2.4 账户日志类型

系统自动根据余额类型和操作方向生成对应的日志类型：

| 余额类型 | 充值（增加） | 扣除（减少） |
|---------|------------|------------|
| legal_balance | 类型1 | 类型2 |
| change_balance | 类型3 | 类型4 |
| lever_balance | 类型5 | 类型6 |
| micro_balance | 类型7 | 类型8 |

## 三、交易系统余额使用分析

### 3.1 现货交易（币币交易）

**使用账户：**
- 买入：扣除 `legal_balance`（法币账户）
- 卖出：扣除 `change_balance`（币币账户）

**余额处理流程：**
1. 下单时锁定对应余额（转入冻结余额）
2. 成交后从冻结余额中扣除
3. 取消订单时解冻余额

**实现位置：**
```
1-Go交易后端/internal/service/spot_trading_service.go (第121-204行)
```

### 3.2 秒合约交易

**使用账户：**
- `legal_balance`（法币账户）

**余额处理流程：**
1. 下单时扣除本金 + 手续费
2. 资金转入 `lock_legal_balance`（冻结法币余额）
3. 结算时根据盈亏返还或扣除

**实现位置：**
```
1-Go交易后端/internal/service/micro_trading_service.go (第107-201行)
```

**关键代码：**
```go
// 锁定投资金额：从法币余额扣除并转入锁定余额
totalLock := req.Amount + fee
tx.Model(&model.UsersWallet{}).
    Where("user_id = ? AND currency = ? AND legal_balance >= ?", userID, req.LegalID, totalLock).
    Updates(map[string]interface{}{
        "legal_balance":      gorm.Expr("legal_balance - ?", totalLock),
        "lock_legal_balance": gorm.Expr("lock_legal_balance + ?", totalLock),
    })
```

### 3.3 合约交易（杠杆交易）

**使用账户：**
- `lever_balance`（合约账户）

**余额处理流程：**
1. 开仓时扣除保证金
2. 持仓期间计算未实现盈亏
3. 平仓时根据实际盈亏调整余额
4. 爆仓时扣除全部保证金

**实现位置：**
```
1-Go交易后端/internal/service/futures_trading_service.go
```

**注意：** Go 后端的合约交易功能尚未完全实现，主要业务逻辑在 PHP 后端。

## 四、安全机制

### 4.1 事务保护

所有余额变动操作都使用数据库事务：
```go
return database.DB.Transaction(func(tx *gorm.DB) error {
    // 1. 查找钱包
    // 2. 更新余额
    // 3. 记录日志
    return tx.Create(&accountLog).Error
})
```

### 4.2 余额检查

扣除操作前检查余额是否充足：
```go
if wallet.LegalBalance+amount < 0 {
    return errors.New("法币余额不足")
}
```

### 4.3 并发控制

现货交易使用乐观锁防止超卖：
```go
Where("id = ? AND change_balance >= ?", wallet.ID, req.Quantity)
```

### 4.4 防止负余额

合约爆仓时防止余额为负：
```go
newBalance := wallet.LeverBalance - position.Caution + finalPnL
if newBalance < 0 {
    newBalance = 0 // 防止负余额
}
```

## 五、使用说明

### 5.1 充值操作示例

1. 进入管理后台 → 用户管理 → 用户列表
2. 找到目标用户，点击操作列的"更多" → "调整余额"
3. 选择余额类型（例如：法币账户）
4. 选择操作类型："充值"
5. 输入金额：100
6. 输入原因："用户申诉补偿"
7. 点击确定 → 确认弹窗 → 提交

**结果：**
- 用户法币账户增加 100 USDT
- 生成类型1的账户变动日志
- 日志信息记录："用户申诉补偿"

### 5.2 扣除操作示例

1. 进入管理后台 → 用户管理 → 用户列表
2. 找到目标用户，点击操作列的"更多" → "调整余额"
3. 选择余额类型（例如：币币账户）
4. 选择操作类型："扣除"
5. 输入金额：50
6. 输入原因："违规交易罚款"
7. 点击确定 → 确认弹窗 → 提交

**结果：**
- 用户币币账户扣除 50 USDT
- 生成类型4的账户变动日志
- 日志信息记录："违规交易罚款"

## 六、账户日志查询

所有余额变动都会记录到 `account_log` 表，可通过以下方式查询：

### 6.1 数据库查询

```sql
SELECT * FROM account_log 
WHERE user_id = 123 
  AND currency = 3 
ORDER BY created_time DESC;
```

### 6.2 管理后台查询

```
管理后台 → 交易管理 → 账户流水
```

可根据用户、币种、日志类型等条件筛选查询。

## 七、注意事项

### 7.1 USDT 币种 ID

本系统中 USDT 的 `currency_id` 为 **3**，这在以下位置定义：

```sql
5-数据库脚本/数据库.sql (第840行)
INSERT INTO `currency` VALUES (3,'USDT',...);
```

### 7.2 余额类型说明

| 账户类型 | 字段名 | 用途 |
|---------|-------|------|
| 法币账户 | legal_balance | 充值提现、秒合约交易 |
| 币币账户 | change_balance | 现货交易 |
| 合约账户 | lever_balance | 杠杆合约交易 |
| 秒合约账户 | micro_balance | 秒合约专用（暂未使用） |

### 7.3 操作权限

此功能仅限管理员使用，需要：
1. 登录管理后台
2. 具有用户管理权限
3. 通过后台 API 鉴权

### 7.4 数据一致性

- ✅ 所有操作都有事务保护
- ✅ 扣除操作会检查余额充足性
- ✅ 每次操作都生成账户日志
- ✅ 支持审计和回溯

## 八、测试建议

### 8.1 功能测试

1. **充值测试**
   - 对各种账户类型进行充值
   - 验证余额增加正确
   - 验证日志生成正确

2. **扣除测试**
   - 余额充足时扣除
   - 余额不足时扣除（应报错）
   - 验证日志生成正确

3. **边界测试**
   - 扣除全部余额
   - 输入极小金额（0.00000001）
   - 输入大金额

### 8.2 安全测试

1. 并发操作测试
2. 事务回滚测试
3. 权限验证测试

### 8.3 日志验证

1. 检查 account_log 表记录
2. 验证日志类型正确
3. 验证操作原因记录完整

## 九、系统架构总结

```
前端（Vue 3 + Arco Design）
    ↓
    调用 API: POST /admin/user/adjust-balance
    ↓
Go 后端 API Handler (admin/user_handler.go)
    ↓
UserAdminService.AdjustBalance()
    ↓
数据库事务执行：
    1. 查找钱包记录（users_wallet 表）
    2. 更新余额字段
    3. 创建账户日志（account_log 表）
    ↓
返回成功/失败结果
    ↓
前端显示提示信息并刷新列表
```

## 十、相关文件清单

| 文件路径 | 说明 |
|---------|------|
| `4-管理后台前端/arco-design-pro-vite/src/views/admin/user/list.vue` | 前端用户列表页面 |
| `1-Go交易后端/internal/api/router/admin_router.go` | API 路由定义 |
| `1-Go交易后端/internal/api/handler/admin/user_handler.go` | API 处理器 |
| `1-Go交易后端/internal/service/admin/user_admin_service.go` | 余额调整核心服务 |
| `1-Go交易后端/internal/service/spot_trading_service.go` | 现货交易服务 |
| `1-Go交易后端/internal/service/micro_trading_service.go` | 秒合约交易服务 |
| `1-Go交易后端/internal/service/futures_trading_service.go` | 合约交易服务 |
| `1-Go交易后端/internal/model/model.go` | 数据模型定义 |

---

**实现完成时间：** 2026年1月25日  
**实现人员：** AI 助手  
**版本号：** 1.0
