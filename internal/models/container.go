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

// OperationLog 操作日志模型（别名）
type OperationLog = ActionLog

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

// Quota 配额模型
type Quota struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ContainerID      uint      `gorm:"uniqueIndex" json:"container_id"`
	// IP地址配额
	IPv4Quota        int       `gorm:"default:-1" json:"ipv4_quota"` // -1 表示无限制
	IPv6Quota        int       `gorm:"default:-1" json:"ipv6_quota"`
	// 端口映射配额
	PortMappingQuota int       `gorm:"default:-1" json:"port_mapping_quota"`
	// 反向代理配额
	ProxyQuota       int       `gorm:"default:-1" json:"proxy_quota"`
	// 流量配额（单位：GB）
	TrafficQuota     int64     `gorm:"default:-1" json:"traffic_quota"` // -1 表示无限制
	TrafficUsed      int64     `gorm:"default:0" json:"traffic_used"`   // 已使用流量
	TrafficResetDate time.Time `json:"traffic_reset_date"`              // 流量重置日期
	// 配额超限处理
	OnExceed         string    `gorm:"default:'warn'" json:"on_exceed"` // warn, limit, stop
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// QuotaUsage 配额使用情况
type QuotaUsage struct {
	ContainerID      uint  `json:"container_id"`
	IPv4Used         int   `json:"ipv4_used"`
	IPv6Used         int   `json:"ipv6_used"`
	PortMappingUsed  int   `json:"port_mapping_used"`
	ProxyUsed        int   `json:"proxy_used"`
	TrafficUsed      int64 `json:"traffic_used"`
	IPv4Quota        int   `json:"ipv4_quota"`
	IPv6Quota        int   `json:"ipv6_quota"`
	PortMappingQuota int   `json:"port_mapping_quota"`
	ProxyQuota       int   `json:"proxy_quota"`
	TrafficQuota     int64 `json:"traffic_quota"`
}
