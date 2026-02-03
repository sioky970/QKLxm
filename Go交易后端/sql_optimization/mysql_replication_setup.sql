-- ============================================================
-- MySQL 主从复制配置脚本
-- 交易所系统 - 读写分离架构
-- ============================================================
-- 执行步骤：
-- 1. 在主库上执行【主库配置】部分
-- 2. 在主库上执行【创建复制用户】部分
-- 3. 在从库上执行【从库配置】部分
-- 4. 在从库上执行【连接主库】部分
-- ============================================================

-- ============================================================
-- 【主库配置】
-- 在主库的 my.ini 中添加以下配置后重启MySQL
-- ============================================================
-- [mysqld]
-- server-id = 1
-- log-bin = mysql-bin
-- binlog_format = ROW
-- binlog_row_image = FULL
-- max_binlog_size = 1G
-- binlog_cache_size = 64M
-- expire_logs_days = 7
-- sync_binlog = 1
-- gtid_mode = ON
-- enforce_gtid_consistency = 1

-- ============================================================
-- 【创建复制用户】
-- 在主库上执行
-- ============================================================

-- 创建复制用户（请修改密码）
CREATE USER 'repl_user'@'%' IDENTIFIED BY 'ReplPassword123!@#';

-- 授予复制权限
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'repl_user'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 记录主库状态（执行完CHANGE MASTER TO后需要）
-- SHOW MASTER STATUS;
-- 记录File和Position的值

-- ============================================================
-- 【主库状态查询】
-- 在主库上执行，记录结果用于配置从库
-- ============================================================

-- 查看主库状态
SHOW MASTER STATUS;

-- 查看二进制日志文件列表
SHOW BINARY LOGS;

-- 查看GTID信息
SELECT @@GLOBAL.GTID_EXECUTED;

-- ============================================================
-- 【从库配置】
-- 在从库的 my.ini 中添加以下配置后重启MySQL
-- ============================================================
-- [mysqld]
-- server-id = 2  # 必须唯一，不同从库使用不同ID
-- relay-log = "C:/ProgramData/MySQL/MySQL Server 5.7/Data/relay-log"
-- relay-log-purge = ON
-- relay-log-recovery = ON
-- read-only = ON  # 从库只读
-- super-read-only = ON  # 超级用户也只读
-- slave-parallel-workers = 4  # 并行复制（MySQL 5.7.6+）
-- slave-parallel-type = LOGICAL_CLOCK
-- log-bin = mysql-bin  # 从库也可以开启二进制日志
-- log-slave-updates = ON  # 从库更新也写入二进制日志

-- ============================================================
-- 【连接主库】
-- 在从库上执行（替换为实际的主库IP和记录的值）
-- ============================================================

-- 停止复制
STOP SLAVE;

-- 重置复制配置
RESET SLAVE ALL;

-- 配置连接信息（使用GTID方式）
CHANGE MASTER TO
    MASTER_HOST='192.168.1.100',  -- 替换为主库IP
    MASTER_USER='repl_user',
    MASTER_PASSWORD='ReplPassword123!@#',  -- 替换为实际密码
    MASTER_PORT=3306,
    MASTER_AUTO_POSITION=1,  -- 使用GTID自动定位
    GET_MASTER_PUBLIC_KEY=1;

-- 启动复制
START SLAVE;

-- ============================================================
-- 【验证复制状态】
-- 在从库上执行
-- ============================================================

-- 查看复制状态
SHOW SLAVE STATUS\G

-- 关键检查项：
-- Slave_IO_Running: Yes
-- Slave_SQL_Running: Yes
-- Seconds_Behind_Master: 0 或 较小的数字
-- Last_IO_Error: 空
-- Last_SQL_Error: 空

-- 查看GTID复制状态
SELECT * FROM performance_schema.replication_applier_status_by_worker;

-- ============================================================
-- 【测试复制】
-- 在主库执行，然后在从库查询验证
-- ============================================================

-- 主库：创建测试数据库
-- CREATE DATABASE test_replication;
-- USE test_replication;
-- CREATE TABLE test_table (id INT, data VARCHAR(100));
-- INSERT INTO test_table VALUES (1, 'test data');
-- SELECT * FROM test_table;

-- 从库：验证数据同步
-- USE test_replication;
-- SELECT * FROM test_table;

-- ============================================================
-- 【监控脚本】
-- 在从库上定期执行
-- ============================================================

-- 检查复制延迟
SELECT 
    CASE 
        WHEN (SELECT COUNT(*) FROM performance_schema.replication_applier_status_by_worker WHERE SERVICE_STATE = 'ON' AND WORKER_ID IS NOT NULL) > 0 
        THEN '复制正常运行'
        ELSE '复制可能停止'
    END AS replication_status,
    (SELECT SUM(APPLIER_TRANSACTION_RETRIES) FROM performance_schema.replication_applier_status_by_worker) AS total_retries,
    (SELECT COUNT(*) FROM performance_schema.replication_applier_status_by_worker WHERE SERVICE_STATE = 'ON') AS active_workers;

-- 查看主从延迟
SHOW SLAVE STATUS\G
-- 查看 Seconds_Behind_Master 列

-- ============================================================
-- 【故障处理】
-- ============================================================

-- 1. 复制停止时，重启复制
START SLAVE;

-- 2. 跳过错误（谨慎使用）
SET GLOBAL sql_slave_skip_counter = 1;
START SLAVE;

-- 3. 重新同步从库（数据不一致时）
-- a. 停止复制
STOP SLAVE;
RESET SLAVE ALL;

-- b. 备份主库数据
-- mysqldump -u root -p --single-transaction --master-data=1 bibi2022 > backup.sql

-- c. 从库恢复数据
-- mysql -u root -p bibi2022 < backup.sql

-- d. 重新配置复制
CHANGE MASTER TO MASTER_HOST='...', ...;
START SLAVE;

-- ============================================================
-- 【多从库配置示例】
-- ============================================================

-- 从库2配置（server-id = 2）
-- 从库3配置（server-id = 3）
-- 以此类推...

-- ============================================================
-- 【读写分离应用配置】
-- 在应用配置文件中设置
-- ============================================================

-- Go应用配置示例（config.yaml）：
-- database:
--   host: "127.0.0.1"  # 主库地址
--   port: 3306
--   user: "root"
--   password: "root123456"
--   dbname: "bibi2022"
--   charset: "utf8"
--   max_idle_conns: 10
--   max_open_conns: 100
-- 
-- read_write_separation:
--   enabled: true
--   slaves:
--     - host: "192.168.1.101"  # 从库1地址
--       port: 3306
--       weight: 10
--     - host: "192.168.1.102"  # 从库2地址
--       port: 3306
--       weight: 10

-- ============================================================
-- 【清理测试】
-- 在主库执行
-- ============================================================

-- DROP DATABASE IF EXISTS test_replication;
