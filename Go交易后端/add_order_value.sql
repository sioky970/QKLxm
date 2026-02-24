-- 添加order_value字段
ALTER TABLE transaction 
ADD COLUMN order_value DECIMAL(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '订单价值(USDT)' 
AFTER deal_number;

-- 更新现有订单的order_value
UPDATE transaction 
SET order_value = price * number 
WHERE order_value = 0 AND price > 0;

-- 查询验证
SELECT id, from_user_id, currency, price, number, order_value, status 
FROM transaction 
ORDER BY id DESC 
LIMIT 5;
