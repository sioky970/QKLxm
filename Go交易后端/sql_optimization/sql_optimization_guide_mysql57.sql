-- ============================================================
-- MySQL 5.7 SQL查询优化指南
-- 交易所系统专用
-- 版本: MySQL 5.7.x
-- ============================================================

-- ============================================================
-- 第一部分：基础优化原则
-- ============================================================

-- 1.1 避免使用SELECT *
-- -------------------------------------------------------

-- 优化前（性能差）
SELECT * FROM transaction WHERE user_id = 100;

-- 优化后（只查询必要字段）
SELECT id, order_no, status, create_time, price, number 
FROM transaction WHERE user_id = 100;

-- 1.2 使用EXPLAIN分析查询
-- -------------------------------------------------------

-- 基本用法
EXPLAIN SELECT * FROM transaction WHERE user_id = 100;

-- 详细输出（MySQL 5.7支持）
EXPLAIN FORMAT=JSON 
SELECT * FROM transaction WHERE user_id = 100;


-- ============================================================
-- 第二部分：分页优化
-- ============================================================

-- 2.1 深度分页优化（最常见性能问题）
-- -------------------------------------------------------

-- 优化前（性能差 - 当offset很大时）
SELECT * FROM transaction ORDER BY create_time DESC LIMIT 100000, 20;

-- 优化方案1：使用游标分页（推荐）
-- 适用于：可按唯一有序字段（如时间、ID）翻页
SELECT * FROM transaction 
WHERE create_time < '2024-01-15 12:00:00' 
ORDER BY create_time DESC LIMIT 20;

-- 优化方案2：子查询+JOIN
-- 适用于：无法使用游标的情况
SELECT t.* FROM (
    SELECT id FROM transaction 
    WHERE user_id = 100 
    ORDER BY create_time DESC 
    LIMIT 100000, 20
) AS tmp JOIN transaction t ON t.id = tmp.id;

-- 优化方案3：延迟关联
SELECT * FROM transaction 
WHERE id IN (
    SELECT id FROM (
        SELECT id FROM transaction 
        WHERE user_id = 100 
        ORDER BY create_time DESC 
        LIMIT 100000, 20
    ) AS tmp
);


-- ============================================================
-- 第三部分：索引优化
-- ============================================================

-- 3.1 确保WHERE子句使用索引
-- -------------------------------------------------------

-- 优化前（无法使用索引）
SELECT * FROM transaction WHERE status LIKE '%已%';

-- 优化后（使用索引）
SELECT * FROM transaction WHERE status = 1;


-- 3.2 复合索引最佳实践
-- -------------------------------------------------------

-- 创建复合索引时，遵循最左前缀原则
-- 索引：(user_id, status, create_time)

-- 以下查询都能使用索引
SELECT * FROM transaction WHERE user_id = 100;
SELECT * FROM transaction WHERE user_id = 100 AND status = 0;
SELECT * FROM transaction WHERE user_id = 100 AND status = 0 AND create_time > '2024-01-01';

-- 以下查询无法使用完整索引
SELECT * FROM transaction WHERE status = 0;  -- 跳过user_id
SELECT * FROM transaction WHERE user_id = 100 AND create_time > '2024-01-01';  -- 跳过status


-- 3.3 避免在索引列上使用函数
-- -------------------------------------------------------

-- 优化前（无法使用索引）
SELECT * FROM transaction 
WHERE DATE(create_time) = '2024-01-15';

-- 优化后（使用索引）
SELECT * FROM transaction 
WHERE create_time >= '2024-01-15 00:00:00' 
AND create_time < '2024-01-16 00:00:00';


-- 3.4 使用覆盖索引避免回表
-- -------------------------------------------------------

-- 索引：(user_id, status, create_time)

-- 优化前（需要回表查询）
SELECT id, order_no, price, number FROM transaction 
WHERE user_id = 100 AND status = 0;

-- 如果查询的列都在索引中，则无需回表
-- EXPLAIN中Extra列显示Using index表示使用覆盖索引


-- ============================================================
-- 第四部分：JOIN优化
-- ============================================================

-- 4.1 小表驱动大表
-- -------------------------------------------------------

-- 优化前（可能使用大表驱动）
SELECT * FROM t 
JOIN users u ON transaction t.user_id = u.id
WHERE t.status = 0;

-- 优化后（使用小表驱动，MySQL 5.7的优化器可能自动优化）
SELECT * FROM users u 
JOIN transaction t ON t.user_id = u.id
WHERE t.status = 0;


-- 4.2 确保JOIN列有索引
-- -------------------------------------------------------

-- 确保被JOIN的列有索引
-- transaction.from_user_id应该有索引
-- users.id是主键，已有索引


