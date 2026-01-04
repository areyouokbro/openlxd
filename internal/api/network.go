package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/openlxd/backend/internal/models"
	"github.com/openlxd/backend/internal/network"
)

// HandleIPPool IP地址池管理
func HandleIPPool(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetIPPool(w, r)
	case "POST":
		handleAddIPRange(w, r)
	case "DELETE":
		handleRemoveIPRange(w, r)
	default:
		respondJSON(w, 405, "不支持的请求方法", nil)
	}
}

// handleGetIPPool 获取IP地址池信息
func handleGetIPPool(w http.ResponseWriter, r *http.Request) {
	// 获取所有 IP 地址
	var ipAddresses []models.IPAddress
	models.DB.Find(&ipAddresses)

	// 统计信息
	ipv4Available := network.GlobalIPPool.GetAvailableIPv4Count()
	ipv6Available := network.GlobalIPPool.GetAvailableIPv6Count()

	respondJSON(w, 200, "成功", map[string]interface{}{
		"ip_addresses":    ipAddresses,
		"ipv4_available":  ipv4Available,
		"ipv6_available":  ipv6Available,
		"total_addresses": len(ipAddresses),
	})
}

// handleAddIPRange 添加IP地址段
func handleAddIPRange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartIP string `json:"start_ip"`
		EndIP   string `json:"end_ip"`
		Gateway string `json:"gateway"`
		Netmask string `json:"netmask"`
		Type    string `json:"type"` // ipv4, ipv6
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	// 验证参数
	if !network.ValidateIP(req.StartIP) || !network.ValidateIP(req.EndIP) {
		respondJSON(w, 400, "无效的IP地址", nil)
		return
	}

	// 添加IP地址段
	err := network.GlobalIPPool.AddIPRange(req.StartIP, req.EndIP, req.Gateway, req.Netmask, req.Type)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("添加IP地址段失败: %v", err), nil)
		return
	}

	models.LogAction("add_ip_range", "", fmt.Sprintf("添加IP地址段: %s - %s", req.StartIP, req.EndIP), "success")
	respondJSON(w, 200, "IP地址段添加成功", nil)
}

// handleRemoveIPRange 删除IP地址段
func handleRemoveIPRange(w http.ResponseWriter, r *http.Request) {
	startIP := r.URL.Query().Get("start_ip")
	endIP := r.URL.Query().Get("end_ip")

	if startIP == "" || endIP == "" {
		respondJSON(w, 400, "缺少参数", nil)
		return
	}

	err := network.GlobalIPPool.RemoveIPRange(startIP, endIP)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("删除IP地址段失败: %v", err), nil)
		return
	}

	models.LogAction("remove_ip_range", "", fmt.Sprintf("删除IP地址段: %s - %s", startIP, endIP), "success")
	respondJSON(w, 200, "IP地址段删除成功", nil)
}

// HandlePortMapping 端口映射管理
func HandlePortMapping(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetPortMappings(w, r)
	case "POST":
		handleAddPortMapping(w, r)
	case "DELETE":
		handleRemovePortMapping(w, r)
	default:
		respondJSON(w, 405, "不支持的请求方法", nil)
	}
}

// handleGetPortMappings 获取端口映射列表
func handleGetPortMappings(w http.ResponseWriter, r *http.Request) {
	containerIDStr := r.URL.Query().Get("container_id")
	
	if containerIDStr != "" {
		// 获取指定容器的端口映射
		containerID, _ := strconv.ParseUint(containerIDStr, 10, 32)
		mappings, err := network.GlobalNATManager.GetContainerMappings(uint(containerID))
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("获取端口映射失败: %v", err), nil)
			return
		}
		respondJSON(w, 200, "成功", mappings)
	} else {
		// 获取所有端口映射
		var mappings []models.PortMapping
		models.DB.Find(&mappings)
		respondJSON(w, 200, "成功", mappings)
	}
}

// handleAddPortMapping 添加端口映射
func handleAddPortMapping(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContainerID  uint   `json:"container_id"`
		ContainerIP  string `json:"container_ip"`
		Protocol     string `json:"protocol"`
		ExternalPort int    `json:"external_port"`
		InternalPort int    `json:"internal_port"`
		Description  string `json:"description"`
		Type         string `json:"type"` // single, range, random
		Count        int    `json:"count"` // 用于端口段
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	// 根据类型添加端口映射
	switch req.Type {
	case "single":
		// 单端口映射
		mapping, err := network.GlobalNATManager.AddPortMapping(
			req.ContainerID, req.ContainerIP, req.Protocol,
			req.ExternalPort, req.InternalPort, req.Description,
		)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("添加端口映射失败: %v", err), nil)
			return
		}
		models.LogAction("add_port_mapping", "", fmt.Sprintf("添加端口映射: %d -> %s:%d", req.ExternalPort, req.ContainerIP, req.InternalPort), "success")
		respondJSON(w, 200, "端口映射添加成功", mapping)

	case "range":
		// 端口段映射
		if req.Count <= 0 {
			respondJSON(w, 400, "端口数量必须大于0", nil)
			return
		}
		err := network.GlobalNATManager.AddPortRange(
			req.ContainerID, req.ContainerIP, req.Protocol,
			req.ExternalPort, req.InternalPort, req.Count, req.Description,
		)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("添加端口段失败: %v", err), nil)
			return
		}
		models.LogAction("add_port_range", "", fmt.Sprintf("添加端口段: %d-%d", req.ExternalPort, req.ExternalPort+req.Count-1), "success")
		respondJSON(w, 200, "端口段添加成功", nil)

	case "random":
		// 随机端口映射
		mapping, err := network.GlobalNATManager.AddRandomPort(
			req.ContainerID, req.ContainerIP, req.Protocol,
			req.InternalPort, req.Description,
		)
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("添加随机端口失败: %v", err), nil)
			return
		}
		models.LogAction("add_random_port", "", fmt.Sprintf("添加随机端口映射: %d -> %s:%d", mapping.ExternalPort, req.ContainerIP, req.InternalPort), "success")
		respondJSON(w, 200, "随机端口映射添加成功", mapping)

	default:
		respondJSON(w, 400, "无效的映射类型", nil)
	}
}

