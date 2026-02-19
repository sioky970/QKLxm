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
						<text class="symbol-name">{{ currentSymbol }}</text>
						<view class="perpetual-tag" v-if="mainTab === 'perpetual'">永续</view>
						<view class="perpetual-tag" v-else>秒级</view>
						<view class="dropdown-icon"></view>
						<text class="symbol-change" :class="symbolChangeClass">{{ symbolChange }}</text>
					</view>
					<view class="symbol-right" @click="goToKline">
						<SvgIcon name="chart" :size="22" color="#333" />
					</view>
				</view>
				
				<!-- 币种选择弹窗 -->
				<SymbolPicker :show="showSymbolPicker" @close="showSymbolPicker = false" @select="onSymbolSelect" />

				<!-- 永续合约内容 -->
				<template v-if="mainTab === 'perpetual'">
					<!-- 顶部合约配置栏 -->
					<view class="contract-config-bar">
						<view class="config-left">
							<text class="config-label">杠杆倍数</text>
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

							<!-- 可用余额 -->
							<view class="available-box">
								<text class="label">可用</text>
								<view class="value-with-btn">
									<text class="value">{{ availableBalance }} USDT</text>
									<view class="add-icon" @click.stop="goToTransfer">+</view>
								</view>
							</view>
							
							<!-- 订单类型 -->
							<picker @change="onOrderTypeChange" :value="orderTypeIndex" :range="orderTypeOptions">
								<view class="form-item order-type">
									<SvgIcon name="about" :size="14" color="#999" />
									<text class="label">{{ orderTypeOptions[orderTypeIndex] }}</text>
									<view class="dropdown-icon gray"></view>
								</view>
							</picker>
							
							<!-- 价格输入 -->
							<view class="form-item price-input" :class="{ disabled: orderType === 'market' }">
								<input 
									v-if="orderType === 'limit'"
									type="digit" 
									v-model="limitPrice"
									placeholder="价格" 
									placeholder-class="placeholder" 
								/>
								<text v-else class="placeholder">市价</text>
							</view>

							<!-- 初始保证金输入 -->
							<view class="input-group">
								<input 
									type="digit" 
									:value="marginAmount"
									@input="onMarginInput"
									placeholder="初始保证金" 
									placeholder-class="placeholder"
									:adjust-position="false"
								/>
								<view class="unit-picker">
									<text>USDT</text>
									<view class="dropdown-icon gray"></view>
								</view>
							</view>

							<!-- 滑动进度条 -->
							<view class="slider-box">
								<slider 
									:value="sliderValue" 
									@change="onSliderChange"
									@changing="onSliderChanging"
									:min="0" 
									:max="100" 
									:step="1"
									:show-value="false"
									class="custom-slider"
									active-color="#3B82F6"
									background-color="#E5E7EB"
								/>
								<view class="slider-marks">
									<text v-for="(percent, index) in sliderPercents" :key="index" class="mark">{{ percent }}%</text>
								</view>
							</view>

							<!-- 选项勾选 -->
							<view class="check-options">
								<view class="check-item" @click="toggleSlippage">
									<view class="checkbox" :class="{ checked: slippageEnabled }">
										<view v-if="slippageEnabled" class="checkbox-icon">✓</view>
									</view>
									<text class="dashed" @click.stop="showSlippageModal = true">滑点容差</text>
								</view>
							</view>

							<!-- 交易操作区 (多/空) -->
							<view class="action-section">
								<view class="info-row">
									<text class="label dashed">保证金</text>
									<text class="value">{{ marginAmount || '0.00' }} USDT</text>
								</view>
								<button class="action-btn long" @click="handleOpenPosition('long')" :disabled="isSubmitting">
									<text class="main">{{ isSubmitting ? '处理中...' : '开多' }}</text>
									<text class="sub">看涨</text>
								</button>
																
								<view class="info-row mt">
									<text class="label dashed">保证金</text>
									<text class="value">{{ marginAmount || '0.00' }} USDT</text>
								</view>
								<button class="action-btn short" @click="handleOpenPosition('short')" :disabled="isSubmitting">
									<text class="main">{{ isSubmitting ? '处理中...' : '开空' }}</text>
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
								<view class="book-item" v-for="(item, index) in sellList" :key="index" @click="selectPrice(item.price)">
									<view class="depth-bar" :style="{ width: item.depth + '%' }"></view>
									<text class="price">{{ item.price }}</text>
									<text class="amount">{{ item.amount }}</text>
								</view>
							</view>
							
							<!-- 当前价 -->
							<view class="current-price" @click="selectPrice(currentPrice)">
								<text class="price-val" :class="[symbolChangeClass, priceChangeClass]">{{ currentPriceFormatted }}</text>
								<text class="price-sub">≈ ${{ currentPriceUsd }}</text>
							</view>
							
							<!-- 买盘 -->
							<view class="book-list buy">
								<view class="book-item" v-for="(item, index) in buyList" :key="index" @click="selectPrice(item.price)">
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
								<view class="value-with-btn">
									<text class="value">{{ availableBalance }} USDT</text>
									<view class="add-icon" @click.stop="goToTransfer">+</view>
								</view>
							</view>

							<!-- 操作按钮 - 横向排列 -->
							<view class="seconds-actions">
								<button class="action-btn long" @click="handleSecondsTrade('buy')" :disabled="isSubmitting">
									<text class="main">{{ isSubmitting ? '处理中...' : '开多' }}</text>
									<text class="sub">看涨</text>
								</button>
								<button class="action-btn short" @click="handleSecondsTrade('sell')" :disabled="isSubmitting">
									<text class="main">{{ isSubmitting ? '处理中...' : '开空' }}</text>
									<text class="sub">看跌</text>
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
								<text class="price-val" :class="[symbolChangeClass, priceChangeClass]">{{ currentPriceFormatted }}</text>
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
							>当前仓位 ({{ positions.length }})</view>
							<view 
								class="tab-item" 
								:class="{ active: activeBottomTab === 'pending' }"
								@click="activeBottomTab = 'pending'"
							>当前委托 ({{ pendingOrders.length }})</view>
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
										<text class="side-tag" :class="order.side">{{ order.side === 'buy' ? '开多' : '开空' }}</text>
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
											{{ formatPrice(order.currentPrice) }}
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
										<text class="detail-value">{{ formatPrice(order.entryPrice) }}</text>
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
										<text class="side-tag" :class="order.side">{{ order.side === 'buy' ? '开多' : '开空' }}</text>
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
										<text class="detail-value">{{ formatPrice(order.entryPrice) }}</text>
									</view>
									<view class="detail-item">
										<text class="detail-label">结算价</text>
										<text class="detail-value">{{ formatPrice(order.settlePrice) }}</text>
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
						<!-- 当前仓位列表 -->
						<template v-if="activeBottomTab === 'position'">
							<view v-if="positions.length === 0" class="empty-state">
								<view class="empty-icon">
									<SvgIcon name="record" :size="48" color="#eee" />
								</view>
								<text class="empty-text">暂无持仓</text>
							</view>
							<view v-else class="position-list">
								<view class="position-item" v-for="pos in positions" :key="pos.id">
									<view class="position-header">
										<view class="left">
											<text class="symbol">{{ getCurrencyNameById(pos.currency_id) }}/USDT</text>
											<text class="side-tag" :class="pos.type === 1 ? 'long' : 'short'">
												{{ pos.type === 1 ? '多' : '空' }} {{ pos.leverage || pos.multiple }}x
											</text>
										</view>
										<view class="pnl" :class="getPositionPnL(pos) >= 0 ? 'profit' : 'loss'">
											{{ getPositionPnL(pos) >= 0 ? '+' : '' }}{{ getPositionPnL(pos).toFixed(2) }} USDT
											<text class="rate">({{ getPositionPnLRate(pos) }}%)</text>
										</view>
									</view>
									<view class="position-info">
										<view class="info-item">
											<text class="label">开仓价</text>
											<text class="value">{{ formatPrice(pos.entry_price || pos.price) }}</text>
										</view>
										<view class="info-item">
											<text class="label">保证金</text>
											<text class="value">{{ formatPrice(pos.margin || pos.caution_money || pos.origin_margin) }}</text>
										</view>
										<view class="info-item">
											<text class="label">止盈/止损</text>
											<text class="value">
												{{ pos.take_profit_amount ? '+' + pos.take_profit_amount : '--' }} / 
												{{ pos.stop_loss_amount ? '-' + pos.stop_loss_amount : '--' }}
											</text>
										</view>
									</view>
									<view class="position-actions">
										<button class="btn tpsl" @click="openTPSLModal(pos)">止盈止损</button>
										<button class="btn close" @click="handleClosePosition(pos)">平仓</button>
									</view>
								</view>
							</view>
						</template>
												
						<!-- 当前委托列表 -->
						<template v-else-if="activeBottomTab === 'pending'">
							<view v-if="pendingOrders.length === 0" class="empty-state">
								<view class="empty-icon">
									<SvgIcon name="record" :size="48" color="#eee" />
								</view>
								<text class="empty-text">暂无挂单</text>
							</view>
							<view v-else class="pending-list">
								<view class="pending-item" v-for="order in pendingOrders" :key="order.id">
									<view class="pending-header">
										<view class="left">
											<text class="symbol">{{ getCurrencyNameById(order.currency_id) }}/USDT</text>
											<text class="side-tag" :class="order.type === 1 ? 'long' : 'short'">
												{{ order.type === 1 ? '多' : '空' }} {{ order.leverage || order.multiple }}x
											</text>
											<text class="order-type">限价单</text>
										</view>
										<text class="status">待成交</text>
									</view>
									<view class="pending-info">
										<view class="info-item">
											<text class="label">限价</text>
											<text class="value">{{ formatPrice(order.limit_price) }}</text>
										</view>
										<view class="info-item">
											<text class="label">保证金</text>
											<text class="value">{{ formatPrice(order.origin_margin || order.origin_caution_money) }}</text>
										</view>
									</view>
									<view class="pending-actions">
										<button class="btn chase" @click="handleChaseOrder(order)">追单</button>
										<button class="btn cancel" @click="handleCancelOrder(order)">取消</button>
									</view>
								</view>
							</view>
						</template>
												
						<!-- 默认空状态 -->
						<template v-else>
							<view class="empty-state">
								<view class="empty-icon">
									<SvgIcon name="record" :size="48" color="#eee" />
								</view>
								<text class="empty-text">暂无记录</text>
							</view>
						</template>
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
		
		<!-- 止盈止损设置弹窗 -->
		<view class="tpsl-mask" v-if="showTPSLModal" @click="closeTPSLModal"></view>
		<view class="tpsl-modal" :class="{ show: showTPSLModal }">
			<view class="modal-header">
				<text class="modal-title">设置止盈止损</text>
				<view class="close-btn" @click="closeTPSLModal">×</view>
			</view>
			
			<view class="tpsl-form">
				<view class="form-item">
					<text class="label">止盈金额 (USDT)</text>
					<input type="digit" v-model="tpslForm.takeProfitAmount" placeholder="输入止盈金额" />
					<text class="hint">当盈利达到此金额时自动平仓</text>
				</view>
				
				<view class="form-item">
					<text class="label">止损金额 (USDT)</text>
					<input type="digit" v-model="tpslForm.stopLossAmount" placeholder="输入止损金额" />
					<text class="hint">当亏损达到此金额时自动平仓</text>
				</view>
			</view>
			
			<button class="tpsl-confirm-btn" @click="submitTPSL">确认设置</button>
		</view>
		
		<!-- 秒合约进度弹窗 -->
		<view class="seconds-progress-overlay" v-if="showSecondsProgress" @click.self="closeSecondsProgress">
			<view class="seconds-progress-modal">
				<!-- 头部 -->
				<view class="progress-header">
					<view class="header-left">
						<text class="symbol">{{ secondsOrderInfo.symbol }}</text>
						<text class="side-tag" :class="secondsOrderInfo.side">
							{{ secondsOrderInfo.side === 'buy' ? '开多' : '开空' }}
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
							<text class="price-value">{{ formatPrice(secondsOrderInfo.entryPrice) }}</text>
						</view>
						<view class="price-item">
							<text class="price-label">当前价格</text>
							<text class="price-value" :class="secondsOrderInfo.currentPrice >= secondsOrderInfo.entryPrice ? 'green' : 'red'">
								{{ formatPrice(secondsOrderInfo.currentPrice) }}
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
						<text class="result-profit" :class="secondsOrderInfo.finalProfit >= 0 ? 'green' : 'red'">
							{{ secondsOrderInfo.finalProfit >= 0 ? '+' : '' }}{{ secondsOrderInfo.finalProfit.toFixed(2) }} USDT
						</text>
					</view>
					
					<!-- 结算详情 -->
					<view class="settle-details">
						<view class="detail-row">
							<text class="detail-label">交易方向</text>
							<text class="detail-value" :class="secondsOrderInfo.side">
								{{ secondsOrderInfo.side === 'buy' ? '开多' : '开空' }}
							</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">开仓价格</text>
							<text class="detail-value">{{ formatPrice(secondsOrderInfo.entryPrice) }}</text>
						</view>
						<view class="detail-row">
							<text class="detail-label">结算价格</text>
							<text class="detail-value">{{ formatPrice(secondsOrderInfo.settlePrice) }}</text>
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
import { ref, computed, onMounted, onUnmounted, watch, getCurrentInstance } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import CustomTabbar from '@/components/CustomTabbar.vue'
import SvgIcon from '@/components/SvgIcon.vue'
import SymbolPicker from '@/components/SymbolPicker.vue'
import marketStore from '@/stores/marketStore.js'
import walletStore from '@/stores/walletStore.js'
import { getLeverageMultiples, contractOpenPosition, contractClosePosition, contractChaseOrder, contractSetTPSL, contractCancelOrder, getContractPositions, getContractPendingOrders, getContractPositionDetail, getContractHistoryPositions, getMicroPeriods, submitMicroOrder, getMicroActiveOrders, getMicroOrderList, getMicroOrderDetail } from '@/utils/api.js'
import wsClient from '@/utils/websocket.js'

