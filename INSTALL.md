# OpenLXD Backend 安装与编译指南

## 快速部署（推荐）

### 使用预编译二进制文件

我们提供了预编译的二进制文件，无需安装 Go 环境即可直接运行。

#### 1. 选择对应平台的二进制文件

```bash
# Linux x86_64 (amd64)
chmod +x bin/openlxd-linux-amd64
sudo mv bin/openlxd-linux-amd64 /usr/local/bin/openlxd
```

#### 2. 配置文件

```bash
# 创建配置目录
sudo mkdir -p /etc/openlxd

# 复制配置文件模板
sudo cp configs/config.yaml /etc/openlxd/config.yaml

# 编辑配置文件
sudo vim /etc/openlxd/config.yaml
```

**必须修改的配置项**：

```yaml
security:
  api_hash: "change-this-to-your-secret-key"  # ⚠️ 必须修改为您的密钥
```

#### 3. 启动服务

```bash
# 方式一：直接运行（前台）
sudo openlxd

# 方式二：使用 systemd 管理（推荐）
sudo cp systemd/openlxd.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable openlxd
sudo systemctl start openlxd
sudo systemctl status openlxd
```

#### 4. 验证安装

```bash
# 测试 API 接口
curl -H "X-API-Hash: your-secret-key" http://localhost:8443/api/system/stats

# 访问 Web 管理界面
# 浏览器打开: http://您的服务器IP:8443
```

---

## 从源码编译

如果您需要自定义编译或贡献代码，请按照以下步骤操作。

### 环境要求

- **Go**: 1.18 或更高版本（推荐 1.22）
- **GCC**: 用于编译 SQLite（CGO 依赖）
- **操作系统**: Linux（推荐 Ubuntu 22.04）

### 安装 Go 环境

#### Ubuntu/Debian

```bash
# 方式一：使用 apt（可能版本较旧）
sudo apt update
sudo apt install -y golang-go gcc

# 方式二：手动安装最新版本（推荐）
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 编译步骤

#### 1. 克隆或下载源码

```bash
cd /opt
git clone https://github.com/yourusername/openlxd-backend.git
cd openlxd-backend
```

#### 2. 下载依赖

```bash
go mod tidy
```

**常见问题**：

如果遇到 `go: cannot find module` 错误，请确保：
- Go 版本 >= 1.18
- 已安装 GCC：`sudo apt install gcc`
- 网络连接正常（需要下载依赖包）

#### 3. 编译

```bash
# 编译当前平台
CGO_ENABLED=1 go build -o openlxd cmd/main.go

# 编译 Linux amd64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o openlxd-linux-amd64 cmd/main.go

# 编译 Linux arm64（需要交叉编译工具链）
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o openlxd-linux-arm64 cmd/main.go
```

#### 4. 安装

```bash
sudo cp openlxd /usr/local/bin/
sudo chmod +x /usr/local/bin/openlxd
```

---

## 配置文件详解

配置文件位置：`/etc/openlxd/config.yaml` 或 `configs/config.yaml`

### 完整配置示例

```yaml
# 服务器配置
server:
  port: 8443                    # API 监听端口
  host: "0.0.0.0"              # 监听地址（0.0.0.0 表示所有网卡）

# 安全配置
security:
  api_hash: "your-secret-key-here"  # API 密钥（必须修改）
  admin_user: "admin"                # Web 管理员用户名
  admin_pass: "admin123"             # Web 管理员密码
  session_secret: "random-string"    # Session 密钥

# 数据库配置
database:
  type: "sqlite"                # 数据库类型（目前仅支持 sqlite）
  path: "lxdapi.db"            # 数据库文件路径

# LXD 配置
lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"  # LXD Unix Socket 路径
  bridge: "lxdbr0"                                 # LXD 网桥名称
  storage_pool: "default"                          # 存储池名称

# 监控配置
monitor:
  traffic_interval: 300        # 流量采集间隔（秒），默认 5 分钟
  enable_auto_stop: true       # 是否启用流量超限自动停机
```

---

## systemd 服务配置

创建服务文件 `/etc/systemd/system/openlxd.service`：

```ini
[Unit]
Description=OpenLXD Backend Service
Documentation=https://github.com/yourusername/openlxd-backend
After=network.target lxd.service
Wants=lxd.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/openlxd-backend
ExecStart=/usr/local/bin/openlxd
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

