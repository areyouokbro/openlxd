# OpenLXD 多租户管理和镜像市场设计文档

## 一、多租户管理系统

### 1.1 用户系统

**数据库表：users**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    role VARCHAR(20) DEFAULT 'user',  -- admin, user
    status VARCHAR(20) DEFAULT 'active',  -- active, suspended, deleted
    created_at DATETIME,
    updated_at DATETIME
);
```

**字段说明：**
- `username`: 用户名（唯一）
- `email`: 邮箱（唯一）
- `password_hash`: 密码哈希（bcrypt）
- `api_key`: API 密钥（用于 WHMCS 等财务系统对接）
- `role`: 角色（admin 管理员、user 普通用户）
- `status`: 状态（active 活跃、suspended 暂停、deleted 已删除）

### 1.2 容器所有权

**修改现有 containers 表：**
```sql
ALTER TABLE containers ADD COLUMN user_id INTEGER;
ALTER TABLE containers ADD COLUMN created_by VARCHAR(100);
```

**字段说明：**
- `user_id`: 容器所属用户ID
- `created_by`: 创建者（用于审计）

### 1.3 认证方式

**两种认证方式：**

1. **JWT Token 认证**（Web 界面使用）
   - 用户登录后获取 JWT Token
   - Token 有效期 24 小时
   - 包含用户ID、角色等信息

2. **API Key 认证**（WHMCS 等财务系统使用）
   - 请求头：`X-API-Key: <api_key>`
   - 每个用户有唯一的 API Key
   - API Key 可以重新生成

### 1.4 权限控制

**权限规则：**
- **普通用户（user）：**
  - 只能查看和管理自己的容器
  - 只能使用自己的配额
  - 不能访问其他用户的资源

- **管理员（admin）：**
  - 可以查看和管理所有容器
  - 可以管理所有用户
  - 可以修改系统配置

---

## 二、WHMCS 兼容 API

### 2.1 API 设计原则

**兼容 lxdapi 项目的 API 规范：**
- 使用 API Key 认证
- RESTful 风格
- JSON 格式响应
- 标准化错误码

### 2.2 核心 API 端点

**用户管理：**
```
POST   /api/v1/users/register          # 注册用户
POST   /api/v1/users/login             # 用户登录
POST   /api/v1/users/regenerate-key    # 重新生成 API Key
GET    /api/v1/users/profile           # 获取用户信息
```

**容器管理（WHMCS 对接）：**
```
POST   /api/v1/whmcs/container/create   # 创建容器（WHMCS）
POST   /api/v1/whmcs/container/start    # 启动容器
POST   /api/v1/whmcs/container/stop     # 停止容器
POST   /api/v1/whmcs/container/restart  # 重启容器
POST   /api/v1/whmcs/container/delete   # 删除容器
GET    /api/v1/whmcs/container/info     # 获取容器信息
POST   /api/v1/whmcs/container/config   # 修改容器配置
```

### 2.3 WHMCS API 请求格式

**请求头：**
```
X-API-Key: <user_api_key>
Content-Type: application/json
```

**创建容器请求示例：**
```json
{
    "name": "container-001",
    "image": "ubuntu/22.04",
    "cpu": 2,
    "memory": "2GB",
    "disk": "20GB",
    "ipv4": "auto",
    "ipv6": "auto"
}
```

**响应示例：**
```json
{
    "success": true,
    "data": {
        "container_id": 123,
        "name": "container-001",
        "status": "running",
        "ipv4": "10.0.0.100",
        "ipv6": "fd42::100"
    },
    "message": "Container created successfully"
}
```

---

## 三、镜像模板市场

### 3.1 镜像数据源

**数据源：** https://images.linuxcontainers.org/

**支持的镜像类型：**
- Ubuntu (18.04, 20.04, 22.04, 24.04)
- Debian (10, 11, 12)
- CentOS (7, 8, 9)
- Alpine (3.16, 3.17, 3.18, 3.19)
- Rocky Linux (8, 9)
- Fedora (38, 39, 40)

### 3.2 镜像管理

**数据库表：images**
```sql
CREATE TABLE images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alias VARCHAR(100) UNIQUE NOT NULL,
    fingerprint VARCHAR(64),
    distribution VARCHAR(50),
    release VARCHAR(50),
    architecture VARCHAR(20) DEFAULT 'amd64',
    variant VARCHAR(50) DEFAULT 'default',
    description TEXT,
    size INTEGER,
    status VARCHAR(20) DEFAULT 'available',  -- available, downloading, imported, failed
    imported_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);
