<!-- frontend/src/views/DbView.vue -->
<template>
  <div class="db-page" @mousemove="onDragMove" @mouseup="onDragEnd">
    <!-- 左侧侧边栏 -->
    <div class="db-sidebar" :style="{ width: sidebarWidth + 'px' }">
      <!-- 服务器 Tab 栏 -->
      <div class="srv-tabs">
        <div
          v-for="srv in servers"
          :key="srv.id"
          class="srv-tab"
          :class="{ active: activeSrvId === srv.id }"
          @click="selectServer(srv)"
        >
          <span>{{ srv.type === 'mysql' ? '🐬' : '🐘' }} {{ srv.name }}</span>
          <span class="tab-close" @click.stop="deleteServer(srv.id)">×</span>
        </div>
        <button class="add-srv-btn" @click="showAddModal = true" title="添加数据库服务器">＋</button>
      </div>

      <!-- 数据库 / 表列表 -->
      <div class="db-tree">
        <div v-if="!activeSrvId" class="tree-empty">请选择或添加服务器</div>
        <div v-else-if="loadingDbs" class="tree-empty">加载中…</div>
        <template v-else>
          <div v-for="db in currentDbs" :key="db">
            <div
              class="db-row"
              :class="{ selected: selectedDb === db }"
              @click="toggleDb(db)"
            >
              <span class="db-expand">{{ expandedDb === db ? '▾' : '▸' }}</span>
              <span class="db-icon">🗄</span>
              <span class="db-name">{{ db }}</span>
            </div>
            <div v-if="expandedDb === db" class="table-list">
              <div v-if="loadingTables" class="tree-empty" style="padding-left:32px">加载中…</div>
              <div
                v-for="t in currentTables"
                :key="t"
                class="table-row"
                @click="fillSelectAll(db, t)"
              >
                <span>☰</span>
                <span class="table-name">{{ t }}</span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 侧边栏拖拽分割线 -->
    <div class="resize-handle" @mousedown="onSidebarDragStart" />

    <!-- 右侧区域（编辑器 + AI 面板） -->
    <div class="db-right">
      <div class="db-main">
        <!-- 工具栏 -->
        <div class="editor-toolbar">
          <span v-if="selectedDb" class="ctx-badge">
            {{ activeSrv?.name }} › {{ selectedDb }}
          </span>
          <button class="run-btn" :disabled="running" @click="runQuery">
            {{ running ? '执行中…' : '▶ 执行  Ctrl+Enter' }}
          </button>
          <span v-if="lastResult" class="result-info">
            {{ lastResult.total }} 行{{ lastResult.total >= 1000 ? '（已截断至 1000）' : '' }}
          </span>
          <button
            class="ai-toggle-btn"
            :class="{ active: aiPanelOpen }"
            @click="aiPanelOpen = !aiPanelOpen"
            title="AI SQL 助手"
          >✦ AI</button>
        </div>

        <!-- CodeMirror 编辑器（高度可拖拽） -->
        <div ref="editorEl" class="sql-editor" :style="{ height: editorHeight + 'px' }" />

        <!-- 编辑器/结果区拖拽分割线 -->
        <div class="editor-resize-handle" @mousedown="onEditorDragStart" />

        <!-- 结果表格 -->
        <div class="result-area">
          <div v-if="queryError" class="query-error">{{ queryError }}</div>
          <n-data-table
            v-else-if="lastResult"
            :columns="resultColumns"
            :data="tableData"
            :max-height="resultAreaHeight"
            size="small"
            :scroll-x="lastResult.columns.length * 120"
          />
        </div>

        <!-- 查询历史折叠面板 -->
        <div class="history-panel">
          <div class="history-header" @click="historyOpen = !historyOpen">
            <span>查询历史 ({{ history.length }})</span>
            <span>{{ historyOpen ? '▾' : '▸' }}</span>
          </div>
          <div v-if="historyOpen" class="history-list">
            <div
              v-for="(h, i) in history"
              :key="i"
              class="history-item"
              @click="fillFromHistory(h)"
            >
              <span class="history-sql">{{ h.sql.slice(0, 80) }}</span>
              <span class="history-meta">
                {{ h.database }} · {{ h.durationMs }}ms · {{ h.error ? '❌' : h.rowsReturned + '行' }}
              </span>
            </div>
            <div v-if="history.length === 0" class="tree-empty">暂无历史</div>
          </div>
        </div>
      </div>

      <!-- AI SQL 助手面板 -->
      <div v-if="aiPanelOpen" class="ai-panel">
        <div class="ai-panel-header">
          <span>✦ AI SQL 助手</span>
          <div style="display:flex;gap:6px;align-items:center;">
            <button v-if="aiMessages.length" class="ai-clear-btn" @click="clearAiHistory" title="清除对话">清除</button>
            <button class="ai-close" @click="aiPanelOpen = false">×</button>
          </div>
        </div>
        <div class="ai-history" ref="aiHistoryEl">
          <div v-for="(msg, i) in aiMessages" :key="i" :class="['ai-msg', `ai-msg-${msg.role}`]">
            <div class="ai-msg-label">{{ msg.role === 'user' ? '你' : 'AI' }}</div>
            <pre class="ai-code">{{ msg.content }}</pre>
            <button v-if="msg.role === 'assistant'" class="ai-insert-btn" @click="insertToEditor(msg.content)">
              插入编辑器
            </button>
          </div>
          <div v-if="aiStreamingContent || aiLoading" class="ai-msg ai-msg-assistant">
            <div class="ai-msg-label">AI</div>
            <pre class="ai-code">{{ aiStreamingContent }}<span v-if="aiLoading" class="ai-cursor">▌</span></pre>
          </div>
        </div>
        <div class="ai-input-row">
          <textarea
            v-model="aiPrompt"
            class="ai-textarea"
            placeholder="描述你想查询的内容…"
            rows="2"
            @keydown.ctrl.enter="askAI"
          />
          <button class="ai-send-btn" :disabled="aiLoading" @click="askAI">
            {{ aiLoading ? '生成中…' : '生成 SQL' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 添加服务器 Modal -->
    <n-modal v-model:show="showAddModal" preset="card" title="添加数据库服务器" style="width:460px">
      <n-form :model="addForm" label-placement="left" label-width="80px">
        <n-form-item label="名称"><n-input v-model:value="addForm.name" /></n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="addForm.type"
            :options="[{label:'MySQL',value:'mysql'},{label:'PostgreSQL',value:'postgres'}]" />
        </n-form-item>
        <n-form-item label="主机"><n-input v-model:value="addForm.host" /></n-form-item>
        <n-form-item label="端口"><n-input-number v-model:value="addForm.port" /></n-form-item>
        <n-form-item label="用户名"><n-input v-model:value="addForm.username" /></n-form-item>
        <n-form-item label="密码"><n-input v-model:value="addForm.password" type="password" show-password-on="click" /></n-form-item>
        <n-form-item label="默认库"><n-input v-model:value="addForm.database" placeholder="可选" /></n-form-item>
      </n-form>
      <template #footer>
        <n-button type="primary" @click="addServer">确定</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { NModal, NForm, NFormItem, NInput, NInputNumber, NButton, NDataTable, NSelect } from 'naive-ui'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { sql } from '@codemirror/lang-sql'
import { oneDark } from '@codemirror/theme-one-dark'
import { defaultKeymap } from '@codemirror/commands'
import { dbApi, type DBServer, type QueryResult, type QueryHistory } from '@/api/db'
import { streamSQLCommand, type AIMessage } from '@/api/ai'

// ── 状态 ──────────────────────────────────────────────
const servers     = ref<DBServer[]>([])
const activeSrvId = ref<number | null>(null)
const selectedDb  = ref<string>('')
const expandedDb  = ref<string>('')
const dbMap       = ref<Record<number, string[]>>({})
const tableMap    = ref<Record<string, string[]>>({})
const loadingDbs  = ref(false)
const loadingTables = ref(false)

const lastResult  = ref<QueryResult | null>(null)
const queryError  = ref('')
const running     = ref(false)

const history     = ref<QueryHistory[]>([])
const historyOpen = ref(false)

const showAddModal = ref(false)
const editorEl     = ref<HTMLElement>()
let editorView: EditorView | null = null

// 侧边栏拖拽
const SIDEBAR_KEY = 'db-sidebar-width'
const sidebarWidth = ref(220)
let sidebarDragging = false
let sidebarDragStartX = 0
let sidebarDragStartWidth = 0

// 编辑器高度拖拽
const EDITOR_HEIGHT_KEY = 'db-editor-height'
const editorHeight = ref(150)
let editorDragging = false
let editorDragStartY = 0
let editorDragStartH = 0

// AI 面板
const aiPanelOpen = ref(false)
const aiPrompt = ref('')
const aiLoading = ref(false)
const aiMessages = ref<AIMessage[]>([])
const aiStreamingContent = ref('')
const aiHistoryEl = ref<HTMLElement>()

const addForm = ref({
  name: '', type: 'mysql', host: '', port: 3306,
  username: '', password: '', database: ''
})

// ── 计算属性 ──────────────────────────────────────────
const activeSrv = computed(() => servers.value.find(s => s.id === activeSrvId.value))

const currentDbs = computed(() =>
  activeSrvId.value ? (dbMap.value[activeSrvId.value] ?? []) : []
)

const currentTables = computed(() => {
  if (!activeSrvId.value || !expandedDb.value) return []
  return tableMap.value[`${activeSrvId.value}:${expandedDb.value}`] ?? []
})

const resultColumns = computed(() =>
  lastResult.value?.columns.map(c => ({
    title: c, key: c, width: 120, ellipsis: { tooltip: true }
  })) ?? []
)

const tableData = computed(() =>
  lastResult.value?.rows.map((r, i) => ({
    _i: i,
    ...Object.fromEntries(lastResult.value!.columns.map((c, j) => [c, r[j]]))
  })) ?? []
)

// 结果区最大高度跟随编辑器高度变化（大约）
const resultAreaHeight = computed(() => Math.max(100, 400 - editorHeight.value))

// ── 生命周期 ──────────────────────────────────────────
onMounted(async () => {
  const saved = localStorage.getItem(SIDEBAR_KEY)
  if (saved) sidebarWidth.value = parseInt(saved, 10)

  const savedEH = localStorage.getItem(EDITOR_HEIGHT_KEY)
  if (savedEH) editorHeight.value = parseInt(savedEH, 10)

  const res = await dbApi.list()
  servers.value = res.data.data ?? []

  const state = EditorState.create({
    doc: 'SELECT * FROM ',
    extensions: [
      sql(),
      oneDark,
      keymap.of([
        ...defaultKeymap,
        { key: 'Ctrl-Enter', run: () => { runQuery(); return true } },
      ]),
      EditorView.theme({
        '&': { height: '100%' },
        '.cm-scroller': { overflow: 'auto' }
      }),
    ],
  })
  editorView = new EditorView({ state, parent: editorEl.value! })
})

onUnmounted(() => {
  editorView?.destroy()
})

// ── 服务器 Tab ────────────────────────────────────────
async function selectServer(srv: DBServer) {
  activeSrvId.value = srv.id
  selectedDb.value = ''
  expandedDb.value = ''
  if (!dbMap.value[srv.id]) {
    loadingDbs.value = true
    try {
      const res = await dbApi.databases(srv.id)
      dbMap.value[srv.id] = res.data.data ?? []
    } finally { loadingDbs.value = false }
  }
}

async function deleteServer(id: number) {
  await dbApi.delete(id)
  if (activeSrvId.value === id) { activeSrvId.value = null; selectedDb.value = '' }
  const res = await dbApi.list()
  servers.value = res.data.data ?? []
}

// ── 数据库 / 表 ───────────────────────────────────────
async function toggleDb(db: string) {
  if (expandedDb.value === db) {
    expandedDb.value = ''
    return
  }
  expandedDb.value = db
  selectedDb.value = db
  const key = `${activeSrvId.value}:${db}`
  if (!tableMap.value[key]) {
    loadingTables.value = true
    try {
      const res = await dbApi.tables(activeSrvId.value!, db)
      tableMap.value[key] = res.data.data ?? []
    } finally { loadingTables.value = false }
  }
}

function fillSelectAll(db: string, table: string) {
  if (!editorView) return
  const sqlStr = `SELECT * FROM ${db}.${table} LIMIT 100`
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: sqlStr }
  })
}

