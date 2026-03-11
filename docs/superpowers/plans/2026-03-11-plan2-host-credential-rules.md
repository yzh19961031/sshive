# SSHive Plan 2: 主机管理 + 凭据 + 高危规则

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现主机资产管理、凭据加密存储/解密、高危指令规则 CRUD，输出完整的 REST API。

**Architecture:** 依赖 Plan 1 的基础设施（db、config、auth、middleware）。Credential 使用 AES-256-GCM 加密存储，密钥从 config.C.EncryptKey 解码。DangerRule 维护编译好的 regexp 内存缓存供 Plan 3 的拦截器使用。

**Tech Stack:** Go, Gin, GORM, pkg/encrypt（Plan 1 实现）

**前置条件:** Plan 1 已完成，`go build ./...` 通过。

---

## 文件结构

```
internal/
├── model/
│   ├── host.go              # Host 结构体
│   ├── credential.go        # Credential 结构体
│   └── dangerrule.go        # DangerRule、DangerEvent 结构体
├── credential/
│   ├── repo.go              # Credential CRUD
│   ├── service.go           # 加密存储/解密逻辑
│   ├── handler.go           # GET/POST /api/credentials
│   └── service_test.go      # 加解密集成测试
├── host/
│   ├── repo.go              # Host CRUD
│   ├── service.go           # 业务逻辑（创建、列表、删除）
│   ├── handler.go           # GET/POST/PUT/DELETE /api/hosts
│   └── service_test.go      # 单元测试
└── dangerrule/
    ├── repo.go              # DangerRule CRUD
    ├── service.go           # 业务逻辑 + regexp 编译缓存
    ├── handler.go           # GET/POST/PUT/DELETE /api/danger-rules
    ├── matcher.go           # 对外暴露：Match(tenantID, cmd) 接口
    └── service_test.go      # 正则匹配单元测试
```

---

## Chunk 1: 模型 + Credential 模块

### Task 1: 添加 Host / Credential / DangerRule 模型

**Files:**
- Create: `internal/model/host.go`
- Create: `internal/model/credential.go`
- Create: `internal/model/dangerrule.go`
- Modify: `internal/db/migrate.go`

- [ ] **Step 1: 创建 Host 模型**

```go
// internal/model/host.go
package model

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// Tags 是 JSON 字符串切片，存储主机标签
type Tags []string

func (t Tags) Value() (driver.Value, error) {
    b, err := json.Marshal(t)
    return string(b), err
}

func (t *Tags) Scan(value any) error {
    if value == nil {
        *t = Tags{}
        return nil
    }
    var bytes []byte
    switch v := value.(type) {
    case []byte:
        bytes = v
    case string:
        bytes = []byte(v)
    default:
        return fmt.Errorf("cannot scan type %T", value)
    }
    return json.Unmarshal(bytes, t)
}

type Host struct {
    ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID     int64     `gorm:"not null;index" json:"tenant_id"`
    Name         string    `gorm:"size:100;not null" json:"name"`
    IP           string    `gorm:"size:45;not null" json:"ip"`
    Port         int       `gorm:"default:22;not null" json:"port"`
    AuthType     string    `gorm:"size:20;not null" json:"auth_type"` // password / key
    CredentialID int64     `gorm:"not null" json:"credential_id"`
    Status       int8      `gorm:"default:1;not null" json:"status"`
    Tags         Tags      `gorm:"type:json" json:"tags"`
    CreatedAt    time.Time `json:"created_at"`
}
```

- [ ] **Step 2: 创建 Credential 模型**

```go
// internal/model/credential.go
package model

import "time"

type Credential struct {
    ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID        int64     `gorm:"not null;index" json:"tenant_id"`
    Name            string    `gorm:"size:100;not null" json:"name"`
    Type            string    `gorm:"size:20;not null" json:"type"` // password / key
    Username        string    `gorm:"size:50;not null" json:"username"`
    EncryptedSecret string    `gorm:"type:text;not null" json:"-"` // 禁止序列化
    CreatedAt       time.Time `json:"created_at"`
}
```

- [ ] **Step 3: 创建 DangerRule / DangerEvent 模型**

