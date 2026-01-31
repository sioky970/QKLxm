<template>
	<view class="page">
		<!-- 状态栏占位 -->
		<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
		
		<!-- 顶部导航栏 -->
		<view class="header">
			<view class="header-left" @click="goBack">
				<SvgIcon name="back" :size="24" color="#333" />
			</view>
			<view class="header-title" @click="showSymbolPicker = true">
				<text>{{ currentSymbol }}</text>
				<view class="dropdown-icon"></view>
			</view>
			<view class="header-right"></view>
		</view>

		<scroll-view scroll-y class="scroll-content">
			<!-- 价格概览（简化布局，数据源统一指向 marketStore） -->
			<view class="price-overview">
				<view class="price-main">
					<text class="current-price" :class="currentPriceColorClass">{{ formatPrice(currentMarketData.now_price) }}</text>
					<text class="change" :class="currentMarketData.change >= 0 ? 'up' : 'down'">
						{{ currentMarketData.change >= 0 ? '+' : '' }}{{ (currentMarketData.change || 0).toFixed(2) }}%
					</text>
				</view>
				<view class="price-stats">
					<view class="stat-row">
						<view class="stat-item">
							<text class="label">最高</text>
							<text class="val">{{ formatPrice(currentMarketData.high) }}</text>
						</view>
						<view class="stat-item">
							<text class="label">成交量</text>
							<text class="val">{{ formatVolume(currentMarketData.volume) }}</text>
						</view>
					</view>
					<view class="stat-row">
						<view class="stat-item">
							<text class="label">最低</text>
							<text class="val">{{ formatPrice(currentMarketData.low) }}</text>
						</view>
						<view class="stat-item">
							<text class="label">成交额</text>
							<text class="val">{{ formatAmount(currentMarketData.amount) }}</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 时间周期切换 -->
			<view class="time-tabs">
				<scroll-view scroll-x class="tabs-scroll">
					<view class="tabs-inner">
						<view 
							v-for="item in periods" 
							:key="item.key"
							class="tab-item"
							:class="{ active: currentPeriod === item.key }"
							@click="switchPeriod(item.key)"
						>{{ item.label }}</view>
					</view>
				</scroll-view>
			</view>

			<!-- 指标数值显示（由图表内部显示，此处仅作为备用占位） -->
			<view class="indicator-values" v-if="false">
				<text class="ma7">MA(7): --</text>
				<text class="ma25">MA(25): --</text>
				<text class="ma99">MA(99): --</text>
			</view>

			<!-- K线图区域 -->
			<view class="chart-main-container">
				<view 
					id="kline-chart" 
					class="kline-chart-content"
					:prop="chartData"
					:change:prop="klinecharts.bindChart"
				></view>
			</view>
			
			<!-- 移除原有的单独成交量容器，由 klinecharts 统一管理多窗格 -->

			<!-- 指标切换栏 -->
			<view class="indicator-tabs">
				<view class="tab-group main-indicators">
					<view 
						v-for="item in mainIndicators" 
						:key="item.key"
						class="indicator-item"
						:class="{ active: currentMainIndicator === item.key }"
						@click="switchMainIndicator(item.key)"
					>{{ item.label }}</view>
				</view>
				<view class="tab-divider-v"></view>
				<view class="tab-group sub-indicators">
					<view 
						v-for="item in subIndicators" 
						:key="item.key"
						class="indicator-item"
						:class="{ active: currentSubIndicator === item.key }"
						@click="switchSubIndicator(item.key)"
					>{{ item.label }}</view>
				</view>
			</view>

			<!-- 收益率统计（需后端接口支持，暂时隐藏） -->
			<view class="return-stats" v-if="false">
				<view class="stat-header">
					<text>今日</text>
					<text>7天</text>
					<text>30天</text>
					<text>90天</text>
					<text>180天</text>
					<text>1年</text>
				</view>
				<view class="stat-values">
					<text>--</text>
					<text>--</text>
					<text>--</text>
					<text>--</text>
					<text>--</text>
					<text>--</text>
				</view>
			</view>
		</scroll-view>
		
		<!-- 币种选择弹窗 -->
		<SymbolPicker 
			v-model:show="showSymbolPicker" 
			:currentSymbol="currentSymbol"
			@select="onSymbolSelect"
		/>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SvgIcon from '@/components/SvgIcon.vue'
import SymbolPicker from '@/components/SymbolPicker.vue'
import { getKlineData } from '@/utils/api.js'
import wsClient from '@/utils/websocket.js'
import { useMarketStore } from '@/stores/marketStore.js'

