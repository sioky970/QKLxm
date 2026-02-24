<template>
  <div class="role-list-container">
    <a-card class="table-card" :bordered="false">
      <template #title>
        <div class="table-header">
          <span>角色管理</span>
          <a-space>
            <a-button type="primary" @click="handleAdd">
              <template #icon><icon-plus /></template>
              新增角色
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
        <template #is_super="{ record }">
          <a-tag :color="record.is_super === 1 ? 'red' : 'blue'">
            {{ record.is_super === 1 ? '超级管理员' : '普通角色' }}
          </a-tag>
        </template>

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
      :width="520"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
    >
      <a-form :model="formData" :label-col-props="{ span: 6 }">
        <a-form-item label="角色名称" required>
          <a-input v-model="formData.name" placeholder="请输入角色名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model="formData.description"
            placeholder="请输入角色描述"
            :max-length="200"
            show-word-limit
          />
        </a-form-item>
        <a-form-item label="超级管理员">
          <a-switch
            v-model="formData.is_super"
            :checked-value="1"
            :unchecked-value="0"
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
  const modalTitle = ref('新增角色');

  const formData = reactive({
    id: undefined as number | undefined,
    name: '',
    description: '',
    is_super: 0,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '角色名称', dataIndex: 'name', width: 180 },
    { title: '描述', dataIndex: 'description', ellipsis: true, tooltip: true },
    { title: '类型', slotName: 'is_super', width: 140 },
    { title: '操作', slotName: 'actions', width: 160, fixed: 'right' },
  ];

  const fetchRoleList = async () => {
    try {
      loading.value = true;
      const response = await axios.get('/admin/role/list');
      if (response.data && response.data.data) {
        tableData.value = response.data.data || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取角色列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleAdd = () => {
    modalTitle.value = '新增角色';
    Object.assign(formData, {
      id: undefined,
      name: '',
      description: '',
      is_super: 0,
    });
    modalVisible.value = true;
  };

  const handleEdit = (record: any) => {
    modalTitle.value = '编辑角色';
    Object.assign(formData, {
      id: record.id,
      name: record.name,
      description: record.description || '',
      is_super: record.is_super ?? 0,
    });
    modalVisible.value = true;
  };

  const handleSubmit = async () => {
    if (!formData.name) {
      Message.warning('请输入角色名称');
      return;
    }
    try {
      if (formData.id) {
        await axios.put('/admin/role/update', {
          id: formData.id,
          name: formData.name,
          description: formData.description,
          is_super: formData.is_super,
        });
        Message.success('更新成功');
      } else {
        await axios.post('/admin/role/create', {
          name: formData.name,
          description: formData.description,
          is_super: formData.is_super,
        });
        Message.success('创建成功');
      }
      modalVisible.value = false;
      fetchRoleList();
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    }
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除角色 ${record.name} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/role/${record.id}`);
          Message.success('删除成功');
          fetchRoleList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleRefresh = () => {
    fetchRoleList();
  };

  onMounted(() => {
    fetchRoleList();
  });
</script>

<style scoped lang="less">
  .role-list-container {
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
