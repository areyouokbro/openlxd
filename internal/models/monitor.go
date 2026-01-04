package models

import "time"

// SystemMetric 系统监控指标
type SystemMetric struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	MemoryTotal   int64     `json:"memory_total"`
	MemoryUsed    int64     `json:"memory_used"`
	DiskUsage     float64   `json:"disk_usage"`
	DiskTotal     int64     `json:"disk_total"`
	DiskUsed      int64     `json:"disk_used"`
	NetworkRxRate float64   `json:"network_rx_rate"`
	NetworkTxRate float64   `json:"network_tx_rate"`
	LoadAverage1  float64   `json:"load_average_1"`
	LoadAverage5  float64   `json:"load_average_5"`
	LoadAverage15 float64   `json:"load_average_15"`
	CreatedAt     time.Time `json:"created_at"`
}

// ContainerMetric 容器监控指标
type ContainerMetric struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ContainerID    uint      `json:"container_id"`
	ContainerName  string    `json:"container_name"`
	Timestamp      time.Time `json:"timestamp"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	MemoryTotal    int64     `json:"memory_total"`
	MemoryUsed     int64     `json:"memory_used"`
	DiskUsage      float64   `json:"disk_usage"`
	DiskTotal      int64     `json:"disk_total"`
	DiskUsed       int64     `json:"disk_used"`
	NetworkRxRate  float64   `json:"network_rx_rate"`
	NetworkTxRate  float64   `json:"network_tx_rate"`
	NetworkRxTotal int64     `json:"network_rx_total"`
	NetworkTxTotal int64     `json:"network_tx_total"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// NetworkTraffic 网络流量统计
type NetworkTraffic struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ContainerID   uint      `json:"container_id"`
	ContainerName string    `json:"container_name"`
	Timestamp     time.Time `json:"timestamp"`
	RxBytes       int64     `json:"rx_bytes"`
	TxBytes       int64     `json:"tx_bytes"`
	RxPackets     int64     `json:"rx_packets"`
	TxPackets     int64     `json:"tx_packets"`
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 指定表名
func (SystemMetric) TableName() string {
	return "system_metrics"
}

func (ContainerMetric) TableName() string {
	return "container_metrics"
}

func (NetworkTraffic) TableName() string {
	return "network_traffic"
}
