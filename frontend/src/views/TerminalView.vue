<!-- frontend/src/views/TerminalView.vue -->
<template>
  <div class="terminal-page">
    <!-- Tab bar -->
    <div class="tab-bar">
      <div v-for="tab in store.tabs" :key="tab.id"
        :class="['tab', tab.id === store.activeId && 'active']"
        @click="store.activeId = tab.id"
        @contextmenu.prevent="showContextMenu($event, tab.id)">
        <div :class="['tab-dot', tab.connected ? 'online' : 'offline']" />
        <span>{{ tab.name }}</span>
        <button class="tab-close" @click.stop="closeTab(tab.id)">×</button>
      </div>

      <button class="new-tab-btn" @click="showHostPicker = true" title="新开终端">＋</button>

      <div class="tab-actions">
        <span class="tab-label">主题</span>
        <n-select
          :value="termTheme.currentId"
          :options="themeOptions"
          size="small"
          style="width:150px"
          @update:value="applyTerminalTheme"
        />
      </div>
    </div>

    <!-- 右键菜单 -->
    <n-dropdown
      placement="bottom-start"
      trigger="manual"
      :x="ctxX" :y="ctxY"
      :options="ctxOptions"
      :show="ctxVisible"
      :on-clickoutside="() => ctxVisible = false"
      @select="handleCtxSelect"
    />

    <!-- Terminal area：v-show 保留所有 tab 的 DOM，切换只改可见性 -->
    <div class="terminal-area">
      <div v-for="tab in store.tabs" :key="tab.id"
        class="terminal-pane"
        v-show="tab.id === store.activeId"
        :ref="el => { if (el) containers[tab.id] = el as HTMLElement }">
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

<script lang="ts">
export default { name: 'TerminalView' }
</script>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, onActivated, nextTick, watch } from 'vue'
import { NSelect, NModal, NSpin, NDropdown } from 'naive-ui'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { useAuthStore } from '@/stores/auth'
import { hostApi } from '@/api/host'
import { useTerminalThemeStore } from '@/stores/terminalTheme'
import { useTerminalStore, type TermTab } from '@/stores/terminal'
import { useThemeStore } from '@/stores/theme'

const auth = useAuthStore()
const termTheme = useTerminalThemeStore()
const store = useTerminalStore()
const themeStore = useThemeStore()
const themeOptions = termTheme.themes.map(t => ({ label: t.name, value: t.id }))

// UI 浅色主题 → 自动切换终端为浅色主题
const LIGHT_UI_THEMES = ['light-clean']
const LIGHT_TERM_THEME = 'github-light'
const DARK_TERM_THEME = 'default'
const LIGHT_TERM_IDS = new Set(['github-light', 'papercolor-light', 'intellij-light'])

watch(() => themeStore.current, (uiTheme) => {
  const wantLight = LIGHT_UI_THEMES.includes(uiTheme)
  const isLightTerm = LIGHT_TERM_IDS.has(termTheme.currentId)
  if (wantLight && !isLightTerm) {
    applyTerminalTheme(LIGHT_TERM_THEME)
  } else if (!wantLight && isLightTerm) {
    applyTerminalTheme(DARK_TERM_THEME)
  }
}, { immediate: true })

// 组件本地：DOM 容器 map 和 xterm 实例 map
const containers: Record<string, HTMLElement> = {}
const xtermMap: Record<string, { term: Terminal; fit: FitAddon }> = {}

// ── 右键菜单 ────────────────────────────────────────────────
const ctxVisible = ref(false)
const ctxX = ref(0)
const ctxY = ref(0)
const ctxTabId = ref('')
const ctxOptions = [
  { label: '复制会话', key: 'duplicate' },
  { label: '关闭会话', key: 'close' },
]
function showContextMenu(e: MouseEvent, tabId: string) {
  ctxTabId.value = tabId; ctxX.value = e.clientX; ctxY.value = e.clientY; ctxVisible.value = true
}
function handleCtxSelect(key: string) {
  ctxVisible.value = false
  const tab = store.getTab(ctxTabId.value)
  if (!tab) return
  if (key === 'close') closeTab(tab.id)
  else if (key === 'duplicate') openTab(tab.hostId, tab.name)
}

