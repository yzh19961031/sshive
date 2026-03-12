package model

type Permission struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string `gorm:"size:100;not null;uniqueIndex" json:"code"`
	Description string `gorm:"size:255" json:"description"`
}

type RolePermission struct {
	RoleID       int64 `gorm:"primaryKey" json:"role_id"`
	PermissionID int64 `gorm:"primaryKey" json:"permission_id"`
}
