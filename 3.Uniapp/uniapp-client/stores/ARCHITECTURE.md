# 全局市场数据状态管理架构设计

## 📐 架构概览

### 当前问题

在原有架构中，多个页面各自订阅WebSocket并处理行情数据：

```
WebSocket服务器
    ↓
    ├─> index.vue (订阅 + 处理)
    ├─> market.vue (订阅 + 处理)
    ├─> trade.vue (订阅 + 处理)
    └─> SymbolPicker.vue (订阅 + 处理)
```

**存在的问题：**
- ❌ 多个页面重复订阅同一个频道
- ❌ 同一条WebSocket消息被多个回调函数处理
- ❌ 数据处理逻辑分散，难以维护
- ❌ 可能出现数据不一致
- ❌ 内存占用较高（多份数据副本）

### 新架构设计

采用**单一数据源（Single Source of Truth）**模式：

```
WebSocket服务器
    ↓
marketStore.js (唯一订阅 + 唯一处理)
    ↓
全局响应式状态 (state.marketDataMap)
    ↓
    ├─> index.vue (只读取)
    ├─> market.vue (只读取)
    ├─> trade.vue (只读取)
    └─> SymbolPicker.vue (只读取)
```

**解决的问题：**
- ✅ 全局只订阅一次WebSocket
- ✅ 消息只处理一次，性能更好
- ✅ 数据处理逻辑集中，易于维护
- ✅ 保证数据一致性
- ✅ 内存占用更低（单份数据）

## 🏗️ 技术选型

### 1. 状态管理方案

**选择：Vue3 Composition API + reactive**

**为什么不使用Pinia？**
- ✅ 更轻量级（Uniapp中Pinia需要额外配置）
- ✅ 原生支持，无需额外依赖
- ✅ 学习成本更低
- ✅ 足够满足需求

**为什么使用reactive而非ref？**
- ✅ Map结构更适合存储多个币种数据
- ✅ 不需要.value访问，代码更简洁
- ✅ 性能更好（深层响应式）

### 2. 数据结构设计

```javascript
state = {
  // Map结构：key为 `${currency_id}_${legal_id}`
  marketDataMap: Map {
    "1_1" => { currency_id: 1, name: "BTC", price: 45678.90, ... },
    "2_1" => { currency_id: 2, name: "ETH", price: 2456.78, ... },
    ...
  },
  
  // Map结构：key为 currency_id
  currencyInfoMap: Map {
    1 => { currency_id: 1, name: "BTC", symbol: "BTC/USDT", ... },
    2 => { currency_id: 2, name: "ETH", symbol: "ETH/USDT", ... },
    ...
  },
  
  // 状态标记
  isInitialized: false,
  hasReceivedWsData: false,
  isInitialLoad: true,
  lastUpdateTime: null
}
```

**为什么使用Map而非Array？**
- ✅ O(1)时间复杂度查找（Array是O(n)）
- ✅ 更新单个币种数据效率高
- ✅ 支持按key快速访问
- ✅ 自动去重（同一个币种不会重复）

**为什么使用复合key `${currency_id}_${legal_id}`？**
- ✅ 唯一标识一个交易对
- ✅ 支持多法币（虽然目前只有USDT）
- ✅ 扩展性好

### 3. 单例模式

```javascript
// 默认导出单例
export default useMarketStore()

// 使用时
import marketStore from '@/stores/marketStore.js'
```

**为什么使用单例？**
- ✅ 确保全局只有一个实例
- ✅ 避免重复初始化
- ✅ 简化使用（不需要每次调用useMarketStore()）

## 🔄 数据流转

### 1. 初始化流程

```
App.vue: onMounted
    ↓
marketStore.initialize()
    ↓
├─> fetchMarketData() (从API加载基础数据)
│   ↓
│   state.marketDataMap.set(...)
│   state.currencyInfoMap.set(...)
│
└─> subscribeMarket() (订阅WebSocket)
    ↓
    wsClient.subscribe('market', handleMarketUpdate)
```

### 2. 数据更新流程

```
WebSocket推送消息
    ↓
wsClient.onMessage
    ↓
handleMarketUpdate(data)
    ↓
├─> 解析data
├─> 构建key: `${currency_id}_${legal_id}`
├─> state.marketDataMap.set(key, updatedData)
└─> state.hasReceivedWsData = true
    ↓
响应式系统自动触发
    ↓
所有使用computed的页面自动更新UI
```

### 3. 页面访问流程

```
页面组件
    ↓
import marketStore
    ↓
const price = computed(() => {
  const data = marketStore.getMarketData(currencyId)
  return data?.price || 0
})
    ↓
模板中使用 {{ price }}
    ↓
当WebSocket更新时，computed自动重新计算
    ↓
UI自动更新
```

## 📊 性能优化

### 1. 减少WebSocket订阅

**优化前：**
- 4个页面/组件各自订阅
- 每条消息触发4次回调
- 消息处理次数：1条消息 × 4个回调 = 4次处理

**优化后：**
- 全局只订阅1次
- 每条消息触发1次回调
- 消息处理次数：1条消息 × 1个回调 = 1次处理

**性能提升：75%** ⚡

### 2. 数据查找优化

**优化前（Array）：**
```javascript
// O(n) 时间复杂度
const data = marketList.find(item => item.currency_id === id)
```