// ── 查询 ──────────────────────────────────────────────
async function runQuery() {
  if (!activeSrvId.value || running.value) return
  const sqlStr = editorView?.state.doc.toString().trim() ?? ''
  if (!sqlStr) return

  running.value = true
  queryError.value = ''
  lastResult.value = null

  const start = Date.now()
  try {
    const res = await dbApi.query(activeSrvId.value, sqlStr, selectedDb.value || undefined)
    lastResult.value = res.data.data
    history.value.unshift({
      sql: sqlStr,
      database: selectedDb.value,
      durationMs: Date.now() - start,
      rowsReturned: res.data.data.total,
      error: '',
      executedAt: new Date()
    })
  } catch (e: any) {
    const msg = e.response?.data?.message ?? '查询失败'
    queryError.value = msg
    history.value.unshift({
      sql: sqlStr, database: selectedDb.value,
      durationMs: Date.now() - start,
      rowsReturned: 0, error: msg, executedAt: new Date()
    })
  } finally {
    running.value = false
    if (history.value.length > 20) history.value = history.value.slice(0, 20)
  }
}

function fillFromHistory(h: QueryHistory) {
  if (!editorView) return
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: h.sql }
  })
}

// ── 拖拽（侧边栏 + 编辑器高度） ──────────────────────
function onSidebarDragStart(e: MouseEvent) {
  sidebarDragging = true
  sidebarDragStartX = e.clientX
  sidebarDragStartWidth = sidebarWidth.value
  e.preventDefault()
}

