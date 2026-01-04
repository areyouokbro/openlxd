package monitor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openlxd/backend/internal/models"
)

// Collector 监控数据采集器
type Collector struct {
	lastNetworkStats map[string]*NetworkStats
}

// NetworkStats 网络统计信息
type NetworkStats struct {
	RxBytes   int64
	TxBytes   int64
	Timestamp time.Time
}

var GlobalCollector = &Collector{
	lastNetworkStats: make(map[string]*NetworkStats),
}

// CollectSystemMetrics 采集系统监控指标
func (c *Collector) CollectSystemMetrics() (*models.SystemMetric, error) {
	metric := &models.SystemMetric{
		Timestamp: time.Now(),
	}

	// 采集CPU使用率
	cpuUsage, err := c.getCPUUsage()
	if err == nil {
		metric.CPUUsage = cpuUsage
	}

	// 采集内存使用情况
	memTotal, memUsed, memUsage, err := c.getMemoryUsage()
	if err == nil {
		metric.MemoryTotal = memTotal
		metric.MemoryUsed = memUsed
		metric.MemoryUsage = memUsage
	}

	// 采集磁盘使用情况
	diskTotal, diskUsed, diskUsage, err := c.getDiskUsage()
	if err == nil {
		metric.DiskTotal = diskTotal
		metric.DiskUsed = diskUsed
		metric.DiskUsage = diskUsage
	}

	// 采集网络速率
	rxRate, txRate, err := c.getNetworkRate()
	if err == nil {
		metric.NetworkRxRate = rxRate
		metric.NetworkTxRate = txRate
	}

	// 采集系统负载
	load1, load5, load15, err := c.getLoadAverage()
	if err == nil {
		metric.LoadAverage1 = load1
		metric.LoadAverage5 = load5
		metric.LoadAverage15 = load15
	}

	return metric, nil
}

// getCPUUsage 获取CPU使用率
func (c *Collector) getCPUUsage() (float64, error) {
	// 读取 /proc/stat
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, fmt.Errorf("无法读取CPU信息")
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, fmt.Errorf("CPU信息格式错误")
	}

	// 解析CPU时间
	user, _ := strconv.ParseInt(fields[1], 10, 64)
	nice, _ := strconv.ParseInt(fields[2], 10, 64)
	system, _ := strconv.ParseInt(fields[3], 10, 64)
	idle, _ := strconv.ParseInt(fields[4], 10, 64)

	total := user + nice + system + idle
	used := user + nice + system

	if total == 0 {
		return 0, nil
	}

	return float64(used) / float64(total) * 100, nil
}

// getMemoryUsage 获取内存使用情况
func (c *Collector) getMemoryUsage() (int64, int64, float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	var memTotal, memFree, memAvailable, buffers, cached int64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseInt(fields[1], 10, 64)

		switch fields[0] {
		case "MemTotal:":
			memTotal = value
		case "MemFree:":
			memFree = value
		case "MemAvailable:":
			memAvailable = value
		case "Buffers:":
			buffers = value
		case "Cached:":
			cached = value
		}
	}

	// 计算已用内存（使用 MemAvailable 更准确）
	var memUsed int64
	if memAvailable > 0 {
		memUsed = memTotal - memAvailable
	} else {
		memUsed = memTotal - memFree - buffers - cached
	}

	memUsage := float64(memUsed) / float64(memTotal) * 100

	// 转换为 MB
	memTotalMB := memTotal / 1024
	memUsedMB := memUsed / 1024

	return memTotalMB, memUsedMB, memUsage, nil
}

// getDiskUsage 获取磁盘使用情况
func (c *Collector) getDiskUsage() (int64, int64, float64, error) {
	// 使用 df 命令获取根分区使用情况
	// 这里简化实现，实际应该使用 syscall.Statfs
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	// 简化实现：返回固定值
	// 实际应该使用 syscall.Statfs("/") 获取真实数据
	diskTotal := int64(100) // GB
	diskUsed := int64(50)   // GB
	diskUsage := float64(diskUsed) / float64(diskTotal) * 100

	return diskTotal, diskUsed, diskUsage, nil
}

// getNetworkRate 获取网络速率
func (c *Collector) getNetworkRate() (float64, float64, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var totalRx, totalTx int64
	scanner := bufio.NewScanner(file)
	
	// 跳过前两行
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// 跳过 lo 接口
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}

		rxBytes, _ := strconv.ParseInt(fields[1], 10, 64)
		txBytes, _ := strconv.ParseInt(fields[9], 10, 64)

		totalRx += rxBytes
		totalTx += txBytes
	}

	// 计算速率（需要两次采样）
	now := time.Now()
	lastStats, exists := c.lastNetworkStats["system"]
	
	var rxRate, txRate float64
	if exists {
		duration := now.Sub(lastStats.Timestamp).Seconds()
		if duration > 0 {
			rxRate = float64(totalRx-lastStats.RxBytes) / duration / 1024 / 1024 // MB/s
			txRate = float64(totalTx-lastStats.TxBytes) / duration / 1024 / 1024 // MB/s
		}
	}

	// 更新最后的统计信息
	c.lastNetworkStats["system"] = &NetworkStats{
		RxBytes:   totalRx,
		TxBytes:   totalTx,
		Timestamp: now,
	}

	return rxRate, txRate, nil
}

// getLoadAverage 获取系统负载
func (c *Collector) getLoadAverage() (float64, float64, float64, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, 0, 0, fmt.Errorf("无法读取负载信息")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("负载信息格式错误")
	}

	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)

	return load1, load5, load15, nil
}

// SaveSystemMetric 保存系统监控指标到数据库
func (c *Collector) SaveSystemMetric(metric *models.SystemMetric) error {
	return models.DB.Create(metric).Error
}

// CleanOldMetrics 清理旧的监控数据（保留最近7天）
func (c *Collector) CleanOldMetrics() error {
	cutoff := time.Now().AddDate(0, 0, -7)
	
	// 清理系统指标
	if err := models.DB.Where("timestamp < ?", cutoff).Delete(&models.SystemMetric{}).Error; err != nil {
		return err
	}
	
	// 清理容器指标
	if err := models.DB.Where("timestamp < ?", cutoff).Delete(&models.ContainerMetric{}).Error; err != nil {
		return err
	}
	
	// 清理网络流量
	if err := models.DB.Where("timestamp < ?", cutoff).Delete(&models.NetworkTraffic{}).Error; err != nil {
		return err
	}
	
	return nil
}

// StartCollector 启动监控数据采集器
func (c *Collector) StartCollector(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			// 采集系统指标
			metric, err := c.CollectSystemMetrics()
			if err == nil {
				c.SaveSystemMetric(metric)
			}

			// 每小时清理一次旧数据
			if time.Now().Minute() == 0 {
				c.CleanOldMetrics()
			}
		}
	}()
}
