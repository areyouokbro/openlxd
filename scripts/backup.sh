#!/bin/bash
#================================================================
# OpenLXD 自动备份脚本
# 版本: 2.0
# 描述: 自动备份数据库和配置文件
#================================================================

set -e

# 配置
BACKUP_DIR="/opt/openlxd/backups"
DB_PATH="/opt/openlxd/lxdapi.db"
CONFIG_DIR="/etc/openlxd"
RETENTION_DAYS=30  # 保留天数

# 颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_info() { echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1"; }
print_success() { echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] [SUCCESS]${NC} $1"; }
print_error() { echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR]${NC} $1"; }

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 生成备份文件名
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_${TIMESTAMP}.tar.gz"

print_info "开始备份..."

# 创建临时目录
TEMP_DIR=$(mktemp -d)
mkdir -p "$TEMP_DIR/database"
mkdir -p "$TEMP_DIR/config"

# 备份数据库
if [ -f "$DB_PATH" ]; then
    print_info "备份数据库..."
    cp "$DB_PATH" "$TEMP_DIR/database/"
    print_success "数据库备份完成"
else
    print_error "数据库文件不存在"
fi

# 备份配置文件
if [ -d "$CONFIG_DIR" ]; then
    print_info "备份配置文件..."
    cp -r "$CONFIG_DIR"/* "$TEMP_DIR/config/"
    print_success "配置文件备份完成"
fi

# 创建备份信息文件
cat > "$TEMP_DIR/backup_info.txt" <<EOF
备份时间: $(date '+%Y-%m-%d %H:%M:%S')
主机名: $(hostname)
操作系统: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)
OpenLXD 版本: 2.0
EOF

# 压缩备份
print_info "压缩备份文件..."
tar -czf "$BACKUP_FILE" -C "$TEMP_DIR" .

# 清理临时目录
rm -rf "$TEMP_DIR"

# 检查备份文件
if [ -f "$BACKUP_FILE" ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    print_success "备份完成: $BACKUP_FILE ($BACKUP_SIZE)"
else
    print_error "备份失败"
    exit 1
fi

# 清理旧备份
print_info "清理 $RETENTION_DAYS 天前的旧备份..."
find "$BACKUP_DIR" -name "backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete
print_success "旧备份清理完成"

# 显示备份列表
print_info "当前备份列表:"
ls -lh "$BACKUP_DIR"/backup_*.tar.gz 2>/dev/null | tail -n 5

print_success "备份任务完成"
