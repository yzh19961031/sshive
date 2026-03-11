package tenant

import (
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(t *model.Tenant) error {
	return db.DB.Create(t).Error
}

func (r *Repo) List(page, pageSize int) ([]model.Tenant, int64, error) {
	var tenants []model.Tenant
	var total int64
	offset := (page - 1) * pageSize
	if err := db.DB.Model(&model.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.DB.Offset(offset).Limit(pageSize).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}
	return tenants, total, nil
}

func (r *Repo) GetByID(id int64) (*model.Tenant, error) {
	var t model.Tenant
	if err := db.DB.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