-- 4.3 避免过多JOIN
-- -------------------------------------------------------

-- 优化前（JOIN多个表）
SELECT * FROM transaction t 
JOIN users u ON t.user_id = u.id
JOIN currency c ON t.currency = c.id
JOIN currency_match m ON t.currency = m.currency AND t.legal = m.legal
WHERE t.id = 1;

-- 优化后（减少JOIN，按需查询）
-- 先查询主表
SELECT * FROM transaction WHERE id = 1;
-- 再查询关联数据


-- ============================================================
-- 第五部分：聚合查询优化
-- ============================================================

-- 5.1 使用近似值代替精确COUNT
-- -------------------------------------------------------

-- 精确COUNT（性能较差）
SELECT COUNT(*) FROM transaction WHERE user_id = 100;

-- 近似值（性能更好，适用于显示"大约"数量）
SELECT TABLE_ROWS FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = 'bibi2022' AND TABLE_NAME = 'transaction';


-- 5.2 批量GROUP BY优化
-- -------------------------------------------------------

-- 优化前
SELECT user_id, SUM(amount) FROM transaction GROUP BY user_id;

-- 使用索引加速GROUP BY
-- 确保GROUP BY的列在索引中
ALTER TABLE transaction ADD INDEX idx_trans_user_amount (user_id, amount);


-- ============================================================
-- 第六部分：MySQL 5.7 特定优化
-- ============================================================

-- 6.1 关闭Query Cache（如未在配置中全局关闭）
-- -------------------------------------------------------

-- MySQL 5.7中Query Cache已弃用，建议关闭
SET GLOBAL query_cache_size = 0;
SET GLOBAL query_cache_type = 0;


-- 6.2 使用STRAIGHT_JOIN强制JOIN顺序
-- -------------------------------------------------------

-- 当优化器JOIN顺序不理想时，可强制指定
SELECT STRAIGHT_JOIN t.*, u.*, c.*
FROM transaction t
STRAIGHT_JOIN users u ON t.user_id = u.id
STRAIGHT_JOIN currency c ON t.currency = c.id
WHERE t.status = 0;


-- 6.3 使用SQL_BIG_RESULT优化GROUP BY
-- -------------------------------------------------------

-- 当GROUP BY结果集很大时
SELECT SQL_BIG_RESULT user_id, COUNT(*) 
FROM transaction 
GROUP BY user_id;


-- 6.4 使用FORCE INDEX强制使用索引
-- -------------------------------------------------------

-- 当优化器选择索引不理想时
SELECT * FROM transaction FORCE INDEX (idx_trans_user_status) 
WHERE user_id = 100 AND status = 0;


-- 6.5 JSON字段查询优化（MySQL 5.7）
-- -------------------------------------------------------

-- 假设有一列extra_data TEXT存储JSON

-- 优化前（需要扫描全表）
SELECT * FROM users WHERE extra_data LIKE '%"vip":"1"%';

-- 优化后1：使用JSON_CONTAINS（MySQL 5.7支持）
SELECT * FROM users 
WHERE JSON_CONTAINS(extra_data, '{"vip":"1"}');

-- 优化后2：为JSON中常用字段添加虚拟列（MySQL 5.7.8+）
ALTER TABLE users ADD vip_level VARCHAR(10) 
GENERATED ALWAYS AS (JSON_UNQUOTE(extra_data->'$.vip')) STORED;

ALTER TABLE users ADD INDEX idx_users_vip (vip_level);


-- ============================================================
-- 第七部分：批量操作优化
-- ============================================================

-- 7.1 批量INSERT优化
-- -------------------------------------------------------

-- 优化前（逐条插入）
INSERT INTO account_log (user_id, amount) VALUES (1, 100);
INSERT INTO account_log (user_id, amount) VALUES (2, 200);
INSERT INTO account_log (user_id, amount) VALUES (3, 300);

-- 优化后（批量插入）
INSERT INTO account_log (user_id, amount) VALUES 
(1, 100), (2, 200), (3, 300);

-- 批量大小建议：每批100-1000条


-- 7.2 批量UPDATE优化
-- -------------------------------------------------------

-- 优化前（逐条更新）
UPDATE users SET status = 1 WHERE id = 1;
UPDATE users SET status = 1 WHERE id = 2;

-- 优化后（批量更新）
UPDATE users SET status = 1 WHERE id IN (1, 2, 3, 4, 5);


-- 7.3 使用LOAD DATA批量导入
-- -------------------------------------------------------

