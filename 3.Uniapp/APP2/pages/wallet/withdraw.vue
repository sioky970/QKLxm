<template>
	<view class="page">
		<!-- 顶部导航栏 -->
		<view class="navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-inner">
				<view class="navbar-left" @click="goBack">
					<SvgIcon name="back" :size="22" color="#1A1A1A" />
				</view>
				<text class="navbar-title">提现</text>
				<view class="navbar-right" @click="goToRecords">
					<text class="navbar-action">提现记录</text>
				</view>
			</view>
		</view>

		<!-- 页面内容 -->
		<scroll-view scroll-y class="page-content" :style="{ paddingTop: navHeight + 'px' }">
			<!-- 账户余额卡片 -->
			<view class="balance-card">
				<view class="balance-header">
				<text class="balance-title">可用余额</text>
				<view class="balance-badge">USDT</view>
			</view>
			<view class="balance-amount">
					<text class="amount-integer">{{ formatBalanceInteger(balance) }}</text>
					<text class="amount-decimal">.{{ formatBalanceDecimal(balance) }}</text>
				</view>
				<view class="balance-footer">
					<text class="balance-hint">仅资金钱包余额可提现</text>
				</view>
				<view class="balance-tip">
					<text class="tip-icon">💡</text>
					<text class="tip-text">请先将其他钱包余额划转至资金钱包后再进行提现</text>
				</view>
			</view>

			<!-- 网络选择 - 两行网格布局 -->
			<view class="form-card">
				<view class="form-header">
					<text class="form-title">选择网络</text>
				</view>
				<view class="network-grid">
					<view 
						v-for="net in availableNetworks" 
						:key="net.name"
						class="network-tag"
						:class="{ active: selectedNetwork === net.name }"
						@click="selectNetwork(net.name)"
					>
						<text class="tag-name">{{ net.name }}</text>
					</view>
				</view>
			</view>

			<!-- 钱包地址输入 -->
			<view class="form-card">
				<view class="form-header">
					<text class="form-title">收款地址</text>
				</view>
				<view class="address-box" :class="{ error: chainAddressError, focus: addressFocused }">
					<input 
						class="address-input" 
						type="text"
						v-model="chainAddress" 
						:placeholder="addressPlaceholder"
						placeholder-class="input-placeholder"
						@input="onChainAddressInput"
						@focus="addressFocused = true"
						@blur="addressFocused = false"
					/>
					<view class="input-clear" v-if="chainAddress" @click="clearAddress">×</view>
				</view>
				<view class="address-tip">
					<text class="tip-error" v-if="chainAddressError">{{ chainAddressError }}</text>
					<text class="tip-warn" v-else>⚠️ 请仔细核对地址，转错无法找回</text>
				</view>
			</view>

			<!-- 提现金额输入 -->
			<view class="form-card">
				<view class="form-header">
					<text class="form-title">提现金额</text>
					<text class="form-max" @click="setMaxAmount">全部提现</text>
				</view>
				<view class="amount-box" :class="{ focus: amountFocused }">
					<text class="amount-unit">USDT</text>
					<input 
						class="amount-input" 
						type="digit" 
						v-model="amount" 
						placeholder="请输入提现金额"
						placeholder-class="input-placeholder"
						@focus="amountFocused = true"
						@blur="amountFocused = false"
					/>
					<button class="amount-all" @click="setMaxAmount">全部</button>
				</view>
				<view class="amount-tips">
					<text class="tip-item">最小提现: {{ minWithdraw }} USDT</text>
				</view>

				<!-- 金额明细 -->
				<view class="amount-detail" v-if="parseFloat(amount) > 0">
					<view class="detail-row">
						<text class="detail-label">提现金额</text>
						<text class="detail-value">{{ formatNumber(parseFloat(amount)) }} USDT</text>
					</view>
					<view class="detail-line"></view>
					<view class="detail-row total">
						<text class="detail-label">实际到账</text>
						<text class="detail-value receive">{{ formatNumber(parseFloat(amount)) }} USDT</text>
					</view>
				</view>
			</view>

			<!-- 支付密码 -->
			<view class="form-card">
				<view class="form-header">
					<text class="form-title">安全验证</text>
				</view>
				
				<!-- 未设置支付密码 -->
				<view class="password-empty" v-if="!hasPayPassword">
					<view class="empty-icon">🔐</view>
					<text class="empty-title">设置支付密码</text>
					<text class="empty-desc">提现需要6位数字支付密码</text>
					<button class="empty-btn" @click="showSetPayPasswordModal = true">立即设置</button>
				</view>
				
				<!-- 已设置支付密码 -->
				<view v-else>
					<view class="password-box" :class="{ focus: passwordFocused }">
						<view class="password-icon">🔒</view>
						<input 
							class="password-input" 
							type="password"
							v-model="payPassword" 
							placeholder="请输入6位支付密码"
							placeholder-class="input-placeholder"
							maxlength="6"
							@focus="passwordFocused = true"
							@blur="passwordFocused = false"
						/>
					</view>
					<text class="password-change" @click="showSetPayPasswordModal = true">修改支付密码</text>
				</view>
			</view>

			<!-- 设置支付密码弹窗 -->
			<view class="modal-mask" v-if="showSetPayPasswordModal" @click="showSetPayPasswordModal = false">
				<view class="modal-box" @click.stop>
					<view class="modal-header">
						<text class="modal-title">{{ hasPayPassword ? '修改支付密码' : '设置支付密码' }}</text>
						<view class="modal-close" @click="showSetPayPasswordModal = false">×</view>
					</view>
					<view class="modal-body">
						<view class="modal-field" v-if="hasPayPassword">
							<text class="field-label">原密码</text>
							<input class="field-input" type="password" v-model="oldPayPassword" placeholder="请输入原支付密码" maxlength="6" />
						</view>
						<view class="modal-field">
							<text class="field-label">新密码</text>
							<input class="field-input" type="password" v-model="newPayPassword" placeholder="请输入6位数字密码" maxlength="6" />
						</view>
						<view class="modal-field">
							<text class="field-label">确认密码</text>
							<input class="field-input" type="password" v-model="confirmPayPassword" placeholder="请再次输入密码" maxlength="6" />
						</view>
					</view>
					<view class="modal-footer">
						<button class="modal-btn cancel" @click="showSetPayPasswordModal = false">取消</button>
						<button class="modal-btn confirm" @click="submitPayPassword" :disabled="settingPayPassword">
							{{ settingPayPassword ? '设置中...' : '确认' }}
						</button>
					</view>
				</view>
			</view>

			<!-- 温馨提示 -->
			<view class="notice-card">
				<view class="notice-header">
					<text class="notice-icon">📋</text>
					<text class="notice-title">温馨提示</text>
				</view>
				<view class="notice-list">
					<text class="notice-item">• 请仔细核对钱包地址，转错将无法找回</text>
					<text class="notice-item">• 提现申请提交后将进入人工审核</text>
					<text class="notice-item">• 审核通过后将在24小时内完成链上转账</text>
					<text class="notice-item">• 到账时间取决于区块链网络拥堵情况</text>
					<text class="notice-item">• 如有疑问，请联系在线客服</text>
				</view>
			</view>

			<!-- 底部占位 -->
			<view class="footer-space"></view>
		</scroll-view>

		<!-- 底部提交按钮 -->
		<view class="submit-bar">
			<button 
				class="submit-btn" 
				:class="{ disabled: !canSubmit }"
				:disabled="!canSubmit || submitting"
				@click="submitWithdraw"
			>
				{{ submitButtonText }}
			</button>
			<text class="submit-tip" v-if="!canSubmit">{{ submitHintText }}</text>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { request } from '@/utils/api.js'
