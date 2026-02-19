<?php
// 生成与PHP MakePassword算法兼容的admin密码
// PHP MakePassword算法: salt = 'ABCDEFG' + md5('a') + md5('d') + md5('m') + md5('i') + md5('n'), 然后 md5(alt)

$password = 'admin123'; // 你的密码
$salt = 'ABCDEFG';
$salt .= md5('a');
$salt .= md5('d');
$salt .= md5('m');
$salt .= md5('i');
$salt .= md5('n');

// 计算alt
$alt = '';
for ($i = 0; $i < strlen($password); $i++) {
    $char = $password[$i];
    $alt .= md5($char);
}

// 计算最终哈希
$hash = md5($salt . $alt);

echo "=== PHP MakePassword 算法 ===";
echo "密码: " . $password;
echo "Salt: " . $salt;
echo "Alt: " . $alt;
echo "Hash: " . $hash;
echo "\n=== SQL更新语句 ===";
echo "UPDATE admin SET password = '" . $hash . "' WHERE username = 'admin';";
