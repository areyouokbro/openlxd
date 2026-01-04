package models

import (
	"fmt"
	"log"
	
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dbType, dsn string) error {
	var err error
	var dialector gorm.Dialector
	
	switch dbType {
	case "sqlite", "":
		if dsn == "" {
			dsn = "lxdapi.db"
		}
		dialector = sqlite.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}
	
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	
	// 自动迁移表结构
	err = DB.AutoMigrate(
		&Container{},
		&IPPool{},
		&PortMapping{},
		&AuditLog{},
		&Config{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	log.Println("数据库初始化成功")
	return nil
}

// LogAction 记录审计日志
func LogAction(action, target, detail, status string) {
	log := AuditLog{
		Action: action,
		Target: target,
		Detail: detail,
		Status: status,
	}
	DB.Create(&log)
}

// GetConfig 获取配置项
func GetConfig(key string) string {
	var config Config
	DB.Where("key = ?", key).First(&config)
	return config.Value
}

// SetConfig 设置配置项
func SetConfig(key, value string) {
	config := Config{Key: key, Value: value}
	DB.Save(&config)
}
