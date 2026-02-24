-- ============================================================
-- MySQL 5.7 数据库索引优化脚本
-- 交易所系统专用
-- 版本: MySQL 5.7.x
-- ============================================================
-- 执行前请务必备份数据库！
-- 建议在低峰期执行，执行前请阅读脚本中的说明
-- ============================================================

-- ------------------------------------------------------------
-- 1. 用户相关表索引优化
-- ------------------------------------------------------------

-- users 用户表
-- 状态索引（认证审核、登录验证等）
ALTER TABLE users ADD INDEX idx_users_status (status);

-- 手机号索引（短信登录、找回密码）
ALTER TABLE users ADD INDEX idx_users_phone (phone);

-- 邮箱索引（邮箱登录）
ALTER TABLE users ADD INDEX idx_users_email (email);

-- 用户邀请码索引
ALTER TABLE users ADD INDEX idx_users_invite_code (invite_code);

-- 组合索引：状态+创建时间（后台审核列表）
ALTER TABLE users ADD INDEX idx_users_status_time (status, created_time);

-- 用户扩展信息表索引
ALTER TABLE users_additional ADD INDEX idx_users_add_user_id (user_id);


-- ------------------------------------------------------------
-- 2. 交易订单表索引优化
-- ------------------------------------------------------------

-- transaction 交易订单表

-- 核心查询索引：用户ID + 状态（订单列表）
ALTER TABLE transaction ADD INDEX idx_trans_from_user_status (from_user_id, status);

-- 核心查询索引：用户ID + 状态 + 创建时间（高级筛选）
ALTER TABLE transaction ADD INDEX idx_trans_from_user_status_time (from_user_id, status, create_time);

-- 对方用户ID索引（查看与某人的交易记录）
ALTER TABLE transaction ADD INDEX idx_trans_to_user (to_user_id);

-- 币种+法币索引（交易对筛选）
ALTER TABLE transaction ADD INDEX idx_trans_currency_legal (currency, legal);

-- 支付方式索引
ALTER TABLE transaction ADD INDEX idx_trans_pay_method (pay_method);

-- 状态+创建时间索引（后台订单管理）
ALTER TABLE transaction ADD INDEX idx_trans_status_time (status, create_time);

-- 组合索引：法币+状态（按交易对查看订单）
ALTER TABLE transaction ADD INDEX idx_trans_legal_status (legal, status);

-- 批量订单号查询优化
ALTER TABLE transaction ADD INDEX idx_trans_order_no (order_no);


-- ------------------------------------------------------------
-- 3. 合约交易表索引优化
-- ------------------------------------------------------------

-- lever_transaction 合约交易订单表

-- 用户ID索引（个人合约记录）
ALTER TABLE lever_transaction ADD INDEX idx_lever_user_id (user_id);

-- 用户+状态索引（合约列表）
ALTER TABLE lever_transaction ADD INDEX idx_lever_user_status (user_id, status);

-- 用户+状态+时间索引（历史记录分页）
ALTER TABLE lever_transaction ADD INDEX idx_lever_user_status_time (user_id, status, create_time DESC);

-- 订单号索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_order_no (order_no);

-- 结算状态索引（结算任务查询）
ALTER TABLE lever_transaction ADD INDEX idx_lever_settled (is_settled, settle_time);

-- 币种+法币索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_currency_legal (currency, legal);

-- 持仓方向索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_direction (direction);

-- 杠杆倍数索引
ALTER TABLE lever_transaction ADD INDEX idx_lever_lever (lever);


-- ------------------------------------------------------------
-- 4. 秒合约订单表索引优化
-- ------------------------------------------------------------

-- micro_order 秒合约订单表

-- 用户ID索引（秒合约记录）
ALTER TABLE micro_order ADD INDEX idx_micro_user_id (user_id);

-- 用户+状态索引（秒合约列表）
ALTER TABLE micro_order ADD INDEX idx_micro_user_status (user_id, status);

-- 用户+状态+时间索引
ALTER TABLE micro_order ADD INDEX idx_micro_user_status_time (user_id, status, create_time DESC);

-- 币种索引
ALTER TABLE micro_order ADD INDEX idx_micro_currency (currency);

-- 订单号索引
ALTER TABLE micro_order ADD INDEX idx_micro_order_no (order_no);

-- 玩法类型索引
ALTER TABLE micro_order ADD INDEX idx_micro_mold (mold);

-- 投注方向索引
ALTER TABLE micro_order ADD INDEX idx_micro_direction (direction);


-- ------------------------------------------------------------
-- 5. 钱包资产表索引优化
-- ------------------------------------------------------------

-- users_wallet 用户钱包表

-- 用户+币种唯一索引（保证每个用户每种币种只有一个钱包）
ALTER TABLE users_wallet ADD UNIQUE INDEX idx_wallet_user_currency (user_id, currency);

-- 状态索引（钱包启用状态）
ALTER TABLE users_wallet ADD INDEX idx_wallet_status (status);

