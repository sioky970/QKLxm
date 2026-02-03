/**
 * K线数据混合获取管理器
 * 实现WebSocket实时推送 + API轮询降级的混合方案
 */

import wsClient from './websocket.js'

class KlineDataFetcher {
	constructor() {
		this.chart = null                    // K线图表实例
		this.currentSymbol = 'BTC/USDT'      // 当前交易对
		this.currentPeriod = '5min'          // 当前周期
		this.wsConnected = false             // WebSocket连接状态
		this.pollingTimer = null             // 轮询定时器
		this.pollingInterval = 30000         // 轮询间隔(ms)
		this.wsCallback = null               // WebSocket回调函数
		this.apiBaseUrl = 'http://localhost:8080' // API基础URL
		this.fallbackMode = false            // 降级模式标志
		
		// 根据周期配置轮询间隔
		this.periodPollingIntervals = {
			'5min': 30000,    // 30秒
			'15min': 60000,   // 1分钟
			'1h': 120000,     // 2分钟
			'4h': 300000,     // 5分钟
			'1d': 600000      // 10分钟
		}
		
		// 周期映射：前端 -> 后端
		this.periodMap = {
			'1m': '1min',   // 后端不支持
			'5m': '5min',
			'15m': '15min',
			'1h': '60min',
			'4h': '4hour',
			'1d': '1day'
		}
		
		// 绑定WebSocket事件
		this.bindWebSocketEvents()
	}
	
	/**
	 * 绑定WebSocket事件监听
	 */
	bindWebSocketEvents() {
		// WebSocket断开事件
		wsClient.on('disconnect', () => {
			console.log('[KlineFetcher] WebSocket断开，准备切换到轮询模式')
			this.wsConnected = false
			this.fallbackMode = true
			// 延迟启动轮询，给重连一点时间
			setTimeout(() => {
				if (!this.wsConnected) {
					this.startPolling()
				}
			}, 5000)
		})
		
		// WebSocket重连成功事件
		wsClient.on('reconnect', () => {
			console.log('[KlineFetcher] WebSocket重连成功，停止轮询')
			this.wsConnected = true
			this.fallbackMode = false
			this.stopPolling()
			// 补充重连期间的数据
			this.loadHistory()
		})
		
		// WebSocket连接成功事件
		wsClient.on('connect', () => {
			console.log('[KlineFetcher] WebSocket连接成功')
			this.wsConnected = true
			this.fallbackMode = false
		})
		
		// WebSocket错误事件
		wsClient.on('error', (err) => {
			console.error('[KlineFetcher] WebSocket错误:', err)
			this.wsConnected = false
		})
	}
	
	/**
	 * 初始化
	 * @param {Object} chart - K线图表实例
	 * @param {String} symbol - 交易对
	 * @param {String} period - 时间周期
	 */
	async init(chart, symbol, period) {
		console.log('[KlineFetcher] 初始化:', symbol, period)
		
		this.chart = chart
		this.currentSymbol = symbol
		this.currentPeriod = period
		
		// 更新轮询间隔
		this.updatePollingInterval(period)
		
		// 1. 快速加载历史数据
		await this.loadHistory()
		
		// 2. 尝试建立WebSocket连接
		this.connectWebSocket()
	}
	
	/**
	 * 更新轮询间隔
	 */
	updatePollingInterval(period) {
		// 将前端周期转换为后端周期
		const backendPeriod = this.periodMap[period] || period
		
		// 根据周期选择轮询间隔
		this.pollingInterval = this.periodPollingIntervals[backendPeriod] 
			|| this.periodPollingIntervals['5min']
		
		console.log('[KlineFetcher] 轮询间隔设置为:', this.pollingInterval, 'ms')
	}
	
	/**
	 * 加载历史数据
	 */
	async loadHistory() {
		try {
			console.log('[KlineFetcher] 开始加载历史数据...')
			
			const backendPeriod = this.periodMap[this.currentPeriod] || this.currentPeriod
			const url = `${this.apiBaseUrl}/api/kline?symbol=${encodeURIComponent(this.currentSymbol)}&period=${backendPeriod}`
			
			const response = await fetch(url)
			const result = await response.json()
			
			if (result.code === 200 && result.data && result.data.length > 0) {
				// 转换数据格式
				const klineData = result.data.map(item => ({
					timestamp: item.time,      // 后端返回毫秒
					open: item.open,
					high: item.high,
					low: item.low,
					close: item.close,
					volume: item.volume
				}))
				
				// 按时间升序排序
				klineData.sort((a, b) => a.timestamp - b.timestamp)
				
				// 更新图表
				if (this.chart) {
					this.chart.applyNewData(klineData)
				}
				
				console.log('[KlineFetcher] 历史数据加载成功:', klineData.length, '条')
				return true
			} else {
				console.warn('[KlineFetcher] 历史数据为空或格式错误')
				return false
			}
		} catch (error) {
			console.error('[KlineFetcher] 加载历史数据失败:', error)
			return false
		}
	}
	
