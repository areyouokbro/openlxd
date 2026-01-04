package network

import (
	"fmt"
	"os"
	"path/filepath"
"os/exec"
	"sync"
	"text/template"

	"github.com/openlxd/backend/internal/models"
)

// ProxyManager 反向代理管理器
type ProxyManager struct {
	mu         sync.RWMutex
	nginxDir   string
	nginxBin   string
}

var GlobalProxyManager = &ProxyManager{
	nginxDir: "/etc/nginx/sites-available",
	nginxBin: "/usr/sbin/nginx",
}

// ProxyConfig 反向代理配置
type ProxyConfig struct {
	ID          uint   `json:"id"`
	ContainerID uint   `json:"container_id"`
	Domain      string `json:"domain"`
	TargetIP    string `json:"target_ip"`
	TargetPort  int    `json:"target_port"`
	SSL         bool   `json:"ssl"`
	CertPath    string `json:"cert_path"`
	KeyPath     string `json:"key_path"`
	Status      string `json:"status"` // active, inactive
}

// nginxConfigTemplate Nginx 配置模板
const nginxConfigTemplate = `server {
    listen 80;
    server_name {{ .Domain }};
    
    {{- if .SSL }}
    listen 443 ssl http2;
    ssl_certificate {{ .CertPath }};
    ssl_certificate_key {{ .KeyPath }};
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    {{- end }}
    
    location / {
        proxy_pass http://{{ .TargetIP }}:{{ .TargetPort }};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # 访问日志
    access_log /var/log/nginx/{{ .Domain }}_access.log;
    error_log /var/log/nginx/{{ .Domain }}_error.log;
}
`

// AddProxy 添加反向代理
func (p *ProxyManager) AddProxy(containerID uint, domain, targetIP string, targetPort int, ssl bool, certPath, keyPath string) (*ProxyConfig, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查域名是否已存在
	var existing models.ProxyConfig
	err := models.DB.Where("domain = ?", domain).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("域名 %s 已被使用", domain)
	}

	// 创建 Nginx 配置
	err = p.createNginxConfig(domain, targetIP, targetPort, ssl, certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("创建 Nginx 配置失败: %v", err)
	}

	// 保存到数据库
	proxy := models.ProxyConfig{
		ContainerID: containerID,
		Domain:      domain,
		TargetIP:    targetIP,
		TargetPort:  targetPort,
		SSL:         ssl,
		CertPath:    certPath,
		KeyPath:     keyPath,
		Status:      "active",
	}
	models.DB.Create(&proxy)

	// 重载 Nginx
	p.reloadNginx()

	return &ProxyConfig{
		ID:          proxy.ID,
		ContainerID: proxy.ContainerID,
		Domain:      proxy.Domain,
		TargetIP:    proxy.TargetIP,
		TargetPort:  proxy.TargetPort,
		SSL:         proxy.SSL,
		CertPath:    proxy.CertPath,
		KeyPath:     proxy.KeyPath,
		Status:      proxy.Status,
	}, nil
}

// RemoveProxy 删除反向代理
func (p *ProxyManager) RemoveProxy(proxyID uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var proxy models.ProxyConfig
	err := models.DB.First(&proxy, proxyID).Error
	if err != nil {
		return fmt.Errorf("反向代理不存在")
	}

	// 删除 Nginx 配置文件
	configPath := filepath.Join(p.nginxDir, proxy.Domain+".conf")
	os.Remove(configPath)
	
	// 删除软链接
	linkPath := filepath.Join("/etc/nginx/sites-enabled", proxy.Domain+".conf")
	os.Remove(linkPath)

	// 从数据库删除
	models.DB.Delete(&proxy)

	// 重载 Nginx
	p.reloadNginx()

	return nil
}

// RemoveContainerProxies 删除容器的所有反向代理
func (p *ProxyManager) RemoveContainerProxies(containerID uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var proxies []models.ProxyConfig
	models.DB.Where("container_id = ?", containerID).Find(&proxies)

	for _, proxy := range proxies {
		// 删除 Nginx 配置文件
		configPath := filepath.Join(p.nginxDir, proxy.Domain+".conf")
		os.Remove(configPath)
		
		// 删除软链接
		linkPath := filepath.Join("/etc/nginx/sites-enabled", proxy.Domain+".conf")
		os.Remove(linkPath)
		
		// 从数据库删除
		models.DB.Delete(&proxy)
	}

	// 重载 Nginx
	p.reloadNginx()

	return nil
}

