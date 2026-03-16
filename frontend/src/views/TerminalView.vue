<!-- frontend/src/views/TerminalView.vue -->
<template>
  <div class="terminal-page">
    <!-- Tab bar -->
    <div class="tab-bar">
      <!-- 当前 active 分屏提示 -->
      <span v-if="splitCount > 1" class="split-hint">
        分屏 {{ activeSplitIdx + 1 }} ↓
      </span>

      <div v-for="tab in store.tabs" :key="tab.id"
        :class="['tab', tab.id === activeSplitTabId && 'active', isTabInSplit(tab.id) && 'in-split']"
        :title="splitCount > 1 ? `点击：将此会话放入分屏 ${activeSplitIdx + 1}` : tab.name"
        @click="assignTabToActiveSplit(tab.id)"
        @contextmenu.prevent="showContextMenu($event, tab.id)">
        <div :class="['tab-dot', tab.connected ? 'online' : 'offline']" />
        <span>{{ tab.name }}</span>
        <button class="tab-close" @click.stop="closeTab(tab.id)" title="关闭会话">×</button>
      </div>

      <button class="new-tab-btn" @click="showHostPicker = true" title="新建终端会话">＋</button>

      <!-- 分屏按钮 -->
      <div class="split-btns">
        <button :class="['split-btn', splitCount === 1 && 'active']"
          @click="setSplitCount(1)"
          title="单屏模式">▣</button>
        <button :class="['split-btn', splitCount === 2 && 'active']"
          :disabled="store.tabs.length < 2"
          @click="setSplitCount(2)"
          :title="store.tabs.length < 2 ? '需先开启 2 个以上会话' : '左右分屏'">⬛⬛</button>
        <button :class="['split-btn', splitCount === 4 && 'active']"
          :disabled="store.tabs.length < 4"
          @click="setSplitCount(4)"
          :title="store.tabs.length < 4 ? '需先开启 4 个以上会话' : '四宫格分屏（最多 4 屏）'">⊞</button>
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
      <div v-for="idx in splitCount" :key="idx - 1"
        :class="['split-cell', (idx - 1) === activeSplitIdx && 'focused']"
        @mousedown="onSplitCellClick(idx - 1)">

        <!-- 分屏格标题栏 -->
        <div class="split-header" @mousedown.stop>
          <div v-if="splitTabIds[idx - 1]"
            :class="['split-dot', getSplitTabConnected(idx - 1) ? 'online' : 'offline']" />
          <span class="split-name">{{ getSplitTabName(idx - 1) }}</span>
          <div class="split-header-actions">
            <!-- 换会话按钮 -->
            <div class="split-picker-wrap">
              <button class="split-hdr-btn"
                :title="`切换此分屏的会话（当前：${getSplitTabName(idx - 1)}）`"
                @click.stop="togglePicker(idx - 1)">⇄</button>
              <!-- 会话选择下拉 -->
              <div v-if="pickerOpen[idx - 1]" class="split-picker" @click.stop>
                <div class="picker-title">选择会话</div>
                <div v-if="store.tabs.length === 0" class="picker-empty">暂无会话，请先点 ＋</div>
                <div v-for="tab in store.tabs" :key="tab.id"
                  :class="['picker-item', tab.id === splitTabIds[idx - 1] && 'current']"
                  @click="pickFromHeader(idx - 1, tab.id)">
                  <div :class="['picker-dot', tab.connected ? 'online' : 'offline']" />
                  <span>{{ tab.name }}</span>
                  <span v-if="tab.id === splitTabIds[idx - 1]" class="picker-check">✓</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- xterm 挂载容器 -->
        <div class="split-pane"
          :ref="(el: any) => { if (el) splitCells[idx - 1] = el }">
        </div>

        <!-- 空分屏占位（绝对定位覆盖 pane） -->
        <div v-if="!splitTabIds[idx - 1]" class="split-placeholder">
          <div class="placeholder-inner">
            <span class="placeholder-icon">⌨</span>
            <span>点击 <b>⇄</b> 选择会话</span>
            <span>或点击上方标签页分配</span>
          </div>
        </div>
      </div>
    </div>

    <!-- picker 遮罩（点击空白关闭） -->
    <div v-if="anyPickerOpen" class="picker-mask" @click="closeAllPickers" />

    <!-- 主机选择 Modal -->
    <n-modal v-model:show="showHostPicker" preset="card" title="选择主机" style="width:480px">
      <div v-if="hostPickerLoading" style="text-align:center;padding:20px"><n-spin /></div>
      <div v-else class="host-picker-list">
        <div v-for="h in pickerHosts" :key="h.id"
          class="picker-row" @click="pickHost(h)">
          <span class="picker-status-dot" />
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

