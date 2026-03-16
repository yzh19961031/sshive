package audit

import (
	"os"
	"strconv"
	"time"

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

	var f CommandFilter
	f.Command = c.Query("command")
	f.Action = c.Query("action")
	if s := c.Query("start_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			f.StartTime = &t
		}
	}
	if s := c.Query("end_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			f.EndTime = &t
		}
	}

	list, total, err := h.svc.ListCommands(tenantID, sessionID, page, pageSize, f)
	if err != nil {
		middleware.NotFound(c, err.Error())
		return
	}
	middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) ListAllCommands(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.ListAllCommands(tenantID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, gin.H{"total": total, "list": list})
}

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
