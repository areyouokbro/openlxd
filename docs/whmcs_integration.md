# WHMCS 插件对接指南

OpenLXD 后端已完全兼容 WHMCS LXD 模块，支持通过 WHMCS 自动化管理容器。

## ✅ 兼容性确认

OpenLXD 后端已实现 WHMCS 模块所需的所有核心功能：

| WHMCS 功能 | API 端点 | 状态 |
|-----------|---------|------|
| 创建容器 | `POST /api/whmcs?action=create` | ✅ 已实现 |
| 暂停容器 | `POST /api/whmcs?action=suspend&hostname={name}` | ✅ 已实现 |
| 恢复容器 | `POST /api/whmcs?action=unsuspend&hostname={name}` | ✅ 已实现 |
| 删除容器 | `POST /api/whmcs?action=terminate&hostname={name}` | ✅ 已实现 |
| 修改密码 | `POST /api/whmcs?action=changepassword&hostname={name}&password={pwd}` | ✅ 已实现 |
| 容器信息 | `GET /api/whmcs?action=info&hostname={name}` | ✅ 已实现 |

## 🔧 WHMCS 服务器配置

### 1. 添加服务器

在 WHMCS 管理后台：

**路径**: `系统设置` → `产品/服务` → `服务器`

**配置参数**:
```
服务器名称: OpenLXD Server 1
主机名: your-domain.com (或 IP 地址)
IP 地址: 156.246.90.151
类型: LXD
用户名: (留空)
密码: (留空)
访问哈希: your-api-key-here
安全: ✓ 使用 SSL
端口: 443
```

### 2. 配置说明

#### 主机名
- **使用域名**: `https://api.yourdomain.com`
- **使用 IP**: `https://156.246.90.151`

#### 访问哈希 (API Key)
从 OpenLXD 配置文件获取：
```bash
cat /etc/openlxd/config.yaml | grep api_hash
```

或从安装日志获取：
```bash
journalctl -u openlxd | grep "API Key"
```

#### SSL 设置
- ✅ **启用 SSL**: 使用 HTTPS (推荐)
- ⚠️ **禁用 SSL**: 仅用于测试环境

## 📦 产品配置

### 1. 创建产品

**路径**: `系统设置` → `产品/服务` → `产品/服务`

**基本设置**:
```
产品类型: 服务器/VPS
产品组: VPS 容器
产品名称: LXD 容器 - 1核1G
```

### 2. 模块设置

**模块**: `LXD`
**服务器**: 选择上面创建的 OpenLXD 服务器

**可配置选项**:
```
主机名: {客户ID}-{产品ID} (自动生成)
镜像: ubuntu/22.04
CPU 核心: 1
内存: 1GB
磁盘: 10GB
```

### 3. 定价设置

根据资源配置设置价格：
```
月付: ¥50.00
季付: ¥135.00 (10% 折扣)
年付: ¥480.00 (20% 折扣)
```

## 🔌 API 端点详解

### 创建容器

**请求**:
```http
POST /api/whmcs?action=create
Content-Type: application/json
X-API-Hash: your-api-key-here

{
  "hostname": "client1-prod1",
  "image": "ubuntu/22.04",
  "cpu": 1,
  "memory": "1GB",
  "disk": "10GB"
}
```

**响应**:
```
success
```

### 暂停容器

**请求**:
```http
POST /api/whmcs?action=suspend&hostname=client1-prod1
X-API-Hash: your-api-key-here
```

**响应**:
```
success
```

### 恢复容器

**请求**:
```http
POST /api/whmcs?action=unsuspend&hostname=client1-prod1
X-API-Hash: your-api-key-here
```

**响应**:
```
success
```

### 删除容器

**请求**:
```http
POST /api/whmcs?action=terminate&hostname=client1-prod1
X-API-Hash: your-api-key-here
```

**响应**:
```
success
```

### 修改密码

**请求**:
```http
POST /api/whmcs?action=changepassword&hostname=client1-prod1&password=NewPassword123
X-API-Hash: your-api-key-here
```

**响应**:
```
success
```

### 获取容器信息

**请求**:
```http
GET /api/whmcs?action=info&hostname=client1-prod1
X-API-Hash: your-api-key-here
```

**响应**:
```json
{
  "hostname": "client1-prod1",
  "ip": "10.0.0.100",
  "status": "Running",
  "cpu": 1,
  "memory": "1GB",
  "disk": "10GB",
  "image": "ubuntu/22.04"
}
```

## 🧪 测试对接

### 使用 curl 测试

