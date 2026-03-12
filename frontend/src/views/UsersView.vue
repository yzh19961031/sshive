<!-- frontend/src/views/UsersView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
    </div>
    <n-empty description="用户管理功能开发中" style="margin-top:60px" />
  </div>
</template>

<script setup lang="ts">
import { NEmpty } from 'naive-ui'
</script>

<style scoped>
.page { height: 100%; padding: 20px 24px; }
.page-header { margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
