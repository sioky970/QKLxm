<template>
	<view class="page">
		<!-- 自定义顶部导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="navbar-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<view class="navbar-title">实名认证</view>
				<view class="navbar-right"></view>
			</view>
		</view>

		<scroll-view scroll-y class="content-scroll">
			<view class="kyc-container">
				<!-- 头部提示 -->
				<view class="header-tips">
					<view class="tips-title">身份验证</view>
					<view class="tips-desc">请确保填写的信息真实有效，我们将严格保护您的个人隐私。</view>
				</view>

				<!-- 信息填写表单 -->
				<view class="form-section">
					<view class="section-label">基本信息</view>
					<view class="input-group">
						<text class="label">姓名</text>
						<input class="input" type="text" v-model="formData.name" placeholder="请输入您的真实姓名" />
					</view>
					<view class="input-group">
						<text class="label">身份证号</text>
						<input class="input" type="idcard" v-model="formData.idCard" placeholder="请输入18位身份证号码" maxlength="18" />
					</view>
					<view class="input-group">
						<text class="label">手机号</text>
						<input class="input" type="tel" v-model="formData.phone" placeholder="请输入11位手机号码" maxlength="11" />
					</view>
					<view class="input-group">
						<text class="label">银行卡号</text>
						<input class="input" type="number" v-model="formData.bankCard" placeholder="请输入您的银行储蓄卡号" />
					</view>
				</view>

				<!-- 证件上传区域 -->
				<view class="upload-section">
					<view class="section-label">上传证件照片</view>
					
					<view class="upload-grid">
						<!-- 身份证正面 -->
						<view class="upload-item" @click="chooseImage('idFront')">
							<view class="upload-box" :class="{ 'has-img': formData.idFront }">
								<image v-if="formData.idFront" :src="formData.idFront" mode="aspectFit" class="preview-img"></image>
								<view v-else class="upload-placeholder">
									<SvgIcon name="kyc" :size="32" color="#3B82F6" />
									<text class="upload-text">身份证正面</text>
								</view>
							</view>
						</view>

						<!-- 身份证反面 -->
						<view class="upload-item" @click="chooseImage('idBack')">
							<view class="upload-box" :class="{ 'has-img': formData.idBack }">
								<image v-if="formData.idBack" :src="formData.idBack" mode="aspectFit" class="preview-img"></image>
								<view v-else class="upload-placeholder">
									<SvgIcon name="kyc" :size="32" color="#3B82F6" />
									<text class="upload-text">身份证反面</text>
								</view>
							</view>
						</view>

						<!-- 银行卡照片 -->
						<view class="upload-item full-width" @click="chooseImage('bankCardImg')">
							<view class="upload-box" :class="{ 'has-img': formData.bankCardImg }">
								<image v-if="formData.bankCardImg" :src="formData.bankCardImg" mode="aspectFit" class="preview-img"></image>
								<view v-else class="upload-placeholder">
									<SvgIcon name="kyc" :size="32" color="#3B82F6" />
									<text class="upload-text">银行卡正面照片</text>
								</view>
							</view>
						</view>
					</view>
				</view>

				<!-- 提交按钮 -->
				<view class="submit-btn-box">
					<button class="submit-btn" :disabled="!isFormValid" @click="handleSubmit">提交认证</button>
					<text class="security-tips">
						您的信息已通过 256 位加密处理
					</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import SvgIcon from '@/components/SvgIcon.vue'

const statusBarHeight = ref(0)

const formData = ref({
	name: '',
	idCard: '',
	phone: '',
	bankCard: '',
	idFront: '',
	idBack: '',
	bankCardImg: ''
})

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
})

const isFormValid = computed(() => {
	const { name, idCard, phone, bankCard, idFront, idBack, bankCardImg } = formData.value
	// 验证手机号：11位数字，以1开头
	const phoneValid = /^1[3-9]\d{9}$/.test(phone)
	return name && idCard.length === 18 && phoneValid && bankCard && idFront && idBack && bankCardImg
})

