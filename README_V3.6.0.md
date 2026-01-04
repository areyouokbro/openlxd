# OpenLXD v3.6.0 Final

## 🎉 100% 兼容 lxdapi WHMCS 插件

OpenLXD 是一个功能完整、生产就绪的 LXD 容器管理系统，现已**完全兼容** lxdapi WHMCS 插件！

## ✨ 主要特性

### 1. lxdapi 完全兼容
- ✅ 支持 `X-API-Hash` 认证头
- ✅ lxdapi 响应格式 `{code, msg, data}`
- ✅ 11 个完全兼容的 API 端点
- ✅ 所有 19 个创建容器参数支持
- ✅ **无需修改 WHMCS 配置，开箱即用！**

### 2. 多租户管理
- 用户注册/登录系统
- JWT Token 认证
- API Key 管理
- 用户角色管理（admin/user）
- 容器所有权隔离

### 3. 镜像模板市场
- 22 个预定义镜像
- 从 linuxcontainers.org 导入
- 异步镜像导入
- 完整的镜像管理

### 4. 容器管理
- 创建、启动、停止、重启、删除
- 暂停/恢复容器 ⭐
- 重装容器 ⭐
- 修改密码 ⭐
- 流量重置 ⭐
- 资源配额管理

### 5. 网络管理
- IP 地址池管理
- 端口映射
- 反向代理
- 流量监控

### 6. 监控和日志
- 系统资源监控
- 容器性能监控
- 网络流量统计
- 操作日志记录

### 7. Web 管理界面
- 现代化的 Web UI
- 容器管理
- 用户管理
- 镜像市场
- 监控仪表板

## 🚀 快速开始

### 1. 下载

```bash
wget https://github.com/areyouokbro/openlxd/releases/download/v3.6.0-final/openlxd
chmod +x openlxd
```

### 2. 运行

```bash
./openlxd
```

### 3. 访问

打开浏览器访问：`http://localhost:8443`

## 📋 lxdapi 兼容 API

| 端点 | 方法 | 功能 |
|------|------|------|
| `/api/system/containers` | POST | 创建容器 |
| `/api/system/containers/{name}/start` | POST | 启动容器 |
| `/api/system/containers/{name}/stop` | POST | 停止容器 |
| `/api/system/containers/{name}/restart` | POST | 重启容器 |
| `/api/system/containers/{name}` | DELETE | 删除容器 |
| `/api/system/containers/{name}` | GET | 获取容器信息 |
| `/api/system/containers/{name}/suspend` | POST | 暂停容器 ⭐ |
| `/api/system/containers/{name}/unsuspend` | POST | 恢复容器 ⭐ |
| `/api/system/containers/{name}/reinstall` | POST | 重装容器 ⭐ |
| `/api/system/containers/{name}/password` | POST | 修改密码 ⭐ |
| `/api/system/containers/{name}/traffic/reset` | POST | 重置流量 ⭐ |

⭐ = v3.6.0 新增功能

## 🔧 WHMCS 集成

### 1. 安装 lxdapi WHMCS 插件

```bash
cp -r lxdapiserver /path/to/whmcs/modules/servers/
```

### 2. 配置 WHMCS 产品

- **服务器类型：** lxdapiserver
- **主机名：** OpenLXD 服务器地址
- **端口：** 8443
- **API Hash：** 用户的 API Key

### 3. 测试

创建订单，WHMCS 会自动调用 OpenLXD API 创建容器！

## 📚 文档

- [兼容性总结](LXDAPI_COMPATIBILITY_SUMMARY.md)
- [测试文档](LXDAPI_COMPATIBILITY_TEST.md)
- [集成指南](LXDAPI_ROUTES_GUIDE.md)
- [完整交付文档](FINAL_DELIVERY_V3.6.0.md)

## 📊 项目统计

- **代码量：** ~14,500 行
- **API 端点：** 60+
- **数据库表：** 9 个
- **Web 页面：** 13 个
- **文档数量：** 10+

## 🎯 兼容性

| 功能 | lxdapi | OpenLXD v3.6.0 |
|------|--------|----------------|
| API 端点路径 | ✅ | ✅ |
| X-API-Hash 认证 | ✅ | ✅ |
| 响应格式 | ✅ | ✅ |
| 创建容器 | ✅ | ✅ |
| 启动/停止/重启 | ✅ | ✅ |
| 删除容器 | ✅ | ✅ |
| 获取容器信息 | ✅ | ✅ |
| 暂停/恢复容器 | ✅ | ✅ |
| 重装容器 | ✅ | ✅ |
| 修改密码 | ✅ | ✅ |
| 流量重置 | ✅ | ✅ |
| **兼容性** | **100%** | **✅ 100%** |

## 🏆 优势

### 相比 lxdapi 的优势

1. **更强大的功能**
   - 完整的 Web 管理界面
   - 多租户管理
   - 镜像模板市场
   - 网络配置管理
   - 监控和日志系统

2. **更好的性能**
   - Go 语言编写，性能更高
   - 原生 LXD API 调用
   - 更低的资源占用

3. **更完善的文档**
   - 详细的 API 文档
   - 完整的测试文档
   - 集成指南

4. **更活跃的维护**
   - 持续更新
   - 快速响应问题
   - 社区支持

## 🔥 使用示例

### 创建容器

```bash
curl -X POST http://localhost:8443/api/system/containers \
  -H "X-API-Hash: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-container",
    "image": "ubuntu:22.04",
    "cpu": 2,
    "memory": 2048,
    "disk": 20480
  }'
```

### 启动容器

```bash
curl -X POST http://localhost:8443/api/system/containers/test-container/start \
  -H "X-API-Hash: your_api_key"
```

### 获取容器信息

```bash
curl -X GET http://localhost:8443/api/system/containers/test-container \
  -H "X-API-Hash: your_api_key"
```

## 📞 支持

- **GitHub:** https://github.com/areyouokbro/openlxd
- **Issues:** https://github.com/areyouokbro/openlxd/issues
- **Releases:** https://github.com/areyouokbro/openlxd/releases

## 📄 许可证

MIT License

---

**OpenLXD v3.6.0 Final** - 生产就绪的容器管理系统

**100% 兼容 lxdapi WHMCS 插件，开箱即用！**