import SvgIcon from '@/components/SvgIcon.vue'
import walletStore from '@/stores/walletStore.js'

// ==================== 状态定义 ====================
const statusBarHeight = ref(0)
const navHeight = ref(88)
const loading = ref(false)
const submitting = ref(false)

// 余额与费率 - 使用资金钱包余额作为可提现余额
const balance = computed(() => walletStore.state.fundBalance || 0)
const minWithdraw = ref(10)
const feeRate = ref(0.001)

// 提现配置
const withdrawConfig = ref({
	networks: [],
	crypto_fee: 1,
	usdt_currency_id: 3
})
const configLoaded = ref(false)

// 可用网络列表（只保留TRC20和ERC20）
const availableNetworks = computed(() => {
	return withdrawConfig.value.networks.filter(net => 
		net.name === 'TRC20' || net.name === 'ERC20'
	)
})

// 区块链相关
const selectedNetwork = ref('')
const chainAddress = ref('')
const chainAddressError = ref('')
const addressFocused = ref(false)

// 支付密码
const hasPayPassword = ref(false)
const showSetPayPasswordModal = ref(false)
const settingPayPassword = ref(false)
const oldPayPassword = ref('')
const newPayPassword = ref('')
const confirmPayPassword = ref('')
const passwordFocused = ref(false)

