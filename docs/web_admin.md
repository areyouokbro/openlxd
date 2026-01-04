# OpenLXD Web 管理界面使用指南

## 概述

OpenLXD 提供了一个简洁易用的 Web 管理界面，可以通过浏览器直接管理 LXD 容器，无需使用命令行或 API。

## 功能特性

- ✅ 用户登录认证
- ✅ 容器列表查看
- ✅ 实时系统统计（容器数量、运行状态、系统负载）
- ✅ 流量统计显示
- ✅ 自动数据刷新（每30秒）
- ✅ 响应式设计，支持移动端访问

## 部署方法

### 方法一：完整安装（推荐）

使用一键安装脚本会自动部署 Web 界面所需的所有文件：

```bash
# 下载并运行安装脚本
wget https://raw.githubusercontent.com/areyouokbro/openlxd/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
```

安装脚本会自动：
1. 下载二进制文件到 `/usr/local/bin/openlxd`
2. 克隆完整项目到 `/opt/openlxd`
3. 配置 systemd 服务，工作目录设置为 `/opt/openlxd`

### 方法二：手动部署

如果你已经安装了 OpenLXD 但没有 Web 界面文件，按以下步骤操作：

#### 1. 克隆项目源码

```bash
# 克隆到 /opt/openlxd
cd /opt
sudo git clone https://github.com/areyouokbro/openlxd.git
```

#### 2. 修改 systemd 服务配置

编辑服务文件：

```bash
sudo nano /etc/systemd/system/openlxd.service
```

确保 `WorkingDirectory` 设置为项目目录：

```ini
[Unit]
Description=OpenLXD Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/openlxd
ExecStart=/usr/local/bin/openlxd
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

#### 3. 重新加载并启动服务

```bash
sudo systemctl daemon-reload
sudo systemctl restart openlxd
sudo systemctl status openlxd
```

### 方法三：从源码编译部署

```bash
# 克隆项目
git clone https://github.com/areyouokbro/openlxd.git
cd openlxd

# 编译
CGO_ENABLED=1 go build -ldflags='-linkmode external -extldflags "-static"' \
  -tags sqlite_omit_load_extension \
  -o bin/openlxd-linux-amd64 \
  cmd/main.go

# 安装
sudo cp bin/openlxd-linux-amd64 /usr/local/bin/openlxd
sudo chmod +x /usr/local/bin/openlxd

# 配置 systemd 服务（WorkingDirectory 指向项目目录）
sudo systemctl restart openlxd
```

## 访问 Web 管理界面

### 1. 获取服务器 IP 地址

```bash
# 查看服务器 IP
hostname -I | awk '{print $1}'
```

### 2. 打开浏览器访问

在浏览器中输入以下地址：

```
http://你的服务器IP:8443/admin/login
```

例如：
```
http://156.246.90.151:8443/admin/login
```

### 3. 登录

**默认登录凭据：**
- 用户名：`admin`
- 密码：`admin123`

> ⚠️ **安全提示**：首次登录后请立即修改默认密码！

## 修改管理员密码

编辑配置文件：

```bash
sudo nano /etc/openlxd/config.yaml
```

修改以下部分：

```yaml
security:
  api_hash: "你的API密钥"
  admin_user: "admin"           # 可以修改用户名
  admin_pass: "你的新密码"       # 修改密码
  session_secret: "会话密钥"
```

重启服务使配置生效：

```bash
sudo systemctl restart openlxd
```

## 管理界面功能说明

### 登录页面 (`/admin/login`)

- 输入用户名和密码
- 登录成功后自动跳转到管理后台
- 登录凭据保存在浏览器本地存储中

### 管理后台 (`/admin/dashboard`)

#### 系统统计卡片

显示以下实时统计信息：
- **总容器数**：系统中所有容器的数量
- **运行中**：当前正在运行的容器数量
- **已停止**：已停止的容器数量
- **系统负载**：服务器当前负载

#### 容器列表

显示所有容器的详细信息：
- **主机名**：容器的主机名
- **IP 地址**：容器的 IPv4 地址
- **镜像**：容器使用的系统镜像
- **状态**：运行中（绿色）或已停止（红色）
- **流量统计**：上传流量 / 下载流量

#### 功能按钮

- **刷新按钮**：手动刷新容器列表和统计数据
- **退出登录**：退出当前登录，返回登录页面

## 常见问题

### Q1: 访问 Web 界面显示 404 错误

**原因**：Web 界面文件未正确部署。

**解决方法**：

```bash
# 检查项目目录是否存在
ls -la /opt/openlxd/web/templates/

