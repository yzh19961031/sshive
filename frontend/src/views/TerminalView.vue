<!-- frontend/src/views/TerminalView.vue -->
<template>
  <div class="terminal-page">
    <!-- Tab bar -->
    <div class="tab-bar">
      <div v-for="tab in tabs" :key="tab.id"
        :class="['tab', tab.id === activeTab && 'active']"
        @click="activeTab = tab.id">
        <div :class="['tab-dot', tab.connected ? 'online' : 'offline']" />
        <span>{{ tab.name }}</span>
        <button class="tab-close" @click.stop="closeTab(tab.id)">×</button>
      </div>
      <div class="tab-actions">
        <n-select
          :value="termTheme.currentId"
          :options="themeOptions"
          size="small"
          style="width:140px"
          @update:value="applyTerminalTheme"
        />
        <button class="split-btn" @click="copySelection" title="复制选中内容">⎘</button>
        <button class="split-btn" @click="showHostPicker = true" title="新开终端">＋</button>
        <button class="split-btn" :class="{ active: splitMode }" @click="toggleSplit" title="分屏">⊞</button>
      </div>
    </div>

    <!-- Terminal area -->
    <div :class="['terminal-area', splitMode && 'split']">
      <div v-for="(tab, i) in displayTabs" :key="tab.id"
        class="terminal-pane"
        :ref="el => setPaneRef(el as HTMLElement, i)">
      </div>
    </div>

    <!-- 主机选择 Modal -->
    <n-modal v-model:show="showHostPicker" preset="card" title="选择主机" style="width:480px">
      <div v-if="hostPickerLoading" style="text-align:center;padding:20px"><n-spin /></div>
      <div v-else class="host-picker-list">
        <div v-for="h in pickerHosts" :key="h.id"
          class="picker-row" @click="pickHost(h)">
          <span class="picker-dot" />
          <span class="picker-name">{{ h.name }}</span>
          <span class="picker-ip">{{ h.ip }}:{{ h.port }}</span>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { NSelect, NModal, NSpin } from 'naive-ui'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { useAuthStore } from '@/stores/auth'
import { hostApi } from '@/api/host'
import { useTerminalThemeStore } from '@/stores/terminalTheme'

const auth = useAuthStore()
const termTheme = useTerminalThemeStore()
const themeOptions = termTheme.themes.map(t => ({ label: t.name, value: t.id }))

interface Tab {
  id: string
  name: string
  hostId: number
  connected: boolean
  term?: Terminal
  fit?: FitAddon
  ws?: WebSocket
}

const tabs = ref<Tab[]>([])
const activeTab = ref('')
const splitMode = ref(false)
const paneRefs: HTMLElement[] = []

// In split mode show up to 4 tabs, otherwise only the active tab
const displayTabs = computed(() =>
  splitMode.value ? tabs.value.slice(0, 4) : tabs.value.filter(t => t.id === activeTab.value)
)

function setPaneRef(el: HTMLElement | null, i: number) {
  if (el) paneRefs[i] = el
}

function applyTerminalTheme(id: string) {
  termTheme.apply(id)
  const newTheme = termTheme.current().theme
  tabs.value.forEach(tab => {
    if (tab.term) tab.term.options.theme = newTheme
  })
}

// 复制选中文本
function copySelection() {
  const activeT = tabs.value.find(t => t.id === activeTab.value)
  if (!activeT?.term) return
  const sel = activeT.term.getSelection()
  if (sel) navigator.clipboard.writeText(sel)
}

// 主机选择器
const showHostPicker = ref(false)
const hostPickerLoading = ref(false)
const pickerHosts = ref<any[]>([])

watch(showHostPicker, async (v) => {
  if (!v) return
  hostPickerLoading.value = true
  try {
    const res = await hostApi.list({ page: 1, page_size: 100 })
    pickerHosts.value = res.data.data?.list ?? []
  } finally {
    hostPickerLoading.value = false
  }
})

function pickHost(h: any) {
  showHostPicker.value = false
  openTab(h.id, h.name)
}

onMounted(async () => {
  // Read pending connections from host list page
  const pending: { hostId: number; hostName: string }[] =
    JSON.parse(sessionStorage.getItem('pendingSSH') ?? '[]')
  sessionStorage.removeItem('pendingSSH')

  for (const p of pending) {
    await openTab(p.hostId, p.hostName)
  }
})

