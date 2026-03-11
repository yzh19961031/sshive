# SSHive Plan 1: 后端基础 + 认证 + 多租户

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建 Go 后端项目骨架，实现配置加载、数据库/Redis 连接、AES 加解密工具、JWT 认证、RBAC 中间件、租户管理、用户管理，输出一个可启动的 HTTP 服务。

**Architecture:** 单体 Go 应用，Gin 框架，GORM + MySQL，Redis 存 JWT 黑名单。认证通过 JWT Bearer Token，RBAC 权限硬编码在角色-权限映射表中，多租户隔离通过中间件从 JWT 提取 tenant_id 注入 context 实现。

**Tech Stack:** Go 1.22+, Gin, GORM, golang-jwt/jwt v5, go-redis/redis v9, golang.org/x/crypto, testify

---

## 文件结构

```
sshive/
├── go.mod
├── go.sum
├── cmd/
│   └── server/
│       └── main.go                        # 程序入口：加载配置、初始化依赖、注册路由、启动服务
├── internal/
│   ├── config/
│   │   └── config.go                      # 从环境变量读取并校验所有配置项
│   ├── db/
│   │   ├── mysql.go                       # GORM 连接初始化
│   │   ├── redis.go                       # Redis 客户端初始化
│   │   └── migrate.go                     # AutoMigrate 所有模型
│   ├── model/
│   │   ├── tenant.go                      # Tenant 结构体
│   │   ├── user.go                        # User、Role、UserRole 结构体
│   │   └── permission.go                  # Permission、RolePermission 结构体
│   ├── auth/
│   │   ├── jwt.go                         # JWT 签发与解析（Claims 定义）
│   │   ├── handler.go                     # POST /api/auth/login, POST /api/auth/logout
│   │   ├── middleware.go                  # JWT 验证中间件、tenant_id 注入 context
│   │   ├── rbac.go                        # 角色→权限映射，RequirePermission 中间件
│   │   └── handler_test.go               # 登录/登出集成测试
│   ├── tenant/
│   │   ├── repo.go                        # Tenant CRUD（数据库操作）
│   │   ├── service.go                     # 业务逻辑（创建租户、列表）
│   │   ├── handler.go                     # GET/POST /api/tenants
│   │   └── service_test.go               # 单元测试
│   └── user/
│       ├── repo.go                        # User CRUD
│       ├── service.go                     # 业务逻辑（创建用户、分配角色、列表）
│       ├── handler.go                     # GET/POST /api/users, PUT /api/users/:id/roles
│       └── service_test.go               # 单元测试
├── pkg/
│   ├── encrypt/
│   │   ├── encrypt.go                     # AES-256-GCM 加密/解密
│   │   └── encrypt_test.go               # 单元测试
│   ├── middleware/
│   │   └── response.go                   # 统一响应格式 {code, message, data}
│   └── ctxkey/
│       └── keys.go                        # context key 常量（避免碰撞）
└── .env.example                           # 环境变量示例文件
```

---

## Chunk 1: 项目初始化与基础设施

### Task 1: Go 模块初始化

**Files:**
- Create: `go.mod`
- Create: `.env.example`
- Create: `cmd/server/main.go`（骨架）

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd /Volumes/data/personal/code/sshive
go mod init github.com/sshive/sshive
```

Expected: 生成 `go.mod`，内容为 `module github.com/sshive/sshive`

- [ ] **Step 2: 安装依赖**

```bash
go get github.com/gin-gonic/gin@v1.9.1
go get gorm.io/gorm@v1.25.7
go get gorm.io/driver/mysql@v1.5.4
go get github.com/redis/go-redis/v9@v9.5.1
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get golang.org/x/crypto@v0.21.0
go get github.com/stretchr/testify@v1.9.0
go mod tidy
```

Expected: `go.sum` 生成，依赖全部 resolved

- [ ] **Step 3: 创建 .env.example**

```bash
cat > .env.example << 'EOF'
SSHIVE_DB_DSN=root:password@tcp(127.0.0.1:3306)/sshive?parseTime=true&charset=utf8mb4
SSHIVE_REDIS_ADDR=127.0.0.1:6379
SSHIVE_REDIS_PASSWORD=
SSHIVE_ENCRYPT_KEY=base64编码的32字节密钥
SSHIVE_CAST_DIR=/tmp/sshive/cast
SSHIVE_JWT_SECRET=your-jwt-secret-here
SSHIVE_PORT=8080
EOF
```

- [ ] **Step 4: 创建 cmd/server/main.go 骨架**

```go
// cmd/server/main.go
package main

