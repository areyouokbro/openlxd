package lxd

import (
	"fmt"
	"log"
	"strings"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

// CreateContainerFixed 创建容器（修复版本 - 支持远程镜像）
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

	// 解析镜像字符串 (例如: "alpine/edge" 或 "images:alpine/edge")
	imageName := req.Image
	imageServer := "images" // 默认使用images远程服务器
	
	// 如果包含冒号，分离服务器和镜像名
	if strings.Contains(imageName, ":") {
		parts := strings.SplitN(imageName, ":", 2)
		imageServer = parts[0]
		imageName = parts[1]
	}

	log.Printf("准备从远程服务器 %s 获取镜像 %s", imageServer, imageName)

	// 连接到远程镜像服务器
	var imageServerClient lxd.ImageServer
	var err error
	
	switch imageServer {
	case "images":
		imageServerClient, err = lxd.ConnectPublicLXD("https://images.lxd.canonical.com", nil)
	case "ubuntu":
		imageServerClient, err = lxd.ConnectSimpleStreams("https://cloud-images.ubuntu.com/releases/", nil)
	case "ubuntu-daily":
		imageServerClient, err = lxd.ConnectSimpleStreams("https://cloud-images.ubuntu.com/daily/", nil)
	default:
		return fmt.Errorf("不支持的镜像服务器: %s", imageServer)
	}
	
	if err != nil {
		return fmt.Errorf("连接镜像服务器失败: %v", err)
	}

	// 获取镜像别名
	imageAlias, _, err := imageServerClient.GetImageAlias(imageName)
	if err != nil {
		return fmt.Errorf("获取镜像别名失败: %v (镜像: %s)", err, imageName)
	}

	log.Printf("找到镜像: %s (指纹: %s)", imageName, imageAlias.Target[:12])

	// 创建容器请求
	instanceReq := api.InstancesPost{
		Name: req.Hostname,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:        "image",
			Fingerprint: imageAlias.Target,
			Server:      imageServerClient.GetConnectionInfo().URL,
			Protocol:    "lxd",
		},
		InstancePut: api.InstancePut{
			Config:  config,
			Devices: devices,
		},
	}

	log.Printf("开始创建容器 %s...", req.Hostname)

	// 创建容器
	op, err := Client.CreateInstance(instanceReq)
	if err != nil {
		return fmt.Errorf("创建容器失败: %v", err)
	}

	// 等待操作完成
	log.Printf("等待容器创建完成...")
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器创建操作失败: %v", err)
	}

	log.Printf("✅ 容器 %s 创建成功！", req.Hostname)

	// 如果指定了密码，设置 root 密码
	if req.Password != "" {
		log.Printf("容器 %s 创建成功，密码将在启动后设置", req.Hostname)
	}

	return nil
}
