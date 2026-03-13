# Plan 6 — Bug Fixes & UX Polish Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Fix replay auth, show host/user names in audit list, add command log viewer, collapsible grouped sidebar, dashboard home page, replay modal, and basic user management.

**Architecture:** Backend changes add a JOIN-based session list query and a stats aggregation endpoint; frontend changes replace the icon-only sidebar with a collapsible grouped nav, add a dashboard page, and overhaul the audit view with inline replay modal and command drawer.

**Tech Stack:** Go 1.21 + Gin + GORM (raw SQL for JOIN queries); Vue 3 + TypeScript + Naive UI + Pinia

---

## Chunk 1: P0 Backend Bug Fixes

### Task 1: Audit ListSessions — JOIN hosts & users tables

**Files:**
- Modify: `internal/audit/service.go`
- Modify: `internal/audit/repo.go`

- [x] **Step 1: Add `SessionListItem` struct to `internal/audit/service.go`**

Add after the existing imports, before `Service` struct:

```go
// SessionListItem is the API response shape for a session list entry.
// It adds host_name and username via JOIN so callers don't need extra round-trips.
type SessionListItem struct {
	ID           int64      `json:"id"`
	TenantID     int64      `json:"tenant_id"`
	UserID       int64      `json:"user_id"`
	HostID       int64      `json:"host_id"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at"`
	ClientIP     string     `json:"client_ip"`
	Status       string     `json:"status"`
	CastFilePath string     `json:"cast_file_path"`
	HostName     string     `json:"host_name"`
	Username     string     `json:"username"`
}
```

Also add `"time"` to imports if not already present (it already is via `StartSessionResult`).

- [x] **Step 2: Change `ListSessions` in `service.go` to return `[]SessionListItem`**

Replace:
```go
func (s *Service) ListSessions(tenantID int64, page, pageSize int) ([]model.Session, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListSessions(tenantID, page, pageSize)
}
```

With:
```go
func (s *Service) ListSessions(tenantID int64, page, pageSize int) ([]SessionListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListSessions(tenantID, page, pageSize)
}
```

- [x] **Step 3: Rewrite `ListSessions` in `repo.go` to use raw SQL JOIN**

Replace the existing `ListSessions` method:
```go
func (r *Repo) ListSessions(tenantID int64, page, pageSize int) ([]model.Session, int64, error) {
	var list []model.Session
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.Session{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
```

With:
```go
func (r *Repo) ListSessions(tenantID int64, page, pageSize int) ([]SessionListItem, int64, error) {
	var total int64
	offset := (page - 1) * pageSize
	if err := db.DB.Model(&model.Session{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SessionListItem
	err := db.DB.Raw(`
		SELECT s.id, s.tenant_id, s.user_id, s.host_id,
		       s.started_at, s.ended_at, s.client_ip, s.status, s.cast_file_path,
		       COALESCE(h.name, '') AS host_name,
		       COALESCE(u.username, '') AS username
		FROM sessions s
		LEFT JOIN hosts h ON h.id = s.host_id
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.tenant_id = ?
		ORDER BY s.started_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, pageSize, offset).Scan(&list).Error
	return list, total, err
}
```

Add `"github.com/sshive/sshive/internal/audit"` note: `SessionListItem` is in the same package so no import needed.

- [x] **Step 4: Verify the build compiles**

```bash
cd /Volumes/data/personal/code/sshive && go build ./...
```

Expected: no errors.

- [x] **Step 5: Commit**

```bash
git add internal/audit/service.go internal/audit/repo.go
git commit -m "feat: audit ListSessions returns host_name and username via JOIN"
```

---

### Task 2: Add GET /api/stats endpoint

**Files:**
- Create: `internal/stats/handler.go`
- Modify: `cmd/server/main.go`

- [x] **Step 1: Create `internal/stats/handler.go`**

```go
package stats

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/pkg/middleware"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Get(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	today := time.Now().UTC().Truncate(24 * time.Hour)

	var totalHosts int64
	db.DB.Table("hosts").Where("tenant_id = ? AND status != 0", tenantID).Count(&totalHosts)

	var todaySessions int64
	db.DB.Table("sessions").Where("tenant_id = ? AND started_at >= ?", tenantID, today).Count(&todaySessions)

	var activeSessions int64
	db.DB.Table("sessions").Where("tenant_id = ? AND status = 'active'", tenantID).Count(&activeSessions)

	var todayDangerEvents int64
	db.DB.Table("danger_events").
		Joins("JOIN sessions ON danger_events.session_id = sessions.id").
		Where("sessions.tenant_id = ? AND danger_events.triggered_at >= ?", tenantID, today).
		Count(&todayDangerEvents)

	middleware.OK(c, gin.H{
		"total_hosts":         totalHosts,
		"today_sessions":      todaySessions,
		"active_sessions":     activeSessions,
		"today_danger_events": todayDangerEvents,
	})
}
```

- [x] **Step 2: Register `GET /api/stats` in `cmd/server/main.go`**

Add import at top of main.go:
```go
"github.com/sshive/sshive/internal/stats"
```

Add handler instantiation near other handlers:
```go
statsHandler := stats.NewHandler()
```

Add route inside the `authed` group:
```go
authed.GET("/stats", statsHandler.Get)
```

- [x] **Step 3: Build to verify**

```bash
cd /Volumes/data/personal/code/sshive && go build ./...
```

Expected: no errors.

- [x] **Step 4: Commit**

```bash
git add internal/stats/handler.go cmd/server/main.go
git commit -m "feat: add GET /api/stats endpoint for dashboard metrics"
```

---

## Chunk 2: P0 Frontend Bug Fixes

### Task 3: Fix replay URL authentication

**Files:**
- Modify: `frontend/src/api/session.ts`
- Modify: `frontend/src/views/AuditReplayView.vue`

- [x] **Step 1: Update `getReplayUrl` to include token**

Replace `frontend/src/api/session.ts` with:

```typescript
// frontend/src/api/session.ts
import { http } from './client'

export const sessionApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/sessions', { params }),
  getLogs: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/logs`, { params }),
  getCommands: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/commands`, { params }),
  getReplayUrl: (id: number) => {
    const token = localStorage.getItem('token') ?? ''
    return `/api/sessions/${id}/replay?token=${encodeURIComponent(token)}`
  },
}
```

- [x] **Step 2: Build frontend to verify TypeScript**

```bash
cd /Volumes/data/personal/code/sshive/frontend && npm run build 2>&1 | head -30
```

Expected: no TypeScript errors (only the build output).

- [x] **Step 3: Commit**

```bash
git add frontend/src/api/session.ts
git commit -m "fix: append JWT token to replay URL for asciinema auth"
```

---

### Task 4: Overhaul AuditSessionsView — names, command drawer, replay modal

**Files:**
- Modify: `frontend/src/views/AuditSessionsView.vue`

- [x] **Step 1: Replace `AuditSessionsView.vue` with full implementation**

```vue
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
let playerInstance: any = null

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString('zh-CN', { hour12: false })
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
  playerInstance = AsciinemaPlayer.create(
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
```

- [x] **Step 2: Build frontend to verify**

```bash
cd /Volumes/data/personal/code/sshive/frontend && npm run build 2>&1 | grep -E "error|Error" | head -20
```

Expected: no errors.

- [x] **Step 3: Commit**

```bash
git add frontend/src/views/AuditSessionsView.vue
git commit -m "feat: audit sessions shows host/user names, adds command drawer and replay modal"
```

---

## Chunk 3: P1 Collapsible Grouped Sidebar

### Task 5: Rewrite MainLayout.vue with collapsible groups

**Files:**
- Modify: `frontend/src/layouts/MainLayout.vue`

The new sidebar has two states stored in `localStorage`:
- `sidebar-collapsed` (`"1"` = collapsed, `"0"` = expanded, default expanded)

Groups: **运维** (主机列表, 终端) | **安全** (审计日志, 高危规则) | **系统** (用户管理, 个人设置)

Expanded: 160px — icon + text + group label. Collapsed: 56px — icon only (tooltip on hover).

- [x] **Step 1: Replace `MainLayout.vue`**

```vue
<!-- frontend/src/layouts/MainLayout.vue -->
<template>
  <div class="layout">
    <nav class="sidebar" :class="{ collapsed }">
      <!-- Logo -->
      <div class="logo-row">
        <div class="logo">S</div>
        <span v-if="!collapsed" class="logo-text">SSHive</span>
      </div>

      <!-- Nav groups -->
      <div class="nav-body">
        <template v-for="group in navGroups" :key="group.label">
          <div v-if="!collapsed" class="group-label">{{ group.label }}</div>
          <div v-else class="group-divider" />
          <router-link
            v-for="item in group.items" :key="item.path"
            :to="item.path"
            class="nav-item"
            :title="collapsed ? item.label : undefined"
          >
            <span class="nav-icon">{{ item.icon }}</span>
            <span v-if="!collapsed" class="nav-text">{{ item.label }}</span>
          </router-link>
        </template>
      </div>

      <!-- Bottom: theme picker + collapse toggle + logout -->
      <div class="sidebar-bottom">
        <div class="nav-item" :title="collapsed ? '切换主题' : undefined" @click="cycleTheme">
          <span class="nav-icon">🎨</span>
          <span v-if="!collapsed" class="nav-text">{{ currentThemeLabel }}</span>
        </div>
        <div class="nav-item" :title="collapsed ? '退出登录' : undefined" @click="auth.logout()">
          <span class="nav-icon">⏻</span>
          <span v-if="!collapsed" class="nav-text">退出登录</span>
        </div>
        <div class="nav-item collapse-btn" @click="toggleCollapse"
          :title="collapsed ? '展开侧边栏' : '收起侧边栏'">
          <span class="nav-icon">{{ collapsed ? '→' : '←' }}</span>
          <span v-if="!collapsed" class="nav-text">收起</span>
        </div>
      </div>
    </nav>

    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'

const auth = useAuthStore()
const themeStore = useThemeStore()
themeStore.init()

// Sidebar collapse state
const collapsed = ref(localStorage.getItem('sidebar-collapsed') === '1')
function toggleCollapse() {
  collapsed.value = !collapsed.value
  localStorage.setItem('sidebar-collapsed', collapsed.value ? '1' : '0')
}

// Theme cycling
const currentThemeLabel = computed(() => {
  const t = themeStore.themes.find(t => t.id === themeStore.current)
  return t?.label ?? themeStore.current
})
function cycleTheme() {
  const ids = themeStore.themes.map(t => t.id)
  const idx = ids.indexOf(themeStore.current)
  themeStore.apply(ids[(idx + 1) % ids.length])
}

const navGroups = [
  {
    label: '运维',
    items: [
      { path: '/hosts',    icon: '⬡', label: '主机列表' },
      { path: '/terminal', icon: '⌨', label: 'Terminal' },
    ],
  },
  {
    label: '安全',
    items: [
      { path: '/audit/sessions',  icon: '📋', label: '审计日志' },
      { path: '/settings/rules',  icon: '⚠', label: '高危规则' },
    ],
  },
  {
    label: '系统',
    items: [
      { path: '/settings/users',   icon: '👥', label: '用户管理' },
      { path: '/settings/profile', icon: '⚙', label: '个人设置' },
    ],
  },
]
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* ── Sidebar ── */
.sidebar {
  width: 160px;
  flex-shrink: 0;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 10px 0 10px;
  transition: width 0.2s ease;
  overflow: hidden;
}
.sidebar.collapsed {
  width: 56px;
}

/* Logo row */
.logo-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  margin-bottom: 12px;
  min-height: 32px;
}
.logo {
  width: 26px; height: 26px;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 12px;
  flex-shrink: 0;
}
.logo-text {
  font-weight: 700;
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

/* Group label */
.group-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-secondary);
  padding: 10px 14px 4px;
  white-space: nowrap;
}
.group-divider {
  height: 1px;
  background: var(--border);
  margin: 8px 10px;
}

/* Nav items */
.nav-body {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 34px;
  padding: 0 12px;
  border-radius: 6px;
  margin: 1px 6px;
  cursor: pointer;
  text-decoration: none;
  color: var(--text-secondary);
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
  position: relative;
}
.nav-item:hover,
.nav-item.router-link-active {
  background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent);
}
.nav-item.router-link-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 4px;
  bottom: 4px;
  width: 2px;
  background: var(--accent);
  border-radius: 2px;
}
.nav-icon { font-size: 15px; flex-shrink: 0; }
.nav-text { font-size: 13px; overflow: hidden; text-overflow: ellipsis; }

/* Bottom section */
.sidebar-bottom {
  border-top: 1px solid var(--border);
  padding-top: 6px;
}
.collapse-btn {
  opacity: 0.6;
}
.collapse-btn:hover { opacity: 1; }

/* Main content */
.content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
}
</style>
```

- [x] **Step 2: Build to verify**

```bash
cd /Volumes/data/personal/code/sshive/frontend && npm run build 2>&1 | grep -E "error|Error" | head -20
```

Expected: no errors.

- [x] **Step 3: Commit**

```bash
git add frontend/src/layouts/MainLayout.vue
git commit -m "feat: collapsible grouped sidebar with 运维/安全/系统 groups"
```

---

## Chunk 4: P2 Dashboard

### Task 6: Stats API, DashboardView, router update

**Files:**
- Create: `frontend/src/api/stats.ts`
- Create: `frontend/src/views/DashboardView.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/stores/auth.ts`

- [x] **Step 1: Create `frontend/src/api/stats.ts`**

```typescript
// frontend/src/api/stats.ts
import { http } from './client'

export interface StatsData {
  total_hosts: number
  today_sessions: number
  active_sessions: number
  today_danger_events: number
}

export const statsApi = {
  get: () => http.get<{ code: number; data: StatsData }>('/stats'),
}
```

- [x] **Step 2: Create `frontend/src/views/DashboardView.vue`**

```vue
<!-- frontend/src/views/DashboardView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">仪表盘</h2>
    </div>

    <!-- Stat cards -->
    <div class="stat-grid">
      <div class="stat-card" v-for="card in statCards" :key="card.label">
        <div class="stat-icon">{{ card.icon }}</div>
        <div class="stat-info">
          <div class="stat-value">{{ card.value }}</div>
          <div class="stat-label">{{ card.label }}</div>
        </div>
      </div>
    </div>

    <!-- Bottom two panels -->
    <div class="bottom-grid">
      <!-- Recent sessions -->
      <div class="panel">
        <div class="panel-title">最近会话</div>
        <n-data-table :columns="sessionColumns" :data="recentSessions"
          :loading="sessionsLoading" :pagination="false" size="small" />
      </div>

      <!-- Quick connect -->
      <div class="panel">
        <div class="panel-title">快捷连接</div>
        <div v-if="quickHosts.length === 0" class="empty-tip">暂无最近连接</div>
        <div v-else class="quick-grid">
          <div v-for="h in quickHosts" :key="h.host_id"
            class="quick-card" @click="openTerminal(h.host_id)">
            <span class="quick-icon">⬡</span>
            <span class="quick-name">{{ h.host_name || h.host_id }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDataTable, NTag } from 'naive-ui'
import { statsApi } from '@/api/stats'
import { sessionApi } from '@/api/session'

const router = useRouter()

// Stats
const stats = ref({ total_hosts: 0, today_sessions: 0, active_sessions: 0, today_danger_events: 0 })
const statCards = computed(() => [
  { icon: '⬡', label: '主机总数',    value: stats.value.total_hosts },
  { icon: '💻', label: '今日会话',    value: stats.value.today_sessions },
  { icon: '🔗', label: '活跃会话',    value: stats.value.active_sessions },
  { icon: '⚠', label: '今日高危拦截', value: stats.value.today_danger_events },
])

// Recent sessions
const recentSessions = ref<any[]>([])
const sessionsLoading = ref(false)
const sessionColumns = [
  { title: '主机', key: 'host_name', render: (row: any) => row.host_name || String(row.host_id) },
  { title: '用户', key: 'username',  render: (row: any) => row.username || String(row.user_id) },
  {
    title: '状态', key: 'status', width: 70,
    render: (row: any) => h(NTag, { type: row.status === 'active' ? 'success' : 'default', size: 'small' },
      { default: () => row.status })
  },
  { title: '开始时间', key: 'started_at', render: (row: any) => new Date(row.started_at).toLocaleString() },
]

// Quick connect hosts (deduplicated from recent sessions)
const quickHosts = computed(() => {
  const seen = new Set<number>()
  const result: any[] = []
  for (const s of recentSessions.value) {
    if (!seen.has(s.host_id)) {
      seen.add(s.host_id)
      result.push(s)
    }
    if (result.length >= 5) break
  }
  return result
})

function openTerminal(hostId: number) {
  router.push({ path: '/terminal', query: { hostId: String(hostId) } })
}

async function loadData() {
  sessionsLoading.value = true
  try {
    const [statsRes, sessRes] = await Promise.all([
      statsApi.get(),
      sessionApi.list({ page: 1, page_size: 5 }),
    ])
    stats.value = statsRes.data.data ?? stats.value
    recentSessions.value = sessRes.data.data?.list ?? []
  } finally {
    sessionsLoading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}
.stat-card {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 18px;
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon { font-size: 28px; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }

.bottom-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.panel {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
}
.panel-title { font-size: 13px; font-weight: 600; margin-bottom: 12px; color: var(--text-secondary); }

.quick-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; }
.quick-card {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: border-color 0.15s, color 0.15s;
}
.quick-card:hover { border-color: var(--accent); color: var(--accent); }
.quick-icon { font-size: 18px; }
.quick-name { font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.empty-tip { color: var(--text-secondary); text-align: center; padding: 20px 0; font-size: 13px; }
</style>
```

- [x] **Step 3: Add `/dashboard` route and update default redirect in `router/index.ts`**

Replace the router file:

```typescript
// frontend/src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: () => import('@/views/DashboardView.vue') },
      { path: 'hosts', component: () => import('@/views/HostListView.vue') },
      { path: 'terminal', component: () => import('@/views/TerminalView.vue') },
      { path: 'sftp/:hostId', component: () => import('@/views/SftpView.vue') },
      { path: 'audit/sessions', component: () => import('@/views/AuditSessionsView.vue') },
      { path: 'audit/replay/:id', component: () => import('@/views/AuditReplayView.vue') },
      { path: 'settings/rules', component: () => import('@/views/DangerRulesView.vue') },
      { path: 'settings/users', component: () => import('@/views/UsersView.vue') },
      { path: 'settings/profile', component: () => import('@/views/ProfileView.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
})
```

- [x] **Step 4: Update login redirect in `stores/auth.ts`**

In `frontend/src/stores/auth.ts`, change `router.push('/hosts')` to `router.push('/dashboard')`.

- [x] **Step 5: Add Dashboard nav item to sidebar**

In `frontend/src/layouts/MainLayout.vue`, add dashboard to the `运维` group's items:

```typescript
{
  label: '运维',
  items: [
    { path: '/dashboard', icon: '📊', label: '仪表盘' },
    { path: '/hosts',     icon: '⬡',  label: '主机列表' },
    { path: '/terminal',  icon: '⌨',  label: 'Terminal' },
  ],
},
```

- [x] **Step 6: Build to verify**

```bash
cd /Volumes/data/personal/code/sshive/frontend && npm run build 2>&1 | grep -E "error|Error" | head -20
```

Expected: no errors.

- [x] **Step 7: Commit**

```bash
git add frontend/src/api/stats.ts frontend/src/views/DashboardView.vue \
        frontend/src/router/index.ts frontend/src/stores/auth.ts \
        frontend/src/layouts/MainLayout.vue
git commit -m "feat: add dashboard page with stats cards, recent sessions, and quick connect"
```

---

## Chunk 5: P4 User Management

### Task 7: Backend — user update, disable, and roles list

**Files:**
- Modify: `internal/user/repo.go`
- Modify: `internal/user/service.go`
- Modify: `internal/user/handler.go`
- Modify: `cmd/server/main.go`

- [x] **Step 1: Add `Update` and `Disable` to `repo.go`**

Add after the existing `SetRoles` method in `internal/user/repo.go`:

```go
func (r *Repo) Update(tenantID, userID int64, fields map[string]any) error {
	return db.DB.Model(&model.User{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Updates(fields).Error
}

func (r *Repo) ListRoles(tenantID int64) ([]model.Role, error) {
	var roles []model.Role
	err := db.DB.Where("tenant_id = ?", tenantID).Find(&roles).Error
	return roles, err
}
```

- [x] **Step 2: Add `UpdateUser`, `DisableUser`, `ListRoles` to `service.go`**

Add after `AssignRoles` in `internal/user/service.go`:

```go
type UpdateUserReq struct {
	Email  string `json:"email"`
	Status *int8  `json:"status"`
}

func (s *Service) UpdateUser(tenantID, userID int64, req UpdateUserReq) error {
	if _, err := s.repo.GetByID(tenantID, userID); err != nil {
		return fmt.Errorf("user not found")
	}
	fields := map[string]any{}
	if req.Email != "" {
		fields["email"] = req.Email
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(tenantID, userID, fields)
}

func (s *Service) DisableUser(tenantID, userID int64) error {
	if _, err := s.repo.GetByID(tenantID, userID); err != nil {
		return fmt.Errorf("user not found")
	}
	return s.repo.Update(tenantID, userID, map[string]any{"status": int8(2)})
}

func (s *Service) ListRoles(tenantID int64) ([]model.Role, error) {
	return s.repo.ListRoles(tenantID)
}
```

- [x] **Step 3: Add `Update`, `Disable`, `GetRoles` handlers to `handler.go`**

Add after `AssignRoles` in `internal/user/handler.go`:

```go
func (h *Handler) Update(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid user id")
		return
	}
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdateUser(tenantID, userID, req); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *Handler) Disable(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid user id")
		return
	}
	if err := h.svc.DisableUser(tenantID, userID); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *Handler) GetRoles(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	roles, err := h.svc.ListRoles(tenantID)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, roles)
}
```

- [x] **Step 4: Register new routes in `cmd/server/main.go`**

In the `authed` group section for users, add after `PUT /api/users/:id/roles`:

```go
authed.PUT("/users/:id", userHandler.Update)
authed.DELETE("/users/:id", userHandler.Disable)
authed.GET("/roles", userHandler.GetRoles)
```

All three need `"user:manage"` permission. Wrap them in the existing permission middleware pattern:

```go
userMgmt := authed.Group("", middleware.RequirePermission("user:manage"))
userMgmt.GET("/users", userHandler.List)
userMgmt.POST("/users", userHandler.Create)
userMgmt.PUT("/users/:id/roles", userHandler.AssignRoles)
userMgmt.PUT("/users/:id", userHandler.Update)
userMgmt.DELETE("/users/:id", userHandler.Disable)
userMgmt.GET("/roles", userHandler.GetRoles)
```

Note: Look at the existing route registration pattern in `cmd/server/main.go` to see how `RequirePermission` is used, and follow the same pattern exactly. The exact structure depends on how the routes are currently grouped.

- [x] **Step 5: Build backend**

```bash
cd /Volumes/data/personal/code/sshive && go build ./...
```

Expected: no errors.

- [x] **Step 6: Commit**

```bash
git add internal/user/repo.go internal/user/service.go internal/user/handler.go cmd/server/main.go
git commit -m "feat: add user update/disable and GET /api/roles endpoints"
```

---

### Task 8: Frontend — UsersView full implementation

**Files:**
- Modify: `frontend/src/views/UsersView.vue`
- Create: `frontend/src/api/user.ts`

- [x] **Step 1: Create `frontend/src/api/user.ts`**

```typescript
// frontend/src/api/user.ts
import { http } from './client'

export interface UserItem {
  id: number
  username: string
  email: string
  status: number
  created_at: string
}

export interface RoleItem {
  id: number
  name: string
  description: string
}

export const userApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/users', { params }),
  create: (data: { username: string; password: string; email?: string }) =>
    http.post('/users', data),
  update: (id: number, data: { email?: string; status?: number }) =>
    http.put(`/users/${id}`, data),
  disable: (id: number) =>
    http.delete(`/users/${id}`),
  assignRoles: (id: number, role_ids: number[]) =>
    http.put(`/users/${id}/roles`, { role_ids }),
  listRoles: () =>
    http.get('/roles'),
}
```

- [x] **Step 2: Replace `UsersView.vue` with full implementation**

```vue
<!-- frontend/src/views/UsersView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
      <n-button type="primary" size="small" @click="openCreate">+ 新建用户</n-button>
    </div>

    <n-data-table :columns="columns" :data="users" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage" />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="modalVisible" preset="card" :title="editingId ? '编辑用户' : '新建用户'"
      style="width:440px" :mask-closable="false">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80px">
        <n-form-item v-if="!editingId" label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="2-50字符" />
        </n-form-item>
        <n-form-item v-if="!editingId" label="密码" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="至少8位" show-password-on="click" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="可选" />
        </n-form-item>
        <n-form-item v-if="editingId" label="状态" path="status">
          <n-select v-model:value="form.status" :options="statusOptions" />
        </n-form-item>
        <n-form-item label="角色" path="role_ids">
          <n-select v-model:value="form.role_ids" multiple :options="roleOptions"
            placeholder="选择角色" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div style="display:flex;justify-content:flex-end;gap:8px">
          <n-button @click="modalVisible = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NDataTable, NButton, NTag, NModal, NForm, NFormItem, NInput, NSelect } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { userApi } from '@/api/user'

const users = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

// Modal state
const modalVisible = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInst>()
const form = ref({ username: '', password: '', email: '', status: 1, role_ids: [] as number[] })

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 2 },
]

const roleOptions = ref<{ label: string; value: number }[]>([])

const rules: FormRules = {
  username: [{ required: true, min: 2, message: '用户名至少2位', trigger: 'blur' }],
  password: [{ required: true, min: 8, message: '密码至少8位', trigger: 'blur' }],
}

function openCreate() {
  editingId.value = null
  form.value = { username: '', password: '', email: '', status: 1, role_ids: [] }
  modalVisible.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  form.value = { username: row.username, password: '', email: row.email ?? '', status: row.status, role_ids: [] }
  modalVisible.value = true
}

async function save() {
  try { await formRef.value?.validate() } catch { return }
  saving.value = true
  try {
    if (editingId.value) {
      await userApi.update(editingId.value, { email: form.value.email, status: form.value.status })
      if (form.value.role_ids.length > 0) {
        await userApi.assignRoles(editingId.value, form.value.role_ids)
      }
    } else {
      const res = await userApi.create({ username: form.value.username, password: form.value.password, email: form.value.email })
      const newId = res.data.data?.id
      if (newId && form.value.role_ids.length > 0) {
        await userApi.assignRoles(newId, form.value.role_ids)
      }
    }
    modalVisible.value = false
    loadPage(pagination.value.page)
  } finally {
    saving.value = false
  }
}

async function disableUser(id: number) {
  await userApi.disable(id)
  loadPage(pagination.value.page)
}

const columns = [
  { title: '用户名', key: 'username' },
  { title: '邮箱', key: 'email', render: (row: any) => row.email || '—' },
  {
    title: '状态', key: 'status', width: 80,
    render: (row: any) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' },
      { default: () => row.status === 1 ? '启用' : '禁用' })
  },
  { title: '创建时间', key: 'created_at', render: (row: any) => new Date(row.created_at).toLocaleString() },
  {
    title: '操作', key: 'actions',
    render: (row: any) => h('div', { style: 'display:flex;gap:6px' }, [
      h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => '编辑' }),
      row.status === 1
        ? h(NButton, { size: 'tiny', type: 'error', onClick: () => disableUser(row.id) }, { default: () => '禁用' })
        : null,
    ])
  },
]

async function loadPage(page: number) {
  loading.value = true
  const res = await userApi.list({ page, page_size: 20 })
  users.value = res.data.data?.list ?? []
  pagination.value.itemCount = res.data.data?.total ?? 0
  pagination.value.page = page
  loading.value = false
}

async function loadRoles() {
  const res = await userApi.listRoles()
  const list: any[] = res.data.data ?? []
  roleOptions.value = list.map(r => ({ label: r.description || r.name, value: r.id }))
}

onMounted(() => {
  loadPage(1)
  loadRoles()
})
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
```

- [x] **Step 3: Build frontend**

```bash
cd /Volumes/data/personal/code/sshive/frontend && npm run build 2>&1 | grep -E "error|Error" | head -20
```

Expected: no errors.

- [x] **Step 4: Build full binary**

```bash
cd /Volumes/data/personal/code/sshive && npm run build --prefix frontend && go build ./...
```

Expected: no errors.

- [x] **Step 5: Commit**

```bash
git add frontend/src/api/user.ts frontend/src/views/UsersView.vue
git commit -m "feat: full user management UI with create/edit/disable and role assignment"
```

---

## Final Steps

- [x] **Build full binary and run smoke test**

```bash
cd /Volumes/data/personal/code/sshive/frontend && npm run build
cd /Volumes/data/personal/code/sshive && go build -o sshive ./cmd/server && echo "build ok"
```

- [x] **Final commit and push**

```bash
git log --oneline -8
git push origin feat/plan6-polish
```
