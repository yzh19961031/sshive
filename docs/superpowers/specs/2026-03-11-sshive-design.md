# SSHive 设计文档

**日期：** 2026-03-11
**状态：** 已确认
**项目名：** SSHive — 多租户 WebSSH / SFTP 运维平台

---

## 1. 项目概述

SSHive 是一个基于 Web 的多租户运维平台，提供 WebSSH 终端、SFTP 文件管理、高危指令拦截和操作审计录制功能。目标覆盖企业内部运维、云服务商托管、个人/小团队三类场景。

---

## 2. 技术栈

| 层次 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| WebSocket | gorilla/websocket |
| SSH/SFTP 协议 | golang.org/x/crypto/ssh |
| 数据库 | MySQL 8.0+ |
| 缓存 | Redis |
| 审计录像存储 | 本地磁盘 / S3 兼容存储（.cast 格式） |
| 前端框架 | Vue 3 + Vite + TypeScript |
| Terminal 渲染 | xterm.js |
| 前端 UI 库 | Naive UI |
| 前端状态管理 | Pinia |

---

## 3. 整体架构

### 3.1 后端模块划分

```
sshive/
├── cmd/
│   └── server/              # 程序入口，读取配置，启动 HTTP 服务
├── internal/
│   ├── auth/                # 登录、JWT 签发与验证、RBAC 中间件
│   ├── tenant/              # 多租户管理（租户 CRUD、租户上下文注入）
│   ├── host/                # 主机资产管理（CRUD、连通性检测）
│   ├── credential/          # 凭据管理（密码/密钥加密存储/解密）
│   ├── ssh/                 # WebSSH 会话（WebSocket 处理、PTY、高危拦截）
│   ├── sftp/                # SFTP 文件管理（上传/下载/删除/重命名/浏览）
│   ├── audit/               # 审计录制写入、会话日志查询
│   ├── dangerrule/          # 高危指令规则管理（CRUD、正则编译缓存）
│   └── config/              # 配置加载（数据库、Redis、存储路径等）
├── pkg/
│   ├── encrypt/             # AES 加解密工具（凭据存储用）
│   ├── cast/                # asciinema .cast 格式写入工具
│   └── middleware/          # 公共 HTTP 中间件
└── web/                     # 前端构建产物（go:embed 嵌入二进制）
```

### 3.2 WebSSH 数据流

```
Browser (xterm.js)
    │  WebSocket  ws://.../api/ws/ssh/:hostId
    ▼
Go SSH Handler
    ├── 1. 验证 JWT + 检查 host:connect 权限
    ├── 2. 从 credential 取凭据并 AES 解密
    ├── 3. 建立 SSH 连接，开启 PTY
    │
    ├── 输入流（Browser → 服务器）
    │       └── 高危拦截中间件
    │               ├── 缓冲输入，遇到回车触发正则检测
    │               ├── 命中规则 → 拦截，返回错误提示给 Browser
    │               └── 未命中 → 透传给 SSH 服务器
    │
    └── 输出流（服务器 → Browser）
            ├── 透传给 xterm.js 渲染
            └── 异步写入审计录制（channel + goroutine）
                    ├── 写 .cast 文件（完整回放）
                    ├── 写 session_logs 表（input/output 原始内容）
                    └── 写 session_commands 表（input 命令，便于检索）
```

---

## 4. 数据库设计（MySQL）

### 4.1 多租户 & 用户权限

```sql
-- 租户
CREATE TABLE tenants (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(100) NOT NULL,
    status      TINYINT NOT NULL DEFAULT 1,  -- 1:正常 0:禁用
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 用户
CREATE TABLE users (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id     BIGINT NOT NULL,
    username      VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email         VARCHAR(100),
    status        TINYINT NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_username (tenant_id, username)
);

-- 角色（预置：super_admin / tenant_admin / operator / auditor）
CREATE TABLE roles (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id   BIGINT NOT NULL,
    name        VARCHAR(50) NOT NULL,
    description VARCHAR(255)
);

-- 用户角色关联（多对多，为细粒度升级预留）
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

-- 权限（第一期硬编码，第二期入库）
CREATE TABLE permissions (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    code        VARCHAR(100) NOT NULL UNIQUE,  -- e.g. host:connect
    description VARCHAR(255)
);

-- 角色权限关联
CREATE TABLE role_permissions (
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);
```

### 4.2 主机资产 & 凭据

```sql
-- 主机
CREATE TABLE hosts (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id     BIGINT NOT NULL,
    name          VARCHAR(100) NOT NULL,
    ip            VARCHAR(45) NOT NULL,
    port          SMALLINT NOT NULL DEFAULT 22,
    auth_type     VARCHAR(20) NOT NULL,  -- password / key
    credential_id BIGINT NOT NULL,
    status        TINYINT NOT NULL DEFAULT 1,
    tags          JSON,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 凭据（密码/私钥 AES 加密存储）
CREATE TABLE credentials (
    id               BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id        BIGINT NOT NULL,
    name             VARCHAR(100) NOT NULL,
    type             VARCHAR(20) NOT NULL,  -- password / key
    username         VARCHAR(50) NOT NULL,
    encrypted_secret TEXT NOT NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 4.3 审计日志

```sql
-- 会话记录
CREATE TABLE sessions (
    id             BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id      BIGINT NOT NULL,
    user_id        BIGINT NOT NULL,
    host_id        BIGINT NOT NULL,
    started_at     DATETIME NOT NULL,
    ended_at       DATETIME,
    client_ip      VARCHAR(45),
    status         VARCHAR(20) NOT NULL DEFAULT 'active',  -- active / closed
    cast_file_path VARCHAR(500)
);

