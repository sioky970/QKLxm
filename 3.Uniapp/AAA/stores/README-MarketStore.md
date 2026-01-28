# 全局市场数据状态管理使用指南

## 📋 概述

`marketStore.js` 提供了一个集中式的市场数据管理方案，所有币价数据通过WebSocket统一接收和存储，各页面只需读取全局状态即可。

## 🎯 核心优势

1. **单一数据源** - 避免多个页面重复订阅WebSocket
2. **性能优化** - 减少WebSocket连接数和消息处理次数
3. **数据一致性** - 所有页面看到的都是同一份数据
4. **易于维护** - 行情数据处理逻辑集中在一个地方
5. **响应式** - 基于Vue3 Composition API，数据变化自动更新UI

## 📦 使用方法

### 1. 在App.vue中初始化

```vue
<script setup>
import { onMounted } from 'vue'
import marketStore from '@/stores/marketStore.js'

onMounted(() => {
  // 初始化市场数据（只需调用一次）
  marketStore.initialize()
})
</script>
```

### 2. 在页面中使用

#### 示例1: Market页面（显示所有交易对）

```vue
<script setup>
import { computed } from 'vue'
import marketStore from '@/stores/marketStore.js'

// 获取所有市场数据（响应式）
const marketList = computed(() => {
  return marketStore.getAllMarketData().map(data => ({
    symbol: data.symbol,
    name: data.name,
    price: data.price.toFixed(data.price < 1 ? 4 : 2),
    change: data.change >= 0 ? `+${data.change.toFixed(2)}%` : `${data.change.toFixed(2)}%`,
    changeClass: data.change >= 0 ? 'up' : 'down',
    volume: formatVolume(data.volume),
    currency_id: data.currency_id,
    legal_id: data.legal_id
  }))
})

// 或者直接使用computed属性
const marketDataList = marketStore.marketDataList
</script>

<template>
  <view v-for="item in marketList" :key="item.currency_id">
    <text>{{ item.name }}</text>
    <text>{{ item.price }}</text>
    <text :class="item.changeClass">{{ item.change }}</text>
  </view>
</template>
```

#### 示例2: Trade页面（显示单个币种）

```vue
<script setup>
import { ref, computed, watch } from 'vue'
import marketStore from '@/stores/marketStore.js'

// 当前选择的币种ID
const currentCurrencyId = ref(8) // 例如TRX的currency_id

// 获取当前币种的市场数据（响应式）
const currentMarketData = computed(() => {
  return marketStore.getMarketData(currentCurrencyId.value)
})

// 当前价格（自动响应WebSocket更新）
const currentPrice = computed(() => {
  return currentMarketData.value?.price?.toFixed(4) || '0.0000'
})

// 涨跌幅
const priceChange = computed(() => {
  const change = currentMarketData.value?.change || 0
  return change >= 0 ? `+${change.toFixed(2)}%` : `${change.toFixed(2)}%`
})

// 监听价格变化
watch(currentPrice, (newPrice, oldPrice) => {
  console.log('价格更新:', oldPrice, '->', newPrice)
  // 可以在这里触发动画效果等
})
</script>

<template>
  <view>
    <text>当前价格: {{ currentPrice }}</text>
    <text>涨跌幅: {{ priceChange }}</text>
  </view>
</template>
```

#### 示例3: Index页面（显示随机5个交易对）

```vue
<script setup>
import { computed } from 'vue'
import marketStore from '@/stores/marketStore.js'

// 获取随机5个交易对
const randomMarketList = computed(() => {
  const allData = marketStore.getAllMarketData()
  
  // 随机打乱
  const shuffled = [...allData].sort(() => Math.random() - 0.5)
  
  // 取前5个
  return shuffled.slice(0, 5).map(data => ({
    symbol: data.name,
    price: data.price.toFixed(2),
    change: data.change >= 0 ? `+${data.change.toFixed(2)}%` : `${data.change.toFixed(2)}%`,
    changeClass: data.change >= 0 ? 'up' : 'down'
  }))
})
</script>

<template>
  <view v-for="(item, index) in randomMarketList" :key="index">
    <text>{{ item.symbol }}</text>
    <text>{{ item.price }}</text>
    <text :class="item.changeClass">{{ item.change }}</text>
  </view>
</template>
```

#### 示例4: SymbolPicker组件（币种选择器）

```vue
<script setup>
import { computed } from 'vue'
import marketStore from '@/stores/marketStore.js'

// 获取所有币种列表
const symbolList = computed(() => {
  return marketStore.getAllMarketData().map(data => ({
    symbol: data.symbol,
    name: data.name,
    quote: 'USDT',
    price: data.price.toFixed(data.price < 1 ? 4 : 2),
    priceNum: data.price,
    change: data.change >= 0 ? `+${data.change.toFixed(2)}%` : `${data.change.toFixed(2)}%`,
    changeNum: data.change,
    changeClass: data.change >= 0 ? 'up' : 'down',
    volume: formatVolume(data.volume),
    volumeNum: data.volume,
    currency_id: data.currency_id,
    legal_id: data.legal_id,
    tag: data.tag,
    tagClass: data.tagClass
  }))
})

// 支持搜索和排序
const filteredSymbols = computed(() => {
  let list = [...symbolList.value]
  
  // 搜索过滤
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toUpperCase()
    list = list.filter(item => item.name.includes(keyword))
  }
  
  // 排序
  if (sortField.value === 'price') {
    list.sort((a, b) => {
      return sortOrder.value === 'asc' ? a.priceNum - b.priceNum : b.priceNum - a.priceNum
    })
  }
  
  return list
})
</script>

<template>
  <view v-for="item in filteredSymbols" :key="item.currency_id">
    <text>{{ item.name }}</text>
    <text>{{ item.price }}</text>
    <text :class="item.changeClass">{{ item.change }}</text>
  </view>
</template>
```