import (
    "log/slog"
    "os"
)

func main() {
    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
    slog.Info("sshive starting")
}
```

- [ ] **Step 5: 验证编译通过**

```bash
go build ./cmd/server/
```

Expected: 无报错，生成二进制文件

- [ ] **Step 6: Commit**

```bash
git init
git add go.mod go.sum .env.example cmd/
git commit -m "chore: init go module and project skeleton"
```

---

### Task 2: 配置加载

**Files:**
- Create: `internal/config/config.go`

- [ ] **Step 1: 创建配置结构体**

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    DBDsn          string
    RedisAddr      string
    RedisPassword  string
    EncryptKey     string // base64，32字节解码后使用
    CastDir        string
    JWTSecret      string
    Port           int
}

var C Config

func Load() error {
    C = Config{
        DBDsn:         requireEnv("SSHIVE_DB_DSN"),
        RedisAddr:     requireEnv("SSHIVE_REDIS_ADDR"),
        RedisPassword: os.Getenv("SSHIVE_REDIS_PASSWORD"),
        EncryptKey:    requireEnv("SSHIVE_ENCRYPT_KEY"),
        CastDir:       getEnvOrDefault("SSHIVE_CAST_DIR", "/data/sshive/cast"),
        JWTSecret:     requireEnv("SSHIVE_JWT_SECRET"),
    }
    portStr := getEnvOrDefault("SSHIVE_PORT", "8080")
    port, err := strconv.Atoi(portStr)
    if err != nil {
        return fmt.Errorf("invalid SSHIVE_PORT: %w", err)
    }
    C.Port = port
    return nil
}

func requireEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        panic(fmt.Sprintf("required env var %s is not set", key))
    }
    return v
}

func getEnvOrDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./internal/config/
```

Expected: 无报错

- [ ] **Step 3: Commit**

```bash
git add internal/config/
git commit -m "feat: add config loader from env vars"
```

---

### Task 3: 统一响应格式与 context key

**Files:**
- Create: `pkg/middleware/response.go`
- Create: `pkg/ctxkey/keys.go`

- [ ] **Step 1: 创建 context key 常量**

```go
// pkg/ctxkey/keys.go
package ctxkey

type contextKey string

const (
    TenantID  contextKey = "tenant_id"
    UserID    contextKey = "user_id"
    UserRoles contextKey = "user_roles"
)
```

- [ ] **Step 2: 创建统一响应格式**

```go
// pkg/middleware/response.go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type Response struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
}

func OK(c *gin.Context, data any) {
    c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

func Fail(c *gin.Context, httpStatus int, code int, msg string) {
    c.JSON(httpStatus, Response{Code: code, Message: msg, Data: nil})
}

func BadRequest(c *gin.Context, msg string) {
    Fail(c, http.StatusBadRequest, 400, msg)
}

func Unauthorized(c *gin.Context, msg string) {
    Fail(c, http.StatusUnauthorized, 401, msg)
}

func Forbidden(c *gin.Context, msg string) {
    Fail(c, http.StatusForbidden, 403, msg)
}

func NotFound(c *gin.Context, msg string) {
    Fail(c, http.StatusNotFound, 404, msg)
}

func InternalError(c *gin.Context, msg string) {
    Fail(c, http.StatusInternalServerError, 500, msg)
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./pkg/...
```

- [ ] **Step 4: Commit**

```bash
git add pkg/
git commit -m "feat: add response helper and context keys"
```

---

### Task 4: 数据库模型

**Files:**
- Create: `internal/model/tenant.go`
- Create: `internal/model/user.go`
- Create: `internal/model/permission.go`

- [ ] **Step 1: 创建 Tenant 模型**

```go
// internal/model/tenant.go
package model

import "time"

type Tenant struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"size:100;not null" json:"name"`
    Status    int8      `gorm:"default:1;not null" json:"status"` // 1:正常 0:禁用
    CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 2: 创建 User / Role 模型**

