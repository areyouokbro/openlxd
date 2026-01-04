package lxd

import (
	"fmt"
	"time"

	lxdapi "github.com/canonical/lxd/shared/api"
)

// CreateSnapshot 创建容器快照
func CreateSnapshot(containerName, snapshotName string, stateful bool) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 如果没有指定快照名称，使用时间戳
	if snapshotName == "" {
		snapshotName = fmt.Sprintf("snap-%d", time.Now().Unix())
	}

	// 创建快照请求
	req := lxdapi.InstanceSnapshotsPost{
		Name:     snapshotName,
		Stateful: stateful,
	}

	// 创建快照
	op, err := Client.CreateInstanceSnapshot(containerName, req)
	if err != nil {
		return fmt.Errorf("创建快照失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("快照操作失败: %v", err)
	}

	return nil
}

// ListSnapshots 列出容器的所有快照
func ListSnapshots(containerName string) ([]lxdapi.InstanceSnapshot, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD 客户端未初始化")
	}

	snapshots, err := Client.GetInstanceSnapshots(containerName)
	if err != nil {
		return nil, fmt.Errorf("获取快照列表失败: %v", err)
	}

	return snapshots, nil
}

// GetSnapshot 获取快照详情
func GetSnapshot(containerName, snapshotName string) (*lxdapi.InstanceSnapshot, error) {
	if Client == nil {
		return nil, fmt.Errorf("LXD 客户端未初始化")
	}

	snapshot, _, err := Client.GetInstanceSnapshot(containerName, snapshotName)
	if err != nil {
		return nil, fmt.Errorf("获取快照详情失败: %v", err)
	}

	return snapshot, nil
}

// RestoreSnapshot 恢复容器到指定快照
func RestoreSnapshot(containerName, snapshotName string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 恢复快照请求
	req := lxdapi.InstancePut{
		Restore: snapshotName,
	}

	// 执行恢复操作
	op, err := Client.UpdateInstance(containerName, req, "")
	if err != nil {
		return fmt.Errorf("恢复快照失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("快照恢复操作失败: %v", err)
	}

	return nil
}

// DeleteSnapshot 删除容器快照
func DeleteSnapshot(containerName, snapshotName string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 删除快照
	op, err := Client.DeleteInstanceSnapshot(containerName, snapshotName)
	if err != nil {
		return fmt.Errorf("删除快照失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("快照删除操作失败: %v", err)
	}

	return nil
}

// RenameSnapshot 重命名容器快照
func RenameSnapshot(containerName, oldName, newName string) error {
	if Client == nil {
		return fmt.Errorf("LXD 客户端未初始化")
	}

	// 重命名快照请求
	req := lxdapi.InstanceSnapshotPost{
		Name: newName,
	}

	// 执行重命名操作
	op, err := Client.RenameInstanceSnapshot(containerName, oldName, req)
	if err != nil {
		return fmt.Errorf("重命名快照失败: %v", err)
	}

	// 等待操作完成
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("快照重命名操作失败: %v", err)
	}

	return nil
}
