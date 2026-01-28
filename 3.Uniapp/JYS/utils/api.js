/**
 * API请求封装
 * 统一处理请求、响应和错误
 */

// API基础配置
const BASE_URL = 'http://localhost:8080/api'

// 请求拦截器
const requestInterceptor = (config) => {
	// 添加token到请求头
	const token = uni.getStorageSync('token')
	if (token) {
		config.header = {
			...config.header,
			'Authorization': `Bearer ${token}`
		}
	}
	return config
}

// 响应拦截器
const responseInterceptor = (response) => {
	const { statusCode, data } = response
	
	if (statusCode === 200) {
		// 后端响应格式: {type: "success/error", message: "xxx", error: "", data: {...}}
		// type为"success"表示成功，"error"表示业务错误
		if (data && data.type === 'success') {
			return Promise.resolve(data)
		} else if (data) {
			// 业务错误，抛出错误信息
			return Promise.reject(new Error(data.error || data.message || '请求失败'))
		} else {
			return Promise.reject(new Error('服务器响应格式错误'))
		}
	} else if (statusCode === 401) {
		// 未授权，清除登录状态
		uni.removeStorageSync('token')
		uni.removeStorageSync('isLoggedIn')
		uni.removeStorageSync('userInfo')
		
		uni.showToast({
			title: '登录已过期，请重新登录',
			icon: 'none',
			duration: 2000
		})
		
		setTimeout(() => {
			uni.navigateTo({
				url: '/pages/login/login'
			})
		}, 2000)
		
		return Promise.reject(new Error('未授权'))
	} else {
		return Promise.reject(new Error(data?.error || data?.message || '网络错误'))
	}
}

// 统一请求方法
const request = (options) => {
	const config = requestInterceptor({
		url: BASE_URL + options.url,
		method: options.method || 'GET',
		data: options.data || {},
		header: {
			'Content-Type': 'application/json',
			...options.header
		}
	})
	
	return new Promise((resolve, reject) => {
		uni.request({
			...config,
			success: (res) => {
				responseInterceptor(res)
					.then(resolve)
					.catch(reject)
			},
			fail: (err) => {
				uni.showToast({
					title: '网络连接失败',
					icon: 'none'
				})
				reject(new Error('网络连接失败'))
			}
		})
	})
}

// ==================== 公开配置API ====================

/**
 * 获取注册配置
 * @returns {Promise<{invite_code_required: boolean}>}
 */
export const getRegisterConfig = () => {
	return request({
		url: '/config/register',
		method: 'GET'
	})
}

// ==================== 用户相关API ====================

/**
 * 用户登录
 * @param {Object} data - 登录参数
 * @param {string} data.user_string - 账号（邮箱或手机号）
 * @param {string} data.password - 密码
 * @param {number} [data.type=1] - 登录类型 1:普通密码 2:手势密码
 * @param {number} [data.area_code_id=0] - 区号ID
 * @returns {Promise<{token: string, user: Object}>}
 */
export const login = (data) => {
	return request({
		url: '/user/login',
		method: 'POST',
		data: {
			user_string: data.user_string,
			password: data.password,
			type: data.type || 1,
			area_code_id: data.area_code_id || 0
		}
	})
}

/**
 * 用户注册
 * @param {Object} data - 注册参数
 * @param {string} data.type - 注册类型 'mobile' 或 'email'
 * @param {string} data.user_string - 手机号或邮箱
 * @param {string} data.password - 密码
 * @param {string} data.re_password - 确认密码
 * @param {string} data.code - 验证码
 * @param {string} [data.extension_code] - 邀请码（可选）
 * @param {number} [data.country_code=0] - 区号ID
 * @returns {Promise}
 */
export const register = (data) => {
	return request({
		url: '/user/register',
		method: 'POST',
		data: {
			type: data.type,
			user_string: data.user_string,
			password: data.password,
			re_password: data.re_password,
			code: data.code,
			extension_code: data.extension_code || '',
			country_code: data.country_code || 0
		}
	})
}