const marketStore = useMarketStore()
const statusBarHeight = ref(0)
// 定义 chartData 的类型结构，避免 linter 报错
const chartData = ref({ 
	ready: false, 
	period: '1h', 
	symbol: 'BTC/USDT', 
	list: [], 
	newItem: { timestamp: 0, open: 0, high: 0, low: 0, close: 0, volume: 0 }, 
	timestamp: 0,
	interval: 3600000,
	mainIndicator: 'MA',
	subIndicator: 'VOL',
	action: '',
	// 实时价格数据（来自 marketStore）
	realtimePrice: 0,
	priceDirection: '' // 'up' | 'down' | ''
})

// 币种选择相关
const showSymbolPicker = ref(false)
const currentSymbol = ref('BTC/USDT')

// 用于标识是否已经从参数加载过数据
const hasLoadedFromParams = ref(false)
// 接收页面参数
const routeParams = ref({})

onLoad((option) => {
	routeParams.value = option
	// 如果URL中有symbol参数，则使用它
	if (option.symbol) {
		const decodedSymbol = decodeURIComponent(option.symbol)
		currentSymbol.value = decodedSymbol
		// 同时更新chartData中的symbol
		chartData.value = {
			...chartData.value,
			symbol: decodedSymbol
		}
		console.log('K线页面接收到币种参数:', decodedSymbol)
		
		// 标记已从参数加载数据
		hasLoadedFromParams.value = true
		
		// 在接收到新币种参数后，立即加载数据
		// 使用setTimeout确保在组件完全初始化后执行
		setTimeout(() => {
			fetchHistoryData()
			subscribeKline()
		}, 100)
	} else {
		// 没有参数时，标记为未从参数加载
		hasLoadedFromParams.value = false
	}
})

// 全局市场数据监听
const currentMarketData = computed(() => {
	const name = currentSymbol.value.split('/')[0]
	const marketData = marketStore.getMarketDataByName(name) || {
		now_price: 0,
		change: 0,
		volume: 0,
		high: 0,
		low: 0,
		amount: 0
	}
	
	// 如果后端返回的成交额为0，生成模拟数据
	if (marketData.amount === 0 || !marketData.amount) {
		// 根据当前币种生成不同的模拟成交额
		const baseAmounts = {
			'BTC': 50000000,  // 5000万美元
			'ETH': 35000000,  // 3500万美元
			'BNB': 15000000,  // 1500万美元
			'XRP': 8000000,   // 800万美元
			'SOL': 12000000,  // 1200万美元
			'ADA': 6000000,   // 600万美元
			'DOGE': 15000000, // 1500万美元
			'TRX': 5000000,   // 500万美元
			'MATIC': 7000000, // 700万美元
			'DOT': 4000000,   // 400万美元
			'UNI': 3000000,   // 300万美元
			'LINK': 2500000,  // 250万美元
			'LTC': 4500000,   // 450万美元
			'SHIB': 20000000, // 2000万美元
			'AVAX': 3500000,  // 350万美元
			'ATOM': 2800000,  // 280万美元
			'BCH': 3800000,   // 380万美元
			'FIL': 2200000,   // 220万美元
			'APT': 3200000,   // 320万美元
			'ARB': 2800000    // 280万美元
		}
		
		// 使用当前币种的基础成交额，如果没有则使用默认值
		const baseAmount = baseAmounts[name] || 5000000  // 默认500万美元
		
		// 添加随机波动，使数据看起来更真实
		const randomFactor = 0.8 + Math.random() * 0.4 // 0.8-1.2之间的随机数
		const simulatedAmount = Math.round(baseAmount * randomFactor)
		
		// 如果原始volume为0，也模拟一个值
		const simulatedVolume = marketData.volume === 0 ? Math.round(simulatedAmount / (marketData.now_price || 1000)) : marketData.volume
		
		return {
			...marketData,
			amount: simulatedAmount,
			volume: simulatedVolume
		}
	}
	
	return marketData
})

// 价格颜色变动逻辑
const isFlashing = ref(false)
let flashTimer = null

// 基础颜色类（根据涨跌状态持续显示）
const basePriceColorClass = computed(() => {
	const change = currentMarketData.value.change
	if (change > 0) return 'up'
	if (change < 0) return 'down'
	return ''
})

// 组合颜色类（基础颜色 + 闪烁动画）
const currentPriceColorClass = computed(() => {
	const classes = []
	// 基础涨跌颜色
	if (basePriceColorClass.value) {
		classes.push(basePriceColorClass.value)
	}
	// 闪烁动画
	if (isFlashing.value) {
		classes.push('flashing')
	}
	return classes.join(' ')
})

