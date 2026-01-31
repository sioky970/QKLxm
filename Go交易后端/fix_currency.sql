-- 修复现有订单的currency字段
-- 根据上下文，这些订单应该是LTC订单（ID=6）

UPDATE transaction 
SET currency = 6 
WHERE id IN (3, 4) AND currency = 0;

-- 查询修复后的结果
SELECT id, from_user_id, currency, legal_id, type, price, number, status 
FROM transaction 
ORDER BY id DESC 
LIMIT 5;