// handleRemovePortMapping 删除端口映射
func handleRemovePortMapping(w http.ResponseWriter, r *http.Request) {
	mappingIDStr := r.URL.Query().Get("id")
	if mappingIDStr == "" {
		respondJSON(w, 400, "缺少映射ID", nil)
		return
	}

	mappingID, err := strconv.ParseUint(mappingIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "无效的映射ID", nil)
		return
	}

	err = network.GlobalNATManager.RemovePortMapping(uint(mappingID))
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("删除端口映射失败: %v", err), nil)
		return
	}

	models.LogAction("remove_port_mapping", "", fmt.Sprintf("删除端口映射: ID %d", mappingID), "success")
	respondJSON(w, 200, "端口映射删除成功", nil)
}

// HandleProxy 反向代理管理
func HandleProxy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetProxies(w, r)
	case "POST":
		handleAddProxy(w, r)
	case "PUT":
		handleUpdateProxy(w, r)
	case "DELETE":
		handleRemoveProxy(w, r)
	default:
		respondJSON(w, 405, "不支持的请求方法", nil)
	}
}

// handleGetProxies 获取反向代理列表
func handleGetProxies(w http.ResponseWriter, r *http.Request) {
	containerIDStr := r.URL.Query().Get("container_id")
	
	if containerIDStr != "" {
		// 获取指定容器的反向代理
		containerID, _ := strconv.ParseUint(containerIDStr, 10, 32)
		proxies, err := network.GlobalProxyManager.GetContainerProxies(uint(containerID))
		if err != nil {
			respondJSON(w, 500, fmt.Sprintf("获取反向代理失败: %v", err), nil)
			return
		}
		respondJSON(w, 200, "成功", proxies)
	} else {
		// 获取所有反向代理
		var proxies []models.ProxyConfig
		models.DB.Find(&proxies)
		respondJSON(w, 200, "成功", proxies)
	}
}

// handleAddProxy 添加反向代理
func handleAddProxy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContainerID uint   `json:"container_id"`
		Domain      string `json:"domain"`
		TargetIP    string `json:"target_ip"`
		TargetPort  int    `json:"target_port"`
		SSL         bool   `json:"ssl"`
		CertPath    string `json:"cert_path"`
		KeyPath     string `json:"key_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	// 验证域名
	if !network.ValidateDomain(req.Domain) {
		respondJSON(w, 400, "无效的域名", nil)
		return
	}

	// 添加反向代理
	proxy, err := network.GlobalProxyManager.AddProxy(
		req.ContainerID, req.Domain, req.TargetIP, req.TargetPort,
		req.SSL, req.CertPath, req.KeyPath,
	)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("添加反向代理失败: %v", err), nil)
		return
	}

	models.LogAction("add_proxy", "", fmt.Sprintf("添加反向代理: %s -> %s:%d", req.Domain, req.TargetIP, req.TargetPort), "success")
	respondJSON(w, 200, "反向代理添加成功", proxy)
}

// handleUpdateProxy 更新反向代理
func handleUpdateProxy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProxyID  uint   `json:"proxy_id"`
		CertPath string `json:"cert_path"`
		KeyPath  string `json:"key_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	err := network.GlobalProxyManager.UpdateProxySSL(req.ProxyID, req.CertPath, req.KeyPath)
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("更新反向代理失败: %v", err), nil)
		return
	}

	models.LogAction("update_proxy", "", fmt.Sprintf("更新反向代理SSL: ID %d", req.ProxyID), "success")
	respondJSON(w, 200, "反向代理更新成功", nil)
}

// handleRemoveProxy 删除反向代理
func handleRemoveProxy(w http.ResponseWriter, r *http.Request) {
	proxyIDStr := r.URL.Query().Get("id")
	if proxyIDStr == "" {
		respondJSON(w, 400, "缺少代理ID", nil)
		return
	}

	proxyID, err := strconv.ParseUint(proxyIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "无效的代理ID", nil)
		return
	}

	err = network.GlobalProxyManager.RemoveProxy(uint(proxyID))
	if err != nil {
		respondJSON(w, 500, fmt.Sprintf("删除反向代理失败: %v", err), nil)
		return
	}

	models.LogAction("remove_proxy", "", fmt.Sprintf("删除反向代理: ID %d", proxyID), "success")
	respondJSON(w, 200, "反向代理删除成功", nil)
}

// HandleNetworkStats 网络统计信息
func HandleNetworkStats(w http.ResponseWriter, r *http.Request) {
	ipv4Available := network.GlobalIPPool.GetAvailableIPv4Count()
	ipv6Available := network.GlobalIPPool.GetAvailableIPv6Count()
	portMappingsCount := network.GlobalNATManager.GetUsedPortsCount()
	proxiesCount := network.GlobalProxyManager.GetProxyCount()

	respondJSON(w, 200, "成功", map[string]interface{}{
		"ipv4_available":      ipv4Available,
		"ipv6_available":      ipv6Available,
		"port_mappings_count": portMappingsCount,
		"proxies_count":       proxiesCount,
	})
}

// respondJSON 返回 JSON 响应
func respondJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"code": code,
		"msg":  message,
		"data": data,
	}
	
	json.NewEncoder(w).Encode(response)
}
