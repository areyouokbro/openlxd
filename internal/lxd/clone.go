package lxd

import (
	"fmt"

	lxdapi "github.com/canonical/lxd/shared/api"
)

// CloneContainer 克隆容器
func CloneContainer(sourceName, targetName string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 创建克隆请求
	req := lxdapi.InstancesPost{
		Name: targetName,
		Source: lxdapi.InstanceSource{
			Type:   "copy",
			Source: sourceName,
		},
		Type: lxdapi.InstanceTypeContainer,
	}

	// 执行克隆操作
	op, err := Client.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("克隆容器失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器克隆操作失败: %v", err)
	}

	return nil
}

// CloneContainerFromSnapshot 从快照克隆容器
func CloneContainerFromSnapshot(sourceName, snapshotName, targetName string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 构建快照源路径
	snapshotSource := fmt.Sprintf("%s/%s", sourceName, snapshotName)

	// 创建克隆请求
	req := lxdapi.InstancesPost{
		Name: targetName,
		Source: lxdapi.InstanceSource{
			Type:   "copy",
			Source: snapshotSource,
		},
		Type: lxdapi.InstanceTypeContainer,
	}

	// 执行克隆操作
	op, err := Client.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("从快照克隆容器失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("快照克隆操作失败: %v", err)
	}

	return nil
}

// CopyContainer 复制容器（带配置）
func CopyContainer(sourceName, targetName string, config map[string]string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 创建复制请求
	req := lxdapi.InstancesPost{
		Name: targetName,
		Source: lxdapi.InstanceSource{
			Type:   "copy",
			Source: sourceName,
		},
		Type: lxdapi.InstanceTypeContainer,
	}

	// 如果提供了配置，添加到请求中
	if len(config) > 0 {
		req.Config = config
	}

	// 执行复制操作
	op, err := Client.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("复制容器失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("容器复制操作失败: %v", err)
	}

	return nil
}
