<template>
  <div class="charge-list-container">
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
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 140px"
            allow-clear
          >
            <a-option :value="1">待审核</a-option>
            <a-option :value="2">已通过</a-option>
            <a-option :value="3">已拒绝</a-option>
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
          <span>充值记录</span>
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
        <template #account="{ record }">
          <a-space direction="vertical" :size="2">
            <span style="font-weight: 500">{{
              record.user_account || '-'
            }}</span>
            <a-tag size="small">UID: {{ record.uid }}</a-tag>
          </a-space>
        </template>

        <template #currency="{ record }">
          <span>{{
            record.currency_name || getCurrencyName(record.currency_id)
          }}</span>
        </template>

        <template #amount="{ record }">
          {{ formatBalance(record.amount) }}
        </template>

        <template #address="{ record }">
          <a-tooltip :content="record.to_address">
            <span style="font-family: monospace">
              {{ formatAddress(record.to_address) }}
            </span>
          </a-tooltip>
        </template>

        <template #image="{ record }">
          <a-image
            v-if="record.image"
            :src="record.image"
            width="40"
            height="40"
            :preview="true"
            alt="凭证"
          />
          <span v-else>-</span>
        </template>

        <template #status="{ record }">
          <a-tag v-if="record.status === 1" color="orange">待审核</a-tag>
          <a-tag v-else-if="record.status === 2" color="green">已通过</a-tag>
          <a-tag v-else-if="record.status === 3" color="red">已拒绝</a-tag>
        </template>

        <template #created_at="{ record }">
          {{ formatTime(record.created_at) }}
        </template>

        <template #actions="{ record }">
          <a-space>
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

    <a-modal
      v-model:visible="rejectModalVisible"
      title="拒绝充值"
      :width="500"
      @ok="handleRejectSubmit"
      @cancel="rejectModalVisible = false"
    >
      <a-form :model="rejectForm" :label-col-props="{ span: 6 }">
        <a-form-item label="充值金额">
          <a-input
            :model-value="`${rejectForm.amount} ${rejectForm.currency_name}`"
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
  const rejectModalVisible = ref(false);
  const currencyList = ref<any[]>([]);
  const currencyMap = reactive<Record<number, string>>({});

  const searchForm = reactive({
    account_number: '',
    currency_id: undefined as number | undefined,
    status: undefined as number | undefined,
  });

  const tableData = ref<any[]>([]);

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const rejectForm = reactive({
    id: 0,
    amount: 0,
    currency_name: '',
    reason: '',
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '用户账号', slotName: 'account', width: 180 },
    { title: '币种', slotName: 'currency', width: 120 },
    { title: '充值金额', dataIndex: 'amount', slotName: 'amount', width: 140 },
    { title: '充值地址', slotName: 'address', width: 220 },
    { title: '凭证', slotName: 'image', width: 80 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '创建时间', slotName: 'created_at', width: 180 },
    {
      title: '备注',
      dataIndex: 'remark',
      width: 200,
      ellipsis: true,
      tooltip: true,
    },
    { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
  ];

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(8);
  };

  const formatTime = (value?: string | number) => {
    if (!value) return '-';
    if (typeof value === 'number' || /^\d+$/.test(String(value))) {
      return dayjs.unix(Number(value)).format('YYYY-MM-DD HH:mm:ss');
    }
    return String(value);
  };

  const formatAddress = (address?: string) => {
    if (!address) return '-';
    if (address.length <= 16) return address;
    return `${address.substring(0, 8)}...${address.substring(
      address.length - 8
    )}`;
  };

  const getCurrencyName = (currencyId?: number) => {
    if (!currencyId) return '-';
    return currencyMap[currencyId] || `币种 ${currencyId}`;
  };

  const fetchCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 100,
      });
      if (response.data && response.data.data) {
        currencyList.value = response.data.data.list || [];
        currencyList.value.forEach((item) => {
          if (item?.id) {
            currencyMap[item.id] = item.name || `币种 ${item.id}`;
          }
        });
      }
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    }
  };

  const fetchChargeList = async () => {
    try {
      loading.value = true;
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        currency_id: searchForm.currency_id,
        status: searchForm.status,
      };

      const response = await axios.post('/admin/wallet/charge/list', params);
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取充值列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchChargeList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.currency_id = undefined;
    searchForm.status = undefined;
    pagination.current = 1;
    fetchChargeList();
  };

  const handleRefresh = () => {
    fetchChargeList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchChargeList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchChargeList();
  };

  const handleApprove = (record: any) => {
    Modal.confirm({
      title: '确认通过',
      content: `确认通过用户 ${
        record.user_account || record.uid
      } 的充值申请吗？`,
      onOk: async () => {
        try {
          await axios.post('/admin/wallet/charge/approve', { id: record.id });
          Message.success('审核通过成功');
          fetchChargeList();
        } catch (error: any) {
          Message.error(error.message || '审核失败');
        }
      },
    });
  };

  const handleReject = (record: any) => {
    Object.assign(rejectForm, {
      id: record.id,
      amount: record.amount,
      currency_name:
        record.currency_name || getCurrencyName(record.currency_id),
      reason: '',
    });
    rejectModalVisible.value = true;
  };

  const handleRejectSubmit = async () => {
    if (!rejectForm.reason) {
      Message.warning('请输入拒绝原因');
      return;
    }
    try {
      await axios.post('/admin/wallet/charge/reject', {
        id: rejectForm.id,
        reason: rejectForm.reason,
      });
      Message.success('已拒绝充值申请');
      rejectModalVisible.value = false;
      fetchChargeList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    }
  };

  onMounted(() => {
    fetchCurrencies();
    fetchChargeList();
  });
</script>

<style scoped lang="less">
  .charge-list-container {
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