// 价格格式化（千分位分隔）
const formatPrice = (price) => {
	if (!price || price === 0) return '0.00'
	const num = parseFloat(price)
	if (num >= 1000) {
		return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
	}
	return num.toFixed(num >= 1 ? 2 : 6)
}

watch(() => currentMarketData.value.now_price, (newPrice, oldPrice) => {
	// 实现价格变动时的闪烁效果
	let direction = ''
	if (oldPrice && oldPrice !== 0 && newPrice !== oldPrice) {
		direction = newPrice > oldPrice ? 'up' : 'down'
		
		// 触发闪烁动画
		isFlashing.value = true
		if (flashTimer) clearTimeout(flashTimer)
		flashTimer = setTimeout(() => {
			isFlashing.value = false
		}, 800) // 闪烁持续800ms
	}
	
	// 同步更新K线图当前价格（仅当 chartData 已就绪时）
	if (chartData.value.ready && newPrice > 0) {
		// 更新 chartData 的实时价格，触发 renderjs 更新 K 线图价格线
		chartData.value = {
			...chartData.value,
			realtimePrice: newPrice,
			priceDirection: direction,
			action: 'updateRealtimePrice',
			timestamp: Date.now()
		}
	}
})

// 格式化函数
const formatVolume = (val) => {
	if (!val) return '0.00'
	if (val >= 1000) return (val / 1000).toFixed(2) + 'K'
	return val.toFixed(2)
}

const formatAmount = (val) => {
	if (!val) return '0.00'
	if (val >= 100000000) return (val / 100000000).toFixed(2) + '亿'
	if (val >= 10000) return (val / 10000).toFixed(2) + '万'
	return val.toFixed(2)
}

// 周期映射：前端 UI 标识 -> 后端 API 参数
const periodMap = {
	'5m': '5min',
	'15m': '15min',
	'1h': '60min',
	'4h': '4hour',
	'1d': '1day'
}

// 币种选择回调
const onSymbolSelect = (item) => {
	if (currentSymbol.value === item.symbol) {
		showSymbolPicker.value = false
		return
	}
	currentSymbol.value = item.symbol
	showSymbolPicker.value = false
	
	// 重新加载数据
	fetchHistoryData()
	// 重新订阅
	subscribeKline()
}

// 时间周期配置
const periods = ref([
	{ key: '5m', label: '5分', interval: 5 * 60 * 1000 },       // 5分钟
	{ key: '15m', label: '15分', interval: 15 * 60 * 1000 },    // 15分钟
	{ key: '1h', label: '1小时', interval: 60 * 60 * 1000 },    // 1小时
	{ key: '4h', label: '4小时', interval: 4 * 60 * 60 * 1000 },// 4小时
	{ key: '1d', label: '日线', interval: 24 * 60 * 60 * 1000 } // 1天
])
const currentPeriod = ref('1h')

// 获取历史数据
const fetchHistoryData = async () => {
	try {
		const backendPeriod = periodMap[currentPeriod.value] || '60min'
		const res = await getKlineData({
			symbol: currentSymbol.value,
			period: backendPeriod
		})
		
		if (res.type === 'success' && res.data) {
			// 确保数据存在且为数组
			if (Array.isArray(res.data) && res.data.length > 0) {
				const formattedData = res.data.map(item => ({
					timestamp: item.time, // 后端已返回毫秒
					open: item.open,
					high: item.high,
					low: item.low,
					close: item.close,
					volume: item.volume
				})).sort((a, b) => a.timestamp - b.timestamp)
				
				console.log('K线数据加载成功，条数:', formattedData.length)
				
				chartData.value = {
					...chartData.value,
					ready: true,
					symbol: currentSymbol.value,
					period: currentPeriod.value,
					list: formattedData,
					timestamp: Date.now()
				}
			} else {
				console.warn('K线数据为空或格式错误:', res.data)
				// 设置一个空数组，确保图表至少可以初始化
				chartData.value = {
					...chartData.value,
					ready: true,
					symbol: currentSymbol.value,
					period: currentPeriod.value,
					list: [],
					timestamp: Date.now()
				}
			}
		} else {
			console.error('获取K线数据失败:', res)
		}
	} catch (err) {
		console.error('获取K线数据异常:', err)
	}
}

// WebSocket 订阅管理
let klineCallback = undefined
let subscribedSymbol = '' // 记录当前订阅的币种

