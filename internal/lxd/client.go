package lxd

import (
	"fmt"
	"log"
	"time"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

var Client lxd.InstanceServer

// GetClient 获取 LXD 客户端
func GetClient() lxd.InstanceServer {
	return Client
}

// InitLXD 初始化 LXD 客户端
func InitLXD(socketPath string) error {
	var err error
	Client, err = lxd.ConnectLXDUnix(socketPath, nil)
	if err != nil {
		return fmt.Errorf("LXD 连接失败: %v", err)
	}

	// 测试连接
	server, _, err := Client.GetServer()
	if err != nil {
		return fmt.Errorf("LXD 服务器信息获取失败: %v", err)
	}

	log.Printf("LXD 连接成功: %s", server.Environment.ServerName)
	return nil
}

// ListContainers 获取所有容器列表
func ListContainers() ([]api.Instance, error) {
	instances, err := Client.GetInstances(api.InstanceTypeContainer)
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %v", err)
	}
	return instances, nil
}

// GetContainer 获取单个容器信息
func GetContainer(name string) (*api.Instance, error) {
	instance, _, err := Client.GetInstance(name)
	if err != nil {
		return nil, fmt.Errorf("获取容器信息失败: %v", err)
	}
	return instance, nil
}

// GetContainerState 获取容器状态
func GetContainerState(name string) (*api.InstanceState, error) {
	state, _, err := Client.GetInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("获取容器状态失败: %v", err)
	}
	return state, nil
}

// CreateContainer 创建容器
func CreateContainer(req CreateContainerRequest) error {
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
			Alias: req.Image,
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

	// 如果指定了密码，设置 root 密码
	if req.Password != "" {
		// 注意：这需要容器启动后执行
		log.Printf("容器 %s 创建成功，密码将在启动后设置", req.Hostname)
	}

	return nil
}

// StartContainer 启动容器
func StartContainer(name string) error {
	reqState := api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
	}

	op, err := Client.UpdateInstanceState(name, reqState, "")
	if err != nil {
		return fmt.Errorf("启动容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器启动操作失败: %v", err)
	}

	log.Printf("容器 %s 启动成功", name)
	return nil
}

// StopContainer 停止容器
func StopContainer(name string) error {
	reqState := api.InstanceStatePut{
		Action:  "stop",
		Timeout: 30,
		Force:   false,
	}

	op, err := Client.UpdateInstanceState(name, reqState, "")
	if err != nil {
		return fmt.Errorf("停止容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器停止操作失败: %v", err)
	}

	log.Printf("容器 %s 停止成功", name)
	return nil
}

// RestartContainer 重启容器
func RestartContainer(name string) error {
	reqState := api.InstanceStatePut{
		Action:  "restart",
		Timeout: 30,
		Force:   false,
	}

	op, err := Client.UpdateInstanceState(name, reqState, "")
	if err != nil {
		return fmt.Errorf("重启容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器重启操作失败: %v", err)
	}

	log.Printf("容器 %s 重启成功", name)
	return nil
}

// DeleteContainer 删除容器
func DeleteContainer(name string) error {
	// 先停止容器
	state, _, _ := Client.GetInstanceState(name)
	if state != nil && state.Status == "Running" {
		StopContainer(name)
		time.Sleep(2 * time.Second)
	}

	// 删除容器
	op, err := Client.DeleteInstance(name)
	if err != nil {
		return fmt.Errorf("删除容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器删除操作失败: %v", err)
	}

	log.Printf("容器 %s 删除成功", name)
	return nil
}

// GetContainerIP 获取容器IP地址
func GetContainerIP(name string) (string, string) {
	state, err := GetContainerState(name)
	if err != nil {
		return "", ""
	}

	var ipv4, ipv6 string
	if eth0, ok := state.Network["eth0"]; ok {
		for _, addr := range eth0.Addresses {
			if addr.Family == "inet" && addr.Scope == "global" {
				ipv4 = addr.Address
			}
			if addr.Family == "inet6" && addr.Scope == "global" {
				ipv6 = addr.Address
			}
		}
	}

	return ipv4, ipv6
}

