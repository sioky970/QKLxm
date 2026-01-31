@echo off
chcp 65001 >nul
echo.
echo ╔═══════════════════════════════════════════════════════════╗
echo ║           Exchange Database Migration Tool                 ║
echo ╚═══════════════════════════════════════════════════════════╝
echo.

cd /d "%~dp0"

echo [INFO] 正在编译迁移工具...
go build -o migrate.exe ./cmd/migrate

if %errorlevel% neq 0 (
    echo [ERROR] 编译失败！
    pause
    exit /b 1
)

echo [INFO] 编译成功，开始执行迁移...
echo.

migrate.exe %*

echo.
echo [INFO] 迁移完成
pause
