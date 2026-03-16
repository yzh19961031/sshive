package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Tags is a JSON string slice for storing host tags
type Tags []string

func (t Tags) Value() (driver.Value, error) {
	b, err := json.Marshal(t)
	return string(b), err
}

func (t *Tags) Scan(value any) error {
	if value == nil {
		*t = Tags{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T", value)
	}
	return json.Unmarshal(bytes, t)
}

type Host struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"not null;index" json:"tenant_id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	IP           string    `gorm:"size:45;not null" json:"ip"`
	Port         int       `gorm:"default:22;not null" json:"port"`
	AuthType     string    `gorm:"size:20;not null" json:"auth_type"` // password / key
	CredentialID int64     `gorm:"not null" json:"credential_id"`
	Status       int8      `gorm:"default:1;not null" json:"status"`
	Tags         Tags      `gorm:"type:json" json:"tags"`
	GroupID      *int64    `gorm:"index" json:"group_id"`
	CreatedAt    time.Time `json:"created_at"`
}
