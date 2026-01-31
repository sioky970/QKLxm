<template>
  <div class="user-detail-container">
    <a-card class="header-card" :bordered="false">
      <template #title>
        <a-space>
          <a-avatar v-if="userDetail?.head_portrait" :size="36">
            <img :src="userDetail.head_portrait" alt="avatar" />
          </a-avatar>
          <div class="header-title">
            <div class="title">用户详情</div>
            <div class="subtitle">{{ userDetail?.account_number || '-' }}</div>
          </div>
        </a-space>
      </template>
      <template #extra>
        <a-space>
          <a-button @click="goBack">返回列表</a-button>
          <a-button type="primary" @click="handleRefresh">
            <template #icon><icon-sync /></template>
            刷新
          </a-button>
        </a-space>
      </template>

      <a-spin :loading="loading">
        <a-descriptions :column="3" size="small" bordered>
          <a-descriptions-item label="用户ID">
            {{ userDetail?.id ?? '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="账号">
            {{ userDetail?.account_number || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="用户类型">
            <a-tag :color="userType.color">{{ userType.label }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="userStatus.color">{{ userStatus.label }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="手机号">
            {{ userDetail?.phone || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="邮箱">
            {{ userDetail?.email || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="昵称">
            {{ userDetail?.nickname || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="注册时间">
            {{ formatTime(userDetail?.time) }}
          </a-descriptions-item>
          <a-descriptions-item label="最后登录">
            {{ formatTime(userDetail?.last_time) }}
          </a-descriptions-item>
          <a-descriptions-item label="最后登录IP">
            {{ userDetail?.last_login_ip || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="邀请码">
            {{ userDetail?.extension_code || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="上级ID">
            {{ userDetail?.parent_id ?? '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="风控等级">
            {{ userDetail?.risk ?? '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="用户等级">
            {{ userDetail?.level ?? '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="黑名单">
            <a-tag :color="userDetail?.is_blacklist === 1 ? 'red' : 'green'">
              {{ userDetail?.is_blacklist === 1 ? '是' : '否' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="备注" :span="3">
            {{ userDetail?.user_remark || '-' }}
          </a-descriptions-item>
        </a-descriptions>
      </a-spin>
    </a-card>

    <a-card class="detail-card" :bordered="false" style="margin-top: 16px">
      <a-tabs type="rounded">
        <a-tab-pane key="kyc" title="实名认证信息">
          <a-descriptions :column="3" size="small" bordered>
            <a-descriptions-item label="实名状态">
              <a-tag :color="realStatus.color">{{ realStatus.label }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="认证标记">
              {{ userDetail?.is_auth || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="认证时间">
              {{ formatTime(userDetail?.new_isreal_time) }}
            </a-descriptions-item>
            <a-descriptions-item label="实名次数">
              {{ userDetail?.zhitui_real_number ?? '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="实名团队数">
              {{ userDetail?.real_teamnumber ?? '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="当日实名团队数">
              {{ userDetail?.today_real_teamnumber ?? '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </a-tab-pane>

        <a-tab-pane key="wallets" title="钱包余额">
          <a-table
            :columns="walletColumns"
            :data="walletData"
            :loading="walletLoading"
            :pagination="false"
            :scroll="{ x: 900 }"
            row-key="currency_id"
          >
            <template #balance="{ record }">
              <span :class="record.balance > 0 ? 'amount-plus' : ''">
                {{ formatBalance(record.balance) }}
              </span>
            </template>
            <template #locked="{ record }">
              <span :class="record.locked > 0 ? 'amount-locked' : ''">
                {{ formatBalance(record.locked) }}
              </span>
            </template>
            <template #total="{ record }">
              <span class="amount-total">
                {{ formatBalance(record.total) }}
              </span>
            </template>
            <template #usdt_value="{ record }">
              <span class="amount-usdt">
                {{ formatBalance(record.usdt_value) }} USDT
              </span>
            </template>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="logs" title="账户流水">
          <a-table
            :columns="accountLogColumns"
            :data="accountLogs"
            :loading="accountLogLoading"
            :pagination="accountLogPagination"
            :scroll="{ x: 1200 }"
            row-key="id"
            @page-change="handleAccountLogPageChange"
            @page-size-change="handleAccountLogPageSizeChange"
          >
            <template #type="{ record }">
              {{ accountLogTypeMap[record.type] || `类型${record.type}` }}
            </template>
            <template #value="{ record }">
              <span :class="record.value >= 0 ? 'amount-plus' : 'amount-minus'">
                {{ formatBalance(record.value) }}
              </span>
            </template>
            <template #created_time="{ record }">
              {{ formatTime(record.created_time) }}
            </template>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="login" title="登录日志">
          <a-empty description="登录日志接口待接入" />
        </a-tab-pane>

        <a-tab-pane key="operation" title="操作日志">
          <a-empty description="操作日志接口待接入" />
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { IconSync } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  interface UserDetail {
    id: number;
    account_number?: string;
    phone?: string;
    email?: string;
    nickname?: string;
    head_portrait?: string;
    status?: number;
    type?: number;
    account_type?: number;
    extension_code?: string;
    parent_id?: number;
    risk?: number;
    level?: number;
    is_blacklist?: number;
    is_real?: number;
    is_realname?: number;
    is_auth?: string;
    new_isreal_time?: number;
    zhitui_real_number?: number;
    real_teamnumber?: number;
    today_real_teamnumber?: number;
    user_remark?: string;
    time?: number;
    last_time?: number;
    last_login_ip?: string;
  }

  interface UserWallet {
    currency_id: number;
    currency_name: string;
    balance: number;
    locked: number;
    total: number;
    usdt_value: number;
  }

  interface AccountLog {
    id: number;
    type: number;
    currency: number;
    value: number;
    info?: string;
    created_time?: number;
  }

  const route = useRoute();
  const router = useRouter();

  const loading = ref(false);
  const walletLoading = ref(false);
  const accountLogLoading = ref(false);
  const userDetail = ref<UserDetail | null>(null);
  const currencyMap = reactive<Record<number, string>>({});
  const walletDataRaw = ref<UserWallet[]>([]);
  const walletData = computed(() => walletDataRaw.value);
  const accountLogs = ref<AccountLog[]>([]);
  const accountLogTypeMap = reactive<Record<number, string>>({});

  const accountLogPagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const userId = computed(() => Number(route.params.id));
  const isValidUserId = computed(
    () => !Number.isNaN(userId.value) && userId.value > 0
  );

  const userStatus = computed(() => {
    const status = userDetail.value?.status;
    const isActive = status === 1 || status === undefined;
    return {
      label: isActive ? '正常' : '冻结',
      color: isActive ? 'green' : 'red',
    };
  });

  const userType = computed(() => {
    const raw = userDetail.value?.account_type ?? userDetail.value?.type;
    if (raw === 2) {
      return { label: '代理用户', color: 'orange' };
    }
    if (raw === 1 || raw === 0) {
      return { label: '普通用户', color: 'blue' };
    }
    return { label: '未知', color: 'gray' };
  });

  const realStatus = computed(() => {
    const raw = userDetail.value?.is_real ?? userDetail.value?.is_realname;
    const isVerified = raw === 1 || raw === 2;
    return {
      label: isVerified ? '已认证' : '未认证',
      color: isVerified ? 'green' : 'red',
    };
  });

  const walletColumns = [
    { 
      title: '币种', 
      dataIndex: 'currency_name',
      width: 120,
      fixed: 'left'
    },
    {
      title: '币种ID',
      dataIndex: 'currency_id',
      width: 100,
    },
    {
      title: '可用余额',
      dataIndex: 'balance',
      slotName: 'balance',
      width: 160,
    },
    {
      title: '冻结余额',
      dataIndex: 'locked',
      slotName: 'locked',
      width: 160,
    },
    {
      title: '总余额',
      dataIndex: 'total',
      slotName: 'total',
      width: 160,
    },
    {
      title: 'USDT估值',
      dataIndex: 'usdt_value',
      slotName: 'usdt_value',
      width: 160,
    },
  ];

  const accountLogColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '类型', dataIndex: 'type', slotName: 'type', width: 200 },
    { title: '币种ID', dataIndex: 'currency', width: 90 },
    { title: '变动金额', dataIndex: 'value', slotName: 'value', width: 140 },
    {
      title: '备注',
      dataIndex: 'info',
      ellipsis: true,
      tooltip: true,
      width: 240,
    },
    {
      title: '时间',
      dataIndex: 'created_time',
      slotName: 'created_time',
      width: 180,
    },
  ];

  const unwrapResponseData = (response: any) => {
    if (response?.data?.data !== undefined) {
      return response.data.data;
    }
    if (response?.data !== undefined) {
      return response.data;
    }
    return response;
  };

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const formatBalance = (value?: number) => {
    if (value === undefined || value === null) return '0.00000000';
    return Number(value).toFixed(8);
  };

  const fetchUserDetail = async () => {
    if (!isValidUserId.value) return;
    loading.value = true;
    try {
      const response = await axios.get(`/admin/user/${userId.value}`);
      userDetail.value = unwrapResponseData(response);
    } catch (error: any) {
      Message.error(error.message || '获取用户详情失败');
    } finally {
      loading.value = false;
    }
  };

  const fetchWallets = async () => {
    if (!isValidUserId.value) return;
    walletLoading.value = true;
    try {
      const response = await axios.get(`/admin/user/${userId.value}/wallets`);
      const payload = unwrapResponseData(response);
      walletDataRaw.value = Array.isArray(payload)
        ? payload
        : payload?.list || [];
    } catch (error: any) {
      Message.error(error.message || '获取钱包信息失败');
    } finally {
      walletLoading.value = false;
    }
  };

  const fetchAccountLogTypes = async () => {
    try {
      const response = await axios.get('/admin/account/log/types');
      const payload = unwrapResponseData(response);
      if (Array.isArray(payload)) {
        payload.forEach((item) => {
          if (item?.type !== undefined) {
            accountLogTypeMap[item.type] = item.name || `类型${item.type}`;
          }
        });
      }
    } catch (error: any) {
      Message.error(error.message || '获取流水类型失败');
    }
  };

  const fetchCurrencyMap = async () => {
    try {
      const response = await axios.post('/admin/currency/list', {
        page: 1,
        page_size: 500,
        is_display: 1,
      });
      const payload = unwrapResponseData(response);
      const list = payload?.list || [];
      list.forEach((item: any) => {
        if (item?.id) {
          currencyMap[item.id] = item.name || `币种${item.id}`;
        }
      });
    } catch (error: any) {
      Message.error(error.message || '获取币种列表失败');
    }
  };

  const fetchAccountLogs = async () => {
    if (!isValidUserId.value) return;
    accountLogLoading.value = true;
    try {
      const response = await axios.post(
        `/admin/account/user/${userId.value}/logs`,
        {
          page: accountLogPagination.current,
          page_size: accountLogPagination.pageSize,
        }
      );
      const payload = unwrapResponseData(response);
      accountLogs.value = payload?.list || [];
      accountLogPagination.total = payload?.total || 0;
    } catch (error: any) {
      Message.error(error.message || '获取账户流水失败');
    } finally {
      accountLogLoading.value = false;
    }
  };

  const handleAccountLogPageChange = (page: number) => {
    accountLogPagination.current = page;
    fetchAccountLogs();
  };

  const handleAccountLogPageSizeChange = (pageSize: number) => {
    accountLogPagination.pageSize = pageSize;
    accountLogPagination.current = 1;
    fetchAccountLogs();
  };

  const handleRefresh = async () => {
    await Promise.all([fetchUserDetail(), fetchWallets(), fetchAccountLogs()]);
  };

  const goBack = () => {
    router.push('/admin/user/list');
  };

  watch(
    () => route.params.id,
    async () => {
      if (!isValidUserId.value) {
        Message.error('用户ID无效');
        return;
      }
      accountLogPagination.current = 1;
      accountLogPagination.total = 0;
      accountLogs.value = [];
      walletDataRaw.value = [];
      userDetail.value = null;
      await Promise.all([
        fetchUserDetail(),
        fetchWallets(),
        fetchAccountLogs(),
        fetchAccountLogTypes(),
        fetchCurrencyMap(),
      ]);
    },
    { immediate: true }
  );
</script>

<style scoped lang="less">
  .user-detail-container {
    padding: 20px;

    .header-card,
    .detail-card {
      :deep(.arco-card-body) {
        padding: 20px;
      }
    }

    .header-title {
      display: flex;
      flex-direction: column;
      gap: 2px;

      .title {
        font-size: 16px;
        font-weight: 600;
        line-height: 1.2;
      }

      .subtitle {
        font-size: 12px;
        color: var(--color-text-3);
      }
    }

    :deep(.arco-tabs-content) {
      padding-top: 16px;
    }

    .amount-plus {
      color: #00b42a;
      font-weight: 500;
    }

    .amount-locked {
      color: #ff7d00;
      font-weight: 500;
    }

    .amount-total {
      color: #165dff;
      font-weight: 600;
    }

    .amount-usdt {
      color: #722ed1;
      font-weight: 500;
    }

    .amount-minus {
      color: #f53f3f;
    }
  }
</style>
