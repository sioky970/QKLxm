<template>
  <div class="kyc-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="账号搜索">
          <a-input
            v-model="searchForm.account"
            placeholder="账号/手机/邮箱"
            style="width: 200px"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="审核状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部状态"
            style="width: 150px"
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
          <span>实名认证列表</span>
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
            <span style="font-weight: 500">
              {{ record.account_number || '-' }}
            </span>
            <span class="sub-text">{{
              record.phone || record.email || '-'
            }}</span>
          </a-space>
        </template>

        <template #card_id="{ record }">
          {{ maskCardId(record.card_id) }}
        </template>

        <template #status="{ record }">
          <a-tag :color="getStatusColor(record.review_status)">
            {{ getStatusLabel(record.review_status) }}
          </a-tag>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #review_time="{ record }">
          {{ formatTime(record.review_time) }}
        </template>

        <template #actions="{ record }">
          <a-space>
            <a-button type="text" size="small" @click="handleView(record)">
              <template #icon><icon-eye /></template>
              查看
            </a-button>
            <a-button
              v-if="isPending(record.review_status)"
              type="text"
              size="small"
              status="success"
              @click="handleApprove(record)"
            >
              <template #icon><icon-check /></template>
              通过
            </a-button>
            <a-button
              v-if="isPending(record.review_status)"
              type="text"
              size="small"
              status="danger"
              @click="handleReject(record)"
            >
              <template #icon><icon-close /></template>
              拒绝
            </a-button>
            <a-button
              type="text"
              size="small"
              status="danger"
              @click="handleDelete(record)"
            >
              <template #icon><icon-delete /></template>
              删除
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-drawer v-model:visible="detailVisible" title="实名认证详情" :width="620">
      <a-spin :loading="detailLoading">
        <div v-if="detailData">
          <a-descriptions :column="1" bordered size="large">
            <a-descriptions-item label="用户账号">
              {{ detailData.user?.account_number || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="手机号">
              {{ detailData.user?.phone || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="邮箱">
              {{ detailData.user?.email || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="姓名">
              {{ detailData.kyc?.name || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="身份证号">
              {{ maskCardId(detailData.kyc?.card_id) }}
            </a-descriptions-item>
            <a-descriptions-item label="手机号">
              {{ detailData.kyc?.phone || detailData.user?.phone || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="区块链地址">
              {{ detailData.kyc?.bank_card || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="审核状态">
              <a-tag :color="getStatusColor(detailData.kyc?.review_status)">
                {{ getStatusLabel(detailData.kyc?.review_status) }}
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="拒绝原因" v-if="detailData.kyc?.review_status === 3">
              <span style="color: #f53f3f">{{ detailData.kyc?.reject_reason || '-' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="提交时间">
              {{ formatTime(detailData.kyc?.create_time) }}
            </a-descriptions-item>
            <a-descriptions-item label="审核时间">
              {{ formatTime(detailData.kyc?.review_time) }}
            </a-descriptions-item>
          </a-descriptions>

          <div class="image-section">
            <div class="image-title">证件照片</div>
            <a-space wrap>
              <a-image
                v-if="detailData.kyc?.front_pic"
                :src="getImageUrl(detailData.kyc.front_pic)"
                width="180"
                height="120"
                fit="cover"
                show-loader
              />
              <a-image
                v-if="detailData.kyc?.reverse_pic"
                :src="getImageUrl(detailData.kyc.reverse_pic)"
                width="180"
                height="120"
                fit="cover"
                show-loader
              />
              <a-image
                v-if="detailData.kyc?.hand_pic"
                :src="getImageUrl(detailData.kyc.hand_pic)"
                width="180"
                height="120"
                fit="cover"
                show-loader
              />
            </a-space>
          </div>
        </div>
      </a-spin>
    </a-drawer>

    <a-modal
      v-model:visible="rejectVisible"
      title="拒绝实名认证"
      :width="420"
      @ok="handleRejectSubmit"
      @cancel="rejectVisible = false"
    >
      <a-form :model="rejectForm" :label-col-props="{ span: 5 }">
        <a-form-item label="ID">
          <a-input :model-value="String(rejectForm.id)" disabled />
        </a-form-item>
        <a-form-item label="原因" required>
          <a-textarea
            v-model="rejectForm.reason"
            placeholder="请输入拒绝原因"
            :max-length="200"
            show-word-limit
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
    IconEye,
    IconCheck,
    IconClose,
    IconDelete,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';
  import { getImageUrl } from '@/utils/image';

  const loading = ref(false);
  const tableData = ref<any[]>([]);
  const detailVisible = ref(false);
  const detailLoading = ref(false);
  const detailData = ref<any>(null);
  const rejectVisible = ref(false);

  const searchForm = reactive({
    account: '',
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
    id: '' as any,
    reason: '',
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '账号', slotName: 'account', width: 200 },
    { title: '姓名', dataIndex: 'name', width: 120 },
    { title: '身份证号', slotName: 'card_id', width: 160 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '提交时间', slotName: 'create_time', width: 180 },
    { title: '审核时间', slotName: 'review_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 260, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const maskCardId = (cardId?: string) => {
    if (!cardId) return '-';
    if (cardId.length <= 8) return cardId;
    return `${cardId.slice(0, 6)}****${cardId.slice(-4)}`;
  };

  const isPending = (status?: number) => status === 1 || status === 0;

  const getStatusLabel = (status?: number) => {
    if (status === 2) return '已通过';
    if (status === 3) return '已拒绝';
    if (status === 1 || status === 0) return '待审核';
    return '未知';
  };

  const getStatusColor = (status?: number) => {
    if (status === 2) return 'green';
    if (status === 3) return 'red';
    if (status === 1 || status === 0) return 'orange';
    return 'gray';
  };

  const fetchKycList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/kyc/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
        account: searchForm.account || undefined,
        status: searchForm.status,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取实名认证列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    fetchKycList();
  };

  const handleReset = () => {
    searchForm.account = '';
    searchForm.status = undefined;
    pagination.current = 1;
    fetchKycList();
  };

  const handleRefresh = () => {
    fetchKycList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchKycList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchKycList();
  };

  const handleView = async (record: any) => {
    detailVisible.value = true;
    detailLoading.value = true;
    try {
      const response = await axios.get(`/admin/kyc/${record.id}`);
      if (response.data && response.data.data) {
        detailData.value = response.data.data;
      } else {
        detailData.value = { kyc: record, user: {} };
      }
    } catch (error: any) {
      detailData.value = { kyc: record, user: {} };
      Message.error(error.message || '获取实名认证详情失败');
    } finally {
      detailLoading.value = false;
    }
  };

  const handleApprove = (record: any) => {
    Modal.confirm({
      title: '确认通过',
      content: `确定通过 ${record.account_number || ''} 的实名认证吗？`,
      onOk: async () => {
        try {
          await axios.post('/admin/kyc/approve', { id: record.id });
          Message.success('审核通过');
          fetchKycList();
        } catch (error: any) {
          Message.error(error.message || '操作失败');
        }
      },
    });
  };

  const handleReject = (record: any) => {
    rejectForm.id = record.id;
    rejectForm.reason = '';
    rejectVisible.value = true;
  };

  const handleRejectSubmit = async () => {
    if (!rejectForm.reason) {
      Message.warning('请输入拒绝原因');
      return;
    }
    try {
      await axios.post('/admin/kyc/reject', {
        id: rejectForm.id,
        reason: rejectForm.reason,
      });
      Message.success('已拒绝');
      rejectVisible.value = false;
      fetchKycList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    }
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后不可恢复，确定删除该实名认证记录吗？',
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/kyc/${record.id}`);
          Message.success('删除成功');
          fetchKycList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  onMounted(() => {
    fetchKycList();
  });
</script>

<style scoped lang="less">
  .kyc-list-container {
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

    .sub-text {
      color: var(--color-text-3);
      font-size: 12px;
    }

    .image-section {
      margin-top: 20px;

      .image-title {
        font-weight: 500;
        margin-bottom: 12px;
      }
    }

    :deep(.arco-table) {
      margin-top: 16px;
    }
  }
</style>
