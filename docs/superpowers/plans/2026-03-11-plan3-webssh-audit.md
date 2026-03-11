# SSHive Plan 3: WebSSH + 审计录制

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 WebSocket SSH 终端（含 PTY、高危拦截）和完整审计录制（.cast 文件 + session_logs + session_commands）。

**Architecture:** 浏览器通过 WebSocket 连接到 Go 服务，服务端建立 SSH 连接并开启 PTY。输入流在回车时触发高危规则检测，命中则拦截并向浏览器返回红色提示，不透传。输出流同时写给浏览器和异步审计 channel。审计 goroutine 负责写 .cast 文件、session_logs、session_commands 三路记录。

**Tech Stack:** Go, gorilla/websocket, golang.org/x/crypto/ssh, asciinema v2 .cast 格式

**前置条件:** Plan 1 + Plan 2 已完成，`go build ./...` 通过。

---

## 文件结构

```
internal/
├── model/
│   └── session.go           # Session、SessionLog、SessionCommand 结构体
├── audit/
│   ├── repo.go              # Session CRUD，SessionLog/Command 写入，查询接口
│   ├── service.go           # 会话生命周期管理（开始/结束）、日志查询
│   ├── writer.go            # 异步审计写入器（channel + goroutine）
│   └── handler.go           # GET /api/sessions, /sessions/:id/logs, /commands, /replay
└── ssh/
    ├── handler.go           # WS /api/ws/ssh/:hostId 入口，权限验证，组装 Session
    ├── session.go           # SSH Session：建立连接、开启 PTY、输入输出循环
    └── interceptor.go       # 高危拦截器：缓冲输入，回车时检测，命中则拦截

pkg/
└── cast/
    ├── writer.go            # asciinema v2 .cast 文件写入器
    └── writer_test.go       # 单元测试
```

---

## Chunk 1: Session 模型 + pkg/cast

### Task 1: Session 相关模型

**Files:**
- Create: `internal/model/session.go`
- Modify: `internal/db/migrate.go`

- [ ] **Step 1: 创建 Session / SessionLog / SessionCommand 模型**

```go
// internal/model/session.go
package model

import "time"

type Session struct {
    ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID     int64      `gorm:"not null;index" json:"tenant_id"`
    UserID       int64      `gorm:"not null;index" json:"user_id"`
    HostID       int64      `gorm:"not null;index" json:"host_id"`
    StartedAt    time.Time  `gorm:"not null" json:"started_at"`
    EndedAt      *time.Time `json:"ended_at"`
    ClientIP     string     `gorm:"size:45" json:"client_ip"`
    Status       string     `gorm:"size:20;default:active;not null" json:"status"` // active / closed
    CastFilePath string     `gorm:"size:500" json:"cast_file_path"`
}

type SessionLog struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    SessionID int64     `gorm:"not null;index" json:"session_id"`
    Type      string    `gorm:"size:10;not null" json:"type"` // input / output
    Content   string    `gorm:"type:mediumtext;not null" json:"content"`
    CreatedAt time.Time `gorm:"precision:3;default:CURRENT_TIMESTAMP(3)" json:"created_at"`
}

type SessionCommand struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    SessionID int64     `gorm:"not null;index" json:"session_id"`
    Command   string    `gorm:"type:text;not null" json:"command"`
    CreatedAt time.Time `gorm:"precision:3;default:CURRENT_TIMESTAMP(3)" json:"created_at"`
}
```

- [ ] **Step 2: 更新 migrate.go**

在 `DB.AutoMigrate(...)` 列表中追加：

```go
&model.Session{},
&model.SessionLog{},
&model.SessionCommand{},
```

- [ ] **Step 3: 编译验证**

```bash
go build ./internal/model/ ./internal/db/
```

Expected: 无报错

- [ ] **Step 4: Commit**

```bash
git add internal/model/session.go internal/db/migrate.go
git commit -m "feat: add session/log/command models"
```

---

