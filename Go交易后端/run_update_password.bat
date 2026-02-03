@echo off
cd /d "d:\工程\交易所源码\OKCoinsgp交易所源码带教程\混合架构版本\1-Go交易后端"
echo 正在编译Go后端...
"C:\Program Files\Go\bin\go.exe" build -o api-new.exe cmd/api/main.go
if %ERRORLEVEL% EQU 0 (
    echo 编译成功！
    echo 正在启动服务...
    api-new.exe
) else (
    echo 编译失败！
)
pause
