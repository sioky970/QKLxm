<template>
	<view>
		<!-- 遮罩层 -->
		<view class="symbol-picker-mask" v-if="show" @click="closePicker"></view>
		
		<!-- 弹窗主体 -->
		<view class="symbol-picker" :class="{ show: show }">
			<!-- 顶部拖拽指示条 -->
			<view class="picker-handle"></view>
			
			<!-- 搜索框 -->
			<view class="picker-search">
				<image class="search-icon" src="/static/icons/search.svg" mode="aspectFit" />
				<input type="text" v-model="searchKeyword" placeholder="搜索" placeholder-class="search-placeholder" />
			</view>
			
			<!-- 排序栏 -->
			<view class="picker-sort">
				<view class="sort-left">
					<view class="sort-btn" @click="toggleSort('name')">
						<text :class="{ active: sortField === 'name' }">名称</text>
						<view class="sort-arrows">
							<view class="arrow-up" :class="{ active: sortField === 'name' && sortOrder === 'asc' }"></view>
							<view class="arrow-down" :class="{ active: sortField === 'name' && sortOrder === 'desc' }"></view>
						</view>
					</view>
					<text class="divider">/</text>
					<view class="sort-btn" @click="toggleSort('volume')">
						<text :class="{ active: sortField === 'volume' }">成交量</text>
						<view class="sort-arrows">
							<view class="arrow-up" :class="{ active: sortField === 'volume' && sortOrder === 'asc' }"></view>
							<view class="arrow-down" :class="{ active: sortField === 'volume' && sortOrder === 'desc' }"></view>
						</view>
					</view>
				</view>
				<view class="sort-right">
					<view class="sort-btn" @click="toggleSort('price')">
						<text :class="{ active: sortField === 'price' }">最新价格</text>
						<view class="sort-arrows">
							<view class="arrow-up" :class="{ active: sortField === 'price' && sortOrder === 'asc' }"></view>
							<view class="arrow-down" :class="{ active: sortField === 'price' && sortOrder === 'desc' }"></view>
						</view>
					</view>
					<text class="divider">/</text>
					<view class="sort-btn" @click="toggleSort('change')">
						<text :class="{ active: sortField === 'change' }">24h涨跌</text>
						<view class="sort-arrows">
							<view class="arrow-up" :class="{ active: sortField === 'change' && sortOrder === 'asc' }"></view>
							<view class="arrow-down" :class="{ active: sortField === 'change' && sortOrder === 'desc' }"></view>
						</view>
					</view>
				</view>
			</view>
			
			<!-- 币种列表 -->
			<scroll-view scroll-y class="picker-list">
				<view 
					v-for="(item, index) in filteredSymbols" 
					:key="index"
					class="symbol-item"
					:class="{ selected: item.symbol === currentSymbol }"
					@click="selectSymbol(item)"
				>
					<view class="item-left">
						<image class="crypto-icon" :src="getCryptoIcon(item.name)" mode="aspectFit" />
						<view class="symbol-info">
							<view class="symbol-row">
								<text class="symbol-name">{{ item.name }}</text>
								<text class="symbol-quote">/{{ item.quote }}</text>
								<text v-if="item.tag" class="tag" :class="item.tagClass">{{ item.tag }}</text>
							</view>
							<text class="volume">Vol {{ item.volume }}</text>
						</view>
					</view>
					<view class="item-right">
						<text class="price">{{ item.price }}</text>
						<text class="change" :class="item.changeClass">{{ item.change }}</text>
					</view>
				</view>
			</scroll-view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { getCryptoIcon } from '@/utils/crypto.js'
import { formatPriceWithPrecision } from '@/utils/format.js'
import marketStore from '@/stores/marketStore.js'

const props = defineProps({
	show: {
		type: Boolean,
		default: false
	},
	currentSymbol: {
		type: String,
		default: 'BTC/USDT'
	}
})

const emit = defineEmits(['update:show', 'select'])

const searchKeyword = ref('')
const sortField = ref('') // name, volume, price, change
const sortOrder = ref('desc') // asc, desc