# 环境变量
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

[Install]
WantedBy=multi-user.target
```

管理服务：

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start openlxd

# 停止服务
sudo systemctl stop openlxd

# 重启服务
sudo systemctl restart openlxd

# 查看状态
sudo systemctl status openlxd

# 查看日志
sudo journalctl -u openlxd -f

# 开机自启
sudo systemctl enable openlxd
```

---

## 故障排查

### 1. 编译错误：`undefined: sqlite3.Error`

**原因**：CGO 未启用或 GCC 未安装

**解决方案**：

```bash
# 安装 GCC
sudo apt install -y gcc

# 启用 CGO 编译
CGO_ENABLED=1 go build -o openlxd cmd/main.go
```

### 2. 运行错误：`cannot open shared object file`

**原因**：缺少动态库

**解决方案**：

```bash
# Ubuntu/Debian
sudo apt install -y libc6

# 或使用静态编译
CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o openlxd cmd/main.go
```

### 3. 无法连接到 LXD

**现象**：日志显示 "无法连接到 LXD"

**解决方案**：

```bash
# 检查 LXD 是否安装
lxd version

# 检查 Socket 文件
ls -la /var/snap/lxd/common/lxd/unix.socket

# 确保以 root 权限运行
sudo openlxd
```

**注意**：如果没有 LXD 环境，后端会自动切换到 Mock 模式，API 接口仍然可用。

### 4. API 返回 401 Unauthorized

**原因**：API 密钥不匹配

**解决方案**：

```bash
# 检查配置文件中的 api_hash
cat /etc/openlxd/config.yaml | grep api_hash

# 检查后端日志中的 API Hash
sudo journalctl -u openlxd | grep "API Hash"

# 确保请求 Header 正确
curl -H "X-API-Hash: your-actual-key" http://localhost:8443/api/system/stats
```

### 5. 端口被占用

**现象**：启动时报错 "address already in use"

**解决方案**：

```bash
# 查找占用 8443 端口的进程
sudo lsof -i :8443

# 或
sudo netstat -tlnp | grep 8443

# 停止占用端口的进程
sudo kill <PID>

# 或修改配置文件中的端口
vim /etc/openlxd/config.yaml
```

---

## 性能优化

### 1. 数据库优化

如果容器数量较多，建议调整 SQLite 配置：

```go
// 在 internal/models/db.go 中添加
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA cache_size=10000")
```

### 2. 流量监控间隔

根据实际需求调整流量采集间隔：

```yaml
monitor:
  traffic_interval: 300  # 默认 5 分钟，可调整为 60（1分钟）或 600（10分钟）
```

### 3. 日志级别

生产环境建议减少日志输出：

```go
// 在 cmd/main.go 中
log.SetFlags(log.LstdFlags)  // 仅输出时间戳
```

---

## 升级指南

### 从旧版本升级

```bash
# 1. 停止服务
sudo systemctl stop openlxd

# 2. 备份数据库
sudo cp /opt/openlxd-backend/lxdapi.db /opt/openlxd-backend/lxdapi.db.backup

# 3. 替换二进制文件
sudo cp bin/openlxd-linux-amd64 /usr/local/bin/openlxd

# 4. 重启服务
sudo systemctl start openlxd

# 5. 验证
sudo systemctl status openlxd
```

---

## 开发环境搭建

### 1. 安装开发工具

```bash
# 安装 Go 工具链
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# 安装代码格式化工具
go install golang.org/x/tools/cmd/goimports@latest
```

### 2. 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/lxd

# 查看覆盖率
go test -cover ./...
```

### 3. 调试

```bash
# 使用 delve 调试
dlv debug cmd/main.go
```

---

## 卸载

```bash
# 停止并禁用服务
sudo systemctl stop openlxd
sudo systemctl disable openlxd

# 删除服务文件
sudo rm /etc/systemd/system/openlxd.service
sudo systemctl daemon-reload

# 删除二进制文件
sudo rm /usr/local/bin/openlxd

# 删除配置文件
sudo rm -rf /etc/openlxd

# 删除数据库（可选）
sudo rm /opt/openlxd-backend/lxdapi.db
```

---

## 获取帮助

- **项目主页**：https://github.com/yourusername/openlxd-backend
- **问题反馈**：https://github.com/yourusername/openlxd-backend/issues
- **文档**：https://github.com/yourusername/openlxd-backend/wiki