	/**
	 * 连接WebSocket并订阅K线频道
	 */
	connectWebSocket() {
		console.log('[KlineFetcher] 连接WebSocket...')
		
		// 确保WebSocket已连接
		if (!wsClient.isConnected) {
			wsClient.connect()
		}
		
		// 取消旧订阅
		if (this.wsCallback) {
			const oldChannel = `kline:${this.currentSymbol}`
			wsClient.unsubscribe(oldChannel, this.wsCallback)
		}
		
		// 创建新的回调函数
		this.wsCallback = (data) => {
			this.wsConnected = true
			this.handleKlineUpdate(data)
		}
		
		// 订阅新频道
		const channel = `kline:${this.currentSymbol}`
		wsClient.subscribe(channel, this.wsCallback)
		
		console.log('[KlineFetcher] WebSocket订阅频道:', channel)
	}
	
	/**
	 * 处理K线WebSocket更新
	 */
	handleKlineUpdate(data) {
		try {
			if (!this.chart) return
			
			// 过滤周期不匹配的数据
			const backendPeriod = this.periodMap[this.currentPeriod] || this.currentPeriod
			if (data.period && data.period !== backendPeriod) {
				return
			}
			
			// 构建K线数据
			const klineItem = {
				timestamp: data.timestamp * 1000,  // 后端推送秒，转为毫秒
				open: data.open,
				high: data.high,
				low: data.low,
				close: data.close,
				volume: data.volume
			}
			
			// 更新图表（klinecharts的updateData方法会自动判断是更新还是新增）
			this.chart.updateData(klineItem)
			
			console.log('[KlineFetcher] WebSocket实时更新:', new Date(klineItem.timestamp).toLocaleTimeString())
			
			// 如果正在轮询，停止轮询
			if (this.pollingTimer) {
				console.log('[KlineFetcher] WebSocket恢复，停止轮询')
				this.stopPolling()
			}
		} catch (error) {
			console.error('[KlineFetcher] 处理K线更新失败:', error)
		}
	}
	
	/**
	 * 启动API轮询
	 */
	startPolling() {
		// 避免重复启动
		if (this.pollingTimer) {
			console.log('[KlineFetcher] 轮询已在运行中')
			return
		}
		
		console.log(`[KlineFetcher] 启动API轮询，间隔${this.pollingInterval}ms`)
		
		// 立即执行一次
		this.loadHistory()
		
		// 启动定时轮询
		this.pollingTimer = setInterval(() => {
			// 如果WebSocket已恢复，停止轮询
			if (this.wsConnected) {
				console.log('[KlineFetcher] WebSocket已恢复，停止轮询')
				this.stopPolling()
				return
			}
			
			console.log('[KlineFetcher] API轮询更新...')
			this.loadHistory()
		}, this.pollingInterval)
	}
	
	/**
	 * 停止API轮询
	 */
	stopPolling() {
		if (this.pollingTimer) {
			clearInterval(this.pollingTimer)
			this.pollingTimer = null
			console.log('[KlineFetcher] API轮询已停止')
		}
	}
	
	/**
	 * 切换交易对
	 */
	async switchSymbol(symbol) {
		console.log('[KlineFetcher] 切换交易对:', symbol)
		
		// 取消旧订阅
		if (this.wsCallback) {
			const oldChannel = `kline:${this.currentSymbol}`
			wsClient.unsubscribe(oldChannel, this.wsCallback)
		}
		
		// 更新交易对
		this.currentSymbol = symbol
		
		// 加载新数据
		await this.loadHistory()
		
		// 重新订阅
		this.connectWebSocket()
	}
	
	/**
	 * 切换周期
	 */
	async switchPeriod(period) {
		console.log('[KlineFetcher] 切换周期:', period)
		
		// 更新周期
		this.currentPeriod = period
		
		// 更新轮询间隔
		this.updatePollingInterval(period)
		
		// 停止旧轮询
		this.stopPolling()
		
		// 加载新数据
		await this.loadHistory()
		
		// 如果在降级模式，重启轮询
		if (this.fallbackMode && !this.wsConnected) {
			this.startPolling()
		}
	}
	
	/**
	 * 获取状态信息
	 */
	getState() {
		return {
			symbol: this.currentSymbol,
			period: this.currentPeriod,
			wsConnected: this.wsConnected,
			fallbackMode: this.fallbackMode,
			polling: !!this.pollingTimer,
			pollingInterval: this.pollingInterval,
			wsState: wsClient.getConnectionState()
		}
	}
	
	/**
	 * 清理资源
	 */
	destroy() {
		console.log('[KlineFetcher] 清理资源...')
		
		// 停止轮询
		this.stopPolling()
		
		// 取消订阅
		if (this.wsCallback) {
			const channel = `kline:${this.currentSymbol}`
			wsClient.unsubscribe(channel, this.wsCallback)
			this.wsCallback = null
		}
		
		// 清空引用
		this.chart = null
	}
}

// 导出单例
export default KlineDataFetcher
