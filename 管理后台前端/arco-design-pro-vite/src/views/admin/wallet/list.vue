<template>
  <div class="wallet-list-container">
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
        <a-form-item label="地址">
          <a-input
            v-model="searchForm.address"
            placeholder="钱包地址"
            style="width: 260px"
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
            <a-option :value="1">正常</a-option>
            <a-option :value="0">冻结</a-option>
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

    <div class="summary-row">
      <a-row :gutter="16">
        <a-col :span="8">
          <a-card class="summary-card" :bordered="false">
            <div class="summary-title">法币总额</div>
            <div class="summary-value">{{
              formatBalance(totals.legal_balance)
            }}</div>
          </a-card>
        </a-col>
        <a-col :span="8">
          <a-card class="summary-card" :bordered="false">
            <div class="summary-title">交易币总额</div>
            <div class="summary-value">{{
              formatBalance(totals.change_balance)
            }}</div>
          </a-card>
        </a-col>
        <a-col :span="8">
          <a-card class="summary-card" :bordered="false">
            <div class="summary-title">杠杆币总额</div>
            <div class="summary-value">{{
              formatBalance(totals.lever_balance)
            }}</div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <a-card class="table-card" :bordered="false" style="margin-top: 16px">
      <template #title>
        <div class="table-header">
          <span>钱包列表</span>
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
        :scroll="{ x: 2200 }"
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

        <template #currency="{ record }">
          <span>{{
            record.currency_name || getCurrencyName(record.currency)
          }}</span>
        </template>

        <template #address="{ record }">
          <a-tooltip :content="record.address">
            <span style="font-family: monospace">
              {{ formatAddress(record.address) }}
            </span>
          </a-tooltip>
        </template>

        <template #legal_balance="{ record }">
          {{ formatBalance(record.legal_balance) }}
        </template>
        <template #change_balance="{ record }">
          {{ formatBalance(record.change_balance) }}
        </template>
        <template #lever_balance="{ record }">
          {{ formatBalance(record.lever_balance) }}
        </template>
        <template #micro_balance="{ record }">
          {{ formatBalance(record.micro_balance) }}
        </template>
        <template #lock_balance="{ record }">
          {{ formatBalance(record.lock_balance) }}
        </template>
        <template #lock_change="{ record }">
          {{ formatBalance(record.lock_change) }}
        </template>
        <template #lock_lever="{ record }">
          {{ formatBalance(record.lock_lever) }}
        </template>

        <template #status="{ record }">
          <a-tag :color="record.status === 1 ? 'green' : 'red'">
            {{ record.status === 1 ? '正常' : '冻结' }}
          </a-tag>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #actions="{ record }">
          <a-space>
            <a-button
              type="text"
              size="small"
              @click="handleViewDetail(record)"
            >
              <template #icon><icon-eye /></template>
              详情
            </a-button>
            <a-button
              type="text"
              size="small"
              @click="handleAdjustBalance(record)"
            >
              <template #icon><icon-swap /></template>
              调整余额
            </a-button>
            <a-button
              type="text"
              size="small"
              :status="record.status === 1 ? 'danger' : 'success'"
              @click="handleToggleStatus(record)"
            >
              <template #icon>
                <icon-lock v-if="record.status === 1" />
                <icon-unlock v-else />
              </template>
              {{ record.status === 1 ? '冻结' : '激活' }}
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="detailModalVisible"
      title="钱包详情"
      :width="560"
      :footer="false"
    >
      <a-descriptions v-if="currentRecord" :column="1" bordered size="large">
        <a-descriptions-item label="钱包ID">{{
          currentRecord.id
        }}</a-descriptions-item>
        <a-descriptions-item label="用户账号">
          {{ currentRecord.account_number || `UID ${currentRecord.user_id}` }}
        </a-descriptions-item>
        <a-descriptions-item label="币种">
          {{
            currentRecord.currency_name ||
            getCurrencyName(currentRecord.currency)
          }}
        </a-descriptions-item>
        <a-descriptions-item label="地址">
          {{ currentRecord.address || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="法币余额">
          {{ formatBalance(currentRecord.legal_balance) }}
        </a-descriptions-item>
        <a-descriptions-item label="币币余额">
          {{ formatBalance(currentRecord.change_balance) }}
        </a-descriptions-item>
        <a-descriptions-item label="合约余额">
          {{ formatBalance(currentRecord.lever_balance) }}
        </a-descriptions-item>
        <a-descriptions-item label="秒合约余额">
          {{ formatBalance(currentRecord.micro_balance) }}
        </a-descriptions-item>
        <a-descriptions-item label="锁定法币">
          {{ formatBalance(currentRecord.lock_balance) }}
        </a-descriptions-item>
        <a-descriptions-item label="锁定币币">
          {{ formatBalance(currentRecord.lock_change) }}
        </a-descriptions-item>
        <a-descriptions-item label="锁定合约">
          {{ formatBalance(currentRecord.lock_lever) }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          {{ currentRecord.status === 1 ? '正常' : '冻结' }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(currentRecord.create_time) }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>

    <a-modal
      v-model:visible="adjustModalVisible"
      title="调整余额"
      :width="520"
      @ok="handleAdjustSubmit"
      @cancel="adjustModalVisible = false"
    >
      <a-form :model="adjustForm" :label-col-props="{ span: 6 }">
        <a-form-item label="钱包ID">
          <a-input :model-value="String(adjustForm.id)" disabled />
        </a-form-item>
        <a-form-item label="钱包类型" required>
          <a-select
            v-model="adjustForm.wallet_type"
            placeholder="选择钱包类型"
          >
            <a-option value="spot">现货钱包</a-option>
            <a-option value="contract">合约钱包</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="余额类型" required>
          <a-select
            v-model="adjustForm.balance_type"
            placeholder="选择余额类型"
          >
            <a-option value="legal">法币余额</a-option>
            <a-option value="change">币币余额</a-option>
            <a-option value="lever">合约余额</a-option>
            <a-option value="micro">秒合约余额</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="调整金额" required>
          <a-input-number
            v-model="adjustForm.amount"
            :min="-999999999"
            :precision="8"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="备注">
          <a-textarea
            v-model="adjustForm.reason"
            placeholder="调整原因（可选）"
          />
        </a-form-item>
        <a-alert type="warning" show-icon>
          正数为增加余额，负数为减少余额，请谨慎操作。
        </a-alert>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconEye,
    IconSwap,
    IconLock,
    IconUnlock,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const loading = ref(false);
  const detailModalVisible = ref(false);
  const adjustModalVisible = ref(false);
  const currentRecord = ref<any>(null);
  const currencyList = ref<any[]>([]);
  const currencyMap = reactive<Record<number, string>>({});
  const route = useRoute();

  // 搜索表单
  const searchForm = reactive({
    account_number: '',
    currency_id: undefined as number | undefined,
    address: '',
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

  const totals = reactive({
    legal_balance: 0,
    change_balance: 0,
    lever_balance: 0,
  });

  // 调整余额表单
  const adjustForm = reactive({
    id: 0,
    balance_type: 'legal',
    wallet_type: 'spot', // spot: 现货钱包, contract: 合约钱包
    amount: 0,
    reason: '',
  });

  // 表格列配置
  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70, fixed: 'left' },
    { title: '用户', slotName: 'user', width: 160, fixed: 'left' },
    { title: '币种', slotName: 'currency', width: 120 },
    { title: '地址', slotName: 'address', width: 220 },
    {
      title: '法币',
      children: [
        {
          title: '余额',
          dataIndex: 'legal_balance',
          slotName: 'legal_balance',
          width: 140,
        },
        {
          title: '冻结',
          dataIndex: 'lock_balance',
          slotName: 'lock_balance',
          width: 140,
        },
      ],
    },
    {
      title: '币币',
      children: [
        {
          title: '余额',
          dataIndex: 'change_balance',
          slotName: 'change_balance',
          width: 140,
        },
        {
          title: '冻结',
          dataIndex: 'lock_change',
          slotName: 'lock_change',
          width: 140,
        },
      ],
    },
    {
      title: '杠杆',
      children: [
        {
          title: '余额',
          dataIndex: 'lever_balance',
          slotName: 'lever_balance',
          width: 140,
        },
        {
          title: '冻结',
          dataIndex: 'lock_lever',
          slotName: 'lock_lever',
          width: 140,
        },
      ],
    },
    {
      title: '秒合约余额',
      dataIndex: 'micro_balance',
      slotName: 'micro_balance',
      width: 140,
    },
    { title: '状态', slotName: 'status', width: 100 },
    { title: '创建时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 220, fixed: 'right' },
  ];

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(8);
  };

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
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

  // 获取币种列表
  const fetchCurrencies = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 200,
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

  // 获取钱包列表
  const fetchWalletList = async () => {
    try {
      loading.value = true;
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        currency_id: searchForm.currency_id,
        address: searchForm.address || undefined,
        status: searchForm.status,
      };
      const response = await axios.post('/admin/wallet/list', params);
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
        const extraTotal = response.data.data.extra_data?.total;
        if (extraTotal) {
          totals.legal_balance = Number(extraTotal.legal_balance || 0);
          totals.change_balance = Number(extraTotal.change_balance || 0);
          totals.lever_balance = Number(extraTotal.lever_balance || 0);
        } else {
          totals.legal_balance = 0;
          totals.change_balance = 0;
          totals.lever_balance = 0;
        }
      }
    } catch (error: any) {
      Message.error(error.message || '获取钱包列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchWalletList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.currency_id = undefined;
    searchForm.address = '';
    searchForm.status = undefined;
    pagination.current = 1;
    fetchWalletList();
  };

  const handleRefresh = () => {
    fetchWalletList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchWalletList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchWalletList();
  };

  const handleViewDetail = (record: any) => {
    currentRecord.value = record;
    detailModalVisible.value = true;
  };

  const handleAdjustBalance = (record: any) => {
    currentRecord.value = record;
    Object.assign(adjustForm, {
      id: record.id,
      balance_type: 'legal',
      wallet_type: 'spot',
      amount: 0,
      reason: '',
    });
    adjustModalVisible.value = true;
  };

  const handleAdjustSubmit = async () => {
    if (!adjustForm.amount) {
      Message.warning('请输入调整金额');
      return;
    }

    try {
      await axios.post('/admin/wallet/update-balance', {
        id: adjustForm.id,
        balance_type: adjustForm.balance_type,
        wallet_type: adjustForm.wallet_type,
        amount: adjustForm.amount,
        reason: adjustForm.reason,
      });
      Message.success('余额调整成功');
      adjustModalVisible.value = false;
      fetchWalletList();
    } catch (error: any) {
      Message.error(error.message || '余额调整失败');
    }
  };

  const handleToggleStatus = (record: any) => {
    const isActive = record.status === 1;
    const action = isActive ? '冻结' : '激活';
    Modal.confirm({
      title: `确认${action}`,
      content: `确定要${action}该钱包吗？`,
      onOk: async () => {
        try {
          await axios.post(
            `/admin/wallet/${isActive ? 'freeze' : 'activate'}`,
            {
              id: record.id,
            }
          );
          Message.success(`${action}成功`);
          fetchWalletList();
        } catch (error: any) {
          Message.error(error.message || `${action}失败`);
        }
      },
    });
  };

  onMounted(() => {
    fetchCurrencies();
    if (route.query.account_number) {
      searchForm.account_number = String(route.query.account_number);
    }
    fetchWalletList();
  });
</script>

<style scoped lang="less">
  .wallet-list-container {
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

    .summary-row {
      margin-top: 16px;
    }

    .summary-card {
      background: linear-gradient(135deg, #2d5cf6, #5e92ff);
      color: #fff;
      text-align: center;

      .summary-title {
        font-size: 14px;
        opacity: 0.9;
      }

      .summary-value {
        margin-top: 8px;
        font-size: 20px;
        font-weight: 600;
        letter-spacing: 0.5px;
      }
    }

    :deep(.arco-table) {
      margin-top: 16px;
    }
  }
</style>
