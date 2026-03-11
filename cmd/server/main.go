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

	api := r.Group("/api")
	api.POST("/auth/login", auth.LoginHandler(config.C.JWTSecret))

	authed := api.Group("/", auth.JWTMiddleware(config.C.JWTSecret))
	authed.POST("/auth/logout", auth.LogoutHandler(config.C.JWTSecret))

	th := tenant.NewHandler()
	authed.GET("/tenants", auth.RequirePermission("tenant:manage"), th.List)
	authed.POST("/tenants", auth.RequirePermission("tenant:manage"), th.Create)

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
