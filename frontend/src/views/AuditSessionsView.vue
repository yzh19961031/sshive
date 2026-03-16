<!-- frontend/src/views/AuditSessionsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">审计日志</h2>
    </div>
    <n-data-table :columns="columns" :data="sessions" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage"
      :scroll-x="820" />

    <!-- 操作记录 Drawer -->
    <n-drawer v-model:show="drawerVisible" :width="760" placement="right">
      <n-drawer-content closable>
        <template #header>
          <span class="drawer-title">操作记录</span>
          <span class="drawer-sub">{{ drawerSessionLabel }}</span>
        </template>

        <!-- 过滤栏 -->
        <div class="filter-bar">
          <n-input
            v-model:value="filterCommand"
            placeholder="请输入命令"
            clearable
            size="small"
            style="width:180px"
            @keyup.enter="applyFilter"
          />
          <n-select
            v-model:value="filterAction"
            :options="actionOptions"
            placeholder="请选择动作"
            clearable
            size="small"
            style="width:130px"
          />
          <n-date-picker
            v-model:value="filterDateRange"
            type="datetimerange"
            clearable
            size="small"
            style="width:360px"
            :shortcuts="dateShortcuts"
          />
          <n-button type="primary" size="small" @click="applyFilter">查询</n-button>
          <n-button size="small" @click="resetFilter">重置</n-button>
        </div>

        <!-- 命令表格 -->
        <n-data-table
          :columns="cmdColumns"
          :data="cmdList"
          :loading="cmdLoading"
          :pagination="cmdPagination"
          remote
          @update:page="loadCmdPage"
          :scroll-x="680"
          size="small"
          style="margin-top:12px"
        />
      </n-drawer-content>
    </n-drawer>

    <!-- 输出详情 Modal -->
    <n-modal v-model:show="outputVisible" preset="card" title="命令输出" style="width:560px">
      <pre class="output-pre">{{ outputContent || '（无输出）' }}</pre>
    </n-modal>

    <!-- 回放 Modal -->
    <n-modal v-model:show="replayVisible" :mask-closable="true" display-directive="show">
      <div class="replay-dialog">
        <div class="replay-header">
          <span class="replay-title">{{ replayTitle }}</span>
          <button class="replay-close" @click="replayVisible = false">✕</button>
        </div>
        <div v-if="!replayReady" class="empty-tip" style="padding:60px 0">录像暂不可用</div>
        <div v-else ref="playerContainer" class="player-wrap" />
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, nextTick } from 'vue'
import {
  NDataTable, NButton, NTag, NDrawer, NDrawerContent,
  NModal, NInput, NSelect, NDatePicker,
} from 'naive-ui'
// @ts-ignore
import * as AsciinemaPlayer from 'asciinema-player'
import 'asciinema-player/dist/bundle/asciinema-player.css'
import { sessionApi, type CommandItem } from '@/api/session'

// ── 会话列表 ───────────────────────────────────────────────
const sessions = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

