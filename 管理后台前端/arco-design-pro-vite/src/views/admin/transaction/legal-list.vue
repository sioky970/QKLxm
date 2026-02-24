<template>
  <div class="legal-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="买家账号">
          <a-input
            v-model="searchForm.account_number"
            placeholder="请输入买家账号"
            style="width: 180px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="卖家账号">
          <a-input
            v-model="searchForm.seller_number"
            placeholder="请输入卖家账号"
            style="width: 180px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="类型">
          <a-select
            v-model="searchForm.type"
            placeholder="全部"
            style="width: 120px"
            allow-clear
          >
            <a-option value="buy">买入</a-option>
            <a-option value="sell">卖出</a-option>
          </a-select>
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
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 140px"
            allow-clear
          >
            <a-option :value="0">待付款</a-option>
            <a-option :value="1">已付款</a-option>
            <a-option :value="2">已完成</a-option>
            <a-option :value="3">已取消</a-option>
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
          <span>法币交易订单</span>
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
        <template #buyer="{ record }">
          <span>{{ record.buyer_account || record.user_id }}</span>
        </template>
        <template #seller="{ record }">
          <span>{{ record.seller_account || record.seller_id }}</span>
        </template>
        <template #type="{ record }">
          <a-tag :color="record.type === 'buy' ? 'green' : 'red'">
            {{ record.type === 'buy' ? '买入' : '卖出' }}
          </a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusColor(record.status)">
            {{ statusText(record.status) }}
          </a-tag>
        </template>
        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>
        <template #actions="{ record }">
          <a-space>
            <a-button
              v-if="record.status === 0"
              type="text"
              size="small"
              status="success"
              @click="handleConfirmPay(record)"
            >
              <template #icon><icon-check /></template>
              确认付款
            </a-button>
            <a-button
              v-if="record.status === 1"
              type="text"
              size="small"
              status="success"
              @click="handleConfirmReceive(record)"
            >
              <template #icon><icon-check /></template>
              确认收款
            </a-button>
            <a-button
              v-if="record.status === 0 || record.status === 1"
              type="text"
              size="small"
              status="danger"
              @click="handleCancel(record)"
            >
              <template #icon><icon-close /></template>
              取消
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconCheck,
    IconClose,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const tableData = ref<any[]>([]);
  const currencyList = ref<any[]>([]);

  const searchForm = reactive({
    account_number: '',
    seller_number: '',
    type: undefined as string | undefined,
    status: undefined as number | undefined,
    currency_id: undefined as number | undefined,
  });

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '买家', slotName: 'buyer', width: 140 },
    { title: '卖家', slotName: 'seller', width: 140 },
    { title: '类型', slotName: 'type', width: 100 },
    { title: '币种', dataIndex: 'currency_name', width: 120 },
    { title: '数量', dataIndex: 'number', width: 120 },
    { title: '单价', dataIndex: 'price', width: 120 },
    { title: '总价', dataIndex: 'total_price', width: 120 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '创建时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 220, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const statusText = (status?: number) => {
    switch (status) {
      case 0:
        return '待付款';
      case 1:
        return '已付款';
      case 2:
        return '已完成';
      case 3:
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
        return 'arcoblue';
      case 2:
        return 'green';
      case 3:
        return 'red';
      default:
        return 'gray';
    }
  };

  const fetchCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 100,
      });
      if (response.data && response.data.data) {
        currencyList.value = response.data.data.list || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    }
  };

  const fetchLegalList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/transaction/legal/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        seller_number: searchForm.seller_number || undefined,
        type: searchForm.type,
        status: searchForm.status,
        currency_id: searchForm.currency_id,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取法币交易失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchLegalList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.seller_number = '';
    searchForm.type = undefined;
    searchForm.status = undefined;
    searchForm.currency_id = undefined;
    pagination.current = 1;
    fetchLegalList();
  };

  const handleRefresh = () => {
    fetchLegalList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchLegalList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchLegalList();
  };

  const handleConfirmPay = (record: any) => {
    Modal.confirm({
      title: '确认付款',
      content: `确定已付款订单 ${record.id} 吗？`,
      onOk: async () => {
        try {
          await axios.post('/admin/transaction/legal/confirm-pay', {
            deal_id: record.id,
          });
          Message.success('确认成功');
          fetchLegalList();
        } catch (error: any) {
          Message.error(error.message || '操作失败');
        }
      },
    });
  };

  const handleConfirmReceive = (record: any) => {
    Modal.confirm({
      title: '确认收款',
      content: `确定完成订单 ${record.id} 吗？`,
      onOk: async () => {
        try {
          await axios.post('/admin/transaction/legal/confirm-receive', {
            deal_id: record.id,
          });
          Message.success('交易完成');
          fetchLegalList();
        } catch (error: any) {
          Message.error(error.message || '操作失败');
        }
      },
    });
  };

  const handleCancel = (record: any) => {
    Modal.confirm({
      title: '确认取消',
      content: `确定取消订单 ${record.id} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.post('/admin/transaction/legal/cancel', {
            deal_id: record.id,
          });
          Message.success('取消成功');
          fetchLegalList();
        } catch (error: any) {
          Message.error(error.message || '取消失败');
        }
      },
    });
  };

  onMounted(() => {
    fetchCurrencies();
    fetchLegalList();
  });
</script>

<style scoped lang="less">
  .legal-list-container {
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
