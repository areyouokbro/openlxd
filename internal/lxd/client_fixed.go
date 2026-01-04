package lxd

import (
	"fmt"
	"log"
	"strings"

	lxd "github.com/canonical/lxd/client"
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

	// 解析镜像名称
	var imageServer lxd.ImageServer
	var imageAlias string
	var err error

	if strings.Contains(req.Image, ":") {
		// 格式: server:image (例如: images:alpine/3.19)
		parts := strings.SplitN(req.Image, ":", 2)
		serverName := parts[0]
		imageAlias = parts[1]

		// 连接到远程镜像服务器
		switch serverName {
		case "images":
			imageServer, err = lxd.ConnectSimpleStreams("https://images.linuxcontainers.org", nil)
		case "ubuntu":
			imageServer, err = lxd.ConnectSimpleStreams("https://cloud-images.ubuntu.com/releases", nil)
		case "ubuntu-daily":
			imageServer, err = lxd.ConnectSimpleStreams("https://cloud-images.ubuntu.com/daily", nil)
		default:
			return fmt.Errorf("不支持的镜像服务器: %s", serverName)
		}

		if err != nil {
			return fmt.Errorf("连接镜像服务器失败: %v", err)
		}
	} else {
		// 使用本地镜像
		imageServer = Client
		imageAlias = req.Image
	}

	// 获取镜像信息
	image, _, err := imageServer.GetImageAlias(imageAlias)
	if err != nil {
		return fmt.Errorf("获取镜像信息失败: %v (镜像: %s)", err, imageAlias)
	}

	// 创建容器请求
	instanceReq := api.InstancesPost{
		Name: req.Hostname,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:        "image",
			Fingerprint: image.Target,
			Server:      imageServer.GetConnectionInfo().URL,
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
