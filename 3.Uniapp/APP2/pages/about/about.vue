<template>
	<view class="page">
		<!-- 自定义顶部导航栏 -->
		<view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
			<view class="navbar-content">
				<view class="navbar-left" @click="goBack">
					<SvgIcon name="back" :size="24" color="#333" />
				</view>
				<view class="navbar-title">关于我们</view>
				<view class="navbar-right"></view>
			</view>
		</view>

		<!-- 滚动区域 -->
		<scroll-view scroll-view scroll-y class="content-scroll">
			<!-- 顶部横幅 -->
			<view class="banner">
				<image class="banner-img" src="https://images.ctfassets.net/o10es7wu5gm1/2gsOMm2A5SOFiZiFUefQfd/bd8194c79d3a0f7c2af022164af6081f/Unsplash-xE5SAT0WeTQ.jpg?fm=avif&w=2586&h=2586&q=65" mode="aspectFill"></image>
				<view class="banner-overlay">
					<image class="coinbase-logo" src="https://static-assets.coinbase.com/ui-infra/illustration/v1/pictogram/svg/light/coinbaseLogoNavigation-4.svg" mode="aspectFit"></image>
					<text class="banner-subtitle">为全球超过 10 亿人带来经济自由</text>
				</view>
			</view>

			<!-- 1. 使命愿景 -->
			<view class="section">
				<view class="section-title">我们的使命</view>
				<view class="section-content">
					<text class="text-p">我们的使命是增加全球的经济自由。我们相信，通过将世界带入链上，我们可以创造一个更加公平、无障碍、高效和透明的经济体系。</text>
					<text class="text-p">加密货币不仅仅是一种新的资产类别——它是重塑金融体系的基础，“链上是新的在线模式”。</text>
				</view>
			</view>

			<!-- 2. 核心价值观 -->
			<view class="section values-section">
				<view class="section-title">核心价值观</view>
				<view class="values-grid">
					<view class="value-item">
						<view class="value-icon-box">
							<SvgIcon name="globe" :size="24" color="#3B82F6" />
						</view>
						<text class="value-name">清晰沟通</text>
						<text class="value-desc">透明、直接、有效的交流，确保使命对齐。</text>
					</view>
					<view class="value-item">
						<view class="value-icon-box">
							<SvgIcon name="gift" :size="24" color="#3B82F6" />
						</view>
						<text class="value-name">正向能量</text>
						<text class="value-desc">保持乐观，为团队文化贡献活力与积极影响。</text>
					</view>
					<view class="value-item">
						<view class="value-icon-box">
							<SvgIcon name="help" :size="24" color="#3B82F6" />
						</view>
						<text class="value-name">持续学习</text>
						<text class="value-desc">将失败视为机会，在谦虚的环境中不断进化。</text>
					</view>
					<view class="value-item">
						<view class="value-icon-box">
							<SvgIcon name="flash-swap" :size="24" color="#3B82F6" />
						</view>
						<text class="value-name">高效执行</text>
						<text class="value-desc">快速决策，以主人翁意识对最终成果负责。</text>
					</view>
				</view>
			</view>

			<!-- 3. 发展里程碑 -->
			<view class="section timeline-section">
				<view class="section-title">成长历程</view>
				<view class="timeline">
					<view class="timeline-item" v-for="(item, index) in milestones" :key="index">
						<view class="timeline-left">
							<text class="time-year">{{ item.year }}</text>
						</view>
						<view class="timeline-right">
							<view class="timeline-dot"></view>
							<view class="timeline-content">
								<text class="milestone-title">{{ item.title }}</text>
								<text class="milestone-desc">{{ item.desc }}</text>
							</view>
						</view>
					</view>
				</view>
			</view>

			<!-- 4. 关键数据 -->
			<view class="stats-grid">
				<view class="stats-item">
					<text class="stats-value">$295B</text>
					<text class="stats-label">季度交易量</text>
				</view>
				<view class="stats-item">
					<text class="stats-value">$516B</text>
					<text class="stats-label">平台资产</text>
				</view>
				<view class="stats-item">
					<text class="stats-value">100+</text>
					<text class="stats-label">覆盖国家</text>
				</view>
				<view class="stats-item">
					<text class="stats-value">1.1亿+</text>
					<text class="stats-label">验证用户</text>
				</view>
			</view>

			<!-- 5. 安全保障 (深度) -->
			<view class="section security-section">
				<view class="section-title">安全保障防线</view>
				<view class="security-card" v-for="(sec, idx) in securityMeasures" :key="idx">
					<view class="sec-header">
						<SvgIcon :name="sec.icon" :size="20" color="#3B82F6" />
						<text class="sec-title">{{ sec.title }}</text>
					</view>
					<text class="sec-desc">{{ sec.content }}</text>
				</view>
			</view>

			<!-- 6. 领导团队 -->
			<view class="section team-section">
				<view class="section-title">执行领导团队</view>
				<scroll-view scroll-x class="team-scroll">
					<view class="team-list">
						<view class="team-member" v-for="(member, midx) in executives" :key="midx">
							<image class="member-avatar" :src="member.avatar" mode="aspectFill"></image>
							<text class="member-name">{{ member.name }}</text>
							<text class="member-role">{{ member.role }}</text>
						</view>
					</view>
				</scroll-view>
			</view>

			<!-- 7. 合规承诺 -->
			<view class="section compliance-section">
				<view class="section-title">合规与信任</view>
				<image class="compliance-img" src="https://images.ctfassets.net/o10es7wu5gm1/V1ohRCN3TNAXc1CfBnFMn/6238184fd2d449d14c6ddf5b391540ff/image__10_.png?fm=avif&w=2640&h=1386&q=65" mode="aspectFit"></image>
				<view class="section-content">
					<text class="text-p">Coinbase 致力于成为最值得信赖的加密货币平台。我们的合规计划植根于传统金融的最佳实践，并结合了创新的合规技术。</text>
					<view class="compliance-list">
						<view class="compliance-item">
							<view class="dot"></view>
							<text class="compliance-text">严格的身份验证与财务合规</text>
						</view>
						<view class="compliance-item">
							<view class="dot"></view>
							<text class="compliance-text">深度的制裁合规监控</text>
						</view>
						<view class="compliance-item">
							<view class="dot"></view>
							<text class="compliance-text">作为 NASDAQ 上市公司（COIN）接受公开审计</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 底部信息 -->
			<view class="footer">
				<text class="footer-text">© 2026 Coinbase, Inc. All rights reserved.</text>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import SvgIcon from '@/components/SvgIcon.vue'

