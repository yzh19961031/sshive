package ssh

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"

	"github.com/sshive/sshive/internal/audit"
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

const blockMsg = "\r\n\033[31m[SSHive] 高危指令已被拦截\033[0m\r\n"

// ResizeMsg 是前端发来的终端尺寸变更消息
type ResizeMsg struct {
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Session 管理一个 WebSSH 会话的完整生命周期
type Session struct {
	host      *model.Host
	sshUser   string
	sshSecret string
	tenantID  int64
	userID    int64
	clientIP  string
	ws        *websocket.Conn
	auditSvc  *audit.Service
}

func NewSession(host *model.Host, sshUser, sshSecret string,
	tenantID, userID int64, clientIP string,
	ws *websocket.Conn, auditSvc *audit.Service) *Session {
	return &Session{
		host:      host,
		sshUser:   sshUser,
		sshSecret: sshSecret,
		tenantID:  tenantID,
		userID:    userID,
		clientIP:  clientIP,
		ws:        ws,
		auditSvc:  auditSvc,
	}
}

// Run 建立 SSH 连接，开启 PTY，运行输入输出循环
func (s *Session) Run(initWidth, initHeight int) error {
	// 1. 建立 SSH 连接
	cfg, err := s.buildSSHConfig()
	if err != nil {
		return fmt.Errorf("build ssh config: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", s.host.IP, s.host.Port)
	sshConn, err := gossh.Dial("tcp", addr, cfg)
	if err != nil {
		return fmt.Errorf("ssh dial %s: %w", addr, err)
	}
	defer sshConn.Close()

	// 2. 开启 Session + PTY
	sshSess, err := sshConn.NewSession()
	if err != nil {
		return fmt.Errorf("new ssh session: %w", err)
	}
	defer sshSess.Close()

	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSess.RequestPty("xterm-256color", initHeight, initWidth, modes); err != nil {
		return fmt.Errorf("request pty: %w", err)
	}

	// 3. 获取 stdin/stdout（必须在 Shell() 启动前调用）
	sshStdin, err := sshSess.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	sshStdout, err := sshSess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := sshSess.Shell(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}

	// 4. 初始化审计
	auditResult, err := s.auditSvc.StartSession(s.tenantID, s.userID, s.host.ID, s.clientIP, initWidth, initHeight)
	if err != nil {
		slog.Warn("start audit session failed", "err", err)
	}
	var asyncW *audit.AsyncWriter
	castPath := ""
	if auditResult != nil {
		asyncW = auditResult.AsyncWriter
		castPath = auditResult.Session.CastFilePath
		defer s.auditSvc.CloseSession(auditResult.Session.ID, asyncW, castPath)
	}

	interceptor := NewInterceptor(s.tenantID)

	// 5. 输出 goroutine：SSH → 浏览器
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		buf := make([]byte, 4096)
		for {
			n, err := sshStdout.Read(buf)
			if n > 0 {
				data := string(buf[:n])
				if writeErr := s.ws.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					return
				}
				if asyncW != nil {
					asyncW.SendOutput(data)
				}
			}
			if err != nil {
				if err != io.EOF {
					slog.Debug("ssh stdout read", "err", err)
				}
				return
			}
		}
	}()

	// 6. 输入循环：浏览器 → 拦截器 → SSH
	for {
		msgType, msg, err := s.ws.ReadMessage()
		if err != nil {
			break
		}
		if msgType == websocket.TextMessage {
			var resize ResizeMsg
			if jsonErr := parseJSON(msg, &resize); jsonErr == nil && resize.Type == "resize" {
				_ = sshSess.WindowChange(resize.Height, resize.Width)
				continue
			}
		}

		data := string(msg)
		result := interceptor.Feed(data)

		if result == nil {
			if _, err := sshStdin.Write(msg); err != nil {
				break
			}
			if asyncW != nil {
				asyncW.SendInput(data)
			}
			continue
		}

		if result.Blocked {
			_ = s.ws.WriteMessage(websocket.BinaryMessage, []byte(blockMsg))
			if asyncW != nil {
				asyncW.SendCommand("[BLOCKED] " + result.Command)
			}
			if auditResult != nil {
				go writeDangerEvent(auditResult.Session.ID, result.Command, result.MatchedRule)
			}
			_, _ = sshStdin.Write([]byte("\r"))
			continue
		}

		if _, err := sshStdin.Write(msg); err != nil {
			break
		}
		if asyncW != nil {
			asyncW.SendInput(data)
			if result.Command != "" {
				asyncW.SendCommand(result.Command)
			}
		}
	}

	<-outputDone
	return nil
}

func (s *Session) buildSSHConfig() (*gossh.ClientConfig, error) {
	var authMethod gossh.AuthMethod
	if s.host.AuthType == "key" {
		signer, err := gossh.ParsePrivateKey([]byte(s.sshSecret))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethod = gossh.PublicKeys(signer)
	} else {
		authMethod = gossh.Password(s.sshSecret)
	}
	return &gossh.ClientConfig{
		User:            s.sshUser,
		Auth:            []gossh.AuthMethod{authMethod},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}, nil
}

func parseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func writeDangerEvent(sessionID int64, command, matchedRule string) {
	event := &model.DangerEvent{
		SessionID:   sessionID,
		Command:     command,
		MatchedRule: matchedRule,
	}
	if err := db.DB.Create(event).Error; err != nil {
		slog.Warn("write danger event failed", "err", err)
	}
}
