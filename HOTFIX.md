# 🔧 Hotfix - glibc 兼容性问题修复

## 问题描述

在 Debian 11 及其他使用较旧 glibc 版本的系统上，运行 OpenLXD 时出现以下错误：

```
/usr/local/bin/openlxd: /lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32' not found
/usr/local/bin/openlxd: /lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.34' not found
```

## 原因分析

原始二进制文件是在 Ubuntu 22.04 (glibc 2.35) 上编译的动态链接版本，依赖较新的 glibc 版本。而 Debian 11 使用 glibc 2.31，导致版本不兼容。

## 解决方案

已重新编译为**完全静态链接**的二进制文件，不依赖任何系统动态库。

### 编译参数

```bash
CGO_ENABLED=1 go build \
  -ldflags='-linkmode external -extldflags "-static"' \
  -tags sqlite_omit_load_extension \
  -o bin/openlxd-linux-amd64 \
  cmd/main.go
```

### 验证静态链接

```bash
$ file openlxd-linux-amd64
openlxd-linux-amd64: ELF 64-bit LSB executable, x86-64, version 1 (GNU/Linux), 
statically linked, BuildID[sha1]=6d8f2e16bc77fe603f3abbc0cc30074418b36d0f, 
for GNU/Linux 3.2.0, not stripped

$ ldd openlxd-linux-amd64
	not a dynamic executable
```

## 兼容性

新的静态链接版本支持：

- ✅ Debian 9+ (Stretch 及更新版本)
- ✅ Ubuntu 18.04+ (Bionic 及更新版本)
- ✅ CentOS 7+ / RHEL 7+
- ✅ Rocky Linux 8+
- ✅ Alpine Linux (musl libc)
- ✅ 任何 Linux 内核 3.2.0+ 的 x86_64 系统

## 如何更新

### 方法 1：重新运行安装脚本（推荐）

```bash
# 停止旧服务
sudo systemctl stop openlxd

# 重新下载并安装
wget https://raw.githubusercontent.com/areyouokbro/openlxd/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh

# 选择选项 1 或 3 从 GitHub 下载
```

### 方法 2：手动替换二进制文件

```bash
# 停止服务
sudo systemctl stop openlxd

# 下载新版本
wget https://github.com/areyouokbro/openlxd/releases/latest/download/openlxd-linux-amd64

# 替换二进制文件
sudo mv openlxd-linux-amd64 /usr/local/bin/openlxd
sudo chmod +x /usr/local/bin/openlxd

# 启动服务
sudo systemctl start openlxd

# 检查状态
sudo systemctl status openlxd
```

## 验证修复

```bash
# 检查二进制文件类型
file /usr/local/bin/openlxd

# 应该显示: statically linked

# 检查服务状态
sudo systemctl status openlxd

# 应该显示: active (running)

# 检查日志
sudo journalctl -u openlxd -n 20
```

## Release 信息

- **版本**: v2.0.0
- **更新时间**: 2026-01-04
- **文件名**: `openlxd-linux-amd64`
- **文件大小**: 15.43 MB (16,180,896 bytes)
- **下载地址**: https://github.com/areyouokbro/openlxd/releases/latest/download/openlxd-linux-amd64

## 注意事项

1. **旧版本文件**: Release 中仍保留 `openlxd-go1.18` (动态链接版本) 供参考，但不推荐使用
2. **配置保留**: 更新二进制文件不会影响现有配置文件和数据库
3. **API Key**: 更新后 API Key 保持不变，无需重新配置财务系统插件

## 问题反馈

如果仍然遇到问题，请在 GitHub Issues 中反馈：
https://github.com/areyouokbro/openlxd/issues

请提供以下信息：
- 操作系统版本 (`cat /etc/os-release`)
- glibc 版本 (`ldd --version`)
- 错误日志 (`journalctl -u openlxd -n 50`)
