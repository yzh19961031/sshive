package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/pkg/ctxkey"
	"github.com/sshive/sshive/pkg/middleware"
)

var rolePermissions = map[string][]string{
	"super_admin": {
		"host:connect", "host:manage", "sftp:access",
		"audit:view", "user:manage", "rule:manage", "tenant:manage",
		"db:read", "db:write",
	},
	"tenant_admin": {
		"host:connect", "host:manage", "sftp:access",
		"audit:view", "user:manage", "rule:manage",
		"db:read", "db:write",
	},
	"operator": {
		"host:connect", "sftp:access", "audit:view",
		"db:read", "db:write",
	},
	"auditor": {
		"audit:view",
		"db:read", // 用于审计页加载服务器下拉列表
	},
}

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

func GetTenantID(ctx context.Context) int64 {
	v, _ := ctx.Value(ctxkey.TenantID).(int64)
	return v
}

func GetUserID(ctx context.Context) int64 {
	v, _ := ctx.Value(ctxkey.UserID).(int64)
	return v
}
