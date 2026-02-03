# 数据库优化回滚方案

## 一、回滚策略概述

本回滚方案针对数据库优化的各个环节制定，确保在任何优化措施导致问题时能够快速恢复到优化前的状态。

### 1.1 适用场景

- 索引添加后导致查询性能下降或死锁
- MySQL配置参数调整后系统不稳定
- 业务代码修改引入新问题
- 数据异常或丢失

### 1.2 回滚原则

1. **先备份后操作**：所有优化操作执行前必须完成数据库备份
2. **分阶段回滚**：支持按模块分次回滚，不影响其他优化措施
3. **快速恢复**：回滚操作应在5分钟内完成
4. **验证确认**：回滚后必须进行功能验证

---

## 二、索引回滚方案

### 2.1 回滚前准备

```sql
-- 创建索引记录表
CREATE TABLE IF NOT EXISTS index_operations_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL,
    index_name VARCHAR(100) NOT NULL,
    index_definition TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50) DEFAULT CURRENT_USER()
);
```

### 2.2 单个索引回滚

```sql
-- 语法：DROP INDEX index_name ON table_name;

-- 示例：回滚transaction表的status索引
DROP INDEX idx_transaction_status ON transaction;

-- 示例：回滚用户相关索引
DROP INDEX idx_users_status ON users;
DROP INDEX idx_users_phone ON users;
DROP INDEX idx_users_email ON users;
```

### 2.3 批量回滚脚本

```sql
-- 批量回滚所有优化添加的索引
-- 交易订单表索引回滚
DROP INDEX IF EXISTS idx_transaction_status ON transaction;
DROP INDEX IF EXISTS idx_transaction_from_user ON transaction;
DROP INDEX IF EXISTS idx_transaction_to_user ON transaction;
DROP INDEX IF EXISTS idx_transaction_user_status ON transaction;
DROP INDEX IF EXISTS idx_transaction_user_time ON transaction;
DROP INDEX IF EXISTS idx_transaction_currency_legal ON transaction;

-- 合约交易表索引回滚
DROP INDEX IF EXISTS idx_lever_user_id ON lever_transaction;
DROP INDEX IF EXISTS idx_lever_status ON lever_transaction;
DROP INDEX IF EXISTS idx_lever_user_status ON lever_transaction;
DROP INDEX IF EXISTS idx_lever_create_time ON lever_transaction;
DROP INDEX IF EXISTS idx_lever_currency_legal ON lever_transaction;
DROP INDEX IF EXISTS idx_lever_settled ON lever_transaction;

-- 秒合约订单表索引回滚
DROP INDEX IF EXISTS idx_micro_user_id ON micro_order;
DROP INDEX IF EXISTS idx_micro_status ON micro_order;
DROP INDEX IF EXISTS idx_micro_user_status ON micro_order;
DROP INDEX IF EXISTS idx_micro_create_time ON micro_order;
DROP INDEX IF EXISTS idx_micro_currency ON micro_order;

-- 钱包资产表索引回滚
DROP INDEX IF EXISTS idx_wallet_user_currency ON users_wallet;
DROP INDEX IF EXISTS idx_wallet_status ON users_wallet;
DROP INDEX IF EXISTS idx_wallet_address ON users_wallet;

-- 充值提现表索引回滚
DROP INDEX IF EXISTS idx_deposit_address_network ON deposit_address;
DROP INDEX IF EXISTS idx_deposit_address_status ON deposit_address;
DROP INDEX IF EXISTS idx_deposit_user_id ON deposit_order;
DROP INDEX IF EXISTS idx_deposit_status ON deposit_order;
DROP INDEX IF EXISTS idx_deposit_create_time ON deposit_order;
DROP INDEX IF EXISTS idx_withdraw_user_id ON users_wallet_out;
DROP INDEX IF EXISTS idx_withdraw_status ON users_wallet_out;
DROP INDEX IF EXISTS idx_withdraw_create_time ON users_wallet_out;

-- 账户流水表索引回滚
DROP INDEX IF EXISTS idx_account_log_user_id ON account_log;
DROP INDEX IF EXISTS idx_account_log_type ON account_log;
DROP INDEX IF EXISTS idx_account_log_currency ON account_log;
DROP INDEX IF EXISTS idx_account_log_user_time ON account_log;
DROP INDEX IF EXISTS idx_account_log_user_type ON account_log;
```

### 2.4 回滚验证

```sql
-- 验证索引已删除
SHOW INDEX FROM transaction;
SHOW INDEX FROM lever_transaction;
SHOW INDEX FROM micro_order;

-- 检查表状态
CHECK TABLE transaction QUICK;
```

---

## 三、MySQL配置回滚方案

### 3.1 配置回滚步骤

1. **停止MySQL服务**

   ```bash
   # Linux
   systemctl stop mysql
   
   # Windows
   net stop MySQL
   ```

2. **恢复原配置文件**

   ```bash
   # 备份当前配置
   cp /etc/mysql/my.cnf /etc/mysql/my.cnf.backup.optimized
   
   # 恢复原配置
   cp /etc/mysql/my.cnf.backup /etc/mysql/my.cnf
   ```

3. **重启MySQL服务**

   ```bash
   # Linux
   systemctl start mysql
   
   # Windows
   net start MySQL
   ```

### 3.2 关键配置参数回滚值

| 参数 | 优化值 | 回滚值 | 说明 |
|------|--------|--------|------|
| innodb_buffer_pool_size | 12G | 4G | 缓冲池大小 |
| max_connections | 500 | 100 | 最大连接数 |
| innodb_log_file_size | 1G | 256M | 日志文件大小 |
| table_open_cache | 4000 | 2000 | 表缓存大小 |
| thread_cache_size | 64 | 8 | 线程缓存大小 |

