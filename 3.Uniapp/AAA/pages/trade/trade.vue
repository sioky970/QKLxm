<template>
	<view class="page">
		<!-- 状态栏占位 -->
		<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
		
		<!-- 顶部导航 Tab -->
		<view class="top-nav">
			<view class="nav-left">
				<view class="nav-item active">现货</view>
				<view class="nav-item" @click="goContract">永续合约</view>
				<view class="nav-item" @click="goDeliveryContract">交割合约</view>
			</view>
		</view>

		<view class="scroll-container">
			<scroll-view scroll-y class="content">
				<!-- 币种信息栏 -->
				<view class="symbol-header">
					<view class="symbol-left" @click="showSymbolPicker = true">
						<text class="symbol-name">{{ currentSymbol }}</text>
						<view class="dropdown-icon"></view>
						<text class="symbol-change" :class="symbolChangeClass">{{ symbolChange }}</text>
					</view>
					<view class="symbol-right" @click="goToKline">
						<SvgIcon name="chart" :size="22" color="#333" />
					</view>
				</view>

				<!-- 交易主区域 -->
				<view class="trade-main">
					<!-- 左侧表单 -->
					<view class="trade-form">
						<view class="side-switch">
							<view 
								class="switch-item buy" 
								:class="{ active: tradeSide === 'buy' }"
								@click="tradeSide = 'buy'"
							>买入</view>
							<view 
								class="switch-item sell" 
								:class="{ active: tradeSide === 'sell' }"
								@click="tradeSide = 'sell'"
							>卖出</view>
						</view>
						
						<picker @change="onOrderTypeChange" :value="orderTypeIndex" :range="orderTypeOptions">
							<view class="order-type-picker">
								<SvgIcon name="about" :size="14" color="#999" />
								<text>{{ orderTypeOptions[orderTypeIndex] }}</text>
								<view class="dropdown-icon"></view>
							</view>
						</picker>

						<view class="input-group">
							<input 
								v-if="orderType === 'market'"
								type="text" 
								disabled 
								placeholder="市价" 
								placeholder-class="placeholder" 
							/>
							<input 
								v-else
								type="digit" 
								v-model="limitPrice"
								placeholder="价格" 
								placeholder-class="placeholder" 
							/>
						</view>

						<view class="input-group">
							<input 
								type="digit" 
								v-model="amount"
								@input="onAmountInput"
								:placeholder="orderType === 'market' && tradeSide === 'buy' ? '总额' : '数量'" 
								placeholder-class="placeholder" 
							/>
							<view class="unit-picker">
								<text>{{ (orderType === 'market' && tradeSide === 'buy') ? 'USDT' : currentCoinName }}</text>
								<view class="dropdown-icon gray"></view>
							</view>
						</view>

						<!-- 滑动进度条 -->
						<view class="slider-box">
							<slider 
								:value="sliderValue" 
								@change="onSliderChange"
								@changing="onSliderChanging"
								:min="0" 
								:max="100" 
								:step="1"
								:show-value="false"
								class="custom-slider"
								active-color="#3B82F6"
								background-color="#E5E7EB"
							/>
							<view class="slider-marks">
								<text v-for="(percent, index) in sliderPercents" :key="index" class="mark">{{ percent }}%</text>
							</view>
						</view>

						<!-- 滑点容差选项 -->
						<view class="slippage-check" @click="toggleSlippage">
							<view class="checkbox" :class="{ checked: slippageEnabled }">
								<view v-if="slippageEnabled" class="checkbox-icon">✓</view>
							</view>
							<text @click.stop="showSlippageModal = true">滑点容差</text>
						</view>

						<view class="asset-info">
							<view class="info-row">
								<view class="label-with-icon">
									<text>可用</text>
									<view class="dropdown-icon small"></view>
								</view>
								<view class="value-with-btn">
									<text>{{ availableBalance }} {{ tradeSide === 'buy' ? 'USDT' : currentCoinName }}</text>
									<view class="add-icon">+</view>
								</view>
							</view>
							<view class="info-row">
								<text class="label">{{ tradeSide === 'buy' ? '可买入' : '可卖出' }}</text>
								<text class="value">{{ canTradeAmount }} {{ tradeSide === 'buy' ? currentCoinName : 'USDT' }}</text>
							</view>
							<view class="info-row">
								<text class="label dashed">预估手续费</text>
								<text class="value">{{ estimatedFee }} USDT</text>
							</view>
						</view>

						<button 
							class="submit-btn" 
							:class="tradeSide"
							@click="handleSubmitOrder"
						>
							{{ tradeSide === 'buy' ? '买入' : '卖出' }} {{ currentCoinName }}
						</button>
					</view>

					<!-- 右侧盘口 -->
					<view class="order-book">
						<view class="book-header">
							<text>委托价格(USDT)</text>
							<text>数量({{ currentCoinName }})</text>
						</view>
						
						<!-- 卖盘 -->
						<view class="book-list sell">
							<view class="book-item" v-for="(item, index) in sellList" :key="index" @click="selectPrice(item.price)">
								<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
								<text class="price">{{ item.price }}</text>
								<text class="amount">{{ item.amount }}</text>
							</view>
						</view>

						<!-- 当前价 -->
						<view class="current-price" @click="selectPrice(currentPrice)">
							<text class="price-val" :class="symbolChangeClass">{{ currentPrice }}</text>
							<text class="price-usd">≈ ${{ currentPriceUsd }}</text>
						</view>

						<!-- 买盘 -->
						<view class="book-list buy">
							<view class="book-item" v-for="(item, index) in buyList" :key="index" @click="selectPrice(item.price)">
								<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
								<text class="price">{{ item.price }}</text>
								<text class="amount">{{ item.amount }}</text>
							</view>
						</view>

						<!-- 比例条 -->
						<view class="ratio-bar">
							<text class="buy-val">{{ buyRatio }}%</text>
							<view class="bar-container">
								<view class="bar-buy" :style="{ width: buyRatio + '%' }"></view>
								<view class="bar-sell" :style="{ width: (100 - buyRatio) + '%' }"></view>
							</view>
							<text class="sell-val">{{ (100 - buyRatio).toFixed(2) }}%</text>
						</view>

						<!-- 精度选择 -->
						<view class="precision-selector">
							<picker @change="onPrecisionChange" :value="precisionIndex" :range="precisionOptions">
								<view class="select-box">
									<text>{{ precisionOptions[precisionIndex] }}</text>
									<view class="dropdown-icon gray"></view>
								</view>
							</picker>
							<SvgIcon name="menu" :size="16" color="#999" />
						</view>
					</view>
				</view>

				<!-- 底部委托 Tab -->
				<view class="bottom-tabs">
					<view class="tab-scroll">
						<view 
							class="tab-item" 
							:class="{ active: activeBottomTab === 'pending' }"
							@click="activeBottomTab = 'pending'"
						>当前委托 ({{ pendingOrderCount }})</view>
						<view 
							class="tab-item" 
							:class="{ active: activeBottomTab === 'assets' }"
							@click="activeBottomTab = 'assets'"
						>持有币种 ({{ holdingCoinCount }})</view>
					</view>
					<view @click="goToOrders">
						<SvgIcon name="document" :size="20" color="#333" />
					</view>
				</view>

				<!-- 列表容器 -->
				<view class="order-list-container">
					<!-- 当前委托列表 -->
					<view v-if="activeBottomTab === 'pending'">
						<view v-if="displayOrders.length === 0" class="empty-state">
							<view class="empty-icon">
								<SvgIcon name="record" :size="48" color="#eee" />
							</view>
							<text class="empty-text">暂无委托订单</text>
						</view>
						<view v-else class="order-list">
							<view 
								v-for="order in displayOrders" 
								:key="order.id" 
								class="order-item"
							>
								<view class="order-header">
									<view class="order-symbol">
										<text class="symbol-name">{{ order.symbol }}</text>
										<text class="order-side" :class="order.side">{{ order.side === 'buy' ? '买入' : '卖出' }}</text>
									</view>
									<text class="order-status">{{ getOrderStatusText(order.status) }}</text>
								</view>
								<view class="order-info">
									<view class="info-row">
										<text class="label">价格</text>
										<text class="value">{{ order.type === 'market' ? '市价' : order.price.toFixed(4) }} USDT</text>
									</view>
									<view class="info-row">
										<text class="label">数量</text>
										<text class="value">{{ order.quantity.toFixed(4) }}</text>
									</view>
									<view class="info-row">
										<text class="label">订单价值</text>
										<text class="value value-highlight">{{ order.order_value ? order.order_value.toFixed(2) : '0.00' }} USDT</text>
									</view>
									<view class="info-row">
										<text class="label">成交</text>
										<text class="value">{{ order.deal_number.toFixed(4) }}/{{ order.quantity.toFixed(4) }}</text>
									</view>
									<view class="info-row">
										<text class="label">时间</text>
										<text class="value">{{ formatOrderTime(order.time) }}</text>
									</view>
								</view>
								<view class="order-actions" v-if="order.status === 0">
									<button 
										class="chase-btn" 
										@click="chaseOrder(order)"
									>追单</button>
									<button 
										class="cancel-btn" 
										@click="cancelOrder(order)"
									>撤销</button>
								</view>
							</view>
						</view>
					</view>

					<!-- 持有币种列表 -->
					<view v-if="activeBottomTab === 'assets'">
						<view v-if="holdingCoins.length === 0" class="empty-state">
							<view class="empty-icon">
								<SvgIcon name="record" :size="48" color="#eee" />
							</view>
							<text class="empty-text">暂无持有币种</text>
						</view>
						<view v-else class="coin-list">
							<view 
								v-for="coin in holdingCoins" 
								:key="coin.currency_id" 
								class="coin-item"
								@click="selectCoinForTrade(coin)"
							>
								<view class="coin-info">
									<text class="coin-name">{{ coin.currency_name }}</text>
									<text class="coin-balance">{{ coin.balance.toFixed(8) }}</text>
								</view>
								<view class="coin-value">
									<text class="value-usdt">≈ {{ (coin.balance * (coin.price || 0)).toFixed(2) }} USDT</text>
								</view>
							</view>
						</view>
					</view>
				</view>
				
				<!-- 内部占位符，防止内容被底部栏遮挡 -->
				<view class="tabbar-placeholder"></view>
			</scroll-view>
		</view>
		
		<!-- 自定义底部栏 -->
		<CustomTabbar :current="2" />
		
		<!-- 币种选择弹窗 -->
		<SymbolPicker 
			v-model:show="showSymbolPicker" 
			:currentSymbol="currentSymbol"
			@select="onSymbolSelect"
		/>
		
		<!-- 滑点容差说明弹窗 -->
		<view class="slippage-mask" v-if="showSlippageModal" @click="showSlippageModal = false"></view>
		<view class="slippage-modal" :class="{ show: showSlippageModal }">
			<view class="modal-title">有滑点容差的市价单</view>
			<view class="modal-content">
				<text>启用滑点的市价单将按限价单执行，生效时间设为"立即成交或取消 (IOC)"，即订单成交后，在订单历史中显示为限价单，而非市价单。</text>
				<text class="mt">若订单金额超出滑点容差允许的深度，剩余未成交的部分将取消。</text>
			</view>
			<button class="modal-btn" @click="showSlippageModal = false">确定</button>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import CustomTabbar from '@/components/CustomTabbar.vue'
