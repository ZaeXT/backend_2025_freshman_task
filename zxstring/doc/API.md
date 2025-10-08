# WebTest API 文档

## 认证方式

大部分API需要通过JWT Token进行认证，有两种方式传递Token：

1. HTTP头部: `Authorization: Bearer <token>`
2. 查询参数: `?token=<token>`

   > 除非特殊说明，只支持HTTP头部`Authorization: Bearer <token>`作为认证方式


## 用户相关接口

### 用户注册

```
POST /v1/users
```

---

**请求参数**

```json
{
	"username": "string",
	"password": "string"
}
```

**响应示例**

```json
{
  "message": "User created successfully",
  "data": {
    "username": "example_user"
  }
}
```

> 注意，username作为用户的唯一标识符不可重复且不可找回，因此需要先确保用户名可用。
用户名可用性检测详见	**检查用户名可用性**


### 检查用户名可用性

```
GET /v1/check-username
```

**请求参数**

| 参数名   | 类型   | 必需 | 说明           |
| ---------- | -------- | ------ | ---------------- |
| username | string | 是   | 要检查的用户名 |

**响应示例**

```json
{
  "available": true,
  "message": "Username is available"
}
```

### 用户登录

用户登录获取Token，有效时长24小时。


```
POST /v1/login
```

**请求参数**

```json
{
  "username": "string",  
  "password": "string"   
}
```

**响应示例**

```json
{
  "message": "Login successful",
  "data": {
    "username": "example_user",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 修改密码

```
POST /v1/change-password
```

**请求参数**

```json
{
  "username": "string",    
  "password": "string",    
  "newPassword": "string"   
}
```

**响应示例**

```json
{
  "message": "Password changed successfully"
}
```


### 获取用户对话ID列表

```
GET /v1/user-ChatContentGet
```

需要认证。只需要请求头包含Token。

**响应示例**

```json
{
  "username": "example_user",
  "cids": [
    "conversation1",
    "conversation2"
  ]
}
```

### 获取用户可用模型

```
GET /v1/available-models
```

需要认证。

**响应示例**

```json
{
    "models": [
        "QWEN3LOCAL",
        "DEEPSEEKR1LOCAL"
    ],
    "permission": 0,
    "username": "zxstring"
}
```

## 聊天相关接口

### AI聊天接口

```
GET /v1/AIChatRequest
```

需要认证。支持SSE流式响应。需要以Token作为请求参数(兼容性)

**请求参数**

| 参数名    | 类型   | 必需 | 说明                    |
| ----------- | -------- |----|-----------------------|
| content   | string | 是  | 用户发送的聊天内容             |
| cid       | string | 是  | 用户对话ID，不提供时会自动生成。获取方式 |
| MODELNAME | string | 是  | 模型名称。                 |

**响应格式**

SSE流式响应，遵循OpenAI API格式：

```
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}

data: [DONE]
```

### 获取聊天历史记录

```
GET /v1/chat-history
```

需要认证。

**请求参数**

| 参数名 | 类型   | 必需 | 说明         |
| -------- | -------- | ------ | -------------- |
| cid    | string | 是   | 对话ID       |
| count  | int    | 是   | 获取记录数量 |

**响应示例**

```json
{
  "records": [
    {
      "id": 1,
      "username": "user",
      "model": "deepseek-r1-250528",
      "role": "user",
      "time": 1700000000000,
      "content": "你好",
      "cid": "11"
    },
    {
      "id": 2,
      "username": "assistant",
      "model": "deepseek-r1-250528",
      "role": "assistant",
      "time": 1700000001000,
      "content": "你好！有什么我可以帮助你的吗？",
      "cid": "11"
    }
  ]
}
```

## 数据模型

### User（用户）

| 字段       | 类型   | 说明                        |
| ------------ | -------- | ----------------------------- |
| username   | string | 用户名（主键）              |
| password   | string | MD5加密后的密码             |
| permission | int    | 用户权限等级（0为普通用户） |

目前使用register注册默认分配权限为0

### ModelConfig（模型信息）

在Model_Config.json配置信息如：

```json
{
  "DEEPSEEKR1LOCAL": { 
    "model_full_name": "deepseek-r1:8b",
    "API_KEY_STORE": "ollama_key", 
    "base_url": "http://localhost:11434/v1/",
    "permission": 0	
  },
  "QWEN3LOCAL": {
    "model_full_name": "qwen3:14b",
    "API_KEY_STORE": "ollama_key",
    "base_url": "http://localhost:11434/v1/",
    "permission": 0
  },
  "VOLCENGINE_DEEPSEEK671B": {
    "model_full_name": "deepseek-r1-250528",
    "API_KEY_STORE": "ARK_API_KEY",
    "base_url": "https://ark.cn-beijing.volces.com/api/v3",
    "permission": 1
  },
  "DOUBAOSEED": {
    "model_full_name": "doubao-seed-1-6-250615",
    "API_KEY_STORE": "ARK_API_KEY",
    "base_url": "https://ark.cn-beijing.volces.com/api/v3",
    "permission": 1
  }
}
```
