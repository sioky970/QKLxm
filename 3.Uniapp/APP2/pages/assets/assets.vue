<template>
	<view class="page">
		<!-- 顶部渐变容器：包含导航栏和总览区 -->
		<view class="header-gradient-container" :style="{ paddingTop: statusBarHeight + 'px' }">
			<!-- 导航栏 -->
			<view class="custom-navbar">
				<view class="navbar-content">
					<text class="nav-title">总览</text>
					<view class="nav-right">
						<view class="icon-btn" @click="goToRecords">
							<SvgIcon name="document" :size="22" color="#fff" />
						</view>
						<view class="icon-btn" @click="handleLogout">
							<SvgIcon name="logout" :size="22" color="#fff" />
						</view>
					</view>
				</view>
			</view>
			
			<!-- 预估总资产区 -->
			<view class="summary-section">
				<view class="summary-header">
					<text class="summary-label">预估总资产</text>
					<view class="eye-icon" @click="toggleVisibility">
						<SvgIcon :name="isAssetsVisible ? 'eye-open' : 'eye-close'" :size="16" color="rgba(255,255,255,0.7)" />
					</view>
				</view>
				
				<view class="total-amount-box">
					<text class="amount-num">{{ isAssetsVisible ? totalBalance : '********' }}</text>
					<text class="currency">USDT</text>
					<SvgIcon name="arrow-down" :size="12" color="#fff" />
				</view>
				
				<view class="usd-equivalent">
					<text class="usd-text">≈${{ isAssetsVisible ? totalUsdValue : '********' }}</text>
				</view>
				
				<view class="profit-row" @click="goToProfitDetail">
					<text class="profit-label">今日盈亏</text>
					<text class="profit-value" :class="profitData.value >= 0 ? 'up' : 'down'">
						{{ isAssetsVisible ? (profitData.value >= 0 ? '+' : '') + profitData.value : '****' }} USDT({{ isAssetsVisible ? profitData.rate : '****' }})
					</text>
					<SvgIcon name="arrow-right" :size="12" :color="profitData.value >= 0 ? '#90EE90' : '#FFB6C1'" />
				</view>
			</view>
			
			<!-- 现货和合约账户余额 -->
			<view class="wallet-balance-section">
				<view class="wallet-balance-card">
					<view class="wallet-balance-header">
						<text class="wallet-balance-title">账户余额</text>
						<view class="transfer-entry-btn" @tap="openTransferPopup">
							<SvgIcon name="swap" :size="14" color="#007AFF" />
							<text class="transfer-entry-text">划转</text>
						</view>
					</view>
					
					<view class="wallet-balance-row">
						<view class="wallet-balance-item" @tap="openTransferPopup">
							<view class="wallet-balance-left">
								<view class="wallet-icon-small spot-icon">
									<SvgIcon name="wallet" :size="16" color="#fff" />
								</view>
								<text class="wallet-balance-name">现货账户</text>
							</view>
							<view class="wallet-balance-right">
								<text class="wallet-balance-amount">{{ isAssetsVisible ? spotBalance : '****' }} USDT</text>
								<SvgIcon name="arrow-right" :size="12" color="rgba(255,255,255,0.6)" />
							</view>
						</view>
						<view class="wallet-balance-item" @tap="openTransferPopup">
							<view class="wallet-balance-left">
								<view class="wallet-icon-small contract-icon">
									<SvgIcon name="chart" :size="16" color="#fff" />
								</view>
								<text class="wallet-balance-name">永续合约账户</text>
							</view>
							<view class="wallet-balance-right">
								<text class="wallet-balance-amount">{{ isAssetsVisible ? contractBalance : '****' }} USDT</text>
								<SvgIcon name="arrow-right" :size="12" color="rgba(255,255,255,0.6)" />
							</view>
						</view>
						<view class="wallet-balance-item" @tap="openTransferPopup">
							<view class="wallet-balance-left">
								<view class="wallet-icon-small delivery-icon">
									<SvgIcon name="position" :size="16" color="#fff" />
								</view>
								<text class="wallet-balance-name">交割合约账户</text>
							</view>
							<view class="wallet-balance-right">
								<text class="wallet-balance-amount">{{ isAssetsVisible ? deliveryBalance : '****' }} USDT</text>
								<SvgIcon name="arrow-right" :size="12" color="rgba(255,255,255,0.6)" />
							</view>
						</view>
					</view>
				</view>
			</view>
		</view>

		<!-- 分隔条 -->
		<view class="divider"></view>

		<!-- 操作按钮区域 - 固定 -->
		<view class="action-buttons-fixed">
			<view class="action-btn" @click="handleAddAsset">
				<SvgIcon name="wallet" :size="24" color="#007AFF" />
				<text class="btn-text">入款</text>
			</view>
			<view class="action-btn" @click="handleWithdraw">
				<SvgIcon name="download" :size="24" color="#007AFF" />
				<text class="btn-text">提币</text>
			</view>
			<view class="action-btn" @click="goToKyc">
				<SvgIcon name="kyc" :size="24" color="#007AFF" />
				<text class="btn-text">实名认证</text>
			</view>
		</view>

		<!-- 账户标签 - 固定 -->
		<view class="section-header-fixed">
			<text class="section-title">我的资产</text>
		</view>

		<view class="scroll-container">
			<scroll-view scroll-y class="content-scroll">

				<!-- 加密货币资产列表 -->
				<view class="crypto-list">
					<view class="crypto-item" v-for="item in cryptoAssets" :key="item.id">
						<view class="crypto-left">
							<image class="crypto-icon" :src="getCryptoIcon(item.symbol)" mode="aspectFit" />
							<view class="crypto-info">
								<text class="crypto-name">{{ item.name }}</text>
								<text class="crypto-symbol">{{ item.symbol }}</text>
							</view>
						</view>
						<view class="crypto-right">
							<view class="amount-line">
								<text class="crypto-amount">{{ isAssetsVisible ? item.balance : '****' }}</text>
							</view>
							<text class="crypto-usd">≈ ${{ isAssetsVisible ? item.usdValue : '****' }}</text>
						</view>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 底部占位符 -->
		<view class="tabbar-placeholder"></view>
		
		<!-- 自定义底部栏 -->
		<CustomTabbar :current="4" />
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SvgIcon from '@/components/SvgIcon.vue'
import CustomTabbar from '@/components/CustomTabbar.vue'
import { getCryptoIcon } from '@/utils/crypto.js'
import walletStore from '@/stores/walletStore.js'
import { formatBalanceWith9Digits, formatPercent } from '@/utils/format.js'


