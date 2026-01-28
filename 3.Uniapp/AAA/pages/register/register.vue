<template>
	<view class="page">
		<!-- 状态栏占位 -->
		<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
		
		<!-- 顶部导航 -->
		<view class="custom-navbar">
			<view class="navbar-content">
				<view class="navbar-left" @click="goBack">
					<text class="back-icon">←</text>
				</view>
				<view class="navbar-title">创建账户</view>
				<view class="navbar-right"></view>
			</view>
		</view>
		
		<scroll-view scroll-y class="content-scroll">
			<view class="register-container">
				<!-- 头部说明 -->
				<view class="header-section">
					<text class="header-title">欢迎加入</text>
					<text class="header-desc">创建您的账户，开始加密货币交易之旅</text>
				</view>
				
				<!-- 标签页切换 -->
				<view class="tab-container">
					<view 
						class="tab-item" 
						:class="{ active: activeTab === 'email' }"
						@click="activeTab = 'email'"
					>
						<text>邮箱注册</text>
					</view>
					<view 
						class="tab-item" 
						:class="{ active: activeTab === 'phone' }"
						@click="activeTab = 'phone'"
					>
						<text>手机号注册</text>
					</view>
				</view>
				
				<!-- 邮箱注册表单 -->
				<view class="form-section" v-if="activeTab === 'email'">
					<view class="input-group">
						<text class="label">邮箱地址</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								type="text" 
								v-model="emailForm.email" 
								placeholder="请输入您的邮箱"
								placeholder-class="placeholder"
							/>
						</view>
					</view>
					
					<view class="input-group">
						<text class="label">设置密码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								:type="showPassword ? 'text' : 'password'" 
								v-model="emailForm.password" 
								placeholder="请输入至少6位密码"
								placeholder-class="placeholder"
							/>
							<view class="eye-btn" @click="showPassword = !showPassword">
								<text class="eye-icon">{{ showPassword ? '👁' : '👁‍🗨' }}</text>
							</view>
						</view>
					</view>
					
					<view class="input-group">
						<text class="label">确认密码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								:type="showConfirmPassword ? 'text' : 'password'" 
								v-model="emailForm.confirmPassword" 
								placeholder="请再次输入密码"
								placeholder-class="placeholder"
							/>
							<view class="eye-btn" @click="showConfirmPassword = !showConfirmPassword">
								<text class="eye-icon">{{ showConfirmPassword ? '👁' : '👁‍🗨' }}</text>
							</view>
						</view>
					</view>
					
					<view class="input-group" v-if="inviteCodeRequired">
						<text class="label">邀请码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								type="text" 
								v-model="emailForm.inviteCode" 
								placeholder="请输入邀请码"
								placeholder-class="placeholder"
							/>
						</view>
					</view>
				</view>
				
				<!-- 手机号注册表单 -->
				<view class="form-section" v-if="activeTab === 'phone'">
					<view class="input-group">
						<text class="label">手机号码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								type="tel" 
								v-model="phoneForm.phone" 
								placeholder="请输入11位手机号"
								placeholder-class="placeholder"
								maxlength="11"
							/>
						</view>
					</view>
					
					<view class="input-group">
						<text class="label">设置密码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								:type="showPassword ? 'text' : 'password'" 
								v-model="phoneForm.password" 
								placeholder="请输入至少6位密码"
								placeholder-class="placeholder"
							/>
							<view class="eye-btn" @click="showPassword = !showPassword">
								<text class="eye-icon">{{ showPassword ? '👁' : '👁‍🗨' }}</text>
							</view>
						</view>
					</view>
					
					<view class="input-group">
						<text class="label">确认密码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								:type="showConfirmPassword ? 'text' : 'password'" 
								v-model="phoneForm.confirmPassword" 
								placeholder="请再次输入密码"
								placeholder-class="placeholder"
							/>
							<view class="eye-btn" @click="showConfirmPassword = !showConfirmPassword">
								<text class="eye-icon">{{ showConfirmPassword ? '👁' : '👁‍🗨' }}</text>
							</view>
						</view>
					</view>
					
					<view class="input-group" v-if="inviteCodeRequired">
						<text class="label">邀请码</text>
						<view class="input-wrapper">
							<input 
								class="input" 
								type="text" 
								v-model="phoneForm.inviteCode" 
								placeholder="请输入邀请码"
								placeholder-class="placeholder"
							/>
						</view>
					</view>
				</view>
				
				<!-- 协议勾选 -->
				<view class="agreement-check" @click="agreeTerms = !agreeTerms">
					<view class="checkbox" :class="{ checked: agreeTerms }">
						<text v-if="agreeTerms">✓</text>
					</view>
					<text class="agreement-text">
						我已阅读并同意
						<text class="link" @click.stop="showAgreementModal('terms')">《服务条款》</text>
						和
						<text class="link" @click.stop="showAgreementModal('privacy')">《隐私政策》</text>
					</text>
				</view>
				
				<!-- 注册按钮 -->
				<button class="register-btn" :disabled="!isFormValid" @click="handleRegister">
					注册
				</button>
				
				<!-- 登录入口 -->
				<view class="login-section">
					<text class="login-text">已有账户？</text>
					<text class="login-link" @click="goToLogin">立即登录</text>
				</view>
			</view>
		</scroll-view>
		
		<!-- 协议弹窗 -->
		<view class="modal-mask" v-if="showModal" @click="closeModal">
			<view class="modal-content" @click.stop>
				<view class="modal-header">
					<text class="modal-title">{{ modalTitle }}</text>
					<view class="modal-close" @click="closeModal">
						<text>×</text>
					</view>
				</view>
				<scroll-view scroll-y class="modal-body">
					<view class="agreement-content" v-if="modalType === 'terms'">
						<text class="section-title">Terms of Service</text>
						<text class="section-text">Last Updated: January 2026</text>
						<text class="section-text">Welcome to Coinbase. By accessing or using our services, you agree to be bound by these Terms of Service.</text>
						
						<text class="section-title">1. Acceptance of Terms</text>
						<text class="section-text">By creating an account or using any Coinbase service, you agree to these Terms. If you do not agree, please do not use our services.</text>
						
						<text class="section-title">2. Eligibility</text>
						<text class="section-text">You must be at least 18 years old and capable of forming a binding contract to use our services. By using Coinbase, you represent that you meet these requirements.</text>
						
						<text class="section-title">3. Account Registration</text>
						<text class="section-text">To access certain features, you must register for an account. You agree to provide accurate, current, and complete information and to update such information to keep it accurate.</text>
						
						<text class="section-title">4. Prohibited Activities</text>
						<text class="section-text">You agree not to engage in any activity that interferes with or disrupts the services, violates any laws, or infringes on the rights of others.</text>
						
						<text class="section-title">5. Fees and Payments</text>
						<text class="section-text">Coinbase charges fees for certain transactions. All fees are disclosed before you complete a transaction and are subject to change.</text>
						
						<text class="section-title">6. Risk Disclosure</text>
						<text class="section-text">Cryptocurrency trading involves substantial risk of loss. You should carefully consider whether trading is appropriate for you in light of your financial condition.</text>
						
						<text class="section-title">7. Limitation of Liability</text>
						<text class="section-text">To the maximum extent permitted by law, Coinbase shall not be liable for any indirect, incidental, special, consequential, or punitive damages.</text>
						
						<text class="section-title">8. Contact Us</text>
						<text class="section-text">If you have any questions about these Terms, please contact us at support@coinbase.com</text>
					</view>
					
					<view class="agreement-content" v-else>
						<text class="section-title">Privacy Policy</text>
						<text class="section-text">Last Updated: January 2026</text>
						<text class="section-text">This Privacy Policy describes how Coinbase collects, uses, and shares information about you.</text>
						
						<text class="section-title">1. Information We Collect</text>
						<text class="section-text">We collect information you provide directly, such as when you create an account, make a transaction, or contact us for support. This includes your name, email, phone number, and identity verification documents.</text>
						
						<text class="section-title">2. How We Use Information</text>
						<text class="section-text">We use the information we collect to provide, maintain, and improve our services, process transactions, send you technical notices and support messages, and respond to your requests.</text>
						
						<text class="section-title">3. Information Sharing</text>
						<text class="section-text">We do not share your personal information with third parties except as described in this policy. We may share information with service providers, for legal compliance, or in connection with a merger or acquisition.</text>
						
						<text class="section-title">4. Data Security</text>
						<text class="section-text">We implement appropriate technical and organizational measures to protect the security of your personal information. All data is encrypted using 256-bit SSL encryption.</text>
						
						<text class="section-title">5. Data Retention</text>
						<text class="section-text">We retain your information for as long as your account is active or as needed to provide you services, comply with legal obligations, and resolve disputes.</text>
						
						<text class="section-title">6. Your Rights</text>
						<text class="section-text">You have the right to access, correct, or delete your personal information. You may also object to or restrict certain processing of your information.</text>
						
						<text class="section-title">7. Cookies and Tracking</text>
						<text class="section-text">We use cookies and similar technologies to collect information about your browsing activities and to personalize your experience.</text>
						
						<text class="section-title">8. Contact Us</text>
						<text class="section-text">For privacy-related inquiries, please contact our Data Protection Officer at privacy@coinbase.com</text>
					</view>
				</scroll-view>
				<view class="modal-footer">
					<button class="modal-btn" @click="closeModal">I Understand</button>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getRegisterConfig, register as apiRegister, login as apiLogin } from '@/utils/api.js'