-- 地址索引（归集查询）
ALTER TABLE users_wallet ADD INDEX idx_wallet_address (address);

-- 合约地址索引
ALTER TABLE users_wallet ADD INDEX idx_wallet_contract (contract_address);


-- ------------------------------------------------------------
-- 6. 充值提现表索引优化
-- ------------------------------------------------------------

-- deposit_address 充值地址表

-- 地址+网络索引
ALTER TABLE deposit_address ADD INDEX idx_dep_addr_network (address, network);

-- 用户+状态索引
ALTER TABLE deposit_address ADD INDEX idx_dep_addr_user_status (user_id, status);


-- deposit_order 充值订单表

-- 用户ID索引
ALTER TABLE deposit_order ADD INDEX idx_dep_order_user_id (user_id);

-- 状态索引
ALTER TABLE deposit_order ADD INDEX idx_dep_order_status (status);

-- 地址索引
ALTER TABLE deposit_order ADD INDEX idx_dep_order_address (address);

-- 用户+状态+时间索引
ALTER TABLE deposit_order ADD INDEX idx_dep_order_user_status_time (user_id, status, create_time DESC);

-- 币种索引
ALTER TABLE deposit_order ADD INDEX idx_dep_order_currency (currency);

-- TXID索引
ALTER TABLE deposit_order ADD INDEX idx_dep_order_txid (tx_hash);


-- users_wallet_out 提现订单表

-- 用户ID索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_user_id (user_id);

-- 状态索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_status (status);

-- 用户+状态+时间索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_user_status_time (user_id, status, create_time DESC);

-- 审核人索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_examine (examine);

-- 币种索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_currency (currency);

-- TXID索引
ALTER TABLE users_wallet_out ADD INDEX idx_withdraw_txid (tx_hash);


-- ------------------------------------------------------------
-- 7. 账户流水表索引优化
-- ------------------------------------------------------------

-- account_log 账户流水表

-- 用户ID索引
ALTER TABLE account_log ADD INDEX idx_acc_log_user_id (user_id);

-- 类型索引（资金流动分类）
ALTER TABLE account_log ADD INDEX idx_acc_log_type (type);

-- 币种索引
ALTER TABLE account_log ADD INDEX idx_acc_log_currency (currency);

-- 用户+时间索引（账户历史）
ALTER TABLE account_log ADD INDEX idx_acc_log_user_time (user_id, created_time DESC);

-- 用户+类型索引（特定类型流水）
ALTER TABLE account_log ADD INDEX idx_acc_log_user_type (user_id, type);

-- 创建时间索引（按时间范围查询）
ALTER TABLE account_log ADD INDEX idx_acc_log_created_time (created_time);


-- ------------------------------------------------------------
-- 8. 管理员与权限表索引优化
-- ------------------------------------------------------------

-- admin 管理员表

-- 状态索引
ALTER TABLE admin ADD INDEX idx_admin_status (status);

-- 用户名索引
ALTER TABLE admin ADD INDEX idx_admin_username (username);


-- admin_role 管理员角色表

-- 角色标识索引
ALTER TABLE admin_role ADD INDEX idx_admin_role_mark (mark);


-- role 角色表

-- 角色状态索引
ALTER TABLE role ADD INDEX idx_role_status (status);


-- access 权限表

-- 父级ID索引
ALTER TABLE access ADD INDEX idx_access_pid (pid);

-- 模块标识索引
ALTER TABLE access ADD INDEX idx_access_module (module);


-- ------------------------------------------------------------
-- 9. KYC认证表索引优化
-- ------------------------------------------------------------

-- kyc_verification KYC认证表

-- 用户ID唯一索引
ALTER TABLE kyc_verification ADD UNIQUE INDEX idx_kyc_user_id (user_id);

-- 状态索引
ALTER TABLE kyc_verification ADD INDEX idx_kyc_status (status);

-- 审核人索引
ALTER TABLE kyc_verification ADD INDEX idx_kyc_examine (examine);

-- 创建时间索引
ALTER TABLE kyc_verification ADD INDEX idx_kyc_create_time (create_time DESC);

-- 证件类型索引
ALTER TABLE kyc_verification ADD INDEX idx_kyc_id_type (id_type);


-- ------------------------------------------------------------
-- 10. 币种与交易对表索引优化
-- ------------------------------------------------------------

-- currency 币种表

-- 状态索引
ALTER TABLE currency ADD INDEX idx_currency_status (status);

-- 排序索引
ALTER TABLE currency ADD INDEX idx_currency_sort (sort);

-- 标识索引
ALTER TABLE currency ADD INDEX idx_currency_mark (mark);


-- currency_match 交易对配置表

-- 法币+币种唯一索引
ALTER TABLE currency_match ADD UNIQUE INDEX idx_match_legal_currency (legal, currency);

-- 状态索引
ALTER TABLE currency_match ADD INDEX idx_match_status (status);

-- 排序索引
ALTER TABLE currency_match ADD INDEX idx_match_sort (sort);