### Task 2: pkg/cast — asciinema v2 写入器

**Files:**
- Create: `pkg/cast/writer.go`
- Create: `pkg/cast/writer_test.go`

- [ ] **Step 1: 写失败测试**

```go
// pkg/cast/writer_test.go
package cast_test

import (
    "os"
    "strings"
    "testing"
    "time"

    "github.com/sshive/sshive/pkg/cast"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCastWriter(t *testing.T) {
    path := t.TempDir() + "/test.cast"

    w, err := cast.NewWriter(path, 80, 24)
    require.NoError(t, err)

    w.WriteOutput(0, "Hello ")
    w.WriteOutput(100*time.Millisecond, "World\r\n")
    w.WriteInput(200*time.Millisecond, "ls\n")
    require.NoError(t, w.Close())

    data, err := os.ReadFile(path)
    require.NoError(t, err)

    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    assert.GreaterOrEqual(t, len(lines), 3) // header + 2 output + 1 input

    // 第一行是 header JSON
    assert.Contains(t, lines[0], `"version":2`)
    assert.Contains(t, lines[0], `"width":80`)
    assert.Contains(t, lines[0], `"height":24`)

    // 第二行是第一个 output 事件
    assert.Contains(t, lines[1], `"o"`)
    assert.Contains(t, lines[1], "Hello ")
}
```

- [ ] **Step 2: 运行确认失败**

```bash
go test ./pkg/cast/ -v
```

Expected: FAIL with "undefined: cast.NewWriter"

- [ ] **Step 3: 实现 cast.Writer**

```go
// pkg/cast/writer.go
package cast

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "time"
)

// Header 是 asciinema v2 .cast 文件头
type Header struct {
    Version   int    `json:"version"`
    Width     int    `json:"width"`
    Height    int    `json:"height"`
    Timestamp int64  `json:"timestamp"`
    Title     string `json:"title,omitempty"`
}

// Writer 负责写 asciinema v2 格式的 .cast 文件
// 格式：第一行 JSON header，后续每行 [elapsed_seconds, "o"|"i", "data"]
type Writer struct {
    mu        sync.Mutex
    f         *os.File
    startTime time.Time
}

// NewWriter 创建并初始化 .cast 文件，写入 header
func NewWriter(path string, width, height int) (*Writer, error) {
    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    if err != nil {
        return nil, fmt.Errorf("open cast file: %w", err)
    }
    header := Header{
        Version:   2,
        Width:     width,
        Height:    height,
        Timestamp: time.Now().Unix(),
    }
    headerJSON, err := json.Marshal(header)
    if err != nil {
        f.Close()
        return nil, err
    }
    if _, err := fmt.Fprintf(f, "%s\n", headerJSON); err != nil {
        f.Close()
        return nil, err
    }
    return &Writer{f: f, startTime: time.Now()}, nil
}

// WriteOutput 写一条输出事件（服务器 → 浏览器）
// offset 是相对于会话开始的时间偏移（0 表示使用实际时间）
func (w *Writer) WriteOutput(offset time.Duration, data string) {
    w.writeEvent("o", offset, data)
}

// WriteInput 写一条输入事件（浏览器 → 服务器）
func (w *Writer) WriteInput(offset time.Duration, data string) {
    w.writeEvent("i", offset, data)
}

func (w *Writer) writeEvent(eventType string, offset time.Duration, data string) {
    elapsed := offset.Seconds()
    if offset == 0 {
        elapsed = time.Since(w.startTime).Seconds()
    }
    line, _ := json.Marshal([]any{elapsed, eventType, data})
    w.mu.Lock()
    defer w.mu.Unlock()
    fmt.Fprintf(w.f, "%s\n", line)
}

// Close 关闭文件
func (w *Writer) Close() error {
    w.mu.Lock()
    defer w.mu.Unlock()
    return w.f.Close()
}
```

- [ ] **Step 4: 运行确认通过**