function onEditorDragStart(e: MouseEvent) {
  editorDragging = true
  editorDragStartY = e.clientY
  editorDragStartH = editorHeight.value
  e.preventDefault()
}

function onDragMove(e: MouseEvent) {
  if (sidebarDragging) {
    const next = Math.min(480, Math.max(160, sidebarDragStartWidth + e.clientX - sidebarDragStartX))
    sidebarWidth.value = next
  }
  if (editorDragging) {
    const next = Math.min(400, Math.max(80, editorDragStartH + e.clientY - editorDragStartY))
    editorHeight.value = next
  }
}

function onDragEnd() {
  if (sidebarDragging) {
    sidebarDragging = false
    localStorage.setItem(SIDEBAR_KEY, String(sidebarWidth.value))
  }
  if (editorDragging) {
    editorDragging = false
    localStorage.setItem(EDITOR_HEIGHT_KEY, String(editorHeight.value))
  }
}

// ── AI SQL 助手 ───────────────────────────────────────
function scrollAiToBottom() {
  nextTick(() => { if (aiHistoryEl.value) aiHistoryEl.value.scrollTop = aiHistoryEl.value.scrollHeight })
}

async function askAI() {
  if (!aiPrompt.value.trim() || aiLoading.value) return
  const userMsg: AIMessage = { role: 'user', content: aiPrompt.value.trim() }
  aiMessages.value.push(userMsg)
  aiPrompt.value = ''
  aiStreamingContent.value = ''
  aiLoading.value = true

  const dbContext = activeSrv.value
    ? `${activeSrv.value.type === 'mysql' ? 'MySQL' : 'PostgreSQL'}${selectedDb.value ? ` · ${selectedDb.value} 数据库` : ''}`
    : ''

  await streamSQLCommand({
    messages: aiMessages.value,
    dbContext,
    onChunk: (text) => { aiStreamingContent.value += text; scrollAiToBottom() },
    onDone: () => {
      if (aiStreamingContent.value) {
        aiMessages.value.push({ role: 'assistant', content: aiStreamingContent.value })
        aiStreamingContent.value = ''
      }
      aiLoading.value = false
      scrollAiToBottom()
    },
    onError: (err) => {
      aiMessages.value.push({ role: 'assistant', content: `错误：${err}` })
      aiStreamingContent.value = ''
      aiLoading.value = false
    },
  })
}