import SvgIcon from '@/components/SvgIcon.vue'
import SymbolPicker from '@/components/SymbolPicker.vue'
import marketStore from '@/stores/marketStore.js'
import { getAllAssetsWithBalance, submitSpotOrder, getSpotOrderList, cancelSpotOrder, chaseSpotOrder } from '@/utils/api.js'
import wsClient from '@/utils/websocket.js'

const statusBarHeight = ref(0)
const timer = ref(null)
const tradeSide = ref('buy')
const orderType = ref('market') // 'market' or 'limit'
const amount = ref('')
const limitPrice = ref('')
const activeBottomTab = ref('pending') // pending, assets

// 用户资产数据
const userAssets = ref([])
const usdtBalance = ref(0) // USDT余额
const currentCoinBalance = ref(0) // 当前交易币种余额

// 当前委托订单列表
const pendingOrders = ref([])
const pendingOrderCount = computed(() => {
	return pendingOrders.value.filter(order => {
		// 只显示未成交的订单（status=0）
		if (order.status !== 0) {
			return false
		}
		// 根据交易方向过滤
		if (tradeSide.value === 'buy') {
			return order.side === 'buy'
		} else {
			return order.side === 'sell'
		}
	}).length
})

// 显示在列表中的订单（过滤已撤销和已完成的）
const displayOrders = computed(() => {
	return pendingOrders.value.filter(order => {
		// 只显示未成交的订单（status=0）
		if (order.status !== 0) {
			return false
		}
		// 根据交易方向过滤
		if (tradeSide.value === 'buy') {
			return order.side === 'buy'
		} else {
			return order.side === 'sell'
		}
	})
})

// 持有币种列表（余额>0）
const holdingCoins = computed(() => {
	return userAssets.value.filter(asset => {
		// 排除USDT，只显示其他币种
		return asset.currency_name !== 'USDT' && asset.balance > 0
	})
})

const holdingCoinCount = computed(() => holdingCoins.value.length)

// 滑块百分比：0%, 25%, 50%, 75%, 100%
const sliderPercents = [0, 25, 50, 75, 100]
const sliderValue = ref(0) // 当前滑块值 (0-100)

// 币种选择相关
const showSymbolPicker = ref(false)
const currentSymbol = ref('TRX/USDT')
const symbolChangeClass = computed(() => {
	const data = currentMarketData.value
	if (!data) return 'down'
	return data.change >= 0 ? 'up' : 'down'
})
const currentSymbolData = ref(null) // 存储完整的币种数据

// 当前币种的市场数据（从全局状态获取）
const currentMarketData = computed(() => {
	if (!currentSymbolData.value) return null
	// 使用 lastUpdateTime 作为响应式触发器，确保 Map 更新时重新计算
	// eslint-disable-next-line no-unused-vars
	const _trigger = marketStore.state.lastUpdateTime
	const key = `${currentSymbolData.value.currency_id}_${currentSymbolData.value.legal_id || 1}`
	return marketStore.state.marketDataMap.get(key) || null
})

// 当前币价（从市场数据获取）
const coinPrice = computed(() => {
	const data = currentMarketData.value
	if (!data) return 0
	// 市场数据字段为price或now_price
	return parseFloat(data.price || data.now_price) || 0
})

// 判断是否有有效的币价数据
const hasPriceData = computed(() => {
	return coinPrice.value > 0
})

// 涨跌幅（从全局市场数据获取）
const symbolChange = computed(() => {
	const data = currentMarketData.value
	if (!data || !hasPriceData.value) return '---'
	const change = data.change || 0
	return change >= 0 ? `+${change.toFixed(2)}%` : `${change.toFixed(2)}%`
})

// 滑点容差弹窗和状态
const showSlippageModal = ref(false)
const slippageEnabled = ref(false) // 滑点容差是否启用

// 切换滑点容差状态
const toggleSlippage = () => {
	slippageEnabled.value = !slippageEnabled.value
	console.log('[Trade] 滑点容差状态:', slippageEnabled.value)
	
	// 保存到localStorage
	uni.setStorageSync('slippageEnabled', slippageEnabled.value)
	
	uni.showToast({
		title: slippageEnabled.value ? '已启用滑点容差' : '已关闭滑点容差',
		icon: 'none',
		duration: 1500
	})
}

