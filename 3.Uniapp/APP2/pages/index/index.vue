<template>
	<view class="page">
		<!-- 已登录显示首页内容 -->
		<template v-if="isLoggedIn">
			<!-- 自定义顶部导航栏 -->
			<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<!-- 品牌Logo -->
				<view class="navbar-left">
					<image class="logo" src="/static/logocn.png" mode="aspectFit"></image>
				</view>
			</view>
		</view>

		<!-- 滚动区域容器 -->
		<view class="scroll-container">
			<scroll-view scroll-y class="content-scroll">
				<!-- 资产显示区 -->
				<view class="asset-section">
					<view class="asset-info">
						<view class="asset-title">
							<text class="title-text">预估总资产(USDT)</text>
							<SvgIcon name="eye-open" :size="14" color="#999" />
						</view>
						<view class="asset-balance">
							<text class="balance-num">{{ assetData.totalBalance }}</text>
							<!-- 七日盈亏折线图 -->
							<view class="asset-chart">
								<svg class="sparkline" width="80" height="30" viewBox="0 0 80 30">
									<path 
										:d="sparklinePath" 
										fill="none" 
										stroke="#00C087" 
										stroke-width="2" 
										stroke-linecap="round" 
										stroke-linejoin="round"
									/>
								</svg>
							</view>
						</view>
						<view class="asset-usd">
							<text class="usd-text">≈${{ assetData.totalBalance }}</text>
						</view>
						<view class="asset-profit">
							<text class="profit-label">今日盈亏</text>
							<text class="profit-value" :class="assetData.profitClass">{{ assetData.todayProfit }} USDT ({{ assetData.profitRate }})</text>
							<SvgIcon name="arrow-down" :size="12" :color="assetData.profitColor" />
						</view>
					</view>
				</view>

				<!-- 活动提示 -->
				<view class="promotion-bar" @click="openNewsPopup">
					<SvgIcon name="gift" :size="20" color="#FF6B00" />
					<text class="promotion-text">Promotion in progress</text>
					<view class="promotion-more">
						<SvgIcon name="menu" :size="20" color="#999" />
					</view>
				</view>

				<!-- 快捷功能区 -->
				<view class="quick-actions">
					<view class="action-row">
						<view class="action-item" v-for="item in quickActions" :key="item.id" @click="handleAction(item.id)">
							<view class="action-icon" :class="item.hot ? 'action-hot' : item.isNew ? 'action-new' : ''">
								<image class="action-img" :src="item.icon" mode="aspectFit"></image>
								<text v-if="item.hot" class="badge hot-badge">HOT</text>
								<text v-if="item.isNew" class="badge new-badge">NEW</text>
							</view>
							<text class="action-label">{{ item.label }}</text>
							<text class="action-sublabel" v-if="item.sublabel">{{ item.sublabel }}</text>
						</view>
					</view>
				</view>

				<!-- 市场列表 -->
				<view class="market-list" @click="goToMarket">
					<!-- 表头 -->
					<view class="market-header-inner">
						<text class="header-item">交易对</text>
						<text class="header-item">最新价</text>
						<text class="header-item">24h涨跌幅</text>
					</view>
					
					<!-- 列表项 -->
					<view class="market-item" v-for="item in marketList" :key="item.symbol" @click.stop="goToTrade(item)">
						<view class="market-left">
							<image class="coin-icon" :src="getCryptoIcon(item.name)" mode="aspectFit"></image>
							<view class="coin-info">
								<text class="coin-symbol">{{ item.name }}</text>
								<text class="coin-volume">Volume {{ item.volume }}</text>
							</view>
						</view>
						<view class="market-center">
							<text class="coin-price">{{ item.price }}</text>
							<text class="coin-price-usd">≈ ${{ item.price }}</text>
						</view>
						<view class="market-right">
							<view class="change-tag" :class="item.changeClass">
								<text class="change-text">{{ item.change }}</text>
							</view>
						</view>
					</view>
				</view>
			</scroll-view>
		</view>
		
		<!-- 底部占位符 -->
		<view class="tabbar-placeholder"></view>
		
		<!-- 自定义底部栏 -->
		<CustomTabbar :current="0" />

		<!-- 新闻弹窗 -->
		<view class="popup-mask" v-if="showNewsPopup" @click="closeNewsPopup" @touchmove.stop.prevent>
			<view class="news-popup" :class="{ 'popup-show': showNewsAnimate }" @click.stop>
				<view class="popup-header">
					<view class="back-btn" v-if="showNewsDetail" @click="backToNewsList">
						<SvgIcon name="back" :size="20" color="#333" />
					</view>
					<text class="popup-title">{{ showNewsDetail ? '公告详情' : '最新公告' }}</text>
					<view class="close-btn" @click="closeNewsPopup">
						<SvgIcon name="menu" :size="20" color="#999" />
					</view>
				</view>
				
				<!-- 新闻列表 -->
				<scroll-view scroll-y class="news-list-scroll" v-if="!showNewsDetail">
					<view class="news-item" v-for="(news, index) in newsList" :key="index" @click="viewNewsDetail(news)">
						<view class="news-top">
							<text class="news-tag" :class="news.type">{{ news.tag }}</text>
						</view>
						<text class="news-title">{{ news.title }}</text>
						<text class="news-content line-clamp">{{ news.content }}</text>
						<view class="news-footer">
							<text class="news-time">{{ news.time }}</text>
							<view class="read-more-box">
								<text class="read-more">阅读全文</text>
								<SvgIcon name="arrow-right" :size="12" color="#3B82F6" />
							</view>
						</view>
					</view>
				</scroll-view>

				<!-- 新闻详情 -->
				<scroll-view scroll-y class="news-detail-scroll" v-else>
					<view class="detail-content">
						<view class="news-top">
							<text class="news-tag" :class="selectedNews.type">{{ selectedNews.tag }}</text>
						</view>
						<text class="detail-title">{{ selectedNews.title }}</text>
						<text class="detail-time">{{ selectedNews.time }}</text>
						<view class="detail-text">
							<text class="text-p">{{ selectedNews.content }}</text>
							<text class="text-p">Coinbase 始终致力于为您提供安全、透明的交易环境。如有任何疑问，请随时联系我们的 24/7 在线客服。</text>
							<text class="text-p">感谢您对我们的支持与信任！</text>
						</view>
					</view>
				</scroll-view>
			</view>
		</view>
		</template>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { onShow, onHide } from '@dcloudio/uni-app'
