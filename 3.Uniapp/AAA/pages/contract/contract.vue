<template>
	<view class="page">
		<!-- 状态栏占位 -->
		<view :style="{ height: statusBarHeight + 'px' }" class="status-bar"></view>
		
		<!-- 顶部导航 Tab -->
		<view class="top-nav">
			<view class="nav-left">
				<view class="nav-item" :class="{ active: mainTab === 'perpetual' }" @click="switchMainTab('perpetual')">永续合约</view>
				<view class="nav-item" :class="{ active: mainTab === 'seconds' }" @click="switchMainTab('seconds')">交割合约</view>
			</view>
		</view>

		<view class="scroll-container">
			<scroll-view scroll-y class="content">
				<!-- 币种信息栏 -->
				<view class="symbol-header">
					<view class="symbol-left" @click="showSymbolPicker = true">
						<text class="symbol-name">{{ currentSymbol }}/USDT</text>
						<view class="perpetual-tag" v-if="mainTab === 'perpetual'">永续</view>
						<view class="perpetual-tag" v-else>秒级</view>
						<view class="dropdown-icon"></view>
						<text class="symbol-change" :class="symbolChangeClass">{{ symbolChange }}</text>
					</view>
					<view class="symbol-right" @click="goToKline">
						<SvgIcon name="chart" :size="22" color="#333" />
					</view>
				</view>

				<!-- 永续合约内容 -->
				<template v-if="mainTab === 'perpetual'">
					<!-- 顶部合约配置栏 -->
					<view class="contract-config-bar">
						<view class="config-left">
							<view class="config-btn leverage" @click="showLeveragePicker">{{ leverageOptions[leverageIndex] }}</view>
						</view>
						<view class="config-right">
							<view class="funding-rate">
								<text class="label">资金费率 (8时)/倒计时</text>
								<text class="value">0.00411%/{{ fundingCountdown }}</text>
							</view>
						</view>
					</view>

					<!-- 交易主区域 -->
					<view class="trade-main">
						<!-- 左侧表单 -->
						<view class="trade-form">
							<!-- 开仓/平仓 切换 -->
							<view class="open-close-tab">
								<view 
									class="tab-item open" 
									:class="{ active: mode === 'open' }"
									@click="mode = 'open'"
								>开仓</view>
								<view 
									class="tab-item close" 
									:class="{ active: mode === 'close' }"
									@click="mode = 'close'"
								>平仓</view>
							</view>

							<!-- 可用余额 -->
							<view class="available-box">
								<text class="label">可用</text>
								<text class="value">0.00 USDT</text>
							</view>
							
							<!-- 订单类型 -->
							<view class="form-item order-type">
								<SvgIcon name="about" :size="14" color="#ccc" />
								<text class="label">市价单</text>
								<view class="dropdown-icon gray"></view>
							</view>

							<!-- 价格输入 (市价状态) -->
							<view class="form-item price-input disabled">
								<text class="placeholder">市价</text>
							</view>

							<!-- 初始保证金输入 -->
							<view class="margin-input-group">
								<view class="stepper-input">
									<view class="minus">—</view>
									<text class="placeholder">初始保证金</text>
									<view class="plus">+</view>
								</view>
								<view class="unit-selector">
									<text>USDT</text>
									<view class="dropdown-icon gray"></view>
								</view>
							</view>

							<!-- 滑动进度条 -->
							<view class="slider-box">
								<view class="slider-line">
									<view class="diamond active"></view>
									<view class="diamond"></view>
									<view class="diamond"></view>
									<view class="diamond"></view>
									<view class="diamond"></view>
								</view>
							</view>

							<!-- 选项勾选 -->
							<view class="check-options">
								<view class="check-item">
									<view class="checkbox"></view>
									<text class="dashed" @click="showSlippageModal = true">滑点容差</text>
								</view>
							</view>

							<!-- 交易操作区 (多/空) -->
							<view class="action-section">
								<view class="info-row">
									<text class="label">可开</text>
									<text class="value">0.00 USDT</text>
								</view>
								<view class="info-row">
									<text class="label dashed">保证金</text>
									<text class="value">0.00 USDT</text>
								</view>
								<button class="action-btn long">
									<text class="main">开多</text>
									<text class="sub">看涨</text>
								</button>

								<view class="info-row mt">
									<text class="label">可开</text>
									<text class="value">0.00 USDT</text>
								</view>
								<view class="info-row">
									<text class="label dashed">保证金</text>
									<text class="value">0.00 USDT</text>
								</view>
								<button class="action-btn short">
									<text class="main">开空</text>
									<text class="sub">看跌</text>
								</button>
							</view>
						</view>

						<!-- 右侧盘口 -->
						<view class="order-book">
							<view class="book-header">
								<view class="col">
									<text>委托价格</text>
									<text class="unit">(USDT)</text>
								</view>
								<view class="col right">
									<text>数量</text>
									<text class="unit">(USDT)</text>
								</view>
							</view>
							
							<!-- 卖盘 -->
							<view class="book-list sell">
								<view class="book-item" v-for="(item, index) in sellList" :key="index">
									<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
									<text class="price">{{ item.price }}</text>
									<text class="amount">{{ item.amount }}</text>
								</view>
							</view>

							<!-- 当前价 -->
							<view class="current-price">
								<text class="price-val green">2,956.86</text>
								<text class="price-sub">2,956.93</text>
							</view>

							<!-- 买盘 -->
							<view class="book-list buy">
								<view class="book-item" v-for="(item, index) in buyList" :key="index">
									<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
									<text class="price">{{ item.price }}</text>
									<text class="amount">{{ item.amount }}</text>
								</view>
							</view>

							<!-- 比例条 -->
							<view class="ratio-bar-section">
								<view class="ratio-text">
									<text class="green">{{ buyRatio }}%</text>
									<text class="red">{{ (100 - buyRatio).toFixed(2) }}%</text>
								</view>
								<view class="bar-container">
									<view class="bar-green" :style="{ width: buyRatio + '%' }"></view>
									<view class="bar-red" :style="{ width: (100 - buyRatio) + '%' }"></view>
								</view>
							</view>

							<!-- 盘口设置 -->
							<view class="book-settings">
								<picker @change="onPrecisionChange" :value="precisionIndex" :range="precisionOptions">
									<view class="precision-picker">
										<text>{{ precisionOptions[precisionIndex] }}</text>
										<view class="dropdown-icon gray"></view>
									</view>
								</picker>
							</view>
						</view>
					</view>
				</template>

				<!-- 秒合约内容 -->
				<template v-else>
					<view class="trade-main seconds-contract">
						<!-- 左侧表单 -->
						<view class="trade-form">
							<!-- 周期选择 -->
							<view class="form-label">结算周期</view>
							<view class="period-selector">
								<view class="period-row">
									<view 
										class="period-item" 
										:class="{ active: periodIndex === 0 }"
										@click="periodIndex = 0"
									>
										<text class="time">{{ periodOptions[0].time }}</text>
										<text class="rate">+{{ periodOptions[0].rate }}%</text>
									</view>
									<view 
										class="period-item" 
										:class="{ active: periodIndex === 1 }"
										@click="periodIndex = 1"
									>
										<text class="time">{{ periodOptions[1].time }}</text>
										<text class="rate">+{{ periodOptions[1].rate }}%</text>
									</view>
								</view>
								<view class="period-row">
									<view 
										class="period-item" 
										:class="{ active: periodIndex === 2 }"
										@click="periodIndex = 2"
									>
										<text class="time">{{ periodOptions[2].time }}</text>
										<text class="rate">+{{ periodOptions[2].rate }}%</text>
									</view>
									<view 
										class="period-item" 
										:class="{ active: periodIndex === 3 }"
										@click="periodIndex = 3"
									>
										<text class="time">{{ periodOptions[3].time }}</text>
										<text class="rate">+{{ periodOptions[3].rate }}%</text>
									</view>
								</view>
								<view class="period-row single">
									<view 
										class="period-item" 
										:class="{ active: periodIndex === 4 }"
										@click="periodIndex = 4"
									>
										<text class="time">{{ periodOptions[4].time }}</text>
										<text class="rate">+{{ periodOptions[4].rate }}%</text>
									</view>
								</view>
							</view>

							<!-- 投入金额 -->
							<view class="form-label">投入金额 <text class="min-tip">(最低 {{ minAmount }} USDT)</text></view>
							<view class="amount-input-box">
								<input class="amount-input" type="digit" v-model="secondsAmount" placeholder="请输入投入金额" />
								<text class="unit">USDT</text>
							</view>
							<view class="quick-amounts">
								<view 
									class="quick-item" 
									v-for="amt in quickAmountOptions" 
									:key="amt"
									@click="secondsAmount = amt"
								>{{ amt }}</view>
							</view>

							<!-- 资产余额 -->
							<view class="balance-info">
								<text class="label">可用余额</text>
								<text class="value">0.00 USDT</text>
							</view>

							<!-- 操作按钮 - 横向排列 -->
							<view class="seconds-actions">
								<button class="action-btn up" @click="handleSecondsTrade('buy')">
									<text>买涨</text>
								</button>
								<button class="action-btn down" @click="handleSecondsTrade('sell')">
									<text>买跌</text>
								</button>
							</view>
						</view>

						<!-- 右侧盘口 (复刻) -->
						<view class="order-book">
							<view class="book-header">
								<view class="col">
									<text>最新价格</text>
									<text class="unit">(USDT)</text>
								</view>
								<view class="col right">
									<text>数量</text>
									<text class="unit">(USDT)</text>
								</view>
							</view>
							
							<view class="book-list sell">
								<view class="book-item" v-for="(item, index) in sellList" :key="index">
									<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
									<text class="price">{{ item.price }}</text>
									<text class="amount">{{ item.amount }}</text>
								</view>
							</view>

							<view class="current-price">
								<text class="price-val green">2,956.86</text>
								<text class="price-sub">指数价格</text>
							</view>

							<view class="book-list buy">
								<view class="book-item" v-for="(item, index) in buyList" :key="index">
									<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
									<text class="price">{{ item.price }}</text>
									<text class="amount">{{ item.amount }}</text>
								</view>
							</view>

							<view class="ratio-bar-section">
								<view class="ratio-text">
									<text class="green">{{ buyRatio }}%</text>
									<text class="red">{{ (100 - buyRatio).toFixed(2) }}%</text>
								</view>
								<view class="bar-container">
									<view class="bar-green" :style="{ width: buyRatio + '%' }"></view>
									<view class="bar-red" :style="{ width: (100 - buyRatio) + '%' }"></view>
								</view>
							</view>
						</view>
					</view>
				</template>

				<!-- 底部委托 Tab -->
				<view class="bottom-tabs">
					<view class="tab-scroll">
						<template v-if="mainTab === 'perpetual'">
							<view 
								class="tab-item" 
								:class="{ active: activeBottomTab === 'position' }"
								@click="activeBottomTab = 'position'"
							>当前仓位 (0)</view>
							<view 
								class="tab-item" 
								:class="{ active: activeBottomTab === 'pending' }"
								@click="activeBottomTab = 'pending'"
							>当前委托 (0)</view>
							<view 
								class="tab-item" 
								:class="{ active: activeBottomTab === 'history' }"
								@click="activeBottomTab = 'history'"
							>历史委托</view>
						</template>
						<template v-else>
							<view 
								class="tab-item" 
								:class="{ active: activeBottomTab === 'seconds_active' }"
								@click="activeBottomTab = 'seconds_active'"
							>进行中 ({{ secondsActiveOrders.length }})</view>
							<view 
								class="tab-item" 
								:class="{ active: activeBottomTab === 'seconds_history' }"
								@click="activeBottomTab = 'seconds_history'"
							>历史结算 ({{ secondsHistoryOrders.length }})</view>
						</template>
					</view>
					<view @click="goToOrders">
						<SvgIcon name="document" :size="20" color="#333" />
					</view>
				</view>

				<!-- 列表容器 -->
				<view class="order-list-container">
					<!-- 秒合约进行中列表 -->
					<template v-if="mainTab === 'seconds' && activeBottomTab === 'seconds_active'">
						<view v-if="secondsActiveOrders.length === 0" class="empty-state">
							<view class="empty-icon">
								<SvgIcon name="record" :size="48" color="#eee" />
							</view>
							<text class="empty-text">暂无进行中的订单</text>
						</view>
						<view v-else class="seconds-order-list">
							<view class="order-item active-order" v-for="order in secondsActiveOrders" :key="order.id" @click="openOrderProgress(order)">
								<!-- 订单头部 -->
								<view class="order-header">
									<view class="order-left">
										<text class="symbol">{{ order.symbol }}</text>
										<text class="side-tag" :class="order.side">{{ order.side === 'buy' ? '买涨' : '买跌' }}</text>
									</view>
									<view class="order-right">
										<view class="countdown-badge">
											<text class="countdown-icon">⏱</text>
											<text class="countdown-value">{{ order.countdown }}s</text>
										</view>
									</view>
								</view>
								
								<!-- 核心数据展示 -->
								<view class="order-core">
									<view class="core-item profit">
										<text class="core-label">预计收益</text>
										<text class="core-value" :class="order.expectedProfit >= 0 ? 'win' : 'lose'">
											{{ order.expectedProfit >= 0 ? '+' : '' }}{{ order.expectedProfit.toFixed(2) }} USDT
										</text>
									</view>
									<view class="core-divider"></view>
									<view class="core-item price">
										<text class="core-label">当前价格</text>
										<text class="core-value" :class="order.currentPrice >= order.entryPrice ? 'green' : 'red'">
											{{ order.currentPrice.toFixed(2) }}
										</text>
									</view>
								</view>
								
								<!-- 订单详情 -->
								<view class="order-details">
									<view class="detail-item">
										<text class="detail-label">投入</text>
										<text class="detail-value">{{ order.amount.toFixed(2) }} USDT</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">周期</text>
										<text class="detail-value">{{ order.period }}s</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">收益率</text>
										<text class="detail-value highlight">{{ order.rate }}%</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">开仓价</text>
										<text class="detail-value">{{ order.entryPrice.toFixed(2) }}</text>
									</view>
								</view>
								
								<!-- 点击提示 -->
								<view class="order-action">
									<text class="action-text">点击查看详情</text>
									<text class="action-arrow">›</text>
								</view>
							</view>
						</view>
					</template>
					
					<!-- 秒合约历史结算列表 -->
					<template v-else-if="mainTab === 'seconds' && activeBottomTab === 'seconds_history'">
						<view v-if="secondsHistoryOrders.length === 0" class="empty-state">
							<view class="empty-icon">
								<SvgIcon name="record" :size="48" color="#eee" />
							</view>
							<text class="empty-text">暂无历史结算记录</text>
						</view>
						<view v-else class="seconds-order-list">
							<view class="order-item history-order" v-for="order in secondsHistoryOrders" :key="order.id">
								<!-- 订单头部 -->
								<view class="order-header">
									<view class="order-left">
										<text class="symbol">{{ order.symbol }}</text>
										<text class="side-tag" :class="order.side">{{ order.side === 'buy' ? '买涨' : '买跌' }}</text>
									</view>
									<view class="order-right">
										<text class="settle-status" :class="order.finalProfit >= 0 ? 'win' : 'lose'">
											{{ order.finalProfit >= 0 ? '盈利' : '亏损' }}
										</text>
									</view>
								</view>
								
								<!-- 结算结果 -->
								<view class="order-result" :class="order.finalProfit >= 0 ? 'win' : 'lose'">
									<text class="result-label">结算收益</text>
									<text class="result-value">
										{{ order.finalProfit >= 0 ? '+' : '' }}{{ order.finalProfit.toFixed(2) }} USDT
									</text>
								</view>
								
								<!-- 订单详情 -->
								<view class="order-details">
									<view class="detail-item">
										<text class="detail-label">投入</text>
										<text class="detail-value">{{ order.amount.toFixed(2) }} USDT</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">周期</text>
										<text class="detail-value">{{ order.period }}s</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">收益率</text>
										<text class="detail-value highlight">{{ order.rate }}%</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">开仓价</text>
										<text class="detail-value">{{ order.entryPrice.toFixed(2) }}</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">结算价</text>
										<text class="detail-value">{{ order.settlePrice.toFixed(2) }}</text>
									</view>
									<view class="detail-item full">
										<text class="detail-label">结算时间</text>
										<text class="detail-value">{{ order.settleTime }}</text>
									</view>
								</view>
							</view>
						</view>
					</template>
					
					<!-- 永续合约空状态 -->
					<template v-else>
						<view class="empty-state">
							<view class="empty-icon">
								<SvgIcon name="record" :size="48" color="#eee" />
							</view>
							<text class="empty-text">暂无记录</text>
						</view>
					</template>
				</view>
				
				<!-- 内部占位符，防止内容被底部栏遮挡 -->
				<view class="tabbar-placeholder"></view>
			</scroll-view>
		</view>
		
		<!-- 自定义底部栏 -->
		<CustomTabbar :current="3" />
		
		<!-- 币种选择弹窗 -->
		<SymbolPicker 
			v-model:show="showSymbolPicker" 
			:currentSymbol="currentSymbol"
			@select="onSymbolSelect"
		/>
		
		<!-- 滑点容差说明弹窗 -->
		<view class="slippage-mask" v-if="showSlippageModal" @click="showSlippageModal = false"></view>
		<view class="slippage-modal" :class="{ show: showSlippageModal }">
			<view class="modal-title">有滑点容差的市价单</view>
			<view class="modal-content">
				<text>启用滑点的市价单将按限价单执行，生效时间设为"立即成交或取消 (IOC)"，即订单成交后，在订单历史中显示为限价单，而非市价单。</text>
				<text class="mt">若订单金额超出滑点容差允许的深度，剩余未成交的部分将取消。</text>
			</view>
			<button class="modal-btn" @click="showSlippageModal = false">确定</button>
		</view>
		
		<!-- 杠杆倍数调整弹窗 -->
		<view class="leverage-mask" v-if="showLeverageModal" @click="showLeverageModal = false"></view>
		<view class="leverage-modal" :class="{ show: showLeverageModal }">
			<view class="modal-header">
				<text class="modal-title">调整杠杆</text>
			</view>
			
			<view class="leverage-selector">
				<view class="leverage-controls">
					<view class="control-btn minus" @click.stop="decreaseLeverage">−</view>
					<text class="leverage-value">{{ leverageOptions[leverageIndex] }}</text>
					<view class="control-btn plus" @click.stop="increaseLeverage">+</view>
				</view>
				
				<view class="leverage-slider">
					<view class="slider-track" @touchstart="onSliderTouchStart" @touchmove="onSliderTouchMove" @touchend="onSliderTouchEnd">
						<view class="slider-progress" :style="{ width: sliderProgress + '%' }"></view>
						<view class="slider-thumb" :class="{ active: isDragging }" :style="{ left: sliderProgress + '%' }"></view>
					</view>
					<view class="slider-labels">
						<text>10x</text>
						<text>30x</text>
						<text>50x</text>
						<text>70x</text>
						<text>90x</text>
						<text>120x</text>
						<text>180x</text>
					</view>
				</view>
				
				<view class="leverage-tips">
					<text class="tip-text">• 当前杠杆倍数最高可开：3,000,000USDT</text>
					<text class="tip-text">杠杆调整将同时影响当前仓位和挂单的杠杆。</text>
					<text class="tip-text">• 选择超过10x杠杆交易会增加强行平仓风险，请注意相关风险。</text>
				</view>
			</view>
			
			<view class="leverage-footer">

			</view>
			
			<button class="leverage-confirm-btn" @click="confirmLeverage">确认</button>
		</view>
		
		<!-- 秒合约进度弹窗 -->
		<view class="seconds-progress-overlay" v-if="showSecondsProgress" @click.self="closeSecondsProgress">
			<view class="seconds-progress-modal">
				<!-- 头部 -->
				<view class="progress-header">
					<view class="header-left">
						<text class="symbol">{{ secondsOrderInfo.symbol }}</text>
						<text class="side-tag" :class="secondsOrderInfo.side">
							{{ secondsOrderInfo.side === 'buy' ? '买涨' : '买跌' }}
						</text>
					</view>
					<view class="close-btn" @click="closeSecondsProgress">
						<text class="close-icon">×</text>
					</view>
				</view>
				
				<!-- 进行中状态 -->
				<template v-if="!isSettled">
					<!-- 倒计时 -->
					<view class="countdown-section">
						<text class="countdown-label">结算倒计时</text>
						<view class="countdown-display">
							<text class="countdown-num">{{ secondsOrderInfo.countdown }}</text>
							<text class="countdown-unit">秒</text>
						</view>
						<view class="period-info">
							<text>周期 {{ secondsOrderInfo.period }}s</text>
							<text class="divider">|</text>
							<text>收益率 {{ secondsOrderInfo.rate }}%</text>
						</view>
					</view>
					
					<!-- 价格信息 -->
					<view class="price-section">
						<view class="price-item">
							<text class="price-label">开仓价格</text>
							<text class="price-value">{{ secondsOrderInfo.entryPrice.toFixed(2) }}</text>
						</view>
						<view class="price-item">
							<text class="price-label">当前价格</text>
							<text class="price-value" :class="secondsOrderInfo.currentPrice >= secondsOrderInfo.entryPrice ? 'green' : 'red'">
								{{ secondsOrderInfo.currentPrice.toFixed(2) }}
							</text>
						</view>
					</view>
					
					<!-- 预计收益 -->
					<view class="profit-section">
						<text class="profit-label">预计收益</text>
						<text class="profit-value" :class="secondsOrderInfo.expectedProfit >= 0 ? 'green' : 'red'">
							{{ secondsOrderInfo.expectedProfit >= 0 ? '+' : '' }}{{ secondsOrderInfo.expectedProfit.toFixed(2) }} USDT
						</text>
					</view>
					
					<!-- 买卖压力进度条 -->
					<view class="pressure-section">
						<text class="pressure-label">买卖压力</text>
						<view class="ratio-bar-section">
							<view class="ratio-text">
								<text class="green">买 {{ secondsOrderInfo.buyPressure }}%</text>
								<text class="red">卖 {{ (100 - secondsOrderInfo.buyPressure).toFixed(2) }}%</text>
							</view>
							<view class="bar-container">
								<view class="bar-green" :style="{ width: secondsOrderInfo.buyPressure + '%' }"></view>
								<view class="bar-red" :style="{ width: (100 - secondsOrderInfo.buyPressure) + '%' }"></view>
							</view>
						</view>
					</view>
					
					<!-- 投入金额 -->
					<view class="amount-section">
						<text class="amount-label">投入金额</text>
						<text class="amount-value">{{ secondsOrderInfo.amount.toFixed(2) }} USDT</text>
					</view>
				</template>
				
				<!-- 已结算状态 -->
				<template v-else>
					<view class="settle-result" :class="secondsOrderInfo.finalProfit >= 0 ? 'win' : 'lose'">
						<text class="result-title">{{ secondsOrderInfo.finalProfit >= 0 ? '恭喜盈利' : '交易亏损' }}</text>
						<text class="result-profit" :class="secondsOrderInfo.finalProfit >= 0 ? 'green' : 'red'">
							{{ secondsOrderInfo.finalProfit >= 0 ? '+' : '' }}{{ secondsOrderInfo.finalProfit.toFixed(2) }} USDT
						</text>
					</view>
					
					<!-- 结算详情 -->
					<view class="settle-details">
						<view class="detail-row">
							<text class="detail-label">交易方向</text>
							<text class="detail-value" :class="secondsOrderInfo.side">
								{{ secondsOrderInfo.side === 'buy' ? '买涨' : '买跌' }}
							</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">开仓价格</text>
							<text class="detail-value">{{ secondsOrderInfo.entryPrice.toFixed(2) }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">结算价格</text>
							<text class="detail-value">{{ secondsOrderInfo.settlePrice.toFixed(2) }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">投入金额</text>
							<text class="detail-value">{{ secondsOrderInfo.amount.toFixed(2) }} USDT</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">结算周期</text>
							<text class="detail-value">{{ secondsOrderInfo.period }}s</text>
						</view>
					</view>
					
					<!-- 关闭按钮 -->
					<button class="close-settle-btn" @click="closeSecondsProgress">确认</button>
				</template>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, getCurrentInstance } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import CustomTabbar from '@/components/CustomTabbar.vue'
import SvgIcon from '@/components/SvgIcon.vue'
import SymbolPicker from '@/components/SymbolPicker.vue'

const statusBarHeight = ref(0)
const timer = ref(null)

// 标签页切换
const mainTab = ref('perpetual') // perpetual: 永续合约, seconds: 秒合约

const mode = ref('open') // open, close
const tradeSide = ref('buy')
const orderType = ref('market')
const amount = ref('')
const limitPrice = ref('')
const leverageOptions = ref(['10x', '20x', '30x', '40x', '50x', '60x', '70x', '80x', '90x', '100x', '120x', '150x', '180x'])
const leverageIndex = ref(4) // 默认 50x
const activeBottomTab = ref('position') // position, pending, history

// 秒合约数据
const periodOptions = ref([
	{ time: '30s', rate: '40' },
	{ time: '60s', rate: '50' },
	{ time: '120s', rate: '60' },
	{ time: '180s', rate: '80' },
	{ time: '300s', rate: '100' }
])
const periodIndex = ref(0)
const secondsAmount = ref('')
const quickAmountOptions = [10, 50, 100, 500, 1000]
const minAmount = 10 // 最低投入金额

const buyRatio = ref(50.00)

// 秒合约订单列表
const secondsActiveOrders = ref([]) // 进行中的订单
const secondsHistoryOrders = ref([]) // 历史结算订单
let orderIdCounter = 1 // 订单ID计数器
const orderTimers = {} // 存储每个订单的定时器

// 秒合约进度弹窗数据
const showSecondsProgress = ref(false)
const isSettled = ref(false) // 是否已结算
const currentOrderId = ref(null) // 当前订单ID
const secondsOrderInfo = ref({
	symbol: 'ETH/USDT',
	side: 'buy', // buy: 买涨, sell: 买跌
	amount: 0,
	period: 30,
	rate: 40,
	entryPrice: 2956.86,
	currentPrice: 2956.86,
	countdown: 30,
	expectedProfit: 0,
	buyPressure: 50,
	finalProfit: 0, // 最终结算收益
	settlePrice: 0  // 结算价格
})
const secondsProgressTimer = ref(null)
const priceUpdateTimer = ref(null)

// 更新订单数据（同步到进行中列表）
const updateOrderInList = (orderId, updates) => {
	const order = secondsActiveOrders.value.find(o => o.id === orderId)
	if (order) {
		Object.assign(order, updates)
	}
	// 如果是当前弹窗显示的订单，同步更新弹窗数据
	if (currentOrderId.value === orderId && showSecondsProgress.value && !isSettled.value) {
		Object.assign(secondsOrderInfo.value, updates)
	}
}

// 价格趋势状态（用于模拟趋势性波动）
const priceTrends = {}

// 启动订单的定时器（独立于弹窗）
const startOrderTimers = (orderId, orderData) => {
	// 清除已有定时器
	if (orderTimers[orderId]) {
		clearInterval(orderTimers[orderId].countdown)
		clearInterval(orderTimers[orderId].priceUpdate)
	}
	
	// 初始化该订单的趋势状态
	priceTrends[orderId] = {
		direction: 0, // -1: 下跌趋势, 0: 震荡, 1: 上涨趋势
		momentum: 0,  // 趋势动量（决定趋势持续时间）
		lastPrice: orderData.entryPrice
	}
	
	// 倒计时定时器
	const countdownTimer = setInterval(() => {
		const order = secondsActiveOrders.value.find(o => o.id === orderId)
		if (!order) {
			clearInterval(countdownTimer)
			return
		}
		
		if (order.countdown > 0) {
			updateOrderInList(orderId, { countdown: order.countdown - 1 })
		} else {
			// 到期结算
			settleOrderById(orderId)
		}
	}, 1000)
	
	// 价格更新定时器（降低频率至1500-2000ms）
	const priceTimer = setInterval(() => {
		const order = secondsActiveOrders.value.find(o => o.id === orderId)
		if (!order) {
			clearInterval(priceTimer)
			delete priceTrends[orderId]
			return
		}
		
		const trend = priceTrends[orderId]
		const basePrice = order.entryPrice
		
		// 趋势动量衰减，偶尔产生新趋势
		if (trend.momentum <= 0) {
			// 30% 概率产生新趋势
			if (Math.random() < 0.3) {
				trend.direction = Math.random() > 0.5 ? 1 : -1
				trend.momentum = Math.floor(Math.random() * 4) + 2 // 持续2-5次更新
			} else {
				trend.direction = 0
			}
		} else {
			trend.momentum--
		}
		
		// 计算价格变动幅度（基于百分比：±0.5% 至 ±1.5%）
		const volatilityPercent = (Math.random() * 1.0 + 0.5) / 100 // 0.5% - 1.5%
		let priceChange = basePrice * volatilityPercent
		
		// 根据趋势方向调整
		if (trend.direction !== 0) {
			// 趋势方向占主导（70%概率跟随趋势）
			if (Math.random() < 0.7) {
				priceChange *= trend.direction
			} else {
				priceChange *= (Math.random() > 0.5 ? 1 : -1)
			}
		} else {
			// 无趋势时随机方向
			priceChange *= (Math.random() > 0.5 ? 1 : -1)
		}
		
		// 基于上次价格计算新价格（平滑过渡）
		const smoothFactor = 0.6 // 平滑系数
		const targetPrice = trend.lastPrice + priceChange
		const newPrice = trend.lastPrice + (targetPrice - trend.lastPrice) * smoothFactor
		
		// 限制价格在入场价的±3%范围内
		const maxDeviation = basePrice * 0.03
		const clampedPrice = Math.max(
			basePrice - maxDeviation,
			Math.min(basePrice + maxDeviation, newPrice)
		)
		
		trend.lastPrice = clampedPrice
		
		// 计算预期收益
		const priceDiff = clampedPrice - basePrice
		const isProfit = order.side === 'buy' ? priceDiff > 0 : priceDiff < 0
		const expectedProfit = isProfit ? order.amount * (order.rate / 100) : -order.amount
		
		// 更新买卖压力（与价格波动同步，变化更平滑）
		const currentPressure = order.buyPressure || 50
		const pressureChange = (Math.random() - 0.5) * 5 // ±2.5%的变化
		const newPressure = Math.max(35, Math.min(65, currentPressure + pressureChange))
		
		updateOrderInList(orderId, {
			currentPrice: clampedPrice,
			expectedProfit: expectedProfit,
			buyPressure: parseFloat(newPressure.toFixed(2))
		})
	}, 1800) // 更新间隔1800ms
	
	orderTimers[orderId] = {
		countdown: countdownTimer,
		priceUpdate: priceTimer
	}
}

// 根据订单ID结算
const settleOrderById = (orderId) => {
	const order = secondsActiveOrders.value.find(o => o.id === orderId)
	if (!order) return
	
	// 清除该订单的定时器
	if (orderTimers[orderId]) {
		clearInterval(orderTimers[orderId].countdown)
		clearInterval(orderTimers[orderId].priceUpdate)
		delete orderTimers[orderId]
	}
	
	// 设置结算数据
	const settledOrder = {
		...order,
		settlePrice: order.currentPrice,
		finalProfit: order.expectedProfit,
		status: 'settled',
		settleTime: new Date().toLocaleString()
	}
	
	// 从进行中列表移除
	const activeIndex = secondsActiveOrders.value.findIndex(o => o.id === orderId)
	if (activeIndex > -1) {
		secondsActiveOrders.value.splice(activeIndex, 1)
	}
	
	// 添加到历史结算列表
	secondsHistoryOrders.value.unshift(settledOrder)
	
	// 如果是当前弹窗显示的订单，更新弹窗状态
	if (currentOrderId.value === orderId && showSecondsProgress.value) {
		secondsOrderInfo.value = settledOrder
		isSettled.value = true
	}
}

// 启动秒合约进度（新建订单）
const startSecondsProgress = (side) => {
	const period = parseInt(periodOptions.value[periodIndex.value].time)
	const rate = parseInt(periodOptions.value[periodIndex.value].rate)
	const amount = parseFloat(secondsAmount.value)
	const entryPrice = 2956.86 + (Math.random() - 0.5) * 10
	const orderId = orderIdCounter++
	const createTime = new Date().toLocaleString()
	
	isSettled.value = false // 重置结算状态
	currentOrderId.value = orderId
	
	const orderData = {
		id: orderId,
		symbol: `${currentSymbol.value}/USDT`,
		side: side,
		amount: amount,
		period: period,
		rate: rate,
		entryPrice: entryPrice,
		currentPrice: entryPrice,
		countdown: period,
		expectedProfit: 0,
		buyPressure: buyRatio.value,
		finalProfit: 0,
		settlePrice: 0,
		createTime: createTime,
		status: 'active' // active: 进行中, settled: 已结算
	}
	
	secondsOrderInfo.value = { ...orderData }
	
	// 添加到进行中列表
	secondsActiveOrders.value.unshift(orderData)
	
	// 启动该订单的独立定时器
	startOrderTimers(orderId, orderData)
	
	showSecondsProgress.value = true
}

// 点击进行中订单打开弹窗
const openOrderProgress = (order) => {
	currentOrderId.value = order.id
	secondsOrderInfo.value = { ...order }
	isSettled.value = false
	showSecondsProgress.value = true
}

// 结算秒合约订单（保留兼容）
const settleSecondsOrder = () => {
	settleOrderById(currentOrderId.value)
}

// 关闭秒合约进度弹窗
const closeSecondsProgress = () => {
	showSecondsProgress.value = false
	// 注意：不清除订单定时器，让订单继续在后台运行
}

// 切换主标签页
const switchMainTab = (tab) => {
	mainTab.value = tab
	if (tab === 'seconds') {
		activeBottomTab.value = 'seconds_active'
	} else {
		activeBottomTab.value = 'position'
	}
}

// 秒合约下单
const handleSecondsTrade = (side) => {
	if (!secondsAmount.value) {
		uni.showToast({ title: '请输入投入金额', icon: 'none' })
		return
	}
	if (parseFloat(secondsAmount.value) < minAmount) {
		uni.showToast({ title: `最低投入 ${minAmount} USDT`, icon: 'none' })
		return
	}
	uni.showLoading({ title: '正在下单...' })
	setTimeout(() => {
		uni.hideLoading()
		uni.showToast({ title: '下单成功', icon: 'success' })
		// 启动秒合约进度弹窗
		setTimeout(() => {
			startSecondsProgress(side)
		}, 500)
	}, 1000)
}

// 滑点容差说明
const showSlippageModal = ref(false)

// 杠杆调整弹窗
const showLeverageModal = ref(false)

// 计算滑块进度 (0-100%)
const sliderProgress = computed(() => {
	return (leverageIndex.value / (leverageOptions.value.length - 1)) * 100
})

// 减少杠杆
const decreaseLeverage = () => {
	if (leverageIndex.value > 0) {
		leverageIndex.value--
	}
}

// 增加杠杆
const increaseLeverage = () => {
	if (leverageIndex.value < leverageOptions.value.length - 1) {
		leverageIndex.value++
	}
}

// 确认杠杆设置
const confirmLeverage = () => {
	showLeverageModal.value = false
	// 这里可以添加实际的杠杆设置逻辑
}

// 显示杠杆调整弹窗
const showLeveragePicker = () => {
	showLeverageModal.value = true
}

// 触摸事件处理相关
const isDragging = ref(false)
const sliderRect = ref(null)

// 初始化滑块区域信息
const initSliderRect = () => {
	// 使用 uni.createSelectorQuery 获取滑块的 DOM 信息
	uni.createSelectorQuery().in(getCurrentInstance().proxy.$scope).select('.leverage-slider .slider-track').boundingClientRect((rect) => {
		if (rect) {
			sliderRect.value = rect
		}
	}).exec()
}

// 触摸开始
const onSliderTouchStart = (e) => {
	isDragging.value = true
	updateLeverageByTouch(e.touches[0].clientX)
}

// 触摸移动
const onSliderTouchMove = (e) => {
	if (!isDragging.value || !sliderRect.value) return
	e.preventDefault() // 防止页面滚动干扰
	updateLeverageByTouch(e.touches[0].clientX)
}

// 触摸结束
const onSliderTouchEnd = () => {
	isDragging.value = false
}

// 根据触摸坐标更新杠杆值
const updateLeverageByTouch = (clientX) => {
	if (!sliderRect.value) return

	// 计算触摸点相对于滑块轨道的偏移量
	const offsetX = clientX - sliderRect.value.left
	// 计算百分比 (0-1)
	const percentage = Math.max(0, Math.min(1, offsetX / sliderRect.value.width))

	// 计算对应的索引
	const maxIndex = leverageOptions.value.length - 1
	const calculatedIndex = Math.round(percentage * maxIndex)

	leverageIndex.value = calculatedIndex
}

// 精度选择相关
const precisionOptions = ref(['0.0001', '0.001', '0.01', '0.1'])
const precisionIndex = ref(2)

const onPrecisionChange = (e) => {
	precisionIndex.value = e.detail.value
}

const onLeverageChange = (e) => {
	leverageIndex.value = e.detail.value
}

// 资金费率倒计时
const fundingCountdown = ref('00:00:00')
const countdownTimer = ref(null)

const updateCountdown = () => {
	const now = new Date()
	// 计算距离下一个整点的时间（美国时间整点与本地整点时间差一致）
	const nextHour = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours() + 1, 0, 0)
	const diff = Math.floor((nextHour - now) / 1000)
	
	const h = Math.floor(diff / 3600).toString().padStart(2, '0')
	const m = Math.floor((diff % 3600) / 60).toString().padStart(2, '0')
	const s = (diff % 60).toString().padStart(2, '0')
	
	fundingCountdown.value = `${h}:${m}:${s}`
}

