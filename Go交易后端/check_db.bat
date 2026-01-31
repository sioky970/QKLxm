@echo off
echo 查询订单数据...
echo.

mysql -uroot -proot123 exchange -e "SELECT id, from_user_id, currency, legal_id, type, price, number, status FROM transaction ORDER BY id DESC LIMIT 5;"

echo.
echo 查询币种数据...
echo.

mysql -uroot -proot123 exchange -e "SELECT id, name FROM currency LIMIT 10;"

pause
