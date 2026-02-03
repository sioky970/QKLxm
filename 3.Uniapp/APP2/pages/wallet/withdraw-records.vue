<template>
	<view class="page">
		<!-- 自定义导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="nav-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<text class="nav-title">提现记录</text>
				<view class="nav-right"></view>
			</view>
		</view>

		<!-- 主内容区域 -->
		<scroll-view 
			scroll-y 
			class="content-scroll" 
			:style="{ paddingTop: navHeight + 'px' }"
			@scrolltolower="loadMore"
		>
			<!-- 加载状态 -->
			<view class="loading-box" v-if="loading && list.length === 0">
				<text class="loading-text">加载中...</text>
			</view>

			<!-- 空状态 -->
			<view class="empty-box" v-if="!loading && list.length === 0">
				<text class="empty-icon">📋</text>
				<text class="empty-text">暂无提现记录</text>
				<button class="goto-withdraw-btn" @click="goToWithdraw">立即提现</button>
			</view>

			<!-- 记录列表 -->
			<view class="record-list" v-if="list.length > 0">
				<view 
					class="record-item" 
					v-for="item in list" 
					:key="item.id"
					@click="showDetail(item)"
				>
					<view class="record-header">
						<view class="header-left">
							<text class="amount">{{ formatNumber(item.number) }} USDT</text>
							<text class="status-badge" :class="'status-' + item.status">
								{{ getStatusText(item.status) }}
							</text>
						</view>
						<text class="create-time">{{ formatTime(item.create_time) }}</text>
					</view>
					
					<view class="record-body">
						<!-- 银行转账模式 -->
						<template v-if="item.withdraw_type === 1">
							<view class="info-row">
								<text class="info-label">收款银行</text>
								<text class="info-value">{{ item.bank_name || '-' }}</text>
							</view>
							<view class="info-row">
								<text class="info-label">收款账户</text>
								<text class="info-value">{{ formatBankCard(item.bank_account) }}</text>
							</view>
						</template>
						<!-- 区块链提币模式 -->
						<template v-else>
							<view class="info-row">
								<text class="info-label">提币网络</text>
								<text class="info-value network-tag">{{ item.network_type || '-' }}</text>
							</view>
							<view class="info-row">
								<text class="info-label">钱包地址</text>
								<text class="info-value address-text">{{ formatChainAddress(item.chain_address || item.address) }}</text>
							</view>
						</template>
						<view class="info-row" v-if="item.status === 3 && item.notes">
							<text class="info-label">拒绝原因</text>
							<text class="info-value reject-reason">{{ item.notes }}</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 加载更多 -->
			<view class="loadmore-box" v-if="list.length > 0">
				<text class="loadmore-text" v-if="hasMore && !loadingMore">下拉加载更多</text>
				<text class="loadmore-text" v-if="loadingMore">加载中...</text>
				<text class="loadmore-text" v-if="!hasMore">没有更多了</text>
			</view>

			<!-- 底部占位 -->
			<view class="bottom-placeholder"></view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { request } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'

// 状态栏高度
const statusBarHeight = ref(0)
const navHeight = ref(88)

// 数据状态
const loading = ref(false)
const loadingMore = ref(false)
const list = ref([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const hasMore = ref(true)

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	navHeight.value = statusBarHeight.value + 44
	
	fetchRecords()
})

// 获取提现记录
const fetchRecords = async (isLoadMore = false) => {
	if (isLoadMore) {
		loadingMore.value = true
	} else {
		loading.value = true
		page.value = 1
		list.value = []
	}
	
	try {
		const res = await request({
			url: '/wallet/withdrawals',
			method: 'GET',
			data: {
				page: page.value,
				page_size: pageSize.value
			}
		})
		
		const newList = res.data?.list || []
		total.value = res.data?.total || 0
		
		if (isLoadMore) {
			list.value = [...list.value, ...newList]
		} else {
			list.value = newList
		}
		
		// 判断是否还有更多
		hasMore.value = list.value.length < total.value
		
	} catch (err) {
		console.error('获取提现记录失败:', err)
		uni.showToast({ title: err.message || '获取记录失败', icon: 'none' })
	} finally {
		loading.value = false
		loadingMore.value = false
	}
}

// 加载更多
const loadMore = () => {
	if (!hasMore.value || loadingMore.value) return
	page.value++
	fetchRecords(true)
}

// 格式化数字
const formatNumber = (num) => {
	if (!num && num !== 0) return '0.00'
	return parseFloat(num).toFixed(2)
}

// 格式化时间
const formatTime = (timestamp) => {
	if (!timestamp) return '-'
	const date = new Date(timestamp * 1000)
	const year = date.getFullYear()
	const month = String(date.getMonth() + 1).padStart(2, '0')
	const day = String(date.getDate()).padStart(2, '0')
	const hour = String(date.getHours()).padStart(2, '0')
	const minute = String(date.getMinutes()).padStart(2, '0')
	return `${year}-${month}-${day} ${hour}:${minute}`
}

// 格式化银行卡号
const formatBankCard = (cardNumber) => {
	if (!cardNumber) return '-'
	const str = String(cardNumber)
	if (str.length <= 8) return str
	return str.slice(0, 4) + ' **** ' + str.slice(-4)
}