// 币种选择相关
const showSymbolPicker = ref(false)
const currentSymbol = ref('ETH')
const symbolChange = ref('-0.87%')
const symbolChangeClass = computed(() => symbolChange.value.startsWith('-') ? 'down' : 'up')

const onSymbolSelect = (item) => {
	currentSymbol.value = item.symbol.split('/')[0]
	symbolChange.value = item.change
}

const sellList = ref([])
const buyList = ref([])

// 生成随机盘口数据
const generateOrderBook = () => {
	const basePrice = 2956.86
	const newSellList = []
	const newBuyList = []
	
	for (let i = 0; i < 8; i++) {
		newSellList.push({
			price: (basePrice + (8 - i) * 0.01).toFixed(2),
			amount: (Math.random() * 150 + 10).toFixed(2) + (Math.random() > 0.5 ? 'K' : ''),
			depth: Math.random() * 90 + 10
		})
		newBuyList.push({
			price: (basePrice - (i + 1) * 0.01).toFixed(2),
			amount: (Math.random() * 800 + 10).toFixed(2) + (Math.random() > 0.5 ? 'K' : ''),
			depth: Math.random() * 90 + 10
		})
	}
	sellList.value = newSellList
	buyList.value = newBuyList
	
	// 更新买卖压力比例 (40% - 60% 之间随机波动)
	buyRatio.value = parseFloat((Math.random() * 20 + 40).toFixed(2))
}

