<template>
  <div class="admin-list-container">
    <a-card class="table-card" :bordered="false">
      <template #title>
        <div class="table-header">
          <span>管理员列表</span>
          <a-space>
            <a-button type="primary" @click="handleAdd">
              <template #icon><icon-plus /></template>
              新增管理员
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
        :pagination="pagination"
        :loading="loading"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #role="{ record }">
          {{ roleMap[record.role_id] || `角色 ${record.role_id}` }}
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
        <a-form-item label="用户名" required>
          <a-input v-model="formData.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item label="角色" required>
          <a-select v-model="formData.role_id" placeholder="请选择角色">
            <a-option v-for="role in roleList" :key="role.id" :value="role.id">
              {{ role.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item
          :label="formData.id ? '新密码' : '密码'"
          :required="!formData.id"
        >
          <a-input-password
            v-model="formData.password"
            placeholder="不修改请留空"
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
  const roleList = ref<any[]>([]);
  const roleMap = reactive<Record<number, string>>({});
  const modalVisible = ref(false);
  const modalTitle = ref('新增管理员');

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const formData = reactive({
    id: undefined as number | undefined,
    username: '',
    role_id: undefined as number | undefined,
    password: '',
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户名', dataIndex: 'username', width: 180 },
    { title: '角色', slotName: 'role', width: 180 },
    { title: '操作', slotName: 'actions', width: 160, fixed: 'right' },
  ];

  const fetchRoles = async () => {
    try {
      const response = await axios.get('/admin/role/list');
      if (response.data && response.data.data) {
        roleList.value = response.data.data || [];
        roleList.value.forEach((role: any) => {
          roleMap[role.id] = role.name;
        });
      }
    } catch (error: any) {
      Message.error(error.message || '获取角色列表失败');
    }
  };

  const fetchAdminList = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/admin/list', {
        page: pagination.current,
        page_size: pagination.pageSize,
      });
      if (response.data && response.data.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取管理员列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleAdd = () => {
    modalTitle.value = '新增管理员';
    Object.assign(formData, {
      id: undefined,
      username: '',
      role_id: roleList.value[0]?.id,
      password: '',
    });
    modalVisible.value = true;
  };

  const handleEdit = (record: any) => {
    modalTitle.value = '编辑管理员';
    Object.assign(formData, {
      id: record.id,
      username: record.username,
      role_id: record.role_id,
      password: '',
    });
    modalVisible.value = true;
  };

  const handleSubmit = async () => {
    if (!formData.username || !formData.role_id) {
      Message.warning('请填写完整信息');
      return;
    }
    if (!formData.id && !formData.password) {
      Message.warning('请设置密码');
      return;
    }

    try {
      if (formData.id) {
        await axios.put('/admin/admin/update', {
          id: formData.id,
          username: formData.username,
          role_id: formData.role_id,
          password: formData.password || undefined,
        });
        Message.success('更新成功');
      } else {
        await axios.post('/admin/admin/create', {
          username: formData.username,
          role_id: formData.role_id,
          password: formData.password,
        });
        Message.success('创建成功');
      }
      modalVisible.value = false;
      fetchAdminList();
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    }
  };

  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除管理员 ${record.username} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/admin/${record.id}`);
          Message.success('删除成功');
          fetchAdminList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleRefresh = () => {
    fetchAdminList();
  };

  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchAdminList();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchAdminList();
  };

  onMounted(() => {
    fetchRoles();
    fetchAdminList();
  });
</script>

<style scoped lang="less">
  .admin-list-container {
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
