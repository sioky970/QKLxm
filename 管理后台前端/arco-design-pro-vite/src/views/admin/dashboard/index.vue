<template>
  <div class="dashboard-container">
    <a-spin :loading="loading" style="width: 100%">
      <!-- 核心数据卡片 -->
      <a-row :gutter="16" style="margin-bottom: 16px">
        <a-col :span="6">
          <a-card :bordered="false" hoverable>
            <a-statistic
              title="平台总用户"
              :value="dashboardData.total_users"
              :precision="0"
              :value-style="{ color: '#165dff' }"
            >
              <template #prefix>
                <icon-user />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card :bordered="false" hoverable>
            <a-statistic
              title="今日新增用户"
              :value="dashboardData.today_new_users"
              :precision="0"
              :value-style="{ color: '#00b42a' }"
            >
              <template #prefix>
                <icon-user-add />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card :bordered="false" hoverable>
            <a-statistic
              title="今日交易量"
              :value="dashboardData.today_trade_amount"
              :precision="2"
              suffix="USDT"
              :value-style="{ color: '#ff7d00' }"
            >
              <template #prefix>
                <icon-swap />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card :bordered="false" hoverable>
            <a-statistic
              title="今日订单数"
              :value="dashboardData.today_trades"
              :precision="0"
              :value-style="{ color: '#722ed1' }"
            >
              <template #prefix>
                <icon-file />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>

      <!-- 待处理事项卡片 -->
      <a-row :gutter="16" style="margin-bottom: 16px">
        <a-col :span="6">
          <a-card :bordered="false">
            <a-statistic
              title="待审核提现"
              :value="dashboardData.pending_withdraw"
              :precision="0"
            >
              <template #prefix>
                <icon-export style="color: #f53f3f" />
              </template>
              <template #suffix>
                <a-button
                  type="text"
                  size="mini"
                  @click="goToPage('/admin/wallet/withdrawal')"
                >
                  去处理 →
                </a-button>
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card :bordered="false">
            <a-statistic
              title="待审核实名"
              :value="dashboardData.pending_kyc"
              :precision="0"
            >
              <template #prefix>
                <icon-idcard style="color: #f77234" />
              </template>
              <template #suffix>
                <a-button
                  type="text"
                  size="mini"
                  @click="goToPage('/admin/system/kyc')"
                >
                  去处理 →
                </a-button>
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card :bordered="false">
            <a-statistic
              title="总充值金额"
              :value="dashboardData.total_deposit"
              :precision="2"
              suffix="USDT"
            >
              <template #prefix>
                <icon-import style="color: #00b42a" />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card :bordered="false">
            <a-statistic
              title="总提现金额"
              :value="dashboardData.total_withdraw"
              :precision="2"
              suffix="USDT"
            >
              <template #prefix>
                <icon-export style="color: #f53f3f" />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>

      <!-- 详细统计 -->
      <a-row :gutter="16">
        <a-col :span="8">
          <a-card title="用户统计" :bordered="false">
            <a-descriptions
              :data="userStatsDescriptions"
              :column="1"
              size="large"
            />
          </a-card>
        </a-col>
        <a-col :span="8">
          <a-card title="交易统计" :bordered="false">
            <a-descriptions
              :data="tradeStatsDescriptions"
              :column="1"
              size="large"
            />
          </a-card>
        </a-col>
        <a-col :span="8">
          <a-card title="快速入口" :bordered="false">
            <a-space direction="vertical" :size="12" style="width: 100%">
              <a-button
                type="outline"
                long
                @click="goToPage('/admin/user/list')"
              >
                <template #icon><icon-user /></template>
                用户管理
              </a-button>
              <a-button
                type="outline"
                long
                @click="goToPage('/admin/currency/list')"
              >
                <template #icon><icon-apps /></template>
                币种管理
              </a-button>
              <a-button
                type="outline"
                long
                @click="goToPage('/admin/wallet/withdrawal')"
              >
                <template #icon><icon-export /></template>
                提现审核
              </a-button>
              <a-button
                type="outline"
                long
                @click="goToPage('/admin/transaction/lever')"
              >
                <template #icon><icon-swap /></template>
                永续合约
              </a-button>
            </a-space>
          </a-card>
        </a-col>
      </a-row>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import {
    IconUser,
    IconUserAdd,
    IconSwap,
    IconFile,
    IconApps,
    IconExport,
    IconImport,
    IconIdcard,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const router = useRouter();
  const loading = ref(false);

  // 仪表盘数据
  const dashboardData = ref({
    total_users: 0,
    today_new_users: 0,
    total_deposit: 0,
    total_withdraw: 0,
    pending_withdraw: 0,
    pending_kyc: 0,
    today_trades: 0,
    today_trade_amount: 0,
  });

  // 用户统计
  const userStats = ref({
    total: 0,
    active: 0,
    frozen: 0,
    verified: 0,
    unverified: 0,
  });

  // 交易统计
  const tradeStats = ref({
    spot_total: 0,
    spot_amount: 0,
    lever_total: 0,
    lever_amount: 0,
    micro_total: 0,
    micro_amount: 0,
  });

  // 用户统计描述列表
  const userStatsDescriptions = computed(() => [
    {
      label: '总用户数',
      value: userStats.value.total.toLocaleString(),
    },
    {
      label: '活跃用户',
      value: userStats.value.active.toLocaleString(),
    },
    {
      label: '冻结用户',
      value: userStats.value.frozen.toLocaleString(),
    },
    {
      label: '已认证',
      value: userStats.value.verified.toLocaleString(),
    },
    {
      label: '未认证',
      value: userStats.value.unverified.toLocaleString(),
    },
  ]);

  // 交易统计描述列表
  const tradeStatsDescriptions = computed(() => [
    {
      label: '现货交易',
      value: `${tradeStats.value.spot_total.toLocaleString()} 笔`,
    },
    {
      label: '永续合约',
      value: `${tradeStats.value.lever_total.toLocaleString()} 笔`,
    },
    {
      label: '交割合约',
      value: `${tradeStats.value.micro_total.toLocaleString()} 笔`,
    },
    {
      label: '总交易量',
      value: `${(
        (tradeStats.value.spot_amount +
          tradeStats.value.lever_amount +
          tradeStats.value.micro_amount) /
        10000
      ).toFixed(2)} 万`,
    },
  ]);

  // 获取仪表盘数据
  const fetchDashboardData = async () => {
    try {
      loading.value = true;
      const response = await axios.get('/admin/statistics/dashboard');
      if (response.data && response.data.data) {
        dashboardData.value = response.data.data;
      }
    } catch (error: any) {
      Message.error(error.message || '获取仪表盘数据失败');
    } finally {
      loading.value = false;
    }
  };

  // 获取用户统计
  const fetchUserStats = async () => {
    try {
      const response = await axios.get('/admin/statistics/user');
      if (response.data && response.data.data) {
        userStats.value = response.data.data;
      }
    } catch (error: any) {
      console.error('获取用户统计失败:', error);
    }
  };

  // 获取交易统计
  const fetchTradeStats = async () => {
    try {
      const response = await axios.get('/admin/statistics/trade');
      if (response.data && response.data.data) {
        tradeStats.value = response.data.data;
      }
    } catch (error: any) {
      console.error('获取交易统计失败:', error);
    }
  };

  // 刷新所有数据
  const refreshData = async () => {
    await Promise.all([
      fetchDashboardData(),
      fetchUserStats(),
      fetchTradeStats(),
    ]);
  };

  // 页面跳转
  const goToPage = (path: string) => {
    router.push(path);
  };

  // 页面加载时获取数据
  onMounted(() => {
    refreshData();

    // 每30秒刷新一次数据
    setInterval(() => {
      refreshData();
    }, 30000);
  });
</script>

<style scoped lang="less">
  .dashboard-container {
    padding: 20px;
    background: var(--color-fill-1);
    min-height: calc(100vh - 60px);

    :deep(.arco-card) {
      margin-bottom: 0;

      .arco-card-body {
        padding: 20px;
      }
    }

    :deep(.arco-statistic) {
      .arco-statistic-title {
        font-size: 14px;
        color: var(--color-text-2);
        margin-bottom: 8px;
      }

      .arco-statistic-content {
        .arco-statistic-value {
          font-size: 28px;
          font-weight: 600;
        }
      }
    }

    :deep(.arco-descriptions-item-label) {
      font-weight: 500;
    }
  }
</style>
