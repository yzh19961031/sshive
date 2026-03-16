<!-- frontend/src/views/AuditSessionsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">审计日志</h2>
    </div>
    <n-data-table :columns="columns" :data="sessions" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage"
      :scroll-x="820" />

    <!-- 命令历史 Drawer -->
    <n-drawer v-model:show="drawerVisible" :width="420" placement="right">
      <n-drawer-content :title="`命令记录 — ${drawerSessionLabel}`" closable>
        <div v-if="commandsLoading" style="text-align:center;padding:40px 0">
          <n-spin />
        </div>
        <div v-else-if="commands.length === 0" class="empty-tip">暂无日志记录</div>
        <div v-else class="cmd-list">
          <div v-for="log in commands" :key="log.id"
            :class="['log-item', log.type === 'input' ? 'log-input' : 'log-output']">
            <span class="log-time">{{ formatTime(log.created_at) }}</span>
            <code class="log-content">
              <span v-if="log.type === 'input'" class="log-prefix">❯ </span>{{ log.content }}
            </code>
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>

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
import { NDataTable, NButton, NTag, NDrawer, NDrawerContent, NSpin, NModal } from 'naive-ui'
// @ts-ignore
import * as AsciinemaPlayer from 'asciinema-player'
import 'asciinema-player/dist/bundle/asciinema-player.css'
import { sessionApi } from '@/api/session'
import type { LogItem } from '@/api/session'

const sessions = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

// Command drawer state
const drawerVisible = ref(false)
const drawerSessionLabel = ref('')
const commandsLoading = ref(false)
const commands = ref<LogItem[]>([])

// Replay modal state
const replayVisible = ref(false)
const replayTitle = ref('')
const replayReady = ref(false)
const playerContainer = ref<HTMLElement>()
let playerInstance: any = null

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString('zh-CN', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function statusTag(status: string): 'success' | 'warning' | 'default' | 'error' {
  const map: Record<string, 'success' | 'warning' | 'default' | 'error'> = {
    active: 'success',
    interrupted: 'warning',
    closed: 'default',
  }
  return map[status] ?? 'default'
}

function formatDuration(row: any): string {
  const end = row.ended_at ? new Date(row.ended_at) : new Date()
  const diff = Math.floor((end.getTime() - new Date(row.started_at).getTime()) / 1000)
  if (diff < 60) return `${diff}s`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ${diff % 60}s`
  return `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`
}

async function openCommands(row: any) {
  drawerSessionLabel.value = `${row.host_name || row.host_id} · ${row.username || row.user_id}`
  drawerVisible.value = true
  commandsLoading.value = true
  commands.value = []
  try {
    const res = await sessionApi.getLogs(row.id, { page: 1, page_size: 200 })
    commands.value = res.data.data?.list ?? []
  } finally {
    commandsLoading.value = false
  }
}

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
        { default: () => '📋 日志' }),
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
.empty-tip { color: var(--text-secondary); text-align: center; padding: 40px 0; }
.cmd-list { display: flex; flex-direction: column; gap: 10px; }
.log-item { display: flex; flex-direction: column; gap: 1px; margin-bottom: 6px; }
.log-time { font-size: 11px; color: var(--text-secondary); }
.log-content { font-family: monospace; font-size: 12px; padding: 3px 8px; border-radius: 4px;
  word-break: break-all; white-space: pre-wrap; display: block; }
.log-input .log-content { background: color-mix(in srgb, var(--accent) 8%, transparent); color: var(--accent); }
.log-output .log-content { background: var(--bg-elevated); color: var(--text-primary); }
.log-prefix { opacity: 0.6; }
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
  width: 28px;
  height: 28px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  flex-shrink: 0;
  transition: background 0.15s;
}
.replay-close:hover { background: rgba(255,255,255,0.15); color: #fff; }
.player-wrap { height: min(65vh, 600px); }
</style>
