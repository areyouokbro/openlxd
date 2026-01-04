package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
)

// HandleSnapshots 处理容器快照请求
func HandleSnapshots(w http.ResponseWriter, r *http.Request) {
	containerName := r.URL.Query().Get("container")
	if containerName == "" {
		respondJSON(w, 400, "容器名称不能为空", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 列出快照
		snapshots, err := lxd.ListSnapshots(containerName)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("获取快照列表失败: %v", err), nil)
			return
		}
		respondJSON(w, 200, "获取成功", snapshots)

	case http.MethodPost:
		// 创建快照
		var req struct {
			SnapshotName string `json:"snapshot_name"`
			Stateful     bool   `json:"stateful"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, "请求参数错误", nil)
			return
		}

		err := lxd.CreateSnapshot(containerName, req.SnapshotName, req.Stateful)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("创建快照失败: %v", err), nil)
			return
		}

		// 记录日志
		models.LogAction("create_snapshot", containerName, fmt.Sprintf("创建快照: %s", req.SnapshotName), "success")
		respondJSON(w, 200, "快照创建成功", nil)

	case http.MethodPut:
		// 恢复快照
		var req struct {
			SnapshotName string `json:"snapshot_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, "请求参数错误", nil)
			return
		}

		err := lxd.RestoreSnapshot(containerName, req.SnapshotName)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("恢复快照失败: %v", err), nil)
			return
		}

		// 记录日志
		models.LogAction("restore_snapshot", containerName, fmt.Sprintf("恢复快照: %s", req.SnapshotName), "success")
		respondJSON(w, 200, "快照恢复成功", nil)

	case http.MethodDelete:
		// 删除快照
		snapshotName := r.URL.Query().Get("snapshot")
		if snapshotName == "" {
			respondJSON(w, 400, "快照名称不能为空", nil)
			return
		}

		err := lxd.DeleteSnapshot(containerName, snapshotName)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("删除快照失败: %v", err), nil)
			return
		}

		// 记录日志
		models.LogAction("delete_snapshot", containerName, fmt.Sprintf("删除快照: %s", snapshotName), "success")
		respondJSON(w, 200, "快照删除成功", nil)

	default:
		respondJSON(w, 405, "Method not allowed", nil)
	}
}

// HandleClone 处理容器克隆请求
func HandleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	var req struct {
		SourceContainer string `json:"source_container"`
		TargetContainer string `json:"target_container"`
		SnapshotName    string `json:"snapshot_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	if req.SourceContainer == "" || req.TargetContainer == "" {
		respondJSON(w, 400, "源容器和目标容器名称不能为空", nil)
		return
	}

	var err error
	if req.SnapshotName != "" {
		// 从快照克隆
		err = lxd.CloneContainerFromSnapshot(req.SourceContainer, req.SnapshotName, req.TargetContainer)
	} else {
		// 直接克隆容器
		err = lxd.CloneContainer(req.SourceContainer, req.TargetContainer)
	}

	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("克隆容器失败: %v", err), nil)
		return
	}

	// 记录日志
	models.LogAction("clone_container", req.TargetContainer, fmt.Sprintf("从 %s 克隆", req.SourceContainer), "success")
	respondJSON(w, 200, "容器克隆成功", nil)
}

// HandleDNS 处理容器 DNS 设置请求
func HandleDNS(w http.ResponseWriter, r *http.Request) {
	containerName := r.URL.Query().Get("container")
	if containerName == "" {
		respondJSON(w, 400, "容器名称不能为空", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 获取 DNS 配置
		dnsServers, err := lxd.GetDNS(containerName)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("获取 DNS 配置失败: %v", err), nil)
			return
		}
		respondJSON(w, 200, "获取成功", map[string]interface{}{
			"dns_servers": dnsServers,
		})

	case http.MethodPost:
		// 设置 DNS 配置
		var req struct {
			DNSServers []string `json:"dns_servers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, "请求参数错误", nil)
			return
		}

		err := lxd.SetDNS(containerName, req.DNSServers)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("设置 DNS 配置失败: %v", err), nil)
			return
		}

		// 记录日志
		models.LogAction("set_dns", containerName, fmt.Sprintf("设置 DNS: %v", req.DNSServers), "success")
		respondJSON(w, 200, "DNS 配置成功", nil)

	default:
		respondJSON(w, 405, "Method not allowed", nil)
	}
}

// HandleExecCommand 处理容器命令执行请求
func HandleExecCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	var req struct {
		Container string   `json:"container"`
		Command   []string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	if req.Container == "" || len(req.Command) == 0 {
		respondJSON(w, 400, "容器名称和命令不能为空", nil)
		return
	}

	output, err := lxd.ExecCommand(req.Container, req.Command)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("执行命令失败: %v", err), nil)
		return
	}

	// 记录日志
	models.LogAction("exec_command", req.Container, fmt.Sprintf("执行命令: %v", req.Command), "success")
	respondJSON(w, 200, "命令执行成功", map[string]interface{}{
		"output": output,
	})
}

// HandleResourceLimits 处理容器资源限制请求
func HandleResourceLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	var req struct {
		Container   string `json:"container"`
		CPULimit    string `json:"cpu_limit"`
		MemoryLimit string `json:"memory_limit"`
		DiskLimit   string `json:"disk_limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	if req.Container == "" {
		respondJSON(w, 400, "容器名称不能为空", nil)
		return
	}

	err := lxd.SetResourceLimits(req.Container, req.CPULimit, req.MemoryLimit, req.DiskLimit)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("设置资源限制失败: %v", err), nil)
		return
	}

	// 记录日志
	models.LogAction("set_limits", req.Container, "设置资源限制", "success")
	respondJSON(w, 200, "资源限制设置成功", nil)
}
