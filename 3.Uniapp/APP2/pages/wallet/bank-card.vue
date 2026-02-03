<template>
	<view class="page">
		<!-- 自定义导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="nav-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<text class="nav-title">银行卡管理</text>
				<view class="nav-right"></view>
			</view>
		</view>

		<!-- 主内容区域 -->
		<scroll-view scroll-y class="content-scroll" :style="{ paddingTop: navHeight + 'px' }">
			<!-- 已绑定银行卡显示 -->
			<view class="card-display" v-if="hasCard && !isEditing">
				<view class="card-box">
					<view class="card-header">
						<text class="bank-name">{{ cardInfo.bank_name }}</text>
					</view>
					<view class="card-number">
						<text>{{ formatCardNumber(cardInfo.bank_account) }}</text>
					</view>
					<view class="card-footer">
						<view class="cardholder">
							<text class="label">持卡人</text>
							<text class="value">{{ cardInfo.real_name }}</text>
						</view>
						<view class="branch" v-if="cardInfo.bank_branch">
							<text class="label">开户行</text>
							<text class="value">{{ cardInfo.bank_branch }}</text>
						</view>
					</view>
				</view>
				<button class="edit-btn" @click="enterEditMode">修改银行卡</button>
			</view>

			<!-- 绑定/编辑表单 -->
			<view class="form-section" v-else>
				<view class="section-card">
					<view class="section-title">持卡人姓名</view>
					<view class="input-box">
						<input 
							class="input" 
							v-model="form.real_name" 
							placeholder="请输入持卡人真实姓名"
							placeholder-class="placeholder"
						/>
					</view>
					<text class="input-tip">请确保与实名认证姓名一致</text>
				</view>

				<view class="section-card">
					<view class="section-title">银行名称</view>
					<view class="input-box">
						<input 
							class="input" 
							v-model="form.bank_name" 
							placeholder="例如：中国工商银行"
							placeholder-class="placeholder"
						/>
					</view>
				</view>

				<view class="section-card">
					<view class="section-title">银行卡号</view>
					<view class="input-box">
						<input 
							class="input" 
							type="number"
							v-model="form.bank_account" 
							placeholder="请输入银行卡号"
							placeholder-class="placeholder"
						/>
					</view>
				</view>

				<view class="section-card">
					<view class="section-title">开户行支行（可选）</view>
					<view class="input-box">
						<input 
							class="input" 
							v-model="form.bank_branch" 
							placeholder="例如：XX省XX市XX支行"
							placeholder-class="placeholder"
						/>
					</view>
				</view>

				<!-- 温馨提示 -->
				<view class="tips-section">
					<view class="tips-title">温馨提示</view>
					<view class="tips-list">
						<text class="tips-item">• 银行卡信息将用于提现到账，请确保准确无误</text>
						<text class="tips-item">• 持卡人必须与实名认证姓名一致</text>
						<text class="tips-item">• 提现时将根据此银行卡进行转账</text>
						<text class="tips-item">• 如需修改，请联系客服确认身份后操作</text>
					</view>
				</view>
			</view>

			<!-- 底部占位 -->
			<view class="bottom-placeholder"></view>
		</scroll-view>

		<!-- 底部按钮 -->
		<view class="submit-section" v-if="isEditing || !hasCard">
			<button 
				class="submit-btn" 
				:class="{ disabled: !canSubmit }"
				:disabled="!canSubmit || submitting"
				@click="submitCard"
			>
				{{ submitting ? '提交中...' : (hasCard ? '保存修改' : '绑定银行卡') }}
			</button>
			<button 
				class="cancel-btn" 
				v-if="hasCard"
				@click="cancelEdit"
			>
				取消
			</button>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { request } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'

// 状态栏高度
const statusBarHeight = ref(0)
const navHeight = ref(88)

// 数据状态
const loading = ref(false)
const submitting = ref(false)
const hasCard = ref(false)
const isEditing = ref(false)

// 银行卡信息
const cardInfo = ref({
	real_name: '',
	bank_name: '',
	bank_account: '',
	bank_branch: ''
})

// 表单数据
const form = ref({
	real_name: '',
	bank_name: '',
	bank_account: '',
	bank_branch: ''
})

// 计算是否可以提交
const canSubmit = computed(() => {
	return form.value.real_name.trim() && 
	       form.value.bank_name.trim() && 
	       form.value.bank_account.trim()
})

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	navHeight.value = statusBarHeight.value + 44
	
	fetchCardInfo()
})

// 获取银行卡信息
const fetchCardInfo = async () => {
	loading.value = true
	try {
		const res = await request({
			url: '/user/cash_info',
			method: 'GET'
		})
		
		if (res.data && res.data.bank_account) {
			hasCard.value = true
			cardInfo.value = res.data
		} else {
			hasCard.value = false
		}
	} catch (err) {
		console.error('获取银行卡信息失败:', err)
		// 如果未绑定，不提示错误
		if (!err.message.includes('未绑定')) {
			uni.showToast({ title: err.message || '获取信息失败', icon: 'none' })
		}
	} finally {
		loading.value = false
	}
}

