package model

import "time"

// DBServer 数据库服务器连接配置
type DBServer struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         int64     `gorm:"not null;index" json:"tenant_id"`
	Name             string    `gorm:"size:100;not null" json:"name"`
	Type             string    `gorm:"size:20;not null" json:"type"` // mysql / postgres
	Host             string    `gorm:"size:128;not null" json:"host"`
	Port             int       `gorm:"not null" json:"port"`
	Username         string    `gorm:"size:100;not null" json:"username"`
	EncryptedPassword string   `gorm:"type:text;not null" json:"-"` // Prevent serialization
	Database         string    `gorm:"size:100" json:"database"`   // 默认连接的库（可空）
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