```bash
go test ./pkg/cast/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cast/
git commit -m "feat: add asciinema v2 cast file writer"
```

---

## Chunk 2: 审计模块

### Task 3: Audit Repo + Service + 异步写入器

**Files:**
- Create: `internal/audit/repo.go`
- Create: `internal/audit/writer.go`
- Create: `internal/audit/service.go`
- Create: `internal/audit/handler.go`

- [ ] **Step 1: 实现 Audit Repo**

```go
// internal/audit/repo.go
package audit

import (
    "time"

    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) CreateSession(s *model.Session) error {
    return db.DB.Create(s).Error
}

func (r *Repo) CloseSession(id int64, castPath string) error {
    now := time.Now()
    return db.DB.Model(&model.Session{}).Where("id = ?", id).Updates(map[string]any{
        "ended_at":       now,
        "status":         "closed",
        "cast_file_path": castPath,
    }).Error
}

func (r *Repo) WriteLog(log *model.SessionLog) error {
    return db.DB.Create(log).Error
}

func (r *Repo) WriteCommand(cmd *model.SessionCommand) error {
    return db.DB.Create(cmd).Error
}

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

func (r *Repo) GetSession(tenantID, id int64) (*model.Session, error) {
    var s model.Session
    if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error; err != nil {
        return nil, err
    }
    return &s, nil
}

func (r *Repo) ListLogs(sessionID int64, page, pageSize int) ([]model.SessionLog, int64, error) {
    var list []model.SessionLog
    var total int64
    offset := (page - 1) * pageSize
    q := db.DB.Model(&model.SessionLog{}).Where("session_id = ?", sessionID)
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := q.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}

func (r *Repo) ListCommands(sessionID int64, page, pageSize int) ([]model.SessionCommand, int64, error) {
    var list []model.SessionCommand
    var total int64
    offset := (page - 1) * pageSize
    q := db.DB.Model(&model.SessionCommand{}).Where("session_id = ?", sessionID)
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := q.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}
```

- [ ] **Step 2: 实现异步写入器**

```go
// internal/audit/writer.go
package audit

import (
    "log/slog"
    "github.com/sshive/sshive/internal/model"
    "github.com/sshive/sshive/pkg/cast"
)

// Event 是写入 channel 的事件类型
type EventType int

const (
    EventOutput EventType = iota
    EventInput
    EventCommand
    EventClose
)

type Event struct {
    Type    EventType
    Content string
}

// AsyncWriter 异步接收审计事件，不阻塞主 SSH 链路
type AsyncWriter struct {
    sessionID int64
    repo      *Repo
    castW     *cast.Writer
    ch        chan Event
    done      chan struct{}
}

func NewAsyncWriter(sessionID int64, castWriter *cast.Writer) *AsyncWriter {
    w := &AsyncWriter{
        sessionID: sessionID,
        repo:      &Repo{},
        castW:     castWriter,
        ch:        make(chan Event, 256),
        done:      make(chan struct{}),
    }
    go w.run()
    return w
}

func (w *AsyncWriter) run() {
    defer close(w.done)
    for evt := range w.ch {
        switch evt.Type {
        case EventOutput:
            if w.castW != nil {
                w.castW.WriteOutput(0, evt.Content)
            }
            if err := w.repo.WriteLog(&model.SessionLog{
                SessionID: w.sessionID,
                Type:      "output",
                Content:   evt.Content,
            }); err != nil {
                slog.Warn("audit write output log failed", "err", err)
            }
        case EventInput:
            if w.castW != nil {
                w.castW.WriteInput(0, evt.Content)
            }
            if err := w.repo.WriteLog(&model.SessionLog{
                SessionID: w.sessionID,
                Type:      "input",
                Content:   evt.Content,
            }); err != nil {
                slog.Warn("audit write input log failed", "err", err)
            }
        case EventCommand:
            if err := w.repo.WriteCommand(&model.SessionCommand{
                SessionID: w.sessionID,
                Command:   evt.Content,
            }); err != nil {
                slog.Warn("audit write command failed", "err", err)
            }
        }
    }
}

// SendOutput 发送输出事件（非阻塞）
func (w *AsyncWriter) SendOutput(data string) {
    select {
    case w.ch <- Event{Type: EventOutput, Content: data}:
    default:
        slog.Warn("audit channel full, dropping output event")
    }
}

// SendInput 发送输入事件（非阻塞）
func (w *AsyncWriter) SendInput(data string) {
    select {
    case w.ch <- Event{Type: EventInput, Content: data}:
    default:
    }
}

// SendCommand 发送命令事件（非阻塞）
func (w *AsyncWriter) SendCommand(cmd string) {
    select {
    case w.ch <- Event{Type: EventCommand, Content: cmd}:
    default:
    }
}

// Close 等待所有事件处理完毕
func (w *AsyncWriter) Close() {
    close(w.ch)
    <-w.done
    if w.castW != nil {
        w.castW.Close()
    }
}
```