const statusBarHeight = ref(0)
const activeTab = ref('email')
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const agreeTerms = ref(true)
const showModal = ref(false)
const modalType = ref('terms')
const modalTitle = ref('')
const isLoading = ref(false)
const inviteCodeRequired = ref(false) // 邀请码是否必填

// 邮箱注册表单
const emailForm = ref({
	email: '',
	password: '',
	confirmPassword: '',
	inviteCode: ''
})

// 手机号注册表单
const phoneForm = ref({
	phone: '',
	password: '',
	confirmPassword: '',
	inviteCode: ''
})

onMounted(async () => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	
	// 获取注册配置
	try {
		const res = await getRegisterConfig()
		if (res.data) {
			inviteCodeRequired.value = res.data.invite_code_required || false
		}
	} catch (err) {
		console.error('获取注册配置失败:', err)
	}
})

// 表单验证
const isFormValid = computed(() => {
	if (!agreeTerms.value) return false
	
	if (activeTab.value === 'email') {
		const { email, password, confirmPassword, inviteCode } = emailForm.value
		const emailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
		const passwordValid = password.length >= 6
		const passwordMatch = password === confirmPassword && confirmPassword.length > 0
		// 如果邀请码必填，则需要验证邀请码
		const inviteValid = !inviteCodeRequired.value || inviteCode.length > 0
		return emailValid && passwordValid && passwordMatch && inviteValid
	} else {
		const { phone, password, confirmPassword, inviteCode } = phoneForm.value
		const phoneValid = /^1[3-9]\d{9}$/.test(phone)
		const passwordValid = password.length >= 6
		const passwordMatch = password === confirmPassword && confirmPassword.length > 0
		// 如果邀请码必填，则需要验证邀请码
		const inviteValid = !inviteCodeRequired.value || inviteCode.length > 0
		return phoneValid && passwordValid && passwordMatch && inviteValid
	}
})

