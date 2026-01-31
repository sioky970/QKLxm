-- =====================================================
-- 修复 user_assets 表索引问题
-- 创建时间: 2026-02-01
-- 问题: 现有唯一索引只包含 user_id，需要改为包含 user_id + wallet_type
-- =====================================================

-- 1. 首先检查并添加 wallet_type 字段（如果不存在）
SET @dbname = DATABASE();
SET @tablename = 'user_assets';
SET @columnname = 'wallet_type';

SET @sql = CONCAT(
    'SELECT COUNT(*) INTO @column_exists FROM information_schema.columns ',
    'WHERE table_schema = ''', @dbname, ''' ',
    'AND table_name = ''', @tablename, ''' ',
    'AND column_name = ''', @columnname, ''''
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 如果字段不存在则添加
SET @add_column = IF(@column_exists = 0, 
    'ALTER TABLE user_assets ADD COLUMN wallet_type VARCHAR(20) NOT NULL DEFAULT ''spot'' COMMENT ''钱包类型: spot=现货钱包, contract=合约钱包'' AFTER user_id',
    'SELECT ''wallet_type column already exists'' as message'
);

PREPARE stmt FROM @add_column;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 删除旧的唯一索引（如果存在）
SET @sql = CONCAT(
    'SELECT COUNT(*) INTO @index_exists FROM information_schema.statistics ',
    'WHERE table_schema = ''', @dbname, ''' ',
    'AND table_name = ''', @tablename, ''' ',
    'AND index_name = ''idx_user_assets_user_id'''
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @drop_index = IF(@index_exists > 0,
    'ALTER TABLE user_assets DROP INDEX idx_user_assets_user_id',
    'SELECT ''idx_user_assets_user_id does not exist'' as message'
);

PREPARE stmt FROM @drop_index;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. 删除可能存在的其他唯一索引
SET @sql = CONCAT(
    'SELECT COUNT(*) INTO @index_exists2 FROM information_schema.statistics ',
    'WHERE table_schema = ''', @dbname, ''' ',
    'AND table_name = ''', @tablename, ''' ',
    'AND index_name = ''idx_user_id'''
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @drop_index2 = IF(@index_exists2 > 0,
    'ALTER TABLE user_assets DROP INDEX idx_user_id',
    'SELECT ''idx_user_id does not exist'' as message'
);

PREPARE stmt FROM @drop_index2;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 4. 创建新的复合唯一索引
SET @sql = CONCAT(
    'SELECT COUNT(*) INTO @new_index_exists FROM information_schema.statistics ',
    'WHERE table_schema = ''', @dbname, ''' ',
    'AND table_name = ''', @tablename, ''' ',
    'AND index_name = ''idx_user_wallet'''
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @create_index = IF(@new_index_exists = 0,
    'ALTER TABLE user_assets ADD UNIQUE INDEX idx_user_wallet (user_id, wallet_type) COMMENT ''用户ID和钱包类型的复合唯一索引''',
    'SELECT ''idx_user_wallet already exists'' as message'
);

PREPARE stmt FROM @create_index;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 5. 为 wallet_type 字段创建普通索引（如果不存在）
SET @sql = CONCAT(
    'SELECT COUNT(*) INTO @wallet_type_index_exists FROM information_schema.statistics ',
    'WHERE table_schema = ''', @dbname, ''' ',
    'AND table_name = ''', @tablename, ''' ',
    'AND index_name = ''idx_wallet_type'''
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @create_wallet_type_index = IF(@wallet_type_index_exists = 0,
    'ALTER TABLE user_assets ADD INDEX idx_wallet_type (wallet_type) COMMENT ''钱包类型索引''',
    'SELECT ''idx_wallet_type already exists'' as message'
);

PREPARE stmt FROM @create_wallet_type_index;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 6. 更新现有数据的 wallet_type 为 'spot'（如果为 NULL 或空）
UPDATE user_assets SET wallet_type = 'spot' WHERE wallet_type IS NULL OR wallet_type = '';

-- 7. 验证迁移结果
SELECT 
    'Migration completed' as status,
    COUNT(*) as total_records,
    SUM(CASE WHEN wallet_type = 'spot' THEN 1 ELSE 0 END) as spot_wallets,
    SUM(CASE WHEN wallet_type = 'contract' THEN 1 ELSE 0 END) as contract_wallets
FROM user_assets;
