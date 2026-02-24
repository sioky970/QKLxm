<template>
  <div class="level-list-container">
    <a-card class="table-card" :bordered="false">
      <template #title>
        <div class="table-header">
          <span>等级管理</span>
          <a-space>
            <a-button type="primary" @click="handleAdd">
              <template #icon><icon-plus /></template>
              新增等级
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
        :pagination="false"
        row-key="id"
      >
        <template #actions="{ record }">
          <a-space>
            <a-button type="text" size="small" @click="handleEdit(record)">
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
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="modalVisible"
      :title="modalTitle"
      :width="560"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
    >
      <a-form :model="formData" :label-col-props="{ span: 7 }">
        <a-form-item label="等级名称" required>
          <a-input v-model="formData.name" placeholder="如：VIP1" />
        </a-form-item>
        <a-form-item label="等级编号" required>
          <a-input-number
            v-model="formData.level"
            :min="0"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="充币数量">
          <a-input-number
            v-model="formData.fill_currency"
            :min="0"
            :precision="4"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="直推数量">
          <a-input-number
            v-model="formData.direct_drive_count"
            :min="0"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="直推金额">
          <a-input-number
            v-model="formData.direct_drive_price"
            :min="0"
            :precision="4"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="最大代数">
          <a-input-number
            v-model="formData.max_algebra"
            :min="0"
            style="width: 100%"
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
    IconPlus,
    IconSync,
    IconEdit,
    IconDelete,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const loading = ref(false);
  const tableData = ref<any[]>([]);
  const modalVisible = ref(false);
  const modalTitle = ref('新增等级');

  const formData = reactive({
    id: undefined as number | undefined,
    name: '',
    level: 0,
    fill_currency: 0,
    direct_drive_count: 0,
    direct_drive_price: 0,
    max_algebra: 0,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '等级名称', dataIndex: 'name', width: 160 },
    { title: '等级编号', dataIndex: 'level', width: 120 },
    { title: '充币数量', dataIndex: 'fill_currency', width: 140 },
    { title: '直推数量', dataIndex: 'direct_drive_count', width: 120 },
    { title: '直推金额', dataIndex: 'direct_drive_price', width: 140 },
    { title: '最大代数', dataIndex: 'max_algebra', width: 120 },
    { title: '操作', slotName: 'actions', width: 160, fixed: 'right' },
  ];

  const fetchLevelList = async () => {
    try {
      loading.value = true;
      const response = await axios.get('/admin/level/list');
      if (response.data && response.data.data) {
        tableData.value = response.data.data || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取等级列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleAdd = () => {
    modalTitle.value = '新增等级';
    Object.assign(formData, {
      id: undefined,
      name: '',
      level: 0,
      fill_currency: 0,
      direct_drive_count: 0,
      direct_drive_price: 0,
      max_algebra: 0,
    });
    modalVisible.value = true;
  };

  const handleEdit = (record: any) => {
    modalTitle.value = '编辑等级';
    Object.assign(formData, {
      id: record.id,
      name: record.name,
      level: record.level,
      fill_currency: record.fill_currency,
      direct_drive_count: record.direct_drive_count,
      direct_drive_price: record.direct_drive_price,
      max_algebra: record.max_algebra,
    });
    modalVisible.value = true;
  };

  const handleSubmit = async () => {
    if (!formData.name) {
      Message.warning('请输入等级名称');
      return;
    }
    try {
      const payload = {
        id: formData.id,
        name: formData.name,
        level: formData.level,
        fill_currency: formData.fill_currency,
        direct_drive_count: formData.direct_drive_count,
        direct_drive_price: formData.direct_drive_price,
        max_algebra: formData.max_algebra,
      };
      if (formData.id) {
        await axios.put('/admin/level/update', payload);
        Message.success('更新成功');
      } else {
        await axios.post('/admin/level/create', payload);
        Message.success('创建成功');
      }
      modalVisible.value = false;
      fetchLevelList();
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    }
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除等级 ${record.name} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/level/${record.id}`);
          Message.success('删除成功');
          fetchLevelList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleRefresh = () => {
    fetchLevelList();
  };

  onMounted(() => {
    fetchLevelList();
  });
</script>

<style scoped lang="less">
  .level-list-container {
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
