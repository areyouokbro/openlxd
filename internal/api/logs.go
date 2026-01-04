package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/openlxd/backend/internal/models"
)

// GetContainerLogs 获取容器操作日志
func GetContainerLogs(w http.ResponseWriter, r *http.Request) {
	containerName := r.URL.Query().Get("container")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	var logs []models.OperationLog
	query := models.DB.Order("created_at DESC").Limit(limit)
	
	if containerName != "" {
		query = query.Where("container_name = ?", containerName)
	}
	
	if err := query.Find(&logs).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, "error", map[string]interface{}{
			"error": "获取日志失败: " + err.Error(),
		})
		return
	}
	
	respondJSON(w, http.StatusOK, "success", map[string]interface{}{
		"logs": logs,
	})
}

// GetSystemLogs 获取系统日志
func GetSystemLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	levelStr := r.URL.Query().Get("level")
	
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// 这里可以从数据库或日志文件中读取系统日志
	// 暂时返回操作日志作为系统日志
	var logs []models.OperationLog
	query := models.DB.Order("created_at DESC").Limit(limit)
	
	if levelStr != "" {
		query = query.Where("status = ?", levelStr)
	}
	
	if err := query.Find(&logs).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, "error", map[string]interface{}{
			"error": "获取系统日志失败: " + err.Error(),
		})
		return
	}
	
	respondJSON(w, http.StatusOK, "success", map[string]interface{}{
		"logs": logs,
	})
}

// GetContainerDetail 获取容器详细信息
func GetContainerDetail(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		respondJSON(w, http.StatusBadRequest, "error", map[string]interface{}{
			"error": "容器名称不能为空",
		})
		return
	}
	
	// 从数据库获取容器信息
	var container models.Container
	if err := models.DB.Where("name = ?", name).First(&container).Error; err != nil {
		respondJSON(w, http.StatusNotFound, "error", map[string]interface{}{
			"error": "容器不存在",
		})
		return
	}
	
	// 获取容器的配额信息
	var quota models.Quota
	models.DB.Where("container_name = ?", name).First(&quota)
	
	// 获取容器的网络信息
	var ipAddresses []models.IPAddress
	models.DB.Where("container_name = ?", name).Find(&ipAddresses)
	
	var portMappings []models.PortMapping
	models.DB.Where("container_name = ?", name).Find(&portMappings)
	
	var proxyConfigs []models.ProxyConfig
	models.DB.Where("container_name = ?", name).Find(&proxyConfigs)
	
	// 获取容器的最近操作日志
	var logs []models.OperationLog
	models.DB.Where("container_name = ?", name).Order("created_at DESC").Limit(10).Find(&logs)
	
	respondJSON(w, http.StatusOK, "success", map[string]interface{}{
		"container":     container,
		"quota":         quota,
		"ip_addresses":  ipAddresses,
		"port_mappings": portMappings,
		"proxy_configs": proxyConfigs,
		"recent_logs":   logs,
	})
}

// GetContainerStats 获取容器资源使用统计
func GetContainerStats(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		respondJSON(w, http.StatusBadRequest, "error", map[string]interface{}{
			"error": "容器名称不能为空",
		})
		return
	}
	
	// 获取最近1小时的容器监控数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	var metrics []models.ContainerMetric
	if err := models.DB.Where("container_name = ? AND timestamp >= ?", name, oneHourAgo).
		Order("timestamp ASC").Find(&metrics).Error; err != nil {
		respondJSON(w, http.StatusInternalServerError, "error", map[string]interface{}{
			"error": "获取统计数据失败: " + err.Error(),
		})
		return
	}
	
	respondJSON(w, http.StatusOK, "success", map[string]interface{}{
		"metrics": metrics,
	})
}

