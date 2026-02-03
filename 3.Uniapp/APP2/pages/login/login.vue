<template>
	<view class="page">
		<!-- 检查登录状态时显示加载中 -->
		<view v-if="isChecking" class="checking-status">
			<view class="loading-spinner"></view>
		</view>
		
		<!-- 登录表单 -->
		<template v-else>
			<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
			
			<view class="login-container">
			<view class="logo-section">
				<image class="logo-img" src="/static/logocn.png" mode="aspectFit"></image>
				<text class="app-slogan">安全、简单的加密货币交易平台</text>
			</view>
			
			<view class="form-section">
				<view class="input-group">
					<view class="input-wrapper">
						<input 
							class="input" 
							type="text" 
							v-model="account" 
							placeholder="邮箱或手机号"
							placeholder-class="placeholder"
						/>
					</view>
				</view>
				
				<view class="input-group">
					<view class="input-wrapper">
						<input 
							class="input" 
							:type="showPassword ? 'text' : 'password'" 
							v-model="password" 
							placeholder="密码"
							placeholder-class="placeholder"
						/>
						<view class="eye-btn" @click="showPassword = !showPassword">
							<text class="eye-icon">{{ showPassword ? '👁' : '👁‍🗨' }}</text>
						</view>
					</view>
				</view>
				
				<view class="forgot-password" @click="handleForgotPassword">
					<text>忘记密码？</text>
				</view>
				
				<button class="login-btn" :disabled="!isFormValid" @click="handleLogin">
					登录
				</button>
			</view>
			
			<view class="register-section">
				<text class="register-text">还没有账户？</text>
				<text class="register-link" @click="goToRegister">立即注册</text>
			</view>
		</view>
		</template>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { login as apiLogin, getUserInfo } from '@/utils/api.js'

const statusBarHeight = ref(0)
const account = ref('')
const password = ref('')
const showPassword = ref(false)
const isLoading = ref(false)
const isChecking = ref(true) // 检查登录状态中

onMounted(async () => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 检查是否已登录
	await checkLoginStatus()
})

// 检查登录状态
const checkLoginStatus = async () => {
	const token = uni.getStorageSync('token')
	const isLoggedIn = uni.getStorageSync('isLoggedIn')
	
	if (token && isLoggedIn) {
		try {
			// 验证token是否有效
			const res = await getUserInfo()
			if (res && res.data) {
				// token有效，直接跳转到首页
				uni.switchTab({
					url: '/pages/index/index'
				})
				return
			}
		} catch (err) {
			// token已过期或无效，清除登录状态
			console.log('[Login] Token已过期:', err.message)
			uni.removeStorageSync('token')
			uni.removeStorageSync('isLoggedIn')
			uni.removeStorageSync('userInfo')
		}
	}
	
	isChecking.value = false
}

const isFormValid = computed(() => {
	return account.value.length > 0 && password.value.length >= 6
})

const handleLogin = async () => {
	if (!isFormValid.value || isLoading.value) return
	
	isLoading.value = true
	uni.showLoading({ title: '登录中...' })
	
	try {
		const res = await apiLogin({
			user_string: account.value,
			password: password.value,
			type: 1,
			area_code_id: 0
		})
		
		uni.hideLoading()
		
		const { token, user } = res.data
		uni.setStorageSync('token', token)
		uni.setStorageSync('isLoggedIn', true)
		uni.setStorageSync('userInfo', {
			id: user.id,
			account: user.account_number,
			nickname: user.email ? user.email.split('@')[0] : user.phone,
			phone: user.phone,
			email: user.email,
			avatar: user.head_portrait || ''
		})
		
		uni.showToast({
			title: '登录成功',
			icon: 'success',
			duration: 1500
		})
		
		setTimeout(() => {
			uni.switchTab({
				url: '/pages/index/index'
			})
		}, 1500)
	} catch (err) {
		uni.hideLoading()
		uni.showToast({
			title: err.message || '登录失败',
			icon: 'none',
			duration: 2000
		})
	} finally {
		isLoading.value = false
	}
}

const handleForgotPassword = () => {
	uni.showToast({
		title: '功能开发中',
		icon: 'none'
	})
}

const goToRegister = () => {
	uni.navigateTo({
		url: '/pages/register/register'
	})
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	min-height: 100vh;
	background: #fff;
}

.checking-status {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: #fff;
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 999;
	
	.loading-spinner {
		width: 60rpx;
		height: 60rpx;
		border: 4rpx solid #f0f0f0;
		border-top-color: #0052FF;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.status-bar {
	width: 100%;
	background: #fff;
}

.login-container {
	padding: 60rpx 40rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
}

.logo-section {
	display: flex;
	flex-direction: column;
	align-items: center;
	margin-bottom: 80rpx;
	
	.logo-img {
		height: 80rpx;
		max-width: 400rpx;
		margin-bottom: 24rpx;
	}
	
	.app-slogan {
		font-size: 26rpx;
		color: #5B616E;
	}
}

.form-section {
	width: 100%;
	
	.input-group {
		margin-bottom: 24rpx;
		
		.input-wrapper {
			height: 100rpx;
			background: #F7F8FA;
			border-radius: 16rpx;
			display: flex;
			align-items: center;
			padding: 0 24rpx;
			border: 2rpx solid transparent;
			
			&:focus-within {
				border-color: #0052FF;
				background: #fff;
			}
			
			.input {
				flex: 1;
				height: 100%;
				font-size: 30rpx;
				color: #0A0B0D;
			}
			
			.eye-btn {
				width: 60rpx;
				height: 60rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				
				.eye-icon {
					font-size: 36rpx;
				}
			}
		}
	}
	
	.forgot-password {
		text-align: right;
		margin-bottom: 40rpx;
		
		text {
			font-size: 26rpx;
			color: #0052FF;
		}
	}
	
	.login-btn {
		width: 100%;
		height: 100rpx;
		background: #0052FF;
		border-radius: 16rpx;
		color: #fff;
		font-size: 32rpx;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		
		&[disabled] {
			background: #E8E8E8;
			color: #9AA0AA;
		}
	}
}

.placeholder {
	color: #9AA0AA;
}

.register-section {
	margin-top: 48rpx;
	display: flex;
	align-items: center;
	gap: 8rpx;
	
	.register-text {
		font-size: 28rpx;
		color: #5B616E;
	}
	
	.register-link {
		font-size: 28rpx;
		color: #0052FF;
		font-weight: 500;
	}
}
</style>