const statusBarHeight = ref(0)
const isAssetsVisible = ref(true)

// 资产数据（从 walletStore 获取）
const totalBalance = computed(() => formatBalanceWith9Digits(walletStore.state.totalBalance))
const totalUsdValue = computed(() => formatBalanceWith9Digits(walletStore.state.totalUsdValue))
const spotBalance = computed(() => formatBalanceWith9Digits(walletStore.state.spotBalance))
const contractBalance = computed(() => formatBalanceWith9Digits(walletStore.state.contractBalance))
const deliveryBalance = computed(() => formatBalanceWith9Digits(walletStore.state.deliveryBalance))
const profitData = computed(() => ({
	value: walletStore.state.todayProfit,
	rate: formatPercent(walletStore.state.todayProfitRate)
}))
const cryptoAssets = computed(() => {
	// 创建数组副本后再排序（避免修改只读数组）
	return [...walletStore.state.assets]
		.sort((a, b) => b.balance - a.balance)  // 按余额降序排序
		.map(asset => ({
			id: asset.id,
			name: asset.name,
			symbol: asset.symbol,
			balance: formatBalanceWith9Digits(asset.balance),
			usdValue: formatBalanceWith9Digits(asset.usdValue)
		}))
})
const isLoading = computed(() => walletStore.state.isLoading)