```go
// internal/model/user.go
package model

import "time"

type User struct {
    ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID     int64     `gorm:"not null;index" json:"tenant_id"`
    Username     string    `gorm:"size:50;not null;uniqueIndex:uk_tenant_username" json:"username"`
    PasswordHash string    `gorm:"size:255;not null" json:"-"`
    Email        string    `gorm:"size:100" json:"email"`
    Status       int8      `gorm:"default:1;not null" json:"status"`
    CreatedAt    time.Time `json:"created_at"`
}

type Role struct {
    ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID    int64  `gorm:"not null" json:"tenant_id"`
    Name        string `gorm:"size:50;not null" json:"name"`
    Description string `gorm:"size:255" json:"description"`
}

type UserRole struct {
    UserID int64 `gorm:"primaryKey" json:"user_id"`
    RoleID int64 `gorm:"primaryKey" json:"role_id"`
}
```

- [ ] **Step 3: 创建 Permission 模型**

```go
// internal/model/permission.go
package model

type Permission struct {
    ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
    Code        string `gorm:"size:100;not null;uniqueIndex" json:"code"`
    Description string `gorm:"size:255" json:"description"`
}

type RolePermission struct {
    RoleID       int64 `gorm:"primaryKey" json:"role_id"`
    PermissionID int64 `gorm:"primaryKey" json:"permission_id"`
}
```

- [ ] **Step 4: 编译验证**

```bash
go build ./internal/model/
```

- [ ] **Step 5: Commit**

```bash
git add internal/model/
git commit -m "feat: add database models (tenant, user, role, permission)"
```

---

### Task 5: 数据库与 Redis 连接

**Files:**
- Create: `internal/db/mysql.go`
- Create: `internal/db/redis.go`
- Create: `internal/db/migrate.go`

- [ ] **Step 1: MySQL 连接**

```go
// internal/db/mysql.go
package db

import (
    "fmt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMySQL(dsn string) error {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    })
    if err != nil {
        return fmt.Errorf("connect mysql: %w", err)
    }
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    DB = db
    return nil
}
```

- [ ] **Step 2: Redis 连接**

```go
// internal/db/redis.go
package db

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(addr, password string) error {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
    })
    if err := rdb.Ping(context.Background()).Err(); err != nil {
        return fmt.Errorf("connect redis: %w", err)
    }
    Redis = rdb
    return nil
}
```

- [ ] **Step 3: 数据库迁移**

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
    )
}
```

- [ ] **Step 4: 编译验证**

```bash
go build ./internal/db/
```

- [ ] **Step 5: Commit**

```bash
git add internal/db/
git commit -m "feat: add mysql and redis initialization with auto-migrate"
```

---

## Chunk 2: AES 加解密 + JWT 认证

### Task 6: AES-256-GCM 加解密

**Files:**
- Create: `pkg/encrypt/encrypt.go`
- Create: `pkg/encrypt/encrypt_test.go`

- [ ] **Step 1: 写失败测试**

```go
// pkg/encrypt/encrypt_test.go
package encrypt_test

import (
    "testing"
    "github.com/sshive/sshive/pkg/encrypt"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
    key := make([]byte, 32)
    for i := range key { key[i] = byte(i) }

    plaintext := "super-secret-password-123"
    ciphertext, err := encrypt.Encrypt(key, plaintext)
    require.NoError(t, err)
    assert.NotEqual(t, plaintext, ciphertext)

    decrypted, err := encrypt.Decrypt(key, ciphertext)
    require.NoError(t, err)
    assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDifferentEachTime(t *testing.T) {
    key := make([]byte, 32)
    c1, _ := encrypt.Encrypt(key, "hello")
    c2, _ := encrypt.Encrypt(key, "hello")
    assert.NotEqual(t, c1, c2) // nonce 随机，每次不同
}

func TestDecryptWrongKey(t *testing.T) {
    key1 := make([]byte, 32)
    key2 := make([]byte, 32)
    key2[0] = 1

    ciphertext, _ := encrypt.Encrypt(key1, "secret")
    _, err := encrypt.Decrypt(key2, ciphertext)
    assert.Error(t, err)
}
```

- [ ] **Step 2: 运行确认失败**

```bash
go test ./pkg/encrypt/ -v
```

Expected: FAIL with "undefined: encrypt.Encrypt"

- [ ] **Step 3: 实现加解密**

```go
// pkg/encrypt/encrypt.go
package encrypt

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
)

// Encrypt 使用 AES-256-GCM 加密，返回 base64 编码的密文
// key 必须是 32 字节
func Encrypt(key []byte, plaintext string) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("new cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("new gcm: %w", err)
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("generate nonce: %w", err)
    }
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Encrypt 返回的 base64 密文
func Decrypt(key []byte, encoded string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return "", fmt.Errorf("base64 decode: %w", err)
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("new cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("new gcm: %w", err)
    }
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", fmt.Errorf("decrypt: %w", err)
    }
    return string(plaintext), nil
}

