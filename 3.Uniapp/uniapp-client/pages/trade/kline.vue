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
			<!-- 价格概览 -->
			<view class="price-overview">
				<view class="price-left">
					<text class="current-price">89,354.83</text>
					<view class="price-sub">
						<text class="price-usd">$89,354.83</text>
						<text class="change down">-0.07%</text>
					</view>
					<view class="tags">
						<text class="tag">POW</text>
						<text class="tag">成交额</text>
						<text class="tag">价格保护</text>
					</view>
				</view>
				<view class="price-right">
					<view class="stat-row">
						<view class="stat-item">
							<text class="label">24h最高价</text>
							<text class="val">89,957.39</text>
						</view>
						<view class="stat-item">
							<text class="label">24h成交量(BTC)</text>
							<text class="val">4,563.23</text>
						</view>
					</view>
					<view class="stat-row">
						<view class="stat-item">
							<text class="label">24h最低价</text>
							<text class="val">89,162.08</text>
						</view>
						<view class="stat-item">
							<text class="label">24h成交额(USDT)</text>
							<text class="val">4.09亿</text>
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

			<!-- 指标数值显示 -->
			<view class="indicator-values">
				<text class="ma7">MA(7): 89,347.67</text>
				<text class="ma25">MA(25): 89,527.91</text>
				<text class="ma99">MA(99): 89,513.48</text>
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

			<!-- 收益率统计 -->
			<view class="return-stats">
				<view class="stat-header">
					<text>今日</text>
					<text>7天</text>
					<text>30天</text>
					<text>90天</text>
					<text>180天</text>
					<text>1年</text>
				</view>
				<view class="stat-values">
					<text class="up">0.08%</text>
					<text class="down">-6.51%</text>
					<text class="up">1.11%</text>
					<text class="down">-21.41%</text>
					<text class="down">-24.39%</text>
					<text class="down">-15.59%</text>
				</view>
			</view>
		</scroll-view>

		<!-- 底部操作栏 -->
		<view class="footer-actions">
			<button class="action-btn buy" @click="goTrade('buy')">买入</button>
			<button class="action-btn sell" @click="goTrade('sell')">卖出</button>
		</view>
		
		<!-- 币种选择弹窗 -->
		<SymbolPicker 
			v-model:show="showSymbolPicker" 
			:currentSymbol="currentSymbol"
			@select="onSymbolSelect"
		/>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import SvgIcon from '@/components/SvgIcon.vue'
import SymbolPicker from '@/components/SymbolPicker.vue'

const statusBarHeight = ref(0)
const chartData = ref({ ready: false, period: '1h' })

// 币种选择相关
const showSymbolPicker = ref(false)
const currentSymbol = ref('BTC/USDT')

// 币种选择回调
const onSymbolSelect = (item) => {
	currentSymbol.value = item.symbol
	// 触发图表重新加载
	chartData.value = {
		...chartData.value,
		symbol: item.symbol,
		timestamp: Date.now()
	}
}

// 时间周期配置
const periods = ref([
	{ key: '1m', label: '1分', interval: 60 * 1000 },           // 1分钟
	{ key: '5m', label: '5分', interval: 5 * 60 * 1000 },       // 5分钟
	{ key: '15m', label: '15分', interval: 15 * 60 * 1000 },    // 15分钟
	{ key: '1h', label: '1小时', interval: 60 * 60 * 1000 },    // 1小时
	{ key: '4h', label: '4小时', interval: 4 * 60 * 60 * 1000 },// 4小时
	{ key: '1d', label: '日线', interval: 24 * 60 * 60 * 1000 } // 1天
])
const currentPeriod = ref('1h')

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

// 切换时间周期
const switchPeriod = (period) => {
	if (currentPeriod.value === period) return
	currentPeriod.value = period
	// 通知 renderjs 更新数据
	const periodConfig = periods.value.find(p => p.key === period)
	chartData.value = { 
		ready: true, 
		period: period,
		interval: periodConfig?.interval || 60 * 60 * 1000,
		timestamp: Date.now() 
	}
}

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
	
	// 延迟触发图表初始化，确保 DOM 已渲染
	setTimeout(() => {
		const periodConfig = periods.value.find(p => p.key === currentPeriod.value)
		chartData.value = { 
			ready: true, 
			period: currentPeriod.value,
			interval: periodConfig?.interval || 60 * 60 * 1000,
			mainIndicator: currentMainIndicator.value,
			subIndicator: currentSubIndicator.value,
			timestamp: Date.now() 
		}
	}, 100)
})

const goBack = () => {
	uni.navigateBack()
}

