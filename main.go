package main

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	
	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
)

//go:embed web/templates/*
var webTemplates embed.FS

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
	} `yaml:"database"`
	LXD struct {
		Socket string `yaml:"socket"`
		Bridge string `yaml:"bridge"`
	} `yaml:"lxd"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

var config Config

func loadConfig() error {
	// 按优先级尝试多个配置文件路径
	configPaths := []string{
		"/etc/openlxd/config.yaml",           // 生产环境路径
		"configs/config.yaml",               // 开发环境路径
		"./config.yaml",                     // 当前目录
		"/opt/openlxd/config.yaml",          // 备用路径
	}
	
	var lastErr error
	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("配置文件解析失败 (%s): %v", path, err)
			}
			log.Printf("成功加载配置文件: %s", path)
			return nil
		}
		lastErr = err
	}
	
	return fmt.Errorf("未找到配置文件，已尝试路径: %v\n最后错误: %v", configPaths, lastErr)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiHash := r.Header.Get("X-API-Hash")
		if apiHash == "" {
			apiHash = r.URL.Query().Get("api_key")
		}
		
		if apiHash != config.Security.APIHash {
			respondJSON(w, 401, "Unauthorized", nil)
			return
		}
		next(w, r)
	}
}

func respondJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func handleListContainers(w http.ResponseWriter, r *http.Request) {
	var containers []models.Container
	models.DB.Find(&containers)
	
	result := make([]map[string]interface{}, 0)
	for _, c := range containers {
		result = append(result, map[string]interface{}{
			"hostname":      c.Hostname,
			"status":        c.Status,
			"ipv4":          c.IPv4,
			"cpus":          c.CPUs,
			"memory":        c.Memory,
			"disk":          c.Disk,
			"traffic_used":  c.TrafficUsed,
			"traffic_limit": c.TrafficLimit,
		})
	}
	
	respondJSON(w, 200, "成功", result)
}

func handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求格式错误", nil)
		return
	}
	
	hostname := req["hostname"].(string)
	cpus := int(req["cpus"].(float64))
	memory := int(req["memory"].(float64))
	disk := int(req["disk"].(float64))
	image := req["image"].(string)
	password := req["password"].(string)
	
	// 创建 LXD 容器
	createReq := lxd.CreateContainerRequest{
		Hostname:     hostname,
		CPUs:         cpus,
		Memory:       memory,
		Disk:         disk,
		Image:        image,
		Password:     password,
		Ingress:      int(req["ingress"].(float64)),
		Egress:       int(req["egress"].(float64)),
		CPUAllowance: int(req["cpu_allowance"].(float64)),
	}
	
	err := lxd.CreateContainer(createReq)
	if err != nil {
		log.Printf("容器创建失败: %v", err)
		respondJSON(w, 500, fmt.Sprintf("容器创建失败: %v", err), nil)
		return
	}
	
	// 自动启动容器
	lxd.StartContainer(hostname)
	
	// 获取真实 IP 地址
	time.Sleep(2 * time.Second) // 等待容器网络初始化
	ipv4 := lxd.GetContainerIP(hostname)
	
	// 保存到数据库
	container := models.Container{
		Hostname:     hostname,
		Status:       "Running",
		Image:        image,
		IPv4:         ipv4,
		CPUs:         cpus,
		Memory:       memory,
		Disk:         disk,
		Ingress:      int(req["ingress"].(float64)),
		Egress:       int(req["egress"].(float64)),
		TrafficLimit: int64(req["traffic_limit"].(float64)) * 1024 * 1024 * 1024,
	}
	models.DB.Create(&container)
	
	models.LogAction("create", hostname, fmt.Sprintf("创建容器: %s", hostname), "success")
	
	respondJSON(w, 200, "容器创建成功", map[string]interface{}{
		"hostname": hostname,
		"status":   "Running",
	})
}

func handleContainerAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondJSON(w, 400, "无效的请求路径", nil)
		return
	}
	
	containerName := parts[4]
	action := r.URL.Query().Get("action")
	
	var err error
	switch action {
	case "start":
		err = lxd.StartContainer(containerName)
		models.DB.Model(&models.Container{}).Where("hostname = ?", containerName).Update("status", "Running")
	case "stop":
		err = lxd.StopContainer(containerName)
		models.DB.Model(&models.Container{}).Where("hostname = ?", containerName).Update("status", "Stopped")
	case "restart":
		err = lxd.RestartContainer(containerName)
	case "reinstall":
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		newImage := req["image"].(string)
		err = lxd.ReinstallContainer(containerName, newImage)
		if err == nil {
			models.DB.Model(&models.Container{}).Where("hostname = ?", containerName).Update("image", newImage)
		}
	case "reset-password":
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		newPassword := req["password"].(string)
		err = lxd.ResetContainerPassword(containerName, newPassword)
	}
	
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("%s 操作失败: %v", action, err), nil)
		return
	}
	
	models.LogAction(action, containerName, fmt.Sprintf("%s 容器", action), "success")
	respondJSON(w, 200, fmt.Sprintf("%s 操作成功", action), nil)
}

func handleDeleteContainer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondJSON(w, 400, "无效的请求路径", nil)
		return
	}
	
	containerName := parts[4]
	
	err := lxd.DeleteContainer(containerName)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("容器删除失败: %v", err), nil)
		return
	}
	
	models.DB.Where("hostname = ?", containerName).Delete(&models.Container{})
	models.LogAction("delete", containerName, "删除容器", "success")
	
	respondJSON(w, 200, "容器删除成功", nil)
}

func handleGetContainer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondJSON(w, 400, "无效的请求路径", nil)
		return
	}
	
	containerName := parts[4]
	
	var container models.Container
	if err := models.DB.Where("hostname = ?", containerName).First(&container).Error; err != nil {
		respondJSON(w, 404, "容器不存在", nil)
		return
	}
	
	respondJSON(w, 200, "成功", map[string]interface{}{
		"hostname":      container.Hostname,
		"status":        container.Status,
		"ipv4":          container.IPv4,
		"traffic_used":  container.TrafficUsed,
		"traffic_limit": container.TrafficLimit,
	})
}

func handleGetCredential(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, 200, "成功", map[string]interface{}{
		"access_code": "demo-access-code-12345",
	})
}

func handleResetTraffic(w http.ResponseWriter, r *http.Request) {
	containerName := r.URL.Query().Get("name")
	models.DB.Model(&models.Container{}).Where("hostname = ?", containerName).Update("traffic_used", 0)
	models.LogAction("reset_traffic", containerName, "重置流量", "success")
	respondJSON(w, 200, "流量重置成功", nil)
}

func handleCreateConsoleToken(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, 200, "成功", map[string]interface{}{
		"token": "console-token-12345",
	})
}

func handleSystemStats(w http.ResponseWriter, r *http.Request) {
	var total, running int64
	models.DB.Model(&models.Container{}).Count(&total)
	models.DB.Model(&models.Container{}).Where("status = ?", "Running").Count(&running)
	
	respondJSON(w, 200, "成功", map[string]interface{}{
		"total_containers":   total,
		"running_containers": running,
		"total_traffic":      1024,
		"sys_mem_usage":      2048,
	})
}

func serveEmbeddedFile(w http.ResponseWriter, filename string) {
	data, err := webTemplates.ReadFile("web/templates/" + filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	// 设置正确的 Content-Type
	contentType := "text/html; charset=utf-8"
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func handleWebUI(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, "index.html")
}

func handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		serveEmbeddedFile(w, "login.html")
		return
	}
}

func handleAdminLoginAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, 405, "仅支持 POST 请求", nil)
		return
	}
	
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}
	
	// 验证用户名和密码
	if req.Username == config.Security.AdminUser && req.Password == config.Security.AdminPass {
		// 登录成功，返回 API Key 作为 token
		respondJSON(w, 200, "登录成功", map[string]string{
			"token": config.Security.APIHash,
		})
	} else {
		respondJSON(w, 401, "用户名或密码错误", nil)
	}
}

func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, "dashboard.html")
}

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatal("配置文件加载失败:", err)
	}
	
	// 初始化数据库
	if err := models.InitDB(config.Database.Type, "lxdapi.db"); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	
	// 初始化 LXD 客户端
	if err := lxd.InitLXD(config.LXD.Socket); err != nil {
		log.Printf("LXD 初始化警告: %v", err)
	}
	
	// 同步 NAT 规则
	if err := lxd.SyncNATRules(); err != nil {
		log.Printf("NAT 规则同步失败: %v", err)
	}
	
	// 启动流量监控
	trafficMonitor := lxd.NewTrafficMonitor(300) // 5分钟采集一次
	trafficMonitor.Start()
	
	log.Printf("OpenLXD 后端启动中...")
	log.Printf("API Hash: %s", config.Security.APIHash)
	log.Printf("LXD 可用: %v", lxd.IsLXDAvailable())
	
	// 路由配置
	// Web 管理界面路由
	http.HandleFunc("/admin/login", handleAdminLogin)
	http.HandleFunc("/admin/api/login", handleAdminLoginAPI)
	http.HandleFunc("/admin/dashboard", handleAdminDashboard)
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
	})
	
	// 默认首页
	http.HandleFunc("/", handleWebUI)
	
	// API 路由
	http.HandleFunc("/api/system/containers", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handleListContainers(w, r)
		} else if r.Method == "POST" {
			handleCreateContainer(w, r)
		}
	}))
	
	http.HandleFunc("/api/system/containers/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/credential") {
			handleGetCredential(w, r)
		} else if strings.Contains(r.URL.Path, "/action") {
			handleContainerAction(w, r)
		} else if r.Method == "DELETE" {
			handleDeleteContainer(w, r)
		} else if r.Method == "GET" {
			handleGetContainer(w, r)
		}
	}))
	
	http.HandleFunc("/api/system/traffic/reset", authMiddleware(handleResetTraffic))
	http.HandleFunc("/api/system/console/create-token", authMiddleware(handleCreateConsoleToken))
	http.HandleFunc("/api/system/stats", authMiddleware(handleSystemStats))
	
	// 启动服务器
	if config.Server.HTTPS {
		if config.Server.AutoTLS && config.Server.Domain != "" {
			// 使用 Let's Encrypt 自动证书
			startHTTPSWithAutoCert()
		} else {
			// 使用手动证书
			startHTTPSWithManualCert()
		}
	} else {
		// HTTP 模式
		addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
		log.Printf("服务器监听 (HTTP): %s\n", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

// 使用 Let's Encrypt 自动证书启动 HTTPS
func startHTTPSWithAutoCert() {
	certDir := config.Server.CertDir
	if certDir == "" {
		certDir = "/etc/openlxd/certs"
	}
	
	// 创建证书目录
	os.MkdirAll(certDir, 0700)
	
	// 配置 autocert
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.Server.Domain),
		Cache:      autocert.DirCache(certDir),
	}
	
	// HTTP 服务器（用于 ACME 验证）
	go func() {
		log.Printf("HTTP 服务器启动 (ACME 验证): :80")
		http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	}()
	
	// HTTPS 服务器
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}
	
	log.Printf("服务器监听 (HTTPS): %s", server.Addr)
	log.Printf("域名: %s", config.Server.Domain)
	log.Printf("证书目录: %s", certDir)
	log.Fatal(server.ListenAndServeTLS("", ""))
}

// 使用手动证书启动 HTTPS
func startHTTPSWithManualCert() {
	certFile := config.Server.CertDir + "/cert.pem"
	keyFile := config.Server.CertDir + "/key.pem"
	
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatal("证书文件不存在: ", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatal("密钥文件不存在: ", keyFile)
	}
	
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("服务器监听 (HTTPS): %s\n", addr)
	log.Printf("证书: %s", certFile)
	log.Printf("密钥: %s", keyFile)
	log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, nil))
}
