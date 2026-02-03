<template>
  <div class="settings-list-container">
    <a-card class="search-card" :bordered="false">
      <a-form :model="searchForm" layout="inline">
        <a-form-item label="关键词">
          <a-input
            v-model="searchForm.keyword"
            placeholder="Key/名称/值"
            style="width: 240px"
            allow-clear
          />
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
          <span>系统配置</span>
          <a-space>
            <a-button @click="handleRefresh">
              <template #icon><icon-sync /></template>
              刷新
            </a-button>
            <a-button type="primary" @click="handleBatchSave">
              <template #icon><icon-save /></template>
              保存全部
            </a-button>
          </a-space>
        </div>
      </template>

      <a-table
        :columns="columns"
        :data="filteredData"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #value="{ record }">
          <a-input v-model="record.value" placeholder="请输入值" />
        </template>

        <template #actions="{ record }">
          <a-button type="text" size="small" @click="handleSave(record)">
            <template #icon><icon-save /></template>
            保存
          </a-button>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconSave,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  const loading = ref(false);
  const settingsData = ref<any[]>([]);

  const searchForm = reactive({
    keyword: '',
  });

  const columns = [
    { title: 'Key', dataIndex: 'key', width: 240 },
    { title: '名称', dataIndex: 'name', width: 200 },
    { title: '值', slotName: 'value' },
    { title: '操作', slotName: 'actions', width: 100, fixed: 'right' },
  ];

  const filteredData = computed(() => {
    if (!searchForm.keyword) {
      return settingsData.value;
    }
    const keyword = searchForm.keyword.trim().toLowerCase();
    return settingsData.value.filter((item) => {
      return (
        String(item.key || '')
          .toLowerCase()
          .includes(keyword) ||
        String(item.name || '')
          .toLowerCase()
          .includes(keyword) ||
        String(item.value || '')
          .toLowerCase()
          .includes(keyword)
      );
    });
  });

  const fetchSettings = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/config/list');
      if (response.data && response.data.data) {
        settingsData.value = response.data.data || [];
      }
    } catch (error: any) {
      Message.error(error.message || '获取系统配置失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    // filteredData is reactive, no extra action needed
  };

  const handleReset = () => {
    searchForm.keyword = '';
  };

  const handleRefresh = () => {
    fetchSettings();
  };

  const handleSave = async (record: any) => {
    try {
      await axios.post('/admin/config/update', {
        key: record.key,
        value: String(record.value ?? ''),
      });
      Message.success('保存成功');
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    }
  };

  const handleBatchSave = async () => {
    try {
      const payload: Record<string, string> = {};
      settingsData.value.forEach((item) => {
        if (item.key) {
          payload[item.key] = String(item.value ?? '');
        }
      });
      await axios.post('/admin/config/batch-update', payload);
      Message.success('批量保存成功');
    } catch (error: any) {
      Message.error(error.message || '批量保存失败');
    }
  };

  onMounted(() => {
    fetchSettings();
  });
</script>

<style scoped lang="less">
  .settings-list-container {
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
  }
</style>
