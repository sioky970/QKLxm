<template>
  <div class="hazard-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="法币">
          <a-select
            v-model="searchForm.legal_id"
            placeholder="全部法币"
            style="width: 160px"
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
        <a-form-item label="风险阈值">
          <a-input-number
            v-model="searchForm.threshold"
            :min="0"
            :step="0.05"
            :precision="2"
            style="width: 140px"
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
          <span>爆仓监控</span>
          <a-space>
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
        :scroll="{ x: 1600 }"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #user="{ record }">
          <a-space direction="vertical" :size="2">
            <span style="font-weight: 500">
              {{ record.account_number || `UID ${record.user_id}` }}
            </span>
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

        <template #caution="{ record }">
          {{ formatBalance(record.caution_money) }}
        </template>

        <template #risk_rate="{ record }">
          <a-tag :color="getRiskColor(record)">
            {{ formatRiskRate(record) }}
          </a-tag>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #actions="{ record }">
          <a-space>
            <a-button type="text" size="small" @click="handleIntervene(record)">
              <template #icon><icon-edit /></template>
              干预价格
            </a-button>
            <a-button
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

    <a-modal
      v-model:visible="interveneVisible"
      title="干预价格"
      :width="420"
      @ok="handleInterveneSubmit"
      @cancel="interveneVisible = false"
    >
      <a-form :model="interveneForm" :label-col-props="{ span: 6 }">
        <a-form-item label="订单ID">
          <a-input :model-value="String(interveneForm.trade_id)" disabled />
        </a-form-item>
        <a-form-item label="干预价格" required>
          <a-input-number
            v-model="interveneForm.update_price"
            :min="0"
            :precision="8"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="写入行情">
          <a-switch
            v-model="interveneForm.write_market"
            :checked-value="1"
            :unchecked-value="0"
          />
        </a-form-item>
      </a-form>
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
    IconEdit,
    IconStop,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const tableData = ref<any[]>([]);
  const legalList = ref<any[]>([]);
  const interveneVisible = ref(false);
  const closeModalVisible = ref(false);

  const searchForm = reactive({
    legal_id: undefined as number | undefined,
    threshold: 0.5,
  });

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const interveneForm = reactive({
    trade_id: '' as any,
    update_price: 0,
    write_market: 0,
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
    {
      title: '保证金',
      dataIndex: 'caution_money',
      slotName: 'caution',
      width: 120,
    },
    { title: '开仓价', dataIndex: 'price', width: 120 },
    { title: '当前价', dataIndex: 'current_price', width: 120 },
    { title: '风险率', slotName: 'risk_rate', width: 120 },
    { title: '开仓时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(8);
  };

  const getRiskRate = (record: any) => {
    if (record.risk_rate && Number(record.risk_rate) > 0) {
      return Number(record.risk_rate);
    }
    const caution = Number(record.caution_money || 0);
    const amount =
      Number(record.number || 0) * Number(record.current_price || 0);
    if (!caution || !amount) return undefined;
    return caution / amount;
  };

  const formatRiskRate = (record: any) => {
    const rate = getRiskRate(record);
    if (rate === undefined || Number.isNaN(rate)) return '-';
    return `${(rate * 100).toFixed(2)}%`;
  };

  const getRiskColor = (record: any) => {
    const rate = getRiskRate(record);
    if (rate === undefined) return 'gray';
    if (rate <= searchForm.threshold) return 'red';
    if (rate <= 1) return 'orange';
    return 'green';
  };

  const fetchLegalCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 200,
      });
      if (response.data && response.data.data) {
        const list = response.data.data.list || [];
        legalList.value = list.filter((item: any) => item.is_legal === 1);
      }
    } catch (error: any) {
      Message.error(error.message || '获取法币列表失败');
    }
  };

  const fetchHazardList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/lever/hazard/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        legal_id: searchForm.legal_id,
        threshold: searchForm.threshold,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取爆仓监控列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchHazardList();
  };

  const handleReset = () => {
    searchForm.legal_id = undefined;
    searchForm.threshold = 0.5;
    pagination.current = 1;
    fetchHazardList();
  };

  const handleRefresh = () => {
    fetchHazardList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchHazardList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchHazardList();
  };

  const handleIntervene = (record: any) => {
    interveneForm.trade_id = record.id;
    interveneForm.update_price = record.current_price || record.price || 0;
    interveneForm.write_market = 0;
    interveneVisible.value = true;
  };

  const handleInterveneSubmit = async () => {
    if (!interveneForm.update_price) {
      Message.warning('请输入干预价格');
      return;
    }
    try {
      await axios.post('/admin/lever/hazard/handle', {
        trade_id: interveneForm.trade_id,
        update_price: interveneForm.update_price,
        write_market: interveneForm.write_market,
      });
      Message.success('干预成功');
      interveneVisible.value = false;
      fetchHazardList();
    } catch (error: any) {
      Message.error(error.message || '干预失败');
    }
  };

  const handleClose = (record: any) => {
    closeForm.id = record.id;
    closeForm.close_price = undefined;
    closeModalVisible.value = true;
  };

  const handleCloseSubmit = async () => {
    try {
      await axios.post('/admin/lever/close', {
        id: closeForm.id,
        close_price: closeForm.close_price,
      });
      Message.success('平仓成功');
      closeModalVisible.value = false;
      fetchHazardList();
    } catch (error: any) {
      Message.error(error.message || '平仓失败');
    }
  };

  onMounted(() => {
    fetchLegalCurrencies();
    fetchHazardList();
  });
</script>

<style scoped lang="less">
  .hazard-list-container {
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
  }
</style>
