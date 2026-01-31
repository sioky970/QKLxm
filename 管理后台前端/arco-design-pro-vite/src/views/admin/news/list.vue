<template>
  <div class="news-list-container">
    <a-card class="table-card" :bordered="false">
      <a-tabs v-model:active-key="activeTab">
        <a-tab-pane key="news" title="新闻列表">
          <div class="table-header">
            <span>新闻列表</span>
            <a-space>
              <a-button type="primary" @click="handleNewsAdd">
                <template #icon><icon-plus /></template>
                新增新闻
              </a-button>
              <a-button @click="handleNewsRefresh">
                <template #icon><icon-sync /></template>
                刷新
              </a-button>
            </a-space>
          </div>

          <a-table
            :columns="newsColumns"
            :data="newsData"
            :pagination="newsPagination"
            :loading="newsLoading"
            :scroll="{ x: 1300 }"
            row-key="id"
            @page-change="handleNewsPageChange"
            @page-size-change="handleNewsPageSizeChange"
          >
            <template #category="{ record }">
              {{ categoryMap[record.category_id] || '-' }}
            </template>

            <template #status="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'orange'">
                {{ record.status === 1 ? '已发布' : '草稿' }}
              </a-tag>
            </template>

            <template #publish_time="{ record }">
              {{ formatTime(record.publish_time) }}
            </template>

            <template #create_time="{ record }">
              {{ formatTime(record.create_time) }}
            </template>

            <template #actions="{ record }">
              <a-space>
                <a-button
                  type="text"
                  size="small"
                  @click="handleNewsEdit(record)"
                >
                  <template #icon><icon-edit /></template>
                  编辑
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  @click="
                    record.status === 1
                      ? handleNewsDraft(record)
                      : handleNewsPublish(record)
                  "
                >
                  <template #icon><icon-send /></template>
                  {{ record.status === 1 ? '设为草稿' : '发布' }}
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  status="danger"
                  @click="handleNewsDelete(record)"
                >
                  <template #icon><icon-delete /></template>
                  删除
                </a-button>
              </a-space>
            </template>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="category" title="分类管理">
          <div class="table-header">
            <span>新闻分类</span>
            <a-space>
              <a-button type="primary" @click="handleCategoryAdd">
                <template #icon><icon-plus /></template>
                新增分类
              </a-button>
              <a-button @click="handleCategoryRefresh">
                <template #icon><icon-sync /></template>
                刷新
              </a-button>
            </a-space>
          </div>

          <a-table
            :columns="categoryColumns"
            :data="categoryData"
            :loading="categoryLoading"
            :pagination="false"
            row-key="id"
          >
            <template #status="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'red'">
                {{ record.status === 1 ? '启用' : '禁用' }}
              </a-tag>
            </template>

            <template #actions="{ record }">
              <a-space>
                <a-button
                  type="text"
                  size="small"
                  @click="handleCategoryEdit(record)"
                >
                  <template #icon><icon-edit /></template>
                  编辑
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  status="danger"
                  @click="handleCategoryDelete(record)"
                >
                  <template #icon><icon-delete /></template>
                  删除
                </a-button>
              </a-space>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <a-modal
      v-model:visible="newsModalVisible"
      :title="newsModalTitle"
      :width="720"
      @ok="handleNewsSubmit"
      @cancel="newsModalVisible = false"
    >
      <a-form :model="newsForm" :label-col-props="{ span: 6 }">
        <a-form-item label="标题" required>
          <a-input v-model="newsForm.title" placeholder="请输入新闻标题" />
        </a-form-item>
        <a-form-item label="分类" required>
          <a-select v-model="newsForm.category_id" placeholder="请选择分类">
            <a-option
              v-for="category in categoryData"
              :key="category.id"
              :value="category.id"
            >
              {{ category.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="作者">
          <a-input v-model="newsForm.author" placeholder="请输入作者" />
        </a-form-item>
        <a-form-item label="内容" required>
          <a-textarea
            v-model="newsForm.content"
            :auto-size="{ minRows: 6, maxRows: 12 }"
            placeholder="请输入新闻内容"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:visible="categoryModalVisible"
      :title="categoryModalTitle"
      :width="520"
      @ok="handleCategorySubmit"
      @cancel="categoryModalVisible = false"
    >
      <a-form :model="categoryForm" :label-col-props="{ span: 6 }">
        <a-form-item label="分类名称" required>
          <a-input v-model="categoryForm.name" placeholder="请输入分类名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model="categoryForm.description"
            placeholder="请输入描述"
            :max-length="200"
            show-word-limit
          />
        </a-form-item>
        <a-form-item label="状态">
          <a-switch
            v-model="categoryForm.status"
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
    IconSend,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const activeTab = ref('news');
  const newsLoading = ref(false);
  const categoryLoading = ref(false);
  const newsData = ref<any[]>([]);
  const categoryData = ref<any[]>([]);
  const categoryMap = reactive<Record<number, string>>({});

  const newsModalVisible = ref(false);
  const newsModalTitle = ref('新增新闻');
  const categoryModalVisible = ref(false);
  const categoryModalTitle = ref('新增分类');

  const newsPagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  const newsForm = reactive({
    id: undefined as number | undefined,
    title: '',
    category_id: undefined as number | undefined,
    author: '',
    content: '',
  });

  const categoryForm = reactive({
    id: undefined as number | undefined,
    name: '',
    description: '',
    status: 1,
  });

  const newsColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    {
      title: '标题',
      dataIndex: 'title',
      width: 240,
      ellipsis: true,
      tooltip: true,
    },
    { title: '分类', slotName: 'category', width: 160 },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '浏览量', dataIndex: 'view_count', width: 100 },
    { title: '发布时间', slotName: 'publish_time', width: 180 },
    { title: '创建时间', slotName: 'create_time', width: 180 },
    { title: '操作', slotName: 'actions', width: 220, fixed: 'right' },
  ];

  const categoryColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '分类名称', dataIndex: 'name', width: 200 },
    { title: '描述', dataIndex: 'description', ellipsis: true, tooltip: true },
    { title: '状态', slotName: 'status', width: 120 },
    { title: '操作', slotName: 'actions', width: 160, fixed: 'right' },
  ];

  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const fetchCategories = async () => {
    try {
      categoryLoading.value = true;
      const response = await axios.get('/admin/news-category/list');
      if (response.data && response.data.data) {
        categoryData.value = response.data.data || [];
        categoryData.value.forEach((item: any) => {
          categoryMap[item.id] = item.name;
        });
      }
    } catch (error: any) {
      Message.error(error.message || '获取分类失败');
    } finally {
      categoryLoading.value = false;
    }
  };

  const fetchNewsList = async () => {
    try {
      newsLoading.value = true;
      const response = await axios.post('/admin/news/list', {
        page: newsPagination.current,
        page_size: newsPagination.pageSize,
      });
      if (response.data && response.data.data) {
        newsData.value = response.data.data.list || [];
        newsPagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取新闻列表失败');
    } finally {
      newsLoading.value = false;
    }
  };

  const handleNewsAdd = () => {
    newsModalTitle.value = '新增新闻';
    Object.assign(newsForm, {
      id: undefined,
      title: '',
      category_id: categoryData.value[0]?.id,
      author: '',
      content: '',
    });
    newsModalVisible.value = true;
  };

  const handleNewsEdit = (record: any) => {
    newsModalTitle.value = '编辑新闻';
    Object.assign(newsForm, {
      id: record.id,
      title: record.title,
      category_id: record.category_id,
      author: record.author || '',
      content: record.content || '',
    });
    newsModalVisible.value = true;
  };

  const handleNewsSubmit = async () => {
    if (!newsForm.title || !newsForm.content) {
      Message.warning('请填写标题和内容');
      return;
    }
    try {
      const payload = {
        title: newsForm.title,
        category_id: newsForm.category_id,
        author: newsForm.author,
        content: newsForm.content,
      };
      if (newsForm.id) {
        await axios.put(`/admin/news/${newsForm.id}`, payload);
        Message.success('更新成功');
      } else {
        await axios.post('/admin/news/create', payload);
        Message.success('创建成功');
      }
      newsModalVisible.value = false;
      fetchNewsList();
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    }
  };

  const handleNewsPublish = async (record: any) => {
    try {
      await axios.post('/admin/news/publish', { id: record.id });
      Message.success('发布成功');
      fetchNewsList();
    } catch (error: any) {
      Message.error(error.message || '发布失败');
    }
  };

  const handleNewsDraft = async (record: any) => {
    try {
      await axios.post('/admin/news/draft', { id: record.id });
      Message.success('已设为草稿');
      fetchNewsList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    }
  };

  const handleNewsDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除新闻 ${record.title} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/news/${record.id}`);
          Message.success('删除成功');
          fetchNewsList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleNewsRefresh = () => {
    fetchNewsList();
  };

  const handleNewsPageChange = (page: number) => {
    newsPagination.current = page;
    fetchNewsList();
  };

  const handleNewsPageSizeChange = (pageSize: number) => {
    newsPagination.pageSize = pageSize;
    newsPagination.current = 1;
    fetchNewsList();
  };

  const handleCategoryAdd = () => {
    categoryModalTitle.value = '新增分类';
    Object.assign(categoryForm, {
      id: undefined,
      name: '',
      description: '',
      status: 1,
    });
    categoryModalVisible.value = true;
  };

  const handleCategoryEdit = (record: any) => {
    categoryModalTitle.value = '编辑分类';
    Object.assign(categoryForm, {
      id: record.id,
      name: record.name,
      description: record.description || '',
      status: record.status ?? 1,
    });
    categoryModalVisible.value = true;
  };

  const handleCategorySubmit = async () => {
    if (!categoryForm.name) {
      Message.warning('请输入分类名称');
      return;
    }
    try {
      const payload = {
        name: categoryForm.name,
        description: categoryForm.description,
        status: categoryForm.status,
      };
      if (categoryForm.id) {
        await axios.put(`/admin/news-category/${categoryForm.id}`, payload);
        Message.success('更新成功');
      } else {
        await axios.post('/admin/news-category/create', payload);
        Message.success('创建成功');
      }
      categoryModalVisible.value = false;
      fetchCategories();
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    }
  };

  const handleCategoryDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除分类 ${record.name} 吗？`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/news-category/${record.id}`);
          Message.success('删除成功');
          fetchCategories();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleCategoryRefresh = () => {
    fetchCategories();
  };

  onMounted(() => {
    fetchCategories();
    fetchNewsList();
  });
</script>

<style scoped lang="less">
  .news-list-container {
    padding: 20px;

    .table-card {
      :deep(.arco-card-body) {
        padding: 20px;
      }
    }

    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 16px;
      font-weight: 500;
      margin-bottom: 16px;
    }
  }
</style>