const startTimer = () => {
	if (timer.value) clearInterval(timer.value)
	timer.value = setInterval(() => {
		generateOrderBook()
	}, 1500)
}

onMounted(() => {
	uni.hideTabBar()
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 0
	generateOrderBook()
	startTimer()
	
	// 启动倒计时
	updateCountdown()
	countdownTimer.value = setInterval(updateCountdown, 1000)
	
	// 初始化滑块区域信息
	initSliderRect()
})

// 页面显示时检查是否需要切换tab
onShow(() => {
	const targetTab = uni.getStorageSync('contractTab')
	if (targetTab) {
		mainTab.value = targetTab
		uni.removeStorageSync('contractTab')
	}
})

onUnmounted(() => {
	if (timer.value) clearInterval(timer.value)
	if (countdownTimer.value) clearInterval(countdownTimer.value)
	// 清理所有订单的定时器和趋势状态
	Object.keys(orderTimers).forEach(orderId => {
		if (orderTimers[orderId]) {
			clearInterval(orderTimers[orderId].countdown)
			clearInterval(orderTimers[orderId].priceUpdate)
		}
		delete priceTrends[orderId]
	})
})

const goToKline = () => {
	uni.navigateTo({ url: '/pages/trade/kline' })
}

const goTrade = () => {
	uni.switchTab({ url: '/pages/trade/trade' })
}

