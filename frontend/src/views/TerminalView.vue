<!-- frontend/src/views/TerminalView.vue -->
<template>
  <div class="terminal-page">
    <!-- Tab bar -->
    <div class="tab-bar">
      <div v-for="tab in store.tabs" :key="tab.id"
        :class="['tab',
          isTabInSplit(tab.id) && 'visible',
          tab.id === activeSplitTabId && 'active']"
        @click="assignTabToActiveSplit(tab.id)"
        @contextmenu.prevent="showContextMenu($event, tab.id)">
        <div :class="['tab-dot', tab.connected ? 'online' : 'offline']" />
        <span>{{ tab.name }}</span>
        <button class="tab-close" @click.stop="closeTab(tab.id)">×</button>
      </div>

      <button class="new-tab-btn" @click="showHostPicker = true" title="新开终端">＋</button>

      <!-- 分屏按钮 -->
      <div class="split-btns">
        <button :class="['split-btn', splitCount === 1 && 'active']"
          @click="setSplitCount(1)" title="单屏">▣</button>
        <button :class="['split-btn', splitCount === 2 && 'active']"
          :disabled="store.tabs.length < 2"
          @click="setSplitCount(2)" title="左右分屏">⬛⬛</button>
        <button :class="['split-btn', splitCount === 4 && 'active']"
          :disabled="store.tabs.length < 4"
          @click="setSplitCount(4)" title="四宫格（最多 4 屏）">⊞</button>
      </div>

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

    <!-- 分屏终端区域 -->
    <div class="terminal-area" :class="`split-${splitCount}`">
      <!-- split-cell 是固定的 grid 单元格，xterm 挂在其中 -->
      <div v-for="idx in splitCount" :key="idx - 1"
        :class="['split-cell', (idx - 1) === activeSplitIdx && 'focused']"
        @mousedown="activeSplitIdx = idx - 1"
        :ref="el => { if (el) splitCells[idx - 1] = el as HTMLElement }">
        <!-- 当该 slot 没有分配 tab 时，显示占位符 -->
        <div v-if="!splitTabIds[idx - 1]" class="split-placeholder">
          <span>点击标签页分配到此分屏</span>
        </div>
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
import { ref, computed, onMounted, onUnmounted, onActivated, nextTick, watch } from 'vue'
import { NSelect, NModal, NSpin, NDropdown } from 'naive-ui'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { useAuthStore } from '@/stores/auth'
import { hostApi } from '@/api/host'
import { useTerminalThemeStore } from '@/stores/terminalTheme'
import { useTerminalStore, type TermTab } from '@/stores/terminal'

const auth = useAuthStore()
const termTheme = useTerminalThemeStore()
const store = useTerminalStore()
const themeOptions = termTheme.themes.map(t => ({ label: t.name, value: t.id }))

// ── 分屏状态 ──────────────────────────────────────────────
const splitCount = ref<1 | 2 | 4>(1)
const activeSplitIdx = ref(0)
// splitTabIds[i] = 分配给第 i 个分屏格的 tabId（'' 表示空）
const splitTabIds = ref<string[]>([''])

// 当前 active 分屏正在显示哪个 tab（用于高亮 tab bar）
const activeSplitTabId = computed(() => splitTabIds.value[activeSplitIdx.value] ?? '')

// 某 tabId 是否在任意分屏中可见
function isTabInSplit(tabId: string) {
  return splitTabIds.value.includes(tabId)
}

// ── 分屏格 DOM 引用 ───────────────────────────────────────
// splitCells[i] = 第 i 个 grid cell 的 HTMLElement
const splitCells: HTMLElement[] = []
// splitXterms[i] = 第 i 个 grid cell 中运行的 {term, fit, ro}
const splitXterms: Record<number, { term: Terminal; fit: FitAddon; ro: ResizeObserver }> = {}
// 每个 tabId 对应的 xterm 挂在哪个 splitIdx（反查）
const tabSplitIdx: Record<string, number> = {}

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
  Object.values(splitXterms).forEach(({ term }) => { term.options.theme = newTheme })
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

// ── 切换分屏数量 ─────────────────────────────────────────────
function setSplitCount(count: 1 | 2 | 4) {
  if (count > store.tabs.length) return
  splitCount.value = count
  // 保留已有分配，扩充或截断 splitTabIds
  const cur = splitTabIds.value.slice(0, count)
  while (cur.length < count) {
    // 找一个尚未被分配的 tab
    const used = new Set(cur.filter(Boolean))
    const next = store.tabs.find(t => !used.has(t.id))
    cur.push(next?.id ?? '')
  }
  // 释放被截断的分屏中的 xterm
  for (let i = count; i < splitTabIds.value.length; i++) {
    disposeSplitXterm(i)
  }
  splitTabIds.value = cur
  if (activeSplitIdx.value >= count) activeSplitIdx.value = 0
  nextTick(initAllSplits)
}

