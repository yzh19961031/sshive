package db

import (
	"github.com/sshive/sshive/internal/model"
)

func Migrate() error {
	return DB.AutoMigrate(
		&model.Tenant{},
		&model.User{},
		&model.Role{},
		&model.UserRole{},
		&model.Permission{},
		&model.RolePermission{},
	)
}
