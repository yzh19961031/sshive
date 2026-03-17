<!-- frontend/src/views/DbView.vue -->
<template>
  <div class="db-page">
    <!-- 左侧：服务器 & 库表树 -->
    <div class="db-sidebar">
      <div class="db-sidebar-header">
        <span>数据库</span>
        <button class="add-btn" @click="showAddModal = true" title="添加数据库服务器">＋</button>
      </div>
      <div v-for="srv in servers" :key="srv.id" class="srv-item">
        <div class="srv-row" @click="toggleServer(srv)">
          <span class="srv-icon">{{ srv.type === 'mysql' ? '🐬' : '🐘' }}</span>
          <span class="srv-name">{{ srv.name }}</span>
          <span class="srv-expand">{{ expandedSrv === srv.id ? '▾' : '▸' }}</span>
        </div>
        <div v-if="expandedSrv === srv.id" class="db-list">
          <div v-if="loadingDbs" style="padding:4px 12px;font-size:11px;color:var(--text-secondary)">加载中…</div>
          <div v-for="db in dbMap[srv.id]" :key="db" class="db-row"
            @click="toggleDb(srv.id, db)">
            <span>🗄 {{ db }}</span>
            <span>{{ expandedDb === `${srv.id}:${db}` ? '▾' : '▸' }}</span>
          </div>
          <div v-if="expandedDb?.startsWith(`${srv.id}:`) && tableMap[expandedDb]">
            <div v-for="t in tableMap[expandedDb]" :key="t"
              class="table-row" @click="fillSelectAll(activeSrvId, expandedDb!.split(':')[1]!, t)">
              ☰ {{ t }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧：编辑器 + 结果 -->
    <div class="db-main">
      <div class="editor-toolbar">
        <n-select v-model:value="activeSrvId" :options="serverOptions"
          placeholder="选择服务器" size="small" style="width:200px" />
        <button class="run-btn" :disabled="running" @click="runQuery">
          {{ running ? '执行中…' : '▶ 执行  Ctrl+Enter' }}
        </button>
        <span v-if="result" class="result-info">
          {{ result.total }} 行{{ result.total >= 1000 ? '（已截断至 1000）' : '' }}
        </span>
      </div>

      <!-- CodeMirror 编辑器挂载点 -->
      <div ref="editorEl" class="sql-editor" />

      <!-- 结果表格 -->
      <div class="result-area">
        <div v-if="queryError" class="query-error">{{ queryError }}</div>
        <n-data-table v-else-if="result"
          :columns="resultColumns"
          :data="result.rows.map((r, i) => ({ _i: i, ...Object.fromEntries(result!.columns.map((c, j) => [c, r[j]])) }))"
          :max-height="300"
          size="small"
          :scroll-x="result.columns.length * 120"
        />
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
        <n-form-item label="密码"><n-input v-model:value="addForm.password" type="password" /></n-form-item>
        <n-form-item label="默认库"><n-input v-model:value="addForm.database" placeholder="可选" /></n-form-item>
      </n-form>
      <template #footer>
        <n-button type="primary" @click="addServer">确定</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NSelect, NModal, NForm, NFormItem, NInput, NInputNumber, NButton, NDataTable } from 'naive-ui'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { sql } from '@codemirror/lang-sql'
import { oneDark } from '@codemirror/theme-one-dark'
import { defaultKeymap } from '@codemirror/commands'
import { dbApi, type DBServer, type QueryResult } from '@/api/db'

const servers = ref<DBServer[]>([])
const activeSrvId = ref<number | null>(null)
const expandedSrv = ref<number | null>(null)
const expandedDb = ref<string | null>(null)
const dbMap = ref<Record<number, string[]>>({})
const tableMap = ref<Record<string, string[]>>({})
const loadingDbs = ref(false)
const result = ref<QueryResult | null>(null)
const queryError = ref('')
const running = ref(false)
const showAddModal = ref(false)
const editorEl = ref<HTMLElement>()
let editorView: EditorView | null = null

