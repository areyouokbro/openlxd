package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/openlxd/backend/internal/quota"
)

// HandleQuota 处理配额管理请求
func HandleQuota(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetQuota(w, r)
	case http.MethodPost:
		handleCreateQuota(w, r)
	case http.MethodPut:
		handleUpdateQuota(w, r)
	case http.MethodDelete:
		handleDeleteQuota(w, r)
	default:
		respondJSON(w, 405, "Method not allowed", nil)
	}
}

// handleGetQuota 获取配额信息
func handleGetQuota(w http.ResponseWriter, r *http.Request) {
	containerIDStr := r.URL.Query().Get("container_id")
	
	if containerIDStr == "" {
		// 获取所有配额
		quotas, err := quota.GlobalQuotaManager.GetAllQuotas()
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("获取配额列表失败: %v", err), nil)
			return
		}
		respondJSON(w, 200, "获取成功", quotas)
		return
	}

	// 获取指定容器的配额
	containerID, err := strconv.ParseUint(containerIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "无效的容器ID", nil)
		return
	}

	quotaInfo, err := quota.GlobalQuotaManager.GetOrCreateQuota(uint(containerID))
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("获取配额失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "获取成功", quotaInfo)
}

// handleCreateQuota 创建配额
func handleCreateQuota(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContainerID      uint   `json:"container_id"`
		IPv4Quota        int    `json:"ipv4_quota"`
		IPv6Quota        int    `json:"ipv6_quota"`
		PortMappingQuota int    `json:"port_mapping_quota"`
		ProxyQuota       int    `json:"proxy_quota"`
		TrafficQuota     int64  `json:"traffic_quota"`
		OnExceed         string `json:"on_exceed"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondJSON(w, 400, "无效的请求数据", nil)
		return
	}

	// 创建配额
	quotaInfo, err := quota.GlobalQuotaManager.GetOrCreateQuota(req.ContainerID)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("创建配额失败: %v", err), nil)
		return
	}

	// 更新配额设置
	updates := map[string]interface{}{
		"ipv4_quota":         req.IPv4Quota,
		"ipv6_quota":         req.IPv6Quota,
		"port_mapping_quota": req.PortMappingQuota,
		"proxy_quota":        req.ProxyQuota,
		"traffic_quota":      req.TrafficQuota,
		"on_exceed":          req.OnExceed,
	}

	err = quota.GlobalQuotaManager.UpdateQuota(req.ContainerID, updates)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("更新配额失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "配额创建成功", quotaInfo)
}

// handleUpdateQuota 更新配额
func handleUpdateQuota(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContainerID      uint                   `json:"container_id"`
		Updates          map[string]interface{} `json:"updates"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondJSON(w, 400, "无效的请求数据", nil)
		return
	}

	err = quota.GlobalQuotaManager.UpdateQuota(req.ContainerID, req.Updates)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("更新配额失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "配额更新成功", nil)
}

// handleDeleteQuota 删除配额
func handleDeleteQuota(w http.ResponseWriter, r *http.Request) {
	containerIDStr := r.URL.Query().Get("container_id")
	if containerIDStr == "" {
		respondJSON(w, 400, "缺少容器ID", nil)
		return
	}

	containerID, err := strconv.ParseUint(containerIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "无效的容器ID", nil)
		return
	}

	err = quota.GlobalQuotaManager.DeleteQuota(uint(containerID))
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("删除配额失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "配额删除成功", nil)
}

// HandleQuotaUsage 处理配额使用情况请求
func HandleQuotaUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	containerIDStr := r.URL.Query().Get("container_id")
	if containerIDStr == "" {
		respondJSON(w, 400, "缺少容器ID", nil)
		return
	}

	containerID, err := strconv.ParseUint(containerIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "无效的容器ID", nil)
		return
	}

	usage, err := quota.GlobalQuotaManager.GetQuotaUsage(uint(containerID))
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("获取配额使用情况失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "获取成功", usage)
}

// HandleQuotaStats 处理配额统计请求
func HandleQuotaStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	stats := quota.GlobalQuotaManager.GetQuotaStats()
	respondJSON(w, 200, "获取成功", stats)
}

// HandleResetTraffic 处理流量重置请求
func HandleResetTraffic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	containerIDStr := r.URL.Query().Get("container_id")
	if containerIDStr == "" {
		respondJSON(w, 400, "缺少容器ID", nil)
		return
	}

	containerID, err := strconv.ParseUint(containerIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "无效的容器ID", nil)
		return
	}

	err = quota.GlobalQuotaManager.ResetTraffic(uint(containerID))
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("重置流量失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "流量重置成功", nil)
}


