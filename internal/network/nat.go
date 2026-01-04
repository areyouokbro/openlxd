package network

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openlxd/backend/internal/models"
	"github.com/openlxd/backend/internal/quota"
)

// NATManager NAT 端口映射管理器
type NATManager struct {
	mu sync.RWMutex
}

var GlobalNATManager = &NATManager{}

// PortMapping 端口映射信息
type PortMapping struct {
	ID            uint   `json:"id"`
	ContainerID   uint   `json:"container_id"`
	ContainerIP   string `json:"container_ip"`
	Protocol      string `json:"protocol"` // tcp, udp
	ExternalPort  int    `json:"external_port"`
	InternalPort  int    `json:"internal_port"`
	Description   string `json:"description"`
	Status        string `json:"status"` // active, inactive
}

// AddPortMapping 添加端口映射
func (n *NATManager) AddPortMapping(containerID uint, containerIP, protocol string, externalPort, internalPort int, description string) (*PortMapping, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 检查配额
	err := quota.GlobalQuotaManager.CheckPortMappingQuota(containerID, 1)
	if err != nil {
		return nil, err
	}

	// 检查端口是否已被使用
	var existing models.PortMapping
	err = models.DB.Where("external_port = ? AND protocol = ?", externalPort, protocol).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("外部端口 %d 已被使用", externalPort)
	}

	// 创建 iptables 规则
	err = n.createIPTablesRule(containerIP, protocol, externalPort, internalPort)
	if err != nil {
		return nil, fmt.Errorf("创建 iptables 规则失败: %v", err)
	}

	// 保存到数据库
	mapping := models.PortMapping{
		ContainerID:  containerID,
		ContainerIP:  containerIP,
		Protocol:     protocol,
		ExternalPort: externalPort,
		InternalPort: internalPort,
		Description:  description,
		Status:       "active",
	}
	models.DB.Create(&mapping)

	return &PortMapping{
		ID:           mapping.ID,
		ContainerID:  mapping.ContainerID,
		ContainerIP:  mapping.ContainerIP,
		Protocol:     mapping.Protocol,
		ExternalPort: mapping.ExternalPort,
		InternalPort: mapping.InternalPort,
		Description:  mapping.Description,
		Status:       mapping.Status,
	}, nil
}

// AddPortRange 添加端口段映射
func (n *NATManager) AddPortRange(containerID uint, containerIP, protocol string, externalStartPort, internalStartPort, count int, description string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 检查配额
	err := quota.GlobalQuotaManager.CheckPortMappingQuota(containerID, count)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		externalPort := externalStartPort + i
		internalPort := internalStartPort + i

		// 检查端口是否已被使用
		var existing models.PortMapping
		err := models.DB.Where("external_port = ? AND protocol = ?", externalPort, protocol).First(&existing).Error
		if err == nil {
			continue // 跳过已使用的端口
		}

		// 创建 iptables 规则
		err = n.createIPTablesRule(containerIP, protocol, externalPort, internalPort)
		if err != nil {
			continue // 跳过失败的规则
		}

		// 保存到数据库
		mapping := models.PortMapping{
			ContainerID:  containerID,
			ContainerIP:  containerIP,
			Protocol:     protocol,
			ExternalPort: externalPort,
			InternalPort: internalPort,
			Description:  fmt.Sprintf("%s (端口段 %d-%d)", description, externalStartPort, externalStartPort+count-1),
			Status:       "active",
		}
		models.DB.Create(&mapping)
	}

	return nil
}

// AddRandomPort 添加随机端口映射
func (n *NATManager) AddRandomPort(containerID uint, containerIP, protocol string, internalPort int, description string) (*PortMapping, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 生成随机端口（10000-65535）
	rand.Seed(time.Now().UnixNano())
	maxAttempts := 100
	
	for i := 0; i < maxAttempts; i++ {
		externalPort := rand.Intn(55535) + 10000

		// 检查端口是否已被使用
		var existing models.PortMapping
		err := models.DB.Where("external_port = ? AND protocol = ?", externalPort, protocol).First(&existing).Error
		if err == nil {
			continue // 端口已被使用，重试
		}

		// 创建 iptables 规则
		err = n.createIPTablesRule(containerIP, protocol, externalPort, internalPort)
		if err != nil {
			continue // 创建失败，重试
		}

		// 保存到数据库
		mapping := models.PortMapping{
			ContainerID:  containerID,
			ContainerIP:  containerIP,
			Protocol:     protocol,
			ExternalPort: externalPort,
			InternalPort: internalPort,
			Description:  description,
			Status:       "active",
		}
		models.DB.Create(&mapping)

		return &PortMapping{
			ID:           mapping.ID,
			ContainerID:  mapping.ContainerID,
			ContainerIP:  mapping.ContainerIP,
			Protocol:     mapping.Protocol,
			ExternalPort: mapping.ExternalPort,
			InternalPort: mapping.InternalPort,
			Description:  mapping.Description,
			Status:       mapping.Status,
		}, nil
	}

	return nil, fmt.Errorf("无法找到可用的随机端口")
}

// RemovePortMapping 删除端口映射
func (n *NATManager) RemovePortMapping(mappingID uint) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var mapping models.PortMapping
	err := models.DB.First(&mapping, mappingID).Error
	if err != nil {
		return fmt.Errorf("端口映射不存在")
	}

	// 删除 iptables 规则
	err = n.deleteIPTablesRule(mapping.ContainerIP, mapping.Protocol, mapping.ExternalPort, mapping.InternalPort)
	if err != nil {
		return fmt.Errorf("删除 iptables 规则失败: %v", err)
	}

	// 从数据库删除
	models.DB.Delete(&mapping)

	return nil
}