**优化后（Map）：**
```javascript
// O(1) 时间复杂度
const data = state.marketDataMap.get(`${id}_1`)
```

**查找速度提升：100倍+**（当有100个币种时）⚡

### 3. 内存优化

**优化前：**
- index.vue: 存储5个币种 × 每个约200字节 = 1KB
- market.vue: 存储100个币种 × 每个约200字节 = 20KB
- trade.vue: 存储1个币种 × 约200字节 = 0.2KB
- SymbolPicker.vue: 存储100个币种 × 约200字节 = 20KB
- **总计：41.2KB**

**优化后：**
- marketStore: 存储100个币种 × 每个约200字节 = 20KB
- 其他页面：只存储引用，约0.1KB × 4 = 0.4KB
- **总计：20.4KB**

**内存节省：50%+** 💾

### 4. 响应式优化

使用`computed`而非`watch`：

```javascript
// ✅ 推荐：computed会自动缓存
const price = computed(() => {
  return marketStore.getMarketData(id)?.price || 0
})

// ❌ 不推荐：每次都要手动watch
watch(() => marketStore.state.marketDataMap, () => {
  // 手动查找和更新...
}, { deep: true })
```

## 🔒 数据一致性保证

### 1. 只读状态

```javascript
// state使用readonly包装，防止外部修改
state: readonly(state)
```

### 2. 统一更新接口

所有数据更新都通过`handleMarketUpdate`，不允许外部直接修改：

```javascript
// ✅ 正确：通过WebSocket更新
// WebSocket推送 → handleMarketUpdate → state更新

// ❌ 错误：外部直接修改
marketStore.state.marketDataMap.set(...) // 会报错
```

### 3. 时间戳标记

每次更新都记录时间戳，便于追踪：

```javascript
{
  ...data,
  timestamp: Date.now()
}
```

## 🎯 扩展性设计

### 1. 支持多法币

虽然目前只有USDT，但设计上已支持多法币：

```javascript
// key: `${currency_id}_${legal_id}`
const btcUsdt = marketStore.getMarketData(1, 1) // BTC/USDT
const btcCny = marketStore.getMarketData(1, 2)  // BTC/CNY (预留)
```

### 2. 支持多频道

可以轻松扩展支持其他频道：

```javascript
// 添加K线数据管理
const handleKlineUpdate = (data) => {
  state.klineDataMap.set(key, data)
}

wsClient.subscribe('kline', handleKlineUpdate)
```

### 3. 支持数据过滤和排序

页面可以基于全局数据做各种过滤和排序：

```javascript
// 只显示涨幅前10
const topGainers = computed(() => {
  return marketStore.getAllMarketData()
    .filter(item => item.change > 0)
    .sort((a, b) => b.change - a.change)
    .slice(0, 10)
})
```

## 🚧 迁移步骤

### 第一步：创建全局状态（已完成）

- ✅ 创建 `stores/marketStore.js`
- ✅ 实现数据管理和WebSocket处理

### 第二步：在App.vue中初始化

```vue
<script setup>
import marketStore from '@/stores/marketStore.js'

onMounted(() => {
  marketStore.initialize()
})
</script>
```

### 第三步：迁移各页面（逐步进行）

1. **market.vue** - 移除本地WebSocket订阅，使用全局状态
2. **index.vue** - 移除本地WebSocket订阅，使用全局状态
3. **trade.vue** - 移除本地WebSocket订阅，使用全局状态
4. **SymbolPicker.vue** - 移除本地WebSocket订阅，使用全局状态

### 第四步：测试验证

- 验证数据实时更新
- 验证多页面数据一致性
- 验证性能优化效果

## 📈 预期收益

| 指标 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| WebSocket订阅数 | 4个 | 1个 | -75% |
| 消息处理次数 | 4次/条 | 1次/条 | -75% |
| 数据查找速度 | O(n) | O(1) | 100倍+ |
| 内存占用 | 41KB | 20KB | -50% |
| 代码维护性 | 分散 | 集中 | ⭐⭐⭐⭐⭐ |
| 数据一致性 | 可能不一致 | 保证一致 | ⭐⭐⭐⭐⭐ |

## 🎓 设计模式

本方案应用了以下设计模式：

1. **单例模式（Singleton）** - 全局只有一个marketStore实例
2. **观察者模式（Observer）** - Vue的响应式系统
3. **发布-订阅模式（Pub-Sub）** - WebSocket消息分发
4. **仓储模式（Repository）** - 统一数据访问接口
5. **单一数据源（Single Source of Truth）** - 所有数据来自一个地方

## 🔮 未来优化方向

1. **数据持久化** - 使用localStorage缓存市场数据，减少API请求
2. **增量更新** - 只更新变化的数据，减少计算量
3. **懒加载** - 按需加载币种数据，减少初始加载时间
4. **数据压缩** - 对历史数据进行压缩存储
5. **离线支持** - 支持离线模式，显示缓存数据

## 📚 参考资料

- [Vue3 Composition API](https://vuejs.org/api/composition-api-setup.html)
- [Vue3 Reactivity API](https://vuejs.org/api/reactivity-core.html)
- [JavaScript Map](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Map)
- [Single Source of Truth](https://en.wikipedia.org/wiki/Single_source_of_truth)
- [Observer Pattern](https://refactoring.guru/design-patterns/observer)