## 🔧 API参考

### 初始化方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `initialize()` | 初始化市场数据并订阅WebSocket | `marketStore.initialize()` |
| `fetchMarketData()` | 从后端重新加载市场数据 | `await marketStore.fetchMarketData()` |

### 数据访问方法

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `getMarketData(currencyId, legalId)` | `currencyId: number`, `legalId: number = 1` | `Object \| null` | 获取指定币种的市场数据 |
| `getMarketDataByName(name)` | `name: string` | `Object \| null` | 根据币种名称获取数据 |
| `getAllMarketData()` | 无 | `Array` | 获取所有市场数据（数组） |
| `marketDataList` | 无 | `ComputedRef<Array>` | 响应式的所有市场数据 |
| `getCurrencyInfo(currencyId)` | `currencyId: number` | `Object \| null` | 获取币种基础信息 |
| `getAllCurrencyInfo()` | 无 | `Array` | 获取所有币种基础信息 |

### WebSocket管理方法

| 方法 | 说明 |
|------|------|
| `subscribeMarket()` | 订阅WebSocket市场行情（自动调用，通常不需手动调用） |
| `unsubscribeMarket()` | 取消订阅WebSocket市场行情 |

### 状态属性

| 属性 | 类型 | 说明 |
|------|------|------|
| `state` | `Readonly<Object>` | 只读的全局状态 |
| `isReady` | `ComputedRef<boolean>` | 数据是否已初始化 |
| `hasData` | `ComputedRef<boolean>` | 是否有数据 |

### 市场数据对象结构

```javascript
{
  currency_id: 1,           // 币种ID
  legal_id: 1,              // 法币ID
  symbol: "BTC/USDT",       // 交易对符号
  name: "BTC",              // 币种名称
  now_price: 45678.90,      // 当前价格
  price: 45678.90,          // 价格（同now_price）
  change: 2.35,             // 涨跌幅（百分比，不带%符号）
  volume: 12345678.90,      // 成交量
  tag: "热门",              // 标签
  tagClass: "hot",          // 标签样式类
  timestamp: 1704096000000  // 更新时间戳
}
```

## 🔄 数据更新流程

```
WebSocket推送币价数据
    ↓
marketStore.handleMarketUpdate() 接收并处理
    ↓
更新 state.marketDataMap (响应式)
    ↓
所有使用computed的页面自动更新UI
```

## ⚠️ 注意事项

1. **只需初始化一次** - 在App.vue的onMounted中调用 `initialize()` 即可
2. **不要重复订阅** - marketStore内部已订阅WebSocket，页面无需再订阅
3. **使用computed** - 页面中使用computed访问数据，确保响应式更新
4. **避免直接修改** - state是只读的，不要直接修改状态
5. **性能优化** - getAllMarketData()会创建新数组，频繁调用时建议使用marketDataList

## 🚀 迁移指南

### 旧代码（每个页面各自订阅）

```javascript
// ❌ 旧方式 - 不推荐
const handleMarketUpdate = (data) => {
  // 处理WebSocket数据...
}

onMounted(() => {
  wsClient.subscribe('market', handleMarketUpdate)
})

onUnmounted(() => {
  wsClient.unsubscribe('market', handleMarketUpdate)
})
```

### 新代码（使用全局状态）

```javascript
// ✅ 新方式 - 推荐
import marketStore from '@/stores/marketStore.js'

const currentPrice = computed(() => {
  const data = marketStore.getMarketData(currencyId.value)
  return data?.price?.toFixed(4) || '0.0000'
})

// 不需要订阅和取消订阅！
```

## 🎨 最佳实践

1. **在App.vue中初始化** - 确保全局只初始化一次
2. **使用computed访问数据** - 保证响应式更新
3. **按需格式化** - 在computed中格式化价格、涨跌幅等
4. **缓存计算结果** - 对于复杂计算，使用computed缓存
5. **watch监听变化** - 需要监听价格变化时使用watch

## 🔍 调试

```javascript
// 查看当前状态
console.log('市场数据数量:', marketStore.state.marketDataMap.size)
console.log('是否已初始化:', marketStore.isReady.value)
console.log('是否有数据:', marketStore.hasData.value)

// 查看所有数据
console.log('所有市场数据:', marketStore.getAllMarketData())

// 查看特定币种
console.log('BTC数据:', marketStore.getMarketDataByName('BTC'))
```