/**
 * 发送验证码
 * @param {string} target - 手机号或邮箱
 * @param {string} type - 类型 'mobile' 或 'email'
 * @returns {Promise<{code?: string}>} 开发模式会返回验证码
 */
export const sendVerifyCode = (target, type = 'mobile') => {
	return request({
		url: '/user/send_code',
		method: 'POST',
		data: {
			target,
			type
		}
	})
}

/**
 * 获取用户信息
 * @returns {Promise<Object>}
 */
export const getUserInfo = () => {
	return request({
		url: '/user/info',
		method: 'GET'
	})
}

/**
 * 重置密码
 * @param {Object} data - 重置密码参数
 * @param {string} data.account - 账号
 * @param {number} data.country_code - 区号ID
 * @param {string} data.password - 新密码
 * @param {string} data.repassword - 确认密码
 * @param {string} data.code - 验证码
 * @returns {Promise}
 */
export const resetPassword = (data) => {
	return request({
		url: '/user/reset_password',
		method: 'POST',
		data
	})
}

// ==================== 钱包资产API ====================

/**
 * 获取资产概览
 * @returns {Promise<{total_balance: number, total_usd_value: number, today_profit: number, today_profit_rate: string, assets: Array}>}
 */
export const getAssetOverview = () => {
	return request({
		url: '/wallet/asset-overview',
		method: 'GET'
	})
}

/**
 * 获取七日盈亏数据
 * @returns {Promise<{profit_loss_data: Array}>}
 */
export const getSevenDaysProfitLoss = () => {
	return request({
		url: '/wallet/seven-days-profit-loss',
		method: 'GET'
	})
}

/**
 * 获取所有启用币种及用户余额（包括余额为0的）
 * @returns {Promise<{total_balance: number, total_usd_value: number, today_profit: number, today_profit_rate: string, assets: Array}>}
 */
export const getAllAssetsWithBalance = () => {
	return request({
		url: '/wallet/all-assets',
		method: 'GET'
	})
}

// ==================== 充值相关API ====================

/**
 * 获取充值地址列表
 * @returns {Promise<Array>} 充值地址列表（按网络类型）
 */
export const getDepositAddresses = () => {
	return request({
		url: '/deposit/addresses',
		method: 'GET'
	})
}

/**
 * 创建充值订单
 * @param {Object} data - 充值信息
 * @param {string} data.network - 网络类型: TRC20/ERC20/BEP20/BRC20
 * @param {number} data.amount - 充值金额(最低10 USDT)
 * @returns {Promise<Object>} 充值订单信息
 */
export const createDepositOrder = (data) => {
	return request({
		url: '/deposit/create',
		method: 'POST',
		data
	})
}

/**
 * 上传转账截图
 * @param {string} filePath - 图片本地路径
 * @param {number} orderId - 订单ID
 * @returns {Promise<Object>} 上传结果
 */
export const uploadDepositScreenshot = (filePath, orderId) => {
	return new Promise((resolve, reject) => {
		const token = uni.getStorageSync('token')
		uni.uploadFile({
			url: BASE_URL + '/deposit/upload',
			filePath: filePath,
			name: 'file',
			formData: {
				order_id: String(orderId)
			},
			header: {
				'Authorization': token ? `Bearer ${token}` : ''
			},
			success: (res) => {
				if (res.statusCode === 200) {
					try {
						const data = JSON.parse(res.data)
						if (data.type === 'success') {
							resolve(data)
						} else {
							reject(new Error(data.error || data.message || '上传失败'))
						}
					} catch (e) {
						reject(new Error('解析响应失败'))
					}
				} else {
					reject(new Error('上传失败'))
				}
			},
			fail: (err) => {
				reject(new Error('网络错误'))
			}
		})
	})
}

