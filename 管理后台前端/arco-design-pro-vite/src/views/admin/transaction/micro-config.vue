<template>
  <div class="micro-config-container">
    <a-card class="table-card" :bordered="false">
      <template #title>
        <div class="table-header">
          <span>秒合约周期配置</span>
          <a-space>
            <a-button type="primary" @click="handleAdd">
              <template #icon><icon-plus /></template>
              添加周期
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
        :data="tableData"
        :loading="loading"
        row-key="id"
      >
        <template #seconds="{ record }">
          <span class="seconds-tag">{{ record.seconds }}秒</span>
        </template>
        <template #profit_ratio="{ record }">
          <span class="profit-ratio">{{ (record.profit_ratio * 100).toFixed(0) }}%</span>
        </template>
        <template #status="{ record }">
          <a-switch
            :model-value="record.status === 1"
            @change="(val) => handleStatusChange(record, val)"
          >
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </a-switch>
        </template>
        <template #actions="{ record }">
          <a-button
            type="text"
            size="small"
            @click="handleEdit(record)"
          >
            <template #icon><icon-edit /></template>
            编辑
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
        </template>
      </a-table>
    </a-card>

    <!-- 添加/编辑弹窗 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="isEdit ? '编辑秒合约周期' : '添加秒合约周期'"
      :width="480"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
    >
      <a-form :model="formData" :label-col-props="{ span: 6 }">
        <a-form-item label="周期(秒)" required>
          <a-input-number
            v-model="formData.seconds"
            placeholder="请输入周期秒数"
            :min="10"
            :max="3600"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="收益回报率" required>
          <a-input-number
            v-model="formData.profit_ratio"
            placeholder="请输入收益回报率"
            :min="1"
            :max="100"
            :precision="2"
            style="width: 100%"
          >
            <template #append>%</template>
          </a-input-number>
        </a-form-item>
        <a-form-item label="状态">
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
  import { ref, reactive, onMounted } from 'vue';
  import { Message, Modal } from '@arco-design/web-vue';
  import { IconPlus, IconSync, IconEdit, IconDelete } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const loading = ref(false);
  const modalVisible = ref(false);
  const isEdit = ref(false);
  const editId = ref<number>(0);
  const tableData = ref<any[]>([]);

  const formData = reactive({
    seconds: 30,
    profit_ratio: 80,
    status: 1,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '周期', dataIndex: 'seconds', slotName: 'seconds', width: 120 },
    { title: '收益回报率', dataIndex: 'profit_ratio', slotName: 'profit_ratio', width: 150 },
    { title: '状态', dataIndex: 'status', slotName: 'status', width: 120 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    { title: '操作', slotName: 'actions', width: 180 },
  ];

  const fetchList = async () => {
    try {
      loading.value = true;
      const response = await axios.get('/admin/micro/seconds/list');
      if (response.data && response.data.data) {
        tableData.value = response.data.data || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取周期配置失败');
    } finally {
      loading.value = false;
    }
  };

  const handleAdd = () => {
    isEdit.value = false;
    editId.value = 0;
    formData.seconds = 30;
    formData.profit_ratio = 80;
    formData.status = 1;
    modalVisible.value = true;
  };

  const handleEdit = (record: any) => {
    isEdit.value = true;
    editId.value = record.id;
    formData.seconds = record.seconds;
    // 将小数转换为百分比显示（如 0.80 -> 80）
    formData.profit_ratio = record.profit_ratio < 1 ? Math.round(record.profit_ratio * 100) : record.profit_ratio;
    formData.status = record.status;
    modalVisible.value = true;
  };

  const handleSubmit = async () => {
    if (!formData.seconds || formData.seconds < 10) {
      Message.error('周期必须大于等于10秒');
      return;
    }
    if (!formData.profit_ratio || formData.profit_ratio <= 0) {
      Message.error('收益回报率必须大于0');
      return;
    }

    try {
      if (isEdit.value) {
        await axios.put(`/admin/micro/seconds/${editId.value}`, {
          seconds: formData.seconds,
          profit_ratio: formData.profit_ratio,
          status: formData.status,
        });
        Message.success('更新成功');
      } else {
        await axios.post('/admin/micro/seconds/create', {
          seconds: formData.seconds,
          profit_ratio: formData.profit_ratio,
          status: formData.status,
        });
        Message.success('创建成功');
      }
      modalVisible.value = false;
      fetchList();
    } catch (error: any) {
      Message.error(error.message || (isEdit.value ? '更新失败' : '创建失败'));
    }
  };

  const handleStatusChange = async (record: any, val: boolean) => {
    try {
      const newStatus = val ? 1 : 0;
      await axios.post('/admin/micro/seconds/status', {
        id: record.id,
        status: newStatus,
      });
      record.status = newStatus;
      Message.success('状态更新成功');
    } catch (error: any) {
      Message.error(error.message || '状态更新失败');
    }
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除 ${record.seconds}秒 的周期配置吗？`,
      onOk: async () => {
        try {
          await axios.delete(`/admin/micro/seconds/${record.id}`);
          Message.success('删除成功');
          fetchList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleRefresh = () => {
    fetchList();
  };

  onMounted(() => {
    fetchList();
  });
</script>

<style scoped lang="less">
  .micro-config-container {
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

    .seconds-tag {
      display: inline-block;
      padding: 4px 12px;
      background: #e8f4ff;
      color: #1890ff;
      border-radius: 4px;
      font-weight: 500;
    }

    .profit-ratio {
      color: #52c41a;
      font-weight: 600;
      font-size: 14px;
    }
  }
</style>
