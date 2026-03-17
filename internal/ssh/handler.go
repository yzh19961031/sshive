package ssh

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/sshive/sshive/internal/audit"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/host"
	"github.com/sshive/sshive/pkg/middleware"
)

func friendlySSHError(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "unable to authenticate") || strings.Contains(s, "no supported methods"):
		return "认证失败：用户名或密码/密钥不正确，或服务器未开启密码登录"
	case strings.Contains(s, "connection refused"):
		return "连接被拒绝：请检查主机 IP 和端口是否正确"
	case strings.Contains(s, "no route to host") || strings.Contains(s, "network is unreachable"):
		return "网络不可达：主机不在线或网络不通"
	case strings.Contains(s, "i/o timeout") || strings.Contains(s, "deadline exceeded") || strings.Contains(s, "timed out"):
		return "连接超时：主机无响应，请检查 IP、端口和防火墙设置"
	case strings.Contains(s, "parse private key") || strings.Contains(s, "cannot decode"):
		return "私钥格式错误：请检查凭据中的私钥内容是否为有效的 PEM 格式"
	case strings.Contains(s, "no such host"):
		return "主机名无法解析：请检查 IP 地址是否正确"
	default:
		return "连接失败：" + s
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type Handler struct {
	auditSvc *audit.Service
	hostSvc  *host.Service
}

func NewHandler() *Handler {
	return &Handler{
		auditSvc: audit.NewService(),
		hostSvc:  host.NewService(),
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

	hwj, err := h.hostSvc.GetWithJumpCredential(tenantID, hostID)
	if err != nil {
		middleware.NotFound(c, "host not found")
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
	sess := NewSession(
		hwj.Host, hwj.SSHUser, hwj.SSHSecret,
		hwj.JumpHost, hwj.JumpSSHUser, hwj.JumpSSHSecret,
		tenantID, userID, clientIP, ws, h.auditSvc,
	)

	if err := sess.Run(width, height); err != nil {
		slog.Info("webssh session ended", "host_id", hostID, "err", err)
		msg := friendlySSHError(err)
		_ = ws.WriteMessage(websocket.BinaryMessage,
			[]byte("\r\n\x1b[31m[SSHive] "+msg+"\x1b[0m\r\n"))
	}
}