/**
 * 获取用户充值记录
 * @param {Object} params - 查询参数
 * @param {number} [params.status=-1] - 状态筛选: -1全部 0待审核 1已通过 2已拒绝 3已取消 4已过期
 * @param {number} [params.page=1] - 页码
 * @param {number} [params.page_size=10] - 每页数量
 * @returns {Promise<{list: Array, total: number}>}
 */
export const getDepositOrders = (params = {}) => {
	return request({
		url: '/deposit/orders',
		method: 'GET',
		data: {
			status: params.status !== undefined ? params.status : -1,
			page: params.page || 1,
			page_size: params.page_size || 10
		}
	})
}

/**
 * 取消充值订单
 * @param {number} orderId - 订单ID
 * @returns {Promise}
 */
export const cancelDepositOrder = (orderId) => {
	return request({
		url: '/deposit/cancel',
		method: 'POST',
		data: {
			order_id: orderId
		}
	})
}

// ==================== 行情API ====================

/**
 * 获取启用的币种列表
 * @returns {Promise<Array>}
 */
export const getCurrencyList = () => {
	return request({
		url: '/currency/list',
		method: 'GET'
	})
}

/**
 * 获取币种行情（最新价格）
 * @returns {Promise<Array>}
 */
export const getCurrencyQuotation = () => {
	return request({
		url: '/quotation/new',
		method: 'GET'
	})
}

/**
 * 获取市场数据
 * @param {Object} data - 市场数据参数
 * @returns {Promise}
 */
export const getMarketData = (data) => {
	return request({
		url: '/market/market',
		method: 'POST',
		data
	})
}

/**
 * 获取K线数据
 * @param {Object} params - K线参数
 * @param {string} params.symbol - 交易对符号，如 'BTC/USDT'
 * @param {string} params.period - 周期，如 '1min', '5min', '15min', '30min', '60min', '1day', '1week', '1mon'
 * @param {number} [params.size=150] - 数据条数
 * @returns {Promise<Array>}
 */
export const getKlineData = (params) => {
	return request({
		url: '/kline',
		method: 'GET',
		data: params
	})
}

// 获取指定交易对的K线数据
export const getSymbolKlineData = (symbol, period = '1min', size = 150) => {
	return request({
		url: '/kline',
		method: 'GET',
		data: {
			symbol,
			period,
			size
		}
	})
}

// ==================== 现货交易 API ====================

/**
 * 提交现货订单
 * @param {Object} data - 订单数据
 * @param {number} data.currency_id - 币种ID
 * @param {number} data.legal_id - 法币ID（通常是USDT）
 * @param {string} data.type - 订单类型: "limit"（限价）或 "market"（市价）
 * @param {string} data.side - 交易方向: "buy"（买入）或 "sell"（卖出）
 * @param {number} data.price - 价格（市价单传当前市价）
 * @param {number} data.quantity - 数量
 * @returns {Promise}
 */
export const submitSpotOrder = (data) => {
	return request({
		url: '/transaction/submit',
		method: 'POST',
		data
	})
}

/**
 * 取消现货订单
 * @param {Object} data
 * @param {number} data.order_id - 订单ID
 * @returns {Promise}
 */
export const cancelSpotOrder = (data) => {
	return request({
		url: '/transaction/cancel',
		method: 'POST',
		data
	})
}

// 追单：修改订单价格并成交
export const chaseSpotOrder = (data) => {
	return request({
		url: '/transaction/chase',
		method: 'POST',
		data
	})
}

/**
 * 获取现货订单列表
 * @param {Object} params
 * @param {string} params.symbol - 交易对（可选）
 * @param {string} params.side - 交易方向（可选）
 * @param {string} params.status - 订单状态（可选）
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @returns {Promise}
 */
export const getSpotOrderList = (params) => {
	return request({
		url: '/transaction/list',
		method: 'GET',
		params
	})
}

// ==================== 合约交易相关API ====================

/**
 * 获取杠杆倍数配置列表（全局统一）
 * @returns {Promise<Array>}
 */
export const getLeverageMultiples = () => {
	return request({
		url: '/admin/lever/multiple/list',
		method: 'GET'
	})
}

