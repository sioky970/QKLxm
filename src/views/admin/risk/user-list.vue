<template>
  <div class="risk-user-container">
    <a-card class="table-card" :bordered="false">
      <template #title>
        <div class="table-header">
          <span>用户风控列表</span>
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
        <template #risk="{ record }">
          <a-select
            v-model="record.risk"
            size="small"
            style="width: 120px"
            @change="handleSetRisk(record)"
          >
            <a-option :value="0">正常</a-option>
            <a-option :value="1">重点关注</a-option>
            <a-option :value="2">高风险</a-option>
          </a-select>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { IconSync } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const loading = ref(false);
  const tableData = ref<any[]>([]);

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const columns = [
    { title: '用户ID', dataIndex: 'id', width: 120 },
    { title: '账号', dataIndex: 'account', width: 200 },
    { title: '风控等级', slotName: 'risk', width: 160 },
  ];

  const fetchRiskUsers = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/risk/users', {
        page: pagination.current,
        page_size: pagination.pageSize,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取用户风控列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSetRisk = async (record: any) => {
    try {
      await axios.post('/admin/risk/users/set', {
        user_id: record.id,
        risk: record.risk,
      });
      Message.success('风控等级已更新');
    } catch (error: any) {
      Message.error(error.message || '更新失败');
    }
  };

  const handleRefresh = () => {
    fetchRiskUsers();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchRiskUsers();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchRiskUsers();
  };

  onMounted(() => {
    fetchRiskUsers();
  });
</script>

<style scoped lang="less">
  .risk-user-container {
    padding: 20px;

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
  }
</style>
