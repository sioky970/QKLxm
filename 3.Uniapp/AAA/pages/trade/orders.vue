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
				<text class="title-text">订单</text>
			</view>
			<view class="nav-right"></view>
		</view>
		
		<!-- Tab 栏 -->
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
				<view class="filter-btn" :class="{ active: filters.orderType }" @click="showFilterPicker('orderType')">
					<text>{{ orderTypeText }}</text>
					<view class="dropdown-icon small"></view>
				</view>
				<view class="filter-btn" :class="{ active: filters.side }" @click="showFilterPicker('side')">
					<text>{{ sideText }}</text>
					<view class="dropdown-icon small"></view>
				</view>
				<view class="filter-btn" :class="{ active: filters.status !== null }" @click="showFilterPicker('status')" v-if="activeTab === 'history'">
					<text>{{ statusText }}</text>
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
			<view class="empty-state" v-if="!loading && orderList.length === 0">
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
			<view class="order-item" v-for="(item, index) in orderList" :key="item.id">
				<view class="order-header">
					<view class="header-left">
						<text class="order-symbol">{{ item.symbol }}</text>
						<text class="order-type-tag" :class="item.type">{{ item.type === 'limit' ? '限价' : '市价' }}</text>
					</view>
					<text class="order-status" :class="getStatusClass(item.status)">{{ getStatusText(item.status) }}</text>
				</view>
				
				<view class="order-info">
					<view class="info-row">
						<text class="label">方向</text>
						<text class="value" :class="item.side">{{ item.side === 'buy' ? '买入' : '卖出' }}</text>
					</view>
					<view class="info-row">
						<text class="label">价格</text>
						<text class="value">{{ formatPrice(item.price) }} USDT</text>
					</view>
					<view class="info-row">
						<text class="label">数量</text>
						<text class="value">{{ formatNumber(item.number) }}</text>
					</view>
					<view class="info-row">
						<text class="label">成交</text>
						<text class="value">{{ formatNumber(item.deal_number) }}</text>
					</view>
					<view class="info-row">
						<text class="label">金额</text>
						<text class="value">{{ formatPrice(item.order_value) }} USDT</text>
					</view>
				</view>
				
				<view class="order-footer">
					<text class="order-time">{{ formatTime(item.time || item.create_time) }}</text>
					<view class="order-actions" v-if="activeTab === 'pending' && item.status === 0">
						<view class="action-btn cancel" @click="handleCancelOrder(item)">撤销</view>
					</view>
				</view>
			</view>
			
			<!-- 加载更多 -->
			<view class="load-more" v-if="orderList.length > 0">
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
import { getSpotOrderList, cancelSpotOrder, getCurrencyList } from '@/utils/api.js'

const statusBarHeight = ref(0)
const activeTab = ref('pending') // 默认显示当前委托（status=0 未成交）
const loading = ref(false)
const isRefreshing = ref(false)
const hasMore = ref(true)
const page = ref(1)
const pageSize = 20

// 防重复请求控制
const isLoadingMore = ref(false) // 加载更多专用锁
const loadRequestId = ref(0) // 请求ID，防止旧请求覆盖新数据
const loadMoreTimer = ref(null) // 防抖定时器
const loadError = ref(false) // 加载失败标记

// 筛选条件
const filters = ref({
	pair: '',
	currencyId: null,
	orderType: '',
	side: '',
	status: null
})

// 订单列表
const orderList = ref([])

// 币种列表
const currencyList = ref([])

// 弹窗控制
const showModal = ref(false)
const modalType = ref('')
const modalTitle = ref('')
const modalOptions = ref([])

// 计算属性
const orderTypeText = computed(() => {
	if (!filters.value.orderType) return '订单类型'
	return filters.value.orderType === 'limit' ? '限价单' : '市价单'
})

const sideText = computed(() => {
	if (!filters.value.side) return '全部方向'
	return filters.value.side === 'buy' ? '买入' : '卖出'
})

const statusText = computed(() => {
	if (filters.value.status === null) return '全部状态'
	const statusMap = { 0: '未成交', 1: '部分成交', 2: '已成交', 3: '已撤销' }
	return statusMap[filters.value.status] || '全部状态'
})

const hasFilters = computed(() => {
	return filters.value.pair || filters.value.orderType || filters.value.side || filters.value.status !== null
})

// 根据 Tab 显示不同的空状态文案
const emptyText = computed(() => {
	switch (activeTab.value) {
		case 'pending':
			return '暂无当前委托'
		case 'filled':
			return '暂无成交记录'
		default:
			return '暂无订单记录'
	}
})

// 方法
const goBack = () => {
	uni.navigateBack()
}

const switchTab = (tab) => {
	if (activeTab.value === tab) return
	console.log('[Orders] 切换Tab:', { from: activeTab.value, to: tab })
	
	activeTab.value = tab
	
	// 如果切换到订单历史Tab，清空状态筛选（确保显示所有状态）
	if (tab === 'history') {
		if (filters.value.status !== null) {
			console.log('[Orders] 切换到订单历史Tab - 清空状态筛选')
			filters.value.status = null
		}
	}
	
	resetAndLoad()
}

