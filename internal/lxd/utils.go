package lxd

import (
	"log"
	"strings"
)

// GetContainerIP 获取容器的 IPv4 地址
func GetContainerIP(name string) string {
	if !IsLXDAvailable() {
		return "10.0.0.100" // Mock IP
	}
	
	state, err := GetContainerState(name)
	if err != nil {
		log.Printf("获取容器状态失败: %v", err)
		return ""
	}
	
	// 从网络接口中提取 IP
	if network, ok := state["network"].(map[string]interface{}); ok {
		for _, iface := range network {
			if ifaceData, ok := iface.(map[string]interface{}); ok {
				if addresses, ok := ifaceData["addresses"].([]interface{}); ok {
					for _, addr := range addresses {
						if addrData, ok := addr.(map[string]interface{}); ok {
							family := addrData["family"].(string)
							address := addrData["address"].(string)
							scope := addrData["scope"].(string)
							
							// 只返回全局 IPv4 地址
							if family == "inet" && scope == "global" {
								return address
							}
						}
					}
				}
			}
		}
	}
	
	return ""
}

// GetContainerStatus 获取容器状态
func GetContainerStatus(name string) string {
	if !IsLXDAvailable() {
		return "Running" // Mock 状态
	}
	
	state, err := GetContainerState(name)
	if err != nil {
		log.Printf("获取容器状态失败: %v", err)
		return "Unknown"
	}
	
	if status, ok := state["status"].(string); ok {
		// 将 LXD 状态转换为友好的状态名
		switch strings.ToLower(status) {
		case "running":
			return "Running"
		case "stopped":
			return "Stopped"
		case "frozen":
			return "Paused"
		default:
			return status
		}
	}
	
	return "Unknown"
}

// ReinstallContainer 重装容器系统
func ReinstallContainer(name, newImage string) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 重装容器: %s, 新镜像: %s", name, newImage)
		return nil
	}
	
	// 1. 获取原容器配置
	container, err := GetContainer(name)
	if err != nil {
		return err
	}
	
	config := container["config"].(map[string]interface{})
	devices := container["devices"].(map[string]interface{})
	
	// 2. 删除旧容器
	if err := DeleteContainer(name); err != nil {
		return err
	}
	
	// 3. 创建新容器（使用相同配置）
	reqBody := map[string]interface{}{
		"name":    name,
		"type":    "container",
		"config":  config,
		"devices": devices,
		"source": map[string]string{
			"type":  "image",
			"alias": newImage,
		},
	}
	
	_, err = lxdRequest("POST", "/1.0/instances", reqBody)
	if err != nil {
		return err
	}
	
	// 4. 启动容器
	StartContainer(name)
	
	log.Printf("容器重装成功: %s", name)
	return nil
}

// ResetContainerPassword 重置容器 root 密码
func ResetContainerPassword(name, newPassword string) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 重置密码: %s", name)
		return nil
	}
	
	// 通过 lxc exec 执行密码重置命令
	execReq := map[string]interface{}{
		"command": []string{"/bin/sh", "-c", "echo 'root:" + newPassword + "' | chpasswd"},
		"wait-for-websocket": false,
		"interactive": false,
	}
	
	_, err := lxdRequest("POST", "/1.0/instances/"+name+"/exec", execReq)
	if err != nil {
		return err
	}
	
	log.Printf("密码重置成功: %s", name)
	return nil
}