// KeyFromBase64 从 base64 字符串解码 32 字节密钥
func KeyFromBase64(b64 string) ([]byte, error) {
    key, err := base64.StdEncoding.DecodeString(b64)
    if err != nil {
        return nil, fmt.Errorf("decode key: %w", err)
    }
    if len(key) != 32 {
        return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
    }
    return key, nil
}
```

- [ ] **Step 4: 运行确认通过**

```bash
go test ./pkg/encrypt/ -v
```

Expected: PASS，3 个测试全部通过

- [ ] **Step 5: Commit**

```bash
git add pkg/encrypt/
git commit -m "feat: add AES-256-GCM encrypt/decrypt"
```

---

### Task 7: JWT 签发与解析

**Files:**
- Create: `internal/auth/jwt.go`

- [ ] **Step 1: 实现 JWT 工具**

```go
// internal/auth/jwt.go
package auth

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

const jwtExpiry = 8 * time.Hour

type Claims struct {
    UserID   int64  `json:"uid"`
    TenantID int64  `json:"tid"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func SignToken(secret string, userID, tenantID int64, username, role string) (string, error) {
    claims := Claims{
        UserID:   userID,
        TenantID: tenantID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ParseToken(secret, tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return []byte(secret), nil
    })
    if err != nil {
        return nil, err
    }
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    return claims, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./internal/auth/
```

Expected: 无报错

- [ ] **Step 3: Commit**

```bash
git add internal/auth/jwt.go
git commit -m "feat: add JWT sign and parse"
```

---

### Task 8: 认证 Handler（登录/登出）

**Files:**
- Create: `internal/auth/handler.go`
- Create: `internal/auth/handler_test.go`

- [ ] **Step 1: 实现登录/登出 Handler**

```go
// internal/auth/handler.go
package auth

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
    "github.com/sshive/sshive/pkg/middleware"
    "golang.org/x/crypto/bcrypt"
)

type loginReq struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    TenantID int64  `json:"tenant_id" binding:"required"`
}

func LoginHandler(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req loginReq
        if err := c.ShouldBindJSON(&req); err != nil {
            middleware.BadRequest(c, err.Error())
            return
        }

        var user model.User
        if err := db.DB.Where("tenant_id = ? AND username = ? AND status = 1",
            req.TenantID, req.Username).First(&user).Error; err != nil {
            middleware.Unauthorized(c, "用户名或密码错误")
            return
        }
        if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
            middleware.Unauthorized(c, "用户名或密码错误")
            return
        }

        // 获取用户角色（取第一个角色）
        var userRole model.UserRole
        var roleName string
        if err := db.DB.Where("user_id = ?", user.ID).First(&userRole).Error; err == nil {
            var role model.Role
            if err := db.DB.First(&role, userRole.RoleID).Error; err == nil {
                roleName = role.Name
            }
        }

        token, err := SignToken(jwtSecret, user.ID, user.TenantID, user.Username, roleName)
        if err != nil {
            middleware.InternalError(c, "生成 token 失败")
            return
        }

        middleware.OK(c, gin.H{"token": token})
    }
}

func LogoutHandler(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenStr == "" {
            middleware.BadRequest(c, "missing token")
            return
        }
        claims, err := ParseToken(jwtSecret, tokenStr)
        if err != nil {
            middleware.OK(c, nil) // token 已失效，直接返回成功
            return
        }
        ttl := time.Until(claims.ExpiresAt.Time)
        if ttl > 0 {
            key := fmt.Sprintf("jwt:blacklist:%s", tokenStr)
            db.Redis.Set(context.Background(), key, 1, ttl)
        }
        middleware.OK(c, nil)
    }
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func IsTokenBlacklisted(tokenStr string) bool {
    key := fmt.Sprintf("jwt:blacklist:%s", tokenStr)
    val, err := db.Redis.Get(context.Background(), key).Result()
    return err == nil && val == "1"
}

// HashPassword 生成 bcrypt 密码哈希
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}
```

- [ ] **Step 2: 写集成测试（使用 httptest）**

```go
// internal/auth/handler_test.go
package auth_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/internal/auth"
    "github.com/stretchr/testify/assert"
)

func setupRouter(secret string) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.POST("/api/auth/login", auth.LoginHandler(secret))
    r.POST("/api/auth/logout", auth.LogoutHandler(secret))
    return r
}

func TestLoginMissingFields(t *testing.T) {
    r := setupRouter("secret")
    body := `{"username": "admin"}`
    req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./internal/auth/ -v -run TestLoginMissingFields
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/auth/
git commit -m "feat: add login/logout handlers with JWT and Redis blacklist"
```

---

### Task 9: JWT 认证中间件 + RBAC

**Files:**
- Create: `internal/auth/middleware.go`
- Create: `internal/auth/rbac.go`

- [ ] **Step 1: 实现 JWT 验证中间件**

```go
// internal/auth/middleware.go
package auth

import (
    "context"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/pkg/ctxkey"
    "github.com/sshive/sshive/pkg/middleware"
)

func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenStr == "" {
            middleware.Unauthorized(c, "missing token")
            c.Abort()
            return
        }
        if IsTokenBlacklisted(tokenStr) {
            middleware.Unauthorized(c, "token revoked")
            c.Abort()
            return
        }
        claims, err := ParseToken(jwtSecret, tokenStr)
        if err != nil {
            middleware.Unauthorized(c, "invalid token")
            c.Abort()
            return
        }
        // 注入 context
        ctx := context.WithValue(c.Request.Context(), ctxkey.TenantID, claims.TenantID)
        ctx = context.WithValue(ctx, ctxkey.UserID, claims.UserID)
        ctx = context.WithValue(ctx, ctxkey.UserRoles, claims.Role)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

- [ ] **Step 2: 实现 RBAC 权限映射**

```go
// internal/auth/rbac.go
package auth

import (
    "context"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/pkg/ctxkey"
    "github.com/sshive/sshive/pkg/middleware"
)

// rolePermissions 角色→权限 硬编码映射（第一期）
var rolePermissions = map[string][]string{
    "super_admin": {
        "host:connect", "host:manage", "sftp:access",
        "audit:view", "user:manage", "rule:manage", "tenant:manage",
    },
    "tenant_admin": {
        "host:connect", "host:manage", "sftp:access",
        "audit:view", "user:manage", "rule:manage",
    },
    "operator": {
        "host:connect", "sftp:access", "audit:view",
    },
    "auditor": {
        "audit:view",
    },
}

// HasPermission 检查角色是否拥有某权限
func HasPermission(role, permission string) bool {
    perms, ok := rolePermissions[role]
    if !ok {
        return false
    }
    for _, p := range perms {
        if p == permission {
            return true
        }
    }
    return false
}

// RequirePermission 返回检查指定权限的中间件
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, _ := c.Request.Context().Value(ctxkey.UserRoles).(string)
        if !HasPermission(role, permission) {
            middleware.Forbidden(c, "permission denied")
            c.Abort()
            return
        }
        c.Next()
    }
}

