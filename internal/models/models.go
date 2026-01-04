package models

import (
	"time"
)

// Container 容器模型
type Container struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Hostname  string    `gorm:"uniqueIndex;size:100" json:"hostname"`
	Status    string    `gorm:"size:20" json:"status"`
	Image     string    `gorm:"size:50" json:"image"`
	IPv4      string    `gorm:"size:50" json:"ipv4"`
	IPv6      string    `gorm:"size:100" json:"ipv6"`
	CPUs      int       `json:"cpus"`
	Memory    int       `json:"memory"`
	Disk      int       `json:"disk"`
	Ingress   int       `json:"ingress"`
	Egress    int       `json:"egress"`
	
	TrafficUsed  int64 `json:"traffic_used"`
	TrafficLimit int64 `json:"traffic_limit"`
	
	IPv4PoolLimit      int `json:"ipv4_pool_limit"`
	IPv4MappingLimit   int `json:"ipv4_mapping_limit"`
	IPv6PoolLimit      int `json:"ipv6_pool_limit"`
	IPv6MappingLimit   int `json:"ipv6_mapping_limit"`
	ReverseProxyLimit  int `json:"reverse_proxy_limit"`
	CPUAllowance       int `json:"cpu_allowance"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IPPool IP地址池
type IPPool struct {
	ID          uint      `gorm:"primaryKey"`
	Address     string    `gorm:"uniqueIndex;size:50"`
	Type        string    `gorm:"size:10"` // ipv4, ipv6
	IsUsed      bool      `gorm:"default:false"`
	ContainerID uint      `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PortMapping 端口映射
type PortMapping struct {
	ID          uint      `gorm:"primaryKey"`
	ContainerID uint      `gorm:"index"`
	Protocol    string    `gorm:"size:10"` // tcp, udp
	HostPort    int       
	ContainerIP string    `gorm:"size:50"`
	ContainerPort int     
	CreatedAt   time.Time
}

// AuditLog 审计日志
type AuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	Action    string    `gorm:"size:50"`
	Target    string    `gorm:"size:100"`
	Detail    string    `gorm:"type:text"`
	Status    string    `gorm:"size:20"`
	CreatedAt time.Time
}

// Config 配置项
type Config struct {
	Key   string `gorm:"primaryKey;size:50"`
	Value string `gorm:"type:text"`
}
