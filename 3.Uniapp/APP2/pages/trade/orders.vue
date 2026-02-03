<template>
	<view class="page">
		<!-- 状态栏占位 -->
		<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
		
		<!-- 顶部导航栏 -->
		<view class="nav-header">
			<view class="nav-back" @click="goBack">
				<view class="back-arrow"></view>
			</view>
			<view class="nav-title">
				<text class="title-text">全部订单</text>
			</view>
			<view class="nav-right"></view>
		</view>
		
		<!-- 交易类型Tab栏 -->
		<view class="type-tab-bar">
			<view 
				class="type-tab-item" 
				:class="{ active: tradeTypeTab === '' }"
				@click="switchTradeType('')"
			>全部</view>
			<view 
				class="type-tab-item" 
				:class="{ active: tradeTypeTab === 'spot' }"
				@click="switchTradeType('spot')"
			>现货</view>
			<view 
				class="type-tab-item" 
				:class="{ active: tradeTypeTab === 'contract' }"
				@click="switchTradeType('contract')"
			>永续合约</view>
			<view 
				class="type-tab-item" 
				:class="{ active: tradeTypeTab === 'micro' }"
				@click="switchTradeType('micro')"
			>交割合约</view>
		</view>
		
		<!-- 状态Tab栏 -->
		<view class="tab-bar">
			<view 
				class="tab-item" 
				:class="{ active: activeTab === 'pending' }"
				@click="switchTab('pending')"
			>当前委托</view>
			<view 
				class="tab-item" 
				:class="{ active: activeTab === 'history' }"
				@click="switchTab('history')"
			>订单历史</view>
			<view 
				class="tab-item" 
				:class="{ active: activeTab === 'filled' }"
				@click="switchTab('filled')"
			>历史成交</view>
		</view>
		
		<!-- 筛选栏 -->
		<view class="filter-bar">
			<scroll-view scroll-x class="filter-scroll">
				<view class="filter-btn" :class="{ active: filters.pair }" @click="showFilterPicker('pair')">
					<text>{{ filters.pair || '全部币对' }}</text>
					<view class="dropdown-icon small"></view>
				</view>
				<view class="filter-btn" :class="{ active: filters.side }" @click="showFilterPicker('side')">
					<text>{{ sideText }}</text>
					<view class="dropdown-icon small"></view>
				</view>
			</scroll-view>
			<view class="filter-icon" @click="resetFilters" v-if="hasFilters">
				<text class="reset-text">重置</text>
			</view>
		</view>
		
		<!-- 列表区域 -->
		<scroll-view 
			scroll-y 
			class="list-container"
			:refresher-enabled="true"
			:refresher-triggered="isRefreshing"
			@refresherrefresh="onRefresh"
			@scrolltolower="onLoadMore"
		>
			<!-- 空状态 -->
			<view class="empty-state" v-if="!loading && filteredOrderList.length === 0">
				<view class="empty-icon">
					<view class="doc-icon">
						<view class="doc-lines">
							<view class="line"></view>
							<view class="line short"></view>
							<view class="line"></view>
						</view>
						<view class="magnifier">
							<view class="mag-circle"></view>
							<view class="mag-handle"></view>
						</view>
					</view>
				</view>
				<text class="empty-text">{{ emptyText }}</text>
			</view>
			
			<!-- 订单列表 -->
			<view class="order-item" v-for="(item, index) in filteredOrderList" :key="getOrderKey(item)">
				<view class="order-header">
					<view class="header-left">
						<text class="order-symbol">{{ item.symbol }}</text>
						<text class="trade-type-tag" :class="item.trade_type">{{ item.trade_type_cn }}</text>
						<text class="order-type-tag" :class="item.order_type">{{ item.order_type === 'limit' ? '限价' : '市价' }}</text>
					</view>
					<text class="order-status" :class="getStatusClass(item)">{{ getStatusText(item) }}</text>
				</view>
				
				<view class="order-info">
					<view class="info-row">
						<text class="label">方向</text>
						<text class="value" :class="getSideClass(item.side)">{{ getSideText(item.side) }}</text>
					</view>
					<view class="info-row">
						<text class="label">价格</text>
						<text class="value">{{ formatPrice(item.price) }} USDT</text>
					</view>
					<view class="info-row">
						<text class="label">数量</text>
						<text class="value">{{ formatNumber(item.quantity) }}</text>
					</view>
					<!-- 合约特有：杠杆 -->
					<view class="info-row" v-if="item.trade_type === 'contract' && item.leverage">
						<text class="label">杠杆</text>
						<text class="value leverage">{{ item.leverage }}x</text>
					</view>
					<!-- 交割合约特有：持续时间 -->
					<view class="info-row" v-if="item.trade_type === 'micro' && item.duration">
						<text class="label">时长</text>
						<text class="value">{{ item.duration }}秒</text>
					</view>
					<!-- 盈亏（已完成订单显示） -->
					<view class="info-row" v-if="showPnL(item)">
						<text class="label">盈亏</text>
						<text class="value" :class="getPnLClass(item.pnl)">{{ formatPnL(item.pnl) }} USDT</text>
					</view>
					<!-- 金额（现货显示） -->
					<view class="info-row" v-if="item.trade_type === 'spot'">
						<text class="label">金额</text>
						<text class="value">{{ formatPrice(item.price * item.quantity) }} USDT</text>
					</view>
				</view>
				
				<view class="order-footer">
					<text class="order-time">{{ formatTime(item.create_time) }}</text>
					<view class="order-actions" v-if="canCancelOrder(item)">
						<view class="action-btn cancel" @click="handleCancelOrder(item)">撤销</view>
					</view>
				</view>
			</view>
			
			<!-- 加载更多 -->
			<view class="load-more" v-if="filteredOrderList.length > 0">
				<view v-if="isLoadingMore" class="loading-indicator">
					<view class="loading-spinner"></view>
					<text>加载中...</text>
				</view>
				<view v-else-if="loadError" class="load-error" @click="retryLoad">
					<text>加载失败，点击重试</text>
				</view>
				<text v-else-if="!hasMore" class="no-more">没有更多了</text>
			</view>
		</scroll-view>
		
		<!-- 筛选弹窗 -->
		<view class="filter-modal" v-if="showModal" @click="closeModal">
			<view class="modal-content" @click.stop>
				<view class="modal-header">
					<text class="modal-title">{{ modalTitle }}</text>
					<view class="modal-close" @click="closeModal">×</view>
				</view>
				<view class="modal-body">
					<view 
						class="modal-option" 
						v-for="(opt, idx) in modalOptions" 
						:key="idx"
						:class="{ active: isOptionSelected(opt.value) }"
						@click="selectOption(opt)"
					>
						<text>{{ opt.label }}</text>
						<view class="check-icon" v-if="isOptionSelected(opt.value)">✓</view>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { getUnifiedOrderList, cancelSpotOrder, contractCancelOrder, getCurrencyList } from '@/utils/api.js'

