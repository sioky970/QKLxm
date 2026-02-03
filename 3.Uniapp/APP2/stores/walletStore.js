/**
 * 全局钱包/余额状态管理
 * 统一管理WebSocket余额推送数据，避免各页面重复订阅
 */
import { reactive, computed, readonly } from 'vue'
import wsClient from '@/utils/websocket.js'
import { getAllAssetsWithBalance, walletTransferApi } from '@/utils/api.js'

// ==================== 全局状态 ====================
const state = reactive({
	// 总资产
	totalBalance: 0,
	totalUsdValue: 0,
	
	// 今日盈亏
	todayProfit: 0,
	todayProfitRate: '0.00%',
	
	// USDT余额（单独存储，因为USDT是基础计价货币）
	usdtBalance: 0,
	
	// 现货账户余额
	spotBalance: 0,
	
	// 永续合约账户余额
	contractBalance: 0,
	
	// 交割合约账户余额
	deliveryBalance: 0,
	
	// 资金账户余额
	fundBalance: 0,
	
	// 资产列表（不包含USDT）
	assets: [],
	
	// 状态标记
	isInitialized: false,
	isLoading: false,
	
	// 当前用户ID
	userId: null,
	
	// 最后更新时间
	lastUpdateTime: null
})

// ==================== WebSocket处理 ====================
let wsSubscribed = false
let currentChannel = null

/**
 * 处理WebSocket余额更新
 */
const handleBalanceUpdate = (data) => {
	if (!data) return
	
	console.log('[WalletStore] 收到余额更新:', data)
	state.lastUpdateTime = Date.now()
	
	// 更新总资产
	if (data.total_balance !== undefined) {
		state.totalBalance = parseFloat(data.total_balance) || 0
		state.totalUsdValue = parseFloat(data.total_balance) || 0
	}
	
	// 更新现货、永续合约和交割账户余额（后端推送的数据是字符串格式）
	if (data.spot_balance !== undefined) {
		state.spotBalance = parseFloat(data.spot_balance) || 0
		console.log('[WalletStore] 现货余额已更新:', state.spotBalance)
	}
	if (data.contract_balance !== undefined) {
		state.contractBalance = parseFloat(data.contract_balance) || 0
		console.log('[WalletStore] 永续合约余额已更新:', state.contractBalance)
	}
	if (data.delivery_balance !== undefined) {
		state.deliveryBalance = parseFloat(data.delivery_balance) || 0
		console.log('[WalletStore] 交割余额已更新:', state.deliveryBalance)
	}
	
	// 更新资金账户余额
	if (data.fund_balance !== undefined) {
		state.fundBalance = parseFloat(data.fund_balance) || 0
		console.log('[WalletStore] 资金账户余额已更新:', state.fundBalance)
	}
	
	// 更新盈亏数据
	if (data.today_profit !== undefined) {
		state.todayProfit = parseFloat(data.today_profit) || 0
	}
	if (data.today_profit_rate !== undefined) {
		state.todayProfitRate = data.today_profit_rate || '0.00%'
	}
	
	// 更新资产列表
	if (data.assets && Array.isArray(data.assets)) {
		// 先提取USDT余额
		const usdtAsset = data.assets.find(asset => asset.currency_name === 'USDT')
		if (usdtAsset) {
			state.usdtBalance = parseFloat(usdtAsset.balance) || 0
			console.log('[WalletStore] USDT余额已更新:', state.usdtBalance)
		}
		
		// 更新其他币种资产列表（不包含USDT）
		state.assets = data.assets
			.filter(asset => asset.currency_name !== 'USDT')
			.sort((a, b) => (b.sort || 0) - (a.sort || 0))
			.map(asset => ({
				id: asset.currency_id,
				name: asset.currency_name,
				symbol: asset.symbol,
				balance: parseFloat(asset.balance) || 0,
				usdValue: parseFloat(asset.usd_value) || 0,
				sort: asset.sort
			}))
	}
}

/**
 * 订阅用户余额频道
 */
const subscribeWallet = (userId) => {
	if (!userId) {
		console.warn('[WalletStore] 未提供用户ID')
		return
	}
	
	// 如果已订阅同一用户，跳过
	if (wsSubscribed && state.userId === userId) {
		return
	}
	
	// 取消旧订阅
	if (currentChannel) {
		wsClient.unsubscribe(currentChannel, handleBalanceUpdate)
	}
	
	// 订阅新频道
	state.userId = userId
	currentChannel = `wallet:${userId}`
	console.log('[WalletStore] 订阅余额频道:', currentChannel)
	wsClient.subscribe(currentChannel, handleBalanceUpdate)
	wsSubscribed = true
}

/**
 * 取消订阅
 */
const unsubscribeWallet = () => {
	if (!wsSubscribed || !currentChannel) return
	
	console.log('[WalletStore] 取消订阅余额频道:', currentChannel)
	wsClient.unsubscribe(currentChannel, handleBalanceUpdate)
	wsSubscribed = false
	currentChannel = null
}

// ==================== 数据加载 ====================
// 请求防抖配置
const FETCH_DEBOUNCE_MS = 5000 // 5秒内不重复请求
let lastFetchTime = 0

