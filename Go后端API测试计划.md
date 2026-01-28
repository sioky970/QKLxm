# Go后端API测试计划

## 文档信息
- **项目名称**: 混合架构交易所系统
- **测试目标**: Go交易后端API
- **创建日期**: 2026-01-24
- **版本**: v1.0

---

## 一、测试范围定义

### 1.1 测试目标
- Go后端API功能正确性验证
- 接口权限控制验证
- 数据完整性和一致性验证
- 异常处理能力验证

### 1.2 API统计
| 分类 | API数量 |
|------|---------|
| 用户端公开接口 | 8个 |
| 用户端登录接口 | 14个 |
| 管理端公开接口 | 1个 |
| 管理端权限接口 | 104个 |
| **总计** | **127个** |

### 1.3 排除范围
以下功能不在测试范围内：
- 邀请返佣管理
- 闪兑管理
- 机器人管理
- 保险管理
- 质押/锁仓管理

---

## 二、测试环境配置

### 2.1 数据库配置
```yaml
数据库类型: MySQL 5.7+
测试数据库: exchange_test
字符集: utf8mb4
```

### 2.2 测试账号
| 角色 | 账号 | 密码 | 用途 |
|------|------|------|------|
| 超级管理员 | admin | admin123 | 全功能测试 |
| 普通管理员 | manager | manager123 | 权限限制测试 |
| 测试用户1 | user001 | user123 | 普通用户功能测试 |
| 测试用户2 | user002 | user123 | 交易对手测试 |
| 风控用户 | riskuser | user123 | 风控场景测试 |

### 2.3 基础测试数据
- 币种: USDT, BTC, ETH
- 交易对: BTC/USDT, ETH/USDT
- 秒合约配置: 30秒、60秒、180秒
- 杠杆倍数: 10x, 20x, 50x, 100x

---

## 三、用户端API测试用例

### 3.1 用户认证模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-AUTH-001 | 用户注册-正常 | `/api/user/register` | POST | 返回成功，创建用户 |
| U-AUTH-002 | 用户注册-重复账号 | `/api/user/register` | POST | 返回错误"账号已存在" |
| U-AUTH-003 | 用户注册-验证码错误 | `/api/user/register` | POST | 返回错误"验证码错误" |
| U-AUTH-004 | 用户登录-正常 | `/api/user/login` | POST | 返回token和用户信息 |
| U-AUTH-005 | 用户登录-密码错误 | `/api/user/login` | POST | 返回错误"密码错误" |
| U-AUTH-006 | 用户登录-账号冻结 | `/api/user/login` | POST | 返回错误"账号已被冻结" |
| U-AUTH-007 | 发送验证码 | `/api/user/send_code` | POST | 返回成功 |
| U-AUTH-008 | 重置密码 | `/api/user/reset_password` | POST | 密码更新成功 |

### 3.2 用户信息模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-INFO-001 | 获取用户信息 | `/api/user/info` | GET | 返回用户详细信息 |
| U-INFO-002 | 无token访问 | `/api/user/info` | GET | 返回401未授权 |
| U-INFO-003 | token过期 | `/api/user/info` | GET | 返回401需重新登录 |
| U-INFO-004 | 更新用户信息 | `/api/user/update` | POST | 更新成功 |
| U-INFO-005 | 修改登录密码 | `/api/user/change_password` | POST | 密码修改成功 |
| U-INFO-006 | 修改支付密码 | `/api/user/change_pay_password` | POST | 设置成功 |

### 3.3 钱包模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-WAL-001 | 获取钱包列表 | `/api/wallet/list` | GET | 返回钱包列表 |
| U-WAL-002 | 获取钱包详情 | `/api/wallet/info` | GET | 返回余额详情 |
| U-WAL-003 | 充值申请 | `/api/wallet/recharge` | POST | 创建充值订单 |
| U-WAL-004 | 提币申请-正常 | `/api/wallet/withdraw` | POST | 创建提币订单 |
| U-WAL-005 | 提币申请-余额不足 | `/api/wallet/withdraw` | POST | 返回错误"余额不足" |
| U-WAL-006 | 提币申请-未实名 | `/api/wallet/withdraw` | POST | 返回错误 |
| U-WAL-007 | 钱包流水查询 | `/api/wallet/logs` | GET | 返回流水列表 |

