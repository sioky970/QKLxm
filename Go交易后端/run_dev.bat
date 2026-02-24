@echo off
REM 使用Air热重载启动交易所后端
REM 使用方法: 直接双击运行，或在命令行中执行 .\run_dev.bat

echo ========================================
echo 启动Go交易所后端 - 热重载模式
echo ========================================
echo.
echo 提示：
echo   - 修改代码后会自动重新编译和重启
echo   - 按 Ctrl+C 可以停止服务
echo   - 日志文件保存在 logs/ 目录
echo   - 构建错误保存在 tmp/build-errors.log
echo.
echo ========================================
echo.

REM 确保tmp目录存在
if not exist tmp mkdir tmp

REM 启动Air
air -c .air.toml

pause
