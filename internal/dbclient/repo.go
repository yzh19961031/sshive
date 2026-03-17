// internal/dbclient/repo.go
package dbclient

import (
	"github.com/sshive/sshive/internal/model"
	"gorm.io/gorm"
)

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) List(tenantID int64) ([]model.DBServer, error) {
	var list []model.DBServer
	return list, r.db.Where("tenant_id = ?", tenantID).Find(&list).Error
}

func (r *Repo) GetByID(tenantID, id int64) (*model.DBServer, error) {
	var s model.DBServer
	return &s, r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error
}

func (r *Repo) Create(s *model.DBServer) error { return r.db.Create(s).Error }
func (r *Repo) Update(s *model.DBServer) error { return r.db.Save(s).Error }
func (r *Repo) Delete(tenantID, id int64) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DBServer{}).Error
}

// CreateQueryLog 写入查询日志
func (r *Repo) CreateQueryLog(log *model.DBQueryLog) error {
	return r.db.Create(log).Error
}

// ListQueryLogs 分页查询日志（仅限本租户）
func (r *Repo) ListQueryLogs(tenantID int64, serverID, userID int64, startTime, endTime string, page, pageSize int) ([]model.DBQueryLog, int64, error) {
	q := r.db.Where("tenant_id = ?", tenantID)
	if serverID > 0 {
		q = q.Where("db_server_id = ?", serverID)
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if startTime != "" {
		q = q.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		q = q.Where("created_at <= ?", endTime)
	}
	var total int64
	q.Model(&model.DBQueryLog{}).Count(&total)
	var list []model.DBQueryLog
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
