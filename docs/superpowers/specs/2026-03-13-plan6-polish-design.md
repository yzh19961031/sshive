# SSHive Plan 6 — Bug 修复 & 体验优化设计文档

**日期**：2026-03-13
**分支**：`feat/plan6-polish`
**优先级**：Bug 修复 > UI 优化 > 用户管理

---

## 一、Bug 修复

### 1.1 审计回放鉴权失败

**现状**：`AsciinemaPlayer.create('/api/sessions/:id/replay', container)` 直接 fetch URL，浏览器不携带 `Authorization` header，后端 JWT 中间件返回 401，播放器无法加载。

**方案**：与 WebSocket 修复保持一致，replay URL 追加 `?token=<jwt>`。

- **后端**：`Replay` handler 无需改动，JWT middleware 已支持 `?token=` query 参数
- **前端**：`sessionApi.getReplayUrl(id)` 改为从 auth store 读取 token 并追加

### 1.2 审计列表显示原始 ID

**现状**：`host_id`、`user_id` 以数字显示，无法直观看出是谁在哪台机器上操作。

**方案**：后端 `ListSessions` JOIN `hosts` 和 `users` 表，response 额外返回 `host_name`、`username`。前端直接渲染，无需额外请求。

- **后端**：`audit/repo.go` 的 `ListSessions` 查询加 LEFT JOIN，新增 `HostName`、`Username` 字段到 `SessionListItem` 结构体
- **前端**：`AuditSessionsView` 列定义替换 ID 列为名称列

### 1.3 命令日志无查看入口

**现状**：`GET /api/sessions/:id/commands` 接口已存在，前端没有入口。

**方案**：审计列表每行加「命令」按钮，点击右侧滑出 `n-drawer`，展示命令列表（时间戳 + 命令内容）。

---

## 二、可折叠侧边栏

### 2.1 导航分组

| 分组 | 菜单项 | 路由 |
|------|--------|------|
| 运维 | 主机列表 | `/hosts` |
| 运维 | 终端 | `/terminal` |
| 安全 | 审计日志 | `/audit/sessions` |
| 安全 | 高危规则 | `/settings/rules` |
| 系统 | 用户管理 | `/settings/users` |
| 系统 | 个人设置 | `/settings/profile` |

### 2.2 折叠行为

- **展开**（默认，160px）：图标 + 文字 + 分组标题
- **收缩**（56px）：仅图标，hover 显示 tooltip
- 折叠状态存 `localStorage`（key: `sidebar-collapsed`）
- 分组标题可单独展开/折叠（状态存 localStorage）
- 折叠按钮位于侧边栏底部

### 2.3 视觉规范

- 当前激活项：`var(--accent)` 背景色 + 左侧 2px 高亮条
- 分组标题：小号灰色文字，折叠时隐藏
- 过渡动画：`transition: width 0.2s ease`

---

## 三、仪表盘首页

**路由**：`/dashboard`（登录后默认跳转）

### 3.1 统计卡片（顶部 4 个）

| 卡片 | 数据来源 |
|------|---------|
| 主机总数 | `GET /api/hosts?page_size=1` → total |
| 今日会话 | `GET /api/stats` → today_sessions |
| 活跃会话 | `GET /api/stats` → active_sessions |
| 今日高危拦截 | `GET /api/stats` → today_danger_events |

### 3.2 新增后端接口

`GET /api/stats`（需认证，tenant 隔离）

```json
{
  "code": 0,
  "data": {
    "total_hosts": 12,
    "today_sessions": 34,
    "active_sessions": 2,
    "today_danger_events": 5
  }
}
```

查询逻辑：
- `total_hosts`：`SELECT COUNT(*) FROM hosts WHERE tenant_id=?`
- `today_sessions`：`SELECT COUNT(*) FROM sessions WHERE tenant_id=? AND started_at >= today`
- `active_sessions`：`SELECT COUNT(*) FROM sessions WHERE tenant_id=? AND status='active'`
- `today_danger_events`：`SELECT COUNT(*) FROM danger_events de JOIN sessions s ON de.session_id=s.id WHERE s.tenant_id=? AND de.triggered_at >= today`

### 3.3 最近会话列表（左下）

展示最近 5 条会话，列：主机名、用户名、开始时间、状态、操作（命令/回放）。
复用 `ListSessions` 接口，`page_size=5`。

### 3.4 快捷连接（右下）

展示最近连接的主机（取最近 5 条会话的去重主机），卡片点击直接打开终端。
数据从最近会话中提取，无需新接口。

---

## 四、审计回放 & 命令查看

### 4.1 回放 Modal

- 触发：审计列表点击「回放」按钮
- 组件：`n-modal`，宽 90vw，高 85vh，`mask-closable`
- 内容：`AsciinemaPlayer`，URL = `/api/sessions/:id/replay?token=<jwt>`
- 标题：`主机名 · 用户名 · 开始时间`
- 会话结束（`cast_file_path` 为空）时显示「录像暂不可用」提示

### 4.2 命令历史 Drawer

- 触发：审计列表 / 仪表盘最近会话点击「命令」按钮
- 组件：`n-drawer`，从右侧滑入，宽 400px
- 列表：时间戳（`HH:mm:ss`）+ 命令内容，等宽字体
- 分页：page_size=100，一次加载，无需翻页
- 空状态：「暂无命令记录」

---

## 五、用户管理（低优先级）

### 5.1 用户列表

表格列：用户名、邮箱、角色、状态、创建时间、操作（编辑/禁用）

### 5.2 新建/编辑用户

Modal 表单：用户名、邮箱、密码（新建必填）、角色（多选）

### 5.3 角色管理

独立 Tab：角色列表，支持新建角色、查看角色权限。
（权限 code 列表固定，不支持自定义）

### 5.4 后端补充接口

- `PUT /api/users/:id` — 编辑用户（邮箱、状态）
- `DELETE /api/users/:id` — 禁用用户（软删除，status=2）
- `GET /api/roles` — 角色列表

---

## 六、实施顺序

| 优先级 | 内容 |
|--------|------|
| P0 | Bug 修复（回放鉴权、审计名称、命令入口） |
| P1 | 可折叠侧边栏 + 分组 |
| P2 | 仪表盘首页（/api/stats + Dashboard 页面） |
| P3 | 审计回放 Modal + 命令 Drawer 整合进审计页 |
| P4 | 用户管理完整版 |

---

## 七、不在此计划范围内

- 主机编辑功能（主机列表只有删除，暂不加编辑）
- 凭据管理独立页面
- 多语言支持
- 移动端适配