// RemoveContainerMappings 删除容器的所有端口映射
func (n *NATManager) RemoveContainerMappings(containerID uint) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var mappings []models.PortMapping
	models.DB.Where("container_id = ?", containerID).Find(&mappings)

	for _, mapping := range mappings {
		// 删除 iptables 规则
		n.deleteIPTablesRule(mapping.ContainerIP, mapping.Protocol, mapping.ExternalPort, mapping.InternalPort)
		
		// 从数据库删除
		models.DB.Delete(&mapping)
	}

	return nil
}

// GetContainerMappings 获取容器的所有端口映射
func (n *NATManager) GetContainerMappings(containerID uint) ([]PortMapping, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var mappings []models.PortMapping
	err := models.DB.Where("container_id = ?", containerID).Find(&mappings).Error
	if err != nil {
		return nil, err
	}

	result := make([]PortMapping, len(mappings))
	for i, m := range mappings {
		result[i] = PortMapping{
			ID:           m.ID,
			ContainerID:  m.ContainerID,
			ContainerIP:  m.ContainerIP,
			Protocol:     m.Protocol,
			ExternalPort: m.ExternalPort,
			InternalPort: m.InternalPort,
			Description:  m.Description,
			Status:       m.Status,
		}
	}

	return result, nil
}

// SyncIPTablesRules 同步 iptables 规则
func (n *NATManager) SyncIPTablesRules() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 清除现有的 OpenLXD 规则
	n.clearOpenLXDRules()

	// 从数据库重新创建所有规则
	var mappings []models.PortMapping
	models.DB.Where("status = ?", "active").Find(&mappings)

	for _, mapping := range mappings {
		n.createIPTablesRule(mapping.ContainerIP, mapping.Protocol, mapping.ExternalPort, mapping.InternalPort)
	}

	return nil
}

// createIPTablesRule 创建 iptables 规则
func (n *NATManager) createIPTablesRule(containerIP, protocol string, externalPort, internalPort int) error {
	// DNAT 规则：外部访问 -> 容器
	dnatCmd := fmt.Sprintf(
		"iptables -t nat -A PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d -m comment --comment 'OpenLXD'",
		protocol, externalPort, containerIP, internalPort,
	)
	
	// FORWARD 规则：允许转发
	forwardCmd := fmt.Sprintf(
		"iptables -A FORWARD -p %s -d %s --dport %d -j ACCEPT -m comment --comment 'OpenLXD'",
		protocol, containerIP, internalPort,
	)

	// MASQUERADE 规则：容器访问外部
	masqCmd := fmt.Sprintf(
		"iptables -t nat -A POSTROUTING -s %s -j MASQUERADE -m comment --comment 'OpenLXD'",
		containerIP,
	)

	// 执行命令
	commands := []string{dnatCmd, forwardCmd, masqCmd}
	for _, cmd := range commands {
		parts := strings.Fields(cmd)
		if err := exec.Command(parts[0], parts[1:]...).Run(); err != nil {
			return err
		}
	}

	return nil
}

// deleteIPTablesRule 删除 iptables 规则
func (n *NATManager) deleteIPTablesRule(containerIP, protocol string, externalPort, internalPort int) error {
	// 删除 DNAT 规则
	dnatCmd := fmt.Sprintf(
		"iptables -t nat -D PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d -m comment --comment 'OpenLXD'",
		protocol, externalPort, containerIP, internalPort,
	)
	
	// 删除 FORWARD 规则
	forwardCmd := fmt.Sprintf(
		"iptables -D FORWARD -p %s -d %s --dport %d -j ACCEPT -m comment --comment 'OpenLXD'",
		protocol, containerIP, internalPort,
	)

	// 执行命令（忽略错误）
	commands := []string{dnatCmd, forwardCmd}
	for _, cmd := range commands {
		parts := strings.Fields(cmd)
		exec.Command(parts[0], parts[1:]...).Run()
	}

	return nil
}

// clearOpenLXDRules 清除所有 OpenLXD 的 iptables 规则
func (n *NATManager) clearOpenLXDRules() {
	// 清除 NAT 表中的 OpenLXD 规则
	exec.Command("bash", "-c", "iptables -t nat -S | grep 'OpenLXD' | sed 's/-A/-D/' | xargs -r -L1 iptables -t nat").Run()
	
	// 清除 FILTER 表中的 OpenLXD 规则
	exec.Command("bash", "-c", "iptables -S | grep 'OpenLXD' | sed 's/-A/-D/' | xargs -r -L1 iptables").Run()
}

// IsPortAvailable 检查端口是否可用
func (n *NATManager) IsPortAvailable(port int, protocol string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var count int64
	models.DB.Model(&models.PortMapping{}).
		Where("external_port = ? AND protocol = ?", port, protocol).
		Count(&count)
	
	return count == 0
}

// GetUsedPortsCount 获取已使用的端口数量
func (n *NATManager) GetUsedPortsCount() int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var count int64
	models.DB.Model(&models.PortMapping{}).
		Where("status = ?", "active").
		Count(&count)
	
	return count
}

// ParsePortRange 解析端口范围字符串
func ParsePortRange(portRange string) (start, end int, err error) {
	parts := strings.Split(portRange, "-")
	if len(parts) == 1 {
		// 单个端口
		port, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("无效的端口号")
		}
		return port, port, nil
	} else if len(parts) == 2 {
		// 端口范围
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("无效的起始端口")
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("无效的结束端口")
		}
		if start > end {
			return 0, 0, fmt.Errorf("起始端口不能大于结束端口")
		}
		return start, end, nil
	}
	
	return 0, 0, fmt.Errorf("无效的端口范围格式")
}
