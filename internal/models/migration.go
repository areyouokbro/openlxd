package models

import (
	"time"
)

// MigrationTask 迁移任务
type MigrationTask struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ContainerName   string    `gorm:"size:255;not null" json:"container_name"`
	SourceHost      string    `gorm:"size:255;not null" json:"source_host"`
	TargetHost      string    `gorm:"size:255;not null" json:"target_host"`
	MigrationType   string    `gorm:"size:50;not null" json:"migration_type"` // live, cold
	Status          string    `gorm:"size:50;not null" json:"status"`          // pending, running, completed, failed, rollback
	Progress        int       `gorm:"default:0" json:"progress"`               // 0-100
	ErrorMessage    string    `gorm:"type:text" json:"error_message,omitempty"`
	StartTime       time.Time `json:"start_time,omitempty"`
	EndTime         time.Time `json:"end_time,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RemoteHost 远程主机配置
type RemoteHost struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Address     string    `gorm:"size:255;not null" json:"address"`     // IP 或域名
	Port        int       `gorm:"default:8443" json:"port"`             // LXD API 端口
	Protocol    string    `gorm:"size:10;default:https" json:"protocol"` // https
	Certificate string    `gorm:"type:text" json:"certificate,omitempty"` // 客户端证书
	Key         string    `gorm:"type:text" json:"key,omitempty"`        // 客户端密钥
	Status      string    `gorm:"size:50;default:active" json:"status"`  // active, inactive
	Description string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MigrationLog 迁移日志
type MigrationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"not null;index" json:"task_id"`
	Level     string    `gorm:"size:20;not null" json:"level"` // info, warning, error
	Message   string    `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (MigrationTask) TableName() string {
	return "migration_tasks"
}

func (RemoteHost) TableName() string {
	return "remote_hosts"
}

func (MigrationLog) TableName() string {
	return "migration_logs"
}
