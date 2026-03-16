package model

import "time"

type Session struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     int64      `gorm:"not null;index" json:"tenant_id"`
	UserID       int64      `gorm:"not null;index" json:"user_id"`
	HostID       int64      `gorm:"not null;index" json:"host_id"`
	StartedAt    time.Time  `gorm:"not null" json:"started_at"`
	EndedAt      *time.Time `json:"ended_at"`
	ClientIP     string     `gorm:"size:45" json:"client_ip"`
	Status       string     `gorm:"size:20;default:active;not null" json:"status"` // active / closed
	CastFilePath string     `gorm:"size:500" json:"cast_file_path"`
}

type SessionLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID int64     `gorm:"not null;index" json:"session_id"`
	Type      string    `gorm:"size:10;not null" json:"type"` // input / output
	Content   string    `gorm:"type:mediumtext;not null" json:"content"`
	CreatedAt time.Time `gorm:"precision:3;default:CURRENT_TIMESTAMP(3)" json:"created_at"`
}

type SessionCommand struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID int64     `gorm:"not null;index" json:"session_id"`
	Command   string    `gorm:"type:text;not null" json:"command"`
	Action    string    `gorm:"size:10;not null;default:execute" json:"action"` // execute / blocked
	Result    string    `gorm:"type:mediumtext" json:"result"`
	CreatedAt time.Time `gorm:"precision:3;default:CURRENT_TIMESTAMP(3)" json:"created_at"`
}