import SvgIcon from '@/components/SvgIcon.vue'
import CustomTabbar from '@/components/CustomTabbar.vue'
import { getCryptoIcon } from '@/utils/crypto.js'
import { getSevenDaysProfitLoss } from '@/utils/api.js'
import { formatBalanceWith9Digits, formatPriceWithPrecision } from '@/utils/format.js'
import marketStore from '@/stores/marketStore.js'
import walletStore from '@/stores/walletStore.js'

// 格式化数字，使用9位有效数字限制（专用于资产显示）
const formatNumber = (num) => {
  return formatBalanceWith9Digits(num)
}

// 登录状态
const isLoggedIn = ref(false)

// 状态栏高度
const statusBarHeight = ref(0)
const showNewsPopup = ref(false)
const showNewsAnimate = ref(false)
const showNewsDetail = ref(false)
const selectedNews = ref({})

const newsList = ref([
	{
		tag: '重要',
		type: 'important',
		time: '2026-01-25',
		title: '关于 Coinbase 系统升级维护公告',
		content: '为了提供更优质的交易体验，我们将于 2026年1月26日 02:00 (UTC) 进行系统升级，预计持续 2 小时。'
	},
	{
		tag: '新币',
		type: 'new',
		time: '2026-01-24',
		title: 'Coinbase 将上线 Base 生态新资产',
		content: '我们很高兴地宣布，Coinbase 将于近期上线多个 Base 生态优质项目，支持 USDT 交易对。'
	},
	{
		tag: '活动',
		type: 'activity',
		time: '2026-01-23',
		title: '邀请好友赚取加密奖励计划',
		content: '现在邀请您的好友加入 Coinbase，双方均可获得价值 10 USDT 的比特币奖励。'
	}
])