// ==================== 永续合约交易API ====================

/**
 * 开仓（市价单/限价单）
 * @param {Object} data - 开仓参数
 * @param {number} data.currency_id - 币种ID
 * @param {number} data.legal_id - 法币ID（通常是 USDT=1）
 * @param {number} data.type - 方向 1=做多 2=做空
 * @param {number} data.order_type - 订单类型 1=市价单 2=限价单
 * @param {number} data.margin - 保证金金额(USDT)
 * @param {number} data.leverage - 杠杆倍数
 * @param {number} [data.limit_price] - 限价价格（限价单必填）
 * @returns {Promise<Object>}
 */
export const contractOpenPosition = (data) => {
	return request({
		url: '/contract/open',
		method: 'POST',
		data
	})
}

/**
 * 手动平仓
 * @param {Object} data - 平仓参数
 * @param {number} data.position_id - 仓位ID
 * @returns {Promise<Object>}
 */
export const contractClosePosition = (data) => {
	return request({
		url: '/contract/close',
		method: 'POST',
		data
	})
}

/**
 * 追单（限价单转市价立即成交）
 * @param {Object} data - 追单参数
 * @param {number} data.order_id - 限价单订单ID
 * @returns {Promise<Object>}
 */
export const contractChaseOrder = (data) => {
	return request({
		url: '/contract/chase',
		method: 'POST',
		data
	})
}

/**
 * 设置止盈止损
 * @param {Object} data - 止盈止损参数
 * @param {number} data.position_id - 仓位ID
 * @param {number} [data.take_profit_amount] - 止盈金额(USDT)
 * @param {number} [data.stop_loss_amount] - 止损金额(USDT)
 * @returns {Promise<Object>}
 */
export const contractSetTPSL = (data) => {
	return request({
		url: '/contract/tpsl',
		method: 'POST',
		data
	})
}

/**
 * 取消限价单
 * @param {number} orderId - 订单ID
 * @returns {Promise<Object>}
 */
export const contractCancelOrder = (orderId) => {
	return request({
		url: `/contract/cancel/${orderId}`,
		method: 'POST'
	})
}

/**
 * 获取仓位列表
 * @param {Object} [params] - 查询参数
 * @param {string} [params.status] - 仓位状态过滤（可选）
 * @returns {Promise<Array>}
 */
export const getContractPositions = (params = {}) => {
	return request({
		url: '/contract/positions',
		method: 'GET',
		data: params
	})
}

/**
 * 获取挂单列表（待成交的限价单）
 * @returns {Promise<Array>}
 */
export const getContractPendingOrders = () => {
	return request({
		url: '/contract/pending',
		method: 'GET'
	})
}

/**
 * 获取仓位详情
 * @param {number} positionId - 仓位ID
 * @returns {Promise<Object>}
 */
export const getContractPositionDetail = (positionId) => {
	return request({
		url: `/contract/position/${positionId}`,
		method: 'GET'
	})
}

/**
 * 获取历史仓位（已平仓）
 * @param {Object} [params] - 查询参数
 * @param {number} [params.page=1] - 页码
 * @param {number} [params.page_size=20] - 每页数量
 * @returns {Promise<Object>}
 */
export const getContractHistoryPositions = (params = {}) => {
	return request({
		url: '/contract/history',
		method: 'GET',
		data: {
			page: params.page || 1,
			page_size: params.page_size || 20
		}
	})
}

// ==================== 统一订单查询API ====================

/**
 * 获取统一订单列表（聚合现货、永续合约、秒合约）
 * @param {Object} [params] - 查询参数
 * @param {string} [params.trade_type] - 交易类型: spot(现货), contract(合约), micro(秒合约)，空则查询全部
 * @param {number} [params.currency_id] - 币种ID筛选
 * @param {number} [params.status] - 状态筛选
 * @param {string} [params.side] - 方向筛选: buy/sell/long/short
 * @param {number} [params.page=1] - 页码
 * @param {number} [params.page_size=20] - 每页数量
 * @returns {Promise<Object>}
 */
