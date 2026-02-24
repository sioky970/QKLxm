-- ============================================================
-- SQL查询优化指南
-- 交易所系统性能优化 - 查询语句优化
-- ============================================================

-- ============================================================
-- 1. 避免SELECT * 查询
-- ============================================================

-- 优化前（错误示例）
SELECT * FROM transaction WHERE id = 100;

-- 优化后（正确示例）
SELECT id, from_user_id, to_user_id, currency, legal, price, number, status, create_time 
FROM transaction WHERE id = 100;

-- 说明：只查询需要的字段，减少网络传输和内存占用


-- ============================================================
-- 2. 使用覆盖索引避免回表
-- ============================================================

-- 优化前（需要回表查询）
SELECT id, status, create_time FROM transaction WHERE user_id = 100 AND status = 0;

-- 优化后（使用覆盖索引）
-- 假设已创建索引 idx_transaction_user_status(user_id, status, create_time)
-- 如果查询的字段都在索引中，MySQL可以直接从索引获取数据，无需回表


-- ============================================================
-- 3. 优化分页查询（深度分页优化）
-- ============================================================

-- 优化前（深度分页性能差）
SELECT * FROM transaction ORDER BY create_time DESC LIMIT 100000, 20;

-- 优化方案1：使用游标分页（推荐）
SELECT * FROM transaction 
WHERE create_time < :last_create_time 
ORDER BY create_time DESC 
LIMIT 20;

-- 优化方案2：使用延迟关联
SELECT t.* FROM (
    SELECT id FROM transaction ORDER BY create_time DESC LIMIT 100000, 20
) AS tmp
JOIN transaction t ON t.id = tmp.id;

-- 优化方案3：记录上次分页位置
SELECT * FROM transaction 
WHERE id < :last_id 
ORDER BY id DESC 
LIMIT 20;


-- ============================================================
-- 4. 优化LIKE查询
-- ============================================================

-- 优化前（无法使用索引）
SELECT * FROM users WHERE account_number LIKE '%123%';

-- 优化方案：如果是前缀匹配，可以使用索引
SELECT * FROM users WHERE account_number LIKE '123%';

-- 替代方案：使用全文索引（适用于中文搜索）
-- ALTER TABLE users ADD FULLTEXT INDEX ft_account (account_number);

-- 或者考虑使用搜索引擎（Elasticsearch）


-- ============================================================
-- 5. 优化JOIN操作
-- ============================================================

-- 优化前（多表JOIN性能差）
SELECT u.*, t.* 
FROM users u 
LEFT JOIN transaction t ON u.id = t.from_user_id 
LEFT JOIN lever_transaction l ON u.id = l.user_id 
WHERE u.status = 1;

-- 优化方案1：减少JOIN表数量，只连接必要的表
SELECT u.id, u.account_number, t.status, t.create_time
FROM users u
INNER JOIN transaction t ON u.id = t.from_user_id
WHERE u.status = 1
LIMIT 100;

-- 优化方案2：先聚合再JOIN
SELECT 
    u.id, 
    u.account_number,
    t.order_count,
    t.last_order_time
FROM users u
INNER JOIN (
    SELECT from_user_id, 
           COUNT(*) as order_count,
           MAX(create_time) as last_order_time
    FROM transaction 
    GROUP BY from_user_id
) t ON u.id = t.from_user_id
WHERE u.status = 1;

-- 优化方案3：确保JOIN字段有索引
-- ALTER TABLE transaction ADD INDEX idx_transaction_from_user (from_user_id);


-- ============================================================
-- 6. 优化COUNT操作
-- ============================================================

-- 优化前（COUNT全表）
SELECT COUNT(*) FROM transaction WHERE status = 0;

-- 优化方案：确保WHERE条件字段有索引
-- 假设已创建索引 idx_transaction_status(status)
-- COUNT(*) 会使用索引扫描，速度大幅提升


-- ============================================================
-- 7. 使用批量操作减少数据库交互
-- ============================================================

-- 优化前（循环单条插入）
INSERT INTO account_log (user_id, value, type, info) VALUES (1, 100, 1, 'test1');
INSERT INTO account_log (user_id, value, type, info) VALUES (2, 200, 1, 'test2');
INSERT INTO account_log (user_id, value, type, info) VALUES (3, 300, 1, 'test3');

-- 优化后（批量插入）
INSERT INTO account_log (user_id, value, type, info) VALUES 
(1, 100, 1, 'test1'),
(2, 200, 1, 'test2'),
(3, 300, 1, 'test3');

-- 注意：MySQL默认接受单条超过4MB的语句会报错
-- 建议每批插入100-500条


-- ============================================================
-- 8. 避免在WHERE条件中使用函数
-- ============================================================

-- 优化前（无法使用索引）
SELECT * FROM transaction WHERE DATE(create_time) = '2024-01-01';

-- 优化后（使用范围查询）
SELECT * FROM transaction 
WHERE create_time >= UNIX_TIMESTAMP('2024-01-01 00:00:00') 
AND create_time < UNIX_TIMESTAMP('2024-01-02 00:00:00');


-- ============================================================
-- 9. 使用EXPLAIN分析查询执行计划
-- ============================================================

-- 示例：分析查询是否使用了索引
EXPLAIN SELECT * FROM transaction WHERE user_id = 100 AND status = 0;

-- 关键字段说明：
-- type: 连接类型，const < eq_ref < ref < range < index < ALL
-- key: 实际使用的索引
-- rows: 扫描的行数
-- Extra: 额外信息，如Using filesort、Using temporary等


-- ============================================================
-- 10. 合理使用临时表
-- ============================================================

-- 当查询需要多次使用中间结果时，使用临时表
CREATE TEMPORARY TABLE temp_user_ids (
    user_id INT PRIMARY KEY
);

INSERT INTO temp_user_ids 
SELECT id FROM users WHERE status = 1;

SELECT * FROM transaction t
INNER JOIN temp_user_ids u ON t.from_user_id = u.user_id;

-- 临时表在会话结束时自动删除


-- ============================================================
-- 11. 优化OR条件
-- ============================================================

-- 优化前（OR可能放弃使用索引）
SELECT * FROM transaction WHERE user_id = 100 OR status = 0;

-- 优化方案：拆分为UNION（每个条件都可能使用索引）
SELECT * FROM transaction WHERE user_id = 100
UNION
SELECT * FROM transaction WHERE status = 0 AND user_id != 100;

-- 或者使用IN代替OR（如果值不多）
SELECT * FROM transaction WHERE user_id IN (100, 200, 300);


-- ============================================================
-- 12. 避免使用不等于查询
-- ============================================================

-- 优化前（!= 可能不使用索引）
SELECT * FROM users WHERE status != 0;

-- 优化方案：如果可能，使用正向查询
SELECT * FROM users WHERE status >= 1;


-- ============================================================
-- 13. 分区表优化（针对大表）
-- ============================================================

-- 按时间范围分区示例
-- ALTER TABLE transaction PARTITION BY RANGE (create_time) (
--     PARTITION p2024_01 VALUES LESS THAN (UNIX_TIMESTAMP('2024-02-01 00:00:00')),
--     PARTITION p2024_02 VALUES LESS THAN (UNIX_TIMESTAMP('2024-03-01 00:00:00')),
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );


-- ============================================================
-- 14. 定期维护优化
-- ============================================================

-- 优化表（回收碎片）
OPTIMIZE TABLE transaction;

-- 分析表（更新索引统计信息）
ANALYZE TABLE transaction;

-- 检查表
CHECK TABLE transaction;
