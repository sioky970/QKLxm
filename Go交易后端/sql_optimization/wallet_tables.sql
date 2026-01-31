-- =====================================================
-- 钱包划转系统数据库表结构
-- 创建时间: 2024
-- 功能: 支持合约账户与现货账户独立体系及资金划转
-- =====================================================

-- 1. 用户钱包资产表（扩展现有user_assets，支持多钱包类型）
CREATE TABLE IF NOT EXISTS `user_wallets` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `wallet_type` VARCHAR(20) NOT NULL COMMENT '钱包类型: spot-现货, contract-合约',
    `currency_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '货币ID(默认USDT)',
    `currency_name` VARCHAR(50) DEFAULT 'USDT' COMMENT '货币名称',
    `available_balance` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '可用余额',
    `locked_balance` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '锁定余额',
    `version` INT NOT NULL DEFAULT 0 COMMENT '版本号(乐观锁)',
    `last_trade_time` BIGINT NOT NULL DEFAULT 0 COMMENT '最后交易时间',
    `create_time` BIGINT NOT NULL COMMENT '创建时间',
    `update_time` BIGINT NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_wallet_type_currency` (`user_id`, `wallet_type`, `currency_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_wallet_type` (`wallet_type`),
    KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户钱包资产表';

-- 2. 钱包划转记录表
CREATE TABLE IF NOT EXISTS `wallet_transfer_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `transfer_no` VARCHAR(64) NOT NULL COMMENT '划转流水号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `from_wallet_type` VARCHAR(20) NOT NULL COMMENT '转出钱包类型: spot-现货, contract-合约',
    `to_wallet_type` VARCHAR(20) NOT NULL COMMENT '转入钱包类型: spot-现货, contract-合约',
    `currency_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '货币ID',
    `currency_name` VARCHAR(50) DEFAULT 'USDT' COMMENT '货币名称',
    `amount` DECIMAL(20,8) NOT NULL COMMENT '划转金额',
    `fee` DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '手续费',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-处理中, 2-成功, 3-失败',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `client_ip` VARCHAR(45) DEFAULT NULL COMMENT '客户端IP',
    `created_time` BIGINT NOT NULL COMMENT '创建时间',
    `completed_time` BIGINT DEFAULT NULL COMMENT '完成时间',
    `error_msg` VARCHAR(500) DEFAULT NULL COMMENT '错误信息',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_transfer_no` (`transfer_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包划转记录表';

-- 3. 钱包余额调整记录表（管理后台使用）
CREATE TABLE IF NOT EXISTS `wallet_adjustment_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `adjustment_no` VARCHAR(64) NOT NULL COMMENT '调整流水号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `wallet_type` VARCHAR(20) NOT NULL COMMENT '钱包类型: spot-现货, contract-合约',
    `currency_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '货币ID',
    `currency_name` VARCHAR(50) DEFAULT 'USDT' COMMENT '货币名称',
    `adjustment_type` TINYINT NOT NULL COMMENT '调整类型: 1-增加, 2-减少',
    `amount` DECIMAL(20,8) NOT NULL COMMENT '调整金额',
    `balance_before` DECIMAL(20,8) NOT NULL COMMENT '调整前余额',
    `balance_after` DECIMAL(20,8) NOT NULL COMMENT '调整后余额',
    `operator_id` BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
    `operator_name` VARCHAR(100) NOT NULL COMMENT '操作员名称',
    `reason` VARCHAR(500) NOT NULL COMMENT '调整原因',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-成功, 2-失败',
    `created_time` BIGINT NOT NULL COMMENT '创建时间',
    `completed_time` BIGINT DEFAULT NULL COMMENT '完成时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_adjustment_no` (`adjustment_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_operator_id` (`operator_id`),
    KEY `idx_wallet_type` (`wallet_type`),
    KEY `idx_created_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包余额调整记录表';

-- 4. 验证表是否创建成功
SELECT 'user_wallets' as table_name, COUNT(*) as record_count FROM user_wallets
UNION ALL
SELECT 'wallet_transfer_records', COUNT(*) FROM wallet_transfer_records
UNION ALL
SELECT 'wallet_adjustment_records', COUNT(*) FROM wallet_adjustment_records;
