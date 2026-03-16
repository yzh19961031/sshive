# Plan 8 — 跳板机（Jump Host）支持

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 支持通过跳板机（ProxyJump）连接目标主机，主机编辑表单可选择跳板机，SSH 连接自动二级跳转

**Architecture:** `hosts` 表新增 `jump_host_id` 自引用字段；SSH 会话建立时若目标主机有跳板机，先连跳板机再 Dial 目标主机，使用 `golang.org/x/crypto/ssh` 原生 ProxyJump 能力；前端主机表单加「跳板机」下拉选择器

**Tech Stack:** Go + GORM + golang.org/x/crypto/ssh；Vue 3 + Naive UI

---

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/model/host.go` | `Host` 结构体加 `JumpHostID *int64`、`JumpHost *Host`（自引用） |
| `internal/host/repo.go` | `GetByID` 预加载 `JumpHost`；`List` 支持返回跳板机信息 |
| `internal/host/service.go` | `GetWithJump(id)` 返回含跳板机凭据的完整主机信息 |
| `internal/ssh/session.go` | `buildSSHConfig` 拆分；新增 `dialWithJump` 当有跳板机时先 Dial 跳板再 Dial 目标 |
| `internal/ssh/handler.go` | `Connect` 时调用 `host.GetWithJump` 获取跳板机凭据并传给 Session |
| `internal/db/migrate.go` | AutoMigrate 触发 `jump_host_id` 列新增 |
| `frontend/src/api/host.ts` | `HostItem` 类型加 `jump_host_id?: number` |
| `frontend/src/views/HostListView.vue` | 新增/编辑主机 Modal 加「跳板机」NSelect，列表显示跳板机标记 |

---

## Chunk 1: 后端 — 数据模型 & 跳转连接

### Task 1: 数据模型扩展

**Files:**
- Modify: `internal/model/host.go`
- Modify: `internal/db/migrate.go`

- [ ] **Step 1: 修改 Host 结构体，加 JumpHostID 字段**

```go
// internal/model/host.go
type Host struct {
    ID         int64  `gorm:"primaryKey"`
    TenantID   int64  `gorm:"not null;index"`
    Name       string `gorm:"size:100;not null"`
    IP         string `gorm:"size:64;not null"`
    Port       int    `gorm:"not null;default:22"`
    AuthType   string `gorm:"size:10;not null;default:password"` // password / key
    Tags       string `gorm:"size:500"`
    JumpHostID *int64 `gorm:"index"`                             // 新增：跳板机 ID（可为空）
    JumpHost   *Host  `gorm:"foreignKey:JumpHostID"`             // 自引用预加载
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

- [ ] **Step 2: 确认 migrate.go 包含 Host，AutoMigrate 会自动加列**

`internal/db/migrate.go` 里 `AutoMigrate` 列表已有 `&model.Host{}`，无需额外改动，重启服务会自动 ALTER TABLE。

- [ ] **Step 3: 启动服务，确认 hosts 表新增 jump_host_id 列**

```bash
go run ./cmd/server/
# 观察日志无报错，连接 MySQL 确认列已加：
# SHOW COLUMNS FROM hosts;
```

- [ ] **Step 4: Commit**

```bash
git add internal/model/host.go
git commit -m "feat(jump): add jump_host_id to Host model"
```

---

### Task 2: Repo & Service 层支持跳板机

**Files:**
- Modify: `internal/host/repo.go`
- Modify: `internal/host/service.go`

- [ ] **Step 1: 修改 repo.GetByID，预加载 JumpHost**

```go
// internal/host/repo.go
func (r *Repo) GetByID(tenantID, id int64) (*model.Host, error) {
    var h model.Host
    err := r.db.Preload("JumpHost").
        Where("id = ? AND tenant_id = ?", id, tenantID).
        First(&h).Error
    return &h, err
}
```

- [ ] **Step 2: 在 service.go 新增 GetWithJumpCredential，返回跳板机解密凭据**

```go
// internal/host/service.go
type HostWithJump struct {
    Host           *model.Host
    SSHUser        string
    SSHSecret      string
    JumpHost       *model.Host // 可为 nil
    JumpSSHUser    string
    JumpSSHSecret  string
}

func (s *Service) GetWithJumpCredential(tenantID, hostID int64) (*HostWithJump, error) {
    host, err := s.repo.GetByID(tenantID, hostID)
    if err != nil {
        return nil, fmt.Errorf("get host: %w", err)
    }
    user, secret, err := s.resolveCredential(host)
    if err != nil {
        return nil, err
    }
    result := &HostWithJump{Host: host, SSHUser: user, SSHSecret: secret}

    if host.JumpHost != nil {
        jUser, jSecret, err := s.resolveCredential(host.JumpHost)
        if err != nil {
            return nil, fmt.Errorf("jump host credential: %w", err)
        }
        result.JumpHost = host.JumpHost
        result.JumpSSHUser = jUser
        result.JumpSSHSecret = jSecret
    }
    return result, nil
}

// resolveCredential 从凭据表解密，提取到私有方法复用
func (s *Service) resolveCredential(h *model.Host) (user, secret string, err error) {
    cred, err := s.credRepo.GetByHostID(h.TenantID, h.ID)
    if err != nil {
        return "", "", err
    }
    secret, err = s.enc.Decrypt(cred.EncryptedSecret)
    return cred.Username, secret, err
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/host/repo.go internal/host/service.go
git commit -m "feat(jump): repo preload JumpHost, service GetWithJumpCredential"
```

---

### Task 3: SSH Session 实现 ProxyJump

**Files:**
- Modify: `internal/ssh/session.go`
- Modify: `internal/ssh/handler.go`

- [ ] **Step 1: 扩展 Session 结构体，加跳板机字段**

```go
// internal/ssh/session.go
type Session struct {
    host          *model.Host
    sshUser       string
    sshSecret     string
    jumpHost      *model.Host // 新增
    jumpSSHUser   string      // 新增
    jumpSSHSecret string      // 新增
    tenantID      int64
    userID        int64
    clientIP      string
    ws            *websocket.Conn
    auditSvc      *audit.Service
}

func NewSession(host *model.Host, sshUser, sshSecret string,
    jumpHost *model.Host, jumpSSHUser, jumpSSHSecret string,
    tenantID, userID int64, clientIP string,
    ws *websocket.Conn, auditSvc *audit.Service) *Session {
    return &Session{
        host: host, sshUser: sshUser, sshSecret: sshSecret,
        jumpHost: jumpHost, jumpSSHUser: jumpSSHUser, jumpSSHSecret: jumpSSHSecret,
        tenantID: tenantID, userID: userID, clientIP: clientIP,
        ws: ws, auditSvc: auditSvc,
    }
}
```

- [ ] **Step 2: 修改 Run()，用 dialTarget() 替换直接 Dial**

```go
// internal/ssh/session.go
func (s *Session) Run(initWidth, initHeight int) error {
    cfg, err := s.buildSSHConfig(s.sshUser, s.sshSecret, s.host.AuthType)
    if err != nil {
        return fmt.Errorf("build ssh config: %w", err)
    }
    sshConn, err := s.dialTarget(cfg)
    if err != nil {
        return fmt.Errorf("dial target: %w", err)
    }
    defer sshConn.Close()
    // ... 其余不变
}

// dialTarget 有跳板机时二级跳转，否则直连
func (s *Session) dialTarget(targetCfg *gossh.ClientConfig) (*gossh.Client, error) {
    targetAddr := fmt.Sprintf("%s:%d", s.host.IP, s.host.Port)

    if s.jumpHost == nil {
        return gossh.Dial("tcp", targetAddr, targetCfg)
    }

    // 1. 连跳板机
    jumpCfg, err := s.buildSSHConfig(s.jumpSSHUser, s.jumpSSHSecret, s.jumpHost.AuthType)
    if err != nil {
        return nil, fmt.Errorf("jump ssh config: %w", err)
    }
    jumpAddr := fmt.Sprintf("%s:%d", s.jumpHost.IP, s.jumpHost.Port)
    jumpClient, err := gossh.Dial("tcp", jumpAddr, jumpCfg)
    if err != nil {
        return nil, fmt.Errorf("dial jump host %s: %w", jumpAddr, err)
    }

    // 2. 通过跳板机 Dial 目标
    conn, err := jumpClient.Dial("tcp", targetAddr)
    if err != nil {
        jumpClient.Close()
        return nil, fmt.Errorf("jump dial target %s: %w", targetAddr, err)
    }

    // 3. 在此连接上做 SSH 握手
    ncc, chans, reqs, err := gossh.NewClientConn(conn, targetAddr, targetCfg)
    if err != nil {
        conn.Close()
        jumpClient.Close()
        return nil, fmt.Errorf("ssh handshake: %w", err)
    }
    return gossh.NewClient(ncc, chans, reqs), nil
}

// buildSSHConfig 改为接受参数，不再用 s.sshUser 等字段（方便跳板机复用）
func (s *Session) buildSSHConfig(user, secret, authType string) (*gossh.ClientConfig, error) {
    var authMethod gossh.AuthMethod
    if authType == "key" {
        signer, err := gossh.ParsePrivateKey([]byte(secret))
        if err != nil {
            return nil, fmt.Errorf("parse private key: %w", err)
        }
        authMethod = gossh.PublicKeys(signer)
    } else {
        authMethod = gossh.Password(secret)
    }
    return &gossh.ClientConfig{
        User:            user,
        Auth:            []gossh.AuthMethod{authMethod},
        HostKeyCallback: gossh.InsecureIgnoreHostKey(),
        Timeout:         15 * time.Second,
    }, nil
}
```

- [ ] **Step 3: 修改 handler.go，改用 GetWithJumpCredential**

```go
// internal/ssh/handler.go  Connect 函数关键部分
hwj, err := h.hostSvc.GetWithJumpCredential(tenantID, hostID)
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
    return
}

sess := NewSession(
    hwj.Host, hwj.SSHUser, hwj.SSHSecret,
    hwj.JumpHost, hwj.JumpSSHUser, hwj.JumpSSHSecret,
    tenantID, userID, clientIP, ws, h.auditSvc,
)
```

- [ ] **Step 4: go build ./... 确认编译通过**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/ssh/session.go internal/ssh/handler.go internal/host/
git commit -m "feat(jump): implement ProxyJump via golang.org/x/crypto/ssh"
```

---

## Chunk 2: 前端 — 主机表单 & 列表

### Task 4: 主机表单加跳板机选择器

**Files:**
- Modify: `frontend/src/api/host.ts`
- Modify: `frontend/src/views/HostListView.vue`

- [ ] **Step 1: host.ts 类型加 jump_host_id**

```typescript
// frontend/src/api/host.ts
export interface HostItem {
  id: number
  name: string
  ip: string
  port: number
  auth_type: string
  tags: string
  jump_host_id?: number   // 新增
  jump_host?: HostItem    // 新增（预加载后返回）
}

export interface CreateHostPayload {
  name: string
  ip: string
  port: number
  auth_type: string
  credential_id: number
  tags?: string
  jump_host_id?: number  // 新增
}
```

- [ ] **Step 2: HostListView.vue 表单加跳板机 NSelect**

在新增/编辑主机的 Modal 表单里，凭据选择器之后加：

```vue
<n-form-item label="跳板机（可选）">
  <n-select
    v-model:value="form.jump_host_id"
    :options="jumpHostOptions"
    placeholder="不使用跳板机"
    clearable
  />
</n-form-item>
```

`jumpHostOptions` 计算属性从已有主机列表生成（排除自身）：

```typescript
const jumpHostOptions = computed(() =>
  hosts.value
    .filter(h => h.id !== form.id)
    .map(h => ({ label: `${h.name} (${h.ip})`, value: h.id }))
)
```

- [ ] **Step 3: 主机列表列加跳板机标识**

在主机列表 columns 里，名称列右侧加小标签：

```typescript
{
  title: '主机', key: 'name',
  render: (row: HostItem) => h('span', [
    row.name,
    row.jump_host_id
      ? h('span', { style: 'margin-left:6px;font-size:10px;color:var(--text-secondary)' }, '⤳跳板')
      : null,
  ])
}
```

- [ ] **Step 4: 前端构建**

```bash
cd frontend && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/host.ts frontend/src/views/HostListView.vue
git commit -m "feat(jump): host form jump host selector, list shows jump indicator"
```

---

## Chunk 3: 联调验收

- [ ] **Step 1: 重新编译启动**

```bash
go build -o server ./cmd/server/ && ./server
```

- [ ] **Step 2: 手动验收**
  1. 创建一台主机 A（跳板机），填入正确凭据
  2. 创建主机 B，跳板机选 A
  3. 主机列表里 B 显示「⤳跳板」标记
  4. 从 Terminal 连接主机 B，能正常打开 Shell（流量经 A → B）
  5. 连接主机 A（无跳板），正常直连

- [ ] **Step 3: 最终 Commit & Push**

```bash
git add -A
git commit -m "feat: jump host support complete"
git push
```
