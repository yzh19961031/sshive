// internal/sftp/session.go
package sftp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	pkgsftp "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/sshive/sshive/internal/model"
)

// Session 管理一个 SFTP WebSocket 会话
type Session struct {
	host      *model.Host
	sshUser   string
	sshSecret string
	ws        *websocket.Conn
	sessionID int64
}

func NewSession(host *model.Host, sshUser, sshSecret string,
	ws *websocket.Conn, sessionID int64) *Session {
	return &Session{
		host:      host,
		sshUser:   sshUser,
		sshSecret: sshSecret,
		ws:        ws,
		sessionID: sessionID,
	}
}

// Run 建立 SFTP 连接，处理命令消息循环
func (s *Session) Run() error {
	sshCfg, err := s.buildSSHConfig()
	if err != nil {
		return fmt.Errorf("build ssh config: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", s.host.IP, s.host.Port)
	sshConn, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	defer sshConn.Close()

	sftpClient, err := pkgsftp.NewClient(sshConn)
	if err != nil {
		return fmt.Errorf("sftp client: %w", err)
	}
	defer sftpClient.Close()

	for {
		msgType, msg, err := s.ws.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.BinaryMessage {
			slog.Debug("sftp binary message received", "bytes", len(msg))
			continue
		}

		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			s.sendError(req.ReqID, "invalid request: "+err.Error())
			continue
		}

		s.handleRequest(sftpClient, req)
	}
	return nil
}

func (s *Session) handleRequest(client *pkgsftp.Client, req Request) {
	switch req.Type {
	case MsgList:
		files, err := listDir(client, req.Path)
		if err != nil {
			s.sendError(req.ReqID, err.Error())
			return
		}
		s.sendJSON(Response{Type: MsgResult, ReqID: req.ReqID, Success: true, Files: files})

	case MsgDownload:
		data, err := downloadFile(client, req.Path)
		if err != nil {
			s.sendError(req.ReqID, err.Error())
			return
		}
		stat, _ := client.Stat(req.Path)
		name := req.Path
		if stat != nil {
			name = stat.Name()
		}
		s.sendJSON(Response{Type: MsgFileData, ReqID: req.ReqID, Success: true,
			Name: name, Size: int64(len(data))})
		_ = s.ws.WriteMessage(websocket.BinaryMessage, data)

	case MsgUpload:
		_, binaryMsg, err := s.ws.ReadMessage()
		if err != nil {
			s.sendError(req.ReqID, "read file data failed")
			return
		}
		if err := uploadFile(client, req.Path, bytes.NewReader(binaryMsg)); err != nil {
			s.sendError(req.ReqID, err.Error())
			return
		}
		s.sendJSON(Response{Type: MsgResult, ReqID: req.ReqID, Success: true})

	case MsgDelete:
		if err := deleteFile(client, req.Path); err != nil {
			s.sendError(req.ReqID, err.Error())
			return
		}
		s.sendJSON(Response{Type: MsgResult, ReqID: req.ReqID, Success: true})

	case MsgRename:
		if err := renameFile(client, req.Path, req.NewPath); err != nil {
			s.sendError(req.ReqID, err.Error())
			return
		}
		s.sendJSON(Response{Type: MsgResult, ReqID: req.ReqID, Success: true})

	case MsgMkdir:
		if err := mkdirAll(client, req.Path); err != nil {
			s.sendError(req.ReqID, err.Error())
			return
		}
		s.sendJSON(Response{Type: MsgResult, ReqID: req.ReqID, Success: true})

	default:
		s.sendError(req.ReqID, "unknown command: "+string(req.Type))
	}
}

func (s *Session) sendJSON(v any) {
	data, _ := json.Marshal(v)
	_ = s.ws.WriteMessage(websocket.TextMessage, data)
}

func (s *Session) sendError(reqID, msg string) {
	s.sendJSON(Response{Type: MsgError, ReqID: reqID, Success: false, Message: msg})
}

func (s *Session) buildSSHConfig() (*ssh.ClientConfig, error) {
	var authMethod ssh.AuthMethod
	if s.host.AuthType == "key" {
		signer, err := ssh.ParsePrivateKey([]byte(s.sshSecret))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(s.sshSecret)
	}
	return &ssh.ClientConfig{
		User:            s.sshUser,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}, nil
}