-- 适合大数据量导入
LOAD DATA LOCAL INFILE '/path/to/data.csv' 
INTO TABLE transaction 
FIELDS TERMINATED BY ',' 
LINES TERMINATED BY '\n' 
(id, order_no, status, ...);


-- ============================================================
-- 第八部分：监控与诊断
-- ============================================================

-- 8.1 查看慢查询日志
-- -------------------------------------------------------

-- 慢查询阈值已在配置中设置为1秒
-- 查看慢查询日志位置：/var/log/mysql/slow.log

-- 分析慢查询（使用mysqldumpslow）
mysqldumpslow -t 10 /var/log/mysql/slow.log


-- 8.2 查看当前运行查询
-- -------------------------------------------------------

-- 查看当前连接和查询
SHOW PROCESSLIST;

-- 查看完整信息
SHOW FULL PROCESSLIST;


-- 8.3 查看索引使用情况
-- -------------------------------------------------------

-- 查看哪些索引被使用
SELECT 
    object_schema,
    object_name,
    index_name,
    cardinality
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE object_schema = 'bibi2022'
ORDER BY count_star DESC;


-- 8.4 查看表统计信息
-- -------------------------------------------------------

-- 查看表的统计信息
SHOW TABLE STATUS LIKE 'transaction';

-- 更新统计信息（重要：在分析查询后执行）
ANALYZE TABLE transaction;


-- ============================================================
-- 第九部分：常见反模式与修复
-- ============================================================

-- 9.1 反模式：NOT IN 子查询
-- -------------------------------------------------------

-- 优化前（可能产生笛卡尔积）
SELECT * FROM users 
WHERE id NOT IN (SELECT user_id FROM transaction WHERE status = 0);

-- 优化后1（使用LEFT JOIN + IS NULL）
SELECT u.* FROM users u
LEFT JOIN transaction t ON u.id = t.user_id AND t.status = 0
WHERE t.user_id IS NULL;

-- 优化后2（使用NOT EXISTS）
SELECT * FROM users u 
WHERE NOT EXISTS (
    SELECT 1 FROM transaction t 
    WHERE t.user_id = u.id AND t.status = 0
);


-- 9.2 反模式：OR条件导致索引失效
-- -------------------------------------------------------

-- 优化前（OR可能导致索引失效）
SELECT * FROM transaction 
WHERE user_id = 100 OR status = 0;

-- 优化后（使用UNION）
SELECT * FROM transaction WHERE user_id = 100
UNION ALL
SELECT * FROM transaction WHERE user_id <> 100 AND status = 0;


-- 9.3 反模式：不等值比较
-- -------------------------------------------------------

-- 优化前
SELECT * FROM transaction WHERE status <> 0;

-- 优化后（根据业务逻辑优化）
SELECT * FROM transaction WHERE status IN (1, 2, 3);


-- ============================================================
-- 第十部分：存储过程优化（MySQL 5.7）
-- ============================================================

-- 10.1 使用LOOP替代游标
-- -------------------------------------------------------

DELIMITER //
CREATE PROCEDURE batch_update_status()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE user_id INT;
    DECLARE cur CURSOR FOR SELECT id FROM users WHERE status = 0;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    OPEN cur;
    
    read_loop: LOOP
        FETCH cur INTO user_id;
        IF done THEN
            LEAVE read_loop;
        END IF;
        
        UPDATE transaction SET status = 1 WHERE user_id = user_id;
    END LOOP;
    
    CLOSE cur;
END //
DELIMITER ;


-- 10.2 使用批量处理优化循环
-- -------------------------------------------------------

DELIMITER //
CREATE PROCEDURE batch_process_transactions()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE min_id INT DEFAULT 0;
    DECLARE max_id INT;
    DECLARE batch_size INT DEFAULT 1000;
    
    -- 获取最大ID
    SELECT MAX(id) INTO max_id FROM transaction;
    
    SET min_id = 0;
    
    WHILE min_id < max_id DO
        -- 批量处理
        UPDATE transaction 
        SET status = 2 
        WHERE id > min_id AND id <= min_id + batch_size;
        
        SET min_id = min_id + batch_size;
    END WHILE;
END //
DELIMITER ;


-- ============================================================
-- 附录：优化检查清单
-- ============================================================

-- 日常检查项
-- 1. 慢查询日志是否有新增的慢查询？
-- 2. 表统计信息是否需要更新（ANALYZE TABLE）？
-- 3. 索引碎片率是否过高（>30%）？
-- 4. 锁等待时间是否过长？

-- 定期维护项
-- 1. 每周：分析慢查询，优化查询语句
-- 2. 每月：更新表统计信息
-- 3. 每季度：审查索引使用情况，删除无用索引
-- 4. 每半年：评估是否需要添加新索引
