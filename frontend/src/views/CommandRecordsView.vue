<!-- frontend/src/views/CommandRecordsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">命令管控</h2>
    </div>

    <n-data-table
      :columns="columns"
      :data="records"
      :loading="loading"
      :pagination="pagination"
      remote
      @update:page="loadPage"
      :scroll-x="860"
    />

    <!-- 输出详情 Modal -->
    <n-modal v-model:show="detailVisible" preset="card" title="命令输出详情" style="width:600px">
      <pre class="output-pre">{{ detailContent || '（无输出）' }}</pre>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NDataTable, NTag, NButton, NModal, NEllipsis } from 'naive-ui'
import { sessionApi, type CommandItem } from '@/api/session'

const records = ref<CommandItem[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

const detailVisible = ref(false)
const detailContent = ref('')

function openDetail(result: string) {
  detailContent.value = result
  detailVisible.value = true
}

const columns = [
  {
    title: '命令', key: 'command', width: 200, ellipsis: { tooltip: true },
    render: (row: CommandItem) => h('code', { class: 'cmd-code' }, row.command),
  },
  {
    title: '动作', key: 'action', width: 80,
    render: (row: CommandItem) => h(NTag,
      { type: row.action === 'blocked' ? 'error' : 'success', size: 'small', round: true },
      { default: () => row.action === 'blocked' ? '阻断' : '执行' }),
  },
  {
    title: '主机', key: 'host_name', width: 140, ellipsis: { tooltip: true },
    render: (row: CommandItem) => row.host_name || '-',
  },
  {
    title: '用户', key: 'username', width: 100, ellipsis: { tooltip: true },
    render: (row: CommandItem) => row.username || '-',
  },
  {
    title: '执行时间', key: 'created_at', width: 165,
    render: (row: CommandItem) => new Date(row.created_at).toLocaleString('zh-CN', { hour12: false }),
  },
  {
    title: '输出结果', key: 'result', width: 180,
    render: (row: CommandItem) => {
      if (!row.result) return h('span', { style: 'color:var(--text-secondary)' }, '--')
      const preview = row.result.replace(/\n/g, ' ').slice(0, 60)
      const hasMore = row.result.length > 60
      return h('div', { class: 'result-cell', onClick: () => openDetail(row.result) }, [
        h('span', { class: 'result-preview' }, hasMore ? preview + '…' : preview),
      ])
    },
  },
]

async function loadPage(page: number) {
  loading.value = true
  try {
    const res = await sessionApi.listAllCommands({ page, page_size: 20 })
    records.value = res.data.data?.list ?? []
    pagination.value.itemCount = res.data.data?.total ?? 0
    pagination.value.page = page
  } finally {
    loading.value = false
  }
}

onMounted(() => loadPage(1))
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
.cmd-code {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  background: color-mix(in srgb, var(--accent) 8%, transparent);
  color: var(--accent);
  padding: 1px 6px;
  border-radius: 4px;
}
.result-cell {
  cursor: pointer;
  max-width: 180px;
}
.result-preview {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}
.result-cell:hover .result-preview {
  color: var(--accent);
}
.output-pre {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--bg-elevated);
  padding: 12px;
  border-radius: 6px;
  max-height: 400px;
  overflow-y: auto;
  color: var(--text-primary);
  margin: 0;
}
</style>
