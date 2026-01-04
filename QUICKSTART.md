# OpenLXD 快速开始指南

## 🚀 30秒快速部署

### 方式 1：直接运行（最简单）

```bash
# 1. 下载
wget https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd
chmod +x openlxd

# 2. 运行（自动创建配置和数据库）
./openlxd
```

**就这么简单！** OpenLXD 会自动：
- ✅ 创建配置文件 `config.yaml`
- ✅ 创建数据库 `openlxd.db`
- ✅ 初始化所有数据表
- ✅ 启动 Web 服务（端口 8443）

### 方式 2：使用安装脚本（生产环境）

```bash
# 1. 下载安装脚本
wget https://raw.githubusercontent.com/areyouokbro/openlxd/master/install.sh

# 2. 运行安装（自动安装 LXD + OpenLXD + 系统服务）
sudo bash install.sh

# 3. 启动服务
sudo systemctl start openlxd

# 4. 查看状态
sudo systemctl status openlxd
```

---

## 📋 前置要求

### 必需
- **LXD** - 容器运行环境
  ```bash
  # Ubuntu/Debian
  sudo snap install lxd
  sudo lxd init --auto
  ```

### 可选
- 无其他依赖！OpenLXD 是单个二进制文件

---

## 🌐 访问 Web 界面

### 1. 打开浏览器

```
http://your-server-ip:8443
```

### 2. 创建用户账户

OpenLXD 启动后，需要创建用户账户：

```bash
curl -X POST http://localhost:8443/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "your-password",
    "role": "admin"
  }'
```

### 3. 登录获取 Token

```bash
curl -X POST http://localhost:8443/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-password"
  }'
```

### 4. 获取 API Key

```bash
curl -X GET http://localhost:8443/api/v1/users/profile \
  -H "Authorization: Bearer <your_jwt_token>"
```

---

## 🔧 配置 WHMCS

OpenLXD 100% 兼容 lxdapi WHMCS 插件！

### 1. 安装 lxdapi WHMCS 模块

```bash
cp -r lxdapiserver /path/to/whmcs/modules/servers/
```

### 2. 配置 WHMCS 产品

在 WHMCS 管理后台：

1. **产品/服务** → **创建新产品**
2. **模块设置：**
   - 模块：lxdapiserver
   - 服务器：选择或创建新服务器
3. **服务器配置：**
   - 主机名：OpenLXD 服务器 IP
   - 端口：8443
   - API Hash：用户的 API Key（从上面获取）

### 3. 测试

创建订单，WHMCS 会自动调用 OpenLXD API 创建容器！

---

## 📊 验证安装

### 1. 检查服务状态

```bash
# 直接运行模式
ps aux | grep openlxd

# 系统服务模式
sudo systemctl status openlxd
```

### 2. 测试 API

```bash
# 创建容器
curl -X POST http://localhost:8443/api/system/containers \
  -H "X-API-Hash: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-container",
    "image": "ubuntu:22.04",
    "cpu": 1,
    "memory": 512,
    "disk": 10240
  }'

# 启动容器
curl -X POST http://localhost:8443/api/system/containers/test-container/start \
  -H "X-API-Hash: your_api_key"

# 获取容器信息
curl -X GET http://localhost:8443/api/system/containers/test-container \
  -H "X-API-Hash: your_api_key"
```

### 3. 查看日志

```bash
# 直接运行模式
tail -f openlxd.log

# 系统服务模式
sudo journalctl -u openlxd -f
```

---

## 🎯 常见问题

### Q: 如何修改端口？

编辑 `config.yaml`：

```yaml
server:
  port: 8080  # 修改为你想要的端口
```

然后重启服务。

### Q: 如何启用 HTTPS？

编辑 `config.yaml`：

```yaml
server:
  https: true
  domain: "your-domain.com"
  auto_tls: true
```

### Q: 数据库文件在哪里？

默认位置：
- 直接运行：`./openlxd.db`
- 系统服务：`/var/lib/openlxd/openlxd.db`

### Q: 如何备份数据？

```bash
# 备份数据库
cp openlxd.db openlxd.db.backup

# 备份配置
cp config.yaml config.yaml.backup
```

### Q: 如何升级？

```bash
# 1. 停止服务
sudo systemctl stop openlxd

# 2. 备份
cp /opt/openlxd/openlxd /opt/openlxd/openlxd.backup
cp /var/lib/openlxd/openlxd.db /var/lib/openlxd/openlxd.db.backup

# 3. 下载新版本
wget -O /opt/openlxd/openlxd https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd
chmod +x /opt/openlxd/openlxd

# 4. 启动服务
sudo systemctl start openlxd
```

### Q: 如何查看所有用户？

```bash
# 使用管理员账户登录后
curl -X GET http://localhost:8443/api/v1/users/list \
  -H "Authorization: Bearer <admin_jwt_token>"
```

### Q: 如何重置管理员密码？

直接编辑数据库或重新创建管理员账户。

---

## 📚 下一步

### 学习更多

- [完整文档](README_V3.6.0.md)
- [API 文档](LXDAPI_COMPATIBILITY_SUMMARY.md)
- [测试指南](LXDAPI_COMPATIBILITY_TEST.md)
- [最终检查报告](FINAL_CHECK_REPORT.md)

### 配置功能

1. **创建用户** - 多租户管理
2. **导入镜像** - 从 linuxcontainers.org
3. **配置网络** - IP 池、端口映射
4. **设置配额** - 资源限制
5. **配置 WHMCS** - 财务系统对接

### 获取帮助

- **GitHub Issues:** https://github.com/areyouokbro/openlxd/issues
- **文档:** https://github.com/areyouokbro/openlxd

---

## 🎉 完成！

现在你已经成功部署了 OpenLXD！

- ✅ 容器管理系统
- ✅ Web 管理界面
- ✅ WHMCS 兼容 API
- ✅ 多租户支持
- ✅ 镜像市场

**开始创建你的第一个容器吧！** 🚀

---

## 🔥 快速命令参考

```bash
# 下载并运行
wget https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd && chmod +x openlxd && ./openlxd

# 创建用户
curl -X POST http://localhost:8443/api/v1/users/register -H "Content-Type: application/json" -d '{"username":"admin","email":"admin@example.com","password":"admin123","role":"admin"}'

# 登录
curl -X POST http://localhost:8443/api/v1/users/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'

# 创建容器
curl -X POST http://localhost:8443/api/system/containers -H "X-API-Hash: YOUR_API_KEY" -H "Content-Type: application/json" -d '{"name":"test","image":"ubuntu:22.04","cpu":1,"memory":512,"disk":10240}'

# 启动容器
curl -X POST http://localhost:8443/api/system/containers/test/start -H "X-API-Hash: YOUR_API_KEY"
```

---

**OpenLXD v3.6.0 Final** - 真正的一键部署！
