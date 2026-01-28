<template>
	<view class="page">
		<!-- 自定义导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="nav-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#1E2329" />
				</view>
				<text class="nav-title">充值记录</text>
				<view class="nav-right"></view>
			</view>
		</view>

		<!-- 订单列表 -->
		<scroll-view 
			scroll-y 
			class="order-list" 
			:style="{ paddingTop: navHeight + 'px' }"
			@scrolltolower="loadMore"
			enhanced
			:bounces="true"
		>
			<!-- 骨架屏加载 -->
			<view v-if="loading && orders.length === 0" class="skeleton-list">
				<view class="skeleton-card" v-for="i in 3" :key="i">
					<view class="skeleton-header">
						<view class="skeleton-line skeleton-title"></view>
						<view class="skeleton-badge"></view>
					</view>
					<view class="skeleton-body">
						<view class="skeleton-line skeleton-amount"></view>
						<view class="skeleton-line skeleton-time"></view>
					</view>
				</view>
			</view>
			
			<!-- 空状态 -->
			<view v-else-if="orders.length === 0" class="empty-state">
				<view class="empty-icon-wrapper">
					<image class="empty-icon" src="/static/icons/record.svg" mode="aspectFit" />
				</view>
				<text class="empty-title">暂无充值记录</text>
				<text class="empty-desc">您还没有任何充值订单</text>
				<view class="empty-action" @click="goToDeposit">
					<text>去充值</text>
				</view>
			</view>
			
			<!-- 订单列表 -->
			<view v-else class="orders-container">
				<view 
					class="order-card" 
					v-for="order in orders" 
					:key="order.id" 
					@click="showDetail(order)"
					:class="{ 'order-card-pressed': activeOrderId === order.id }"
					@touchstart="handleTouchStart(order.id)"
					@touchend="handleTouchEnd"
				>
					<!-- 左侧状态条 -->
					<view class="status-bar" :class="getStatusBarClass(order.status)"></view>
					
					<view class="card-content">
						<!-- 卡片头部 -->
						<view class="card-header">
							<view class="header-left">
								<view class="network-badge">
									<text>{{ order.network }}</text>
								</view>
								<text class="order-id">{{ formatOrderNo(order.order_no) }}</text>
							</view>
							<view class="status-tag" :class="getStatusClass(order.status)">
								<view class="status-dot"></view>
								<text>{{ getStatusText(order.status) }}</text>
							</view>
						</view>
						
						<!-- 卡片主体 -->
						<view class="card-body">
							<view class="amount-section">
								<text class="amount-label">充值金额</text>
								<view class="amount-row">
									<text class="amount-value" :class="getAmountClass(order.status)">
										{{ formatAmount(order.amount) }}
									</text>
									<text class="amount-unit">USDT</text>
								</view>
							</view>
							<view class="time-section">
								<text class="time-label">创建时间</text>
								<text class="time-value">{{ formatTime(order.create_time) }}</text>
							</view>
						</view>
						
						<!-- 卡片底部操作 -->
						<view class="card-footer" v-if="order.status === 0">
							<view class="footer-hint">
								<text>等待管理员审核</text>
							</view>
							<view 
								class="action-btn cancel-btn" 
								@click.stop="cancelOrder(order)"
								:class="{ 'btn-pressed': cancelingOrderId === order.id }"
								@touchstart.stop="cancelingOrderId = order.id"
								@touchend.stop="cancelingOrderId = null"
							>
								<text>取消订单</text>
							</view>
						</view>
					</view>
					
					<!-- 箭头指示 -->
					<view class="card-arrow">
						<text class="arrow-icon">›</text>
					</view>
				</view>
				
				<!-- 加载更多 -->
				<view class="load-more" v-if="hasMore">
					<view v-if="loading" class="loading-spinner">
						<view class="spinner"></view>
						<text>加载中...</text>
					</view>
					<text v-else class="load-hint">上拉加载更多</text>
				</view>
				
				<!-- 没有更多 -->
				<view class="no-more" v-else-if="orders.length > 0">
					<view class="divider-line"></view>
					<text>已加载全部</text>
					<view class="divider-line"></view>
				</view>
			</view>
		</scroll-view>

		<!-- 订单详情弹窗 -->
		<view 
			class="popup-overlay" 
			v-if="showDetailPopup" 
			@click="closeDetail"
			:class="{ 'overlay-visible': popupAnimating }"
		>
			<view 
				class="detail-sheet" 
				@click.stop
				:class="{ 'sheet-visible': popupAnimating }"
			>
				<!-- 弹窗头部 -->
				<view class="sheet-header">
					<view class="sheet-handle"></view>
					<view class="header-row">
						<text class="sheet-title">订单详情</text>
						<view class="close-btn" @click="closeDetail">
							<SvgIcon name="close" :size="18" color="#848E9C" />
						</view>
					</view>
				</view>
				
				<!-- 弹窗内容 -->
				<scroll-view scroll-y class="sheet-body" v-if="currentOrder">
					<!-- 状态卡片 -->
					<view class="status-card" :class="getStatusClass(currentOrder.status)">
						<view class="status-icon-wrapper">
							<text class="status-emoji">{{ getStatusEmoji(currentOrder.status) }}</text>
						</view>
						<text class="status-main">{{ getStatusText(currentOrder.status) }}</text>
						<text class="status-sub">{{ getStatusDesc(currentOrder.status) }}</text>
					</view>
					
					<!-- 金额展示 -->
					<view class="amount-display">
						<text class="display-amount" :class="getAmountClass(currentOrder.status)">
							{{ formatAmount(currentOrder.amount) }}
						</text>
						<text class="display-unit">USDT</text>
					</view>
					
					<!-- 详情信息 -->
					<view class="detail-section">
						<view class="detail-item">
							<text class="item-label">订单编号</text>
							<view class="item-value-row">
								<text class="item-value" selectable>{{ currentOrder.order_no }}</text>
								<view class="copy-btn" @click="copyOrderNo">
									<text>复制</text>
								</view>
							</view>
						</view>
						
						<view class="detail-item">
							<text class="item-label">充值网络</text>
							<view class="network-value">
								<text class="item-value">{{ currentOrder.network }}</text>
							</view>
						</view>
						
						<view class="detail-item">
							<text class="item-label">创建时间</text>
							<text class="item-value">{{ formatFullTime(currentOrder.create_time) }}</text>
						</view>
						
						<view class="detail-item" v-if="currentOrder.review_time">
							<text class="item-label">审核时间</text>
							<text class="item-value">{{ formatFullTime(currentOrder.review_time) }}</text>
						</view>
						
						<view class="detail-item" v-if="currentOrder.admin_remark">
							<text class="item-label">审核备注</text>
							<text class="item-value remark-text">{{ currentOrder.admin_remark }}</text>
						</view>
					</view>
					
					<!-- 截图预览 -->
					<view class="screenshot-card" v-if="currentOrder.screenshot">
						<text class="screenshot-title">转账截图</text>
						<view class="screenshot-wrapper" @click="previewImage(currentOrder.screenshot)">
							<image 
								class="screenshot-image" 
								:src="getScreenshotUrl(currentOrder.screenshot)" 
								mode="widthFix"
							/>
							<view class="screenshot-overlay">
								<text>点击查看大图</text>
							</view>
						</view>
					</view>
				</scroll-view>
				
				<!-- 底部安全区 -->
				<view class="sheet-footer"></view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { getDepositOrders, cancelDepositOrder } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'

