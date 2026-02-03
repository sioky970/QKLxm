<template>
	<view class="page">
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="nav-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<text class="nav-title">添加资金</text>
				<view class="nav-right" @click="goToRecords" style="visibility: hidden;">
					<text class="record-text">记录</text>
				</view>
			</view>
		</view>

		<scroll-view scroll-y class="content-scroll" :style="{ paddingTop: navHeight + 'px' }">
			<view class="currency-section">
				<view class="currency-info">
					<image class="currency-icon" src="/static/crypto-icons/usdt.svg" mode="aspectFit" />
					<view class="currency-detail">
						<text class="currency-name">USDT</text>
						<text class="currency-desc">Tether USD</text>
					</view>
				</view>
			</view>

			<view class="section-card">
				<view class="section-title">选择网络</view>
				<view class="network-list">
					<view 
						class="network-item" 
						v-for="addr in filteredAddresses" 
						:key="addr.network"
						:class="{ active: selectedNetwork === addr.network }"
						@click="selectNetwork(addr)"
					>
						<text class="network-name">{{ addr.network }}</text>
						<text class="network-check" v-if="selectedNetwork === addr.network">✓</text>
					</view>
				</view>
				<view class="no-network" v-if="filteredAddresses.length === 0 && !loading">
					<text>暂无可用的充值网络</text>
				</view>
			</view>

			<view class="section-card" v-if="currentAddress">
				<view class="section-title">钱包地址</view>
				
				<view class="qrcode-box">
					<QRCode 
						v-if="currentAddress.address" 
						:text="currentAddress.address" 
						:size="150"
						:margin="8"
					/>
				</view>
				
				<view class="address-box">
					<text class="address-text" selectable>{{ currentAddress.address }}</text>
					<view class="copy-btn" @click="copyAddress">
						<text>复制</text>
					</view>
				</view>
			</view>

			<view class="bottom-placeholder"></view>
		</scroll-view>

		<view class="submit-section">
			<button 
				class="submit-btn" 
				:class="{ disabled: !canSubmit }"
				:disabled="!canSubmit || submitting"
				@click="submitOrder"
			>
				{{ submitting ? '提交中...' : '提交' }}
			</button>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getDepositAddresses, createDepositOrder, uploadDepositScreenshot } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'
import QRCode from '@/components/QRCode.vue'

const statusBarHeight = ref(0)
const navHeight = ref(88)

const loading = ref(false)
const submitting = ref(false)
const addresses = ref([])
const selectedNetwork = ref('')
const currentAddress = ref(null)

const filteredAddresses = computed(() => {
	return addresses.value.filter(addr => 
		addr.network === 'TRC20' || addr.network === 'ERC20'
	)
})

const canSubmit = computed(() => {
	return selectedNetwork.value && currentAddress.value
})

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	navHeight.value = statusBarHeight.value + 44
	
	fetchAddresses()
})

const fetchAddresses = async () => {
	loading.value = true
	try {
		const res = await getDepositAddresses()
		addresses.value = res.data || []
		
		if (filteredAddresses.value.length > 0) {
			selectNetwork(filteredAddresses.value[0])
		}
	} catch (err) {
		console.error('获取充值地址失败:', err)
		uni.showToast({ title: err.message || '获取充值地址失败', icon: 'none' })
	} finally {
		loading.value = false
	}
}

const selectNetwork = (addr) => {
	selectedNetwork.value = addr.network
	currentAddress.value = addr
}

const copyAddress = () => {
	if (!currentAddress.value?.address) return
	
	uni.setClipboardData({
		data: currentAddress.value.address,
		success: () => {
			uni.showToast({ title: '复制成功', icon: 'success' })
		}
	})
}