// 格式化区块链地址
const formatChainAddress = (address) => {
	if (!address) return '-'
	const str = String(address)
	if (str.length <= 16) return str
	return str.slice(0, 8) + '...' + str.slice(-6)
}

// 获取状态文本
const getStatusText = (status) => {
	const statusMap = {
		1: '待审核',
		2: '已通过',
		3: '已拒绝'
	}
	return statusMap[status] || '未知'
}

// 显示详情
const showDetail = (item) => {
	const statusText = getStatusText(item.status)
	let content = `提现金额: ${formatNumber(item.number)} USDT\n`
	content += `申请时间: ${formatTime(item.create_time)}\n`
	content += `状态: ${statusText}\n`
	
	// 根据提现类型显示不同信息
	if (item.withdraw_type === 1) {
		content += `收款银行: ${item.bank_name || '-'}\n`
		content += `收款账户: ${formatBankCard(item.bank_account)}\n`
	} else {
		content += `提币网络: ${item.network_type || '-'}\n`
		content += `钱包地址: ${item.chain_address || item.address || '-'}\n`
	}
	
	if (item.status === 2 && item.update_time) {
		content += `审核时间: ${formatTime(item.update_time)}\n`
	}
	
	if (item.status === 3 && item.notes) {
		content += `拒绝原因: ${item.notes}`
	}
	
	uni.showModal({
		title: '提现详情',
		content: content,
		showCancel: false,
		confirmText: '确定'
	})
}

// 跳转到提现
const goToWithdraw = () => {
	uni.redirectTo({ url: '/pages/wallet/withdraw' })
}

// 返回上一页
const goBack = () => {
	uni.navigateBack()
}
</script>

<style scoped>
.page {
	min-height: 100vh;
	background: #F5F6FA;
	position: relative;
}

/* 导航栏 */
.custom-navbar {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: #fff;
	z-index: 999;
	border-bottom: 1px solid #E5E7EB;
}

.navbar-content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 44px;
	padding: 0 16px;
}

.nav-left, .nav-right {
	width: 60px;
}

.nav-title {
	font-size: 17px;
	font-weight: 600;
	color: #1F2937;
}

/* 主内容区域 */
.content-scroll {
	height: 100vh;
	padding: 12px 16px 40px;
	box-sizing: border-box;
}

/* 加载状态 */
.loading-box {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 60px 0;
	width: 100%;
	max-width: 600px;
	margin: 0 auto;
}

.loading-text {
	font-size: 14px;
	color: #9CA3AF;
}

/* 空状态 */
.empty-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80px 0;
	width: 100%;
	max-width: 600px;
	margin: 0 auto;
}

.empty-icon {
	font-size: 64px;
	margin-bottom: 16px;
}

.empty-text {
	font-size: 15px;
	color: #6B7280;
	margin-bottom: 24px;
}

.goto-withdraw-btn {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: #fff;
	border: none;
	border-radius: 20px;
	padding: 10px 32px;
	font-size: 15px;
	height: auto;
	line-height: 1.5;
}

/* 记录列表 */
.record-list {
	display: flex;
	flex-direction: column;
	gap: 12px;
	width: 100%;
	max-width: 600px;
	margin: 0 auto;
}

.record-item {
	background: #fff;
	border-radius: 12px;
	padding: 16px;
	width: 100%;
	box-sizing: border-box;
	box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.record-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 12px;
	padding-bottom: 12px;
	border-bottom: 1px solid #F3F4F6;
}

.header-left {
	display: flex;
	align-items: center;
	gap: 8px;
}

.amount {
	font-size: 18px;
	font-weight: 700;
	color: #1F2937;
}

.status-badge {
	font-size: 12px;
	padding: 2px 8px;
	border-radius: 4px;
	font-weight: 500;
}

.status-badge.status-1 {
	background: #FEF3C7;
	color: #92400E;
}

.status-badge.status-2 {
	background: #D1FAE5;
	color: #065F46;
}

.status-badge.status-3 {
	background: #FEE2E2;
	color: #991B1B;
}

.create-time {
	font-size: 12px;
	color: #9CA3AF;
}

.record-body {
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.info-row {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
}

.info-label {
	font-size: 13px;
	color: #6B7280;
	flex-shrink: 0;
}

.info-value {
	font-size: 13px;
	color: #1F2937;
	text-align: right;
	flex: 1;
	margin-left: 12px;
}

.reject-reason {
	color: #DC2626;
}

.network-tag {
	background: #EEF2FF;
	color: #4F46E5;
	padding: 2px 8px;
	border-radius: 4px;
	font-weight: 500;
}

.address-text {
	font-family: monospace;
	font-size: 12px;
	word-break: break-all;
}

/* 加载更多 */
.loadmore-box {
	display: flex;
	justify-content: center;
	padding: 20px 0;
	width: 100%;
	max-width: 600px;
	margin: 0 auto;
}

.loadmore-text {
	font-size: 13px;
	color: #9CA3AF;
}

/* 底部占位 */
.bottom-placeholder {
	height: 20px;
}
</style>