```bash
# 设置变量
API_URL="https://your-domain.com"
API_KEY="your-api-key-here"

# 测试创建容器
curl -X POST "$API_URL/api/whmcs?action=create" \
  -H "X-API-Hash: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test-container",
    "image": "ubuntu/22.04",
    "cpu": 1,
    "memory": "1GB",
    "disk": "10GB"
  }'

# 测试获取容器信息
curl "$API_URL/api/whmcs?action=info&hostname=test-container" \
  -H "X-API-Hash: $API_KEY"

# 测试暂停容器
curl -X POST "$API_URL/api/whmcs?action=suspend&hostname=test-container" \
  -H "X-API-Hash: $API_KEY"

# 测试恢复容器
curl -X POST "$API_URL/api/whmcs?action=unsuspend&hostname=test-container" \
  -H "X-API-Hash: $API_KEY"

# 测试删除容器
curl -X POST "$API_URL/api/whmcs?action=terminate&hostname=test-container" \
  -H "X-API-Hash: $API_KEY"
```

## 📝 自定义 WHMCS 模块（可选）

如果需要自定义 WHMCS 模块，可以参考以下代码：

### lib/Api.php

```php
<?php

namespace LXD;

class Api {
    private $apiUrl;
    private $apiKey;
    
    public function __construct($hostname, $apiKey) {
        $this->apiUrl = "https://{$hostname}";
        $this->apiKey = $apiKey;
    }
    
    public function createContainer($params) {
        return $this->request('POST', '/api/whmcs?action=create', [
            'hostname' => $params['hostname'],
            'image' => $params['image'],
            'cpu' => $params['cpu'],
            'memory' => $params['memory'],
            'disk' => $params['disk'],
        ]);
    }
    
    public function suspendContainer($hostname) {
        return $this->request('POST', "/api/whmcs?action=suspend&hostname={$hostname}");
    }
    
    public function unsuspendContainer($hostname) {
        return $this->request('POST', "/api/whmcs?action=unsuspend&hostname={$hostname}");
    }
    
    public function terminateContainer($hostname) {
        return $this->request('POST', "/api/whmcs?action=terminate&hostname={$hostname}");
    }
    
    public function changePassword($hostname, $password) {
        return $this->request('POST', "/api/whmcs?action=changepassword&hostname={$hostname}&password={$password}");
    }
    
    public function getContainerInfo($hostname) {
        return $this->request('GET', "/api/whmcs?action=info&hostname={$hostname}");
    }
    
    private function request($method, $endpoint, $data = null) {
        $ch = curl_init($this->apiUrl . $endpoint);
        
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'X-API-Hash: ' . $this->apiKey,
            'Content-Type: application/json',
        ]);
        
        if ($data && $method === 'POST') {
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
        }
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        
        if ($httpCode !== 200) {
            throw new \Exception("API request failed: " . $response);
        }
        
        return $response;
    }
}
```

## ❓ 常见问题

### Q1: WHMCS 提示"连接失败"

**原因**: SSL 证书问题或 API Key 错误

**解决方案**:
1. 检查 HTTPS 是否正常工作
2. 验证 API Key 是否正确
3. 检查防火墙是否开放 443 端口

### Q2: 容器创建失败

**原因**: 镜像不存在或资源不足

**解决方案**:
1. 在 OpenLXD 后台预先下载镜像
2. 检查服务器资源是否充足

### Q3: 如何查看 API 调用日志

```bash
# 查看 OpenLXD 日志
sudo journalctl -u openlxd -f

# 查看 WHMCS 模块日志
tail -f /path/to/whmcs/modules/servers/lxd/debug.log
```

### Q4: 支持哪些镜像

OpenLXD 支持所有 LXD 官方镜像：
- Ubuntu 22.04, 20.04, 18.04
- Debian 12, 11, 10
- CentOS 7, 8
- Rocky Linux 8, 9
- Alpine Linux

### Q5: 如何自定义容器配置

在 WHMCS 产品配置中添加自定义字段：
```
configoption1: CPU 核心数
configoption2: 内存大小
configoption3: 磁盘大小
configoption4: 镜像选择
```

## 🔒 安全建议

1. **使用 HTTPS**: 始终启用 SSL/TLS 加密
2. **保护 API Key**: 不要在客户端暴露 API Key
3. **限制 IP 访问**: 在防火墙中限制 WHMCS 服务器 IP
4. **定期更新**: 保持 OpenLXD 和 WHMCS 模块最新版本
5. **监控日志**: 定期检查 API 调用日志

## 📞 技术支持

如有问题，请：
1. 查看 [OpenLXD 文档](https://github.com/areyouokbro/openlxd)
2. 提交 [GitHub Issue](https://github.com/areyouokbro/openlxd/issues)
3. 加入社区讨论

## 🎉 完成

现在您已经成功配置了 WHMCS 与 OpenLXD 的对接！客户可以通过 WHMCS 自动购买和管理 LXD 容器了。
