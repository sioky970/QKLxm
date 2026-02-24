<template>
	<view class="page">
		<!-- 自定义导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="nav-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<text class="nav-title">充值</text>
				<view class="nav-right" @click="goToRecords">
					<text class="record-text">充值记录</text>
				</view>
			</view>
		</view>

		<!-- 主内容区域 -->
		<scroll-view scroll-y class="content-scroll" :style="{ paddingTop: navHeight + 'px' }">
			<!-- 币种显示 -->
			<view class="currency-section">
				<view class="currency-info">
					<image class="currency-icon" src="/static/crypto-icons/usdt.svg" mode="aspectFit" />
					<view class="currency-detail">
						<text class="currency-name">USDT</text>
						<text class="currency-desc">Tether USD</text>
					</view>
				</view>
			</view>

			<!-- 网络选择 -->
			<view class="section-card">
				<view class="section-title">选择充值网络</view>
				<view class="network-list">
					<view 
						class="network-item" 
						v-for="addr in addresses" 
						:key="addr.network"
						:class="{ active: selectedNetwork === addr.network }"
						@click="selectNetwork(addr)"
					>
						<text class="network-name">{{ addr.network }}</text>
						<text class="network-check" v-if="selectedNetwork === addr.network">✓</text>
					</view>
				</view>
				<view class="no-network" v-if="addresses.length === 0 && !loading">
					<text>暂无可用的充值网络</text>
				</view>
			</view>

			<!-- 充值地址 -->
			<view class="section-card" v-if="currentAddress">
				<view class="section-title">充值地址</view>
				
				<!-- 二维码 -->
				<view class="qrcode-box">
					<QRCode 
						v-if="currentAddress.address" 
						:text="currentAddress.address" 
						:size="150"
						:margin="8"
					/>
				</view>
				
				<!-- 地址文本 -->
				<view class="address-box">
					<text class="address-text" selectable>{{ currentAddress.address }}</text>
					<view class="copy-btn" @click="copyAddress">
						<text>复制</text>
					</view>
				</view>
			</view>

			<!-- 充值金额 -->
			<view class="section-card">
				<view class="section-title">充值金额</view>
				<view class="amount-input-box">
					<text class="amount-prefix">USDT</text>
					<input 
						class="amount-input" 
						type="digit" 
						v-model="amount" 
						placeholder="请输入充值金额"
						placeholder-class="placeholder"
					/>
				</view>
				<text class="amount-tip">最小充值金额: 10 USDT</text>
			</view>

			<!-- 转账截图 -->
			<view class="section-card">
				<view class="section-title">上传转账截图</view>
				<view class="upload-box" @click="chooseImage">
					<image v-if="screenshotPath" class="preview-img" :src="screenshotPath" mode="aspectFit" />
					<view v-else class="upload-placeholder">
						<text class="upload-icon">+</text>
						<text class="upload-text">点击上传截图</text>
						<text class="upload-tip">支持jpg/png/gif，最大20MB</text>
					</view>
				</view>
			</view>

			<!-- 温馨提示 -->
			<view class="tips-section">
				<view class="tips-title">温馨提示</view>
				<view class="tips-list">
					<text class="tips-item">• 请确保转账金额与填写金额一致</text>
					<text class="tips-item">• 请勿向非 {{ selectedNetwork }} 网络地址充值</text>
					<text class="tips-item">• 充值订单24小时内有效，过期需重新提交</text>
					<text class="tips-item">• 审核通过后，USDT将自动到账</text>
					<text class="tips-item">• 如有疑问，请联系在线客服</text>
				</view>
			</view>

			<!-- 底部占位 -->
			<view class="bottom-placeholder"></view>
		</scroll-view>

		<!-- 底部提交按钮 -->
		<view class="submit-section">
			<button 
				class="submit-btn" 
				:class="{ disabled: !canSubmit }"
				:disabled="!canSubmit || submitting"
				@click="submitOrder"
			>
				{{ submitting ? '提交中...' : '提交充值订单' }}
			</button>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getDepositAddresses, createDepositOrder, uploadDepositScreenshot } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'
import QRCode from '@/components/QRCode.vue'

// 状态栏高度
const statusBarHeight = ref(0)
const navHeight = ref(88)

// 数据状态
const loading = ref(false)
const submitting = ref(false)
const addresses = ref([])
const selectedNetwork = ref('')
const currentAddress = ref(null)
const amount = ref('')
const screenshotPath = ref('')
const currentOrderId = ref(null)

// 计算是否可以提交
const canSubmit = computed(() => {
	return selectedNetwork.value && 
		   currentAddress.value && 
		   amount.value && 
		   parseFloat(amount.value) >= 10 &&
		   screenshotPath.value
})

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	navHeight.value = statusBarHeight.value + 44
	
	fetchAddresses()
})

