<!-- frontend/src/views/AuditSessionsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">审计日志</h2>
    </div>
    <n-data-table :columns="columns" :data="sessions" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage" />

    <!-- 命令历史 Drawer -->
    <n-drawer v-model:show="drawerVisible" :width="420" placement="right">
      <n-drawer-content :title="`命令记录 — ${drawerSessionLabel}`" closable>
        <div v-if="commandsLoading" style="text-align:center;padding:40px 0">
          <n-spin />
        </div>
        <div v-else-if="commands.length === 0" class="empty-tip">暂无命令记录</div>
        <div v-else class="cmd-list">
          <div v-for="cmd in commands" :key="cmd.id" class="cmd-item">
            <span class="cmd-time">{{ formatTime(cmd.created_at) }}</span>
            <code class="cmd-text">{{ cmd.command }}</code>
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>

    <!-- 回放 Modal -->
    <n-modal v-model:show="replayVisible" :mask-closable="true"
      style="width:90vw;height:85vh" preset="card" :title="replayTitle">
      <div v-if="!replayReady" class="empty-tip">录像暂不可用</div>
      <div v-else ref="playerContainer" class="player-wrap" />
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

const sessions = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

// Command drawer state
const drawerVisible = ref(false)
const drawerSessionLabel = ref('')
const commandsLoading = ref(false)
const commands = ref<any[]>([])

// Replay modal state
const replayVisible = ref(false)
const replayTitle = ref('')
const replayReady = ref(false)
const playerContainer = ref<HTMLElement>()

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString('zh-CN', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

async function openCommands(row: any) {
  drawerSessionLabel.value = `${row.host_name || row.host_id} · ${row.username || row.user_id}`
  drawerVisible.value = true
  commandsLoading.value = true
  commands.value = []
  try {
    const res = await sessionApi.getCommands(row.id, { page: 1, page_size: 100 })
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
  AsciinemaPlayer.create(
    sessionApi.getReplayUrl(row.id),
    playerContainer.value,
    { cols: 220, rows: 50, fit: 'width', theme: 'monokai' }
  )
}

const columns = [
  { title: '主机', key: 'host_name', render: (row: any) => row.host_name || String(row.host_id) },
  { title: '用户', key: 'username', render: (row: any) => row.username || String(row.user_id) },
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
      h(NButton, { size: 'tiny', onClick: () => openCommands(row) }, { default: () => '命令' }),
      h(NButton, { size: 'tiny', type: 'primary', onClick: () => openReplay(row) }, { default: () => '回放' }),
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
.empty-tip { color: var(--text-secondary); text-align: center; padding: 40px 0; }
.cmd-list { display: flex; flex-direction: column; gap: 10px; }
.cmd-item { display: flex; flex-direction: column; gap: 2px; }
.cmd-time { font-size: 11px; color: var(--text-secondary); }
.cmd-text { font-family: monospace; font-size: 13px; background: var(--bg-elevated); padding: 4px 8px; border-radius: 4px; word-break: break-all; white-space: pre-wrap; }
.player-wrap { height: calc(85vh - 80px); }
</style>
