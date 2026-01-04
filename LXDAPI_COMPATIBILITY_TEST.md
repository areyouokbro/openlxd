# lxdapi 兼容性测试文档

## 📋 测试清单

### 前置条件

1. ✅ 编译成功（bin/openlxd-lxdapi）
2. ⏳ 添加 lxdapi 路由到 main.go
3. ⏳ 启动服务
4. ⏳ 创建测试用户并获取 API Key

### API 端点测试

#### 1. 测试认证（X-API-Hash）

```bash
# 使用 X-API-Hash 认证头
curl -X GET http://localhost:8443/api/system/containers/test \
  -H "X-API-Hash: your_api_key"

# 应该返回 lxdapi 格式响应：
# {"code": 200, "msg": "...", "data": {...}}
```

#### 2. 测试创建容器

```bash
curl -X POST http://localhost:8443/api/system/containers \
  -H "X-API-Hash: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "lxd11451123456",
    "image": "ubuntu:22.04",
    "username": "user_123",
    "password": "testpass123",
    "cpu": 2,
    "memory": 2048,
    "disk": 20480,
    "ingress": 100,
    "egress": 100,
    "traffic_limit": 100,
    "cpu_allowance": 100,
    "io_read": 100,
    "io_write": 50,
    "processes_limit": 512,
    "allow_nesting": true,
    "memory_swap": true,
    "privileged": false
  }'

# 预期响应：
# {
#   "code": 200,
#   "msg": "创建容器成功",
#   "data": {
#     "name": "lxd11451123456",
#     "ipv4": "10.x.x.x",
#     "ipv6": "..."
#   }
# }
```

#### 3. 测试启动容器

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/start \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "启动容器成功", "data": null}
```

#### 4. 测试停止容器

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/stop \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "停止容器成功", "data": null}
```

#### 5. 测试重启容器

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/restart \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "重启容器成功", "data": null}
```

#### 6. 测试获取容器信息

```bash
curl -X GET http://localhost:8443/api/system/containers/lxd11451123456 \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {
#   "code": 200,
#   "msg": "获取容器信息成功",
#   "data": {
#     "name": "lxd11451123456",
#     "image": "ubuntu:22.04",
#     "status": "running",
#     "cpu": 2,
#     "memory": 2048,
#     "disk": 20,
#     "ipv4": "10.x.x.x",
#     "ipv6": "...",
#     "created_at": "..."
#   }
# }
```

#### 7. 测试暂停容器

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/suspend \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "暂停容器成功", "data": null}
```

#### 8. 测试恢复容器

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/unsuspend \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "恢复容器成功", "data": null}
```

#### 9. 测试重装容器

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/reinstall \
  -H "X-API-Hash: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "ubuntu:22.04"
  }'

# 预期响应：
# {"code": 200, "msg": "重装容器成功", "data": null}
```

#### 10. 测试修改密码

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/password \
  -H "X-API-Hash: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "password": "newpass123"
  }'

# 预期响应：
# {"code": 200, "msg": "修改密码成功", "data": null}
```

#### 11. 测试流量重置

```bash
curl -X POST http://localhost:8443/api/system/containers/lxd11451123456/traffic/reset \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "流量重置成功", "data": null}
```

#### 12. 测试删除容器

```bash
curl -X DELETE http://localhost:8443/api/system/containers/lxd11451123456 \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 200, "msg": "删除容器成功", "data": null}
```

## 🔍 错误测试

### 1. 测试无效 API Key

```bash
curl -X GET http://localhost:8443/api/system/containers/test \
  -H "X-API-Hash: invalid_key"

# 预期响应：
# {"code": 401, "msg": "Invalid API key", "data": null}
```

### 2. 测试缺少 API Key

```bash
curl -X GET http://localhost:8443/api/system/containers/test

# 预期响应：
# {"code": 401, "msg": "Missing API key", "data": null}
```

### 3. 测试容器不存在

```bash
curl -X GET http://localhost:8443/api/system/containers/nonexistent \
  -H "X-API-Hash: your_api_key"

# 预期响应：
# {"code": 404, "msg": "容器不存在或无权限", "data": null}
```

### 4. 测试无权限访问

```bash
# 使用用户A的API Key访问用户B的容器
curl -X GET http://localhost:8443/api/system/containers/user_b_container \
  -H "X-API-Hash: user_a_api_key"