const openNewsPopup = () => {
	showNewsPopup.value = true
	setTimeout(() => {
		showNewsAnimate.value = true
	}, 10)
}

const closeNewsPopup = () => {
	showNewsAnimate.value = false
	setTimeout(() => {
		showNewsPopup.value = false
		showNewsDetail.value = false
		selectedNews.value = {}
	}, 300)
}

const viewNewsDetail = (news) => {
	selectedNews.value = news
	showNewsDetail.value = true
}

const backToNewsList = () => {
	showNewsDetail.value = false
	selectedNews.value = {}
}

// 获取系统信息
onMounted(() => {
	uni.hideTabBar()
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 检查登录状态
	checkLoginStatus()
	
	// 不再需要初始化加载资产数据，walletStore 在 App.vue 中已初始化
	if (isLoggedIn.value) {
		// 获取七日盈亏数据（仅用于折线图）
		fetchSevenDaysData()
	}
})

// 页面显示时检查登录状态，不再重复请求资产数据
onShow(() => {
	checkLoginStatus()
	// walletStore 在 App.vue 中已初始化，WebSocket 会持续推送数据
	// 如果需要强制刷新，可调用：
	if (isLoggedIn.value && !walletStore.state.isInitialized) {
		walletStore.fetchAssetData()
	}
	
	// 获取七日盈亏数据（仅用于折线图）
	if (isLoggedIn.value) {
		fetchSevenDaysData()
	}
})

// 页面隐藏时不做特殊处理，保持WebSocket连接
onHide(() => {
	// 保持WebSocket连接以便后台接收数据
})

// 页面卸载时不需要任何操作
onUnmounted(() => {
	// 不再需要取消订阅，使用全局marketStore
	// WebSocket连接由App.vue统一管理
})

// 检查登录状态
const checkLoginStatus = () => {
	const logged = uni.getStorageSync('isLoggedIn')
	isLoggedIn.value = !!logged
	
	if (!logged) {
		// 未登录，跳转到登录页
		uni.redirectTo({
			url: '/pages/login/login'
		})
	}
}

// 资产数据（从wallet store获取）
const assetData = computed(() => {
	const totalBalance = formatNumber(walletStore.state.totalBalance)
	const profit = walletStore.state.todayProfit
	const profitRate = walletStore.state.todayProfitRate
	
	return {
		totalBalance: totalBalance,
		totalUsdValue: totalBalance,
		todayProfit: profit >= 0 ? `+${formatNumber(profit)}` : formatNumber(profit),
		profitRate: profitRate,
		profitClass: profit >= 0 ? 'up' : 'down',
		profitColor: profit >= 0 ? '#00C087' : '#FF4444'
	}
})

// 加载状态（保留以便展示loading）
const isLoadingAssets = computed(() => walletStore.state.isLoading)

// 七日数据点（用于折线图）
const sparklinePoints = ref([25, 22, 26, 15, 18, 12, 10])
let lastSevenDaysFetchTime = 0 // 用于防止重复请求

