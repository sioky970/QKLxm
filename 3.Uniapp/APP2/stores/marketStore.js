/**
 * 全局市场数据状态管理
 * 统一管理WebSocket行情推送数据，避免各页面重复订阅和处理
 */
import { reactive, computed, readonly } from 'vue'
import wsClient from '@/utils/websocket.js'
import { getCurrencyList, getCurrencyQuotation } from '@/utils/api.js'

// 预设价格字典（仅用于首次加载且未接收WS数据时）
const CRYPTO_PRESET_PRICES = {
	'BTC': 45678.90,
	'ETH': 2456.78,
	'BNB': 345.67,
	'XRP': 0.5234,
	'SOL': 98.76,
	'ADA': 0.4567,
	'DOGE': 0.0789,
	'TRX': 0.0956,
	'MATIC': 0.8765,
	'DOT': 6.789,
	'UNI': 5.678,
	'LINK': 14.567,
	'LTC': 78.90,
	'SHIB': 0.000009876,
	'AVAX': 34.56,
	'ATOM': 9.876,
	'BCH': 234.56,
	'FIL': 4.567,
	'APT': 8.901,
	'ARB': 1.234
}

// ==================== 全局状态 ====================
const state = reactive({
	// 所有市场数据 Map: `${currency_id}_${legal_id}` -> marketData
	marketDataMap: new Map(),
	
	// 币种基础信息 Map: currency_id -> currencyInfo
	currencyInfoMap: new Map(),
	
	// 状态标记
	isInitialized: false,
	hasReceivedWsData: false,
	isInitialLoad: true,
	
	// 最后更新时间
	lastUpdateTime: null
})

// ==================== WebSocket处理 ====================
let wsSubscribed = false
let wsBatchSubscribed = false

/**
 * 处理单条WebSocket行情更新
 * 支持完整格式和精简格式
 * 优化: 复用对象，减少GC压力
 */
const handleSingleMarketUpdate = (data) => {
	if (!data) return
	
	// 检测并转换精简格式 (i=currency_id 是精简格式的标识)
	let normalizedData = data
	if (data.i !== undefined) {
		// 精简格式转换为标准格式
		normalizedData = {
			currency_id: data.i,
			legal_id: data.l || 1,
			currency_name: data.n || '',
			now_price: data.p,
			price: data.p,
			change: data.c,           // 已是数字格式
			volume: data.v,
			high: data.h,
			low: data.w,              // 精简格式用w表示low
			open: data.o,
			decimal_scale: data.d     // 价格精度
		}
	}
	
	if (!normalizedData.currency_id) {
		return
	}
	
	// 标记已接收过WebSocket数据
	state.hasReceivedWsData = true
	state.lastUpdateTime = Date.now()
	
	// 构建唯一key
	const key = `${normalizedData.currency_id}_${normalizedData.legal_id || 1}`
	
	// 获取现有数据（复用对象）
	let existing = state.marketDataMap.get(key)
	
	// 解析数值
	const priceNum = parseFloat(normalizedData.now_price || normalizedData.price) || 0
	let changeNum = 0
	if (typeof normalizedData.change === 'string') {
		changeNum = parseFloat(normalizedData.change.replace('%', '')) || 0
	} else {
		changeNum = parseFloat(normalizedData.change) || 0
	}
	const volumeNum = parseFloat(normalizedData.volume) || 0
	const currencyName = normalizedData.currency_name || normalizedData.name || (existing ? existing.name : '') || ''
	
	// 获取精度（优先使用推送的精度，其次使用现有值，默认4）
	const decimalScale = normalizedData.decimal_scale || (existing ? existing.decimal_scale : null) || 4
	
	// 优化: 复用对象，直接修改属性而非创建新对象
	if (existing) {
		// 直接更新现有对象的属性，避免创建新对象
		existing.now_price = priceNum
		existing.price = priceNum
		existing.change = changeNum
		existing.volume = volumeNum
		existing.high = parseFloat(normalizedData.high) || existing.high || 0
		existing.low = parseFloat(normalizedData.low) || existing.low || 0
		existing.amount = parseFloat(normalizedData.amount) || existing.amount || 0
		existing.timestamp = Date.now()
		// 更新精度（如果有新值）
		if (normalizedData.decimal_scale) {
			existing.decimal_scale = normalizedData.decimal_scale
		}
		// 更新名称（如果有新值）
		if (currencyName && currencyName !== existing.name) {
			existing.name = currencyName
			existing.symbol = `${currencyName}/USDT`
		}
	} else {
		// 首次创建对象
		state.marketDataMap.set(key, {
			currency_id: normalizedData.currency_id,
			legal_id: normalizedData.legal_id || 1,
			symbol: normalizedData.symbol || `${currencyName}/USDT`,
			name: currencyName,
			now_price: priceNum,
			price: priceNum,
			change: changeNum,
			volume: volumeNum,
			high: parseFloat(normalizedData.high) || 0,
			low: parseFloat(normalizedData.low) || 0,
			amount: parseFloat(normalizedData.amount) || 0,
			decimal_scale: decimalScale, // 存储精度
			timestamp: Date.now()
		})
	}
}

