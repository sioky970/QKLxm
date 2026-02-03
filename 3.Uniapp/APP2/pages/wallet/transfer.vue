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
				<view class="transfer-box">
					<view class="transfer-main">
						<!-- 转出账户 -->
						<view class="transfer-item">
							<view class="transfer-label">从</view>
							<picker @change="onFromChange" :value="fromIndex" :range="wallets" range-key="name" class="account-picker">
								<view class="picker-content">
									<view class="account-brief">
										<view class="account-icon" :style="{ background: wallets[fromIndex].color }">
											<SvgIcon :name="wallets[fromIndex].icon" :size="18" color="#fff" />
										</view>
										<text class="account-name">{{ wallets[fromIndex].name }}</text>
									</view>
									<SvgIcon name="arrow-down" :size="16" color="#999" />
								</view>
							</picker>
						</view>
						
						<!-- 分割线及交换按钮 -->
						<view class="transfer-divider">
							<view class="divider-line"></view>
							<view class="swap-btn-mini" @tap="swapDirection">
								<SvgIcon name="swap" :size="20" color="#007AFF" />
							</view>
						</view>
						
						<!-- 转入账户 -->
						<view class="transfer-item">
							<view class="transfer-label">到</view>
							<picker @change="onToChange" :value="toIndex" :range="wallets" range-key="name" class="account-picker">
								<view class="picker-content">
									<view class="account-brief">
										<view class="account-icon" :style="{ background: wallets[toIndex].color }">
											<SvgIcon :name="wallets[toIndex].icon" :size="18" color="#fff" />
										</view>
										<text class="account-name">{{ wallets[toIndex].name }}</text>
									</view>
									<SvgIcon name="arrow-down" :size="16" color="#999" />
								</view>
							</picker>
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
import { onShow, onLoad } from '@dcloudio/uni-app'
import SvgIcon from '@/components/SvgIcon.vue'
import walletStore from '@/stores/walletStore.js'
import { walletTransferApi } from '@/utils/api.js'

const statusBarHeight = ref(0)
const fromWallet = ref('spot') // 转出钱包
const toWallet = ref('contract') // 转入钱包
let initialWalletType = null // 存储从URL参数传入的钱包类型
const transferAmount = ref('')
const amountError = ref('')
const isTransferring = ref(false)
const sliderValue = ref(0)

const wallets = [
	{ id: 'spot', name: '现货账户', icon: 'wallet', color: 'linear-gradient(135deg, #52c41a 0%, #389e0d 100%)' },
	{ id: 'contract', name: '永续合约账户', icon: 'chart', color: 'linear-gradient(135deg, #fa8c16 0%, #d46b08 100%)' },
	{ id: 'delivery', name: '交割合约账户', icon: 'position', color: 'linear-gradient(135deg, #722ed1 0%, #531dab 100%)' },
	{ id: 'fund', name: '资金账户', icon: 'wallet', color: 'linear-gradient(135deg, #00C853 0%, #69F0AE 100%)' }
]

const fromIndex = computed(() => wallets.findIndex(w => w.id === fromWallet.value))
const toIndex = computed(() => wallets.findIndex(w => w.id === toWallet.value))

const onFromChange = (e) => {
	const index = e.detail.value
	setFromWallet(wallets[index].id)
}

const onToChange = (e) => {
	const index = e.detail.value
	setToWallet(wallets[index].id)
}

// 获取钱包余额
const spotBalance = computed(() => walletStore.state.spotBalance)
const contractBalance = computed(() => walletStore.state.contractBalance)
const deliveryBalance = computed(() => walletStore.state.deliveryBalance)
const fundBalance = computed(() => walletStore.state.fundBalance)