// 币种选择回调
const onSymbolSelect = (item) => {
	currentSymbol.value = item.symbol
	
	// 保存当前选择的币种信息到localStorage，用于页面刷新后恢复
	const symbolData = {
		symbol: item.symbol,
		currency_id: item.currency_id,
		legal_id: item.legal_id
	}
	uni.setStorageSync('lastTradeSymbol', JSON.stringify(symbolData))
	currentSymbolData.value = symbolData
	
	// 基于币种价格更新盘口数据（从全局状态获取价格）
	const marketData = marketStore.getMarketData(item.currency_id, item.legal_id)
	const price = marketData ? marketData.price : 0.2959
	generateOrderBookWithPrice(price)
}

const orderTypeOptions = ['市价单', '限价单']
const orderTypeIndex = computed(() => orderType.value === 'market' ? 0 : 1)

const onOrderTypeChange = (e) => {
	const newType = e.detail.value === 0 ? 'market' : 'limit'
	orderType.value = newType
	
	// 切换到限价单时，自动填充当前币价
	if (newType === 'limit') {
		const price = parseFloat(currentPrice.value) || 0
		if (price > 0) {
			limitPrice.value = price.toString()
		}
	} else {
		// 切换到市价单时，清空限价
		limitPrice.value = ''
	}
	
	// 清空数量
	amount.value = ''
	// 重置滑块
	sliderValue.value = 0
}

const estimatedFee = computed(() => {
	const val = parseFloat(amount.value) || 0
	if (val <= 0) return '0.00'
	
	// 手续费计算：买入USDT的 0.031%
	const feeRate = 0.00031
	
	let orderValue = 0
	if (orderType.value === 'market') {
		if (tradeSide.value === 'buy') {
			// 市价买入：总额就是USDT金额
			orderValue = val
		} else {
			// 市价卖出：数量 * 当前价格
			const price = parseFloat(currentPrice.value) || 0
			orderValue = val * price
		}
	} else {
		// 限价单：数量 * 限价
		const price = parseFloat(limitPrice.value) || 0
		orderValue = val * price
	}
	
	const fee = orderValue * feeRate
	return fee.toFixed(6)
})

// 可用余额（根据买入/卖出显示不同币种）
const availableBalance = computed(() => {
	if (tradeSide.value === 'buy') {
		// 买入时显示USDT余额
		return usdtBalance.value.toFixed(2)
	} else {
		// 卖出时显示当前币种余额
		return currentCoinBalance.value.toFixed(4)
	}
})

// 可买入/可卖出数量（根据输入金额和币价计算）
const canTradeAmount = computed(() => {
	// 没有币价数据时显示 "---"
	if (currentPrice.value === '---') return '---'
	
	const price = parseFloat(currentPrice.value) || 0
	if (price <= 0) return '---'
	
	const inputAmount = parseFloat(amount.value) || 0
	
	if (tradeSide.value === 'buy') {
		if (orderType.value === 'market') {
			// 市价买入：输入的是USDT金额，计算可买入的币种数量
			const canBuy = inputAmount / price
			return canBuy.toFixed(4)
		} else {
			// 限价买入：输入的是币种数量，计算需要的USDT
			const limitPriceVal = parseFloat(limitPrice.value) || price
			const needUsdt = inputAmount * limitPriceVal
			return needUsdt.toFixed(2)
		}
	} else {
		if (orderType.value === 'market') {
			// 市价卖出：输入的是币种数量，计算可卖出的USDT
			const canSell = inputAmount * price
			return canSell.toFixed(2)
		} else {
			// 限价卖出：输入的是币种数量，计算可卖出的USDT
			const limitPriceVal = parseFloat(limitPrice.value) || price
			const canSell = inputAmount * limitPriceVal
			return canSell.toFixed(2)
		}
	}
})

const sellList = ref([])
const buyList = ref([])
const buyRatio = ref(50)

// 当前价格（从全局市场数据获取）
const currentPrice = computed(() => {
	const data = currentMarketData.value
	if (data && data.price > 0) {
		return data.price.toFixed(4)
	}
	// 如果没有全局数据，使用从市场页面传递的价格
	if (currentSymbolData.value && currentSymbolData.value.price) {
		return parseFloat(currentSymbolData.value.price).toFixed(4)
	}
	// 没有有效数据时显示 "---"
	return '---'
})

const currentPriceUsd = computed(() => {
	if (currentPrice.value === '---') return '---'
	const price = parseFloat(currentPrice.value)
	return price.toFixed(8)
})

// 当前币种名称（从 symbol 中提取，如 "BTC/USDT" -> "BTC"）
const currentCoinName = computed(() => {
	if (currentSymbol.value) {
		return currentSymbol.value.split('/')[0]
	}
	return 'TRX'
})

// 精度选择相关
const precisionOptions = ref(['0.0001', '0.001', '0.01', '0.1'])
const precisionIndex = ref(0)

// 动态生成精度选项 (基于当前价格)
const updatePrecisionOptions = (price) => {
	const p = parseFloat(price)
	if (p >= 1000) {
		precisionOptions.value = ['0.01', '0.1', '1', '10']
	} else if (p >= 1) {
		precisionOptions.value = ['0.001', '0.01', '0.1', '1']
	} else {
		precisionOptions.value = ['0.0001', '0.001', '0.01', '0.1']
	}
}

// 格式化价格（根据选中的精度）
const formatPrice = (price) => {
	const precision = parseFloat(precisionOptions.value[precisionIndex.value])
	if (precision >= 1) {
		return (Math.round(price / precision) * precision).toFixed(0)
	}
	const decimals = precisionOptions.value[precisionIndex.value].split('.')[1].length
	return (Math.round(price / precision) * precision).toFixed(decimals)
}

// 生成随机盘口数据（基于指定价格）
const generateOrderBookWithPrice = (basePrice = 0) => {
	// 无效价格时，生成占位数据
	if (!basePrice || basePrice <= 0) {
		const placeholderList = []
		for (let i = 0; i < 5; i++) {
			placeholderList.push({
				price: '---',
				amount: '---',
				depth: 0
			})
		}
		sellList.value = placeholderList
		buyList.value = placeholderList
		buyRatio.value = 50
		return
	}
	
	const newSellList = []
	const newBuyList = []
	
	// 更新精度选项
	updatePrecisionOptions(basePrice)
	
	const precision = parseFloat(precisionOptions.value[precisionIndex.value])
	
	for (let i = 0; i < 5; i++) {
		// 卖盘价格递增
		const sellPrice = basePrice + (5 - i) * precision
		newSellList.push({
			price: formatPrice(sellPrice),
			amount: (Math.random() * 900 + 100).toFixed(2) + 'K',
			depth: Math.random() * 90 + 10
		})
		
		// 买盘价格递减
		const buyPrice = basePrice - (i + 1) * precision
		newBuyList.push({
			price: formatPrice(buyPrice),
			amount: (Math.random() * 900 + 100).toFixed(2) + 'K',
			depth: Math.random() * 90 + 10
		})
	}
	sellList.value = newSellList
	buyList.value = newBuyList
	
	// 随机生成买卖压力比例 (40% - 60% 之间波动)
	buyRatio.value = (Math.random() * 20 + 40).toFixed(2)
}

// 生成随机盘口数据（使用当前价格）
const generateOrderBook = () => {
	// 从全局市场数据获取实时价格
	const data = currentMarketData.value
	let basePrice = 0
	if (data && data.price > 0) {
		basePrice = data.price
	} else if (currentSymbolData.value && currentSymbolData.value.price) {
		basePrice = parseFloat(currentSymbolData.value.price)
	}
	// 如果没有价格数据，传入0会生成占位数据
	generateOrderBookWithPrice(basePrice)
}

