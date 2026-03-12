package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sshive/sshive/internal/config"
	"github.com/sshive/sshive/internal/model"
	"github.com/sshive/sshive/pkg/cast"
)

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

	castWriter, err := cast.NewWriter(castPath, width, height)
	if err != nil {
		return nil, fmt.Errorf("create cast writer: %w", err)
	}

	asyncW := NewAsyncWriter(sess.ID, castWriter)
	return &StartSessionResult{Session: sess, AsyncWriter: asyncW}, nil
}

func (s *Service) CloseSession(sessID int64, asyncW *AsyncWriter, castPath string) {
	asyncW.Close()
	_ = s.repo.CloseSession(sessID, castPath)
}

func (s *Service) ListSessions(tenantID int64, page, pageSize int) ([]model.Session, int64, error) {
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