const statusBarHeight = ref(0)
const tradeTypeTab = ref('') // 交易类型：'', 'spot', 'contract', 'micro'
const activeTab = ref('history') // 状态Tab：'pending', 'history', 'filled' - 默认显示订单历史
const loading = ref(false)
const isRefreshing = ref(false)
const hasMore = ref(true)
const page = ref(1)
const pageSize = 20

// 防重复请求控制
const isLoadingMore = ref(false)
const loadRequestId = ref(0)
const loadMoreTimer = ref(null)
const loadError = ref(false)

// 筛选条件
const filters = ref({
	pair: '',
	currencyId: null,
	side: ''
})

// 订单列表
const allOrders = ref([]) // 存储所有订单原始数据
const orderList = ref([]) // 兼容旧代码

// 前端筛选后的订单列表
const filteredOrderList = computed(() => {
	let result = [...allOrders.value]
	console.log('[Orders] 筛选前订单数:', result.length, '交易类型:', tradeTypeTab.value, '状态Tab:', activeTab.value)
	
	// 1. 按交易类型筛选
	if (tradeTypeTab.value) {
		result = result.filter(order => order.trade_type === tradeTypeTab.value)
		console.log('[Orders] 交易类型筛选后:', result.length)
	}
	
	// 2. 按状态筛选
	if (activeTab.value === 'pending') {
		result = result.filter(order => order.status === 0)
		console.log('[Orders] 当前委托筛选后:', result.length)
	} else if (activeTab.value === 'filled') {
		// 历史成交: 现货status=2, 合约status=3(已平仓), 交割合约status=1(已结算)
		result = result.filter(order => {
			if (order.trade_type === 'spot') return order.status === 2
			if (order.trade_type === 'contract') return order.status === 3
			if (order.trade_type === 'micro') return order.status === 1
			return false
		})
		console.log('[Orders] 历史成交筛选后:', result.length)
	}
	// 'history' 显示所有状态，不筛选
	
	// 3. 按币对筛选
	if (filters.value.pair) {
		result = result.filter(order => {
			const symbol = order.symbol || ''
			return symbol.includes(filters.value.pair)
		})
	}
	
	// 4. 按方向筛选
	if (filters.value.side) {
		result = result.filter(order => order.side === filters.value.side)
	}
	
	console.log('[Orders] 最终筛选结果:', result.length)
	return result
})