-- 完整输入输出记录（审计用）
CREATE TABLE session_logs (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL,
    type       VARCHAR(10) NOT NULL,  -- input / output
    content    MEDIUMTEXT NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_session_id (session_id)
);

-- 命令记录（输入命令单独存，便于检索）
CREATE TABLE session_commands (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL,
    command    TEXT NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_session_id (session_id)
);

-- 高危指令触发记录
CREATE TABLE danger_events (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id   BIGINT NOT NULL,
    matched_rule VARCHAR(255) NOT NULL,
    command      TEXT NOT NULL,
    triggered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 4.4 高危指令规则

```sql
-- 规则（支持租户级自定义）
CREATE TABLE danger_rules (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id   BIGINT NOT NULL,               -- 0 表示全局规则
    pattern     VARCHAR(500) NOT NULL,          -- 正则表达式
    description VARCHAR(255),
    action      VARCHAR(20) NOT NULL DEFAULT 'block',
    enabled     TINYINT NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 预置规则示例：
-- rm\s+-rf\s+/
-- :(){ :|:& };:
-- dd\s+if=.*of=/dev/
-- shutdown|reboot|halt
-- mkfs\..*
```

---

## 5. API 设计（REST）

```
POST   /api/auth/login               # 登录，返回 JWT
POST   /api/auth/logout              # 登出

GET    /api/tenants                  # 租户列表（super_admin）
POST   /api/tenants                  # 创建租户

GET    /api/users                    # 用户列表
POST   /api/users                    # 创建用户
PUT    /api/users/:id/roles          # 分配角色

GET    /api/hosts                    # 主机列表
POST   /api/hosts                    # 添加主机
PUT    /api/hosts/:id                # 编辑主机
DELETE /api/hosts/:id                # 删除主机

GET    /api/credentials              # 凭据列表
POST   /api/credentials              # 添加凭据

GET    /api/sessions                 # 会话列表
GET    /api/sessions/:id/logs        # 会话日志
GET    /api/sessions/:id/commands    # 会话命令记录
GET    /api/sessions/:id/replay      # 获取 .cast 文件（回放）

GET    /api/danger-rules             # 高危规则列表
POST   /api/danger-rules             # 创建规则
PUT    /api/danger-rules/:id         # 编辑规则
DELETE /api/danger-rules/:id         # 删除规则

WS     /api/ws/ssh/:hostId           # WebSSH 连接
WS     /api/ws/sftp/:hostId          # SFTP 连接
```

---

## 6. 前端设计

### 6.1 内置主题（6个）

| 主题名 | 风格 |
|--------|------|
| 暗黑专业 | 深色 #0d1117，蓝色主色 #1f6feb（GitHub Dark 风） |
| 赛博霓虹 | 极深黑 #070b14，青色霓虹 #00e5ff |
| 浅色简约 | 白色背景，紫色主色 #6366f1 |
| Solarized Dark | 深青 #002b36，经典护眼配色 |
| Dracula | 深紫 #282a36，高饱和紫绿配色 |
| Nord | 冷蓝灰 #2e3440，北欧极简风 |

实现：CSS 变量 + Vue 响应式，切换主题替换根节点 class，偏好保存 localStorage。

### 6.2 页面结构

```
整体布局：左侧图标窄边栏（44px） + 右侧内容区

页面列表：
├── /hosts              主机列表（分组 + 卡片/列表视图切换）
├── /terminal           WebSSH Terminal（Tab 模式 + 分屏）
├── /sftp/:hostId       SFTP 文件管理
├── /audit/sessions     审计会话列表
├── /audit/replay/:id   会话录像回放（xterm.js）
├── /settings/rules     高危指令规则配置
├── /settings/users     用户管理
├── /settings/hosts     主机管理
└── /settings/profile   个人设置（主题切换在此）
```

### 6.3 Terminal 核心交互

- 默认 Tab 模式，每个 SSH 会话一个 Tab，可关闭
- 工具栏提供分屏按钮，支持 2/4 分屏
- 高危指令触发时，Terminal 内显示红色拦截提示
- 支持 xterm.js 全屏模式

---

## 7. 安全设计

- 凭据（密码/私钥）AES-256-GCM 加密存储，密钥从环境变量读取
- JWT 有效期 8 小时，Redis 维护黑名单支持强制登出
- 高危拦截在服务端执行，客户端无法绕过
- 多租户数据隔离：所有查询强制带 tenant_id 条件
- SFTP 操作记录在 session_logs（type=sftp）

---

## 8. 第一期交付范围

- [x] 多租户管理（租户/用户/角色 CRUD）
- [x] RBAC 权限控制（4个固定角色）
- [x] 主机 & 凭据管理
- [x] WebSSH Terminal（Tab + 分屏）
- [x] 高危指令拦截（正则规则，租户可配置）
- [x] 操作审计（.cast 录制 + session_logs + session_commands）
- [x] 审计回放（xterm.js）
- [x] SFTP 基础文件管理
- [x] 前端 6 主题切换

## 9. 第二期规划

- [ ] 批量执行（多主机同时执行命令）
- [ ] SSH 密钥统一管理与下发
- [ ] 告警通知（邮件 / 企业微信 / 钉钉 Webhook）
- [ ] RBAC 细粒度升级（自定义角色 + 权限入库）
- [ ] 主机健康检测（定时连通性探测）