// 获取七日盈亏数据
const fetchSevenDaysData = async (force = false) => {
	// 防抖：30秒内不重复请求（七日数据变化较慢）
	const now = Date.now()
	if (!force && lastSevenDaysFetchTime && (now - lastSevenDaysFetchTime < 30000)) {
		return
	}
	lastSevenDaysFetchTime = now
	
	try {
		const res = await getSevenDaysProfitLoss()
		const data = res.data
		
		if (data.profit_loss_data && data.profit_loss_data.length > 0) {
			// 将盈亏数据转换为图表坐标（0-30范围）
			const values = data.profit_loss_data.map(item => item.net_profit_loss)
			const max = Math.max(...values)
			const min = Math.min(...values)
			const range = max - min || 1
			
			sparklinePoints.value = values.map(val => {
				return 30 - ((val - min) / range) * 30
			})
		}
	} catch (err) {
		console.error('获取七日数据失败:', err)
	}
}

// 计算平滑曲线路径 (使用贝塞尔曲线)
const sparklinePath = computed(() => {
	const points = sparklinePoints.value
	const width = 80
	const height = 30
	const step = width / (points.length - 1)
	
	let path = `M 0 ${points[0]}`
	
	for (let i = 0; i < points.length - 1; i++) {
		const x1 = i * step
		const y1 = points[i]
		const x2 = (i + 1) * step
		const y2 = points[i+1]
		
		// 控制点使得曲线平滑
		const cx = (x1 + x2) / 2
		path += ` C ${cx} ${y1}, ${cx} ${y2}, ${x2} ${y2}`
	}
	return path
})

// 快捷功能数据
const quickActions = ref([
	{ id: 'deposit', label: '添加资金', sublabel: '', icon: '/static/crypto-icons/flash-swap.svg', hot: true },
	{ id: 'kyc', label: '实名认证', sublabel: '', icon: '/static/crypto-icons/kyc.svg' },
	{ id: 'contract', label: '合约交易', sublabel: '', icon: '/static/crypto-icons/contract.svg' },
	{ id: 'about', label: '关于我们', sublabel: '', icon: '/static/crypto-icons/about.svg' }
])

// 当前选中的标签
const currentTab = ref('up')

// 缓存随机选择的5个币种ID，避免每次computed都重新随机
const randomSymbolIds = ref([])

// 市场数据（从全局marketStore获取）
const marketList = computed(() => {
	// 使用响应式状态
	const mapData = marketStore.state.marketDataMap
	const allData = Array.from(mapData.values())
	
	if (allData.length === 0) {
		return []
	}
	
	// 如果还没有随机选择过，或者数据长度变化了，则重新随机选择
	if (randomSymbolIds.value.length === 0 || randomSymbolIds.value.length > allData.length) {
		// 随机打乱并取前5个的ID
		const shuffled = [...allData].sort(() => Math.random() - 0.5)
		const randomFive = shuffled.slice(0, 5)
		randomSymbolIds.value = randomFive.map(d => `${d.currency_id}_${d.legal_id}`)
	}
	
	// 根据缓存的ID获取对应的数据
	const selectedData = []
	for (const id of randomSymbolIds.value) {
		const data = allData.find(d => `${d.currency_id}_${d.legal_id}` === id)
		if (data) {
			selectedData.push(data)
		}
	}
	
	// 如果选中的数据不足5个，补充新的
	if (selectedData.length < 5 && allData.length >= 5) {
		const shuffled = [...allData].sort(() => Math.random() - 0.5)
		const randomFive = shuffled.slice(0, 5)
		randomSymbolIds.value = randomFive.map(d => `${d.currency_id}_${d.legal_id}`)
		return randomFive.map(data => {
			const changeNum = data.change || 0
			const name = data.name || ''
			// 使用动态精度（从WebSocket推送的decimal_scale）
			const precision = data.decimal_scale || 4
			return {
				symbol: name,
				name: name,
				price: formatPriceWithPrecision(data.price || 0, precision),
				volume: formatVolume(data.volume),
				change: changeNum >= 0 ? `+${changeNum.toFixed(2)}%` : `${changeNum.toFixed(2)}%`,
				changeClass: changeNum >= 0 ? 'up' : 'down',
				currency_id: data.currency_id,
				legal_id: data.legal_id
			}
		})
	}
	
	// 格式化数据
	const formatted = selectedData.map(data => {
		const changeNum = data.change || 0
		const name = data.name || ''
		// 使用动态精度（从WebSocket推送的decimal_scale）
		const precision = data.decimal_scale || 4
		return {
			symbol: name,
			name: name,
			price: formatPriceWithPrecision(data.price || 0, precision),
			volume: formatVolume(data.volume),
			change: changeNum >= 0 ? `+${changeNum.toFixed(2)}%` : `${changeNum.toFixed(2)}%`,
			changeClass: changeNum >= 0 ? 'up' : 'down',
			currency_id: data.currency_id,
			legal_id: data.legal_id
		}
	})
	
	return formatted
})