const goToOrders = () => {
	uni.navigateTo({ url: '/pages/trade/orders' })
}
</script>

<style scoped lang="scss">
.page {
	width: 100%;
	height: 100vh;
	background: #FFFFFF;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.status-bar {
	width: 100%;
	flex-shrink: 0;
	background: #fff;
}

.scroll-container {
	flex: 1;
	height: 0;
	min-height: 0;
	overflow: hidden;
}

.content {
	height: 100%;
}

.top-nav {
	height: 101rpx;
	display: flex;
	align-items: center;
	padding: 0 30rpx;
	flex-shrink: 0;
	border-bottom: 1rpx solid #f5f5f5;

	.nav-left {
		display: flex;
		gap: 40rpx;
		
		.nav-item {
			font-size: 37rpx;
			font-weight: bold;
			color: #999;
			&.active {
				color: #333;
				font-size: 41rpx;
			}
		}
	}
}

.symbol-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 30rpx;
	
	.symbol-left {
		display: flex;
		align-items: baseline;
		gap: 12rpx;
		.symbol-name { font-size: 40rpx; font-weight: bold; color: #333; }
		.perpetual-tag {
			font-size: 20rpx;
			color: #999;
			background: #f5f5f5;
			padding: 4rpx 8rpx;
			border-radius: 4rpx;
			line-height: 1;
			margin-bottom: 6rpx;
		}
		.symbol-change { font-size: 24rpx; &.down { color: #F6465D; } &.up { color: #00c087; } }
	}
}

/* 顶部合约配置栏 */
.contract-config-bar {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0 30rpx 20rpx;
	
	.config-left {
		display: flex;
		gap: 16rpx;
		
		.config-btn {
			height: 56rpx;
			padding: 0 24rpx;
			border: 1rpx solid #eee;
			border-radius: 8rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 26rpx;
			color: #333;
			
			&.mode {
				width: 56rpx;
				padding: 0;
			}
		}
	}
	
	.config-right {
		.funding-rate {
			display: flex;
			flex-direction: column;
			align-items: flex-end;
			
			.label {
				font-size: 20rpx;
				color: #999;
				border-bottom: 1rpx dotted #ccc;
			}
			.value {
				font-size: 22rpx;
				color: #333;
				font-weight: 500;
				margin-top: 4rpx;
			}
		}
	}
}

.trade-main {
	display: flex;
	padding: 0 30rpx;
	gap: 20rpx;
	
	.trade-form {
		flex: 1.1;
		
		/* 开仓/平仓 Tab */
		.open-close-tab {
			display: flex;
			height: 64rpx;
			background: #f5f5f5;
			border-radius: 8rpx;
			overflow: hidden;
			margin-bottom: 24rpx;
			position: relative;
			
			.tab-item {
				flex: 1;
				display: flex;
				align-items: center;
				justify-content: center;
				font-size: 30rpx;
				color: #333;
				z-index: 1;
				transition: all 0.2s;
				
				&.active {
					background: #fff;
					font-weight: 600;
				}
				
				&.open.active {
					clip-path: polygon(0 0, 85% 0, 100% 100%, 0% 100%);
					background: #eee;
				}
				&.close.active {
					clip-path: polygon(15% 0, 100% 0, 100% 100%, 0% 100%);
					background: #eee;
				}
			}
		}

		.available-box {
			display: flex;
			justify-content: space-between;
			margin-bottom: 20rpx;
			font-size: 26rpx;
			.label { color: #999; }
			.value { color: #333; font-weight: 500; }
		}
		
		.form-item {
			height: 72rpx;
			background: #f5f5f5;
			border-radius: 8rpx;
			display: flex;
			align-items: center;
			padding: 0 20rpx;
			margin-bottom: 20rpx;
			
			&.order-type {
				gap: 12rpx;
				.label { flex: 1; font-size: 28rpx; color: #333; }
			}
			
			&.price-input {
				justify-content: center;
				&.disabled {
					.placeholder { color: #999; font-size: 28rpx; }
				}
			}
		}

		.margin-input-group {
			display: flex;
			gap: 12rpx;
			margin-bottom: 30rpx;
			
			.stepper-input {
				flex: 1.5;
				height: 72rpx;
				background: #f5f5f5;
				border-radius: 8rpx;
				display: flex;
				align-items: center;
				padding: 0 20rpx;
				
				.minus, .plus {
					font-size: 32rpx;
					color: #999;
					width: 40rpx;
					text-align: center;
				}
				.placeholder {
					flex: 1;
					text-align: center;
					font-size: 24rpx;
					color: #ccc;
				}
			}
			
			.unit-selector {
				flex: 1;
				height: 72rpx;
				background: #f5f5f5;
				border-radius: 8rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				gap: 8rpx;
				font-size: 26rpx;
				color: #333;
			}
		}
		
		.slider-box {
			margin: 40rpx 0;
			padding: 0 10rpx;
			.slider-line {
				height: 2rpx;
				background: #eee;
				display: flex;
				justify-content: space-between;
				align-items: center;
				position: relative;
				
				.diamond {
					width: 12rpx;
					height: 12rpx;
					background: #fff;
					border: 2rpx solid #eee;
					transform: rotate(45deg);
					&.active {
						border-color: #333;
						width: 16rpx;
						height: 16rpx;
						z-index: 1;
					}
				}
			}
		}

		.check-options {
			display: flex;
			flex-direction: column;
			gap: 24rpx;
			margin-bottom: 30rpx;
			
			.check-item {
				display: flex;
				align-items: center;
				gap: 12rpx;
				
				.checkbox {
					width: 32rpx;
					height: 32rpx;
					border: 2rpx solid #ccc;
					border-radius: 4rpx;
				}
				text {
					font-size: 26rpx;
					color: #333;
					&.dashed {
						border-bottom: 1rpx dotted #ccc;
					}
				}
			}
		}

		.action-section {
			.info-row {
				display: flex;
				justify-content: space-between;
				margin-bottom: 12rpx;
				font-size: 24rpx;
				.label { color: #999; &.dashed { border-bottom: 1rpx dotted #ccc; } }
				.value { color: #333; font-weight: 500; }
				&.mt { margin-top: 30rpx; }
			}
			
			.action-btn {
				height: 88rpx;
				border-radius: 12rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				padding: 0 32rpx;
				color: #fff;
				margin-top: 8rpx;
				border: none;
				position: relative;
				
				.main { font-size: 32rpx; font-weight: bold; }
				.sub { 
					font-size: 24rpx; 
					opacity: 0.9; 
					position: absolute;
					right: 32rpx;
				}
				
				&.long { background: #00B894; }
				&.short { background: #F6465D; }
				
				&::after { border: none; }
			}
		}
	}
	
	.order-book {
		flex: 0.9;
		
		.book-header {
			display: flex;
			justify-content: space-between;
			margin-bottom: 16rpx;
			.col {
				display: flex;
				flex-direction: column;
				font-size: 20rpx;
				color: #999;
				&.right { align-items: flex-end; }
			}
		}
		
		.book-list {
			.book-item {
				height: 44rpx;
				display: flex;
				justify-content: space-between;
				align-items: center;
				position: relative;
				.depth-bar { 
					position: absolute; 
					right: 0; 
					top: 0; 
					bottom: 0; 
					opacity: 0.1; 
					transition: width 0.5s ease-out;
				}
				.price { font-size: 22rpx; font-weight: 500; }
				.amount { font-size: 22rpx; color: #333; }
			}
			&.sell { .book-item .price { color: #F6465D; } .book-item .depth-bar { background: #F6465D; } }
			&.buy { .book-item .price { color: #00B894; } .book-item .depth-bar { background: #00B894; } }
		}
		
		.current-price {
			padding: 20rpx 0;
			display: flex;
			flex-direction: column;
			align-items: center;
			.price-val { font-size: 40rpx; font-weight: bold; &.green { color: #00B894; } }
			.price-sub { font-size: 22rpx; color: #999; margin-top: 4rpx; border-bottom: 1rpx dotted #ccc; }
		}

		.ratio-bar-section {
			margin-top: 20rpx;
			.ratio-text {
				display: flex;
				justify-content: space-between;
				font-size: 20rpx;
				margin-bottom: 8rpx;
				.green { color: #00B894; }
				.red { color: #F6465D; }
			}
			.bar-container {
				height: 6rpx;
				background: #eee;
				border-radius: 3rpx;
				display: flex;
				overflow: hidden;
				.bar-green { 
					background: #00B894; 
					transition: width 0.8s ease-in-out;
				}
				.bar-red { 
					background: #F6465D; 
					margin-left: 2rpx; 
					transition: width 0.8s ease-in-out;
				}
			}
		}

		.book-settings {
			margin-top: 24rpx;
			width: 100%;
			
			uni-picker {
				width: 100%;
			}
			
			.precision-picker {
				width: 100%;
				height: 56rpx;
				background: #f5f5f5;
				border-radius: 4rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				gap: 10rpx;
				font-size: 24rpx;
				color: #333;
			}
		}
	}

	/* 秒合约表单样式 */
	&.seconds-contract {
		.form-label {
			font-size: 26rpx;
			color: #333;
			font-weight: 500;
			margin-bottom: 16rpx;
			
			.min-tip {
				font-size: 22rpx;
				color: #999;
				font-weight: normal;
			}
		}

		.period-selector {
			display: flex;
			flex-direction: column;
			gap: 12rpx;
			margin-bottom: 24rpx;

			.period-row {
				display: flex;
				gap: 12rpx;
				
				&.single {
					.period-item {
						width: calc(50% - 6rpx);
					}
				}
			}

			.period-item {
				flex: 1;
				height: 76rpx;
				background: #f5f5f5;
				border-radius: 8rpx;
				display: flex;
				flex-direction: column;
				align-items: center;
				justify-content: center;
				border: 2rpx solid transparent;

				.time { font-size: 26rpx; color: #333; font-weight: bold; }
				.rate { font-size: 18rpx; color: #00B894; margin-top: 4rpx; }

				&.active {
					background: rgba(59, 130, 246, 0.08);
					border-color: #3B82F6;
					.time { color: #3B82F6; }
					.rate { color: #3B82F6; }
				}
			}
		}

		.amount-input-box {
			height: 80rpx;
			background: #f5f5f5;
			border-radius: 8rpx;
			display: flex;
			align-items: center;
			padding: 0 20rpx;
			margin-bottom: 16rpx;

			.amount-input {
				flex: 1;
				font-size: 30rpx;
				color: #333;
			}
			.unit { font-size: 26rpx; color: #333; font-weight: 500; }
		}

		.quick-amounts {
			display: flex;
			justify-content: space-between;
			margin-bottom: 20rpx;

			.quick-item {
				flex: 1;
				height: 56rpx;
				margin: 0 4rpx;
				background: #f5f5f5;
				border-radius: 4rpx;
				font-size: 24rpx;
				color: #666;
				display: flex;
				align-items: center;
				justify-content: center;
				
				&:first-child { margin-left: 0; }
				&:last-child { margin-right: 0; }
			}
		}

		.balance-info {
			display: flex;
			justify-content: space-between;
			margin-bottom: 24rpx;
			.label { font-size: 24rpx; color: #999; }
			.value { font-size: 24rpx; color: #333; font-weight: bold; }
		}

		.seconds-actions {
			display: flex;
			flex-direction: column;
			gap: 16rpx;

			.action-btn {
				width: 100%;
				height: 80rpx;
				border-radius: 8rpx;
				display: flex;
				align-items: center;
				justify-content: center;
				color: #fff;
				font-size: 28rpx;
				font-weight: bold;
				border: none;
				margin: 0;
				padding: 0;

				&.up { background: #00B894; }
				&.down { background: #F6465D; }
				
				&::after { border: none; }
			}
		}
	}
}

.bottom-tabs {
	margin-top: 40rpx; display: flex; align-items: center; padding: 0 30rpx;
	.tab-scroll {
		flex: 1; display: flex; gap: 40rpx; height: 88rpx; align-items: center;
		.tab-item { font-size: 28rpx; color: #999; font-weight: 500; &.active { color: #333; border-bottom: 4rpx solid #F7D100; padding-bottom: 4rpx; } }
	}
}

.order-list-container {
	min-height: 300rpx;
	padding: 24rpx 30rpx;
	
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 80rpx 0;
		
		.empty-icon {
			margin-bottom: 24rpx;
		}
		
		.empty-text {
			font-size: 26rpx;
			color: #999;
		}
	}
	
	/* 秒合约订单列表样式 */
	.seconds-order-list {
		.order-item {
			background: #fff;
			border-radius: 12rpx;
			padding: 24rpx;
			margin-bottom: 20rpx;
			border: 1rpx solid #eee;
			
			/* 订单头部 */
			.order-header {
				display: flex;
				justify-content: space-between;
				align-items: center;
				margin-bottom: 16rpx;
				
				.order-left {
					display: flex;
					align-items: center;
					gap: 12rpx;
					
					.symbol {
						font-size: 30rpx;
						font-weight: bold;
						color: #333;
					}
					
					.side-tag {
						font-size: 20rpx;
						padding: 4rpx 12rpx;
						border-radius: 4rpx;
						font-weight: 500;
						
						&.buy {
							background: rgba(0, 184, 148, 0.1);
							color: #00B894;
						}
						&.sell {
							background: rgba(246, 70, 93, 0.1);
							color: #F6465D;
						}
					}
				}
				
				.order-right {
					.countdown-badge {
						display: flex;
						align-items: center;
						gap: 4rpx;
						
						.countdown-icon {
							display: none;
						}
						
						.countdown-value {
							font-size: 32rpx;
							font-weight: bold;
							color: #3B82F6;
						}
					}
					
					.settle-status {
						font-size: 22rpx;
						font-weight: 500;
						padding: 4rpx 12rpx;
						border-radius: 4rpx;
						
						&.win {
							background: rgba(0, 184, 148, 0.1);
							color: #00B894;
						}
						&.lose {
							background: rgba(246, 70, 93, 0.1);
							color: #F6465D;
						}
					}
				}
			}
			
			/* 核心数据展示 */
			.order-core {
				display: flex;
				align-items: stretch;
				background: #f8f9fa;
				border-radius: 8rpx;
				padding: 16rpx;
				margin-bottom: 16rpx;
				
				.core-item {
					flex: 1;
					text-align: center;
					
					.core-label {
						display: block;
						font-size: 22rpx;
						color: #999;
						margin-bottom: 6rpx;
					}
					
					.core-value {
						display: block;
						font-size: 30rpx;
						font-weight: bold;
						color: #333;
						
						&.win { color: #00B894; }
						&.lose { color: #F6465D; }
						&.green { color: #00B894; }
						&.red { color: #F6465D; }
					}
				}
				
				.core-divider {
					width: 1rpx;
					background: #e5e5e5;
					margin: 0 16rpx;
				}
			}
			
			/* 结算结果展示 */
			.order-result {
				display: flex;
				justify-content: space-between;
				align-items: center;
				padding: 16rpx 20rpx;
				border-radius: 8rpx;
				margin-bottom: 16rpx;
				
				&.win {
					background: rgba(0, 184, 148, 0.08);
				}
				
				&.lose {
					background: rgba(246, 70, 93, 0.08);
				}
				
				.result-label {
					font-size: 24rpx;
					color: #666;
				}
				
				.result-value {
					font-size: 32rpx;
					font-weight: bold;
					color: #00B894;
				}
			}
			
			&.history-order .order-result.lose .result-value {
				color: #F6465D;
			}
			
			&.history-order .order-result.win .result-value {
				color: #00B894;
			}
			
			/* 订单详情 */
			.order-details {
				display: flex;
				flex-wrap: wrap;
				padding-top: 12rpx;
				border-top: 1rpx solid #f0f0f0;
				
				.detail-item {
					width: 50%;
					display: flex;
					justify-content: space-between;
					align-items: center;
					padding: 8rpx 0;
					padding-right: 16rpx;
					box-sizing: border-box;
					
					&:nth-child(even) {
						padding-right: 0;
						padding-left: 16rpx;
					}
					
					&.full {
						width: 100%;
						padding-right: 0;
						padding-left: 0;
						margin-top: 4rpx;
					}
					
					.detail-label {
						font-size: 24rpx;
						color: #999;
					}
					
					.detail-value {
						font-size: 24rpx;
						color: #333;
						font-weight: 500;
						
						&.highlight {
							color: #3B82F6;
						}
						
						&.green { color: #00B894; }
						&.red { color: #F6465D; }
					}
				}
			}
			
			/* 点击提示 */
			.order-action {
				display: flex;
				justify-content: center;
				align-items: center;
				gap: 6rpx;
				margin-top: 16rpx;
				padding-top: 12rpx;
				border-top: 1rpx dashed #eee;
				
				.action-text {
					font-size: 22rpx;
					color: #3B82F6;
				}
				
				.action-arrow {
					font-size: 28rpx;
					color: #3B82F6;
				}
			}
		}
	}
}

.dropdown-icon {
	width: 0; height: 0; border-left: 8rpx solid transparent; border-right: 8rpx solid transparent; border-top: 10rpx solid #333; margin-top: 6rpx;
	&.gray { border-top-color: #999; }
}

.tabbar-placeholder { height: 120rpx; padding-bottom: env(safe-area-inset-bottom); }
.placeholder { color: #ccc; }

/* 滑点容差弹窗 */
.slippage-mask {
	position: fixed;
	left: 0;
	right: 0;
	top: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 999;
}

.slippage-modal {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	background: #fff;
	border-radius: 32rpx 32rpx 0 0;
	padding: 48rpx 40rpx calc(48rpx + env(safe-area-inset-bottom));
	z-index: 1000;
	transform: translateY(100%);
	transition: transform 0.3s ease;
	
	&.show {
		transform: translateY(0);
	}
	
	.modal-title {
		font-size: 36rpx;
		font-weight: 600;
		color: #333;
		margin-bottom: 32rpx;
	}
	
	.modal-content {
		display: flex;
		flex-direction: column;
		font-size: 28rpx;
		color: #666;
		line-height: 1.7;
		
		.mt {
			margin-top: 24rpx;
		}
	}
	
	.modal-btn {
		margin-top: 48rpx;
		width: 100%;
		height: 96rpx;
		background: #0052FF;
		border-radius: 48rpx;
		color: #fff;
		font-size: 32rpx;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		
		&::after {
			border: none;
		}
	}
}

/* 杠杆调整弹窗 */
.leverage-mask {
	position: fixed;
	left: 0;
	right: 0;
	top: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 999;
}

.leverage-modal {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	background: #fff;
	border-radius: 32rpx 32rpx 0 0;
	padding: 48rpx 40rpx calc(48rpx + env(safe-area-inset-bottom));
	z-index: 1000;
	transform: translateY(100%);
	transition: transform 0.3s ease;
	
	&.show {
		transform: translateY(0);
	}
	
	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 32rpx;
		
		.modal-title {
			font-size: 36rpx;
			font-weight: 600;
			color: #333;
		}
	}
	
	.leverage-selector {
		.leverage-controls {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 32rpx;
			margin-bottom: 48rpx;
			
			.control-btn {
				width: 80rpx;
				height: 80rpx;
				border-radius: 50%;
				display: flex;
				align-items: center;
				justify-content: center;
				font-size: 40rpx;
				font-weight: bold;
				background: #f5f5f5;
				color: #333;
				
				&.minus {
					transform: translateY(-4rpx);
				}
				
				&.plus {
					transform: translateY(4rpx);
				}
			}
			
			.leverage-value {
				font-size: 64rpx;
				font-weight: bold;
				color: #333;
			}
		}
		
		.leverage-slider {
			margin-bottom: 48rpx;
			
			.slider-track {
				height: 8rpx;
				background: #f0f0f0;
				border-radius: 4rpx;
				position: relative;
				margin-bottom: 20rpx;
				
				.slider-progress {
					height: 100%;
					background: #007AFF; /* Coinbase blue */
					border-radius: 4rpx;
				}
				
				.slider-thumb {
					width: 40rpx;
					height: 40rpx;
					background: #007AFF; /* Coinbase blue */
					border-radius: 50%;
					position: absolute;
					top: 50%;
					transform: translate(-50%, -50%);
					box-shadow: 0 4rpx 12rpx rgba(0, 122, 255, 0.3);
					transition: background-color 0.2s ease, transform 0.2s ease; /* 添加过渡效果 */
							
					&.active {
						background-color: #0051C9; /* 拖动时更深的蓝色 */
						transform: translate(-50%, -50%) scale(1.1); /* 拖动时稍微放大 */
					}
				}
			}
			
			.slider-labels {
				display: flex;
				justify-content: space-between;
				font-size: 24rpx;
				color: #999;
			}
		}
		
		.leverage-tips {
			.tip-text {
				display: block;
				font-size: 24rpx;
				color: #666;
				line-height: 1.6;
				margin-bottom: 16rpx;
				
				&:last-child {
					margin-bottom: 0;
				}
				
				.link {
					color: #007AFF; /* Coinbase blue */
					text-decoration: underline;
				}
			}
		}
	}
	
	.leverage-footer {
		margin: 48rpx 0;
		padding: 32rpx 0;
		border-top: 1rpx solid #f0f0f0;
		
		.footer-link {
			display: flex;
			justify-content: space-between;
			align-items: center;
			font-size: 28rpx;
			color: #333;
			
			.link-arrow {
				color: #999;
			}
		}
	}
	
	.leverage-confirm-btn {
		width: 100%;
		height: 96rpx;
		background: #007AFF; /* Coinbase blue */
		border-radius: 48rpx;
		color: #fff;
		font-size: 32rpx;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		
		&::after {
			border: none;
		}
	}
}

/* 秒合约进度弹窗样式 */
.seconds-progress-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.6);
	z-index: 9999;
	display: flex;
	align-items: flex-start;
	justify-content: center;
	padding-top: 120rpx;
}

.seconds-progress-modal {
	width: 90%;
	max-width: 680rpx;
	background: #fff;
	border-radius: 24rpx;
	padding: 32rpx;
	box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.2);
	
	.progress-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 32rpx;
		
		.header-left {
			display: flex;
			align-items: center;
			gap: 16rpx;
			
			.symbol {
				font-size: 34rpx;
				font-weight: bold;
				color: #333;
			}
			
			.side-tag {
				font-size: 22rpx;
				padding: 6rpx 16rpx;
				border-radius: 6rpx;
				font-weight: 500;
				
				&.buy {
					background: rgba(0, 184, 148, 0.1);
					color: #00B894;
				}
				&.sell {
					background: rgba(246, 70, 93, 0.1);
					color: #F6465D;
				}
			}
		}
		
		.close-btn {
			width: 60rpx;
			height: 60rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			background: #f0f0f0;
			border-radius: 50%;
			transition: all 0.2s ease;
			
			&:active {
				background: #e0e0e0;
				transform: scale(0.95);
			}
			
			.close-icon {
				font-size: 44rpx;
				color: #666;
				font-weight: 300;
				line-height: 1;
				margin-top: -4rpx;
			}
		}
	}
	
	.countdown-section {
		text-align: center;
		padding: 32rpx 0;
		background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
		border-radius: 16rpx;
		margin-bottom: 32rpx;
		
		.countdown-label {
			font-size: 24rpx;
			color: #999;
			display: block;
			margin-bottom: 16rpx;
		}
		
		.countdown-display {
			display: flex;
			align-items: baseline;
			justify-content: center;
			
			.countdown-num {
				font-size: 80rpx;
				font-weight: bold;
				color: #3B82F6;
				line-height: 1;
			}
			
			.countdown-unit {
				font-size: 28rpx;
				color: #666;
				margin-left: 8rpx;
			}
		}
		
		.period-info {
			margin-top: 16rpx;
			font-size: 24rpx;
			color: #666;
			
			.divider {
				margin: 0 16rpx;
				color: #ddd;
			}
		}
	}
	
	.price-section {
		display: flex;
		justify-content: space-between;
		padding: 24rpx 0;
		border-bottom: 1rpx solid #f0f0f0;
		
		.price-item {
			.price-label {
				display: block;
				font-size: 24rpx;
				color: #999;
				margin-bottom: 8rpx;
			}
			
			.price-value {
				font-size: 32rpx;
				font-weight: bold;
				color: #333;
				
				&.green { color: #00B894; }
				&.red { color: #F6465D; }
			}
		}
	}
	
	.profit-section {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24rpx 0;
		border-bottom: 1rpx solid #f0f0f0;
		
		.profit-label {
			font-size: 26rpx;
			color: #666;
		}
		
		.profit-value {
			font-size: 36rpx;
			font-weight: bold;
			
			&.green { color: #00B894; }
			&.red { color: #F6465D; }
		}
	}
	
	.pressure-section {
		padding: 24rpx 0;
		border-bottom: 1rpx solid #f0f0f0;
		
		.pressure-label {
			display: block;
			font-size: 24rpx;
			color: #999;
			margin-bottom: 16rpx;
		}
		
		.ratio-bar-section {
			.ratio-text {
				display: flex;
				justify-content: space-between;
				font-size: 24rpx;
				margin-bottom: 12rpx;
				
				.green { color: #00B894; }
				.red { color: #F6465D; }
			}
			
			.bar-container {
				display: flex;
				height: 12rpx;
				border-radius: 6rpx;
				overflow: hidden;
				
				.bar-green {
					background: #00B894;
					transition: width 0.3s ease;
				}
				
				.bar-red {
					background: #F6465D;
					transition: width 0.3s ease;
				}
			}
		}
	}
	
	.amount-section {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 24rpx;
		
		.amount-label {
			font-size: 26rpx;
			color: #666;
		}
		
		.amount-value {
			font-size: 28rpx;
			font-weight: bold;
			color: #333;
		}
	}
	
	/* 结算结果样式 */
	.settle-result {
		text-align: center;
		padding: 48rpx 0;
		
		.result-title {
			display: block;
			font-size: 36rpx;
			font-weight: bold;
			color: #333;
			margin-bottom: 16rpx;
		}
		
		.result-profit {
			display: block;
			font-size: 56rpx;
			font-weight: bold;
			
			&.green { color: #00B894; }
			&.red { color: #F6465D; }
		}
		
		&.win {
			background: linear-gradient(135deg, rgba(0, 184, 148, 0.08) 0%, rgba(0, 184, 148, 0.02) 100%);
			border-radius: 16rpx;
		}
		
		&.lose {
			background: linear-gradient(135deg, rgba(246, 70, 93, 0.08) 0%, rgba(246, 70, 93, 0.02) 100%);
			border-radius: 16rpx;
		}
	}
	
	.settle-details {
		margin-top: 32rpx;
		padding: 24rpx;
		background: #f8f9fa;
		border-radius: 12rpx;
		
		.detail-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 16rpx 0;
			border-bottom: 1rpx solid #eee;
			
			&:last-child {
				border-bottom: none;
			}
			
			.detail-label {
				font-size: 26rpx;
				color: #999;
			}
			
			.detail-value {
				font-size: 26rpx;
				color: #333;
				font-weight: 500;
				
				&.buy { color: #00B894; }
				&.sell { color: #F6465D; }
			}
		}
	}
	
	.close-settle-btn {
		margin-top: 32rpx;
		width: 100%;
		height: 88rpx;
		background: #3B82F6;
		border-radius: 12rpx;
		color: #fff;
		font-size: 30rpx;
		font-weight: bold;
		border: none;
		display: flex;
		align-items: center;
		justify-content: center;
		
		&::after { border: none; }
	}
}
</style>
