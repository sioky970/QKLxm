<template>
	<view class="page">
		<!-- 状态栏占位 -->
		<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
		
		<!-- 市场表头 -->
		<view class="market-header">
			<view class="header-item" @click="handleSort('symbol')">
				<text>交易对</text>
				<view class="sort-icons">
					<view class="sort-up" :class="{ active: sortField === 'symbol' && sortOrder === 'asc' }"></view>
					<view class="sort-down" :class="{ active: sortField === 'symbol' && sortOrder === 'desc' }"></view>
				</view>
			</view>
			<view class="header-item" @click="handleSort('price')">
				<text>最新价</text>
				<view class="sort-icons">
					<view class="sort-up" :class="{ active: sortField === 'price' && sortOrder === 'asc' }"></view>
					<view class="sort-down" :class="{ active: sortField === 'price' && sortOrder === 'desc' }"></view>
				</view>
			</view>
			<view class="header-item" @click="handleSort('change')">
				<text>24h涨跌幅</text>
				<view class="sort-icons">
					<view class="sort-up" :class="{ active: sortField === 'change' && sortOrder === 'asc' }"></view>
					<view class="sort-down" :class="{ active: sortField === 'change' && sortOrder === 'desc' }"></view>
				</view>
			</view>
		</view>

		<view class="scroll-container">
			<scroll-view scroll-y class="market-list">
				<view class="market-item" v-for="(item, index) in sortedMarketList" :key="index" @click="goToTrade(item)">
					<view class="item-left">
						<image class="coin-icon" :src="getCryptoIcon(item.name)" mode="aspectFit"></image>
						<view class="coin-info">
							<text class="coin-symbol">{{ item.symbol.split('/')[0] }}</text>
							<text class="coin-volume">Volume {{ item.volume }}</text>
						</view>
					</view>
					
					<view class="item-center">
						<text class="coin-price">{{ item.price }}</text>
						<text class="coin-usd">≈ ${{ item.usd }}</text>
					</view>
					
					<view class="item-right">
						<view class="change-tag" :class="item.changeClass">
							<text class="change-text">{{ item.change }}</text>
						</view>
					</view>
				</view>
				
				<!-- 内部占位符，防止内容被底部栏遮挡 -->
				<view class="tabbar-placeholder"></view>
			</scroll-view>
		</view>
		
		<!-- 自定义底部栏 -->
		<CustomTabbar :current="1" />
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { onShow, onHide } from '@dcloudio/uni-app'
import CustomTabbar from '@/components/CustomTabbar.vue'
import { getCryptoIcon } from '@/utils/crypto.js'
import marketStore from '@/stores/marketStore.js'
import { formatCurrency, formatNumber } from '@/utils/format.js'

const statusBarHeight = ref(0)
const sortField = ref('volume') // 默认按成交量排序
const sortOrder = ref('desc')

// 从全局marketStore获取市场数据
const marketList = computed(() => {
	// 使用响应式状态
	const mapData = marketStore.state.marketDataMap
	const allData = Array.from(mapData.values())
	
	if (allData.length === 0) {
		return []
	}
	
	// 格式化显示数据
	return allData.map(data => {
		const priceNum = data.price || 0
		const changeNum = data.change || 0
		const volumeNum = data.volume || 0
		const name = data.name || ''
		
		return {
			symbol: `${name}/USDT`,
			name: name,
			currency_id: data.currency_id,
			legal_id: data.legal_id,
			price: formatCurrency(priceNum, 2),
			priceNum: priceNum,
			usd: formatCurrency(priceNum, 2),
			volume: formatVolume(volumeNum),
			volumeNum: volumeNum,
			change: changeNum >= 0 ? `+${formatNumber(changeNum, 2)}%` : `${formatNumber(changeNum, 2)}%`,
			changeNum: changeNum,
			changeClass: changeNum >= 0 ? 'up' : 'down'
		}
	})
})

// 格式化交易量
const formatVolume = (volume) => {
	if (!volume) return '0'
	if (volume >= 1000000000) {
		return (volume / 1000000000).toFixed(2) + 'B'
	} else if (volume >= 1000000) {
		return (volume / 1000000).toFixed(2) + 'M'
	} else if (volume >= 1000) {
		return (volume / 1000).toFixed(2) + 'K'
	}
	return volume.toFixed(2)
}

// 排序后的列表
const sortedMarketList = computed(() => {
	if (!sortField.value) return marketList.value
	
	const list = [...marketList.value]
	list.sort((a, b) => {
		let valA, valB
		
		if (sortField.value === 'symbol') {
			valA = a.symbol
			valB = b.symbol
		} else if (sortField.value === 'price') {
			valA = a.priceNum
			valB = b.priceNum
		} else if (sortField.value === 'change') {
			valA = a.changeNum
			valB = b.changeNum
		} else if (sortField.value === 'volume') {
			valA = a.volumeNum
			valB = b.volumeNum
		}
		
		if (valA < valB) return sortOrder.value === 'asc' ? -1 : 1
		if (valA > valB) return sortOrder.value === 'asc' ? 1 : -1
		return 0
	})
	return list
})

