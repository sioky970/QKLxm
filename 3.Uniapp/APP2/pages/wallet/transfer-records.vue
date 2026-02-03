<template>
	<view class="page">
		<!-- 导航栏 -->
		<view class="navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="nav-content">
				<view class="nav-left" @click="goBack">
					<text class="nav-back">‹</text>
				</view>
				<text class="nav-title">划转记录</text>
				<view class="nav-right"></view>
			</view>
		</view>

		<!-- 主体内容 -->
		<view class="content" :style="{ marginTop: (statusBarHeight + 44) + 'px' }">
			<!-- 记录列表 -->
			<view class="record-list" v-if="records.length > 0">
				<view class="record-item" v-for="item in records" :key="item.id">
					<view class="record-header">
						<view class="record-direction">
							<text class="direction-text">{{ formatWalletType(item.from_wallet_type) }}</text>
							<SvgIcon name="arrow-right" :size="14" color="#007AFF" />
							<text class="direction-text">{{ formatWalletType(item.to_wallet_type) }}</text>
						</view>
						<view class="record-status" :class="getStatusClass(item.status)">
							<text class="status-text">{{ formatStatus(item.status) }}</text>
						</view>
					</view>
					
					<view class="record-body">
						<view class="record-row">
							<text class="row-label">划转金额</text>
							<text class="row-value amount">{{ item.amount }} USDT</text>
						</view>
						<view class="record-row">
							<text class="row-label">手续费</text>
							<text class="row-value">{{ item.fee }} USDT</text>
						</view>
						<view class="record-row">
							<text class="row-label">流水号</text>
							<text class="row-value no">{{ item.transfer_no }}</text>
						</view>
						<view class="record-row">
							<text class="row-label">时间</text>
							<text class="row-value">{{ formatTime(item.created_time) }}</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 空状态 -->
			<view class="empty-state" v-else-if="!isLoading">
				<SvgIcon name="empty" :size="80" color="#ddd" />
				<text class="empty-text">暂无划转记录</text>
			</view>

			<!-- 加载更多 -->
			<view class="load-more" v-if="hasMore">
				<text class="load-text">{{ isLoading ? '加载中...' : '上拉加载更多' }}</text>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import SvgIcon from '@/components/SvgIcon.vue'
import { walletTransferApi } from '@/utils/api.js'

const statusBarHeight = ref(0)
const records = ref([])
const page = ref(1)
const pageSize = ref(20)
const hasMore = ref(true)
const isLoading = ref(false)

const walletTypeMap = {
	'spot': '现货账户',
	'contract': '永续合约账户',
	'delivery': '交割合约账户'
}

const statusMap = {
	1: '处理中',
	2: '成功',
	3: '失败'
}

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	fetchRecords()
})

async function fetchRecords(isLoadMore = false) {
	if (isLoading.value) return
	
	isLoading.value = true
	
	try {
		const res = await walletTransferApi.getTransferRecords({
			page: page.value,
			page_size: pageSize.value
		})
		
		if (res.code === 0 && res.data) {
			const list = res.data.list || []
			
			if (isLoadMore) {
				records.value = [...records.value, ...list]
			} else {
				records.value = list
			}
			
			hasMore.value = list.length >= pageSize.value
			
			if (list.length > 0) {
				page.value++
			}
		}
	} catch (error) {
		console.error('获取划转记录失败:', error)
	} finally {
		isLoading.value = false
	}
}

function formatWalletType(type) {
	return walletTypeMap[type] || type
}

function formatStatus(status) {
	return statusMap[status] || '未知'
}

function getStatusClass(status) {
	switch (status) {
		case 2: return 'success'
		case 3: return 'failed'
		default: return 'pending'
	}
}

function formatTime(timestamp) {
	const date = new Date(timestamp * 1000)
	const year = date.getFullYear()
	const month = String(date.getMonth() + 1).padStart(2, '0')
	const day = String(date.getDate()).padStart(2, '0')
	const hour = String(date.getHours()).padStart(2, '0')
	const minute = String(date.getMinutes()).padStart(2, '0')
	return `${year}-${month}-${day} ${hour}:${minute}`
}

function goBack() {
	uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.page {
	min-height: 100vh;
	background-color: #f5f5f5;
}

.navbar {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: #fff;
	z-index: 100;
	border-bottom: 1px solid #f0f0f0;
}

.nav-content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 44px;
	padding: 0 15px;
}

.nav-left, .nav-right {
	width: 40px;
}

.nav-back {
	font-size: 24px;
	color: #333;
}

.nav-title {
	font-size: 17px;
	font-weight: 600;
	color: #333;
}

.content {
	padding: 15px;
}

.record-list {
	
}

.record-item {
	background: #fff;
	border-radius: 12px;
	padding: 15px;
	margin-bottom: 12px;
}

.record-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-bottom: 12px;
	border-bottom: 1px solid #f0f0f0;
	margin-bottom: 12px;
}

.record-direction {
	display: flex;
	align-items: center;
	
	.direction-text {
		font-size: 14px;
		font-weight: 500;
		color: #333;
	}
	
	.direction-arrow {
		margin: 0 8px;
	}
}

.record-status {
	padding: 4px 10px;
	border-radius: 4px;
	
	&.success {
		background: #e8f5e9;
		
		.status-text {
			color: #34C759;
		}
	}
	
	&.failed {
		background: #ffebee;
		
		.status-text {
			color: #FF3B30;
		}
	}
	
	&.pending {
		background: #fff3e0;
		
		.status-text {
			color: #FF9500;
		}
	}
}

.status-text {
	font-size: 12px;
}

.record-body {
	
}

.record-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 6px 0;
}

.row-label {
	font-size: 13px;
	color: #999;
}

.row-value {
	font-size: 13px;
	color: #333;
	
	&.amount {
		font-size: 15px;
		font-weight: 600;
		color: #007AFF;
	}
	
	&.no {
		font-size: 12px;
		color: #666;
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 100px 0;
}

.empty-text {
	font-size: 14px;
	color: #999;
	margin-top: 15px;
}

.load-more {
	text-align: center;
	padding: 15px;
}

.load-text {
	font-size: 13px;
	color: #999;
}
</style>
