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
			<!-- 加载中 -->
			<view class="loading-container" v-if="loading">
				<view class="loading-spinner"></view>
				<text class="loading-text">加载中...</text>
			</view>

			<!-- 已通过认证 -->
			<view class="kyc-container" v-else-if="kycStatus === 2">
				<view class="status-card success">
					<view class="status-icon">✓</view>
					<view class="status-title">实名认证已通过</view>
					<view class="status-desc">您的身份信息已验证成功</view>
				</view>
				<view class="info-card">
					<view class="info-row">
						<text class="info-label">真实姓名</text>
						<text class="info-value">{{ kycInfo.name }}</text>
					</view>
					<view class="info-row">
						<text class="info-label">身份证号</text>
						<text class="info-value">{{ kycInfo.card_id_masked }}</text>
					</view>
					<view class="info-row">
						<text class="info-label">认证时间</text>
						<text class="info-value">{{ formatTime(kycInfo.review_time) }}</text>
					</view>
				</view>
			</view>

			<!-- 审核中 -->
			<view class="kyc-container" v-else-if="kycStatus === 1">
				<view class="status-card pending">
					<view class="status-icon">⏳</view>
					<view class="status-title">实名认证审核中</view>
					<view class="status-desc">您的认证信息正在审核，请耐心等待</view>
					<view class="status-time">提交时间: {{ formatTime(kycInfo.submit_time) }}</view>
				</view>
				<view class="tips-card">
					<text class="tips-text">• 审核一般在1-3个工作日内完成</text>
					<text class="tips-text">• 审核结果将通过站内信通知您</text>
					<text class="tips-text">• 如有疑问请联系在线客服</text>
				</view>
			</view>

			<!-- 被拒绝 -->
			<view class="kyc-container" v-else-if="kycStatus === 3">
				<view class="status-card rejected">
					<view class="status-icon">✕</view>
					<view class="status-title">实名认证未通过</view>
					<view class="status-desc">{{ kycInfo.reject_reason || '您提交的信息未通过审核' }}</view>
				</view>
				<button class="resubmit-btn" @click="showForm = true">重新提交</button>
			</view>

			<!-- 未提交/重新提交表单 -->
			<view class="kyc-container" v-if="kycStatus === 0 || showForm">
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
						<text class="input-error" v-if="formData.idCard && !isIdCardValid">身份证号格式不正确</text>
					</view>
					<view class="input-group">
						<text class="label">手机号</text>
						<input class="input" type="tel" v-model="formData.phone" placeholder="请输入11位手机号码" maxlength="11" />
					</view>
					<!-- 隐藏提款地址，提交时使用占位值 -->
					<view class="input-group" v-if="false">
						<text class="label">提款地址</text>
						<input class="input" type="text" v-model="formData.bankCard" placeholder="请输入您的提款地址（支持TRC、BRC等主流网络）" />
					</view>
				</view>

				<!-- 证件上传区域 -->
				<view class="upload-section">
					<view class="section-label">上传证件照片</view>
					
					<view class="upload-grid">
						<!-- 身份证正面 -->
						<view class="upload-item" @click="chooseImage('idFront')">
							<view class="upload-box" :class="{ 'has-img': formData.idFront, 'uploading': uploadingType === 'idFront' }">
								<image v-if="formData.idFront" :src="formData.idFront" mode="aspectFit" class="preview-img"></image>
								<view v-else class="upload-placeholder">
									<SvgIcon name="kyc" :size="32" color="#0052FF" />
									<text class="upload-text">身份证正面</text>
								</view>
								<view v-if="uploadingType === 'idFront'" class="upload-loading">
									<view class="loading-spinner small"></view>
								</view>
							</view>
						</view>

						<!-- 身份证反面 -->
						<view class="upload-item" @click="chooseImage('idBack')">
							<view class="upload-box" :class="{ 'has-img': formData.idBack, 'uploading': uploadingType === 'idBack' }">
								<image v-if="formData.idBack" :src="formData.idBack" mode="aspectFit" class="preview-img"></image>
								<view v-else class="upload-placeholder">
									<SvgIcon name="kyc" :size="32" color="#0052FF" />
									<text class="upload-text">身份证反面</text>
								</view>
								<view v-if="uploadingType === 'idBack'" class="upload-loading">
									<view class="loading-spinner small"></view>
								</view>
							</view>
						</view>
					</view>
				</view>

				<!-- 提交按钮 -->
				<view class="submit-btn-box">
					<button class="submit-btn" :disabled="!isFormValid || submitting" @click="handleSubmit">
						{{ submitting ? '提交中...' : '提交认证' }}
					</button>
					<text class="security-tips">您的信息已通过 256 位加密处理</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { request, BASE_URL } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'

const statusBarHeight = ref(0)
const loading = ref(true)
const submitting = ref(false)
const uploadingType = ref('')
const showForm = ref(false)

// KYC状态: 0=未提交 1=待审核 2=已通过 3=已拒绝
const kycStatus = ref(0)
const kycInfo = ref({})

