# OpenLXD v3.5.0 更新说明

## 概述

v3.5.0 是一个重大更新，添加了多租户管理系统、WHMCS 兼容 API 和镜像模板市场。

## 新增功能

### 1. 多租户管理系统

**数据模型：**
- `internal/models/user.go` - 用户模型
- `internal/models/container.go` - 添加 UserID 和 CreatedBy 字段

**认证系统：**
- `internal/auth/auth.go` - JWT Token 和 API Key 生成/验证
- `internal/auth/middleware.go` - 认证中间件

**用户 API：**
- `internal/api/user.go` - 用户管理 API

**功能：**
- 用户注册和登录
- JWT Token 认证（Web 界面）
- API Key 认证（WHMCS 对接）
- 用户角色管理（admin/user）
- 容器所有权管理

### 2. WHMCS 兼容 API

**文件：**
- `internal/api/whmcs.go` - WHMCS 兼容 API

**端点：**
- POST `/api/v1/whmcs/container/create` - 创建容器
- POST `/api/v1/whmcs/container/start` - 启动容器
- POST `/api/v1/whmcs/container/stop` - 停止容器
- POST `/api/v1/whmcs/container/restart` - 重启容器
- POST `/api/v1/whmcs/container/delete` - 删除容器
- GET `/api/v1/whmcs/container/info` - 获取容器信息
- POST `/api/v1/whmcs/container/config` - 更新容器配置

**认证：**
- 请求头：`X-API-Key: <user_api_key>`

### 3. 镜像模板市场

**数据模型：**
- `internal/models/image.go` - 镜像模型

**镜像 API：**
- `internal/api/image.go` - 镜像管理 API

**LXD 集成：**
- `internal/lxd/client.go` - 添加镜像管理方法
- `internal/lxd/wrapper.go` - LXD 客户端包装器

**端点：**
- GET `/api/v1/images/list` - 获取本地镜像列表
- GET `/api/v1/images/remote` - 获取远程镜像列表
- POST `/api/v1/images/import` - 导入镜像
- DELETE `/api/v1/images/:alias` - 删除镜像
- POST `/api/v1/images/sync` - 同步镜像

**支持的镜像：**
- Ubuntu: 24.04, 22.04, 20.04, 18.04
- Debian: 12, 11, 10
- CentOS: 9-Stream, 8-Stream, 7
- Alpine: 3.19, 3.18, 3.17, 3.16
- Rocky Linux: 9, 8
- Fedora: 40, 39, 38
- Arch Linux: current
- Oracle Linux: 9, 8

### 4. Web 界面更新

**JavaScript 模块：**
- `web/static/user.js` - 用户管理模块
- `web/static/images.js` - 镜像市场模块

**HTML 页面：**
- `web/templates/login.html` - 登录/注册页面（全新设计）

**功能：**
- 用户信息显示
- API 密钥管理
- 用户列表和管理（管理员）
- 本地镜像列表
- 远程镜像浏览和导入
- 按发行版过滤镜像

## 数据库迁移

需要在数据库初始化时添加新表：

```go
// 在 internal/models/db.go 的 InitDB 函数中添加：
db.AutoMigrate(&User{})
db.AutoMigrate(&Image{})

// 更新 Container 表
db.AutoMigrate(&Container{})
```

## main.go 修改

### 1. 添加导入

```go
import (
    // ... 现有导入 ...
    "openlxd/internal/auth"
)
```

### 2. 在 setupRoutes 函数中添加路由

