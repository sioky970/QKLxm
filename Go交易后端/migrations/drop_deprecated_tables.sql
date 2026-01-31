-- =====================================================
-- 删除废弃表迁移脚本
-- 创建时间: 2026-02-01
-- 功能: 删除已废弃的 users_wallet 表及相关索引
-- 注意: 执行前请确保所有数据已迁移到 user_assets 表
-- =====================================================

-- 1. 删除 users_wallet 表的外键约束（如果有）
-- 注意：根据实际情况调整外键名称
-- ALTER TABLE users_wallet DROP FOREIGN KEY IF EXISTS fk_users_wallet_user;
-- ALTER TABLE users_wallet DROP FOREIGN KEY IF EXISTS fk_users_wallet_currency;

-- 2. 删除 users_wallet 表的索引
DROP INDEX IF EXISTS idx_user_currency ON users_wallet;
DROP INDEX IF EXISTS idx_user_id ON users_wallet;
DROP INDEX IF EXISTS idx_currency ON users_wallet;

-- 3. 删除 users_wallet 表
-- 警告：此操作不可逆，请确保数据已备份或迁移
DROP TABLE IF EXISTS users_wallet;

-- 4. 验证删除结果
SELECT 'users_wallet table dropped successfully' AS status;

-- 5. 显示当前剩余的资产相关表
SHOW TABLES LIKE '%asset%';
SHOW TABLES LIKE '%wallet%';
