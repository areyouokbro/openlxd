package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Port       int    `yaml:"port"`
		Host       string `yaml:"host"`
		HTTPS      bool   `yaml:"https"`
		Domain     string `yaml:"domain"`
		CertDir    string `yaml:"cert_dir"`
		AutoTLS    bool   `yaml:"auto_tls"`
	} `yaml:"server"`
	Security struct {
		APIHash       string `yaml:"api_hash"`
		AdminUser     string `yaml:"admin_user"`
		AdminPass     string `yaml:"admin_pass"`
		SessionSecret string `yaml:"session_secret"`
	} `yaml:"security"`
	Database struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"database"`
	LXD struct {
		Socket string `yaml:"socket"`
		Bridge string `yaml:"bridge"`
	} `yaml:"lxd"`
}

var GlobalConfig Config

// LoadConfig 加载配置文件
func LoadConfig() error {
	// 按优先级尝试多个配置文件路径
	configPaths := []string{
		"/etc/openlxd/config.yaml",  // 生产环境路径
		"configs/config.yaml",        // 开发环境路径
		"./config.yaml",              // 当前目录
		"/opt/openlxd/config.yaml",   // 备用路径
	}

	var lastErr error
	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &GlobalConfig); err != nil {
				return fmt.Errorf("配置文件解析失败 (%s): %v", path, err)
			}
			log.Printf("成功加载配置文件: %s", path)
			
			// 设置默认值
			if GlobalConfig.Database.Path == "" {
				GlobalConfig.Database.Path = "/var/lib/openlxd/openlxd.db"
			}
			if GlobalConfig.LXD.Socket == "" {
				GlobalConfig.LXD.Socket = "/var/snap/lxd/common/lxd/unix.socket"
			}
			if GlobalConfig.LXD.Bridge == "" {
				GlobalConfig.LXD.Bridge = "lxdbr0"
			}
			if GlobalConfig.Server.CertDir == "" {
				GlobalConfig.Server.CertDir = "/etc/openlxd/certs"
			}
			
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("未找到配置文件，已尝试路径: %v\n最后错误: %v", configPaths, lastErr)
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return &GlobalConfig
}
