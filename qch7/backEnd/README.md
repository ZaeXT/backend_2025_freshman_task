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

### 权限分级
当前通过 `JWT` 的 `role` 字段做基础控制（`free`/`pro`/`admin`），可在中间件中扩展到模型访问控制。