const goTrade = (side) => {
	// 存储交易方向，供 trade 页面读取
	uni.setStorageSync('tradeSide', side)
	uni.switchTab({
		url: '/pages/trade/trade'
	})
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
			subIndicatorPaneId: null
		}
	},
	methods: {
		bindChart(newVal, oldVal, ownerVm, vm) {
			// 当 prop 变化时触发
			if (newVal && newVal.ready) {
				const interval = newVal.interval || 60 * 60 * 1000
				const period = newVal.period || '1h'
				const action = newVal.action
				
				// 处理指标切换
				if (this.chart && action === 'switchMainIndicator') {
					this.switchMainIndicator(newVal.mainIndicator)
					return
				}
				if (this.chart && action === 'switchSubIndicator') {
					this.switchSubIndicator(newVal.subIndicator)
					return
				}
				
				// 如果图表已初始化且周期变化，只更新数据
				if (this.chart && this.chartPeriod !== period) {
					this.chartInterval = interval
					this.chartPeriod = period
					this.updateData()
					console.log('Period changed to:', period)
				} else if (!this.chart) {
					this.chartInterval = interval
					this.chartPeriod = period
					this.chartMainIndicator = newVal.mainIndicator || 'MA'
					this.chartSubIndicator = newVal.subIndicator || 'VOL'
					this.loadScript()
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
				setTimeout(() => this.initChart(), 100)
				return
			}
			
			// 检查元素尺寸
			if (el.clientWidth === 0 || el.clientHeight === 0) {
				console.log('KLine DOM size is 0, retrying...')
				setTimeout(() => this.initChart(), 100)
				return
			}
			
			if (this.chart) {
				console.log('Chart already initialized')
				return
			}

			try {
				console.log('Initializing KLineChart...')
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
							priceMark: {
								last: {
									show: true,
									upColor: '#00c087',
									downColor: '#F6465D',
									line: { show: true, style: 'dash', size: 1 },
									text: { show: true, color: '#fff', size: 12, paddingLeft: 4, paddingTop: 4, paddingRight: 4, paddingBottom: 4, borderRadius: 2 }
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

				this.updateData()
				
				// 关键：强制延迟 resize 确保渲染
				setTimeout(() => {
					if (this.chart) {
						this.chart.resize()
						console.log('KLineChart resize completed')
					}
				}, 200)
				
				console.log('KLineChart initialized successfully')
			} catch (err) {
				console.error('KLine init fail:', err)
			}
		},
		updateData() {
			const dataList = []
			let lastClose = 89354.83
			const interval = this.chartInterval
			const now = Math.floor(Date.now() / interval) * interval
			
			// 根据周期调整波动率
			let volatilityFactor = 0.002
			if (interval >= 24 * 60 * 60 * 1000) {
				volatilityFactor = 0.02  // 日线波动更大
			} else if (interval >= 4 * 60 * 60 * 1000) {
				volatilityFactor = 0.01  // 4小时
			} else if (interval >= 60 * 60 * 1000) {
				volatilityFactor = 0.006 // 1小时
			} else if (interval >= 15 * 60 * 1000) {
				volatilityFactor = 0.004 // 15分钟
			} else if (interval >= 5 * 60 * 1000) {
				volatilityFactor = 0.003 // 5分钟
			}
			
			for (let i = 500; i > 0; i--) {
				const timestamp = now - i * interval
				const volatility = lastClose * volatilityFactor
				const open = lastClose
				const change = (Math.random() - 0.5) * volatility
				const close = open + change
				const high = Math.max(open, close) + Math.random() * (volatility * 0.5)
				const low = Math.min(open, close) - Math.random() * (volatility * 0.5)
				const volume = (Math.random() * 100 + 50) * (interval / 60000) // 成交量与周期成比例
				
				dataList.push({
					timestamp,
					open,
					high,
					low,
					close,
					volume
				})
				lastClose = close
			}
			this.chart.applyNewData(dataList)
			console.log('KLineChart data applied, period:', this.chartPeriod, 'count:', dataList.length)
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
	
	.price-left {
		.current-price {
			font-size: 56rpx;
			font-weight: bold;
			color: #333;
			display: block;
		}
		
		.price-sub {
			display: flex;
			align-items: center;
			gap: 12rpx;
			margin-top: 10rpx;
			
			.price-usd { font-size: 24rpx; color: #999; }
			.change { font-size: 24rpx; &.down { color: #F6465D; } &.up { color: #00c087; } }
		}
		
		.tags {
			display: flex;
			gap: 12rpx;
			margin-top: 12rpx;
			.tag {
				font-size: 20rpx;
				color: #B88A44;
				background: #FFF9F0;
				padding: 2rpx 8rpx;
				border-radius: 4rpx;
			}
		}
	}
	
	.price-right {
		display: flex;
		flex-direction: column;
		gap: 16rpx;
		
		.stat-row {
			display: flex;
			gap: 30rpx;
			
			.stat-item {
				display: flex;
				flex-direction: column;
				align-items: flex-end;
				
				.label { font-size: 20rpx; color: #999; }
				.val { font-size: 20rpx; color: #333; font-weight: 500; margin-top: 4rpx; }
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
	
	.kline-chart-content {
		height: 100%;
		width: 100%;
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

.footer-actions {
	padding: 20rpx 30rpx;
	padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
	display: flex;
	gap: 20rpx;
	border-top: 1rpx solid #f5f5f5;
	
	.action-btn {
		flex: 1;
		height: 84rpx;
		border-radius: 42rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 32rpx;
		font-weight: bold;
		color: #fff;
		&.buy { background: #00c087; }
		&.sell { background: #F6465D; }
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