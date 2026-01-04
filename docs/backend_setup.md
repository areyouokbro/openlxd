# 后端部署手册

OpenLXD 后端是整个系统的核心，负责与 LXD 通信、管理数据库以及处理来自财务系统的请求。

## 环境要求

- **操作系统**: Ubuntu 20.04+ 或 Debian 11+
- **Go 语言**: 1.18 或更高版本
- **LXD/Incus**: 已安装并初始化的 LXD 环境
- **权限**: 需要 root 权限以管理 iptables 和访问 LXD Unix Socket

## 安装步骤

### 方法一：使用一键安装脚本 (推荐)

1. **安装 LXD 环境**:
   ```bash
   sudo bash Shell/lxd_install.sh
   ```
2. **安装 OpenLXD 后端**:
   ```bash
   sudo bash Shell/openlxd_backend_install.sh
   ```

### 方法二：手动安装

1. **克隆源码**:
   ```bash
   git clone <your-repo-url>
   cd Backend
   ```
2. **安装依赖**:
   ```bash
   go mod tidy
   ```
3. **编译并运行**:
   ```bash
   go build -o openlxd cmd/main.go
   sudo ./openlxd
   ```
服务默认监听 `8443` 端口。

## 关键配置

### 数据库
系统默认使用 SQLite 数据库 (`lxdapi.db`)。如需切换至 MySQL，请修改 `internal/models/db.go` 中的连接逻辑。

### 网络设置
- **NAT 转发**: 系统会自动调用 `iptables`。请确保系统已安装 `iptables` 且内核支持 NAT。
- **LXD 桥接**: 默认使用 `lxdbr0`。如果您的网桥名称不同，请在 `internal/lxd/network.go` 中修改。

## 生产环境建议
- 使用 `systemd` 管理服务进程。
- 使用 Nginx 进行反向代理并配置 SSL 证书。
- 定期备份 `lxdapi.db` 数据库文件。
