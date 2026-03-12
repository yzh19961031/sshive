package model

import "time"

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"not null;index" json:"tenant_id"`
	Username     string    `gorm:"size:50;not null;uniqueIndex:uk_tenant_username" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Email        string    `gorm:"size:100" json:"email"`
	Status       int8      `gorm:"default:1;not null" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Role struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    int64  `gorm:"not null" json:"tenant_id"`
	Name        string `gorm:"size:50;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}

type UserRole struct {
	UserID int64 `gorm:"primaryKey" json:"user_id"`
	RoleID int64 `gorm:"primaryKey" json:"role_id"`
}