// 币种列表
const currencyList = ref([])

// 弹窗控制
const showModal = ref(false)
const modalType = ref('')
const modalTitle = ref('')
const modalOptions = ref([])

// 计算属性
const sideText = computed(() => {
	if (!filters.value.side) return '全部方向'
	const sideMap = { 'buy': '买入', 'sell': '卖出', 'long': '做多', 'short': '做空' }
	return sideMap[filters.value.side] || '全部方向'
})

const hasFilters = computed(() => filters.value.pair || filters.value.side)

const emptyText = computed(() => {
	switch (activeTab.value) {
		case 'pending': return '暂无当前委托'
		case 'filled': return '暂无成交记录'
		default: return '暂无订单记录'
	}
})

// 方法
const goBack = () => uni.navigateBack()

const switchTradeType = (type) => {
	if (tradeTypeTab.value === type) return
	tradeTypeTab.value = type
	// 前端筛选，不需要重新加载数据
}

const switchTab = (tab) => {
	if (activeTab.value === tab) return
	activeTab.value = tab
	// 前端筛选，不需要重新加载数据
}

const resetAndLoad = () => {
	if (loadMoreTimer.value) { clearTimeout(loadMoreTimer.value); loadMoreTimer.value = null }
	page.value = 1
	hasMore.value = true
	loadError.value = false
	isLoadingMore.value = false
	allOrders.value = []
	orderList.value = []
	loadOrders(true)
}

const getOrderKey = (order) => `${order.trade_type}_${order.id}`

const loadOrders = async (isFirstLoad = false) => {
	if (loading.value) return
	if (!isFirstLoad && isLoadingMore.value) return
	if (!isFirstLoad && !hasMore.value) return
	
	const currentRequestId = ++loadRequestId.value
	loading.value = true
	if (!isFirstLoad) isLoadingMore.value = true
	loadError.value = false
	
	try {
		// 前端筛选模式：只传分页参数，不传筛选条件
		const params = {
			page: page.value,
			page_size: pageSize
		}
		
		console.log('[Orders] 发起请求(前端筛选模式):', params)
		const res = await getUnifiedOrderList(params)
		
		if (currentRequestId !== loadRequestId.value) return
		
		console.log('[Orders] API响应:', res)
		if (res && res.data) {
			const newOrders = res.data.list || []
			console.log('[Orders] 获取到订单数:', newOrders.length)
			if (page.value === 1) {
				allOrders.value = newOrders
				orderList.value = newOrders // 兼容
			} else {
				const existingKeys = new Set(allOrders.value.map(o => getOrderKey(o)))
				const uniqueNewOrders = newOrders.filter(o => !existingKeys.has(getOrderKey(o)))
				allOrders.value = [...allOrders.value, ...uniqueNewOrders]
				orderList.value = allOrders.value // 兼容
			}
			hasMore.value = newOrders.length >= pageSize
		} else {
			if (page.value === 1) {
				allOrders.value = []
				orderList.value = []
			}
			hasMore.value = false
		}
	} catch (err) {
		console.error('[Orders] 加载失败:', err)
		if (currentRequestId !== loadRequestId.value) return
		loadError.value = true
		if (page.value > 1) page.value--
		if (page.value === 1) {
			allOrders.value = []
			orderList.value = []
		}
	} finally {
		if (currentRequestId === loadRequestId.value) {
			loading.value = false
			isLoadingMore.value = false
			isRefreshing.value = false
		}
	}
}