// 表单数据
const formData = ref({
	name: '',
	idCard: '',
	phone: '',
	bankCard: '',
	idFront: '',
	idBack: '',
	bankCardImg: ''
})

// 上传后的图片URL
const uploadedUrls = ref({
	idFront: '',
	idBack: '',
	bankCardImg: ''
})

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	fetchKYCStatus()
})

// 获取KYC状态
const fetchKYCStatus = async () => {
	loading.value = true
	try {
		const res = await request({
			url: '/kyc/status',
			method: 'GET'
		})
		if (res.type === 'ok' && res.data) {
			kycStatus.value = res.data.review_status || 0
			kycInfo.value = res.data
		}
	} catch (err) {
		// 未找到记录表示未提交
		kycStatus.value = 0
	} finally {
		loading.value = false
	}
}

// 身份证号校验（基本格式检查）
const isIdCardValid = computed(() => {
	const id = formData.value.idCard
	if (!id) return false
	// 只检查基本格式：15位或18位数字，18位最后一位可以是X
	const pattern = /^(\d{15}|\d{17}[\dXx])$/
	return pattern.test(id)
})

// 表单验证
const isFormValid = computed(() => {
	const { name, idCard, phone } = formData.value
	const { idFront, idBack } = uploadedUrls.value
	const phoneValid = /^1[3-9]\d{9}$/.test(phone)
	return name && isIdCardValid.value && phoneValid && idFront && idBack
})

const goBack = () => {
	uni.navigateBack()
}

// 选择并上传图片
const chooseImage = (type) => {
	if (uploadingType.value) return
	
	uni.chooseImage({
		count: 1,
		sizeType: ['compressed'],
		sourceType: ['album', 'camera'],
		success: async (res) => {
			const tempPath = res.tempFilePaths[0]
			formData.value[type] = tempPath
			
			// 上传图片
			uploadingType.value = type
			try {
				const uploadRes = await uploadImage(tempPath)
				if (uploadRes) {
					uploadedUrls.value[type] = uploadRes
				} else {
					uni.showToast({ title: '图片上传失败', icon: 'none' })
					formData.value[type] = ''
				}
			} catch (err) {
				console.error('上传失败:', err)
				uni.showToast({ title: '图片上传失败', icon: 'none' })
				formData.value[type] = ''
			} finally {
				uploadingType.value = ''
			}
		}
	})
}

// 上传图片到服务器
const uploadImage = (filePath) => {
	return new Promise((resolve, reject) => {
		const token = uni.getStorageSync('token')
		uni.uploadFile({
			url: BASE_URL + '/upload',
			filePath: filePath,
			name: 'file',
			header: {
				'Authorization': token ? `Bearer ${token}` : ''
			},
			success: (res) => {
				try {
					const data = JSON.parse(res.data)
					if (data.type === 'ok' && data.data && data.data.url) {
						resolve(data.data.url)
					} else {
						resolve(null)
					}
				} catch (e) {
					resolve(null)
				}
			},
			fail: (err) => {
				reject(err)
			}
		})
	})
}

