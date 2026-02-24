<template>
  <div class="currency-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="币种名称">
          <a-input
            v-model="searchForm.keyword"
            placeholder="请输入币种名称"
            style="width: 200px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="币种类型">
          <a-select
            v-model="searchForm.type"
            placeholder="全部类型"
            style="width: 150px"
            allow-clear
          >
            <a-option value="coin">数字货币</a-option>
            <a-option value="token">代币</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="启用状态">
          <a-select
            v-model="searchForm.is_display"
            placeholder="全部"
            style="width: 120px"
            allow-clear
          >
            <a-option :value="1">启用</a-option>
            <a-option :value="0">停用</a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><icon-search /></template>
              搜索
            </a-button>
            <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <a-card class="table-card" :bordered="false" style="margin-top: 16px">
      <template #title>
        <div class="table-header">
          <span>币种列表</span>
          <a-space>
            <a-button @click="handleRefresh">
              <template #icon><icon-sync /></template>
              刷新
            </a-button>
          </a-space>
        </div>
      </template>

      <div class="table-scroll" @scroll="handleScroll">
        <a-table
          :columns="columns"
          :data="tableData"
          :pagination="false"
          :loading="loading"
          row-key="id"
        >
          <template #logo="{ record }">
            <a-avatar :size="40" shape="square">
              <img v-if="record.logo" :src="record.logo" alt="logo" />
              <span v-else>{{ record.name.substring(0, 1) }}</span>
            </a-avatar>
          </template>

          <template #name="{ record }">
            <a-space>
              <span style="font-weight: 500">{{ record.name }}</span>
              <a-tag v-if="record.is_legal === 1" color="blue" size="small"
                >法币</a-tag
              >
            </a-space>
          </template>

          <template #type="{ record }">
            <a-tag :color="record.type === 'coin' ? 'arcoblue' : 'green'">
              {{ record.type === 'coin' ? '数字货币' : '代币' }}
            </a-tag>
          </template>

          <template #features="{ record }">
            <a-space>
              <a-tag v-if="record.is_change === 1" color="green" size="small"
                >币币</a-tag
              >
              <a-tag v-if="record.is_lever === 1" color="orange" size="small"
                >杠杆</a-tag
              >
              <a-tag v-if="record.is_micro === 1" color="purple" size="small"
                >秒合约</a-tag
              >
            </a-space>
          </template>

          <template #rate="{ record }">
            <a-tag color="blue">{{ formatRate(record.rate) }}</a-tag>
          </template>

          <template #latest_price="{ record }">
            <a-tag color="green">{{ formatPrice(record.latest_price) }}</a-tag>
          </template>

          <template #kline_count="{ record }">
            <span>{{ formatCount(record.kline_count) }}</span>
          </template>

          <template #last_quote_time="{ record }">
            <span>{{ formatTime(record.last_quote_time) }}</span>
          </template>

          <template #is_display="{ record }">
            <a-switch
              v-model="record.is_display"
              :checked-value="1"
              :unchecked-value="0"
              @change="handleToggleDisplay(record)"
            >
              <template #checked>启用</template>
              <template #unchecked>停用</template>
            </a-switch>
          </template>

          <template #actions="{ record }">
            <a-button type="text" size="small" @click="handleOpenRisk(record)">
              设置风控
            </a-button>
          </template>
        </a-table>
        <div v-if="loadingMore" class="table-loading">加载中...</div>
        <div v-else-if="isFinished" class="table-loading">已加载全部</div>
      </div>
    </a-card>

    <a-modal
      v-model:visible="riskModalVisible"
      title="币种风控设置"
      :width="600"
      @ok="handleRiskSubmit"
      @cancel="riskModalVisible = false"
    >
      <a-form
        :model="riskForm"
        :label-col-props="{ span: 10 }"
        :wrapper-col-props="{ span: 14 }"
      >
        <a-alert type="warning" banner class="risk-warning" :hide-icon="false">
          <template #title>
            <span>风险控制说明</span>
          </template>
          <span
            >风控设置用于控制特定币种的交易结果。请注意：<strong>盈利</strong>指用户交易盈利，<strong>亏损</strong>指用户交易亏损。这些设置将直接影响交易的最终结果。</span
          >
        </a-alert>
        <a-form-item label="币种名称">
          <a-typography-text strong>{{ riskForm.name }}</a-typography-text>
        </a-form-item>

        <a-divider orientation="left" style="margin: 24px 0 16px 0">
          <div style="display: flex; align-items: center">
            <span>概率风控</span>
            <a-tooltip content="通过设置盈利概率来控制交易结果的概率分布">
              <icon-question-circle
                style="margin-left: 8px; color: var(--color-text-3)"
              />
            </a-tooltip>
          </div>
        </a-divider>
        <a-alert
          type="info"
          banner
          closable
          class="risk-info"
          :hide-icon="false"
        >
          <template #title>
            <span>关于“盈利”概念的说明</span>
          </template>
          <span
            >在风控设置中，“盈利”指交易完成后用户的盈亏状态。例如，当设置为“盈利”时，交易结果会被调整为用户盈利；设置为“亏损”时，交易结果会被调整为用户亏损。</span
          >
        </a-alert>
        <a-form-item label="启用概率控">
          <a-switch
            v-model="riskForm.risk_prob_enabled"
            :checked-value="1"
            :unchecked-value="0"
          >
            <template #checked>开启</template>
            <template #unchecked>关闭</template>
          </a-switch>
        </a-form-item>
        <a-form-item
          v-show="riskForm.risk_prob_enabled === 1"
          label="盈利概率(%)"
          :rules="[
            {
              required: riskForm.risk_prob_enabled === 1,
              message: '请输入盈利概率',
            },
            {
              type: 'number',
              min: 0,
              max: 100,
              message: '盈利概率应在0-100之间',
            },
          ]"
        >
          <template #help>
            <span
              >当用户交易时，有
              <strong>{{ riskForm.risk_profit_probability }}%</strong>
              的概率使交易结果为“盈利”，<strong
                >{{ 100 - riskForm.risk_profit_probability }}%</strong
              >
              的概率使交易结果为“亏损”</span
            >
          </template>
          <a-slider
            v-model="riskForm.risk_profit_probability"
            :min="0"
            :max="100"
            :step="1"
            style="width: 80%; margin-right: 16px"
            :tip-format="(value) => `${value}%`"
          />
          <a-input-number
            v-model="riskForm.risk_profit_probability"
            :min="0"
            :max="100"
            :precision="0"
            style="width: 80px"
          />
        </a-form-item>

        <a-divider orientation="left" style="margin: 24px 0 16px 0">
          <div style="display: flex; align-items: center">
            <span>金额风控</span>
            <a-tooltip content="针对特定交易金额范围设置风控规则">
              <icon-question-circle
                style="margin-left: 8px; color: var(--color-text-3)"
              />
            </a-tooltip>
          </div>
        </a-divider>
        <a-form-item label="启用金额控">
          <a-switch
            v-model="riskForm.risk_money_enabled"
            :checked-value="1"
            :unchecked-value="0"
          >
            <template #checked>开启</template>
            <template #unchecked>关闭</template>
          </a-switch>
        </a-form-item>
        <div v-show="riskForm.risk_money_enabled === 1">
          <a-form-item
            label="金额范围"
            :rules="[
              {
                required: riskForm.risk_money_enabled === 1,
                message: '请输入金额范围',
              },
            ]"
          >
            <a-space style="width: 100%">
              <a-input-number
                v-model="riskForm.risk_money_min"
                :min="0"
                :precision="8"
                style="width: 120px"
                placeholder="最小值"
              />
              <span style="color: var(--color-text-4)">-</span>
              <a-input-number
                v-model="riskForm.risk_money_max"
                :min="0"
                :precision="8"
                style="width: 120px"
                placeholder="最大值"
              />
            </a-space>
          </a-form-item>
          <a-form-item
            label="金额控结果"
            :rules="[
              {
                required: riskForm.risk_money_enabled === 1,
                message: '请选择金额控结果',
              },
            ]"
          >
            <a-radio-group v-model="riskForm.risk_money_result">
              <a-radio :value="0">
                <span>不干预</span>
                <a-tooltip content="不对交易结果进行任何干预">
                  <icon-question-circle
                    style="
                      margin-left: 4px;
                      color: var(--color-text-3);
                      font-size: 12px;
                    "
                  />
                </a-tooltip>
              </a-radio>
              <a-radio :value="1">
                <span>盈利</span>
                <a-tooltip content="确保在此金额范围内的交易结果为用户盈利">
                  <icon-question-circle
                    style="
                      margin-left: 4px;
                      color: var(--color-text-3);
                      font-size: 12px;
                    "
                  />
                </a-tooltip>
              </a-radio>
              <a-radio :value="-1">
                <span>亏损</span>
                <a-tooltip content="确保在此金额范围内的交易结果为用户亏损">
                  <icon-question-circle
                    style="
                      margin-left: 4px;
                      color: var(--color-text-3);
                      font-size: 12px;
                    "
                  />
                </a-tooltip>
              </a-radio>
            </a-radio-group>
          </a-form-item>
        </div>

        <a-divider orientation="left" style="margin: 24px 0 16px 0">
          <div style="display: flex; align-items: center">
            <span>时间风控</span>
            <a-tooltip content="在特定时间段内应用风控规则">
              <icon-question-circle
                style="margin-left: 8px; color: var(--color-text-3)"
              />
            </a-tooltip>
          </div>
        </a-divider>
        <a-form-item label="启用时间段控">
          <a-switch
            v-model="riskForm.risk_time_enabled"
            :checked-value="1"
            :unchecked-value="0"
          >
            <template #checked>开启</template>
            <template #unchecked>关闭</template>
          </a-switch>
        </a-form-item>
        <div v-show="riskForm.risk_time_enabled === 1">
          <a-form-item
            label="时间范围"
            :rules="[
              {
                required: riskForm.risk_time_enabled === 1,
                message: '请输入时间范围',
              },
            ]"
          >
            <a-space>
              <a-time-picker
                v-model="riskForm.risk_time_start"
                format="HH:mm"
                placeholder="开始时间"
                style="width: 120px"
              />
              <span style="color: var(--color-text-4)">至</span>
              <a-time-picker
                v-model="riskForm.risk_time_end"
                format="HH:mm"
                placeholder="结束时间"
                style="width: 120px"
              />
            </a-space>
          </a-form-item>
          <a-form-item
            label="时间控结果"
            :rules="[
              {
                required: riskForm.risk_time_enabled === 1,
                message: '请选择时间控结果',
              },
            ]"
          >
            <a-radio-group v-model="riskForm.risk_time_result">
              <a-radio :value="0">
                <span>不干预</span>
                <a-tooltip content="在该时间段内不对交易结果进行任何干预">
                  <icon-question-circle
                    style="
                      margin-left: 4px;
                      color: var(--color-text-3);
                      font-size: 12px;
                    "
                  />
                </a-tooltip>
              </a-radio>
              <a-radio :value="1">
                <span>盈利</span>
                <a-tooltip content="在该时间段内确保交易结果为用户盈利">
                  <icon-question-circle
                    style="
                      margin-left: 4px;
                      color: var(--color-text-3);
                      font-size: 12px;
                    "
                  />
                </a-tooltip>
              </a-radio>
              <a-radio :value="-1">
                <span>亏损</span>
                <a-tooltip content="在该时间段内确保交易结果为用户亏损">
                  <icon-question-circle
                    style="
                      margin-left: 4px;
                      color: var(--color-text-3);
                      font-size: 12px;
                    "
                  />
                </a-tooltip>
              </a-radio>
            </a-radio-group>
          </a-form-item>
        </div>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, onUnmounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconQuestionCircle,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const loading = ref(false);
  const loadingMore = ref(false);
  const isFinished = ref(false);
  const riskModalVisible = ref(false);

  // 防抖定时器
  let scrollTimer: ReturnType<typeof setTimeout> | null = null;
  // 当前加载的页码，防止重复加载
  const loadingPage = ref<number | null>(null);

  const searchForm = reactive({
    keyword: '',
    type: undefined as string | undefined,
    is_display: undefined as number | undefined,
  });

  const tableData = ref<any[]>([]);

  const pagination = reactive({
    current: 1,
    pageSize: 30,
    total: 0,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: 'Logo', dataIndex: 'logo', slotName: 'logo', width: 80 },
    { title: '币种名称', dataIndex: 'name', slotName: 'name' },
    { title: '类型', dataIndex: 'type', slotName: 'type', width: 100 },
    { title: '功能', slotName: 'features', width: 180 },
    { title: '小数位', dataIndex: 'decimal_scale', width: 80 },
    { title: '汇率(USDT)', dataIndex: 'rate', slotName: 'rate', width: 120 },
    {
      title: '实时价格(USDT)',
      dataIndex: 'latest_price',
      slotName: 'latest_price',
      width: 150,
    },
    {
      title: 'K线数据(1min)',
      dataIndex: 'kline_count',
      slotName: 'kline_count',
      width: 140,
    },
    {
      title: '上次获取时间',
      dataIndex: 'last_quote_time',
      slotName: 'last_quote_time',
      width: 180,
    },
    { title: '排序', dataIndex: 'sort', width: 80 },
    {
      title: '状态',
      dataIndex: 'is_display',
      slotName: 'is_display',
      width: 100,
    },
    { title: '操作', slotName: 'actions', width: 120, fixed: 'right' },
  ];

  const formatRate = (rate?: number) => {
    const value = typeof rate === 'number' ? rate : 0;
    return value.toFixed(4);
  };

  const formatPrice = (price?: number) => {
    const value = typeof price === 'number' ? price : 0;
    return value > 0 ? value.toFixed(6) : '-';
  };

  const formatCount = (count?: number) => {
    if (typeof count !== 'number') return '-';
    return count.toLocaleString();
  };

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('zh-CN', { hour12: false });
  };

  const sortByMarketCap = (list: any[]) =>
    list.sort((a: any, b: any) => Number(b.sort ?? 0) - Number(a.sort ?? 0));

  // 根据 ID 去重
  const deduplicateById = (list: any[]) => {
    const seen = new Set();
    return list.filter((item) => {
      if (seen.has(item.id)) {
        return false;
      }
      seen.add(item.id);
      return true;
    });
  };

  const fetchCurrencyList = async (reset = false) => {
    // 如果正在加载或已到底，直接返回
    if (loading.value || loadingMore.value || isFinished.value) return;

    // 重置时清空数据
    if (reset) {
      pagination.current = 1;
      tableData.value = [];
      isFinished.value = false;
      loadingPage.value = null;
    }

    // 检查是否已经在加载当前页码
    const targetPage = pagination.current;
    if (loadingPage.value === targetPage) {
      return;
    }

    // 标记当前加载的页码
    loadingPage.value = targetPage;

    try {
      if (targetPage === 1) {
        loading.value = true;
      } else {
        loadingMore.value = true;
      }

      const params = {
        page: targetPage,
        page_size: pagination.pageSize,
        keyword: searchForm.keyword || undefined,
        type: searchForm.type,
        is_display: searchForm.is_display,
      };

      const response = await axios.post('/admin/currency/list', params);
      if (response.data && response.data.data) {
        const list = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;

        if (reset || targetPage === 1) {
          // 重置时直接设置数据
          tableData.value = sortByMarketCap(list);
        } else {
          // 追加数据并去重
          const merged = deduplicateById([...tableData.value, ...list]);
          tableData.value = sortByMarketCap(merged);
        }

        // 检查是否已加载全部
        if (list.length === 0 || tableData.value.length >= pagination.total) {
          isFinished.value = true;
        } else {
          // 只有加载成功且未到底时才增加页码
          pagination.current = targetPage + 1;
        }
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    } finally {
      loading.value = false;
      loadingMore.value = false;
      loadingPage.value = null;
    }
  };

  const handleSearch = () => {
    fetchCurrencyList(true);
  };

  const handleReset = () => {
    searchForm.keyword = '';
    searchForm.type = undefined;
    searchForm.is_display = undefined;
    fetchCurrencyList(true);
  };

  const handleRefresh = () => {
    fetchCurrencyList(true);
  };

  // 带防抖的滚动处理函数
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement;
    if (!target) return;

    // 清除之前的定时器
    if (scrollTimer) {
      clearTimeout(scrollTimer);
    }

    // 设置防抖延迟
    scrollTimer = setTimeout(() => {
      // 检查是否已经在加载或已到底
      if (loading.value || loadingMore.value || isFinished.value) {
        return;
      }

      // 检查是否滚动到底部
      const { scrollTop } = target;
      const { clientHeight } = target;
      const { scrollHeight } = target;

      if (scrollTop + clientHeight >= scrollHeight - 50) {
        fetchCurrencyList();
      }
    }, 150); // 150ms 防抖延迟
  };

  // 组件卸载时清理定时器
  onUnmounted(() => {
    if (scrollTimer) {
      clearTimeout(scrollTimer);
      scrollTimer = null;
    }
  });

  const handleToggleDisplay = async (record: any) => {
    try {
      await axios.post('/admin/currency/toggle-display', {
        id: record.id,
        is_display: record.is_display,
      });
      Message.success('启用状态已更新');
    } catch (error: any) {
      Message.error(error.message || '更新失败');
      record.is_display = record.is_display === 1 ? 0 : 1;
    }
  };

  const riskForm = reactive({
    id: 0,
    name: '',
    risk_prob_enabled: 0,
    risk_profit_probability: 50,
    risk_money_enabled: 0,
    risk_money_min: 0,
    risk_money_max: 0,
    risk_money_result: 0,
    risk_time_enabled: 0,
    risk_time_start: '',
    risk_time_end: '',
    risk_time_result: 0,
  });

  const handleOpenRisk = (record: any) => {
    // 确保时间格式正确
    const formatTimeValue = (timeStr: string) => {
      if (!timeStr) return undefined;
      // 如果是时间字符串格式，则直接返回
      if (/^\d{2}:\d{2}$/.test(timeStr)) {
        return timeStr;
      }
      // 否则返回undefined
      return undefined;
    };

    Object.assign(riskForm, {
      id: record.id,
      name: record.name,
      risk_prob_enabled: record.risk_prob_enabled ?? 0,
      risk_profit_probability: record.risk_profit_probability ?? 50,
      risk_money_enabled: record.risk_money_enabled ?? 0,
      risk_money_min: record.risk_money_min ?? 0,
      risk_money_max: record.risk_money_max ?? 0,
      risk_money_result: record.risk_money_result ?? 0,
      risk_time_enabled: record.risk_time_enabled ?? 0,
      risk_time_start: formatTimeValue(record.risk_time_start),
      risk_time_end: formatTimeValue(record.risk_time_end),
      risk_time_result: record.risk_time_result ?? 0,
    });
    riskModalVisible.value = true;
  };

  const handleRiskSubmit = async () => {
    try {
      await axios.post('/admin/currency/update-risk', {
        id: riskForm.id,
        risk_prob_enabled: riskForm.risk_prob_enabled,
        risk_profit_probability: riskForm.risk_profit_probability,
        risk_money_enabled: riskForm.risk_money_enabled,
        risk_money_min: riskForm.risk_money_min,
        risk_money_max: riskForm.risk_money_max,
        risk_money_result: riskForm.risk_money_result,
        risk_time_enabled: riskForm.risk_time_enabled,
        risk_time_start: riskForm.risk_time_start,
        risk_time_end: riskForm.risk_time_end,
        risk_time_result: riskForm.risk_time_result,
      });
      Message.success('风控设置已保存');
      riskModalVisible.value = false;
      fetchCurrencyList(true);
    } catch (error: any) {
      Message.error(error.message || '风控设置失败');
    }
  };

  onMounted(() => {
    fetchCurrencyList(true);
  });
</script>

<style scoped lang="less">
  .risk-control-container {
    .arco-divider {
      margin: 24px 0;
    }

    .arco-form-item {
      margin-bottom: 16px;
    }
  }

  .risk-section {
    margin: 16px 0;
    padding: 16px;
    background-color: var(--color-fill-1);
    border-radius: 4px;
  }

  .risk-warning {
    margin-bottom: 16px;
  }

  .risk-info {
    margin-bottom: 16px;
  }
</style>

<style lang="less">
  .currency-list-container {
    padding: 20px;

    .search-card {
      :deep(.arco-card-body) {
        padding: 20px;
      }
    }

    .table-card {
      :deep(.arco-card-body) {
        padding: 20px;
      }

      .table-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 16px;
        font-weight: 500;
      }
    }

    :deep(.arco-table) {
      margin-top: 16px;
    }

    .table-scroll {
      max-height: 640px;
      overflow: auto;
    }

    .table-loading {
      padding: 12px 0;
      text-align: center;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }
</style>