# 预期响应：
# {"code": 404, "msg": "容器不存在或无权限", "data": null}
```

## 📊 兼容性验证

### lxdapi WHMCS 插件测试

1. **安装 lxdapi WHMCS 模块**
   ```bash
   cp -r lxdapiserver /path/to/whmcs/modules/servers/
   ```

2. **配置 WHMCS 产品**
   - 服务器类型：lxdapiserver
   - 主机名：OpenLXD 服务器地址
   - 端口：8443
   - API Hash：用户的 API Key

3. **测试 WHMCS 功能**
   - ✅ 创建订单
   - ✅ 自动开通容器
   - ✅ 暂停服务
   - ✅ 恢复服务
   - ✅ 删除服务
   - ✅ 重装系统
   - ✅ 修改密码

## 🎯 测试结果

### 预期结果

| 功能 | 状态 | 备注 |
|------|------|------|
| X-API-Hash 认证 | ⏳ 待测试 | 兼容 lxdapi |
| 创建容器 | ⏳ 待测试 | 支持所有参数 |
| 启动容器 | ⏳ 待测试 | |
| 停止容器 | ⏳ 待测试 | |
| 重启容器 | ⏳ 待测试 | |
| 删除容器 | ⏳ 待测试 | |
| 获取容器信息 | ⏳ 待测试 | |
| 暂停容器 | ⏳ 待测试 | 新功能 |
| 恢复容器 | ⏳ 待测试 | 新功能 |
| 重装容器 | ⏳ 待测试 | 新功能 |
| 修改密码 | ⏳ 待测试 | 新功能 |
| 流量重置 | ⏳ 待测试 | 新功能 |
| 响应格式 | ⏳ 待测试 | {code, msg, data} |
| 容器命名 | ⏳ 待测试 | lxd11451{userid}{serviceid} |
| 权限隔离 | ⏳ 待测试 | 用户只能访问自己的容器 |

## 📝 测试脚本

创建自动化测试脚本：

```bash
#!/bin/bash

API_KEY="your_api_key"
BASE_URL="http://localhost:8443"
CONTAINER_NAME="lxd11451test123"

echo "=== lxdapi 兼容性测试 ==="

# 1. 创建容器
echo "1. 创建容器..."
curl -s -X POST $BASE_URL/api/system/containers \
  -H "X-API-Hash: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$CONTAINER_NAME\",
    \"image\": \"ubuntu:22.04\",
    \"cpu\": 1,
    \"memory\": 512,
    \"disk\": 10240
  }" | jq .

# 2. 获取容器信息
echo "2. 获取容器信息..."
curl -s -X GET $BASE_URL/api/system/containers/$CONTAINER_NAME \
  -H "X-API-Hash: $API_KEY" | jq .

# 3. 停止容器
echo "3. 停止容器..."
curl -s -X POST $BASE_URL/api/system/containers/$CONTAINER_NAME/stop \
  -H "X-API-Hash: $API_KEY" | jq .

# 4. 启动容器
echo "4. 启动容器..."
curl -s -X POST $BASE_URL/api/system/containers/$CONTAINER_NAME/start \
  -H "X-API-Hash: $API_KEY" | jq .

# 5. 暂停容器
echo "5. 暂停容器..."
curl -s -X POST $BASE_URL/api/system/containers/$CONTAINER_NAME/suspend \
  -H "X-API-Hash: $API_KEY" | jq .

# 6. 恢复容器
echo "6. 恢复容器..."
curl -s -X POST $BASE_URL/api/system/containers/$CONTAINER_NAME/unsuspend \
  -H "X-API-Hash: $API_KEY" | jq .

# 7. 删除容器
echo "7. 删除容器..."
curl -s -X DELETE $BASE_URL/api/system/containers/$CONTAINER_NAME \
  -H "X-API-Hash: $API_KEY" | jq .

echo "=== 测试完成 ==="
```

保存为 `test_lxdapi.sh` 并执行：

```bash
chmod +x test_lxdapi.sh
./test_lxdapi.sh
```

## 🚀 下一步

1. ⏳ 在 main.go 中添加 lxdapi 路由
2. ⏳ 启动服务并运行测试
3. ⏳ 修复发现的问题
4. ⏳ 使用 lxdapi WHMCS 插件进行实际测试
5. ⏳ 更新文档
6. ⏳ 提交代码到 GitHub

## 📌 注意事项

1. **容器命名规则**：lxdapi 使用 `lxd11451{userid}{serviceid}` 格式
2. **API Key 管理**：确保每个用户有唯一的 API Key
3. **权限隔离**：用户只能管理自己的容器
4. **响应格式**：必须使用 `{code, msg, data}` 格式
5. **错误处理**：所有错误都应返回正确的 HTTP 状态码和错误消息
