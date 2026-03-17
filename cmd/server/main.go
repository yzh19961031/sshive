package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/internal/ai"
	"github.com/sshive/sshive/internal/audit"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/config"
	"github.com/sshive/sshive/internal/credential"
	"github.com/sshive/sshive/internal/dangerrule"
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/dbclient"
	"github.com/sshive/sshive/internal/host"
	"github.com/sshive/sshive/internal/hostgroup"
	sftpmodule "github.com/sshive/sshive/internal/sftp"
	sshmodule "github.com/sshive/sshive/internal/ssh"
	"github.com/sshive/sshive/internal/stats"
	"github.com/sshive/sshive/internal/tenant"
	"github.com/sshive/sshive/internal/user"
	"github.com/sshive/sshive/web"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := config.Load(*configPath); err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

	if err := db.InitMySQL(config.C.DB.DSN); err != nil {
		slog.Error("init mysql failed", "err", err)
		os.Exit(1)
	}

	if err := db.InitRedis(config.C.Redis.Addr, config.C.Redis.Password); err != nil {
		slog.Error("init redis failed", "err", err)
		os.Exit(1)
	}

	if err := db.Migrate(); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	if err := db.Seed(); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}

	// 清理服务重启残留的 active 会话
	audit.NewService().RecoverStaleSessions()

	r := gin.Default()

	api := r.Group("/api")
	api.POST("/auth/login", auth.LoginHandler(config.C.JWTSecret))

	authed := api.Group("/", auth.JWTMiddleware(config.C.JWTSecret))
	authed.POST("/auth/logout", auth.LogoutHandler(config.C.JWTSecret))

	th := tenant.NewHandler()
	// 公开接口：供登录页获取租户列表（无需认证）
	api.GET("/tenants/public", th.ListPublic)
	authed.GET("/tenants", auth.RequirePermission("tenant:manage"), th.List)
	authed.POST("/tenants", auth.RequirePermission("tenant:manage"), th.Create)

	uh := user.NewHandler()
	authed.GET("/users", auth.RequirePermission("user:manage"), uh.List)
	authed.POST("/users", auth.RequirePermission("user:manage"), uh.Create)
	authed.PUT("/users/:id/roles", auth.RequirePermission("user:manage"), uh.AssignRoles)
	authed.PUT("/users/:id", auth.RequirePermission("user:manage"), uh.Update)
	authed.DELETE("/users/:id", auth.RequirePermission("user:manage"), uh.Disable)
	authed.GET("/roles", auth.RequirePermission("user:manage"), uh.GetRoles)

	// 凭据管理
	credH := credential.NewHandler()
	authed.GET("/credentials", auth.RequirePermission("host:manage"), credH.List)
	authed.POST("/credentials", auth.RequirePermission("host:manage"), credH.Create)

	// 主机分组管理
	groupH := hostgroup.NewHandler()
	authed.GET("/host-groups", auth.RequirePermission("host:manage"), groupH.List)
	authed.POST("/host-groups", auth.RequirePermission("host:manage"), groupH.Create)
	authed.PUT("/host-groups/:id", auth.RequirePermission("host:manage"), groupH.Update)
	authed.DELETE("/host-groups/:id", auth.RequirePermission("host:manage"), groupH.Delete)

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

	// 审计日志
	auditH := audit.NewHandler()
	authed.GET("/sessions", auth.RequirePermission("audit:view"), auditH.ListSessions)
	authed.GET("/sessions/:id/logs", auth.RequirePermission("audit:view"), auditH.ListLogs)
	authed.GET("/sessions/:id/commands", auth.RequirePermission("audit:view"), auditH.ListCommands)
	authed.GET("/sessions/:id/replay", auth.RequirePermission("audit:view"), auditH.Replay)
	authed.GET("/commands", auth.RequirePermission("audit:view"), auditH.ListAllCommands)

	// 统计数据
	statsH := stats.NewHandler()
	// /stats 返回当前租户的聚合统计，登录即可访问（无需特定权限）
	authed.GET("/stats", statsH.Get)

	// DB Client
	dbRepo := dbclient.NewRepo(db.DB)
	dbSvc := dbclient.NewService(dbRepo)
	dbH := dbclient.NewHandler(dbSvc, dbRepo)
	authed.GET("/db-servers", auth.RequirePermission("db:read"), dbH.List)
	authed.POST("/db-servers", auth.RequirePermission("db:write"), dbH.Create)
	authed.DELETE("/db-servers/:id", auth.RequirePermission("db:write"), dbH.Delete)
	authed.POST("/db-servers/:id/query", auth.RequirePermission("db:write"), dbH.Query)
	authed.GET("/db-servers/:id/databases", auth.RequirePermission("db:read"), dbH.Databases)
	authed.GET("/db-servers/:id/databases/:db/tables", auth.RequirePermission("db:read"), dbH.Tables)

	// WebSSH
	sshH := sshmodule.NewHandler()
	authed.GET("/ws/ssh/:hostId", auth.RequirePermission("host:connect"), sshH.Connect)

	// SFTP
	sftpH := sftpmodule.NewHandler()
	authed.GET("/ws/sftp/:hostId", auth.RequirePermission("sftp:access"), sftpH.Connect)

	// AI Shell
	aiSvc := ai.NewService()
	aiH := ai.NewHandler(aiSvc)
	authed.POST("/ai/shell", aiH.Shell)

	// Static files (frontend SPA)
	distFS, err := fs.Sub(web.Static, "dist")
	if err != nil {
		slog.Error("embed fs error", "err", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(distFS))
	r.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if strings.HasPrefix(urlPath, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}
		// Check if the exact file exists; if so, serve it via http.FileServer
		// which handles MIME types, ETags, and Range requests correctly.
		filePath := strings.TrimPrefix(urlPath, "/")
		if filePath == "" {
			filePath = "index.html"
		}
		if f, err := distFS.Open(filePath); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA fallback: unknown paths → index.html
		data, _ := fs.ReadFile(distFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	addr := fmt.Sprintf(":%d", config.C.Port)
	slog.Info("server starting", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
