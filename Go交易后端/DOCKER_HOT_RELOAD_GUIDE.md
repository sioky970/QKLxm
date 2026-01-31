# Docker容器热重载配置指南

## 概述

本文档说明如何在Docker容器中配置热重载，实现修改代码后自动重新编译和重启。

## 工作原理

```
宿主机代码目录 --volume挂载--> 容器内/app目录
                                   ↓
                              容器内Air监听
                                   ↓
                            检测到文件变化
                                   ↓
                            自动编译+重启
```

## 快速开始

### 方式1：使用启动脚本（推荐）
```bash
# Windows
.\run_docker_dev.bat

# 首次启动会下载镜像（约2-3分钟）
# 后续启动约30秒
```

### 方式2：直接使用docker-compose
```bash
# 启动所有服务（带热重载）
docker-compose -f docker-compose.dev.yml up --build

# 后台运行
docker-compose -f docker-compose.dev.yml up -d

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f api

# 停止所有服务
docker-compose -f docker-compose.dev.yml down
```

## 服务访问

| 服务 | 地址 | 说明 |
|-----|------|------|
| API | http://localhost:8080 | RESTful API |
| WebSocket | ws://localhost:8081 | 实时行情推送 |
| MySQL | localhost:3306 | 数据库 |
| Redis | localhost:6379 | 缓存 |

## 配置文件说明

### Dockerfile.dev
```dockerfile
FROM golang:1.21-alpine

# 安装Air（热重载工具）
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# 利用Docker缓存层
COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080 8081

# 启动Air监听
CMD ["air", "-c", ".air.toml"]
```

### docker-compose.dev.yml
```yaml
services:
  api:
    build:
      dockerfile: Dockerfile.dev
    volumes:
      # ✅ 关键：挂载代码目录
      - .:/app
      # ❌ 排除这些目录（避免权限问题）
      - /app/vendor
      - /app/tmp
      # ✅ 缓存Go modules（加速编译）
      - go-modules:/go/pkg/mod
    environment:
      # 容器内使用服务名连接
      - DB_HOST=mysql
      - REDIS_HOST=redis
```

## 开发流程

### 1. 启动开发环境
```bash
.\run_docker_dev.bat

# 等待所有服务就绪
✓ MySQL健康检查通过
✓ Redis健康检查通过
✓ API服务启动成功
```

### 2. 修改代码
```go
// 在宿主机编辑文件
// internal/api/handler/user_handler.go

func (h *UserHandler) GetProfile(c *gin.Context) {
    // 添加日志
    logger.Info("获取用户信息")  // <- 新增这行
    // ...
}

// 保存文件（Ctrl+S）
```

### 3. 自动重载
```
容器内Air检测到变化：
[cyan][watcher] 检测到文件变化: internal/api/handler/user_handler.go
[yellow][build] 开始构建...
[green][runner] 重启服务...
[white][app] ✓ 服务启动成功
```

### 4. 测试新代码
```bash
# 新的代码立即生效
curl http://localhost:8080/api/user/profile
```

## 性能对比

| 操作 | 本地开发 | Docker开发（热重载） | Docker生产 |
|-----|---------|-------------------|----------|
| 首次启动 | 15秒 | 30秒 | 10秒 |
| 代码修改重载 | 12秒 | 15-20秒 | 需重新构建 |
| 资源占用 | 低 | 中 | 低 |
| 环境一致性 | ❌ | ✅ | ✅ |

## 优势

### ✅ 环境一致性
```
开发人员A: Windows + MySQL 8.0
开发人员B: macOS + MySQL 5.7
测试环境: Linux + MySQL 8.0

使用Docker后 → 所有人环境完全一致
```

### ✅ 依赖隔离
```
不同项目使用不同的Go版本、MySQL版本
无需担心版本冲突
```

### ✅ 快速部署
```bash
# 新成员加入团队
git clone <repository>
cd project
.\run_docker_dev.bat

# 3分钟后环境就绪
```

### ✅ 代码热重载
```
保存代码 → 自动编译 → 自动重启
无需手动操作
```

## 劣势

### ⚠️ 首次启动慢
```
下载镜像：1-2分钟
构建镜像：1分钟
启动服务：30秒
总计：2-3分钟

解决：首次启动后，后续只需30秒
```

### ⚠️ 重载稍慢
```
本地：12秒
容器：15-20秒（多了文件监听和容器通信开销）

影响：可接受，开发体验仍然流畅
```

### ⚠️ 资源占用
```
Docker Desktop: 2GB基础开销
MySQL容器: 400MB
Redis容器: 50MB
API容器: 300MB（包含Go编译器）
总计: 约2.75GB

适合: 16GB内存以上的开发机
不适合: 8GB以下的机器（建议本地开发）
```

## 常见问题

### Q1: 修改代码后没有自动重载？

