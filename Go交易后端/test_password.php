<?php
function MakePassword($password) {
    $salt = "ABCDEFG";
    $passwordChars = str_split($password);
    foreach ($passwordChars as $char) {
        $salt .= md5($char);
    }
    return md5($salt);
}
echo "admin -> " . MakePassword("admin") . "\n";
echo "Expected: e3b6712b0207804be7d52c92d766e102\n";
