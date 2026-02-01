-- 插入一个测试用户，密码为 "test123"
INSERT INTO users (phone, password, status, create_time, update_time) 
VALUES ('13900000001', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 获取刚插入的用户ID
SELECT LAST_INSERT_ID() as user_id;