function insertToEditor(content: string) {
  if (!editorView) return
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: content.trim() }
  })
}

function clearAiHistory() {
  aiMessages.value = []
  aiStreamingContent.value = ''
}

// ── 添加服务器 ────────────────────────────────────────
async function addServer() {
  try {
    await dbApi.create(addForm.value as any)
    showAddModal.value = false
    addForm.value = { name: '', type: 'mysql', host: '', port: 3306, username: '', password: '', database: '' }
    const res = await dbApi.list()
    servers.value = res.data.data ?? []
  } catch (e: any) {
    alert(e.response?.data?.message ?? '添加失败')
  }
}
</script>

<style scoped>
.db-page {
  display: flex; height: 100%; overflow: hidden; user-select: none;
}

/* ── 侧边栏 ── */
.db-sidebar {
  display: flex; flex-direction: column;
  flex-shrink: 0; overflow: hidden;
  border-right: 1px solid var(--border);
  background: var(--bg-elevated);
}

.srv-tabs {
  display: flex; flex-wrap: wrap; gap: 2px;
  padding: 6px 8px; border-bottom: 1px solid var(--border);
  flex-shrink: 0; min-height: 36px;
}
.srv-tab {
  display: flex; align-items: center; gap: 4px;
  padding: 3px 8px; border-radius: 4px; cursor: pointer;
  font-size: 11px; color: var(--text-secondary);
  background: var(--bg-primary); border: 1px solid var(--border);
  white-space: nowrap; max-width: 120px;
}
.srv-tab.active {
  color: var(--accent); border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}