// ── 把 tab 分配到当前 active 分屏 ───────────────────────────
async function assignTabToActiveSplit(tabId: string) {
  const idx = activeSplitIdx.value
  const oldTabId = splitTabIds.value[idx]
  if (oldTabId === tabId) return // 已经在这里了，focus 即可

  // 如果该 tab 已经在另一个 split，先从那里移除
  const prevIdx = tabSplitIdx[tabId]
  if (prevIdx !== undefined && prevIdx !== idx) {
    disposeSplitXterm(prevIdx)
    splitTabIds.value[prevIdx] = ''
    delete tabSplitIdx[tabId]
  }

  // 释放当前 split 里的旧 xterm
  if (oldTabId) {
    disposeSplitXterm(idx)
    delete tabSplitIdx[oldTabId]
  }

  splitTabIds.value[idx] = tabId
  await nextTick()
  await mountXtermToSplit(idx)
}

// ── 创建 xterm 并挂到第 idx 个 split-cell ────────────────────
async function mountXtermToSplit(idx: number) {
  await nextTick()
  const tabId = splitTabIds.value[idx]
  if (!tabId) return
  const tab = store.getTab(tabId)
  if (!tab) return
  const cell = splitCells[idx]
  if (!cell) return

  // 已挂载则直接 fit
  if (splitXterms[idx]) {
    splitXterms[idx].fit.fit()
    splitXterms[idx].term.focus()
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
  term.open(cell)

  // 回放历史输出
  for (const chunk of tab.outputBuf) {
    term.write(chunk instanceof Uint8Array ? chunk : new TextEncoder().encode(chunk as string))
  }
  fit.fit()
  term.focus()

  const ro = new ResizeObserver(() => splitXterms[idx]?.fit.fit())
  ro.observe(cell)
  splitXterms[idx] = { term, fit, ro }
  tabSplitIdx[tabId] = idx

  // ── WebSocket ─────────────────────────────────────────────
  if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
    bindWsToTerm(tab, term)
    term.onData(d => { if (tab.ws!.readyState === WebSocket.OPEN) tab.ws!.send(d) })
    term.onResize(({ cols, rows }) => {
      if (tab.ws!.readyState === WebSocket.OPEN)
        tab.ws!.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
    })
    return
  }

  // 新建 WS
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
  term.onData(d => { if (ws.readyState === WebSocket.OPEN) ws.send(d) })
  term.onResize(({ cols, rows }) => {
    if (ws.readyState === WebSocket.OPEN)
      ws.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
  })
}

function bindWsToTerm(tab: TermTab, term: Terminal) {
  tab.ws!.onmessage = (e) => {
    const data = e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : e.data
    term.write(data)
    store.appendOutput(tab.id, data)
  }
}

// ── 释放某个 split-cell 的 xterm ────────────────────────────
function disposeSplitXterm(idx: number) {
  const x = splitXterms[idx]
  if (!x) return
  x.ro.disconnect()
  x.term.dispose()
  delete splitXterms[idx]
  // 让 WS 改为纯缓冲模式（不再写入已销毁的 xterm）
  const tabId = splitTabIds.value[idx]
  if (tabId) {
    const tab = store.getTab(tabId)
    if (tab?.ws) {
      tab.ws.onmessage = (e) => {
        const data = e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : e.data
        store.appendOutput(tab.id, data)
      }
    }
  }
}

// ── 打开 / 关闭 Tab ──────────────────────────────────────────
async function openTab(hostId: number, name: string) {
  const id = `tab-${Date.now()}-${hostId}`
  const tab: TermTab = { id, name, hostId, connected: false, everConnected: false, outputBuf: [] }
  store.addTab(tab)
  await nextTick()
  // 分配到当前 active split（自动替换该格的旧 tab）
  await assignTabToActiveSplit(id)
}

function closeTab(id: string) {
  // 从所有持有该 tab 的 split 中释放
  splitTabIds.value.forEach((tabId, idx) => {
    if (tabId === id) {
      disposeSplitXterm(idx)
      splitTabIds.value[idx] = ''
      delete tabSplitIdx[id]
    }
  })
  store.removeTab(id)
}

// ── 初始化所有分屏 ───────────────────────────────────────────
async function initAllSplits() {
  await nextTick()
  for (let i = 0; i < splitCount.value; i++) {
    if (splitTabIds.value[i]) {
      await mountXtermToSplit(i)
    }
  }
}

// ── 恢复 store 中已有 tabs ──────────────────────────────────
async function restoreExistingTabs() {
  const tabs = store.tabs
  if (!tabs.length) return
  // 初始化 splitTabIds：每个 split 格自动分配一个 tab
  const newIds: string[] = []
  for (let i = 0; i < splitCount.value; i++) {
    newIds.push(tabs[i]?.id ?? '')
  }
  splitTabIds.value = newIds
  await initAllSplits()
}

// ── pendingSSH ───────────────────────────────────────────────
async function processPendingSSH() {
  const pending: { hostId: number; hostName: string }[] =
    JSON.parse(sessionStorage.getItem('pendingSSH') ?? '[]')
  if (!pending.length) return
  sessionStorage.removeItem('pendingSSH')
  for (const p of pending) await openTab(p.hostId, p.hostName)
}

onMounted(async () => {
  await restoreExistingTabs()
  await processPendingSSH()
})

