// internal/sftp/handler.go
package sftp

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/credential"
	"github.com/sshive/sshive/internal/host"
	"github.com/sshive/sshive/pkg/middleware"
)

func friendlySFTPError(err error) string {
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
		msg := friendlySFTPError(err)
		data, _ := json.Marshal(Response{Type: MsgError, Success: false, Message: msg})
		_ = ws.WriteMessage(websocket.TextMessage, data)
	}
}
