package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/openlxd/backend/internal/auth"
	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
	"gorm.io/gorm"
)

// LXDAPIRouter lxdapi路由处理器（兼容http.ServeMux）
type LXDAPIRouter struct {
	handler *LXDAPIHandler
	db      *gorm.DB
}

// NewLXDAPIRouter 创建lxdapi路由处理器
func NewLXDAPIRouter(db *gorm.DB, lxdClient *lxd.ClientWrapper) *LXDAPIRouter {
	return &LXDAPIRouter{
		handler: NewLXDAPIHandler(db, lxdClient),
		db:      db,
	}
}

// ServeHTTP 实现http.Handler接口
func (router *LXDAPIRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// API Key认证（支持X-API-Key和X-API-Hash）
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.Header.Get("X-API-Hash")
	}
	
	if apiKey == "" {
		RespondLXDAPIError(w, "Missing API key", http.StatusUnauthorized)
		return
	}
	
	// 验证API Key
	var user models.User
	if err := router.db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
		RespondLXDAPIError(w, "Invalid API key", http.StatusUnauthorized)
		return
	}
	
	// 检查用户状态
	if !user.IsActive() {
		RespondLXDAPIError(w, "User account is not active", http.StatusForbidden)
		return
	}
	
	// 将用户信息存入上下文
	ctx := context.WithValue(r.Context(), auth.UserContextKey, &user)
	r = r.WithContext(ctx)
	
	// 路由分发
	path := r.URL.Path
	method := r.Method
	
	// 移除前缀 /api/system
	path = strings.TrimPrefix(path, "/api/system")
	
	switch {
	case path == "/containers" && method == "POST":
		router.handler.CreateContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/start") && method == "POST":
		router.handler.StartContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/stop") && method == "POST":
		router.handler.StopContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/restart") && method == "POST":
		router.handler.RestartContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/suspend") && method == "POST":
		router.handler.SuspendContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/unsuspend") && method == "POST":
		router.handler.UnsuspendContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/reinstall") && method == "POST":
		router.handler.ReinstallContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/password") && method == "POST":
		router.handler.ChangePassword(w, r)
	case strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/traffic/reset") && method == "POST":
		router.handler.ResetTraffic(w, r)
	case strings.HasPrefix(path, "/containers/") && method == "DELETE":
		router.handler.DeleteContainer(w, r)
	case strings.HasPrefix(path, "/containers/") && method == "GET":
		router.handler.GetContainerInfo(w, r)
	default:
		RespondLXDAPIError(w, "Not found", http.StatusNotFound)
	}
}

// extractContainerName 从路径中提取容器名称
func extractContainerName(path string) string {
	// 移除 /containers/ 前缀
	path = strings.TrimPrefix(path, "/containers/")
	
	// 移除操作后缀
	suffixes := []string{"/start", "/stop", "/restart", "/suspend", "/unsuspend", 
		"/reinstall", "/password", "/traffic/reset"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return strings.TrimSuffix(path, suffix)
		}
	}
	
	return path
}