// 启动/重置定时器
const startTimer = () => {
	if (timer.value) clearInterval(timer.value)
	
	// 频率逻辑：省略值越大（Index越大），更新频率低2倍
	// Index 0: 1500ms
	// Index 1: 3000ms
	// Index 2: 6000ms
	// Index 3: 12000ms
	const baseFrequency = 1500
	const interval = baseFrequency * Math.pow(2, precisionIndex.value)
	
	timer.value = setInterval(() => {
		generateOrderBook()
	}, interval)
}

const onPrecisionChange = (e) => {
	precisionIndex.value = e.detail.value
	generateOrderBook() // 立即更新一次显示
	startTimer() // 根据新精度重新启动定时器，调整频率
}

// 检查是否有从其他页面传递的交易方向
const checkTradeSide = () => {
	const side = uni.getStorageSync('tradeSide')
	if (side === 'buy' || side === 'sell') {
		tradeSide.value = side
		uni.removeStorageSync('tradeSide') // 使用后清除
	}
}

// 滑块改变事件
const onSliderChange = (e) => {
	const percent = e.detail.value
	sliderValue.value = percent
	calculateAmountBySlider(percent)
}

// 滑动过程中实时计算
const onSliderChanging = (e) => {
	const percent = e.detail.value
	sliderValue.value = percent
	calculateAmountBySlider(percent)
}

// 根据滑块百分比计算金额
const calculateAmountBySlider = (percent) => {
	const price = parseFloat(currentPrice.value) || 0
	
	if (tradeSide.value === 'buy') {
		// 买入：使用USDT余额的百分比
		const usdtAmount = usdtBalance.value * (percent / 100)
		if (orderType.value === 'market') {
			// 市价买入：直接填充USDT总额
			amount.value = usdtAmount.toFixed(2)
		} else {
			// 限价买入：使用当前币价或限价计算可购买的币种数量
			const limitPriceVal = parseFloat(limitPrice.value) || price
			if (limitPriceVal > 0) {
				amount.value = (usdtAmount / limitPriceVal).toFixed(4)
			} else {
				amount.value = '0'
			}
			// 如果限价输入框为空，自动填充当前币价
			if (!limitPrice.value && price > 0) {
				limitPrice.value = price.toFixed(4)
			}
		}
	} else {
		// 卖出：使用当前币种余额的百分比
		const maxAmount = currentCoinBalance.value
		const coinAmount = maxAmount * (percent / 100)
		amount.value = coinAmount.toFixed(4)
		// 如果限价输入框为空，自动填充当前币价
		if (orderType.value === 'limit' && !limitPrice.value && price > 0) {
			limitPrice.value = price.toFixed(4)
		}
	}
}

// 数量输入框变化时的处理
const onAmountInput = (e) => {
	const inputVal = parseFloat(e.detail.value) || 0
	
	if (tradeSide.value === 'sell') {
		// 卖出时：限制最大值为持有数量
		const maxAmount = currentCoinBalance.value
		if (inputVal > maxAmount) {
			amount.value = maxAmount.toFixed(4)
		}
		// 同步更新滑块百分比
		if (maxAmount > 0) {
			const percent = Math.min((inputVal / maxAmount) * 100, 100)
			sliderValue.value = percent
		}
	} else if (tradeSide.value === 'buy') {
		// 买入时：同步更新滑块百分比
		if (orderType.value === 'market') {
			// 市价买入：输入的是USDT，限制最大值为USDT余额
			const maxUsdt = usdtBalance.value
			if (inputVal > maxUsdt) {
				amount.value = maxUsdt.toFixed(2)
			}
			if (maxUsdt > 0) {
				const percent = Math.min((inputVal / maxUsdt) * 100, 100)
				sliderValue.value = percent
			}
		} else {
			// 限价买入：输入的是币种数量，根据USDT余额和限价计算最大可买数量
			const price = parseFloat(limitPrice.value) || parseFloat(currentPrice.value) || 0
			if (price > 0) {
				const maxAmount = usdtBalance.value / price
				if (inputVal > maxAmount) {
					amount.value = maxAmount.toFixed(4)
				}
				if (maxAmount > 0) {
					const percent = Math.min((inputVal / maxAmount) * 100, 100)
					sliderValue.value = percent
				}
			}
		}
	}
}

// 点击盘口价格填充（仅限价单有效）
const selectPrice = (price) => {
	if (orderType.value === 'limit') {
		// 只有限价单时才能点击盘口价格
		// 跳过无效价格
		if (price === '---') return
		
		const priceValue = parseFloat(price)
		if (priceValue > 0) {
			limitPrice.value = priceValue.toString()
		}
	}
}

// 获取用户资产数据
const fetchUserAssets = async () => {
	try {
		const res = await getAllAssetsWithBalance()
		console.log('[Trade] getAllAssetsWithBalance 原始响应:', res)
		console.log('[Trade] res.data:', res.data)
		
		const data = res.data
		const assets = data.assets || []
		userAssets.value = assets
		
		console.log('[Trade] assets数组:', assets)
		if (assets.length > 0) {
			console.log('[Trade] assets数组示例（前3个）:', assets.slice(0, 3))
		}
		
		// 查找USDT余额
		const usdtAsset = assets.find(asset => asset.currency_name === 'USDT')
		console.log('[Trade] USDT资产对象:', usdtAsset)
		if (usdtAsset) {
			const balanceValue = usdtAsset.balance
			console.log('[Trade] USDT balance原始值:', balanceValue, 'typeof:', typeof balanceValue)
			usdtBalance.value = typeof balanceValue === 'string' ? parseFloat(balanceValue) : (balanceValue || 0)
			console.log('[Trade] USDT解析后余额:', usdtBalance.value)
		} else {
			usdtBalance.value = 0
			console.log('[Trade] ⚠️ 未找到USDT资产')
		}
		
		// 查找当前交易币种余额
		const coinName = currentCoinName.value
		const coinAsset = assets.find(asset => asset.currency_name === coinName)
		console.log(`[Trade] ${coinName}资产对象:`, coinAsset)
		if (coinAsset) {
			const balanceValue = coinAsset.balance
			console.log(`[Trade] ${coinName} balance原始值:`, balanceValue, 'typeof:', typeof balanceValue)
			currentCoinBalance.value = typeof balanceValue === 'string' ? parseFloat(balanceValue) : (balanceValue || 0)
			console.log(`[Trade] ${coinName}解析后余额:`, currentCoinBalance.value)
		} else {
			currentCoinBalance.value = 0
			console.log(`[Trade] ⚠️ 未找到${coinName}资产`)
		}
		
		console.log('[Trade] 更新后的余额:', {
			usdt: usdtBalance.value,
			[coinName]: currentCoinBalance.value
		})
	} catch (err) {
		console.error('[Trade] ❌ 获取资产失败:', err)
		uni.showToast({
			title: err.message || '获取资产失败',
			icon: 'none'
		})
	}
}

// 获取当前委托订单
const fetchPendingOrders = async () => {
	try {
		console.log('[Trade] 开始获取委托订单...')
		
		// 添加时间戳防止缓存
		const result = await getSpotOrderList({
			status: 0, // 0=未成交
			page: 1,
			page_size: 50,
			_t: Date.now() // 防止缓存
		})
		
		console.log('[Trade] API返回结果:', result)
		
		if (result.type === 'success' && result.data) {
			const orders = result.data.list || []
			
			console.log('[Trade] 原始订单数据:', JSON.stringify(orders, null, 2))
			
			// 后端已经返回完整的数据结构，直接使用
			pendingOrders.value = orders.map(order => ({
				...order,
				// 确保字段存在（兼容处理）
				quantity: order.number || 0,
				deal_number: order.deal_number || 0
			}))
			
			console.log('[Trade] 委托订单更新:', pendingOrders.value.length, pendingOrders.value)
			console.log('[Trade] 第一个订单详情:', JSON.stringify(pendingOrders.value[0], null, 2))
		} else {
			console.error('[Trade] API返回错误:', result)
		}
	} catch (err) {
		console.error('[Trade] 获取委托订单失败:', err)
	}
}