const onRefresh = () => {
	if (loadMoreTimer.value) { clearTimeout(loadMoreTimer.value); loadMoreTimer.value = null }
	isRefreshing.value = true
	page.value = 1
	hasMore.value = true
	loadError.value = false
	isLoadingMore.value = false
	loadOrders(true)
}

// 带防抖的加载更多
const onLoadMore = () => {
	// 严格条件检查
	if (loading.value || isLoadingMore.value || !hasMore.value || loadError.value) {
		return
	}
	
	// 清除之前的防抖定时器
	if (loadMoreTimer.value) {
		clearTimeout(loadMoreTimer.value)
	}
	
	// 防抖：300ms 内不重复触发
	loadMoreTimer.value = setTimeout(() => {
		loadMoreTimer.value = null
		
		// 再次检查条件
		if (loading.value || isLoadingMore.value || !hasMore.value) {
			return
		}
		
		page.value++
		loadOrders(false)
	}, 300)
}

// 重试加载
const retryLoad = () => {
	loadError.value = false
	if (page.value === 1) {
		loadOrders(true)
	} else {
		loadOrders(false)
	}
}

const showFilterPicker = (type) => {
	modalType.value = type
	switch (type) {
		case 'pair':
			modalTitle.value = '选择币对'
			modalOptions.value = [
				{ label: '全部币对', value: '' },
				...currencyList.value.map(c => ({ label: `${c.name}/USDT`, value: c.name, currencyId: c.id }))
			]
			break
		case 'side':
			modalTitle.value = '交易方向'
			if (tradeTypeTab.value === 'spot') {
				modalOptions.value = [
					{ label: '全部方向', value: '' },
					{ label: '买入', value: 'buy' },
					{ label: '卖出', value: 'sell' }
				]
			} else if (tradeTypeTab.value === 'contract' || tradeTypeTab.value === 'micro') {
				modalOptions.value = [
					{ label: '全部方向', value: '' },
					{ label: '做多', value: 'long' },
					{ label: '做空', value: 'short' }
				]
			} else {
				modalOptions.value = [
					{ label: '全部方向', value: '' },
					{ label: '买入', value: 'buy' },
					{ label: '卖出', value: 'sell' },
					{ label: '做多', value: 'long' },
					{ label: '做空', value: 'short' }
				]
			}
			break
	}
	showModal.value = true
}

const closeModal = () => {
	showModal.value = false
}

const isOptionSelected = (value) => {
	switch (modalType.value) {
		case 'pair': return filters.value.pair === value
		case 'side': return filters.value.side === value
	}
	return false
}

const selectOption = (opt) => {
	switch (modalType.value) {
		case 'pair':
			filters.value.pair = opt.value
			filters.value.currencyId = opt.currencyId || null
			break
		case 'side':
			filters.value.side = opt.value
			break
	}
	closeModal()
	// 前端筛选，不需要重新加载数据
}

const resetFilters = () => {
	filters.value = { pair: '', currencyId: null, side: '' }
	// 前端筛选，不需要重新加载数据
}

const handleCancelOrder = async (order) => {
	uni.showModal({
		title: '确认撤销',
		content: `确定要撤销订单 ${order.symbol} 吗？`,
		success: async (res) => {
			if (res.confirm) {
				try {
					if (order.trade_type === 'spot') {
						await cancelSpotOrder({ order_id: order.id })
					} else if (order.trade_type === 'contract') {
						await contractCancelOrder(order.id)
					} else {
						uni.showToast({ title: '交割合约不支持撤销', icon: 'none' })
						return
					}
					uni.showToast({ title: '撤销成功', icon: 'success' })
					const idx = orderList.value.findIndex(o => getOrderKey(o) === getOrderKey(order))
					if (idx !== -1) orderList.value.splice(idx, 1)
				} catch (err) {
					uni.showToast({ title: err.message || '撤销失败', icon: 'none' })
				}
			}
		}
	})
}