/**
 * 处理WebSocket行情更新（兼容单条和批量）
 * 这是唯一处理WebSocket推送的地方
 */
const handleMarketUpdate = (data) => {
	// 处理单条数据
	handleSingleMarketUpdate(data)
}

/**
 * 处理批量WebSocket行情更新
 * 优化: 支持后端批量推送的消息格式
 */
const handleMarketBatchUpdate = (dataArray) => {
	if (!Array.isArray(dataArray)) {
		return
	}
	
	// 批量处理所有数据
	dataArray.forEach(data => {
		handleSingleMarketUpdate(data)
	})
}

/**
 * 订阅WebSocket行情（全局只订阅一次）
 * 优化: 同时订阅单条和批量消息频道
 */
const subscribeMarket = () => {
	if (!wsSubscribed) {
		wsClient.subscribe('market', handleMarketUpdate)
		wsSubscribed = true
	}
	// 订阅批量消息频道（后端优化后会使用此频道）
	if (!wsBatchSubscribed) {
		wsClient.subscribe('market_batch', handleMarketBatchUpdate)
		wsBatchSubscribed = true
	}
}

/**
 * 取消订阅WebSocket行情
 */
const unsubscribeMarket = () => {
	if (wsSubscribed) {
		wsClient.unsubscribe('market', handleMarketUpdate)
		wsSubscribed = false
	}
	if (wsBatchSubscribed) {
		wsClient.unsubscribe('market_batch', handleMarketBatchUpdate)
		wsBatchSubscribed = false
	}
}

// ==================== 数据加载 ====================
// 请求防抖配置
const FETCH_DEBOUNCE_MS = 5000 // 5秒内不重复请求
let lastFetchTime = 0

/**
 * 从后端加载市场数据
 * @param {boolean} force - 是否强制刷新（忽略防抖）
 */
const fetchMarketData = async (force = false) => {
	// 防抖检查：5秒内不重复请求（除非强制刷新）
	const now = Date.now()
	if (!force && lastFetchTime && (now - lastFetchTime < FETCH_DEBOUNCE_MS)) {
		console.log('[MarketStore] 请求过于频繁，跳过 (距上次请求:', now - lastFetchTime, 'ms)')
		return
	}
	lastFetchTime = now
	
	try {
		// 1. 获取启用的币种列表
		const currencyRes = await getCurrencyList()
		const enabledCurrencies = currencyRes.data || []
		
		if (!Array.isArray(enabledCurrencies)) {
			console.error('[MarketStore] 币种列表格式错误:', enabledCurrencies)
			return
		}
		
		// 2. 获取行情数据
		const quotationRes = await getCurrencyQuotation()
		const quotations = quotationRes.data || []
		
		if (!Array.isArray(quotations)) {
			console.error('[MarketStore] 行情数据格式错误:', quotations)
			return
		}
		
		// 3. 创建币种ID到币种信息的映射
		const currencyMap = {}
		enabledCurrencies.forEach(currency => {
			currencyMap[currency.id] = currency
			state.currencyInfoMap.set(currency.id, {
				currency_id: currency.id,
				legal_id: 1,
				name: currency.name,
				symbol: `${currency.name}/USDT`,
				tag: '',
				tagClass: ''
			})
		})
		
		// 4. 处理行情数据
		let processedCount = 0
		quotations.forEach(item => {
			const currency = currencyMap[item.currency_id]
			if (!currency) {
				return
			}
			
			// 获取价格
			let priceNum = parseFloat(item.now_price) || 0
			
			// 构建唯一key
			const dataKey = `${item.currency_id}_${item.legal_id || 1}`
			
			// 预设价格逻辑优化：优先使用现有数据，避免价格归零
			const existingData = state.marketDataMap.get(dataKey)
			
			if (priceNum === 0) {
				// 1. 优先使用之前保存的价格数据
				if (existingData && existingData.price > 0) {
					priceNum = existingData.price
				}
				// 2. 其次使用预设价格（仅首次加载且无WS数据时）
				else if (!state.hasReceivedWsData && CRYPTO_PRESET_PRICES[currency.name]) {
					priceNum = CRYPTO_PRESET_PRICES[currency.name]
				}
			}
			
			const changeNum = parseFloat(item.change) || 0
			const volumeNum = parseFloat(item.volume) || 0
			
			// 存储市场数据
			state.marketDataMap.set(dataKey, {
				currency_id: item.currency_id,
				legal_id: item.legal_id || 1,
				symbol: `${currency.name}/USDT`,
				name: currency.name,
				now_price: priceNum,
				price: priceNum,
				change: changeNum,
				volume: volumeNum,
				high: parseFloat(item.high) || 0,
				low: parseFloat(item.low) || 0,
				amount: parseFloat(item.amount) || 0,
				tag: '',
				tagClass: '',
				timestamp: Date.now()
			})
			processedCount++
		})
		
		// 首次加载完成后，标记为false
		state.isInitialLoad = false
		state.isInitialized = true
	} catch (err) {
		console.error('[MarketStore] 加载市场数据失败:', err)
	}
}