onMounted(() => {
	// #ifndef H5
	uni.hideTabBar()
	// #endif
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// walletStore 在 App.vue 中已初始化，无需手动调用
	console.log('[Assets] walletStore 已就绪，余额:', walletStore.state.totalBalance)
})

// 页面显示时不需要重复刷新，walletStore 会通过 WebSocket 自动更新
onShow(() => {
	// walletStore 已通过 WebSocket 持续接收更新
	// 如果需要强制刷新，可调用：
	if (!walletStore.state.isInitialized) {
		walletStore.fetchAssetData()
	}
	console.log('[Assets] 页面显示，余额数据来自 walletStore')
})

const toggleVisibility = () => {
	isAssetsVisible.value = !isAssetsVisible.value
}

const goToRecords = () => {
	uni.navigateTo({ url: '/pages/trade/orders' })
}

const goToProfitDetail = () => {
	uni.showToast({ title: '盈亏详情', icon: 'none' })
}

// 跳转资金划转页面
const goToTransfer = () => {
	uni.navigateTo({ url: '/pages/wallet/transfer' })
}

// 充值
const handleAddAsset = () => {
	uni.navigateTo({ url: '/pages/wallet/deposit' })
}

// 提现
const handleWithdraw = () => {
	uni.navigateTo({ url: '/pages/wallet/withdraw' })
}

// 跳转到实名认证
const goToKyc = () => {
	uni.navigateTo({
		url: '/pages/kyc/kyc'
	})
}

// 退出登录
const handleLogout = () => {
	uni.showModal({
		title: '提示',
		content: '确定要退出登录吗？',
		success: (res) => {
			if (res.confirm) {
				// 清除登录状态
				uni.removeStorageSync('isLoggedIn')
				uni.removeStorageSync('userInfo')
				uni.removeStorageSync('token')
				
				uni.showToast({
					title: '已退出登录',
					icon: 'success',
					duration: 1500
				})
				
				setTimeout(() => {
					// 跳转到登录页
					uni.redirectTo({
						url: '/pages/login/login'
					})
				}, 1500)
			}
		}
	})
}

// 打开划转页面
const openTransferPopup = () => {
	console.log('[Transfer] 跳转到划转页面')
	uni.navigateTo({ url: '/pages/wallet/transfer' })
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: linear-gradient(180deg, #fafbfc 0%, #ffffff 100%);
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

/* 顶部渐变容器 */
.header-gradient-container {
	flex-shrink: 0;
	background: linear-gradient(135deg, #0066FF 0%, #0052D9 45%, #0041A8 100%);
	position: relative;
	z-index: 999;
	
	/* 添加微妙的光泽效果 */
	&::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 160rpx;
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.15) 0%, rgba(255, 255, 255, 0) 100%);
		pointer-events: none;
	}
	
	/* 底部柔和过渡 */
	&::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 60rpx;
		background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.03) 100%);
		pointer-events: none;
	}
}

/* 自定义导航栏 */
.custom-navbar {
	position: relative;
	z-index: 10;
	
	.navbar-content {
		height: 88rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 30rpx;
	}
	
	.nav-title {
		font-size: 36rpx;
		font-weight: 600;
		color: #fff;
		text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.1);
		letter-spacing: 0.5rpx;
	}
	
	.nav-right {
		display: flex;
		align-items: center;
		gap: 24rpx;
		
		.icon-btn {
			width: 48rpx;
			height: 48rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.12);
			backdrop-filter: blur(10rpx);
			transition: all 0.3s ease;
			
			&:active {
				background: rgba(255, 255, 255, 0.25);
				transform: scale(0.95);
			}
		}
	}
}