// 表单数据
const amount = ref('')
const payPassword = ref('')
const amountFocused = ref(false)

// ==================== 计算属性 ====================
const fee = computed(() => {
	const amt = parseFloat(amount.value) || 0
	return amt * feeRate.value
})

const realAmount = computed(() => {
	const amt = parseFloat(amount.value) || 0
	return Math.max(0, amt - fee.value)
})

const canSubmit = computed(() => {
	const amt = parseFloat(amount.value) || 0
	return amt >= minWithdraw.value && 
		amt <= balance.value && 
		payPassword.value.length === 6 &&
		selectedNetwork.value && 
		chainAddress.value && 
		!chainAddressError.value
})

const addressPlaceholder = computed(() => {
	if (!selectedNetwork.value) return '请先选择网络'
	return `请输入${selectedNetwork.value}钱包地址`
})

const submitButtonText = computed(() => {
	if (submitting.value) return '提交中...'
	return '提交提现申请'
})

const submitHintText = computed(() => {
	const amt = parseFloat(amount.value) || 0
	if (amt <= 0) return '请输入提现金额'
	if (amt < minWithdraw.value) return `最小提现 ${minWithdraw.value} USDT`
	if (amt > balance.value) return '余额不足'
	if (payPassword.value.length !== 6) return '请输入6位支付密码'
	if (!selectedNetwork.value) return '请选择网络'
	if (!chainAddress.value) return '请输入钱包地址'
	if (chainAddressError.value) return '地址格式错误'
	return ''
})

// ==================== 生命周期 ====================
onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	navHeight.value = statusBarHeight.value + 44
	
	fetchWithdrawConfig()
	fetchUserInfo()
})



// ==================== 方法定义 ====================
const fetchWithdrawConfig = async () => {
	try {
		const res = await request({
			url: '/wallet/withdraw-config',
			method: 'GET'
		})
		if (res.data) {
			withdrawConfig.value = {
				networks: res.data.networks || [],
				crypto_fee: res.data.crypto_fee || 1,
				usdt_currency_id: res.data.usdt_currency_id || 3
			}
			if (availableNetworks.value.length > 0) {
				selectedNetwork.value = availableNetworks.value[0].name
			}
		}
		configLoaded.value = true
	} catch (err) {
		console.error('获取提现配置失败:', err)
		configLoaded.value = true
	}
}

const fetchUserInfo = async () => {
	loading.value = true
	try {
		const balanceRes = await request({
			url: '/wallet/asset-overview',
			method: 'GET'
		})
		balance.value = balanceRes.data?.total_balance || 0
		
		try {
			const userRes = await request({
				url: '/user/info',
				method: 'GET'
			})
			hasPayPassword.value = userRes.data?.has_pay_password || false
		} catch (err) {
			console.log('获取用户信息失败')
		}
	} catch (err) {
		uni.showToast({ title: '获取数据失败', icon: 'none' })
	} finally {
		loading.value = false
	}
}

const selectNetwork = (network) => {
	selectedNetwork.value = network
	if (chainAddress.value) {
		onChainAddressInput()
	}
}

const getCurrentNetworkFee = () => {
	const net = availableNetworks.value.find(n => n.name === selectedNetwork.value)
	return net ? net.fee : 0
}

const clearAddress = () => {
	chainAddress.value = ''
	chainAddressError.value = ''
}

