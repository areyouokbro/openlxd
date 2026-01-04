package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
	lxdapi "github.com/canonical/lxd/shared/api"
)

// HandleListContainers 获取容器列表
func HandleListContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondContainerJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 从数据库获取容器列表
	var containers []models.Container
	if err := models.DB.Find(&containers).Error; err != nil {
		RespondContainerJSON(w, 500, "获取容器列表失败: "+err.Error(), nil)
		return
	}

	// 获取LXD客户端
	client := lxd.GetClient()
	if client == nil {
		// LXD未连接，返回数据库中的容器信息
		RespondContainerJSON(w, 200, "获取成功", containers)
		return
	}

	// 从LXD获取实时状态
	var result []map[string]interface{}
	for _, container := range containers {
		// 获取容器状态
		state, _, err := client.GetInstanceState(container.Hostname)
		if err != nil {
			// 如果获取失败，使用数据库中的信息
			result = append(result, map[string]interface{}{
				"id":          container.ID,
				"name":        container.Hostname,
				"status":      container.Status,
				"image":       container.Image,
				"cpu":         container.CPUs,
				"memory":      container.Memory,
				"disk":        container.Disk,
				"ipv4":        container.IPv4,
				"ipv6":        container.IPv6,
				"created_at":  container.CreatedAt,
			})
			continue
		}

		// 提取IP地址
		ipv4, ipv6 := extractIPFromContainerState(state)

		// 构建返回数据
		result = append(result, map[string]interface{}{
			"id":          container.ID,
			"name":        container.Hostname,
			"status":      strings.ToLower(state.Status),
			"image":       container.Image,
			"cpu":         container.CPUs,
			"memory":      container.Memory,
			"disk":        container.Disk,
			"ipv4":        ipv4,
			"ipv6":        ipv6,
			"created_at":  container.CreatedAt,
			"cpu_usage":   state.CPU.Usage,
			"memory_usage": state.Memory.Usage,
		})
	}

	RespondContainerJSON(w, 200, "获取成功", result)
}