### 3.3 紧急回滚命令

```sql
-- 如果修改后出现严重性能问题，可通过SQL命令临时调整
SET GLOBAL innodb_buffer_pool_size = 4294967296;
SET GLOBAL max_connections = 100;
SET GLOBAL thread_cache_size = 8;
```

---

## 四、业务代码回滚方案

### 4.1 代码回滚策略

#### 4.1.1 缓存层回滚

1. **禁用缓存层**（在代码中注释缓存调用）

   ```go
   // 注释掉缓存获取
   // assets, err := cache.GetUserAssets(userID)
   
   // 直接从数据库获取
   assets, err := GetUserAssetsFromDB(userID)
   ```

2. **回滚缓存配置**

   ```go
   // 恢复原始配置
   cacheCfg := &cache.CacheConfig{
       EnableUserAssetsCache: false,
       EnableBalanceCache:    false,
       EnableMatchCache:      false,
       EnableMarketCache:     false,
       EnableStatisticsCache: false,
   }
   ```

#### 4.1.2 读写分离回滚

1. **统一使用主库**

   ```go
   // 回滚前：GetDB根据SQL类型选择库
   db := database.GetDB(sql)
   
   // 回滚后：统一使用主库
   db := database.MasterDB
   ```

2. **恢复原有数据库初始化**

   ```go
   // 注释掉读写分离初始化
   // database.InitWithReadWriteSeparation(&database.DBConfig{...})
   
   // 恢复原有初始化
   database.Init(cfg)
   ```

---

## 五、数据恢复方案

### 5.1 使用备份恢复

```bash
# 执行恢复脚本
./restore_database.bat ./backup/bibi2022_20240115_120000.sql.gz
```

### 5.2 时间点恢复

```bash
# 如果需要恢复到特定时间点
mysqlbinlog --stop-datetime="2024-01-15 12:00:00" binlog.000001 | mysql -u root -p
```

### 5.3 表级恢复

```sql
-- 恢复单个表
DROP TABLE IF EXISTS transaction;
SOURCE ./backup/transaction_table_backup.sql;
```

---

## 六、回滚执行流程

### 6.1 触发条件

1. **自动触发**
   - 错误日志中出现大量死锁警告
   - 查询响应时间超过阈值（>5秒）
   - 连接数持续达到上限
   - 系统可用性低于99%

2. **手动触发**
   - 业务方反馈功能异常
   - 监控指标出现异常波动
   - 运维人员判断需要回滚

### 6.2 执行流程

```
┌─────────────────────────────────────────────────────────┐
│                    回滚执行流程                          │
├─────────────────────────────────────────────────────────┤
│  1. 问题确认                                            │
│     ├── 确认问题影响范围                                 │
│     ├── 评估是否需要回滚                                 │
│     └── 选择回滚策略                                     │
├─────────────────────────────────────────────────────────┤
│  2. 通知相关方                                          │
│     ├── 通知运维团队                                     │
│     ├── 通知开发团队                                     │
│     └── 通知业务负责人                                   │
├─────────────────────────────────────────────────────────┤
│  3. 执行回滚                                            │
│     ├── 按优先级执行回滚操作                             │
│     ├── 索引回滚 → 配置回滚 → 代码回滚 → 数据恢复        │
│     └── 每步验证                                         │
├─────────────────────────────────────────────────────────┤
│  4. 验证确认                                            │
│     ├── 功能测试                                         │
│     ├── 性能验证                                         │
│     └── 监控确认                                         │
├─────────────────────────────────────────────────────────┤
│  5. 总结报告                                            │
│     ├── 问题分析                                         │
│     ├── 回滚记录                                         │
│     └── 改进措施                                         │
└─────────────────────────────────────────────────────────┘
```

### 6.3 回滚决策矩阵

| 问题类型 | 影响程度 | 回滚优先级 | 回滚方式 |
|---------|---------|-----------|---------|
| 数据丢失/错误 | 严重 | P0 | 立即恢复备份 |
| 大面积查询超时 | 严重 | P0 | 立即回滚索引和配置 |
| 死锁频繁 | 中等 | P1 | 逐步回滚索引 |
| 部分功能异常 | 中等 | P1 | 回滚相关代码 |
| 性能小幅下降 | 轻微 | P2 | 观察 + 微调 |
| 监控指标波动 | 轻微 | P3 | 继续监控 |

---

## 七、回滚后验证清单

### 7.1 功能验证

- [ ] 用户登录功能正常
- [ ] 订单创建和查询正常
- [ ] 充值提现功能正常
- [ ] KYC认证流程正常
- [ ] 账户资产显示正确
- [ ] 交易记录完整

### 7.2 性能验证

- [ ] 查询响应时间恢复正常
- [ ] 无长时间阻塞的查询
- [ ] 连接数在正常范围
- [ ] 无异常死锁

### 7.3 监控验证

- [ ] 慢查询日志无新增异常
- [ ] 错误日志无新增错误
- [ ] 系统资源使用正常
- [ ] 数据库健康状态良好

---

## 八、联系人与责任

| 角色 | 职责 | 联系方式 |
|-----|------|---------|
| DBA负责人 | 执行数据库相关回滚 | [待填写] |
| 运维负责人 | 服务重启和配置恢复 | [待填写] |
| 开发负责人 | 代码回滚和修复 | [待填写] |
| 项目经理 | 回滚决策和沟通 | [待填写] |

---

## 九、附件清单

1. [add_indexes.sql](add_indexes.sql) - 索引创建脚本（备份）
2. [my.cnf.optimized](my.cnf.optimized) - 优化配置（备份）
3. [backup_database.bat](backup_database.bat) - 备份脚本
4. [restore_database.bat](restore_database.bat) - 恢复脚本
