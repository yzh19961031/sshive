package credential

import (
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(c *model.Credential) error {
	return db.DB.Create(c).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.Credential, int64, error) {
	var list []model.Credential
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.Credential{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repo) GetByID(tenantID, id int64) (*model.Credential, error) {
	var c model.Credential
	if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) Delete(tenantID, id int64) error {
	return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Credential{}).Error
}
