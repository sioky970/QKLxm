#!/bin/bash
# 添加管理员账号脚本

# 进入项目目录
cd /d/工程/交易所开发/QKLxm-main/Go交易后端

echo "开始添加管理员账号..."

# 使用Go运行一个简单的数据库操作来添加管理员
go run -exec "env" cmd/scripts/add_admin.go 2>/dev/null || echo "通过数据库命令添加..."

# 如果Go脚本不可用，直接使用MySQL命令
docker exec exchange_mysql mysql -uroot -proot123 -e "
USE exchange_db;

-- 检查是否已存在admin用户
SELECT id, username, phone, level FROM users WHERE username = 'admin' OR phone = 'admin';

-- 如果不存在则创建管理员
INSERT INTO users (phone, username, password, level, status, create_time, update_time)
VALUES ('admin', 'admin', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.PGblOYNBlxNXkD0vPe', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE level = 1, status = 1;

SELECT id, username, phone, level FROM users WHERE username = 'admin';
"

echo "管理员账号检查完成"
