<template>
  <div class="deposit-address-container">
    <!-- 搜索区域 -->
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="网络类型">
          <a-select
            v-model="searchForm.network"
            placeholder="全部"
            style="width: 150px"
            allow-clear
          >
            <a-option value="TRC20">TRC20</a-option>
            <a-option value="ERC20">ERC20</a-option>
            <a-option value="BEP20">BEP20</a-option>
            <a-option value="BRC20">BRC20</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model="searchForm.status"
            placeholder="全部"
            style="width: 120px"
            allow-clear
          >
            <a-option :value="1">启用</a-option>
            <a-option :value="0">禁用</a-option>
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

    <!-- 表格区域 -->
    <a-card class="table-card" :bordered="false" style="margin-top: 16px">
      <template #title>
        <div class="table-header">
          <span>充值地址列表</span>
          <a-space>
            <a-button type="primary" @click="handleAdd">
              <template #icon><icon-plus /></template>
              新增地址
            </a-button>
            <a-button @click="handleRefresh">
              <template #icon><icon-sync /></template>
              刷新
            </a-button>
          </a-space>
        </div>
      </template>

      <a-table
        :columns="columns"
        :data="filteredData"
        :pagination="pagination"
        :loading="loading"
        :scroll="{ x: 1200 }"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #network="{ record }">
          <a-tag :color="getNetworkColor(record.network)">
            {{ record.network }}
          </a-tag>
        </template>

        <template #address="{ record }">
          <a-tooltip :content="record.address">
            <span class="address-text">
              {{ formatAddress(record.address) }}
            </span>
          </a-tooltip>
          <a-button
            type="text"
            size="mini"
            @click="copyAddress(record.address)"
          >
            <template #icon><icon-copy /></template>
          </a-button>
        </template>

        <template #qr_code="{ record }">
          <a-image
            v-if="record.qr_code"
            :src="getImageUrl(record.qr_code)"
            width="40"
            height="40"
            :preview="true"
            alt="二维码"
          />
          <span v-else class="text-gray">未设置</span>
        </template>

        <template #status="{ record }">
          <a-switch
            :model-value="record.status === 1"
            :loading="record.statusLoading"
            @change="(val) => handleStatusChange(record, val)"
          >
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </a-switch>
        </template>

        <template #sort="{ record }">
          <a-tag>{{ record.sort }}</a-tag>
        </template>

        <template #create_time="{ record }">
          {{ formatTime(record.create_time) }}
        </template>

        <template #actions="{ record }">
          <a-space>
            <a-button type="text" size="small" @click="handleEdit(record)">
              <template #icon><icon-edit /></template>
              编辑
            </a-button>
            <a-popconfirm
              content="确定要删除这个充值地址吗？"
              @ok="handleDelete(record)"
            >
              <a-button type="text" size="small" status="danger">
                <template #icon><icon-delete /></template>
                删除
              </a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 新增/编辑弹窗 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="isEdit ? '编辑充值地址' : '新增充值地址'"
      :width="550"
      :ok-loading="submitLoading"
      @ok="handleSubmit"
      @cancel="handleCancel"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        :label-col-props="{ span: 5 }"
        :wrapper-col-props="{ span: 19 }"
      >
        <a-form-item label="网络类型" field="network">
          <a-select
            v-model="formData.network"
            placeholder="请选择网络类型"
            :disabled="isEdit"
          >
            <a-option value="TRC20">TRC20 (波场)</a-option>
            <a-option value="ERC20">ERC20 (以太坊)</a-option>
            <a-option value="BEP20">BEP20 (币安智能链)</a-option>
            <a-option value="BRC20">BRC20 (比特币)</a-option>
          </a-select>
        </a-form-item>

        <a-form-item label="充值地址" field="address">
          <a-input
            v-model="formData.address"
            placeholder="请输入充值地址"
            :max-length="128"
            show-word-limit
          />
        </a-form-item>

        <a-form-item label="二维码图片" field="qr_code">
          <a-input
            v-model="formData.qr_code"
            placeholder="请输入二维码图片路径，如: /uploads/qrcode/xxx.png"
          />
          <template #extra>
            <span class="form-extra">支持 jpg、png、gif 格式</span>
          </template>
        </a-form-item>

        <a-form-item label="排序" field="sort">
          <a-input-number
            v-model="formData.sort"
            placeholder="数值越大越靠前"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="状态" field="status">
          <a-radio-group v-model="formData.status">
            <a-radio :value="1">启用</a-radio>
            <a-radio :value="0">禁用</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted, watch } from 'vue';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconPlus,
    IconEdit,
    IconDelete,
    IconCopy,
  } from '@arco-design/web-vue/es/icon';
  import dayjs from 'dayjs';
  import {
    getDepositAddressList,
    createDepositAddress,
    updateDepositAddress,
    deleteDepositAddress,
    type DepositAddress,
  } from '@/api/deposit';

  const loading = ref(false);
  const modalVisible = ref(false);
  const isEdit = ref(false);
  const submitLoading = ref(false);
  const formRef = ref<FormInstance>();

  const tableData = ref<(DepositAddress & { statusLoading?: boolean })[]>([]);

  const searchForm = reactive({
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

  const formData = reactive({
    id: 0,
    network: '',
    address: '',
    qr_code: '',
    sort: 0,
    status: 1,
  });

  const formRules = {
    network: [{ required: true, message: '请选择网络类型' }],
    address: [
      { required: true, message: '请输入充值地址' },
      { minLength: 20, message: '地址长度不能少于20个字符' },
    ],
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '网络类型', slotName: 'network', width: 120 },
    { title: '充值地址', slotName: 'address', width: 300 },
    { title: '二维码', slotName: 'qr_code', width: 100 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '排序', slotName: 'sort', width: 80 },
    { title: '创建时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 160, fixed: 'right' as const },
  ];

  // 过滤后的完整数据
  const allFilteredData = computed(() => {
    let data = tableData.value;
    if (searchForm.network) {
      data = data.filter((item) => item.network === searchForm.network);
    }
    if (searchForm.status !== undefined) {
      data = data.filter((item) => item.status === searchForm.status);
    }
    return data;
  });

  // 分页后的数据
  const filteredData = computed(() => {
    const start = (pagination.current - 1) * pagination.pageSize;
    return allFilteredData.value.slice(start, start + pagination.pageSize);
  });

  // 监听过滤数据变化，更新总数
  watch(
    () => allFilteredData.value.length,
    (newTotal) => {
      pagination.total = newTotal;
    },
    { immediate: true }
  );

  const getNetworkColor = (network: string) => {
    const colors: Record<string, string> = {
      TRC20: 'green',
      ERC20: 'blue',
      BEP20: 'orange',
      BRC20: 'purple',
    };
    return colors[network] || 'gray';
  };

  const formatAddress = (address?: string) => {
    if (!address) return '-';
    if (address.length <= 20) return address;
    return `${address.substring(0, 10)}...${address.substring(
      address.length - 10
    )}`;
  };

  const formatTime = (value?: number) => {
    if (!value) return '-';
    return dayjs.unix(value).format('YYYY-MM-DD HH:mm:ss');
  };

  const getImageUrl = (path: string) => {
    if (!path) return '';
    if (path.startsWith('http')) return path;
    return `${window.location.origin}${path}`;
  };

  const copyAddress = (address: string) => {
    navigator.clipboard
      .writeText(address)
      .then(() => {
        Message.success('地址已复制到剪贴板');
      })
      .catch(() => {
        Message.error('复制失败');
      });
  };

  const fetchAddressList = async () => {
    try {
      loading.value = true;
      const response = await getDepositAddressList();
      if (response.data) {
        tableData.value = (response.data as any).data || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取充值地址列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
  };

  const handleReset = () => {
    searchForm.network = undefined;
    searchForm.status = undefined;
    pagination.current = 1;
  };

  const handleRefresh = () => {
    fetchAddressList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
  };

  const handleAdd = () => {
    isEdit.value = false;
    Object.assign(formData, {
      id: 0,
      network: '',
      address: '',
      qr_code: '',
      sort: 0,
      status: 1,
    });
    modalVisible.value = true;
  };

  const handleEdit = (record: DepositAddress) => {
    isEdit.value = true;
    Object.assign(formData, {
      id: record.id,
      network: record.network,
      address: record.address,
      qr_code: record.qr_code,
      sort: record.sort,
      status: record.status,
    });
    modalVisible.value = true;
  };

  const handleSubmit = async () => {
    try {
      const valid = await formRef.value?.validate();
      if (valid) return;

      submitLoading.value = true;

      if (isEdit.value) {
        await updateDepositAddress(formData.id, {
          address: formData.address,
          qr_code: formData.qr_code,
          status: formData.status,
          sort: formData.sort,
        });
        Message.success('更新成功');
      } else {
        await createDepositAddress({
          network: formData.network,
          address: formData.address,
          qr_code: formData.qr_code,
          status: formData.status,
          sort: formData.sort,
        });
        Message.success('创建成功');
      }

      modalVisible.value = false;
      fetchAddressList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    } finally {
      submitLoading.value = false;
    }
  };

  const handleCancel = () => {
    formRef.value?.resetFields();
    modalVisible.value = false;
  };

  const handleStatusChange = async (record: any, checked: boolean) => {
    const newStatus = checked ? 1 : 0;
    const statusText = checked ? '启用' : '禁用';

    Modal.confirm({
      title: '确认操作',
      content: `确定要${statusText}该充值地址吗？`,
      onOk: async () => {
        try {
          record.statusLoading = true;
          await updateDepositAddress(record.id, { status: newStatus });
          record.status = newStatus;
          Message.success(`${statusText}成功`);
        } catch (error: any) {
          Message.error(error.message || '操作失败');
        } finally {
          record.statusLoading = false;
        }
      },
    });
  };

  const handleDelete = async (record: DepositAddress) => {
    try {
      await deleteDepositAddress(record.id);
      Message.success('删除成功');
      fetchAddressList();
    } catch (error: any) {
      Message.error(error.message || '删除失败');
    }
  };

  onMounted(() => {
    fetchAddressList();
  });
</script>

<style scoped lang="less">
  .deposit-address-container {
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

    .address-text {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 13px;
    }

    .text-gray {
      color: #999;
    }

    .form-extra {
      color: #999;
      font-size: 12px;
    }
  }
</style>