const goBack = () => {
	uni.navigateBack()
}

const handleRegister = async () => {
	if (!isFormValid.value || isLoading.value) return
	
	isLoading.value = true
	uni.showLoading({ title: '注册中...' })
	
	try {
		let registerData = {}
		
		if (activeTab.value === 'email') {
			registerData = {
				type: 'email',
				user_string: emailForm.value.email,
				password: emailForm.value.password,
				re_password: emailForm.value.confirmPassword,
				code: '000000', // 使用测试验证码跳过验证
				extension_code: emailForm.value.inviteCode || '',
				country_code: 0
			}
		} else {
			registerData = {
				type: 'mobile',
				user_string: phoneForm.value.phone,
				password: phoneForm.value.password,
				re_password: phoneForm.value.confirmPassword,
				code: '000000', // 使用测试验证码跳过验证
				extension_code: phoneForm.value.inviteCode || '',
				country_code: 0
			}
		}
		
		await apiRegister(registerData)
		
		uni.hideLoading()
		uni.showToast({
			title: '注册成功',
			icon: 'success',
			duration: 1500
		})
		
		// 注册成功后自动登录
		setTimeout(async () => {
			try {
				// 使用注册时的账号密码自动登录
				const loginData = {
					user_string: activeTab.value === 'email' ? emailForm.value.email : phoneForm.value.phone,
					password: activeTab.value === 'email' ? emailForm.value.password : phoneForm.value.password,
					type: 1,
					area_code_id: 0
				}
				
				const loginRes = await apiLogin(loginData)
				
				// 保存登录信息
				const { token, user } = loginRes.data
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
				
				// 跳转到首页
				uni.reLaunch({
					url: '/pages/index/index'
				})
			} catch (loginErr) {
				console.error('自动登录失败:', loginErr)
				// 登录失败则返回登录页
				uni.navigateBack()
			}
		}, 1500)
	} catch (err) {
		console.error('注册失败详情:', err)
		uni.hideLoading()
		
		// 提取错误信息
		let errorMsg = '注册失败'
		if (err.message) {
			errorMsg = err.message
		} else if (typeof err === 'string') {
			errorMsg = err
		}
		
		// 特殊处理常见错误
		if (errorMsg.includes('Duplicate entry') || errorMsg.includes('account_number')) {
			errorMsg = '该账号已被注册，请更换手机号或邮箱'
		}
		
		uni.showToast({
			title: errorMsg,
			icon: 'none',
			duration: 3000
		})
	} finally {
		isLoading.value = false
	}
}

