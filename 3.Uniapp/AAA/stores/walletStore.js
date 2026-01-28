/**
 * 全局钱包/余额状态管理
 * 统一管理WebSocket余额推送数据，避免各页面重复订阅
 */
import { reactive, computed, readonly } from 'vue'
import wsClient from '@/utils/websocket.js'
import { getAllAssetsWithBalance } from '@/utils/api.js'

// ==================== 全局状态 ====================
const state = reactive({
	// 总资产
	totalBalance: 0,
	totalUsdValue: 0,
	
	// 今日盈亏
	todayProfit: 0,
	todayProfitRate: '0.00%',
	
	// 资产列表
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
		state.totalBalance = data.total_balance
		state.totalUsdValue = data.total_balance // USDT 1:1
	}
	
	// 更新盈亏数据
	if (data.today_profit !== undefined) {
		state.todayProfit = data.today_profit
	}
	if (data.today_profit_rate !== undefined) {
		state.todayProfitRate = data.today_profit_rate
	}
	
	// 更新资产列表
	if (data.assets && Array.isArray(data.assets)) {
		state.assets = data.assets
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
/**
 * 从后端加载资产数据
 */
const fetchAssetData = async () => {
	if (state.isLoading) return
	
	state.isLoading = true
	try {
		const res = await getAllAssetsWithBalance()
		const data = res.data
		
		// 更新状态
		state.totalBalance = data.total_balance || 0
		state.totalUsdValue = data.total_balance || 0
		state.todayProfit = data.today_profit || 0
		state.todayProfitRate = data.today_profit_rate || '0.00%'
		
		// 更新资产列表
		state.assets = (data.assets || [])
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
