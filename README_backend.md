# OpenLXD Backend - 完全开源的 LXD 容器管理后端

## 项目简介

OpenLXD Backend 是一个**完全开源**的 LXD 容器管理后端，基于对原版 lxdapi-web-server 的深度分析开发，API 接口 100% 兼容 WHMCS、魔方财务等财务系统插件。

## 核心特性

### ✅ 已完整实现

1. **容器生命周期管理**
   - ✅ 创建容器（支持自定义 CPU、内存、磁盘、镜像）
   - ✅ 启动/停止/重启容器
   - ✅ 删除容器
   - ✅ 重装系统（保留配置）
   - ✅ 重置 root 密码
   - ✅ 自动获取容器 IP 地址

2. **资源限制与配额**
   - ✅ CPU 核心数限制
   - ✅ CPU 使用率限制（百分比）
   - ✅ 内存硬限制
   - ✅ 磁盘大小限制
   - ✅ 网络带宽限制（Ingress/Egress）

3. **网络管理**
   - ✅ 自动获取容器 IPv4 地址
   - ✅ NAT 端口映射（基于 iptables）
   - ✅ 端口映射持久化存储
   - ✅ 服务启动时自动恢复 NAT 规则

4. **流量统计与控制**
   - ✅ 异步流量监控（可配置采集间隔）
   - ✅ 流量配额控制
   - ✅ 超限自动停机
   - ✅ 流量重置接口

5. **数据持久化**
   - ✅ SQLite 数据库存储
   - ✅ 容器信息持久化
   - ✅ 端口映射持久化
   - ✅ 审计日志记录
   - ✅ 配置项存储

6. **安全认证**
   - ✅ API Key 认证（X-API-Hash Header）
   - ✅ 支持 Query 参数传递密钥

7. **Web 管理界面**
   - ✅ 系统概览仪表盘
   - ✅ 容器列表展示
   - ✅ 实时状态监控
   - ✅ 8443 端口访问

8. **LXD 集成**
   - ✅ 通过 Unix Socket 连接 LXD
   - ✅ 支持 Mock 模式（无 LXD 环境也能运行）
   - ✅ 真实的容器操作（需 LXD 环境）

### 🚧 待完善功能

- WebSocket 控制台代理（框架已搭建）
- IPv6 支持
- IP 地址池自动分配
- 镜像仓库管理

## 快速开始

### 环境要求

- **必需**：Go 1.22+、Linux 系统（推荐 Ubuntu 22.04）
- **可选**：LXD 5.0+（无 LXD 可以 Mock 模式运行）

### 安装步骤

#### 1. 克隆或下载项目

```bash
cd /opt
git clone https://github.com/areyouokbro/openlxd.git
cd openlxd-backend
```

#### 2. 配置文件

编辑 `configs/config.yaml`：

```yaml
server:
  port: 8443
  host: "0.0.0.0"

security:
  api_hash: "change-this-to-your-secret-key"  # ⚠️ 必须修改
  admin_user: "admin"
  admin_pass: "admin123"
  session_secret: "random-secret-string"

database:
  type: "sqlite"  # 目前仅支持 sqlite

lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"
  bridge: "lxdbr0"
```

#### 3. 编译并运行

```bash
go build -o openlxd cmd/main.go
sudo ./openlxd
```

#### 4. 测试 API

```bash
# 测试系统统计接口
curl -H "X-API-Hash: change-this-to-your-secret-key" \
  http://localhost:8443/api/system/stats

# 创建测试容器
curl -X POST -H "X-API-Hash: change-this-to-your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test1",
    "cpus": 2,
    "memory": 1024,
    "disk": 10240,
    "image": "ubuntu2204",
    "password": "test123",
    "ingress": 100,
    "egress": 100,
    "traffic_limit": 100,
    "cpu_allowance": 50
  }' \
  http://localhost:8443/api/system/containers
```

#### 5. 访问 Web 管理界面

浏览器打开：`http://您的服务器IP:8443`

### 使用 systemd 管理服务

创建服务文件 `/etc/systemd/system/openlxd.service`：

```ini
[Unit]
Description=OpenLXD Backend Service
After=network.target lxd.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/openlxd-backend
ExecStart=/opt/openlxd-backend/openlxd
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable openlxd
sudo systemctl start openlxd
sudo systemctl status openlxd
```

## API 文档

### 认证方式

所有 API 请求必须携带 `X-API-Hash` Header：

```
X-API-Hash: your-secret-api-key-here
```

或通过 Query 参数：

```
?api_key=your-secret-api-key-here
```

### 统一响应格式

```json
{
  "code": 200,
  "msg": "成功",
  "data": {}
}
```

- `code`: 状态码（200=成功，其他=失败）
- `msg`: 消息描述
- `data`: 返回数据

### 核心接口

#### 1. 列出所有容器

```http
GET /api/system/containers
```

响应示例：