const statusBarHeight = ref(0)

// 发展里程碑数据
const milestones = ref([
	{ year: '2012', title: 'Coinbase 创立', desc: '由 Brian Armstrong 在旧金山创立，旨在让比特币买卖更简单。' },
	{ year: '2013', title: 'A轮融资', desc: '获得 a16z 领投的 2500 万美元融资，开始进入快速扩张期。' },
	{ year: '2017', title: '全球领先', desc: '成为全球交易量最大的加密货币平台之一，推动行业走向大众。' },
	{ year: '2021', title: '纳斯达克上市', desc: '在 NASDAQ 挂牌上市（COIN），成为首家上市的加密货币交易所。' },
	{ year: '2025', title: '经济自由', desc: '管理资产超 5160 亿美元，服务超过 100 个国家的验证用户。' }
])

// 安全措施数据
const securityMeasures = ref([
	{ icon: 'wallet', title: '98% 离线冷存储', content: '绝大部分数字资产存储在安全的离线钱包中，有效防范黑客攻击。' },
	{ icon: 'kyc', title: '银行级加密标准', content: '采用 AES-256 加密技术和严格的身份验证（2FA），保护账户隐私。' },
	{ icon: 'document', title: '全额准备金保障', content: '遵循 1:1 资产持有政策，未经用户许可绝不将资产用于借贷。' }
])

// 领导团队数据
const executives = ref([
	{ name: 'Brian Armstrong', role: 'CEO & 联合创始人', avatar: 'https://images.ctfassets.net/o10es7wu5gm1/2gsOMm2A5SOFiZiFUefQfd/bd8194c79d3a0f7c2af022164af6081f/Unsplash-xE5SAT0WeTQ.jpg?w=200&h=200&q=80' },
	{ name: 'Emilie Choi', role: '总裁兼首席运营官', avatar: 'https://images.ctfassets.net/o10es7wu5gm1/5btpBgFB1sHcbmqLExyYvK/2736742a45b5cc96fbec1117cb49e728/working-at-cb-1.jpg?w=200&h=200&q=80' },
	{ name: 'Alesia Haas', role: '首席财务官', avatar: 'https://images.ctfassets.net/o10es7wu5gm1/V1ohRCN3TNAXc1CfBnFMn/6238184fd2d449d14c6ddf5b391540ff/image__10_.png?w=200&h=200&q=80' }
])

onMounted(() => {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
})

const goBack = () => {
	uni.navigateBack()
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #FFFFFF;
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
	
	.navbar-left {
		width: 60rpx;
		display: flex;
		align-items: center;
	}
	
	.navbar-title {
		font-size: 32rpx;
		font-weight: bold;
		color: #333;
	}
	
	.navbar-right {
		width: 60rpx;
	}
}

.content-scroll {
	flex: 1;
	overflow: hidden;
}

.banner {
	position: relative;
	width: 100%;
	height: 400rpx;
	
	.banner-img {
		width: 100%;
		height: 100%;
	}
	
	.banner-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.4);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 0 40rpx;
		
		.coinbase-logo {
			width: 300rpx;
			height: 80rpx;
			margin-bottom: 20rpx;
		}
		
		.banner-subtitle {
			color: #FFFFFF;
			font-size: 28rpx;
			text-align: center;
		}
	}
}

