# API接口文档

## 基础信息

- **基础URL**: `http://localhost:8080`
- **认证方式**: JWT Bearer Token
- **内容类型**: `application/json`

---

## 1. 用户认证相关

### 1.1 用户注册

**接口**: `POST /api/register`

**请求体**:
```json
{
  "username": "string",
  "password": "string"
}
```

**响应示例**:
```json
{
  "message": "注册成功"
}
```

**状态码**:
- `200`: 注册成功
- `400`: 请求格式错误
- `409`: 用户名已存在

---

### 1.2 用户登录

**接口**: `POST /api/login`

**请求体**:
```json
{
  "username": "string",
  "password": "string"
}
```

**响应示例**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "testuser",
  "level": 1
}
```

**字段说明**:
- `token`: JWT令牌，后续请求需在请求头中携带
- `username`: 用户名
- `level`: 用户等级（1=普通用户，2=高级用户）

**状态码**:
- `200`: 登录成功
- `401`: 用户名或密码错误

---

### 1.3 用户升级

**接口**: `POST /api/upgrade`

**请求头**:
```
Authorization: Bearer <your_jwt_token>
```

**请求体**:
```json
{
  "answer": "杭电助手"
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "升级成功！",
  "level": 2
}
```

**说明**:
- 需要回答正确答案"杭电助手"才能升级
- 仅level为1的用户可以升级
- 升级后可使用高级AI模型

**状态码**:
- `200`: 升级成功`
- `400`: 已经是高级用户
- `401`: 答案错误或未授权

---

## 2. 对话管理

### 2.1 获取对话列表

**接口**: `GET /api/conversations`

**请求头**:
```
Authorization: Bearer <your_jwt_token>
```

**响应示例**:
```json
[
  {
    "id": 1,
    "user_id": 1,
    "title": "我的第一个对话",
    "created_at": "2025-10-08T12:00:00Z"
  },
  {
    "id": 2,
    "user_id": 1,
    "title": "关于AI的讨论",
    "created_at": "2025-10-08T14:30:00Z"
  }
]
```

**说明**:
- 返回当前用户的所有对话
- 按创建时间降序排列（最新的在前）

**状态码**:
- `200`: 成功
- `401`: 未授权

---

### 2.2 创建新对话

**接口**: `POST /api/conversation/create`

**请求头**:
```
Authorization: Bearer <your_jwt_token>
```

**请求体**:
```json
{
  "title": "新对话标题"
}
```

**响应示例**:
```json
{
  "id": 123
}
```

**说明**:
- 返回新创建的对话ID
- 该ID用于后续发送消息和查询历史

**状态码**:
- `200`: 创建成功
- `401`: 未授权
- `500`: 服务器错误

---

### 2.3 获取消息列表

**接口**: `GET /api/messages?conversation_id={id}`

**请求头**:
```
Authorization: Bearer <your_jwt_token>
```

**查询参数**:
- `conversation_id`: 对话ID（必需）

**响应示例**:
```json
[
  {
    "id": 1,
    "conversation_id": 1,
    "role": "user",
    "content": "你好，请介绍一下你自己",
    "created_at": "2025-10-08T12:00:00Z"
  },
  {
    "id": 2,
    "conversation_id": 1,
    "role": "assistant",
    "content": "你好！我是AI助手，很高兴为您服务...",
    "created_at": "2025-10-08T12:00:05Z"
  }
]
```

**字段说明**:
- `role`: 消息角色
    - `user`: 用户发送的消息
    - `assistant`: AI回复的消息
- `content`: 消息内容

**状态码**:
- `200`: 成功
- `400`: 缺少conversation_id参数
- `401`: 未授权

---

## 3. 聊天功能

### 3.1 普通聊天（一次性返回）

**接口**: `POST /api/chat`

**请求头**:
```
Authorization: Bearer <your_jwt_token>
```

**请求体**:
```json
{
  "conversation_id": 1,
  "message": "什么是人工智能？",
  "model": "ep-20241008120000-xxxxx"
}
```

**响应示例**:
```json
{
  "response": "人工智能（Artificial Intelligence，AI）是计算机科学的一个分支..."
}
```

**字段说明**:
- `conversation_id`: 对话ID
- `message`: 用户发送的消息
- `model`: 要使用的AI模型名称

**模型权限**:
- 普通用户（level=1）：只能使用基础模型
- 高级用户（level=2）：可使用所有模型（包括模型名包含"ADVANCED"的高级模型）

**状态码**:
- `200`: 成功
- `401`: 未授权
- `403`: 权限不足（普通用户尝试使用高级模型）
- `500`: AI调用失败

---

### 3.2 流式聊天

**接口**: `POST /api/chat/stream`

**请求头**:
```
Authorization: Bearer <your_jwt_token>
```

**请求体**:
```json
{
  "conversation_id": 1,
  "message": "写一首关于春天的诗",
  "model": "ep-20241008120000-xxxxx"
}
```



## 4. 权限说明

### 用户等级体系

| 等级 | 名称 | 模型权限 | 升级方式 |
|------|------|----------|------|
| 1 | 普通用户 | 仅基础模型 | 答题升级 |
| 2 | 高级用户 | 所有模型 | /    |

### 模型访问控制

- **基础模型**: 所有用户都可使用
- **高级模型**: 模型名包含"ADVANCED"，仅level=2用户可用

### 升级流程

1. 调用`/api/upgrade`接口
2. 提交答案"杭电助手"
3. 验证通过后level升级为2
4. 获得高级模型访问权限

---

## 5. 错误响应

### 错误格式

所有错误响应遵循统一格式：

```json
{
  "error": "错误描述信息"
}
```

或纯文本错误消息（根据接口而定）

### 常见HTTP状态码

| 状态码 | 说明 | 常见原因 |
|--------|------|----------|
| 200 | 成功 | 请求正常处理 |
| 400 | 请求错误 | 参数缺失或格式错误 |
| 401 | 未授权 | token无效、过期或未提供 |
| 403 | 禁止访问 | 权限不足 |
| 405 | 方法不允许 | 使用了错误的HTTP方法 |
| 409 | 冲突 | 资源已存在（如用户名重复） |
| 500 | 服务器错误 | 内部错误或外部服务调用失败 |

### 常见错误示例

**1. Token过期**
```
Status: 401
Body: Unauthorized
```

**2. 权限不足**
```
Status: 403
Body: 权限不足，高级模型需要高级用户
```

**3. 参数错误**
```
Status: 400
Body: conversation_id required
```

**4. AI调用失败**
```
Status: 500
Body: AI调用失败: API返回错误状态: 400
```

---

## 6. 认证流程

### 完整认证流程

```
1. 用户注册/登录
   POST /api/register 或 /api/login
   ↓
2. 获取JWT Token
   Response: { "token": "xxx", "username": "xxx", "level": 1 }
   ↓
3. 保存Token（前端）
   localStorage.setItem('token', token)
   ↓
4. 后续请求携带Token
   Header: Authorization: Bearer <token>
   ↓
5. 服务器验证Token
   提取用户信息 → 执行业务逻辑
```

### Token说明

- **格式**: JWT (JSON Web Token)
- **有效期**: 24小时
- **包含信息**: 用户ID、用户名、用户等级
- **使用方式**: 在请求头中添加`Authorization: Bearer <token>`

---

