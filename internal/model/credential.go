package model

import "time"

type Credential struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        int64     `gorm:"not null;index" json:"tenant_id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	Type            string    `gorm:"size:20;not null" json:"type"` // password / key
	Username        string    `gorm:"size:50;not null" json:"username"`
	EncryptedSecret string    `gorm:"type:text;not null" json:"-"` // Prevent serialization
	CreatedAt       time.Time `json:"created_at"`
}
