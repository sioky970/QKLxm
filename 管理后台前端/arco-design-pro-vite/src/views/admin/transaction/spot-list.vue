<template>
  <div class="spot-list-container">
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
        <a-form-item label="订单状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 160px"
            allow-clear
          >
            <a-option :value="0">待成交</a-option>
            <a-option :value="1">部分成交</a-option>
            <a-option :value="2">已成交</a-option>
            <a-option :value="3">已撤销</a-option>
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
          <span>币币交易列表</span>
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
        :scroll="{ x: 1400 }"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #user="{ record }">
          <a-space direction="vertical" :size="2">
            <span>{{
              record.account_number || `用户${record.from_user_id}`
            }}</span>
            <a-tag size="small" color="arcoblue"
              >ID: {{ record.from_user_id }}</a-tag
            >
          </a-space>
        </template>

        <template #symbol="{ record }">
          <span>{{ getSymbol(record.currency, record.legal) }}</span>
        </template>

        <template #side="{ record }">
          <a-tag :color="record.type === 1 ? 'green' : 'red'">
            {{ record.type === 1 ? '买入' : '卖出' }}
          </a-tag>
        </template>

        <template #status="{ record }">
          <a-tag :color="statusColor(record.status)">
            {{ statusText(record.status) }}
          </a-tag>
        </template>

        <template #time="{ record }">
          {{ formatTime(record.time) }}
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
              @click="handleCancel(record)"
            >
              <template #icon><icon-delete /></template>
              撤销
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 订单详情弹窗 -->
    <a-modal
      v-model:visible="detailModalVisible"
      title="订单详情"
      :width="600"
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
        <a-descriptions-item label="卖方用户ID">{{
          detailData.from_user_id
        }}</a-descriptions-item>
        <a-descriptions-item label="买方用户ID">{{
          detailData.to_user_id || '-'
        }}</a-descriptions-item>
        <a-descriptions-item label="交易对">{{
          getSymbol(detailData.currency, detailData.legal)
        }}</a-descriptions-item>
        <a-descriptions-item label="方向">
          <a-tag :color="detailData.type === 1 ? 'green' : 'red'">
            {{ detailData.type === 1 ? '买入' : '卖出' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="委托价">{{
          detailData.price
        }}</a-descriptions-item>
        <a-descriptions-item label="委托数量">{{
          detailData.number
        }}</a-descriptions-item>
        <a-descriptions-item label="成交量">{{
          detailData.deal
        }}</a-descriptions-item>
        <a-descriptions-item label="手续费">{{
          detailData.fee
        }}</a-descriptions-item>
        <a-descriptions-item label="下单时间" :span="2">
          {{ formatTime(detailData.time) }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconDelete,
    IconEye,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const detailModalVisible = ref(false);
  const tableData = ref<any[]>([]);
  const currencyMap = reactive<Record<number, string>>({});
  const detailData = ref<any>({});

  const searchForm = reactive({
    account_number: '',
    status: undefined as number | undefined,
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
    { title: '用户', slotName: 'user', width: 160 },
    { title: '交易对', slotName: 'symbol', width: 140 },
    { title: '方向', slotName: 'side', width: 100 },
    { title: '委托价', dataIndex: 'price', width: 120 },
    { title: '数量', dataIndex: 'number', width: 120 },
    { title: '成交量', dataIndex: 'deal', width: 120 },
    { title: '手续费', dataIndex: 'fee', width: 120 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '时间', slotName: 'time', width: 180 },
    { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const statusText = (status?: number) => {
    switch (status) {
      case 0:
        return '待成交';
      case 1:
        return '部分成交';
      case 2:
        return '已成交';
      case 3:
        return '已撤销';
      default:
        return '未知';
    }
  };

  const statusColor = (status?: number) => {
    switch (status) {
      case 0:
        return 'orange';
      case 1:
        return 'arcoblue';
      case 2:
        return 'green';
      case 3:
        return 'red';
      default:
        return 'gray';
    }
  };

  const getSymbol = (currencyId?: number, legalId?: number) => {
    const currency = currencyId ? currencyMap[currencyId] : '';
    const legal = legalId ? currencyMap[legalId] : '';
    if (currency && legal) {
      return `${currency}/${legal}`;
    }
    return `${currencyId || '-'} / ${legalId || '-'}`;
  };

  const fetchCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 300,
      });
      if (response.data && response.data.data) {
        const list = response.data.data.list || [];
        list.forEach((item: any) => {
          if (item?.id) {
            currencyMap[item.id] = item.name || `币种${item.id}`;
          }
        });
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    }
  };

  const fetchSpotList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/transaction/spot/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        status: searchForm.status,
      });

      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取现货交易失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchSpotList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.status = undefined;
    pagination.current = 1;
    fetchSpotList();
  };

  const handleRefresh = () => {
    fetchSpotList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchSpotList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchSpotList();
  };

  const handleDetail = (record: any) => {
    detailData.value = { ...record };
    detailModalVisible.value = true;
  };

  const handleCancel = (record: any) => {
    Modal.confirm({
      title: '确认撤销',
      content: `确定撤销订单 ${record.id} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.post('/admin/transaction/cancel', { id: record.id });
          Message.success('撤销成功');
          fetchSpotList();
        } catch (error: any) {
          Message.error(error.message || '撤销失败');
        }
      },
    });
  };

  onMounted(() => {
    fetchCurrencies();
    fetchSpotList();
  });
</script>

<style scoped lang="less">
  .spot-list-container {
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
