-- 创建交割合约配置表
CREATE TABLE IF NOT EXISTS delivery_contracts (
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建交割合约持仓表
CREATE TABLE IF NOT EXISTS delivery_positions (
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
    INDEX idx_user_contract (user_id, contract_id),
    INDEX idx_user_status (user_id, status),
    INDEX idx_symbol_status (symbol, status)
);