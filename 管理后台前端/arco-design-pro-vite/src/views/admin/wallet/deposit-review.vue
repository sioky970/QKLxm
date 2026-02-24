<template>
  <div class="deposit-review-container">
    <!-- 搜索区域 -->
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="订单号">
          <a-input
            v-model="searchForm.order_no"
            placeholder="请输入订单号"
            style="width: 200px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="用户ID">
          <a-input-number
            v-model="searchForm.user_id"
            placeholder="请输入用户ID"
            style="width: 140px"
            :min="1"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="网络类型">
          <a-select
            v-model="searchForm.network"
            placeholder="全部网络"
            style="width: 140px"
            allow-clear
          >
            <a-option value="TRC20">TRC20</a-option>
            <a-option value="ERC20">ERC20</a-option>
            <a-option value="BEP20">BEP20</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部状态"
            style="width: 140px"
            allow-clear
          >
            <a-option :value="0">待审核</a-option>
            <a-option :value="1">已通过</a-option>
            <a-option :value="2">已拒绝</a-option>
            <a-option :value="3">已取消</a-option>
            <a-option :value="4">已过期</a-option>
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
          <a-statistic title="待审核" :value="stats.pending" :value-style="{ color: '#ff7d00' }">
            <template #suffix>
              <span style="font-size: 14px">笔</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="已通过" :value="stats.approved" :value-style="{ color: '#00b42a' }">
            <template #suffix>
              <span style="font-size: 14px">笔</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="已拒绝" :value="stats.rejected" :value-style="{ color: '#f53f3f' }">
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
            :value-style="{ color: '#165dff' }"
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
          <span>充值订单列表</span>
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
        :scroll="{ x: 1500 }"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #order_no="{ record }">
          <a-tooltip :content="record.order_no">
            <span class="order-no">{{ formatOrderNo(record.order_no) }}</span>
          </a-tooltip>
          <a-button type="text" size="mini" @click="copyText(record.order_no)">
            <template #icon><icon-copy /></template>
          </a-button>
        </template>

        <template #user="{ record }">
          <a-space direction="vertical" :size="2">
            <span class="user-account">{{ record.account_number || '-' }}</span>
            <a-tag size="small" color="arcoblue">
              UID: {{ record.user_id }}
            </a-tag>
          </a-space>
        </template>

        <template #network="{ record }">
          <a-tag color="purple">{{ record.network }}</a-tag>
        </template>

        <template #amount="{ record }">
          <span class="amount-value">{{ formatAmount(record.amount) }} USDT</span>
        </template>

        <template #address="{ record }">
          <a-tooltip :content="record.address">
            <span class="address-text">{{ formatAddress(record.address) }}</span>
          </a-tooltip>
          <a-button type="text" size="mini" @click="copyText(record.address)">
            <template #icon><icon-copy /></template>
          </a-button>
        </template>

        <template #screenshot="{ record }">
          <template v-if="record.screenshot">
            <a-button type="text" size="small" @click="handleViewScreenshot(record)">
              <template #icon><icon-image /></template>
              查看截图
            </a-button>
          </template>
          <span v-else class="text-gray">未上传</span>
        </template>

        <template #status="{ record }">
          <a-tag v-if="record.status === 0" color="orange">待审核</a-tag>
          <a-tag v-else-if="record.status === 1" color="green">已通过</a-tag>
          <a-tag v-else-if="record.status === 2" color="red">已拒绝</a-tag>
          <a-tag v-else-if="record.status === 3" color="gray">已取消</a-tag>
          <a-tag v-else-if="record.status === 4" color="gray">已过期</a-tag>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #actions="{ record }">
          <a-space v-if="record.status === 0">
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
      title="审核通过"
      :width="600"
      :ok-loading="submitLoading"
      @ok="handleApproveSubmit"
      @cancel="approveModalVisible = false"
    >
      <a-alert type="info" style="margin-bottom: 16px">
        请仔细核对用户上传的转账截图和充值金额后再进行审核
      </a-alert>

      <a-descriptions :column="2" bordered style="margin-bottom: 16px">
        <a-descriptions-item label="订单号">
          {{ approveForm.order_no }}
        </a-descriptions-item>
        <a-descriptions-item label="用户ID">
          {{ approveForm.user_id }}
        </a-descriptions-item>
        <a-descriptions-item label="网络类型">
          <a-tag color="purple">{{ approveForm.network }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="充值金额">
          <span style="font-weight: 600; color: #165dff">
            {{ formatAmount(approveForm.amount) }} USDT
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="充值地址" :span="2">
          <span style="font-family: monospace; word-break: break-all">
            {{ approveForm.address }}
          </span>
        </a-descriptions-item>
      </a-descriptions>

      <!-- 截图预览 -->
      <div v-if="approveForm.screenshot" class="screenshot-preview">
        <p style="margin-bottom: 8px; font-weight: 500">转账截图：</p>
        <a-image
          :src="getScreenshotUrl(approveForm.screenshot)"
          :width="300"
          fit="contain"
        />
      </div>

      <a-form :model="approveForm" :label-col-props="{ span: 4 }" style="margin-top: 16px">
        <a-form-item label="审核备注">
          <a-textarea
            v-model="approveForm.remark"
            placeholder="请输入审核备注（选填）"
            :max-length="200"
            show-word-limit
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 拒绝弹窗 -->
    <a-modal
      v-model:visible="rejectModalVisible"
      title="拒绝充值"
      :width="500"
      :ok-loading="submitLoading"
      ok-text="确认拒绝"
      @ok="handleRejectSubmit"
      @cancel="rejectModalVisible = false"
    >
      <a-alert type="warning" style="margin-bottom: 16px">
        拒绝后，用户需重新提交充值申请
      </a-alert>

      <a-descriptions :column="1" bordered style="margin-bottom: 16px">
        <a-descriptions-item label="订单号">
          {{ rejectForm.order_no }}
        </a-descriptions-item>
        <a-descriptions-item label="充值金额">
          {{ formatAmount(rejectForm.amount) }} USDT
        </a-descriptions-item>
        <a-descriptions-item label="用户账号">
          {{ rejectForm.account_number || `UID: ${rejectForm.user_id}` }}
        </a-descriptions-item>
      </a-descriptions>

      <a-form :model="rejectForm" :label-col-props="{ span: 6 }">
        <a-form-item label="拒绝原因" required>
          <a-textarea
            v-model="rejectForm.remark"
            placeholder="请输入拒绝原因，如：截图不清晰、金额不符等"
            :max-length="200"
            show-word-limit
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 详情弹窗 -->
    <a-modal
      v-model:visible="detailModalVisible"
      title="充值订单详情"
      :width="700"
      :footer="false"
    >
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="订单ID">
          {{ detailData.id }}
        </a-descriptions-item>
        <a-descriptions-item label="订单号">
          {{ detailData.order_no }}
        </a-descriptions-item>
        <a-descriptions-item label="用户ID">
          {{ detailData.user_id }}
        </a-descriptions-item>
        <a-descriptions-item label="用户账号">
          {{ detailData.account_number || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="网络类型">
          <a-tag color="purple">{{ detailData.network }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="充值金额">
          <span style="font-weight: 600; color: #165dff">
            {{ formatAmount(detailData.amount) }} USDT
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="充值地址" :span="2">
          <span style="font-family: monospace; word-break: break-all">
            {{ detailData.address }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag v-if="detailData.status === 0" color="orange">待审核</a-tag>
          <a-tag v-else-if="detailData.status === 1" color="green">已通过</a-tag>
          <a-tag v-else-if="detailData.status === 2" color="red">已拒绝</a-tag>
          <a-tag v-else-if="detailData.status === 3" color="gray">已取消</a-tag>
          <a-tag v-else-if="detailData.status === 4" color="gray">已过期</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="审核人">
          {{ detailData.admin_id || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(detailData.create_time) }}
        </a-descriptions-item>
        <a-descriptions-item label="审核时间">
          {{ detailData.review_time ? formatTime(detailData.review_time) : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="审核备注" :span="2">
          {{ detailData.admin_remark || '-' }}
        </a-descriptions-item>
      </a-descriptions>

      <!-- 截图预览 -->
      <div v-if="detailData.screenshot" class="screenshot-preview" style="margin-top: 16px">
        <p style="margin-bottom: 8px; font-weight: 500">转账截图：</p>
        <a-image
          :src="getScreenshotUrl(detailData.screenshot)"
          :width="400"
          fit="contain"
        />
      </div>
    </a-modal>

    <!-- 截图预览弹窗 -->
    <a-modal
      v-model:visible="screenshotModalVisible"
      title="转账截图"
      :width="600"
      :footer="false"
    >
      <div style="text-align: center">
        <a-image
          :src="getScreenshotUrl(currentScreenshot)"
          :width="500"
          fit="contain"
        />
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, computed } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconCheck,
    IconClose,
    IconCopy,
    IconEye,
    IconImage,
  } from '@arco-design/web-vue/es/icon';
  import dayjs from 'dayjs';
  import {
    getDepositOrderList,
    approveDepositOrder,
    rejectDepositOrder,
    type DepositOrder,
  } from '@/api/deposit';

  const loading = ref(false);
  const submitLoading = ref(false);
  const approveModalVisible = ref(false);
  const rejectModalVisible = ref(false);
  const detailModalVisible = ref(false);
  const screenshotModalVisible = ref(false);
  const currentScreenshot = ref('');

  const tableData = ref<DepositOrder[]>([]);

  // 后端 API 基础地址
  const API_BASE_URL = computed(() => {
    // 从当前环境获取，或使用默认值
    return import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  });

  const getScreenshotUrl = (path: string) => {
    if (!path) return '';
    if (path.startsWith('http')) return path;
    // /uploads 开头的路径直接返回（由 Vite 代理处理）
    if (path.startsWith('/uploads')) return path;
    return `${API_BASE_URL.value}${path}`;
  };

  const searchForm = reactive({
    order_no: '',
    user_id: undefined as number | undefined,
    network: undefined as string | undefined,
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
    order_no: '',
    user_id: 0,
    network: '',
    amount: 0,
    address: '',
    screenshot: '',
    remark: '',
  });

  const rejectForm = reactive({
    id: 0,
    order_no: '',
    user_id: 0,
    account_number: '',
    amount: 0,
    remark: '',
  });

  const detailData = reactive<Partial<DepositOrder>>({});

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '订单号', slotName: 'order_no', width: 180 },
    { title: '用户', slotName: 'user', width: 140 },
    { title: '网络', slotName: 'network', width: 90 },
    { title: '充值金额', slotName: 'amount', width: 140 },
    { title: '充值地址', slotName: 'address', width: 200 },
    { title: '转账截图', slotName: 'screenshot', width: 100 },
    { title: '状态', slotName: 'status', width: 90 },
    { title: '创建时间', slotName: 'create_time', width: 170 },
    {
      title: '审核备注',
      dataIndex: 'admin_remark',
      width: 150,
      ellipsis: true,
      tooltip: true,
    },
    { title: '操作', slotName: 'actions', width: 140, fixed: 'right' as const },
  ];

  const formatAmount = (value?: number) => {
    if (value === undefined || value === null) return '-';
    return Number(value).toFixed(2);
  };

  const formatOrderNo = (orderNo?: string) => {
    if (!orderNo) return '-';
    if (orderNo.length <= 16) return orderNo;
    return `${orderNo.substring(0, 8)}...${orderNo.substring(orderNo.length - 6)}`;
  };

  const formatAddress = (address?: string) => {
    if (!address) return '-';
    if (address.length <= 16) return address;
    return `${address.substring(0, 8)}...${address.substring(address.length - 8)}`;
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
      if (item.status === 0) {
        pending += 1;
        pendingAmount += item.amount;
      } else if (item.status === 1) {
        approved += 1;
      } else if (item.status === 2) {
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
        order_no: searchForm.order_no || undefined,
        user_id: searchForm.user_id,
        network: searchForm.network,
        status: searchForm.status,
      };

      const response = await getDepositOrderList(params);
      if (response.data) {
        const resData = response.data as any;
        tableData.value = resData.data || [];
        pagination.total = resData.count || 0;

        // 计算统计数据
        calculateStats();
      }
    } catch (error: any) {
      Message.error(error.message || '获取充值订单列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchList();
  };

  const handleReset = () => {
    searchForm.order_no = '';
    searchForm.user_id = undefined;
    searchForm.network = undefined;
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

  const handleViewScreenshot = (record: DepositOrder) => {
    currentScreenshot.value = record.screenshot;
    screenshotModalVisible.value = true;
  };

  const handleApprove = (record: DepositOrder) => {
    Object.assign(approveForm, {
      id: record.id,
      order_no: record.order_no,
      user_id: record.user_id,
      network: record.network,
      amount: record.amount,
      address: record.address,
      screenshot: record.screenshot,
      remark: '',
    });
    approveModalVisible.value = true;
  };

  const handleApproveSubmit = async () => {
    try {
      submitLoading.value = true;
      await approveDepositOrder({
        order_id: approveForm.id,
        remark: approveForm.remark || undefined,
      });
      Message.success('审核通过成功，已增加用户余额');
      approveModalVisible.value = false;
      fetchList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    } finally {
      submitLoading.value = false;
    }
  };

  const handleReject = (record: DepositOrder) => {
    Object.assign(rejectForm, {
      id: record.id,
      order_no: record.order_no,
      user_id: record.user_id,
      account_number: record.account_number || '',
      amount: record.amount,
      remark: '',
    });
    rejectModalVisible.value = true;
  };

  const handleRejectSubmit = async () => {
    if (!rejectForm.remark) {
      Message.warning('请输入拒绝原因');
      return;
    }

    try {
      submitLoading.value = true;
      await rejectDepositOrder({
        order_id: rejectForm.id,
        remark: rejectForm.remark,
      });
      Message.success('已拒绝该充值申请');
      rejectModalVisible.value = false;
      fetchList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    } finally {
      submitLoading.value = false;
    }
  };

  const handleViewDetail = (record: DepositOrder) => {
    Object.assign(detailData, record);
    detailModalVisible.value = true;
  };

  onMounted(() => {
    fetchList();
  });
</script>

<style scoped lang="less">
  .deposit-review-container {
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

    .order-no {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 12px;
      color: #165dff;
    }

    .user-account {
      font-weight: 500;
    }

    .amount-value {
      font-weight: 600;
      color: #165dff;
    }

    .address-text {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 12px;
    }

    .text-gray {
      color: #c9cdd4;
    }

    .screenshot-preview {
      background: #f7f8fa;
      padding: 16px;
      border-radius: 4px;
      text-align: center;
    }
  }
</style>