```go
// internal/model/dangerrule.go
package model

import "time"

type DangerRule struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID    int64     `gorm:"not null;index" json:"tenant_id"` // 0=全局规则
    Pattern     string    `gorm:"size:500;not null" json:"pattern"`
    Description string    `gorm:"size:255" json:"description"`
    Action      string    `gorm:"size:20;default:block;not null" json:"action"`
    Enabled     int8      `gorm:"default:1;not null" json:"enabled"`
    CreatedAt   time.Time `json:"created_at"`
}

type DangerEvent struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    SessionID   int64     `gorm:"not null;index" json:"session_id"`
    MatchedRule string    `gorm:"size:255;not null" json:"matched_rule"`
    Command     string    `gorm:"type:text;not null" json:"command"`
    TriggeredAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"triggered_at"`
}
```

- [ ] **Step 4: 更新 migrate.go，加入新模型**

```go
// internal/db/migrate.go
package db

import (
    "github.com/sshive/sshive/internal/model"
)

func Migrate() error {
    return DB.AutoMigrate(
        &model.Tenant{},
        &model.User{},
        &model.Role{},
        &model.UserRole{},
        &model.Permission{},
        &model.RolePermission{},
        // Plan 2 新增
        &model.Host{},
        &model.Credential{},
        &model.DangerRule{},
        &model.DangerEvent{},
    )
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./internal/model/ ./internal/db/
```

Expected: 无报错

- [ ] **Step 6: Commit**

```bash
git add internal/model/ internal/db/migrate.go
git commit -m "feat: add host, credential, danger rule models"
```

---

### Task 2: Credential 模块（含加密）

**Files:**
- Create: `internal/credential/repo.go`
- Create: `internal/credential/service.go`
- Create: `internal/credential/handler.go`
- Create: `internal/credential/service_test.go`

- [ ] **Step 1: 实现 Credential Repo**

```go
// internal/credential/repo.go
package credential

import (
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(c *model.Credential) error {
    return db.DB.Create(c).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.Credential, int64, error) {
    var list []model.Credential
    var total int64
    offset := (page - 1) * pageSize
    q := db.DB.Model(&model.Credential{}).Where("tenant_id = ?", tenantID)
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := q.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}

func (r *Repo) GetByID(tenantID, id int64) (*model.Credential, error) {
    var c model.Credential
    if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&c).Error; err != nil {
        return nil, err
    }
    return &c, nil
}

func (r *Repo) Delete(tenantID, id int64) error {
    return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Credential{}).Error
}
```

- [ ] **Step 2: 实现 Credential Service**

```go
// internal/credential/service.go
package credential

import (
    "encoding/base64"
    "fmt"

    "github.com/sshive/sshive/internal/config"
    "github.com/sshive/sshive/internal/model"
    "github.com/sshive/sshive/pkg/encrypt"
)

type Service struct {
    repo *Repo
}

func NewService() *Service {
    return &Service{repo: &Repo{}}
}

type CreateReq struct {
    Name     string `json:"name" binding:"required,max=100"`
    Type     string `json:"type" binding:"required,oneof=password key"`
    Username string `json:"username" binding:"required,max=50"`
    Secret   string `json:"secret" binding:"required"` // 明文密码或私钥，仅创建时传入
}

func (s *Service) encryptionKey() ([]byte, error) {
    return base64.StdEncoding.DecodeString(config.C.EncryptKey)
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.Credential, error) {
    key, err := s.encryptionKey()
    if err != nil {
        return nil, fmt.Errorf("load encrypt key: %w", err)
    }
    encrypted, err := encrypt.Encrypt(key, req.Secret)
    if err != nil {
        return nil, fmt.Errorf("encrypt secret: %w", err)
    }
    c := &model.Credential{
        TenantID:        tenantID,
        Name:            req.Name,
        Type:            req.Type,
        Username:        req.Username,
        EncryptedSecret: encrypted,
    }
    if err := s.repo.Create(c); err != nil {
        return nil, fmt.Errorf("save credential: %w", err)
    }
    return c, nil
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.Credential, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.List(tenantID, page, pageSize)
}

