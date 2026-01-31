@echo off
cd /d "%~dp0"
echo 正在添加order_value字段...
go run cmd/add_order_value/main.go
pause