// ── 分屏状态 ─────────────────────────────────────────────
const splitCount = ref<1 | 2 | 4>(1)
const activeSplitIdx = ref(0)
// splitTabIds[i] = 第 i 格显示的 tabId（'' = 空）
const splitTabIds = ref<string[]>([''])

const activeSplitTabId = computed(() => splitTabIds.value[activeSplitIdx.value] ?? '')

function isTabInSplit(tabId: string) {
  return splitTabIds.value.some(id => id === tabId)
}

// ── split-cell 内的 picker 状态 ──────────────────────────
const pickerOpen = ref<boolean[]>([false, false, false, false])
const anyPickerOpen = computed(() => pickerOpen.value.some(Boolean))
function togglePicker(idx: number) {
  const next = !pickerOpen.value[idx]
  pickerOpen.value = [false, false, false, false]
  pickerOpen.value[idx] = next
}
function closeAllPickers() {
  pickerOpen.value = [false, false, false, false]
}

// ── split-cell DOM 引用 ───────────────────────────────────
const splitCells: HTMLElement[] = []
const splitXterms: Record<number, { term: Terminal; fit: FitAddon; ro: ResizeObserver }> = {}
const tabSplitIdx: Record<string, number> = {}

// ── split-cell 辅助 ──────────────────────────────────────
function getSplitTabName(idx: number) {
  const id = splitTabIds.value[idx]
  if (!id) return `分屏 ${idx + 1}`
  return store.getTab(id)?.name ?? '未知'
}
function getSplitTabConnected(idx: number) {
  const id = splitTabIds.value[idx]
  return id ? (store.getTab(id)?.connected ?? false) : false
}

// ── 右键菜单 ────────────────────────────────────────────
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

// ── 主题 ────────────────────────────────────────────────
function applyTerminalTheme(id: string) {
  termTheme.apply(id)
  const newTheme = termTheme.current().theme
  Object.values(splitXterms).forEach(({ term }) => { term.options.theme = newTheme })
}

// ── 主机选择器 ───────────────────────────────────────────
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

// ── 分屏数量切换 ─────────────────────────────────────────
function setSplitCount(count: 1 | 2 | 4) {
  if (count > store.tabs.length) return
  splitCount.value = count
  const cur = splitTabIds.value.slice(0, count)
  while (cur.length < count) {
    const used = new Set(cur.filter(Boolean))
    const next = store.tabs.find(t => !used.has(t.id))
    cur.push(next?.id ?? '')
  }
  for (let i = count; i < splitTabIds.value.length; i++) {
    disposeSplitXterm(i)
  }
  splitTabIds.value = cur
  if (activeSplitIdx.value >= count) activeSplitIdx.value = 0
  nextTick(initAllSplits)
}

// ── 点击分屏格 ───────────────────────────────────────────
function onSplitCellClick(idx: number) {
  activeSplitIdx.value = idx
  closeAllPickers()
  const tabId = splitTabIds.value[idx]
  if (tabId) nextTick(() => splitXterms[idx]?.term.focus())
}

// ── 通过 tab bar 分配 tab 到当前 active 分屏 ─────────────
async function assignTabToActiveSplit(tabId: string) {
  await assignTabToSplit(activeSplitIdx.value, tabId)
}

// ── 通过 header picker 选择 tab ──────────────────────────
async function pickFromHeader(idx: number, tabId: string) {
  closeAllPickers()
  activeSplitIdx.value = idx
  await assignTabToSplit(idx, tabId)
}

async function assignTabToSplit(idx: number, tabId: string) {
  const oldTabId = splitTabIds.value[idx]
  if (oldTabId === tabId) {
    nextTick(() => splitXterms[idx]?.term.focus())
    return
  }
  // 如果该 tab 已在另一个 split，先从那移除
  const prevIdx = tabSplitIdx[tabId]
  if (prevIdx !== undefined && prevIdx !== idx) {
    disposeSplitXterm(prevIdx)
    splitTabIds.value[prevIdx] = ''
    delete tabSplitIdx[tabId]
  }
  // 释放当前 split 的旧 xterm
  if (oldTabId) {
    disposeSplitXterm(idx)
    delete tabSplitIdx[oldTabId]
  }
  splitTabIds.value[idx] = tabId
  await nextTick()
  await mountXtermToSplit(idx)
}