// 取消订单
const cancelOrder = async (order) => {
	uni.showModal({
		title: '确认撤销',
		content: `确定要撤销该笔${order.side === 'buy' ? '买入' : '卖出'}订单吗？`,
		success: async (res) => {
			if (res.confirm) {
				uni.showLoading({ title: '处理中...' })
				
				try {
					const result = await cancelSpotOrder({
						order_id: order.id
					})
					
					uni.hideLoading()
					
					if (result.type === 'success') {
						uni.showToast({
							title: '撤销成功',
							icon: 'success'
						})
						
						// 刷新订单列表和余额
						await fetchPendingOrders()
						await fetchUserAssets()
					} else {
						uni.showToast({
							title: result.message || '撤销失败',
							icon: 'none'
						})
					}
				} catch (err) {
					uni.hideLoading()
					console.error('[Trade] 撤销订单失败:', err)
					uni.showToast({
						title: err.message || '网络错误',
						icon: 'none'
					})
				}
			}
		}
	})
}

// 追单：以当前市场价修改订单
const chaseOrder = async (order) => {
	// 获取当前币价
	const currentPrice = coinPrice.value
	
	console.log('[Trade] 追单检查:', {
		coinPrice: currentPrice,
		currentMarketData: currentMarketData.value,
		currentSymbolData: currentSymbolData.value,
		currentSymbol: currentSymbol.value,
		marketStoreSize: marketStore.state.marketDataMap.size,
		order: order
	})
	
	// 如果currentMarketData存在但价格为0，打印详细数据
	if (currentMarketData.value) {
		console.log('[Trade] 市场数据详情:', {
			price: currentMarketData.value.price,
			now_price: currentMarketData.value.now_price,
			allFields: Object.keys(currentMarketData.value)
		})
	}
	
	if (!currentPrice || currentPrice <= 0) {
		uni.showToast({
			title: '无法获取市场价格',
			icon: 'none'
		})
		return
	}
	
	uni.showModal({
		title: '追单确认',
		content: `将以当前市场价 ${currentPrice.toFixed(4)} USDT 修改订单价格，立即成交。是否继续？`,
		success: async (res) => {
			if (res.confirm) {
				uni.showLoading({ title: '处理中...' })
				
				try {
					const result = await chaseSpotOrder({
						order_id: order.id,
						new_price: currentPrice
					})
					
					uni.hideLoading()
					
					if (result.type === 'success') {
						uni.showToast({
							title: '追单成功',
							icon: 'success'
						})
						
						// 刷新订单列表和余额
						await fetchPendingOrders()
						await fetchUserAssets()
					} else {
						uni.showToast({
							title: result.message || '追单失败',
							icon: 'none'
						})
					}
				} catch (err) {
					uni.hideLoading()
					console.error('[Trade] 追单失败:', err)
					uni.showToast({
						title: err.message || '网络错误',
						icon: 'none'
					})
				}
			}
		}
	})
}

	// 选择币种进行交易
const selectCoinForTrade = (coin) => {
	// 切换到卖出模式
	tradeSide.value = 'sell'
	
	// 更新当前交易对
	const symbolData = {
		symbol: `${coin.currency_name}/USDT`,
		currency_id: coin.currency_id,
		legal_id: 3,
		price: coin.price || 0
	}
	
	currentSymbol.value = symbolData.symbol
	currentSymbolData.value = symbolData
	
	// 保存到本地存储
	uni.setStorageSync('lastTradeSymbol', JSON.stringify(symbolData))
	
	// 更新余额
	currentCoinBalance.value = coin.balance
	
	// 切换到当前委托tab
	activeBottomTab.value = 'pending'
	
	console.log('[Trade] 切换交易对:', symbolData)
}

// 格式化订单状态
const getOrderStatusText = (status) => {
	const statusMap = {
		0: '未成交',
		1: '部分成交',
		2: '已成交',
		3: '已撤销'
	}
	return statusMap[status] || '未知'
}

// 格式化订单时间
const formatOrderTime = (timestamp) => {
	if (!timestamp) return '-'
	
	const date = new Date(timestamp * 1000)
	const month = (date.getMonth() + 1).toString().padStart(2, '0')
	const day = date.getDate().toString().padStart(2, '0')
	const hour = date.getHours().toString().padStart(2, '0')
	const minute = date.getMinutes().toString().padStart(2, '0')
	
	return `${month}-${day} ${hour}:${minute}`
}

// 提交订单
const handleSubmitOrder = async () => {
	// 检查是否有有效的币价数据
	if (currentPrice.value === '---' || !hasPriceData.value) {
		uni.showToast({
			title: '暂无币价数据，请稍后再试',
			icon: 'none'
		})
		return
	}
	
	// 校验输入
	const inputAmount = parseFloat(amount.value)
	if (!inputAmount || inputAmount <= 0) {
		uni.showToast({
			title: '请输入交易数量',
			icon: 'none'
		})
		return
	}
	
	// 限价单需要检查价格
	let orderPrice = parseFloat(currentPrice.value) || 0
	if (orderType.value === 'limit') {
		orderPrice = parseFloat(limitPrice.value)
		if (!orderPrice || orderPrice <= 0) {
			uni.showToast({
				title: '请输入限价',
				icon: 'none'
			})
			return
		}
	}
	
	// 获取币种ID和法币ID
	const currencyId = currentSymbolData.value?.currency_id
	const legalId = currentSymbolData.value?.legal_id || 3 // 默认USDT
	
	if (!currencyId) {
		uni.showToast({
			title: '请选择交易对',
			icon: 'none'
		})
		return
	}
	
	// 计算实际数量（市价买入时，输入的是USDT金额，需要转换为币种数量）
	let orderQuantity = inputAmount
	if (orderType.value === 'market' && tradeSide.value === 'buy') {
		// 市价买入：输入的是USDT，需要除以当前价格得到币种数量
		if (orderPrice <= 0) {
			uni.showToast({
				title: '获取市场价格失败',
				icon: 'none'
			})
			return
		}
		orderQuantity = inputAmount / orderPrice
	}
	
	// 检查余额
	if (tradeSide.value === 'buy') {
		const requiredUsdt = orderPrice * orderQuantity
		if (requiredUsdt > usdtBalance.value) {
			uni.showToast({
				title: 'USDT余额不足',
				icon: 'none'
			})
			return
		}
	} else {
		if (orderQuantity > currentCoinBalance.value) {
			uni.showToast({
				title: `${currentCoinName.value}余额不足`,
				icon: 'none'
			})
			return
		}
	}
	
	// 显示确认对话框
	const action = tradeSide.value === 'buy' ? '买入' : '卖出'
	const orderTypeText = orderType.value === 'market' ? '市价' : '限价'
	const totalValue = orderPrice * orderQuantity
	
	uni.showModal({
		title: `确认${action}订单`,
		content: `${orderTypeText}${action} ${orderQuantity.toFixed(8)} ${currentCoinName.value}\n价格: ${orderPrice.toFixed(4)} USDT\n总价: ${totalValue.toFixed(2)} USDT`,
		success: async (res) => {
			if (res.confirm) {
				// 用户确认，提交订单
				uni.showLoading({
					title: '提交中...'
				})
				
				try {
					const result = await submitSpotOrder({
						currency_id: currencyId,
						legal_id: legalId,
						type: orderType.value,
						side: tradeSide.value,
						price: orderPrice,
						quantity: orderQuantity
					})
					
					uni.hideLoading()
					
					if (result.type === 'success') {
						uni.showToast({
							title: '订单提交成功',
							icon: 'success'
						})
						
						// 清空输入
						amount.value = ''
						if (orderType.value === 'limit') {
							limitPrice.value = ''
						}
						sliderValue.value = 0
						
						// 刷新余额和订单列表
						await fetchUserAssets()
						await fetchPendingOrders()
						
						// 显示订单详情
						const orderData = result.data
						console.log('[Trade] 订单提交成功:', orderData)
					} else {
						uni.showToast({
							title: result.message || '订单提交失败',
							icon: 'none'
						})
					}
				} catch (err) {
					uni.hideLoading()
					console.error('[Trade] 提交订单失败:', err)
					uni.showToast({
						title: err.message || '网络错误',
						icon: 'none'
					})
				}
			}
		}
	})
}

