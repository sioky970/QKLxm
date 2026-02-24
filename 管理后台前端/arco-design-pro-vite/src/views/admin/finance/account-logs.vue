<template>
  <div class="account-logs-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="用户账号">
          <a-input
            v-model="searchForm.account_number"
            placeholder="账号/UID"
            style="width: 200px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="币种">
          <a-select
            v-model="searchForm.currency_id"
            placeholder="全部币种"
            style="width: 160px"
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
        <a-form-item label="类型">
          <a-select
            v-model="searchForm.type"
            placeholder="全部类型"
            style="width: 200px"
            allow-clear
          >
            <a-option
              v-for="item in logTypeList"
              :key="item.type"
              :value="item.type"
            >
              {{ item.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="时间">
          <a-range-picker
            v-model="searchForm.dateRange"
            style="width: 260px"
            value-format="YYYY-MM-DD"
          />
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
          <span>财务流水</span>
          <a-space>
            <a-button @click="handleRefresh">
              <template #icon><icon-sync /></template>
              刷新
            </a-button>
            <a-button @click="handleExport">
              <template #icon><icon-download /></template>
              导出
            </a-button>
          </a-space>
        </div>
      </template>

      <a-table
        :columns="columns"
        :data="tableData"
        :pagination="pagination"
        :loading="loading"
        :scroll="{ x: 1400 }"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #value="{ record }">
          <span :class="record.value >= 0 ? 'amount-plus' : 'amount-minus'">
            {{ formatBalance(record.value) }}
          </span>
        </template>

        <template #type="{ record }">
          {{ record.type_name || logTypeMap[record.type] || '-' }}
        </template>

        <template #created_time="{ record }">
          {{ formatTime(record.created_time) }}
        </template>

        <template #actions="{ record }">
          <a-button type="text" size="small" @click="handleView(record)">
            <template #icon><icon-eye /></template>
            详情
          </a-button>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="detailVisible"
      title="流水详情"
      :width="520"
      :footer="false"
    >
      <a-spin :loading="detailLoading">
        <a-descriptions v-if="detailData" :column="1" bordered size="large">
          <a-descriptions-item label="流水ID">
            {{ detailData.log?.id || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="用户账号">
            {{ detailData.user?.account_number || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="币种">
            {{ detailData.currency?.name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="类型">
            {{ detailData.type_name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="变动金额">
            {{ formatBalance(detailData.log?.value) }}
          </a-descriptions-item>
          <a-descriptions-item label="备注">
            {{ detailData.log?.info || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="时间">
            {{ formatTime(detailData.log?.created_time) }}
          </a-descriptions-item>
        </a-descriptions>
      </a-spin>
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
    IconDownload,
    IconEye,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';
  import { getToken } from '@/utils/auth';

  const loading = ref(false);
  const tableData = ref<any[]>([]);
  const currencyList = ref<any[]>([]);
  const logTypeList = ref<any[]>([]);
  const logTypeMap = reactive<Record<number, string>>({});

  const detailVisible = ref(false);
  const detailLoading = ref(false);
  const detailData = ref<any>(null);

  const searchForm = reactive({
    account_number: '',
    currency_id: undefined as number | undefined,
    type: undefined as number | undefined,
    dateRange: [] as string[],
  });

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户账号', dataIndex: 'account_number', width: 160 },
    { title: '币种', dataIndex: 'currency_name', width: 120 },
    { title: '类型', slotName: 'type', width: 220 },
    { title: '变动金额', slotName: 'value', width: 140 },
    { title: '备注', dataIndex: 'info', ellipsis: true, tooltip: true },
    { title: '时间', slotName: 'created_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 100, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(8);
  };

  const getDateParams = () => {
    if (searchForm.dateRange && searchForm.dateRange.length === 2) {
      const [start, end] = searchForm.dateRange;
      return { start_time: start, end_time: end };
    }
    return { start_time: undefined, end_time: undefined };
  };

  const fetchCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 200,
      });
      if (response.data && response.data.data) {
        currencyList.value = response.data.data.list || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    }
  };

  const fetchLogTypes = async () => {
    try {
      const response = await axios.get('/admin/account/log/types');
      if (response.data && response.data.data) {
        logTypeList.value = response.data.data || [];
        logTypeList.value.forEach((item: any) => {
          logTypeMap[item.type] = item.name;
        });
      }
    } catch (error: any) {
      Message.error(error.message || '获取流水类型失败');
    }
  };

  const fetchAccountLogs = async () => {
    try {
      loading.value = true;
      const dateParams = getDateParams();
      const response = await axios.post('/admin/account/log/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        currency_id: searchForm.currency_id || 0,
        type: searchForm.type,
        start_time: dateParams.start_time,
        end_time: dateParams.end_time,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取账户流水失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchAccountLogs();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.currency_id = undefined;
    searchForm.type = undefined;
    searchForm.dateRange = [];
    pagination.current = 1;
    fetchAccountLogs();
  };

  const handleRefresh = () => {
    fetchAccountLogs();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchAccountLogs();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchAccountLogs();
  };

  const handleView = async (record: any) => {
    detailVisible.value = true;
    detailLoading.value = true;
    try {
      const response = await axios.get(`/admin/account/log/${record.id}`);
      if (response.data && response.data.data) {
        detailData.value = response.data.data;
      }
    } catch (error: any) {
      Message.error(error.message || '获取详情失败');
    } finally {
      detailLoading.value = false;
    }
  };

  const handleExport = async () => {
    try {
      const dateParams = getDateParams();
      const downloadClient = axios.create({
        baseURL: import.meta.env.VITE_API_BASE_URL,
        responseType: 'blob',
        headers: {
          Authorization: `Bearer ${getToken()}`,
        },
      });

      const response = await downloadClient.get('/admin/account/log/export', {
        params: {
          account_number: searchForm.account_number || undefined,
          currency_id: searchForm.currency_id || undefined,
          type: searchForm.type || undefined,
          start_time: dateParams.start_time,
          end_time: dateParams.end_time,
        },
      });

      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      const disposition = response.headers['content-disposition'];
      let filename = 'account_logs.csv';
      if (disposition) {
        const match = disposition.match(/filename=([^;]+)/i);
        if (match && match[1]) {
          filename = decodeURIComponent(match[1].trim());
        }
      }
      link.href = url;
      link.download = filename;
      link.click();
      URL.revokeObjectURL(url);
    } catch (error: any) {
      Message.error(error.message || '导出失败');
    }
  };

  onMounted(() => {
    fetchCurrencies();
    fetchLogTypes();
    fetchAccountLogs();
  });
</script>

<style scoped lang="less">
  .account-logs-container {
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