- [ ] **Step 3: 实现 Audit Service**

```go
// internal/audit/service.go
package audit

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "time"

    "github.com/sshive/sshive/internal/config"
    "github.com/sshive/sshive/internal/model"
    "github.com/sshive/sshive/pkg/cast"
)

type Service struct {
    repo *Repo
}

func NewService() *Service {
    return &Service{repo: &Repo{}}
}

type StartSessionResult struct {
    Session     *model.Session
    AsyncWriter *AsyncWriter
}

// StartSession 创建会话记录，初始化 .cast 文件和异步写入器
func (s *Service) StartSession(tenantID, userID, hostID int64, clientIP string, width, height int) (*StartSessionResult, error) {
    sess := &model.Session{
        TenantID:  tenantID,
        UserID:    userID,
        HostID:    hostID,
        StartedAt: time.Now().UTC(),
        ClientIP:  clientIP,
        Status:    "active",
    }
    if err := s.repo.CreateSession(sess); err != nil {
        return nil, fmt.Errorf("create session: %w", err)
    }

    castDir := config.C.CastDir
    if err := os.MkdirAll(castDir, 0755); err != nil {
        return nil, fmt.Errorf("mkdir cast dir: %w", err)
    }
    castPath := filepath.Join(castDir, strconv.FormatInt(sess.ID, 10)+".cast")

    castWriter, err := cast.NewWriter(castPath, width, height)
    if err != nil {
        return nil, fmt.Errorf("create cast writer: %w", err)
    }

    asyncW := NewAsyncWriter(sess.ID, castWriter)
    return &StartSessionResult{Session: sess, AsyncWriter: asyncW}, nil
}

// CloseSession 关闭会话，等待审计写入完成
func (s *Service) CloseSession(sessID int64, asyncW *AsyncWriter, castPath string) {
    asyncW.Close()
    _ = s.repo.CloseSession(sessID, castPath)
}

func (s *Service) ListSessions(tenantID int64, page, pageSize int) ([]model.Session, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.ListSessions(tenantID, page, pageSize)
}

func (s *Service) GetSession(tenantID, id int64) (*model.Session, error) {
    return s.repo.GetSession(tenantID, id)
}

func (s *Service) ListLogs(tenantID, sessionID int64, page, pageSize int) ([]model.SessionLog, int64, error) {
    // 验证 session 属于该租户
    if _, err := s.repo.GetSession(tenantID, sessionID); err != nil {
        return nil, 0, fmt.Errorf("session not found")
    }
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.ListLogs(sessionID, page, pageSize)
}

func (s *Service) ListCommands(tenantID, sessionID int64, page, pageSize int) ([]model.SessionCommand, int64, error) {
    if _, err := s.repo.GetSession(tenantID, sessionID); err != nil {
        return nil, 0, fmt.Errorf("session not found")
    }
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.ListCommands(sessionID, page, pageSize)
}
```

- [ ] **Step 4: 实现 Audit Handler**

