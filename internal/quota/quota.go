package quota

import (
	"fmt"
	"sync"
	"time"

	"github.com/openlxd/backend/internal/models"
)

// QuotaManager 配额管理器
type QuotaManager struct {
	mu sync.RWMutex
}

var GlobalQuotaManager = &QuotaManager{}

// GetOrCreateQuota 获取或创建容器配额
func (q *QuotaManager) GetOrCreateQuota(containerID uint) (*models.Quota, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var quota models.Quota
	err := models.DB.Where("container_id = ?", containerID).First(&quota).Error
	if err != nil {
		// 创建默认配额
		quota = models.Quota{
			ContainerID:      containerID,
			IPv4Quota:        -1, // 无限制
			IPv6Quota:        -1,
			PortMappingQuota: -1,
			ProxyQuota:       -1,
			TrafficQuota:     -1,
			TrafficUsed:      0,
			TrafficResetDate: time.Now().AddDate(0, 1, 0), // 下个月重置
			OnExceed:         "warn",
		}
		models.DB.Create(&quota)
	}

	return &quota, nil
}

// UpdateQuota 更新容器配额
func (q *QuotaManager) UpdateQuota(containerID uint, updates map[string]interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	var quota models.Quota
	err := models.DB.Where("container_id = ?", containerID).First(&quota).Error
	if err != nil {
		return fmt.Errorf("配额不存在")
	}

	err = models.DB.Model(&quota).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("更新配额失败: %v", err)
	}

	return nil
}

// GetQuotaUsage 获取配额使用情况
func (q *QuotaManager) GetQuotaUsage(containerID uint) (*models.QuotaUsage, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// 获取配额设置
	quota, err := q.GetOrCreateQuota(containerID)
	if err != nil {
		return nil, err
	}

	// 统计实际使用情况
	var ipv4Used, ipv6Used int64
	models.DB.Model(&models.IPAddress{}).
		Where("container_id = ? AND type = ? AND status = ?", containerID, "ipv4", "used").
		Count(&ipv4Used)
	models.DB.Model(&models.IPAddress{}).
		Where("container_id = ? AND type = ? AND status = ?", containerID, "ipv6", "used").
		Count(&ipv6Used)

	var portMappingUsed int64
	models.DB.Model(&models.PortMapping{}).
		Where("container_id = ? AND status = ?", containerID, "active").
		Count(&portMappingUsed)

	var proxyUsed int64
	models.DB.Model(&models.ProxyConfig{}).
		Where("container_id = ? AND status = ?", containerID, "active").
		Count(&proxyUsed)

	usage := &models.QuotaUsage{
		ContainerID:      containerID,
		IPv4Used:         int(ipv4Used),
		IPv6Used:         int(ipv6Used),
		PortMappingUsed:  int(portMappingUsed),
		ProxyUsed:        int(proxyUsed),
		TrafficUsed:      quota.TrafficUsed,
		IPv4Quota:        quota.IPv4Quota,
		IPv6Quota:        quota.IPv6Quota,
		PortMappingQuota: quota.PortMappingQuota,
		ProxyQuota:       quota.ProxyQuota,
		TrafficQuota:     quota.TrafficQuota,
	}

	return usage, nil
}

// CheckIPv4Quota 检查 IPv4 配额
func (q *QuotaManager) CheckIPv4Quota(containerID uint) error {
	usage, err := q.GetQuotaUsage(containerID)
	if err != nil {
		return err
	}

	if usage.IPv4Quota == -1 {
		return nil // 无限制
	}

	if usage.IPv4Used >= usage.IPv4Quota {
		return fmt.Errorf("IPv4 地址配额已用完 (%d/%d)", usage.IPv4Used, usage.IPv4Quota)
	}

	return nil
}

// CheckIPv6Quota 检查 IPv6 配额
func (q *QuotaManager) CheckIPv6Quota(containerID uint) error {
	usage, err := q.GetQuotaUsage(containerID)
	if err != nil {
		return err
	}

	if usage.IPv6Quota == -1 {
		return nil // 无限制
	}

	if usage.IPv6Used >= usage.IPv6Quota {
		return fmt.Errorf("IPv6 地址配额已用完 (%d/%d)", usage.IPv6Used, usage.IPv6Quota)
	}

	return nil
}

// CheckPortMappingQuota 检查端口映射配额
func (q *QuotaManager) CheckPortMappingQuota(containerID uint, count int) error {
	usage, err := q.GetQuotaUsage(containerID)
	if err != nil {
		return err
	}

	if usage.PortMappingQuota == -1 {
		return nil // 无限制
	}

	if usage.PortMappingUsed+count > usage.PortMappingQuota {
		return fmt.Errorf("端口映射配额不足 (当前: %d, 需要: %d, 配额: %d)", 
			usage.PortMappingUsed, count, usage.PortMappingQuota)
	}

	return nil
}

