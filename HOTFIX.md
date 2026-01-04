# 🔧 Hotfix - 配置文件路径和 glibc 兼容性问题修复

## 修复历史

### 2026-01-04 更新 2：配置文件路径问题

**问题描述**：
```
2026/01/04 01:24:36 配置文件加载失败:open configs/config.yaml: no such file or directory
```

**原因**：程序硬编码使用相对路径 `configs/config.yaml`，但生产环境配置文件在 `/etc/openlxd/config.yaml`

**解决方案**：修改配置加载逻辑，按优先级尝试多个路径：
1. `/etc/openlxd/config.yaml` (生产环境，优先)
2. `configs/config.yaml` (开发环境)
3. `./config.yaml` (当前目录)
4. `/opt/openlxd/config.yaml` (备用路径)

---

### 2026-01-04 更新 1：glibc 兼容性问题

**问题描述**：
```
/usr/local/bin/openlxd: /lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32' not found
/usr/local/bin/openlxd: /lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.34' not found
```

**原因**：原始二进制文件是在 Ubuntu 22.04 (glibc 2.35) 上编译的动态链接版本，依赖较新的 glibc 版本。而 Debian 11 使用 glibc 2.31，导致版本不兼容。

**解决方案**：重新编译为**完全静态链接**的二进制文件，不依赖任何系统动态库。

---

## 当前版本信息

- **版本**: v2.0.0 (最新修复)
- **更新时间**: 2026-01-04
- **文件名**: `openlxd-linux-amd64`
- **文件大小**: 15.43 MB (16,180,896 bytes)
- **编译方式**: 静态链接 + 多路径配置加载
- **下载地址**: https://github.com/areyouokbro/openlxd/releases/latest/download/openlxd-linux-amd64

## 兼容性

新版本支持：

- ✅ Debian 9+ (Stretch 及更新版本)
- ✅ Ubuntu 18.04+ (Bionic 及更新版本)
- ✅ CentOS 7+ / RHEL 7+
- ✅ Rocky Linux 8+
- ✅ Alpine Linux (musl libc)
- ✅ 任何 Linux 内核 3.2.0+ 的 x86_64 系统

## 如何更新

### 方法 1：手动替换二进制文件（推荐，最快）

```bash
# 停止服务
sudo systemctl stop openlxd

# 备份旧版本（可选）
sudo cp /usr/local/bin/openlxd /usr/local/bin/openlxd.backup

# 下载最新版本
wget https://github.com/areyouokbro/openlxd/releases/latest/download/openlxd-linux-amd64

# 替换二进制文件
sudo mv openlxd-linux-amd64 /usr/local/bin/openlxd
sudo chmod +x /usr/local/bin/openlxd

# 启动服务
sudo systemctl start openlxd

# 检查状态
sudo systemctl status openlxd

# 查看日志确认配置文件加载成功
sudo journalctl -u openlxd -n 20
```

### 方法 2：重新运行安装脚本

```bash
# 停止服务
sudo systemctl stop openlxd

# 下载并运行安装脚本
wget https://raw.githubusercontent.com/areyouokbro/openlxd/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh

# 选择选项 1 或 3 从 GitHub 下载
```

## 验证修复

### 1. 检查二进制文件

```bash
# 检查是否为静态链接
ldd /usr/local/bin/openlxd
# 应该显示: not a dynamic executable

# 检查文件类型
file /usr/local/bin/openlxd
# 应该显示: statically linked
```

### 2. 检查服务状态

```bash
# 检查服务是否运行
sudo systemctl status openlxd
# 应该显示: active (running)

# 查看启动日志
sudo journalctl -u openlxd -n 20
# 应该看到: "成功加载配置文件: /etc/openlxd/config.yaml"
```

### 3. 测试 API

```bash
# 获取 API Key
API_KEY=$(sudo cat /etc/openlxd/.api_key)

# 测试 API
curl -H "X-API-Hash: $API_KEY" http://localhost:8443/api/system/stats

# 应该返回 JSON 格式的系统状态
```

## 技术细节

### 静态链接编译

```bash
CGO_ENABLED=1 go build \
  -ldflags='-linkmode external -extldflags "-static"' \
  -tags sqlite_omit_load_extension \
  -o bin/openlxd-linux-amd64 \
  cmd/main.go
```

### 配置文件加载逻辑

程序按以下优先级查找配置文件：

1. `/etc/openlxd/config.yaml` - 生产环境标准路径（推荐）
2. `configs/config.yaml` - 开发环境相对路径
3. `./config.yaml` - 当前目录
4. `/opt/openlxd/config.yaml` - 备用安装路径

如果所有路径都找不到配置文件，程序会输出详细的错误信息。

## 常见问题

### Q1: 更新后服务无法启动

**A**: 检查配置文件是否存在：
```bash
ls -l /etc/openlxd/config.yaml
```

如果不存在，重新运行安装脚本或手动创建配置文件。

### Q2: 仍然提示 glibc 版本错误

**A**: 确认下载的是最新版本：
```bash
# 检查文件大小（应该是 15.43 MB）
ls -lh /usr/local/bin/openlxd

# 检查是否为静态链接
ldd /usr/local/bin/openlxd
```

如果不是静态链接，重新下载最新版本。

### Q3: API 无法访问

**A**: 检查防火墙和端口：
```bash
# 检查端口是否监听
sudo netstat -tlnp | grep 8443

# 检查防火墙规则
sudo iptables -L -n | grep 8443
```

## Release 信息

- **GitHub Release**: https://github.com/areyouokbro/openlxd/releases/tag/v2.0.0
- **源码仓库**: https://github.com/areyouokbro/openlxd
- **问题反馈**: https://github.com/areyouokbro/openlxd/issues

## 注意事项

1. ✅ **配置保留**：更新二进制文件不会影响现有配置文件和数据库
2. ✅ **API Key 保持**：更新后 API Key 保持不变，无需重新配置财务系统插件
3. ✅ **数据安全**：容器数据和流量统计数据不受影响
4. ⚠️ **备份建议**：更新前建议备份配置文件和数据库

## 问题反馈

如果仍然遇到问题，请在 GitHub Issues 中反馈：
https://github.com/areyouokbro/openlxd/issues

请提供以下信息：
- 操作系统版本：`cat /etc/os-release`
- glibc 版本：`ldd --version`
- 二进制文件信息：`file /usr/local/bin/openlxd`
- 服务状态：`systemctl status openlxd`
- 错误日志：`journalctl -u openlxd -n 50`