// DecryptSecret 解密指定凭据的 secret，供 SSH/SFTP 模块调用
func (s *Service) DecryptSecret(tenantID, id int64) (username, secret string, err error) {
    c, err := s.repo.GetByID(tenantID, id)
    if err != nil {
        return "", "", fmt.Errorf("credential not found: %w", err)
    }
    key, err := s.encryptionKey()
    if err != nil {
        return "", "", fmt.Errorf("load encrypt key: %w", err)
    }
    plain, err := encrypt.Decrypt(key, c.EncryptedSecret)
    if err != nil {
        return "", "", fmt.Errorf("decrypt: %w", err)
    }
    return c.Username, plain, nil
}
```

- [ ] **Step 3: 写 Service 测试（不依赖数据库）**

```go
// internal/credential/service_test.go
package credential_test

import (
    "encoding/base64"
    "os"
    "testing"

    "github.com/sshive/sshive/internal/config"
    "github.com/sshive/sshive/pkg/encrypt"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
    // 模拟 32 字节密钥
    raw := make([]byte, 32)
    for i := range raw { raw[i] = byte(i + 1) }
    key := base64.StdEncoding.EncodeToString(raw)
    os.Setenv("SSHIVE_ENCRYPT_KEY", key)
    config.C.EncryptKey = key

    secret := "my-private-key-content"
    encKey, err := base64.StdEncoding.DecodeString(config.C.EncryptKey)
    require.NoError(t, err)

    encrypted, err := encrypt.Encrypt(encKey, secret)
    require.NoError(t, err)
    assert.NotEmpty(t, encrypted)

    decrypted, err := encrypt.Decrypt(encKey, encrypted)
    require.NoError(t, err)
    assert.Equal(t, secret, decrypted)
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/credential/ -v -run TestEncryptDecryptRoundtrip
```

Expected: PASS

- [ ] **Step 5: 实现 Credential Handler**

```go
// internal/credential/handler.go
package credential

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/internal/auth"
    "github.com/sshive/sshive/pkg/middleware"
)

type Handler struct {
    svc *Service
}

func NewHandler() *Handler {
    return &Handler{svc: NewService()}
}

func (h *Handler) List(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    list, total, err := h.svc.List(tenantID, page, pageSize)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) Create(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    cred, err := h.svc.Create(tenantID, req)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, cred)
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/credential/
git commit -m "feat: add credential module with AES encryption"
```

---

## Chunk 2: Host 模块

### Task 3: Host CRUD

**Files:**
- Create: `internal/host/repo.go`
- Create: `internal/host/service.go`
- Create: `internal/host/handler.go`
- Create: `internal/host/service_test.go`

- [ ] **Step 1: 实现 Host Repo**

```go
// internal/host/repo.go
package host

import (
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(h *model.Host) error {
    return db.DB.Create(h).Error
}

func (r *Repo) Update(h *model.Host) error {
    return db.DB.Save(h).Error
}

func (r *Repo) Delete(tenantID, id int64) error {
    return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Host{}).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.Host, int64, error) {
    var list []model.Host
    var total int64
    offset := (page - 1) * pageSize
    q := db.DB.Model(&model.Host{}).Where("tenant_id = ?", tenantID)
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := q.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}

func (r *Repo) GetByID(tenantID, id int64) (*model.Host, error) {
    var h model.Host
    if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&h).Error; err != nil {
        return nil, err
    }
    return &h, nil
}
```

- [ ] **Step 2: 实现 Host Service**

```go
// internal/host/service.go
package host

import (
    "fmt"
    "github.com/sshive/sshive/internal/model"
)

type Service struct {
    repo *Repo
}

func NewService() *Service {
    return &Service{repo: &Repo{}}
}

type CreateReq struct {
    Name         string      `json:"name" binding:"required,max=100"`
    IP           string      `json:"ip" binding:"required"`
    Port         int         `json:"port"`
    AuthType     string      `json:"auth_type" binding:"required,oneof=password key"`
    CredentialID int64       `json:"credential_id" binding:"required"`
    Tags         model.Tags  `json:"tags"`
}

type UpdateReq struct {
    Name         string      `json:"name" binding:"required,max=100"`
    IP           string      `json:"ip" binding:"required"`
    Port         int         `json:"port"`
    AuthType     string      `json:"auth_type" binding:"required,oneof=password key"`
    CredentialID int64       `json:"credential_id" binding:"required"`
    Tags         model.Tags  `json:"tags"`
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.Host, error) {
    port := req.Port
    if port == 0 { port = 22 }
    h := &model.Host{
        TenantID:     tenantID,
        Name:         req.Name,
        IP:           req.IP,
        Port:         port,
        AuthType:     req.AuthType,
        CredentialID: req.CredentialID,
        Status:       1,
        Tags:         req.Tags,
    }
    if err := s.repo.Create(h); err != nil {
        return nil, fmt.Errorf("create host: %w", err)
    }
    return h, nil
}