// ResetContainerPassword 重置容器密码
func ResetContainerPassword(name, password string) error {
	// 构建执行命令
	execReq := api.InstanceExecPost{
		Command: []string{"/bin/bash", "-c", fmt.Sprintf("echo 'root:%s' | chpasswd", password)},
	}

	op, err := Client.ExecInstance(name, execReq, nil)
	if err != nil {
		return fmt.Errorf("执行密码重置命令失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("密码重置操作失败: %v", err)
	}

	log.Printf("容器 %s 密码重置成功", name)
	return nil
}

// ReinstallContainer 重装容器系统
func ReinstallContainer(name, newImage string) error {
	// 获取原容器配置
	instance, _, err := Client.GetInstance(name)
	if err != nil {
		return fmt.Errorf("获取容器配置失败: %v", err)
	}

	// 删除原容器
	err = DeleteContainer(name)
	if err != nil {
		return fmt.Errorf("删除原容器失败: %v", err)
	}

	// 等待删除完成
	time.Sleep(2 * time.Second)

	// 使用新镜像创建容器
	instanceReq := api.InstancesPost{
		Name: name,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:  "image",
			Alias: newImage,
		},
		InstancePut: api.InstancePut{
			Config:  instance.Config,
			Devices: instance.Devices,
		},
	}

	op, err := Client.CreateInstance(instanceReq)
	if err != nil {
		return fmt.Errorf("重建容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器重建操作失败: %v", err)
	}

	log.Printf("容器 %s 重装成功，新镜像: %s", name, newImage)
	return nil
}

// CreateContainerRequest 创建容器请求结构
type CreateContainerRequest struct {
	Hostname     string
	CPUs         int
	Memory       int
	Disk         int
	Image        string
	Password     string
	Ingress      int
	Egress       int
	CPUAllowance int
}

// ImageInfo 镜像信息
type ImageInfo struct {
	Alias        string
	Fingerprint  string
	Architecture string
	Description  string
	Size         int64
}

// ListImages 获取镜像列表
func ListImages() ([]ImageInfo, error) {
	images, err := Client.GetImages()
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %v", err)
	}

	var imageList []ImageInfo
	for _, img := range images {
		alias := ""
		if len(img.Aliases) > 0 {
			alias = img.Aliases[0].Name
		}

		imageList = append(imageList, ImageInfo{
			Alias:        alias,
			Fingerprint:  img.Fingerprint,
			Architecture: img.Architecture,
			Description:  img.Properties["description"],
			Size:         img.Size,
		})
	}

	return imageList, nil
}

// ImportImage 从远程导入镜像
func ImportImage(alias, architecture string) error {
	// 连接到远程镜像服务器
	remote, err := lxd.ConnectSimpleStreams("https://images.linuxcontainers.org", nil)
	if err != nil {
		return fmt.Errorf("连接远程镜像服务器失败: %v", err)
	}

	// 复制镜像请求
	req := lxd.ImageCopyArgs{
		Aliases: []api.ImageAlias{{Name: alias}},
	}

	// 从远程复制镜像
	op, err := Client.CopyImage(remote, api.Image{
		Filename: alias,
	}, &req)
	if err != nil {
		return fmt.Errorf("复制镜像失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("镜像导入操作失败: %v", err)
	}

	log.Printf("镜像 %s 导入成功", alias)
	return nil
}

// DeleteImage 删除镜像
func DeleteImage(fingerprint string) error {
	op, err := Client.DeleteImage(fingerprint)
	if err != nil {
		return fmt.Errorf("删除镜像失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("镜像删除操作失败: %v", err)
	}

	log.Printf("镜像 %s 删除成功", fingerprint)
	return nil
}

// GetImage 获取镜像信息
func GetImage(fingerprint string) (*api.Image, error) {
	image, _, err := Client.GetImage(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("获取镜像信息失败: %v", err)
	}
	return image, nil
}