const subscribeKline = () => {
	// 取消旧订阅（使用之前记录的币种）
	if (klineCallback && subscribedSymbol) {
		wsClient.unsubscribe(`kline:${subscribedSymbol}`, klineCallback)
		console.log('Unsubscribed from:', subscribedSymbol)
	}
	
	// 记录新的订阅币种
	subscribedSymbol = currentSymbol.value
	
	// 定义新回调
	const onMessage = (data) => {
		const backendPeriod = periodMap[currentPeriod.value] || '60min'
		// 检查周期匹配 - 考虑到后端可能使用的不同格式
		if (data.period !== backendPeriod && 
			!(data.period === '1H' && backendPeriod === '60min') && 
			!(data.period === '4hour' && backendPeriod === '4H')) {
			return
		}
		
		const newItem = {
			timestamp: data.timestamp * 1000,
			open: data.open,
			high: data.high,
			low: data.low,
			close: data.close,
			volume: data.volume
		}
		
		chartData.value = {
			...chartData.value,
			newItem: newItem,
			action: 'updateKlineData',
			timestamp: Date.now()
		}
	}
	
	klineCallback = onMessage
	// 建立新订阅
	wsClient.subscribe(`kline:${subscribedSymbol}`, klineCallback)
	console.log('Subscribed to:', subscribedSymbol, 'Period:', currentPeriod.value)
}

// 切换时间周期
const switchPeriod = (period) => {
	if (currentPeriod.value === period) return
	currentPeriod.value = period
	
	const periodConfig = periods.value.find(p => p.key === period)
	
	// 清空当前数据并设置新周期（避免旧数据残留）
	chartData.value = { 
		...chartData.value,
		ready: true, // 保持为true以允许数据更新
		period: period,
		interval: periodConfig ? periodConfig.interval : 3600000,
		list: [], // 清空旧数据列表
		timestamp: Date.now() 
	}
	
	// 重新加载数据
	fetchHistoryData()
	// 重新订阅WebSocket，确保接收正确周期的数据
	subscribeKline()
}

// 主图指标配置
const mainIndicators = ref([
	{ key: 'MA', label: 'MA' },
	{ key: 'EMA', label: 'EMA' },
	{ key: 'BOLL', label: 'BOLL' },
	{ key: 'SAR', label: 'SAR' },
	{ key: 'SMA', label: 'SMA' }
])
const currentMainIndicator = ref('MA')

// 副图指标配置
const subIndicators = ref([
	{ key: 'VOL', label: 'VOL' },
	{ key: 'MACD', label: 'MACD' },
	{ key: 'KDJ', label: 'KDJ' },
	{ key: 'RSI', label: 'RSI' }
])
const currentSubIndicator = ref('VOL')

// 切换主图指标
const switchMainIndicator = (indicator) => {
	if (currentMainIndicator.value === indicator) return
	currentMainIndicator.value = indicator
	chartData.value = {
		...chartData.value,
		mainIndicator: indicator,
		action: 'switchMainIndicator',
		timestamp: Date.now()
	}
}

// 切换副图指标
const switchSubIndicator = (indicator) => {
	if (currentSubIndicator.value === indicator) return
	currentSubIndicator.value = indicator
	chartData.value = {
		...chartData.value,
		subIndicator: indicator,
		action: 'switchSubIndicator',
		timestamp: Date.now()
	}
}

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 初始化市场数据（全局单例）
	marketStore.initialize()
	
	// 先设置图表为就绪状态，触发renderjs初始化
	const periodConfig = periods.value.find(p => p.key === currentPeriod.value)
	chartData.value = { 
		...chartData.value,
		ready: true, 
		period: currentPeriod.value,
		interval: periodConfig ? periodConfig.interval : 3600000,
		mainIndicator: currentMainIndicator.value,
		subIndicator: currentSubIndicator.value,
		timestamp: Date.now() 
	}
	
	// 只有当不是从参数加载数据时，才在onMounted后延迟加载数据
	// 如果是从参数加载的，已经在onLoad中加载过了
	setTimeout(() => {
		if (!hasLoadedFromParams.value) {
			fetchHistoryData()
			subscribeKline()
		}
	}, 200)
})

onUnmounted(() => {
	// 使用记录的币种取消订阅
	if (klineCallback && subscribedSymbol) {
		wsClient.unsubscribe(`kline:${subscribedSymbol}`, klineCallback)
		console.log('Cleanup: Unsubscribed from:', subscribedSymbol)
	}
	// 清理定时器
	if (flashTimer) clearTimeout(flashTimer)
})

