# OpenLXD 快速安装指南

## 🚀 一条命令完成安装

适用于**纯净系统**，无需任何前置依赖：

```bash
curl -fsSL https://raw.githubusercontent.com/areyouokbro/openlxd/master/scripts/install.sh | sudo bash
```

或者使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/areyouokbro/openlxd/master/scripts/install.sh | sudo bash
```

## ✅ 支持的系统

- Ubuntu 18.04+
- Debian 9+
- CentOS 7+
- Rocky Linux 8+
- AlmaLinux 8+

## 📋 安装过程

脚本会自动完成以下步骤：

1. ✅ 检测操作系统类型和版本
2. ✅ 安装必要依赖（wget、curl、ca-certificates、file）
3. ✅ 从 GitHub 下载最新版本二进制文件（~16MB）
4. ✅ 创建安装目录和配置目录
5. ✅ 生成安全的 API Key 和配置文件
6. ✅ 配置 systemd 服务
7. ✅ 配置防火墙规则（开放 8443 端口）
8. ✅ 启动服务并验证安装

## ⏱️ 安装时间

- 国内服务器：约 1-2 分钟
- 国外服务器：约 30-60 秒

## 🎉 安装完成后

### 访问 Web 管理界面

```
http://你的服务器IP:8443/admin/login
```

### 默认登录凭据

- **用户名**：`admin`
- **密码**：`admin123`

> ⚠️ **重要**：首次登录后请立即修改默认密码！

### 查看 API Key

```bash
sudo cat /etc/openlxd/config.yaml | grep api_hash
```

### 服务管理命令

```bash
# 查看服务状态
sudo systemctl status openlxd

# 查看实时日志
sudo journalctl -u openlxd -f

# 重启服务
sudo systemctl restart openlxd

# 停止服务
sudo systemctl stop openlxd

# 启动服务
sudo systemctl start openlxd
```

## 🔧 常见问题

### Q1: 没有 curl 和 wget 怎么办？

**Debian/Ubuntu**:
```bash
sudo apt-get update
sudo apt-get install -y curl
```

**CentOS/Rocky**:
```bash
sudo yum install -y curl
```

然后重新运行安装命令。

### Q2: 安装失败怎么办？

查看详细日志：
```bash
sudo journalctl -u openlxd -n 100
```

或联系支持：https://github.com/areyouokbro/openlxd/issues

### Q3: 如何卸载？

```bash
# 停止服务
sudo systemctl stop openlxd
sudo systemctl disable openlxd

# 删除文件
sudo rm -f /usr/local/bin/openlxd
sudo rm -f /etc/systemd/system/openlxd.service
sudo rm -rf /etc/openlxd
sudo rm -rf /opt/openlxd

# 重新加载 systemd
sudo systemctl daemon-reload
```

### Q4: 如何更新到最新版本？

```bash
# 停止服务
sudo systemctl stop openlxd

# 下载最新版本
wget https://github.com/areyouokbro/openlxd/releases/latest/download/openlxd-linux-amd64

# 替换二进制文件
sudo mv openlxd-linux-amd64 /usr/local/bin/openlxd
sudo chmod +x /usr/local/bin/openlxd

# 启动服务
sudo systemctl start openlxd
```

### Q5: 无法访问 Web 界面？

检查防火墙：
```bash
# 检查端口是否监听
sudo netstat -tlnp | grep 8443

# 手动开放端口（UFW）
sudo ufw allow 8443/tcp

# 手动开放端口（firewalld）
sudo firewall-cmd --permanent --add-port=8443/tcp
sudo firewall-cmd --reload
```

## 📚 更多文档

- [完整安装指南](INSTALL.md)
- [Web 管理界面文档](docs/web_admin.md)
- [API 文档](docs/api_reference.md)
- [插件集成](docs/plugin_integration.md)

## 💬 获取帮助

- GitHub Issues: https://github.com/areyouokbro/openlxd/issues
- 文档: https://github.com/areyouokbro/openlxd

## 🎯 下一步

1. 登录 Web 管理界面
2. 修改默认管理员密码
3. 配置 LXD 环境（如果还没有）
4. 创建第一个容器
5. 集成财务系统（可选）

祝您使用愉快！🎉