const submitOrder = async () => {
	if (!canSubmit.value || submitting.value) return
	
	submitting.value = true
	
	try {
		const orderRes = await createDepositOrder({
			network: selectedNetwork.value,
			amount: 0
		})
		
		const order = orderRes.data
		
		uni.showLoading({ title: '上传截图中...' })
		await uploadDepositScreenshot('', order.id)
		uni.hideLoading()
		
		uni.showModal({
			title: '提交成功',
			content: `订单号: ${order.order_no}\n\n请等待管理员审核，审核通过后将自动到账。`,
			showCancel: false,
			confirmText: '查看记录',
			success: (res) => {
				if (res.confirm) {
					goToRecords()
				}
			}
		})
		
	} catch (err) {
		console.error('提交订单失败:', err)
		uni.showToast({ title: err.message || '提交失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}

const goBack = () => {
	uni.navigateBack()
}

const goToRecords = () => {
	uni.navigateTo({ url: '/pages/wallet/deposit-records' })
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	min-height: 100vh;
	background: #f5f5f5;
}

.custom-navbar {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: #fff;
	z-index: 100;
	
	.navbar-content {
		height: 88rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 30rpx;
	}
	
	.nav-left {
		width: 60rpx;
		height: 60rpx;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	
	.nav-title {
		font-size: 34rpx;
		font-weight: 600;
		color: #333;
	}
	
	.nav-right {
		.record-text {
			font-size: 28rpx;
			color: #007AFF;
		}
	}
}

.content-scroll {
	min-height: 100vh;
	padding-bottom: 160rpx;
}

.currency-section {
	background: #fff;
	padding: 30rpx;
	margin-bottom: 20rpx;
	
	.currency-info {
		display: flex;
		align-items: center;
		
		.currency-icon {
			width: 80rpx;
			height: 80rpx;
			margin-right: 24rpx;
		}
		
		.currency-detail {
			display: flex;
			flex-direction: column;
			
			.currency-name {
				font-size: 36rpx;
				font-weight: 600;
				color: #333;
			}
			
			.currency-desc {
				font-size: 26rpx;
				color: #999;
				margin-top: 6rpx;
			}
		}
	}
}

.section-card {
	background: #fff;
	padding: 30rpx;
	margin-bottom: 20rpx;
	
	.section-title {
		font-size: 30rpx;
		font-weight: 600;
		color: #333;
		margin-bottom: 24rpx;
	}
}

.network-list {
	display: flex;
	flex-wrap: wrap;
	gap: 20rpx;
	
	.network-item {
		flex: 0 0 calc(50% - 10rpx);
		height: 80rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 24rpx;
		background: #f5f5f5;
		border-radius: 12rpx;
		border: 2rpx solid transparent;
		
		&.active {
			background: #E6F7FF;
			border-color: #007AFF;
		}
		
		.network-name {
			font-size: 28rpx;
			color: #333;
		}
		
		.network-check {
			font-size: 28rpx;
			color: #007AFF;
		}
	}
}

.no-network {
	text-align: center;
	padding: 40rpx;
	color: #999;
}

.qrcode-box {
	display: flex;
	justify-content: center;
	margin-bottom: 30rpx;
	
	.qrcode-img {
		width: 300rpx;
		height: 300rpx;
		border: 1rpx solid #eee;
		border-radius: 12rpx;
	}
	
	.qrcode-placeholder {
		width: 300rpx;
		height: 300rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f5f5f5;
		border-radius: 12rpx;
		
		text {
			font-size: 24rpx;
			color: #999;
		}
	}
}

.address-box {
	display: flex;
	align-items: center;
	background: #f5f5f5;
	padding: 20rpx;
	border-radius: 12rpx;
	
	.address-text {
		flex: 1;
		font-size: 24rpx;
		color: #666;
		word-break: break-all;
		line-height: 1.5;
	}
	
	.copy-btn {
		margin-left: 20rpx;
		padding: 12rpx 24rpx;
		background: #007AFF;
		border-radius: 8rpx;
		
		text {
			font-size: 24rpx;
			color: #fff;
		}
	}
}

.bottom-placeholder {
	height: 40rpx;
}

.submit-section {
	display: none;
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	background: #fff;
	padding: 20rpx 30rpx;
	padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
	box-shadow: 0 -4rpx 12rpx rgba(0, 0, 0, 0.05);
	
	.submit-btn {
		width: 100%;
		height: 88rpx;
		background: linear-gradient(135deg, #007AFF 0%, #0052D9 100%);
		color: #fff;
		font-size: 32rpx;
		font-weight: 600;
		border-radius: 44rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		
		&.disabled {
			background: #ccc;
		}
	}
}
</style>
