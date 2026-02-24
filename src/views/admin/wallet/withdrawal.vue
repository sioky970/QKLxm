<template>
  <div class="withdrawal-list-container">
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
            style="width: 150px"
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
        <a-form-item label="审核状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部状态"
            style="width: 150px"
            allow-clear
          >
            <a-option :value="1">待审核</a-option>
            <a-option :value="2">审核通过</a-option>
            <a-option :value="3">审核拒绝</a-option>
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
          <span>提现审核列表</span>
          <a-space>
            <a-badge :count="pendingCount" :offset="[5, 0]">
              <a-button @click="handleFilterPending">
                <template #icon><icon-clock-circle /></template>
                待审核
              </a-button>
            </a-badge>
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
          <a-space direction="vertical" :size="4">
            <span style="font-weight: 500">{{ record.account_number }}</span>
            <a-tag size="small">ID: {{ record.user_id }}</a-tag>
          </a-space>
        </template>

        <template #currency="{ record }">
          <a-space>
            <span>{{ record.currency_name }}</span>
          </a-space>
        </template>

        <template #amount="{ record }">
          <a-space direction="vertical" :size="4">
            <span style="font-weight: 500; color: #f53f3f">
              {{ record.number }} {{ record.currency_name }}
            </span>
            <span style="font-size: 12px; color: var(--color-text-3)">
              手续费: {{ record.rate }}%
            </span>
          </a-space>
        </template>

        <template #address="{ record }">
          <a-tooltip :content="record.address">
            <span style="font-family: monospace">
              {{
                record.address
                  ? `${record.address.substring(
                      0,
                      10
                    )}...${record.address.substring(
                      record.address.length - 10
                    )}`
                  : '-'
              }}
            </span>
          </a-tooltip>
        </template>

        <template #status="{ record }">
          <a-tag v-if="record.status === 1" color="orange">
            <template #icon><icon-clock-circle /></template>
            待审核
          </a-tag>
          <a-tag v-else-if="record.status === 2" color="green">
            <template #icon><icon-check-circle /></template>
            已通过
          </a-tag>
          <a-tag v-else-if="record.status === 3" color="red">
            <template #icon><icon-close-circle /></template>
            已拒绝
          </a-tag>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #actions="{ record }">
          <a-space direction="vertical" :size="8">
            <a-button
              type="text"
              size="small"
              @click="handleViewDetail(record)"
            >
              <template #icon><icon-eye /></template>
              查看详情
            </a-button>
            <a-button
              v-if="record.status === 1"
              type="text"
              size="small"
              status="success"
              @click="handleApprove(record)"
            >
              <template #icon><icon-check /></template>
              通过
            </a-button>
            <a-button
              v-if="record.status === 1"
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

    <!-- 详情弹窗 -->
    <a-modal
      v-model:visible="detailModalVisible"
      title="提现详情"
      :width="600"
      :footer="false"
    >
      <a-descriptions
        v-if="currentRecord"
        :data="detailData"
        :column="1"
        bordered
        size="large"
      >
        <template #label="{ label }">
          <span style="font-weight: 500">{{ label }}</span>
        </template>
      </a-descriptions>
    </a-modal>

    <!-- 审核通过弹窗 -->
    <a-modal
      v-model:visible="approveModalVisible"
      title="审核通过"
      :width="500"
      @ok="handleApproveSubmit"
      @cancel="approveModalVisible = false"
    >
      <a-form :model="approveForm" :label-col-props="{ span: 6 }">
        <a-form-item label="提现金额">
          <a-input
            :model-value="`${approveForm.number} ${approveForm.currency_name}`"
            disabled
          />
        </a-form-item>
        <a-form-item label="提现地址">
          <a-input :model-value="approveForm.address" disabled />
        </a-form-item>
        <a-form-item label="审核方式" required>
          <a-radio-group v-model="approveForm.method">
            <a-radio value="manual">手动转账</a-radio>
            <a-radio value="auto">自动转账</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="approveForm.method === 'manual'" label="交易哈希">
          <a-input
            v-model="approveForm.txid"
            placeholder="请输入链上交易哈希"
          />
        </a-form-item>
        <a-form-item label="备注">
          <a-textarea
            v-model="approveForm.notes"
            placeholder="审核备注（可选）"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 拒绝弹窗 -->
    <a-modal
      v-model:visible="rejectModalVisible"
      title="拒绝提现"
      :width="500"
      @ok="handleRejectSubmit"
      @cancel="rejectModalVisible = false"
    >
      <a-form :model="rejectForm" :label-col-props="{ span: 6 }">
        <a-form-item label="提现金额">
          <a-input
            :model-value="`${rejectForm.number} ${rejectForm.currency_name}`"
            disabled
          />
        </a-form-item>
        <a-form-item label="拒绝原因" required>
          <a-textarea
            v-model="rejectForm.reason"
            placeholder="请输入拒绝原因"
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconClockCircle,
    IconCheckCircle,
    IconCloseCircle,
    IconEye,
    IconCheck,
    IconClose,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const detailModalVisible = ref(false);
  const approveModalVisible = ref(false);
  const rejectModalVisible = ref(false);
  const currentRecord = ref<any>(null);
  const currencyList = ref<any[]>([]);

  // 搜索表单
  const searchForm = reactive({
    account_number: '',
    currency_id: undefined as number | undefined,
    status: undefined as number | undefined,
  });

  // 表格数据
  const tableData = ref<any[]>([]);

  // 分页配置
  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  // 审核通过表单
  const approveForm = reactive({
    id: 0,
    number: 0,
    currency_name: '',
    address: '',
    method: 'manual',
    txid: '',
    notes: '',
  });

  // 拒绝表单
  const rejectForm = reactive({
    id: 0,
    number: 0,
    currency_name: '',
    reason: '',
  });

  // 待审核数量
  const pendingCount = computed(() => {
    return tableData.value.filter((item) => item.status === 1).length;
  });

  // 详情数据
  const detailData = computed(() => {
    if (!currentRecord.value) return [];
    const record = currentRecord.value;
    return [
      { label: '提现ID', value: record.id },
      { label: '用户账号', value: record.account_number },
      { label: '用户ID', value: record.user_id },
      { label: '币种', value: record.currency_name },
      { label: '提现数量', value: `${record.number} ${record.currency_name}` },
      { label: '手续费率', value: `${record.rate}%` },
      { label: '提现地址', value: record.address || '-' },
      {
        label: '审核状态',
        value: (() => {
          if (record.status === 1) return '待审核';
          if (record.status === 2) return '已通过';
          return '已拒绝';
        })(),
      },
      { label: '申请时间', value: formatTime(record.create_time) },
      {
        label: '审核时间',
        value: record.update_time ? formatTime(record.update_time) : '-',
      },
    ];
  });

  // 表格列配置
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
    },
    {
      title: '用户账号',
      slotName: 'account',
      width: 150,
    },
    {
      title: '币种',
      slotName: 'currency',
      width: 100,
    },
    {
      title: '提现数量',
      slotName: 'amount',
      width: 180,
    },
    {
      title: '提现地址',
      slotName: 'address',
      width: 200,
    },
    {
      title: '审核状态',
      slotName: 'status',
      width: 120,
    },
    {
      title: '申请时间',
      slotName: 'create_time',
      width: 180,
    },
    {
      title: '操作',
      slotName: 'actions',
      width: 120,
      fixed: 'right',
    },
  ];

  // 格式化时间
  const formatTime = (timestamp: number) => {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  // 获取币种列表
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
      console.error('获取币种列表失败:', error);
    }
  };

  // 获取提现列表
  const fetchWithdrawalList = async () => {
    try {
      loading.value = true;
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        currency_id: searchForm.currency_id,
        status: searchForm.status,
      };

      const response = await axios.post('/admin/wallet/withdrawals', params);

      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取提现列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 搜索
  const handleSearch = () => {
    pagination.current = 1;
    fetchWithdrawalList();
  };

  // 重置
  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.currency_id = undefined;
    searchForm.status = undefined;
    pagination.current = 1;
    fetchWithdrawalList();
  };

  // 筛选待审核
  const handleFilterPending = () => {
    searchForm.status = 1;
    handleSearch();
  };

  // 刷新
  const handleRefresh = () => {
    fetchWithdrawalList();
  };

  // 翻页
  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchWithdrawalList();
  };

  // 改变每页条数
  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchWithdrawalList();
  };

  // 查看详情
  const handleViewDetail = (record: any) => {
    currentRecord.value = record;
    detailModalVisible.value = true;
  };

  // 审核通过
  const handleApprove = (record: any) => {
    Object.assign(approveForm, {
      id: record.id,
      number: record.number,
      currency_name: record.currency_name,
      address: record.address,
      method: 'manual',
      txid: '',
      notes: '',
    });
    approveModalVisible.value = true;
  };

  // 提交审核通过
  const handleApproveSubmit = async () => {
    try {
      if (approveForm.method === 'manual' && !approveForm.txid) {
        Message.warning('请输入交易哈希');
        return;
      }

      await axios.post('/admin/wallet/withdrawals/approve', {
        id: approveForm.id,
        method: approveForm.method,
        txid: approveForm.txid,
        notes: approveForm.notes,
      });

      Message.success('审核通过成功');
      approveModalVisible.value = false;
      fetchWithdrawalList();
    } catch (error: any) {
      Message.error(error.message || '审核失败');
    }
  };

  // 拒绝
  const handleReject = (record: any) => {
    Object.assign(rejectForm, {
      id: record.id,
      number: record.number,
      currency_name: record.currency_name,
      reason: '',
    });
    rejectModalVisible.value = true;
  };

  // 提交拒绝
  const handleRejectSubmit = async () => {
    try {
      if (!rejectForm.reason) {
        Message.warning('请输入拒绝原因');
        return;
      }

      await axios.post('/admin/wallet/withdrawals/reject', {
        id: rejectForm.id,
        reason: rejectForm.reason,
      });

      Message.success('已拒绝提现申请');
      rejectModalVisible.value = false;
      fetchWithdrawalList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    }
  };

  // 页面加载时获取数据
  onMounted(() => {
    fetchCurrencies();
    fetchWithdrawalList();
  });
</script>

<style scoped lang="less">
  .withdrawal-list-container {
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