// 格式化银行卡号（显示前4位和后4位）
const formatCardNumber = (cardNumber) => {
	if (!cardNumber) return ''
	const str = String(cardNumber)
	if (str.length <= 8) return str
	return str.slice(0, 4) + ' **** **** ' + str.slice(-4)
}

// 进入编辑模式
const enterEditMode = () => {
	form.value = { ...cardInfo.value }
	isEditing.value = true
}

// 取消编辑
const cancelEdit = () => {
	form.value = {
		real_name: '',
		bank_name: '',
		bank_account: '',
		bank_branch: ''
	}
	isEditing.value = false
}

// 提交银行卡信息
const submitCard = async () => {
	if (!canSubmit.value) return
	
	submitting.value = true
	try {
		await request({
			url: '/user/cash_info',
			method: 'POST',
			data: {
				real_name: form.value.real_name.trim(),
				bank_name: form.value.bank_name.trim(),
				bank_account: form.value.bank_account.trim(),
				bank_branch: form.value.bank_branch.trim()
			}
		})
		
		uni.showToast({ 
			title: hasCard.value ? '修改成功' : '绑定成功', 
			icon: 'success' 
		})
		
		// 重新获取信息
		await fetchCardInfo()
		isEditing.value = false
		
	} catch (err) {
		console.error('提交失败:', err)
		uni.showToast({ title: err.message || '操作失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}

// 返回上一页
const goBack = () => {
	uni.navigateBack()
}
</script>

<style scoped>
.page {
	min-height: 100vh;
	background: #F5F6FA;
	position: relative;
}

/* 导航栏 */
.custom-navbar {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: #fff;
	z-index: 999;
	border-bottom: 1px solid #E5E7EB;
}

.navbar-content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 44px;
	padding: 0 16px;
}

.nav-left, .nav-right {
	width: 60px;
}

.nav-title {
	font-size: 17px;
	font-weight: 600;
	color: #1F2937;
}

/* 主内容区域 */
.content-scroll {
	height: 100vh;
	padding: 12px 16px 100px;
}

/* 银行卡展示卡片 */
.card-display {
	margin-bottom: 16px;
}

.card-box {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	border-radius: 16px;
	padding: 24px;
	color: #fff;
	margin-bottom: 16px;
}

.card-header {
	margin-bottom: 24px;
}

.bank-name {
	font-size: 18px;
	font-weight: 600;
}

.card-number {
	font-size: 20px;
	letter-spacing: 2px;
	margin-bottom: 20px;
	font-family: monospace;
}

.card-footer {
	display: flex;
	justify-content: space-between;
}

.cardholder, .branch {
	display: flex;
	flex-direction: column;
}

.label {
	font-size: 12px;
	opacity: 0.8;
	margin-bottom: 4px;
}

.value {
	font-size: 14px;
	font-weight: 500;
}

.edit-btn {
	background: #fff;
	color: #667eea;
	border: 1px solid #667eea;
	border-radius: 8px;
	height: 44px;
	line-height: 44px;
	font-size: 15px;
	font-weight: 500;
}

/* 表单区域 */
.form-section {
	margin-bottom: 16px;
}

.section-card {
	background: #fff;
	border-radius: 12px;
	padding: 16px;
	margin-bottom: 12px;
}

.section-title {
	font-size: 14px;
	font-weight: 600;
	color: #1F2937;
	margin-bottom: 12px;
}

.input-box {
	background: #F9FAFB;
	border: 1px solid #E5E7EB;
	border-radius: 8px;
	padding: 0 12px;
	height: 44px;
}

.input {
	height: 100%;
	font-size: 15px;
	color: #1F2937;
}

.placeholder {
	color: #9CA3AF;
}

.input-tip {
	font-size: 12px;
	color: #6B7280;
	margin-top: 8px;
	display: block;
}

/* 温馨提示 */
.tips-section {
	background: #FEF3C7;
	border-radius: 12px;
	padding: 16px;
	margin-top: 12px;
}

.tips-title {
	font-size: 14px;
	font-weight: 600;
	color: #92400E;
	margin-bottom: 12px;
	display: flex;
	align-items: center;
}

.tips-list {
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.tips-item {
	font-size: 13px;
	color: #92400E;
	line-height: 1.5;
}

/* 底部提交区域 */
.bottom-placeholder {
	height: 80px;
}

.submit-section {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	background: #fff;
	padding: 12px 16px;
	padding-bottom: calc(12px + env(safe-area-inset-bottom));
	box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
	z-index: 100;
}

.submit-btn {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: #fff;
	border: none;
	border-radius: 8px;
	height: 48px;
	line-height: 48px;
	font-size: 16px;
	font-weight: 600;
	margin-bottom: 8px;
}

.submit-btn.disabled {
	background: #D1D5DB;
	color: #9CA3AF;
}

.cancel-btn {
	background: #F3F4F6;
	color: #6B7280;
	border: none;
	border-radius: 8px;
	height: 44px;
	line-height: 44px;
	font-size: 15px;
}
</style>
