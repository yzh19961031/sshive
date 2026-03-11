package auth

import (
	"context"
	"fmt"
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
			middleware.OK(c, nil)
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

func IsTokenBlacklisted(tokenStr string) bool {
	key := fmt.Sprintf("jwt:blacklist:%s", tokenStr)
	val, err := db.Redis.Get(context.Background(), key).Result()
	return err == nil && val == "1"
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