onActivated(async () => {
  await restoreExistingTabs()
  await processPendingSSH()
  nextTick(() => Object.values(splitXterms).forEach(({ fit }) => fit.fit()))
})

// 销毁：只清理 xterm，WS 留在 store
onUnmounted(() => {
  Object.keys(splitXterms).forEach(k => {
    const idx = Number(k)
    disposeSplitXterm(idx)
  })
})
</script>

<style scoped>
.terminal-page { display: flex; flex-direction: column; height: 100%; background: var(--terminal-bg); }

/* ── Tab bar ── */
.tab-bar {
  display: flex; align-items: center;
  background: color-mix(in srgb, var(--terminal-bg) 85%, #000);
  border-bottom: 1px solid color-mix(in srgb, var(--terminal-fg) 15%, transparent);
  height: 36px; padding: 0 8px; gap: 2px; overflow-x: auto; flex-shrink: 0;
}
.tab {
  display: flex; align-items: center; gap: 6px; padding: 4px 10px;
  border-radius: 4px 4px 0 0; cursor: pointer; font-size: 12px;
  color: color-mix(in srgb, var(--terminal-fg) 45%, transparent); white-space: nowrap;
  border: 1px solid transparent; border-bottom: none; transition: all 0.15s;
}
.tab:hover { color: color-mix(in srgb, var(--terminal-fg) 75%, transparent); }
.tab.visible { color: color-mix(in srgb, var(--terminal-fg) 70%, transparent); }
.tab.active {
  background: var(--terminal-bg); color: var(--terminal-fg);
  border-color: color-mix(in srgb, var(--terminal-fg) 20%, transparent);
}
.tab-dot { width: 6px; height: 6px; border-radius: 50%; }
.tab-dot.online { background: var(--success); }
.tab-dot.offline { background: var(--text-muted); }
.tab-close { background: none; border: none; color: inherit; cursor: pointer;
  padding: 0 2px; opacity: 0.5; font-size: 14px; line-height: 1; }
.tab-close:hover { opacity: 1; color: var(--danger); }
.new-tab-btn {
  background: transparent; border: 1px solid color-mix(in srgb, var(--terminal-fg) 20%, transparent);
  color: color-mix(in srgb, var(--terminal-fg) 70%, transparent);
  border-radius: 4px; padding: 2px 8px; cursor: pointer; font-size: 16px; line-height: 1;
  flex-shrink: 0; transition: all 0.15s;
}
.new-tab-btn:hover { border-color: color-mix(in srgb, var(--terminal-fg) 50%, transparent); color: var(--terminal-fg); }

/* ── 分屏按钮 ── */
.split-btns {
  display: flex; gap: 2px; margin-left: 4px; flex-shrink: 0;
}
.split-btn {
  background: transparent;
  border: 1px solid color-mix(in srgb, var(--terminal-fg) 15%, transparent);
  color: color-mix(in srgb, var(--terminal-fg) 50%, transparent);
  border-radius: 3px; padding: 2px 5px; cursor: pointer; font-size: 11px;
  line-height: 1.4; transition: all 0.15s;
}
.split-btn:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--terminal-fg) 40%, transparent);
  color: var(--terminal-fg);
}
.split-btn.active {
  background: color-mix(in srgb, var(--terminal-fg) 15%, transparent);
  color: var(--terminal-fg);
  border-color: color-mix(in srgb, var(--terminal-fg) 30%, transparent);
}
.split-btn:disabled { opacity: 0.3; cursor: not-allowed; }

.tab-actions { margin-left: auto; display: flex; gap: 6px; align-items: center; }
.tab-label { font-size: 12px; color: color-mix(in srgb, var(--terminal-fg) 55%, transparent); white-space: nowrap; }

/* ── 终端区域 + 分屏 Grid ── */
.terminal-area {
  flex: 1; overflow: hidden; background: var(--terminal-bg);
  display: grid; gap: 2px;
}
.terminal-area.split-1  { grid-template-columns: 1fr; grid-template-rows: 1fr; }
.terminal-area.split-2  { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr; }
.terminal-area.split-4  { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }

.split-cell {
  overflow: hidden; position: relative;
  border: 1px solid transparent;
  transition: border-color 0.15s;
}
.split-cell.focused {
  border-color: color-mix(in srgb, var(--terminal-fg) 30%, transparent);
}

/* 空分屏占位符 */
.split-placeholder {
  height: 100%; display: flex; align-items: center; justify-content: center;
}
.split-placeholder span {
  font-size: 12px;
  color: color-mix(in srgb, var(--terminal-fg) 25%, transparent);
  user-select: none;
}

/* 主机选择弹窗 */
.host-picker-list { display: flex; flex-direction: column; gap: 4px; max-height: 400px; overflow-y: auto; }
.picker-row { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-radius: 6px;
  cursor: pointer; transition: background 0.15s; }
.picker-row:hover { background: color-mix(in srgb, var(--accent) 10%, transparent); }
.picker-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--success); flex-shrink: 0; }
.picker-name { font-size: 13px; color: var(--text-primary); flex: 1; }
.picker-ip { font-size: 12px; color: var(--text-secondary); }
</style>
