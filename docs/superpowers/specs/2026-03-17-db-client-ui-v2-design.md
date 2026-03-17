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

1. 重构侧边栏为「服务器 Tab + 数据库/表分层」结构（方案 B）
2. 支持拖拽调整侧边栏宽度
3. 数据库选中高亮 + 自动传入查询上下文
4. DbView 内嵌查询历史面板（最近 20 条）
5. 审计日志页新增「数据库查询」Tab（需 `audit:view` 权限）

---

## 设计

### 1. 前端：侧边栏重构（方案 B）

**结构变化：**

```
旧：
  服务器行（可展开）
    数据库行（可展开）
      表行

新：
  [服务器 Tab: prod-mysql] [analytics-pg] [＋]
  ─────────────────────────────────────────
  🗄 aqchat
  🗄 sshive  ◀（选中：蓝色左边框 + 背景高亮）
    ├ ☰ hosts
    ├ ☰ sessions  ◀（hover 高亮）
    └ ☰ users
  🗄 sys
```

**关键交互：**
- 点击服务器 Tab 切换，加载该服务器的数据库列表
- 点击数据库：展开表列表 + 高亮选中（`selectedDb` ref）+ 自动将当前数据库设为查询默认上下文
- 点击表名：填充 `SELECT * FROM <table> LIMIT 100`（不带 db 前缀，因为连接时已指定数据库）
- 服务器 Tab 右侧的「＋」按钮打开添加服务器 Modal

**溢出修复：**
- 侧边栏容器 `overflow: hidden`，内部数据库/表列表用独立 `overflow-y: auto` 容器
- 表名超长用 `text-overflow: ellipsis` 截断

### 2. 前端：拖拽调整宽度

在侧边栏与主区域之间插入 `<div class="resize-handle">`（宽 4px，cursor: ew-resize）。

实现：
- `mousedown` 记录起始 X 和当前宽度
- `mousemove`（监听 document）计算 delta，更新侧边栏宽度（限制 160px~480px）
- `mouseup` 清理监听，持久化宽度到 `localStorage`（key: `db-sidebar-width`）

### 3. 前端：查询历史面板

DbView 底部增加折叠面板（默认折叠）：
- 标题栏：「查询历史 (N)」，点击展开/折叠
- 内容：最近 20 条查询记录，每条显示 SQL 前 80 字符 + 耗时 + 行数 + 时间
- 点击记录：回填到编辑器
- 历史记录存储在内存（`ref<QueryHistory[]>`），不做本地持久化（刷新清空）

### 4. 后端：db_query_logs 表

```go
type DBQueryLog struct {
    ID          int64  `gorm:"primaryKey"`
    TenantID    int64  `gorm:"not null;index"`
    DBServerID  int64  `gorm:"not null;index"`
    UserID      int64  `gorm:"not null;index"`
    Database    string `gorm:"size:100"`
    SQL         string `gorm:"type:text;not null"`
    DurationMs  int64  // 执行耗时（毫秒）
    RowsReturned int
    ErrorMsg    string `gorm:"type:text"`  // 空表示成功
    CreatedAt   time.Time
}
```

写入时机：`service.Query()` 执行完成后，异步 goroutine 写日志（不阻塞响应）。
写入内容：不记录查询结果，只记录元信息。

### 5. 后端：查询日志 API

```
GET /api/db-query-logs
```

权限：`audit:view`（审计员和管理员可查看全部，operator 调用返回 403）

查询参数：
- `db_server_id` — 按服务器过滤
- `user_id` — 按用户过滤
- `page` / `page_size` — 分页，默认 20
- `start_time` / `end_time` — 时间范围（RFC3339）

响应：标准 `{"code":0,"data":[...],"total":N}`

### 6. 前端：审计日志页新 Tab

在 `AuditSessionsView.vue`（或新建 `DbAuditView.vue`）添加「数据库查询」Tab：

- 列：时间、用户、服务器、数据库、SQL（截断）、耗时、行数/错误
- 支持按服务器和时间范围过滤
- 点击行展开完整 SQL

入口：侧边栏「安全」分组加「DB 查询审计」菜单项，路由 `/db-audit`。

---

## 权限设计

复用现有权限码，不新增：

| 操作 | 权限码 | 角色 |
|------|--------|------|
| 执行查询 / 管理服务器 | `db:write` | super_admin, tenant_admin, operator |
| 查看服务器列表 | `db:read` | super_admin, tenant_admin, operator |
| 查看全部查询审计 | `audit:view` | super_admin, tenant_admin, auditor |
| DbView 内查询历史 | — | 前端内存，无需权限接口 |

> 注：operator 无法查看其他人的查询日志（403），只能在 DbView 里看自己当前 session 的历史。

---

## 文件变更清单

| 文件 | 变更类型 |
|------|---------|
| `internal/model/dbquerylog.go` | 新建 |
| `internal/db/migrate.go` | 修改（加 DBQueryLog）|
| `internal/dbclient/service.go` | 修改（Query 后异步写日志）|
| `internal/dbclient/repo.go` | 修改（加 DBQueryLog CRUD）|
| `internal/dbclient/handler.go` | 修改（加 ListQueryLogs handler）|
| `cmd/server/main.go` | 修改（注册 /api/db-query-logs 路由）|
| `frontend/src/views/DbView.vue` | 重构（侧边栏、拖拽、查询历史）|
| `frontend/src/views/DbAuditView.vue` | 新建 |
| `frontend/src/api/db.ts` | 修改（加 queryLogs API）|
| `frontend/src/router/index.ts` | 修改（加 /db-audit 路由）|
| `frontend/src/layouts/MainLayout.vue` | 修改（加菜单项）|

---

## 不在本次范围内

- 权限系统重构（用户提到后续单独规划）
- DB 连接编辑（Update 接口，后续补充）
- 查询历史本地持久化（localStorage，后续视需求添加）
- 写操作拦截（INSERT/UPDATE/DELETE 关键词检测，后续安全加固）