const statusBarHeight = ref(0)
const timer = ref(null)

// 标签页切换
const mainTab = ref('perpetual') // perpetual: 永续合约, seconds: 秒合约

// 已删除 mode 变量（开仓/平仓切换）
const tradeSide = ref('buy')
const orderType = ref('market') // 'market' 或 'limit'
const amount = ref('')
const limitPrice = ref('')

// 用户资产数据（从wallet store获取）
const userAssets = computed(() => walletStore.state.assets)
const usdtBalance = computed(() => walletStore.state.usdtBalance)
const contractBalance = computed(() => walletStore.state.contractBalance)
const deliveryBalance = computed(() => walletStore.state.deliveryBalance)
const currentCoinBalance = computed(() => {
	const coinName = currentCoinName.value
	const coinAsset = walletStore.state.assets.find(asset => asset.name === coinName)
	return coinAsset ? coinAsset.balance : 0
})

// 币种选择相关
const showSymbolPicker = ref(false)
const currentSymbol = ref('ETH/USDT')
const currentSymbolData = ref(null) // 存储完整的币种数据

// 当前币种的市场数据（从全局状态获取）
const currentMarketData = computed(() => {
	if (!currentSymbolData.value) return null
	// 使用 lastUpdateTime 作为响应式触发器，确保 Map 更新时重新计算
	// eslint-disable-next-line no-unused-vars
	const _trigger = marketStore.state.lastUpdateTime
	const key = `${currentSymbolData.value.currency_id}_${currentSymbolData.value.legal_id || 1}`
	return marketStore.state.marketDataMap.get(key) || null
})

// 涨跌幅（从全局市场数据获取）
const symbolChangeClass = computed(() => {
	const data = currentMarketData.value
	if (!data) return 'down'
	return data.change >= 0 ? 'up' : 'down'
})

const symbolChange = computed(() => {
	const data = currentMarketData.value
	if (!data) return '---'
	const change = data.change || 0
	return change >= 0 ? `+${change.toFixed(2)}%` : `${change.toFixed(2)}%`
})

// 当前币价（从市场数据获取）
const coinPrice = computed(() => {
	const data = currentMarketData.value
	if (!data) return 0
	return parseFloat(data.price || data.now_price) || 0
})

// 当前币种名称（从 symbol 中提取）
const currentCoinName = computed(() => {
	if (currentSymbol.value) {
		return currentSymbol.value.split('/')[0]
	}
	return 'ETH'
})

// 可用余额（根据标签页显示对应账户余额）
const availableBalance = computed(() => {
	if (mainTab.value === 'seconds') {
		// 交割合约显示交割合约账户余额
		return deliveryBalance.value.toFixed(4)
	} else {
		// 永续合约显示永续合约账户余额
		return contractBalance.value.toFixed(4)
	}
})

// 价格变化追踪
const previousPrice = ref(0) // 上一次的价格
const priceChangeClass = ref('') // 价格变化颜色类：'price-up' 或 'price-down'
const priceChangeTimer = ref(null) // 颜色变化定时器

// 当前价格（从全局市场数据获取）
const currentPrice = computed(() => {
	const data = currentMarketData.value
	if (data && data.price > 0) {
		// 使用动态精度（从WebSocket推送的decimal_scale）
		const precision = data.decimal_scale || 4
		// 使用parseFloat去除尾随零（如 122.95590000 → 122.9559）
		return parseFloat(data.price.toFixed(precision)).toString()
	}
	// 如果没有全局数据，使用从币种选择传递的价格
	if (currentSymbolData.value && currentSymbolData.value.price) {
		return parseFloat(parseFloat(currentSymbolData.value.price).toFixed(4)).toString()
	}
	// 没有有效数据时显示 "---"
	return '---'
})

// 监听价格变化，触发颜色变化
watch(currentPrice, (newPrice, oldPrice) => {
	if (newPrice === '---' || oldPrice === '---') return
	
	const newVal = parseFloat(newPrice)
	const oldVal = parseFloat(oldPrice)
	
	// 跳过初始化时的变化
	if (previousPrice.value === 0) {
		previousPrice.value = newVal
		return
	}
	
	// 比较价格变化
	if (newVal > previousPrice.value) {
		// 价格上涨 - 绿色
		priceChangeClass.value = 'price-up'
	} else if (newVal < previousPrice.value) {
		// 价格下跌 - 红色
		priceChangeClass.value = 'price-down'
	} else {
		// 价格不变
		return
	}
	
	// 更新上一次价格
	previousPrice.value = newVal
	
	// 清除之前的定时器
	if (priceChangeTimer.value) {
		clearTimeout(priceChangeTimer.value)
	}
	
	// 1秒后恢复到默认颜色（基于涨跌幅）
	priceChangeTimer.value = setTimeout(() => {
		priceChangeClass.value = ''
	}, 1000)
})

// 当前价格格式化显示（带千分位，动态精度）
const currentPriceFormatted = computed(() => {
	if (currentPrice.value === '---') return '---'
	const price = parseFloat(currentPrice.value)
	// 获取动态精度
	const data = currentMarketData.value
	const precision = data?.decimal_scale || 4
	// 千分位格式化，使用动态精度
	return price.toLocaleString('en-US', {
		minimumFractionDigits: 0,
		maximumFractionDigits: precision
	})
})