const goToLogin = () => {
	uni.navigateBack()
}

const showAgreementModal = (type) => {
	modalType.value = type
	modalTitle.value = type === 'terms' ? 'Terms of Service' : 'Privacy Policy'
	showModal.value = true
}

const closeModal = () => {
	showModal.value = false
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	min-height: 100vh;
	background: #fff;
	display: flex;
	flex-direction: column;
}

.status-bar {
	width: 100%;
	background: #fff;
}

.custom-navbar {
	background: #fff;
	flex-shrink: 0;
	
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
	
	.back-icon {
		font-size: 40rpx;
		color: #0A0B0D;
	}
	
	.navbar-title {
		font-size: 32rpx;
		font-weight: 600;
		color: #0A0B0D;
	}
}

.content-scroll {
	flex: 1;
	overflow: hidden;
}

.register-container {
	padding: 40rpx;
}

.header-section {
	margin-bottom: 40rpx;
	
	.header-title {
		display: block;
		font-size: 48rpx;
		font-weight: bold;
		color: #0A0B0D;
		margin-bottom: 12rpx;
	}
	
	.header-desc {
		font-size: 28rpx;
		color: #5B616E;
	}
}

/* 标签页样式 */
.tab-container {
	display: flex;
	background: #F7F8FA;
	border-radius: 12rpx;
	padding: 6rpx;
	margin-bottom: 40rpx;
	
	.tab-item {
		flex: 1;
		height: 80rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10rpx;
		transition: all 0.2s;
		
		text {
			font-size: 28rpx;
			color: #5B616E;
			font-weight: 500;
		}
		
		&.active {
			background: #fff;
			box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.08);
			
			text {
				color: #0052FF;
				font-weight: 600;
			}
		}
	}
}

