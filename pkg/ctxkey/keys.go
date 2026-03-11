package ctxkey

type contextKey string

const (
	TenantID  contextKey = "tenant_id"
	UserID    contextKey = "user_id"
	UserRoles contextKey = "user_roles"
)