// USD等价价格（动态精度显示）
const currentPriceUsd = computed(() => {
	if (currentPrice.value === '---') return '---'
	const price = parseFloat(currentPrice.value)
	// 获取动态精度
	const data = currentMarketData.value
	const precision = data?.decimal_scale || 4
	// 去除尾随零
	return parseFloat(price.toFixed(precision)).toString()
})

// 订单类型选项
const orderTypeOptions = ['市价单', '限价单']
const orderTypeIndex = computed(() => orderType.value === 'market' ? 0 : 1)

// 订单类型切换处理
const onOrderTypeChange = (e) => {
	const newType = e.detail.value === 0 ? 'market' : 'limit'
	orderType.value = newType
	
	// 切换到限价单时，自动填充当前市场价格
	if (newType === 'limit') {
		const price = coinPrice.value || 0
		if (price > 0) {
			limitPrice.value = price.toString()
		}
	} else {
		// 切换到市价单时，清空限价
		limitPrice.value = ''
	}
	
	// 清空数量
	amount.value = ''
	
	console.log('[Contract] 订单类型切换:', newType)
}

// 点击盘口价格填充（仅限价单有效）
const selectPrice = (price) => {
	if (orderType.value === 'limit') {
		// 只有限价单时才能点击盘口价格
		const priceValue = parseFloat(price)
		if (priceValue > 0) {
			limitPrice.value = priceValue.toString()
			console.log('[Contract] 选择价格:', priceValue)
		}
	}
}

// 杠杆倍数配置（从后端动态加载）
const leverageOptions = ref(['10x', '20x', '30x', '40x', '50x', '60x', '70x', '80x', '90x', '100x', '120x', '150x', '180x'])
const leverageIndex = ref(4) // 默认 50x

// 从后端获取杠杆倍数配置
const fetchLeverageMultiples = async () => {
	try {
		const result = await getLeverageMultiples()
		console.log('[Contract] 杠杆倍数配置:', result)
		
		if (result && result.data && result.data.length > 0) {
			// 过滤type=1（倍数）的配置，并去重
			const multiplesSet = new Set(
				result.data
					.filter(item => item.type === 1)
					.map(item => parseInt(item.value)) // 转换为数字进行去重
			)
			
			// 转换为数组并排序，然后格式化为字符串
			const multiples = Array.from(multiplesSet)
				.sort((a, b) => a - b) // 按数值排序
				.map(value => `${value}x`) // 格式化为带x的字符串
			
			if (multiples.length > 0) {
				leverageOptions.value = multiples
				// 调整默认选中项，确保索引不越界
				if (leverageIndex.value >= multiples.length) {
					leverageIndex.value = Math.floor(multiples.length / 2) // 选择中间值
				}
				console.log('[Contract] 杠杆倍数加载成功（已去重）:', multiples)
			}
		}
	} catch (err) {
		console.error('[Contract] 获取杠杆倍数失败:', err)
		// 使用默认配置
	}
}

const activeBottomTab = ref('position') // position, pending, history

// 永续合约仓位和订单数据
const positions = ref([]) // 当前持仓列表
const pendingOrders = ref([]) // 待成交挂单列表
const historyOrders = ref([]) // 历史委托列表
const isLoading = ref(false) // 加载状态
const isSubmitting = ref(false) // 提交状态

// 当前选中的仓位（用于止盈止损弹窗）
const selectedPosition = ref(null)
const showTPSLModal = ref(false)
const tpslForm = ref({
	takeProfitAmount: '',
	stopLossAmount: ''
})

// 滑动组件相关（永续合约）
const marginAmount = ref('') // 初始保证金数量
const sliderValue = ref(0) // 滑块值（0-100）
const sliderPercents = [0, 25, 50, 75, 100] // 滑块刻度

// 秒合约数据
const periodOptions = ref([
	{ time: '30s', rate: '40', seconds: 30 },
	{ time: '60s', rate: '50', seconds: 60 },
	{ time: '120s', rate: '60', seconds: 120 },
	{ time: '180s', rate: '80', seconds: 180 },
	{ time: '300s', rate: '100', seconds: 300 }
])
const periodIndex = ref(0)
const secondsAmount = ref('')
const quickAmountOptions = [10, 50, 100, 500, 1000]
const minAmount = 10 // 最低投入金额
const secondsSubmitting = ref(false) // 秒合约下单提交状态

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

// 控盘阈值（最后3秒进入控盘阶段，与后端保持一致）
const CONTROL_THRESHOLD_SECONDS = 3

// 检测当前价格趋势是否与预设结果一致
const isPriceTrendMatchingPreset = (order, currentPrice) => {
	const preProfitResult = order.preProfitResult
	if (preProfitResult === 0) {
		// 无预设，不控盘
		return true
	}
	
	const entryPrice = order.entryPrice
	const priceDiff = currentPrice - entryPrice
	
	// 买涨(buy): 当前价 > 开仓价 = 盈利
	// 买跌(sell): 当前价 < 开仓价 = 盈利
	const isCurrentlyProfiting = order.side === 'buy' ? priceDiff > 0 : priceDiff < 0
	const shouldProfit = preProfitResult === 1
	
	return isCurrentlyProfiting === shouldProfit
}

// 计算控盘价格（让显示价格符合预设结果）
const calculateControlledPrice = (order, realPrice, remainingSeconds) => {
	const preProfitResult = order.preProfitResult
	if (preProfitResult === 0) {
		return realPrice // 无预设，不控盘
	}
	
	const entryPrice = order.entryPrice
	const shouldProfit = preProfitResult === 1
	
	// 计算目标价格（让结果符合预设）
	// 价格偏移基于入场价的千分之几
	const minOffset = entryPrice * 0.001 // 最小偏移0.1%
	const maxOffset = entryPrice * 0.003 // 最大偏移0.3%
	
	// 根据剩余时间计算偏移量（最后越接近结算，偏移越明显）
	const timeProgress = 1 - (remainingSeconds / CONTROL_THRESHOLD_SECONDS)
	const offset = minOffset + (maxOffset - minOffset) * timeProgress
	
	// 添加随机微小波动，让价格看起来自然
	const randomFactor = 0.8 + Math.random() * 0.4 // 0.8~1.2
	const adjustedOffset = offset * randomFactor
	
	if (order.side === 'buy') {
		// 买涨
		if (shouldProfit) {
			// 预设盈利，价格必须高于开仓价
			return Math.max(entryPrice + adjustedOffset, entryPrice + minOffset)
		} else {
			// 预设亏损，价格必须低于开仓价
			return Math.min(entryPrice - adjustedOffset, entryPrice - minOffset)
		}
	} else {
		// 买跌
		if (shouldProfit) {
			// 预设盈利，价格必须低于开仓价
			return Math.min(entryPrice - adjustedOffset, entryPrice - minOffset)
		} else {
			// 预设亏损，价格必须高于开仓价
			return Math.max(entryPrice + adjustedOffset, entryPrice + minOffset)
		}
	}
}

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
	
	// 价格更新定时器（使用真实WebSocket币价 + 最后2秒控盘）
	const priceTimer = setInterval(() => {
		const order = secondsActiveOrders.value.find(o => o.id === orderId)
		if (!order) {
			clearInterval(priceTimer)
			delete priceTrends[orderId]
			return
		}
		
		const trend = priceTrends[orderId]
		const basePrice = order.entryPrice
		const remainingSeconds = order.countdown
		
		// 获取真实WebSocket币价
		let realPrice = basePrice
		const currencyId = order.currencyId || currentSymbolData.value?.currency_id
		const legalId = currentSymbolData.value?.legal_id || 3 // 默认USDT
		
		if (currencyId) {
			const marketData = marketStore.getMarketData(currencyId, legalId)
			if (marketData && marketData.price > 0) {
				realPrice = marketData.price
			}
		}
		
		let displayPrice = realPrice
		
		// 控盘逻辑：最后2秒进入控盘阶段
		if (remainingSeconds <= CONTROL_THRESHOLD_SECONDS && remainingSeconds > 0) {
			// 检测当前价格趋势是否与预设结果一致
			if (!isPriceTrendMatchingPreset(order, realPrice)) {
				// 不一致，进行控盘，计算符合预设的价格
				displayPrice = calculateControlledPrice(order, realPrice, remainingSeconds)
				console.log(`[控盘] 订单${orderId} 剩余${remainingSeconds}秒, 真实价=${realPrice.toFixed(2)}, 控盘价=${displayPrice.toFixed(2)}, 预设=${order.preProfitResult}`)
			}
		}
		
		// 如果没有获取到真实价格，使用趋势模拟价格
		if (realPrice === basePrice && trend) {
			// 趋势动量衰减，偶尔产生新趋势
			if (trend.momentum <= 0) {
				if (Math.random() < 0.3) {
					trend.direction = Math.random() > 0.5 ? 1 : -1
					trend.momentum = Math.floor(Math.random() * 4) + 2
				} else {
					trend.direction = 0
				}
			} else {
				trend.momentum--
			}
			
			const volatilityPercent = (Math.random() * 1.0 + 0.5) / 100
			let priceChange = basePrice * volatilityPercent
			
			if (trend.direction !== 0) {
				if (Math.random() < 0.7) {
					priceChange *= trend.direction
				} else {
					priceChange *= (Math.random() > 0.5 ? 1 : -1)
				}
			} else {
				priceChange *= (Math.random() > 0.5 ? 1 : -1)
			}
			
			const smoothFactor = 0.6
			const targetPrice = trend.lastPrice + priceChange
			const newPrice = trend.lastPrice + (targetPrice - trend.lastPrice) * smoothFactor
			
			const maxDeviation = basePrice * 0.03
			displayPrice = Math.max(basePrice - maxDeviation, Math.min(basePrice + maxDeviation, newPrice))
			
			trend.lastPrice = displayPrice
			
			// 控盘阶段复用模拟价格时也要考虑控盘
			if (remainingSeconds <= CONTROL_THRESHOLD_SECONDS && remainingSeconds > 0) {
				if (!isPriceTrendMatchingPreset(order, displayPrice)) {
					displayPrice = calculateControlledPrice(order, displayPrice, remainingSeconds)
				}
			}
		}
		
		// 最后3秒：直接显示最终结算结果（与后端风控结果保持一致）
		let finalDisplayPrice = displayPrice
		let finalExpectedProfit = order.expectedProfit
		
		if (remainingSeconds <= CONTROL_THRESHOLD_SECONDS && remainingSeconds > 0 && order.preProfitResult !== 0) {
			// 有风控预设，直接计算并显示最终结算结果
			const settlement = calculateFinalSettlement(order)
			finalDisplayPrice = settlement.settlePrice
			finalExpectedProfit = settlement.finalProfit
			console.log(`[最后${remainingSeconds}秒] 订单${orderId} 显示最终结算结果: 价格=${finalDisplayPrice.toFixed(2)}, 盈亏=${finalExpectedProfit.toFixed(2)}`)
		} else {
			// 正常计算预期收益 - 盈亏1:1对称
			const priceDiff = displayPrice - basePrice
			const isProfit = order.side === 'buy' ? priceDiff > 0 : priceDiff < 0
			finalExpectedProfit = isProfit ? order.amount * (order.rate / 100) : -order.amount * (order.rate / 100)
		}
		
		// 更新买卖压力
		const currentPressure = order.buyPressure || 50
		const pressureChange = (Math.random() - 0.5) * 5
		const newPressure = Math.max(35, Math.min(65, currentPressure + pressureChange))
		
		updateOrderInList(orderId, {
			currentPrice: finalDisplayPrice,
			expectedProfit: finalExpectedProfit,
			buyPressure: parseFloat(newPressure.toFixed(2))
		})
	}, 1800) // 更新间隔1800ms
	
	orderTimers[orderId] = {
		countdown: countdownTimer,
		priceUpdate: priceTimer
	}
}

