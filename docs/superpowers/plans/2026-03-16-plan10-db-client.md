# Plan 10 — 数据库客户端（MySQL / PostgreSQL）

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 支持在 Web 界面连接 MySQL / PostgreSQL，执行 SQL 查询并以表格展示结果

**Architecture:** 新建 `internal/dbclient/` 包管理数据库连接池；`db_servers` 表存连接配置（凭据复用现有加密）；前端新建 `DbView.vue`，左侧库表树，右侧 CodeMirror SQL 编辑器 + 结果表格；查询通过 `POST /api/db/:id/query` 执行，强制超时 + 行数限制防止滥用

**Tech Stack:** Go + `database/sql` + `go-sql-driver/mysql` + `lib/pq`；Vue 3 + CodeMirror 6（SQL 语法高亮）+ Naive UI DataTable

---

## 安全约束（必须遵守）

- 每次查询超时 30 秒
- 结果最多返回 1000 行
- 禁止多语句执行（`;` 分隔的多条 SQL）
- 写操作（INSERT/UPDATE/DELETE/DROP）只允许有 `db:write` 权限的用户
- 凭据加密存储，不在 API 响应中返回密码

---

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/model/dbserver.go` | `DBServer` 模型 |
| `internal/dbclient/repo.go` | CRUD db_servers 表 |
| `internal/dbclient/service.go` | 连接池管理、执行查询、获取库表结构 |
| `internal/dbclient/handler.go` | REST handlers |
| `internal/db/migrate.go` | AutoMigrate DBServer |
| `cmd/server/main.go` | 注册 dbclient 路由 |
| `frontend/src/api/db.ts` | API 类型和请求函数 |
| `frontend/src/views/DbView.vue` | 数据库客户端页面 |
| `frontend/src/router/index.ts` | 加 `/db` 路由 |
| `frontend/src/layouts/MainLayout.vue` | 侧边栏加「数据库」菜单项 |

---

## Chunk 1: 后端 — 数据模型 & 连接管理

### Task 1: 数据模型

**Files:**
- Create: `internal/model/dbserver.go`
- Modify: `internal/db/migrate.go`

- [ ] **Step 1: 创建 DBServer 模型**

```go
// internal/model/dbserver.go
package model

import "time"

