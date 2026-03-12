package model

import "time"

type DangerRule struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    int64     `gorm:"not null;index" json:"tenant_id"` // 0 = global rule
	Pattern     string    `gorm:"size:500;not null" json:"pattern"`
	Description string    `gorm:"size:255" json:"description"`
	Action      string    `gorm:"size:20;default:block;not null" json:"action"`
	Enabled     int8      `gorm:"default:1;not null" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type DangerEvent struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID   int64     `gorm:"not null;index" json:"session_id"`
	MatchedRule string    `gorm:"size:255;not null" json:"matched_rule"`
	Command     string    `gorm:"type:text;not null" json:"command"`
	TriggeredAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"triggered_at"`
}
