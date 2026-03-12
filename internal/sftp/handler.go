// internal/sftp/handler.go
package sftp

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/credential"
	"github.com/sshive/sshive/internal/host"
	"github.com/sshive/sshive/pkg/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024 * 1024,
	WriteBufferSize: 1024 * 1024,
}

type Handler struct {
	hostSvc *host.Service
	credSvc *credential.Service
}

func NewHandler() *Handler {
	return &Handler{
		hostSvc: host.NewService(),
		credSvc: credential.NewService(),
	}
}

func (h *Handler) Connect(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())

	hostID, err := strconv.ParseInt(c.Param("hostId"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid host id")
		return
	}

	hostInfo, err := h.hostSvc.GetByID(tenantID, hostID)
	if err != nil {
		middleware.NotFound(c, "host not found")
		return
	}

	sshUser, sshSecret, err := h.credSvc.DecryptSecret(tenantID, hostInfo.CredentialID)
	if err != nil {
		middleware.InternalError(c, "load credential failed")
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("sftp websocket upgrade failed", "err", err)
		return
	}
	defer ws.Close()

	sess := NewSession(hostInfo, sshUser, sshSecret, ws, 0)
	if err := sess.Run(); err != nil {
		slog.Info("sftp session ended", "host_id", hostID, "err", err)
	}
}