// ── 主题 ─────────────────────────────────────────────────────
function applyTerminalTheme(id: string) {
  termTheme.apply(id)
  const newTheme = termTheme.current().theme
  Object.values(xtermMap).forEach(({ term }) => { term.options.theme = newTheme })
}

// ── 主机选择器 ───────────────────────────────────────────────
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
function pickHost(h: any) { showHostPicker.value = false; openTab(h.id, h.name) }

// ── 打开 / 关闭 Tab ──────────────────────────────────────────
async function openTab(hostId: number, name: string) {
  const id = `tab-${Date.now()}-${hostId}`
  const tab: TermTab = { id, name, hostId, connected: false, everConnected: false, outputBuf: [] }
  store.addTab(tab)
  await nextTick()          // 等 v-show DOM 渲染
  await attachXterm(tab)   // 创建 xterm + 连接 WS
}

function closeTab(id: string) {
  const x = xtermMap[id]
  if (x) { x.term.dispose(); delete xtermMap[id] }
  delete containers[id]
  store.removeTab(id)
}

// ── 初始化 / 重连 xterm ──────────────────────────────────────
async function attachXterm(tab: TermTab) {
  await nextTick()
  const container = containers[tab.id]
  if (!container) return

  // 如果已有 xterm（keep-alive 路径），直接 fit 就好
  if (xtermMap[tab.id]) {
    xtermMap[tab.id]!.fit.fit()
    return
  }

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

  // 重放历史输出（组件重建时恢复内容）
  for (const chunk of tab.outputBuf) {
    term.write(chunk instanceof Uint8Array ? chunk : new TextEncoder().encode(chunk as string))
  }
  fit.fit()
  xtermMap[tab.id] = { term, fit }

  // ── WebSocket ─────────────────────────────────────────────
  // 如果已有活跃 WS（全局 store 中），重新绑定 onmessage 即可
  if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
    bindWsToTerm(tab, term)
    term.onData(data => { if (tab.ws!.readyState === WebSocket.OPEN) tab.ws!.send(data) })
    term.onResize(({ cols, rows }) => {
      if (tab.ws!.readyState === WebSocket.OPEN)
        tab.ws!.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
    })
    return
  }

  // 否则新建 WS
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const ws = new WebSocket(`${proto}//${location.host}/api/ws/ssh/${tab.hostId}?token=${auth.token}`)
  ws.binaryType = 'arraybuffer'
  tab.ws = ws

  ws.onopen = () => {
    ws.send(JSON.stringify({ type: 'init', width: term.cols, height: term.rows }))
    tab.connected = true; tab.everConnected = true
  }
  bindWsToTerm(tab, term)
  ws.onclose = (e) => {
    tab.connected = false
    if (tab.everConnected) term.write('\r\n\x1b[33m[连接已断开]\x1b[0m\r\n')
    else {
      const reason = e.code === 4001 ? '认证失败' : e.code === 4003 ? '无权限' : `错误码 ${e.code}`
      term.write(`\r\n\x1b[31m[连接失败: ${reason}]\x1b[0m\r\n`)
    }
  }

  term.onData(data => { if (ws.readyState === WebSocket.OPEN) ws.send(data) })
  term.onResize(({ cols, rows }) => {
    if (ws.readyState === WebSocket.OPEN)
      ws.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
  })

  const ro = new ResizeObserver(() => xtermMap[tab.id]?.fit.fit())
  ro.observe(container)
}

function bindWsToTerm(tab: TermTab, term: Terminal) {
  tab.ws!.onmessage = (e) => {
    const data = e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : e.data
    term.write(data)
    store.appendOutput(tab.id, data)
  }
}