// 获取充值地址列表
const fetchAddresses = async () => {
	loading.value = true
	try {
		const res = await getDepositAddresses()
		addresses.value = res.data || []
		
		// 默认选中第一个
		if (addresses.value.length > 0) {
			selectNetwork(addresses.value[0])
		}
	} catch (err) {
		console.error('获取充值地址失败:', err)
		uni.showToast({ title: err.message || '获取充值地址失败', icon: 'none' })
	} finally {
		loading.value = false
	}
}

// 选择网络
const selectNetwork = (addr) => {
	selectedNetwork.value = addr.network
	currentAddress.value = addr
}

// 复制地址
const copyAddress = () => {
	if (!currentAddress.value?.address) return
	
	uni.setClipboardData({
		data: currentAddress.value.address,
		success: () => {
			uni.showToast({ title: '复制成功', icon: 'success' })
		}
	})
}

// 选择图片
const chooseImage = () => {
	uni.chooseImage({
		count: 1,
		sizeType: ['compressed'],
		sourceType: ['album', 'camera'],
		success: (res) => {
			const tempPath = res.tempFilePaths[0]
			const size = res.tempFiles[0].size
			
			// 检查文件大小 (20MB)
			if (size > 20 * 1024 * 1024) {
				uni.showToast({ title: '图片大小不能超过20MB', icon: 'none' })
				return
			}
			
			screenshotPath.value = tempPath
		}
	})
}

// 提交订单
const submitOrder = async () => {
	if (!canSubmit.value || submitting.value) return
	
	// 验证金额
	const amountNum = parseFloat(amount.value)
	if (isNaN(amountNum) || amountNum < 10) {
		uni.showToast({ title: '充值金额不能小于10 USDT', icon: 'none' })
		return
	}
	
	submitting.value = true
	
	try {
		// 1. 创建订单
		const orderRes = await createDepositOrder({
			network: selectedNetwork.value,
			amount: amountNum
		})
		
		const order = orderRes.data
		currentOrderId.value = order.id
		
		// 2. 上传截图
		uni.showLoading({ title: '上传截图中...' })
		await uploadDepositScreenshot(screenshotPath.value, order.id)
		uni.hideLoading()
		
		// 3. 提交成功
		uni.showModal({
			title: '提交成功',
			content: `订单号: ${order.order_no}\n充值金额: ${amountNum} USDT\n\n请等待管理员审核，审核通过后将自动到账。`,
			showCancel: false,
			confirmText: '查看记录',
			success: (res) => {
				if (res.confirm) {
					goToRecords()
				}
			}
		})
		
		// 重置表单
		amount.value = ''
		screenshotPath.value = ''
		currentOrderId.value = null
		
	} catch (err) {
		console.error('提交订单失败:', err)
		uni.showToast({ title: err.message || '提交失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}

// 返回上一页
const goBack = () => {
	uni.navigateBack()
}

// 跳转充值记录
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

/* 导航栏 */
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

/* 币种显示 */
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

/* 通用卡片 */
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

/* 网络选择 */
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

/* 二维码 */
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

/* 地址显示 */
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

/* 金额输入 */
.amount-input-box {
	display: flex;
	align-items: center;
	background: #f5f5f5;
	padding: 0 24rpx;
	border-radius: 12rpx;
	height: 88rpx;
	
	.amount-prefix {
		font-size: 30rpx;
		color: #333;
		font-weight: 500;
		margin-right: 20rpx;
	}
	
	.amount-input {
		flex: 1;
		font-size: 32rpx;
		color: #333;
	}
	
	.placeholder {
		color: #999;
	}
}

.amount-tip {
	font-size: 24rpx;
	color: #999;
	margin-top: 16rpx;
}

/* 上传截图 */
.upload-box {
	width: 100%;
	height: 300rpx;
	background: #f5f5f5;
	border-radius: 12rpx;
	border: 2rpx dashed #ddd;
	display: flex;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	
	.preview-img {
		width: 100%;
		height: 100%;
	}
	
	.upload-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		
		.upload-icon {
			font-size: 80rpx;
			color: #ccc;
			line-height: 1;
		}
		
		.upload-text {
			font-size: 28rpx;
			color: #666;
			margin-top: 16rpx;
		}
		
		.upload-tip {
			font-size: 24rpx;
			color: #999;
			margin-top: 10rpx;
		}
	}
}

/* 温馨提示 */
.tips-section {
	background: #fff;
	padding: 30rpx;
	
	.tips-title {
		font-size: 28rpx;
		font-weight: 600;
		color: #FF9500;
		margin-bottom: 20rpx;
	}
	
	.tips-list {
		display: flex;
		flex-direction: column;
		gap: 12rpx;
		
		.tips-item {
			font-size: 24rpx;
			color: #666;
			line-height: 1.6;
		}
	}
}

.bottom-placeholder {
	height: 40rpx;
}

/* 提交按钮 */
.submit-section {
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
