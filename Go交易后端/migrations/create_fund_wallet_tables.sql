-- ============================================================
-- 资金钱包系统数据库迁移脚本
-- 创建时间: 2024-01-24
-- 功能: 支持资金钱包（Fund Wallet）与现货、合约账户的划转
-- 特点: 资金钱包不参与交易，仅用于资金存储
-- ============================================================

-- 确保使用正确的数据库
USE bibi2022;

-- ============================================================
-- 1. 创建资金钱包表 (fund_wallets)
-- ============================================================
DROP TABLE IF EXISTS `fund_wallets`;
CREATE TABLE `fund_wallets` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `currency_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '币种ID（默认USDT）',
    `currency_name` VARCHAR(50) NOT NULL DEFAULT 'USDT' COMMENT '币种名称',
    `available_balance` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '可用余额',
    `locked_balance` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '锁定余额',
    `total_deposit` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '累计充值金额',
    `total_withdraw` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '累计提现金额',
    `version` INT NOT NULL DEFAULT 0 COMMENT '版本号（乐观锁）',
    `last_trade_time` BIGINT NOT NULL DEFAULT 0 COMMENT '最后交易时间',
    `create_time` BIGINT NOT NULL COMMENT '创建时间',
    `update_time` BIGINT NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_fund_user_currency` (`user_id`, `currency_id`),
    KEY `idx_fund_user_id` (`user_id`),
    KEY `idx_fund_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资金钱包表';

-- ============================================================
-- 2. 创建资金钱包划转记录表 (fund_wallet_transfers)
-- ============================================================
DROP TABLE IF EXISTS `fund_wallet_transfers`;
CREATE TABLE `fund_wallet_transfers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `transfer_no` VARCHAR(64) NOT NULL COMMENT '划转单号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `transfer_type` TINYINT NOT NULL COMMENT '划转类型：1=充值到资金钱包，2=从资金钱包提走，3=转入现货，4=从现货转入，5=转入合约，6=从合约转入，7=转入交割，8=从交割转入',
    `from_wallet_type` VARCHAR(20) DEFAULT NULL COMMENT '转出钱包类型（spot/contract/delivery/fund）',
    `to_wallet_type` VARCHAR(20) DEFAULT NULL COMMENT '转入钱包类型（spot/contract/delivery/fund）',
    `currency_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '币种ID',
    `currency_name` VARCHAR(50) NOT NULL DEFAULT 'USDT' COMMENT '币种名称',
    `amount` DECIMAL(20,8) NOT NULL COMMENT '划转金额',
    `fee` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '手续费',
    `balance_before` DECIMAL(20,8) NOT NULL COMMENT '操作前余额',
    `balance_after` DECIMAL(20,8) NOT NULL COMMENT '操作后余额',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0=待处理，1=成功，2=失败',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `client_ip` VARCHAR(50) DEFAULT NULL COMMENT '客户端IP',
    `created_time` BIGINT NOT NULL COMMENT '创建时间',
    `completed_time` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_fund_transfer_no` (`transfer_no`),
    KEY `idx_fund_transfer_user` (`user_id`),
    KEY `idx_fund_transfer_type` (`transfer_type`),
    KEY `idx_fund_transfer_status` (`status`),
    KEY `idx_fund_transfer_created` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资金钱包划转记录表';

-- ============================================================
-- 3. 为现有用户创建资金钱包记录
-- 假设已有用户表，遍历所有用户创建初始资金钱包
-- ============================================================
-- 注意：此脚本仅初始化结构，实际用户钱包创建由应用服务处理
-- INSERT INTO fund_wallets (user_id, currency_id, currency_name, available_balance, create_time, update_time)
-- SELECT id, 1, 'USDT', 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP() FROM users;

-- ============================================================
-- 4. 添加索引优化（如果表已存在）
-- ============================================================
-- 为 fund_wallets 添加索引（如果不存在）
-- CREATE INDEX idx_fund_user_currency ON fund_wallets (user_id, currency_id);
-- CREATE INDEX idx_fund_update_time ON fund_wallets (update_time);

-- ============================================================
-- 5. 设置表注释
-- ============================================================
ALTER TABLE `fund_wallets` COMMENT = '资金钱包表：用于存储用户资金，支持与现货、合约账户自由划转';
ALTER TABLE `fund_wallet_transfers` COMMENT = '资金钱包划转记录表：记录所有资金钱包的划转操作';

-- ============================================================
-- 6. 验证表创建
-- ============================================================
SELECT COUNT(*) AS fund_wallets_count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'fund_wallets';
SELECT COUNT(*) AS fund_transfers_count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'fund_wallet_transfers';