// 根据风控预设结果计算最终结算价格和盈亏
const calculateFinalSettlement = (order) => {
	const preProfitResult = order.preProfitResult
	const basePrice = order.entryPrice
	const amount = order.amount
	const rate = order.rate / 100
	
	// 如果没有风控预设，使用当前价格计算 - 盈亏1:1对称
	if (preProfitResult === 0) {
		const priceDiff = order.currentPrice - basePrice
		const isProfit = order.side === 'buy' ? priceDiff > 0 : priceDiff < 0
		const finalProfit = isProfit ? amount * rate : -amount * rate
		return {
			settlePrice: order.currentPrice,
			finalProfit: finalProfit
		}
	}
	
	// 有风控预设，根据预设结果计算
	const shouldProfit = preProfitResult === 1
	
	// 计算结算价格（确保结果符合风控预设）
	const minOffset = basePrice * 0.001 // 最小偏移0.1%
	const randomFactor = 0.9 + Math.random() * 0.2 // 0.9~1.1
	const offset = minOffset * randomFactor
	
	let settlePrice
	if (order.side === 'buy') {
		// 买涨
		if (shouldProfit) {
			// 盈利：结算价 > 开仓价
			settlePrice = basePrice + offset
		} else {
			// 亏损：结算价 < 开仓价
			settlePrice = basePrice - offset
		}
	} else {
		// 买跌
		if (shouldProfit) {
			// 盈利：结算价 < 开仓价
			settlePrice = basePrice - offset
		} else {
			// 亏损：结算价 > 开仓价
			settlePrice = basePrice + offset
		}
	}
	
	// 计算最终盈亏 - 盈亏1:1对称
	const finalProfit = shouldProfit ? amount * rate : -amount * rate
	
	return {
		settlePrice: settlePrice,
		finalProfit: finalProfit
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
	
	// 根据风控预设计算最终结算结果
	const settlement = calculateFinalSettlement(order)
	
	// 设置结算数据
	const settledOrder = {
		...order,
		settlePrice: settlement.settlePrice,
		finalProfit: settlement.finalProfit,
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
const handleSecondsTrade = async (side) => {
	if (!secondsAmount.value) {
		uni.showToast({ title: '请输入投入金额', icon: 'none' })
		return
	}
	const amount = parseFloat(secondsAmount.value)
	if (amount < minAmount) {
		uni.showToast({ title: `最低投入 ${minAmount} USDT`, icon: 'none' })
		return
	}
	// 校验交割合约账户余额
	if (amount > deliveryBalance.value) {
		uni.showToast({ title: '交割合约账户余额不足', icon: 'none' })
		return
	}
	
	// 检查是否已有进行中的订单
	if (secondsActiveOrders.value.length > 0) {
		uni.showToast({ title: '已有进行中的订单', icon: 'none' })
		return
	}
	
	if (secondsSubmitting.value) return
	secondsSubmitting.value = true
	
	uni.showLoading({ title: '正在下单...' })
	try {
		const period = periodOptions.value[periodIndex.value]
		const result = await submitMicroOrder({
			currency_id: currentSymbolData.value?.currency_id || 1,
			direction: side === 'buy' ? 'rise' : 'fall',
			seconds: period.seconds,
			amount: amount
		})
		
		uni.hideLoading()
		uni.showToast({ title: '下单成功', icon: 'success' })
		console.log('[Contract] 下单成功:', result)
		
		// 清空输入
		secondsAmount.value = ''
		
		// walletStore 会通过 WebSocket 自动更新余额
		
		// 刷新进行中订单列表
		await fetchMicroActiveOrders()
		
		// 如果有进行中订单，打开进度弹窗
		if (secondsActiveOrders.value.length > 0) {
			const newOrder = secondsActiveOrders.value[0]
			openOrderProgress(newOrder)
		}
	} catch (err) {
		uni.hideLoading()
		console.error('[Contract] 秒合约下单失败:', err)
		uni.showToast({ title: err.message || '下单失败', icon: 'none' })
	} finally {
		secondsSubmitting.value = false
	}
}

// 滑点容差说明和状态
const showSlippageModal = ref(false)
const slippageEnabled = ref(false) // 滑点容差是否启用

// 切换滑点容差状态
const toggleSlippage = () => {
	slippageEnabled.value = !slippageEnabled.value
	console.log('[Contract] 滑点容差状态:', slippageEnabled.value)
	
	// 保存到localStorage
	uni.setStorageSync('slippageEnabled', slippageEnabled.value)
	
	uni.showToast({
		title: slippageEnabled.value ? '已启用滑点容差' : '已关闭滑点容差',
		icon: 'none',
		duration: 1500
	})
}

// 币种选择回调
const onSymbolSelect = (item) => {
	currentSymbol.value = item.symbol
	
	// 保存当前选择的币种信息到localStorage，用于页面刷新后恢复
	const symbolData = {
		symbol: item.symbol,
		currency_id: item.currency_id,
		legal_id: item.legal_id
	}
	uni.setStorageSync('lastContractSymbol', JSON.stringify(symbolData))
	currentSymbolData.value = symbolData
	
	// 基于币种价格更新盘口数据（从全局状态获取价格）
	const marketData = marketStore.getMarketData(item.currency_id, item.legal_id)
	const price = marketData ? marketData.price : 2956.86
	generateOrderBookWithPrice(price)
	
	console.log('[Contract] 币种选择:', item.symbol, '完整数据:', item)
	console.log('[Contract] 保存到storage:', symbolData)
	console.log('[Contract] currentSymbolData.value已更新:', currentSymbolData.value)
}

// ==================== 永续合约交易核心功能 ====================

// 获取当前杠杆倍数（数字）
const currentLeverage = computed(() => {
	const leverageStr = leverageOptions.value[leverageIndex.value] || '10x'
	return parseInt(leverageStr.replace('x', ''))
})

// 开仓操作（做多/做空）
const handleOpenPosition = async (side) => {
	// side: 'long' = 做多, 'short' = 做空
	const type = side === 'long' ? 1 : 2
	
	// 验证参数
	if (!marginAmount.value || parseFloat(marginAmount.value) <= 0) {
		uni.showToast({ title: '请输入保证金金额', icon: 'none' })
		return
	}
	
	const margin = parseFloat(marginAmount.value)
	if (margin < 10) {
		uni.showToast({ title: '最小保证金为10 USDT', icon: 'none' })
		return
	}
	
	// 校验永续合约账户余额
	if (margin > contractBalance.value) {
		uni.showToast({ title: '永续合约账户余额不足', icon: 'none' })
		return
	}
	
	// 限价单需要验证价格
	if (orderType.value === 'limit') {
		if (!limitPrice.value || parseFloat(limitPrice.value) <= 0) {
			uni.showToast({ title: '请输入限价价格', icon: 'none' })
			return
		}
	}
	
	// 确保已选择币种
	if (!currentSymbolData.value || !currentSymbolData.value.currency_id) {
		uni.showToast({ title: '请先选择交易币种', icon: 'none' })
		return
	}
	
	// ========== 调试日志：开仓前检查 ==========
	console.log('[Contract] ========== 开仓前检查 ==========')
	console.log('[Contract] currentSymbol.value:', currentSymbol.value)
	console.log('[Contract] currentSymbolData.value:', JSON.stringify(currentSymbolData.value))
	const storageCheck = uni.getStorageSync('lastContractSymbol')
	console.log('[Contract] Storage中的币种:', storageCheck)
	console.log('[Contract] ======================================')
	
	isSubmitting.value = true
	uni.showLoading({ title: '正在开仓...' })
	
	try {
		const orderData = {
			currency_id: currentSymbolData.value.currency_id,
			legal_id: currentSymbolData.value.legal_id || 1,
			type: type,
			order_type: orderType.value === 'market' ? 1 : 2,
			margin: margin,
			leverage: currentLeverage.value
		}
		
		// 限价单添加价格
		if (orderType.value === 'limit') {
			orderData.limit_price = parseFloat(limitPrice.value)
		}
		
		console.log('[Contract] 开仓参数:', orderData)
		const res = await contractOpenPosition(orderData)
		console.log('[Contract] 开仓结果:', res)
		
		uni.hideLoading()
		
		// 显示成功提示
		const data = res.data
		const successMsg = orderType.value === 'market' 
			? `开仓成功\n开仓价: ${data.entry_price}\n手续费: ${data.fee} USDT`
			: `挂单成功\n限价: ${data.limit_price}`
		
		uni.showToast({
			title: orderType.value === 'market' ? '开仓成功' : '挂单成功',
			icon: 'success'
		})
		
		// 清空输入
		marginAmount.value = ''
		sliderValue.value = 0
		if (orderType.value === 'limit') {
			limitPrice.value = ''
		}
		
		// 刷新数据
		await fetchContractData()
		// walletStore 会通过 WebSocket 自动更新余额
		
	} catch (err) {
		uni.hideLoading()
		console.error('[Contract] 开仓失败:', err)
		uni.showToast({
			title: err.message || '开仓失败',
			icon: 'none',
			duration: 3000
		})
	} finally {
		isSubmitting.value = false
	}
}

// 获取合约数据（仓位和挂单）
const fetchContractData = async () => {
	isLoading.value = true
	try {
		// 并行获取仓位和挂单
		const [posRes, pendingRes] = await Promise.all([
			getContractPositions(),
			getContractPendingOrders()
		])
		
		console.log('[Contract] 仓位接口返回:', posRes)
		console.log('[Contract] 挂单接口返回:', pendingRes)
		
		// 更新仓位列表（处理不同的返回格式）
		let allPositions = []
		if (posRes.data) {
			// 如果是数组直接使用，否则尝试获取list属性或包装为数组
			if (Array.isArray(posRes.data)) {
				allPositions = posRes.data
			} else if (Array.isArray(posRes.data.list)) {
				allPositions = posRes.data.list
			} else if (posRes.data.positions && Array.isArray(posRes.data.positions)) {
				allPositions = posRes.data.positions
			} else if (typeof posRes.data === 'object' && posRes.data.id) {
				// 单个仓位对象
				allPositions = [posRes.data]
			}
		}
		// 过滭status=1的持仓中仓位
		positions.value = allPositions.filter(p => p && p.status === 1)
		
		// 输出详细的仓位数据供调试
		if (positions.value.length > 0) {
			console.log('[Contract] 仓位详细数据:', JSON.stringify(positions.value[0], null, 2))
			console.log('[Contract] 仓位字段:', Object.keys(positions.value[0]))
		}
		
		// 更新挂单列表
		let allPending = []
		if (pendingRes.data) {
			if (Array.isArray(pendingRes.data)) {
				allPending = pendingRes.data
			} else if (Array.isArray(pendingRes.data.list)) {
				allPending = pendingRes.data.list
			}
		}
		pendingOrders.value = allPending
		
		console.log('[Contract] 当前持仓:', positions.value.length)
		console.log('[Contract] 当前挂单:', pendingOrders.value.length)
	} catch (err) {
		console.error('[Contract] 获取合约数据失败:', err)
	} finally {
		isLoading.value = false
	}
}

// 获取历史委托
const fetchHistoryOrders = async () => {
	try {
		const res = await getContractHistoryPositions({ page: 1, page_size: 50 })
		console.log('[Contract] 历史委托:', res)
		historyOrders.value = res.data?.list || res.data || []
	} catch (err) {
		console.error('[Contract] 获取历史委托失败:', err)
	}
}

// ========== 秒合约数据获取函数 ==========

// 获取秒合约周期配置
const fetchMicroPeriods = async () => {
	try {
		const res = await getMicroPeriods()
		console.log('[Contract] 秒合约周期配置:', res)
		if (res.data && Array.isArray(res.data)) {
			// 转换后端数据格式为前端格式
			periodOptions.value = res.data.map(item => ({
				time: `${item.seconds}s`,
				rate: (item.profit_ratio * 100).toFixed(0),
				seconds: item.seconds
			}))
			console.log('[Contract] 周期配置已加载:', periodOptions.value)
		}
	} catch (err) {
		console.error('[Contract] 获取秒合约周期配置失败:', err)
		// 保持默认配置
	}
}

// 获取秒合约进行中订单
const fetchMicroActiveOrders = async () => {
	try {
		const res = await getMicroActiveOrders()
		console.log('[Contract] 秒合约进行中订单响应:', res)
		
		// 后端直接返回订单对象（不是包裹在order字段中）
		// 有订单时: { type: "success", data: {id, user_id, ...} }
		// 无订单时: { type: "success", data: null, message: "暂无进行中的订单" }
		if (res.data && res.data.id) {
			// 后端直接返回订单对象
			const order = res.data
			const mappedOrder = mapMicroOrderToFrontend(order)
			secondsActiveOrders.value = [mappedOrder]
			console.log('[Contract] 进行中订单映射后:', mappedOrder)
			
			// 启动倒计时定时器
			startOrderTimers(mappedOrder.id, mappedOrder)
		} else {
			secondsActiveOrders.value = []
			console.log('[Contract] 暂无进行中订单')
		}
	} catch (err) {
		console.error('[Contract] 获取秒合约进行中订单失败:', err)
		secondsActiveOrders.value = []
	}
}

// 获取秒合约历史结算订单
const fetchMicroHistoryOrders = async () => {
	try {
		const res = await getMicroOrderList({ page: 1, page_size: 50 })
		console.log('[Contract] 秒合约历史订单:', res)
		if (res.data && Array.isArray(res.data.list)) {
			// 过滤已结算的订单 (status=1)
			secondsHistoryOrders.value = res.data.list
				.filter(order => order.status === 1)
				.map(order => mapMicroOrderToFrontend(order))
		} else if (res.data && Array.isArray(res.data)) {
			secondsHistoryOrders.value = res.data
				.filter(order => order.status === 1)
				.map(order => mapMicroOrderToFrontend(order))
		} else {
			secondsHistoryOrders.value = []
		}
	} catch (err) {
		console.error('[Contract] 获取秒合约历史订单失败:', err)
		secondsHistoryOrders.value = []
	}
}

// 映射后端秒合约订单到前端格式
// 后端返回字段: id, user_id, currency_id, currency_name, direction("rise"/"fall"), 
//                amount, seconds, profit_ratio, open_price, end_price, fact_profits, 
//                profit_result(1/-1), pre_profit_result(0/1/-1), status(0/1), expected_end_time, create_time, settle_time
const mapMicroOrderToFrontend = (order) => {
	const now = Date.now()
	// 使用 expected_end_time 计算剩余时间
	const endTime = order.expected_end_time ? new Date(order.expected_end_time).getTime() : (now + order.seconds * 1000)
	const remainingSeconds = Math.max(0, Math.ceil((endTime - now) / 1000))
	
	// 盈利率从后端返回的 profit_ratio (小数形式，如0.4表示40%)
	const rate = order.profit_ratio ? (order.profit_ratio * 100) : 50
	
	// 计算预期收益
	const amount = order.amount || order.number || 0
	const currentPrice = order.end_price || order.open_price || 0
	const entryPrice = order.open_price || 0
	
	// direction 字段: "rise"=买涨, "fall"=买跌，或 type: 1=买涨, 2=买跌
	let side = 'buy'
	if (order.direction === 'fall' || order.type === 2) {
		side = 'sell'
	}
	
	// 计算预期收益（进行中订单）- 盈亏1:1对称
	let expectedProfit = 0
	if (order.status === 0 && entryPrice > 0 && currentPrice > 0) {
		const priceDiff = currentPrice - entryPrice
		const isProfit = (side === 'buy' && priceDiff > 0) || (side === 'sell' && priceDiff < 0)
		expectedProfit = isProfit ? amount * (rate / 100) : -amount * (rate / 100)
	}
	
	// 已结算订单使用 fact_profits
	const finalProfit = order.fact_profits || 0
	
	// 预设结果 (0=无预设, 1=盈利, -1=亏损)
	// 注意: 后台管理使用0/1/2(2表示输)，数据库存傂为0/1/-1
	const preProfitResult = order.pre_profit_result || 0
	
	return {
		id: order.id,
		symbol: order.currency_name ? `${order.currency_name}/USDT` : `${getCurrencyNameById(order.currency_id)}/USDT`,
		side: side,
		amount: amount,
		period: order.seconds || 30,
		rate: rate,
		entryPrice: entryPrice,
		currentPrice: currentPrice,
		countdown: remainingSeconds,
		expectedProfit: expectedProfit,
		buyPressure: 50,
		finalProfit: finalProfit,
		settlePrice: order.end_price || 0,
		status: order.status === 0 ? 'active' : 'settled',
		settleTime: order.settle_time || '',
		createTime: order.create_time || '',
		// 风控相关
		currencyId: order.currency_id,
		preProfitResult: preProfitResult // 预设结果（用于前端控盘）
	}
}

// 手动平仓
const handleClosePosition = async (position) => {
	uni.showModal({
		title: '确认平仓',
		content: `确定要平掉这个${position.type === 1 ? '多' : '空'}仓吗？`,
		success: async (res) => {
			if (res.confirm) {
				uni.showLoading({ title: '正在平仓...' })
				try {
					const result = await contractClosePosition({ position_id: position.id })
					console.log('[Contract] 平仓结果:', result)
					
					uni.hideLoading()
					
					const data = result.data
					const pnlText = data.pnl >= 0 ? `+${data.pnl.toFixed(2)}` : data.pnl.toFixed(2)
					
					uni.showToast({
						title: `平仓成功\n盈亏: ${pnlText} USDT`,
						icon: 'none',
						duration: 3000
					})
					
					// 刷新数据
					await fetchContractData()
					// walletStore 会通过 WebSocket 自动更新余额
				} catch (err) {
					uni.hideLoading()
					console.error('[Contract] 平仓失败:', err)
					uni.showToast({ title: err.message || '平仓失败', icon: 'none' })
				}
			}
		}
	})
}

// 取消限价单
const handleCancelOrder = async (order) => {
	uni.showModal({
		title: '确认取消',
		content: '确定要取消这个限价单吗？',
		success: async (res) => {
			if (res.confirm) {
				uni.showLoading({ title: '正在取消...' })
				try {
					const result = await contractCancelOrder(order.id)
					console.log('[Contract] 取消结果:', result)
					
					uni.hideLoading()
					uni.showToast({ title: '取消成功', icon: 'success' })
					
					// 刷新数据
					await fetchContractData()
					// walletStore 会通过 WebSocket 自动更新余额
				} catch (err) {
					uni.hideLoading()
					console.error('[Contract] 取消失败:', err)
					uni.showToast({ title: err.message || '取消失败', icon: 'none' })
				}
			}
		}
	})
}

// 追单（限价单转市价）
const handleChaseOrder = async (order) => {
	uni.showModal({
		title: '确认追单',
		content: '确定要将这个限价单转为市价立即成交吗？',
		success: async (res) => {
			if (res.confirm) {
				uni.showLoading({ title: '正在追单...' })
				try {
					const result = await contractChaseOrder({ order_id: order.id })
					console.log('[Contract] 追单结果:', result)
					
					uni.hideLoading()
					uni.showToast({ title: '追单成功', icon: 'success' })
					
					// 刷新数据
					await fetchContractData()
				} catch (err) {
					uni.hideLoading()
					console.error('[Contract] 追单失败:', err)
					uni.showToast({ title: err.message || '追单失败', icon: 'none' })
				}
			}
		}
	})
}

// 打开止盈止损设置弹窗
const openTPSLModal = (position) => {
	selectedPosition.value = position
	tpslForm.value = {
		takeProfitAmount: position.take_profit_amount ? position.take_profit_amount.toString() : '',
		stopLossAmount: position.stop_loss_amount ? position.stop_loss_amount.toString() : ''
	}
	showTPSLModal.value = true
}

// 关闭止盈止损弹窗
const closeTPSLModal = () => {
	showTPSLModal.value = false
	selectedPosition.value = null
}

// 提交止盈止损设置
const submitTPSL = async () => {
	if (!selectedPosition.value) return
	
	const tp = parseFloat(tpslForm.value.takeProfitAmount) || 0
	const sl = parseFloat(tpslForm.value.stopLossAmount) || 0
	
	if (tp <= 0 && sl <= 0) {
		uni.showToast({ title: '请设置止盈或止损金额', icon: 'none' })
		return
	}
	
	uni.showLoading({ title: '正在设置...' })
	try {
		const data = {
			position_id: selectedPosition.value.id
		}
		if (tp > 0) data.take_profit_amount = tp
		if (sl > 0) data.stop_loss_amount = sl
		
		const result = await contractSetTPSL(data)
		console.log('[Contract] 止盈止损设置结果:', result)
		
		uni.hideLoading()
		uni.showToast({ title: '设置成功', icon: 'success' })
		
		closeTPSLModal()
		await fetchContractData()
	} catch (err) {
		uni.hideLoading()
		console.error('[Contract] 止盈止损设置失败:', err)
		uni.showToast({ title: err.message || '设置失败', icon: 'none' })
	}
}

// 格式化价格显示（自动去除尾随零）
const formatPrice = (value, precision = null) => {
	if (value === null || value === undefined || isNaN(value)) return '--'
	const num = parseFloat(value)
	if (isNaN(num)) return '--'
	
	let decimals
	if (precision !== null) {
		// 使用指定精度
		decimals = precision
	} else {
		// 根据数值大小决定小数位数
		if (num >= 1000) decimals = 2
		else if (num >= 1) decimals = 4
		else decimals = 6
	}
	
	// 使用parseFloat自动去除尾随零（如 122.95590000 → 122.9559）
	return parseFloat(num.toFixed(decimals)).toString()
}

// 获取仓位盈亏 - 总是使用前端实时计算（不再依赖后端静态数据）
const getPositionPnL = (position) => {
	return getRealTimePnL(position)
}

// 获取仓位盈亏率 - 总是使用前端实时计算
const getPositionPnLRate = (position) => {
	return getRealTimePnLRate(position)
}

// 根据currency_id获取币种名称
const getCurrencyNameById = (currencyId) => {
	if (!currencyId) return 'BTC'
	
	// 从 marketStore 中查找
	const allData = marketStore.getAllMarketData()
	const currency = allData.find(item => item.currency_id === currencyId)
	
	if (currency && currency.name) {
		return currency.name
	}
	
	// 如果找不到，返回默认值
	console.warn('[Contract] 找不到currency_id对应的币种名称:', currencyId)
	return 'UNKNOWN'
}

// 前端计算仓位未实现盈亏（实时计算，基于当前价格）
const calcUnrealizedPnL = (position) => {
	// 获取开仓价格：优先 entry_price，否则 price
	const entryPrice = position.entry_price || position.price
	if (!entryPrice || entryPrice <= 0) return 0
	
	// 获取当前价格：总是从 marketStore 实时获取，确保最新价格
	const currentMarketData = marketStore.getMarketData(position.currency_id, position.legal_id || 1)
	const currentPrice = currentMarketData?.price || entryPrice
	
	// 获取保证金和杠杆
	const margin = position.origin_margin || position.origin_caution_money || position.margin || 0
	const leverage = position.leverage || position.multiple || 1
	
	// 计算仓位价值和盈亏
	const positionValue = margin * leverage
	const priceChangeRate = (currentPrice - entryPrice) / entryPrice
	
	if (position.type === 1) {
		// 做多：价格上涨盈利
		return positionValue * priceChangeRate
	} else {
		// 做空：价格下跌盈利
		return positionValue * (-priceChangeRate)
	}
}

// 前端计算仓位盈亏率（实时计算）
const calcPnLRate = (position) => {
	const pnl = calcUnrealizedPnL(position)
	const margin = position.margin || position.caution_money || position.origin_margin || position.origin_caution_money
	if (margin > 0) {
		return (pnl / margin * 100).toFixed(2)
	}
	return '0.00'
}

// 实时盈亏计算的 computed 属性（用于响应式更新）
const getRealTimePnL = (position) => {
	// 强制依赖 marketStore 的更新
	const _trigger = marketStore.state.lastUpdateTime
	return calcUnrealizedPnL(position)
}

const getRealTimePnLRate = (position) => {
	// 强制依赖 marketStore 的更新
	const _trigger = marketStore.state.lastUpdateTime
	return calcPnLRate(position)
}

// 点击历史委托tab
const handleHistoryTabClick = async () => {
	activeBottomTab.value = 'history'
	await fetchHistoryOrders()
}

// 获取平仓类型文本
const getCloseTypeText = (closeType) => {
	const types = {
		1: '手动平仓',
		2: '爆仓',
		3: '止盈',
		4: '止损'
	}
	return types[closeType] || '未知'
}

// 从存储中恢复币种选择
const restoreSymbolFromStorage = () => {
	const savedSymbol = uni.getStorageSync('lastContractSymbol')
	if (savedSymbol) {
		try {
			const data = typeof savedSymbol === 'string' ? JSON.parse(savedSymbol) : savedSymbol
			currentSymbol.value = data.symbol || 'ETH/USDT'
			currentSymbolData.value = data
			
			// 基于币种价格更新盘口数据
			const marketData = marketStore.getMarketData(data.currency_id, data.legal_id)
			const price = marketData ? marketData.price : 2956.86
			generateOrderBookWithPrice(price)
			
			console.log('[Contract] 恢复币种选择:', data.symbol)
		} catch (err) {
			console.error('[Contract] 解析币种数据失败:', err)
			initDefaultSymbol()
		}
	} else {
		initDefaultSymbol()
	}
}

// 初始化默认币种
const initDefaultSymbol = () => {
	// 尝试从 marketStore 获取 ETH 的数据
	const ethData = marketStore.getMarketDataByName('ETH')
	if (ethData && ethData.currency_id) {
		currentSymbolData.value = {
			symbol: ethData.symbol || 'ETH/USDT',
			currency_id: ethData.currency_id,
			legal_id: ethData.legal_id || 1
		}
		console.log('[Contract] 使用marketStore初始化默认币种:', currentSymbolData.value)
	} else {
		// marketStore 可能还未加载，使用默认 currency_id
		currentSymbolData.value = {
			symbol: 'ETH/USDT',
			currency_id: 2, // ETH 默认 ID
			legal_id: 1
		}
		console.log('[Contract] 使用默认配置初始化币种:', currentSymbolData.value)
	}
}

// WebSocket订阅管理
let positionChannel = null
let microChannel = null // 秒合约结算推送频道

const setupWebSocketSubscriptions = () => {
	// 获取当前用户ID
	const userInfo = uni.getStorageSync('userInfo')
	if (!userInfo || !userInfo.id) {
		console.log('[Contract] 用户未登录，跳过WebSocket订阅')
		return
	}
	
	const userId = userInfo.id
	// 不再订阅 wallet 频道，walletStore 已统一管理
	positionChannel = `position:${userId}`
	microChannel = `micro:${userId}` // 秒合约结算频道
	
	// 确保WebSocket已连接
	wsClient.connect()
	
	// 订阅仓位更新频道
	wsClient.subscribe(positionChannel, (data) => {
		console.log('[Contract] 收到仓位更新推送:', data)
		handlePositionUpdate(data)
	})
	
	// 订阅秒合约结算频道
	wsClient.subscribe(microChannel, (data) => {
		console.log('[Contract] 收到秒合约结算推送:', data)
		handleMicroOrderSettled(data)
	})
	
	console.log('[Contract] WebSocket订阅完成:', { positionChannel, microChannel })
	console.log('[Contract] 余额数据由 walletStore 统一管理')
}

const cleanupWebSocketSubscriptions = () => {
	// 不需要取消订阅 wallet 频道，walletStore 会统一管理
	if (positionChannel) {
		wsClient.unsubscribe(positionChannel)
		positionChannel = null
	}
	if (microChannel) {
		wsClient.unsubscribe(microChannel)
		microChannel = null
	}
	console.log('[Contract] WebSocket订阅已清理')
}

// 处理秒合约结算推送
const handleMicroOrderSettled = (data) => {
	if (!data || data.type !== 'micro_order_settled') return
	
	const order = data.order
	if (!order) return
	
	console.log('[Contract] 秒合约订单结算:', order)
	
	// 清除该订单的定时器
	if (orderTimers[order.id]) {
		clearInterval(orderTimers[order.id].countdown)
		clearInterval(orderTimers[order.id].priceUpdate)
		delete orderTimers[order.id]
	}
	delete priceTrends[order.id]
	
	// 映射后端订单数据
	const settledOrder = mapMicroOrderToFrontend(order)
	settledOrder.status = 'settled'
	
	// 从进行中列表移除
	const activeIndex = secondsActiveOrders.value.findIndex(o => o.id === order.id)
	if (activeIndex > -1) {
		secondsActiveOrders.value.splice(activeIndex, 1)
	}
	
	// 添加到历史结算列表
	secondsHistoryOrders.value.unshift(settledOrder)
	
	// 如果是当前弹窗显示的订单，更新弹窗状态
	if (currentOrderId.value === order.id && showSecondsProgress.value) {
		secondsOrderInfo.value = settledOrder
		isSettled.value = true
	}
	
	// 显示结算结果提示
	const profit = order.profit_result || 0
	const profitText = profit >= 0 ? `+${profit.toFixed(2)}` : profit.toFixed(2)
	uni.showToast({
		title: `秒合约结算: ${profitText} USDT`,
		icon: 'none',
		duration: 3000
	})
	
	// walletStore 会通过 WebSocket 自动更新余额
}

// 处理仓位更新推送
const handlePositionUpdate = (data) => {
	if (!data) return
	
	const { type, position } = data
	
	// 根据事件类型处理
	switch (type) {
		case 'opened':
			// 新开仓位
			uni.showToast({ title: '开仓成功', icon: 'success' })
			fetchContractData()
			// walletStore 会通过 WebSocket 自动更新余额
			break
		case 'closed':
			// 仓位已平仓
			const pnl = position?.pnl || 0
			const pnlText = pnl >= 0 ? `+${pnl.toFixed(2)}` : pnl.toFixed(2)
			uni.showToast({ title: `已平仓，盈亏: ${pnlText}`, icon: 'none', duration: 3000 })
			fetchContractData()
			// walletStore 会通过 WebSocket 自动更新余额
			break
		case 'liquidated':
			// 爆仓
			uni.showModal({
				title: '爆仓提醒',
				content: '您的仓位已被强平，请注意风险控制',
				showCancel: false
			})
			fetchContractData()
			// walletStore 会通过 WebSocket 自动更新余额
			break
		case 'take_profit':
			// 止盈触发
			uni.showToast({ title: '止盈触发，已自动平仓', icon: 'success', duration: 3000 })
			fetchContractData()
			// walletStore 会通过 WebSocket 自动更新余额
			break
		case 'stop_loss':
			// 止损触发
			uni.showToast({ title: '止损触发，已自动平仓', icon: 'none', duration: 3000 })
			fetchContractData()
			// walletStore 会通过 WebSocket 自动更新余额
			break
		case 'limit_filled':
			// 限价单成交
			uni.showToast({ title: '限价单已成交', icon: 'success' })
			fetchContractData()
			break
		default:
			// 其他更新，刷新数据
			fetchContractData()
	}
}

// 余额数据由 walletStore 统一管理，不再需要 handleBalanceUpdate 函数

// 格式化数字：直接截断，不四舍五入，确保不超出余额
const floorFixed = (num, decimals) => {
	if (isNaN(num) || num === null) return '0'
	const factor = Math.pow(10, decimals)
	const result = (Math.floor(num * factor + 0.0000000001) / factor).toFixed(decimals)
	return result
}

// 滑动组件处理函数（永续合约）
// 滑块改变事件
const onSliderChange = (e) => {
	const percent = e.detail.value
	sliderValue.value = percent
	// 直接计算并赋值，确保不超出余额（使用当前标签页的可用余额）
	const maxUsdt = parseFloat(availableBalance.value) || 0
	if (percent === 100) {
		marginAmount.value = floorFixed(maxUsdt, 2)
	} else {
		const usdtAmount = maxUsdt * (percent / 100)
		marginAmount.value = floorFixed(usdtAmount, 2)
	}
	
	// 如果是限价单且限价输入框为空，自动填充当前币价
	if (orderType.value === 'limit' && !limitPrice.value) {
		const price = coinPrice.value || 0
		if (price > 0) {
			limitPrice.value = price.toFixed(4)
		}
	}
}

// 滑动过程中实时计算
const onSliderChanging = (e) => {
	const percent = e.detail.value
	sliderValue.value = percent
	// 直接内联计算，确保不超出余额（使用当前标签页的可用余额）
	const maxUsdt = parseFloat(availableBalance.value) || 0
	if (percent === 100) {
		marginAmount.value = floorFixed(maxUsdt, 2)
	} else {
		const usdtAmount = maxUsdt * (percent / 100)
		marginAmount.value = floorFixed(usdtAmount, 2)
	}
	
	// 如果是限价单且限价输入框为空，自动填充当前币价
	if (orderType.value === 'limit' && !limitPrice.value) {
		const price = coinPrice.value || 0
		if (price > 0) {
			limitPrice.value = price.toFixed(4)
		}
	}
}

// 根据滑块百分比计算保证金（保留用于其他调用）
const calculateMarginBySlider = (percent) => {
	// 永续合约使用当前标签页的可用余额的百分比，确保不超出余额
	const maxUsdt = parseFloat(availableBalance.value) || 0
	if (percent === 100) {
		marginAmount.value = floorFixed(maxUsdt, 2)
	} else {
		const usdtAmount = maxUsdt * (percent / 100)
		marginAmount.value = floorFixed(usdtAmount, 2)
	}
	
	// 如果是限价单且限价输入框为空，自动填充当前币价
	if (orderType.value === 'limit' && !limitPrice.value) {
		const price = coinPrice.value || 0
		if (price > 0) {
			limitPrice.value = price.toFixed(4)
		}
	}
}

// 保证金输入框变化时的处理
const onMarginInput = (e) => {
	const inputVal = e.detail.value
	
	// 立即更新marginAmount，确保响应式同步
	marginAmount.value = inputVal
	
	// 解析数值用于计算
	const numVal = parseFloat(inputVal) || 0
	
	// 限制最大值为USDT余额
	const maxUsdt = usdtBalance.value
	if (numVal > maxUsdt) {
		marginAmount.value = floorFixed(maxUsdt, 2)
	}
	
	// 同步更新滑块百分比
	if (maxUsdt > 0) {
		const percent = Math.min((numVal / maxUsdt) * 100, 100)
		sliderValue.value = percent
	}
}

// 杆杆调整弹窗
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

const sellList = ref([])
const buyList = ref([])

// 生成随机盘口数据（基于指定价格，6档深度）
const generateOrderBookWithPrice = (basePrice = 0) => {
	// 无效价格时，使用默认价格
	if (!basePrice || basePrice <= 0) {
		basePrice = 2956.86
	}
	
	const newSellList = []
	const newBuyList = []
	
	for (let i = 0; i < 6; i++) {
		newSellList.push({
			price: (basePrice + (6 - i) * 0.01).toFixed(2),
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

// 生成随机盘口数据（使用当前价格）
const generateOrderBook = () => {
	// 从全局市场数据获取实时价格
	const data = currentMarketData.value
	let basePrice = 0
	if (data && data.price > 0) {
		basePrice = data.price
	}
	generateOrderBookWithPrice(basePrice)
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
	
	// 恢复或初始化币种选择
	restoreSymbolFromStorage()
	
	// 生成盘口数据并启动定时器
	generateOrderBook()
	startTimer()
	
	// 启动倒计时
	updateCountdown()
	countdownTimer.value = setInterval(updateCountdown, 1000)
	
	// 初始化滑块区域信息
	initSliderRect()
	
	// 恢复滑点容差状态
	const savedSlippageEnabled = uni.getStorageSync('slippageEnabled')
	if (savedSlippageEnabled !== undefined && savedSlippageEnabled !== '') {
		slippageEnabled.value = savedSlippageEnabled
	}
	
	// 订阅WebSocket频道（余额实时推送）
	setupWebSocketSubscriptions()
	
	// walletStore 在 App.vue 中已初始化，不需要再次获取资产数据
	
	// 获取合约数据（仓位和挂单）
	fetchContractData()
	
	// 获取杠杆倍数配置
	fetchLeverageMultiples()
	
	// 获取秒合约周期配置
	fetchMicroPeriods()
	
	// 获取秒合约进行中订单
	fetchMicroActiveOrders()
	
	// 获取秒合约历史订单
	fetchMicroHistoryOrders()
})

// 页面显示时检查是否需要切换tab
onShow(() => {
	const targetTab = uni.getStorageSync('contractTab')
	if (targetTab) {
		mainTab.value = targetTab
		uni.removeStorageSync('contractTab')
	}
	
	// 刷新余额数据（从划转页面返回时需要）
	walletStore.fetchAssetData(true)
	
	// 刷新合约数据
	fetchContractData()
	
	// 刷新秒合约数据
	fetchMicroActiveOrders()
	fetchMicroHistoryOrders()
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
	
	// 取消WebSocket订阅
	cleanupWebSocketSubscriptions()
})

const goToKline = () => {
	uni.navigateTo({ url: `/pages/trade/kline?symbol=${encodeURIComponent(currentSymbol.value)}` })
}

const goTrade = () => {
	uni.switchTab({ url: '/pages/trade/trade' })
}

const goToOrders = () => {
	uni.navigateTo({ url: '/pages/trade/orders' })
}

// 跳转到划转页面
const goToTransfer = () => {
	uni.navigateTo({
		url: '/pages/wallet/transfer'
	})
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
		align-items: center;
		gap: 16rpx;
		min-width: 280rpx; // 增加最小宽度
		
		.config-label {
			font-size: 24rpx;
			color: #666;
			flex-shrink: 0;
		}
		
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
			flex-shrink: 0;
			
			&.mode {
				width: 56rpx;
				padding: 0;
			}
			
			&.leverage {
				min-width: 100rpx; // 保证按钮有足够空间
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

		.available-box {
			display: flex;
			justify-content: space-between;
			margin-bottom: 20rpx;
			font-size: 26rpx;
			.label { color: #999; }
			.value { color: #333; font-weight: 500; }
			
			.value-with-btn {
				display: flex;
				align-items: center;
				gap: 8rpx;
				
				.add-icon {
					width: 28rpx;
					height: 28rpx;
					background: #fcc419;
					color: #fff;
					border-radius: 50%;
					display: flex;
					align-items: center;
					justify-content: center;
					font-size: 20rpx;
					transition: all 0.2s ease;
					
					&:active {
						transform: scale(0.9);
						background: #e6b017;
					}
				}
			}
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
						
				input {
					flex: 1;
					font-size: 28rpx;
					color: #333;
					text-align: center;
				}
						
				&.disabled {
					.placeholder { color: #999; font-size: 28rpx; }
				}
			}
		}

		.input-group {
			height: 80rpx;
			background: #f5f5f5;
			border-radius: 8rpx;
			display: flex;
			align-items: center;
			padding: 0 24rpx;
			margin-bottom: 20rpx;
			
			input {
				flex: 1;
				font-size: 28rpx;
				color: #333;
			}
			
			.unit-picker {
				display: flex;
				align-items: center;
				gap: 8rpx;
				border-left: 1rpx solid #ddd;
				padding-left: 20rpx;
				font-size: 24rpx;
				color: #333;
			}
		}
		
		.slider-box {
			margin: 40rpx 0;
			position: relative;
			
			.custom-slider {
				width: 100%;
				margin: 0;
			}
			
			.slider-marks {
				display: flex;
				justify-content: space-between;
				margin-top: 16rpx;
				
				.mark {
					font-size: 22rpx;
					color: #999;
					text-align: center;
					flex: 1;
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
				cursor: pointer;
				
				.checkbox {
					width: 32rpx;
					height: 32rpx;
					border: 2rpx solid #d0d0d0;
					border-radius: 6rpx;
					display: flex;
					align-items: center;
					justify-content: center;
					transition: all 0.2s;
					background: #fff;
					
					&.checked {
						background: #3B82F6;
						border-color: #3B82F6;
					}
					
					.checkbox-icon {
						color: #fff;
						font-size: 20rpx;
						font-weight: bold;
					}
				}
				text {
					font-size: 26rpx;
					color: #333;
					&.dashed {
						border-bottom: 1rpx dotted #ccc;
						cursor: pointer;
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
				cursor: pointer;
				transition: background-color 0.2s;
					
				&:hover {
					background-color: rgba(0, 0, 0, 0.02);
				}
					
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
			cursor: pointer;
			transition: transform 0.2s;
				
			&:active {
				transform: scale(0.98);
			}
				
			.price-val { 
				font-size: 40rpx; 
				font-weight: bold;
				transition: color 0.3s ease;
				
				// 基于涨跌幅的静态颜色
				&.green { color: #00B894; }
				&.up { color: #00B894; }
				&.down { color: #F6465D; }
				
				// 基于价格变化的动态颜色（优先级更高）
				&.price-up { 
					color: #00B894 !important; 
					animation: priceFlash 0.5s ease;
				}
				&.price-down { 
					color: #F6465D !important; 
					animation: priceFlash 0.5s ease;
				}
			}
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
			
			.value-with-btn {
				display: flex;
				align-items: center;
				gap: 8rpx;
				
				.add-icon {
					width: 28rpx;
					height: 28rpx;
					background: #fcc419;
					color: #fff;
					border-radius: 50%;
					display: flex;
					align-items: center;
					justify-content: center;
					font-size: 20rpx;
					transition: all 0.2s ease;
					
					&:active {
						transform: scale(0.9);
						background: #e6b017;
					}
				}
			}
		}

		.seconds-actions {
			display: flex;
			flex-direction: column;
			gap: 16rpx;

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
				width: 100%;
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

/* 价格变化闪烁动画 */
@keyframes priceFlash {
	0% {
		opacity: 0.6;
		transform: scale(1.05);
	}
	50% {
		opacity: 1;
		transform: scale(1.1);
	}
	100% {
		opacity: 1;
		transform: scale(1);
	}
}

/* 永续合约仓位列表样式 */
.position-list, .pending-list, .history-list {
	padding: 20rpx;
}

.position-item, .pending-item, .history-item {
	background: #fff;
	border-radius: 16rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
	
	.position-header, .pending-header, .history-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16rpx;
		
		.left {
			display: flex;
			align-items: center;
			gap: 12rpx;
			
			.symbol {
				font-size: 30rpx;
				font-weight: bold;
				color: #333;
			}
			
			.side-tag {
				font-size: 22rpx;
				padding: 4rpx 12rpx;
				border-radius: 6rpx;
				font-weight: 500;
				
				&.long {
					background: rgba(0, 184, 148, 0.1);
					color: #00B894;
				}
				&.short {
					background: rgba(246, 70, 93, 0.1);
					color: #F6465D;
				}
			}
			
			.order-type {
				font-size: 22rpx;
				color: #666;
				background: #f5f5f5;
				padding: 4rpx 12rpx;
				border-radius: 6rpx;
			}
		}
		
		.pnl {
			font-size: 28rpx;
			font-weight: bold;
			
			&.profit { color: #00B894; }
			&.loss { color: #F6465D; }
			
			.rate {
				font-size: 22rpx;
				margin-left: 8rpx;
			}
		}
		
		.status {
			font-size: 24rpx;
			color: #FFA500;
		}
	}
	
	.position-info, .pending-info, .history-info {
		display: flex;
		flex-wrap: wrap;
		gap: 16rpx;
		padding: 16rpx 0;
		border-top: 1rpx solid #f0f0f0;
		border-bottom: 1rpx solid #f0f0f0;
		
		.info-item {
			flex: 1;
			min-width: 30%;
			
			.label {
				display: block;
				font-size: 22rpx;
				color: #999;
				margin-bottom: 4rpx;
			}
			
			.value {
				font-size: 26rpx;
				color: #333;
				font-weight: 500;
			}
		}
	}
	
	.position-actions, .pending-actions {
		display: flex;
		gap: 16rpx;
		margin-top: 16rpx;
		
		.btn {
			flex: 1;
			height: 72rpx;
			border-radius: 8rpx;
			font-size: 26rpx;
			font-weight: 500;
			border: none;
			
			&::after { border: none; }
			
			&.tpsl {
				background: #f0f0f0;
				color: #333;
			}
			
			&.close {
				background: #F6465D;
				color: #fff;
			}
			
			&.chase {
				background: #3B82F6;
				color: #fff;
			}
			
			&.cancel {
				background: #f0f0f0;
				color: #666;
			}
		}
	}
}

/* 止盈止损弹窗样式 */
.tpsl-mask {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	z-index: 9998;
}

.tpsl-modal {
	position: fixed;
	left: 50%;
	top: 50%;
	transform: translate(-50%, -50%) scale(0.9);
	width: 85%;
	max-width: 640rpx;
	background: #fff;
	border-radius: 20rpx;
	padding: 32rpx;
	z-index: 9999;
	opacity: 0;
	visibility: hidden;
	transition: all 0.25s ease;
	
	&.show {
		opacity: 1;
		visibility: visible;
		transform: translate(-50%, -50%) scale(1);
	}
	
	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 32rpx;
		
		.modal-title {
			font-size: 34rpx;
			font-weight: bold;
			color: #333;
		}
		
		.close-btn {
			width: 56rpx;
			height: 56rpx;
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 40rpx;
			color: #999;
			background: #f5f5f5;
			border-radius: 50%;
		}
	}
	
	.tpsl-form {
		.form-item {
			margin-bottom: 24rpx;
			
			.label {
				display: block;
				font-size: 26rpx;
				color: #333;
				font-weight: 500;
				margin-bottom: 12rpx;
			}
			
			input {
				width: 100%;
				height: 88rpx;
				background: #f5f5f5;
				border-radius: 12rpx;
				padding: 0 24rpx;
				font-size: 28rpx;
				box-sizing: border-box;
			}
			
			.hint {
				display: block;
				font-size: 22rpx;
				color: #999;
				margin-top: 8rpx;
			}
		}
	}
	
	.tpsl-confirm-btn {
		width: 100%;
		height: 88rpx;
		background: #3B82F6;
		border-radius: 12rpx;
		color: #fff;
		font-size: 30rpx;
		font-weight: bold;
		border: none;
		margin-top: 16rpx;
		
		&::after { border: none; }
	}
}
</style>
