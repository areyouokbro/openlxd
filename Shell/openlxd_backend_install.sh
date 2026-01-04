#!/bin/bash

# OpenLXD - 后端服务一键安装脚本
# 支持系统: Ubuntu 20.04+, Debian 11+

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}开始安装 OpenLXD 后端服务...${NC}"

# 1. 安装 Go 环境 (如果未安装)
if ! command -v go &> /dev/null; then
    echo -e "${GREEN}正在安装 Go 环境...${NC}"
    apt-get update
    apt-get install -y golang-go
fi

# 2. 准备工作目录
INSTALL_DIR="/opt/openlxd"
mkdir -p $INSTALL_DIR

# 3. 编译后端程序
echo -e "${GREEN}正在编译后端程序...${NC}"
# 假设脚本在项目根目录下运行，或者从远程拉取
# 这里演示从当前目录的 Backend 文件夹编译
if [ -d "./Backend" ]; then
    cd ./Backend
    go build -o $INSTALL_DIR/openlxd-backend cmd/main.go
    cd ..
else
    echo -e "${RED}错误: 未找到 Backend 源码目录。请在项目根目录下运行此脚本。${NC}"
    exit 1
fi

# 4. 创建 systemd 服务
echo -e "${GREEN}正在配置 systemd 服务...${NC}"
cat <<EOF > /etc/systemd/system/openlxd.service
[Unit]
Description=OpenLXD Backend Service
After=network.target lxd.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/openlxd-backend
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 5. 启动服务
systemctl daemon-reload
systemctl enable openlxd
systemctl start openlxd

echo -e "${GREEN}OpenLXD 后端服务安装完成并已启动！${NC}"
echo -e "${GREEN}服务状态:${NC}"
systemctl status openlxd --no-pager
