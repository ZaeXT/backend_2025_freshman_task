## 后端服务（Gin + MongoDB + DashScope 兼容AI）

### 快速开始
1. 复制 `.env.example` 为 `.env` 并填好变量（或通过系统环境变量注入）
2. 启动本地 MongoDB（或修改 `MONGO_URI` 指向你的实例）
3. 安装依赖并运行

```bash
go mod tidy
go run ./...
```

服务启动后默认监听 `:8080`。

### 环境变量
- `PORT`：服务端口，默认 `:8080`
- `MONGO_URI`：Mongo 连接串，默认 `mongodb://localhost:27017`
- `MONGO_DB`：数据库名，默认 `qa_app`
- `JWT_SECRET`：JWT 密钥
- `DASHSCOPE_API_KEY`：阿里云百炼 API Key（兼容 OpenAI 风格）
- `AI_BASE_URL`：默认 `https://dashscope.aliyuncs.com/compatible-mode/v1`
- `AI_MODEL`：默认 `qwen-plus`

### 接口

#### 健康检查
```
GET /ping
```

#### 注册
```
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "a@b.com",
  "username": "alice",
  "password": "secret123"
}
```

#### 登录
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "a@b.com",
  "password": "secret123"
}

返回：{"token":"<JWT>"}
```

#### 聊天（非流式）
```
POST /api/v1/chat
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "title": "小测试",
  "model": "qwen-plus",
  "messages": [
    {"role":"system","content":"You are a helpful assistant."},
    {"role":"user","content":"你好！"}
  ],
  "stream": false
}
```

返回：
```
{
  "conversationId": "...",
  "content": "...AI回复..."
}
```

可用模型将根据用户角色自动校验（见“权限分级与模型访问”）。

#### 聊天（SSE流式）
```
POST /api/v1/chat
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "conversationId": "...可选...",
  "messages": [
    {"role":"user","content":"给我讲个笑话"}
  ],
  "stream": true
}
```

浏览器/前端以 EventSource 或自实现方式接收 `event: message\n data: <delta>` 片段。

#### 查询会话（含模型）
```
GET /api/v1/conversations/:id
Authorization: Bearer <JWT>
```

返回示例：
```
{
  "id": "66f1d7...",
  "title": "新对话",
  "model": "qwen-plus",
  "createdAt": "2025-10-09T12:34:56Z",
  "updatedAt": "2025-10-09T12:35:10Z"
}
```

用于校验该会话最终使用的模型（已包含服务端按角色判定的结果）。

#### 设置用户角色（仅管理员）
```
PUT /api/v1/users/:id/role
Authorization: Bearer <JWT(管理员)>
Content-Type: application/json

{
  "role": "free | pro | admin"
}
```
返回：`204 No Content`

### 权限分级与模型访问
系统通过 `JWT` 中的 `role` 字段进行访问控制，角色有：`free` / `pro` / `admin`。聊天接口对模型的访问限制如下：

- free：仅允许 `qwen-plus`（若未指定 `model`，默认使用 `qwen-plus`）
- pro：允许 `qwen3-max` 与 `qwen-plus`（若未指定，默认使用 `qwen3-max`）
- admin：允许任意模型（若未指定，默认使用服务端配置的 `AI_MODEL`）

当请求指定了不被该角色允许的模型时，接口将返回 `403`。创建新会话时，服务端会将最终判定的模型写入 `conversation.model`，可通过上述“查询会话”接口核对。