// 格式化成交量
const formatVolume = (volume) => {
	if (!volume) return '0'
	if (volume >= 1000000000) {
		return (volume / 1000000000).toFixed(2) + 'B'
	} else if (volume >= 1000000) {
		return (volume / 1000000).toFixed(2) + 'M'
	} else if (volume >= 1000) {
		return (volume / 1000).toFixed(2) + 'K'
	}
	return volume.toString()
}

// 切换标签
const switchTab = (tab) => {
	currentTab.value = tab
}

// 快捷操作
const handleAction = (id) => {
	switch (id) {
		case 'deposit':
			uni.navigateTo({ url: '/pages/wallet/deposit' })
			break
		case 'contract':
			uni.switchTab({ url: '/pages/contract/contract' })
			break
		case 'kyc':
			uni.navigateTo({ url: '/pages/kyc/kyc' })
			break
		case 'position':
			uni.switchTab({ url: '/pages/trade/trade' })
			break
		case 'record':
			uni.navigateTo({ url: '/pages/trade/orders' })
			break
		case 'market':
			uni.switchTab({ url: '/pages/market/market' })
			break
		case 'about':
			uni.navigateTo({ url: '/pages/about/about' })
			break
		default:
			uni.showToast({ title: `点击了${id}`, icon: 'none' })
	}
}

// 跳转市场页
const goToMarket = () => {
	uni.switchTab({ url: '/pages/market/market' })
}

