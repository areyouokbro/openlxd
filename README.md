# OpenLXD

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://golang.org/)
[![Version](https://img.shields.io/badge/version-3.6.0--final-brightgreen.svg)](https://github.com/areyouokbro/openlxd/releases)
[![Platform](https://img.shields.io/badge/platform-Linux-lightgrey.svg)](https://www.linux.org/)

> 🚀 生产就绪的 LXD 容器管理系统 - 100% 兼容 lxdapi WHMCS 插件

OpenLXD 是一个功能完整、生产就绪的 LXD 容器管理系统，提供完整的 RESTful API、Web 管理界面、多租户管理，并**完全兼容 lxdapi WHMCS 插件**。

## 🎉 v3.6.0 Final - 重大更新

### ✨ 新功能

- **🚀 零配置启动** - 下载即用，自动创建配置和数据库
- **🔌 100% lxdapi 兼容** - 直接使用 lxdapi WHMCS 插件，无需修改
- **👥 多租户管理** - 完整的用户系统和权限管理
- **🖼️ 镜像模板市场** - 22 个预定义镜像，一键导入
- **📦 一键安装脚本** - 生产环境快速部署

## 🚀 快速开始（30秒）

### 方式 1：直接运行（推荐）

```bash
# 下载
wget https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd
chmod +x openlxd

# 运行（自动创建配置和数据库）
./openlxd
```

访问：`http://your-server-ip:8443`

### 方式 2：使用安装脚本（生产环境）

```bash
# 下载并运行安装脚本
wget https://raw.githubusercontent.com/areyouokbro/openlxd/master/install.sh
sudo bash install.sh

# 启动服务
sudo systemctl start openlxd
```

**就这么简单！** 无需任何配置，下载即用！

## ✨ 核心特性

### 🎯 容器管理
- ✅ 创建、启动、停止、重启、删除
- ✅ 暂停/恢复容器
- ✅ 重装系统
- ✅ 修改密码
- ✅ 流量重置
- ✅ 资源配额管理（CPU、内存、磁盘）

### 👥 多租户管理
- ✅ 用户注册/登录系统
- ✅ JWT Token 认证
- ✅ API Key 管理
- ✅ 用户角色管理（admin/user）
- ✅ 容器所有权隔离

### 🔌 WHMCS 集成
- ✅ **100% 兼容 lxdapi WHMCS 插件**
- ✅ 支持 X-API-Hash 认证
- ✅ lxdapi 响应格式
- ✅ 11 个兼容 API 端点
- ✅ 无需修改 WHMCS 配置

### 🖼️ 镜像模板市场
- ✅ 22 个预定义镜像
- ✅ 从 linuxcontainers.org 导入
- ✅ 支持 Ubuntu、Debian、CentOS、Alpine、Rocky、Fedora 等
- ✅ 异步镜像导入
- ✅ 完整的镜像管理

### 🌐 网络管理
- ✅ IP 地址池管理（IPv4/IPv6）
- ✅ NAT 端口映射
- ✅ 反向代理配置
- ✅ 流量监控和统计

### 📊 监控和日志
- ✅ 系统资源监控
- ✅ 容器性能监控
- ✅ 网络流量统计
- ✅ 操作日志记录

### 🎨 Web 管理界面
- ✅ 现代化的 Web UI
- ✅ 容器管理界面
- ✅ 用户管理界面
- ✅ 镜像市场界面
- ✅ 监控仪表板

## 📋 lxdapi 兼容 API

OpenLXD 提供 11 个完全兼容 lxdapi 的 API 端点：

| 端点 | 方法 | 功能 |
|------|------|------|
| `/api/system/containers` | POST | 创建容器 |
| `/api/system/containers/{name}/start` | POST | 启动容器 |
| `/api/system/containers/{name}/stop` | POST | 停止容器 |
| `/api/system/containers/{name}/restart` | POST | 重启容器 |
| `/api/system/containers/{name}` | DELETE | 删除容器 |
| `/api/system/containers/{name}` | GET | 获取容器信息 |
| `/api/system/containers/{name}/suspend` | POST | 暂停容器 |
| `/api/system/containers/{name}/unsuspend` | POST | 恢复容器 |
| `/api/system/containers/{name}/reinstall` | POST | 重装容器 |
| `/api/system/containers/{name}/password` | POST | 修改密码 |
| `/api/system/containers/{name}/traffic/reset` | POST | 重置流量 |

### 认证方式

支持两种认证头：
- `X-API-Key` (OpenLXD 原生)
- `X-API-Hash` (lxdapi 兼容)

### 响应格式

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {...}
}
```

## 🔧 WHMCS 集成

### 1. 安装 lxdapi WHMCS 模块

```bash
cp -r lxdapiserver /path/to/whmcs/modules/servers/
```

### 2. 配置 WHMCS 产品

在 WHMCS 管理后台：

1. **产品/服务** → **创建新产品**
2. **模块设置：**
   - 模块：lxdapiserver
   - 服务器：选择或创建新服务器
3. **服务器配置：**
   - 主机名：OpenLXD 服务器 IP
   - 端口：8443
   - API Hash：用户的 API Key

### 3. 获取 API Key

```bash
# 创建用户
curl -X POST http://localhost:8443/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "your-password",
    "role": "admin"
  }'

# 登录
curl -X POST http://localhost:8443/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-password"
  }'

# 获取 API Key
curl -X GET http://localhost:8443/api/v1/users/profile \
  -H "Authorization: Bearer <your_jwt_token>"
```

**无需任何其他配置！** WHMCS 会自动调用 OpenLXD API 管理容器。

## 📊 项目统计

- **代码量：** 14,564 行
- **API 端点：** 70+
- **数据库表：** 9 个
- **Web 页面：** 13 个
- **支持镜像：** 22 个
- **文档数量：** 15+

## 🎯 功能对比

| 功能 | lxdapi | OpenLXD v3.6.0 |
|------|--------|----------------|
| API 端点路径 | ✅ | ✅ |
| X-API-Hash 认证 | ✅ | ✅ |
| 响应格式 | ✅ | ✅ |
| 容器管理 | ✅ | ✅ |
| 暂停/恢复 | ✅ | ✅ |
| 重装系统 | ✅ | ✅ |
| 修改密码 | ✅ | ✅ |
| 流量重置 | ✅ | ✅ |
| **多租户管理** | ❌ | ✅ |
| **镜像市场** | ❌ | ✅ |
| **Web 界面** | ❌ | ✅ |
| **网络管理** | ❌ | ✅ |
| **监控日志** | ❌ | ✅ |
| **兼容性** | **100%** | **✅ 100%** |

## 📚 文档

- [快速开始指南](QUICKSTART.md) - 30秒快速部署
- [完整文档](README_V3.6.0.md) - 详细功能说明
- [兼容性总结](LXDAPI_COMPATIBILITY_SUMMARY.md) - lxdapi 兼容性
- [测试文档](LXDAPI_COMPATIBILITY_TEST.md) - API 测试指南
- [最终检查报告](FINAL_CHECK_REPORT.md) - 完整检查报告

## 🏗️ 架构

```
OpenLXD
├── 后端 (Go)
│   ├── API 服务器 (70+ 端点)
│   ├── 多租户管理
│   ├── lxdapi 兼容层
│   ├── 镜像管理
│   ├── 网络管理
│   └── 监控系统
├── 前端 (HTML/JS)
│   ├── 管理界面
│   ├── 用户管理
│   ├── 镜像市场
│   └── 监控仪表板
└── 数据库 (SQLite)
    ├── 用户表
    ├── 容器表
    ├── 镜像表
    └── 其他 6 个表
```

## 🔥 使用示例

### 创建容器

```bash
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
```

### 启动容器

```bash
curl -X POST http://localhost:8443/api/system/containers/test-container/start \
  -H "X-API-Hash: your_api_key"
```

### 获取容器信息

```bash
curl -X GET http://localhost:8443/api/system/containers/test-container \
  -H "X-API-Hash: your_api_key"
```

## 🛠️ 系统要求

### 必需
- **操作系统：** Ubuntu 18.04+ / Debian 9+ / CentOS 7+
- **LXD：** 已安装并初始化
  ```bash
  sudo snap install lxd
  sudo lxd init --auto
  ```

### 可选
- 无其他依赖！OpenLXD 是单个二进制文件

## 🔧 配置

OpenLXD 会自动创建默认配置文件 `config.yaml`：

```yaml
server:
  port: 8443
  host: "0.0.0.0"

database:
  type: "sqlite"
  path: "./openlxd.db"

lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"
  bridge: "lxdbr0"
```

所有配置都可以根据需要修改。

## 📞 支持

- **GitHub Issues：** https://github.com/areyouokbro/openlxd/issues
- **文档：** https://github.com/areyouokbro/openlxd
- **Releases：** https://github.com/areyouokbro/openlxd/releases

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

---

## 🎉 立即开始

```bash
# 下载
wget https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd
chmod +x openlxd

# 运行
./openlxd
```

**就这么简单！** 🚀

---

**OpenLXD v3.6.0 Final** - 生产就绪的容器管理系统

**100% 兼容 lxdapi WHMCS 插件 | 零配置启动 | 开箱即用**