# 如果不存在，克隆项目
cd /opt
sudo git clone https://github.com/areyouokbro/openlxd.git

# 检查 systemd 服务的工作目录
sudo systemctl cat openlxd.service | grep WorkingDirectory

# 如果工作目录不是 /opt/openlxd，修改服务配置
sudo nano /etc/systemd/system/openlxd.service
# 添加或修改: WorkingDirectory=/opt/openlxd

# 重新加载并重启服务
sudo systemctl daemon-reload
sudo systemctl restart openlxd
```

### Q2: 登录后显示"加载失败，请检查 API Key 是否正确"

**原因**：API Key 配置不正确或服务未正常运行。

**解决方法**：

```bash
# 检查服务状态
sudo systemctl status openlxd

# 查看日志
sudo journalctl -u openlxd -n 50

# 检查 API Key
sudo cat /etc/openlxd/.api_key

# 确认配置文件中的 API Key 一致
sudo cat /etc/openlxd/config.yaml | grep api_hash
```

### Q3: 容器列表为空但实际有容器

**原因**：LXD 未正确初始化或权限问题。

**解决方法**：

```bash
# 检查 LXD 是否运行
sudo lxc list

# 检查 LXD socket 权限
sudo ls -la /var/snap/lxd/common/lxd/unix.socket

# 确认配置文件中的 socket 路径正确
sudo cat /etc/openlxd/config.yaml | grep socket

# 重启服务
sudo systemctl restart openlxd
```

### Q4: 无法访问 8443 端口

**原因**：防火墙阻止了 8443 端口。

**解决方法**：

```bash
# UFW 防火墙
sudo ufw allow 8443/tcp
sudo ufw reload

# iptables 防火墙
sudo iptables -A INPUT -p tcp --dport 8443 -j ACCEPT
sudo iptables-save | sudo tee /etc/iptables/rules.v4

# firewalld 防火墙
sudo firewall-cmd --permanent --add-port=8443/tcp
sudo firewall-cmd --reload

# 检查端口是否开放
sudo netstat -tlnp | grep 8443
```

### Q5: 修改密码后无法登录

**原因**：配置文件格式错误或服务未重启。

**解决方法**：

```bash
# 检查配置文件语法
sudo cat /etc/openlxd/config.yaml

# 确保 YAML 格式正确（注意缩进）
# 重启服务
sudo systemctl restart openlxd

# 查看日志确认配置加载成功
sudo journalctl -u openlxd -n 20
```

## 安全建议

1. **修改默认密码**
   - 首次登录后立即修改 `admin_pass`
   - 使用强密码（至少12位，包含大小写字母、数字、特殊字符）

2. **使用 HTTPS**
   - 在生产环境中，建议使用 Nginx 或 Caddy 反向代理
   - 配置 SSL 证书（Let's Encrypt）

3. **限制访问 IP**
   - 使用防火墙限制只有特定 IP 可以访问 8443 端口
   - 例如：`sudo ufw allow from 你的IP to any port 8443`

4. **定期更新**
   - 定期更新 OpenLXD 到最新版本
   - 关注 GitHub Release 获取安全更新

## Nginx 反向代理配置示例

如果需要使用域名和 HTTPS 访问：

```nginx
server {
    listen 80;
    server_name openlxd.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name openlxd.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/openlxd.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/openlxd.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8443;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 技术细节

### 文件结构

```
/opt/openlxd/
├── web/
│   └── templates/
│       ├── login.html       # 登录页面
│       ├── dashboard.html   # 管理后台
│       └── index.html       # 默认首页
├── configs/
│   └── config.yaml          # 配置文件模板
└── bin/
    └── openlxd-linux-amd64  # 二进制文件
```

### API 端点

Web 界面使用以下 API 端点：

- `POST /admin/api/login` - 登录认证
- `GET /api/system/containers` - 获取容器列表
- `GET /api/system/stats` - 获取系统统计

所有 API 请求需要在 Header 中携带 `X-API-Hash` 认证。

## 相关文档

- [安装指南](../INSTALL.md)
- [API 参考](./api_reference.md)
- [后端部署](./backend_setup.md)
- [插件集成](./plugin_integration.md)

## 问题反馈

如有问题或建议，请在 GitHub 提交 Issue：
https://github.com/areyouokbro/openlxd/issues