// HandleCreateContainer 创建容器
func HandleCreateContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondContainerJSON(w, 405, "Method not allowed", nil)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Image    string `json:"image"`
		CPU      int    `json:"cpu"`
		Memory   int    `json:"memory"`
		Disk     int    `json:"disk"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondContainerJSON(w, 400, "Invalid request: "+err.Error(), nil)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Image == "" {
		RespondContainerJSON(w, 400, "Name and image are required", nil)
		return
	}

	// 检查容器名称是否已存在
	var existing models.Container
	if err := models.DB.Where("hostname = ?", req.Name).First(&existing).Error; err == nil {
		RespondContainerJSON(w, 400, "Container name already exists", nil)
		return
	}

	// 获取LXD客户端
	client := lxd.GetClient()
	if client == nil {
		RespondContainerJSON(w, 500, "LXD client not available", nil)
		return
	}

	// 创建容器配置
	config := map[string]string{
		"security.nesting": "true",
	}
	if req.Password != "" {
		config["user.password"] = req.Password
	}

	// 设置资源限制
	if req.CPU > 0 {
		config["limits.cpu"] = strconv.Itoa(req.CPU)
	}
	if req.Memory > 0 {
		config["limits.memory"] = strconv.Itoa(req.Memory) + "MB"
	}

	// 创建容器请求
	createReq := lxdapi.InstancesPost{
		Name: req.Name,
		Source: lxdapi.InstanceSource{
			Type:  "image",
			Alias: req.Image,
		},
		Type: "container",
		InstancePut: lxdapi.InstancePut{
			Config: config,
		},
	}

	// 创建容器
	op, err := client.CreateInstance(createReq)
	if err != nil {
		RespondContainerJSON(w, 500, "Failed to create container: "+err.Error(), nil)
		return
	}

	// 等待操作完成
	if err := op.Wait(); err != nil {
		RespondContainerJSON(w, 500, "Failed to wait for container creation: "+err.Error(), nil)
		return
	}

	// 保存到数据库
	container := models.Container{
		Hostname: req.Name,
		Status:   "Stopped",
		Image:    req.Image,
		CPUs:     req.CPU,
		Memory:   req.Memory,
		Disk:     req.Disk,
	}

	if err := models.DB.Create(&container).Error; err != nil {
		RespondContainerJSON(w, 500, "Failed to save container to database: "+err.Error(), nil)
		return
	}

	RespondContainerJSON(w, 200, "Container created successfully", container)
}

// HandleContainerAction 处理容器操作（启动、停止、重启、删除）
func HandleContainerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		RespondContainerJSON(w, 405, "Method not allowed", nil)
		return
	}

	var req struct {
		Name   string `json:"name"`
		Action string `json:"action"` // start, stop, restart, delete
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondContainerJSON(w, 400, "Invalid request: "+err.Error(), nil)
		return
	}

	if req.Name == "" {
		RespondContainerJSON(w, 400, "Container name is required", nil)
		return
	}

	// 获取LXD客户端
	client := lxd.GetClient()
	if client == nil {
		RespondContainerJSON(w, 500, "LXD client not available", nil)
		return
	}

	var err error

	switch req.Action {
	case "start":
		reqState := lxdapi.InstanceStatePut{
			Action:  "start",
			Timeout: -1,
		}
		_, err = client.UpdateInstanceState(req.Name, reqState, "")
	case "stop":
		reqState := lxdapi.InstanceStatePut{
			Action:  "stop",
			Timeout: -1,
			Force:   true,
		}
		_, err = client.UpdateInstanceState(req.Name, reqState, "")
	case "restart":
		reqState := lxdapi.InstanceStatePut{
			Action:  "restart",
			Timeout: -1,
			Force:   true,
		}
		_, err = client.UpdateInstanceState(req.Name, reqState, "")
	case "delete":
		// 先停止容器
		state, _, _ := client.GetInstanceState(req.Name)
		if state != nil && state.Status == "Running" {
			reqState := lxdapi.InstanceStatePut{
				Action:  "stop",
				Timeout: -1,
				Force:   true,
			}
			stopOp, _ := client.UpdateInstanceState(req.Name, reqState, "")
			if stopOp != nil {
				stopOp.Wait()
			}
		}
		// 删除容器
		delOp, err := client.DeleteInstance(req.Name)
		if err == nil && delOp != nil {
			if err := delOp.Wait(); err == nil {
				// 从数据库删除
				models.DB.Where("hostname = ?", req.Name).Delete(&models.Container{})
			}
		}
	default:
		RespondContainerJSON(w, 400, "Invalid action", nil)
		return
	}

	if err != nil {
		RespondContainerJSON(w, 500, "Failed to "+req.Action+" container: "+err.Error(), nil)
		return
	}

	// 更新数据库状态
	if req.Action != "delete" {
		var status string
		switch req.Action {
		case "start":
			status = "Running"
		case "stop":
			status = "Stopped"
		case "restart":
			status = "Running"
		}
		models.DB.Model(&models.Container{}).Where("hostname = ?", req.Name).Update("status", status)
	}

	RespondContainerJSON(w, 200, "Action completed successfully", nil)
}

// extractIPFromContainerState 从容器状态中提取IP地址
func extractIPFromContainerState(state *lxdapi.InstanceState) (string, string) {
	var ipv4, ipv6 string
	if eth0, ok := state.Network["eth0"]; ok {
		for _, addr := range eth0.Addresses {
			if addr.Family == "inet" && addr.Scope == "global" {
				ipv4 = addr.Address
			} else if addr.Family == "inet6" && addr.Scope == "global" {
				ipv6 = addr.Address
			}
		}
	}
	return ipv4, ipv6
}

// RespondContainerJSON 返回JSON响应
func RespondContainerJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
		"msg":  message,
		"data": data,
	})
}
