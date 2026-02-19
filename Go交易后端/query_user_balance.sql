-- 查询用户13333333333的ID
SELECT id, account_number, phone, email FROM users WHERE account_number = '13333333333' OR phone = '13333333333';

-- 查询用户13333333333的所有账户余额
SELECT 
    ua.id,
    ua.user_id,
    ua.wallet_type,
    ua.usdt_balance,
    ua.usdt_locked,
    ua.currency_balances,
    ua.currency_locked,
    ua.delivery_margin,
    ua.delivery_pnl,
    ua.total_value_usdt,
    u.account_number
FROM user_assets ua
LEFT JOIN users u ON ua.user_id = u.id
WHERE ua.user_id IN (
    SELECT id FROM users WHERE account_number = '13333333333' OR phone = '13333333333'
)
ORDER BY ua.wallet_type;