onMounted(() => {
	uni.hideTabBar()
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 确保 marketStore 已初始化
	if (!marketStore.isReady.value) {
		marketStore.initialize()
	}
})

// 页面显示时刷新数据
onShow(() => {
	marketStore.fetchMarketData()
})

// 页面隐藏时不需要任何操作
onHide(() => {
	// 保持使用全局marketStore
})

// 页面卸载时不需要任何操作
onUnmounted(() => {
	// 不再需要取消订阅，使用全局marketStore
	// WebSocket连接由App.vue统一管理
})

const handleSort = (field) => {
	if (sortField.value === field) {
		sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
	} else {
		sortField.value = field
		sortOrder.value = 'desc' // 默认降序
	}
}

const goToTrade = (item) => {
	// 保存币种完整信息到storage，供交易页面使用
	const tradeSymbol = {
		symbol: item.symbol,
		name: item.name,
		price: item.price,
		change: item.change,
		changeClass: item.changeClass,
		currency_id: item.currency_id,
		legal_id: item.legal_id
	}
	
	console.log('[Market] 跳转到交易页面，币种信息:', tradeSymbol)
	uni.setStorageSync('currentTradeSymbol', JSON.stringify(tradeSymbol))
	
	uni.switchTab({
		url: '/pages/trade/trade'
	})
}
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
	background: #fff;
}

.market-header {
	height: 88rpx;
	flex-shrink: 0;
	display: flex;
	padding: 0 30rpx;
	border-bottom: 1rpx solid #f0f0f0;
	background: #fff;
	
	.header-item {
		display: flex;
		align-items: center;
		font-size: 24rpx;
		color: #999;
		
		&:nth-child(1) { flex: 2.5; }
		&:nth-child(2) { flex: 2; justify-content: flex-start; padding-left: 20rpx; }
		&:nth-child(3) { flex: 1.5; justify-content: flex-end; }
		
		.sort-icons {
			margin-left: 8rpx;
			display: flex;
			flex-direction: column;
			gap: 4rpx;
			
			.sort-up, .sort-down {
				width: 0;
				height: 0;
				border-left: 6rpx solid transparent;
				border-right: 6rpx solid transparent;
			}
			.sort-up {
				border-bottom: 8rpx solid #ccc;
				&.active { border-bottom-color: #3B82F6; }
			}
			.sort-down {
				border-top: 8rpx solid #ccc;
				&.active { border-top-color: #3B82F6; }
			}
		}
	}
}

.scroll-container {
	flex: 1;
	min-height: 0;
	overflow: hidden;
}

.market-list {
	height: 100%;
	background: #fff;
	
	.market-item {
		display: flex;
		align-items: center;
		padding: 30rpx;
		border-bottom: 1rpx solid #f8f8f8;
		
		&:active {
			background-color: #f9f9f9;
		}
		
		.item-left {
			flex: 2.5;
			display: flex;
			align-items: center;
			
			.coin-icon {
				width: 64rpx;
				height: 64rpx;
				margin-right: 20rpx;
			}
			
			.coin-info {
				display: flex;
				flex-direction: column;
				
				.coin-symbol {
					font-size: 30rpx;
					font-weight: bold;
					color: #333;
					margin-bottom: 4rpx;
				}
				
				.coin-volume {
					font-size: 24rpx;
					color: #999;
				}
			}
		}
		
		.item-center {
			flex: 2;
			display: flex;
			flex-direction: column;
			align-items: flex-start;
			padding-left: 20rpx;
			
			.coin-price {
				font-size: 28rpx;
				font-weight: bold;
				color: #333;
				margin-bottom: 4rpx;
			}
			
			.coin-usd {
				font-size: 22rpx;
				color: #999;
			}
		}
		
		.item-right {
			flex: 1.5;
			display: flex;
			justify-content: flex-end;
			
			.change-tag {
				width: 140rpx;
				height: 60rpx;
				border-radius: 8rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				
				&.up { background-color: #00c087; }
				&.down { background-color: #ff4444; }
				
				.change-text {
					color: #fff;
					font-size: 24rpx;
					font-weight: bold;
				}
			}
		}
	}
	
	.bottom-padding {
		height: 40rpx;
	}
}

.tabbar-placeholder {
	height: 120rpx;
	padding-bottom: env(safe-area-inset-bottom);
	flex-shrink: 0;
}
</style>
