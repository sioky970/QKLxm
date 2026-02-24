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
    `total_balance` DECIMAL(20,8) GENERATED ALWAYS AS (available_balance + locked_balance) STORED COMMENT '总余额',
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

-- 4. 插入初始用户钱包数据（从现有user_assets迁移）
-- 注意: 此脚本需要在数据迁移工具执行后手动运行
-- INSERT INTO user_wallets (user_id, wallet_type, currency_id, currency_name, available_balance, locked_balance, create_time, update_time)
-- SELECT 
--     user_id,
--     'spot' as wallet_type,
--     1 as currency_id,
--     'USDT' as currency_name,
--     usdt_balance,
--     usdt_locked,
--     create_time,
--     update_time
-- FROM user_assets;

-- 5. 创建钱包划转的存储过程（可选，用于复杂业务逻辑）
DELIMITER //
CREATE PROCEDURE `sp_wallet_transfer`(
    IN p_user_id BIGINT,
    IN p_from_wallet VARCHAR(20),
    IN p_to_wallet VARCHAR(20),
    IN p_currency_id BIGINT,
    IN p_amount DECIMAL(20,8),
    IN p_transfer_no VARCHAR(64)
)
BEGIN
    DECLARE v_from_balance DECIMAL(20,8);
    DECLARE v_to_balance DECIMAL(20,8);
    DECLARE v_lock_version INT;
    DECLARE v_error_msg VARCHAR(500);
    DECLARE v_success INT DEFAULT 0;
    
    DECLARE CONTINUE HANDLER FOR SQLEXCEPTION 
    BEGIN
        GET DIAGNOSTICS CONDITION 1 v_error_msg = MESSAGE_TEXT;
        UPDATE wallet_transfer_records 
        SET status = 3, error_msg = v_error_msg, completed_time = UNIX_TIMESTAMP()
        WHERE transfer_no = p_transfer_no;
        ROLLBACK;
    END;
    
    START TRANSACTION;
    
    -- 检查源钱包余额
    SELECT available_balance, version INTO v_from_balance, v_lock_version
    FROM user_wallets 
    WHERE user_id = p_user_id 
        AND wallet_type = p_from_wallet 
        AND currency_id = p_currency_id
    FOR UPDATE;
    
    IF v_from_balance < p_amount THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '余额不足';
    END IF;
    
    -- 扣减源钱包余额
    UPDATE user_wallets 
    SET available_balance = available_balance - p_amount,
        update_time = UNIX_TIMESTAMP(),
        version = version + 1
    WHERE user_id = p_user_id 
        AND wallet_type = p_from_wallet 
        AND currency_id = p_currency_id
        AND version = v_lock_version;
    
    IF ROW_COUNT() = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '并发更新失败，请重试';
    END IF;
    
    -- 增加目标钱包余额
    UPDATE user_wallets 
    SET available_balance = available_balance + p_amount,
        update_time = UNIX_TIMESTAMP()
    WHERE user_id = p_user_id 
        AND wallet_type = p_to_wallet 
        AND currency_id = p_currency_id;
    
    -- 如果目标钱包不存在，则创建
    IF ROW_COUNT() = 0 THEN
        INSERT INTO user_wallets (user_id, wallet_type, currency_id, currency_name, available_balance, locked_balance, create_time, update_time)
        VALUES (p_user_id, p_to_wallet, p_currency_id, 'USDT', p_amount, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
    END IF;
    
    -- 更新划转记录状态
    UPDATE wallet_transfer_records 
    SET status = 2, completed_time = UNIX_TIMESTAMP()
    WHERE transfer_no = p_transfer_no;
    
    COMMIT;
    SET v_success = 1;
    
    SELECT v_success as success, v_from_balance as from_balance, p_amount as amount;
END //
DELIMITER ;

-- 6. 创建索引优化查询性能
-- 注意: 如果索引已存在，此命令会忽略错误
-- ALTER TABLE user_wallets ADD INDEX idx_user_wallet (user_id, wallet_type);
-- ALTER TABLE wallet_transfer_records ADD INDEX idx_user_transfer (user_id, created_time);
-- ALTER TABLE wallet_adjustment_records ADD INDEX idx_user_adjustment (user_id, created_time);