const canCancelOrder = (order) => activeTab.value === 'pending' && order.status === 0 && order.trade_type !== 'micro'

// 格式化函数
const formatPrice = (price) => {
	if (!price && price !== 0) return '---'
	return parseFloat(price).toFixed(4)
}

const formatNumber = (num) => {
	if (!num && num !== 0) return '0'
	return parseFloat(num).toFixed(6)
}

const formatTime = (timestamp) => {
	if (!timestamp) return '---'
	const date = new Date(timestamp * 1000)
	const month = String(date.getMonth() + 1).padStart(2, '0')
	const day = String(date.getDate()).padStart(2, '0')
	const hour = String(date.getHours()).padStart(2, '0')
	const minute = String(date.getMinutes()).padStart(2, '0')
	return `${month}-${day} ${hour}:${minute}`
}

const formatPnL = (pnl) => {
	if (pnl === null || pnl === undefined) return '---'
	const num = parseFloat(pnl)
	return (num >= 0 ? '+' : '') + num.toFixed(2)
}

const getStatusClass = (order) => {
	const status = order.status
	const tradeType = order.trade_type
	if (tradeType === 'spot') {
		return { 0: 'pending', 1: 'partial', 2: 'filled', 3: 'canceled' }[status] || ''
	} else if (tradeType === 'contract') {
		return { 0: 'pending', 1: 'holding', 2: 'closing', 3: 'closed' }[status] || ''
	} else if (tradeType === 'micro') {
		return { 0: 'pending', 1: 'filled' }[status] || ''
	}
	return ''
}

const getStatusText = (order) => {
	const status = order.status
	const tradeType = order.trade_type
	if (tradeType === 'spot') {
		return { 0: '未成交', 1: '部分成交', 2: '已成交', 3: '已撤销' }[status] || '未知'
	} else if (tradeType === 'contract') {
		return { 0: '挂单中', 1: '持仓中', 2: '平仓中', 3: '已平仓' }[status] || '未知'
	} else if (tradeType === 'micro') {
		return { 0: '进行中', 1: '已结算' }[status] || '未知'
	}
	return '未知'
}

const getSideClass = (side) => (side === 'buy' || side === 'long') ? 'buy' : 'sell'

const getSideText = (side) => ({ 'buy': '买入', 'sell': '卖出', 'long': '做多', 'short': '做空' }[side] || side)

const getPnLClass = (pnl) => {
	if (pnl === null || pnl === undefined) return ''
	return parseFloat(pnl) >= 0 ? 'profit' : 'loss'
}

const showPnL = (order) => {
	if (order.trade_type === 'contract' && order.status === 3) return true
	if (order.trade_type === 'micro' && order.status === 1) return true
	return false
}

// 加载币种列表
const loadCurrencyList = async () => {
	try {
		const res = await getCurrencyList()
		if (res && res.data) {
			currencyList.value = res.data.filter(c => c.name !== 'USDT')
		}
	} catch (err) {
		console.error('[Orders] 加载币种列表失败:', err)
	}
}

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	loadCurrencyList()
	loadOrders(true)
})

onUnmounted(() => {
	if (loadMoreTimer.value) {
		clearTimeout(loadMoreTimer.value)
		loadMoreTimer.value = null
	}
})
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #FFFFFF;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.status-bar {
	width: 100%;
	flex-shrink: 0;
}

/* 交易类型Tab栏 */
.type-tab-bar {
	display: flex;
	padding: 0 30rpx;
	gap: 24rpx;
	flex-shrink: 0;
	border-bottom: 1rpx solid #f5f5f5;
	background: #FAFAFA;
	
	.type-tab-item {
		font-size: 26rpx;
		color: #666;
		padding: 20rpx 16rpx;
		position: relative;
		border-radius: 8rpx;
		
		&.active {
			color: #F7A100;
			font-weight: 500;
			background: #FFF8E1;
		}
	}
}