.form-section {
	.input-group {
		margin-bottom: 28rpx;
		
		.label {
			display: block;
			font-size: 26rpx;
			color: #5B616E;
			margin-bottom: 12rpx;
		}
		
		.input-wrapper {
			height: 96rpx;
			background: #F7F8FA;
			border-radius: 12rpx;
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
				font-size: 28rpx;
				color: #0A0B0D;
			}
			
			.eye-btn {
				width: 60rpx;
				height: 60rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				
				.eye-icon {
					font-size: 32rpx;
				}
			}
		}
		
		.code-wrapper {
			padding-right: 12rpx;
			
			.code-btn {
				padding: 0 24rpx;
				height: 64rpx;
				background: #0052FF;
				border-radius: 8rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				flex-shrink: 0;
				
				text {
					font-size: 24rpx;
					color: #fff;
					white-space: nowrap;
				}
				
				&.disabled {
					background: #E8E8E8;
					
					text {
						color: #9AA0AA;
					}
				}
			}
		}
	}
}

.placeholder {
	color: #9AA0AA;
}

.agreement-check {
	display: flex;
	align-items: flex-start;
	gap: 12rpx;
	margin: 32rpx 0;
	
	.checkbox {
		width: 40rpx;
		height: 40rpx;
		border: 2rpx solid #E8E8E8;
		border-radius: 8rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		margin-top: 4rpx;
		
		&.checked {
			background: #0052FF;
			border-color: #0052FF;
			
			text {
				color: #fff;
				font-size: 24rpx;
			}
		}
	}
	
	.agreement-text {
		font-size: 24rpx;
		color: #5B616E;
		line-height: 1.5;
		
		.link {
			color: #0052FF;
		}
	}
}

.register-btn {
	width: 100%;
	height: 96rpx;
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
	
	&:active:not([disabled]) {
		opacity: 0.9;
	}
}

.login-section {
	margin-top: 40rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	
	.login-text {
		font-size: 28rpx;
		color: #5B616E;
	}
	
	.login-link {
		font-size: 28rpx;
		color: #0052FF;
		font-weight: 500;
	}
}

/* 协议弹窗 */
.modal-mask {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 1000;
	display: flex;
	align-items: center;
	justify-content: center;
}

.modal-content {
	width: 90%;
	max-height: 80vh;
	background: #fff;
	border-radius: 24rpx;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.modal-header {
	padding: 32rpx;
	border-bottom: 1rpx solid #E8E8E8;
	display: flex;
	align-items: center;
	justify-content: space-between;
	flex-shrink: 0;
	
	.modal-title {
		font-size: 32rpx;
		font-weight: 600;
		color: #0A0B0D;
	}
	
	.modal-close {
		width: 56rpx;
		height: 56rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		
		text {
			font-size: 48rpx;
			color: #9AA0AA;
			line-height: 1;
		}
	}
}

.modal-body {
	flex: 1;
	max-height: 60vh;
	padding: 32rpx;
}

.agreement-content {
	.section-title {
		display: block;
		font-size: 28rpx;
		font-weight: 600;
		color: #0A0B0D;
		margin-top: 32rpx;
		margin-bottom: 16rpx;
		
		&:first-child {
			margin-top: 0;
		}
	}
	
	.section-text {
		display: block;
		font-size: 26rpx;
		color: #5B616E;
		line-height: 1.6;
		margin-bottom: 12rpx;
	}
}

.modal-footer {
	padding: 24rpx 32rpx;
	border-top: 1rpx solid #E8E8E8;
	flex-shrink: 0;
	
	.modal-btn {
		width: 100%;
		height: 88rpx;
		background: #0052FF;
		border-radius: 12rpx;
		color: #fff;
		font-size: 30rpx;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		
		&:active {
			opacity: 0.9;
		}
	}
}
</style>
