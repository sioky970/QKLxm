-- =====================================================
-- 交割合约功能数据库迁移脚本
-- 创建时间: 2026-02-01
-- 功能: 添加交割合约账户支持，扩展三账户体系
-- =====================================================

-- 1. 为 user_assets 表添加交割合约专用字段
ALTER TABLE user_assets 
ADD COLUMN delivery_margin DECIMAL(20,8) DEFAULT 0 COMMENT '交割合约保证金' AFTER usdt_locked;

ALTER TABLE user_assets 
ADD COLUMN delivery_pnl DECIMAL(20,8) DEFAULT 0 COMMENT '交割合约盈亏' AFTER delivery_margin;

-- 2. 创建交割合约配置表
CREATE TABLE delivery_contracts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    symbol VARCHAR(50) NOT NULL UNIQUE COMMENT '合约代码',
    base_asset VARCHAR(20) NOT NULL COMMENT '基础资产',
    quote_asset VARCHAR(20) NOT NULL COMMENT '计价资产',
    contract_size DECIMAL(20,8) NOT NULL COMMENT '合约乘数',
    min_price DECIMAL(20,8) NOT NULL COMMENT '最小价格变动单位',
    max_leverage INT NOT NULL COMMENT '最大杠杆倍数',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    delivery_date DATETIME NOT NULL COMMENT '交割日期',
    settle_type TINYINT NOT NULL COMMENT '交割方式: 1=现金, 2=实物',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_symbol (symbol),
    KEY idx_status (status),
    KEY idx_delivery_date (delivery_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交割合约配置表';

-- 3. 创建交割合约持仓表
CREATE TABLE delivery_positions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    contract_id BIGINT NOT NULL COMMENT '合约ID',
    symbol VARCHAR(50) NOT NULL COMMENT '合约代码',
    side TINYINT NOT NULL COMMENT '方向: 1=做多, 2=做空',
    size DECIMAL(20,8) NOT NULL COMMENT '持仓数量',
    entry_price DECIMAL(20,8) NOT NULL COMMENT '开仓价格',
    leverage INT NOT NULL COMMENT '杠杆倍数',
    margin DECIMAL(20,8) NOT NULL COMMENT '保证金',
    current_price DECIMAL(20,8) DEFAULT 0 COMMENT '当前价格',
    unrealized_pnl DECIMAL(20,8) DEFAULT 0 COMMENT '未实现盈亏',
    take_profit_price DECIMAL(20,8) COMMENT '止盈价格',
    stop_loss_price DECIMAL(20,8) COMMENT '止损价格',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1=持仓, 0=已平仓',
    delivery_status TINYINT NOT NULL DEFAULT 0 COMMENT '交割状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_user_contract (user_id, contract_id),
    KEY idx_user_status (user_id, status),
    KEY idx_symbol_status (symbol, status),
    KEY idx_delivery_status (delivery_status),
    FOREIGN KEY (contract_id) REFERENCES delivery_contracts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交割合约持仓表';

-- 4. 更新钱包类型枚举，确保支持所有三种类型
-- 注意：这里只是确保数据一致性
UPDATE user_assets SET wallet_type = 'delivery' WHERE wallet_type = 'delivery';

-- 5. 插入默认交割合约配置数据
INSERT INTO delivery_contracts (symbol, base_asset, quote_asset, contract_size, min_price, max_leverage, delivery_date, settle_type, status) VALUES
('BTC-2024-03', 'BTC', 'USDT', 0.001, 0.01, 100, '2024-03-29 16:00:00', 1, 1),
('ETH-2024-03', 'ETH', 'USDT', 0.01, 0.001, 100, '2024-03-29 16:00:00', 1, 1),
('BNB-2024-03', 'BNB', 'USDT', 0.01, 0.001, 100, '2024-03-29 16:00:00', 1, 1),
('BTC-2024-06', 'BTC', 'USDT', 0.001, 0.01, 100, '2024-06-28 16:00:00', 1, 1),
('ETH-2024-06', 'ETH', 'USDT', 0.01, 0.001, 100, '2024-06-28 16:00:00', 1, 1);

-- 6. 创建账户类型映射视图
CREATE OR REPLACE VIEW account_types AS
SELECT 
    user_id,
    wallet_type,
    CASE 
        WHEN wallet_type = 'spot' THEN '现货账户'
        WHEN wallet_type = 'contract' THEN '永续合约账户'
        WHEN wallet_type = 'delivery' THEN '交割合约账户'
        ELSE '未知账户'
    END as account_type_name,
    usdt_balance,
    usdt_locked,
    delivery_margin,
    delivery_pnl,
    total_value_usdt
FROM user_assets;

-- 7. 验证迁移结果
SELECT 'user_assets表字段验证' as check_item, 
       COUNT(*) as total_count,
       COUNT(delivery_margin) as delivery_margin_count,
       COUNT(delivery_pnl) as delivery_pnl_count
FROM user_assets;

SELECT 'delivery_contracts表验证' as check_item, COUNT(*) as total_contracts FROM delivery_contracts;

SELECT 'delivery_positions表验证' as check_item, COUNT(*) as total_positions FROM delivery_positions;

-- 8. 显示迁移完成状态
SELECT '交割合约功能数据库迁移完成' AS status, NOW() AS completion_time;