// GetTenantID 从 context 提取 tenant_id
func GetTenantID(ctx context.Context) int64 {
    v, _ := ctx.Value(ctxkey.TenantID).(int64)
    return v
}

// GetUserID 从 context 提取 user_id
func GetUserID(ctx context.Context) int64 {
    v, _ := ctx.Value(ctxkey.UserID).(int64)
    return v
}
```

- [ ] **Step 3: 写 RBAC 单元测试**

```go
// internal/auth/rbac_test.go
package auth_test

import (
    "testing"
    "github.com/sshive/sshive/internal/auth"
    "github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
    assert.True(t, auth.HasPermission("super_admin", "tenant:manage"))
    assert.True(t, auth.HasPermission("operator", "host:connect"))
    assert.False(t, auth.HasPermission("operator", "user:manage"))
    assert.False(t, auth.HasPermission("auditor", "host:connect"))
    assert.False(t, auth.HasPermission("unknown_role", "host:connect"))
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/auth/ -v -run TestHasPermission
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/auth/middleware.go internal/auth/rbac.go internal/auth/rbac_test.go
git commit -m "feat: add JWT middleware and RBAC permission check"
```

---

## Chunk 3: 租户与用户管理

### Task 10: 租户管理（Repo + Service + Handler）

**Files:**
- Create: `internal/tenant/repo.go`
- Create: `internal/tenant/service.go`
- Create: `internal/tenant/handler.go`
- Create: `internal/tenant/service_test.go`

- [ ] **Step 1: 实现 Tenant Repo**

```go
// internal/tenant/repo.go
package tenant

import (
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(t *model.Tenant) error {
    return db.DB.Create(t).Error
}

func (r *Repo) List(page, pageSize int) ([]model.Tenant, int64, error) {
    var tenants []model.Tenant
    var total int64
    offset := (page - 1) * pageSize
    if err := db.DB.Model(&model.Tenant{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := db.DB.Offset(offset).Limit(pageSize).Find(&tenants).Error; err != nil {
        return nil, 0, err
    }
    return tenants, total, nil
}

func (r *Repo) GetByID(id int64) (*model.Tenant, error) {
    var t model.Tenant
    if err := db.DB.First(&t, id).Error; err != nil {
        return nil, err
    }
    return &t, nil
}
```

- [ ] **Step 2: 实现 Tenant Service**

```go
// internal/tenant/service.go
package tenant

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
    Name string `json:"name" binding:"required,min=2,max=100"`
}

func (s *Service) Create(req CreateReq) (*model.Tenant, error) {
    t := &model.Tenant{Name: req.Name, Status: 1}
    if err := s.repo.Create(t); err != nil {
        return nil, fmt.Errorf("create tenant: %w", err)
    }
    return t, nil
}

func (s *Service) List(page, pageSize int) ([]model.Tenant, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.List(page, pageSize)
}
```

- [ ] **Step 3: 实现 Tenant Handler**

```go
// internal/tenant/handler.go
package tenant

import (
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/pkg/middleware"
)

type Handler struct {
    svc *Service
}

func NewHandler() *Handler {
    return &Handler{svc: NewService()}
}

func (h *Handler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    tenants, total, err := h.svc.List(page, pageSize)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": tenants})
}

func (h *Handler) Create(c *gin.Context) {
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    t, err := h.svc.Create(req)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, t)
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/tenant/
git commit -m "feat: add tenant CRUD (repo/service/handler)"
```

---

### Task 11: 用户管理（Repo + Service + Handler）

**Files:**
- Create: `internal/user/repo.go`
- Create: `internal/user/service.go`
- Create: `internal/user/handler.go`
- Create: `internal/user/service_test.go`

- [ ] **Step 1: 实现 User Repo**

```go
// internal/user/repo.go
package user

import (
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(u *model.User) error {
    return db.DB.Create(u).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.User, int64, error) {
    var users []model.User
    var total int64
    offset := (page - 1) * pageSize
    q := db.DB.Model(&model.User{}).Where("tenant_id = ?", tenantID)
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    if err := q.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
        return nil, 0, err
    }
    return users, total, nil
}

func (r *Repo) GetByID(tenantID, userID int64) (*model.User, error) {
    var u model.User
    if err := db.DB.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&u).Error; err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *Repo) SetRoles(userID int64, roleIDs []int64) error {
    return db.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
            return err
        }
        for _, rid := range roleIDs {
            if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil {
                return err
            }
        }
        return nil
    })
}
```

- [ ] **Step 2: 注意 Repo 需要导入 gorm**

在 `repo.go` 顶部添加导入：

```go
import (
    "gorm.io/gorm"
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/model"
)
```

- [ ] **Step 3: 实现 User Service**

```go
// internal/user/service.go
package user

