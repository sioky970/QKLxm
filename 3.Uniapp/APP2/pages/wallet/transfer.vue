<template>
	<view class="page">
		<!-- 自定义导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="nav-left" @tap="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<text class="nav-title">资金划转</text>
				<view class="placeholder"></view>
			</view>
		</view>

		<view class="content">
			<!-- 划转方向选择卡片 -->
			<view class="direction-card">
				<view class="direction-title">划转方向</view>
				<view class="direction-selector">
					<!-- 转出账户 -->
					<view class="account-box" :class="{ active: transferDirection === 'spot_to_contract' }" @tap="setDirection('spot_to_contract')">
						<view class="account-icon spot-icon">
							<SvgIcon name="wallet" :size="28" color="#fff" />
						</view>
						<view class="account-info">
							<text class="account-name">现货账户</text>
							<text class="account-balance">{{ formatBalance(spotBalance) }} USDT</text>
						</view>
						<view class="check-icon" v-if="transferDirection === 'spot_to_contract'">
							<SvgIcon name="check" :size="20" color="#007AFF" />
						</view>
					</view>

					<!-- 交换按钮 -->
					<view class="swap-btn" @tap="swapDirection">
						<view class="swap-icon">
							<SvgIcon name="swap" :size="24" color="#007AFF" />
						</view>
					</view>

					<!-- 转入账户 -->
					<view class="account-box" :class="{ active: transferDirection === 'contract_to_spot' }" @tap="setDirection('contract_to_spot')">
						<view class="account-icon contract-icon">
							<SvgIcon name="chart" :size="28" color="#fff" />
						</view>
						<view class="account-info">
							<text class="account-name">合约账户</text>
							<text class="account-balance">{{ formatBalance(contractBalance) }} USDT</text>
						</view>
						<view class="check-icon" v-if="transferDirection === 'contract_to_spot'">
							<SvgIcon name="check" :size="20" color="#007AFF" />
						</view>
					</view>
				</view>
			</view>

			<!-- 金额输入卡片 -->
			<view class="amount-card">
				<view class="amount-header">
					<text class="amount-label">划转金额</text>
					<text class="available-balance">可用: {{ formatBalance(availableBalance) }} USDT</text>
				</view>
				<view class="amount-input-wrapper">
					<text class="currency-symbol">USDT</text>
					<input
						type="digit"
						class="amount-input"
						v-model="transferAmount"
						placeholder="请输入划转金额"
						:disabled="isTransferring"
						@input="onAmountInput"
					/>
					<view class="max-btn" @tap="setMaxAmount" v-if="!isTransferring">全部</view>
				</view>
				<view class="amount-error" v-if="amountError">{{ amountError }}</view>

				<!-- 百分比滑动条 -->
				<view class="percentage-slider-section">
					<view class="slider-container">
						<slider
							class="percentage-slider"
							:min="0"
							:max="100"
							:step="1"
							:value="sliderValue"
							@change="onSliderChange"
							@changing="onSliderChanging"
							:disabled="isTransferring || availableBalance <= 0"
							activeColor="#007AFF"
							backgroundColor="#E5E5E5"
							block-size="20"
						/>
					</view>
					<view class="percentage-labels">
						<view
							class="percentage-btn"
							:class="{ active: sliderValue === 0 }"
							@tap="setPercentage(0)"
						>
							<text>0%</text>
						</view>
						<view
							class="percentage-btn"
							:class="{ active: sliderValue === 25 }"
							@tap="setPercentage(25)"
						>
							<text>25%</text>
						</view>
						<view
							class="percentage-btn"
							:class="{ active: sliderValue === 50 }"
							@tap="setPercentage(50)"
						>
							<text>50%</text>
						</view>
						<view
							class="percentage-btn"
							:class="{ active: sliderValue === 75 }"
							@tap="setPercentage(75)"
						>
							<text>75%</text>
						</view>
						<view
							class="percentage-btn"
							:class="{ active: sliderValue === 100 }"
							@tap="setPercentage(100)"
						>
							<text>100%</text>
						</view>
					</view>
					<view class="current-percentage" v-if="sliderValue > 0">
						<text class="percentage-value">{{ sliderValue }}%</text>
						<text class="percentage-amount">≈ {{ formatBalance(calculateAmountByPercentage(sliderValue)) }} USDT</text>
					</view>
				</view>
			</view>

			<!-- 划转记录入口 -->
			<view class="records-entry" @tap="goToRecords">
				<text class="records-text">划转记录</text>
				<SvgIcon name="arrow-right" :size="16" color="#999" />
			</view>

			<!-- 温馨提示 -->
			<view class="tips-card">
				<view class="tips-title">
					<SvgIcon name="info" :size="16" color="#999" />
					<text>温馨提示</text>
				</view>
				<view class="tips-list">
					<view class="tip-item">
						<text class="tip-dot">·</text>
						<text class="tip-text">划转实时到账，无手续费</text>
					</view>
					<view class="tip-item">
						<text class="tip-dot">·</text>
						<text class="tip-text">现货账户用于币币交易</text>
					</view>
					<view class="tip-item">
						<text class="tip-dot">·</text>
						<text class="tip-text">合约账户用于合约交易</text>
					</view>
				</view>
			</view>
		</view>

		<!-- 底部确认按钮 -->
		<view class="bottom-bar">
			<view
				class="confirm-btn"
				:class="{ disabled: !canSubmit, loading: isTransferring }"
				@tap="handleTransfer"
			>
				<text v-if="!isTransferring">确认划转</text>
				<view class="loading-spinner" v-else>
					<view class="spinner"></view>
					<text>划转中...</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SvgIcon from '@/components/SvgIcon.vue'
