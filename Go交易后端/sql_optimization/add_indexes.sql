-- MySQL数据库索引优化脚本
-- 执行前请备份数据库，并在非业务高峰期执行
-- 此脚本为交易所系统添加缺失的索引，优化查询性能

-- ============================================================
-- 1. 用户相关表索引优化
-- ============================================================

-- 为用户表添加常用查询索引
-- 状态查询索引（用于查找特定状态的用户）
ALTER TABLE users ADD INDEX idx_users_status (status);

-- 手机号索引（用于登录验证）
ALTER TABLE users ADD INDEX idx_users_phone (phone);

-- 邮箱索引（用于登录验证）
ALTER TABLE users ADD INDEX idx_users_email (email);

-- 账户类型索引（区分普通用户和代理用户）
ALTER TABLE users ADD INDEX idx_users_account_type (account_type);

-- 注册时间索引（用于按时间查询用户）
ALTER TABLE users ADD INDEX idx_users_time (time);

-- ============================================================
-- 2. 交易订单表索引优化
-- ============================================================

-- Transaction 订单表索引优化
-- 状态索引（查询待处理订单）
ALTER TABLE transaction ADD INDEX idx_transaction_status (status);

-- 用户ID索引（查询用户订单）
ALTER TABLE transaction ADD INDEX idx_transaction_from_user (from_user_id);
ALTER TABLE transaction ADD INDEX idx_transaction_to_user (to_user_id);

-- 复合索引：用户+状态（最常见查询模式）
ALTER TABLE transaction ADD INDEX idx_transaction_user_status (from_user_id, status);
ALTER TABLE transaction ADD INDEX idx_transaction_user_time (from_user_id, create_time);

-- 币种和法币索引（交易对查询）
ALTER TABLE transaction ADD INDEX idx_transaction_currency_legal (currency, legal);

-- ============================================================
-- 3. 合约交易表索引优化
-- ============================================================

-- LeverTransaction 合约交易表索引优化
-- 用户索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_user_id (user_id);

-- 状态索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_status (status);

-- 复合索引：用户+状态
ALTER TABLE lever_transaction ADD INDEX idx_lever_user_status (user_id, status);

-- 创建时间索引（历史记录查询）
ALTER TABLE lever_transaction ADD INDEX idx_lever_create_time (create_time);

-- 币种和法币索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_currency_legal (currency, legal);

-- 结算状态索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_settled (settled);

-- ============================================================
-- 4. 秒合约交易表索引优化
-- ============================================================

-- MicroOrder 秒合约订单表索引优化
-- 用户索引
ALTER TABLE micro_order ADD INDEX idx_micro_user_id (user_id);

-- 状态索引
ALTER TABLE micro_order ADD INDEX idx_micro_status (status);

-- 复合索引：用户+状态
ALTER TABLE micro_order ADD INDEX idx_micro_user_status (user_id, status);

-- 创建时间索引
ALTER TABLE micro_order ADD INDEX idx_micro_create_time (create_time);

-- 币种索引
ALTER TABLE micro_order ADD INDEX idx_micro_currency (currency);

-- ============================================================
-- 5. 钱包与资产表索引优化
-- ============================================================

-- UsersWallet 钱包表索引优化
-- 用户+币种复合索引（最常见查询）
ALTER TABLE users_wallet ADD INDEX idx_wallet_user_currency (user_id, currency);

-- 状态索引
ALTER TABLE users_wallet ADD INDEX idx_wallet_status (status);

-- 地址索引（充值查询）
ALTER TABLE users_wallet ADD INDEX idx_wallet_address (address);

-- UsersAssets 资产表已有 user_id 唯一索引，检查是否需要其他索引
-- 创建时间索引（按时间排序）
ALTER TABLE user_assets ADD INDEX idx_assets_create_time (create_time);

-- ============================================================
-- 6. 充值提现表索引优化
-- ============================================================

-- 充值地址表索引
ALTER TABLE deposit_address ADD INDEX idx_deposit_address_network (network);
ALTER TABLE deposit_address ADD INDEX idx_deposit_address_status (status);

-- 充值订单表索引
ALTER TABLE deposit_order ADD INDEX idx_deposit_user_id (user_id);
ALTER TABLE deposit_order ADD INDEX idx_deposit_status (status);
ALTER TABLE deposit_order ADD INDEX idx_deposit_create_time (create_time);
ALTER TABLE deposit_order ADD INDEX idx_deposit_order_no (order_no);
ALTER TABLE deposit_order ADD INDEX idx_deposit_user_status (user_id, status);

