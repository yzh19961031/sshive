package user

import (
	"gorm.io/gorm"

	"github.com/sshive/sshive/internal/db"
	"github.com/sshive/sshive/internal/model"
)

type Repo struct{}

func (r *Repo) Create(u *model.User) error {
	return db.DB.Create(u).Error
}

func (r *Repo) List(tenantID int64, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	offset := (page - 1) * pageSize
	q := db.DB.Model(&model.User{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *Repo) GetByID(tenantID, userID int64) (*model.User, error) {
	var u model.User
	if err := db.DB.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) SetRoles(userID int64, roleIDs []int64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repo) Update(tenantID, userID int64, fields map[string]any) error {
	return db.DB.Model(&model.User{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Updates(fields).Error
}

func (r *Repo) ListRoles(tenantID int64) ([]model.Role, error) {
	var roles []model.Role
	err := db.DB.Where("tenant_id = ?", tenantID).Find(&roles).Error
	return roles, err
}