import (
    "fmt"
    "github.com/sshive/sshive/internal/auth"
    "github.com/sshive/sshive/internal/model"
)

type Service struct {
    repo *Repo
}

func NewService() *Service {
    return &Service{repo: &Repo{}}
}

type CreateReq struct {
    Username string `json:"username" binding:"required,min=2,max=50"`
    Password string `json:"password" binding:"required,min=8"`
    Email    string `json:"email"`
}

type AssignRolesReq struct {
    RoleIDs []int64 `json:"role_ids" binding:"required"`
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.User, error) {
    hash, err := auth.HashPassword(req.Password)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }
    u := &model.User{
        TenantID:     tenantID,
        Username:     req.Username,
        PasswordHash: hash,
        Email:        req.Email,
        Status:       1,
    }
    if err := s.repo.Create(u); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }
    return u, nil
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.User, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.List(tenantID, page, pageSize)
}

func (s *Service) AssignRoles(tenantID, userID int64, req AssignRolesReq) error {
    if _, err := s.repo.GetByID(tenantID, userID); err != nil {
        return fmt.Errorf("user not found")
    }
    return s.repo.SetRoles(userID, req.RoleIDs)
}
```

- [ ] **Step 4: 实现 User Handler**

```go
// internal/user/handler.go
package user

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
    users, total, err := h.svc.List(tenantID, page, pageSize)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, gin.H{"total": total, "list": users})
}