// GetContainerProxies 获取容器的所有反向代理
func (p *ProxyManager) GetContainerProxies(containerID uint) ([]ProxyConfig, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var proxies []models.ProxyConfig
	err := models.DB.Where("container_id = ?", containerID).Find(&proxies).Error
	if err != nil {
		return nil, err
	}

	result := make([]ProxyConfig, len(proxies))
	for i, proxy := range proxies {
		result[i] = ProxyConfig{
			ID:          proxy.ID,
			ContainerID: proxy.ContainerID,
			Domain:      proxy.Domain,
			TargetIP:    proxy.TargetIP,
			TargetPort:  proxy.TargetPort,
			SSL:         proxy.SSL,
			CertPath:    proxy.CertPath,
			KeyPath:     proxy.KeyPath,
			Status:      proxy.Status,
		}
	}

	return result, nil
}

// UpdateProxySSL 更新反向代理的 SSL 配置
func (p *ProxyManager) UpdateProxySSL(proxyID uint, certPath, keyPath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var proxy models.ProxyConfig
	err := models.DB.First(&proxy, proxyID).Error
	if err != nil {
		return fmt.Errorf("反向代理不存在")
	}

	// 更新数据库
	proxy.SSL = true
	proxy.CertPath = certPath
	proxy.KeyPath = keyPath
	models.DB.Save(&proxy)

	// 重新创建 Nginx 配置
	err = p.createNginxConfig(proxy.Domain, proxy.TargetIP, proxy.TargetPort, proxy.SSL, proxy.CertPath, proxy.KeyPath)
	if err != nil {
		return fmt.Errorf("更新 Nginx 配置失败: %v", err)
	}

	// 重载 Nginx
	p.reloadNginx()

	return nil
}

// SyncNginxConfigs 同步 Nginx 配置
func (p *ProxyManager) SyncNginxConfigs() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 从数据库重新创建所有配置
	var proxies []models.ProxyConfig
	models.DB.Where("status = ?", "active").Find(&proxies)

	for _, proxy := range proxies {
		p.createNginxConfig(proxy.Domain, proxy.TargetIP, proxy.TargetPort, proxy.SSL, proxy.CertPath, proxy.KeyPath)
	}

	// 重载 Nginx
	p.reloadNginx()

	return nil
}

// createNginxConfig 创建 Nginx 配置文件
func (p *ProxyManager) createNginxConfig(domain, targetIP string, targetPort int, ssl bool, certPath, keyPath string) error {
	// 确保目录存在
	os.MkdirAll(p.nginxDir, 0755)
	os.MkdirAll("/etc/nginx/sites-enabled", 0755)

	// 解析模板
	tmpl, err := template.New("nginx").Parse(nginxConfigTemplate)
	if err != nil {
		return err
	}

	// 创建配置文件
	configPath := filepath.Join(p.nginxDir, domain+".conf")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 渲染模板
	data := struct {
		Domain     string
		TargetIP   string
		TargetPort int
		SSL        bool
		CertPath   string
		KeyPath    string
	}{
		Domain:     domain,
		TargetIP:   targetIP,
		TargetPort: targetPort,
		SSL:        ssl,
		CertPath:   certPath,
		KeyPath:    keyPath,
	}
	
	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	// 创建软链接到 sites-enabled
	linkPath := filepath.Join("/etc/nginx/sites-enabled", domain+".conf")
	os.Remove(linkPath) // 删除旧的软链接
	os.Symlink(configPath, linkPath)

	return nil
}

// reloadNginx 重载 Nginx 配置
func (p *ProxyManager) reloadNginx() error {
	// 测试配置
	testCmd := exec.Command(p.nginxBin, "-t")
	if err := testCmd.Run(); err != nil {
		return fmt.Errorf("Nginx 配置测试失败")
	}

	// 重载配置
	reloadCmd := exec.Command(p.nginxBin, "-s", "reload")
	if err := reloadCmd.Run(); err != nil {
		return fmt.Errorf("Nginx 重载失败")
	}

	return nil
}

// IsNginxInstalled 检查 Nginx 是否已安装
func (p *ProxyManager) IsNginxInstalled() bool {
	_, err := os.Stat(p.nginxBin)
	return err == nil
}

// GetProxyCount 获取反向代理数量
func (p *ProxyManager) GetProxyCount() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var count int64
	models.DB.Model(&models.ProxyConfig{}).
		Where("status = ?", "active").
		Count(&count)
	
	return count
}

// ValidateDomain 验证域名格式
func ValidateDomain(domain string) bool {
	// 简单的域名验证
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	// 可以添加更复杂的验证逻辑
	return true
}
