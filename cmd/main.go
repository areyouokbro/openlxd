package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/openlxd/backend/internal/config"
	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
)

//go:embed ../web
var webFS embed.FS

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func main() {
	log.Println("OpenLXD 启动中...")

	// 1. 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	cfg := config.GetConfig()
	log.Println("配置加载成功")

	// 2. 初始化数据库
	dbPath := cfg.Database.Path
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}
	if err := models.InitDB(dbPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Println("数据库初始化成功")

	// 3. 初始化 LXD 客户端
	if err := lxd.InitLXD(cfg.LXD.Socket); err != nil {
		log.Fatalf("LXD 初始化失败: %v", err)
	}
	log.Println("LXD 客户端初始化成功")

	// 4. 同步 LXD 容器到数据库
	if err := syncContainersFromLXD(); err != nil {
		log.Printf("警告: 容器同步失败: %v", err)
	}

	// 5. 设置 HTTP 路由
	mux := http.NewServeMux()
	setupRoutes(mux)

	// 6. 启动 HTTP(S) 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器（支持 HTTPS）
	go func() {
		if cfg.Server.HTTPS {
			certFile := filepath.Join(cfg.Server.CertDir, "server.crt")
			keyFile := filepath.Join(cfg.Server.CertDir, "server.key")

			// 如果证书不存在，生成自签名证书
			if _, err := os.Stat(certFile); os.IsNotExist(err) {
				log.Println("证书不存在，生成自签名证书...")
				if err := generateSelfSignedCert(certFile, keyFile); err != nil {
					log.Fatalf("证书生成失败: %v", err)
				}
			}

			log.Printf("HTTPS 服务器启动在 https://%s", addr)
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS 服务器启动失败: %v", err)
			}
		} else {
			log.Printf("HTTP 服务器启动在 http://%s", addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP 服务器启动失败: %v", err)
			}
		}
	}()

	log.Println("OpenLXD 启动成功！")
	log.Printf("访问地址: https://%s", addr)
	log.Printf("管理员账号: %s", cfg.Security.AdminUser)
	log.Printf("API Hash: %s", cfg.Security.APIHash)

	// 7. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}
	log.Println("服务器已关闭")
}

// setupRoutes 设置 HTTP 路由
func setupRoutes(mux *http.ServeMux) {
	// Web 管理界面
	mux.HandleFunc("/", handleWebUI)
	mux.HandleFunc("/admin", handleAdminLogin)
	mux.HandleFunc("/admin/login", handleAdminLogin)
	mux.HandleFunc("/admin/api/login", handleAdminLoginAPI)
	mux.HandleFunc("/admin/dashboard", handleAdminDashboard)

	// API 路由（需要认证）
	mux.HandleFunc("/api/system/containers", authMiddleware(handleContainers))
	mux.HandleFunc("/api/system/containers/", authMiddleware(handleContainerOperations))
	mux.HandleFunc("/api/system/stats", authMiddleware(handleSystemStats))
	mux.HandleFunc("/api/system/traffic/reset", authMiddleware(handleResetTraffic))
}

// authMiddleware 认证中间件
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := config.GetConfig()
		apiHash := r.Header.Get("X-API-Hash")
		if apiHash == "" {
			apiHash = r.URL.Query().Get("api_key")
		}

		if apiHash != cfg.Security.APIHash {
			respondJSON(w, 401, "Unauthorized", nil)
			return
		}
		next(w, r)
	}
}

// respondJSON 返回 JSON 响应
func respondJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// handleWebUI 处理 Web 首页
func handleWebUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	serveEmbeddedFile(w, "index.html")
}

// handleAdminLogin 处理管理员登录页面
func handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		serveEmbeddedFile(w, "login.html")
		return
	}
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// handleAdminLoginAPI 处理管理员登录 API
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

	cfg := config.GetConfig()
	if req.Username == cfg.Security.AdminUser && req.Password == cfg.Security.AdminPass {
		respondJSON(w, 200, "登录成功", map[string]interface{}{
			"token":   cfg.Security.APIHash,
			"api_key": cfg.Security.APIHash,
		})
		return
	}

	respondJSON(w, 401, "用户名或密码错误", nil)
}

// handleAdminDashboard 处理管理员控制台
func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, "dashboard.html")
}

// handleContainers 处理容器列表和创建
func handleContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		handleListContainers(w, r)
	} else if r.Method == "POST" {
		handleCreateContainer(w, r)
	} else {
		respondJSON(w, 405, "不支持的请求方法", nil)
	}
}

