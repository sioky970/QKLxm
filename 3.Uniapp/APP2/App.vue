<script>
import wsClient from '@/utils/websocket.js'
import marketStore from '@/stores/marketStore.js'
import walletStore from '@/stores/walletStore.js'

export default {
	onLaunch: function() {
		console.log('[App] Launch')
		
		// 全局初始化WebSocket连接
		const isLoggedIn = uni.getStorageSync('isLoggedIn')
		console.log('[App] 登录状态:', isLoggedIn)
		
		if (isLoggedIn) {
			console.log('[App] 开始连接WebSocket...')
			wsClient.connect()
			marketStore.initialize()
			
			// 初始化 walletStore（统一管理余额数据）
			const userInfo = uni.getStorageSync('userInfo')
			if (userInfo && userInfo.id) {
				console.log('[App] 初始化 walletStore，用户ID:', userInfo.id)
				walletStore.initialize()
			}
		}
	},
	onShow: function() {
		console.log('[App] Show')
		
		const isLoggedIn = uni.getStorageSync('isLoggedIn')
		if (isLoggedIn) {
			// WebSocket 断线重连
			if (!wsClient.isConnected) {
				console.log('[App] WebSocket 未连接，启动重连')
				wsClient.connect()
			}
			
			// 确保 marketStore 已初始化
			if (!marketStore.isReady.value) {
				console.log('[App] marketStore 未初始化，执行初始化')
				marketStore.initialize()
			} else {
				// 已初始化的情况下，主动刷新数据以确保数据最新
				console.log('[App] 刷新 marketStore 数据')
				marketStore.fetchMarketData()
			}
			
			// 确保 walletStore 也初始化
			const userInfo = uni.getStorageSync('userInfo')
			if (userInfo && userInfo.id) {
				if (!walletStore.isReady.value) {
					walletStore.initialize()
				} else {
					// 刷新余额数据
					walletStore.fetchAssetData()
				}
			}
		}
	},
	onHide: function() {
		console.log('[App] Hide')
	}
}
</script>

<style>
	/*每个页面公共css */
</style>