// 监听交易方向变化，重置滑块并刷新订单列表
watch(tradeSide, () => {
	sliderValue.value = 0
	amount.value = ''
	// 切换交易方向时刷新委托订单
	fetchPendingOrders()
})

// 监听币种变化，更新余额
watch(currentCoinName, () => {
	fetchUserAssets()
})

// 监听全局市场数据变化，立即同步更新盘口数据
watch(currentMarketData, (newData, oldData) => {
	// 移除oldData判断，确保首次加载也能触发
	if (newData && newData.price > 0) {
		const oldPrice = oldData ? oldData.price : 0
		const newPrice = newData.price
		
		// 只有价格变化时才更新
		if (oldPrice !== newPrice) {
			console.log('[Trade] 币价更新:', {
				name: newData.name,
				oldPrice,
				newPrice,
				isFirstLoad: !oldData
			})
			// 立即更新盘口数据，确保买卖盘价格与实时币价高度同步
			generateOrderBook()
		}
	}
}, { immediate: false, deep: false })

// 检查并加载从市场页面传递的币种信息
const checkSymbolFromMarket = () => {
	// 优先检查是否有从市场页面跳转过来的币种数据
	let symbolData = uni.getStorageSync('currentTradeSymbol')
	let isFromMarket = false
	
	if (symbolData) {
		isFromMarket = true
		// 使用后清除，避免下次加载
		uni.removeStorageSync('currentTradeSymbol')
	} else {
		// 如果没有跳转数据，尝试恢复上次选择的币种（支持页面刷新）
		symbolData = uni.getStorageSync('lastTradeSymbol')
	}
	
	if (symbolData) {
		try {
			const data = JSON.parse(symbolData)
			console.log(`[Trade] ${isFromMarket ? '加载市场跳转' : '恢复上次'}币种数据:`, data)
			
			// 更新显示信息
			currentSymbol.value = data.symbol || 'TRX/USDT'
			currentSymbolData.value = data
			
			// 如果是从市场跳转过来的，保存到lastTradeSymbol以便下次恢复
			if (isFromMarket) {
				uni.setStorageSync('lastTradeSymbol', symbolData)
			}
			
			// 基于币种价格更新盘口数据（从全局状态获取）
			const marketData = marketStore.getMarketData(data.currency_id, data.legal_id)
			const price = marketData ? marketData.price : (parseFloat(data.price) || 0.2959)
			generateOrderBookWithPrice(price)
		} catch (err) {
			console.error('[Trade] 解析币种数据失败:', err)
		}
	} else {
		// 没有保存的币种数据时，使用默认币种（TRX/USDT）初始化
		initDefaultSymbol()
	}
}

// 初始化默认币种
const initDefaultSymbol = () => {
	// 尝试从 marketStore 获取 TRX 的数据
	const trxData = marketStore.getMarketDataByName('TRX')
	if (trxData && trxData.currency_id) {
		currentSymbolData.value = {
			symbol: trxData.symbol || 'TRX/USDT',
			currency_id: trxData.currency_id,
			legal_id: trxData.legal_id || 1
		}
		console.log('[Trade] 使用marketStore初始化默认币种:', currentSymbolData.value)
	} else {
		// marketStore 可能还未加载，使用默认 currency_id
		currentSymbolData.value = {
			symbol: 'TRX/USDT',
			currency_id: 8,  // TRX 默认 ID
			legal_id: 1
		}
		console.log('[Trade] 使用默认配置初始化币种:', currentSymbolData.value)
	}
}

onMounted(() => {
	uni.hideTabBar()
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 检查交易方向
	checkTradeSide()
	
	// 检查并加载从市场页面传递的币种信息（会初始化默认币种）
	checkSymbolFromMarket()
	
	// 恢复滑点容差状态
	const savedSlippage = uni.getStorageSync('slippageEnabled')
	if (savedSlippage !== undefined && savedSlippage !== '') {
		slippageEnabled.value = savedSlippage
		console.log('[Trade] 恢复滑点容差状态:', slippageEnabled.value)
	}
	
	// 订阅WebSocket频道（余额和订单状态实时推送）
	setupWebSocketSubscriptions()
	
	// 获取用户资产数据
	fetchUserAssets()
	
	// 获取当前委托订单
	fetchPendingOrders()
	
	// 初始化盘口
	generateOrderBook()
	startTimer()
})

onUnmounted(() => {
	if (timer.value) {
		clearInterval(timer.value)
	}
	// 取消WebSocket订阅
	cleanupWebSocketSubscriptions()
})

// WebSocket订阅管理
let walletChannel = null
let orderChannel = null

const setupWebSocketSubscriptions = () => {
	// 获取当前用户ID
	const userInfo = uni.getStorageSync('userInfo')
	if (!userInfo || !userInfo.id) {
		console.log('[Trade] 用户未登录，跳过WebSocket订阅')
		return
	}
	
	const userId = userInfo.id
	walletChannel = `wallet:${userId}`
	orderChannel = `order:${userId}`
	
	// 确保WebSocket已连接
	wsClient.connect()
	
	// 订阅余额更新频道
	wsClient.subscribe(walletChannel, (data) => {
		console.log('[Trade] 收到余额更新推送:', data)
		handleBalanceUpdate(data)
	})
	
	// 订阅订单状态更新频道
	wsClient.subscribe(orderChannel, (data) => {
		console.log('[Trade] 收到订单状态更新推送:', data)
		handleOrderUpdate(data)
	})
	
	console.log('[Trade] WebSocket订阅完成:', { walletChannel, orderChannel })
}

const cleanupWebSocketSubscriptions = () => {
	if (walletChannel) {
		wsClient.unsubscribe(walletChannel)
		walletChannel = null
	}
	if (orderChannel) {
		wsClient.unsubscribe(orderChannel)
		orderChannel = null
	}
	console.log('[Trade] WebSocket订阅已清理')
}

// 处理余额更新推送
const handleBalanceUpdate = (data) => {
	if (!data || !data.assets) return
	
	// 更新用户资产列表
	userAssets.value = data.assets
	
	// 更新USDT余额
	const usdtAsset = data.assets.find(asset => asset.currency_name === 'USDT')
	if (usdtAsset) {
		usdtBalance.value = usdtAsset.balance || 0
		console.log('[Trade] USDT余额已更新:', usdtBalance.value)
	}
	
	// 更新当前交易币种余额
	const coinName = currentCoinName.value
	const coinAsset = data.assets.find(asset => asset.currency_name === coinName)
	if (coinAsset) {
		currentCoinBalance.value = coinAsset.balance || 0
		console.log('[Trade] 币种余额已更新:', coinName, currentCoinBalance.value)
	}
	
	uni.showToast({
		title: '余额已更新',
		icon: 'none',
		duration: 1500
	})
}

