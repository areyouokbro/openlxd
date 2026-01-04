#!/bin/bash
#================================================================
# OpenLXD Backend 一键安装脚本
# 版本: 2.1
# 作者: OpenLXD Team
# 描述: 自动安装 OpenLXD 后端服务，支持纯净系统零依赖安装
#================================================================

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
GITHUB_REPO="areyouokbro/openlxd"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/openlxd-linux-amd64"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
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
}

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    elif [ -f /etc/redhat-release ]; then
        OS="centos"
    else
        OS="unknown"
    fi
    
    print_info "检测到操作系统: $OS $OS_VERSION"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用 root 用户或 sudo 运行此脚本"
        exit 1
    fi
}

# 安装基础依赖
install_dependencies() {
    print_info "安装系统依赖..."
    
    case $OS in
        ubuntu|debian)
            # 更新软件源（静默处理错误）
            apt-get update -qq 2>/dev/null || true
            
            # 安装必要工具
            DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
                wget \
                curl \
                ca-certificates \
                file \
                >/dev/null 2>&1
            ;;
        centos|rhel|rocky|almalinux)
            yum install -y -q \
                wget \
                curl \
                ca-certificates \
                file \
                >/dev/null 2>&1
            ;;
        *)
            print_warning "未知的操作系统，尝试继续..."
            ;;
    esac
    
    print_success "依赖安装完成"
}

# 生成随机 API Key
generate_api_key() {
    if command -v openssl &> /dev/null; then
        openssl rand -hex 32
    else
        # 如果没有 openssl，使用 /dev/urandom
        cat /dev/urandom | tr -dc 'a-f0-9' | fold -w 64 | head -n 1
    fi
}

# 下载二进制文件
download_binary() {
    print_info "从 GitHub 下载最新版本..."
    print_info "下载地址: $DOWNLOAD_URL"
    print_info "文件大小: ~16MB，请耐心等待..."
    
    # 创建临时下载目录
    TMP_FILE="/tmp/openlxd-download-$$"
    
    # 使用 wget 下载（带进度条和重试）
    if command -v wget &> /dev/null; then
        wget --timeout=60 \
             --tries=3 \
             --show-progress \
             --progress=bar:force \
             -O "$TMP_FILE" \
             "$DOWNLOAD_URL" 2>&1 | grep -v "^$"
    elif command -v curl &> /dev/null; then
        curl -L --max-time 60 \
             --retry 3 \
             --progress-bar \
             -o "$TMP_FILE" \
             "$DOWNLOAD_URL"
    else
        print_error "未找到 wget 或 curl，无法下载文件"
        print_info "请手动安装: apt-get install wget 或 yum install wget"
        exit 1
    fi
    
    # 检查下载是否成功
    if [ ! -f "$TMP_FILE" ]; then
        print_error "下载失败"
        exit 1
    fi
    
    # 检查文件大小
    FILE_SIZE=$(stat -c%s "$TMP_FILE" 2>/dev/null || stat -f%z "$TMP_FILE" 2>/dev/null || echo "0")
    if [ "$FILE_SIZE" -lt 10000000 ]; then
        print_error "下载的文件大小不正确 ($FILE_SIZE bytes)"
        rm -f "$TMP_FILE"
        exit 1
    fi
    
    print_success "下载完成"
    
    # 移动到目标位置
    mv "$TMP_FILE" "$BIN_PATH"
    chmod +x "$BIN_PATH"
}

# 创建安装目录
create_directories() {
    print_info "创建安装目录..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "/var/log/openlxd"
    
    print_success "目录创建完成"
}

# 创建配置文件
create_config() {
    print_info "配置 OpenLXD..."
    
    # 生成 API Key
    API_KEY=$(generate_api_key)
    
    # 获取服务器 IP
    SERVER_IP=$(hostname -I | awk '{print $1}')
    
    # 创建配置文件
    cat > "$CONFIG_DIR/config.yaml" << EOF
server:
  port: 8443
  host: "0.0.0.0"
  https: true                     # 启用 HTTPS
  domain: ""                      # 留空使用自签名证书
  cert_dir: "/etc/openlxd/certs" # 证书存储目录
  auto_tls: false                 # 使用自签名证书，不使用 Let's Encrypt

security:
  api_hash: "$API_KEY"
  admin_user: "admin"
  admin_pass: "admin123"
  session_secret: "$(generate_api_key)"

database:
  type: "sqlite"
  path: "$INSTALL_DIR/lxdapi.db"

lxd:
  socket: "/var/snap/lxd/common/lxd/unix.socket"
  bridge: "lxdbr0"
  storage_pool: "default"

monitor:
  traffic_interval: 300
  enable_auto_stop: true
EOF
    
    print_success "配置文件创建完成"
    print_warning "API Key: $API_KEY"
    print_warning "请妥善保管此密钥！"
}