/* 状态Tab栏 */
/* 顶部导航栏 */
.nav-header {
	display: flex;
	align-items: center;
	padding: 0 30rpx;
	height: 88rpx;
	flex-shrink: 0;
	
	.nav-back {
		width: 60rpx;
		height: 60rpx;
		display: flex;
		align-items: center;
		justify-content: flex-start;
		
		.back-arrow {
			width: 20rpx;
			height: 20rpx;
			border-left: 4rpx solid #333;
			border-bottom: 4rpx solid #333;
			transform: rotate(45deg);
		}
	}
	
	.nav-title {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		
		.title-text {
			font-size: 34rpx;
			font-weight: 600;
			color: #333;
		}
	}
	
	.nav-right {
		width: 60rpx;
		height: 60rpx;
	}
}

/* Tab 栏 */
.tab-bar {
	display: flex;
	padding: 0 30rpx;
	gap: 48rpx;
	flex-shrink: 0;
	border-bottom: 1rpx solid #f5f5f5;
	
	.tab-item {
		font-size: 30rpx;
		color: #999;
		padding: 24rpx 0;
		position: relative;
		
		&.active {
			color: #333;
			font-weight: 600;
			
			&::after {
				content: '';
				position: absolute;
				bottom: 0;
				left: 50%;
				transform: translateX(-50%);
				width: 48rpx;
				height: 6rpx;
				background: #F7D100;
				border-radius: 3rpx;
			}
		}
	}
}

/* 筛选栏 */
.filter-bar {
	display: flex;
	align-items: center;
	padding: 20rpx 30rpx;
	flex-shrink: 0;
	
	.filter-scroll {
		flex: 1;
		white-space: nowrap;
	}
	
	.filter-btn {
		display: inline-flex;
		align-items: center;
		gap: 8rpx;
		padding: 16rpx 24rpx;
		background: #F5F5F5;
		border-radius: 8rpx;
		margin-right: 16rpx;
		
		text {
			font-size: 26rpx;
			color: #666;
		}
		
		&.active {
			background: #FFF8E1;
			
			text {
				color: #F7A100;
			}
			
			.dropdown-icon.small {
				border-top-color: #F7A100;
			}
		}
	}
	
	.filter-icon {
		flex-shrink: 0;
		padding: 8rpx 16rpx;
		
		.reset-text {
			font-size: 26rpx;
			color: #F7A100;
		}
	}
}

/* 列表区域 */
.list-container {
	flex: 1;
	overflow: hidden;
}

/* 空状态 */
.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding-top: 200rpx;
	
	.empty-icon {
		width: 200rpx;
		height: 200rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 32rpx;
		
		.doc-icon {
			position: relative;
			width: 120rpx;
			height: 140rpx;
			border: 4rpx solid #ddd;
			border-radius: 12rpx;
			background: #fff;
			
			.doc-lines {
				padding: 24rpx 16rpx;
				display: flex;
				flex-direction: column;
				gap: 12rpx;
				
				.line {
					height: 8rpx;
					background: #eee;
					border-radius: 4rpx;
					
					&.short {
						width: 60%;
					}
				}
			}
			
			.magnifier {
				position: absolute;
				bottom: -10rpx;
				left: -20rpx;
				
				.mag-circle {
					width: 50rpx;
					height: 50rpx;
					border: 4rpx solid #ddd;
					border-radius: 50%;
					background: #fff;
				}
				
				.mag-handle {
					position: absolute;
					bottom: -16rpx;
					left: 36rpx;
					width: 24rpx;
					height: 6rpx;
					background: #ddd;
					transform: rotate(45deg);
					border-radius: 3rpx;
				}
			}
		}
	}
	
	.empty-text {
		font-size: 28rpx;
		color: #999;
	}
}

