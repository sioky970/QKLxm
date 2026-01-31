@echo off
chcp 65001 >nul
echo ============================================================
echo MySQL数据库恢复脚本 - 交易所系统
echo ============================================================

rem 配置参数
set MYSQL_HOST=127.0.0.1
set MYSQL_PORT=3306
set MYSQL_USER=root
set MYSQL_PASSWORD=root123456
set DATABASE_NAME=bibi2022
set BACKUP_DIR=./backup

rem 检查参数
if "%~1"=="" (
    echo [ERROR] 请指定要恢复的备份文件！
    echo 用法: restore_database.bat ^<backup_file^>
    echo.
    echo 可用的备份文件:
    dir /b "%BACKUP_DIR%\*.sql.gz" 2>nul
    echo.
    echo 可用的SQL文件:
    dir /b "%BACKUP_DIR%\*.sql" 2>nul
    exit /b 1
)

set BACKUP_FILE=%~1

rem 检查文件是否存在
if not exist "%BACKUP_FILE%" (
    echo [ERROR] 备份文件不存在: %BACKUP_FILE%
    exit /b 1
)

echo ============================================================
echo 数据库恢复确认
echo ============================================================
echo 警告：此操作将覆盖当前数据库中的所有数据！
echo.
echo 目标数据库: %DATABASE_NAME%
echo 备份文件: %BACKUP_FILE%
echo.

set /p CONFIRM="确认执行恢复操作？(y/n): "
if not "%CONFIRM%"=="y" (
    echo [INFO] 已取消恢复操作
    exit /b 0
)

rem 解压备份文件（如果是gzip格式）
if "%BACKUP_FILE:~-3%"==".gz" (
    echo [INFO] 解压备份文件...
    gunzip -c "%BACKUP_FILE%" > "%BACKUP_FILE:.gz=.sql"
    set SQL_FILE="%BACKUP_FILE:.gz=.sql"
) else (
    set SQL_FILE="%BACKUP_FILE%"
)

rem 确认解压后的文件存在
if not exist %SQL_FILE% (
    echo [ERROR] SQL文件不存在: %SQL_FILE%
    exit /b 1
)

echo [INFO] 开始恢复数据库...
echo [INFO] 如果文件较大，此过程可能需要几分钟...

rem 执行数据库恢复
mysql -h%MYSQL_HOST% -P%MYSQL_PORT% -u%MYSQL_USER% -p%MYSQL_PASSWORD% %DATABASE_NAME% < %SQL_FILE% 2>&1

if %errorlevel% equ 0 (
    echo [SUCCESS] 数据库恢复成功！
    
    rem 清理临时文件
    if "%BACKUP_FILE:~-3%"==".gz" (
        del "%SQL_FILE%" 2>nul
    )
    
    echo [INFO] 数据库已恢复到备份时的状态
) else (
    echo [ERROR] 数据库恢复失败！
    echo [INFO] 请检查MySQL服务是否运行，以及数据库凭据是否正确
    exit /b 1
)

echo ============================================================
echo 恢复完成！
echo ============================================================
goto :eof

:eof