/**
 * 初始化市场数据
 */
const initialize = async () => {
	if (state.isInitialized) {
		return
	}
	
	// 立即标记为已初始化，避免重复调用
	state.isInitialized = true
	
	// 订阅WebSocket（不等待API）
	subscribeMarket()
	
	// 异步加载数据（不阻塞页面显示）
	fetchMarketData()
}

// ==================== 数据访问接口 ====================
/**
 * 根据currency_id和legal_id获取市场数据
 */
const getMarketData = (currencyId, legalId = 1) => {
	const key = `${currencyId}_${legalId}`
	return state.marketDataMap.get(key) || null
}

/**
 * 根据币种名称获取市场数据
 */
const getMarketDataByName = (name) => {
	for (const [key, data] of state.marketDataMap.entries()) {
		if (data.name === name) {
			return data
		}
	}
	return null
}

/**
 * 获取所有市场数据（数组形式）
 */
const getAllMarketData = () => {
	return Array.from(state.marketDataMap.values())
}

/**
 * 获取所有市场数据（响应式computed）
 */
const marketDataList = computed(() => {
	return getAllMarketData()
})

/**
 * 获取币种基础信息
 */
const getCurrencyInfo = (currencyId) => {
	return state.currencyInfoMap.get(currencyId) || null
}

/**
 * 获取所有币种基础信息
 */
const getAllCurrencyInfo = () => {
	return Array.from(state.currencyInfoMap.values())
}

/**
 * 获取币种价格精度
 * @param {number} currencyId - 币种ID
 * @param {number} legalId - 法币ID，默认1
 * @returns {number} 价格精度（小数位数），默认4
 */
const getPrecision = (currencyId, legalId = 1) => {
	const key = `${currencyId}_${legalId}`
	const data = state.marketDataMap.get(key)
	return data?.decimal_scale || 4
}

/**
 * 根据币种名称获取价格精度
 * @param {string} name - 币种名称
 * @returns {number} 价格精度（小数位数），默认4
 */
const getPrecisionByName = (name) => {
	const upperName = String(name || '').toUpperCase().trim()
	for (const [, data] of state.marketDataMap) {
		if (data.name?.toUpperCase() === upperName) {
			return data.decimal_scale || 4
		}
	}
	return 4
}

/**
 * 格式化价格（根据币种精度）
 * @param {number} price - 价格
 * @param {number} currencyId - 币种ID
 * @param {number} legalId - 法币ID，默认1
 * @returns {string} 格式化后的价格（自动去除尾随零）
 */
const formatPriceByPrecision = (price, currencyId, legalId = 1) => {
	const precision = getPrecision(currencyId, legalId)
	const num = parseFloat(price)
	if (isNaN(num)) return '0'
	// 格式化并去除尾随零
	return parseFloat(num.toFixed(precision)).toString()
}

// ==================== 导出 ====================
export const useMarketStore = () => {
	return {
		// 状态（只读）
		state: readonly(state),
		
		// 初始化
		initialize,
		fetchMarketData,
		
		// WebSocket管理
		subscribeMarket,
		unsubscribeMarket,
		
		// 数据访问
		getMarketData,
		getMarketDataByName,
		getAllMarketData,
		marketDataList,
		getCurrencyInfo,
		getAllCurrencyInfo,
		
		// 精度相关
		getPrecision,
		getPrecisionByName,
		formatPriceByPrecision,
		
		// 计算属性
		isReady: computed(() => state.isInitialized),
		hasData: computed(() => state.marketDataMap.size > 0)
	}
}

// 默认导出单例
export default useMarketStore()