func (s *Service) Update(tenantID, id int64, req UpdateReq) (*model.Host, error) {
    h, err := s.repo.GetByID(tenantID, id)
    if err != nil {
        return nil, fmt.Errorf("host not found")
    }
    h.Name = req.Name
    h.IP = req.IP
    h.Port = req.Port
    if h.Port == 0 { h.Port = 22 }
    h.AuthType = req.AuthType
    h.CredentialID = req.CredentialID
    h.Tags = req.Tags
    if err := s.repo.Update(h); err != nil {
        return nil, fmt.Errorf("update host: %w", err)
    }
    return h, nil
}

func (s *Service) Delete(tenantID, id int64) error {
    return s.repo.Delete(tenantID, id)
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.Host, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.List(tenantID, page, pageSize)
}

func (s *Service) GetByID(tenantID, id int64) (*model.Host, error) {
    return s.repo.GetByID(tenantID, id)
}
```

- [ ] **Step 3: 实现 Host Handler**

```go
// internal/host/handler.go
package host

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/internal/auth"
    "github.com/sshive/sshive/pkg/middleware"
)

type Handler struct {
    svc *Service
}

func NewHandler() *Handler {
    return &Handler{svc: NewService()}
}

func (h *Handler) List(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    list, total, err := h.svc.List(tenantID, page, pageSize)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) Create(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    host, err := h.svc.Create(tenantID, req)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, host)
}

func (h *Handler) Update(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid id")
        return
    }
    var req UpdateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    host, err := h.svc.Update(tenantID, id, req)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, host)
}

func (h *Handler) Delete(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid id")
        return
    }
    if err := h.svc.Delete(tenantID, id); err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, nil)
}
```

- [ ] **Step 4: 编译验证**

```bash
go build ./internal/host/
```

Expected: 无报错

- [ ] **Step 5: Commit**

```bash
git add internal/host/
git commit -m "feat: add host CRUD module"
```

---

## Chunk 3: 高危规则模块

### Task 4: DangerRule CRUD + 正则匹配器

**Files:**
- Create: `internal/dangerrule/repo.go`
- Create: `internal/dangerrule/service.go`
- Create: `internal/dangerrule/matcher.go`
- Create: `internal/dangerrule/handler.go`
- Create: `internal/dangerrule/service_test.go`

- [ ] **Step 1: 实现 DangerRule Repo**

```go
// internal/dangerrule/repo.go
package dangerrule

import (
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(rule *model.DangerRule) error {
    return db.DB.Create(rule).Error
}

func (r *Repo) Update(rule *model.DangerRule) error {
    return db.DB.Save(rule).Error
}

func (r *Repo) Delete(tenantID, id int64) error {
    return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DangerRule{}).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.DangerRule, int64, error) {
    var list []model.DangerRule
    var total int64
    offset := (page - 1) * pageSize
    q := db.DB.Model(&model.DangerRule{}).Where("tenant_id = ?", tenantID)
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := q.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}

// ListActive 返回租户规则 + 全局规则（tenant_id=0），供匹配器使用
func (r *Repo) ListActive(tenantID int64) ([]model.DangerRule, error) {
    var list []model.DangerRule
    err := db.DB.Where("(tenant_id = ? OR tenant_id = 0) AND enabled = 1", tenantID).Find(&list).Error
    return list, err
}
```

- [ ] **Step 2: 实现正则匹配器（含缓存）**

```go
// internal/dangerrule/matcher.go
package dangerrule

import (
    "regexp"
    "sync"
)

// Matcher 维护编译好的正则表达式缓存
type Matcher struct {
    mu      sync.RWMutex
    repo    *Repo
    cache   map[int64][]*compiledRule // tenantID -> 编译好的规则列表
}

type compiledRule struct {
    id      int64
    pattern string
    re      *regexp.Regexp
}

var GlobalMatcher = &Matcher{
    repo:  &Repo{},
    cache: make(map[int64][]*compiledRule),
}