// 状态栏高度
const statusBarHeight = ref(0)
const navHeight = ref(88)

// 数据状态
const loading = ref(false)
const orders = ref([])
const currentStatus = ref(-1)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const hasMore = ref(true)

// 详情弹窗
const showDetailPopup = ref(false)
const popupAnimating = ref(false)
const currentOrder = ref(null)

// 交互状态
const activeOrderId = ref(null)
const cancelingOrderId = ref(null)

// API基础地址
const BASE_URL = 'http://localhost:8080'

// 状态选项
const statusOptions = [
	{ label: '全部', value: -1 },
	{ label: '待审核', value: 0 },
	{ label: '已完成', value: 1 },
	{ label: '已拒绝', value: 2 },
	{ label: '已取消', value: 3 },
	{ label: '已过期', value: 4 }
]

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	navHeight.value = statusBarHeight.value + 44
	
	fetchOrders()
})

// 获取订单列表
const fetchOrders = async (isLoadMore = false) => {
	if (loading.value) return
	
	loading.value = true
	try {
		const res = await getDepositOrders({
			status: currentStatus.value,
			page: page.value,
			page_size: pageSize.value
		})
		
		const data = res.data
		if (isLoadMore) {
			orders.value = [...orders.value, ...(data.list || [])]
		} else {
			orders.value = data.list || []
		}
		total.value = data.total || 0
		hasMore.value = orders.value.length < total.value
	} catch (err) {
		console.error('获取充值记录失败:', err)
		uni.showToast({ title: err.message || '获取记录失败', icon: 'none' })
	} finally {
		loading.value = false
	}
}