const validateChainAddress = (network, address) => {
	if (!address) return '请输入钱包地址'
	
	const addr = address.trim()
	
	switch (network) {
		case 'TRC20':
			if (!addr.startsWith('T')) return 'TRC20地址必须以T开头'
			if (addr.length !== 34) return 'TRC20地址长度必须为34位'
			break
		case 'ERC20':
			if (!addr.startsWith('0x')) return 'ERC20地址必须以0x开头'
			if (addr.length !== 42) return 'ERC20地址长度必须为42位'
			break
	}
	
	return ''
}

const onChainAddressInput = () => {
	if (chainAddress.value && selectedNetwork.value) {
		chainAddressError.value = validateChainAddress(selectedNetwork.value, chainAddress.value)
	} else {
		chainAddressError.value = ''
	}
}

const submitPayPassword = async () => {
	if (newPayPassword.value.length !== 6) {
		uni.showToast({ title: '请输入6位支付密码', icon: 'none' })
		return
	}
	if (!/^\d{6}$/.test(newPayPassword.value)) {
		uni.showToast({ title: '支付密码必须为6位数字', icon: 'none' })
		return
	}
	if (newPayPassword.value !== confirmPayPassword.value) {
		uni.showToast({ title: '两次密码输入不一致', icon: 'none' })
		return
	}
	if (hasPayPassword.value && oldPayPassword.value.length !== 6) {
		uni.showToast({ title: '请输入原支付密码', icon: 'none' })
		return
	}
	
	settingPayPassword.value = true
	try {
		await request({
			url: '/user/change_pay_password',
			method: 'POST',
			header: { 'Content-Type': 'application/x-www-form-urlencoded' },
			data: {
				old_password: hasPayPassword.value ? oldPayPassword.value : '',
				new_password: newPayPassword.value
			}
		})
		
		uni.showToast({ title: '设置成功', icon: 'success' })
		hasPayPassword.value = true
		showSetPayPasswordModal.value = false
		oldPayPassword.value = ''
		newPayPassword.value = ''
		confirmPayPassword.value = ''
	} catch (err) {
		uni.showToast({ title: err.message || '设置失败', icon: 'none' })
	} finally {
		settingPayPassword.value = false
	}
}

const formatNumber = (num) => {
	if (!num && num !== 0) return '0.00'
	return parseFloat(num).toFixed(2)
}

const formatBalanceInteger = (num) => {
	const val = parseFloat(num) || 0
	return Math.floor(val).toLocaleString()
}

const formatBalanceDecimal = (num) => {
	const val = parseFloat(num) || 0
	return val.toFixed(2).split('.')[1]
}

const setMaxAmount = () => {
	amount.value = String(balance.value)
}