### 3.4 币币交易模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-SPOT-001 | 提交买单-限价 | `/api/transaction/submit` | POST | 创建限价买单 |
| U-SPOT-002 | 提交卖单-限价 | `/api/transaction/submit` | POST | 创建限价卖单 |
| U-SPOT-003 | 提交买单-市价 | `/api/transaction/submit` | POST | 市价成交 |
| U-SPOT-004 | 余额不足下单 | `/api/transaction/submit` | POST | 返回错误 |
| U-SPOT-005 | 取消订单 | `/api/transaction/cancel` | POST | 订单取消 |
| U-SPOT-006 | 取消已成交订单 | `/api/transaction/cancel` | POST | 返回错误 |
| U-SPOT-007 | 当前委托查询 | `/api/transaction/list` | GET | 返回未成交订单 |
| U-SPOT-008 | 历史订单查询 | `/api/transaction/history` | GET | 返回历史订单 |

### 3.5 杠杆交易模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-LEV-001 | 开仓-做多 | `/api/lever/submit` | POST | 创建多单持仓 |
| U-LEV-002 | 开仓-做空 | `/api/lever/submit` | POST | 创建空单持仓 |
| U-LEV-003 | 保证金不足 | `/api/lever/submit` | POST | 返回错误 |
| U-LEV-004 | 平仓 | `/api/lever/close` | POST | 计算盈亏结算 |
| U-LEV-005 | 持仓列表 | `/api/lever/position` | GET | 返回持仓列表 |
| U-LEV-006 | 历史记录 | `/api/lever/history` | GET | 返回历史记录 |

### 3.6 秒合约模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-MIC-001 | 下单-看涨 | `/api/micro/submit` | POST | 创建看涨订单 |
| U-MIC-002 | 下单-看跌 | `/api/micro/submit` | POST | 创建看跌订单 |
| U-MIC-003 | 余额不足 | `/api/micro/submit` | POST | 返回错误 |
| U-MIC-004 | 订单列表 | `/api/micro/list` | GET | 返回订单列表 |

### 3.7 KYC实名认证模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-KYC-001 | 提交认证 | `/api/kyc/submit` | POST | 创建认证申请 |
| U-KYC-002 | 重复提交 | `/api/kyc/submit` | POST | 返回错误 |
| U-KYC-003 | 身份证格式错误 | `/api/kyc/submit` | POST | 返回错误 |
| U-KYC-004 | 查询认证状态 | `/api/kyc/status` | GET | 返回认证状态 |

### 3.8 行情模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| U-MKT-001 | 最新行情 | `/api/quotation/new` | GET | 返回所有币种行情 |
| U-MKT-002 | 市场数据 | `/api/market/market` | POST | 返回交易对详情 |
| U-MKT-003 | K线数据 | `/api/kline` | GET | 返回K线数据 |

---

## 四、管理端API测试用例

### 4.1 管理员认证模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-AUTH-001 | 管理员登录-正常 | `/api/admin/login` | POST | 返回token |
| A-AUTH-002 | 管理员登录-密码错误 | `/api/admin/login` | POST | 返回错误 |
| A-AUTH-003 | 管理员登录-账号禁用 | `/api/admin/login` | POST | 返回错误 |
| A-AUTH-004 | 获取管理员信息 | `/api/admin/info` | GET | 返回管理员信息 |
| A-AUTH-005 | 退出登录 | `/api/admin/logout` | POST | 退出成功 |
| A-AUTH-006 | 修改密码 | `/api/admin/change-password` | POST | 修改成功 |