/* 预估资产区 */
.summary-section {
	padding: 30rpx 30rpx 36rpx;
	position: relative;
	z-index: 10;
	
	.summary-header {
		display: flex;
		align-items: center;
		margin-bottom: 24rpx;
		gap: 12rpx;
		
		.summary-label {
			font-size: 28rpx;
			font-weight: 500;
			color: rgba(255, 255, 255, 0.95);
			letter-spacing: 0.5rpx;
		}
		
		.eye-icon {
			display: flex;
			align-items: center;
			padding: 4rpx;
			border-radius: 50%;
			transition: background-color 0.2s ease;
			
			&:active {
				background-color: rgba(255, 255, 255, 0.15);
			}
		}
	}
	
	.total-amount-box {
		display: flex;
		align-items: baseline;
		gap: 12rpx;
		margin-bottom: 16rpx;
		
		.amount-num {
			font-size: 60rpx;
			font-weight: 700;
			color: #fff;
			font-family: 'DIN Alternate', 'Roboto', 'Helvetica Neue', Helvetica, Arial, sans-serif;
			text-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
			letter-spacing: 1rpx;
		}
		
		.currency {
			font-size: 34rpx;
			font-weight: 600;
			color: rgba(255, 255, 255, 0.95);
			margin-right: 4rpx;
			text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.1);
		}
	}
	
	.usd-equivalent {
		margin-bottom: 32rpx;
		
		.usd-text {
			font-size: 28rpx;
			font-weight: 500;
			color: rgba(255, 255, 255, 0.85);
			letter-spacing: 0.3rpx;
		}
	}
	
	.profit-row {
		display: flex;
		align-items: center;
		gap: 10rpx;
		padding: 16rpx 20rpx;
		background: rgba(255, 255, 255, 0.08);
		border-radius: 16rpx;
		backdrop-filter: blur(10rpx);
		
		.profit-label {
			font-size: 26rpx;
			font-weight: 500;
			color: rgba(255, 255, 255, 0.95);
			letter-spacing: 0.3rpx;
		}
		
		.profit-value {
			font-size: 26rpx;
			font-weight: 600;
			letter-spacing: 0.3rpx;
			
			&.up {
				color: #7FFF7F;
				text-shadow: 0 2rpx 6rpx rgba(127, 255, 127, 0.3);
			}
			
			&.down {
				color: #FFB6C1;
				text-shadow: 0 2rpx 6rpx rgba(255, 182, 193, 0.3);
			}
		}
	}
}

.scroll-container {
	flex: 1;
	min-height: 0;
	overflow: hidden;
	background: #ffffff;
}

.content-scroll {
	height: 100%;
	background: #ffffff;
}

.divider {
	height: 20rpx;
	background: linear-gradient(180deg, #f8f9fa 0%, #ffffff 100%);
	margin: 0;
	box-shadow: 0 -4rpx 12rpx rgba(0, 0, 0, 0.02);
	flex-shrink: 0;
}

/* 操作按钮区域 - 固定 */
.action-buttons-fixed {
	display: flex;
	justify-content: space-around;
	padding: 36rpx 30rpx 24rpx;
	background: #ffffff;
	flex-shrink: 0;
	position: relative;
	z-index: 10;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.03);
	
	.action-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14rpx;
		padding: 16rpx 24rpx;
		border-radius: 16rpx;
		transition: all 0.3s ease;
		
		.btn-text {
			font-size: 26rpx;
			font-weight: 500;
			color: #333333;
			letter-spacing: 0.3rpx;
		}
		
		&:active {
			background-color: #f5f7fa;
			transform: translateY(2rpx);
		}
	}
}

/* 区块标题 */
.section-header {
	padding: 24rpx 30rpx 16rpx;
	
	.section-title {
		font-size: 30rpx;
		font-weight: 600;
		color: #1a1a1a;
		letter-spacing: 0.5rpx;
	}
}

/* 区块标题 - 固定 */
.section-header-fixed {
	padding: 24rpx 30rpx 16rpx;
	background: #ffffff;
	flex-shrink: 0;
	position: relative;
	z-index: 9;
	border-bottom: 1rpx solid #f0f0f0;
	
	.section-title {
		font-size: 30rpx;
		font-weight: 600;
		color: #1a1a1a;
		letter-spacing: 0.5rpx;
	}
}