// 从全局marketStore获取市场数据
const symbolList = computed(() => {
	const allData = marketStore.getAllMarketData()
	
	if (allData.length === 0) {
		return []
	}
	
	// 格式化为符号选择器需要的格式
	return allData.map(data => {
		const priceNum = data.price || 0
		const changeNum = data.change || 0
		const volumeNum = data.volume || 0
		// 使用动态精度（从WebSocket推送的decimal_scale）
		const precision = data.decimal_scale || 4
		
		return {
			symbol: `${data.name}/USDT`,
			name: data.name,
			quote: 'USDT',
			price: formatPriceWithPrecision(priceNum, precision),
			priceNum: priceNum,
			change: changeNum >= 0 ? `+${changeNum.toFixed(2)}%` : `${changeNum.toFixed(2)}%`,
			changeNum: changeNum,
			changeClass: changeNum >= 0 ? 'up' : 'down',
			volume: formatVolume(volumeNum),
			volumeNum: volumeNum,
			tag: data.tag || '',
			tagClass: data.tagClass || '',
			currency_id: data.currency_id,
			legal_id: data.legal_id
		}
	})
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
	return volume.toFixed(2)
}

// 切换排序
const toggleSort = (field) => {
	if (sortField.value === field) {
		// 同一字段，切换顺序：desc -> asc -> 无
		if (sortOrder.value === 'desc') {
			sortOrder.value = 'asc'
		} else {
			sortField.value = ''
			sortOrder.value = 'desc'
		}
	} else {
		sortField.value = field
		sortOrder.value = 'desc'
	}
}

// 过滤和排序后的列表
const filteredSymbols = computed(() => {
	let list = [...symbolList.value]
	
	// 搜索过滤
	if (searchKeyword.value) {
		const keyword = searchKeyword.value.toUpperCase()
		list = list.filter(item => item.name.includes(keyword) || item.symbol.includes(keyword))
	}
	
	// 排序
	if (sortField.value) {
		list.sort((a, b) => {
			let valA, valB
			switch (sortField.value) {
				case 'name':
					valA = a.name
					valB = b.name
					return sortOrder.value === 'asc' 
						? valA.localeCompare(valB) 
						: valB.localeCompare(valA)
				case 'volume':
					valA = a.volumeNum
					valB = b.volumeNum
					break
				case 'price':
					valA = a.priceNum
					valB = b.priceNum
					break
				case 'change':
					valA = a.changeNum
					valB = b.changeNum
					break
			}
			return sortOrder.value === 'asc' ? valA - valB : valB - valA
		})
	}
	
	return list
})

// 关闭弹窗
const closePicker = () => {
	emit('update:show', false)
}

// 选择币种
const selectSymbol = (item) => {
	emit('select', item)
	closePicker()
}

// 重置搜索
watch(() => props.show, (val) => {
	if (val) {
		searchKeyword.value = ''
		// 弹窗打开时，直接使用全局marketStore的数据
		// 不需要再订阅WebSocket或获取数据
	}
})

onMounted(() => {
	// 不需要任何操作，直接使用全局marketStore
})

onUnmounted(() => {
	// 不需要任何操作，使用全局marketStore
})
</script>

<style scoped lang="scss">
.symbol-picker-mask {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 998;
}

.symbol-picker {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	height: 85vh;
	background: #fff;
	border-radius: 24rpx 24rpx 0 0;
	z-index: 999;
	display: flex;
	flex-direction: column;
	transform: translateY(100%);
	transition: transform 0.3s ease-out;
	
	&.show {
		transform: translateY(0);
	}
	
	.picker-handle {
		width: 60rpx;
		height: 8rpx;
		background: #ddd;
		border-radius: 4rpx;
		margin: 20rpx auto;
	}
	
	.picker-search {
		margin: 0 30rpx 20rpx;
		height: 72rpx;
		background: #f5f5f5;
		border-radius: 36rpx;
		display: flex;
		align-items: center;
		padding: 0 24rpx;
		gap: 16rpx;
		
		.search-icon {
			width: 32rpx;
			height: 32rpx;
		}
		
		input {
			flex: 1;
			font-size: 28rpx;
			color: #333;
		}
	}
	
	.picker-sort {
		display: flex;
		justify-content: space-between;
		padding: 20rpx 30rpx;
		border-bottom: 1rpx solid #f5f5f5;
		font-size: 22rpx;
		color: #999;
		
		.sort-left, .sort-right {
			display: flex;
			align-items: center;
			gap: 4rpx;
		}
		
		.sort-btn {
			display: flex;
			align-items: center;
			gap: 4rpx;
			
			text.active {
				color: #333;
				font-weight: bold;
			}
		}
		
		.sort-arrows {
			display: flex;
			flex-direction: column;
			gap: 2rpx;
			
			.arrow-up, .arrow-down {
				width: 0;
				height: 0;
				border-left: 6rpx solid transparent;
				border-right: 6rpx solid transparent;
			}
			
			.arrow-up {
				border-bottom: 8rpx solid #ccc;
				&.active { border-bottom-color: #333; }
			}
			
			.arrow-down {
				border-top: 8rpx solid #ccc;
				&.active { border-top-color: #333; }
			}
		}
		
		.divider {
			margin: 0 8rpx;
		}
	}
	
	.picker-list {
		flex: 1;
		min-height: 0;
	}
	
	.symbol-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24rpx 30rpx;
		
		&.selected {
			background: #fffbe6;
		}
		
		.item-left {
			display: flex;
			align-items: center;
			gap: 20rpx;
			
			.crypto-icon {
				width: 48rpx;
				height: 48rpx;
				border-radius: 50%;
			}
			
			.symbol-info {
				.symbol-row {
					display: flex;
					align-items: center;
					gap: 8rpx;
					
					.symbol-name {
						font-size: 32rpx;
						font-weight: bold;
						color: #333;
					}
					
					.symbol-quote {
						font-size: 24rpx;
						color: #999;
					}
					
					.tag {
						font-size: 20rpx;
						padding: 2rpx 10rpx;
						border-radius: 4rpx;
						margin-left: 8rpx;
						
						&.protect {
							color: #f0b90b;
							border: 1rpx solid #f0b90b;
						}
						
						&.free {
							color: #00c087;
							border: 1rpx solid #00c087;
						}
						
						&.new {
							color: #9966ff;
							border: 1rpx solid #9966ff;
						}
						
						&.ai {
							color: #00bcd4;
							border: 1rpx solid #00bcd4;
						}
						
						&.meme {
							color: #ff6b6b;
							border: 1rpx solid #ff6b6b;
						}
						
						&.stable {
							color: #4caf50;
							border: 1rpx solid #4caf50;
						}
					}
				}
				
				.volume {
					font-size: 22rpx;
					color: #999;
					margin-top: 8rpx;
				}
			}
		}
		
		.item-right {
			text-align: right;
			
			.price {
				font-size: 30rpx;
				font-weight: bold;
				color: #333;
				display: block;
			}
			
			.change {
				font-size: 24rpx;
				margin-top: 8rpx;
				display: block;
				
				&.up { color: #00c087; }
				&.down { color: #F6465D; }
			}
		}
	}
}

.search-placeholder {
	color: #999;
}
</style>
