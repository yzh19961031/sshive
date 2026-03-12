package ssh

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/sshive/sshive/internal/audit"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/credential"
	"github.com/sshive/sshive/internal/host"
	"github.com/sshive/sshive/pkg/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type Handler struct {
	auditSvc *audit.Service
	hostSvc  *host.Service
	credSvc  *credential.Service
}

func NewHandler() *Handler {
	return &Handler{
		auditSvc: audit.NewService(),
		hostSvc:  host.NewService(),
		credSvc:  credential.NewService(),
	}
}

func (h *Handler) Connect(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	userID := auth.GetUserID(c.Request.Context())

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
		slog.Error("websocket upgrade failed", "err", err)
		return
	}
	defer ws.Close()

	width, height := 220, 50
	_, initMsg, err := ws.ReadMessage()
	if err == nil {
		var init struct {
			Type   string `json:"type"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		}
		if jsonErr := parseJSON(initMsg, &init); jsonErr == nil && init.Width > 0 {
			width = init.Width
			height = init.Height
		}
	}

	clientIP := c.ClientIP()
	sess := NewSession(hostInfo, sshUser, sshSecret, tenantID, userID, clientIP, ws, h.auditSvc)

	if err := sess.Run(width, height); err != nil {
		slog.Info("webssh session ended", "host_id", hostID, "err", err)
	}
}