// ── 挂载 xterm 到第 idx 个 split-cell ─────────────────────
async function mountXtermToSplit(idx: number) {
  await nextTick()
  const tabId = splitTabIds.value[idx]
  if (!tabId) return
  const tab = store.getTab(tabId)
  if (!tab) return
  const cell = splitCells[idx]
  if (!cell) return

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

  for (const chunk of tab.outputBuf) {
    term.write(chunk instanceof Uint8Array ? chunk : new TextEncoder().encode(chunk as string))
  }
  fit.fit()
  term.focus()

  const ro = new ResizeObserver(() => splitXterms[idx]?.fit.fit())
  ro.observe(cell)
  splitXterms[idx] = { term, fit, ro }
  tabSplitIdx[tabId] = idx

  if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
    bindWsToTerm(tab, term)
    term.onData(d => { if (tab.ws!.readyState === WebSocket.OPEN) tab.ws!.send(d) })
    term.onResize(({ cols, rows }) => {
      if (tab.ws!.readyState === WebSocket.OPEN)
        tab.ws!.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
    })
    return
  }

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

// ── 释放 split-cell 的 xterm ────────────────────────────
function disposeSplitXterm(idx: number) {
  const x = splitXterms[idx]
  if (!x) return
  x.ro.disconnect()
  x.term.dispose()
  delete splitXterms[idx]
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

// ── 打开 / 关闭 Tab ─────────────────────────────────────
async function openTab(hostId: number, name: string) {
  const id = `tab-${Date.now()}-${hostId}`
  const tab: TermTab = { id, name, hostId, connected: false, everConnected: false, outputBuf: [] }
  store.addTab(tab)
  await nextTick()
  await assignTabToActiveSplit(id)
}

function closeTab(id: string) {
  splitTabIds.value.forEach((tabId, idx) => {
    if (tabId === id) {
      disposeSplitXterm(idx)
      splitTabIds.value[idx] = ''
      delete tabSplitIdx[id]
    }
  })
  store.removeTab(id)
}

// ── 初始化 / 恢复所有分屏 ───────────────────────────────
async function initAllSplits() {
  await nextTick()
  for (let i = 0; i < splitCount.value; i++) {
    if (splitTabIds.value[i]) await mountXtermToSplit(i)
  }
}

async function restoreExistingTabs() {
  const tabs = store.tabs
  if (!tabs.length) return
  const newIds: string[] = []
  for (let i = 0; i < splitCount.value; i++) {
    newIds.push(tabs[i]?.id ?? '')
  }
  splitTabIds.value = newIds
  await initAllSplits()
}

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

onUnmounted(() => {
  Object.keys(splitXterms).forEach(k => disposeSplitXterm(Number(k)))
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
.split-hint {
  font-size: 11px; color: var(--accent); white-space: nowrap;
  padding: 2px 6px; background: color-mix(in srgb, var(--accent) 12%, transparent);
  border-radius: 3px; flex-shrink: 0; margin-right: 4px;
}
.tab {
  display: flex; align-items: center; gap: 6px; padding: 4px 10px;
  border-radius: 4px 4px 0 0; cursor: pointer; font-size: 12px;
  color: color-mix(in srgb, var(--terminal-fg) 45%, transparent); white-space: nowrap;
  border: 1px solid transparent; border-bottom: none; transition: all 0.15s;
}
.tab:hover { color: color-mix(in srgb, var(--terminal-fg) 75%, transparent); }
.tab.in-split { color: color-mix(in srgb, var(--terminal-fg) 65%, transparent); }
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
.split-btns { display: flex; gap: 2px; margin-left: 4px; flex-shrink: 0; }
.split-btn {
  background: transparent;
  border: 1px solid color-mix(in srgb, var(--terminal-fg) 15%, transparent);
  color: color-mix(in srgb, var(--terminal-fg) 50%, transparent);
  border-radius: 3px; padding: 2px 6px; cursor: pointer; font-size: 11px;
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

/* ── 终端区域 Grid ── */
.terminal-area {
  flex: 1; overflow: hidden; background: var(--terminal-bg);
  display: grid; gap: 2px;
}
.terminal-area.split-1 { grid-template-columns: 1fr; grid-template-rows: 1fr; }
.terminal-area.split-2 { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr; }
.terminal-area.split-4 { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }

/* ── 分屏格 ── */
.split-cell {
  display: flex; flex-direction: column; overflow: hidden; position: relative;
  border: 1px solid transparent; transition: border-color 0.15s;
}
.split-cell.focused {
  border-color: color-mix(in srgb, var(--accent) 60%, transparent);
}

/* ── 分屏格标题栏 ── */
.split-header {
  display: flex; align-items: center; gap: 5px;
  height: 22px; padding: 0 6px; flex-shrink: 0;
  background: color-mix(in srgb, var(--terminal-bg) 70%, #000);
  border-bottom: 1px solid color-mix(in srgb, var(--terminal-fg) 10%, transparent);
  user-select: none;
}
.split-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.split-dot.online { background: var(--success); }
.split-dot.offline { background: var(--text-muted); }
.split-name {
  font-size: 11px; color: color-mix(in srgb, var(--terminal-fg) 60%, transparent);
  flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.split-header-actions { display: flex; gap: 2px; flex-shrink: 0; }
.split-hdr-btn {
  background: transparent; border: none;
  color: color-mix(in srgb, var(--terminal-fg) 45%, transparent);
  cursor: pointer; font-size: 12px; padding: 0 4px; border-radius: 3px;
  line-height: 1.6; transition: all 0.15s;
}
.split-hdr-btn:hover {
  background: color-mix(in srgb, var(--terminal-fg) 12%, transparent);
  color: var(--terminal-fg);
}

/* ── 会话选择下拉 ── */
.split-picker-wrap { position: relative; }
.split-picker {
  position: absolute; top: 100%; right: 0; z-index: 200;
  background: var(--bg-elevated); border: 1px solid var(--border);
  border-radius: 6px; min-width: 180px; padding: 4px 0;
  box-shadow: 0 8px 24px rgba(0,0,0,0.3);
}
.picker-title {
  font-size: 11px; color: var(--text-secondary);
  padding: 4px 12px 6px; border-bottom: 1px solid var(--border); margin-bottom: 2px;
}
.picker-empty { font-size: 12px; color: var(--text-secondary); padding: 8px 12px; }
.picker-item {
  display: flex; align-items: center; gap: 7px;
  padding: 6px 12px; cursor: pointer; font-size: 12px; color: var(--text-primary);
  transition: background 0.12s;
}
.picker-item:hover { background: color-mix(in srgb, var(--accent) 10%, transparent); }
.picker-item.current { color: var(--accent); }
.picker-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.picker-dot.online { background: var(--success); }
.picker-dot.offline { background: var(--text-muted); }
.picker-check { margin-left: auto; font-size: 12px; color: var(--accent); }
.picker-mask {
  position: fixed; inset: 0; z-index: 199;
}

/* ── 分屏 pane（xterm 容器） ── */
.split-pane { flex: 1; overflow: hidden; }

/* ── 空分屏占位 ── */
.split-placeholder {
  position: absolute; inset: 22px 0 0 0; /* 22px = header 高度 */
  display: flex; align-items: center; justify-content: center;
  pointer-events: none;
}
.placeholder-inner {
  display: flex; flex-direction: column; align-items: center; gap: 6px;
}
.placeholder-icon { font-size: 28px; opacity: 0.2; }
.placeholder-inner span {
  font-size: 12px; color: color-mix(in srgb, var(--terminal-fg) 25%, transparent);
}
.placeholder-inner b { color: color-mix(in srgb, var(--terminal-fg) 40%, transparent); }

/* ── 主机选择弹窗 ── */
.host-picker-list { display: flex; flex-direction: column; gap: 4px; max-height: 400px; overflow-y: auto; }
.picker-row { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-radius: 6px;
  cursor: pointer; transition: background 0.15s; }
.picker-row:hover { background: color-mix(in srgb, var(--accent) 10%, transparent); }
.picker-status-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--success); flex-shrink: 0; }
.picker-name { font-size: 13px; color: var(--text-primary); flex: 1; }
.picker-ip { font-size: 12px; color: var(--text-secondary); }
</style>
