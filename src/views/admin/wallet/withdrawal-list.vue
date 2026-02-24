<template>
  <div class="withdrawal-container">
    <!-- 搜索区域 -->
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="用户账号">
          <a-input
            v-model="searchForm.account_number"
            placeholder="请输入用户账号"
            style="width: 180px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部状态"
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

    <!-- 统计卡片 -->
    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="待审核" :value="stats.pending">
            <template #suffix>
              <span style="font-size: 14px">笔</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="已通过" :value="stats.approved">
            <template #suffix>
              <span style="font-size: 14px">笔</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="已拒绝" :value="stats.rejected">
            <template #suffix>
              <span style="font-size: 14px">笔</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic
            title="待审核金额"
            :value="stats.pendingAmount"
            :precision="2"
          >
            <template #suffix>
              <span style="font-size: 14px">USDT</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- 表格区域 -->
    <a-card class="table-card" :bordered="false" style="margin-top: 16px">
      <template #title>
        <div class="table-header">
          <span>提现申请列表</span>
          <a-button @click="handleRefresh">
            <template #icon><icon-sync /></template>
            刷新
          </a-button>
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
            <span class="user-account">{{ record.account_number || '-' }}</span>
            <a-tag size="small" color="arcoblue">
              UID: {{ record.user_id }}
            </a-tag>
          </a-space>
        </template>

        <template #currency="{ record }">
          <span>{{ record.currency_name || `币种${record.currency_id}` }}</span>
        </template>

        <template #network_type="{ record }">
          <a-tag color="purple">
            {{ record.network_type || 'TRC20' }}
          </a-tag>
        </template>

        <template #amount="{ record }">
          <a-space direction="vertical" :size="2">
            <span class="amount-value">{{ formatAmount(record.number) }}</span>
            <span class="amount-real">
              实际到账: {{ formatAmount(record.real_number) }}
            </span>
          </a-space>
        </template>

        <template #address="{ record }">
          <a-tooltip :content="getValidAddress(record)">
            <span class="address-text chain-address">
              {{ formatAddress(getValidAddress(record)) }}
            </span>
          </a-tooltip>
          <a-button 
            type="text" 
            size="mini" 
            @click="copyText(getValidAddress(record))"
            :disabled="!getValidAddress(record)"
          >
            <template #icon><icon-copy /></template>
          </a-button>
        </template>

        <template #status="{ record }">
          <a-tag v-if="record.status === 1" color="orange">待审核</a-tag>
          <a-tag v-else-if="record.status === 2" color="green">已通过</a-tag>
          <a-tag v-else-if="record.status === 3" color="red">已拒绝</a-tag>
        </template>

        <template #txid="{ record }">
          <template v-if="record.txid">
            <a-tooltip :content="record.txid">
              <span class="txid-text">{{ formatAddress(record.txid) }}</span>
            </a-tooltip>
            <a-button type="text" size="mini" @click="copyText(record.txid)">
              <template #icon><icon-copy /></template>
            </a-button>
          </template>
          <span v-else class="text-gray">-</span>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #actions="{ record }">
          <a-space v-if="record.status === 1">
            <a-button
              type="text"
              size="small"
              status="success"
              @click="handleApprove(record)"
            >
              <template #icon><icon-check /></template>
              通过
            </a-button>
            <a-button
              type="text"
              size="small"
              status="danger"
              @click="handleReject(record)"
            >
              <template #icon><icon-close /></template>
              拒绝
            </a-button>
          </a-space>
          <a-button
            v-else
            type="text"
            size="small"
            @click="handleViewDetail(record)"
          >
            <template #icon><icon-eye /></template>
            详情
          </a-button>
        </template>
      </a-table>
    </a-card>

    <!-- 审核通过弹窗 -->
    <a-modal
      v-model:visible="approveModalVisible"
      title="审核通过 - 区块链提现"
      :width="700"
      :ok-loading="submitLoading"
      @ok="handleApproveSubmit"
      @cancel="approveModalVisible = false"
    >
      <a-alert type="info" style="margin-bottom: 16px">
        <template #icon><icon-send /></template>
        区块链提现 - {{ approveForm.network_type || 'TRC20' }}，请核对钱包地址后进行打款
      </a-alert>

      <!-- 区块链地址信息 -->
      <a-descriptions
        title="区块链收款地址"
        :column="1"
        bordered
        style="margin-bottom: 16px"
      >
        <a-descriptions-item label="网络类型">
          <a-tag color="purple">{{ approveForm.network_type || 'TRC20' }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="钱包地址">
          <span style="font-family: monospace; word-break: break-all; color: #165dff">
            {{ getValidAddress(approveForm) || '-' }}
          </span>
          <a-button 
            v-if="getValidAddress(approveForm)" 
            type="text" 
            size="mini" 
            @click="copyText(getValidAddress(approveForm))"
          >
            <template #icon><icon-copy /></template>
          </a-button>
        </a-descriptions-item>
      </a-descriptions>

      <a-divider />

      <a-form :model="approveForm" :label-col-props="{ span: 6 }">
        <a-form-item label="提现金额">
          <a-input
            :model-value="`${approveForm.number} ${approveForm.currency_name}`"
            disabled
          />
        </a-form-item>
        <a-form-item label="区块链地址">
          <a-input :model-value="getValidAddress(approveForm)" disabled />
        </a-form-item>
        <a-form-item label="处理方式">
          <a-radio-group v-model="approveForm.method">
            <a-radio value="manual">手动转账</a-radio>
            <a-radio value="auto">自动转账</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="approveForm.method === 'manual'" label="交易哈希">
          <a-input
            v-model="approveForm.txid"
            placeholder="请输入链上交易哈希（选填）"
          />
        </a-form-item>
        <a-form-item label="备注">
          <a-textarea
            v-model="approveForm.notes"
            placeholder="请输入备注信息（选填）"
            :max-length="200"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 拒绝弹窗 -->
    <a-modal
      v-model:visible="rejectModalVisible"
      title="拒绝提现"
      :width="500"
      :ok-loading="submitLoading"
      ok-text="确认拒绝"
      @ok="handleRejectSubmit"
      @cancel="rejectModalVisible = false"
    >
      <a-alert type="warning" style="margin-bottom: 16px">
        拒绝后，提现金额将退还至用户账户余额
      </a-alert>
      <a-form :model="rejectForm" :label-col-props="{ span: 6 }">
        <a-form-item label="提现金额">
          <a-input
            :model-value="`${rejectForm.number} ${rejectForm.currency_name}`"
            disabled
          />
        </a-form-item>
        <a-form-item label="用户账号">
          <a-input :model-value="rejectForm.account_number" disabled />
        </a-form-item>
        <a-form-item label="拒绝原因" required>
          <a-textarea
            v-model="rejectForm.reason"
            placeholder="请输入拒绝原因"
            :max-length="200"
            show-word-limit
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 详情弹窗 -->
    <a-modal
      v-model:visible="detailModalVisible"
      title="区块链提现详情"
      :width="700"
      :footer="false"
    >
      <a-alert type="info" style="margin-bottom: 16px">
        <template #icon><icon-send /></template>
        区块链提现 - {{ (detailData as any).network_type || 'TRC20' }}
      </a-alert>

      <!-- 区块链提现信息 -->
      <a-descriptions
        title="区块链收款地址"
        :column="1"
        bordered
        style="margin-bottom: 16px"
      >
        <a-descriptions-item label="网络类型">
          <a-tag color="purple">{{ (detailData as any).network_type || 'TRC20' }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="钱包地址">
          <span style="font-family: monospace; word-break: break-all">
            {{ getValidAddress(detailData as any) || '-' }}
          </span>
          <a-button 
            v-if="getValidAddress(detailData as any)" 
            type="text" 
            size="mini" 
            @click="copyText(getValidAddress(detailData as any))"
          >
            <template #icon><icon-copy /></template>
          </a-button>
        </a-descriptions-item>
      </a-descriptions>

      <a-divider />

      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="订单ID">
          {{ detailData.id }}
        </a-descriptions-item>
        <a-descriptions-item label="用户账号">
          {{ detailData.account_number || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="币种">
          {{ detailData.currency_name || `币种${detailData.currency_id}` }}
        </a-descriptions-item>
        <a-descriptions-item label="网络类型">
          <a-tag color="purple">{{ (detailData as any).network_type || 'TRC20' }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="提现金额">
          {{ formatAmount(detailData.number) }}
        </a-descriptions-item>
        <a-descriptions-item label="实际到账">
          {{ formatAmount(detailData.real_number) }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag v-if="detailData.status === 1" color="orange">待审核</a-tag>
          <a-tag v-else-if="detailData.status === 2" color="green">
            已通过
          </a-tag>
          <a-tag v-else-if="detailData.status === 3" color="red">已拒绝</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="交易哈希">
          <template v-if="detailData.txid">
            <span style="font-family: monospace; word-break: break-all; color: #00b42a">
              {{ detailData.txid }}
            </span>
            <a-button type="text" size="mini" @click="copyText(detailData.txid)">
              <template #icon><icon-copy /></template>
            </a-button>
          </template>
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(detailData.create_time) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ formatTime(detailData.update_time) }}
        </a-descriptions-item>
        <a-descriptions-item label="备注" :span="2">
          {{ detailData.notes || '-' }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconCheck,
    IconClose,
    IconCopy,
    IconEye,
    IconSend,
  } from '@arco-design/web-vue/es/icon';
  import dayjs from 'dayjs';
  import {
    getWithdrawalList,
    approveWithdrawal,
    rejectWithdrawal,
    type WithdrawalItem,
  } from '@/api/withdrawal';

  const loading = ref(false);
  const submitLoading = ref(false);
  const approveModalVisible = ref(false);
  const rejectModalVisible = ref(false);
  const detailModalVisible = ref(false);

  const tableData = ref<WithdrawalItem[]>([]);

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

  const stats = reactive({
    pending: 0,
    approved: 0,
    rejected: 0,
    pendingAmount: 0,
  });

  const approveForm = reactive({
    id: 0,
    number: 0,
    currency_name: '',
    address: '',
    method: 'manual',
    txid: '',
    notes: '',
    // 区块链提现相关
    network_type: '',
    chain_address: '',
  });

  const rejectForm = reactive({
    id: 0,
    number: 0,
    currency_name: '',
    account_number: '',
    reason: '',
  });

  const detailData = reactive<Partial<WithdrawalItem>>({});

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户', slotName: 'user', width: 150 },
    { title: '币种', slotName: 'currency', width: 100 },
    { title: '网络类型', slotName: 'network_type', width: 120 },
    { title: '提现金额', slotName: 'amount', width: 160 },
    { title: '区块链地址', slotName: 'address', width: 220 },
    { title: '状态', slotName: 'status', width: 100 },
    { title: '交易哈希', slotName: 'txid', width: 180 },
    { title: '申请时间', slotName: 'create_time', width: 170 },
    {
      title: '备注',
      dataIndex: 'notes',
      width: 150,
      ellipsis: true,
      tooltip: true,
    },
    { title: '操作', slotName: 'actions', width: 140, fixed: 'right' as const },
  ];

  const formatAmount = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(4);
  };

  const formatAddress = (address?: string) => {
    if (!address) return '-';
    if (address.length <= 16) return address;
    return `${address.substring(0, 8)}...${address.substring(
      address.length - 8
    )}`;
  };

  /**
   * 获取有效的区块链地址
   * 过滤掉 pending_verification 等无效值
   */
  const getValidAddress = (record: any): string => {
    const chainAddr = record.chain_address;
    const addr = record.address;
    
    // 优先使用 chain_address
    if (chainAddr && chainAddr !== 'pending_verification' && chainAddr.trim() !== '') {
      return chainAddr;
    }
    
    // 其次使用 address
    if (addr && addr !== 'pending_verification' && addr.trim() !== '') {
      return addr;
    }
    
    // 尝试使用 to_adddress
    if (record.to_adddress && record.to_adddress !== 'pending_verification' && record.to_adddress.trim() !== '') {
      return record.to_adddress;
    }
    
    // 所有地址都无效
    return '待补充地址';
  };

  const formatTime = (value?: number) => {
    if (!value) return '-';
    return dayjs.unix(value).format('YYYY-MM-DD HH:mm:ss');
  };

  const copyText = (text: string) => {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        Message.success('已复制到剪贴板');
      })
      .catch(() => {
        Message.error('复制失败');
      });
  };

  const calculateStats = () => {
    let pending = 0;
    let approved = 0;
    let rejected = 0;
    let pendingAmount = 0;

    tableData.value.forEach((item) => {
      if (item.status === 1) {
        pending += 1;
        pendingAmount += item.number;
      } else if (item.status === 2) {
        approved += 1;
      } else if (item.status === 3) {
        rejected += 1;
      }
    });

    stats.pending = pending;
    stats.approved = approved;
    stats.rejected = rejected;
    stats.pendingAmount = pendingAmount;
  };

  const fetchList = async () => {
    try {
      loading.value = true;
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        account_number: searchForm.account_number || undefined,
        status: searchForm.status,
      };

      const response = await getWithdrawalList(params);
      if (response.data) {
        const resData = response.data as any;
        tableData.value = resData.data?.list || [];
        pagination.total = resData.data?.total || 0;

        // 计算统计数据
        calculateStats();
      }
    } catch (error: any) {
      Message.error(error.message || '获取提现列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchList();
  };

  const handleReset = () => {
    searchForm.account_number = '';
    searchForm.status = undefined;
    pagination.current = 1;
    fetchList();
  };

  const handleRefresh = () => {
    fetchList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchList();
  };

  const handleApprove = (record: WithdrawalItem) => {
    const validAddress = getValidAddress(record);
    Object.assign(approveForm, {
      id: record.id,
      number: record.number,
      currency_name: record.currency_name || `币种${record.currency_id}`,
      address: validAddress,
      method: 'manual',
      txid: '',
      notes: '',
      // 区块链提现相关
      network_type: (record as any).network_type || 'TRC20',
      chain_address: validAddress,
    });
    approveModalVisible.value = true;
  };

  const handleApproveSubmit = async () => {
    try {
      submitLoading.value = true;
      await approveWithdrawal({
        id: approveForm.id,
        method: approveForm.method,
        txid: approveForm.txid || undefined,
        notes: approveForm.notes || undefined,
      });
      Message.success('审核通过成功');
      approveModalVisible.value = false;
      fetchList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    } finally {
      submitLoading.value = false;
    }
  };

  const handleReject = (record: WithdrawalItem) => {
    Object.assign(rejectForm, {
      id: record.id,
      number: record.number,
      currency_name: record.currency_name || `币种${record.currency_id}`,
      account_number: record.account_number || '-',
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
      submitLoading.value = true;
      await rejectWithdrawal({
        id: rejectForm.id,
        reason: rejectForm.reason,
      });
      Message.success('已拒绝该提现申请');
      rejectModalVisible.value = false;
      fetchList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    } finally {
      submitLoading.value = false;
    }
  };

  const handleViewDetail = (record: WithdrawalItem) => {
    Object.assign(detailData, record);
    detailModalVisible.value = true;
  };

  onMounted(() => {
    fetchList();
  });
</script>

<style scoped lang="less">
  .withdrawal-container {
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

    .user-account {
      font-weight: 500;
    }

    .amount-value {
      font-weight: 600;
      color: #165dff;
    }

    .amount-real {
      font-size: 12px;
      color: #86909c;
    }

    .address-text {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 12px;
    }

    .chain-address {
      color: #722ed1;
    }

    .network-tag {
      font-size: 11px;
      color: #86909c;
      margin-top: 4px;
    }

    .txid-text {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 12px;
      color: #00b42a;
    }

    .text-gray {
      color: #c9cdd4;
    }
  }
</style>