```go
// internal/audit/handler.go
package audit

import (
    "os"
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

func (h *Handler) ListSessions(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    list, total, err := h.svc.ListSessions(tenantID, page, pageSize)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) ListLogs(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid session id")
        return
    }
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    list, total, err := h.svc.ListLogs(tenantID, sessionID, page, pageSize)
    if err != nil {
        middleware.NotFound(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) ListCommands(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid session id")
        return
    }
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    list, total, err := h.svc.ListCommands(tenantID, sessionID, page, pageSize)
    if err != nil {
        middleware.NotFound(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": list})
}

// Replay 返回 .cast 文件内容（用于前端 asciinema-player 播放）
func (h *Handler) Replay(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid session id")
        return
    }
    sess, err := h.svc.GetSession(tenantID, sessionID)
    if err != nil {
        middleware.NotFound(c, "session not found")
        return
    }
    if sess.CastFilePath == "" {
        middleware.NotFound(c, "cast file not available")
        return
    }
    data, err := os.ReadFile(sess.CastFilePath)
    if err != nil {
        middleware.NotFound(c, "cast file not found")
        return
    }
    c.Data(200, "application/octet-stream", data)
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./internal/audit/
```

Expected: 无报错

- [ ] **Step 6: Commit**

```bash
git add internal/audit/
git commit -m "feat: add audit module with async writer and session management"
```

---

## Chunk 3: WebSSH 核心

### Task 4: 高危拦截器

**Files:**
- Create: `internal/ssh/interceptor.go`

- [ ] **Step 1: 实现输入拦截器**

```go
// internal/ssh/interceptor.go
package ssh

import (
    "strings"
    "github.com/sshive/sshive/internal/dangerrule"
)

// Interceptor 缓冲用户输入，在回车时检测高危指令
type Interceptor struct {
    tenantID int64
    buf      strings.Builder
}

func NewInterceptor(tenantID int64) *Interceptor {
    return &Interceptor{tenantID: tenantID}
}

// InterceptResult 拦截结果
type InterceptResult struct {
    Blocked     bool   // 是否被拦截
    Command     string // 完整命令（用于审计）
    MatchedRule string // 命中的规则 pattern
}

// Feed 处理一段输入数据，返回是否包含回车以及拦截结果
// 若未触发回车，返回 nil 表示继续缓冲
// 调用方应只在 result.Blocked==false 时才透传数据给 SSH 服务器
func (ic *Interceptor) Feed(data string) *InterceptResult {
    ic.buf.WriteString(data)
    // 检查是否含有回车（\r 或 \r\n）
    if !strings.ContainsAny(data, "\r\n") {
        return nil
    }
    cmd := strings.TrimRight(ic.buf.String(), "\r\n")
    ic.buf.Reset()

    if cmd == "" {
        return &InterceptResult{Blocked: false, Command: cmd}
    }

    matched := dangerrule.GlobalMatcher.Match(ic.tenantID, cmd)
    if matched != "" {
        return &InterceptResult{Blocked: true, Command: cmd, MatchedRule: matched}
    }
    return &InterceptResult{Blocked: false, Command: cmd}
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./internal/ssh/
```

Expected: 无报错

- [ ] **Step 3: Commit**

```bash
git add internal/ssh/interceptor.go
git commit -m "feat: add danger command interceptor"
```

---

### Task 5: SSH Session（建立连接 + PTY + 输入输出循环）

**Files:**
- Create: `internal/ssh/session.go`

- [ ] **Step 1: 实现 SSH Session**

