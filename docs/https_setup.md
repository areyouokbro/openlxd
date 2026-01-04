# OpenLXD HTTPS 配置指南

## 概述

OpenLXD 支持两种 HTTPS 配置方式：
1. **自动 Let's Encrypt 证书**（推荐，免费，自动续期）
2. **手动证书**（适用于已有证书或内网环境）

## 方式一：Let's Encrypt 自动证书（推荐）

### 前提条件

1. **拥有一个域名**（例如：`api.yourdomain.com`）
2. **域名已解析到服务器 IP**
3. **服务器 80 和 443 端口可访问**（Let's Encrypt 需要 80 端口验证）

### 配置步骤

#### 1. 编辑配置文件

```bash
sudo nano /etc/openlxd/config.yaml
```

修改以下配置：

```yaml
server:
  port: 443                      # HTTPS 默认端口
  host: "0.0.0.0"
  https: true                    # 启用 HTTPS
  domain: "api.yourdomain.com"   # 替换为你的域名
  cert_dir: "/etc/openlxd/certs" # 证书存储目录
  auto_tls: true                 # 启用自动证书
```

#### 2. 确保域名已解析

```bash
# 检查域名解析
nslookup api.yourdomain.com

# 或使用 dig
dig api.yourdomain.com +short
```

确保返回的 IP 地址是你的服务器 IP。

#### 3. 开放必要端口

```bash
# UFW 防火墙
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# firewalld
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

#### 4. 重启服务

```bash
sudo systemctl restart openlxd
```

#### 5. 查看日志确认证书申请成功

```bash
sudo journalctl -u openlxd -f
```

你应该看到类似以下日志：

```
HTTP 服务器启动 (ACME 验证): :80
服务器监听 (HTTPS): 0.0.0.0:443
域名: api.yourdomain.com
证书目录: /etc/openlxd/certs
```

#### 6. 访问测试

```bash
# 测试 HTTPS 访问
curl https://api.yourdomain.com/api/system/stats \
  -H "X-API-Hash: your-api-key"
```

Web 管理界面：
```
https://api.yourdomain.com/admin/login
```

### 证书自动续期

Let's Encrypt 证书有效期为 90 天，但 `autocert` 会自动续期，无需手动操作。

证书文件存储在：`/etc/openlxd/certs/`

## 方式二：手动证书

### 适用场景

- 已有 SSL 证书（购买的或企业内部 CA 签发）
- 内网环境（无法使用 Let's Encrypt）
- 需要使用通配符证书

### 配置步骤

#### 1. 准备证书文件

将证书文件放到 `/etc/openlxd/certs/` 目录：

```bash
sudo mkdir -p /etc/openlxd/certs
sudo cp your-cert.pem /etc/openlxd/certs/cert.pem
sudo cp your-key.pem /etc/openlxd/certs/key.pem
sudo chmod 600 /etc/openlxd/certs/*.pem
```

#### 2. 编辑配置文件

```bash
sudo nano /etc/openlxd/config.yaml
```

修改配置：

```yaml
server:
  port: 443
  host: "0.0.0.0"
  https: true                    # 启用 HTTPS
  domain: ""                     # 手动证书不需要域名
  cert_dir: "/etc/openlxd/certs" # 证书目录
  auto_tls: false                # 禁用自动证书
```

#### 3. 重启服务

```bash
sudo systemctl restart openlxd
```

#### 4. 查看日志

```bash
sudo journalctl -u openlxd -n 20
```

应该看到：

```
服务器监听 (HTTPS): 0.0.0.0:443
证书: /etc/openlxd/certs/cert.pem
密钥: /etc/openlxd/certs/key.pem
```

## 生成自签名证书（测试用）

如果只是测试，可以生成自签名证书：

```bash
# 创建证书目录
sudo mkdir -p /etc/openlxd/certs

# 生成自签名证书（有效期 365 天）
sudo openssl req -x509 -newkey rsa:4096 \
  -keyout /etc/openlxd/certs/key.pem \
  -out /etc/openlxd/certs/cert.pem \
  -days 365 -nodes \
  -subj "/CN=localhost"

# 设置权限
sudo chmod 600 /etc/openlxd/certs/*.pem
```

然后按照"方式二"配置即可。

> ⚠️ **注意**：自签名证书会导致浏览器显示安全警告，仅适用于测试环境。

## WHMCS 插件配置

配置 WHMCS 插件时，使用 HTTPS 地址：

```
API URL: https://api.yourdomain.com
API Key: your-api-key-here
```

确保勾选"验证 SSL 证书"选项（如果使用 Let's Encrypt 或正规 CA 证书）。

## 常见问题

### Q1: 证书申请失败

**可能原因**：
- 域名未正确解析到服务器
- 80 端口被占用或防火墙阻止
- 域名配置错误

**解决方法**：

```bash
# 检查域名解析
nslookup your-domain.com

# 检查 80 端口是否被占用
sudo netstat -tlnp | grep :80

# 查看详细日志
sudo journalctl -u openlxd -n 100
```

### Q2: 浏览器显示"不安全"

**Let's Encrypt 证书**：
- 等待几分钟，证书申请需要时间
- 检查域名是否正确解析

**自签名证书**：
- 这是正常现象，自签名证书不被浏览器信任
- 生产环境请使用 Let's Encrypt 或购买证书

### Q3: WHMCS 插件连接失败

**错误信息**：`SSL routines:ssl3_get_record:wrong version number`

**原因**：配置文件中 `https: false`，但 WHMCS 使用 HTTPS 连接

**解决方法**：
```bash
# 编辑配置文件
sudo nano /etc/openlxd/config.yaml

# 修改为
server:
  https: true
  auto_tls: true
  domain: "your-domain.com"

# 重启服务
sudo systemctl restart openlxd
```

### Q4: 证书过期

**Let's Encrypt**：
- 自动续期，无需手动操作
- 如果续期失败，检查 80 端口是否可访问

**手动证书**：
- 需要手动更新证书文件
- 更新后重启服务：`sudo systemctl restart openlxd`

### Q5: 如何从 HTTP 切换到 HTTPS

```bash
# 1. 编辑配置
sudo nano /etc/openlxd/config.yaml

# 2. 修改配置
server:
  https: true
  auto_tls: true
  domain: "your-domain.com"

# 3. 开放端口
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 4. 重启服务
sudo systemctl restart openlxd

# 5. 更新 WHMCS 插件配置
# 将 API URL 从 http:// 改为 https://
```

## 安全建议

1. **使用 Let's Encrypt**：免费、自动续期、受信任
2. **定期检查证书有效期**：`openssl x509 -in /etc/openlxd/certs/cert.pem -noout -dates`
3. **保护私钥文件**：`chmod 600 /etc/openlxd/certs/key.pem`
4. **使用强密码**：修改默认的 admin 密码
5. **限制访问 IP**：使用防火墙限制只有特定 IP 可访问

## 相关文档

- [安装指南](../INSTALL.md)
- [Web 管理界面](./web_admin.md)
- [API 文档](./api_reference.md)
- [插件集成](./plugin_integration.md)

## 问题反馈

如有问题，请在 GitHub 提交 Issue：
https://github.com/areyouokbro/openlxd/issues