// 切换状态筛选
const changeStatus = (status) => {
	if (currentStatus.value === status) return
	currentStatus.value = status
	page.value = 1
	orders.value = []
	fetchOrders()
}

// 加载更多
const loadMore = () => {
	if (!hasMore.value || loading.value) return
	page.value++
	fetchOrders(true)
}

// 取消订单
const cancelOrder = async (order) => {
	uni.showModal({
		title: '确认取消',
		content: '确定要取消这个充值订单吗？取消后无法恢复。',
		confirmColor: '#F6465D',
		success: async (res) => {
			if (res.confirm) {
				try {
					await cancelDepositOrder(order.id)
					uni.showToast({ title: '订单已取消', icon: 'success' })
					page.value = 1
					fetchOrders()
				} catch (err) {
					uni.showToast({ title: err.message || '取消失败', icon: 'none' })
				}
			}
		}
	})
}

// 显示详情
const showDetail = (order) => {
	currentOrder.value = order
	showDetailPopup.value = true
	nextTick(() => {
		popupAnimating.value = true
	})
}

// 关闭详情
const closeDetail = () => {
	popupAnimating.value = false
	setTimeout(() => {
		showDetailPopup.value = false
		currentOrder.value = null
	}, 300)
}

// 复制订单号
const copyOrderNo = () => {
	if (!currentOrder.value) return
	uni.setClipboardData({
		data: currentOrder.value.order_no,
		success: () => {
			uni.showToast({ title: '已复制', icon: 'success' })
		}
	})
}

// 获取状态文本
const getStatusText = (status) => {
	const map = {
		0: '审核中',
		1: '已完成',
		2: '已拒绝',
		3: '已取消',
		4: '已过期'
	}
	return map[status] || '未知'
}

// 获取状态描述
const getStatusDesc = (status) => {
	const map = {
		0: '您的充值申请正在等待审核',
		1: '充值成功，资金已到账',
		2: '充值申请未通过审核',
		3: '您已取消该充值订单',
		4: '订单超时，已自动过期'
	}
	return map[status] || ''
}

// 获取状态Emoji
const getStatusEmoji = (status) => {
	const map = {
		0: '⏳',
		1: '✅',
		2: '❌',
		3: '🚫',
		4: '⏰'
	}
	return map[status] || '❓'
}

// 获取状态样式类
const getStatusClass = (status) => {
	const map = {
		0: 'status-pending',
		1: 'status-success',
		2: 'status-rejected',
		3: 'status-canceled',
		4: 'status-expired'
	}
	return map[status] || ''
}

// 获取状态条样式类
const getStatusBarClass = (status) => {
	const map = {
		0: 'bar-pending',
		1: 'bar-success',
		2: 'bar-rejected',
		3: 'bar-canceled',
		4: 'bar-expired'
	}
	return map[status] || ''
}

// 获取金额样式类
const getAmountClass = (status) => {
	if (status === 1) return 'amount-success'
	if (status === 2) return 'amount-failed'
	return ''
}