.section {
	padding: 40rpx 30rpx;
	
	.section-title {
		font-size: 36rpx;
		font-weight: bold;
		color: #000;
		margin-bottom: 24rpx;
		display: flex;
		align-items: center;
		
		&::before {
			content: "";
			width: 8rpx;
			height: 32rpx;
			background: #3B82F6;
			margin-right: 16rpx;
			border-radius: 4rpx;
		}
	}
	
	.section-content {
		.text-p {
			display: block;
			font-size: 28rpx;
			color: #666;
			line-height: 1.6;
			margin-bottom: 20rpx;
		}
	}
}

/* 核心价值观 */
.values-grid {
	display: flex;
	flex-wrap: wrap;
	gap: 20rpx;
	
	.value-item {
		width: calc(50% - 10rpx);
		background: #F8FAFC;
		padding: 30rpx 20rpx;
		border-radius: 16rpx;
		display: flex;
		flex-direction: column;
		
		.value-icon-box {
			width: 60rpx;
			height: 60rpx;
			margin-bottom: 16rpx;
		}
		
		.value-name {
			font-size: 30rpx;
			font-weight: bold;
			color: #333;
			margin-bottom: 8rpx;
		}
		
		.value-desc {
			font-size: 24rpx;
			color: #999;
			line-height: 1.4;
		}
	}
}

/* 发展里程碑 */
.timeline {
	padding-left: 20rpx;
	
	.timeline-item {
		display: flex;
		margin-bottom: 40rpx;
		
		&:last-child {
			margin-bottom: 0;
		}
		
		.timeline-left {
			width: 100rpx;
			padding-top: 4rpx;
			
			.time-year {
				font-size: 28rpx;
				font-weight: bold;
				color: #3B82F6;
			}
		}
		
		.timeline-right {
			flex: 1;
			position: relative;
			padding-left: 40rpx;
			padding-bottom: 20rpx;
			border-left: 2rpx solid #E5E7EB;
			
			.timeline-dot {
				position: absolute;
				left: -7rpx;
				top: 14rpx;
				width: 12rpx;
				height: 12rpx;
				background: #3B82F6;
				border-radius: 50%;
				border: 4rpx solid #FFF;
				box-shadow: 0 0 0 4rpx rgba(59, 130, 246, 0.2);
			}
			
			.milestone-title {
				display: block;
				font-size: 30rpx;
				font-weight: bold;
				color: #333;
				margin-bottom: 8rpx;
			}
			
			.milestone-desc {
				font-size: 26rpx;
				color: #666;
				line-height: 1.5;
			}
		}
	}
}

/* 安全保障卡片 */
.security-card {
	background: #FFFFFF;
	border: 1rpx solid #F1F5F9;
	border-radius: 16rpx;
	padding: 30rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.02);
	
	.sec-header {
		display: flex;
		align-items: center;
		gap: 16rpx;
		margin-bottom: 16rpx;
		
		.sec-title {
			font-size: 30rpx;
			font-weight: bold;
			color: #333;
		}
	}
	
	.sec-desc {
		font-size: 26rpx;
		color: #666;
		line-height: 1.6;
	}
}

/* 团队 */
.team-scroll {
	width: 100%;
	
	.team-list {
		display: flex;
		padding: 10rpx 0;
	}
	
	.team-member {
		flex-shrink: 0;
		width: 240rpx;
		margin-right: 30rpx;
		display: flex;
		flex-direction: column;
		align-items: center;
		
		.member-avatar {
			width: 160rpx;
			height: 160rpx;
			border-radius: 50%;
			background: #F1F5F9;
			margin-bottom: 20rpx;
		}
		
		.member-name {
			font-size: 28rpx;
			font-weight: bold;
			color: #333;
			text-align: center;
			margin-bottom: 4rpx;
		}
		
		.member-role {
			font-size: 22rpx;
			color: #999;
			text-align: center;
		}
	}
}

.stats-grid {
	display: flex;
	flex-wrap: wrap;
	background: #F8FAFC;
	padding: 40rpx 20rpx;
	margin: 0 30rpx;
	border-radius: 16rpx;
	
	.stats-item {
		width: 50%;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 20rpx 0;
		
		.stats-value {
			font-size: 40rpx;
			font-weight: bold;
			color: #3B82F6;
			margin-bottom: 8rpx;
		}
		
		.stats-label {
			font-size: 24rpx;
			color: #999;
		}
	}
}

.compliance-section {
	.compliance-img {
		width: 100%;
		height: 300rpx;
		border-radius: 12rpx;
		margin-bottom: 30rpx;
	}
	
	.compliance-list {
		margin-top: 20rpx;
		
		.compliance-item {
			display: flex;
			align-items: center;
			gap: 16rpx;
			margin-bottom: 16rpx;
			
			.compliance-text {
				font-size: 26rpx;
				color: #444;
			}
			
			.dot {
				width: 12rpx;
				height: 12rpx;
				border-radius: 50%;
				background-color: #3B82F6;
				flex-shrink: 0;
			}
		}
	}
}

.footer {
	padding: 60rpx 30rpx;
	text-align: center;
	border-top: 1rpx solid #F0F0F0;
	
	.footer-text {
		font-size: 22rpx;
		color: #BBB;
	}
}
</style>
