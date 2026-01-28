<script>
import wsClient from '@/utils/websocket.js'
import marketStore from '@/stores/marketStore.js'

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
		}
	},
	onShow: function() {
		console.log('[App] Show')
		
		const isLoggedIn = uni.getStorageSync('isLoggedIn')
		if (isLoggedIn && !wsClient.isConnected) {
			wsClient.connect()
			if (!marketStore.isReady.value) {
				marketStore.initialize()
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