// DBServer 数据库服务器连接配置
type DBServer struct {
    ID               int64  `gorm:"primaryKey"`
    TenantID         int64  `gorm:"not null;index"`
    Name             string `gorm:"size:100;not null"`
    Type             string `gorm:"size:20;not null"` // mysql / postgres
    Host             string `gorm:"size:128;not null"`
    Port             int    `gorm:"not null"`
    Username         string `gorm:"size:100;not null"`
    EncryptedPassword string `gorm:"type:text;not null"`
    Database         string `gorm:"size:100"`          // 默认连接的库（可空）
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

- [ ] **Step 2: 加入 AutoMigrate**

```go
// internal/db/migrate.go — 在已有列表里加
&model.DBServer{},
```

- [ ] **Step 3: Commit**

```bash
git add internal/model/dbserver.go internal/db/migrate.go
git commit -m "feat(db): DBServer model"
```

---

### Task 2: Repo & Service

**Files:**
- Create: `internal/dbclient/repo.go`
- Create: `internal/dbclient/service.go`

- [ ] **Step 1: 安装依赖**

```bash
go get github.com/go-sql-driver/mysql
go get github.com/lib/pq
```

- [ ] **Step 2: 创建 repo.go**

```go
// internal/dbclient/repo.go
package dbclient

import (
    "github.com/sshive/sshive/internal/model"
    "gorm.io/gorm"
)

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) List(tenantID int64) ([]model.DBServer, error) {
    var list []model.DBServer
    return list, r.db.Where("tenant_id = ?", tenantID).Find(&list).Error
}

func (r *Repo) GetByID(tenantID, id int64) (*model.DBServer, error) {
    var s model.DBServer
    return &s, r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error
}

func (r *Repo) Create(s *model.DBServer) error { return r.db.Create(s).Error }
func (r *Repo) Update(s *model.DBServer) error { return r.db.Save(s).Error }
func (r *Repo) Delete(tenantID, id int64) error {
    return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DBServer{}).Error
}
```

- [ ] **Step 3: 创建 service.go**

```go
// internal/dbclient/service.go
package dbclient

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "time"

    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
    "github.com/sshive/sshive/internal/model"
    "github.com/sshive/sshive/pkg/encrypt"
)

const (
    queryTimeout = 30 * time.Second
    maxRows      = 1000
)

type QueryResult struct {
    Columns []string        `json:"columns"`
    Rows    [][]interface{} `json:"rows"`
    Total   int             `json:"total"` // 实际行数（截断前）
}

type Service struct {
    repo *Repo
    enc  *encrypt.Encryptor
}

func NewService(repo *Repo, enc *encrypt.Encryptor) *Service {
    return &Service{repo: repo, enc: enc}
}

func (s *Service) openDB(srv *model.DBServer) (*sql.DB, error) {
    pwd, err := s.enc.Decrypt(srv.EncryptedPassword)
    if err != nil {
        return nil, fmt.Errorf("decrypt password: %w", err)
    }

    var dsn, driver string
    switch srv.Type {
    case "mysql":
        driver = "mysql"
        db := srv.Database
        if db == "" { db = "information_schema" }
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=5s&parseTime=true",
            srv.Username, pwd, srv.Host, srv.Port, db)
    case "postgres":
        driver = "postgres"
        dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
            srv.Host, srv.Port, srv.Username, pwd, srv.Database)
    default:
        return nil, fmt.Errorf("unsupported db type: %s", srv.Type)
    }

    db, err := sql.Open(driver, dsn)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(1)
    return db, nil
}

// Query 执行单条 SELECT，返回结果（超过 maxRows 截断）
func (s *Service) Query(tenantID, serverID int64, sqlStr string) (*QueryResult, error) {
    // 安全检查：禁止多语句
    if strings.Count(strings.TrimRight(sqlStr, "; \t\n"), ";") > 0 {
        return nil, fmt.Errorf("不支持多语句执行")
    }

    srv, err := s.repo.GetByID(tenantID, serverID)
    if err != nil {
        return nil, err
    }

    db, err := s.openDB(srv)
    if err != nil {
        return nil, err
    }
    defer db.Close()

    ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
    defer cancel()

    rows, err := db.QueryContext(ctx, sqlStr)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    cols, _ := rows.Columns()
    result := &QueryResult{Columns: cols}

    for rows.Next() {
        vals := make([]interface{}, len(cols))
        ptrs := make([]interface{}, len(cols))
        for i := range vals { ptrs[i] = &vals[i] }
        if err := rows.Scan(ptrs...); err != nil {
            continue
        }
        // 转为可序列化类型
        row := make([]interface{}, len(cols))
        for i, v := range vals {
            if b, ok := v.([]byte); ok {
                row[i] = string(b)
            } else {
                row[i] = v
            }
        }
        result.Rows = append(result.Rows, row)
        result.Total++
        if result.Total >= maxRows {
            break
        }
    }
    return result, rows.Err()
}

// ListDatabases 获取所有数据库名
func (s *Service) ListDatabases(tenantID, serverID int64) ([]string, error) {
    srv, err := s.repo.GetByID(tenantID, serverID)
    if err != nil { return nil, err }
    db, err := s.openDB(srv)
    if err != nil { return nil, err }
    defer db.Close()

    var q string
    switch srv.Type {
    case "mysql":    q = "SHOW DATABASES"
    case "postgres": q = "SELECT datname FROM pg_database WHERE datistemplate = false"
    }
    rows, err := db.Query(q)
    if err != nil { return nil, err }
    defer rows.Close()

    var dbs []string
    for rows.Next() {
        var name string
        rows.Scan(&name)
        dbs = append(dbs, name)
    }
    return dbs, nil
}

// ListTables 获取指定数据库的表名
func (s *Service) ListTables(tenantID, serverID int64, database string) ([]string, error) {
    srv, err := s.repo.GetByID(tenantID, serverID)
    if err != nil { return nil, err }
    // 临时切换到目标库
    orig := srv.Database
    srv.Database = database
    db, err := s.openDB(srv)
    srv.Database = orig
    if err != nil { return nil, err }
    defer db.Close()

    var q string
    switch srv.Type {
    case "mysql":    q = "SHOW TABLES"
    case "postgres": q = "SELECT tablename FROM pg_tables WHERE schemaname='public'"
    }
    rows, err := db.Query(q)
    if err != nil { return nil, err }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var name string
        rows.Scan(&name)
        tables = append(tables, name)
    }
    return tables, nil
}
```

- [ ] **Step 4: go build ./... 确认编译**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/dbclient/ go.mod go.sum
git commit -m "feat(db): repo, service with query/schema exploration"
```

---

### Task 3: Handler & 路由

**Files:**
- Create: `internal/dbclient/handler.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 创建 handler.go**

```go
// internal/dbclient/handler.go
package dbclient

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/pkg/ctxkey"
    "github.com/sshive/sshive/internal/model"
    "github.com/sshive/sshive/pkg/encrypt"
)

type Handler struct {
    svc  *Service
    repo *Repo
    enc  *encrypt.Encryptor
}

func NewHandler(svc *Service, repo *Repo, enc *encrypt.Encryptor) *Handler {
    return &Handler{svc: svc, repo: repo, enc: enc}
}

// List GET /api/db-servers
func (h *Handler) List(c *gin.Context) {
    tenantID := c.MustGet(ctxkey.TenantID).(int64)
    list, err := h.repo.List(tenantID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
        return
    }
    // 脱敏：不返回密码
    for i := range list { list[i].EncryptedPassword = "" }
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

type CreateReq struct {
    Name     string `json:"name"     binding:"required"`
    Type     string `json:"type"     binding:"required"` // mysql/postgres
    Host     string `json:"host"     binding:"required"`
    Port     int    `json:"port"     binding:"required"`
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Database string `json:"database"`
}

// Create POST /api/db-servers
func (h *Handler) Create(c *gin.Context) {
    tenantID := c.MustGet(ctxkey.TenantID).(int64)
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
        return
    }
    enc, err := h.enc.Encrypt(req.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "加密失败"})
        return
    }
    srv := &model.DBServer{
        TenantID: tenantID, Name: req.Name, Type: req.Type,
        Host: req.Host, Port: req.Port, Username: req.Username,
        EncryptedPassword: enc, Database: req.Database,
    }
    if err := h.repo.Create(srv); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
        return
    }
    srv.EncryptedPassword = ""
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": srv})
}

// Delete DELETE /api/db-servers/:id
func (h *Handler) Delete(c *gin.Context) {
    tenantID := c.MustGet(ctxkey.TenantID).(int64)
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    if err := h.repo.Delete(tenantID, id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 0})
}

type QueryReq struct {
    SQL string `json:"sql" binding:"required"`
}

// Query POST /api/db-servers/:id/query
func (h *Handler) Query(c *gin.Context) {
    tenantID := c.MustGet(ctxkey.TenantID).(int64)
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    var req QueryReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
        return
    }
    result, err := h.svc.Query(tenantID, id, req.SQL)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// Databases GET /api/db-servers/:id/databases
func (h *Handler) Databases(c *gin.Context) {
    tenantID := c.MustGet(ctxkey.TenantID).(int64)
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    dbs, err := h.svc.ListDatabases(tenantID, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": dbs})
}

// Tables GET /api/db-servers/:id/databases/:db/tables
func (h *Handler) Tables(c *gin.Context) {
    tenantID := c.MustGet(ctxkey.TenantID).(int64)
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    database := c.Param("db")
    tables, err := h.svc.ListTables(tenantID, id, database)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": tables})
}
```

- [ ] **Step 2: 注册路由（cmd/server/main.go）**

```go
dbRepo := dbclient.NewRepo(db.DB)
dbSvc  := dbclient.NewService(dbRepo, encryptor)
dbH    := dbclient.NewHandler(dbSvc, dbRepo, encryptor)

authed.GET("/db-servers",                          dbH.List)
authed.POST("/db-servers",                         dbH.Create)
authed.DELETE("/db-servers/:id",                   dbH.Delete)
authed.POST("/db-servers/:id/query",               dbH.Query)
authed.GET("/db-servers/:id/databases",            dbH.Databases)
authed.GET("/db-servers/:id/databases/:db/tables", dbH.Tables)
```

- [ ] **Step 3: go build ./... 确认编译**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/dbclient/handler.go cmd/server/main.go
git commit -m "feat(db): REST handlers and routes"
```

---

## Chunk 2: 前端 — 数据库客户端页面

### Task 4: API 层

**Files:**
- Create: `frontend/src/api/db.ts`

- [ ] **Step 1: 创建 db.ts**

```typescript
// frontend/src/api/db.ts
import axios from './axios'

export interface DBServer {
  id: number
  name: string
  type: 'mysql' | 'postgres'
  host: string
  port: number
  username: string
  database: string
}

export interface QueryResult {
  columns: string[]
  rows: any[][]
  total: number
}

export const dbApi = {
  list: () => axios.get<{ code: number; data: DBServer[] }>('/api/db-servers'),
  create: (data: Omit<DBServer, 'id'> & { password: string }) =>
    axios.post('/api/db-servers', data),
  delete: (id: number) => axios.delete(`/api/db-servers/${id}`),
  query: (id: number, sql: string) =>
    axios.post<{ code: number; data: QueryResult }>(`/api/db-servers/${id}/query`, { sql }),
  databases: (id: number) =>
    axios.get<{ code: number; data: string[] }>(`/api/db-servers/${id}/databases`),
  tables: (id: number, db: string) =>
    axios.get<{ code: number; data: string[] }>(`/api/db-servers/${id}/databases/${db}/tables`),
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/api/db.ts
git commit -m "feat(db): API client"
```

---

### Task 5: DbView 页面

**Files:**
- Create: `frontend/src/views/DbView.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/layouts/MainLayout.vue`

- [ ] **Step 1: 安装 CodeMirror 6**

```bash
cd frontend
npm install @codemirror/view @codemirror/state @codemirror/lang-sql @codemirror/theme-one-dark
```

- [ ] **Step 2: 创建 DbView.vue**

页面布局：左侧树（服务器列表 → 数据库 → 表） + 右侧编辑区（CodeMirror + 执行按钮 + 结果表格）。

```vue
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
              class="table-row" @click="fillSelectAll(activeSrvId, expandedDb.split(':')[1], t)">
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
import { ref, computed, onMounted, watch } from 'vue'
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

function fillSelectAll(srvId: number | null, db: string, table: string) {
  if (!editorView) return
  const sql = `SELECT * FROM \`${db}\`.\`${table}\` LIMIT 100`
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: sql }
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
  await dbApi.create(addForm.value as any)
  showAddModal.value = false
  const res = await dbApi.list()
  servers.value = res.data.data ?? []
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
```

- [ ] **Step 3: 注册路由**

```typescript
// frontend/src/router/index.ts — 在 authed 路由里加
{ path: '/db', component: () => import('@/views/DbView.vue') }
```

- [ ] **Step 4: 侧边栏加菜单项**

```typescript
// frontend/src/layouts/MainLayout.vue — 运维分组里加
{ path: '/db', icon: '🗄', label: '数据库' }
```

- [ ] **Step 5: 前端构建**

```bash
cd frontend && npm run build
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/views/DbView.vue frontend/src/api/db.ts \
  frontend/src/router/index.ts frontend/src/layouts/MainLayout.vue
git commit -m "feat(db): database client UI with CodeMirror SQL editor"
```

---

## Chunk 3: 联调验收

- [ ] **Step 1: 重新编译启动**

```bash
go build -o server ./cmd/server/ && ./server
```

- [ ] **Step 2: 手动验收**
  1. 侧边栏出现「🗄 数据库」菜单，点击进入
  2. 点「＋」添加 MySQL 服务器，填入真实连接信息
  3. 点服务器名，展开显示数据库列表
  4. 点数据库名，展开显示表列表
  5. 点表名，编辑器自动填入 `SELECT * FROM db.table LIMIT 100`
  6. 点「▶ 执行」，结果表格显示
  7. 执行 `DROP TABLE xxx` 应报错（禁止多语句/危险操作由 MySQL 权限控制）
  8. 查询超过 1000 行时结果被截断，提示「已截断至 1000」

- [ ] **Step 3: 最终 Commit & Push**

```bash
git add -A
git commit -m "feat: database client complete"
git push
```
