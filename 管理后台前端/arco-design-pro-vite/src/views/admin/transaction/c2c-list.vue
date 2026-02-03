<template>
  <div class="c2c-list-container">
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
          <span>C2C交易订单</span>
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
          <a-tag :color="statusColor(record.is_sure)">
            {{ statusText(record.is_sure) }}
          </a-tag>
        </template>
        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>
        <template #actions="{ record }">
          <a-space>
            <a-button type="text" size="small" @click="handleBack(record)">
              <template #icon><icon-undo /></template>
              撤回发布
            </a-button>
            <a-button
              type="text"
              size="small"
              status="danger"
              @click="handleDelete(record)"
            >
              <template #icon><icon-delete /></template>
              删除发布
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconUndo,
    IconDelete,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const tableData = ref<any[]>([]);

  const searchForm = reactive({
    account_number: '',
    seller_number: '',
    type: undefined as string | undefined,
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
    { title: '状态', slotName: 'status', width: 120 },
    { title: '创建时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const statusText = (status?: number) => {
    switch (status) {
      case 0:
        return '未完成';
      case 1:
        return '已完成';
      case 2:
        return '已取消';
      case 3:
        return '已付款';
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
      case 3:
        return 'arcoblue';
      default:
        return 'gray';
    }
  };

  const fetchC2cList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/transaction/c2c/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        seller_number: searchForm.seller_number || undefined,
        type: searchForm.type,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取C2C订单失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchC2cList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.seller_number = '';
    searchForm.type = undefined;
    pagination.current = 1;
    fetchC2cList();
  };

  const handleRefresh = () => {
    fetchC2cList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchC2cList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchC2cList();
  };

  const handleBack = (record: any) => {
    Modal.confirm({
      title: '确认撤回',
      content: `确定撤回发布 ${record.legal_deal_send_id} 吗？`,
      onOk: async () => {
        try {
          await axios.post('/admin/transaction/c2c/back', {
            send_id: record.legal_deal_send_id,
          });
          Message.success('撤回成功');
          fetchC2cList();
        } catch (error: any) {
          Message.error(error.message || '撤回失败');
        }
      },
    });
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除发布 ${record.legal_deal_send_id} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(
            `/admin/transaction/c2c/send/${record.legal_deal_send_id}`
          );
          Message.success('删除成功');
          fetchC2cList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  fetchC2cList();
</script>

<style scoped lang="less">
  .c2c-list-container {
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
