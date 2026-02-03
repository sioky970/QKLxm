<template>
  <div class="lever-multiple-container">
    <a-card class="table-card" :bordered="false">
      <template #title>
        <div class="table-header">
          <span>永续合约倍数配置</span>
          <a-space>
            <a-button type="primary" @click="handleAdd">
              <template #icon><icon-plus /></template>
              添加配置
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
        <template #type="{ record }">
          <a-tag :color="record.type === 'multiple' ? 'blue' : 'purple'">
            {{ record.type === 'multiple' ? '倍数' : '手数' }}
          </a-tag>
        </template>
        <template #actions="{ record }">
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

    <a-modal
      v-model:visible="modalVisible"
      title="添加永续合约倍数配置"
      :width="420"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
    >
      <a-form :model="formData" :label-col-props="{ span: 6 }">
        <a-form-item label="类型" required>
          <a-radio-group v-model="formData.type">
            <a-radio value="multiple">倍数</a-radio>
            <a-radio value="hand">手数</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item label="数值" required>
          <a-input v-model="formData.value" placeholder="如：10x 或 0.1" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message, Modal } from '@arco-design/web-vue';
  import { IconPlus, IconSync, IconDelete } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const loading = ref(false);
  const modalVisible = ref(false);
  const tableData = ref<any[]>([]);

  const formData = reactive({
    type: 'multiple',
    value: '',
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '类型', dataIndex: 'type', slotName: 'type', width: 120 },
    { title: '数值', dataIndex: 'value', width: 200 },
    { title: '操作', slotName: 'actions', width: 120 },
  ];

  const fetchList = async () => {
    try {
      loading.value = true;
      const response = await axios.get('/admin/lever/multiple/list');
      if (response.data && response.data.data) {
        tableData.value = response.data.data || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取倍数配置失败');
    } finally {
      loading.value = false;
    }
  };

  const handleRefresh = () => {
    fetchList();
  };

  const handleAdd = () => {
    formData.type = 'multiple';
    formData.value = '';
    modalVisible.value = true;
  };

  const handleSubmit = async () => {
    if (!formData.value) {
      Message.warning('请输入数值');
      return;
    }
    try {
      await axios.post('/admin/lever/multiple/create', {
        type: formData.type,
        value: formData.value,
      });
      Message.success('创建成功');
      modalVisible.value = false;
      fetchList();
    } catch (error: any) {
      Message.error(error.message || '创建失败');
    }
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除配置 ${record.value} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/lever/multiple/${record.id}`);
          Message.success('删除成功');
          fetchList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  onMounted(() => {
    fetchList();
  });
</script>

<style scoped lang="less">
  .lever-multiple-container {
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
  }
</style>
