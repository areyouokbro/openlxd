package lxd

import (
	"log"
	"time"
	
	"github.com/openlxd/backend/internal/models"
)

// TrafficMonitor 流量监控器
type TrafficMonitor struct {
	interval time.Duration
	stopChan chan bool
}

// NewTrafficMonitor 创建流量监控器
func NewTrafficMonitor(intervalSeconds int) *TrafficMonitor {
	return &TrafficMonitor{
		interval: time.Duration(intervalSeconds) * time.Second,
		stopChan: make(chan bool),
	}
}

// Start 启动流量监控
func (tm *TrafficMonitor) Start() {
	go func() {
		ticker := time.NewTicker(tm.interval)
		defer ticker.Stop()
		
		log.Printf("流量监控已启动，采集间隔: %v", tm.interval)
		
		for {
			select {
			case <-ticker.C:
				tm.collectTraffic()
			case <-tm.stopChan:
				log.Println("流量监控已停止")
				return
			}
		}
	}()
}

// Stop 停止流量监控
func (tm *TrafficMonitor) Stop() {
	tm.stopChan <- true
}

// collectTraffic 采集所有容器的流量数据
func (tm *TrafficMonitor) collectTraffic() {
	var containers []models.Container
	models.DB.Where("status = ?", "Running").Find(&containers)
	
	if len(containers) == 0 {
		return
	}
	
	log.Printf("采集流量数据，运行中容器数: %d", len(containers))
	
	for _, container := range containers {
		traffic := tm.getContainerTraffic(container.Hostname)
		if traffic > 0 {
			// 更新数据库中的流量使用量
			models.DB.Model(&models.Container{}).
				Where("hostname = ?", container.Hostname).
				Update("traffic_used", traffic)
			
			// 检查是否超过流量限制
			if container.TrafficLimit > 0 && traffic >= container.TrafficLimit {
				log.Printf("容器 %s 流量已达上限，自动停止", container.Hostname)
				StopContainer(container.Hostname)
				models.DB.Model(&models.Container{}).
					Where("hostname = ?", container.Hostname).
					Update("status", "Stopped")
				models.LogAction("auto_stop", container.Hostname, "流量超限自动停止", "success")
			}
		}
	}
}

// getContainerTraffic 获取容器的流量使用量（字节）
func (tm *TrafficMonitor) getContainerTraffic(name string) int64 {
	if !IsLXDAvailable() {
		// Mock 模式：返回随机增长的流量
		var container models.Container
		models.DB.Where("hostname = ?", name).First(&container)
		return container.TrafficUsed + 1024*1024*10 // 每次增加 10MB
	}
	
	state, err := GetContainerState(name)
	if err != nil {
		return 0
	}
	
	var totalBytes int64 = 0
	
	// 从网络接口统计流量
	if network, ok := state["network"].(map[string]interface{}); ok {
		for _, iface := range network {
			if ifaceData, ok := iface.(map[string]interface{}); ok {
				if counters, ok := ifaceData["counters"].(map[string]interface{}); ok {
					if bytesReceived, ok := counters["bytes_received"].(float64); ok {
						totalBytes += int64(bytesReceived)
					}
					if bytesSent, ok := counters["bytes_sent"].(float64); ok {
						totalBytes += int64(bytesSent)
					}
				}
			}
		}
	}
	
	return totalBytes
}