**检查清单**：
```bash
# 1. 确认volume挂载正确
docker-compose -f docker-compose.dev.yml exec api ls -la /app
# 应该能看到宿主机的所有文件

# 2. 查看Air日志
docker-compose -f docker-compose.dev.yml logs -f api

# 3. 确认.air.toml配置
docker-compose -f docker-compose.dev.yml exec api cat .air.toml
```

### Q2: 编译错误如何查看？

```bash
# 查看容器日志
docker-compose -f docker-compose.dev.yml logs api

# 进入容器查看
docker-compose -f docker-compose.dev.yml exec api sh
cat /app/tmp/build-errors.log
```

### Q3: 如何调试？

**方式1：日志调试**
```go
// 代码中添加日志
logger.Debug("变量值: %+v", someVar)

// 容器日志实时显示
docker-compose logs -f api
```

**方式2：远程调试（高级）**
```yaml
# docker-compose.dev.yml
api:
  ports:
    - "2345:2345"  # Delve调试端口
  security_opt:
    - "apparmor=unconfined"
  cap_add:
    - SYS_PTRACE
  command: dlv debug --headless --listen=:2345 --api-version=2 ./cmd/api
```

### Q4: 如何重置环境？

```bash
# 停止并删除所有容器和数据
docker-compose -f docker-compose.dev.yml down -v

# 重新启动（全新环境）
docker-compose -f docker-compose.dev.yml up --build
```

### Q5: 容器内如何连接数据库？

```yaml
# 容器内使用服务名
environment:
  - DB_HOST=mysql      # 不是127.0.0.1
  - REDIS_HOST=redis   # 不是localhost
```

```go
// config/config.go需要支持环境变量
viper.BindEnv("database.host", "DB_HOST")
viper.BindEnv("redis.host", "REDIS_HOST")
```

### Q6: 如何查看容器资源占用？

```bash
# 实时监控
docker stats exchange-api-dev

# 查看网络
docker network inspect exchange-network
```

## 对比其他方案

### 方案A：纯本地开发（当前方式）
```
✅ 编译最快（10秒）
✅ 资源占用低
❌ 环境不一致
❌ 依赖管理麻烦
推荐：个人开发
```

### 方案B：Docker热重载（本方案）
```
✅ 环境完全一致
✅ 依赖隔离
✅ 快速部署
⚠️ 重载稍慢（15秒）
⚠️ 资源占用中等
推荐：团队协作开发
```

### 方案C：Docker生产镜像
```
✅ 镜像小（15MB）
✅ 启动快（5秒）
❌ 不支持热重载
❌ 每次修改需重新构建
推荐：生产部署
```

## 最佳实践

### 1. 开发模式选择矩阵

| 场景 | 推荐方案 |
|-----|---------|
| 个人开发，配置简单 | 纯本地 + Air |
| 团队协作，环境复杂 | Docker热重载 |
| CI/CD自动化测试 | Docker生产镜像 |
| 生产部署 | Kubernetes + Docker |

### 2. 混合方案（推荐）

```yaml
# 日常开发
本地运行：.\run_dev.bat（Air热重载）
Redis/MySQL：Docker容器

# 集成测试
全部Docker：.\run_docker_dev.bat

# 生产部署
Docker镜像：docker-compose.prod.yml
```

### 3. 文件监听优化

```toml
# .air.toml
[build]
  # 只监听关键目录
  include_dir = ["cmd", "internal", "config"]
  
  # 排除频繁变化的目录
  exclude_dir = ["tmp", "logs", "uploads", ".git"]
  
  # 延迟1秒，合并多次保存
  delay = 1000
```

## 交易所项目特殊考虑

### ⚠️ WebSocket连接
```
热重载会断开所有WebSocket连接
客户端需要实现自动重连机制

建议：
- 开发时使用短连接测试
- 生产使用Nginx负载均衡 + 零停机部署
```

### ⚠️ 调度器任务
```go
// 热重载时调度器会重启
// 确保任务幂等性
func (s *Scheduler) Process() {
    // ✅ 使用分布式锁（Redis）
    locked := redis.SetNX("lock:scheduler", "1", 60*time.Second)
    if !locked {
        return  // 其他实例正在运行
    }
    defer redis.Del("lock:scheduler")
    
    // 处理任务...
}
```

### ⚠️ 进行中的交易
```
热重载建议：
1. 等待交易完成（监控pending订单数量）
2. 或使用灰度发布（多实例轮流重启）
3. 数据库事务确保一致性
```

## 总结

**Docker热重载适合**：
- ✅ 团队协作开发
- ✅ 环境一致性要求高
- ✅ 快速搭建开发环境
- ✅ 16GB+内存的开发机

**不适合**：
- ❌ 个人开发且环境简单
- ❌ 8GB以下内存的机器
- ❌ 需要极致编译速度
- ❌ 频繁调试性能问题

**当前项目建议**：
保持**混合方案**最灵活：
- 日常开发：本地运行 + Air（快速）
- 集成测试：Docker热重载（一致性）
- 生产部署：Docker镜像（稳定）
