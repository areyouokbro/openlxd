#!/bin/bash

# OpenLXD 一键安装脚本
# 适用于 Ubuntu/Debian 系统

set -e

echo "================================"
echo "OpenLXD v3.6.0 一键安装脚本"
echo "================================"
echo ""

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then 
    echo "请使用 root 权限运行此脚本"
    echo "使用方法: sudo bash install.sh"
    exit 1
fi

# 检测系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VER=$VERSION_ID
else
    echo "无法检测操作系统"
    exit 1
fi

echo "检测到系统: $OS $VER"
echo ""

# 1. 安装 LXD（如果未安装）
echo "步骤 1/5: 检查 LXD..."
if ! command -v lxd &> /dev/null; then
    echo "LXD 未安装，正在安装..."
    
    # 检测是否有 snap
    if command -v snap &> /dev/null; then
        echo "使用 snap 安装 LXD..."
        snap install lxd
        lxd init --auto
    else
        # 使用 apt 安装 LXD
        echo "使用 apt 安装 LXD..."
        apt-get update
        apt-get install -y lxd lxd-client
        
        # 初始化 LXD
        cat <<EOF | lxd init --preseed
config: {}
networks:
- config:
    ipv4.address: auto
    ipv6.address: auto
  description: ""
  name: lxdbr0
  type: ""
  project: default
storage_pools:
- config:
    size: 30GB
  description: ""
  name: default
  driver: dir
profiles:
- config: {}
  description: ""
  devices:
    eth0:
      name: eth0
      network: lxdbr0
      type: nic
    root:
      path: /
      pool: default
      type: disk
  name: default
projects: []
cluster: null
EOF
    fi
    
    echo "✓ LXD 安装完成"
else
    echo "✓ LXD 已安装"
fi
echo ""

# 2. 创建安装目录
echo "步骤 2/5: 创建安装目录..."
mkdir -p /opt/openlxd
mkdir -p /etc/openlxd
mkdir -p /var/lib/openlxd
mkdir -p /var/log/openlxd
echo "✓ 目录创建完成"
echo ""

# 3. 下载 OpenLXD
echo "步骤 3/5: 下载 OpenLXD..."
DOWNLOAD_URL="https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd"

if command -v wget &> /dev/null; then
    wget -O /opt/openlxd/openlxd "$DOWNLOAD_URL"
elif command -v curl &> /dev/null; then
    curl -L -o /opt/openlxd/openlxd "$DOWNLOAD_URL"
else
    echo "错误: 需要 wget 或 curl"
    exit 1
fi

chmod +x /opt/openlxd/openlxd
echo "✓ OpenLXD 下载完成"
echo ""

# 4. 创建配置文件
echo "步骤 4/5: 创建配置文件..."
cat > /etc/openlxd/config.yaml << 'EOF'
server:
  port: 8443
  host: "0.0.0.0"
  https: false
  domain: "localhost"
  cert_dir: "/etc/openlxd/certs"
  auto_tls: false

security:
  api_hash: "default-api-key-please-change"
  admin_user: "admin"
  admin_pass: "admin123"
  session_secret: "default-secret-please-change"

database:
  type: "sqlite"
  path: "/var/lib/openlxd/openlxd.db"

lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"
  bridge: "lxdbr0"
EOF
echo "✓ 配置文件创建完成"
echo ""

# 5. 创建 systemd 服务
echo "步骤 5/5: 创建系统服务..."
cat > /etc/systemd/system/openlxd.service << 'EOF'
[Unit]
Description=OpenLXD Container Management System
After=network.target lxd.service
Requires=lxd.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/openlxd
ExecStart=/opt/openlxd/openlxd
Restart=always
RestartSec=5
StandardOutput=append:/var/log/openlxd/openlxd.log
StandardError=append:/var/log/openlxd/openlxd.log

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable openlxd
echo "✓ 系统服务创建完成"
echo ""

echo "================================"
echo "安装完成！"
echo "================================"
echo ""
echo "启动服务:"
echo "  sudo systemctl start openlxd"
echo ""
echo "查看状态:"
echo "  sudo systemctl status openlxd"
echo ""
echo "查看日志:"
echo "  sudo tail -f /var/log/openlxd/openlxd.log"
echo ""
echo "访问 Web 界面:"
echo "  http://$(hostname -I | awk '{print $1}'):8443"
echo ""
echo "默认管理员账户:"
echo "  用户名: admin"
echo "  密码: admin123"
echo ""
echo "重要提示:"
echo "  1. 请修改 /etc/openlxd/config.yaml 中的默认密码"
echo "  2. 首次使用需要创建用户并获取 API Key"
echo "  3. 配置 WHMCS 时使用用户的 API Key"
echo ""
echo "文档: https://github.com/areyouokbro/openlxd"
echo ""
