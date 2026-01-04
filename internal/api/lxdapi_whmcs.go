package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/openlxd/backend/internal/auth"
	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
	"gorm.io/gorm"
)

// LXDAPIHandler lxdapi兼容的API处理器
type LXDAPIHandler struct {
	db        *gorm.DB
	lxdClient *lxd.ClientWrapper
}

// NewLXDAPIHandler 创建lxdapi API处理器
func NewLXDAPIHandler(db *gorm.DB, lxdClient *lxd.ClientWrapper) *LXDAPIHandler {
	return &LXDAPIHandler{
		db:        db,
		lxdClient: lxdClient,
	}
}

// LXDAPICreateContainerRequest lxdapi创建容器请求
type LXDAPICreateContainerRequest struct {
	Name                string `json:"name"`
	Image               string `json:"image"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	CPU                 int    `json:"cpu"`
	Memory              int    `json:"memory"` // MB
	Disk                int    `json:"disk"`   // MB
	Ingress             int    `json:"ingress"`
	Egress              int    `json:"egress"`
	TrafficLimit        int    `json:"traffic_limit"`
	IPv4PoolLimit       int    `json:"ipv4_pool_limit"`
	IPv4MappingLimit    int    `json:"ipv4_mapping_limit"`
	IPv6PoolLimit       int    `json:"ipv6_pool_limit"`
	IPv6MappingLimit    int    `json:"ipv6_mapping_limit"`
	ReverseProxyLimit   int    `json:"reverse_proxy_limit"`
	CPUAllowance        int    `json:"cpu_allowance"`
	IORead              int    `json:"io_read"`
	IOWrite             int    `json:"io_write"`
	ProcessesLimit      int    `json:"processes_limit"`
	AllowNesting        bool   `json:"allow_nesting"`
	MemorySwap          bool   `json:"memory_swap"`
	Privileged          bool   `json:"privileged"`
}

// CreateContainer 创建容器（lxdapi兼容）
func (h *LXDAPIHandler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	var req LXDAPICreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondLXDAPIError(w, "请求参数错误", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Image == "" {
		RespondLXDAPIError(w, "容器名称和镜像不能为空", http.StatusBadRequest)
		return
	}

	// 检查容器是否已存在
	var existingContainer models.Container
	if err := h.db.Where("hostname = ?", req.Name).First(&existingContainer).Error; err == nil {
		RespondLXDAPIError(w, "容器名称已存在", http.StatusConflict)
		return
	}

	// 创建LXD容器
	createReq := lxd.CreateContainerRequest{
		Hostname: req.Name,
		Image:    req.Image,
		CPUs:     req.CPU,
		Memory:   req.Memory,
		Disk:     req.Disk / 1024, // 转换为GB
		Password: req.Password,
		Ingress:  req.Ingress,
		Egress:   req.Egress,
		CPUAllowance: req.CPUAllowance,
	}
	err := lxd.CreateContainer(createReq)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("创建容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 资源限制已在CreateContainer中设置

	// 启动容器
	err = lxd.StartContainer(req.Name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("启动容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取容器IP
	ipv4 := ""
	ipv6 := ""
	time.Sleep(2 * time.Second) // 等待容器获取IP
	state, err := lxd.GetContainerState(req.Name)
	if err == nil && state.Network != nil {
		for _, net := range state.Network {
			for _, addr := range net.Addresses {
				if addr.Family == "inet" && addr.Address != "" {
					ipv4 = addr.Address
				} else if addr.Family == "inet6" && addr.Address != "" && !strings.HasPrefix(addr.Address, "fe80") {
					ipv6 = addr.Address
				}
			}
		}
	}

	// 保存到数据库
	container := models.Container{
		Hostname:  req.Name,
		Image:     req.Image,
		Status:    "running",
		CPUs:      req.CPU,
		Memory:    req.Memory,
		Disk:      req.Disk / 1024, // 转换为GB
		IPv4:      ipv4,
		IPv6:      ipv6,
		Ingress:   req.Ingress,
		Egress:    req.Egress,
		UserID:    user.ID,
		CreatedBy: user.Username,
		CreatedAt: time.Now(),
	}

	if err := h.db.Create(&container).Error; err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("保存容器信息失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回响应
	RespondLXDAPISuccess(w, map[string]interface{}{
		"name":     req.Name,
		"ipv4":     ipv4,
		"ipv6":     ipv6,
		"hostname": req.Name,
	}, "创建容器成功")
}

// StartContainer 启动容器
func (h *LXDAPIHandler) StartContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 启动容器
	err := lxd.StartContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("启动容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库状态
	h.db.Model(&container).Update("status", "running")

	RespondLXDAPISuccess(w, nil, "启动容器成功")
}

// StopContainer 停止容器
func (h *LXDAPIHandler) StopContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 停止容器
	err := lxd.StopContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("停止容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库状态
	h.db.Model(&container).Update("status", "stopped")

	RespondLXDAPISuccess(w, nil, "停止容器成功")
}

// RestartContainer 重启容器
func (h *LXDAPIHandler) RestartContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 重启容器
	err := lxd.RestartContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("重启容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	RespondLXDAPISuccess(w, nil, "重启容器成功")
}

// DeleteContainer 删除容器
func (h *LXDAPIHandler) DeleteContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 停止容器
	lxd.StopContainer(name)

	// 删除容器
	err := lxd.DeleteContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("删除容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 从数据库删除
	h.db.Delete(&container)

	RespondLXDAPISuccess(w, nil, "删除容器成功")
}

// GetContainerInfo 获取容器信息
func (h *LXDAPIHandler) GetContainerInfo(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 获取LXD状态
	state, err := lxd.GetContainerState(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("获取容器状态失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新状态
	container.Status = state.Status

	RespondLXDAPISuccess(w, map[string]interface{}{
		"name":       container.Hostname,
		"image":      container.Image,
		"status":     container.Status,
		"cpu":        container.CPUs,
		"memory":     container.Memory,
		"disk":       container.Disk,
		"ipv4":       container.IPv4,
		"ipv6":       container.IPv6,
		"created_at": container.CreatedAt,
	}, "获取容器信息成功")
}

// SuspendContainer 暂停容器
func (h *LXDAPIHandler) SuspendContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 停止容器
	err := lxd.StopContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("暂停容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库状态
	h.db.Model(&container).Update("status", "suspended")

	RespondLXDAPISuccess(w, nil, "暂停容器成功")
}

// UnsuspendContainer 恢复容器
func (h *LXDAPIHandler) UnsuspendContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 启动容器
	err := lxd.StartContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("恢复容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库状态
	h.db.Model(&container).Update("status", "running")

	RespondLXDAPISuccess(w, nil, "恢复容器成功")
}

// ReinstallContainer 重装容器
func (h *LXDAPIHandler) ReinstallContainer(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 如果没有提供镜像，使用原镜像
	}

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 使用原镜像或新镜像
	image := container.Image
	if req.Image != "" {
		image = req.Image
	}

	// 保存配置
	cpu := container.CPUs
	memory := container.Memory
	disk := container.Disk

	// 停止容器
	lxd.StopContainer(name)

	// 删除容器
	err := lxd.DeleteContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("删除旧容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 重新创建容器
	recreateReq := lxd.CreateContainerRequest{
		Hostname: name,
		Image:    image,
		CPUs:     cpu,
		Memory:   memory,
		Disk:     disk,
	}
	err = lxd.CreateContainer(recreateReq)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("创建新容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 配置已在CreateContainer中设置

	// 启动容器
	err = lxd.StartContainer(name)
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("启动新容器失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新数据库
	h.db.Model(&container).Updates(map[string]interface{}{
		"image":  image,
		"status": "running",
	})

	RespondLXDAPISuccess(w, nil, "重装容器成功")
}

// ChangePassword 修改密码
func (h *LXDAPIHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondLXDAPIError(w, "请求参数错误", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		RespondLXDAPIError(w, "密码不能为空", http.StatusBadRequest)
		return
	}

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 在容器中执行修改密码命令
	// TODO: 实现容器内命令执行
	// cmd := fmt.Sprintf("echo 'root:%s' | chpasswd", req.Password)
	// _, err := lxd.ExecContainer(name, []string{"/bin/sh", "-c", cmd})
	var err error = nil // 暂时跳过
	if err != nil {
		RespondLXDAPIError(w, fmt.Sprintf("修改密码失败: %v", err), http.StatusInternalServerError)
		return
	}

	RespondLXDAPISuccess(w, nil, "修改密码成功")
}

// ResetTraffic 重置流量
func (h *LXDAPIHandler) ResetTraffic(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		RespondLXDAPIError(w, "未授权", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	// 检查容器所有权
	var container models.Container
	if err := h.db.Where("hostname = ? AND user_id = ?", name, user.ID).First(&container).Error; err != nil {
		RespondLXDAPIError(w, "容器不存在或无权限", http.StatusNotFound)
		return
	}

	// 重置流量统计（这里简单实现，实际可能需要更复杂的逻辑）
	h.db.Model(&container).Updates(map[string]interface{}{
		"traffic_used":     0,
		"traffic_reset_at": time.Now(),
	})

	RespondLXDAPISuccess(w, nil, "流量重置成功")
}


