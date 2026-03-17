package model

import "time"

// DBQueryLog 记录每次 DB 客户端查询
type DBQueryLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"not null;index"           json:"tenant_id"`
	DBServerID   int64     `gorm:"not null;index"           json:"db_server_id"`
	UserID       int64     `gorm:"not null;index"           json:"user_id"`
	Database     string    `gorm:"size:100"                 json:"database"`
	SQL          string    `gorm:"type:text;not null"       json:"sql"`
	DurationMs   int64     `                                json:"duration_ms"`
	RowsReturned int       `                                json:"rows_returned"`
	ErrorMsg     string    `gorm:"type:text"                json:"error_msg"` // 空表示成功
	CreatedAt    time.Time `                                json:"created_at"`
}
