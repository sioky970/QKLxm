<template>
	<view class="page">
		<!-- 自定义顶部导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="navbar-left" @click="goBack">
					<SvgIcon name="back" :size="20" color="#fff" />
				</view>
				<text class="nav-title">客服</text>
				<view class="navbar-right"></view>
			</view>
		</view>

		<!-- 客服内容区 -->
		<view class="customer-service-content">
			<!-- WebView容器 -->
			<view class="webview-container" v-if="showWebView">
				<web-view 
					:src="customerServiceUrl" 
					@load="onWebViewLoad"
					@error="onWebViewError"
				></web-view>
			</view>
			
			<!-- 备用方案：iframe容器（H5环境） -->
			<view class="iframe-container" v-if="showIframe">
				<iframe 
					:src="customerServiceUrl"
					class="customer-service-iframe"
					@load="onIframeLoad"
					@error="onIframeError"
				></iframe>
			</view>
			
			<!-- 备用方案：显示链接和复制按钮 -->
			<view class="fallback-content" v-if="showFallback">
				<view class="fallback-header">
					<SvgIcon name="customer-service" :size="48" color="#007AFF" />
					<text class="fallback-title">联系客服</text>
					<text class="fallback-subtitle">点击下方链接或复制链接到浏览器访问</text>
				</view>
				
				<view class="link-container">
					<view class="link-box" @click="openInBrowser">
						<text class="link-text">{{ customerServiceUrl }}</text>
						<SvgIcon name="expand" :size="16" color="#007AFF" />
					</view>
					
					<view class="action-buttons">
						<view class="copy-btn" @click="copyLink">
							<SvgIcon name="copy" :size="16" color="#fff" />
							<text class="btn-text">复制链接</text>
						</view>
						<view class="open-btn" @click="openInBrowser">
							<SvgIcon name="globe" :size="16" color="#fff" />
							<text class="btn-text">打开链接</text>
						</view>
					</view>
				</view>
				
				<view class="instruction">
					<text class="instruction-title">使用说明：</text>
					<text class="instruction-text">1. 点击"打开链接"在新窗口中打开客服</text>
					<text class="instruction-text">2. 或者复制链接到浏览器地址栏访问</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SvgIcon from '@/components/SvgIcon.vue'

// 客服链接
const customerServiceUrl = 'https://xjb.axd05kh.cfd/chat/index?channelId=ee30ef2870614d838edc5185f286abcf'

// 状态栏高度
const statusBarHeight = ref(0)

// 显示状态
const showWebView = ref(false)
const showIframe = ref(false)
const showFallback = ref(false)

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 根据平台选择合适的显示方式
	// #ifdef H5
	showIframe.value = true
	// #endif
	
	// #ifdef APP-PLUS
	showWebView.value = true
	// #endif
	
	// #ifdef MP-WEIXIN || MP-ALIPAY || MP-BAIDU
	showFallback.value = true
	// #endif
})

onShow(() => {
	// 页面显示时的逻辑
})

// WebView事件
const onWebViewLoad = () => {
	console.log('WebView加载完成')
}

const onWebViewError = () => {
	console.log('WebView加载失败，切换到备用方案')
	showWebView.value = false
	showFallback.value = true
}

// iframe事件
const onIframeLoad = () => {
	console.log('iframe加载完成')
}

const onIframeError = () => {
	console.log('iframe加载失败，切换到备用方案')
	showIframe.value = false
	showFallback.value = true
}

// 返回功能
const goBack = () => {
	uni.navigateBack()
}

// 复制链接
const copyLink = () => {
	uni.setClipboardData({
		data: customerServiceUrl,
		success: () => {
			uni.showToast({
				title: '链接已复制',
				icon: 'success',
				duration: 2000
			})
		}
	})
}

// 在浏览器中打开
const openInBrowser = () => {
	// #ifdef H5
	window.open(customerServiceUrl, '_blank')
	// #endif
	
	// #ifdef APP-PLUS
	plus.runtime.openURL(customerServiceUrl)
	// #endif
	
	// #ifdef MP-WEIXIN || MP-ALIPAY || MP-BAIDU
	// 小程序环境下复制链接
	uni.showModal({
		title: '提示',
		content: '请复制链接到浏览器中打开客服页面',
		confirmText: '复制链接',
		success: (res) => {
			if (res.confirm) {
				copyLink()
			}
		}
	})
	// #endif
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
	background: #007AFF;
	flex-shrink: 0;
	z-index: 999;
	box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.1);
	
	.navbar-content {
		height: 88rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 30rpx;
	}
	
	.navbar-left {
		width: 60rpx;
		display: flex;
		align-items: center;
		justify-content: flex-start;
		padding: 8rpx;
		border-radius: 8rpx;
		transition: background-color 0.2s ease;
		
		&:active {
			background-color: rgba(255, 255, 255, 0.1);
		}
	}
	
	.nav-title {
		flex: 1;
		text-align: center;
		font-size: 32rpx;
		font-weight: 600;
		color: #fff;
	}
	
	.navbar-right {
		width: 60rpx;
	}
}

/* 客服内容区 */
.customer-service-content {
	flex: 1;
	position: relative;
	background: #fff;
}

/* WebView容器 */
.webview-container {
	width: 100%;
	height: 100%;
}

/* iframe容器 */
.iframe-container {
	width: 100%;
	height: 100%;
	position: relative;
	
	.customer-service-iframe {
		width: 100%;
		height: 100%;
		border: none;
	}
}

/* 备用内容 */
.fallback-content {
	padding: 40rpx 30rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 40rpx;
	
	.fallback-header {
		text-align: center;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16rpx;
		
		.fallback-title {
			font-size: 36rpx;
			font-weight: 600;
			color: #1a1a1a;
		}
		
		.fallback-subtitle {
			font-size: 26rpx;
			color: #666;
		}
	}
	
	.link-container {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 24rpx;
		
		.link-box {
			background: #f8f9fa;
			border: 2rpx solid #e9ecef;
			border-radius: 12rpx;
			padding: 20rpx 24rpx;
			display: flex;
			align-items: center;
			justify-content: space-between;
			transition: border-color 0.2s ease;
			
			&:active {
				border-color: #007AFF;
			}
			
			.link-text {
				flex: 1;
				font-size: 26rpx;
				color: #333;
				word-break: break-all;
				margin-right: 12rpx;
			}
		}
		
		.action-buttons {
			display: flex;
			gap: 20rpx;
			
			.copy-btn, .open-btn {
				flex: 1;
				display: flex;
				align-items: center;
				justify-content: center;
				gap: 8rpx;
				padding: 16rpx 20rpx;
				border-radius: 12rpx;
				font-size: 26rpx;
				font-weight: 500;
				transition: transform 0.2s ease;
				
				&:active {
					transform: scale(0.95);
				}
			}
			
			.copy-btn {
				background: #007AFF;
				color: #fff;
			}
			
			.open-btn {
				background: #28a745;
				color: #fff;
			}
			
			.btn-text {
				font-size: 26rpx;
				color: inherit;
			}
		}
	}
	
	.instruction {
		width: 100%;
		background: #f8f9fa;
		border-radius: 12rpx;
		padding: 20rpx;
		display: flex;
		flex-direction: column;
		gap: 8rpx;
		
		.instruction-title {
			font-size: 28rpx;
			font-weight: 600;
			color: #495057;
			margin-bottom: 8rpx;
		}
		
		.instruction-text {
			font-size: 26rpx;
			color: #6c757d;
			line-height: 1.5;
		}
	}
}
</style>