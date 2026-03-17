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
            <!-- 数据库行 -->
            <div
              class="db-row"
              :class="{ selected: selectedDb === db }"
              @click="toggleDb(db)"
            >
              <span class="db-expand">{{ expandedDb === db ? '▾' : '▸' }}</span>
              <span class="db-icon">🗄</span>
              <span class="db-name">{{ db }}</span>
            </div>
            <!-- 表列表（内联展开） -->
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

    <!-- 拖拽分割线 -->
    <div class="resize-handle" @mousedown="onDragStart" />

    <!-- 右侧主区域 -->
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
      </div>

      <!-- CodeMirror 编辑器 -->
      <div ref="editorEl" class="sql-editor" />

      <!-- 结果表格 -->
      <div class="result-area">
        <div v-if="queryError" class="query-error">{{ queryError }}</div>
        <n-data-table
          v-else-if="lastResult"
          :columns="resultColumns"
          :data="tableData"
          :max-height="280"
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { NModal, NForm, NFormItem, NInput, NInputNumber, NButton, NDataTable, NSelect } from 'naive-ui'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { sql } from '@codemirror/lang-sql'
import { oneDark } from '@codemirror/theme-one-dark'
import { defaultKeymap } from '@codemirror/commands'
import { dbApi, type DBServer, type QueryResult, type QueryHistory } from '@/api/db'

// ── 状态 ──────────────────────────────────────────────
const servers     = ref<DBServer[]>([])
const activeSrvId = ref<number | null>(null)
const selectedDb  = ref<string>('')
const expandedDb  = ref<string>('')
const dbMap       = ref<Record<number, string[]>>({})     // srvId -> db[]
const tableMap    = ref<Record<string, string[]>>({})      // "srvId:db" -> table[]
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

// 拖拽
const SIDEBAR_KEY = 'db-sidebar-width'
const sidebarWidth = ref(220)
let dragging = false
let dragStartX = 0
let dragStartWidth = 0

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

// ── 生命周期 ──────────────────────────────────────────
onMounted(async () => {
  // 恢复侧边栏宽度
  const saved = localStorage.getItem(SIDEBAR_KEY)
  if (saved) sidebarWidth.value = parseInt(saved, 10)

  // 加载服务器列表
  const res = await dbApi.list()
  servers.value = res.data.data ?? []

  // 初始化 CodeMirror
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
        '&': { height: '150px' },
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

// ── 拖拽调宽 ──────────────────────────────────────────
function onDragStart(e: MouseEvent) {
  dragging = true
  dragStartX = e.clientX
  dragStartWidth = sidebarWidth.value
  e.preventDefault()
}

function onDragMove(e: MouseEvent) {
  if (!dragging) return
  const next = Math.min(480, Math.max(160, dragStartWidth + e.clientX - dragStartX))
  sidebarWidth.value = next
}

function onDragEnd() {
  if (!dragging) return
  dragging = false
  localStorage.setItem(SIDEBAR_KEY, String(sidebarWidth.value))
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

/* 服务器 Tab 栏 */
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

/* 数据库 / 表树 */
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

/* ── 拖拽分割线 ── */
.resize-handle {
  width: 4px; flex-shrink: 0; cursor: ew-resize;
  background: var(--border);
  transition: background 0.15s;
}
.resize-handle:hover { background: var(--accent); }

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

.sql-editor { flex-shrink: 0; border-bottom: 1px solid var(--border); }

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
</style>
