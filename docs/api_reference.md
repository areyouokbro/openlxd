# API 接口参考

所有请求均应发送至 `http://<server_ip>:8443/api/system`。

## 容器管理

### 获取容器列表
- **URL**: `/containers`
- **Method**: `GET`
- **Response**: 返回所有容器的 JSON 数组。

### 创建容器
- **URL**: `/containers`
- **Method**: `POST`
- **Payload**:
```json
{
  "name": "test-container",
  "image": "ubuntu/22.04",
  "cpu": 1,
  "memory": 512,
  "disk": 10240,
  "password": "your-password"
}
```

### 容器操作
- **URL**: `/containers/:name/action`
- **Method**: `POST`
- **Query Params**: `action=start|stop|restart|reinstall`

### 删除容器
- **URL**: `/containers/:name`
- **Method**: `DELETE`

## 流量管理

### 重置流量统计
- **URL**: `/traffic/reset`
- **Method**: `POST`
- **Query Params**: `name=container-name`

## 错误码说明
- `200`: 操作成功。
- `400`: 请求参数错误。
- `500`: 后端处理失败（如 LXD 通信错误）。
