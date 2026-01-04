package models

import (
	"time"
)

// Container 容器模型
type Container struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Hostname     string    `gorm:"uniqueIndex;not null" json:"hostname"`
	Status       string    `gorm:"default:'Stopped'" json:"status"`
	Image        string    `json:"image"`
	IPv4         string    `json:"ipv4"`
	IPv6         string    `json:"ipv6"`
	CPUs         int       `json:"cpus"`
	Memory       int       `json:"memory"`        // MB
	Disk         int       `json:"disk"`          // GB
	Ingress      int       `json:"ingress"`       // Mbps
	Egress       int       `json:"egress"`        // Mbps
	TrafficUsed  int64     `json:"traffic_used"`  // Bytes
	TrafficLimit int64     `json:"traffic_limit"` // Bytes
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ActionLog 操作日志模型
type ActionLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Action      string    `json:"action"`
	Container   string    `json:"container"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// NetworkConfig 网络配置模型
type NetworkConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ContainerID  uint      `json:"container_id"`
	Type         string    `json:"type"` // independent, nat, proxy
	IPv4         string    `json:"ipv4"`
	IPv6         string    `json:"ipv6"`
	PortMappings string    `gorm:"type:text" json:"port_mappings"` // JSON格式
	Domain       string    `json:"domain"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IPAddress IP地址模型
type IPAddress struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	IP          string    `gorm:"uniqueIndex;not null" json:"ip"`
	Type        string    `json:"type"` // ipv4, ipv6
	Status      string    `gorm:"default:'available'" json:"status"` // available, used, reserved
	ContainerID uint      `json:"container_id"`
	Gateway     string    `json:"gateway"`
	Netmask     string    `json:"netmask"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PortMapping 端口映射模型
type PortMapping struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ContainerID  uint      `json:"container_id"`
	ContainerIP  string    `json:"container_ip"`
	Protocol     string    `json:"protocol"` // tcp, udp
	ExternalPort int       `gorm:"index" json:"external_port"`
	InternalPort int       `json:"internal_port"`
	Description  string    `json:"description"`
	Status       string    `gorm:"default:'active'" json:"status"` // active, inactive
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProxyConfig 反向代理配置模型
type ProxyConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ContainerID uint      `json:"container_id"`
	Domain      string    `gorm:"uniqueIndex;not null" json:"domain"`
	TargetIP    string    `json:"target_ip"`
	TargetPort  int       `json:"target_port"`
	SSL         bool      `json:"ssl"`
	CertPath    string    `json:"cert_path"`
	KeyPath     string    `json:"key_path"`
	Status      string    `gorm:"default:'active'" json:"status"` // active, inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Quota 配额模型（为后续配额系统做准备）
type Quota struct {
	ID               uint `gorm:"primaryKey" json:"id"`
	UserID           uint `json:"user_id"`
	IPv4Quota        int  `json:"ipv4_quota"`
	IPv6Quota        int  `json:"ipv6_quota"`
	IPv4MappingQuota int  `json:"ipv4_mapping_quota"`
	IPv6MappingQuota int  `json:"ipv6_mapping_quota"`
	ProxyQuota       int  `json:"proxy_quota"`
}