// 格式化时间
const formatTime = (timestamp) => {
	if (!timestamp) return '-'
	const date = new Date(timestamp * 1000)
	return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

// 提交KYC
const handleSubmit = async () => {
	if (!isFormValid.value || submitting.value) return
	
	submitting.value = true
	try {
		const res = await request({
			url: '/kyc/submit',
			method: 'POST',
			data: {
				name: formData.value.name,
				card_id: formData.value.idCard,
				phone: formData.value.phone,
				bank_card: formData.value.bankCard || 'pending_verification',
				front_pic: uploadedUrls.value.idFront,
				reverse_pic: uploadedUrls.value.idBack,
				bank_pic: ''
			}
		})
		
		if (res.type === 'ok') {
			uni.showToast({
				title: '提交成功，请等待审核',
				icon: 'success',
				duration: 2000
			})
			// 刷新状态
			setTimeout(() => {
				showForm.value = false
				fetchKYCStatus()
			}, 1500)
		} else {
			uni.showToast({ title: res.message || '提交失败', icon: 'none' })
		}
	} catch (err) {
		console.error('提交失败:', err)
		uni.showToast({ title: err.message || '提交失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #F7F8FA;
	display: flex;
	flex-direction: column;
}

.custom-navbar {
	background: #FFFFFF;
	flex-shrink: 0;
	z-index: 999;
	border-bottom: 1px solid #E8E8E8;
	
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
		font-weight: 600;
		color: #1A1A1A;
	}
}

.content-scroll {
	flex: 1;
	overflow: hidden;
}

/* 加载状态 */
.loading-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 100rpx 0;
}

.loading-spinner {
	width: 60rpx;
	height: 60rpx;
	border: 4rpx solid #E8E8E8;
	border-top-color: #0052FF;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
	
	&.small {
		width: 40rpx;
		height: 40rpx;
		border-width: 3rpx;
	}
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.loading-text {
	margin-top: 20rpx;
	font-size: 28rpx;
	color: #666;
}

.kyc-container {
	padding: 40rpx 30rpx;
}

/* 状态卡片 */
.status-card {
	background: #FFFFFF;
	border-radius: 24rpx;
	padding: 60rpx 40rpx;
	text-align: center;
	margin-bottom: 30rpx;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.04);
	
	.status-icon {
		width: 100rpx;
		height: 100rpx;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 48rpx;
		margin: 0 auto 24rpx;
	}
	
	.status-title {
		font-size: 36rpx;
		font-weight: 600;
		margin-bottom: 16rpx;
	}
	
	.status-desc {
		font-size: 28rpx;
		color: #666;
		line-height: 1.5;
	}
	
	.status-time {
		font-size: 24rpx;
		color: #999;
		margin-top: 20rpx;
	}
	
	&.success {
		.status-icon {
			background: #E8F5E9;
			color: #4CAF50;
		}
		.status-title {
			color: #4CAF50;
		}
	}
	
	&.pending {
		.status-icon {
			background: #FFF3E0;
			color: #FF9800;
		}
		.status-title {
			color: #FF9800;
		}
	}
	
	&.rejected {
		.status-icon {
			background: #FFEBEE;
			color: #F44336;
		}
		.status-title {
			color: #F44336;
		}
	}
}

/* 信息卡片 */
.info-card {
	background: #FFFFFF;
	border-radius: 16rpx;
	padding: 20rpx 30rpx;
	
	.info-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24rpx 0;
		border-bottom: 1rpx solid #F0F0F0;
		
		&:last-child {
			border-bottom: none;
		}
	}
	
	.info-label {
		font-size: 28rpx;
		color: #666;
	}
	
	.info-value {
		font-size: 28rpx;
		color: #1A1A1A;
		font-weight: 500;
	}
}

/* 提示卡片 */
.tips-card {
	background: #F0F6FF;
	border-radius: 16rpx;
	padding: 30rpx;
	
	.tips-text {
		display: block;
		font-size: 26rpx;
		color: #0052FF;
		line-height: 2;
	}
}

/* 重新提交按钮 */
.resubmit-btn {
	width: 100%;
	height: 96rpx;
	background: #0052FF;
	color: #FFFFFF;
	border-radius: 48rpx;
	font-size: 32rpx;
	font-weight: 600;
	border: none;
	margin-top: 30rpx;
}

/* 表单区域 */
.header-tips {
	margin-bottom: 40rpx;
	
	.tips-title {
		font-size: 48rpx;
		font-weight: bold;
		color: #1A1A1A;
		margin-bottom: 16rpx;
	}
	
	.tips-desc {
		font-size: 26rpx;
		color: #666;
		line-height: 1.5;
	}
}

.section-label {
	font-size: 32rpx;
	font-weight: 600;
	color: #1A1A1A;
	margin-bottom: 24rpx;
	padding-left: 16rpx;
	border-left: 8rpx solid #0052FF;
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
			color: #666;
			margin-bottom: 16rpx;
		}
		
		.input {
			height: 88rpx;
			background: #F7F8FA;
			border: 2rpx solid #E8E8E8;
			border-radius: 12rpx;
			padding: 0 24rpx;
			font-size: 28rpx;
			color: #1A1A1A;
			
			&:focus {
				border-color: #0052FF;
				background: #FFFFFF;
			}
		}
		
		.input-error {
			display: block;
			font-size: 24rpx;
			color: #F44336;
			margin-top: 10rpx;
			padding-left: 10rpx;
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
			border: 2rpx dashed #D0D0D0;
			border-radius: 24rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			overflow: hidden;
			position: relative;
			
			&.has-img {
				border-style: solid;
				border-color: #0052FF;
			}
			
			&.uploading {
				opacity: 0.7;
			}
			
			.upload-placeholder {
				display: flex;
				flex-direction: column;
				align-items: center;
				
				.upload-text {
					font-size: 24rpx;
					color: #0052FF;
					margin-top: 16rpx;
				}
			}
			
			.preview-img {
				width: 100%;
				height: 100%;
			}
			
			.upload-loading {
				position: absolute;
				top: 0;
				left: 0;
				right: 0;
				bottom: 0;
				background: rgba(255, 255, 255, 0.8);
				display: flex;
				align-items: center;
				justify-content: center;
			}
		}
	}
}

.submit-btn-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 24rpx;
	padding-bottom: 60rpx;
	
	.submit-btn {
		width: 100%;
		height: 100rpx;
		background: #0052FF;
		color: #FFFFFF;
		border-radius: 50rpx;
		font-size: 32rpx;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		box-shadow: 0 8rpx 24rpx rgba(0, 82, 255, 0.3);
		
		&[disabled] {
			background: #D0D0D0;
			color: #999;
			box-shadow: none;
		}
		
		&:active:not([disabled]) {
			opacity: 0.9;
			transform: scale(0.99);
		}
	}
	
	.security-tips {
		font-size: 22rpx;
		color: #999;
	}
}
</style>
