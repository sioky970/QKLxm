<template>
  <div class="placeholder-container">
    <a-result status="info" :title="title || '页面开发中'">
      <template #subtitle>
        {{ subtitle || '该功能正在开发中，敬请期待' }}
      </template>
      <template #extra>
        <a-space>
          <a-button type="primary" @click="goBack">返回</a-button>
          <a-button @click="goHome">返回首页</a-button>
        </a-space>
      </template>
    </a-result>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router';

  defineProps<{
    title?: string;
    subtitle?: string;
  }>();

  const router = useRouter();

  const goBack = () => {
    router.back();
  };

  const goHome = () => {
    router.push('/admin/dashboard');
  };
</script>

<style scoped lang="less">
  .placeholder-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400px;
  }
</style>