-- ------------------------------------------------------------
-- 11. 杠杆配置表索引优化
-- ------------------------------------------------------------

-- lever_config 杠杆配置表

-- 币种索引
ALTER TABLE lever_config ADD INDEX idx_lever_config_currency (currency);

-- 状态索引
ALTER TABLE lever_config ADD INDEX idx_lever_config_status (status);


-- ------------------------------------------------------------
-- 12. 活动配置表索引优化
-- ------------------------------------------------------------

-- activity_config 活动配置表

-- 状态索引
ALTER TABLE activity_config ADD INDEX idx_activity_status (status);

-- 时间范围索引
ALTER TABLE activity_config ADD INDEX idx_activity_time (start_time, end_time);

-- 类型索引
ALTER TABLE activity_config ADD INDEX idx_activity_type (type);


-- ------------------------------------------------------------
-- 13. 自动撤单配置表索引优化
-- ------------------------------------------------------------

-- revoke_config 自动撤单配置表

-- 用户ID索引
ALTER TABLE revoke_config ADD INDEX idx_revoke_user_id (user_id);

-- 状态索引
ALTER TABLE revoke_config ADD INDEX idx_revoke_status (status);


-- ------------------------------------------------------------
-- 14. 统计数据表索引优化
-- ------------------------------------------------------------

-- statistics 统计数据表

-- 日期索引
ALTER TABLE statistics ADD UNIQUE INDEX idx_stats_date (date);

-- 类型索引
ALTER TABLE statistics ADD INDEX idx_stats_type (type);


-- ------------------------------------------------------------
-- 15. 白名单地址表索引优化
-- ------------------------------------------------------------

-- whitelist_address 白名单地址表

-- 地址唯一索引
ALTER TABLE whitelist_address ADD UNIQUE INDEX idx_whitelist_address (address, network);

-- 状态索引
ALTER TABLE whitelist_address ADD INDEX idx_whitelist_status (status);


-- ------------------------------------------------------------
-- 16. 验证日志表索引优化
-- ------------------------------------------------------------

-- verify_log 验证日志表

-- 用户ID索引
ALTER TABLE verify_log ADD INDEX idx_verify_log_user_id (user_id);

-- 类型索引
ALTER TABLE verify_log ADD INDEX idx_verify_log_type (type);

-- 创建时间索引
ALTER TABLE verify_log ADD INDEX idx_verify_log_created_time (created_time DESC);


-- ------------------------------------------------------------
-- 17. 通知公告表索引优化
-- ------------------------------------------------------------

-- notice 通知公告表

-- 状态索引
ALTER TABLE notice ADD INDEX idx_notice_status (status);

-- 类型索引
ALTER TABLE notice ADD INDEX idx_notice_type (type);

-- 创建时间索引
ALTER TABLE notice ADD INDEX idx_notice_create_time (create_time DESC);


-- ------------------------------------------------------------
-- 18. 常见问题表索引优化
-- ------------------------------------------------------------

-- faq 常见问题表

-- 分类索引
ALTER TABLE faq ADD INDEX idx_faq_category (category);

-- 状态索引
ALTER TABLE faq ADD INDEX idx_faq_status (status);


-- ------------------------------------------------------------
-- 19. 合作伙伴表索引优化
-- ------------------------------------------------------------

-- partner 合作伙伴表

-- 状态索引
ALTER TABLE partner ADD INDEX idx_partner_status (status);

-- 排序索引
ALTER TABLE partner ADD INDEX idx_partner_sort (sort);


-- ------------------------------------------------------------
-- 20. 短信验证码表索引优化
-- ------------------------------------------------------------

-- sms_code 短信验证码表

-- 手机号+类型索引
ALTER TABLE sms_code ADD INDEX idx_sms_phone_type (phone, type);

-- 创建时间索引（过期清理）
ALTER TABLE sms_code ADD INDEX idx_sms_create_time (created_time);


-- ------------------------------------------------------------
-- 索引验证查询
-- ------------------------------------------------------------

-- 查看指定表的索引
-- SHOW INDEX FROM transaction;

-- 查看索引使用情况
-- SELECT * FROM mysql.innodb_index_stats WHERE table_name = 'transaction';

-- 查看慢查询（优化后监测）
-- SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 100;


-- ------------------------------------------------------------
-- 注意事项
-- ------------------------------------------------------------

-- 1. 执行前请备份数据库
-- 2. 建议在低峰期执行，大表索引创建可能需要较长时间
-- 3. 如果表数据量较大，建议使用以下方式避免锁表：
--    ALTER TABLE table_name ADD INDEX idx_xxx (column) ALGORITHM=INPLACE, LOCK=NONE;

-- 4. 检查索引创建进度（执行长时间索引创建时）
--    SHOW PROCESSLIST;
--    -- 查看State列是否显示 "Finished reading altred table"

-- 5. 回滚：如果需要删除索引
--    ALTER TABLE table_name DROP INDEX idx_xxx;