const addForm = ref({ name: '', type: 'mysql', host: '', port: 3306, username: '', password: '', database: '' })

const serverOptions = computed(() =>
  servers.value.map(s => ({ label: `${s.name} (${s.type})`, value: s.id }))
)

const resultColumns = computed(() =>
  result.value?.columns.map(c => ({ title: c, key: c, width: 120, ellipsis: { tooltip: true } })) ?? []
)

onMounted(async () => {
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
      EditorView.theme({ '&': { height: '160px' }, '.cm-scroller': { overflow: 'auto' } }),
    ],
  })
  editorView = new EditorView({ state, parent: editorEl.value! })
})

async function toggleServer(srv: DBServer) {
  if (expandedSrv.value === srv.id) { expandedSrv.value = null; return }
  expandedSrv.value = srv.id
  activeSrvId.value = srv.id
  if (!dbMap.value[srv.id]) {
    loadingDbs.value = true
    try {
      const res = await dbApi.databases(srv.id)
      dbMap.value[srv.id] = res.data.data ?? []
    } finally { loadingDbs.value = false }
  }
}

async function toggleDb(srvId: number, db: string) {
  const key = `${srvId}:${db}`
  if (expandedDb.value === key) { expandedDb.value = null; return }
  expandedDb.value = key
  if (!tableMap.value[key]) {
    const res = await dbApi.tables(srvId, db)
    tableMap.value[key] = res.data.data ?? []
  }
}

function fillSelectAll(_srvId: number | null, db: string, table: string) {
  if (!editorView) return
  const sqlStr = `SELECT * FROM ${db}.${table} LIMIT 100`
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: sqlStr }
  })
}

async function runQuery() {
  if (!activeSrvId.value || running.value) return
  const sqlStr = editorView?.state.doc.toString().trim() ?? ''
  if (!sqlStr) return
  running.value = true
  queryError.value = ''
  result.value = null
  try {
    const res = await dbApi.query(activeSrvId.value, sqlStr)
    result.value = res.data.data
  } catch (e: any) {
    queryError.value = e.response?.data?.message ?? '查询失败'
  } finally { running.value = false }
}

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
.db-page { display: flex; height: 100%; overflow: hidden; }
.db-sidebar {
  width: 220px; flex-shrink: 0; overflow-y: auto;
  border-right: 1px solid var(--border);
  background: var(--bg-elevated);
}
.db-sidebar-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 12px; font-size: 13px; font-weight: 600;
  border-bottom: 1px solid var(--border);
}
.add-btn { background: none; border: none; cursor: pointer; font-size: 18px;
  color: var(--accent); line-height: 1; }
.srv-row, .db-row, .table-row {
  display: flex; align-items: center; gap: 6px; padding: 5px 12px;
  cursor: pointer; font-size: 12px; color: var(--text-primary);
}
.srv-row:hover, .db-row:hover, .table-row:hover {
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}
.srv-expand, .db-row span:last-child { margin-left: auto; color: var(--text-secondary); }
.db-list { padding-left: 8px; }
.db-row { padding-left: 20px; }
.table-row { padding-left: 32px; color: var(--text-secondary); }

.db-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.editor-toolbar {
  display: flex; align-items: center; gap: 10px; padding: 8px 12px;
  border-bottom: 1px solid var(--border); flex-shrink: 0;
}
.run-btn {
  background: var(--accent); color: #fff; border: none;
  border-radius: 4px; padding: 4px 12px; cursor: pointer; font-size: 12px;
}
.run-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.result-info { font-size: 12px; color: var(--text-secondary); margin-left: auto; }
.sql-editor { flex-shrink: 0; border-bottom: 1px solid var(--border); }
.result-area { flex: 1; overflow: auto; padding: 8px; }
.query-error {
  color: var(--danger); font-size: 12px;
  background: color-mix(in srgb, var(--danger) 8%, transparent);
  padding: 8px 12px; border-radius: 4px;
}
</style>
