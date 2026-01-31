/**
 * WebSocket客户端工具类
 * 用于实时接收服务器推送的行情数据、K线数据等
 */

// WebSocket地址配置
// 开发环境使用本地后端，生产环境使用线上服务器
const DEV_WS_URL = 'ws://localhost:8081/ws'
const PROD_WS_URL = 'wss://api.gamaoxo223.shop/ws'

// 判断运行环境
const isDev = process.env.NODE_ENV === 'development'
const WS_URL = isDev ? DEV_WS_URL : PROD_WS_URL

// 调试信息（仅开发模式）
if (isDev) {
	console.log('[WS] 开发模式，使用本地WebSocket:', DEV_WS_URL)
}

class WebSocketClient {
	constructor() {
		this.socket = null
		this.url = WS_URL
		this.isConnected = false
		this.reconnectTimer = null
		this.reconnectAttempts = 0
		this.maxReconnectAttempts = 5
		this.reconnectInterval = 3000
		this.subscriptions = new Map() // channel -> callbacks[]
		this.messageQueue = [] // 待发送的消息队列
	}

	/**
	 * 连接WebSocket
	 */
	connect() {
		if (this.isConnected) {
			console.log('[WS] 已连接，跳过')
			return
		}

		console.log('[WS] 开始连接:', this.url)

		// #ifdef H5
		// H5环境使用原生WebSocket
		try {
			this.socket = new WebSocket(this.url)
			
			this.socket.onopen = () => {
				console.log('[WS] H5连接成功')
				// 优化: 移除延迟，直接处理连接成功
				this.isConnected = true
				this.reconnectAttempts = 0
				this.flushMessageQueue()
				this.resubscribeAll()
			}
			
			this.socket.onmessage = (event) => {
				this.handleRawMessage(event.data)
			}
			
			this.socket.onclose = () => {
				console.log('[WS] H5连接关闭')
				this.isConnected = false
				this.scheduleReconnect()
			}
			
			this.socket.onerror = (err) => {
				console.error('[WS] H5连接错误:', err)
				this.isConnected = false
			}
		} catch (err) {
			console.error('[WS] H5创建WebSocket失败:', err)
			this.scheduleReconnect()
		}
		// #endif

		// #ifndef H5
		// 非H5环境使用uni.connectSocket
		this.socket = uni.connectSocket({
			url: this.url,
			success: () => {
				console.log('[WS] uni.connectSocket调用成功')
			},
			fail: (err) => {
				console.error('[WS] Socket创建失败:', err)
				this.scheduleReconnect()
			}
		})

		this.socket.onOpen(() => {
			console.log('[WS] 连接成功')
			this.isConnected = true
			this.reconnectAttempts = 0
			this.flushMessageQueue()
			this.resubscribeAll()
		})

		this.socket.onMessage((res) => {
			this.handleRawMessage(res.data)
		})

		this.socket.onClose(() => {
			console.log('[WS] 连接关闭')
			this.isConnected = false
			this.scheduleReconnect()
		})

		this.socket.onError((err) => {
			console.error('[WS] 连接错误:', err)
			this.isConnected = false
			this.scheduleReconnect()
		})
		// #endif
	}

	/**
	 * 处理原始消息数据
	 */
	handleRawMessage(data) {
		try {
			const messages = data.trim().split('\n').filter(msg => msg.trim())
			messages.forEach(msgStr => {
				try {
					const message = JSON.parse(msgStr)
					this.handleMessage(message)
				} catch (parseErr) {
					// 静默处理
				}
			})
		} catch (err) {
			// 静默处理
		}
	}

	/**
	 * 处理接收到的消息
	 */
	handleMessage(message) {
		const { type, channel, data } = message

		if (type === 'subscribed') {
			console.log('[WS] 订阅成功:', channel)
			return
		}

		if (type === 'unsubscribed') {
			console.log('[WS] 取消订阅成功:', channel)
			return
		}

		// 优化: 处理批量消息类型 (后端聚合推送)
		if (type === 'market_batch' && channel === 'market') {
			// 批量消息发送到 market_batch 频道
			const callbacks = this.subscriptions.get('market_batch')
			if (callbacks && callbacks.length > 0) {
				callbacks.forEach(callback => {
					try {
						callback(data) // data 是数组
					} catch (err) {
						console.error('[WS] 批量回调执行失败:', err)
					}
				})
			}
			return
		}

		if (type === 'channel_message' && channel) {
			const callbacks = this.subscriptions.get(channel)
			if (callbacks && callbacks.length > 0) {
				callbacks.forEach(callback => {
					try {
						callback(data)
					} catch (err) {
						console.error('[WS] 回调执行失败:', err)
					}
				})
			}
		}
	}