// ── pendingSSH（来自主机列表快捷连接）───────────────────────
async function processPendingSSH() {
  const pending: { hostId: number; hostName: string }[] =
    JSON.parse(sessionStorage.getItem('pendingSSH') ?? '[]')
  if (!pending.length) return
  sessionStorage.removeItem('pendingSSH')
  for (const p of pending) await openTab(p.hostId, p.hostName)
}

// ── 组件重建时恢复已有 tabs（store 中有数据但 xterm 未初始化）─
async function restoreExistingTabs() {
  await nextTick()
  for (const tab of store.tabs) {
    await attachXterm(tab)
  }
}

onMounted(async () => {
  await restoreExistingTabs()
  await processPendingSSH()
})

onActivated(async () => {
  await restoreExistingTabs()
  await processPendingSSH()
  nextTick(() => Object.values(xtermMap).forEach(({ fit }) => fit.fit()))
})

// 组件销毁只清理 xterm DOM，WS 留在 store 里
onUnmounted(() => {
  Object.values(xtermMap).forEach(({ term }) => term.dispose())
  Object.keys(xtermMap).forEach(k => delete xtermMap[k])
})
</script>

<style scoped>
.terminal-page { display: flex; flex-direction: column; height: 100%; background: var(--terminal-bg); }
.tab-bar { display: flex; align-items: center; background: color-mix(in srgb, var(--terminal-bg) 85%, #000);
  border-bottom: 1px solid color-mix(in srgb, var(--terminal-fg) 15%, transparent); height: 36px; padding: 0 8px; gap: 2px; overflow-x: auto; flex-shrink: 0; }
.tab { display: flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 4px 4px 0 0;
  cursor: pointer; font-size: 12px; color: color-mix(in srgb, var(--terminal-fg) 55%, transparent); white-space: nowrap;
  border: 1px solid transparent; border-bottom: none; transition: all 0.15s; }
.tab:hover { color: color-mix(in srgb, var(--terminal-fg) 80%, transparent); }
.tab.active { background: var(--terminal-bg); color: var(--terminal-fg);
  border-color: color-mix(in srgb, var(--terminal-fg) 20%, transparent); }
.tab-dot { width: 6px; height: 6px; border-radius: 50%; }
.tab-dot.online { background: var(--success); }
.tab-dot.offline { background: var(--text-muted); }
.tab-close { background: none; border: none; color: inherit; cursor: pointer;
  padding: 0 2px; opacity: 0.5; font-size: 14px; line-height: 1; }
.tab-close:hover { opacity: 1; color: var(--danger); }
.new-tab-btn { background: transparent; border: 1px solid color-mix(in srgb, var(--terminal-fg) 20%, transparent);
  color: color-mix(in srgb, var(--terminal-fg) 70%, transparent);
  border-radius: 4px; padding: 2px 8px; cursor: pointer; font-size: 16px; line-height: 1;
  flex-shrink: 0; transition: all 0.15s; }
.new-tab-btn:hover { border-color: color-mix(in srgb, var(--terminal-fg) 50%, transparent); color: var(--terminal-fg); }
.tab-actions { margin-left: auto; display: flex; gap: 6px; align-items: center; }
.tab-label { font-size: 12px; color: color-mix(in srgb, var(--terminal-fg) 55%, transparent); white-space: nowrap; }
.terminal-area { flex: 1; overflow: hidden; display: flex; background: var(--terminal-bg); }
.terminal-pane { flex: 1; overflow: hidden; }
.host-picker-list { display: flex; flex-direction: column; gap: 4px; max-height: 400px; overflow-y: auto; }
.picker-row { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-radius: 6px;
  cursor: pointer; transition: background 0.15s; }
.picker-row:hover { background: color-mix(in srgb, var(--accent) 10%, transparent); }
.picker-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--success); flex-shrink: 0; }
.picker-name { font-size: 13px; color: var(--text-primary); flex: 1; }
.picker-ip { font-size: 12px; color: var(--text-secondary); }
</style>
