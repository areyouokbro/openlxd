package models

import (
	"time"
)

// Image 镜像模型
type Image struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Alias        string    `gorm:"uniqueIndex;size:100;not null" json:"alias"` // 如 ubuntu/22.04
	Fingerprint  string    `gorm:"size:64" json:"fingerprint"`
	Distribution string    `gorm:"size:50" json:"distribution"` // ubuntu, debian, centos等
	Release      string    `gorm:"size:50" json:"release"`      // 22.04, 11, 9等
	Architecture string    `gorm:"size:20;default:amd64" json:"architecture"` // amd64, arm64
	Variant      string    `gorm:"size:50;default:default" json:"variant"` // default, cloud
	Description  string    `gorm:"type:text" json:"description"`
	Size         int64     `json:"size"` // 字节
	Status       string    `gorm:"size:20;default:available" json:"status"` // available, downloading, imported, failed
	ImportedAt   *time.Time `json:"imported_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Image) TableName() string {
	return "images"
}

// IsImported 检查镜像是否已导入
func (i *Image) IsImported() bool {
	return i.Status == "imported"
}

// IsAvailable 检查镜像是否可用
func (i *Image) IsAvailable() bool {
	return i.Status == "available" || i.Status == "imported"
}