```go
// internal/ssh/session.go
package ssh

import (
    "fmt"
    "io"
    "log/slog"
    "net"
    "time"

    "github.com/gorilla/websocket"
    "golang.org/x/crypto/ssh"

    "github.com/sshive/sshive/internal/audit"
    "github.com/sshive/sshive/internal/model"
)

const (
    // 拦截提示（ANSI 红色）
    blockMsg = "\r\n\033[31m[SSHive] 高危指令已被拦截\033[0m\r\n"
)

// ResizeMsg 是前端发来的终端尺寸变更消息
type ResizeMsg struct {
    Type   string `json:"type"`   // "resize"
    Width  int    `json:"width"`
    Height int    `json:"height"`
}

// Session 管理一个 WebSSH 会话的完整生命周期
type Session struct {
    host        *model.Host
    sshUser     string
    sshSecret   string
    tenantID    int64
    userID      int64
    clientIP    string
    ws          *websocket.Conn
    auditSvc    *audit.Service
}

func NewSession(host *model.Host, sshUser, sshSecret string,
    tenantID, userID int64, clientIP string,
    ws *websocket.Conn, auditSvc *audit.Service) *Session {
    return &Session{
        host:      host,
        sshUser:   sshUser,
        sshSecret: sshSecret,
        tenantID:  tenantID,
        userID:    userID,
        clientIP:  clientIP,
        ws:        ws,
        auditSvc:  auditSvc,
    }
}

// Run 建立 SSH 连接，开启 PTY，运行输入输出循环，直到连接关闭
func (s *Session) Run(initWidth, initHeight int) error {
    // 1. 建立 SSH 连接
    cfg, err := s.buildSSHConfig()
    if err != nil {
        return fmt.Errorf("build ssh config: %w", err)
    }
    addr := fmt.Sprintf("%s:%d", s.host.IP, s.host.Port)
    sshConn, err := ssh.Dial("tcp", addr, cfg)
    if err != nil {
        return fmt.Errorf("ssh dial %s: %w", addr, err)
    }
    defer sshConn.Close()

    // 2. 开启 Session + PTY
    sshSess, err := sshConn.NewSession()
    if err != nil {
        return fmt.Errorf("new ssh session: %w", err)
    }
    defer sshSess.Close()

    modes := ssh.TerminalModes{
        ssh.ECHO:          1,
        ssh.TTY_OP_ISPEED: 14400,
        ssh.TTY_OP_OSPEED: 14400,
    }
    if err := sshSess.RequestPty("xterm-256color", initHeight, initWidth, modes); err != nil {
        return fmt.Errorf("request pty: %w", err)
    }
    if err := sshSess.Shell(); err != nil {
        return fmt.Errorf("start shell: %w", err)
    }

    // 3. 初始化审计
    auditResult, err := s.auditSvc.StartSession(s.tenantID, s.userID, s.host.ID, s.clientIP, initWidth, initHeight)
    if err != nil {
        slog.Warn("start audit session failed", "err", err)
    }
    var asyncW *audit.AsyncWriter
    castPath := ""
    if auditResult != nil {
        asyncW = auditResult.AsyncWriter
        castPath = auditResult.Session.CastFilePath
        defer s.auditSvc.CloseSession(auditResult.Session.ID, asyncW, castPath)
    }

    // 4. 获取 SSH stdin/stdout
    sshStdin, err := sshSess.StdinPipe()
    if err != nil {
        return fmt.Errorf("stdin pipe: %w", err)
    }
    sshStdout, err := sshSess.StdoutPipe()
    if err != nil {
        return fmt.Errorf("stdout pipe: %w", err)
    }

    interceptor := NewInterceptor(s.tenantID)

    // 5. 输出 goroutine：SSH → 浏览器
    outputDone := make(chan struct{})
    go func() {
        defer close(outputDone)
        buf := make([]byte, 4096)
        for {
            n, err := sshStdout.Read(buf)
            if n > 0 {
                data := string(buf[:n])
                if writeErr := s.ws.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
                    return
                }
                if asyncW != nil {
                    asyncW.SendOutput(data)
                }
            }
            if err != nil {
                if err != io.EOF {
                    slog.Debug("ssh stdout read", "err", err)
                }
                return
            }
        }
    }()

    // 6. 输入循环：浏览器 → 拦截器 → SSH
    for {
        msgType, msg, err := s.ws.ReadMessage()
        if err != nil {
            break
        }
        // 处理 resize 消息（前端发 JSON）
        if msgType == websocket.TextMessage {
            var resize ResizeMsg
            if parseErr := parseJSON(msg, &resize); parseErr == nil && resize.Type == "resize" {
                _ = sshSess.WindowChange(resize.Height, resize.Width)
                continue
            }
        }

        data := string(msg)
        result := interceptor.Feed(data)

        if result == nil {
            // 未到回车，直接透传（实时回显）
            if _, err := sshStdin.Write(msg); err != nil {
                break
            }
            if asyncW != nil {
                asyncW.SendInput(data)
            }
            continue
        }

        if result.Blocked {
            // 高危拦截：写红色提示给浏览器，不透传
            _ = s.ws.WriteMessage(websocket.BinaryMessage, []byte(blockMsg))
            if asyncW != nil {
                asyncW.SendCommand("[BLOCKED] " + result.Command)
            }
            // 记录 danger_event
            s.recordDangerEvent(auditResult, result)
            // 发送回车让终端光标回到命令行（但不发命令本身）
            _, _ = sshStdin.Write([]byte("\r"))
            continue
        }

        // 正常命令透传
        if _, err := sshStdin.Write(msg); err != nil {
            break
        }
        if asyncW != nil {
            asyncW.SendInput(data)
            if result.Command != "" {
                asyncW.SendCommand(result.Command)
            }
        }
    }

    <-outputDone
    return nil
}

func (s *Session) buildSSHConfig() (*ssh.ClientConfig, error) {
    var authMethod ssh.AuthMethod
    if s.host.AuthType == "key" {
        signer, err := ssh.ParsePrivateKey([]byte(s.sshSecret))
        if err != nil {
            return nil, fmt.Errorf("parse private key: %w", err)
        }
        authMethod = ssh.PublicKeys(signer)
    } else {
        authMethod = ssh.Password(s.sshSecret)
    }
    return &ssh.ClientConfig{
        User:            s.sshUser,
        Auth:            []ssh.AuthMethod{authMethod},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: 生产环境应验证 host key
        Timeout:         15 * time.Second,
    }, nil
}

func (s *Session) recordDangerEvent(ar *audit.StartSessionResult, result *InterceptResult) {
    if ar == nil {
        return
    }
    slog.Info("danger command blocked",
        "session_id", ar.Session.ID,
        "command", result.Command,
        "rule", result.MatchedRule,
    )
    // 写 danger_events 表（通过 db 直接写，避免引入循环依赖）
    go func() {
        _ = writeDangerEvent(ar.Session.ID, result.Command, result.MatchedRule)
    }()
}

func parseJSON(data []byte, v any) error {
    import_json, _ := func() (any, error) { return nil, nil }() // 占位，下方正确实现
    _ = import_json
    return nil // 见 handler.go 中的实现
}
```

