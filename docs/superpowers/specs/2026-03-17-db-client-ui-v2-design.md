# DB 客户端 UI v2 + 查询审计日志 设计文档

**日期：** 2026-03-17
**状态：** 已确认
**关联计划：** Plan 10.1

---

## 背景

Plan 10 实现了基础数据库客户端（DbView），上线后发现以下问题：
1. 侧边栏表名列表溢出到主区域（overflow bug）
2. 树形层级视觉不清晰（服务器/数据库/表三层混在一起）
3. 缺少侧边栏宽度调整能力
4. 缺少数据库选中高亮状态
5. 缺少 DB 查询审计日志

---

## 目标

1. 重构侧边栏为「服务器 Tab + 数据库/表分层」结构
2. 支持拖拽调整侧边栏宽度（持久化）
3. 数据库选中高亮 + 自动传入查询上下文
4. DbView 内嵌查询历史面板（最近 20 条，内存存储）
5. 新建独立审计页 `DbAuditView.vue`，展示全租户 DB 查询日志（需 `audit:view` 权限）

---

## 设计

### 1. 前端：侧边栏重构

**结构：**

```
[Tab: prod-mysql ×] [Tab: analytics-pg ×] [＋ 添加]
──────────────────────────────────────────────────
🗄 aqchat
🗄 sshive   ← 选中：蓝色左边框 + 浅蓝背景
  ☰ hosts
  ☰ sessions   ← hover 高亮
  ☰ users
🗄 sys
```

**关键交互：**
- 点击 Tab 切换服务器，加载该服务器的数据库列表
- 点击数据库：切换展开/折叠 + 设置 `selectedDb`（蓝色左边框高亮）
- 点击表名：填充 `SELECT * FROM <table> LIMIT 100`（不带 db 前缀，由选中的数据库上下文决定）
- Tab 上「×」删除服务器（调用 DELETE 接口）

**溢出修复：**
- 侧边栏容器 `overflow: hidden`，数据库/表列表独立设 `overflow-y: auto`
- 表名超长用 `text-overflow: ellipsis` 截断，侧边栏宽度为容器边界

### 2. 前端：拖拽调整宽度

在侧边栏和主区域之间插入 4px 宽的 `resize-handle`（cursor: ew-resize）。

实现：
- `mousedown` 记录起始 X 和当前宽度
- `mousemove`（监听 document）计算 delta，更新 CSS 变量（宽度限制 160px~480px）
- `mouseup` 清理事件监听 + 写入 `localStorage`（key: `db-sidebar-width`）
- **`onMounted` 时从 `localStorage` 读取宽度初始化**（默认 220px）

### 3. 后端：Query API 新增 database 参数

**现有接口：** `POST /api/db-servers/:id/query`

**请求体扩展：**
```json
{
  "sql": "SELECT * FROM sessions LIMIT 100",
  "database": "sshive"   // 新增，可选；若传入则 override 服务器默认库
}
```

**service.Query 签名变更：**
```go
func (s *Service) Query(tenantID, serverID, userID int64, sqlStr, database string) (*QueryResult, error)
```

`openDB` 中：若 `database != ""`，临时覆盖 `srv.Database`（每次 `GetByID` 返回新对象，无并发问题）。

**`userID` 来源：** handler 层从 JWT context 提取（`auth.GetUserID(c.Request.Context())`），传入 service。

### 4. 后端：DBQueryLog 模型

```go
// internal/model/dbquerylog.go
type DBQueryLog struct {
    ID           int64  `gorm:"primaryKey"`
    TenantID     int64  `gorm:"not null;index"`
    DBServerID   int64  `gorm:"not null;index"`
    UserID       int64  `gorm:"not null;index"`
    Database     string `gorm:"size:100"`
    SQL          string `gorm:"type:text;not null"`
    DurationMs   int64
    RowsReturned int
    ErrorMsg     string `gorm:"type:text"` // 空表示成功
    CreatedAt    time.Time
}
```

**写入时机：** `service.Query()` 执行完成（无论成功失败）后，异步 goroutine 写日志，不阻塞响应。不记录查询结果数据。

### 5. 后端：查询日志 API

```
GET /api/db-query-logs
```

权限：`auth.RequirePermission("audit:view")`

查询参数：
- `db_server_id` int（可选）
- `user_id` int（可选）
- `page` / `page_size`（默认 20）
- `start_time` / `end_time` RFC3339（可选）

响应格式（与项目现有审计 API 一致）：
```json
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 100
  }
}
```

### 6. 前端：查询历史面板（DbView 内）

底部折叠面板，不需要接口调用，纯前端内存维护：

```typescript
interface QueryHistory {
  sql: string
  database: string
  durationMs: number
  rowsReturned: number
  error: string
  executedAt: Date
}
```

- 执行成功/失败后推入 `history` ref（最多保留 20 条，超出丢弃最旧）
- 点击记录：回填 SQL 到编辑器
- 刷新页面清空（不持久化）

### 7. 前端：DbAuditView.vue（独立新页面）

新建 `frontend/src/views/DbAuditView.vue`，**不复用** AuditSessionsView。

功能：
- 列：执行时间、用户、服务器、数据库、SQL（截断 80 字符）、耗时(ms)、行数/错误标识
- 过滤：服务器下拉（调用 `GET /api/db-servers`）+ 时间范围
- 点击行展开完整 SQL（使用 NDataTable 行展开功能）
- 无数据时显示空状态提示

入口：侧边栏「安全」分组加「DB 查询审计」菜单项，路由 `/db-audit`。

### 8. 权限

复用现有权限码，不新增。调整 `auditor` 角色增加 `db:read`（用于在审计页加载服务器下拉列表）：

```go
"auditor": {"audit:view", "db:read"},  // 新增 db:read
```

| 操作 | 权限码 | 角色 |
|------|--------|------|
| 执行查询 / 管理服务器 | `db:write` | super_admin, tenant_admin, operator |
| 查看服务器列表 | `db:read` | super_admin, tenant_admin, operator, **auditor** |
| 查看全部查询审计 | `audit:view` | super_admin, tenant_admin, auditor |
| DbView 内查询历史 | — | 前端内存，无需接口权限 |

---

## 文件变更清单

| 文件 | 变更类型 |
|------|---------|
| `internal/model/dbquerylog.go` | 新建 |
| `internal/db/migrate.go` | 修改（加 DBQueryLog AutoMigrate）|
| `internal/dbclient/service.go` | 修改（Query 新增 userID/database 参数，异步写日志）|
| `internal/dbclient/repo.go` | 修改（加 DBQueryLog List/Create 方法）|
| `internal/dbclient/handler.go` | 修改（Query 传 userID/database；新增 ListQueryLogs handler）|
| `internal/auth/rbac.go` | 修改（auditor 加 db:read 权限）|
| `cmd/server/main.go` | 修改（注册 GET /api/db-query-logs 路由）|
| `frontend/src/api/db.ts` | 修改（query 加 database 参数；加 listQueryLogs 方法）|
| `frontend/src/views/DbView.vue` | 重构（侧边栏 Tab 结构、拖拽、查询历史面板）|
| `frontend/src/views/DbAuditView.vue` | 新建 |
| `frontend/src/router/index.ts` | 修改（加 /db-audit 路由）|
| `frontend/src/layouts/MainLayout.vue` | 修改（安全分组加 DB 查询审计菜单项）|

---

## 不在本次范围内

- 权限系统重构（后续单独规划）
- DB 连接编辑（Update 接口，后续补充）
- 写操作拦截（INSERT/UPDATE/DELETE 关键词检测，后续安全加固）
- 查询历史本地 localStorage 持久化
