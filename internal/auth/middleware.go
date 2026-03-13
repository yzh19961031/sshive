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
		// Try Authorization header first; fall back to ?token= query param (WebSocket clients
		// cannot set custom headers, so they pass the JWT as a query parameter).
		authHeader := c.GetHeader("Authorization")
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
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
		ctx := context.WithValue(c.Request.Context(), ctxkey.TenantID, claims.TenantID)
		ctx = context.WithValue(ctx, ctxkey.UserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxkey.UserRoles, claims.Role)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
