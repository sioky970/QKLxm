<template>
  <div class="seller-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="账号">
          <a-input
            v-model="searchForm.account_number"
            placeholder="请输入账号"
            style="width: 200px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 140px"
            allow-clear
          >
            <a-option :value="0">待审核</a-option>
            <a-option :value="1">已通过</a-option>
            <a-option :value="2">已拒绝</a-option>
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
          <span>商家列表</span>
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
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #account="{ record }">
          <a-space direction="vertical" :size="2">
            <span style="font-weight: 500">{{ record.account_number }}</span>
            <span class="text-muted">{{ record.real_name || '-' }}</span>
          </a-space>
        </template>
        <template #contact="{ record }">
          <a-space direction="vertical" :size="2">
            <span>{{ record.phone || '-' }}</span>
            <span class="text-muted">{{ record.email || '-' }}</span>
          </a-space>
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
              @click="handleApprove(record)"
            >
              <template #icon><icon-check /></template>
              通过
            </a-button>
            <a-button
              v-if="record.status === 0"
              type="text"
              size="small"
              status="danger"
              @click="handleReject(record)"
            >
              <template #icon><icon-close /></template>
              拒绝
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="rejectModalVisible"
      title="拒绝商家申请"
      :width="420"
      @ok="handleRejectSubmit"
      @cancel="rejectModalVisible = false"
    >
      <a-form :model="rejectForm" :label-col-props="{ span: 6 }">
        <a-form-item label="拒绝原因" required>
          <a-textarea v-model="rejectForm.reason" placeholder="请输入原因" />
        </a-form-item>
      </a-form>
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
    IconCheck,
    IconClose,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const tableData = ref<any[]>([]);
  const rejectModalVisible = ref(false);

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

  const rejectForm = reactive({
    id: 0,
    reason: '',
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '账号/实名', slotName: 'account', width: 180 },
    { title: '联系方式', slotName: 'contact', width: 200 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '申请时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 160 },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const statusText = (status?: number) => {
    switch (status) {
      case 0:
        return '待审核';
      case 1:
        return '已通过';
      case 2:
        return '已拒绝';
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

  const fetchSellerList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/transaction/seller/list', {
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
      Message.error(error.message || '获取商家列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchSellerList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.status = undefined;
    pagination.current = 1;
    fetchSellerList();
  };

  const handleRefresh = () => {
    fetchSellerList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchSellerList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchSellerList();
  };

  const handleApprove = (record: any) => {
    Modal.confirm({
      title: '确认通过',
      content: `确定通过商家 ${record.account_number} 吗？`,
      onOk: async () => {
        try {
          await axios.post('/admin/transaction/seller/approve', {
            id: record.id,
          });
          Message.success('审核通过');
          fetchSellerList();
        } catch (error: any) {
          Message.error(error.message || '操作失败');
        }
      },
    });
  };

  const handleReject = (record: any) => {
    rejectForm.id = record.id;
    rejectForm.reason = '';
    rejectModalVisible.value = true;
  };

  const handleRejectSubmit = async () => {
    if (!rejectForm.reason) {
      Message.warning('请输入拒绝原因');
      return;
    }
    try {
      await axios.post('/admin/transaction/seller/reject', {
        id: rejectForm.id,
        reason: rejectForm.reason,
      });
      Message.success('已拒绝');
      rejectModalVisible.value = false;
      fetchSellerList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    }
  };

  onMounted(() => {
    fetchSellerList();
  });
</script>

<style scoped lang="less">
  .seller-list-container {
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

    .text-muted {
      color: var(--color-text-3);
      font-size: 12px;
    }
  }
</style>
