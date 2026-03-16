package audit

import (
	"time"

	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) CreateSession(s *model.Session) error {
	return db.DB.Create(s).Error
}

func (r *Repo) CloseSession(id int64, castPath string) error {
	now := time.Now()
	return db.DB.Model(&model.Session{}).Where("id = ?", id).Updates(map[string]any{
		"ended_at":       now,
		"status":         "closed",
		"cast_file_path": castPath,
	}).Error
}

func (r *Repo) WriteLog(log *model.SessionLog) error {
	return db.DB.Create(log).Error
}

func (r *Repo) WriteCommand(cmd *model.SessionCommand) error {
	return db.DB.Create(cmd).Error
}

func (r *Repo) ListSessions(tenantID int64, page, pageSize int) ([]SessionListItem, int64, error) {
	var total int64
	offset := (page - 1) * pageSize
	if err := db.DB.Model(&model.Session{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SessionListItem
	err := db.DB.Raw(`
		SELECT s.id, s.tenant_id, s.user_id, s.host_id,
		       s.started_at, s.ended_at, s.client_ip, s.status, s.cast_file_path,
		       COALESCE(h.name, '') AS host_name,
		       COALESCE(u.username, '') AS username
		FROM sessions s
		LEFT JOIN hosts h ON h.id = s.host_id AND h.tenant_id = s.tenant_id
		LEFT JOIN users u ON u.id = s.user_id AND u.tenant_id = s.tenant_id
		WHERE s.tenant_id = ?
		ORDER BY s.started_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, pageSize, offset).Scan(&list).Error
	return list, total, err
}

func (r *Repo) GetSession(tenantID, id int64) (*model.Session, error) {
	var s model.Session
	if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repo) ListLogs(sessionID int64, page, pageSize int) ([]model.SessionLog, int64, error) {
	var list []model.SessionLog
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.SessionLog{}).Where("session_id = ?", sessionID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repo) ListCommands(sessionID int64, page, pageSize int, f CommandFilter) ([]model.SessionCommand, int64, error) {
	var list []model.SessionCommand
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.SessionCommand{}).Where("session_id = ?", sessionID)
	if f.Command != "" {
		q = q.Where("command LIKE ?", "%"+f.Command+"%")
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", f.EndTime)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repo) ListAllCommands(tenantID int64, page, pageSize int) ([]CommandListItem, int64, error) {
	var total int64
	offset := (page - 1) * pageSize
	if err := db.DB.Model(&model.SessionCommand{}).
		Joins("JOIN sessions ON sessions.id = session_commands.session_id AND sessions.tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []CommandListItem
	err := db.DB.Raw(`
		SELECT sc.id, sc.session_id, sc.command, sc.action, sc.result, sc.created_at,
		       COALESCE(h.name, '') AS host_name,
		       COALESCE(u.username, '') AS username
		FROM session_commands sc
		JOIN sessions s ON s.id = sc.session_id AND s.tenant_id = ?
		LEFT JOIN hosts h ON h.id = s.host_id AND h.tenant_id = s.tenant_id
		LEFT JOIN users u ON u.id = s.user_id AND u.tenant_id = s.tenant_id
		ORDER BY sc.created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, pageSize, offset).Scan(&list).Error
	return list, total, err
}

func (r *Repo) UpdateCastPath(id int64, castPath string) error {
	return db.DB.Model(&model.Session{}).Where("id = ?", id).
		Update("cast_file_path", castPath).Error
}

func (r *Repo) RecoverStaleSessions() error {
	now := time.Now()
	return db.DB.Model(&model.Session{}).
		Where("status = ? AND ended_at IS NULL", "active").
		Updates(map[string]any{
			"status":   "interrupted",
			"ended_at": now,
		}).Error
}