// Reload 重新从数据库加载并编译指定租户的规则
func (m *Matcher) Reload(tenantID int64) error {
    rules, err := m.repo.ListActive(tenantID)
    if err != nil {
        return err
    }
    compiled := make([]*compiledRule, 0, len(rules))
    for _, r := range rules {
        re, err := regexp.Compile(r.Pattern)
        if err != nil {
            continue // 跳过无效正则
        }
        compiled = append(compiled, &compiledRule{
            id:      r.ID,
            pattern: r.Pattern,
            re:      re,
        })
    }
    m.mu.Lock()
    m.cache[tenantID] = compiled
    m.mu.Unlock()
    return nil
}

// Match 检查命令是否命中规则，返回命中的规则 pattern（未命中返回空字符串）
func (m *Matcher) Match(tenantID int64, command string) string {
    m.mu.RLock()
    rules, ok := m.cache[tenantID]
    m.mu.RUnlock()
    if !ok {
        // 首次使用时懒加载
        _ = m.Reload(tenantID)
        m.mu.RLock()
        rules = m.cache[tenantID]
        m.mu.RUnlock()
    }
    for _, r := range rules {
        if r.re.MatchString(command) {
            return r.pattern
        }
    }
    return ""
}
```

- [ ] **Step 3: 写正则匹配单元测试（不依赖数据库）**

```go
// internal/dangerrule/service_test.go
package dangerrule_test

import (
    "regexp"
    "testing"

    "github.com/stretchr/testify/assert"
)

// 直接测试正则逻辑，不依赖数据库
func TestDangerPatterns(t *testing.T) {
    cases := []struct {
        pattern string
        cmd     string
        match   bool
    }{
        {`rm\s+-rf\s+/`, "rm -rf /", true},
        {`rm\s+-rf\s+/`, "rm -rf /tmp/log", true},
        {`rm\s+-rf\s+/`, "rm -rf ./mydir", false},
        {`shutdown|reboot|halt`, "shutdown now", true},
        {`shutdown|reboot|halt`, "reboot", true},
        {`shutdown|reboot|halt`, "echo hello", false},
        {`mkfs\..*`, "mkfs.ext4 /dev/sdb", true},
        {`mkfs\..*`, "ls /dev/sdb", false},
        {`dd\s+if=.*of=/dev/`, "dd if=/dev/zero of=/dev/sda", true},
        {`dd\s+if=.*of=/dev/`, "dd if=image.iso of=/tmp/out", false},
    }
    for _, tc := range cases {
        re := regexp.MustCompile(tc.pattern)
        got := re.MatchString(tc.cmd)
        assert.Equal(t, tc.match, got, "pattern=%q cmd=%q", tc.pattern, tc.cmd)
    }
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
go test ./internal/dangerrule/ -v -run TestDangerPatterns
```

Expected: PASS（10 个子用例全部通过）

- [ ] **Step 5: 实现 DangerRule Service**

```go
// internal/dangerrule/service.go
package dangerrule

import (
    "fmt"
    "regexp"

    "github.com/sshive/sshive/internal/model"
)

type Service struct {
    repo *Repo
}

func NewService() *Service {
    return &Service{repo: &Repo{}}
}

type CreateReq struct {
    Pattern     string `json:"pattern" binding:"required,max=500"`
    Description string `json:"description" binding:"max=255"`
    Action      string `json:"action" binding:"omitempty,oneof=block warn"`
}

type UpdateReq struct {
    Pattern     string `json:"pattern" binding:"required,max=500"`
    Description string `json:"description" binding:"max=255"`
    Action      string `json:"action" binding:"omitempty,oneof=block warn"`
    Enabled     *int8  `json:"enabled"`
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.DangerRule, error) {
    if _, err := regexp.Compile(req.Pattern); err != nil {
        return nil, fmt.Errorf("invalid regex: %w", err)
    }
    action := req.Action
    if action == "" { action = "block" }
    rule := &model.DangerRule{
        TenantID:    tenantID,
        Pattern:     req.Pattern,
        Description: req.Description,
        Action:      action,
        Enabled:     1,
    }
    if err := s.repo.Create(rule); err != nil {
        return nil, fmt.Errorf("create rule: %w", err)
    }
    // 刷新匹配器缓存
    _ = GlobalMatcher.Reload(tenantID)
    return rule, nil
}

func (s *Service) Update(tenantID, id int64, req UpdateReq) (*model.DangerRule, error) {
    rules, _, err := s.repo.List(tenantID, 1, 1000)
    if err != nil {
        return nil, err
    }
    var rule *model.DangerRule
    for i := range rules {
        if rules[i].ID == id {
            rule = &rules[i]
            break
        }
    }
    if rule == nil {
        return nil, fmt.Errorf("rule not found")
    }
    if _, err := regexp.Compile(req.Pattern); err != nil {
        return nil, fmt.Errorf("invalid regex: %w", err)
    }
    rule.Pattern = req.Pattern
    rule.Description = req.Description
    if req.Action != "" { rule.Action = req.Action }
    if req.Enabled != nil { rule.Enabled = *req.Enabled }
    if err := s.repo.Update(rule); err != nil {
        return nil, fmt.Errorf("update rule: %w", err)
    }
    _ = GlobalMatcher.Reload(tenantID)
    return rule, nil
}

func (s *Service) Delete(tenantID, id int64) error {
    if err := s.repo.Delete(tenantID, id); err != nil {
        return err
    }
    _ = GlobalMatcher.Reload(tenantID)
    return nil
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.DangerRule, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.List(tenantID, page, pageSize)
}
```

- [ ] **Step 6: 实现 DangerRule Handler**

```go
// internal/dangerrule/handler.go
package dangerrule

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/internal/auth"
    "github.com/sshive/sshive/pkg/middleware"
)