import walletStore from '@/stores/walletStore.js'
import { walletTransferApi } from '@/utils/api.js'

const statusBarHeight = ref(0)
const transferDirection = ref('spot_to_contract')
const transferAmount = ref('')
const amountError = ref('')
const isTransferring = ref(false)
const sliderValue = ref(0)

// 获取钱包余额
const spotBalance = computed(() => walletStore.state.spotBalance)
const contractBalance = computed(() => walletStore.state.contractBalance)

// 可用余额（根据划转方向）
const availableBalance = computed(() => {
	return transferDirection.value === 'spot_to_contract' ? spotBalance.value : contractBalance.value
})

// 格式化余额显示
const formatBalance = (value) => {
	const num = parseFloat(value) || 0
	return num.toFixed(4)
}

// 根据百分比计算金额
const calculateAmountByPercentage = (percentage) => {
	return (availableBalance.value * percentage) / 100
}

// 根据金额计算百分比
const calculatePercentageByAmount = (amount) => {
	if (availableBalance.value <= 0) return 0
	const percentage = (amount / availableBalance.value) * 100
	return Math.min(100, Math.max(0, Math.round(percentage)))
}

// 是否可以提交
const canSubmit = computed(() => {
	const amount = parseFloat(transferAmount.value)
	return amount > 0 && amount <= availableBalance.value && !isTransferring.value && !amountError.value
})

// 设置划转方向
const setDirection = (direction) => {
	transferDirection.value = direction
	transferAmount.value = ''
	amountError.value = ''
	sliderValue.value = 0
}

// 切换划转方向
const swapDirection = () => {
	transferDirection.value = transferDirection.value === 'spot_to_contract' ? 'contract_to_spot' : 'spot_to_contract'
	transferAmount.value = ''
	amountError.value = ''
	sliderValue.value = 0
}

// 金额输入处理
const onAmountInput = () => {
	validateInput()
	// 同步更新滑动条
	const amount = parseFloat(transferAmount.value) || 0
	sliderValue.value = calculatePercentageByAmount(amount)
}

// 滑动条变化处理（滑动结束）
const onSliderChange = (e) => {
	sliderValue.value = e.detail.value
	updateAmountBySlider()
}

// 滑动条拖动中处理
const onSliderChanging = (e) => {
	sliderValue.value = e.detail.value
	updateAmountBySlider()
}

// 根据滑动条更新金额
const updateAmountBySlider = () => {
	if (availableBalance.value > 0) {
		const amount = calculateAmountByPercentage(sliderValue.value)
		// 如果滑动到100%，使用全部余额，避免精度问题
		if (sliderValue.value === 100) {
			transferAmount.value = availableBalance.value.toString()
		} else {
			transferAmount.value = amount.toFixed(4)
		}
		validateInput()
	}
}

// 设置百分比
const setPercentage = (percentage) => {
	sliderValue.value = percentage
	updateAmountBySlider()
}

// 设置最大金额
const setMaxAmount = () => {
	transferAmount.value = availableBalance.value.toString()
	amountError.value = ''
	sliderValue.value = 100
}