function statusTag(status: string): 'success' | 'warning' | 'default' | 'error' {
  return ({ active: 'success', interrupted: 'warning', closed: 'default' } as any)[status] ?? 'default'
}
function formatDuration(row: any): string {
  const end = row.ended_at ? new Date(row.ended_at) : new Date()
  const diff = Math.floor((end.getTime() - new Date(row.started_at).getTime()) / 1000)
  if (diff < 60) return `${diff}s`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ${diff % 60}s`
  return `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`
}

// ── 操作记录 Drawer ────────────────────────────────────────
const drawerVisible = ref(false)
const drawerSessionLabel = ref('')
const drawerSessionId = ref(0)

// 过滤条件
const filterCommand = ref('')
const filterAction = ref<string | null>(null)
const filterDateRange = ref<[number, number] | null>(null)

const actionOptions = [
  { label: '执行', value: 'execute' },
  { label: '阻断', value: 'blocked' },
]
const dateShortcuts = {
  '今天': () => {
    const s = new Date(); s.setHours(0, 0, 0, 0)
    const e = new Date(); e.setHours(23, 59, 59, 999)
    return [s.getTime(), e.getTime()] as [number, number]
  },
  '近 7 天': () => {
    const e = new Date()
    const s = new Date(e.getTime() - 7 * 86400_000)
    return [s.getTime(), e.getTime()] as [number, number]
  },
}

// 命令表格
const cmdList = ref<CommandItem[]>([])
const cmdLoading = ref(false)
const cmdPagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

// 输出详情
const outputVisible = ref(false)
const outputContent = ref('')

const cmdColumns = [
  {
    title: '命令', key: 'command', width: 200, ellipsis: { tooltip: true },
    render: (row: CommandItem) => h('code', { class: 'cmd-code' }, row.command),
  },
  {
    title: '动作', key: 'action', width: 72,
    render: (row: CommandItem) => h(NTag,
      { type: row.action === 'blocked' ? 'error' : 'success', size: 'small', round: true },
      { default: () => row.action === 'blocked' ? '阻断' : '执行' }),
  },
  {
    title: '执行时间', key: 'created_at', width: 160,
    render: (row: CommandItem) => new Date(row.created_at).toLocaleString('zh-CN', { hour12: false }),
  },
  {
    title: '输出结果', key: 'result',
    render: (row: CommandItem) => {
      if (!row.result) return h('span', { style: 'color:var(--text-secondary)' }, '--')
      const preview = row.result.replace(/\n/g, ' ').slice(0, 50)
      const hasMore = row.result.length > 50
      return h('span', {
        class: 'result-link',
        onClick: () => { outputContent.value = row.result; outputVisible.value = true },
      }, hasMore ? preview + '…' : preview)
    },
  },
]

function buildFilterParams() {
  const params: Record<string, any> = {}
  if (filterCommand.value) params.command = filterCommand.value
  if (filterAction.value) params.action = filterAction.value
  if (filterDateRange.value) {
    params.start_time = new Date(filterDateRange.value[0]).toISOString()
    params.end_time = new Date(filterDateRange.value[1]).toISOString()
  }
  return params
}

async function loadCmdPage(page: number) {
  cmdLoading.value = true
  try {
    const res = await sessionApi.getCommands(drawerSessionId.value, {
      page, page_size: 20, ...buildFilterParams(),
    })
    cmdList.value = res.data.data?.list ?? []
    cmdPagination.value.itemCount = res.data.data?.total ?? 0
    cmdPagination.value.page = page
  } finally {
    cmdLoading.value = false
  }
}

function applyFilter() { loadCmdPage(1) }
function resetFilter() {
  filterCommand.value = ''
  filterAction.value = null
  filterDateRange.value = null
  loadCmdPage(1)
}

function openCommands(row: any) {
  drawerSessionId.value = row.id
  drawerSessionLabel.value = `${row.host_name || row.host_id} · ${row.username || row.user_id}`
  filterCommand.value = ''
  filterAction.value = null
  filterDateRange.value = null
  cmdList.value = []
  drawerVisible.value = true
  loadCmdPage(1)
}

// ── 回放 Modal ────────────────────────────────────────────
const replayVisible = ref(false)
const replayTitle = ref('')
const replayReady = ref(false)
const playerContainer = ref<HTMLElement>()
let playerInstance: any = null

async function openReplay(row: any) {
  replayTitle.value = `${row.host_name || row.host_id} · ${row.username || row.user_id} · ${new Date(row.started_at).toLocaleString()}`
  replayReady.value = !!row.cast_file_path
  replayVisible.value = true
  if (!replayReady.value) return
  await nextTick()
  if (!playerContainer.value) return
  playerContainer.value.innerHTML = ''
  if (playerInstance?.dispose) playerInstance.dispose()
  playerInstance = AsciinemaPlayer.create(
    sessionApi.getReplayUrl(row.id),
    playerContainer.value,
    { cols: 160, rows: 40, fit: 'width', theme: 'monokai', autoPlay: true }
  )
}

// ── 会话列表列定义 ─────────────────────────────────────────
const columns = [
  { title: '主机', key: 'host_name', width: 160, ellipsis: { tooltip: true },
    render: (row: any) => row.host_name || String(row.host_id) },
  { title: '用户', key: 'username', width: 90, ellipsis: { tooltip: true },
    render: (row: any) => row.username || String(row.user_id) },
  { title: '客户端 IP', key: 'client_ip', width: 110, ellipsis: true },
  {
    title: '状态', key: 'status', width: 80,
    render: (row: any) => h(NTag,
      { type: statusTag(row.status), size: 'small', round: true },
      { default: () => row.status })
  },
  { title: '时长', key: 'duration', width: 75,
    render: (row: any) => formatDuration(row) },
  { title: '开始时间', key: 'started_at', width: 165,
    render: (row: any) => new Date(row.started_at).toLocaleString('zh-CN', { hour12: false }) },
  {
    title: '操作', key: 'actions', width: 120, fixed: 'right' as const,
    render: (row: any) => h('div', { style: 'display:flex;gap:4px' }, [
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => openCommands(row) },
        { default: () => '📋 操作记录' }),
      h(NButton, { size: 'tiny', quaternary: true, type: 'primary',
        disabled: !row.cast_file_path, onClick: () => openReplay(row) },
        { default: () => '▶ 回放' }),
    ])
  },
]

async function loadPage(page: number) {
  loading.value = true
  try {
    const res = await sessionApi.list({ page, page_size: 20 })
    sessions.value = res.data.data?.list ?? []
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

.drawer-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.drawer-sub { font-size: 12px; color: var(--text-secondary); margin-left: 8px; }

.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 0 0 4px;
  border-bottom: 1px solid var(--border);
  padding-bottom: 12px;
}

.cmd-code {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  background: color-mix(in srgb, var(--accent) 8%, transparent);
  color: var(--accent);
  padding: 1px 6px;
  border-radius: 4px;
  white-space: nowrap;
}
.result-link {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
  max-width: 220px;
}
.result-link:hover { color: var(--accent); }

.output-pre {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--bg-elevated);
  padding: 12px;
  border-radius: 6px;
  max-height: 420px;
  overflow-y: auto;
  color: var(--text-primary);
  margin: 0;
}

.empty-tip { color: var(--text-secondary); text-align: center; padding: 40px 0; }

.replay-dialog {
  width: min(90vw, 1000px);
  background: #1a1a2e;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 24px 80px rgba(0,0,0,0.6);
}
.replay-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #16213e;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.replay-title {
  font-size: 13px;
  color: rgba(255,255,255,0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: calc(100% - 40px);
}
.replay-close {
  background: rgba(255,255,255,0.08);
  border: none;
  color: rgba(255,255,255,0.6);
  width: 28px; height: 28px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  flex-shrink: 0;
  transition: background 0.15s;
}
.replay-close:hover { background: rgba(255,255,255,0.15); color: #fff; }
.player-wrap { height: min(65vh, 600px); }
</style>
