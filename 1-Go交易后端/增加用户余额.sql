-- 为用户 14444444444 增加 1000000 USDT 余额
-- 执行时间: 2026-01-26

-- ==========================================
-- 第一步：查询用户信息
-- ==========================================
SELECT '查询用户信息' as step;
SELECT id, account_number, phone, status 
FROM users 
WHERE account_number='14444444444' OR phone='14444444444'
LIMIT 1;

-- ==========================================
-- 第二步：查询当前钱包余额
-- ==========================================
SELECT '查询USDT钱包余额' as step;
SELECT 
    w.id as wallet_id,
    w.user_id,
    w.currency,
    w.usdt_balance as current_balance,
    w.lock_usdt_balance,
    u.account_number
FROM users_wallet w
INNER JOIN users u ON w.user_id = u.id
WHERE (u.account_number='14444444444' OR u.phone='14444444444')
  AND w.currency = 3
LIMIT 1;

-- ==========================================
-- 第三步：更新余额（增加 1000000 USDT）
-- ==========================================
SELECT '开始更新余额' as step;

-- 使用事务保证数据一致性
START TRANSACTION;

-- 更新 USDT 余额
UPDATE users_wallet w
INNER JOIN users u ON w.user_id = u.id
SET w.usdt_balance = w.usdt_balance + 1000000
WHERE (u.account_number='14444444444' OR u.phone='14444444444')
  AND w.currency = 3;

-- 记录账户日志
INSERT INTO account_log (
    user_id,
    value,
    currency,
    type,
    info,
    created_time
)
SELECT 
    u.id,
    1000000,
    3,
    101,  -- 类型101表示余额增加
    '后台充值 - 管理员手动添加',
    UNIX_TIMESTAMP()
FROM users u
WHERE (u.account_number='14444444444' OR u.phone='14444444444')
LIMIT 1;

-- 提交事务
COMMIT;

SELECT '余额更新成功' as result;

-- ==========================================
-- 第四步：验证结果
-- ==========================================
SELECT '验证更新后的余额' as step;
SELECT 
    w.id as wallet_id,
    w.user_id,
    w.usdt_balance as new_balance,
    w.lock_usdt_balance,
    u.account_number,
    '增加了 1000000 USDT' as operation
FROM users_wallet w
INNER JOIN users u ON w.user_id = u.id
WHERE (u.account_number='14444444444' OR u.phone='14444444444')
  AND w.currency = 3
LIMIT 1;

-- 查看账户日志
SELECT '查看账户日志' as step;
SELECT 
    al.id,
    al.user_id,
    al.value,
    al.currency,
    al.type,
    al.info,
    FROM_UNIXTIME(al.created_time) as created_at
FROM account_log al
INNER JOIN users u ON al.user_id = u.id
WHERE (u.account_number='14444444444' OR u.phone='14444444444')
  AND al.currency = 3
ORDER BY al.created_time DESC
LIMIT 5;

SELECT '操作完成！' as status;