const goBack = () => {
	uni.navigateBack()
}

const chooseImage = (type) => {
	uni.chooseImage({
		count: 1,
		sizeType: ['compressed'],
		sourceType: ['album', 'camera'],
		success: (res) => {
			formData.value[type] = res.tempFilePaths[0]
		}
	})
}

const handleSubmit = () => {
	uni.showLoading({ title: '提交中...' })
	setTimeout(() => {
		uni.hideLoading()
		uni.showToast({
			title: '提交成功，请等待审核',
			icon: 'success',
			duration: 2000
		})
		setTimeout(() => {
			uni.navigateBack()
		}, 2000)
	}, 1500)
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #F8FAFC;
	display: flex;
	flex-direction: column;
}

.custom-navbar {
	background: #FFFFFF;
	flex-shrink: 0;
	z-index: 999;
	
	.navbar-content {
		height: 88rpx;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 30rpx;
	}
	
	.navbar-left, .navbar-right {
		width: 60rpx;
	}
	
	.navbar-title {
		font-size: 32rpx;
		font-weight: bold;
		color: #333;
	}
}

.content-scroll {
	flex: 1;
	overflow: hidden;
}

.kyc-container {
	padding: 40rpx 30rpx;
}

.header-tips {
	margin-bottom: 40rpx;
	
	.tips-title {
		font-size: 48rpx;
		font-weight: bold;
		color: #000;
		margin-bottom: 16rpx;
	}
	
	.tips-desc {
		font-size: 26rpx;
		color: #64748B;
		line-height: 1.5;
	}
}

.section-label {
	font-size: 32rpx;
	font-weight: bold;
	color: #000;
	margin-bottom: 24rpx;
	padding-left: 16rpx;
	border-left: 8rpx solid #3B82F6;
}

.form-section {
	background: #FFFFFF;
	border-radius: 24rpx;
	padding: 30rpx;
	margin-bottom: 40rpx;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.02);
	
	.input-group {
		margin-bottom: 30rpx;
		
		&:last-child {
			margin-bottom: 0;
		}
		
		.label {
			display: block;
			font-size: 26rpx;
			color: #64748B;
			margin-bottom: 16rpx;
		}
		
		.input {
			height: 88rpx;
			background: #F1F5F9;
			border-radius: 12rpx;
			padding: 0 24rpx;
			font-size: 28rpx;
			color: #333;
		}
	}
}

.upload-section {
	margin-bottom: 60rpx;
	
	.upload-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 20rpx;
		
		.upload-item {
			width: calc(50% - 10rpx);
			
			&.full-width {
				width: 100%;
			}
		}
		
		.upload-box {
			height: 240rpx;
			background: #FFFFFF;
			border: 2rpx dashed #E2E8F0;
			border-radius: 24rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			overflow: hidden;
			
			&.has-img {
				border-style: solid;
				border-color: #3B82F6;
			}
			
			.upload-placeholder {
				display: flex;
				flex-direction: column;
				align-items: center;
				
				.upload-text {
					font-size: 24rpx;
					color: #3B82F6;
					margin-top: 16rpx;
				}
			}
			
			.preview-img {
				width: 100%;
				height: 100%;
			}
		}
	}
}

.submit-btn-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 24rpx;
	
	.submit-btn {
		width: 100%;
		height: 100rpx;
		background: #3B82F6;
		color: #FFFFFF;
		border-radius: 50rpx;
		font-size: 32rpx;
		font-weight: bold;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		
		&[disabled] {
			background: #CBD5E1;
			color: #94A3B8;
		}
		
		&:active {
			opacity: 0.9;
		}
	}
	
	.security-tips {
		display: flex;
		align-items: center;
		gap: 8rpx;
		font-size: 22rpx;
		color: #94A3B8;
	}
}
</style>