### 4.2 用户管理模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-USR-001 | 用户列表 | `/api/admin/user/list` | POST | 返回用户列表 |
| A-USR-002 | 用户列表-筛选 | `/api/admin/user/list` | POST | 返回筛选结果 |
| A-USR-003 | 用户详情 | `/api/admin/user/:id` | GET | 返回用户详情 |
| A-USR-004 | 编辑用户 | `/api/admin/user/:id` | PUT | 更新成功 |
| A-USR-005 | 冻结用户 | `/api/admin/user/freeze` | POST | 状态变为冻结 |
| A-USR-006 | 解冻用户 | `/api/admin/user/activate` | POST | 状态恢复正常 |
| A-USR-007 | 重置密码 | `/api/admin/user/reset-password` | POST | 密码重置成功 |
| A-USR-008 | 调节余额-增加 | `/api/admin/user/adjust-balance` | POST | 余额增加 |
| A-USR-009 | 调节余额-扣减 | `/api/admin/user/adjust-balance` | POST | 余额扣减 |
| A-USR-010 | 批量风控设置 | `/api/admin/user/batch-risk` | POST | 批量更新成功 |
| A-USR-011 | 用户钱包查询 | `/api/admin/user/:id/wallets` | GET | 返回用户钱包 |
| A-USR-012 | 用户搜索 | `/api/admin/user/search` | POST | 返回匹配用户 |
| A-USR-013 | 用户导出 | `/api/admin/user/export` | GET | 下载CSV文件 |

### 4.3 提币审核模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-WIT-001 | 提币列表 | `/api/admin/wallet/withdrawals` | POST | 返回提币列表 |
| A-WIT-002 | 提币列表-状态筛选 | `/api/admin/wallet/withdrawals` | POST | 返回筛选列表 |
| A-WIT-003 | 提币详情 | `/api/admin/wallet/withdrawals/:id` | GET | 返回详情 |
| A-WIT-004 | 提币审核通过 | `/api/admin/wallet/withdrawals/approve` | POST | 状态变为通过 |
| A-WIT-005 | 提币审核拒绝 | `/api/admin/wallet/withdrawals/reject` | POST | 状态变为拒绝 |
| A-WIT-006 | 重复审核 | `/api/admin/wallet/withdrawals/approve` | POST | 返回错误 |

### 4.4 充值审核模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-CHG-001 | 充值列表 | `/api/admin/wallet/charge/list` | POST | 返回充值列表 |
| A-CHG-002 | 充值审核通过 | `/api/admin/wallet/charge/approve` | POST | 用户余额增加 |
| A-CHG-003 | 充值审核拒绝 | `/api/admin/wallet/charge/reject` | POST | 状态变为拒绝 |

### 4.5 KYC审核模块 (P0)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-KYC-001 | KYC列表 | `/api/admin/kyc/list` | POST | 返回认证列表 |
| A-KYC-002 | KYC列表-状态筛选 | `/api/admin/kyc/list` | POST | 返回筛选列表 |
| A-KYC-003 | KYC详情 | `/api/admin/kyc/:id` | GET | 返回认证详情 |
| A-KYC-004 | KYC审核通过 | `/api/admin/kyc/approve` | POST | 状态变为通过 |
| A-KYC-005 | KYC审核拒绝 | `/api/admin/kyc/reject` | POST | 状态变为拒绝 |
| A-KYC-006 | 删除KYC记录 | `/api/admin/kyc/:id` | DELETE | 记录删除 |

