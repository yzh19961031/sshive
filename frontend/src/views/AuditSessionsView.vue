<!-- frontend/src/views/AuditSessionsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">审计日志</h2>
    </div>
    <n-data-table :columns="columns" :data="sessions" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage" />
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDataTable, NButton, NTag } from 'naive-ui'
import { sessionApi } from '@/api/session'

const router = useRouter()
const sessions = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '主机 ID', key: 'host_id', width: 100 },
  { title: '用户 ID', key: 'user_id', width: 100 },
  { title: '客户端 IP', key: 'client_ip', width: 140 },
  {
    title: '状态', key: 'status', width: 80,
    render: (row: any) => h(NTag, { type: row.status === 'active' ? 'success' : 'default', size: 'small' },
      { default: () => row.status })
  },
  { title: '开始时间', key: 'started_at', render: (row: any) => new Date(row.started_at).toLocaleString() },
  {
    title: '操作', key: 'actions',
    render: (row: any) => h('div', { style: 'display:flex;gap:6px' }, [
      h(NButton, { size: 'tiny', onClick: () => router.push(`/audit/replay/${row.id}`) }, { default: () => '回放' }),
    ])
  },
]

async function loadPage(page: number) {
  loading.value = true
  const res = await sessionApi.list({ page, page_size: 20 })
  sessions.value = res.data.data?.list ?? []
  pagination.value.itemCount = res.data.data?.total ?? 0
  pagination.value.page = page
  loading.value = false
}

onMounted(() => loadPage(1))
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
