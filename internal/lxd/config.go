package lxd

import (
	"fmt"
	"strings"

	lxdapi "github.com/canonical/lxd/shared/api"
)

// SetDNS 设置容器的 DNS 服务器
func SetDNS(containerName string, dnsServers []string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 获取当前容器配置
	container, etag, err := Client.GetInstance(containerName)
	if err != nil {
		return fmt.Errorf("获取容器配置失败: %v", err)
	}

	// 更新 DNS 配置
	if container.Config == nil {
		container.Config = make(map[string]string)
	}

	// 设置 DNS 服务器（使用逗号分隔）
	container.Config["user.dns.servers"] = strings.Join(dnsServers, ",")

	// 更新容器配置
	op, err := Client.UpdateInstance(containerName, container.Writable(), etag)
	if err != nil {
		return fmt.Errorf("更新 DNS 配置失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("DNS 配置更新操作失败: %v", err)
	}

	return nil
}

// GetDNS 获取容器的 DNS 服务器配置
func GetDNS(containerName string) ([]string, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD 客户端未初始化")
	}

	// 获取容器配置
	container, _, err := Client.GetInstance(containerName)
	if err != nil {
		return nil, fmt.Errorf("获取容器配置失败: %v", err)
	}

	// 读取 DNS 配置
	dnsConfig, exists := container.Config["user.dns.servers"]
	if !exists || dnsConfig == "" {
		return []string{}, nil
	}

	// 分割 DNS 服务器列表
	dnsServers := strings.Split(dnsConfig, ",")
	return dnsServers, nil
}

// SetConfig 设置容器配置项
func SetConfig(containerName, key, value string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 获取当前容器配置
	container, etag, err := Client.GetInstance(containerName)
	if err != nil {
		return fmt.Errorf("获取容器配置失败: %v", err)
	}

	// 更新配置
	if container.Config == nil {
		container.Config = make(map[string]string)
	}
	container.Config[key] = value

	// 更新容器配置
	op, err := Client.UpdateInstance(containerName, container.Writable(), etag)
	if err != nil {
		return fmt.Errorf("更新配置失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("配置更新操作失败: %v", err)
	}

	return nil
}

// GetConfig 获取容器配置项
func GetConfig(containerName, key string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("LXD 客户端未初始化")
	}

	// 获取容器配置
	container, _, err := Client.GetInstance(containerName)
	if err != nil {
		return "", fmt.Errorf("获取容器配置失败: %v", err)
	}

	// 读取配置项
	value, exists := container.Config[key]
	if !exists {
		return "", nil
	}

	return value, nil
}

// SetResourceLimits 设置容器资源限制
func SetResourceLimits(containerName string, cpuLimit, memoryLimit, diskLimit string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 获取当前容器配置
	container, etag, err := Client.GetInstance(containerName)
	if err != nil {
		return fmt.Errorf("获取容器配置失败: %v", err)
	}

	// 更新资源限制
	if container.Config == nil {
		container.Config = make(map[string]string)
	}

	if cpuLimit != "" {
		container.Config["limits.cpu"] = cpuLimit
	}
	if memoryLimit != "" {
		container.Config["limits.memory"] = memoryLimit
	}
	if diskLimit != "" {
		container.Config["limits.disk"] = diskLimit
	}

	// 更新容器配置
	op, err := Client.UpdateInstance(containerName, container.Writable(), etag)
	if err != nil {
		return fmt.Errorf("更新资源限制失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("资源限制更新操作失败: %v", err)
	}

	return nil
}

// ExecCommand 在容器中执行命令
func ExecCommand(containerName string, command []string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("LXD 客户端未初始化")
	}

	// 创建执行请求
	req := lxdapi.InstanceExecPost{
		Command:     command,
		WaitForWS:   false,
		Interactive: false,
		Environment: map[string]string{
			"TERM": "xterm",
		},
	}

	// 执行命令
	op, err := Client.ExecInstance(containerName, req, nil)
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return "", fmt.Errorf("命令执行操作失败: %v", err)
	}

	// 获取操作结果
	opAPI := op.Get()
	
	// 检查返回码
	if opAPI.Metadata == nil {
		return "", fmt.Errorf("无法获取命令执行结果")
	}

	return "命令执行成功", nil
}
