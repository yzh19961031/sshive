package model

import "time"

type HostGroup struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  int64     `gorm:"not null;index" json:"tenant_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