### 4.6 秒合约管理模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-MIC-001 | 金额配置列表 | `/api/admin/micro/number/list` | GET | 返回金额配置 |
| A-MIC-002 | 创建金额配置 | `/api/admin/micro/number/create` | POST | 创建成功 |
| A-MIC-003 | 删除金额配置 | `/api/admin/micro/number/:id` | DELETE | 删除成功 |
| A-MIC-004 | 时间配置列表 | `/api/admin/micro/seconds/list` | GET | 返回时间配置 |
| A-MIC-005 | 创建时间配置 | `/api/admin/micro/seconds/create` | POST | 创建成功 |
| A-MIC-006 | 更新时间配置 | `/api/admin/micro/seconds/:id` | PUT | 更新成功 |
| A-MIC-007 | 秒合约订单列表 | `/api/admin/micro/order/list` | POST | 返回订单列表 |
| A-MIC-008 | 秒合约订单详情 | `/api/admin/micro/order/:id` | GET | 返回订单详情 |
| A-MIC-009 | 编辑订单风控 | `/api/admin/micro/order/:id` | PUT | 风控设置成功 |
| A-MIC-010 | 批量风控设置 | `/api/admin/micro/order/batch-risk` | POST | 批量设置成功 |

### 4.7 杠杆交易管理模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-LEV-001 | 杠杆订单列表 | `/api/admin/lever/list` | POST | 返回订单列表 |
| A-LEV-002 | 杠杆订单详情 | `/api/admin/lever/:id` | GET | 返回订单详情 |
| A-LEV-003 | 倍数配置列表 | `/api/admin/lever/multiple/list` | GET | 返回倍数配置 |
| A-LEV-004 | 创建倍数配置 | `/api/admin/lever/multiple/create` | POST | 创建成功 |
| A-LEV-005 | 风险率列表 | `/api/admin/lever/hazard/list` | POST | 返回高风险订单 |
| A-LEV-006 | 价格干预 | `/api/admin/lever/hazard/handle` | POST | 干预成功 |
| A-LEV-007 | 强制平仓 | `/api/admin/lever/close` | POST | 平仓成功 |
| A-LEV-008 | 导出订单 | `/api/admin/lever/export` | GET | 下载CSV文件 |

### 4.8 法币/C2C管理模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-LEG-001 | 法币订单列表 | `/api/admin/legal/order/list` | POST | 返回订单列表 |
| A-LEG-002 | 法币统计 | `/api/admin/legal/statistics` | GET | 返回统计数据 |
| A-LEG-003 | 取消法币订单 | `/api/admin/legal/cancel` | POST | 订单取消 |
| A-LEG-004 | 确认付款 | `/api/admin/legal/confirm-pay` | POST | 状态更新 |
| A-LEG-005 | 确认收款 | `/api/admin/legal/confirm-receive` | POST | 交易完成 |
| A-C2C-001 | C2C订单列表 | `/api/admin/c2c/order/list` | POST | 返回订单列表 |
| A-C2C-002 | C2C统计 | `/api/admin/c2c/statistics` | GET | 返回统计数据 |
| A-C2C-003 | 导出C2C订单 | `/api/admin/c2c/export` | GET | 下载CSV文件 |

### 4.9 账户流水模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-LOG-001 | 流水列表 | `/api/admin/account/log/list` | POST | 返回流水列表 |
| A-LOG-002 | 流水详情 | `/api/admin/account/log/:id` | GET | 返回流水详情 |
| A-LOG-003 | 流水类型列表 | `/api/admin/account/log/types` | GET | 返回所有类型 |
| A-LOG-004 | 盈亏统计 | `/api/admin/account/profits` | GET | 返回盈亏统计 |
| A-LOG-005 | 用户流水查询 | `/api/admin/account/user/:user_id/logs` | POST | 返回用户流水 |
| A-LOG-006 | 导出流水 | `/api/admin/account/log/export` | GET | 下载CSV文件 |

### 4.10 统计功能模块 (P2)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-STA-001 | 仪表盘数据 | `/api/admin/statistics/dashboard` | GET | 返回汇总数据 |
| A-STA-002 | 用户统计 | `/api/admin/statistics/user` | GET | 返回用户统计 |
| A-STA-003 | 交易统计 | `/api/admin/statistics/trade` | GET | 返回交易统计 |
| A-STA-004 | 财务统计 | `/api/admin/statistics/finance` | GET | 返回财务统计 |
| A-STA-005 | 交易榜单 | `/api/admin/statistics/top-traders` | GET | 返回Top交易用户 |