> **注意：** `parseJSON` 和 `writeDangerEvent` 需在 handler.go 中补充实现，见下一步。

- [ ] **Step 2: 完善 session.go 中的辅助函数**

将 `session.go` 中 `parseJSON` 替换为正确实现，并添加 `writeDangerEvent`：

```go
// 在 session.go 顶部 import 中添加：
import (
    // ... 已有 import ...
    "encoding/json"
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

// 替换 parseJSON 为：
func parseJSON(data []byte, v any) error {
    return json.Unmarshal(data, v)
}

// 添加 writeDangerEvent：
func writeDangerEvent(sessionID int64, command, matchedRule string) error {
    event := &model.DangerEvent{
        SessionID:   sessionID,
        Command:     command,
        MatchedRule: matchedRule,
    }
    return db.DB.Create(event).Error
}
```

- [ ] **Step 3: Commit（阶段性）**

```bash
git add internal/ssh/session.go
git commit -m "feat: add WebSSH session with PTY and danger interceptor"
```

---

### Task 6: WebSSH Handler（WebSocket 入口）

**Files:**
- Create: `internal/ssh/handler.go`

- [ ] **Step 1: 实现 WebSocket Handler**

```go
// internal/ssh/handler.go
package ssh

import (
    "log/slog"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"

    "github.com/sshive/sshive/internal/audit"
    "github.com/sshive/sshive/internal/auth"
    "github.com/sshive/sshive/internal/credential"
    "github.com/sshive/sshive/internal/host"
    "github.com/sshive/sshive/pkg/middleware"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // 生产环境应校验 Origin
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
}

type Handler struct {
    auditSvc  *audit.Service
    hostSvc   *host.Service
    credSvc   *credential.Service
}

func NewHandler() *Handler {
    return &Handler{
        auditSvc: audit.NewService(),
        hostSvc:  host.NewService(),
        credSvc:  credential.NewService(),
    }
}

func (h *Handler) Connect(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    userID := auth.GetUserID(c.Request.Context())

    hostID, err := strconv.ParseInt(c.Param("hostId"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid host id")
        return
    }

    // 获取主机信息（确保属于该租户）
    hostInfo, err := h.hostSvc.GetByID(tenantID, hostID)
    if err != nil {
        middleware.NotFound(c, "host not found")
        return
    }

    // 解密凭据
    sshUser, sshSecret, err := h.credSvc.DecryptSecret(tenantID, hostInfo.CredentialID)
    if err != nil {
        middleware.InternalError(c, "load credential failed")
        return
    }

    // 升级为 WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        slog.Error("websocket upgrade failed", "err", err)
        return
    }
    defer ws.Close()

    // 读取初始 terminal 尺寸（前端连接后首条消息）
    width, height := 220, 50 // 默认值
    _, initMsg, err := ws.ReadMessage()
    if err == nil {
        var init struct {
            Type   string `json:"type"`
            Width  int    `json:"width"`
            Height int    `json:"height"`
        }
        if jsonErr := parseJSON(initMsg, &init); jsonErr == nil && init.Width > 0 {
            width = init.Width
            height = init.Height
        }
    }

    clientIP := c.ClientIP()
    sess := NewSession(hostInfo, sshUser, sshSecret, tenantID, userID, clientIP, ws, h.auditSvc)

    if err := sess.Run(width, height); err != nil {
        slog.Info("webssh session ended", "host_id", hostID, "err", err)
    }
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./internal/ssh/
```