// handleListContainers 获取容器列表
func handleListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := models.GetAllContainers()
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("获取容器列表失败: %v", err), nil)
		return
	}

	result := make([]map[string]interface{}, 0)
	for _, c := range containers {
		result = append(result, map[string]interface{}{
			"hostname":      c.Hostname,
			"status":        c.Status,
			"ipv4":          c.IPv4,
			"ipv6":          c.IPv6,
			"cpus":          c.CPUs,
			"memory":        c.Memory,
			"disk":          c.Disk,
			"traffic_used":  c.TrafficUsed,
			"traffic_limit": c.TrafficLimit,
			"created_at":    c.CreatedAt,
		})
	}

	respondJSON(w, 200, "成功", result)
}

// handleCreateContainer 创建容器
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
	ipv4, ipv6 := lxd.GetContainerIP(hostname)

	// 保存到数据库
	container := models.Container{
		Hostname:     hostname,
		Status:       "Running",
		Image:        image,
		IPv4:         ipv4,
		IPv6:         ipv6,
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
		"ipv4":     ipv4,
		"ipv6":     ipv6,
	})
}

// handleContainerOperations 处理容器操作
func handleContainerOperations(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondJSON(w, 400, "无效的请求路径", nil)
		return
	}

	containerName := parts[4]

	if r.Method == "GET" {
		handleGetContainer(w, r, containerName)
	} else if r.Method == "DELETE" {
		handleDeleteContainer(w, r, containerName)
	} else if r.Method == "POST" {
		action := r.URL.Query().Get("action")
		handleContainerAction(w, r, containerName, action)
	} else {
		respondJSON(w, 405, "不支持的请求方法", nil)
	}
}

// handleGetContainer 获取容器详情
func handleGetContainer(w http.ResponseWriter, r *http.Request, containerName string) {
	container, err := models.GetContainerByHostname(containerName)
	if err != nil {
		respondJSON(w, 404, "容器不存在", nil)
		return
	}

	// 从 LXD 获取实时状态
	state, err := lxd.GetContainerState(containerName)
	if err == nil {
		container.Status = state.Status
		ipv4, ipv6 := extractIPFromState(state)
		if ipv4 != "" {
			container.IPv4 = ipv4
		}
		if ipv6 != "" {
			container.IPv6 = ipv6
		}
	}

	respondJSON(w, 200, "成功", map[string]interface{}{
		"hostname":      container.Hostname,
		"status":        container.Status,
		"ipv4":          container.IPv4,
		"ipv6":          container.IPv6,
		"cpus":          container.CPUs,
		"memory":        container.Memory,
		"disk":          container.Disk,
		"traffic_used":  container.TrafficUsed,
		"traffic_limit": container.TrafficLimit,
		"created_at":    container.CreatedAt,
	})
}

// handleContainerAction 处理容器操作
func handleContainerAction(w http.ResponseWriter, r *http.Request, containerName, action string) {
	var err error
	switch action {
	case "start":
		err = lxd.StartContainer(containerName)
		if err == nil {
			models.UpdateContainerStatus(containerName, "Running")
		}
	case "stop":
		err = lxd.StopContainer(containerName)
		if err == nil {
			models.UpdateContainerStatus(containerName, "Stopped")
		}
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
	default:
		respondJSON(w, 400, "不支持的操作", nil)
		return
	}

	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("%s 操作失败: %v", action, err), nil)
		return
	}

	models.LogAction(action, containerName, fmt.Sprintf("%s 容器", action), "success")
	respondJSON(w, 200, fmt.Sprintf("%s 操作成功", action), nil)
}

// handleDeleteContainer 删除容器
func handleDeleteContainer(w http.ResponseWriter, r *http.Request, containerName string) {
	err := lxd.DeleteContainer(containerName)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("容器删除失败: %v", err), nil)
		return
	}

	models.DeleteContainer(containerName)
	models.LogAction("delete", containerName, "删除容器", "success")

	respondJSON(w, 200, "容器删除成功", nil)
}

// handleSystemStats 获取系统统计信息
func handleSystemStats(w http.ResponseWriter, r *http.Request) {
	containers, _ := models.GetAllContainers()
	total := len(containers)
	running := 0
	for _, c := range containers {
		if c.Status == "Running" {
			running++
		}
	}

	respondJSON(w, 200, "成功", map[string]interface{}{
		"total_containers":   total,
		"running_containers": running,
	})
}