// 验证输入
const validateInput = () => {
	const amount = parseFloat(transferAmount.value)
	if (isNaN(amount) || amount < 0) {
		amountError.value = '请输入有效的金额'
		return false
	}
	if (amount > availableBalance.value) {
		amountError.value = '划转金额不能超过可用余额'
		return false
	}
	amountError.value = ''
	return true
}

// 处理划转
const handleTransfer = async () => {
	if (!canSubmit.value) return
	if (!validateInput()) return

	isTransferring.value = true
	const amount = parseFloat(transferAmount.value)

	const fromWallet = transferDirection.value === 'spot_to_contract' ? 'spot' : 'contract'
	const toWallet = transferDirection.value === 'spot_to_contract' ? 'contract' : 'spot'

	try {
		console.log('[Transfer] 开始划转:', { fromWallet, toWallet, amount })

		const res = await walletTransferApi.transfer({
			from_wallet: fromWallet,
			to_wallet: toWallet,
			amount: amount
		})

		console.log('[Transfer] 划转响应:', res)

		if (res.data && res.data.success) {
			uni.showToast({
				title: '划转成功',
				icon: 'success',
				duration: 2000
			})

			// 静默刷新余额数据（不显示加载状态）
			walletStore.fetchAssetData(false)

			// 延迟返回
			setTimeout(() => {
				goBack()
			}, 1500)
		} else {
			uni.showToast({
				title: res.data?.error_msg || '划转失败',
				icon: 'none',
				duration: 2000
			})
		}
	} catch (error) {
		console.error('[Transfer] 划转错误:', error)
		// 显示后端返回的具体错误信息
		const errorMsg = error.message || '划转失败，请重试'
		uni.showToast({
			title: errorMsg,
			icon: 'none',
			duration: 2500
		})
	} finally {
		isTransferring.value = false
	}
}

// 返回上一页
const goBack = () => {
	uni.navigateBack()
}

// 跳转到划转记录
const goToRecords = () => {
	uni.navigateTo({ url: '/pages/wallet/transfer-records' })
}

// 页面显示时刷新数据
onShow(() => {
	walletStore.fetchAssetData(true)
})

onMounted(() => {
	// 获取状态栏高度
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
})
</script>

<style scoped lang="scss">
.page {
	min-height: 100vh;
	background: #f5f6fa;
	display: flex;
	flex-direction: column;
}

/* 自定义导航栏 */
.custom-navbar {
	background: #fff;
	position: sticky;
	top: 0;
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
		font-size: 36rpx;
		font-weight: 600;
		color: #333;
	}

	.placeholder {
		width: 60rpx;
	}
}

/* 内容区域 */
.content {
	flex: 1;
	padding: 20rpx;
	padding-bottom: 140rpx;
}

