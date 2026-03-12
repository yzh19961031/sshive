package host

import (
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(h *model.Host) error {
	return db.DB.Create(h).Error
}

func (r *Repo) Update(h *model.Host) error {
	return db.DB.Save(h).Error
}

func (r *Repo) Delete(tenantID, id int64) error {
	return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Host{}).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.Host, int64, error) {
	var list []model.Host
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.Host{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repo) GetByID(tenantID, id int64) (*model.Host, error) {
	var h model.Host
	if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}
