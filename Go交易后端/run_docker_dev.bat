@echo off
REM Docker开发环境启动脚本 - 支持热重载
REM 使用方法: .\run_docker_dev.bat

echo ========================================
echo 启动Docker开发环境 - 热重载模式
echo ========================================
echo.
echo 服务列表：
echo   - MySQL:     localhost:3306
echo   - Redis:     localhost:6379
echo   - API:       http://localhost:8080
echo   - WebSocket: ws://localhost:8081
echo.
echo 提示：
echo   - 修改代码后会自动重新编译（容器内Air监听）
echo   - 首次启动需要下载镜像（约2-3分钟）
echo   - 后续启动更快（约30秒）
echo   - 按 Ctrl+C 停止所有服务
echo.
echo ========================================
echo.

REM 确保Docker正在运行
docker info >nul 2>&1
if errorlevel 1 (
    echo [错误] Docker未运行，请先启动Docker Desktop
    pause
    exit /b 1
)

REM 启动所有服务
echo [启动] 正在启动开发环境...
docker-compose -f docker-compose.dev.yml up --build

pause