/* 加密货币列表 */
.crypto-list {
	padding: 20rpx 30rpx 30rpx;
	background: #ffffff;
	
	.crypto-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 28rpx 0;
		border-bottom: 1rpx solid #f0f0f0;
		transition: background-color 0.2s ease;
		
		&:active {
			background-color: #f8f8f8;
		}
		
		.crypto-left {
			display: flex;
			align-items: center;
			gap: 24rpx;
			flex: 1;
			
			.crypto-icon {
				width: 80rpx;
				height: 80rpx;
				border-radius: 50%;
				background-color: #f5f5f5;
				padding: 4rpx;
			}
			
			.crypto-info {
				display: flex;
				flex-direction: column;
				gap: 10rpx;
				
				.crypto-name {
					font-size: 30rpx;
					font-weight: 500;
					color: #1a1a1a;
					line-height: 1.2;
				}
				
				.crypto-symbol {
					font-size: 24rpx;
					color: #999999;
					font-weight: 400;
					letter-spacing: 0.5rpx;
				}
			}
		}
		
		.crypto-right {
			display: flex;
			flex-direction: column;
			align-items: flex-end;
			gap: 8rpx;
			min-width: 200rpx;
			
			.amount-line {
				.crypto-amount {
					font-size: 32rpx;
					font-weight: 600;
					color: #1a1a1a;
					font-family: 'DIN Alternate', 'Roboto', 'Helvetica Neue', Helvetica, Arial, sans-serif;
					letter-spacing: 0.5rpx;
					line-height: 1.2;
				}
			}
			
			.crypto-usd {
				font-size: 24rpx;
				color: #8c8c8c;
				font-weight: 400;
				letter-spacing: 0.3rpx;
			}
		}
		
		&:last-child {
			border-bottom: none;
		}
	}
}

.tabbar-placeholder {
	height: 120rpx;
	padding-bottom: env(safe-area-inset-bottom);
	flex-shrink: 0;
}

/* 现货和合约账户余额 */
.wallet-balance-section {
	padding: 0 30rpx 30rpx;
	position: relative;
	z-index: 10;
}

.wallet-balance-card {
	background: rgba(255, 255, 255, 0.15);
	border-radius: 20rpx;
	padding: 24rpx;
	backdrop-filter: blur(10rpx);
}

.wallet-balance-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 20rpx;
	padding-bottom: 16rpx;
	border-bottom: 1rpx solid rgba(255, 255, 255, 0.1);
}

.wallet-balance-title {
	font-size: 28rpx;
	font-weight: 500;
	color: rgba(255, 255, 255, 0.9);
}

.transfer-entry-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	padding: 10rpx 20rpx;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 30rpx;
	transition: all 0.2s ease;
	
	&:active {
		background: rgba(255, 255, 255, 0.3);
		transform: scale(0.95);
	}
}

.transfer-entry-text {
	font-size: 26rpx;
	font-weight: 500;
	color: #fff;
}

.wallet-balance-row {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.wallet-balance-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx 16rpx;
	background: rgba(255, 255, 255, 0.08);
	border-radius: 12rpx;
	transition: background-color 0.2s ease;
	
	&:active {
		background: rgba(255, 255, 255, 0.15);
	}
}

.wallet-balance-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.wallet-icon-small {
	width: 48rpx;
	height: 48rpx;
	border-radius: 12rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.spot-icon {
	background: linear-gradient(135deg, #007AFF 0%, #00C6FF 100%);
}

.contract-icon {
	background: linear-gradient(135deg, #FF9500 0%, #FFB800 100%);
}

.wallet-balance-name {
	font-size: 28rpx;
	font-weight: 500;
	color: rgba(255, 255, 255, 0.95);
}

.wallet-balance-right {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.wallet-balance-amount {
	font-size: 28rpx;
	font-weight: 600;
	color: #fff;
	font-family: 'DIN Alternate', 'Roboto', 'Helvetica Neue', Helvetica, Arial, sans-serif;
}


</style>