const resetAndLoad = () => {
	// 取消正在进行的加载
	if (loadMoreTimer.value) {
		clearTimeout(loadMoreTimer.value)
		loadMoreTimer.value = null
	}
	
	page.value = 1
	hasMore.value = true
	loadError.value = false
	isLoadingMore.value = false
	orderList.value = []
	loadOrders(true) // true 表示首次加载
}

const loadOrders = async (isFirstLoad = false) => {
	// 严格的防重复请求检查
	if (loading.value) {
		console.log('[Orders] 请求正在进行中，跳过')
		return
	}
	
	// 非首次加载时检查加载更多锁
	if (!isFirstLoad && isLoadingMore.value) {
		console.log('[Orders] 加载更多正在进行中，跳过')
		return
	}
	
	// 没有更多数据时直接返回
	if (!isFirstLoad && !hasMore.value) {
		console.log('[Orders] 没有更多数据')
		return
	}
	
	// 生成请求ID
	const currentRequestId = ++loadRequestId.value
	
	loading.value = true
	if (!isFirstLoad) {
		isLoadingMore.value = true
	}
	loadError.value = false
	
	try {
		// 构建请求参数
		const params = {
			page: page.value,
			page_size: pageSize,
			_t: Date.now() // 防止缓存
		}
		
		// 根据 Tab 设置状态筛选
		if (activeTab.value === 'pending') {
			params.status = 0 // 未成交 - 当前委托Tab固定筛选条件
			console.log('[Orders] 当前委托 Tab - 设置 status=0')
		} else if (activeTab.value === 'filled') {
			params.status = 2 // 已成交 - 历史成交Tab固定筛选条件
			console.log('[Orders] 历史成交 Tab - 设置 status=2')
		} else if (activeTab.value === 'history' && filters.value.status !== null) {
			// 订单历史 Tab 且有状态筛选时，使用筛选状态
			params.status = filters.value.status
			console.log('[Orders] 订单历史 Tab - 使用筛选状态:', filters.value.status)
		} else if (activeTab.value === 'history') {
			// 订单历史 Tab 且无状态筛选时，不设置 status 参数（显示所有状态）
			console.log('[Orders] 订单历史 Tab - 不设置状态筛选（显示所有状态）')
		}
		
		// 币种筛选
		if (filters.value.currencyId) {
			params.currency_id = filters.value.currencyId
		}
		
		// 方向筛选
		if (filters.value.side) {
			// 修复：后端期望 buy=1, sell=2（不是0和1）
			params.side = filters.value.side  // 直接传递 'buy' 或 'sell' 字符串
		}
		
		console.log('[Orders] 发起请求:', { page: page.value, requestId: currentRequestId, params })
		
		const res = await getSpotOrderList(params)
		
		console.log('[Orders] API 响应:', res)
		
		// 检查请求ID，防止旧请求覆盖新数据
		if (currentRequestId !== loadRequestId.value) {
			console.log('[Orders] 请求已过期，丢弃结果:', currentRequestId)
			return
		}
		
		// 后端返回格式: { type: 'success', data: { list, total, page, size } }
		if (res && res.data) {
			const newOrders = res.data.list || []
			
			console.log('[Orders] 收到订单数据:', { 
				tab: activeTab.value, 
				count: newOrders.length, 
				statusFilter: params.status,
				orders: newOrders.map(o => ({ id: o.id, status: o.status, symbol: o.symbol }))
			})
			
			// 订单类型筛选（前端筛选）
			let filteredOrders = newOrders
			if (filters.value.orderType) {
				filteredOrders = newOrders.filter(o => o.type === filters.value.orderType)
			}
			
			if (page.value === 1) {
				orderList.value = filteredOrders
			} else {
				// 去重合并，防止重复数据
				const existingIds = new Set(orderList.value.map(o => o.id))
				const uniqueNewOrders = filteredOrders.filter(o => !existingIds.has(o.id))
				orderList.value = [...orderList.value, ...uniqueNewOrders]
			}
			
			// 判断是否还有更多
			hasMore.value = newOrders.length >= pageSize
			console.log('[Orders] 加载成功:', { 
				loaded: newOrders.length, 
				total: orderList.value.length,
				hasMore: hasMore.value 
			})
		} else {
			console.log('[Orders] API 响应无数据:', res)
			if (page.value === 1) {
				orderList.value = []
			}
			hasMore.value = false
		}
	} catch (err) {
		console.error('[Orders] 加载订单失败:', err)
		
		// 检查请求ID
		if (currentRequestId !== loadRequestId.value) {
			return
		}
		
		loadError.value = true
		
		// 加载失败时回退页码（非首页时）
		if (page.value > 1) {
			page.value--
		}
		
		if (page.value === 1) {
			orderList.value = []
		}
	} finally {
		// 检查请求ID，确保是当前请求
		if (currentRequestId === loadRequestId.value) {
			loading.value = false
			isLoadingMore.value = false
			isRefreshing.value = false
		}
	}
}