// handleResetTraffic 重置流量
func handleResetTraffic(w http.ResponseWriter, r *http.Request) {
	containerName := r.URL.Query().Get("name")
	models.DB.Model(&models.Container{}).Where("hostname = ?", containerName).Update("traffic_used", 0)
	models.LogAction("reset_traffic", containerName, "重置流量", "success")
	respondJSON(w, 200, "流量重置成功", nil)
}

// serveEmbeddedFile 提供嵌入的静态文件
func serveEmbeddedFile(w http.ResponseWriter, filename string) {
	var path string
	if strings.HasSuffix(filename, ".js") || strings.HasSuffix(filename, ".css") {
		path = "../web/static/" + filename
	} else {
		path = "../web/templates/" + filename
	}

	data, err := webFS.ReadFile(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		log.Printf("文件未找到: %s (尝试路径: %s)", filename, path)
		return
	}

	contentType := "text/html; charset=utf-8"
	if strings.HasSuffix(filename, ".js") {
		contentType = "application/javascript; charset=utf-8"
	} else if strings.HasSuffix(filename, ".css") {
		contentType = "text/css; charset=utf-8"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// syncContainersFromLXD 从 LXD 同步容器到数据库
func syncContainersFromLXD() error {
	instances, err := lxd.ListContainers()
	if err != nil {
		return err
	}

	log.Printf("从 LXD 同步 %d 个容器", len(instances))

	for _, inst := range instances {
		// 检查数据库中是否已存在
		var existing models.Container
		err := models.DB.Where("hostname = ?", inst.Name).First(&existing).Error

		if err != nil {
			// 不存在，创建新记录
			state, _ := lxd.GetContainerState(inst.Name)
			ipv4, ipv6 := "", ""
			if state != nil {
				ipv4, ipv6 = extractIPFromState(state)
			}

			container := models.Container{
				Hostname: inst.Name,
				Status:   inst.Status,
				Image:    extractImageAlias(inst.Config),
				IPv4:     ipv4,
				IPv6:     ipv6,
				CPUs:     extractCPUs(inst.Config),
				Memory:   extractMemory(inst.Config),
				Disk:     extractDisk(inst.Devices),
			}
			models.DB.Create(&container)
			log.Printf("同步容器: %s", inst.Name)
		} else {
			// 已存在，更新状态
			models.UpdateContainerStatus(inst.Name, inst.Status)
		}
	}

	return nil
}

// 辅助函数：从配置中提取信息
func extractImageAlias(config map[string]string) string {
	if alias, ok := config["image.alias"]; ok {
		return alias
	}
	if alias, ok := config["volatile.base_image"]; ok {
		return alias
	}
	return "unknown"
}

func extractCPUs(config map[string]string) int {
	if cpus, ok := config["limits.cpu"]; ok {
		var count int
		fmt.Sscanf(cpus, "%d", &count)
		return count
	}
	return 1
}

func extractMemory(config map[string]string) int {
	if mem, ok := config["limits.memory"]; ok {
		var size int
		fmt.Sscanf(mem, "%dMB", &size)
		if size == 0 {
			fmt.Sscanf(mem, "%dGB", &size)
			size *= 1024
		}
		return size
	}
	return 512
}

func extractDisk(devices map[string]map[string]string) int {
	if root, ok := devices["root"]; ok {
		if size, ok := root["size"]; ok {
			var diskSize int
			fmt.Sscanf(size, "%dGB", &diskSize)
			if diskSize == 0 {
				fmt.Sscanf(size, "%dMB", &diskSize)
				diskSize /= 1024
			}
			return diskSize
		}
	}
	return 10
}

func extractIPFromState(state *api.InstanceState) (string, string) {
	var ipv4, ipv6 string
	if eth0, ok := state.Network["eth0"]; ok {
		for _, addr := range eth0.Addresses {
			if addr.Family == "inet" && addr.Scope == "global" {
				ipv4 = addr.Address
			}
			if addr.Family == "inet6" && addr.Scope == "global" {
				ipv6 = addr.Address
			}
		}
	}
	return ipv4, ipv6
}

// generateSelfSignedCert 生成自签名证书
func generateSelfSignedCert(certFile, keyFile string) error {
	// 创建证书目录
	if err := os.MkdirAll(filepath.Dir(certFile), 0755); err != nil {
		return err
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"OpenLXD"},
			CommonName:   "OpenLXD Self-Signed Certificate",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 添加 IP 地址和域名
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	template.DNSNames = []string{"localhost"}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// 保存证书
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// 保存私钥
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	log.Println("自签名证书生成成功")
	return nil
}