```go
func setupRoutes(mux *http.ServeMux) {
    // ... 现有路由 ...
    
    // 用户管理 API
    userAPI := api.NewUserAPI(models.DB)
    mux.HandleFunc("/api/v1/users/register", userAPI.Register)
    mux.HandleFunc("/api/v1/users/login", userAPI.Login)
    mux.HandleFunc("/api/v1/users/profile", jwtAuthMiddleware(userAPI.GetProfile))
    mux.HandleFunc("/api/v1/users/regenerate-key", jwtAuthMiddleware(userAPI.RegenerateAPIKey))
    mux.HandleFunc("/api/v1/users/list", jwtAuthMiddleware(adminOnly(userAPI.ListUsers)))
    mux.HandleFunc("/api/v1/users/status", jwtAuthMiddleware(adminOnly(userAPI.UpdateUserStatus)))
    mux.HandleFunc("/api/v1/users/role", jwtAuthMiddleware(adminOnly(userAPI.UpdateUserRole)))
    mux.HandleFunc("/api/v1/users/containers", jwtAuthMiddleware(userAPI.GetUserContainers))
    
    // WHMCS 兼容 API
    lxdClient := lxd.NewClient()
    whmcsAPI := api.NewWHMCSAPI(models.DB, lxdClient)
    mux.HandleFunc("/api/v1/whmcs/container/create", apiKeyAuthMiddleware(whmcsAPI.CreateContainer))
    mux.HandleFunc("/api/v1/whmcs/container/start", apiKeyAuthMiddleware(whmcsAPI.StartContainer))
    mux.HandleFunc("/api/v1/whmcs/container/stop", apiKeyAuthMiddleware(whmcsAPI.StopContainer))
    mux.HandleFunc("/api/v1/whmcs/container/restart", apiKeyAuthMiddleware(whmcsAPI.RestartContainer))
    mux.HandleFunc("/api/v1/whmcs/container/delete", apiKeyAuthMiddleware(whmcsAPI.DeleteContainer))
    mux.HandleFunc("/api/v1/whmcs/container/info", apiKeyAuthMiddleware(whmcsAPI.GetContainerInfo))
    mux.HandleFunc("/api/v1/whmcs/container/config", apiKeyAuthMiddleware(whmcsAPI.UpdateContainerConfig))
    
    // 镜像管理 API
    imageAPI := api.NewImageAPI(models.DB, lxdClient)
    mux.HandleFunc("/api/v1/images/list", jwtAuthMiddleware(imageAPI.ListImages))
    mux.HandleFunc("/api/v1/images/remote", jwtAuthMiddleware(imageAPI.GetRemoteImages))
    mux.HandleFunc("/api/v1/images/import", jwtAuthMiddleware(adminOnly(imageAPI.ImportImage)))
    mux.HandleFunc("/api/v1/images/sync", jwtAuthMiddleware(adminOnly(imageAPI.SyncImages)))
    
    // 删除镜像需要特殊处理（从 URL 获取 alias）
    mux.HandleFunc("/api/v1/images/", jwtAuthMiddleware(adminOnly(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "DELETE" {
            imageAPI.DeleteImage(w, r)
        } else {
            http.NotFound(w, r)
        }
    })))
}
```

### 3. 添加新的中间件

```go
// jwtAuthMiddleware JWT 认证中间件
func jwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return auth.AuthMiddleware(models.DB)(http.HandlerFunc(next)).ServeHTTP
}

// apiKeyAuthMiddleware API Key 认证中间件
func apiKeyAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return auth.APIKeyMiddleware(models.DB)(http.HandlerFunc(next)).ServeHTTP
}

// adminOnly 管理员权限中间件
func adminOnly(next http.HandlerFunc) http.HandlerFunc {
    return auth.AdminMiddleware(http.HandlerFunc(next)).ServeHTTP
}
```

## 依赖包

需要安装以下新依赖：

```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

## 配置文件

可选：在 config.yaml 中添加 JWT 密钥配置：

```yaml
security:
  jwt_secret: "your-secret-key-change-in-production"
  token_expiration: 24h
```

## 测试

### 1. 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```

### 2. 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### 3. WHMCS API（使用 API Key）

```bash
curl -X POST http://localhost:8080/api/v1/whmcs/container/create \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"name":"test-container","image":"ubuntu/22.04","cpu":2,"memory":"2GB","disk":"20GB"}'
```

### 4. 镜像管理

```bash
# 获取远程镜像列表
curl -X GET http://localhost:8080/api/v1/images/remote \
  -H "Authorization: Bearer your-jwt-token"

# 导入镜像
curl -X POST http://localhost:8080/api/v1/images/import \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"alias":"ubuntu/22.04","architecture":"amd64"}'
```

## 升级步骤

1. 备份数据库
2. 更新代码
3. 安装新依赖
4. 修改 main.go 添加新路由
5. 修改 internal/models/db.go 添加数据库迁移
6. 重新编译
7. 重启服务

## 注意事项

1. **JWT 密钥**：生产环境必须修改 `internal/auth/auth.go` 中的 JWTSecret
2. **首个管理员**：第一个注册的用户默认为普通用户，需要手动修改数据库设置为管理员
3. **API 兼容性**：新的 API 使用 `/api/v1/` 前缀，旧 API 保持不变
4. **镜像导入**：镜像导入是异步操作，可能需要几分钟时间

## 破坏性变更

- 无破坏性变更，完全向后兼容

## 已知问题

- 镜像导入进度无法实时查看（计划在后续版本添加 WebSocket 支持）
- 首个用户需要手动设置为管理员

## 下一步计划

- 完善容器迁移功能
- 添加集群管理
- 添加多租户配额管理
- WebSocket 实时通知