/**
 * 从后端加载资产数据
 * @param {boolean} force - 是否强制刷新（忽略防抖）
 */
const fetchAssetData = async (force = false) => {
	if (state.isLoading) return
	
	// 防抖检查：5秒内不重复请求（除非强制刷新）
	const now = Date.now()
	if (!force && lastFetchTime && (now - lastFetchTime < FETCH_DEBOUNCE_MS)) {
		console.log('[WalletStore] 请求过于频繁，跳过 (距上次请求:', now - lastFetchTime, 'ms)')
		return
	}
	lastFetchTime = now
	
	state.isLoading = true
	try {
		// 并行加载资产数据、钱包余额和资金账户余额
		const [assetsRes, balanceRes, fundBalanceRes] = await Promise.all([
			getAllAssetsWithBalance(),
			walletTransferApi.getBalance().catch(() => null),
			walletTransferApi.getFundBalance().catch(() => null)
		])
		
		// 处理资产数据
		if (assetsRes && assetsRes.data) {
			const data = assetsRes.data
			
			state.totalBalance = data.total_balance || 0
			state.totalUsdValue = data.total_balance || 0
			state.todayProfit = data.today_profit || 0
			state.todayProfitRate = data.today_profit_rate || '0.00%'
			
			// 更新资产列表
			const assets = data.assets || []
				
			// 先提取USDT余额
			const usdtAsset = assets.find(asset => asset.currency_name === 'USDT')
			if (usdtAsset) {
				state.usdtBalance = usdtAsset.balance || 0
				console.log('[WalletStore] USDT余额已更新:', state.usdtBalance)
			}
				
			// 更新其他币种资产列表（不包含USDT）
			state.assets = assets
				.filter(asset => asset.currency_name !== 'USDT')
				.sort((a, b) => (b.sort || 0) - (a.sort || 0))
				.map(asset => ({
					id: asset.currency_id,
					name: asset.currency_name,
					symbol: asset.symbol,
					balance: asset.balance,
					usdValue: asset.usd_value,
					sort: asset.sort
				}))
		}
		
		// 处理现货、永续合约和交割账户余额
		if (balanceRes && balanceRes.data) {
			const balanceData = balanceRes.data
			state.spotBalance = parseFloat(balanceData.spot_balance) || 0
			state.contractBalance = parseFloat(balanceData.contract_balance) || 0
			state.deliveryBalance = parseFloat(balanceData.delivery_balance) || 0
			
			console.log('[WalletStore] 钱包余额已更新: 现货=', state.spotBalance, ', 永续合约=', state.contractBalance, ', 交割=', state.deliveryBalance)
		} else {
			console.log('[WalletStore] 钱包余额API返回异常: balanceRes=', balanceRes)
		}
		
		// 处理资金账户余额
		if (fundBalanceRes && fundBalanceRes.data) {
			const fundData = fundBalanceRes.data
			state.fundBalance = parseFloat(fundData.available_balance) || 0
			console.log('[WalletStore] 资金账户余额已更新:', state.fundBalance)
		} else {
			console.log('[WalletStore] 资金账户余额API返回异常: fundBalanceRes=', fundBalanceRes)
		}
		
		// 更新总余额（包含资金账户）
		const calculatedTotal = state.spotBalance + state.contractBalance + state.deliveryBalance + state.fundBalance
		if (calculatedTotal > 0 && Math.abs(calculatedTotal - state.totalBalance) > 0.0001) {
			state.totalBalance = calculatedTotal
			state.totalUsdValue = calculatedTotal
		}
		
		state.isInitialized = true
		state.lastUpdateTime = Date.now()
		console.log('[WalletStore] 资产数据加载完成')
	} catch (err) {
		console.error('[WalletStore] 加载资产数据失败:', err)
	} finally {
		state.isLoading = false
	}
}

/**
 * 初始化钱包数据
 */
const initialize = async () => {
	// 获取用户信息
	const userInfo = uni.getStorageSync('userInfo')
	if (!userInfo || !userInfo.id) {
		console.warn('[WalletStore] 用户未登录')
		return
	}
	
	// 订阅WebSocket
	subscribeWallet(userInfo.id)
	
	// 加载初始数据
	await fetchAssetData()
}

/**
 * 重置状态（退出登录时调用）
 */
const reset = () => {
	unsubscribeWallet()
	state.totalBalance = 0
	state.totalUsdValue = 0
	state.todayProfit = 0
	state.todayProfitRate = '0.00%'
	state.usdtBalance = 0
	state.spotBalance = 0
	state.contractBalance = 0
	state.deliveryBalance = 0
	state.fundBalance = 0
	state.assets = []
	state.isInitialized = false
	state.userId = null
}

// ==================== 导出 ====================
export const useWalletStore = () => {
	return {
		// 状态（只读）
		state: readonly(state),
		
		// 初始化
		initialize,
		fetchAssetData,
		reset,
		
		// WebSocket管理
		subscribeWallet,
		unsubscribeWallet,
		
		// 计算属性
		isReady: computed(() => state.isInitialized),
		isLoading: computed(() => state.isLoading)
	}
}

// 默认导出单例
export default useWalletStore()
