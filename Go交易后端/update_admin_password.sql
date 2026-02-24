-- 临时修改admin密码为PHP算法加密的密码
-- 这样Go后端(已修改为PHP兼容算法)就能验证通过

USE bibi2022;

-- 显示当前密码
SELECT id, username, password, '当前密码' as note FROM admin WHERE username = 'admin';

-- 更新为PHP MakePassword算法加密的'admin'
-- PHP算法: salt = 'ABCDEFG' + md5('a') + md5('d') + md5('m') + md5('i') + md5('n'), 然后 md5(salt)
UPDATE admin SET password = '979d7eed0b8835d92a0be361a630475b' WHERE username = 'admin';

-- 显示更新后的密码
SELECT id, username, password, '新密码(PHP兼容)' as note FROM admin WHERE username = 'admin';

-- 说明：
-- 旧密码: e3b6712b0207804be7d52c92d766e102 (未知来源)
-- 新密码: 979d7eed0b8835d92a0be361a630475b (PHP MakePassword('admin'))
