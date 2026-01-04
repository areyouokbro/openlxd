#!/bin/bash
#================================================================
# OpenLXD 管理工具
# 版本: 2.0
# 描述: OpenLXD 后端的命令行管理工具
#================================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
SERVICE_NAME="openlxd"
CONFIG_DIR="/etc/openlxd"
LOG_DIR="/var/log/openlxd"
DB_PATH="/opt/openlxd/lxdapi.db"

# 打印函数
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 显示菜单
show_menu() {
    clear
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   OpenLXD 管理工具${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "1)  启动服务"
    echo "2)  停止服务"
    echo "3)  重启服务"
    echo "4)  查看状态"
    echo "5)  查看日志（实时）"
    echo "6)  查看日志（最近50行）"
    echo "7)  查看配置"
    echo "8)  编辑配置"
    echo "9)  查看 API Key"
    echo "10) 重新生成 API Key"
    echo "11) 数据库备份"
    echo "12) 数据库恢复"
    echo "13) 清理日志"
    echo "14) 系统信息"
    echo "15) 性能监控"
    echo "16) 容器列表"
    echo "17) 测试 API"
    echo "18) 更新服务"
    echo "19) 卸载服务"
    echo "0)  退出"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo ""
}

# 启动服务
start_service() {
    print_info "启动 OpenLXD 服务..."
    systemctl start $SERVICE_NAME
    sleep 2
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败"
        systemctl status $SERVICE_NAME --no-pager
    fi
}

# 停止服务
stop_service() {
    print_info "停止 OpenLXD 服务..."
    systemctl stop $SERVICE_NAME
    sleep 1
    if ! systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务已停止"
    else
        print_error "服务停止失败"
    fi
}

# 重启服务
restart_service() {
    print_info "重启 OpenLXD 服务..."
    systemctl restart $SERVICE_NAME
    sleep 2
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务重启成功"
    else
        print_error "服务重启失败"
        systemctl status $SERVICE_NAME --no-pager
    fi
}

# 查看状态
show_status() {
    print_info "服务状态:"
    echo ""
    systemctl status $SERVICE_NAME --no-pager
    echo ""
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务运行中"
        
        # 显示端口监听
        print_info "端口监听:"
        netstat -tlnp 2>/dev/null | grep 8443 || ss -tlnp | grep 8443
        
        # 显示进程信息
        print_info "进程信息:"
        ps aux | grep openlxd | grep -v grep
    else
        print_warning "服务未运行"
    fi
}

# 查看实时日志
show_logs_live() {
    print_info "实时日志（按 Ctrl+C 退出）:"
    journalctl -u $SERVICE_NAME -f
}

# 查看最近日志
show_logs_recent() {
    print_info "最近 50 行日志:"
    journalctl -u $SERVICE_NAME -n 50 --no-pager
}

# 查看配置
show_config() {
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        print_info "当前配置:"
        echo ""
        cat "$CONFIG_DIR/config.yaml"
    else
        print_error "配置文件不存在"
    fi
}

# 编辑配置
edit_config() {
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        ${EDITOR:-vim} "$CONFIG_DIR/config.yaml"
        print_warning "配置已修改，需要重启服务生效"
        read -p "是否立即重启服务? (y/n): " choice
        if [ "$choice" = "y" ]; then
            restart_service
        fi
    else
        print_error "配置文件不存在"
    fi
}

# 查看 API Key
show_api_key() {
    if [ -f "$CONFIG_DIR/.api_key" ]; then
        API_KEY=$(cat "$CONFIG_DIR/.api_key")
        print_info "当前 API Key:"
        echo ""
        echo -e "${GREEN}${API_KEY}${NC}"
        echo ""
    else
        print_warning "API Key 文件不存在，从配置文件读取..."
        API_KEY=$(grep "api_hash:" "$CONFIG_DIR/config.yaml" | awk '{print $2}' | tr -d '"')
        echo ""
        echo -e "${GREEN}${API_KEY}${NC}"
        echo ""
    fi
}

# 重新生成 API Key
regenerate_api_key() {
    print_warning "重新生成 API Key 将使旧的密钥失效"
    read -p "确定要继续吗? (y/n): " choice
    
    if [ "$choice" = "y" ]; then
        NEW_KEY=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)
        
        # 更新配置文件
        sed -i "s/api_hash:.*/api_hash: \"${NEW_KEY}\"/" "$CONFIG_DIR/config.yaml"
        
        # 保存到文件
        echo "$NEW_KEY" > "$CONFIG_DIR/.api_key"
        chmod 600 "$CONFIG_DIR/.api_key"
        
        print_success "新的 API Key:"
        echo ""
        echo -e "${GREEN}${NEW_KEY}${NC}"
        echo ""
        print_warning "请更新财务系统中的 API Key"
        
        restart_service
    fi
}

