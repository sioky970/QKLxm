<template>
  <div class="lever-list-container">
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
            style="width: 140px"
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
        <a-form-item label="法币">
          <a-select
            v-model="searchForm.legal_id"
            placeholder="全部法币"
            style="width: 140px"
            allow-clear
          >
            <a-option
              v-for="currency in legalList"
              :key="currency.id"
              :value="currency.id"
            >
              {{ currency.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="方向">
          <a-select
            v-model="searchForm.type"
            placeholder="全部"
            style="width: 120px"
            allow-clear
          >
            <a-option value="buy">做多</a-option>
            <a-option value="sell">做空</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 120px"
            allow-clear
          >
            <a-option :value="0">持仓中</a-option>
            <a-option :value="1">已平仓</a-option>
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
          <span>永续合约订单</span>
          <a-space>
            <a-button type="outline" @click="handleExport">
              <template #icon><icon-download /></template>
              导出
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
          <span>{{ record.currency_name }}/{{ record.legal_name }}</span>
        </template>

        <template #type="{ record }">
          <a-tag :color="record.type === 1 ? 'green' : 'red'">
            {{ record.type === 1 ? '做多' : '做空' }}
          </a-tag>
        </template>

        <template #status="{ record }">
          <a-tag :color="record.status === 0 ? 'orange' : 'green'">
            {{ record.status === 0 ? '持仓中' : '已平仓' }}
          </a-tag>
        </template>

        <template #profits="{ record }">
          <span
            :class="record.fact_profits >= 0 ? 'amount-plus' : 'amount-minus'"
          >
            {{ formatBalance(record.fact_profits) }}
          </span>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
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
              status="danger"
              @click="handleClose(record)"
            >
              <template #icon><icon-stop /></template>
              强制平仓
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
          <a-tag :color="detailData.status === 0 ? 'orange' : 'green'">
            {{ detailData.status === 0 ? '持仓中' : '已平仓' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="用户账号">{{
          detailData.account_number
        }}</a-descriptions-item>
        <a-descriptions-item label="用户ID">{{
          detailData.user_id
        }}</a-descriptions-item>
        <a-descriptions-item label="交易对"
          >{{ detailData.currency_name }}/{{
            detailData.legal_name
          }}</a-descriptions-item
        >
        <a-descriptions-item label="方向">
          <a-tag :color="detailData.type === 1 ? 'green' : 'red'">
            {{ detailData.type === 1 ? '做多' : '做空' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="杆杆倍数"
          >{{ detailData.multiple }}x</a-descriptions-item
        >
        <a-descriptions-item label="手数">{{
          detailData.share
        }}</a-descriptions-item>
        <a-descriptions-item label="开仓价">
          <span style="color: var(--color-text-2)">{{
            formatBalance(detailData.price)
          }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="当前价">
          <span style="color: var(--color-primary-6)">{{
            formatBalance(detailData.current_price || detailData.update_price)
          }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="保证金">
          <span>{{ formatBalance(detailData.caution_money) }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="盈亏">
          <span
            :class="
              detailData.fact_profits >= 0 ? 'amount-plus' : 'amount-minus'
            "
          >
            {{ formatBalance(detailData.fact_profits) }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="开仓时间">
          {{ formatTime(detailData.create_time) }}
        </a-descriptions-item>
        <a-descriptions-item label="平仓时间">
          {{
            detailData.complete_time
              ? formatTime(detailData.complete_time)
              : '-'
          }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>

    <a-modal
      v-model:visible="closeModalVisible"
      title="强制平仓"
      :width="420"
      @ok="handleCloseSubmit"
      @cancel="closeModalVisible = false"
    >
      <a-form :model="closeForm" :label-col-props="{ span: 6 }">
        <a-form-item label="订单ID">
          <a-input :model-value="String(closeForm.id)" disabled />
        </a-form-item>
        <a-form-item label="平仓价格">
          <a-input-number
            v-model="closeForm.close_price"
            :min="0"
            :precision="8"
            style="width: 100%"
            placeholder="为空则使用当前价"
          />
        </a-form-item>
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
    IconStop,
    IconEye,
    IconDownload,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const closeModalVisible = ref(false);
  const detailModalVisible = ref(false);
  const tableData = ref<any[]>([]);
  const currencyList = ref<any[]>([]);
  const legalList = ref<any[]>([]);
  const detailData = ref<any>({});

  const searchForm = reactive({
    account_number: '',
    currency_id: undefined as number | undefined,
    legal_id: undefined as number | undefined,
    status: undefined as number | undefined,
    type: undefined as string | undefined,
  });

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const closeForm = reactive({
    id: '' as any,
    close_price: undefined as number | undefined,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '用户', slotName: 'user', width: 160 },
    { title: '交易对', slotName: 'symbol', width: 140 },
    { title: '方向', slotName: 'type', width: 100 },
    { title: '倍数', dataIndex: 'multiple', width: 80 },
    { title: '手数', dataIndex: 'share', width: 80 },
    { title: '开仓价', dataIndex: 'price', width: 120 },
    { title: '当前价', dataIndex: 'current_price', width: 120 },
    { title: '保证金', dataIndex: 'caution_money', width: 120 },
    { title: '盈亏', slotName: 'profits', width: 120 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '开仓时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 160, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(8);
  };

  const fetchCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 200,
      });
      if (response.data && response.data.data) {
        const list = response.data.data.list || [];
        currencyList.value = list.filter(
          (item: any) => item.is_change === 1 || item.is_lever === 1
        );
        legalList.value = list.filter((item: any) => item.is_legal === 1);
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    }
  };

  const fetchLeverList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/lever/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        currency_id: searchForm.currency_id,
        legal_id: searchForm.legal_id,
        status: searchForm.status,
        type: searchForm.type,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取永续合约失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchLeverList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.currency_id = undefined;
    searchForm.legal_id = undefined;
    searchForm.status = undefined;
    searchForm.type = undefined;
    pagination.current = 1;
    fetchLeverList();
  };

  const handleRefresh = () => {
    fetchLeverList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchLeverList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchLeverList();
  };

  const handleClose = (record: any) => {
    closeForm.id = record.id;
    closeForm.close_price = undefined;
    closeModalVisible.value = true;
  };

  const handleDetail = (record: any) => {
    detailData.value = { ...record };
    detailModalVisible.value = true;
  };

  const handleExport = async () => {
    try {
      Message.loading('正在导出...');
      const response = await axios.get('/admin/lever/export', {
        responseType: 'blob',
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute(
        'download',
        `lever_orders_${dayjs().format('YYYYMMDD_HHmmss')}.csv`
      );
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      Message.success('导出成功');
    } catch (error: any) {
      Message.error(error.message || '导出失败');
    }
  };

  const handleCloseSubmit = async () => {
    try {
      await axios.post('/admin/lever/close', {
        id: closeForm.id,
        close_price: closeForm.close_price,
      });
      Message.success('平仓成功');
      closeModalVisible.value = false;
      fetchLeverList();
    } catch (error: any) {
      Message.error(error.message || '平仓失败');
    }
  };

  onMounted(() => {
    fetchCurrencies();
    fetchLeverList();
  });
</script>

<style scoped lang="less">
  .lever-list-container {
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
