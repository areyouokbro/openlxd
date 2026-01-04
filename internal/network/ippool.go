package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/openlxd/backend/internal/models"
)

// IPPool IP地址池管理
type IPPool struct {
	mu sync.RWMutex
}

var GlobalIPPool = &IPPool{}

// IPAddress IP地址信息
type IPAddress struct {
	ID          uint   `json:"id"`
	IP          string `json:"ip"`
	Type        string `json:"type"` // ipv4, ipv6
	Status      string `json:"status"` // available, used, reserved
	ContainerID uint   `json:"container_id"`
	Gateway     string `json:"gateway"`
	Netmask     string `json:"netmask"`
}

// AllocateIPv4 分配 IPv4 地址
func (p *IPPool) AllocateIPv4(containerID uint) (*IPAddress, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 从数据库查找可用的 IPv4 地址
	var ipAddr models.IPAddress
	err := models.DB.Where("type = ? AND status = ?", "ipv4", "available").First(&ipAddr).Error
	if err != nil {
		return nil, fmt.Errorf("没有可用的 IPv4 地址")
	}

	// 标记为已使用
	ipAddr.Status = "used"
	ipAddr.ContainerID = containerID
	models.DB.Save(&ipAddr)

	return &IPAddress{
		ID:          ipAddr.ID,
		IP:          ipAddr.IP,
		Type:        ipAddr.Type,
		Status:      ipAddr.Status,
		ContainerID: ipAddr.ContainerID,
		Gateway:     ipAddr.Gateway,
		Netmask:     ipAddr.Netmask,
	}, nil
}

// AllocateIPv6 分配 IPv6 地址
func (p *IPPool) AllocateIPv6(containerID uint) (*IPAddress, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 从数据库查找可用的 IPv6 地址
	var ipAddr models.IPAddress
	err := models.DB.Where("type = ? AND status = ?", "ipv6", "available").First(&ipAddr).Error
	if err != nil {
		return nil, fmt.Errorf("没有可用的 IPv6 地址")
	}

	// 标记为已使用
	ipAddr.Status = "used"
	ipAddr.ContainerID = containerID
	models.DB.Save(&ipAddr)

	return &IPAddress{
		ID:          ipAddr.ID,
		IP:          ipAddr.IP,
		Type:        ipAddr.Type,
		Status:      ipAddr.Status,
		ContainerID: ipAddr.ContainerID,
		Gateway:     ipAddr.Gateway,
		Netmask:     ipAddr.Netmask,
	}, nil
}

// ReleaseIP 释放 IP 地址
func (p *IPPool) ReleaseIP(ipID uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var ipAddr models.IPAddress
	err := models.DB.First(&ipAddr, ipID).Error
	if err != nil {
		return fmt.Errorf("IP 地址不存在")
	}

	// 标记为可用
	ipAddr.Status = "available"
	ipAddr.ContainerID = 0
	models.DB.Save(&ipAddr)

	return nil
}

// ReleaseContainerIPs 释放容器的所有 IP 地址
func (p *IPPool) ReleaseContainerIPs(containerID uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 释放所有该容器的 IP 地址
	models.DB.Model(&models.IPAddress{}).
		Where("container_id = ?", containerID).
		Updates(map[string]interface{}{
			"status":       "available",
			"container_id": 0,
		})

	return nil
}

// GetAvailableIPv4Count 获取可用 IPv4 地址数量
func (p *IPPool) GetAvailableIPv4Count() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var count int64
	models.DB.Model(&models.IPAddress{}).
		Where("type = ? AND status = ?", "ipv4", "available").
		Count(&count)
	return count
}

// GetAvailableIPv6Count 获取可用 IPv6 地址数量
func (p *IPPool) GetAvailableIPv6Count() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var count int64
	models.DB.Model(&models.IPAddress{}).
		Where("type = ? AND status = ?", "ipv6", "available").
		Count(&count)
	return count
}

// AddIPRange 添加 IP 地址段
func (p *IPPool) AddIPRange(startIP, endIP, gateway, netmask, ipType string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)

	if start == nil || end == nil {
		return fmt.Errorf("无效的 IP 地址")
	}

	// 生成 IP 地址列表
	for ip := start; !ip.Equal(end); ip = nextIP(ip) {
		ipAddr := models.IPAddress{
			IP:      ip.String(),
			Type:    ipType,
			Status:  "available",
			Gateway: gateway,
			Netmask: netmask,
		}
		models.DB.Create(&ipAddr)
	}

	// 添加结束 IP
	ipAddr := models.IPAddress{
		IP:      endIP,
		Type:    ipType,
		Status:  "available",
		Gateway: gateway,
		Netmask: netmask,
	}
	models.DB.Create(&ipAddr)

	return nil
}

// RemoveIPRange 删除 IP 地址段
func (p *IPPool) RemoveIPRange(startIP, endIP string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 删除指定范围内的 IP 地址（仅删除未使用的）
	models.DB.Where("ip >= ? AND ip <= ? AND status = ?", startIP, endIP, "available").
		Delete(&models.IPAddress{})

	return nil
}

// GetContainerIPs 获取容器的所有 IP 地址
func (p *IPPool) GetContainerIPs(containerID uint) ([]IPAddress, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var ipAddrs []models.IPAddress
	err := models.DB.Where("container_id = ?", containerID).Find(&ipAddrs).Error
	if err != nil {
		return nil, err
	}

	result := make([]IPAddress, len(ipAddrs))
	for i, ip := range ipAddrs {
		result[i] = IPAddress{
			ID:          ip.ID,
			IP:          ip.IP,
			Type:        ip.Type,
			Status:      ip.Status,
			ContainerID: ip.ContainerID,
			Gateway:     ip.Gateway,
			Netmask:     ip.Netmask,
		}
	}

	return result, nil
}

// nextIP 获取下一个 IP 地址
func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}
	return next
}

// ValidateIP 验证 IP 地址格式
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsIPv4 判断是否为 IPv4 地址
func IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// IsIPv6 判断是否为 IPv6 地址
func IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}