# 配置 systemd 服务
setup_systemd() {
    print_info "配置 systemd 服务..."
    
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=OpenLXD Backend Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=$BIN_PATH
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable openlxd >/dev/null 2>&1
    
    print_success "systemd 服务配置完成"
}

# 配置防火墙
setup_firewall() {
    print_info "配置防火墙..."
    
    # UFW
    if command -v ufw &> /dev/null; then
        ufw allow 8443/tcp >/dev/null 2>&1
        print_success "UFW 防火墙规则已添加 (8443)"
    # firewalld
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=8443/tcp >/dev/null 2>&1
        firewall-cmd --reload >/dev/null 2>&1
        print_success "firewalld 防火墙规则已添加 (8443)"
    # iptables
    elif command -v iptables &> /dev/null; then
        iptables -A INPUT -p tcp --dport 8443 -j ACCEPT >/dev/null 2>&1
        print_success "iptables 防火墙规则已添加 (8443)"
    else
        print_warning "未检测到防火墙，请手动开放 8443 端口"
    fi
}

# 启动服务
start_service() {
    print_info "启动 OpenLXD 服务..."
    
    systemctl start openlxd
    
    # 等待服务启动
    sleep 3
    
    # 检查服务状态
    if systemctl is-active --quiet openlxd; then
        print_success "服务启动成功"
        return 0
    else
        print_error "服务启动失败，请查看日志: journalctl -u openlxd -n 50"
        return 1
    fi
}

# 验证安装
verify_installation() {
    print_info "验证安装..."
    
    # 检查二进制文件
    if [ ! -f "$BIN_PATH" ]; then
        print_error "二进制文件不存在"
        return 1
    fi
    
    # 检查配置文件
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        print_error "配置文件不存在"
        return 1
    fi
    
    # 检查服务状态
    if ! systemctl is-active --quiet openlxd; then
        print_warning "服务未运行"
        return 1
    fi
    
    # 测试 API
    API_KEY=$(grep "api_hash:" "$CONFIG_DIR/config.yaml" | awk '{print $2}' | tr -d '"')
    if command -v curl &> /dev/null; then
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
            -H "X-API-Hash: $API_KEY" \
            http://localhost:8443/api/system/stats)
        
        if [ "$HTTP_CODE" = "200" ]; then
            print_success "API 测试通过"
        else
            print_warning "API 测试失败 (HTTP $HTTP_CODE)"
        fi
    fi
    
    return 0
}

# 显示安装信息
show_install_info() {
    SERVER_IP=$(hostname -I | awk '{print $1}')
    API_KEY=$(grep "api_hash:" "$CONFIG_DIR/config.yaml" | awk '{print $2}' | tr -d '"')
    
    echo ""
    echo "========================================"
    echo "   OpenLXD 安装完成！"
    echo "========================================"
    echo ""
    echo "安装信息:"
    echo "  安装目录: $INSTALL_DIR"
    echo "  配置目录: $CONFIG_DIR"
    echo "  二进制文件: $BIN_PATH"
    echo "  日志文件: /var/log/openlxd/"
    echo ""
    echo "API 信息:"
    echo "  API Key: $API_KEY"
    echo "  API 地址: https://$SERVER_IP:8443"
    echo "  Web 管理: https://$SERVER_IP:8443/admin/login"
    echo ""
    echo "Web 登录凭据:"
    echo "  用户名: admin"
    echo "  密码: admin123"
    echo ""
    echo "服务管理:"
    echo "  启动服务: systemctl start openlxd"
    echo "  停止服务: systemctl stop openlxd"
    echo "  重启服务: systemctl restart openlxd"
    echo "  查看状态: systemctl status openlxd"
    echo "  查看日志: journalctl -u openlxd -f"
    echo ""
    echo "重要提示:"
    echo "  1. 请妃善保管 API Key"
    echo "  2. 建议修改默认管理员密码"
    echo "  3. 默认使用自签名证书，浏览器会提示不安全（忽略即可）"
    echo "  4. WHMCS 对接请使用 https://$SERVER_IP:8443 和上述 API Key"
    echo "  5. WHMCS 中需要禁用 SSL 验证或使用 Nginx 反向代理配置正式证书"
    echo ""
    echo "========================================"
    echo ""
}

# 主函数
main() {
    show_banner
    
    # 检查 root 权限
    check_root
    
    # 检测操作系统
    detect_os
    
    # 安装依赖
    install_dependencies
    
    # 下载二进制文件
    download_binary
    
    # 创建目录
    create_directories
    
    # 创建配置
    create_config
    
    # 配置 systemd
    setup_systemd
    
    # 配置防火墙
    setup_firewall
    
    # 启动服务
    if start_service; then
        # 验证安装
        verify_installation
        
        # 显示安装信息
        show_install_info
        
        print_success "安装完成！"
    else
        print_error "安装过程中出现错误"
        print_info "请查看日志: journalctl -u openlxd -n 50"
        exit 1
    fi
}

# 运行主函数
main
