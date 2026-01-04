package lxd

import (
	"fmt"
	"log"
	"time"
)

// CreateContainerRequest 创建容器请求
type CreateContainerRequest struct {
	Hostname     string
	CPUs         int
	Memory       int
	Disk         int
	Image        string
	Ingress      int
	Egress       int
	Password     string
	CPUAllowance int
}

// CreateContainer 创建容器
func CreateContainer(req CreateContainerRequest) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 创建容器: %s (镜像: %s, CPU: %d, 内存: %dMB)", 
			req.Hostname, req.Image, req.CPUs, req.Memory)
		return nil
	}
	
	// 构建容器配置
	config := map[string]string{
		"limits.cpu":    fmt.Sprintf("%d", req.CPUs),
		"limits.memory": fmt.Sprintf("%dMB", req.Memory),
	}
	
	// CPU 使用率限制
	if req.CPUAllowance > 0 && req.CPUAllowance < 100 {
		config["limits.cpu.allowance"] = fmt.Sprintf("%d%%", req.CPUAllowance)
	}
	
	// 网络带宽限制
	if req.Ingress > 0 {
		config["limits.network.ingress"] = fmt.Sprintf("%dMbit", req.Ingress)
	}
	if req.Egress > 0 {
		config["limits.network.egress"] = fmt.Sprintf("%dMbit", req.Egress)
	}
	
	// 设置 root 密码
	if req.Password != "" {
		config["user.user-data"] = fmt.Sprintf(`#cloud-config
password: %s
chpasswd: { expire: False }
ssh_pwauth: True
`, req.Password)
	}
	
	// 设备配置
	devices := map[string]map[string]string{
		"root": {
			"type": "disk",
			"path": "/",
			"pool": "default",
			"size": fmt.Sprintf("%dMB", req.Disk),
		},
		"eth0": {
			"type":    "nic",
			"nictype": "bridged",
			"parent":  "lxdbr0",
			"name":    "eth0",
		},
	}
	
	// 创建容器请求体
	reqBody := map[string]interface{}{
		"name": req.Hostname,
		"type": "container",
		"config": config,
		"devices": devices,
		"source": map[string]string{
			"type":  "image",
			"alias": req.Image,
		},
	}
	
	// 发送创建请求
	_, err := lxdRequest("POST", "/1.0/instances", reqBody)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	
	log.Printf("容器创建成功: %s", req.Hostname)
	return nil
}

// StartContainer 启动容器
func StartContainer(name string) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 启动容器: %s", name)
		return nil
	}
	
	reqBody := map[string]interface{}{
		"action":  "start",
		"timeout": 30,
	}
	
	_, err := lxdRequest("PUT", "/1.0/instances/"+name+"/state", reqBody)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	
	log.Printf("容器启动成功: %s", name)
	return nil
}

// StopContainer 停止容器
func StopContainer(name string) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 停止容器: %s", name)
		return nil
	}
	
	reqBody := map[string]interface{}{
		"action":  "stop",
		"timeout": 30,
		"force":   false,
	}
	
	_, err := lxdRequest("PUT", "/1.0/instances/"+name+"/state", reqBody)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	
	log.Printf("容器停止成功: %s", name)
	return nil
}

// RestartContainer 重启容器
func RestartContainer(name string) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 重启容器: %s", name)
		return nil
	}
	
	reqBody := map[string]interface{}{
		"action":  "restart",
		"timeout": 30,
		"force":   false,
	}
	
	_, err := lxdRequest("PUT", "/1.0/instances/"+name+"/state", reqBody)
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}
	
	log.Printf("容器重启成功: %s", name)
	return nil
}

// DeleteContainer 删除容器
func DeleteContainer(name string) error {
	if !IsLXDAvailable() {
		log.Printf("[Mock] 删除容器: %s", name)
		return nil
	}
	
	// 先停止容器
	state, _ := GetContainerState(name)
	if state != nil {
		if status, ok := state["status"].(string); ok && status == "Running" {
			StopContainer(name)
			time.Sleep(2 * time.Second)
		}
	}
	
	// 删除容器
	_, err := lxdRequest("DELETE", "/1.0/instances/"+name, nil)
	if err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}
	
	log.Printf("容器删除成功: %s", name)
	return nil
}
