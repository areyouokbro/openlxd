# 财务插件集成指南

## WHMCS 集成

1. **上传插件**: 将 `Fmis/whmcs/lxdapiserver` 目录上传到 WHMCS 的 `/modules/servers/` 目录下。
2. **创建服务器**:
   - 进入 WHMCS 后台 -> Setup -> Products/Services -> Servers。
   - 添加新服务器，类型选择 `Lxdapiserver`。
   - 填写后端 API 的 IP 地址、端口（默认 8443）并勾选 SSL（如果已配置）。
3. **创建产品**:
   - 在产品设置的 `Module Settings` 选项卡中选择 `Lxdapiserver`。
   - 配置 CPU、内存、磁盘等参数。

## 魔方财务 (ZJMF) 集成

1. **上传插件**: 将 `Fmis/zjmf/lxdapiserver` 上传到魔方财务的服务器插件目录。
2. **添加接口**: 在魔方后台添加服务器接口，选择对应的 LXD 插件。
3. **同步商品**: 在商品设置中关联该接口并配置资源规格。

## 常见问题

### 1. 连接测试失败
- 检查后端服务是否正在运行。
- 检查防火墙是否放行了 8443 端口。
- 检查 API 地址是否填写正确（包含 http:// 或 https://）。

### 2. 容器创建成功但无法联网
- 检查宿主机的 NAT 转发是否开启：`sysctl -w net.ipv4.ip_forward=1`。
- 检查 LXD 的网桥配置是否正确。
