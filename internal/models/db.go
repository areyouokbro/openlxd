package models

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		&User{},
		&Image{},
		&Container{},
		&ActionLog{},
		&NetworkConfig{},
		&IPAddress{},
		&PortMapping{},
		&ProxyConfig{},
		&Quota{},
		&SystemMetric{},
		&ContainerMetric{},
		&NetworkTraffic{},
		&MigrationTask{},
		&RemoteHost{},
		&MigrationLog{},
	)
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	log.Println("数据库初始化成功")
	return nil
}

// LogAction 记录操作日志
func LogAction(action, container, description, status string) {
	log := ActionLog{
		Action:      action,
		Container:   container,
		Description: description,
		Status:      status,
	}
	DB.Create(&log)
}

// GetContainerByHostname 根据主机名获取容器
func GetContainerByHostname(hostname string) (*Container, error) {
	var container Container
	err := DB.Where("hostname = ?", hostname).First(&container).Error
	return &container, err
}

// UpdateContainerStatus 更新容器状态
func UpdateContainerStatus(hostname, status string) error {
	return DB.Model(&Container{}).Where("hostname = ?", hostname).Update("status", status).Error
}

// UpdateContainerIP 更新容器IP地址
func UpdateContainerIP(hostname, ipv4, ipv6 string) error {
	updates := map[string]interface{}{}
	if ipv4 != "" {
		updates["ipv4"] = ipv4
	}
	if ipv6 != "" {
		updates["ipv6"] = ipv6
	}
	return DB.Model(&Container{}).Where("hostname = ?", hostname).Updates(updates).Error
}

// GetAllContainers 获取所有容器
func GetAllContainers() ([]Container, error) {
	var containers []Container
	err := DB.Find(&containers).Error
	return containers, err
}

// DeleteContainer 删除容器记录
func DeleteContainer(hostname string) error {
	return DB.Where("hostname = ?", hostname).Delete(&Container{}).Error
}
