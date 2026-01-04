#!/bin/bash
#================================================================
# OpenLXD Backend 一键安装脚本
# 版本: 2.0
# 作者: OpenLXD Team
# 描述: 自动安装 OpenLXD 后端服务，支持多种安装方式
#================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
INSTALL_DIR="/opt/openlxd"
CONFIG_DIR="/etc/openlxd"
BIN_PATH="/usr/local/bin/openlxd"
SERVICE_FILE="/etc/systemd/system/openlxd.service"
LOG_FILE="/var/log/openlxd-install.log"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# 显示横幅
show_banner() {
    clear
    echo -e "${GREEN}"
    cat << "EOF"
   ____                   _    __  _______ 
  / __ \____  ___  ____  | |  / / |  __ \ 
 / / / / __ \/ _ \/ __ \ | | / /  | |  | |
/ /_/ / /_/ /  __/ / / / | |/ /   | |  | |
\____/ .___/\___/_/ /_/  |___/    |_____/ 
    /_/                                    
    
OpenLXD Backend - 开源 LXD 容器管理后端
版本: 2.0 | 完全开源 | 生产级
EOF
    echo -e "${NC}"
    echo ""
}

# 检查 root 权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "此脚本必须以 root 权限运行"
        echo "请使用: sudo $0"
        exit 1
    fi
}

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VER=$VERSION_ID
    else
        print_error "无法检测操作系统"
        exit 1
    fi
    
    print_info "检测到操作系统: $OS $VER"
}

# 检查系统要求
check_requirements() {
    print_info "检查系统要求..."
    
    # 检查内存
    total_mem=$(free -m | awk '/^Mem:/{print $2}')
    if [ "$total_mem" -lt 1024 ]; then
        print_warning "内存不足 1GB，可能影响性能"
    fi
    
    # 检查磁盘空间
    available_space=$(df -m / | awk 'NR==2 {print $4}')
    if [ "$available_space" -lt 1024 ]; then
        print_error "磁盘空间不足 1GB"
        exit 1
    fi
    
    print_success "系统要求检查通过"
}

# 安装依赖
install_dependencies() {
    print_info "安装系统依赖..."
    
    case $OS in
        ubuntu|debian)
            apt-get update -qq
            apt-get install -y curl wget tar gzip unzip iptables sqlite3 net-tools >> "$LOG_FILE" 2>&1
            ;;
        centos|rhel|rocky)
            yum install -y curl wget tar gzip unzip iptables sqlite net-tools >> "$LOG_FILE" 2>&1
            ;;
        *)
            print_warning "未知的操作系统，跳过依赖安装"
            ;;
    esac
    
    print_success "依赖安装完成"
}

# 选择安装方式
select_install_method() {
    echo ""
    echo "请选择安装方式:"
    echo "1) 使用预编译二进制文件（推荐，快速）"
    echo "2) 从源码编译（需要 Go 环境）"
    echo "3) 从 GitHub 下载最新版本"
    echo ""
    read -p "请输入选项 [1-3]: " choice
    
    case $choice in
        1) install_from_binary ;;
        2) install_from_source ;;
        3) install_from_github ;;
        *) print_error "无效的选项"; exit 1 ;;
    esac
}

# 从二进制文件安装
install_from_binary() {
    print_info "从预编译二进制文件安装..."
    
    # 检查是否存在二进制文件
    if [ -f "./bin/openlxd-linux-amd64" ]; then
        cp ./bin/openlxd-linux-amd64 "$BIN_PATH"
        chmod +x "$BIN_PATH"
        print_success "二进制文件安装完成"
    else
        print_error "未找到预编译二进制文件"
        print_info "尝试从 GitHub 下载..."
        install_from_github
    fi
}

# 从源码编译安装
install_from_source() {
    print_info "从源码编译安装..."
    
    # 检查 Go 是否已安装
    if ! command -v go &> /dev/null; then
        print_info "未检测到 Go 环境，正在安装..."
        install_golang
    fi
    
    # 编译
    print_info "开始编译..."
    cd "$(dirname "$0")/.."
    CGO_ENABLED=1 go build -o "$BIN_PATH" cmd/main.go >> "$LOG_FILE" 2>&1
    
    if [ $? -eq 0 ]; then
        chmod +x "$BIN_PATH"
        print_success "编译完成"
    else
        print_error "编译失败，请查看日志: $LOG_FILE"
        exit 1
    fi
}

# 安装 Go 环境
install_golang() {
    GO_VERSION="1.22.0"
    GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    
    print_info "下载 Go ${GO_VERSION}..."
    wget -q --show-progress "$GO_URL" -O /tmp/go.tar.gz
    
    print_info "安装 Go..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    # 添加到 PATH
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    
    print_success "Go 安装完成"
}

# 从 GitHub 下载
install_from_github() {
    print_info "从 GitHub 下载最新版本..."
    
    GITHUB_REPO="yourusername/openlxd-backend"
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/openlxd-linux-amd64"
    
    wget -q --show-progress "$DOWNLOAD_URL" -O "$BIN_PATH"
    
    if [ $? -eq 0 ]; then
        chmod +x "$BIN_PATH"
        print_success "下载完成"
    else
        print_error "下载失败"
        exit 1
    fi
}

