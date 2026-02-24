-- =====================================================
-- user_assets 表添加 wallet_type 字段迁移脚本
-- 创建时间: 2026-02-01
-- 功能: 支持现货钱包(spot)和合约钱包(contract)分离
-- =====================================================

-- 1. 添加 wallet_type 字段
ALTER TABLE user_assets 
ADD COLUMN wallet_type VARCHAR(20) NOT NULL DEFAULT 'spot' 
COMMENT '钱包类型: spot=现货钱包, contract=合约钱包' 
AFTER user_id;

-- 2. 删除旧的唯一索引（如果存在）
-- 注意：根据实际数据库情况调整索引名称
-- DROP INDEX IF EXISTS idx_user_id ON user_assets;

-- 3. 创建新的复合唯一索引
CREATE UNIQUE INDEX idx_user_wallet 
ON user_assets(user_id, wallet_type) 
COMMENT '用户ID和钱包类型的复合唯一索引';

-- 4. 为 wallet_type 字段创建普通索引（用于快速查询特定类型的钱包）
CREATE INDEX idx_wallet_type 
ON user_assets(wallet_type) 
COMMENT '钱包类型索引';

-- 5. 将现有数据标记为现货钱包（兼容旧数据）
-- 注意：如果系统之前已经运行过，所有数据都在现货钱包中
UPDATE user_assets 
SET wallet_type = 'spot' 
WHERE wallet_type IS NULL OR wallet_type = '';

-- 6. 验证迁移结果
SELECT 
    wallet_type,
    COUNT(*) as count,
    SUM(usdt_balance) as total_balance,
    SUM(usdt_locked) as total_locked
FROM user_assets 
GROUP BY wallet_type;

-- 7. 显示迁移完成状态
SELECT 'user_assets table migration completed successfully' AS status;
