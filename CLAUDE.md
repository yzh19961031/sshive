# SSHive — Claude 工作指南

## 语言约定

**始终使用中文回复用户。**

**每个 Plan 的所有任务完成并推送到远端后，在 MEMORY.md 中将对应 Plan 标记为 ✅ 已完成。**

**每次本地 commit 完成后，必须立即执行 `git push` 推送到远端仓库，不得遗漏。**

**每个 Plan 开始前必须从 main 拉一个独立分支（命名规范：`feat/planN-<简短描述>`），所有开发在该分支完成；Plan 全部任务通过后合并回 main，特性分支保留不删除。**

---

## 项目简介

多租户 WebSSH / SFTP 运维平台，支持高危指令拦截和操作审计录制。

设计文档：`docs/superpowers/specs/2026-03-11-sshive-design.md`

---

## 技术栈

- **后端**：Go + Gin + gorilla/websocket + golang.org/x/crypto/ssh
- **数据库**：MySQL 8.0+（ORM：GORM）
- **缓存**：Redis
- **前端**：Vue 3 + Vite + TypeScript + Naive UI + Pinia
- **Terminal**：xterm.js
- **审计录像**：asciinema v2 `.cast` 格式

---

## 项目结构

```
sshive/
├── cmd/server/          # 程序入口
├── internal/
│   ├── auth/            # JWT、RBAC 中间件
│   ├── tenant/          # 多租户管理
│   ├── host/            # 主机资产
│   ├── credential/      # 凭据加密存储
│   ├── ssh/             # WebSSH 会话、高危拦截
│   ├── sftp/            # SFTP 文件管理
│   ├── audit/           # 审计录制与查询
│   ├── dangerrule/      # 高危规则管理
│   └── config/          # 配置加载
├── pkg/
│   ├── encrypt/         # AES-256-GCM 加解密
│   ├── cast/            # .cast 文件写入
│   └── middleware/      # 公共中间件
├── web/                 # 前端构建产物（go:embed）
├── docs/                # 设计文档
└── CLAUDE.md
```

---

## 核心约定

### 多租户隔离
- **所有数据库查询必须带 `tenant_id` 条件**，禁止跨租户查询
- 从 JWT claims 中提取 `tenant_id`，通过 context 传递

### RBAC 角色
| 角色 | 权限范围 |
|------|---------|
| `super_admin` | 管理所有租户、用户、主机 |
| `tenant_admin` | 管理本租户的用户和主机 |
| `operator` | 连接主机（SSH/SFTP）、查看自己的审计日志 |
| `auditor` | 只读查看所有审计日志和录像 |

权限 code 列表：`host:connect` `host:manage` `sftp:access` `audit:view` `user:manage` `rule:manage`

### 凭据安全
- 密码和私钥使用 AES-256-GCM 加密，密钥从环境变量 `SSHIVE_ENCRYPT_KEY` 读取
- 凭据内容**禁止**出现在日志中

### 高危拦截
- 拦截发生在服务端，用户按回车后、命令发送前
- 正则匹配 `danger_rules` 表，`tenant_id=0` 为全局规则
- 拦截时向 WebSocket 写入红色错误提示，**不透传**给 SSH 服务器
- 触发记录写入 `danger_events` 表

### 审计录制
- 每个 SSH 会话对应一个 `.cast` 文件，路径存 `sessions.cast_file_path`
- 同步写 `session_logs`（完整 input/output）和 `session_commands`（input 命令）
- 写入通过 channel + goroutine 异步处理，**不阻塞**主 SSH 链路

### 前端主题
- 6 个内置主题：`dark-pro` `cyber-neon` `light-clean` `solarized` `dracula` `nord`
- 使用 CSS 变量实现，切换主题替换 `<html>` 的 `data-theme` 属性
- 用户偏好存 `localStorage`，登录后自动恢复

---

## 开发规范

- **错误处理**：统一返回 `{"code": int, "message": string, "data": any}`
- **分页**：统一用 `page` + `page_size` 参数，默认 `page_size=20`
- **时间**：数据库存 UTC，API 返回 RFC3339 格式
- **日志**：使用 `slog`（Go 标准库），生产环境 JSON 格式
- **测试**：业务逻辑层（internal/\*）写单元测试，handler 层写集成测试

## API 规范

所有 API 路径以 `/api` 开头，需要认证的接口带 `Authorization: Bearer <token>` 头。

WebSocket 接口：
- `/api/ws/ssh/:hostId` — WebSSH 连接
- `/api/ws/sftp/:hostId` — SFTP 连接

---

## 环境变量

```bash
SSHIVE_DB_DSN=user:pass@tcp(127.0.0.1:3306)/sshive?parseTime=true
SSHIVE_REDIS_ADDR=127.0.0.1:6379
SSHIVE_ENCRYPT_KEY=<32字节随机密钥的base64>
SSHIVE_CAST_DIR=/data/sshive/cast        # 录像存储目录
SSHIVE_JWT_SECRET=<JWT签名密钥>
SSHIVE_PORT=8080
```