.srv-tab span:first-child {
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.tab-close {
  color: var(--text-secondary); font-size: 13px; line-height: 1;
  flex-shrink: 0; opacity: 0.5;
}
.tab-close:hover { opacity: 1; color: var(--danger); }
.add-srv-btn {
  background: none; border: none; cursor: pointer;
  font-size: 16px; color: var(--accent); padding: 0 4px; line-height: 1;
}

.db-tree {
  flex: 1; overflow-y: auto; overflow-x: hidden;
}
.tree-empty {
  padding: 12px; font-size: 11px; color: var(--text-secondary); text-align: center;
}
.db-row {
  display: flex; align-items: center; gap: 4px;
  padding: 5px 10px; cursor: pointer; font-size: 12px;
  color: var(--text-primary); border-left: 2px solid transparent;
}
.db-row:hover { background: color-mix(in srgb, var(--accent) 6%, transparent); }
.db-row.selected {
  color: var(--accent); border-left-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  font-weight: 600;
}
.db-expand { color: var(--text-secondary); font-size: 10px; width: 10px; flex-shrink: 0; }
.db-icon { flex-shrink: 0; }
.db-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }

.table-list { background: color-mix(in srgb, var(--bg-primary) 50%, transparent); }
.table-row {
  display: flex; align-items: center; gap: 6px;
  padding: 3px 10px 3px 28px; cursor: pointer;
  font-size: 11px; color: var(--text-secondary);
}
.table-row:hover {
  background: color-mix(in srgb, var(--accent) 6%, transparent);
  color: var(--text-primary);
}
.table-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ── 侧边栏拖拽分割线 ── */
.resize-handle {
  width: 4px; flex-shrink: 0; cursor: ew-resize;
  background: var(--border); transition: background 0.15s;
}
.resize-handle:hover { background: var(--accent); }

/* ── 右侧区域（编辑器 + AI 面板） ── */
.db-right {
  flex: 1; display: flex; flex-direction: row; overflow: hidden;
}

/* ── 主区域 ── */
.db-main {
  flex: 1; display: flex; flex-direction: column; overflow: hidden;
}
.editor-toolbar {
  display: flex; align-items: center; gap: 10px;
  padding: 6px 12px; border-bottom: 1px solid var(--border); flex-shrink: 0;
}
.ctx-badge {
  font-size: 11px; color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  padding: 2px 8px; border-radius: 10px;
}
.run-btn {
  background: var(--accent); color: #fff; border: none;
  border-radius: 4px; padding: 4px 12px; cursor: pointer; font-size: 12px;
}
.run-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.result-info { font-size: 11px; color: var(--text-secondary); margin-left: auto; }
.ai-toggle-btn {
  background: none; border: 1px solid var(--border);
  border-radius: 4px; padding: 3px 10px; cursor: pointer;
  font-size: 11px; color: var(--text-secondary);
}
.ai-toggle-btn:hover, .ai-toggle-btn.active {
  color: var(--accent); border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}