// 跳转交易页面
const goToTrade = (item) => {
	uni.setStorageSync('currentSymbol', item.name)
	uni.switchTab({
		url: '/pages/trade/trade'
	})
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #F5F5F5;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

/* 自定义导航栏 */
.custom-navbar {
	background: #FFFFFF;
	flex-shrink: 0;
	z-index: 999;
	box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
	
	.navbar-content {
		height: 88rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 30rpx;
	}
	
	.navbar-left {
		display: flex;
		align-items: center;
		
		.logo {
			height: 56rpx;
			max-width: 280rpx;
		}
	}
	
	.navbar-right {
		display: flex;
		gap: 24rpx;
		
		.icon-btn {
			width: 44rpx;
			height: 44rpx;
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}
}

/* 内容区域 */
.scroll-container {
	flex: 1;
	min-height: 0;
	overflow: hidden;
}

.content-scroll {
	height: 100%;
	padding-bottom: 20rpx;
}

/* 资产显示区 */
.asset-section {
	margin: 24rpx 30rpx;
	padding: 30rpx 0;
	background: #FFFFFF;
	border-radius: 16rpx;
	
	.asset-info {
		padding: 0 30rpx;
		
		.asset-title {
			display: flex;
			align-items: center;
			gap: 12rpx;
			margin-bottom: 20rpx;
			
			.title-text {
				font-size: 28rpx;
				color: #333;
			}
		}
		
		.asset-balance {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 12rpx;
			
			.balance-num {
				font-size: 64rpx;
				font-weight: bold;
				color: #000;
				line-height: 1.2;
			}
			
			.asset-chart {
				width: 160rpx;
				height: 60rpx;
				display: flex;
				align-items: center;
				justify-content: flex-end;
				
				.sparkline {
					width: 100%;
					height: 100%;
				}
			}
		}
		
		.asset-usd {
			margin-bottom: 24rpx;
			
			.usd-text {
				font-size: 28rpx;
				color: #999;
			}
		}
		
		.asset-profit {
			display: flex;
			align-items: center;
			gap: 8rpx;
			
			.profit-label {
				font-size: 26rpx;
				color: #999;
				border-bottom: 1rpx dashed #ccc;
			}
			
			.profit-value {
				font-size: 26rpx;
				font-weight: 500;
				
				&.up {
					color: #00C087;
				}
				
				&.down {
					color: #FF4444;
				}
			}
		}
	}
}

/* 活动提示条 */
.promotion-bar {
	margin: 0 30rpx 24rpx;
	background: #FFFFFF;
	border-radius: 16rpx;
	padding: 24rpx 30rpx;
	display: flex;
	align-items: center;
	
	.promotion-text {
		flex: 1;
		margin-left: 16rpx;
		font-size: 28rpx;
		color: #333;
	}
	
	.promotion-more {
		width: 40rpx;
		height: 40rpx;
		display: flex;
		align-items: center;
		justify-content: center;
	}
}

/* 快捷功能 */
.quick-actions {
	margin: 0 30rpx 24rpx;
	background: #FFFFFF;
	border-radius: 16rpx;
	padding: 24rpx 0;
	
	.action-row {
		display: flex;
		justify-content: space-around;
	}
	
	.action-item {
		width: 25%;
		display: flex;
		flex-direction: column;
		align-items: center;
		
		.action-icon {
			width: 80rpx;
			height: 80rpx;
			margin-bottom: 10rpx;
			position: relative;
			
			.action-img {
				width: 100%;
				height: 100%;
			}
			
			.badge {
				position: absolute;
				top: -6rpx;
				right: -10rpx;
				font-size: 16rpx;
				color: #FFFFFF;
				padding: 2rpx 10rpx;
				border-radius: 12rpx;
				font-weight: bold;
			}
			
			.hot-badge {
				background: #FF4444;
			}
			
			.new-badge {
				background: #3B82F6;
			}
		}
		
		.action-label {
			font-size: 24rpx;
			color: #333;
			margin-bottom: 2rpx;
			text-align: center;
		}
		
		.action-sublabel {
			font-size: 20rpx;
			color: #999;
			text-align: center;
		}
	}
}

/* 市场列表 */
.market-list {
	margin: 0 30rpx;
	background: #FFFFFF;
	border-radius: 16rpx;
	padding: 0 30rpx;
	
	.market-header-inner {
		display: flex;
		justify-content: space-between;
		padding: 24rpx 0;
		border-bottom: 1rpx solid #F0F0F0;
		
		.header-item {
			font-size: 24rpx;
			color: #999;
			
			&:nth-child(1) {
				flex: 2.5;
			}
			
			&:nth-child(2) {
				flex: 2;
				text-align: left;
				padding-left: 20rpx;
			}
			
			&:nth-child(3) {
				flex: 1.5;
				text-align: right;
			}
		}
	}
	
	.market-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 28rpx 0;
		border-bottom: 1rpx solid #F0F0F0;
		
		&:last-child {
			border-bottom: none;
		}
		
		.market-left {
			flex: 2.5;
			display: flex;
			align-items: center;
			
			.coin-icon {
				width: 64rpx;
				height: 64rpx;
				border-radius: 50%;
				margin-right: 20rpx;
			}
			
			.coin-info {
				display: flex;
				flex-direction: column;
				
				.coin-symbol {
					font-size: 28rpx;
					font-weight: bold;
					color: #333;
					margin-bottom: 8rpx;
				}
				
				.coin-volume {
					font-size: 22rpx;
					color: #999;
				}
			}
		}
		
		.market-center {
			flex: 2;
			display: flex;
			flex-direction: column;
			align-items: flex-start;
			padding-left: 20rpx;
			
			.coin-price {
				font-size: 30rpx;
				font-weight: bold;
				color: #333;
				margin-bottom: 8rpx;
			}
			
			.coin-price-usd {
				font-size: 22rpx;
				color: #999;
			}
		}
		
		.market-right {
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
				
				&.up {
					background-color: #00c087;
				}
				
				&.down {
					background-color: #ff4444;
				}
				
				.change-text {
					color: #fff;
					font-size: 24rpx;
					font-weight: bold;
				}
			}
		}
	}
}