const submitWithdraw = async () => {
	if (!canSubmit.value) return
	
	const amt = parseFloat(amount.value)
	
	const confirmed = await new Promise((resolve) => {
		uni.showModal({
			title: '确认提现',
			content: `确认提现 ${formatNumber(amt)} USDT 到 ${selectedNetwork.value} 地址吗？\n\n钱包地址: ${chainAddress.value.slice(0, 8)}...${chainAddress.value.slice(-6)}`,
			confirmText: '确认',
			cancelText: '取消',
			success: (res) => resolve(res.confirm)
		})
	})
	
	if (!confirmed) return
	
	submitting.value = true
	try {
		await request({
			url: '/wallet/withdraw',
			method: 'POST',
			data: {
				currency_id: withdrawConfig.value.usdt_currency_id,
				amount: amt,
				pay_password: payPassword.value,
				network_type: selectedNetwork.value,
				chain_address: chainAddress.value.trim()
			}
		})
		
		uni.showToast({ 
			title: '提现申请已提交', 
			icon: 'success',
			duration: 2000
		})
		
		setTimeout(() => {
			goToRecords()
		}, 2000)
		
	} catch (err) {
		uni.showToast({ title: err.message || '提现失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}

const goToRecords = () => {
	uni.navigateTo({ url: '/pages/wallet/withdraw-records' })
}

const goBack = () => {
	uni.navigateBack()
}
</script>

<style scoped lang="scss">
// ==================== 变量定义 ====================
$primary: #0052FF;
$primary-dark: #003ECB;
$primary-light: #E8F0FF;
$text-primary: #1A1A1A;
$text-secondary: #666666;
$text-tertiary: #999999;
$bg-page: #F5F7FA;
$bg-card: #FFFFFF;
$border-color: #E8E8E8;
$success: #00C853;
$warning: #FF9800;
$error: #F44336;
$radius-sm: 8px;
$radius-md: 12px;
$radius-lg: 16px;

// ==================== 页面基础 ====================
.page {
	min-height: 100vh;
	background: $bg-page;
}

// ==================== 导航栏 ====================
.navbar {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: $bg-card;
	z-index: 999;
	box-shadow: 0 1px 0 rgba(0, 0, 0, 0.05);
}

.navbar-inner {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 44px;
	padding: 0 16px;
}

.navbar-left, .navbar-right {
	width: 72px;
}

.navbar-left {
	display: flex;
	align-items: center;
}

.navbar-title {
	font-size: 17px;
	font-weight: 600;
	color: $text-primary;
}

.navbar-right {
	display: flex;
	justify-content: flex-end;
}

.navbar-action {
	font-size: 14px;
	color: $primary;
	font-weight: 500;
}

// ==================== 页面内容 ====================
.page-content {
	height: 100vh;
	padding: 16px;
	box-sizing: border-box;
}

// ==================== 余额卡片 ====================
.balance-card {
	background: linear-gradient(135deg, $primary 0%, $primary-dark 100%);
	border-radius: $radius-lg;
	padding: 24px;
	margin-bottom: 16px;
	position: relative;
	overflow: hidden;
	
	&::after {
		content: '';
		position: absolute;
		top: -40%;
		right: -20%;
		width: 160px;
		height: 160px;
		background: radial-gradient(circle, rgba(255,255,255,0.12) 0%, transparent 70%);
		border-radius: 50%;
	}
}

.balance-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 12px;
}

.refresh-btn {
	width: 32px;
	height: 32px;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
}

.refresh-icon {
	font-size: 16px;
}

.balance-title {
	font-size: 13px;
	color: rgba(255, 255, 255, 0.8);
}

.balance-badge {
	background: rgba(255, 255, 255, 0.2);
	padding: 3px 8px;
	border-radius: 10px;
	font-size: 11px;
	color: #FFFFFF;
	font-weight: 600;
}

.balance-amount {
	display: flex;
	align-items: baseline;
	margin-bottom: 8px;
}

.amount-integer {
	font-size: 32px;
	font-weight: 700;
	color: #FFFFFF;
}

.amount-decimal {
	font-size: 18px;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.7);
}

.balance-hint {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.6);
	}

	.balance-tip {
		margin-top: 12px;
		padding: 10px 12px;
		background: rgba(255, 255, 255, 0.15);
		border-radius: 8px;
		display: flex;
		align-items: flex-start;
		gap: 8px;

		.tip-icon {
			font-size: 14px;
			line-height: 1.4;
		}

		.tip-text {
			font-size: 12px;
			color: rgba(255, 255, 255, 0.9);
			line-height: 1.4;
			flex: 1;
		}
	}

// ==================== 表单卡片 ====================
.form-card {
	background: $bg-card;
	border-radius: $radius-md;
	padding: 16px;
	margin-bottom: 12px;
}

.form-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 12px;
}

.form-title {
	font-size: 14px;
	font-weight: 600;
	color: $text-primary;
}

.form-max {
	font-size: 13px;
	color: $primary;
	font-weight: 500;
}

// ==================== 网络选择 - 两行网格布局 ====================
.network-grid {
	display: flex;
	flex-wrap: wrap;
	gap: 8px;
}

.network-tag {
	flex: 0 0 calc(33.33% - 6px);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 10px 8px;
	background: $bg-page;
	border: 2px solid transparent;
	border-radius: $radius-sm;
	transition: all 0.2s ease;
	
	&.active {
		background: $primary-light;
		border-color: $primary;
		
		.tag-name {
			color: $primary;
		}
		
		.tag-fee {
			color: $primary;
		}
	}
	
	&:active {
		transform: scale(0.96);
	}
}

