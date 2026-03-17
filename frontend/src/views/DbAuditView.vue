<!-- frontend/src/views/DbAuditView.vue -->
<template>
  <div style="padding: 16px; height: 100%; display: flex; flex-direction: column; gap: 12px;">
    <div style="display: flex; align-items: center; justify-content: space-between; flex-shrink: 0;">
      <h2 style="font-size: 16px; font-weight: 600;">DB 查询审计</h2>
    </div>

    <!-- 过滤条件 -->
    <div style="display: flex; gap: 10px; flex-wrap: wrap; flex-shrink: 0;">
      <n-select
        v-model:value="filterServerId"
        :options="serverOptions"
        placeholder="全部服务器"
        clearable
        size="small"
        style="width: 180px"
      />
      <n-date-picker
        v-model:value="dateRange"
        type="daterange"
        size="small"
        clearable
        placeholder="时间范围"
      />
      <n-button size="small" type="primary" @click="load">查询</n-button>
    </div>

    <!-- 数据表格 -->
    <n-data-table
      :columns="columns"
      :data="logs"
      :loading="loading"
      :pagination="pagination"
      :row-key="(row: DBQueryLog) => row.id"
      :expanded-row-keys="expandedKeys"
      @update:expanded-row-keys="(keys: DataTableRowKey[]) => { expandedKeys = keys }"
      size="small"
      flex-height
      style="flex: 1; min-height: 0;"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { NDataTable, NSelect, NDatePicker, NButton, NTag } from 'naive-ui'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'
import { dbApi, type DBServer, type DBQueryLog } from '@/api/db'

const servers      = ref<DBServer[]>([])
const logs         = ref<DBQueryLog[]>([])
const total        = ref(0)
const loading      = ref(false)
const filterServerId = ref<number | null>(null)
const dateRange    = ref<[number, number] | null>(null)
const expandedKeys = ref<DataTableRowKey[]>([])
const page         = ref(1)
const pageSize     = ref(20)

const serverOptions = computed(() =>
  servers.value.map(s => ({ label: `${s.name} (${s.type})`, value: s.id }))
)

const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  onChange: (p: number) => { page.value = p; load() },
  onUpdatePageSize: (ps: number) => { pageSize.value = ps; page.value = 1; load() },
}))

const columns: DataTableColumns<DBQueryLog> = [
  {
    type: 'expand' as const,
    expandable: () => true,
    renderExpand: (row: DBQueryLog) =>
      h('pre', { style: 'margin:0;padding:8px;font-size:11px;white-space:pre-wrap;word-break:break-all;' }, row.sql)
  },
  { title: '时间', key: 'created_at', width: 160,
    render: (row: DBQueryLog) => new Date(row.created_at).toLocaleString() },
  { title: '数据库', key: 'database', width: 120 },
  { title: 'SQL', key: 'sql', ellipsis: { tooltip: false },
    render: (row: DBQueryLog) => row.sql.slice(0, 80) + (row.sql.length > 80 ? '…' : '') },
  { title: '耗时(ms)', key: 'duration_ms', width: 90 },
  { title: '行数', key: 'rows_returned', width: 70 },
  { title: '状态', key: 'error_msg', width: 80,
    render: (row: DBQueryLog) =>
      row.error_msg
        ? h(NTag, { type: 'error', size: 'small' }, () => '失败')
        : h(NTag, { type: 'success', size: 'small' }, () => '成功')
  },
]

async function load() {
  loading.value = true
  try {
    const params: Record<string, any> = { page: page.value, page_size: pageSize.value }
    if (filterServerId.value) params.db_server_id = filterServerId.value
    if (dateRange.value) {
      params.start_time = new Date(dateRange.value[0]).toISOString()
      params.end_time   = new Date(dateRange.value[1] + 86400000).toISOString()
    }
    const res = await dbApi.listQueryLogs(params)
    logs.value  = res.data.data?.list ?? []
    total.value = res.data.data?.total ?? 0
  } finally { loading.value = false }
}

onMounted(async () => {
  const res = await dbApi.list()
  servers.value = res.data.data ?? []
  await load()
})
</script>