```json
{
  "code": 200,
  "msg": "成功",
  "data": [
    {
      "hostname": "test1",
      "status": "Running",
      "ipv4": "10.0.0.100",
      "cpus": 2,
      "memory": 1024,
      "disk": 10240,
      "traffic_used": 1073741824,
      "traffic_limit": 107374182400
    }
  ]
}
```

#### 2. 创建容器

```http
POST /api/system/containers
Content-Type: application/json

{
  "hostname": "test1",
  "cpus": 2,
  "memory": 1024,
  "disk": 10240,
  "image": "ubuntu2204",
  "password": "yourpassword",
  "ingress": 100,
  "egress": 100,
  "traffic_limit": 100,
  "cpu_allowance": 50
}
```

#### 3. 容器操作

```http
POST /api/system/containers/{name}/action?action={action_type}
```

支持的 `action_type`：
- `start`: 启动容器
- `stop`: 停止容器
- `restart`: 重启容器
- `reinstall`: 重装系统（需传 `{"image": "ubuntu2204"}`）
- `reset-password`: 重置密码（需传 `{"password": "newpass"}`）

#### 4. 删除容器

```http
DELETE /api/system/containers/{name}
```

#### 5. 获取容器信息

```http
GET /api/system/containers/{name}
```

#### 6. 获取访问凭证

```http
GET /api/system/containers/{name}/credential
```

#### 7. 重置流量

```http
POST /api/system/traffic/reset?name={container_name}
```

#### 8. 系统统计

```http
GET /api/system/stats
```

## 与财务系统集成

### WHMCS 插件配置

1. 将 `Fmis/whmcs/lxdapiserver` 目录复制到 WHMCS 的 `modules/servers/` 目录
2. 在 WHMCS 后台 → 系统设置 → 服务器 → 添加新服务器：
   - **服务器类型**：lxdapiserver
   - **主机名**：您的服务器 IP 或域名
   - **API Hash**：与 config.yaml 中的 `api_hash` 一致
3. 创建产品时选择该服务器即可

### 魔方财务配置

类似 WHMCS，将对应插件复制到魔方财务的插件目录并配置。

## 项目结构

```
openlxd-backend-v2/
├── cmd/
│   └── main.go              # 主程序入口
├── internal/
│   ├── lxd/
│   │   ├── client.go        # LXD HTTP API 客户端
│   │   ├── container.go     # 容器操作
│   │   ├── utils.go         # 工具函数（IP获取、密码重置）
│   │   ├── traffic.go       # 流量监控
│   │   └── nat.go           # NAT 端口映射
│   └── models/
│       ├── models.go        # 数据库模型
│       └── db.go            # 数据库初始化
├── configs/
│   └── config.yaml          # 配置文件
├── web/
│   └── templates/
│       └── index.html       # Web 管理界面
├── lxdapi.db                # SQLite 数据库（自动生成）
└── README.md
```

## 故障排查

### 1. 无法连接到 LXD

**现象**：日志显示 "无法连接到 LXD"

**解决方案**：
- 检查 LXD 是否已安装：`lxd version`
- 检查 Socket 路径：`ls -la /var/snap/lxd/common/lxd/unix.socket`
- 确保以 root 权限运行后端：`sudo ./openlxd`
- 如果不需要真实容器管理，可以忽略此警告（Mock 模式）

### 2. 端口映射不生效

**现象**：外部无法访问容器端口

**解决方案**：
- 检查 iptables 规则：`iptables -t nat -L -n -v`
- 确保 IP 转发已开启：`echo 1 > /proc/sys/net/ipv4/ip_forward`
- 检查防火墙规则：`ufw status`

### 3. 流量统计不准确

**现象**：流量数据异常

**解决方案**：
- 检查流量监控是否启动：查看日志中的 "流量监控已启动"
- 调整采集间隔（在 main.go 中修改 `NewTrafficMonitor(300)`）
- 检查 LXD 容器网络接口状态

### 4. API 返回 401 Unauthorized

**现象**：所有 API 请求都返回 401

**解决方案**：
- 检查 `X-API-Hash` Header 是否正确
- 确认 config.yaml 中的 `api_hash` 配置
- 查看后端日志中的 "API Hash" 输出

## 开发指南

### 添加新功能

1. 在 `internal/lxd/` 中实现核心逻辑
2. 在 `cmd/main.go` 中添加 API 路由
3. 更新数据库模型（如需要）
4. 编写测试

### 调试模式

在 `cmd/main.go` 中启用详细日志：

```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 致谢

本项目基于 [xkatld/lxdapi-web-server](https://github.com/xkatld/lxdapi-web-server) 的 API 规范开发，感谢原作者的开创性工作。

## 联系方式

- 项目主页：https://github.com/areyouokbro/openlxd
- 问题反馈：https://github.com/areyouokbro/openlxd/issues
- 文档：https://github.com/areyouokbro/openlxd/wiki