.tag-name {
	font-size: 13px;
	font-weight: 600;
	color: $text-primary;
	margin-bottom: 2px;
}

.tag-fee {
	font-size: 11px;
	color: $text-tertiary;
}

.network-info {
	margin-top: 10px;
	padding-top: 10px;
	border-top: 1px dashed $border-color;
}

.info-text {
	font-size: 12px;
	color: $text-secondary;
}

// ==================== 地址输入 ====================
.address-box {
	background: $bg-page;
	border: 2px solid $border-color;
	border-radius: $radius-sm;
	padding: 12px 14px;
	display: flex;
	align-items: center;
	gap: 10px;
	transition: all 0.2s ease;
	
	&.focus {
		border-color: $primary;
		background: #FFFFFF;
	}
	
	&.error {
		border-color: $error;
	}
}

.address-input {
	flex: 1;
	font-size: 14px;
	font-family: 'Monaco', monospace;
	color: $text-primary;
	background: transparent;
}

.input-clear {
	width: 20px;
	height: 20px;
	background: #CCC;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #FFFFFF;
	font-size: 14px;
	line-height: 1;
}

.address-tip {
	margin-top: 8px;
	padding: 0 2px;
}

.tip-error {
	font-size: 12px;
	color: $error;
}

.tip-warn {
	font-size: 12px;
	color: $warning;
}

.input-placeholder {
	color: #AAAAAA;
}

// ==================== 金额输入 ====================
.amount-box {
	background: $bg-page;
	border: 2px solid $border-color;
	border-radius: $radius-sm;
	padding: 0 14px;
	height: 56px;
	display: flex;
	align-items: center;
	gap: 12px;
	transition: all 0.2s ease;
	
	&.focus {
		border-color: $primary;
		background: #FFFFFF;
	}
}

.amount-unit {
	font-size: 14px;
	font-weight: 700;
	color: $primary;
	padding-right: 12px;
	border-right: 1px solid $border-color;
}

.amount-input {
	flex: 1;
	font-size: 22px;
	font-weight: 600;
	color: $text-primary;
	background: transparent;
}

.amount-all {
	background: rgba($primary, 0.1);
	color: $primary;
	border: none;
	border-radius: 6px;
	padding: 6px 12px;
	font-size: 12px;
	font-weight: 600;
	height: auto;
	line-height: 1;
}

.amount-tips {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 10px;
	margin-top: 10px;
}

.tip-item {
	font-size: 12px;
	color: $text-tertiary;
}

.tip-sep {
	color: $border-color;
}

// ==================== 金额明细 ====================
.amount-detail {
	background: $bg-page;
	border-radius: $radius-sm;
	padding: 14px;
	margin-top: 14px;
}

.detail-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 6px 0;
	
	&.total {
		padding-top: 10px;
	}
}

.detail-line {
	height: 1px;
	background: linear-gradient(90deg, transparent, $border-color, transparent);
	margin: 4px 0;
}

.detail-label {
	font-size: 13px;
	color: $text-secondary;
}

.detail-value {
	font-size: 14px;
	color: $text-primary;
	font-weight: 500;
	
	&.fee {
		color: $error;
	}
	
	&.receive {
		font-size: 17px;
		font-weight: 700;
		color: $primary;
	}
}

// ==================== 支付密码 ====================
.password-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 20px 0;
	gap: 6px;
}

.empty-icon {
	font-size: 36px;
	margin-bottom: 6px;
}

.empty-title {
	font-size: 15px;
	font-weight: 600;
	color: $text-primary;
}

.empty-desc {
	font-size: 13px;
	color: $text-tertiary;
}

.empty-btn {
	margin-top: 10px;
	background: $primary;
	color: #FFFFFF;
	border: none;
	border-radius: $radius-sm;
	padding: 10px 28px;
	font-size: 14px;
	font-weight: 600;
}

.password-box {
	background: $bg-page;
	border: 2px solid $border-color;
	border-radius: $radius-sm;
	padding: 12px 14px;
	display: flex;
	align-items: center;
	gap: 12px;
	transition: all 0.2s ease;
	
	&.focus {
		border-color: $primary;
		background: #FFFFFF;
	}
}

