package lxd

import (
	"fmt"
	"log"
	"os/exec"
	
	"github.com/openlxd/backend/internal/models"
)

// AddPortMapping 添加端口映射
func AddPortMapping(containerName, protocol string, hostPort int, containerIP string, containerPort int) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 添加端口映射: %s:%d -> %s:%d (%s)", 
			"0.0.0.0", hostPort, containerIP, containerPort, protocol)
		return nil
	}
	
	// 添加 DNAT 规则（外部访问 -> 容器）
	dnatCmd := fmt.Sprintf(
		"iptables -t nat -A PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
		protocol, hostPort, containerIP, containerPort,
	)
	
	if err := exec.Command("sh", "-c", dnatCmd).Run(); err != nil {
		return fmt.Errorf("添加 DNAT 规则失败: %w", err)
	}
	
	// 添加 FORWARD 规则（允许转发）
	forwardCmd := fmt.Sprintf(
		"iptables -A FORWARD -p %s -d %s --dport %d -j ACCEPT",
		protocol, containerIP, containerPort,
	)
	
	if err := exec.Command("sh", "-c", forwardCmd).Run(); err != nil {
		return fmt.Errorf("添加 FORWARD 规则失败: %w", err)
	}
	
	// 保存到数据库
	var container models.Container
	models.DB.Where("hostname = ?", containerName).First(&container)
	
	mapping := models.PortMapping{
		ContainerID:   container.ID,
		Protocol:      protocol,
		HostPort:      hostPort,
		ContainerIP:   containerIP,
		ContainerPort: containerPort,
	}
	models.DB.Create(&mapping)
	
	log.Printf("端口映射添加成功: %s:%d -> %s:%d", "0.0.0.0", hostPort, containerIP, containerPort)
	return nil
}

// RemovePortMapping 删除端口映射
func RemovePortMapping(hostPort int, protocol string) error {
	var mapping models.PortMapping
	if err := models.DB.Where("host_port = ? AND protocol = ?", hostPort, protocol).First(&mapping).Error; err != nil {
		return fmt.Errorf("端口映射不存在")
	}
	
	if !IsLXDAvailable() {
		log.Printf("[Mock] 删除端口映射: %d (%s)", hostPort, protocol)
		models.DB.Delete(&mapping)
		return nil
	}
	
	// 删除 DNAT 规则
	dnatCmd := fmt.Sprintf(
		"iptables -t nat -D PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
		protocol, hostPort, mapping.ContainerIP, mapping.ContainerPort,
	)
	exec.Command("sh", "-c", dnatCmd).Run()
	
	// 删除 FORWARD 规则
	forwardCmd := fmt.Sprintf(
		"iptables -D FORWARD -p %s -d %s --dport %d -j ACCEPT",
		protocol, mapping.ContainerIP, mapping.ContainerPort,
	)
	exec.Command("sh", "-c", forwardCmd).Run()
	
	// 从数据库删除
	models.DB.Delete(&mapping)
	
	log.Printf("端口映射删除成功: %d", hostPort)
	return nil
}

// SyncNATRules 同步 NAT 规则（服务启动时调用）
func SyncNATRules() error {
	if !IsLXDAvailable() {
		log.Println("[Mock] 跳过 NAT 规则同步")
		return nil
	}
	
	var mappings []models.PortMapping
	models.DB.Find(&mappings)
	
	log.Printf("同步 NAT 规则，共 %d 条", len(mappings))
	
	for _, mapping := range mappings {
		// 重新添加规则
		dnatCmd := fmt.Sprintf(
			"iptables -t nat -A PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
			mapping.Protocol, mapping.HostPort, mapping.ContainerIP, mapping.ContainerPort,
		)
		exec.Command("sh", "-c", dnatCmd).Run()
		
		forwardCmd := fmt.Sprintf(
			"iptables -A FORWARD -p %s -d %s --dport %d -j ACCEPT",
			mapping.Protocol, mapping.ContainerIP, mapping.ContainerPort,
		)
		exec.Command("sh", "-c", forwardCmd).Run()
	}
	
	log.Println("NAT 规则同步完成")
	return nil
}