// 处理订单状态更新推送
const handleOrderUpdate = (data) => {
	if (!data || !data.order_id) return
	
	const { order_id, status, deal_number, price } = data
	
	// 查找并更新订单列表中的订单
	const orderIndex = pendingOrders.value.findIndex(o => o.id === order_id)
	if (orderIndex !== -1) {
		if (status === 2) {
			// 订单已成交，从当前委托列表中移除
			pendingOrders.value.splice(orderIndex, 1)
			console.log('[Trade] 订单已成交并移除:', order_id)
		} else if (status === 1) {
			// 部分成交，更新订单信息
			pendingOrders.value[orderIndex].deal_number = deal_number
			pendingOrders.value[orderIndex].status = status
			console.log('[Trade] 订单部分成交:', order_id, deal_number)
		}
	}
	
	// 刷新订单列表确保数据一致
	fetchPendingOrders()
}

// 页面显示时检查交易方向和币种信息（从其他页面切换回来时）
onShow(() => {
	checkTradeSide()
	checkSymbolFromMarket()
	// 不再需要订阅WebSocket，使用全局marketStore的数据
	// 刷新资产数据和订单列表
	fetchUserAssets()
	fetchPendingOrders()
})

const goToKline = () => {
	uni.navigateTo({
		url: '/pages/trade/kline'
	})
}

const goContract = () => {
	uni.switchTab({
		url: '/pages/contract/contract'
	})
}

const goDeliveryContract = () => {
	// 设置标记，让合约页面切换到交割合约tab
	uni.setStorageSync('contractTab', 'seconds')
	uni.switchTab({
		url: '/pages/contract/contract'
	})
}