const onRefresh = () => {
	// 取消正在进行的加载更多
	if (loadMoreTimer.value) {
		clearTimeout(loadMoreTimer.value)
		loadMoreTimer.value = null
	}
	
	console.log('[Orders] 下拉刷新:', { activeTab: activeTab.value, statusFilter: filters.value.status })
	
	isRefreshing.value = true
	page.value = 1
	hasMore.value = true
	loadError.value = false
	isLoadingMore.value = false
	
	// 如果当前是订单历史Tab，清空状态筛选（确保显示所有状态）
	if (activeTab.value === 'history') {
		if (filters.value.status !== null) {
			console.log('[Orders] 刷新 - 订单历史Tab清空状态筛选')
			filters.value.status = null
		}
	}
	
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
				...currencyList.value.map(c => ({ 
					label: `${c.name}/USDT`, 
					value: c.name,
					currencyId: c.id
				}))
			]
			break
		case 'orderType':
			modalTitle.value = '订单类型'
			modalOptions.value = [
				{ label: '全部类型', value: '' },
				{ label: '限价单', value: 'limit' },
				{ label: '市价单', value: 'market' }
			]
			break
		case 'side':
			modalTitle.value = '交易方向'
			modalOptions.value = [
				{ label: '全部方向', value: '' },
				{ label: '买入', value: 'buy' },
				{ label: '卖出', value: 'sell' }
			]
			break
		case 'status':
			modalTitle.value = '订单状态'
			modalOptions.value = [
				{ label: '全部状态', value: null },
				{ label: '未成交', value: 0 },
				{ label: '部分成交', value: 1 },
				{ label: '已成交', value: 2 },
				{ label: '已撤销', value: 3 }
			]
			break
	}
	
	showModal.value = true
}

const closeModal = () => {
	showModal.value = false
}

const isOptionSelected = (value) => {
	switch (modalType.value) {
		case 'pair':
			return filters.value.pair === value
		case 'orderType':
			return filters.value.orderType === value
		case 'side':
			return filters.value.side === value
		case 'status':
			return filters.value.status === value
	}
	return false
}

const selectOption = (opt) => {
	switch (modalType.value) {
		case 'pair':
			filters.value.pair = opt.value
			filters.value.currencyId = opt.currencyId || null
			break
		case 'orderType':
			filters.value.orderType = opt.value
			break
		case 'side':
			filters.value.side = opt.value
			break
		case 'status':
			filters.value.status = opt.value
			break
	}
	closeModal()
	resetAndLoad()
}

const resetFilters = () => {
	filters.value = {
		pair: '',
		currencyId: null,
		orderType: '',
		side: '',
		status: null
	}
	resetAndLoad()
}

const handleCancelOrder = async (order) => {
	uni.showModal({
		title: '确认撤销',
		content: `确定要撤销订单 ${order.symbol} 吗？`,
		success: async (res) => {
			if (res.confirm) {
				try {
					await cancelSpotOrder({ order_id: order.id })
					uni.showToast({ title: '撤销成功', icon: 'success' })
					// 从列表中移除
					const idx = orderList.value.findIndex(o => o.id === order.id)
					if (idx !== -1) {
						orderList.value.splice(idx, 1)
					}
				} catch (err) {
					console.error('[Orders] 撤销订单失败:', err)
					uni.showToast({ title: err.message || '撤销失败', icon: 'none' })
				}
			}
		}
	})
}

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

const getStatusClass = (status) => {
	const classMap = {
		0: 'pending',
		1: 'partial',
		2: 'filled',
		3: 'canceled'
	}
	return classMap[status] || ''
}

const getStatusText = (status) => {
	const textMap = {
		0: '未成交',
		1: '部分成交',
		2: '已成交',
		3: '已撤销'
	}
	return textMap[status] || '未知'
}

// 加载币种列表
const loadCurrencyList = async () => {
	try {
		const res = await getCurrencyList()
		// 后端返回格式: { type: 'success', data: [...] }
		if (res && res.data) {
			currencyList.value = res.data.filter(c => c.name !== 'USDT')
		}
	} catch (err) {
		console.error('[Orders] 加载币种列表失败:', err)
	}
}

// 检查订单状态分布的辅助函数（用于调试）
const debugOrderStats = () => {
	const stats = {}
	orderList.value.forEach(order => {
		const statusText = getStatusText(order.status)
		stats[statusText] = (stats[statusText] || 0) + 1
	})
	console.log('[Orders] 当前订单状态分布:', stats, {
		tab: activeTab.value,
		filter: filters.value.status,
		total: orderList.value.length
	})
}

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	loadCurrencyList()
	loadOrders(true)
	
	// 添加一个延时日志，方便检查初始数据
	setTimeout(() => {
		debugOrderStats()
	}, 1000)
})

onUnmounted(() => {
	// 清理防抖定时器
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
