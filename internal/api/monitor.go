package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/openlxd/backend/internal/models"
	"github.com/openlxd/backend/internal/monitor"
)

// HandleSystemMetrics 处理系统监控指标请求
func HandleSystemMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 获取查询参数
	hoursStr := r.URL.Query().Get("hours")
	hours := 1 // 默认1小时
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	// 查询最近N小时的数据
	var metrics []models.SystemMetric
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	err := models.DB.Where("timestamp >= ?", startTime).
		Order("timestamp ASC").
		Find(&metrics).Error

	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("查询失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "获取成功", metrics)
}

// HandleCurrentSystemMetrics 处理当前系统监控指标请求
func HandleCurrentSystemMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 实时采集系统指标
	metric, err := monitor.GlobalCollector.CollectSystemMetrics()
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("采集失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "获取成功", metric)
}

// HandleContainerMetrics 处理容器监控指标请求
func HandleContainerMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 获取查询参数
	containerIDStr := r.URL.Query().Get("container_id")
	hoursStr := r.URL.Query().Get("hours")
	
	hours := 1 // 默认1小时
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	// 查询最近N小时的数据
	var metrics []models.ContainerMetric
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	query := models.DB.Where("timestamp >= ?", startTime)
	if containerIDStr != "" {
		containerID, err := strconv.ParseUint(containerIDStr, 10, 32)
		if err == nil {
			query = query.Where("container_id = ?", containerID)
		}
	}
	
	err := query.Order("timestamp ASC").Find(&metrics).Error
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("查询失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "获取成功", metrics)
}

// HandleNetworkTraffic 处理网络流量请求
func HandleNetworkTraffic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 获取查询参数
	containerIDStr := r.URL.Query().Get("container_id")
	hoursStr := r.URL.Query().Get("hours")
	
	hours := 24 // 默认24小时
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	// 查询最近N小时的数据
	var traffic []models.NetworkTraffic
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	query := models.DB.Where("timestamp >= ?", startTime)
	if containerIDStr != "" {
		containerID, err := strconv.ParseUint(containerIDStr, 10, 32)
		if err == nil {
			query = query.Where("container_id = ?", containerID)
		}
	}
	
	err := query.Order("timestamp ASC").Find(&traffic).Error
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("查询失败: %v", err), nil)
		return
	}

	respondJSON(w, 200, "获取成功", traffic)
}

// HandleResourceStats 处理资源统计请求
func HandleResourceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 获取所有容器
	var containers []models.Container
	err := models.DB.Find(&containers).Error
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("查询容器失败: %v", err), nil)
		return
	}

	// 统计每个容器的资源使用情况
	var stats []map[string]interface{}
	
	for _, container := range containers {
		// 查询最近24小时的指标
		var metrics []models.ContainerMetric
		startTime := time.Now().Add(-24 * time.Hour)
		err := models.DB.Where("container_id = ? AND timestamp >= ?", container.ID, startTime).
			Find(&metrics).Error
		
		if err != nil || len(metrics) == 0 {
			continue
		}

		// 计算平均值和最大值
		var cpuSum, memSum, diskSum float64
		var cpuMax, memMax, diskMax float64
		var rxTotal, txTotal int64

		for _, m := range metrics {
			cpuSum += m.CPUUsage
			memSum += m.MemoryUsage
			diskSum += m.DiskUsage
			
			if m.CPUUsage > cpuMax {
				cpuMax = m.CPUUsage
			}
			if m.MemoryUsage > memMax {
				memMax = m.MemoryUsage
			}
			if m.DiskUsage > diskMax {
				diskMax = m.DiskUsage
			}
			
			rxTotal = m.NetworkRxTotal
			txTotal = m.NetworkTxTotal
		}

		count := float64(len(metrics))
		
		// 查询配额使用情况
		var ipv4Count, ipv6Count, portMappingCount, proxyCount int64
		models.DB.Model(&models.IPAddress{}).Where("container_id = ? AND type = ?", container.ID, "ipv4").Count(&ipv4Count)
		models.DB.Model(&models.IPAddress{}).Where("container_id = ? AND type = ?", container.ID, "ipv6").Count(&ipv6Count)
		models.DB.Model(&models.PortMapping{}).Where("container_id = ?", container.ID).Count(&portMappingCount)
		models.DB.Model(&models.ProxyConfig{}).Where("container_id = ?", container.ID).Count(&proxyCount)

		stat := map[string]interface{}{
			"container_id":        container.ID,
			"container_name":      container.Hostname,
			"cpu_usage_avg":       cpuSum / count,
			"cpu_usage_max":       cpuMax,
			"memory_usage_avg":    memSum / count,
			"memory_usage_max":    memMax,
			"disk_usage_avg":      diskSum / count,
			"disk_usage_max":      diskMax,
			"network_rx_total":    rxTotal,
			"network_tx_total":    txTotal,
			"ipv4_count":          ipv4Count,
			"ipv6_count":          ipv6Count,
			"port_mapping_count":  portMappingCount,
			"proxy_count":         proxyCount,
		}
		
		stats = append(stats, stat)
	}

	respondJSON(w, 200, "获取成功", stats)
}

// HandleMonitorDashboard 处理监控仪表板数据请求
func HandleMonitorDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, 405, "Method not allowed", nil)
		return
	}

	// 获取当前系统指标
	systemMetric, _ := monitor.GlobalCollector.CollectSystemMetrics()

	// 获取容器数量统计
	var totalContainers, runningContainers int64
	models.DB.Model(&models.Container{}).Count(&totalContainers)
	models.DB.Model(&models.Container{}).Where("status = ?", "Running").Count(&runningContainers)

	// 获取网络资源统计
	var totalIPv4, usedIPv4, totalIPv6, usedIPv6 int64
	models.DB.Model(&models.IPAddress{}).Where("type = ?", "ipv4").Count(&totalIPv4)
	models.DB.Model(&models.IPAddress{}).Where("type = ? AND status = ?", "ipv4", "used").Count(&usedIPv4)
	models.DB.Model(&models.IPAddress{}).Where("type = ?", "ipv6").Count(&totalIPv6)
	models.DB.Model(&models.IPAddress{}).Where("type = ? AND status = ?", "ipv6", "used").Count(&usedIPv6)

	// 获取端口映射和反向代理统计
	var totalPortMappings, totalProxies int64
	models.DB.Model(&models.PortMapping{}).Count(&totalPortMappings)
	models.DB.Model(&models.ProxyConfig{}).Count(&totalProxies)

	// 获取最近24小时的系统指标
	var recentMetrics []models.SystemMetric
	startTime := time.Now().Add(-24 * time.Hour)
	models.DB.Where("timestamp >= ?", startTime).
		Order("timestamp DESC").
		Limit(288). // 每5分钟一个点，24小时=288个点
		Find(&recentMetrics)

	dashboard := map[string]interface{}{
		"current_system": systemMetric,
		"containers": map[string]interface{}{
			"total":   totalContainers,
			"running": runningContainers,
			"stopped": totalContainers - runningContainers,
		},
		"network": map[string]interface{}{
			"ipv4_total":       totalIPv4,
			"ipv4_used":        usedIPv4,
			"ipv4_available":   totalIPv4 - usedIPv4,
			"ipv6_total":       totalIPv6,
			"ipv6_used":        usedIPv6,
			"ipv6_available":   totalIPv6 - usedIPv6,
			"port_mappings":    totalPortMappings,
			"proxies":          totalProxies,
		},
		"recent_metrics": recentMetrics,
	}

	respondJSON(w, 200, "获取成功", dashboard)
}