func (h *Handler) Create(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    u, err := h.svc.Create(tenantID, req)
    if err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, u)
}

func (h *Handler) AssignRoles(c *gin.Context) {
    tenantID := auth.GetTenantID(c.Request.Context())
    userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        middleware.BadRequest(c, "invalid user id")
        return
    }
    var req AssignRolesReq
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.BadRequest(c, err.Error())
        return
    }
    if err := h.svc.AssignRoles(tenantID, userID, req); err != nil {
        middleware.InternalError(c, err.Error())
        return
    }
    middleware.OK(c, nil)
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./internal/...
```

Expected: 无报错

- [ ] **Step 6: Commit**

```bash
git add internal/user/
git commit -m "feat: add user CRUD and role assignment"
```

---

### Task 12: 组装服务器入口

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 完整实现 main.go**

```go
// cmd/server/main.go
package main

import (
    "fmt"
    "log/slog"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/sshive/sshive/internal/auth"
    "github.com/sshive/sshive/internal/config"
    "github.com/sshive/sshive/internal/db"
    "github.com/sshive/sshive/internal/tenant"
    "github.com/sshive/sshive/internal/user"
)

func main() {
    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

    if err := config.Load(); err != nil {
        slog.Error("load config failed", "err", err)
        os.Exit(1)
    }

    if err := db.InitMySQL(config.C.DBDsn); err != nil {
        slog.Error("init mysql failed", "err", err)
        os.Exit(1)
    }

    if err := db.InitRedis(config.C.RedisAddr, config.C.RedisPassword); err != nil {
        slog.Error("init redis failed", "err", err)
        os.Exit(1)
    }

    if err := db.Migrate(); err != nil {
        slog.Error("migrate failed", "err", err)
        os.Exit(1)
    }

    r := gin.Default()

    // 公开路由
    api := r.Group("/api")
    api.POST("/auth/login", auth.LoginHandler(config.C.JWTSecret))

    // 需要认证的路由
    authed := api.Group("/", auth.JWTMiddleware(config.C.JWTSecret))
    authed.POST("/auth/logout", auth.LogoutHandler(config.C.JWTSecret))

    // 租户管理（super_admin only）
    th := tenant.NewHandler()
    authed.GET("/tenants", auth.RequirePermission("tenant:manage"), th.List)
    authed.POST("/tenants", auth.RequirePermission("tenant:manage"), th.Create)

    // 用户管理
    uh := user.NewHandler()
    authed.GET("/users", auth.RequirePermission("user:manage"), uh.List)
    authed.POST("/users", auth.RequirePermission("user:manage"), uh.Create)
    authed.PUT("/users/:id/roles", auth.RequirePermission("user:manage"), uh.AssignRoles)

    addr := fmt.Sprintf(":%d", config.C.Port)
    slog.Info("server starting", "addr", addr)
    if err := r.Run(addr); err != nil {
        slog.Error("server error", "err", err)
        os.Exit(1)
    }
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./cmd/server/
```

Expected: 生成二进制，无报错

- [ ] **Step 3: 运行冒烟测试（需要本地 MySQL + Redis）**

```bash
# 先确保 MySQL 中存在 sshive 数据库
# mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS sshive CHARACTER SET utf8mb4;"

SSHIVE_DB_DSN="root:password@tcp(127.0.0.1:3306)/sshive?parseTime=true&charset=utf8mb4" \
SSHIVE_REDIS_ADDR="127.0.0.1:6379" \
SSHIVE_ENCRYPT_KEY="AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" \
SSHIVE_JWT_SECRET="test-secret" \
./server &

# 测试登录接口（期望 401，因为还没有用户）
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"test","tenant_id":1}' | jq .

kill %1
```

Expected: `{"code":401,"message":"用户名或密码错误","data":null}`

- [ ] **Step 4: 最终 Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: wire server entry point with all routes"
```

---

## 测试运行汇总

运行 Plan 1 所有测试：

```bash
go test ./pkg/encrypt/... -v
go test ./internal/auth/... -v
go build ./...
```

Expected:
```
pkg/encrypt: PASS (3 tests)
internal/auth: PASS (2 tests)
build: SUCCESS
```
