package lxd

import (
	"fmt"
	"log"

	"github.com/canonical/lxd/shared/api"
)

// CreateContainerFixed 创建容器（修复版本）
func CreateContainerFixed(req CreateContainerRequest) error {
	// 构建容器配置
	config := map[string]string{
		"limits.cpu":    fmt.Sprintf("%d", req.CPUs),
		"limits.memory": fmt.Sprintf("%dMB", req.Memory),
	}

	// 如果指定了网络限制
	if req.Ingress > 0 {
		config["limits.network.priority"] = "10"
	}

	// 构建设备配置
	devices := map[string]map[string]string{
		"root": {
			"path": "/",
			"pool": "default",
			"type": "disk",
			"size": fmt.Sprintf("%dGB", req.Disk),
		},
		"eth0": {
			"name":    "eth0",
			"type":    "nic",
			"nictype": "bridged",
			"parent":  "lxdbr0",
		},
	}

	// 创建容器请求
	instanceReq := api.InstancesPost{
		Name: req.Hostname,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:  "image",
			Alias: req.Image,  // 直接使用镜像别名（例如: images:alpine/edge）
		},
		InstancePut: api.InstancePut{
			Config:  config,
			Devices: devices,
		},
	}

	// 创建容器
	op, err := Client.CreateInstance(instanceReq)
	if err != nil {
		return fmt.Errorf("创建容器失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器创建操作失败: %v", err)
	}

	log.Printf("容器 %s 创建成功", req.Hostname)

	// 如果指定了密码，设置 root 密码
	if req.Password != "" {
		log.Printf("容器 %s 创建成功，密码将在启动后设置", req.Hostname)
	}

	return nil
}