	/**
	 * 订阅频道
	 */
	subscribe(channel, callback) {
		console.log('[WS] 订阅频道:', channel)
		
		if (!this.subscriptions.has(channel)) {
			this.subscriptions.set(channel, [])
		}

		const callbacks = this.subscriptions.get(channel)
		if (!callbacks.includes(callback)) {
			callbacks.push(callback)
		}

		this.send({
			type: 'subscribe',
			channel: channel,
			timestamp: Date.now()
		})
	}

	/**
	 * 取消订阅频道
	 */
	unsubscribe(channel, callback) {
		if (!this.subscriptions.has(channel)) {
			return
		}

		if (callback) {
			const callbacks = this.subscriptions.get(channel)
			const index = callbacks.indexOf(callback)
			if (index > -1) {
				callbacks.splice(index, 1)
			}

			if (callbacks.length === 0) {
				this.subscriptions.delete(channel)
				this.send({
					type: 'unsubscribe',
					channel: channel,
					timestamp: Date.now()
				})
			}
		} else {
			this.subscriptions.delete(channel)
			this.send({
				type: 'unsubscribe',
				channel: channel,
				timestamp: Date.now()
			})
		}
	}

	/**
	 * 发送消息
	 */
	send(data) {
		const message = JSON.stringify(data)

		if (this.isConnected && this.socket) {
			// #ifdef H5
			try {
				// 检查连接状态
				if (this.socket.readyState === WebSocket.OPEN) {
					this.socket.send(message)
				} else if (this.socket.readyState === WebSocket.CONNECTING) {
					// 连接中，加入队列等待
					console.log('[WS] 连接中，消息加入队列')
					this.messageQueue.push(message)
				} else {
					// 连接已关闭或关闭中
					console.warn('[WS] 连接不可用，消息加入队列')
					this.isConnected = false
					this.messageQueue.push(message)
				}
			} catch (err) {
				console.error('[WS] H5发送失败:', err)
				this.messageQueue.push(message)
			}
			// #endif

			// #ifndef H5
			this.socket.send({
				data: message,
				fail: (err) => {
					this.messageQueue.push(message)
				}
			})
			// #endif
		} else {
			this.messageQueue.push(message)
		}
	}

	/**
	 * 发送队列中的消息
	 */
	flushMessageQueue() {
		console.log('[WS] 发送队列消息, 数量:', this.messageQueue.length)
		while (this.messageQueue.length > 0) {
			const message = this.messageQueue.shift()
			if (this.isConnected && this.socket) {
				// #ifdef H5
				try {
					this.socket.send(message)
				} catch (err) {
					this.messageQueue.unshift(message)
					break
				}
				// #endif

				// #ifndef H5
				this.socket.send({
					data: message,
					fail: (err) => {
						this.messageQueue.unshift(message)
					}
				})
				// #endif
			}
		}
	}

	/**
	 * 重新订阅所有频道
	 */
	resubscribeAll() {
		console.log('[WS] 重新订阅所有频道, 数量:', this.subscriptions.size)
		for (const channel of this.subscriptions.keys()) {
			this.send({
				type: 'subscribe',
				channel: channel,
				timestamp: Date.now()
			})
		}
	}

	/**
	 * 计划重连
	 */
	scheduleReconnect() {
		if (this.reconnectTimer) {
			return
		}

		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			console.error('[WS] 超过最大重连次数，停止重连')
			return
		}

		this.reconnectAttempts++
		console.log('[WS] 计划重连, 尝试次数:', this.reconnectAttempts)

		this.reconnectTimer = setTimeout(() => {
			this.reconnectTimer = null
			this.connect()
		}, this.reconnectInterval)
	}

	/**
	 * 关闭连接
	 */
	close() {
		console.log('[WS] 关闭连接')
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer)
			this.reconnectTimer = null
		}

		if (this.socket) {
			// #ifdef H5
			this.socket.close()
			// #endif

			// #ifndef H5
			this.socket.close()
			// #endif
			this.socket = null
		}

		this.isConnected = false
		this.subscriptions.clear()
		this.messageQueue = []
	}
}

// 创建全局实例
const wsClient = new WebSocketClient()

export default wsClient
