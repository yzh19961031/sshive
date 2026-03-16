package audit

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sshive/sshive/internal/config"
	"github.com/sshive/sshive/internal/model"
	"github.com/sshive/sshive/pkg/cast"
)

// SessionListItem 是会话列表的 API 响应类型，包含 host_name 和 username 避免前端额外请求。
type SessionListItem struct {
	ID           int64      `json:"id"`
	TenantID     int64      `json:"tenant_id"`
	UserID       int64      `json:"user_id"`
	HostID       int64      `json:"host_id"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at"`
	ClientIP     string     `json:"client_ip"`
	Status       string     `json:"status"`
	CastFilePath string     `json:"cast_file_path"`
	HostName     string     `json:"host_name"`
	Username     string     `json:"username"`
}

type Service struct {
	repo *Repo
}

func NewService() *Service {
	return &Service{repo: &Repo{}}
}

type StartSessionResult struct {
	Session     *model.Session
	AsyncWriter *AsyncWriter
}

func (s *Service) StartSession(tenantID, userID, hostID int64, clientIP string, width, height int) (*StartSessionResult, error) {
	sess := &model.Session{
		TenantID:  tenantID,
		UserID:    userID,
		HostID:    hostID,
		StartedAt: time.Now().UTC(),
		ClientIP:  clientIP,
		Status:    "active",
	}
	if err := s.repo.CreateSession(sess); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	castDir := config.C.CastDir
	if err := os.MkdirAll(castDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir cast dir: %w", err)
	}
	castPath := filepath.Join(castDir, strconv.FormatInt(sess.ID, 10)+".cast")
	sess.CastFilePath = castPath

	castWriter, err := cast.NewWriter(castPath, width, height)
	if err != nil {
		return nil, fmt.Errorf("create cast writer: %w", err)
	}

	// 立即持久化 cast_file_path，避免服务重启时路径丢失
	if err := s.repo.UpdateCastPath(sess.ID, castPath); err != nil {
		slog.Warn("update cast path failed", "err", err, "session_id", sess.ID)
	}
	sess.CastFilePath = castPath

	asyncW := NewAsyncWriter(sess.ID, castWriter)
	return &StartSessionResult{Session: sess, AsyncWriter: asyncW}, nil
}

func (s *Service) CloseSession(sessID int64, asyncW *AsyncWriter, castPath string) {
	asyncW.Close()
	_ = s.repo.CloseSession(sessID, castPath)
}

func (s *Service) ListSessions(tenantID int64, page, pageSize int) ([]SessionListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListSessions(tenantID, page, pageSize)
}

func (s *Service) GetSession(tenantID, id int64) (*model.Session, error) {
	return s.repo.GetSession(tenantID, id)
}

func (s *Service) ListLogs(tenantID, sessionID int64, page, pageSize int) ([]model.SessionLog, int64, error) {
	if _, err := s.repo.GetSession(tenantID, sessionID); err != nil {
		return nil, 0, fmt.Errorf("session not found")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListLogs(sessionID, page, pageSize)
}

func (s *Service) ListCommands(tenantID, sessionID int64, page, pageSize int) ([]model.SessionCommand, int64, error) {
	if _, err := s.repo.GetSession(tenantID, sessionID); err != nil {
		return nil, 0, fmt.Errorf("session not found")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListCommands(sessionID, page, pageSize)
}

// CommandListItem 是全局命令记录列表的 API 响应类型
type CommandListItem struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Command   string    `json:"command"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	CreatedAt time.Time `json:"created_at"`
	HostName  string    `json:"host_name"`
	Username  string    `json:"username"`
}

// ListAllCommands 列出租户下所有会话的命令记录（全局视图）
func (s *Service) ListAllCommands(tenantID int64, page, pageSize int) ([]CommandListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListAllCommands(tenantID, page, pageSize)
}

// RecoverStaleSessions 在服务启动时调用，清理因重启残留的 active 会话
func (s *Service) RecoverStaleSessions() {
	if err := s.repo.RecoverStaleSessions(); err != nil {
		slog.Warn("recover stale sessions failed", "err", err)
		return
	}
	slog.Info("stale sessions recovered")
}
