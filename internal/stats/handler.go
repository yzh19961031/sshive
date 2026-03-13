package stats

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/pkg/middleware"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Get(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	today := time.Now().UTC().Truncate(24 * time.Hour)

	var totalHosts int64
	if res := db.DB.Table("hosts").Where("tenant_id = ? AND status != 0", tenantID).Count(&totalHosts); res.Error != nil {
		middleware.InternalError(c, "stats query failed")
		return
	}

	var todaySessions int64
	if res := db.DB.Table("sessions").Where("tenant_id = ? AND started_at >= ?", tenantID, today).Count(&todaySessions); res.Error != nil {
		middleware.InternalError(c, "stats query failed")
		return
	}

	var activeSessions int64
	if res := db.DB.Table("sessions").Where("tenant_id = ? AND status = 'active'", tenantID).Count(&activeSessions); res.Error != nil {
		middleware.InternalError(c, "stats query failed")
		return
	}

	var todayDangerEvents int64
	if res := db.DB.Table("danger_events").
		Joins("JOIN sessions ON danger_events.session_id = sessions.id").
		Where("sessions.tenant_id = ? AND danger_events.triggered_at >= ?", tenantID, today).
		Count(&todayDangerEvents); res.Error != nil {
		middleware.InternalError(c, "stats query failed")
		return
	}

	middleware.OK(c, gin.H{
		"total_hosts":         totalHosts,
		"today_sessions":      todaySessions,
		"active_sessions":     activeSessions,
		"today_danger_events": todayDangerEvents,
	})
}