-- 提现订单表索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_user_id (user_id);
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_status (status);
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_create_time (create_time);
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_order_no (order_no);
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_user_status (user_id, status);

-- ============================================================
-- 7. 账户流水表索引优化
-- ============================================================

-- AccountLog 账户流水表索引
-- 用户索引
ALTER TABLE account_log ADD INDEX idx_account_log_user_id (user_id);

-- 类型索引（按交易类型查询）
ALTER TABLE account_log ADD INDEX idx_account_log_type (type);

-- 币种索引
ALTER TABLE account_log ADD INDEX idx_account_log_currency (currency);

-- 复合索引：用户+时间（最常见的流水查询）
ALTER TABLE account_log ADD INDEX idx_account_log_user_time (user_id, created_time);

-- 复合索引：用户+类型
ALTER TABLE account_log ADD INDEX idx_account_log_user_type (user_id, type);

-- ============================================================
-- 8. KYC认证表索引优化
-- ============================================================

-- KYC认证表索引
ALTER TABLE user_real ADD INDEX idx_real_user_id (user_id);

-- 审核状态索引
ALTER TABLE user_real ADD INDEX idx_real_review_status (review_status);

-- 复合索引：用户+状态
ALTER TABLE user_real ADD INDEX idx_real_user_status (user_id, review_status);

-- 提交时间索引
ALTER TABLE user_real ADD INDEX idx_real_create_time (create_time);

-- ============================================================
-- 9. 新闻公告表索引优化
-- ============================================================

-- 新闻表索引
ALTER TABLE news ADD INDEX idx_news_type (type);
ALTER TABLE news ADD INDEX idx_news_status (status);
ALTER TABLE news ADD INDEX idx_news_create_time (create_time);

-- ============================================================
-- 10. 消息通知表索引优化
-- ============================================================

-- 用户消息表索引
ALTER TABLE user_message ADD INDEX idx_message_user_id (user_id);
ALTER TABLE user_message ADD INDEX idx_message_readed (readed);
ALTER TABLE user_message ADD INDEX idx_message_user_readed (user_id, readed);
ALTER TABLE user_message ADD INDEX idx_message_create_time (create_time);

-- ============================================================
-- 11. 法币交易表索引优化
-- ============================================================

-- 法币交易相关表索引
ALTER TABLE legal_deal ADD INDEX idx_legal_deal_user_id (user_id);
ALTER TABLE legal_deal ADD INDEX idx_legal_deal_status (status);
ALTER TABLE legal_deal ADD INDEX idx_legal_deal_create_time (create_time);

ALTER TABLE legal_deal_send ADD INDEX idx_legal_send_user_id (user_id);
ALTER TABLE legal_deal_send ADD INDEX idx_legal_send_status (status);
ALTER TABLE legal_deal_send ADD INDEX idx_legal_send_type (type);

-- ============================================================
-- 12. 交易对配置表索引优化
-- ============================================================

-- CurrencyMatch 交易对配置表索引
ALTER TABLE currency_matches ADD INDEX idx_match_legal_id (legal_id);
ALTER TABLE currency_matches ADD INDEX idx_match_currency_id (currency_id);
ALTER TABLE currency_matches ADD INDEX idx_match_display (is_display);

-- Currency 币种表索引
ALTER TABLE currency ADD INDEX idx_currency_display (is_display);
ALTER TABLE currency ADD INDEX idx_currency_type (type);
ALTER TABLE currency ADD INDEX idx_currency_legal (is_legal);
ALTER TABLE currency ADD INDEX idx_currency_lever (is_lever);

-- ============================================================
-- 创建索引说明
-- ============================================================
-- 注意：创建索引会短暂锁定表，建议在非业务高峰期执行
-- 对于大表，建议使用 ALGORITHM=INPLACE 和 LOCK=NONE 选项（如果MySQL版本支持）
-- 例如：ALTER TABLE transaction ADD INDEX idx_transaction_status (status), ALGORITHM=INPLACE, LOCK=NONE;

-- 执行完成后，建议运行 ANALYZE TABLE 更新统计信息
-- ANALYZE TABLE transaction, lever_transaction, micro_order, users_wallet, user_assets;