// CheckProxyQuota 检查反向代理配额
func (q *QuotaManager) CheckProxyQuota(containerID uint) error {
	usage, err := q.GetQuotaUsage(containerID)
	if err != nil {
		return err
	}

	if usage.ProxyQuota == -1 {
		return nil // 无限制
	}

	if usage.ProxyUsed >= usage.ProxyQuota {
		return fmt.Errorf("反向代理配额已用完 (%d/%d)", usage.ProxyUsed, usage.ProxyQuota)
	}

	return nil
}

// CheckTrafficQuota 检查流量配额
func (q *QuotaManager) CheckTrafficQuota(containerID uint) error {
	usage, err := q.GetQuotaUsage(containerID)
	if err != nil {
		return err
	}

	if usage.TrafficQuota == -1 {
		return nil // 无限制
	}

	if usage.TrafficUsed >= usage.TrafficQuota {
		return fmt.Errorf("流量配额已用完 (%d GB/%d GB)", 
			usage.TrafficUsed, usage.TrafficQuota)
	}

	return nil
}

// AddTrafficUsage 增加流量使用量
func (q *QuotaManager) AddTrafficUsage(containerID uint, bytes int64) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	var quota models.Quota
	err := models.DB.Where("container_id = ?", containerID).First(&quota).Error
	if err != nil {
		return err
	}

	// 转换为 GB
	gb := bytes / (1024 * 1024 * 1024)
	quota.TrafficUsed += gb

	// 检查是否需要重置流量
	if time.Now().After(quota.TrafficResetDate) {
		quota.TrafficUsed = 0
		quota.TrafficResetDate = time.Now().AddDate(0, 1, 0)
	}

	models.DB.Save(&quota)

	// 检查是否超限
	if quota.TrafficQuota != -1 && quota.TrafficUsed >= quota.TrafficQuota {
		q.handleQuotaExceed(containerID, "traffic")
	}

	return nil
}

// ResetTraffic 重置流量统计
func (q *QuotaManager) ResetTraffic(containerID uint) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	return models.DB.Model(&models.Quota{}).
		Where("container_id = ?", containerID).
		Updates(map[string]interface{}{
			"traffic_used":       0,
			"traffic_reset_date": time.Now().AddDate(0, 1, 0),
		}).Error
}

// handleQuotaExceed 处理配额超限
func (q *QuotaManager) handleQuotaExceed(containerID uint, quotaType string) {
	var quota models.Quota
	err := models.DB.Where("container_id = ?", containerID).First(&quota).Error
	if err != nil {
		return
	}

	switch quota.OnExceed {
	case "warn":
		// 记录警告日志
		models.LogAction("quota_exceed", "", 
			fmt.Sprintf("容器 %d 的 %s 配额已超限", containerID, quotaType), "warning")
	
	case "limit":
		// 限制新的资源分配（在检查函数中已经实现）
		models.LogAction("quota_limit", "", 
			fmt.Sprintf("容器 %d 的 %s 配额已达上限，限制新分配", containerID, quotaType), "warning")
	
	case "stop":
		// 停止容器
		models.LogAction("quota_stop", "", 
			fmt.Sprintf("容器 %d 的 %s 配额已超限，自动停止容器", containerID, quotaType), "error")
		// TODO: 调用 LXD API 停止容器
	}
}

// GetAllQuotas 获取所有配额
func (q *QuotaManager) GetAllQuotas() ([]models.Quota, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var quotas []models.Quota
	err := models.DB.Find(&quotas).Error
	return quotas, err
}

// DeleteQuota 删除配额
func (q *QuotaManager) DeleteQuota(containerID uint) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	return models.DB.Where("container_id = ?", containerID).Delete(&models.Quota{}).Error
}

// SetDefaultQuota 设置默认配额
func (q *QuotaManager) SetDefaultQuota(ipv4, ipv6, portMapping, proxy int, traffic int64) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 更新所有未设置配额的容器
	return models.DB.Model(&models.Quota{}).
		Where("ipv4_quota = -1 AND ipv6_quota = -1").
		Updates(map[string]interface{}{
			"ipv4_quota":         ipv4,
			"ipv6_quota":         ipv6,
			"port_mapping_quota": portMapping,
			"proxy_quota":        proxy,
			"traffic_quota":      traffic,
		}).Error
}

// GetQuotaStats 获取配额统计信息
func (q *QuotaManager) GetQuotaStats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var totalQuotas int64
	models.DB.Model(&models.Quota{}).Count(&totalQuotas)

	var exceedCount int64
	models.DB.Model(&models.Quota{}).
		Where("traffic_quota != -1 AND traffic_used >= traffic_quota").
		Count(&exceedCount)

	return map[string]interface{}{
		"total_quotas":  totalQuotas,
		"exceed_count":  exceedCount,
		"warning_count": 0, // TODO: 实现警告计数
	}
}