/* 订单项 */
.order-item {
	padding: 24rpx 30rpx;
	border-bottom: 1rpx solid #f5f5f5;
	
	.order-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20rpx;
		
		.header-left {
			display: flex;
			align-items: center;
			gap: 16rpx;
		}
		
		.order-symbol {
			font-size: 32rpx;
			font-weight: 600;
			color: #333;
		}
		
		.trade-type-tag {
			font-size: 20rpx;
			padding: 4rpx 10rpx;
			border-radius: 4rpx;
			background: #E8F5E9;
			color: #4CAF50;
			
			&.spot {
				background: #E3F2FD;
				color: #1976D2;
			}
			
			&.contract {
				background: #FFF3E0;
				color: #FF9800;
			}
			
			&.micro {
				background: #F3E5F5;
				color: #9C27B0;
			}
		}
		
		.order-type-tag {
			font-size: 22rpx;
			padding: 4rpx 12rpx;
			border-radius: 4rpx;
			background: #F5F5F5;
			color: #666;
			
			&.limit {
				background: #E3F2FD;
				color: #1976D2;
			}
			
			&.market {
				background: #FFF3E0;
				color: #F57C00;
			}
		}
		
		.order-status {
			font-size: 26rpx;
			
			&.pending { color: #F7A100; }
			&.partial { color: #2196F3; }
			&.filled { color: #00B894; }
			&.canceled { color: #999; }
			&.holding { color: #FF9800; }
			&.closing { color: #9C27B0; }
			&.closed { color: #4CAF50; }
		}
	}
	
	.order-info {
		display: flex;
		flex-wrap: wrap;
		gap: 16rpx 0;
		
		.info-row {
			width: 50%;
			display: flex;
			gap: 8rpx;
			
			.label {
				font-size: 26rpx;
				color: #999;
				min-width: 60rpx;
			}
			
			.value {
				font-size: 26rpx;
				color: #333;
				
				&.buy { color: #00B894; }
				&.sell { color: #E74C3C; }
				&.leverage { color: #FF9800; font-weight: 500; }
				&.profit { color: #00B894; font-weight: 500; }
				&.loss { color: #E74C3C; font-weight: 500; }
			}
		}
	}
	
	.order-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 20rpx;
		
		.order-time {
			font-size: 24rpx;
			color: #bbb;
		}
		
		.order-actions {
			display: flex;
			gap: 16rpx;
			
			.action-btn {
				font-size: 26rpx;
				padding: 8rpx 24rpx;
				border-radius: 6rpx;
				
				&.cancel {
					background: #FFF5F5;
					color: #E74C3C;
				}
			}
		}
	}
}

/* 加载更多 */
.load-more {
	text-align: center;
	padding: 32rpx;
	
	text {
		font-size: 26rpx;
		color: #999;
	}
	
	.no-more {
		font-size: 26rpx;
		color: #ccc;
	}
	
	.loading-indicator {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12rpx;
		
		text {
			font-size: 26rpx;
			color: #666;
		}
	}
	
	.loading-spinner {
		width: 32rpx;
		height: 32rpx;
		border: 4rpx solid #eee;
		border-top-color: #F7A100;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}
	
	.load-error {
		padding: 16rpx 32rpx;
		
		text {
			font-size: 26rpx;
			color: #E74C3C;
		}
	}
}

@keyframes spin {
	0% { transform: rotate(0deg); }
	100% { transform: rotate(360deg); }
}

/* 下拉图标 */
.dropdown-icon {
	width: 0;
	height: 0;
	border-left: 10rpx solid transparent;
	border-right: 10rpx solid transparent;
	border-top: 12rpx solid #333;
	
	&.small {
		border-left-width: 8rpx;
		border-right-width: 8rpx;
		border-top-width: 10rpx;
		border-top-color: #999;
	}
}

/* 筛选弹窗 */
.filter-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 100;
	display: flex;
	align-items: flex-end;
	
	.modal-content {
		width: 100%;
		background: #fff;
		border-radius: 24rpx 24rpx 0 0;
		max-height: 70vh;
		overflow: hidden;
	}
	
	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 32rpx;
		border-bottom: 1rpx solid #f5f5f5;
		
		.modal-title {
			font-size: 32rpx;
			font-weight: 600;
			color: #333;
		}
		
		.modal-close {
			font-size: 48rpx;
			color: #999;
			line-height: 1;
		}
	}
	
	.modal-body {
		padding: 16rpx 0;
		max-height: 50vh;
		overflow-y: auto;
	}
	
	.modal-option {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 28rpx 32rpx;
		
		text {
			font-size: 30rpx;
			color: #333;
		}
		
		.check-icon {
			color: #F7A100;
			font-size: 32rpx;
			font-weight: bold;
		}
		
		&.active {
			background: #FFF8E1;
			
			text {
				color: #F7A100;
				font-weight: 500;
			}
		}
	}
}
</style>
