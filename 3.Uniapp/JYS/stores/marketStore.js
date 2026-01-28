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

/**
 * 处理WebSocket行情更新
 * 这是唯一处理WebSocket推送的地方
 */
const handleMarketUpdate = (data) => {
	if (!data || !data.currency_id) {
		return
	}
	
	// 标记已接收过WebSocket数据
	state.hasReceivedWsData = true
	state.lastUpdateTime = Date.now()
	
	// 构建唯一key
	const key = `${data.currency_id}_${data.legal_id || 1}`
	
	// 获取现有数据或创建新数据
	const existing = state.marketDataMap.get(key) || {}
	
	// 获取价格，完全使用WebSocket推送的数据，不再使用预设价格
	const priceNum = parseFloat(data.now_price || data.price) || 0
	// change 可能是字符串如 "+1.23%" 或数字
	let changeNum = 0
	if (typeof data.change === 'string') {
		changeNum = parseFloat(data.change.replace('%', '')) || 0
	} else {
		changeNum = parseFloat(data.change) || 0
	}
	const volumeNum = parseFloat(data.volume) || 0
	
	// 获取币种名称：优先使用 currency_name，其次 name，最后 existing.name
	const currencyName = data.currency_name || data.name || existing.name || ''
	
	// 更新数据
	const updatedData = {
		...existing,
		currency_id: data.currency_id,
		legal_id: data.legal_id || 1,
		symbol: data.symbol || existing.symbol || `${currencyName}/USDT`,
		name: currencyName,
		now_price: priceNum,
		price: priceNum,
		change: changeNum,
		volume: volumeNum,
		timestamp: Date.now()
	}
	
	state.marketDataMap.set(key, updatedData)
}

/**
 * 订阅WebSocket行情（全局只订阅一次）
 */
const subscribeMarket = () => {
	if (wsSubscribed) {
		return
	}
	
	wsClient.subscribe('market', handleMarketUpdate)
	wsSubscribed = true
}

/**
 * 取消订阅WebSocket行情
 */
const unsubscribeMarket = () => {
	if (!wsSubscribed) {
		return
	}
	
	wsClient.unsubscribe('market', handleMarketUpdate)
	wsSubscribed = false
}

// ==================== 数据加载 ====================
/**
 * 从后端加载市场数据
 */
const fetchMarketData = async () => {
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
			
			// 预设价格只在首次加载且未接收WS数据时使用
			if (priceNum === 0 && state.isInitialLoad && !state.hasReceivedWsData && CRYPTO_PRESET_PRICES[currency.name]) {
				priceNum = CRYPTO_PRESET_PRICES[currency.name]
			}
			
			const changeNum = parseFloat(item.change) || 0
			const volumeNum = parseFloat(item.volume) || 0
			
			// 构建唯一key
			const key = `${item.currency_id}_${item.legal_id || 1}`
			
			// 存储市场数据
			state.marketDataMap.set(key, {
				currency_id: item.currency_id,
				legal_id: item.legal_id || 1,
				symbol: `${currency.name}/USDT`,
				name: currency.name,
				now_price: priceNum,
				price: priceNum,
				change: changeNum,
				volume: volumeNum,
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
		
		// 计算属性
		isReady: computed(() => state.isInitialized),
		hasData: computed(() => state.marketDataMap.size > 0)
	}
}

// 默认导出单例
export default useMarketStore()
