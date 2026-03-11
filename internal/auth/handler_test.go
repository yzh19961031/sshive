package auth_test

import (
	"bytes"
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