type Handler struct {
    svc *Service
}

func NewHandler() *Handler {
    return &Handler{svc: NewService()}
}

func (h *Handler) List(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    list, total, err := h.svc.List(tenantID, page, pageSize)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) Create(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    rule, err := h.svc.Create(tenantID, req)
    if err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    middleware.OK(c, rule)
}

func (h *Handler) Update(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid id")
        return
    }
    var req UpdateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    rule, err := h.svc.Update(tenantID, id, req)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, rule)
}

func (h *Handler) Delete(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid id")
        return
    }
    if err := h.svc.Delete(tenantID, id); err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, nil)
}
```

- [ ] **Step 7: 编译验证**

```bash
go build ./internal/dangerrule/
```

Expected: 无报错

- [ ] **Step 8: Commit**

```bash
git add internal/dangerrule/
git commit -m "feat: add danger rule CRUD with regex matcher cache"
```

---

### Task 5: 注册 Plan 2 路由到 main.go

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 在 main.go 中注册新路由**

在 `authed` 路由组后追加：

```go
// 凭据管理
credH := credential.NewHandler()
authed.GET("/credentials", auth.RequirePermission("host:manage"), credH.List)
authed.POST("/credentials", auth.RequirePermission("host:manage"), credH.Create)

// 主机管理
hostH := host.NewHandler()
authed.GET("/hosts", auth.RequirePermission("host:manage"), hostH.List)
authed.POST("/hosts", auth.RequirePermission("host:manage"), hostH.Create)
authed.PUT("/hosts/:id", auth.RequirePermission("host:manage"), hostH.Update)
authed.DELETE("/hosts/:id", auth.RequirePermission("host:manage"), hostH.Delete)

// 高危规则管理
ruleH := dangerrule.NewHandler()
authed.GET("/danger-rules", auth.RequirePermission("rule:manage"), ruleH.List)
authed.POST("/danger-rules", auth.RequirePermission("rule:manage"), ruleH.Create)
authed.PUT("/danger-rules/:id", auth.RequirePermission("rule:manage"), ruleH.Update)
authed.DELETE("/danger-rules/:id", auth.RequirePermission("rule:manage"), ruleH.Delete)
```

并在 import 中加入：

```go
"github.com/sshive/sshive/internal/credential"
"github.com/sshive/sshive/internal/host"
"github.com/sshive/sshive/internal/dangerrule"
```

- [ ] **Step 2: 编译验证**

```bash
go build ./cmd/server/
```

Expected: 无报错

- [ ] **Step 3: 运行全部测试**

```bash
go test ./... -v
```

Expected: 所有测试 PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: register host/credential/danger-rule routes in main"
```

- [ ] **Step 5: Push 到 GitHub**

```bash
git push origin main
```

---

## 测试汇总

```bash
go test ./internal/credential/... -v   # 1 test: PASS
go test ./internal/dangerrule/... -v   # 10 sub-cases: PASS
go build ./...                         # SUCCESS
```