```

**字段说明：**
- `alias`: 镜像别名（如 ubuntu/22.04）
- `fingerprint`: LXD 镜像指纹
- `distribution`: 发行版（ubuntu, debian, centos 等）
- `release`: 版本号（22.04, 11, 9 等）
- `architecture`: 架构（amd64, arm64）
- `variant`: 变体（default, cloud）
- `status`: 状态（available 可用、downloading 下载中、imported 已导入、failed 失败）

### 3.3 镜像 API 端点

```
GET    /api/v1/images/list              # 获取镜像列表
GET    /api/v1/images/remote            # 获取远程镜像列表
POST   /api/v1/images/import            # 导入镜像
DELETE /api/v1/images/:alias            # 删除镜像
GET    /api/v1/images/:alias/info       # 获取镜像信息
```

### 3.4 镜像导入流程

1. 用户从远程镜像列表选择镜像
2. 调用 `/api/v1/images/import` 接口
3. 后端从 images.linuxcontainers.org 下载镜像
4. 导入到 LXD
5. 更新数据库状态
6. 返回导入结果

**导入请求示例：**
```json
{
    "alias": "ubuntu/22.04",
    "architecture": "amd64"
}
```

---

## 四、Web 界面更新

### 4.1 新增页面

**用户管理页面：**
- 用户列表（管理员）
- 用户详情
- API 密钥管理
- 登录/注册页面

**镜像市场页面：**
- 远程镜像列表（可搜索、过滤）
- 本地镜像列表
- 镜像导入进度
- 从镜像创建容器

### 4.2 界面布局

**新增 Tab 标签：**
- 用户管理（仅管理员可见）
- 镜像市场

**登录页面：**
- 用户名/密码登录
- 记住登录状态
- 注册链接

---

## 五、实施步骤

### 阶段1：多租户管理（3-4天）

**Day 1：用户系统**
- 创建 users 表
- 实现用户注册、登录 API
- 实现 JWT Token 生成和验证
- 实现 API Key 生成和验证

**Day 2：权限控制**
- 修改 containers 表添加 user_id
- 实现权限中间件
- 修改现有 API 添加权限检查
- 实现用户管理 API

**Day 3：WHMCS API**
- 实现 WHMCS 兼容的容器管理 API
- 实现 API Key 认证中间件
- 编写 API 文档

**Day 4：Web 界面**
- 实现登录/注册页面
- 实现用户管理页面
- 实现 API 密钥管理页面

### 阶段2：镜像模板市场（2-3天）

**Day 5：镜像管理后端**
- 创建 images 表
- 实现镜像列表 API
- 实现远程镜像获取（从 images.linuxcontainers.org）
- 实现镜像导入功能

**Day 6：镜像管理前端**
- 实现镜像市场页面
- 实现远程镜像浏览
- 实现镜像导入界面
- 实现从镜像创建容器

**Day 7：测试和发布**
- 功能测试
- API 测试
- 编写文档
- 提交代码并发布新版本

---

## 六、技术栈

**后端：**
- Go 1.21.13
- JWT 认证：`github.com/golang-jwt/jwt/v5`
- 密码加密：`golang.org/x/crypto/bcrypt`
- LXD 客户端：`github.com/canonical/lxd/client`

**前端：**
- Vanilla JavaScript
- Fetch API
- LocalStorage（存储 Token）

**数据库：**
- SQLite + GORM

---

## 七、安全考虑

1. **密码安全：**
   - 使用 bcrypt 加密密码
   - 密码强度验证（最少8位）

2. **API 安全：**
   - API Key 使用随机生成（64位十六进制）
   - JWT Token 有效期限制
   - HTTPS 传输（生产环境）

3. **权限隔离：**
   - 严格的用户权限检查
   - 容器所有权验证
   - 防止越权访问

4. **输入验证：**
   - 所有 API 输入验证
   - SQL 注入防护（GORM ORM）
   - XSS 防护

---

## 八、兼容性

**与现有功能兼容：**
- 现有容器管理功能保持不变
- 向后兼容现有 API
- 管理员可以访问所有功能

**WHMCS 插件兼容：**
- 遵循 lxdapi 项目的 API 规范
- 标准化的请求/响应格式
- 详细的错误码和消息

---

## 九、预期成果

**功能完整度：** 95% → 98%

**新增内容：**
- 2 个数据库表（users, images）
- 15+ 个 API 端点
- 3 个 Web 页面
- ~2000 行代码

**最终版本：** v3.5.0 或 v4.0.0（重大更新）