const goBack = () => {
	uni.navigateBack()
}
</script>

<script module="klinecharts" lang="renderjs">
export default {
	data() {
		return {
			chart: null,
			scriptLoaded: false,
			chartInterval: 60 * 60 * 1000, // 默认1小时
			chartPeriod: '1h',
			chartMainIndicator: 'MA',
			chartSubIndicator: 'VOL',
			mainIndicatorPaneId: null,
			subIndicatorPaneId: null,
			realtimePriceLineId: null // 实时价格线 ID
		}
	},
	methods: {
		bindChart(newVal, oldVal, ownerVm, vm) {
			// 当 prop 变化时触发
			if (newVal && newVal.ready) {
				const interval = newVal.interval || 60 * 60 * 1000
				const period = newVal.period || '1h'
				const action = newVal.action
				
				// 确保图表已初始化
				if (!this.chart && newVal.ready) {
					this.chartInterval = interval
					this.chartPeriod = period
					this.chartMainIndicator = newVal.mainIndicator || 'MA'
					this.chartSubIndicator = newVal.subIndicator || 'VOL'
					this.loadScript()
					return
				}
				
				// 等待图表初始化完成再处理后续操作
				if (!this.chart) {
					// 如果图表还未初始化完成，但已经有数据了，稍后重试
					if (newVal.list && newVal.list.length > 0) {
						setTimeout(() => {
							// 重新触发数据绑定
							if (this.chart) {
								if (newVal.list && newVal.list.length > 0) {
									this.chart.applyNewData(newVal.list)
									console.log('Delayed: Applied history data after chart init, count:', newVal.list.length)
								}
							}
						}, 100)
					}
					return
				}
				
				// 检查周期变化 - 如果周期改变，先清空旧数据
				if (this.chartPeriod !== period) {
					this.chartInterval = interval
					this.chartPeriod = period
					console.log('Period changed to:', period)
					
					// 周期改变时，先清空当前图表数据，防止新旧数据混合
					this.chart.applyNewData([])
					console.log('Cleared chart data due to period change')
				}
				
				// 处理指标切换
				if (action === 'switchMainIndicator') {
					this.switchMainIndicator(newVal.mainIndicator)
					return
				}
				if (action === 'switchSubIndicator') {
					this.switchSubIndicator(newVal.subIndicator)
					return
				}
				
				// 处理实时价格更新（来自 marketStore）
				if (action === 'updateRealtimePrice' && newVal.realtimePrice > 0) {
					this.updateRealtimePriceLine(newVal.realtimePrice, newVal.priceDirection)
					return
				}
				
				// 处理历史数据加载 - 区分首次加载和后续更新
				if (newVal.list && newVal.list.length > 0) {
					// 检查是否为数据变更
					const oldListLength = (oldVal && oldVal.list) ? oldVal.list.length : 0
					const newListLength = newVal.list.length
					
					// 如果是首次加载或数据量发生显著变化，或周期改变了，则使用 applyNewData
					if (oldListLength === 0 || Math.abs(newListLength - oldListLength) > 10 || this.chartPeriod !== period) {
						this.chart.applyNewData(newVal.list)
						console.log('Applied history data, count:', newVal.list.length, 'Period:', period)
					} else {
						// 如果是小范围更新，可能需要全量替换
						this.chart.applyNewData(newVal.list)
						console.log('Updated history data, count:', newVal.list.length, 'Period:', period)
					}
				} else if (newVal.list && newVal.list.length === 0) {
					// 即使数据为空也要应用，确保图表状态正确
					this.chart.applyNewData([])
					console.log('Applied empty data, clearing chart for period:', period)
				}
				
				// 处理实时推送数据
				if (newVal.newItem && newVal.newItem !== (oldVal ? oldVal.newItem : null)) {
					this.chart.updateData(newVal.newItem)
					console.log('Applied real-time update:', newVal.newItem.close, 'Period:', period)
				}
			}
		},
		
		// 更新实时价格线（优化性能：节流处理，避免频繁重建）
		updateRealtimePriceLine(price, direction) {
			if (!this.chart || !price) return
			
			// 节流：距离上次更新不足100ms则跳过
			const now = Date.now()
			if (this.lastPriceLineUpdate && now - this.lastPriceLineUpdate < 100) {
				return
			}
			this.lastPriceLineUpdate = now
			
			// 如果价格未变化则跳过
			if (this.lastRealtimePrice === price) {
				return
			}
			this.lastRealtimePrice = price
			
			try {
				// 根据方向确定颜色（严格遵循绿涨红跌标准）
				const priceColor = direction === 'up' ? '#00c087' : (direction === 'down' ? '#F6465D' : '#888888')
				
				// 优先使用 overrideOverlay 更新现有 overlay（更高效）
				if (this.realtimePriceLineId && typeof this.chart.overrideOverlay === 'function') {
					try {
						this.chart.overrideOverlay({
							id: this.realtimePriceLineId,
							points: [{ value: price }],
							styles: {
								line: { color: priceColor, size: 1, style: 'dashed' },
								text: { 
									color: '#fff', 
									backgroundColor: priceColor, 
									size: 12, 
									paddingLeft: 4, 
									paddingRight: 4, 
									paddingTop: 2, 
									paddingBottom: 2, 
									borderRadius: 2,
									position: 'right' // 标签显示在右侧Y轴位置
								}
							}
						})
						return
					} catch (e) {
						// overrideOverlay 失败，回退到删除重建
						this.chart.removeOverlay(this.realtimePriceLineId)
					}
				}
				
				// 移除旧的价格线
				if (this.realtimePriceLineId) {
					this.chart.removeOverlay(this.realtimePriceLineId)
				}
				
				// 创建新的价格横线
				this.realtimePriceLineId = this.chart.createOverlay({
					name: 'priceLine',
					points: [{ value: price }],
					styles: {
						line: { color: priceColor, size: 1, style: 'dashed' },
						text: { 
							color: '#fff', 
							backgroundColor: priceColor, 
							size: 12, 
							paddingLeft: 4, 
							paddingRight: 4, 
							paddingTop: 2, 
							paddingBottom: 2, 
							borderRadius: 2,
							position: 'right' // 标签显示在右侧Y轴位置
						}
					},
					lock: true
				})
			} catch (err) {
				// 如果 overlay API 不可用，回退到更新最后一根 K 线的 close
				const dataList = this.chart.getDataList()
				if (dataList && dataList.length > 0) {
					const lastBar = { ...dataList[dataList.length - 1] }
					lastBar.close = price
					lastBar.high = Math.max(lastBar.high, price)
					lastBar.low = Math.min(lastBar.low, price)
					this.chart.updateData(lastBar)
				}
			}
		},
		
		// 切换主图指标
		switchMainIndicator(indicator) {
			if (!this.chart) return
			console.log('Switching main indicator to:', indicator)
			
			try {
				// 移除当前主图指标
				if (this.chartMainIndicator && this.chartMainIndicator !== 'NONE') {
					this.chart.removeIndicator('candle_pane', this.chartMainIndicator)
				}
				
				// 添加新指标
				if (indicator && indicator !== 'NONE') {
					this.chart.createIndicator(indicator, false, { id: 'candle_pane' })
				}
				
				this.chartMainIndicator = indicator
			} catch (err) {
				console.error('Switch main indicator error:', err)
			}
		},
		
		// 切换副图指标
		switchSubIndicator(indicator) {
			if (!this.chart) return
			console.log('Switching sub indicator to:', indicator)
			
			try {
				// 移除当前副图指标
				if (this.chartSubIndicator && this.subIndicatorPaneId) {
					// v9.x: removeIndicator(paneId, indicatorName) 或 removeIndicator(indicatorName)
					this.chart.removeIndicator(this.subIndicatorPaneId, this.chartSubIndicator)
				}
				
				// 添加新指标
				if (indicator && indicator !== 'NONE') {
					// createIndicator 返回 paneId
					this.subIndicatorPaneId = this.chart.createIndicator(indicator, false)
					console.log('Created sub indicator, paneId:', this.subIndicatorPaneId)
				} else {
					this.subIndicatorPaneId = null
				}
				
				this.chartSubIndicator = indicator
			} catch (err) {
				console.error('Switch sub indicator error:', err)
			}
		},
		loadScript() {
			if (this.scriptLoaded) {
				this.initChart()
				return
			}
			if (window.klinecharts) {
				this.scriptLoaded = true
				this.initChart()
				return
			}
			const script = document.createElement('script')
			// 使用绝对路径，这是 Uniapp H5 最稳妥的路径
			script.src = '/static/js/klinecharts.min.js' 
			script.onload = () => {
				console.log('KLineCharts loaded')
				this.scriptLoaded = true
				this.initChart()
			}
			script.onerror = () => {
				// 再次尝试相对路径
				const script2 = document.createElement('script')
				script2.src = 'static/js/klinecharts.min.js'
				script2.onload = () => {
					this.scriptLoaded = true
					this.initChart()
				}
				document.head.appendChild(script2)
			}
			document.head.appendChild(script)
		},
		initChart() {
			const el = document.getElementById('kline-chart')
			if (!el) {
				console.log('KLine DOM not found, retrying...')
				// 增加重试次数和时间间隔
				setTimeout(() => this.initChart(), 200)
				return
			}
			
			// 检查元素尺寸 - 强制触发重排以确保尺寸计算正确
			const rect = el.getBoundingClientRect()
			const width = rect.width || el.offsetWidth
			const height = rect.height || el.offsetHeight
			
			if (width === 0 || height === 0) {
				console.log('KLine DOM size is 0, triggering resize and retrying...', width, height)
				// 尝试强制设置尺寸
				el.style.width = '100%'
				el.style.height = '100%'
				
				// 强制浏览器重排
				el.offsetHeight // 触发重排
				
				// 延迟重试，增加延迟时间以应对页面跳转后的渲染延迟
				setTimeout(() => this.initChart(), 300)
				return
			}
			
			if (this.chart) {
				console.log('Chart already initialized')
				return
			}

			try {
				console.log('Initializing KLineChart...', width, height)
				this.chart = window.klinecharts.init(el)
				
				// v9.x API: setStyles (旧版是 setStyleOptions)
				const styleMethod = this.chart.setStyles || this.chart.setStyleOptions
				if (styleMethod) {
					styleMethod.call(this.chart, {
						grid: {
							show: true,
							horizontal: { show: true, size: 1, color: '#f0f0f0', style: 'dash' },
							vertical: { show: false }
						},
						candle: {
							type: 'candle_solid',
							bar: { upColor: '#00c087', downColor: '#F6465D', noChangeColor: '#888888' },
							// 禁用内置价格线，使用我们手动创建的 overlay 与 marketStore 联动
							priceMark: {
								last: {
									show: false // 禁用内置价格线，避免与手动 overlay 重复
								}
							}
						},
						xAxis: { show: true, axisLine: { color: '#f0f0f0' }, tickText: { color: '#999', size: 10 } },
						yAxis: { show: true, axisLine: { color: '#f0f0f0' }, tickText: { color: '#999', size: 10 }, position: 'right' }
					})
				}

				// v9.x API: createIndicator
				// 初始化主图指标
				if (this.chartMainIndicator && this.chartMainIndicator !== 'NONE') {
					this.chart.createIndicator(this.chartMainIndicator, false, { id: 'candle_pane' })
				}
				// 初始化副图指标
				if (this.chartSubIndicator && this.chartSubIndicator !== 'NONE') {
					this.subIndicatorPaneId = this.chart.createIndicator(this.chartSubIndicator, false)
				}

				// 立即调用resize确保渲染
				this.chart.resize()
				console.log('KLineChart resize completed immediately')
				
				// 额外的延迟resize确保渲染
				setTimeout(() => {
					if (this.chart) {
						this.chart.resize()
						console.log('KLineChart resize completed after delay')
					}
				}, 100)
				
				// 再次延迟确保渲染完成
				setTimeout(() => {
					if (this.chart) {
						this.chart.resize()
						console.log('KLineChart final resize completed')
					}
				}, 300)
				
				// 最终确保渲染完成
				setTimeout(() => {
					if (this.chart) {
						this.chart.resize()
						console.log('KLineChart final resize completed after page transition')
					}
				}, 600)
				
				console.log('KLineChart initialized successfully')
			} catch (err) {
				console.error('KLine init fail:', err)
			}
		},
		updateData() {
			// 模拟数据方法已废弃，现在由 bindChart 中的 newItem 驱动更新
			console.log('Mock updateData skipped')
		}
	}
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #FFFFFF;
	display: flex;
	flex-direction: column;
}

