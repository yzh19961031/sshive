// internal/dbclient/handler.go
package dbclient

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/model"
)

type Handler struct {
	svc  *Service
	repo *Repo
}

func NewHandler(svc *Service, repo *Repo) *Handler {
	return &Handler{svc: svc, repo: repo}
}

// List GET /api/db-servers
func (h *Handler) List(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	list, err := h.repo.List(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// 脱敏：不返回密码
	for i := range list {
		list[i].EncryptedPassword = ""
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

type CreateReq struct {
	Name     string `json:"name"     binding:"required"`
	Type     string `json:"type"     binding:"required,oneof=mysql postgres"` // mysql/postgres
	Host     string `json:"host"     binding:"required"`
	Port     int    `json:"port"     binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Database string `json:"database"`
}

// Create POST /api/db-servers
func (h *Handler) Create(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	enc, err := h.svc.EncryptPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "加密失败"})
		return
	}
	srv := &model.DBServer{
		TenantID: tenantID, Name: req.Name, Type: req.Type,
		Host: req.Host, Port: req.Port, Username: req.Username,
		EncryptedPassword: enc, Database: req.Database,
	}
	if err := h.repo.Create(srv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	srv.EncryptedPassword = ""
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": srv})
}

// Delete DELETE /api/db-servers/:id
func (h *Handler) Delete(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}
	if err := h.repo.Delete(tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

type QueryReq struct {
	SQL string `json:"sql" binding:"required"`
}

// Query POST /api/db-servers/:id/query
func (h *Handler) Query(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}
	var req QueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	result, err := h.svc.Query(tenantID, id, req.SQL)
	if err != nil {
		if strings.Contains(err.Error(), "不支持多语句") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// Databases GET /api/db-servers/:id/databases
func (h *Handler) Databases(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}
	dbs, err := h.svc.ListDatabases(tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dbs})
}

// Tables GET /api/db-servers/:id/databases/:db/tables
func (h *Handler) Tables(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}
	database := c.Param("db")
	tables, err := h.svc.ListTables(tenantID, id, database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tables})
}
