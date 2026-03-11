package model

import "time"

type Tenant struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Status    int8      `gorm:"default:1;not null" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