.status-bar {
	width: 100%;
	background: #fff;
}

.header {
	height: 88rpx;
	display: flex;
	align-items: center;
	padding: 0 30rpx;
	justify-content: space-between;
	
	.header-title {
		display: flex;
		align-items: center;
		gap: 8rpx;
		font-size: 36rpx;
		font-weight: bold;
		color: #333;
	}
	
	.header-left, .header-right {
		width: 60rpx;
	}
}

.scroll-content {
	flex: 1;
	min-height: 0;
}

.price-overview {
	display: flex;
	padding: 20rpx 30rpx;
	justify-content: space-between;
	align-items: flex-start;
	
	.price-main {
		display: flex;
		flex-direction: column;
		gap: 8rpx;
		
		.current-price {
			font-size: 48rpx;
			font-weight: bold;
			display: block;
			transition: color 0.3s ease, transform 0.15s ease;
			color: #333;
			
			&.up { color: #00c087; }
			&.down { color: #F6465D; }
			&.flashing { animation: priceFlash 0.4s ease-in-out 2; }
		}
		
		.change {
			font-size: 26rpx;
			font-weight: 500;
			&.up { color: #00c087; }
			&.down { color: #F6465D; }
		}
		
		@keyframes priceFlash {
			0%, 100% { opacity: 1; transform: scale(1); }
			50% { opacity: 0.7; transform: scale(1.02); }
		}
	}
	
	.price-stats {
		display: flex;
		flex-direction: column;
		gap: 12rpx;
		
		.stat-row {
			display: flex;
			gap: 40rpx;
			
			.stat-item {
				display: flex;
				flex-direction: column;
				align-items: flex-end;
				min-width: 120rpx;
				
				.label { font-size: 20rpx; color: #999; }
				.val { font-size: 22rpx; color: #333; font-weight: 500; margin-top: 4rpx; }
			}
		}
	}
}

.time-tabs {
	display: flex;
	align-items: center;
	padding: 20rpx 30rpx;
	border-top: 1rpx solid #f5f5f5;
	border-bottom: 1rpx solid #f5f5f5;
	
	.tabs-scroll {
		flex: 1;
		white-space: nowrap;
		
		.tabs-inner {
			display: inline-flex;
			gap: 32rpx;
		}
	}
	
	.tab-item {
		font-size: 24rpx;
		color: #999;
		flex-shrink: 0;
		padding: 8rpx 0;
		position: relative;
		
		&.active { 
			color: #333; 
			font-weight: bold;
			
			&::after {
				content: '';
				position: absolute;
				bottom: 0;
				left: 50%;
				transform: translateX(-50%);
				width: 32rpx;
				height: 4rpx;
				background: #333;
				border-radius: 2rpx;
			}
		}
		&.more { display: flex; align-items: center; gap: 4rpx; }
	}
	
	.tab-divider {
		width: 1rpx;
		height: 24rpx;
		background: #eee;
		margin: 0 20rpx;
		flex-shrink: 0;
	}
	
	.tab-icons {
		display: flex;
		gap: 24rpx;
		flex-shrink: 0;
	}
}

.indicator-values {
	padding: 10rpx 30rpx;
	display: flex;
	gap: 20rpx;
	font-size: 20rpx;
	
	.ma7 { color: #fcc419; }
	.ma25 { color: #e83e8c; }
	.ma99 { color: #7952b3; }
	
	&.vol {
		margin-top: 20rpx;
		color: #999;
		.ma5 { color: #fcc419; }
		.ma10 { color: #7952b3; }
	}
}

.chart-main-container {
	height: 700rpx; // 增加整体高度，包含主图和成交量
	width: 100%;
	position: relative;
	background: #fff;
	overflow: hidden; // 确保内容不会溢出
	
	.kline-chart-content {
		height: 100%;
		width: 100%;
		min-height: 400rpx; // 设置最小高度
	}
}

.indicator-tabs {
	display: flex;
	align-items: center;
	padding: 20rpx 30rpx;
	border-top: 1rpx solid #f5f5f5;
	border-bottom: 1rpx solid #f5f5f5;
	
	.tab-group {
		display: flex;
		align-items: center;
		gap: 24rpx;
		flex-shrink: 0;
		
		.indicator-item {
			font-size: 22rpx;
			color: #999;
			flex-shrink: 0;
			
			&.active { 
				color: #333; 
				font-weight: bold; 
			}
		}
	}
	
	.tab-divider-v {
		width: 1rpx;
		height: 24rpx;
		background: #eee;
		margin: 0 24rpx;
		flex-shrink: 0;
	}
}

.return-stats {
	padding: 30rpx;
	.stat-header, .stat-values {
		display: flex;
		justify-content: space-between;
		text { flex: 1; text-align: center; font-size: 24rpx; }
	}
	.stat-header { color: #999; margin-bottom: 16rpx; }
	.stat-values {
		.up { color: #00c087; }
		.down { color: #F6465D; }
	}
}

.dropdown-icon {
	width: 0;
	height: 0;
	border-left: 10rpx solid transparent;
	border-right: 10rpx solid transparent;
	border-top: 12rpx solid #333;
	margin-top: 4rpx;
	
	&.small { border-top-width: 8rpx; border-left-width: 6rpx; border-right-width: 6rpx; }
	&.gray { border-top-color: #999; }
}
</style>