# 配置安装目录
setup_directories() {
    print_info "创建安装目录..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p /var/log/openlxd
    
    print_success "目录创建完成"
}

# 配置文件
setup_config() {
    print_info "配置 OpenLXD..."
    
    # 生成随机 API Key
    API_KEY=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)
    
    # 创建配置文件
    cat > "$CONFIG_DIR/config.yaml" <<EOF
# OpenLXD Backend 配置文件
# 自动生成于: $(date)

server:
  port: 8443
  host: "0.0.0.0"

security:
  api_hash: "${API_KEY}"
  admin_user: "admin"
  admin_pass: "admin123"
  session_secret: "$(openssl rand -hex 16 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)"

database:
  type: "sqlite"
  path: "${INSTALL_DIR}/lxdapi.db"

lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"
  bridge: "lxdbr0"
  storage_pool: "default"

monitor:
  traffic_interval: 300
  enable_auto_stop: true
EOF
    
    print_success "配置文件创建完成"
    print_warning "API Key: ${API_KEY}"
    print_warning "请妥善保管此密钥！"
    
    # 保存密钥到文件
    echo "$API_KEY" > "$CONFIG_DIR/.api_key"
    chmod 600 "$CONFIG_DIR/.api_key"
}

# 配置 systemd 服务
setup_systemd() {
    print_info "配置 systemd 服务..."
    
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=OpenLXD Backend Service
Documentation=https://github.com/yourusername/openlxd-backend
After=network.target lxd.service
Wants=lxd.service

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${BIN_PATH}
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
Environment="CONFIG_PATH=${CONFIG_DIR}/config.yaml"

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    print_success "systemd 服务配置完成"
}

# 配置防火墙
setup_firewall() {
    print_info "配置防火墙..."
    
    # 检查防火墙类型
    if command -v ufw &> /dev/null; then
        ufw allow 8443/tcp >> "$LOG_FILE" 2>&1
        print_success "UFW 防火墙规则已添加"
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=8443/tcp >> "$LOG_FILE" 2>&1
        firewall-cmd --reload >> "$LOG_FILE" 2>&1
        print_success "firewalld 防火墙规则已添加"
    else
        print_warning "未检测到防火墙，请手动开放 8443 端口"
    fi
}

# 启动服务
start_service() {
    print_info "启动 OpenLXD 服务..."
    
    systemctl enable openlxd >> "$LOG_FILE" 2>&1
    systemctl start openlxd
    
    sleep 3
    
    if systemctl is-active --quiet openlxd; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败，请查看日志: journalctl -u openlxd -n 50"
        exit 1
    fi
}

# 验证安装
verify_installation() {
    print_info "验证安装..."
    
    API_KEY=$(cat "$CONFIG_DIR/.api_key")
    
    # 测试 API
    response=$(curl -s -H "X-API-Hash: $API_KEY" http://localhost:8443/api/system/stats 2>/dev/null || echo "error")
    
    if echo "$response" | grep -q "code"; then
        print_success "API 测试通过"
    else
        print_warning "API 测试失败，但服务可能仍在启动中"
    fi
}

# 显示安装信息
show_install_info() {
    API_KEY=$(cat "$CONFIG_DIR/.api_key")
    SERVER_IP=$(hostname -I | awk '{print $1}')
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   OpenLXD 安装完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}安装信息:${NC}"
    echo "  安装目录: $INSTALL_DIR"
    echo "  配置目录: $CONFIG_DIR"
    echo "  二进制文件: $BIN_PATH"
    echo "  日志文件: /var/log/openlxd/"
    echo ""
    echo -e "${BLUE}API 信息:${NC}"
    echo "  API Key: $API_KEY"
    echo "  API 地址: http://${SERVER_IP}:8443"
    echo "  Web 管理: http://${SERVER_IP}:8443"
    echo ""
    echo -e "${BLUE}服务管理:${NC}"
    echo "  启动服务: systemctl start openlxd"
    echo "  停止服务: systemctl stop openlxd"
    echo "  重启服务: systemctl restart openlxd"
    echo "  查看状态: systemctl status openlxd"
    echo "  查看日志: journalctl -u openlxd -f"
    echo ""
    echo -e "${BLUE}管理工具:${NC}"
    echo "  管理脚本: openlxd-cli"
    echo "  配置文件: $CONFIG_DIR/config.yaml"
    echo ""
    echo -e "${YELLOW}重要提示:${NC}"
    echo "  1. 请妥善保管 API Key"
    echo "  2. 建议修改默认管理员密码"
    echo "  3. 如需集成财务系统，请使用上述 API Key"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo ""
}

# 主函数
main() {
    show_banner
    check_root
    detect_os
    check_requirements
    
    print_info "开始安装 OpenLXD Backend..."
    echo ""
    
    install_dependencies
    select_install_method
    setup_directories
    setup_config
    setup_systemd
    setup_firewall
    start_service
    verify_installation
    
    show_install_info
    
    print_success "安装完成！"
}

# 执行主函数
main "$@"
