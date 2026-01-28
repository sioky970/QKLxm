/**
 * 格式化货币显示
 * @param {number} amount - 金额
 * @param {number} decimals - 小数位数，默认为2
 * @returns {string} 格式化后的金额字符串
 */
export function formatCurrency(amount, decimals = 2) {
	if (typeof amount !== 'number' || isNaN(amount)) {
		return '0.00';
	}
	
	// 处理非常小的数字，避免科学计数法显示
	if (Math.abs(amount) < Math.pow(10, -decimals) && amount !== 0) {
		return '<' + Math.pow(10, -decimals).toFixed(decimals);
	}
	
	return amount.toLocaleString(undefined, {
		minimumFractionDigits: decimals,
		maximumFractionDigits: decimals
	});
}

/**
 * 格式化数字显示
 * @param {number} num - 数字
 * @param {number} decimals - 小数位数，默认为2
 * @returns {string} 格式化后的数字字符串
 */
export function formatNumber(num, decimals = 2) {
	if (typeof num !== 'number' || isNaN(num)) {
		return '0.00';
	}
	
	// 处理非常小的数字，避免科学计数法显示
	if (Math.abs(num) < Math.pow(10, -decimals) && num !== 0) {
		return '<' + Math.pow(10, -decimals).toFixed(decimals);
	}
	
	return num.toFixed(decimals);
}

/**
 * 格式化百分比显示
 * @param {number|string} value - 百分比值
 * @returns {string} 格式化后的百分比字符串
 */
export function formatPercent(value) {
	if (typeof value === 'string') {
		// 如果已经是字符串，直接返回
		return value.includes('%') ? value : value + '%';
	}
	if (typeof value !== 'number' || isNaN(value)) {
		return '0.00%';
	}
	const prefix = value >= 0 ? '+' : '';
	return prefix + value.toFixed(2) + '%';
}

/**
 * 格式化余额显示（限制9位有效数字）
 * 专用于主页(index.vue)和资产页(assets.vue)的余额显示
 * 
 * @param {number} amount - 金额数值
 * @returns {string} 格式化后的金额字符串
 * 
 * 规则：
 * 1. 最多显示9位有效数字（小数点前后的每一位都计入）
 * 2. 直接截取，不四舍五入
 * 3. 特殊情况：极小数额（<0.0000001）显示为0
 * 4. 整数部分超过9位时，只保留前9位整数，小数部分舍弃
 * 
 * 示例：
 * - 1000000.0000000032 → 1000000.00（前9位：100000000，小数点后保留2位00）
 * - 0.00023404 → 0.0002340（前9位：000023404）
 * - 123456789.1 → 123456789（前9位：123456789）
 * - 987654321.999999 → 987654321（前9位：987654321）
 */
export function formatBalanceWith9Digits(amount) {
	// 类型检查
	if (typeof amount !== 'number' || isNaN(amount) || amount === null || amount === undefined) {
		return '0';
	}
	
	// 处理0和负数
	if (amount === 0) {
		return '0';
	}
	
	// 处理负数（保留符号）
	const isNegative = amount < 0;
	const absAmount = Math.abs(amount);
	
	// 处理极小数（小于0.0000001）显示为0
	if (absAmount < 0.0000001) {
		return '0';
	}
	
	// 转换为字符串进行处理
	const amountStr = absAmount.toString();
	
	// 处理科学计数法
	let normalizedStr;
	if (amountStr.includes('e')) {
		normalizedStr = absAmount.toFixed(20); // 先展开科学计数法
	} else {
		normalizedStr = amountStr;
	}
	
	// 移除小数点，统计所有数字
	const digitsOnly = normalizedStr.replace('.', '').replace(/^0+/, ''); // 移除前导0
	
	// 如果全是0，返回0
	if (digitsOnly.length === 0 || digitsOnly === '0') {
		return '0';
	}
	
	// 分离整数和小数部分
	const parts = normalizedStr.split('.');
	const integerPart = parts[0];
	const decimalPart = parts[1] || '';
	
	// 去除整数部分前导0（但至少保留一个0）
	let integerDigits = integerPart.replace(/^0+/, '') || '0';
	const integerDigitCount = integerDigits === '0' ? 0 : integerDigits.length;
	
	// 如果整数部分就已经>=9位
	if (integerDigitCount >= 9) {
		// 只保留前9位整数，小数部分舍弃
		const result = integerDigits.substring(0, 9);
		return isNegative ? '-' + result : result;
	}
	
	// 整数部分<9位，需要从小数部分补充
	const remainingDigits = 9 - integerDigitCount;
	
	// 构建结果
	let result;
	if (integerDigitCount === 0) {
		// 整数部分全是0（如0.00023404）
		// 找到小数部分第一个非零数字的位置
		const firstNonZeroIndex = decimalPart.search(/[1-9]/);
		if (firstNonZeroIndex === -1) {
			return '0';
		}
		// 保留"0." + 前导0 + 9位有效数字
		const zerosBeforeFirstDigit = decimalPart.substring(0, firstNonZeroIndex);
		const significantDigits = decimalPart.substring(firstNonZeroIndex);
		const digitsToKeep = significantDigits.substring(0, 9);
		result = '0.' + zerosBeforeFirstDigit + digitsToKeep;
	} else {
		// 整数部分有非零数字
		// 截取小数部分（不四舍五入）
		const decimalToKeep = decimalPart.substring(0, remainingDigits);
		if (decimalToKeep.length > 0) {
			result = integerDigits + '.' + decimalToKeep;
		} else {
			result = integerDigits;
		}
	}
	
	// 移除尾随的0和小数点（仅限小数部分）
	if (result.includes('.')) {
		// 只有包含小数点时才移除尾随0
		result = result.replace(/\.?0+$/, '');
	}
	
	// 如果结果是空字符串或只有小数点，返回0
	if (result === '' || result === '.') {
		return '0';
	}
	
	return isNegative ? '-' + result : result;
}