// 格式化订单号
const formatOrderNo = (orderNo) => {
	if (!orderNo || orderNo.length <= 12) return orderNo
	return orderNo.substring(0, 6) + '...' + orderNo.substring(orderNo.length - 6)
}

// 格式化金额
const formatAmount = (amount) => {
	if (!amount) return '0.00'
	const num = parseFloat(amount)
	return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 8 })
}

// 格式化时间
const formatTime = (timestamp) => {
	if (!timestamp) return '-'
	const date = new Date(timestamp * 1000)
	const month = String(date.getMonth() + 1).padStart(2, '0')
	const day = String(date.getDate()).padStart(2, '0')
	const hour = String(date.getHours()).padStart(2, '0')
	const minute = String(date.getMinutes()).padStart(2, '0')
	return `${month}-${day} ${hour}:${minute}`
}

// 格式化完整时间
const formatFullTime = (timestamp) => {
	if (!timestamp) return '-'
	const date = new Date(timestamp * 1000)
	const year = date.getFullYear()
	const month = String(date.getMonth() + 1).padStart(2, '0')
	const day = String(date.getDate()).padStart(2, '0')
	const hour = String(date.getHours()).padStart(2, '0')
	const minute = String(date.getMinutes()).padStart(2, '0')
	const second = String(date.getSeconds()).padStart(2, '0')
	return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

// 获取截图URL
const getScreenshotUrl = (path) => {
	if (!path) return ''
	if (path.startsWith('http')) return path
	return BASE_URL + path
}

// 预览图片
const previewImage = (path) => {
	uni.previewImage({
		urls: [getScreenshotUrl(path)]
	})
}

// 返回上一页
const goBack = () => {
	uni.navigateBack()
}

// 去充值
const goToDeposit = () => {
	uni.navigateTo({ url: '/pages/wallet/deposit' })
}

// 订单项触摸开始
const handleTouchStart = (orderId) => {
	activeOrderId.value = orderId
}

// 订单项触摸结束
const handleTouchEnd = () => {
	setTimeout(() => {
		activeOrderId.value = null
	}, 150)
}
</script>

<style scoped lang="scss">
// 主题色彩变量（参考币安配色）
$primary-color: #F0B90B;      // 币安黄
$success-color: #0ECB81;      // 成功绿
$danger-color: #F6465D;       // 危险红
$warning-color: #F0B90B;      // 警告黄
$text-primary: #1E2329;       // 主文字
$text-secondary: #474D57;     // 次要文字
$text-tertiary: #848E9C;      // 辅助文字
$bg-primary: #FAFAFA;         // 主背景
$bg-card: #FFFFFF;            // 卡片背景
$border-color: #EAECEF;       // 边框色

.page {
	width: 100%;
	min-height: 100vh;
	background: $bg-primary;
}

/* 导航栏 */
.custom-navbar {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: $bg-card;
	z-index: 100;
	box-shadow: 0 1rpx 0 $border-color;
	
	.navbar-content {
		height: 88rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 24rpx;
	}
	
	.nav-left {
		width: 72rpx;
		height: 72rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		transition: background-color 0.2s;
		margin-left: -12rpx;
		
		&:active {
			background-color: rgba(0, 0, 0, 0.04);
		}
	}
	
	.nav-title {
		font-size: 36rpx;
		font-weight: 600;
		color: $text-primary;
		letter-spacing: 1rpx;
	}
	
	.nav-right {
		width: 72rpx;
	}
}

/* 订单列表 */
.order-list {
	min-height: 100vh;
	padding: 24rpx;
	box-sizing: border-box;
}

/* 骨架屏 */
.skeleton-list {
	.skeleton-card {
		background: $bg-card;
		border-radius: 16rpx;
		padding: 32rpx;
		margin-bottom: 20rpx;
		
		.skeleton-header {
			display: flex;
			justify-content: space-between;
			margin-bottom: 24rpx;
		}
		
		.skeleton-body {
			display: flex;
			justify-content: space-between;
		}
		
		.skeleton-line {
			background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
			background-size: 200% 100%;
			animation: shimmer 1.5s infinite;
			border-radius: 8rpx;
		}
		
		.skeleton-title {
			width: 200rpx;
			height: 32rpx;
		}
		
		.skeleton-badge {
			width: 100rpx;
			height: 40rpx;
			background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
			background-size: 200% 100%;
			animation: shimmer 1.5s infinite;
			border-radius: 20rpx;
		}
		
		.skeleton-amount {
			width: 180rpx;
			height: 48rpx;
		}
		
		.skeleton-time {
			width: 140rpx;
			height: 28rpx;
		}
	}
}

@keyframes shimmer {
	0% { background-position: 200% 0; }
	100% { background-position: -200% 0; }
}

/* 空状态 */
.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 120rpx 60rpx;
	
	.empty-icon-wrapper {
		width: 160rpx;
		height: 160rpx;
		background: linear-gradient(135deg, #f5f5f5 0%, #e8e8e8 100%);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 32rpx;
		
		.empty-icon {
			width: 80rpx;
			height: 80rpx;
			opacity: 0.5;
		}
	}
	
	.empty-title {
		font-size: 32rpx;
		font-weight: 600;
		color: $text-primary;
		margin-bottom: 12rpx;
	}
	
	.empty-desc {
		font-size: 26rpx;
		color: $text-tertiary;
		margin-bottom: 40rpx;
	}
	
	.empty-action {
		padding: 20rpx 60rpx;
		background: $primary-color;
		border-radius: 8rpx;
		
		text {
			font-size: 28rpx;
			font-weight: 600;
			color: $text-primary;
		}
		
		&:active {
			opacity: 0.85;
		}
	}
}

/* 订单卡片 */
.orders-container {
	padding-bottom: 40rpx;
}

.order-card {
	background: $bg-card;
	border-radius: 16rpx;
	margin-bottom: 20rpx;
	display: flex;
	overflow: hidden;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
	transition: all 0.15s ease;
	
	&.order-card-pressed {
		transform: scale(0.98);
		box-shadow: 0 1rpx 4rpx rgba(0, 0, 0, 0.08);
	}
	
	.status-bar {
		width: 8rpx;
		flex-shrink: 0;
		
		&.bar-pending { background: $warning-color; }
		&.bar-success { background: $success-color; }
		&.bar-rejected { background: $danger-color; }
		&.bar-canceled { background: #B7BDC6; }
		&.bar-expired { background: #B7BDC6; }
	}
	
	.card-content {
		flex: 1;
		padding: 28rpx 24rpx;
		min-width: 0;
	}
	
	.card-arrow {
		width: 60rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		
		.arrow-icon {
			font-size: 40rpx;
			color: #B7BDC6;
			font-weight: 300;
		}
	}
}

.card-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 20rpx;
	
	.header-left {
		display: flex;
		align-items: center;
		gap: 16rpx;
		min-width: 0;
		flex: 1;
	}
	
	.network-badge {
		padding: 6rpx 16rpx;
		background: rgba(240, 185, 11, 0.1);
		border-radius: 6rpx;
		flex-shrink: 0;
		
		text {
			font-size: 22rpx;
			font-weight: 600;
			color: $primary-color;
		}
	}
	
	.order-id {
		font-size: 24rpx;
		color: $text-tertiary;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.status-tag {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 8rpx 16rpx;
	border-radius: 6rpx;
	flex-shrink: 0;
	
	.status-dot {
		width: 12rpx;
		height: 12rpx;
		border-radius: 50%;
	}
	
	text {
		font-size: 22rpx;
		font-weight: 500;
	}
	
	&.status-pending {
		background: rgba(240, 185, 11, 0.1);
		.status-dot { background: $warning-color; }
		text { color: #C99400; }
	}
	
	&.status-success {
		background: rgba(14, 203, 129, 0.1);
		.status-dot { background: $success-color; }
		text { color: #0ECB81; }
	}
	
	&.status-rejected {
		background: rgba(246, 70, 93, 0.1);
		.status-dot { background: $danger-color; }
		text { color: #F6465D; }
	}
	
	&.status-canceled,
	&.status-expired {
		background: rgba(132, 142, 156, 0.1);
		.status-dot { background: #848E9C; }
		text { color: #848E9C; }
	}
}

.card-body {
	display: flex;
	justify-content: space-between;
	align-items: flex-end;
	
	.amount-section {
		.amount-label {
			font-size: 22rpx;
			color: $text-tertiary;
			display: block;
			margin-bottom: 8rpx;
		}
		
		.amount-row {
			display: flex;
			align-items: baseline;
			gap: 8rpx;
		}
		
		.amount-value {
			font-size: 44rpx;
			font-weight: 700;
			color: $text-primary;
			font-family: 'DIN Alternate', -apple-system, sans-serif;
			
			&.amount-success { color: $success-color; }
			&.amount-failed { color: $danger-color; }
		}
		
		.amount-unit {
			font-size: 24rpx;
			color: $text-secondary;
			font-weight: 500;
		}
	}
	
	.time-section {
		text-align: right;
		
		.time-label {
			font-size: 22rpx;
			color: $text-tertiary;
			display: block;
			margin-bottom: 8rpx;
		}
		
		.time-value {
			font-size: 24rpx;
			color: $text-tertiary;
		}
	}
}

.card-footer {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-top: 20rpx;
	padding-top: 20rpx;
	border-top: 1rpx solid $border-color;
	
	.footer-hint {
		text {
			font-size: 24rpx;
			color: $text-tertiary;
		}
	}
	
	.action-btn {
		padding: 12rpx 28rpx;
		border-radius: 6rpx;
		transition: all 0.15s ease;
		
		text {
			font-size: 24rpx;
			font-weight: 500;
		}
		
		&.cancel-btn {
			background: rgba(246, 70, 93, 0.1);
			
			text { color: $danger-color; }
			
			&.btn-pressed,
			&:active {
				background: rgba(246, 70, 93, 0.2);
				transform: scale(0.95);
			}
		}
	}
}

/* 加载更多 */
.load-more {
	padding: 40rpx 0;
	display: flex;
	justify-content: center;
	
	.loading-spinner {
		display: flex;
		align-items: center;
		gap: 16rpx;
		
		.spinner {
			width: 32rpx;
			height: 32rpx;
			border: 3rpx solid $border-color;
			border-top-color: $primary-color;
			border-radius: 50%;
			animation: spin 0.8s linear infinite;
		}
		
		text {
			font-size: 26rpx;
			color: $text-tertiary;
		}
	}
	
	.load-hint {
		font-size: 26rpx;
		color: $text-tertiary;
	}
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.no-more {
	padding: 40rpx 0;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 20rpx;
	
	.divider-line {
		width: 80rpx;
		height: 1rpx;
		background: $border-color;
	}
	
	text {
		font-size: 24rpx;
		color: $text-tertiary;
	}
}

/* 详情弹窗 */
.popup-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0);
	z-index: 200;
	display: flex;
	align-items: flex-end;
	transition: background 0.3s ease;
	
	&.overlay-visible {
		background: rgba(0, 0, 0, 0.5);
	}
}

.detail-sheet {
	width: 100%;
	background: $bg-card;
	border-radius: 32rpx 32rpx 0 0;
	max-height: 85vh;
	display: flex;
	flex-direction: column;
	transform: translateY(100%);
	transition: transform 0.3s cubic-bezier(0.32, 0.72, 0, 1);
	
	&.sheet-visible {
		transform: translateY(0);
	}
}

.sheet-header {
	padding: 16rpx 32rpx 24rpx;
	
	.sheet-handle {
		width: 64rpx;
		height: 8rpx;
		background: #E0E0E0;
		border-radius: 4rpx;
		margin: 0 auto 20rpx;
	}
	
	.header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	
	.sheet-title {
		font-size: 36rpx;
		font-weight: 600;
		color: $text-primary;
	}
	
	.close-btn {
		width: 64rpx;
		height: 64rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #F5F5F5;
		border-radius: 50%;
		transition: background 0.2s;
		
		&:active {
			background: #E8E8E8;
		}
	}
}

.sheet-body {
	flex: 1;
	overflow-y: auto;
	padding: 0 32rpx;
}

.sheet-footer {
	height: calc(32rpx + env(safe-area-inset-bottom));
}

/* 状态卡片 */
.status-card {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 40rpx;
	border-radius: 16rpx;
	margin-bottom: 32rpx;
	
	.status-icon-wrapper {
		width: 80rpx;
		height: 80rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 16rpx;
		
		.status-emoji {
			font-size: 56rpx;
		}
	}
	
	.status-main {
		font-size: 32rpx;
		font-weight: 600;
		margin-bottom: 8rpx;
	}
	
	.status-sub {
		font-size: 24rpx;
	}
	
	&.status-pending {
		background: rgba(240, 185, 11, 0.08);
		.status-main { color: #C99400; }
		.status-sub { color: $text-tertiary; }
	}
	
	&.status-success {
		background: rgba(14, 203, 129, 0.08);
		.status-main { color: $success-color; }
		.status-sub { color: $text-tertiary; }
	}
	
	&.status-rejected {
		background: rgba(246, 70, 93, 0.08);
		.status-main { color: $danger-color; }
		.status-sub { color: $text-tertiary; }
	}
	
	&.status-canceled,
	&.status-expired {
		background: rgba(132, 142, 156, 0.08);
		.status-main { color: $text-secondary; }
		.status-sub { color: $text-tertiary; }
	}
}

/* 金额展示 */
.amount-display {
	display: flex;
	align-items: baseline;
	justify-content: center;
	gap: 12rpx;
	padding: 20rpx 0 40rpx;
	
	.display-amount {
		font-size: 72rpx;
		font-weight: 700;
		color: $text-primary;
		font-family: 'DIN Alternate', -apple-system, sans-serif;
		
		&.amount-success { color: $success-color; }
		&.amount-failed { color: $danger-color; }
	}
	
	.display-unit {
		font-size: 32rpx;
		color: $text-secondary;
		font-weight: 500;
	}
}

/* 详情信息 */
.detail-section {
	background: #FAFAFA;
	border-radius: 16rpx;
	padding: 8rpx 24rpx;
	margin-bottom: 32rpx;
	
	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24rpx 0;
		border-bottom: 1rpx solid $border-color;
		
		&:last-child {
			border-bottom: none;
		}
		
		.item-label {
			font-size: 28rpx;
			color: $text-tertiary;
			flex-shrink: 0;
		}
		
		.item-value-row {
			display: flex;
			align-items: center;
			gap: 16rpx;
		}
		
		.item-value {
			font-size: 28rpx;
			color: $text-primary;
			text-align: right;
			word-break: break-all;
			
			&.remark-text {
				color: $text-secondary;
				max-width: 400rpx;
			}
		}
		
		.copy-btn {
			padding: 8rpx 20rpx;
			background: rgba(240, 185, 11, 0.1);
			border-radius: 6rpx;
			
			text {
				font-size: 24rpx;
				color: $primary-color;
				font-weight: 500;
			}
			
			&:active {
				background: rgba(240, 185, 11, 0.2);
			}
		}
		
		.network-value {
			padding: 8rpx 16rpx;
			background: rgba(240, 185, 11, 0.1);
			border-radius: 6rpx;
			
			.item-value {
				color: $primary-color;
				font-weight: 500;
			}
		}
	}
}

/* 截图卡片 */
.screenshot-card {
	margin-bottom: 32rpx;
	
	.screenshot-title {
		font-size: 28rpx;
		color: $text-tertiary;
		display: block;
		margin-bottom: 16rpx;
	}
	
	.screenshot-wrapper {
		position: relative;
		border-radius: 16rpx;
		overflow: hidden;
		
		.screenshot-image {
			width: 100%;
			display: block;
		}
		
		.screenshot-overlay {
			position: absolute;
			bottom: 0;
			left: 0;
			right: 0;
			padding: 20rpx;
			background: linear-gradient(transparent, rgba(0, 0, 0, 0.5));
			
			text {
				font-size: 24rpx;
				color: #fff;
			}
		}
	}
}
</style>
