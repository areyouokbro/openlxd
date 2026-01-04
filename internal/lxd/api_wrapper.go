package lxd

import "fmt"

// ClientWrapper LXD客户端包装器
type ClientWrapper struct{}

// NewClient 创建LXD客户端实例
func NewClient() *ClientWrapper {
	return &ClientWrapper{}
}

// CreateContainer 创建容器
func (c *ClientWrapper) CreateContainer(config ContainerConfig) error {
	req := CreateContainerRequest{
		Hostname: config.Name,
		CPUs:     config.CPU,
		Memory:   config.Memory,
		Disk:     config.Disk,
		Image:    config.Image,
	}
	return CreateContainer(req)
}

// StartContainer 启动容器
func (c *ClientWrapper) StartContainer(name string) error {
	return StartContainer(name)
}

// StopContainer 停止容器
func (c *ClientWrapper) StopContainer(name string) error {
	return StopContainer(name)
}

// RestartContainer 重启容器
func (c *ClientWrapper) RestartContainer(name string) error {
	return RestartContainer(name)
}

// DeleteContainer 删除容器
func (c *ClientWrapper) DeleteContainer(name string) error {
	return DeleteContainer(name)
}

// GetContainerState 获取容器状态
func (c *ClientWrapper) GetContainerState(name string) (*ContainerState, error) {
	state, err := GetContainerState(name)
	if err != nil {
		return nil, err
	}
	
	return &ContainerState{
		Status: state.Status,
		CPU:    state.CPU.Usage,
		Memory: state.Memory.Usage,
		Disk:   state.Disk["root"].Usage,
	}, nil
}

// SetPassword 设置容器密码
func (c *ClientWrapper) SetPassword(name, username, password string) error {
	return ResetContainerPassword(name, password)
}

// SetCPULimit 设置CPU限制
func (c *ClientWrapper) SetCPULimit(name string, cpus int) error {
	instance, _, err := Client.GetInstance(name)
	if err != nil {
		return err
	}
	instance.Config["limits.cpu"] = fmt.Sprintf("%d", cpus)
	op, err := Client.UpdateInstance(name, instance.Writable(), "")
	if err != nil {
		return err
	}
	return op.Wait()
}

// SetMemoryLimit 设置内存限制
func (c *ClientWrapper) SetMemoryLimit(name string, memoryMB int) error {
	instance, _, err := Client.GetInstance(name)
	if err != nil {
		return err
	}
	instance.Config["limits.memory"] = fmt.Sprintf("%dMB", memoryMB)
	op, err := Client.UpdateInstance(name, instance.Writable(), "")
	if err != nil {
		return err
	}
	return op.Wait()
}

// SetDiskLimit 设置磁盘限制
func (c *ClientWrapper) SetDiskLimit(name string, diskGB int) error {
	// 磁盘限制需要通过设备配置修改
	instance, _, err := Client.GetInstance(name)
	if err != nil {
		return err
	}
	if instance.Devices["root"] != nil {
		instance.Devices["root"]["size"] = fmt.Sprintf("%dGB", diskGB)
	}
	op, err := Client.UpdateInstance(name, instance.Writable(), "")
	if err != nil {
		return err
	}
	return op.Wait()
}

// ListImages 获取镜像列表
func (c *ClientWrapper) ListImages() ([]ImageInfo, error) {
	return ListImages()
}

// ImportImage 导入镜像
func (c *ClientWrapper) ImportImage(alias, architecture string) error {
	return ImportImage(alias, architecture)
}

// DeleteImage 删除镜像
func (c *ClientWrapper) DeleteImage(fingerprint string) error {
	return DeleteImage(fingerprint)
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	Name   string
	Image  string
	CPU    int
	Memory int
	Disk   int
}

// ContainerState 容器状态
type ContainerState struct {
	Status string
	CPU    int64
	Memory int64
	Disk   int64
}