.sql-editor { flex-shrink: 0; border-bottom: 1px solid var(--border); overflow: hidden; }

/* 编辑器/结果区拖拽分割线 */
.editor-resize-handle {
  height: 4px; flex-shrink: 0; cursor: ns-resize;
  background: var(--border); transition: background 0.15s;
}
.editor-resize-handle:hover { background: var(--accent); }

.result-area { flex: 1; overflow: auto; padding: 8px; min-height: 0; }
.query-error {
  color: var(--danger); font-size: 12px;
  background: color-mix(in srgb, var(--danger) 8%, transparent);
  padding: 8px 12px; border-radius: 4px;
}

/* ── 查询历史 ── */
.history-panel {
  flex-shrink: 0; border-top: 1px solid var(--border);
  max-height: 180px; display: flex; flex-direction: column;
}
.history-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 5px 12px; font-size: 11px; font-weight: 600;
  color: var(--text-secondary); cursor: pointer; flex-shrink: 0;
  background: var(--bg-elevated);
}
.history-header:hover { color: var(--text-primary); }
.history-list { overflow-y: auto; flex: 1; }
.history-item {
  padding: 4px 12px; cursor: pointer; font-size: 11px;
  display: flex; flex-direction: column; gap: 1px;
  border-bottom: 1px solid var(--border);
}
.history-item:hover { background: color-mix(in srgb, var(--accent) 6%, transparent); }
.history-sql { color: var(--text-primary); font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.history-meta { color: var(--text-secondary); font-size: 10px; }

/* ── AI 面板 ── */
.ai-panel {
  width: 280px; flex-shrink: 0; display: flex; flex-direction: column;
  border-left: 1px solid var(--border); background: var(--bg-elevated);
  overflow: hidden;
}
.ai-panel-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 12px; border-bottom: 1px solid var(--border); flex-shrink: 0;
  font-size: 12px; font-weight: 600; color: var(--text-primary);
}
.ai-clear-btn {
  background: none; border: 1px solid var(--border); border-radius: 3px;
  font-size: 10px; color: var(--text-secondary); cursor: pointer; padding: 1px 6px;
}
.ai-clear-btn:hover { color: var(--danger); border-color: var(--danger); }
.ai-close {
  background: none; border: none; cursor: pointer;
  font-size: 16px; color: var(--text-secondary); line-height: 1;
}
.ai-close:hover { color: var(--text-primary); }
.ai-history {
  flex: 1; overflow-y: auto; padding: 8px; display: flex; flex-direction: column; gap: 8px;
}
.ai-msg { display: flex; flex-direction: column; gap: 4px; }
.ai-msg-label { font-size: 10px; font-weight: 600; color: var(--text-secondary); }
.ai-msg-user .ai-msg-label { color: var(--accent); }
.ai-code {
  margin: 0; padding: 6px 8px; border-radius: 4px; font-size: 11px;
  font-family: monospace; white-space: pre-wrap; word-break: break-all;
  background: var(--bg-primary); color: var(--text-primary); line-height: 1.5;
}
.ai-cursor { animation: blink 1s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }
.ai-insert-btn {
  align-self: flex-start; background: var(--accent); color: #fff;
  border: none; border-radius: 3px; font-size: 10px; padding: 2px 8px;
  cursor: pointer; margin-top: 2px;
}
.ai-insert-btn:hover { opacity: 0.85; }
.ai-input-row {
  padding: 8px; border-top: 1px solid var(--border); flex-shrink: 0;
  display: flex; flex-direction: column; gap: 6px;
}
.ai-textarea {
  resize: none; border: 1px solid var(--border); border-radius: 4px;
  padding: 6px 8px; font-size: 12px; font-family: inherit;
  background: var(--bg-primary); color: var(--text-primary);
  outline: none; width: 100%; box-sizing: border-box;
}
.ai-textarea:focus { border-color: var(--accent); }
.ai-send-btn {
  background: var(--accent); color: #fff; border: none;
  border-radius: 4px; padding: 5px 12px; cursor: pointer;
  font-size: 12px; align-self: flex-end;
}
.ai-send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