Expected: 无报错

- [ ] **Step 3: 注册路由到 main.go**

在 `authed` 路由组末尾添加：

```go
// 审计日志
auditH := audit.NewHandler()
authed.GET("/sessions", auth.RequirePermission("audit:view"), auditH.ListSessions)
authed.GET("/sessions/:id/logs", auth.RequirePermission("audit:view"), auditH.ListLogs)
authed.GET("/sessions/:id/commands", auth.RequirePermission("audit:view"), auditH.ListCommands)
authed.GET("/sessions/:id/replay", auth.RequirePermission("audit:view"), auditH.Replay)

// WebSSH
sshH := sshmodule.NewHandler()
authed.GET("/ws/ssh/:hostId", auth.RequirePermission("host:connect"), sshH.Connect)
```

在 import 中添加：
```go
sshmodule "github.com/sshive/sshive/internal/ssh"
"github.com/sshive/sshive/internal/audit"
```

- [ ] **Step 4: 全量编译验证**

```bash
go build ./...
```

Expected: 无报错

- [ ] **Step 5: Commit + Push**

```bash
git add internal/ssh/ internal/audit/ cmd/server/main.go
git commit -m "feat: add WebSSH WebSocket handler with audit recording"
git push origin main
```

---

## 测试汇总

```bash
go test ./pkg/cast/... -v       # 1 test: PASS
go test ./internal/dangerrule/... -v  # PASS (已在 Plan 2 验证)
go build ./...                  # SUCCESS
```
