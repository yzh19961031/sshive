package hostgroup

import (
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(g *model.HostGroup) error {
	return db.DB.Create(g).Error
}

func (r *Repo) Update(g *model.HostGroup) error {
	return db.DB.Save(g).Error
}

func (r *Repo) Delete(tenantID, id int64) error {
	return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.HostGroup{}).Error
}

func (r *Repo) List(tenantID int64) ([]model.HostGroup, error) {
	var list []model.HostGroup
	if err := db.DB.Where("tenant_id = ?", tenantID).Order("id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repo) GetByID(tenantID, id int64) (*model.HostGroup, error) {
	var g model.HostGroup
	if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}
