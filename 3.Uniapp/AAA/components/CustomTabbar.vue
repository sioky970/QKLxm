<template>
	<view class="custom-tabbar">
		<view 
			v-for="(item, index) in tabList" 
			:key="index"
			class="tab-item"
			:class="{ active: current === index }"
			@click="switchTab(index)"
		>
			<view class="tab-icon">
				<SvgIcon :name="item.icon" :size="24" :color="current === index ? activeColor : normalColor" />
			</view>
			<text class="tab-text" :style="{ color: current === index ? activeColor : normalColor }">
				{{ item.text }}
			</text>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import SvgIcon from './SvgIcon.vue'

const props = defineProps({
	current: {
		type: Number,
		default: 0
	}
})

const emit = defineEmits(['change'])

// 配置项
const activeColor = '#0052FF'
const normalColor = '#666666'

// 底部导航配置
const tabList = ref([
	{ 
		text: '首页', 
		icon: 'home',
		pagePath: '/pages/index/index'
	},
	{ 
		text: '市场', 
		icon: 'chart',
		pagePath: '/pages/market/market'
	},
	{ 
		text: '交易', 
		icon: 'swap',
		pagePath: '/pages/trade/trade'
	},
	{ 
		text: '合约', 
		icon: 'document',
		pagePath: '/pages/contract/contract'
	},
	{ 
		text: '资产', 
		icon: 'wallet',
		pagePath: '/pages/assets/assets'
	}
])

// 切换标签
const switchTab = (index) => {
	emit('change', index)
	uni.switchTab({
		url: tabList.value[index].pagePath
	})
}
</script>

<style scoped lang="scss">
.custom-tabbar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	height: 120rpx;
	background: #FFFFFF;
	display: flex;
	align-items: center;
	justify-content: space-around;
	box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
	z-index: 1000;
	padding-bottom: env(safe-area-inset-bottom);
	
	.tab-item {
		flex: 1;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		transition: all 0.3s;
		
		.tab-icon {
			margin-bottom: 6rpx;
			display: flex;
			align-items: center;
			justify-content: center;
		}
		
		.tab-text {
			font-size: 22rpx;
			transition: color 0.3s;
		}
		
		&.active {
			.tab-icon {
				transform: scale(1.1);
			}
			
			.tab-text {
				font-weight: 600;
			}
		}
	}
}
</style>