// 可用余额（根据转出钱包）
const availableBalance = computed(() => {
	if (fromWallet.value === 'spot') {
		return spotBalance.value
	} else if (fromWallet.value === 'contract') {
		return contractBalance.value
	} else if (fromWallet.value === 'delivery') {
		return deliveryBalance.value
	} else if (fromWallet.value === 'fund') {
		return fundBalance.value
	}
	return 0
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

// 设置转出钱包
const setFromWallet = (wallet) => {
	fromWallet.value = wallet
	// 如果转出钱包与转入钱包相同，切换转入钱包
	if (toWallet.value === wallet) {
		if (wallet === 'spot') {
			toWallet.value = 'contract'
		} else if (wallet === 'contract') {
			toWallet.value = 'delivery'
		} else if (wallet === 'delivery') {
			toWallet.value = 'fund'
		} else {
			toWallet.value = 'spot'
		}
	}
	transferAmount.value = ''
	amountError.value = ''
	sliderValue.value = 0
}

// 设置转入钱包
const setToWallet = (wallet) => {
	toWallet.value = wallet
	// 如果转入钱包与转出钱包相同，切换转出钱包
	if (fromWallet.value === wallet) {
		if (wallet === 'contract') {
			fromWallet.value = 'spot'
		} else if (wallet === 'delivery') {
			fromWallet.value = 'contract'
		} else if (wallet === 'fund') {
			fromWallet.value = 'delivery'
		} else {
			fromWallet.value = 'contract'
		}
	}
	transferAmount.value = ''
	amountError.value = ''
	sliderValue.value = 0
}

// 快捷转账到交割账户
const quickTransfer = (from, to) => {
	fromWallet.value = from
	toWallet.value = to
	transferAmount.value = ''
	amountError.value = ''
	sliderValue.value = 0
}

// 切换划转方向
const swapDirection = () => {
	const temp = fromWallet.value
	fromWallet.value = toWallet.value
	toWallet.value = temp
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

	try {
		console.log('[Transfer] 开始划转:', { fromWallet: fromWallet.value, toWallet: toWallet.value, amount })

		const res = await walletTransferApi.transfer({
			from_wallet: fromWallet.value,
			to_wallet: toWallet.value,
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

// 页面加载时读取URL参数
onLoad((options) => {
	if (options && options.walletType) {
		const validTypes = ['spot', 'contract', 'delivery', 'fund']
		if (validTypes.includes(options.walletType)) {
			initialWalletType = options.walletType
			console.log('[Transfer] 收到钱包类型参数:', initialWalletType)
		}
	}
})

// 页面显示时刷新数据
onShow(() => {
	walletStore.fetchAssetData(true)
})

onMounted(() => {
	// 获取状态栏高度
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 如果有URL参数，设置初始钱包
	if (initialWalletType) {
		const validWallets = ['spot', 'contract', 'delivery', 'fund']
		if (validWallets.includes(initialWalletType)) {
			// 设置转出钱包为目标钱包，转入钱包为默认的下一个钱包
			fromWallet.value = initialWalletType
			
			// 根据转出钱包设置合理的转入钱包
			if (initialWalletType === 'spot') {
				toWallet.value = 'contract'
			} else if (initialWalletType === 'contract') {
				toWallet.value = 'delivery'
			} else if (initialWalletType === 'delivery') {
				toWallet.value = 'fund'
			} else {
				toWallet.value = 'spot'
			}
			
			console.log('[Transfer] 初始化钱包:', { from: fromWallet.value, to: toWallet.value })
		}
	}
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

	.transfer-box {
		.transfer-main {
			display: flex;
			flex-direction: column;
		}

		.transfer-item {
			display: flex;
			align-items: center;
			padding: 10rpx 0;

			.transfer-label {
				width: 60rpx;
				font-size: 28rpx;
				color: #999;
			}

			.account-picker {
				flex: 1;

				.picker-content {
					display: flex;
					align-items: center;
					justify-content: space-between;
					padding: 24rpx;
					background: #f8f9fa;
					border-radius: 16rpx;

					.account-brief {
						display: flex;
						align-items: center;
						gap: 16rpx;

						.account-icon {
							width: 48rpx;
							height: 48rpx;
							border-radius: 50%;
							display: flex;
							align-items: center;
							justify-content: center;
						}

						.account-name {
							font-size: 30rpx;
							font-weight: 500;
							color: #333;
						}
					}
				}
			}
		}

		.transfer-divider {
			position: relative;
			height: 20rpx;
			display: flex;
			align-items: center;
			margin-left: 60rpx;
			margin-right: 20rpx;

			.divider-line {
				flex: 1;
				height: 1rpx;
				background: #f0f0f0;
			}

			.swap-btn-mini {
				position: absolute;
				right: 40rpx;
				width: 56rpx;
				height: 56rpx;
				background: #fff;
				border-radius: 50%;
				box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
				display: flex;
				align-items: center;
				justify-content: center;
				z-index: 2;
				transition: all 0.3s ease;

				&:active {
					transform: scale(0.9) rotate(180deg);
				}
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
