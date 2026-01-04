package lxd

import (
	"fmt"
	"log"
	"strings"
	"time"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

var Client lxd.InstanceServer

// Connect 连接到LXD
func Connect(socketPath string) error {
	var err error
	Client, err = lxd.ConnectLXDUnix(socketPath, nil)
	if err != nil {
		return fmt.Errorf("连接LXD失败: %v", err)
	}

	// 测试连接
	_, _, err = Client.GetServer()
	if err != nil {
		return fmt.Errorf("获取LXD服务器信息失败: %v", err)
	}

	log.Println("LXD连接成功")
	return nil
}

// ListContainers 获取所有容器列表
func ListContainers() ([]api.Instance, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD未连接")
	}

	containers, err := Client.GetInstances(api.InstanceTypeContainer)
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %v", err)
	}

	return containers, nil
}

// GetContainer 获取单个容器信息
func GetContainer(name string) (*api.Instance, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD未连接")
	}

	container, _, err := Client.GetInstance(name)
	if err != nil {
		return nil, fmt.Errorf("获取容器信息失败: %v", err)
	}

	return container, nil
}

// GetContainerState 获取容器状态
func GetContainerState(name string) (*api.InstanceState, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD未连接")
	}

	state, _, err := Client.GetInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("获取容器状态失败: %v", err)
	}

	return state, nil
}

// CreateContainer 创建容器
func CreateContainer(name, image string, cpu, memory, disk int) error {
	if Client == nil {
		return fmt.Errorf("LXD未连接")
	}

	// 构建容器配置
	req := api.InstancesPost{
		Name: name,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:  "image",
			Alias: image,
		},
		InstancePut: api.InstancePut{
			Config: map[string]string{
				"limits.cpu":    fmt.Sprintf("%d", cpu),
				"limits.memory": fmt.Sprintf("%dMB", memory),
			},
			Devices: map[string]map[string]string{
				"root": {
					"type": "disk",
					"path": "/",
					"pool": "default",
					"size": fmt.Sprintf("%dGB", disk),
				},
			},
		},
	}

	// 创建容器
	op, err := Client.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("创建容器失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("等待容器创建失败: %v", err)
	}

	log.Printf("容器 %s 创建成功", name)
	return nil
}

// StartContainer 启动容器
func StartContainer(name string) error {
	if Client == nil {
		return fmt.Errorf("LXD未连接")
	}

	req := api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
	}

	op, err := Client.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("启动容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("等待容器启动失败: %v", err)
	}

	log.Printf("容器 %s 启动成功", name)
	return nil
}

// StopContainer 停止容器
func StopContainer(name string) error {
	if Client == nil {
		return fmt.Errorf("LXD未连接")
	}

	req := api.InstanceStatePut{
		Action:  "stop",
		Timeout: 30,
		Force:   false,
	}

	op, err := Client.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("停止容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("等待容器停止失败: %v", err)
	}

	log.Printf("容器 %s 停止成功", name)
	return nil
}

// RestartContainer 重启容器
func RestartContainer(name string) error {
	if Client == nil {
		return fmt.Errorf("LXD未连接")
	}

	req := api.InstanceStatePut{
		Action:  "restart",
		Timeout: 30,
		Force:   false,
	}

	op, err := Client.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("重启容器失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("等待容器重启失败: %v", err)
	}

	log.Printf("容器 %s 重启成功", name)
	return nil
}

// DeleteContainer 删除容器
func DeleteContainer(name string) error {
	if Client == nil {
		return fmt.Errorf("LXD未连接")
	}

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
		return fmt.Errorf("等待容器删除失败: %v", err)
	}

	log.Printf("容器 %s 删除成功", name)
	return nil
}

// GetContainerIP 获取容器IP地址
func GetContainerIP(name string) (string, error) {
	state, err := GetContainerState(name)
	if err != nil {
		return "", err
	}

	// 获取eth0网卡的IPv4地址
	if network, ok := state.Network["eth0"]; ok {
		for _, addr := range network.Addresses {
			if addr.Family == "inet" && addr.Scope == "global" {
				return addr.Address, nil
			}
		}
	}

	return "", fmt.Errorf("未找到容器IP地址")
}

// ExecuteCommand 在容器中执行命令
func ExecuteCommand(name string, command []string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("LXD未连接")
	}

	req := api.InstanceExecPost{
		Command:     command,
		WaitForWS:   true,
		Interactive: false,
	}

	op, err := Client.ExecInstance(name, req, nil)
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	err = op.Wait()
	if err != nil {
		return "", fmt.Errorf("等待命令执行失败: %v", err)
	}

	return "", nil
}

// SetRootPassword 设置容器root密码
func SetRootPassword(name, password string) error {
	command := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf("echo 'root:%s' | chpasswd", password),
	}

	_, err := ExecuteCommand(name, command)
	if err != nil {
		return fmt.Errorf("设置密码失败: %v", err)
	}

	log.Printf("容器 %s 密码设置成功", name)
	return nil
}

// ListImages 获取镜像列表
func ListImages() ([]api.Image, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD未连接")
	}

	images, err := Client.GetImages()
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %v", err)
	}

	return images, nil
}

// GetImageAlias 获取镜像别名
func GetImageAlias(fingerprint string) string {
	if Client == nil {
		return ""
	}

	aliases, err := Client.GetImageAliases()
	if err != nil {
		return ""
	}

	for _, alias := range aliases {
		if strings.HasPrefix(alias.Target, fingerprint[:12]) {
			return alias.Name
		}
	}

	return fingerprint[:12]
}
