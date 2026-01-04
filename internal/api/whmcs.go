package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/openlxd/backend/internal/auth"
	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
	"gorm.io/gorm"
)

// WHMCSAPI WHMCS兼容API
type WHMCSAPI struct {
	db        *gorm.DB
	lxdClient *lxd.ClientWrapper
}

// NewWHMCSAPI 创建WHMCS API实例
func NewWHMCSAPI(db *gorm.DB, lxdClient *lxd.ClientWrapper) *WHMCSAPI {
	return &WHMCSAPI{
		db:        db,
		lxdClient: lxdClient,
	}
}

// CreateContainerRequest 创建容器请求
type CreateContainerRequest struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	CPU      int    `json:"cpu"`
	Memory   string `json:"memory"` // 如 "2GB"
	Disk     string `json:"disk"`   // 如 "20GB"
	IPv4     string `json:"ipv4"`   // "auto" 或具体IP
	IPv6     string `json:"ipv6"`   // "auto" 或具体IP
	Password string `json:"password,omitempty"`
}

// ContainerActionRequest 容器操作请求
type ContainerActionRequest struct {
	Name string `json:"name"`
}

// ContainerConfigRequest 容器配置请求
type ContainerConfigRequest struct {
	Name   string `json:"name"`
	CPU    int    `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Disk   string `json:"disk,omitempty"`
}

// WHMCSResponse WHMCS响应格式
type WHMCSResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
}

// CreateContainer 创建容器（WHMCS）
func (api *WHMCSAPI) CreateContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Image == "" {
		api.respondError(w, "Name and image are required", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.CPU == 0 {
		req.CPU = 1
	}
	if req.Memory == "" {
		req.Memory = "1GB"
	}
	if req.Disk == "" {
		req.Disk = "10GB"
	}

	// 解析内存和磁盘大小
	memoryMB, err := parseSize(req.Memory)
	if err != nil {
		api.respondError(w, "Invalid memory size", http.StatusBadRequest)
		return
	}

	diskGB, err := parseSize(req.Disk)
	if err != nil {
		api.respondError(w, "Invalid disk size", http.StatusBadRequest)
		return
	}

	// 分配IP地址
	var ipv4, ipv6 string
	if req.IPv4 == "auto" || req.IPv4 == "" {
		// 自动分配IPv4
		var ip models.IPAddress
		if err := api.db.Where("type = ? AND status = ?", "ipv4", "available").First(&ip).Error; err == nil {
			ipv4 = ip.IP
			api.db.Model(&ip).Updates(map[string]interface{}{"status": "used"})
		}
	} else {
		ipv4 = req.IPv4
	}

	if req.IPv6 == "auto" || req.IPv6 == "" {
		// 自动分配IPv6
		var ip models.IPAddress
		if err := api.db.Where("type = ? AND status = ?", "ipv6", "available").First(&ip).Error; err == nil {
			ipv6 = ip.IP
			api.db.Model(&ip).Updates(map[string]interface{}{"status": "used"})
		}
	} else {
		ipv6 = req.IPv6
	}

	// 创建LXD容器
	config := lxd.ContainerConfig{
		Name:   req.Name,
		Image:  req.Image,
		CPU:    req.CPU,
		Memory: memoryMB,
		Disk:   diskGB,
	}

	if err := api.lxdClient.CreateContainer(config); err != nil {
		api.respondError(w, fmt.Sprintf("Failed to create container: %v", err), http.StatusInternalServerError)
		return
	}

	// 启动容器
	if err := api.lxdClient.StartContainer(req.Name); err != nil {
		api.respondError(w, fmt.Sprintf("Container created but failed to start: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置密码（如果提供）
	if req.Password != "" {
		if err := api.lxdClient.SetPassword(req.Name, "root", req.Password); err != nil {
			// 密码设置失败不影响容器创建
			fmt.Printf("Warning: Failed to set password: %v\n", err)
		}
	}

	// 保存到数据库
	container := models.Container{
		Hostname:  req.Name,
		Status:    "Running",
		Image:     req.Image,
		IPv4:      ipv4,
		IPv6:      ipv6,
		CPUs:      req.CPU,
		Memory:    memoryMB,
		Disk:      diskGB,
		UserID:    user.ID,
		CreatedBy: user.Username,
	}

	if err := api.db.Create(&container).Error; err != nil {
		api.respondError(w, "Failed to save container to database", http.StatusInternalServerError)
		return
	}

	// 更新IP地址关联
	if ipv4 != "" {
		api.db.Model(&models.IPAddress{}).Where("ip = ?", ipv4).Update("container_id", container.ID)
	}
	if ipv6 != "" {
		api.db.Model(&models.IPAddress{}).Where("ip = ?", ipv6).Update("container_id", container.ID)
	}

	api.respondSuccess(w, map[string]interface{}{
		"container_id": container.ID,
		"name":         container.Hostname,
		"status":       container.Status,
		"ipv4":         container.IPv4,
		"ipv6":         container.IPv6,
		"cpu":          container.CPUs,
		"memory":       fmt.Sprintf("%dMB", container.Memory),
		"disk":         fmt.Sprintf("%dGB", container.Disk),
	}, "Container created successfully")
}

// StartContainer 启动容器（WHMCS）
func (api *WHMCSAPI) StartContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ContainerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := api.db.Where("hostname = ?", req.Name).First(&container).Error; err != nil {
		api.respondError(w, "Container not found", http.StatusNotFound)
		return
	}

	if !user.IsAdmin() && container.UserID != user.ID {
		api.respondError(w, "Access denied", http.StatusForbidden)
		return
	}

	// 启动容器
	if err := api.lxdClient.StartContainer(req.Name); err != nil {
		api.respondError(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库
	api.db.Model(&container).Update("status", "Running")

	api.respondSuccess(w, map[string]interface{}{
		"name":   req.Name,
		"status": "Running",
	}, "Container started successfully")
}

// StopContainer 停止容器（WHMCS）
func (api *WHMCSAPI) StopContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ContainerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := api.db.Where("hostname = ?", req.Name).First(&container).Error; err != nil {
		api.respondError(w, "Container not found", http.StatusNotFound)
		return
	}

	if !user.IsAdmin() && container.UserID != user.ID {
		api.respondError(w, "Access denied", http.StatusForbidden)
		return
	}

	// 停止容器
	if err := api.lxdClient.StopContainer(req.Name); err != nil {
		api.respondError(w, fmt.Sprintf("Failed to stop container: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库
	api.db.Model(&container).Update("status", "Stopped")

	api.respondSuccess(w, map[string]interface{}{
		"name":   req.Name,
		"status": "Stopped",
	}, "Container stopped successfully")
}

// RestartContainer 重启容器（WHMCS）
func (api *WHMCSAPI) RestartContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ContainerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := api.db.Where("hostname = ?", req.Name).First(&container).Error; err != nil {
		api.respondError(w, "Container not found", http.StatusNotFound)
		return
	}

	if !user.IsAdmin() && container.UserID != user.ID {
		api.respondError(w, "Access denied", http.StatusForbidden)
		return
	}

	// 重启容器
	if err := api.lxdClient.RestartContainer(req.Name); err != nil {
		api.respondError(w, fmt.Sprintf("Failed to restart container: %v", err), http.StatusInternalServerError)
		return
	}

	api.respondSuccess(w, map[string]interface{}{
		"name":   req.Name,
		"status": "Running",
	}, "Container restarted successfully")
}

// DeleteContainer 删除容器（WHMCS）
func (api *WHMCSAPI) DeleteContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ContainerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := api.db.Where("hostname = ?", req.Name).First(&container).Error; err != nil {
		api.respondError(w, "Container not found", http.StatusNotFound)
		return
	}

	if !user.IsAdmin() && container.UserID != user.ID {
		api.respondError(w, "Access denied", http.StatusForbidden)
		return
	}

	// 删除LXD容器
	if err := api.lxdClient.DeleteContainer(req.Name); err != nil {
		api.respondError(w, fmt.Sprintf("Failed to delete container: %v", err), http.StatusInternalServerError)
		return
	}

	// 释放IP地址
	if container.IPv4 != "" {
		api.db.Model(&models.IPAddress{}).Where("ip = ?", container.IPv4).Updates(map[string]interface{}{
			"status":       "available",
			"container_id": 0,
		})
	}
	if container.IPv6 != "" {
		api.db.Model(&models.IPAddress{}).Where("ip = ?", container.IPv6).Updates(map[string]interface{}{
			"status":       "available",
			"container_id": 0,
		})
	}

	// 从数据库删除
	api.db.Delete(&container)

	api.respondSuccess(w, map[string]interface{}{
		"name": req.Name,
	}, "Container deleted successfully")
}

// GetContainerInfo 获取容器信息（WHMCS）
func (api *WHMCSAPI) GetContainerInfo(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		api.respondError(w, "Container name is required", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := api.db.Where("hostname = ?", name).First(&container).Error; err != nil {
		api.respondError(w, "Container not found", http.StatusNotFound)
		return
	}

	if !user.IsAdmin() && container.UserID != user.ID {
		api.respondError(w, "Access denied", http.StatusForbidden)
		return
	}

	// 获取LXD容器状态
	state, err := api.lxdClient.GetContainerState(name)
	if err == nil {
		container.Status = state.Status
	}

	api.respondSuccess(w, map[string]interface{}{
		"container_id": container.ID,
		"name":         container.Hostname,
		"status":       container.Status,
		"image":        container.Image,
		"ipv4":         container.IPv4,
		"ipv6":         container.IPv6,
		"cpu":          container.CPUs,
		"memory":       fmt.Sprintf("%dMB", container.Memory),
		"disk":         fmt.Sprintf("%dGB", container.Disk),
		"created_at":   container.CreatedAt,
	}, "Container info retrieved successfully")
}

// UpdateContainerConfig 更新容器配置（WHMCS）
func (api *WHMCSAPI) UpdateContainerConfig(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		api.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ContainerConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := api.db.Where("hostname = ?", req.Name).First(&container).Error; err != nil {
		api.respondError(w, "Container not found", http.StatusNotFound)
		return
	}

	if !user.IsAdmin() && container.UserID != user.ID {
		api.respondError(w, "Access denied", http.StatusForbidden)
		return
	}

	updates := make(map[string]interface{})

	// 更新CPU
	if req.CPU > 0 {
		if err := api.lxdClient.SetCPULimit(req.Name, req.CPU); err != nil {
			api.respondError(w, fmt.Sprintf("Failed to update CPU: %v", err), http.StatusInternalServerError)
			return
		}
		updates["cpus"] = req.CPU
	}

	// 更新内存
	if req.Memory != "" {
		memoryMB, err := parseSize(req.Memory)
		if err != nil {
			api.respondError(w, "Invalid memory size", http.StatusBadRequest)
			return
		}
		if err := api.lxdClient.SetMemoryLimit(req.Name, memoryMB); err != nil {
			api.respondError(w, fmt.Sprintf("Failed to update memory: %v", err), http.StatusInternalServerError)
			return
		}
		updates["memory"] = memoryMB
	}

	// 更新磁盘
	if req.Disk != "" {
		diskGB, err := parseSize(req.Disk)
		if err != nil {
			api.respondError(w, "Invalid disk size", http.StatusBadRequest)
			return
		}
		if err := api.lxdClient.SetDiskLimit(req.Name, diskGB); err != nil {
			api.respondError(w, fmt.Sprintf("Failed to update disk: %v", err), http.StatusInternalServerError)
			return
		}
		updates["disk"] = diskGB
	}

	// 更新数据库
	if len(updates) > 0 {
		api.db.Model(&container).Updates(updates)
	}

	api.respondSuccess(w, map[string]interface{}{
		"name":   req.Name,
		"cpu":    container.CPUs,
		"memory": fmt.Sprintf("%dMB", container.Memory),
		"disk":   fmt.Sprintf("%dGB", container.Disk),
	}, "Container configuration updated successfully")
}

// respondError 返回错误响应
func (api *WHMCSAPI) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(WHMCSResponse{
		Success: false,
		Message: message,
		Error:   message,
	})
}

// respondSuccess 返回成功响应
func (api *WHMCSAPI) respondSuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WHMCSResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// parseSize 解析大小字符串（如 "2GB", "1024MB"）
func parseSize(sizeStr string) (int, error) {
	var size int
	var unit string

	_, err := fmt.Sscanf(sizeStr, "%d%s", &size, &unit)
	if err != nil {
		// 尝试只解析数字
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			return 0, fmt.Errorf("invalid size format")
		}
		return size, nil
	}

	// 转换为标准单位
	switch unit {
	case "GB", "G", "gb", "g":
		return size, nil
	case "MB", "M", "mb", "m":
		return size, nil
	case "TB", "T", "tb", "t":
		return size * 1024, nil
	default:
		return size, nil
	}
}