export const getUnifiedOrderList = (params = {}) => {
	return request({
		url: '/orders/all',
		method: 'GET',
		data: {
			trade_type: params.trade_type || '',
			currency_id: params.currency_id || 0,
			status: params.status,
			side: params.side || '',
			page: params.page || 1,
			page_size: params.page_size || 20
		}
	})
}

/**
 * 获取订单统计数据
 * @returns {Promise<Object>}
 */
export const getOrderStatistics = () => {
	return request({
		url: '/orders/statistics',
		method: 'GET'
	})
}

/**
 * 获取订单详情
 * @param {number} orderId - 订单ID
 * @param {string} tradeType - 交易类型: spot/contract/micro
 * @returns {Promise<Object>}
 */
export const getOrderDetail = (orderId, tradeType) => {
	return request({
		url: '/orders/detail',
		method: 'GET',
		data: {
			order_id: orderId,
			trade_type: tradeType
		}
	})
}

// ==================== 秒合约交易API ====================

/**
 * 获取秒合约周期配置列表
 * @returns {Promise<Array>}
 */
export const getMicroPeriods = () => {
	return request({
		url: '/micro/periods',
		method: 'GET'
	})
}

/**
 * 秒合约下单
 * @param {Object} data - 下单参数
 * @param {number} data.currency_id - 币种ID
 * @param {string} data.direction - 方向: "rise"(买涨) 或 "fall"(买跌)
 * @param {number} data.seconds - 周期秒数(30/60/120/180/300)
 * @param {number} data.amount - 投入金额(USDT)，最低10
 * @returns {Promise<Object>}
 */
export const submitMicroOrder = (data) => {
	return request({
		url: '/micro/submit',
		method: 'POST',
		data
	})
}

/**
 * 获取秒合约进行中订单
 * @returns {Promise<Object>}
 */
export const getMicroActiveOrders = () => {
	return request({
		url: '/micro/active',
		method: 'GET'
	})
}

/**
 * 获取秒合约订单列表（历史结算）
 * @param {Object} [params] - 查询参数
 * @param {number} [params.page=1] - 页码
 * @param {number} [params.page_size=20] - 每页数量
 * @returns {Promise<Object>}
 */
export const getMicroOrderList = (params = {}) => {
	return request({
		url: '/micro/orders',
		method: 'GET',
		data: {
			page: params.page || 1,
			page_size: params.page_size || 20
		}
	})
}

/**
 * 获取秒合约订单详情
 * @param {number} orderId - 订单ID
 * @returns {Promise<Object>}
 */
export const getMicroOrderDetail = (orderId) => {
	return request({
		url: `/micro/order/${orderId}`,
		method: 'GET'
	})
}

// ==================== 导出默认实例 ====================

export default {
	getRegisterConfig,
	login,
	register,
	sendVerifyCode,
	getUserInfo,
	resetPassword,
	getAssetOverview,
	getSevenDaysProfitLoss,
	getAllAssetsWithBalance,
	getCurrencyList,
	getCurrencyQuotation,
	getMarketData,
	getKlineData,
	submitSpotOrder,
	cancelSpotOrder,
	getSpotOrderList,
	getLeverageMultiples,
	// 永续合约API
	contractOpenPosition,
	contractClosePosition,
	contractChaseOrder,
	contractSetTPSL,
	contractCancelOrder,
	getContractPositions,
	getContractPendingOrders,
	getContractPositionDetail,
	getContractHistoryPositions,
	// 统一订单API
	getUnifiedOrderList,
	getOrderStatistics,
	getOrderDetail,
	// 秒合约API
	getMicroPeriods,
	submitMicroOrder,
	getMicroActiveOrders,
	getMicroOrderList,
	getMicroOrderDetail,
	// 充值API
	getDepositAddresses,
	createDepositOrder,
	uploadDepositScreenshot,
	getDepositOrders,
	cancelDepositOrder,
	request
}
