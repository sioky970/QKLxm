# 热重载开发指南

## 快速开始

### 方式1：使用批处理脚本（推荐）
```bash
# Windows
.\run_dev.bat

# 或直接双击 run_dev.bat 文件
```

### 方式2：直接使用Air命令
```bash
air
# 或指定配置文件
air -c .air.toml
```

## 工作原理

Air会监听项目中的Go和YAML文件变化，一旦检测到修改：
1. 自动编译代码（保存到 `tmp/exchange-api.exe`）
2. 停止旧进程
3. 启动新编译的程序
4. 输出彩色日志便于观察

## 监听范围

### ✅ 会触发重新编译的文件
- `*.go` - 所有Go源代码文件
- `*.yaml` - 配置文件（如 `config/config.yaml`）

### ❌ 被排除的目录/文件
- `tmp/` - 临时构建文件
- `vendor/` - 依赖包
- `logs/` - 日志文件
- `uploads/` - 上传文件
- `*_test.go` - 测试文件
- `*.log` - 日志文件
- `*.md` - 文档文件

## 配置说明

热重载配置文件：`.air.toml`

关键配置项：
```toml
[build]
  # 编译命令
  cmd = "go build -o ./tmp/exchange-api.exe ./cmd/api"
  
  # 运行参数
  args_bin = ["-c", "config/config.yaml"]
  
  # 延迟1秒避免频繁重启
  delay = 1000
  
  # 监听扩展名
  include_ext = ["go", "yaml"]
```

## 日志输出

- **控制台**：实时显示彩色日志
  - 紫色(magenta)：主日志
  - 青色(cyan)：文件监听器
  - 黄色(yellow)：构建过程
  - 绿色(green)：运行状态
  - 白色(white)：应用输出

- **构建错误**：`tmp/build-errors.log`
- **应用日志**：`logs/app.log`

## 常见问题

### Q1: Air命令找不到？
```bash
# 安装Air
go install github.com/air-verse/air@latest

# 确保 %GOPATH%\bin 在环境变量PATH中
# 默认路径: %USERPROFILE%\go\bin
```

### Q2: 修改代码后没有自动重启？
- 检查文件扩展名是否在 `include_ext` 中
- 检查文件/目录是否在排除列表中
- 查看控制台是否有错误提示

### Q3: 频繁重启怎么办？
- 调整 `delay` 参数（默认1000ms）
- 检查是否有程序自动修改文件（如IDE格式化）

### Q4: 端口被占用？
```bash
# 停止旧进程
Get-Process | Where-Object {$_.ProcessName -eq "exchange-api"} | Stop-Process -Force

# 或使用Ctrl+C停止Air
```

## 性能对比

| 操作方式 | 启动时间 | 修改后重启 |
|---------|---------|-----------|
| 手动 `go run` | 15-20秒 | 15-20秒 |
| 手动 `go build` | 10-15秒 | 10-15秒 |
| **Air热重载** | **15-20秒** | **12-18秒** |

优势：
- ✅ 无需手动停止/启动
- ✅ 保存即重载，开发流畅
- ✅ 彩色日志，错误醒目
- ✅ 构建错误自动捕获

## 与Docker结合使用

见后续文档：[DOCKER_HOT_RELOAD_GUIDE.md](./DOCKER_HOT_RELOAD_GUIDE.md)

---

## 快速参考卡片

### 方案选择决策树

```
开始
  ↓
是团队开发？
  ├─ 是 → 环境复杂？
  │        ├─ 是 → Docker热重载
  │        └─ 否 → 本地Air热重载
  └─ 否 → 个人开发
           ├─ 机器配置 ≥16GB → 任意方案
           ├─ 机器配置 8-16GB → 本地Air
           └─ 机器配置 <8GB → 手动编译
```

### 方案对比矩阵

| 特性 | 手动编译 | 本地Air | Docker热重载 |
|-----|---------|---------|-------------|
| **编译速度** | 10-15秒 | 12-18秒 | 15-25秒 |
| **资源占用** | 低(200MB) | 低(300MB) | 高(2.7GB) |
| **环境一致性** | ❌ | ❌ | ✅ |
| **学习成本** | 低 | 低 | 中 |
| **团队协作** | ❌ | ⚠️ | ✅ |
| **自动重载** | ❌ | ✅ | ✅ |
| **依赖隔离** | ❌ | ❌ | ✅ |

### 针对交易所项目的特别说明

#### ⚠️ 关键组件热重载影响

**WebSocket连接**
- 影响：所有连接会断开
- 解决：客户端实现心跳+自动重连
- 代码示例：
```javascript
// 前端WebSocket自动重连
let ws;
let reconnectTimer;

function connect() {
    ws = new WebSocket('ws://localhost:8081/ws');
    
    ws.onclose = () => {
        console.log('连接断开，3秒后重连...');
        clearTimeout(reconnectTimer);
        reconnectTimer = setTimeout(connect, 3000);
    };
    
    ws.onopen = () => {
        console.log('WebSocket已连接');
        clearTimeout(reconnectTimer);
    };
}

connect();
```

**调度器任务**
- 影响：定时任务会重启
- 已修复：使用atomic确保优雅停止
- 建议：生产环境使用Redis分布式锁

**数据库连接**
- 影响：连接池会重建
- 已优化：SetConnMaxIdleTime配置生效
- 注意：未提交事务会回滚（正确行为）

**Redis连接**
- 影响：连接会重建
- 已优化：超时配置确保快速恢复
- 持久化数据不受影响

### 最佳实践建议

**当前项目推荐配置**：
```bash
# 开发环境（快速迭代）
方案：本地Air热重载
启动：.\run_dev.bat
优势：最快的编译速度
适用：日常功能开发

# 测试环境（集成测试）
方案：Docker热重载
启动：.\run_docker_dev.bat
优势：环境完全一致
适用：提交前的完整测试

# 生产环境（线上部署）
方案：Docker镜像
部署：docker-compose.prod.yml
优势：稳定、轻量、快速
适用：正式发布
```

### 开发工作流建议

```
1. 日常开发
   ├─ 使用本地Air热重载
   ├─ 快速迭代功能
   └─ 实时查看效果

2. 提交前测试
   ├─ 切换到Docker环境
   ├─ 完整集成测试
   └─ 确保无环境差异

3. 代码审查
   ├─ Git提交代码
   ├─ CI/CD自动构建
   └─ Docker镜像测试

4. 生产部署
   ├─ 使用生产Docker镜像
   ├─ 滚动更新
   └─ 零停机部署
```