const goToOrders = () => {
	uni.navigateTo({
		url: '/pages/trade/orders'
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

.scroll-container {
	flex: 1;
	height: 0;
	min-height: 0;
	overflow: hidden;
}

.content {
	height: 100%;
}

/* 顶部导航 */
.top-nav {
	height: 101rpx; // 88rpx * 1.15 = 101.2rpx
	display: flex;
	align-items: center;
	padding: 0 30rpx;
	flex-shrink: 0;
	border-bottom: 1rpx solid #f5f5f5; // 增加浅灰色分割线

	.nav-left {
		display: flex;
		gap: 40rpx;
		
		.nav-item {
			font-size: 37rpx; // 32rpx * 1.15 = 36.8rpx
			font-weight: bold;
			color: #999;
			position: relative;
			
			&.active {
				color: #333;
				font-size: 41rpx; // 36rpx * 1.15 = 41.4rpx
			}
		}
	}
}

/* 币种信息 */
.symbol-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 30rpx;
	
	.symbol-left {
		display: flex;
		align-items: baseline;
		gap: 12rpx;
		
		.symbol-name {
			font-size: 44rpx;
			font-weight: bold;
			color: #333;
		}
		
		.symbol-change {
			font-size: 24rpx;
			margin-left: 8rpx;
			&.down { color: #F6465D; }
			&.up { color: #00c087; }
		}
	}
	
	.symbol-right {
		display: flex;
		align-items: center;
		gap: 30rpx;
		
		.more-btn {
			position: relative;
			.dot {
				position: absolute;
				top: 0;
				right: -4rpx;
				width: 12rpx;
				height: 12rpx;
				background: #fcc419;
				border-radius: 50%;
				border: 2rpx solid #fff;
			}
		}
	}
}

/* 交易主区 */
.trade-main {
	display: flex;
	padding: 0 30rpx;
	gap: 30rpx;
	
	.trade-form {
		flex: 1.2;
		
		.side-switch {
			display: flex;
			height: 68rpx; // 减少约5%的高度 (从72rpx到68rpx)
			background: #f5f5f5;
			border-radius: 34rpx; // 更加圆润 (高度的一半)
			overflow: hidden;
			margin-bottom: 24rpx;
			
			.switch-item {
				flex: 1;
				display: flex;
				align-items: center;
				justify-content: center;
				font-size: 28rpx;
				color: #666;
				transition: all 0.2s;
				
				&.buy.active {
					background: #00c087;
					color: #fff;
					clip-path: polygon(0 0, 85% 0, 100% 100%, 0% 100%);
					border-radius: 34rpx 0 0 34rpx;
				}
				&.sell.active {
					background: #F6465D; // 调整为更专业的交易红
					color: #fff;
					border-radius: 0 34rpx 34rpx 0;
				}
			}
		}
		
		.order-type-picker {
			height: 80rpx;
			border: 1rpx solid #eee;
			border-radius: 8rpx;
			display: flex;
			align-items: center;
			padding: 0 20rpx;
			margin-bottom: 20rpx;
			font-size: 28rpx;
			color: #333;
			gap: 10rpx;
			
			text { flex: 1; margin-left: 10rpx; }
		}
		
		.input-group {
			height: 80rpx;
			background: #f5f5f5;
			border-radius: 8rpx;
			display: flex;
			align-items: center;
			padding: 0 24rpx;
			margin-bottom: 20rpx;
			
			input {
				flex: 1;
				font-size: 28rpx;
				color: #333;
			}
			
			.unit-picker {
				display: flex;
				align-items: center;
				gap: 8rpx;
				border-left: 1rpx solid #ddd;
				padding-left: 20rpx;
				font-size: 24rpx;
				color: #333;
			}
		}
		
		.slider-box {
			margin: 40rpx 0;
			position: relative;
			
			.custom-slider {
				width: 100%;
				margin: 0;
			}
			
			.slider-marks {
				display: flex;
				justify-content: space-between;
				margin-top: 16rpx;
				
				.mark {
					font-size: 22rpx;
					color: #999;
					text-align: center;
					flex: 1;
				}
			}
		}
		
		.slippage-check {
			display: flex;
			align-items: center;
			gap: 12rpx;
			margin-bottom: 30rpx;
			cursor: pointer;
			
			.checkbox {
				width: 32rpx;
				height: 32rpx;
				border: 2rpx solid #d0d0d0;
				border-radius: 6rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				transition: all 0.2s;
				background: #fff;
				
				&.checked {
					background: #3B82F6;
					border-color: #3B82F6;
				}
				
				.checkbox-icon {
					color: #fff;
					font-size: 20rpx;
					font-weight: bold;
				}
			}
			
			text { 
				font-size: 24rpx; 
				color: #333; 
				border-bottom: 1rpx dotted #ccc;
				cursor: pointer;
			}
		}
		
		.asset-info {
			margin-bottom: 40rpx;
			.info-row {
				display: flex;
				justify-content: space-between;
				margin-bottom: 16rpx;
				font-size: 24rpx;
				
				.label { color: #999; &.dashed { border-bottom: 1rpx dotted #ccc; } }
				.value { color: #333; font-weight: 500; }
				
				.label-with-icon {
					display: flex;
					align-items: center;
					gap: 4rpx;
					color: #999;
				}
				
				.value-with-btn {
					display: flex;
					align-items: center;
					gap: 8rpx;
					.add-icon {
						width: 28rpx;
						height: 28rpx;
						background: #fcc419;
						color: #fff;
						border-radius: 50%;
						display: flex;
						align-items: center;
						justify-content: center;
						font-size: 20rpx;
					}
				}
			}
		}
		
		.submit-btn {
			height: 84rpx; // 减少约5%的高度 (从88rpx到84rpx)
			border-radius: 42rpx; // 更加圆润 (高度的一半)
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 32rpx;
			font-weight: bold;
			color: #fff;
			transition: all 0.2s;
			&.buy { background: #00c087; }
			&.sell { background: #F6465D; }
		}
	}
	
	.order-book {
		flex: 1;
		
		.book-header {
			display: flex;
			justify-content: space-between;
			margin-bottom: 20rpx;
			text { font-size: 20rpx; color: #999; }
		}
		
		.book-list {
			.book-item {
				height: 48rpx;
				display: flex;
				justify-content: space-between;
				align-items: center;
				position: relative;
				cursor: pointer;
				transition: background 0.2s;
				
				&:active {
					background: rgba(0, 0, 0, 0.05);
				}
				
				.depth-bar {
					position: absolute;
					right: 0;
					top: 0;
					bottom: 0;
					opacity: 0.1;
					transition: width 0.8s ease-in-out;
				}
				
				.price { font-size: 24rpx; font-weight: 500; position: relative; z-index: 1; }
				.amount { font-size: 24rpx; color: #333; position: relative; z-index: 1; }
			}
			
			&.sell {
				.book-item .price { color: #F6465D; }
				.book-item .depth-bar { background: #F6465D; }
			}
			&.buy {
				.book-item .price { color: #00c087; }
				.book-item .depth-bar { background: #00c087; }
			}
		}
		
		.current-price {
			padding: 24rpx 0;
			display: flex;
			flex-direction: column;
			align-items: center;
			cursor: pointer;
			transition: background 0.2s;
			
			&:active {
				background: rgba(0, 0, 0, 0.05);
			}
			
			.price-val { font-size: 36rpx; font-weight: bold; &.up { color: #00c087; } &.down { color: #F6465D; } }
			.price-usd { font-size: 20rpx; color: #999; margin-top: 4rpx; }
		}
		
		.ratio-bar {
			margin-top: 20rpx;
			display: flex;
			align-items: center;
			gap: 10rpx;
			font-size: 20rpx;
			
			.buy-val { color: #00c087; }
			.sell-val { color: #F6465D; }
			
			.bar-container {
				flex: 1;
				height: 6rpx;
				display: flex;
				border-radius: 3rpx;
				overflow: hidden;
				
				.bar-buy { background: #00c087; transition: width 0.8s ease-in-out; }
				.bar-sell { background: #F6465D; margin-left: 2rpx; transition: width 0.8s ease-in-out; }
			}
		}
		
		.precision-selector {
			margin-top: 30rpx;
			display: flex;
			align-items: center;
			justify-content: space-between;
			
			.select-box {
				background: #f5f5f5;
				height: 54rpx;
				padding: 0 16rpx;
				border-radius: 4rpx;
				display: flex;
				align-items: center;
				gap: 10rpx;
				font-size: 24rpx;
				color: #333;
			}
		}
	}
}

/* 底部 Tab */
.bottom-tabs {
	margin-top: 40rpx;
	display: flex;
	align-items: center;
	padding: 0 30rpx;
	
	.tab-scroll {
		flex: 1;
		display: flex;
		gap: 40rpx;
		height: 88rpx;
		align-items: center;
		
		.tab-item {
			font-size: 28rpx;
			color: #999;
			font-weight: 500;
			&.active {
				color: #333;
				border-bottom: 4rpx solid #F7D100;
				padding-bottom: 4rpx;
			}
		}
	}
}

.order-list-container {
	min-height: 300rpx;
	padding: 40rpx 30rpx;
	
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60rpx 0;
		
		.empty-icon {
			margin-bottom: 20rpx;
		}
		
		.empty-text {
			font-size: 24rpx;
			color: #999;
		}
	}
	
	/* 订单列表 */
	.order-list {
		display: flex;
		flex-direction: column;
		gap: 20rpx;
		
		.order-item {
			background: #f8f8f8;
			border-radius: 12rpx;
			padding: 24rpx;
			position: relative;
			
			.order-header {
				display: flex;
				justify-content: space-between;
				align-items: center;
				margin-bottom: 16rpx;
				
				.order-symbol {
					display: flex;
					align-items: center;
					gap: 12rpx;
					
					.symbol-name {
						font-size: 30rpx;
						font-weight: 600;
						color: #333;
					}
					
					.order-side {
						font-size: 24rpx;
						padding: 4rpx 12rpx;
						border-radius: 4rpx;
						font-weight: 500;
						
						&.buy {
							color: #16c784;
							background: #e8f8f0;
						}
						
						&.sell {
							color: #ea3943;
							background: #ffeaec;
						}
					}
				}
				
				.order-status {
					font-size: 24rpx;
					color: #999;
				}
			}
			
			.order-info {
				display: flex;
				flex-direction: column;
				gap: 12rpx;
				margin-bottom: 20rpx;
				
				.info-row {
					display: flex;
					justify-content: space-between;
					font-size: 26rpx;
					
					.label {
						color: #999;
					}
					
					.value {
						color: #333;
						font-weight: 500;
					}
					
					.value-highlight {
						color: #16c784;
						font-weight: 600;
					}
				}
			}
			
			.order-actions {
				display: flex;
				gap: 16rpx;
				margin-top: 12rpx;
				
				.chase-btn {
					flex: 1;
					height: 64rpx;
					background: #fff;
					border: 2rpx solid #16c784;
					color: #16c784;
					border-radius: 8rpx;
					font-size: 28rpx;
					font-weight: 500;
				}
				
				.cancel-btn {
					flex: 1;
					height: 64rpx;
					background: #fff;
					border: 2rpx solid #ea3943;
					color: #ea3943;
					border-radius: 8rpx;
					font-size: 28rpx;
					font-weight: 500;
				}
			}
		}
	}
	
	/* 币种列表 */
	.coin-list {
		display: flex;
		flex-direction: column;
		gap: 16rpx;
		
		.coin-item {
			background: #f8f8f8;
			border-radius: 12rpx;
			padding: 24rpx;
			display: flex;
			justify-content: space-between;
			align-items: center;
			
			.coin-info {
				display: flex;
				flex-direction: column;
				gap: 8rpx;
				
				.coin-name {
					font-size: 30rpx;
					font-weight: 600;
					color: #333;
				}
				
				.coin-balance {
					font-size: 26rpx;
					color: #666;
				}
			}
			
			.coin-value {
				.value-usdt {
					font-size: 26rpx;
					color: #999;
				}
			}
		}
	}
}

/* 通用组件 */
.dropdown-icon {
	width: 0;
	height: 0;
	border-left: 8rpx solid transparent;
	border-right: 8rpx solid transparent;
	border-top: 10rpx solid #333;
	margin-top: 6rpx;
	
	&.gray { border-top-color: #999; }
	&.small { border-top-width: 8rpx; border-left-width: 6rpx; border-right-width: 6rpx; }
}

.tabbar-placeholder {
	height: 120rpx;
	padding-bottom: env(safe-area-inset-bottom);
	flex-shrink: 0;
}

.placeholder {
	color: #ccc;
}

/* 滑点容差弹窗 */
.slippage-mask {
	position: fixed;
	left: 0;
	right: 0;
	top: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 999;
}

.slippage-modal {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	background: #fff;
	border-radius: 32rpx 32rpx 0 0;
	padding: 48rpx 40rpx calc(48rpx + env(safe-area-inset-bottom));
	z-index: 1000;
	transform: translateY(100%);
	transition: transform 0.3s ease;
	
	&.show {
		transform: translateY(0);
	}
	
	.modal-title {
		font-size: 36rpx;
		font-weight: 600;
		color: #333;
		margin-bottom: 32rpx;
	}
	
	.modal-content {
		display: flex;
		flex-direction: column;
		font-size: 28rpx;
		color: #666;
		line-height: 1.7;
		
		.mt {
			margin-top: 24rpx;
		}
	}
	
	.modal-btn {
		margin-top: 48rpx;
		width: 100%;
		height: 96rpx;
		background: #0052FF;
		border-radius: 48rpx;
		color: #fff;
		font-size: 32rpx;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		
		&::after {
			border: none;
		}
	}
}
</style>
