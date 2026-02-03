@echo off
chcp 65001 >nul
echo ============================================================
echo MySQL数据库备份脚本 - 交易所系统
echo ============================================================

rem 配置参数
set BACKUP_DIR=./backup
set MYSQL_HOST=127.0.0.1
set MYSQL_PORT=3306
set MYSQL_USER=root
set MYSQL_PASSWORD=root123456
set DATABASE_NAME=bibi2022
set BACKUP_DATE=%date:~0,4%%date:~5,2%%date:~8,2%
set BACKUP_TIME=%time:~0,2%%time:~3,2%%time:~6,2%
set BACKUP_FILENAME=%BACKUP_DIR%\%DATABASE_NAME%_%BACKUP_DATE%_%BACKUP_TIME%

rem 创建备份目录
if not exist "%BACKUP_DIR%" mkdir "%BACKUP_DIR%"

echo [INFO] 开始备份数据库: %DATABASE_NAME%
echo [INFO] 备份时间: %BACKUP_DATE% %BACKUP_TIME%
echo [INFO] 备份路径: %BACKUP_FILENAME%.sql

rem 使用mysqldump进行备份
mysqldump -h%MYSQL_HOST% -P%MYSQL_PORT% -u%MYSQL_USER% -p%MYSQL_PASSWORD% ^
    --routines ^
    --triggers ^
    --events ^
    --single-transaction ^
    --master-data=2 ^
    --flush-logs ^
    %DATABASE_NAME% > "%BACKUP_FILENAME%.sql" 2>> "%BACKUP_DIR%\backup.log"

rem 检查备份结果
if %errorlevel% equ 0 (
    echo [SUCCESS] 数据库备份成功！
    
    rem 压缩备份文件（如果安装了gzip）
    gzip "%BACKUP_FILENAME%.sql"
    
    if exist "%BACKUP_FILENAME%.sql.gz" (
        echo [INFO] 备份文件已压缩: %BACKUP_FILENAME%.sql.gz
        for %%I in ("%BACKUP_FILENAME%.sql.gz") do set FILE_SIZE=%%~zI
        echo [INFO] 压缩后文件大小: %FILE_SIZE% bytes
    )
    
    rem 删除旧备份（保留最近7天）
    forfiles /p "%BACKUP_DIR%" /m *.sql.gz /d -7 /c "cmd /c del /q @path" 2>nul
    echo [INFO] 已清理7天前的旧备份文件
    
    rem 同步到远程存储（如果有配置）
    call :sync_to_remote "%BACKUP_FILENAME%.sql.gz"
) else (
    echo [ERROR] 数据库备份失败！请检查错误日志: %BACKUP_DIR%\backup.log
    exit /b 1
)

echo ============================================================
echo 备份完成！
echo ============================================================
goto :eof

:sync_to_remote
setlocal
set LOCAL_FILE=%~1
set REMOTE_PATH=backup@remote-server:/home/backup/

echo [INFO] 同步备份文件到远程存储...
pscp -l 5000 -pw %MYSQL_PASSWORD% "%LOCAL_FILE%" "%REMOTE_PATH%" 2>> "%BACKUP_DIR%\backup.log"
if %errorlevel% equ 0 (
    echo [SUCCESS] 文件已同步到远程存储
) else (
    echo [WARN] 远程同步失败，请手动检查
)
endlocal
goto :eof

:eof