.tabbar-placeholder {
	height: 120rpx;
	padding-bottom: env(safe-area-inset-bottom);
	flex-shrink: 0;
}

/* 弹窗样式 */
.popup-mask {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.6);
	z-index: 2000;
	display: flex;
	align-items: flex-end;
}

.news-popup {
	width: 100%;
	background: #FFFFFF;
	border-radius: 32rpx 32rpx 0 0;
	padding: 40rpx 0;
	max-height: 70vh;
	transform: translateY(100%);
	transition: transform 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
	display: flex;
	flex-direction: column;
	
	&.popup-show {
		transform: translateY(0);
	}
	
	.popup-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 40rpx 30rpx;
		border-bottom: 1rpx solid #F0F0F0;
		
		.popup-title {
			font-size: 34rpx;
			font-weight: bold;
			color: #333;
		}
		
		.close-btn {
			padding: 10rpx;
		}
	}
}

.news-list-scroll {
	flex: 1;
	overflow: hidden;
	padding: 0 40rpx;
}

/* 公共新闻顶部样式（用于列表和详情） */
.news-top {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 16rpx;
	
	.news-tag {
		font-size: 20rpx;
		padding: 4rpx 12rpx;
		border-radius: 4rpx;
		font-weight: bold;
		
		&.important {
			background: #FFF1F0;
			color: #FF4D4F;
		}
		
		&.new {
			background: #E6F7FF;
			color: #1890FF;
		}
		
		&.activity {
			background: #F6FFED;
			color: #52C41A;
		}
	}
	
	.news-time {
		font-size: 24rpx;
		color: #999;
	}
}

.news-item {
	padding: 30rpx 0;
	border-bottom: 1rpx solid #F5F5F5;
	
	&:last-child {
		border-bottom: none;
	}
	
	.news-title {
		font-size: 30rpx;
		font-weight: bold;
		color: #333;
		margin-bottom: 12rpx;
		display: block;
	}
	
	.news-content {
		font-size: 26rpx;
		color: #666;
		line-height: 1.5;
		margin-bottom: 16rpx;
		display: block;
		
		&.line-clamp {
			display: -webkit-box;
			-webkit-box-orient: vertical;
			-webkit-line-clamp: 2;
			overflow: hidden;
		}
	}
	
	.news-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 10rpx;
		
		.news-time {
			font-size: 24rpx;
			color: #999;
		}
		
		.read-more-box {
			display: flex;
			align-items: center;
			gap: 8rpx;
			margin-right: 40rpx;
			
			.read-more {
				font-size: 24rpx;
				color: #3B82F6;
			}
		}
	}
}

.news-detail-scroll {
	flex: 1;
	overflow: hidden;
	padding: 0 40rpx;
	
	.detail-content {
		padding: 30rpx 0;
		
		.detail-title {
			font-size: 36rpx;
			font-weight: bold;
			color: #333;
			margin-bottom: 12rpx;
			display: block;
			line-height: 1.4;
		}
		
		.detail-time {
			font-size: 24rpx;
			color: #999;
			margin-bottom: 30rpx;
			display: block;
		}
		
		.detail-text {
			.text-p {
				font-size: 28rpx;
				color: #444;
				line-height: 1.8;
				margin-bottom: 24rpx;
				display: block;
			}
		}
	}
}

.back-btn {
	padding: 10rpx;
	margin-right: 10rpx;
}
</style>
