# lxdapi 路由添加指南

## 需要在 main.go 中添加的代码

### 1. 导入新的 API 处理器

在 main.go 的导入部分添加（如果还没有）：

```go
import (
    // ... 现有导入 ...
    "github.com/gorilla/mux"
)
```

### 2. 创建 lxdapi 处理器实例

在 `setupRoutes` 函数中，创建 API 处理器之后添加：

```go
// 创建 lxdapi 兼容的 API 处理器
lxdapiHandler := api.NewLXDAPIHandler(db, lxdClientWrapper)
```

### 3. 添加 lxdapi 兼容路由

在路由设置部分添加以下代码：

```go
// ========================================
// lxdapi 兼容路由（使用 X-API-Hash 认证）
// ========================================

// 创建 lxdapi API 路由组
lxdapiRouter := mux.NewRouter()
lxdapiRouter.Use(auth.APIKeyMiddleware(db))

// 容器管理（lxdapi 格式）
lxdapiRouter.HandleFunc("/api/system/containers", lxdapiHandler.CreateContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/start", lxdapiHandler.StartContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/stop", lxdapiHandler.StopContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/restart", lxdapiHandler.RestartContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}", lxdapiHandler.DeleteContainer).Methods("DELETE")
lxdapiRouter.HandleFunc("/api/system/containers/{name}", lxdapiHandler.GetContainerInfo).Methods("GET")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/suspend", lxdapiHandler.SuspendContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/unsuspend", lxdapiHandler.UnsuspendContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/reinstall", lxdapiHandler.ReinstallContainer).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/password", lxdapiHandler.ChangePassword).Methods("POST")
lxdapiRouter.HandleFunc("/api/system/containers/{name}/traffic/reset", lxdapiHandler.ResetTraffic).Methods("POST")

// 将 lxdapi 路由挂载到主路由
http.Handle("/api/system/", lxdapiRouter)
```

### 4. 完整示例

如果你的 main.go 使用 gorilla/mux，完整的路由设置应该类似：

```go
func setupRoutes(db *gorm.DB, lxdClient *lxd.Client) http.Handler {
    r := mux.NewRouter()
    
    // 创建 API 处理器
    userAPI := api.NewUserAPI(db)
    whmcsAPI := api.NewWHMCSAPI(db, lxdClientWrapper)
    imageAPI := api.NewImageAPI(db, lxdClientWrapper)
    lxdapiHandler := api.NewLXDAPIHandler(db, lxdClientWrapper)
    
    // 公开路由
    r.HandleFunc("/api/v1/users/register", userAPI.Register).Methods("POST")
    r.HandleFunc("/api/v1/users/login", userAPI.Login).Methods("POST")
    
    // JWT 认证路由
    jwtRouter := r.PathPrefix("/api/v1").Subrouter()
    jwtRouter.Use(auth.AuthMiddleware(db))
    jwtRouter.HandleFunc("/users/profile", userAPI.GetProfile).Methods("GET")
    // ... 其他 JWT 路由 ...
    
    // API Key 认证路由（原 OpenLXD WHMCS API）
    apiKeyRouter := r.PathPrefix("/api/v1/whmcs").Subrouter()
    apiKeyRouter.Use(auth.APIKeyMiddleware(db))
    apiKeyRouter.HandleFunc("/container/create", whmcsAPI.CreateContainer).Methods("POST")
    // ... 其他 WHMCS 路由 ...
    
    // lxdapi 兼容路由（使用 X-API-Hash 认证）
    lxdapiRouter := r.PathPrefix("/api/system").Subrouter()
    lxdapiRouter.Use(auth.APIKeyMiddleware(db))
    lxdapiRouter.HandleFunc("/containers", lxdapiHandler.CreateContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/start", lxdapiHandler.StartContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/stop", lxdapiHandler.StopContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/restart", lxdapiHandler.RestartContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}", lxdapiHandler.DeleteContainer).Methods("DELETE")
    lxdapiRouter.HandleFunc("/containers/{name}", lxdapiHandler.GetContainerInfo).Methods("GET")
    lxdapiRouter.HandleFunc("/containers/{name}/suspend", lxdapiHandler.SuspendContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/unsuspend", lxdapiHandler.UnsuspendContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/reinstall", lxdapiHandler.ReinstallContainer).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/password", lxdapiHandler.ChangePassword).Methods("POST")
    lxdapiRouter.HandleFunc("/containers/{name}/traffic/reset", lxdapiHandler.ResetTraffic).Methods("POST")
    
    return r
}
```

## 路由对照表

| lxdapi 路由 | OpenLXD 处理器 | 功能 |
|-------------|---------------|------|
| `POST /api/system/containers` | `CreateContainer` | 创建容器 |
| `POST /api/system/containers/{name}/start` | `StartContainer` | 启动容器 |
| `POST /api/system/containers/{name}/stop` | `StopContainer` | 停止容器 |
| `POST /api/system/containers/{name}/restart` | `RestartContainer` | 重启容器 |
| `DELETE /api/system/containers/{name}` | `DeleteContainer` | 删除容器 |
| `GET /api/system/containers/{name}` | `GetContainerInfo` | 获取容器信息 |
| `POST /api/system/containers/{name}/suspend` | `SuspendContainer` | 暂停容器 |
| `POST /api/system/containers/{name}/unsuspend` | `UnsuspendContainer` | 恢复容器 |
| `POST /api/system/containers/{name}/reinstall` | `ReinstallContainer` | 重装容器 |
| `POST /api/system/containers/{name}/password` | `ChangePassword` | 修改密码 |
| `POST /api/system/containers/{name}/traffic/reset` | `ResetTraffic` | 重置流量 |

## 测试路由

添加路由后，可以使用以下命令测试：

```bash
# 测试创建容器
curl -X POST http://localhost:8443/api/system/containers \
  -H "X-API-Hash: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-container",
    "image": "ubuntu:22.04",
    "cpu": 2,
    "memory": 2048,
    "disk": 20480
  }'

# 测试启动容器
curl -X POST http://localhost:8443/api/system/containers/test-container/start \
  -H "X-API-Hash: your_api_key"

# 测试获取容器信息
curl -X GET http://localhost:8443/api/system/containers/test-container \
  -H "X-API-Hash: your_api_key"
```

## 注意事项

1. **认证头兼容性**：中间件已经支持 `X-API-Key` 和 `X-API-Hash` 两种认证头
2. **响应格式**：所有 lxdapi 路由使用 `{code, msg, data}` 格式
3. **容器命名**：支持 lxdapi 的特殊命名规则（如 `lxd11451123456`）
4. **错误处理**：所有错误都返回 lxdapi 格式的响应

## 编译和运行

```bash
cd /home/ubuntu/openlxd-final
go build -o bin/openlxd-lxdapi
./bin/openlxd-lxdapi
```
