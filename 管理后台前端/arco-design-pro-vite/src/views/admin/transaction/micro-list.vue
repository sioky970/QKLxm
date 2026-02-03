<template>
  <div class="micro-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="用户账号">
          <a-input
            v-model="searchForm.account_number"
            placeholder="请输入用户账号"
            style="width: 200px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="币种">
          <a-select
            v-model="searchForm.currency_id"
            placeholder="全部币种"
            style="width: 180px"
            allow-clear
          >
            <a-option
              v-for="currency in currencyList"
              :key="currency.id"
              :value="currency.id"
            >
              {{ currency.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 140px"
            allow-clear
          >
            <a-option :value="0">待结算</a-option>
            <a-option :value="1">已结算</a-option>
            <a-option :value="2">已取消</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="结果">
          <a-select
            v-model="searchForm.result"
            placeholder="全部"
            style="width: 120px"
            allow-clear
          >
            <a-option value="win">赢</a-option>
            <a-option value="lose">输</a-option>
            <a-option value="draw">平</a-option>
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
          <span>交割合约订单</span>
          <a-space>
            <a-button
              :disabled="selectedRowKeys.length === 0"
              @click="handleBatchRisk"
            >
              <template #icon><icon-settings /></template>
              批量风控
            </a-button>
            <a-button @click="handleRefresh">
              <template #icon><icon-sync /></template>
              刷新
            </a-button>
          </a-space>
        </div>
      </template>

      <a-table
        :columns="columns"
        :data="tableData"
        :pagination="pagination"
        :loading="loading"
        :scroll="{ x: 1800 }"
        :row-selection="rowSelection"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #user="{ record }">
          <a-space direction="vertical" :size="2">
            <span style="font-weight: 500">{{ record.account_number }}</span>
            <a-tag size="small">ID: {{ record.user_id }}</a-tag>
          </a-space>
        </template>

        <template #symbol="{ record }">
          <span>{{ record.currency_name }}/USDT</span>
        </template>

        <template #type="{ record }">
          <a-tag :color="record.type === 1 ? 'green' : 'red'">
            {{ record.type === 1 ? '买涨' : '买跌' }}
          </a-tag>
        </template>

        <template #status="{ record }">
          <a-tag :color="statusColor(record.status)">
            {{ statusText(record.status) }}
          </a-tag>
        </template>

        <template #result="{ record }">
          <a-tag :color="resultColor(record.profit_result)">
            {{ resultText(record.profit_result) }}
          </a-tag>
        </template>

        <template #profits="{ record }">
          <span
            :class="record.fact_profits >= 0 ? 'amount-plus' : 'amount-minus'"
          >
            {{ formatBalance(record.fact_profits) }}
          </span>
        </template>

        <template #created_at="{ record }">
          {{ formatDateTime(record.created_at) }}
        </template>

        <template #actions="{ record }">
          <a-space>
            <a-button type="text" size="small" @click="handleDetail(record)">
              <template #icon><icon-eye /></template>
              详情
            </a-button>
            <a-button
              v-if="record.status === 0"
              type="text"
              size="small"
              @click="handleRisk(record)"
            >
              <template #icon><icon-settings /></template>
              风控
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 订单详情弹窗 -->
    <a-modal
      v-model:visible="detailModalVisible"
      title="订单详情"
      :width="700"
      :footer="false"
    >
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="订单ID">{{
          detailData.id
        }}</a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="statusColor(detailData.status)">
            {{ statusText(detailData.status) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="用户账号">{{
          detailData.account_number
        }}</a-descriptions-item>
        <a-descriptions-item label="用户ID">{{
          detailData.user_id
        }}</a-descriptions-item>
        <a-descriptions-item label="币种"
          >{{ detailData.currency_name }}/USDT</a-descriptions-item
        >
        <a-descriptions-item label="方向">
          <a-tag :color="detailData.type === 1 ? 'green' : 'red'">
            {{ detailData.type === 1 ? '买涨' : '买跌' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="秒数"
          >{{ detailData.seconds }}秒</a-descriptions-item
        >
        <a-descriptions-item label="下单金额">{{
          formatBalance(detailData.number)
        }}</a-descriptions-item>
        <a-descriptions-item label="赔率"
          >{{ detailData.profit_ratio }}%</a-descriptions-item
        >
        <a-descriptions-item label="结果">
          <a-tag :color="resultColor(detailData.profit_result)">
            {{ resultText(detailData.profit_result) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="开仓价">
          <span style="color: var(--color-text-2)">{{
            formatBalance(detailData.open_price)
          }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="结算价">
          <span style="color: var(--color-primary-6)">{{
            formatBalance(detailData.end_price) || '-'
          }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="盈亏">
          <span
            :class="
              (detailData.fact_profits || 0) >= 0
                ? 'amount-plus'
                : 'amount-minus'
            "
          >
            {{ formatBalance(detailData.fact_profits) }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="预设结果">
          <a-tag v-if="detailData.pre_profit_result === 1" color="green"
            >赢</a-tag
          >
          <a-tag v-else-if="detailData.pre_profit_result === 2" color="red"
            >输</a-tag
          >
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item label="下单时间" :span="2">
          {{ formatDateTime(detailData.created_at) }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>

    <a-modal
      v-model:visible="riskModalVisible"
      title="设置风控"
      :width="420"
      @ok="handleRiskSubmit"
      @cancel="riskModalVisible = false"
    >
      <a-form :model="riskForm" :label-col-props="{ span: 6 }">
        <a-form-item label="订单ID">
          <a-input :model-value="String(riskForm.id)" disabled />
        </a-form-item>
        <a-form-item label="预设结果">
          <a-select v-model="riskForm.pre_profit_result" placeholder="不干预">
            <a-option :value="0">不干预</a-option>
            <a-option :value="1">赢</a-option>
            <a-option :value="2">输</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="结算价格">
          <a-input-number
            v-model="riskForm.end_price"
            :min="0"
            :precision="8"
            style="width: 100%"
            placeholder="可选"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:visible="batchModalVisible"
      title="批量风控"
      :width="420"
      @ok="handleBatchSubmit"
      @cancel="batchModalVisible = false"
    >
      <a-form :model="batchForm" :label-col-props="{ span: 6 }">
        <a-form-item label="选择结果" required>
          <a-select v-model="batchForm.pre_profit_result">
            <a-option :value="0">不干预</a-option>
            <a-option :value="1">赢</a-option>
            <a-option :value="2">输</a-option>
          </a-select>
        </a-form-item>
        <a-alert type="warning" show-icon>
          将对选中的订单批量应用风控设置。
        </a-alert>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconSettings,
    IconEye,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const riskModalVisible = ref(false);
  const batchModalVisible = ref(false);
  const detailModalVisible = ref(false);
  const tableData = ref<any[]>([]);
  const currencyList = ref<any[]>([]);
  const selectedRowKeys = ref<number[]>([]);
  const detailData = ref<any>({});

  const searchForm = reactive({
    account_number: '',
    currency_id: undefined as number | undefined,
    status: undefined as number | undefined,
    result: undefined as string | undefined,
  });

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const riskForm = reactive({
    id: '' as any,
    pre_profit_result: 0,
    end_price: undefined as number | undefined,
  });

  const batchForm = reactive({
    pre_profit_result: 0,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '用户', slotName: 'user', width: 160 },
    { title: '币种', slotName: 'symbol', width: 140 },
    { title: '方向', slotName: 'type', width: 100 },
    { title: '秒数', dataIndex: 'seconds', width: 100 },
    { title: '金额', dataIndex: 'number', width: 120 },
    { title: '赔率', dataIndex: 'profit_ratio', width: 100 },
    { title: '开仓价', dataIndex: 'open_price', width: 120 },
    { title: '结算价', dataIndex: 'end_price', width: 120 },
    { title: '盈亏', slotName: 'profits', width: 120 },
    { title: '结果', slotName: 'result', width: 100 },
    { title: '状态', slotName: 'status', width: 100 },
    { title: '下单时间', slotName: 'created_at', width: 180 },
    { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
  ];

  const rowSelection = reactive({
    selectedRowKeys: [] as number[],
    onChange: (keys: number[]) => {
      selectedRowKeys.value = keys;
      rowSelection.selectedRowKeys = keys;
    },
  });

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(8);
  };

  const formatDateTime = (value?: string | number) => {
    if (!value) return '-';
    const date = dayjs(value);
    return date.isValid() ? date.format('YYYY-MM-DD HH:mm:ss') : String(value);
  };

  const statusText = (status?: number) => {
    switch (status) {
      case 0:
        return '待结算';
      case 1:
        return '已结算';
      case 2:
        return '已取消';
      default:
        return '未知';
    }
  };

  const statusColor = (status?: number) => {
    switch (status) {
      case 0:
        return 'orange';
      case 1:
        return 'green';
      case 2:
        return 'red';
      default:
        return 'gray';
    }
  };

  const resultText = (result?: number) => {
    switch (result) {
      case 0:
        return '平';
      case 1:
        return '赢';
      case 2:
        return '输';
      default:
        return '-';
    }
  };

  const resultColor = (result?: number) => {
    switch (result) {
      case 1:
        return 'green';
      case 2:
        return 'red';
      default:
        return 'gray';
    }
  };

  const fetchCurrencyList = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 200,
      });
      if (response.data && response.data.data) {
        currencyList.value = response.data.data.list || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种失败');
    }
  };

  const fetchMicroList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/micro/order/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        currency_id: searchForm.currency_id,
        status: searchForm.status,
        result: searchForm.result,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取交割合约订单失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchMicroList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.currency_id = undefined;
    searchForm.status = undefined;
    searchForm.result = undefined;
    pagination.current = 1;
    fetchMicroList();
  };

  const handleRefresh = () => {
    fetchMicroList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchMicroList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchMicroList();
  };

  const handleRisk = (record: any) => {
    Object.assign(riskForm, {
      id: record.id,
      pre_profit_result: record.pre_profit_result ?? 0,
      end_price: record.end_price || undefined,
    });
    riskModalVisible.value = true;
  };

  const handleDetail = (record: any) => {
    detailData.value = { ...record };
    detailModalVisible.value = true;
  };

  const handleRiskSubmit = async () => {
    try {
      await axios.put(`/admin/micro/order/${riskForm.id}`, {
        pre_profit_result: riskForm.pre_profit_result,
        end_price: riskForm.end_price,
      });
      Message.success('风控设置成功');
      riskModalVisible.value = false;
      fetchMicroList();
    } catch (error: any) {
      Message.error(error.message || '设置失败');
    }
  };

  const handleBatchRisk = () => {
    batchForm.pre_profit_result = 0;
    batchModalVisible.value = true;
  };

  const handleBatchSubmit = async () => {
    if (selectedRowKeys.value.length === 0) {
      Message.warning('请选择订单');
      return;
    }
    try {
      await axios.post('/admin/micro/order/batch-risk', {
        order_ids: selectedRowKeys.value,
        pre_profit_result: batchForm.pre_profit_result,
      });
      Message.success('批量设置成功');
      batchModalVisible.value = false;
      fetchMicroList();
    } catch (error: any) {
      Message.error(error.message || '批量设置失败');
    }
  };

  onMounted(() => {
    fetchCurrencyList();
    fetchMicroList();
  });
</script>

<style scoped lang="less">
  .micro-list-container {
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

    .amount-plus {
      color: #00b42a;
    }

    .amount-minus {
      color: #f53f3f;
    }
  }
</style>
