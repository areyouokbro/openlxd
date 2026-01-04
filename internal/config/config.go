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
		"./config.yaml",              // 当前目录（最高优先级）
		"configs/config.yaml",        // 开发环境路径
		"/etc/openlxd/config.yaml",  // 生产环境路径
		"/opt/openlxd/config.yaml",   // 备用路径
	}

	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &GlobalConfig); err != nil {
				return fmt.Errorf("配置文件解析失败 (%s): %v", path, err)
			}
			log.Printf("成功加载配置文件: %s", path)
			setDefaults()
			return nil
		}
	}

	// 未找到配置文件，使用默认配置并创建配置文件
	log.Println("未找到配置文件，使用默认配置")
	loadDefaultConfig()
	
	// 尝试创建默认配置文件
	if err := createDefaultConfigFile("./config.yaml"); err != nil {
		log.Printf("警告: 无法创建配置文件: %v", err)
	}
	
	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return &GlobalConfig
}


// loadDefaultConfig 加载默认配置
func loadDefaultConfig() {
	GlobalConfig.Server.Port = 8443
	GlobalConfig.Server.Host = "0.0.0.0"
	GlobalConfig.Server.HTTPS = false
	GlobalConfig.Server.Domain = "localhost"
	GlobalConfig.Server.CertDir = "./certs"
	GlobalConfig.Server.AutoTLS = false
	
	GlobalConfig.Security.APIHash = "default-api-key-please-change"
	GlobalConfig.Security.AdminUser = "admin"
	GlobalConfig.Security.AdminPass = "admin123"
	GlobalConfig.Security.SessionSecret = "default-secret-please-change"
	
	GlobalConfig.Database.Type = "sqlite"
	GlobalConfig.Database.Path = "./openlxd.db"
	
	GlobalConfig.LXD.Socket = "/var/snap/lxd/common/lxd/unix.socket"
	GlobalConfig.LXD.Bridge = "lxdbr0"
	
	log.Println("已加载默认配置")
}

// setDefaults 设置默认值
func setDefaults() {
	if GlobalConfig.Database.Path == "" {
		GlobalConfig.Database.Path = "./openlxd.db"
	}
	if GlobalConfig.LXD.Socket == "" {
		GlobalConfig.LXD.Socket = "/var/snap/lxd/common/lxd/unix.socket"
	}
	if GlobalConfig.LXD.Bridge == "" {
		GlobalConfig.LXD.Bridge = "lxdbr0"
	}
	if GlobalConfig.Server.CertDir == "" {
		GlobalConfig.Server.CertDir = "./certs"
	}
	if GlobalConfig.Server.Port == 0 {
		GlobalConfig.Server.Port = 8443
	}
	if GlobalConfig.Server.Host == "" {
		GlobalConfig.Server.Host = "0.0.0.0"
	}
}

// createDefaultConfigFile 创建默认配置文件
func createDefaultConfigFile(path string) error {
	configContent := `server:
  port: 8443
  host: "0.0.0.0"
  https: false
  domain: "localhost"
  cert_dir: "./certs"
  auto_tls: false

security:
  api_hash: "default-api-key-please-change"
  admin_user: "admin"
  admin_pass: "admin123"
  session_secret: "default-secret-please-change"

database:
  type: "sqlite"
  path: "./openlxd.db"

lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"
  bridge: "lxdbr0"
`
	
	if err := os.WriteFile(path, []byte(configContent), 0644); err != nil {
		return err
	}
	
	log.Printf("已创建默认配置文件: %s", path)
	return nil
}