async function openTab(hostId: number, name: string) {
  const id = `tab-${Date.now()}-${hostId}`
  const tab: Tab = { id, name, hostId, connected: false }
  tabs.value.push(tab)
  activeTab.value = id

  await nextTick()
  await initTerminal(tab)
}

async function initTerminal(tab: Tab) {
  const idx = displayTabs.value.findIndex(t => t.id === tab.id)
  const container = paneRefs[idx]
  if (!container) return

  const term = new Terminal({
    theme: termTheme.current().theme,
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    fontSize: 14,
    cursorBlink: true,
    allowTransparency: true,
  })
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(container)
  fit.fit()
  tab.term = term
  tab.fit = fit

  // Connect WebSocket
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const ws = new WebSocket(`${proto}//${location.host}/api/ws/ssh/${tab.hostId}?token=${auth.token}`)
  ws.binaryType = 'arraybuffer'
  tab.ws = ws

  ws.onopen = () => {
    ws.send(JSON.stringify({ type: 'init', width: term.cols, height: term.rows }))
    tab.connected = true
  }
  ws.onmessage = (e) => {
    const data = e.data instanceof ArrayBuffer
      ? new Uint8Array(e.data)
      : e.data
    term.write(data)
  }
  ws.onclose = () => {
    tab.connected = false
    term.write('\r\n\x1b[31m[连接已断开]\x1b[0m\r\n')
  }

  term.onData(data => {
    if (ws.readyState === WebSocket.OPEN) ws.send(data)
  })
  term.onResize(({ cols, rows }) => {
    if (ws.readyState === WebSocket.OPEN)
      ws.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
  })

  // Watch container resize
  const ro = new ResizeObserver(() => fit.fit())
  ro.observe(container)
}

function closeTab(id: string) {
  const tab = tabs.value.find(t => t.id === id)
  if (tab?.ws) tab.ws.close()
  if (tab?.term) tab.term.dispose()
  tabs.value = tabs.value.filter(t => t.id !== id)
  if (activeTab.value === id && tabs.value.length) {
    const last = tabs.value[tabs.value.length - 1]
    if (last) activeTab.value = last.id
  }
}

function toggleSplit() {
  splitMode.value = !splitMode.value
  nextTick(() => tabs.value.forEach(t => t.fit?.fit()))
}

onUnmounted(() => {
  tabs.value.forEach(t => { t.ws?.close(); t.term?.dispose() })
})
</script>

<style scoped>
.terminal-page { display: flex; flex-direction: column; height: 100%; background: var(--terminal-bg); }
.tab-bar { display: flex; align-items: center; background: var(--bg-sidebar);
  border-bottom: 1px solid var(--border); height: 36px; padding: 0 8px; gap: 2px; overflow-x: auto; }
.tab { display: flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 4px 4px 0 0;
  cursor: pointer; font-size: 12px; color: var(--text-secondary); white-space: nowrap;
  border: 1px solid transparent; border-bottom: none; transition: all 0.15s; }
.tab.active { background: var(--terminal-bg); color: var(--text-primary);
  border-color: var(--border); }
.tab-dot { width: 6px; height: 6px; border-radius: 50%; }
.tab-dot.online { background: var(--success); }
.tab-dot.offline { background: var(--text-muted); }
.tab-close { background: none; border: none; color: inherit; cursor: pointer;
  padding: 0 2px; opacity: 0.5; font-size: 14px; line-height: 1; }
.tab-close:hover { opacity: 1; color: var(--danger); }
.tab-actions { margin-left: auto; display: flex; gap: 4px; align-items: center; }
.split-btn { background: transparent; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: 4px; padding: 2px 8px; cursor: pointer; font-size: 13px; }
.split-btn.active { border-color: var(--accent); color: var(--accent); }
.terminal-area { flex: 1; overflow: hidden; display: flex; gap: 2px; background: var(--border); }
.terminal-pane { flex: 1; overflow: hidden; background: var(--terminal-bg); }
.host-picker-list { display: flex; flex-direction: column; gap: 4px; max-height: 400px; overflow-y: auto; }
.picker-row { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-radius: 6px;
  cursor: pointer; transition: background 0.15s; }
.picker-row:hover { background: color-mix(in srgb, var(--accent) 10%, transparent); }
.picker-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--success); flex-shrink: 0; }
.picker-name { font-size: 13px; color: var(--text-primary); flex: 1; }
.picker-ip { font-size: 12px; color: var(--text-secondary); }
</style>
