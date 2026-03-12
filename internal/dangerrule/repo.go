package dangerrule

import (
	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(rule *model.DangerRule) error {
	return db.DB.Create(rule).Error
}

func (r *Repo) Update(rule *model.DangerRule) error {
	return db.DB.Save(rule).Error
}

func (r *Repo) Delete(tenantID, id int64) error {
	return db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DangerRule{}).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.DangerRule, int64, error) {
	var list []model.DangerRule
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.DangerRule{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListActive 返回租户规则 + 全局规则（tenant_id=0），供匹配器使用
func (r *Repo) ListActive(tenantID int64) ([]model.DangerRule, error) {
	var list []model.DangerRule
	err := db.DB.Where("(tenant_id = ? OR tenant_id = 0) AND enabled = 1", tenantID).Find(&list).Error
	return list, err
}

func (r *Repo) GetByID(tenantID, id int64) (*model.DangerRule, error) {
	var rule model.DangerRule
	if err := db.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}