/* 方向选择卡片 */
.direction-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 30rpx;
	margin-bottom: 20rpx;

	.direction-title {
		font-size: 32rpx;
		font-weight: 600;
		color: #333;
		margin-bottom: 24rpx;
	}

	.direction-selector {
		display: flex;
		flex-direction: column;
		gap: 20rpx;
	}

	.account-box {
		display: flex;
		align-items: center;
		padding: 24rpx;
		background: #f8f9fa;
		border-radius: 16rpx;
		border: 2rpx solid transparent;
		transition: all 0.3s ease;

		&.active {
			background: #f0f7ff;
			border-color: #007AFF;
		}

		.account-icon {
			width: 64rpx;
			height: 64rpx;
			border-radius: 50%;
			display: flex;
			align-items: center;
			justify-content: center;
			margin-right: 20rpx;

			&.spot-icon {
				background: linear-gradient(135deg, #52c41a 0%, #389e0d 100%);
			}

			&.contract-icon {
				background: linear-gradient(135deg, #fa8c16 0%, #d46b08 100%);
			}
		}

		.account-info {
			flex: 1;
			display: flex;
			flex-direction: column;

			.account-name {
				font-size: 30rpx;
				font-weight: 500;
				color: #333;
				margin-bottom: 8rpx;
			}

			.account-balance {
				font-size: 26rpx;
				color: #666;
			}
		}

		.check-icon {
			width: 40rpx;
			height: 40rpx;
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}

	.swap-btn {
		display: flex;
		justify-content: center;
		padding: 10rpx 0;

		.swap-icon {
			width: 72rpx;
			height: 72rpx;
			background: #fff;
			border-radius: 50%;
			box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
			display: flex;
			align-items: center;
			justify-content: center;
			transition: transform 0.3s ease;

			&:active {
				transform: rotate(180deg);
			}
		}
	}
}

/* 金额输入卡片 */
.amount-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 30rpx;
	margin-bottom: 20rpx;

	.amount-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 24rpx;

		.amount-label {
			font-size: 32rpx;
			font-weight: 600;
			color: #333;
		}

		.available-balance {
			font-size: 26rpx;
			color: #666;
		}
	}

	.amount-input-wrapper {
		display: flex;
		align-items: center;
		padding: 24rpx;
		background: #f8f9fa;
		border-radius: 16rpx;

		.currency-symbol {
			font-size: 32rpx;
			font-weight: 600;
			color: #333;
			margin-right: 20rpx;
		}

		.amount-input {
			flex: 1;
			font-size: 40rpx;
			font-weight: 600;
			color: #333;
			height: 60rpx;
		}

		.max-btn {
			padding: 12rpx 24rpx;
			background: #007AFF;
			border-radius: 8rpx;
			font-size: 26rpx;
			color: #fff;
			font-weight: 500;

			&:active {
				opacity: 0.8;
			}
		}
	}

	.amount-error {
		margin-top: 16rpx;
		font-size: 26rpx;
		color: #ff4d4f;
	}

	/* 百分比滑动条 */
	.percentage-slider-section {
		margin-top: 30rpx;
		padding-top: 20rpx;
		border-top: 1rpx solid #f0f0f0;

		.slider-container {
			padding: 10rpx 0;

			.percentage-slider {
				width: 100%;
			}
		}

		.percentage-labels {
			display: flex;
			justify-content: space-between;
			padding: 0 10rpx;
			margin-top: 10rpx;

			.percentage-btn {
				padding: 12rpx 20rpx;
				background: #f5f5f5;
				border-radius: 8rpx;
				transition: all 0.2s ease;

				text {
					font-size: 24rpx;
					color: #666;
				}

				&.active {
					background: #007AFF;

					text {
						color: #fff;
					}
				}

				&:active {
					opacity: 0.8;
				}
			}
		}

		.current-percentage {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 16rpx;
			margin-top: 20rpx;
			padding: 16rpx;
			background: #f0f7ff;
			border-radius: 12rpx;

			.percentage-value {
				font-size: 32rpx;
				font-weight: 600;
				color: #007AFF;
			}

			.percentage-amount {
				font-size: 26rpx;
				color: #666;
			}
		}
	}
}

/* 划转记录入口 */
.records-entry {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 30rpx;
	background: #fff;
	border-radius: 24rpx;
	margin-bottom: 20rpx;

	.records-text {
		font-size: 30rpx;
		color: #333;
	}

	&:active {
		background: #f8f9fa;
	}
}

/* 温馨提示卡片 */
.tips-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 30rpx;

	.tips-title {
		display: flex;
		align-items: center;
		gap: 12rpx;
		margin-bottom: 20rpx;

		text {
			font-size: 28rpx;
			font-weight: 500;
			color: #333;
		}
	}

	.tips-list {
		display: flex;
		flex-direction: column;
		gap: 16rpx;
	}

	.tip-item {
		display: flex;
		align-items: flex-start;
		gap: 12rpx;

		.tip-dot {
			font-size: 30rpx;
			color: #999;
			line-height: 1.4;
		}

		.tip-text {
			font-size: 26rpx;
			color: #666;
			line-height: 1.6;
		}
	}
}

/* 底部确认按钮 */
.bottom-bar {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	padding: 20rpx 30rpx 40rpx;
	background: #fff;
	box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.05);

	.confirm-btn {
		height: 96rpx;
		background: linear-gradient(135deg, #007AFF 0%, #0056D6 100%);
		border-radius: 48rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.3s ease;

		text {
			font-size: 32rpx;
			font-weight: 600;
			color: #fff;
		}

		&.disabled {
			background: #d9d9d9;
		}

		&.loading {
			opacity: 0.8;
		}

		&:active:not(.disabled) {
			transform: scale(0.98);
		}
	}

	.loading-spinner {
		display: flex;
		align-items: center;
		gap: 16rpx;

		.spinner {
			width: 32rpx;
			height: 32rpx;
			border: 4rpx solid rgba(255, 255, 255, 0.3);
			border-top-color: #fff;
			border-radius: 50%;
			animation: spin 1s linear infinite;
		}

		text {
			font-size: 30rpx;
			color: #fff;
		}
	}
}

@keyframes spin {
	to {
		transform: rotate(360deg);
	}
}
</style>
