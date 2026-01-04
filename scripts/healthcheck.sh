#!/bin/bash
#================================================================
# OpenLXD 健康检查脚本
# 版本: 2.0
# 描述: 检查服务健康状态，可用于监控告警
#================================================================

# 配置
API_KEY_FILE="/etc/openlxd/.api_key"
API_URL="http://localhost:8443/api/system/stats"
SERVICE_NAME="openlxd"

# 退出码
EXIT_OK=0
EXIT_WARNING=1
EXIT_CRITICAL=2
EXIT_UNKNOWN=3

# 检查服务状态
check_service() {
    if systemctl is-active --quiet $SERVICE_NAME; then
        return 0
    else
        echo "CRITICAL: OpenLXD 服务未运行"
        return $EXIT_CRITICAL
    fi
}

# 检查端口监听
check_port() {
    if netstat -tlnp 2>/dev/null | grep -q ":8443" || ss -tlnp 2>/dev/null | grep -q ":8443"; then
        return 0
    else
        echo "CRITICAL: 端口 8443 未监听"
        return $EXIT_CRITICAL
    fi
}

# 检查 API 响应
check_api() {
    if [ ! -f "$API_KEY_FILE" ]; then
        echo "WARNING: API Key 文件不存在"
        return $EXIT_WARNING
    fi
    
    API_KEY=$(cat "$API_KEY_FILE")
    
    response=$(curl -s -w "\n%{http_code}" -H "X-API-Hash: $API_KEY" --max-time 5 "$API_URL" 2>/dev/null)
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        return 0
    else
        echo "CRITICAL: API 响应异常 (HTTP $http_code)"
        return $EXIT_CRITICAL
    fi
}

# 检查数据库
check_database() {
    DB_PATH="/opt/openlxd/lxdapi.db"
    
    if [ ! -f "$DB_PATH" ]; then
        echo "CRITICAL: 数据库文件不存在"
        return $EXIT_CRITICAL
    fi
    
    # 检查数据库是否可读
    if sqlite3 "$DB_PATH" "SELECT 1;" > /dev/null 2>&1; then
        return 0
    else
        echo "CRITICAL: 数据库损坏或无法访问"
        return $EXIT_CRITICAL
    fi
}

# 检查磁盘空间
check_disk() {
    usage=$(df / | awk 'NR==2 {print $5}' | tr -d '%')
    
    if [ "$usage" -gt 90 ]; then
        echo "CRITICAL: 磁盘使用率 ${usage}%"
        return $EXIT_CRITICAL
    elif [ "$usage" -gt 80 ]; then
        echo "WARNING: 磁盘使用率 ${usage}%"
        return $EXIT_WARNING
    else
        return 0
    fi
}

# 检查内存
check_memory() {
    mem_usage=$(free | awk '/Mem:/ {printf "%.0f", $3/$2 * 100}')
    
    if [ "$mem_usage" -gt 90 ]; then
        echo "WARNING: 内存使用率 ${mem_usage}%"
        return $EXIT_WARNING
    else
        return 0
    fi
}

# 主函数
main() {
    exit_code=$EXIT_OK
    messages=()
    
    # 执行所有检查
    check_service || { exit_code=$?; messages+=("服务检查失败"); }
    check_port || { exit_code=$?; messages+=("端口检查失败"); }
    check_api || { exit_code=$?; messages+=("API检查失败"); }
    check_database || { exit_code=$?; messages+=("数据库检查失败"); }
    check_disk || { exit_code=$?; messages+=("磁盘检查失败"); }
    check_memory || { exit_code=$?; messages+=("内存检查失败"); }
    
    # 输出结果
    if [ $exit_code -eq $EXIT_OK ]; then
        echo "OK: OpenLXD 运行正常"
    elif [ $exit_code -eq $EXIT_WARNING ]; then
        echo "WARNING: ${messages[*]}"
    else
        echo "CRITICAL: ${messages[*]}"
    fi
    
    exit $exit_code
}

main "$@"