# 数据库备份
backup_database() {
    if [ -f "$DB_PATH" ]; then
        BACKUP_DIR="/opt/openlxd/backups"
        mkdir -p "$BACKUP_DIR"
        
        BACKUP_FILE="$BACKUP_DIR/lxdapi_$(date +%Y%m%d_%H%M%S).db"
        
        print_info "备份数据库..."
        cp "$DB_PATH" "$BACKUP_FILE"
        
        if [ -f "$BACKUP_FILE" ]; then
            print_success "备份完成: $BACKUP_FILE"
            
            # 压缩备份
            gzip "$BACKUP_FILE"
            print_success "已压缩: ${BACKUP_FILE}.gz"
        else
            print_error "备份失败"
        fi
    else
        print_error "数据库文件不存在"
    fi
}

# 数据库恢复
restore_database() {
    BACKUP_DIR="/opt/openlxd/backups"
    
    if [ ! -d "$BACKUP_DIR" ]; then
        print_error "备份目录不存在"
        return
    fi
    
    print_info "可用的备份:"
    echo ""
    ls -lh "$BACKUP_DIR"/*.db.gz 2>/dev/null || echo "无备份文件"
    echo ""
    
    read -p "请输入备份文件完整路径: " backup_file
    
    if [ -f "$backup_file" ]; then
        print_warning "恢复数据库将覆盖当前数据"
        read -p "确定要继续吗? (y/n): " choice
        
        if [ "$choice" = "y" ]; then
            stop_service
            
            # 备份当前数据库
            cp "$DB_PATH" "${DB_PATH}.before_restore"
            
            # 恢复
            if [[ "$backup_file" == *.gz ]]; then
                gunzip -c "$backup_file" > "$DB_PATH"
            else
                cp "$backup_file" "$DB_PATH"
            fi
            
            print_success "数据库已恢复"
            start_service
        fi
    else
        print_error "备份文件不存在"
    fi
}

# 清理日志
clean_logs() {
    print_info "清理日志..."
    
    # 清理 systemd 日志
    journalctl --vacuum-time=7d >> /dev/null 2>&1
    
    # 清理应用日志
    if [ -d "$LOG_DIR" ]; then
        find "$LOG_DIR" -name "*.log" -mtime +7 -delete
    fi
    
    print_success "日志清理完成"
}

# 系统信息
show_system_info() {
    print_info "系统信息:"
    echo ""
    
    echo "操作系统: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
    echo "内核版本: $(uname -r)"
    echo "CPU 核心: $(nproc)"
    echo "总内存: $(free -h | awk '/^Mem:/ {print $2}')"
    echo "可用内存: $(free -h | awk '/^Mem:/ {print $7}')"
    echo "磁盘使用: $(df -h / | awk 'NR==2 {print $5}')"
    echo ""
    
    print_info "OpenLXD 信息:"
    echo ""
    echo "服务状态: $(systemctl is-active $SERVICE_NAME)"
    echo "启动时间: $(systemctl show $SERVICE_NAME --property=ActiveEnterTimestamp --value)"
    echo "配置目录: $CONFIG_DIR"
    echo "数据库大小: $(du -h $DB_PATH 2>/dev/null | cut -f1 || echo '未知')"
    echo ""
}

# 性能监控
show_performance() {
    print_info "性能监控（按 Ctrl+C 退出）:"
    echo ""
    
    while true; do
        clear
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}   OpenLXD 性能监控${NC}"
        echo -e "${GREEN}========================================${NC}"
        echo ""
        
        # CPU 使用率
        cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
        echo "CPU 使用率: ${cpu_usage}%"
        
        # 内存使用
        mem_info=$(free -h | awk '/^Mem:/ {print $3 "/" $2}')
        echo "内存使用: $mem_info"
        
        # OpenLXD 进程
        if pgrep openlxd > /dev/null; then
            openlxd_cpu=$(ps aux | grep openlxd | grep -v grep | awk '{sum+=$3} END {print sum}')
            openlxd_mem=$(ps aux | grep openlxd | grep -v grep | awk '{sum+=$4} END {print sum}')
            echo "OpenLXD CPU: ${openlxd_cpu}%"
            echo "OpenLXD MEM: ${openlxd_mem}%"
        else
            echo "OpenLXD: 未运行"
        fi
        
        echo ""
        echo "更新时间: $(date '+%Y-%m-%d %H:%M:%S')"
        
        sleep 3
    done
}

# 容器列表
show_containers() {
    API_KEY=$(cat "$CONFIG_DIR/.api_key" 2>/dev/null)
    
    if [ -z "$API_KEY" ]; then
        print_error "无法读取 API Key"
        return
    fi
    
    print_info "获取容器列表..."
    
    response=$(curl -s -H "X-API-Hash: $API_KEY" http://localhost:8443/api/system/containers)
    
    if echo "$response" | grep -q "code"; then
        echo ""
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        print_error "API 请求失败"
    fi
}

# 测试 API
test_api() {
    API_KEY=$(cat "$CONFIG_DIR/.api_key" 2>/dev/null)
    
    if [ -z "$API_KEY" ]; then
        print_error "无法读取 API Key"
        return
    fi
    
    print_info "测试 API 连接..."
    
    response=$(curl -s -w "\n%{http_code}" -H "X-API-Hash: $API_KEY" http://localhost:8443/api/system/stats)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ]; then
        print_success "API 测试通过"
        echo ""
        echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    else
        print_error "API 测试失败 (HTTP $http_code)"
    fi
}

# 更新服务
update_service() {
    print_warning "此功能将从 GitHub 下载最新版本"
    read -p "确定要继续吗? (y/n): " choice
    
    if [ "$choice" = "y" ]; then
        print_info "停止服务..."
        stop_service
        
        print_info "备份当前版本..."
        cp /usr/local/bin/openlxd /usr/local/bin/openlxd.backup
        
        print_info "下载最新版本..."
        wget -q --show-progress https://github.com/yourusername/openlxd-backend/releases/latest/download/openlxd-linux-amd64 -O /usr/local/bin/openlxd
        
        if [ $? -eq 0 ]; then
            chmod +x /usr/local/bin/openlxd
            print_success "更新完成"
            start_service
        else
            print_error "更新失败，恢复备份..."
            mv /usr/local/bin/openlxd.backup /usr/local/bin/openlxd
            start_service
        fi
    fi
}

# 卸载服务
uninstall_service() {
    print_error "警告: 此操作将完全卸载 OpenLXD"
    read -p "确定要继续吗? (yes/no): " choice
    
    if [ "$choice" = "yes" ]; then
        print_info "停止服务..."
        systemctl stop $SERVICE_NAME
        systemctl disable $SERVICE_NAME
        
        print_info "删除文件..."
        rm -f /usr/local/bin/openlxd
        rm -f /etc/systemd/system/openlxd.service
        systemctl daemon-reload
        
        read -p "是否删除配置和数据? (y/n): " del_data
        if [ "$del_data" = "y" ]; then
            rm -rf /etc/openlxd
            rm -rf /opt/openlxd
        fi
        
        print_success "卸载完成"
    else
        print_info "取消卸载"
    fi
}

# 主循环
main() {
    # 检查 root 权限
    if [[ $EUID -ne 0 ]]; then
        print_error "此脚本需要 root 权限"
        exit 1
    fi
    
    while true; do
        show_menu
        read -p "请选择操作 [0-19]: " choice
        
        case $choice in
            1) start_service ;;
            2) stop_service ;;
            3) restart_service ;;
            4) show_status ;;
            5) show_logs_live ;;
            6) show_logs_recent ;;
            7) show_config ;;
            8) edit_config ;;
            9) show_api_key ;;
            10) regenerate_api_key ;;
            11) backup_database ;;
            12) restore_database ;;
            13) clean_logs ;;
            14) show_system_info ;;
            15) show_performance ;;
            16) show_containers ;;
            17) test_api ;;
            18) update_service ;;
            19) uninstall_service ;;
            0) print_info "退出"; exit 0 ;;
            *) print_error "无效的选项" ;;
        esac
        
        echo ""
        read -p "按 Enter 继续..."
    done
}

# 执行主函数
main "$@"