.password-icon {
	font-size: 18px;
}

.password-input {
	flex: 1;
	font-size: 16px;
	color: $text-primary;
	background: transparent;
	letter-spacing: 6px;
}

.password-change {
	font-size: 12px;
	color: $primary;
	margin-top: 8px;
	display: block;
	padding-left: 2px;
}

// ==================== 弹窗 ====================
.modal-mask {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 9999;
}

.modal-box {
	background: $bg-card;
	border-radius: $radius-lg;
	width: 85%;
	max-width: 360px;
	overflow: hidden;
}

.modal-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 16px 18px;
	border-bottom: 1px solid $border-color;
}

.modal-title {
	font-size: 16px;
	font-weight: 600;
	color: $text-primary;
}

.modal-close {
	font-size: 22px;
	color: $text-tertiary;
	line-height: 1;
	padding: 4px;
}

.modal-body {
	padding: 18px;
}

.modal-field {
	margin-bottom: 14px;
	
	&:last-child {
		margin-bottom: 0;
	}
}

.field-label {
	display: block;
	font-size: 13px;
	color: $text-secondary;
	margin-bottom: 6px;
}

.field-input {
	width: 100%;
	height: 44px;
	background: $bg-page;
	border: 1px solid $border-color;
	border-radius: $radius-sm;
	padding: 0 12px;
	font-size: 15px;
	box-sizing: border-box;
	
	&:focus {
		border-color: $primary;
		background: #FFFFFF;
	}
}

.modal-footer {
	display: flex;
	gap: 10px;
	padding: 14px 18px 18px;
}

.modal-btn {
	flex: 1;
	height: 44px;
	border-radius: $radius-sm;
	font-size: 15px;
	font-weight: 600;
	border: none;
	
	&.cancel {
		background: #F0F0F0;
		color: $text-secondary;
	}
	
	&.confirm {
		background: $primary;
		color: #FFFFFF;
		
		&:disabled {
			background: #CCCCCC;
		}
	}
}

// ==================== 提示卡片 ====================
.notice-card {
	background: $primary-light;
	border: 1px solid #D6E4FF;
	border-radius: $radius-md;
	padding: 16px;
	margin-top: 12px;
}

.notice-header {
	display: flex;
	align-items: center;
	gap: 6px;
	margin-bottom: 12px;
}

.notice-icon {
	font-size: 14px;
}

.notice-title {
	font-size: 14px;
	font-weight: 600;
	color: $primary;
}

.notice-list {
	display: flex;
	flex-direction: column;
	gap: 6px;
}

.notice-item {
	font-size: 12px;
	color: #4A7FD0;
	line-height: 1.5;
}

// ==================== 底部提交 ====================
.footer-space {
	height: 110px;
}

.submit-bar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	background: linear-gradient(180deg, rgba($bg-page, 0) 0%, $bg-page 30%);
	padding: 16px;
	padding-bottom: calc(16px + env(safe-area-inset-bottom));
	z-index: 100;
}

.submit-btn {
	width: 100%;
	background: $primary;
	color: #FFFFFF;
	border: none;
	border-radius: $radius-md;
	height: 50px;
	font-size: 16px;
	font-weight: 600;
	box-shadow: 0 4px 12px rgba($primary, 0.3);
	transition: all 0.2s ease;
	
	&:active {
		background: $primary-dark;
		transform: scale(0.98);
	}
	
	&.disabled {
		background: #D0D0D0;
		box-shadow: none;
	}
}

.submit-tip {
	display: block;
	text-align: center;
	font-size: 12px;
	color: $text-tertiary;
	margin-top: 8px;
}

// ==================== 响应式 ====================
@media screen and (min-width: 414px) {
	.page-content {
		padding: 20px;
	}
	
	.balance-card {
		padding: 28px;
	}
	
	.amount-integer {
		font-size: 36px;
	}
}

@media screen and (min-width: 768px) {
	.page-content {
		max-width: 540px;
		margin: 0 auto;
	}
}
</style>