### 4.11 风控管理模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-RSK-001 | 风控配置获取 | `/api/admin/risk/config` | GET | 返回风控配置 |
| A-RSK-002 | 风控配置更新 | `/api/admin/risk/config` | POST | 更新成功 |
| A-RSK-003 | 风控用户列表 | `/api/admin/risk/users` | POST | 返回风控用户 |
| A-RSK-004 | 设置用户风控 | `/api/admin/risk/users/set` | POST | 设置成功 |
| A-RSK-005 | 交易对风控列表 | `/api/admin/risk/matches` | GET | 返回交易对风控 |
| A-RSK-006 | 设置交易对风控 | `/api/admin/risk/matches/set` | POST | 设置成功 |

### 4.12 权限管理模块 (P1)

| 用例ID | 测试场景 | API | 方法 | 预期结果 |
|--------|----------|-----|------|----------|
| A-ADM-001 | 管理员列表 | `/api/admin/admin/list` | POST | 返回管理员列表 |
| A-ADM-002 | 创建管理员 | `/api/admin/admin/create` | POST | 创建成功 |
| A-ADM-003 | 更新管理员 | `/api/admin/admin/update` | PUT | 更新成功 |
| A-ADM-004 | 删除管理员 | `/api/admin/admin/:id` | DELETE | 删除成功 |
| A-ROL-001 | 角色列表 | `/api/admin/role/list` | GET | 返回角色列表 |
| A-ROL-002 | 创建角色 | `/api/admin/role/create` | POST | 创建成功 |
| A-ROL-003 | 分配权限 | `/api/admin/role/assign-permissions` | POST | 分配成功 |
| A-ROL-004 | 删除角色 | `/api/admin/role/:id` | DELETE | 删除成功 |

---

## 五、权限验证测试

### 5.1 用户端权限测试

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| P-USR-001 | 未登录访问需登录接口 | 返回401 |
| P-USR-002 | 用户A访问用户B数据 | 返回403或只返回自己数据 |
| P-USR-003 | 冻结用户执行交易 | 返回错误 |

### 5.2 管理端权限测试

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| P-ADM-001 | 用户端token访问管理端 | 返回401 |
| P-ADM-002 | 普通管理员访问超级管理员功能 | 返回403 |
| P-ADM-003 | 无权限管理员删除用户 | 返回403 |
| P-ADM-004 | 超级管理员访问所有功能 | 全部成功 |

---

## 六、异常场景测试

| 用例ID | 测试场景 | 测试数据 | 预期结果 |
|--------|----------|----------|----------|
| E-001 | 缺少必填参数 | 缺少required字段 | 返回400 |
| E-002 | 参数类型错误 | 字符串传给数字字段 | 返回400 |
| E-003 | 参数值越界 | 超大数值 | 返回错误 |
| E-004 | SQL注入测试 | `' OR 1=1 --` | 正确转义 |
| E-005 | XSS测试 | `<script>alert(1)</script>` | 正确转义 |

---

## 七、测试优先级

| 优先级 | 模块 | 用例数 | 说明 |
|--------|------|--------|------|
| P0 | 认证、钱包、交易、审核 | 约60个 | 核心业务流程 |
| P1 | 用户管理、KYC、风控、流水 | 约40个 | 重要管理功能 |
| P2 | 统计、配置、导出 | 约27个 | 辅助功能 |

---

## 八、测试执行顺序

```
阶段1: 环境准备和服务启动验证
  ↓
阶段2: 用户端认证和基础功能测试 (P0)
  ↓
阶段3: 管理端认证和审核功能测试 (P0)
  ↓
阶段4: 交易功能完整流程测试 (P0)
  ↓
阶段5: 管理端其他功能测试 (P1)
  ↓
阶段6: 权限和异常场景测试
  ↓
阶段7: 统计和辅助功能测试 (P2)
  ↓
阶段8: 回归测试和总结
```
