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
		// Plan 2 new models
		&model.HostGroup{},
		&model.Host{},
		&model.Credential{},
		&model.DangerRule{},
		&model.DangerEvent{},
		// Plan 3 new models
		&model.Session{},
		&model.SessionLog{},
		&model.SessionCommand{},
		// Plan 10 new models
		&model.DBServer{},
		// Plan 11 new models
		&model.DBQueryLog{},
	)
}
