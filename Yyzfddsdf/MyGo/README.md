# AI 对话系统 API 文档
### 1. 概述
本 API 基于 Go 语言和 Gin 框架构建，是一套完整的 AI 对话系统接口，支持用户认证、对话管理、Token 充值、实时 WebSocket 通信等核心功能。系统采用 JWT（JSON Web Token）实现身份验证，并通过时间戳验证机制防止重放攻击，保障接口安全性。
##### 技术栈
###### 开发语言：Go 1.18+
###### Web 框架：Gin v1.9+
###### ORM 库：GORM v2.0+
###### 数据库：MySQL 8.0+
###### 认证方式：JWT（JSON Web Token）
###### 实时通信：WebSocket
###### AI 交互：集成 Ollama API（支持多模型调用）
### 2. 认证机制
#### 2.1 JWT Token 认证
所有需要用户权限的 API 端点（标记为 “需认证”），必须在请求头中携带有效的 JWT Token，格式如下：
```
http
Authorization: Bearer <your_jwt_token>
```
###### Token 有效期：24 小时（从生成时间起算）
###### Token 包含信息：用户 ID、邮箱、用户名（用于权限校验）
#### 2.2 时间戳验证（防重放攻击）
为防止请求被恶意重放，所有需认证的接口必须额外携带 X-Timestamp  X-Nonce  X-Signature 请求头，值为当前 Unix 时间戳（秒级）。
服务器会验证时间戳有效性，仅接受与服务器当前时间偏差在 10 秒内的请求。
每个nonce在10分钟内只允许使用一次，重复使用会被拒绝请求。
X-Signature由请求头、请求方法、路由、URL参数、请求体共同拼接为字符串，再由密钥签名，在服务端校验，以防止数据篡改。(拼接顺序：HTTP方法 & 路由 & 时间戳 & nonce & 查询参数 & 请求体)

###### X-Signature 签名生成规则：
###### 1. 构建签名字符串，格式为：HTTP方法&请求路径&时间戳&nonce&查询参数&请求体
###### 2. 使用 HMAC-SHA256 算法和共享密钥对签名字符串进行签名
###### 3. 将签名结果转换为十六进制字符串

###### 签名字符串构建示例：
###### GET请求（带查询参数）：
```
GET&/api/conversations&1716215400&5d8e1b7b8e4b4a9d9c7e1f8a7b8c9d0e&page=1&pageSize=10
```
###### POST请求（带请求体）：
```
POST&/api/recharge&1716215400&5d8e1b7b8e4b4a9d9c7e1f8a7b8c9d0e&{"tokenAmount":100}
```
###### WebSocket连接请求：
```
GET&/ws/ai&1716215400&5d8e1b7b8e4b4a9d9c7e1f8a7b8c9d0e
```

##### 2.2.1 Python 时间戳生成示例
```
Python
import time
import requests
import json
import secrets
import hashlib
import hmac

def _build_sign_string(method, path, timestamp, nonce, query_params=None, body=None):
    """构建签名字符串，与服务端保持一致"""
    # 1. 添加HTTP方法
    sign_parts = [method.upper()]
        
    # 2. 添加路由
    sign_parts.append(path)
        
    # 3. 添加时间戳
    sign_parts.append(str(timestamp))
        
    # 4. 添加nonce
    sign_parts.append(nonce)
        
    # 5. 添加查询参数（按键排序）
    if query_params:
        query_parts = []
        for key in sorted(query_params.keys()):
            value = query_params[key]
            # 处理不同类型的参数值
            if isinstance(value, list):
                for v in sorted(value):
                    query_parts.append(f"{key}={v}")
            else:
                query_parts.append(f"{key}={value}")
        if query_parts:
            sign_parts.append("&".join(query_parts))
        
    # 6. 添加请求体（如果有）
    if body:
        sign_parts.append(body)
            
    # 使用&连接所有部分
    return "&".join(sign_parts)

def _generate_nonce():
    """生成加密安全的随机nonce"""
    return secrets.token_hex(16)  # 生成32位十六进制随机字符串

# 示例1: PUT请求包含请求体
# 1. 准备请求参数
method = "PUT"
path = "/api/conversations/123/title"  # 注意这里必须包含 /title
timestamp = str(int(time.time()))
nonce = _generate_nonce()
query_params = None  # 没有查询参数

# 2. 准备请求体
request_body = {
    "title": "更新的对话标题"  # 只有title字段，没有system_prompt
}
body_json = json.dumps(request_body, separators=(',', ':'))

# 3. 构建签名字符串
sign_string = _build_sign_string(method, path, timestamp, nonce, query_params, body_json)
print(f"签名字符串: {sign_string}")

# 4. 生成签名
signature = hmac.new(
    secret_key.encode('utf-8'),
    sign_string.encode('utf-8'),
    hashlib.sha256
).hexdigest()

# 5. 构造请求头
headers = {
    "Authorization": "Bearer your_jwt_token_here",
    "X-Timestamp": timestamp,
    "X-Nonce": nonce,
    "X-Signature": signature,
    "Content-Type": "application/json"
}

# 6. 发送请求
try:
    response = requests.put(
        url="http://localhost:8080" + path,
        headers=headers,
        data=body_json
    )
    print(f"PUT请求响应状态码: {response.status_code}")
    print(f"PUT请求响应结果: {response.json()}")
except Exception as e:
    print(f"PUT请求发生错误: {e}")

# 示例2：GET请求带查询参数（获取对话列表） 
print("=== 示例：GET请求带查询参数（获取对话列表）===") 
# 1. 准备请求参数 
method = "GET" 
path = "/api/conversations" 
timestamp = str(int(time.time())) 
nonce = _generate_nonce() 

# 2. 准备查询参数 
query_params = {
    "page": 1,
    "pageSize": 5
}

# GET请求没有请求体 
body_json = None

# 3. 构建签名字符串 
sign_string = _build_sign_string(method, path, timestamp, nonce, query_params, body_json) 
print(f"签名字符串: {sign_string}") 

# 4. 生成签名 
secret_key = "your_shared_secret_key" 
signature = hmac.new( 
    secret_key.encode('utf-8'), 
    sign_string.encode('utf-8'), 
    hashlib.sha256 
).hexdigest() 

# 5. 构造请求头（包含认证信息） 
headers = { 
    "Authorization": "Bearer your_jwt_token_here",  # 替换为实际的JWT token
    "X-Timestamp": timestamp, 
    "X-Nonce": nonce, 
    "X-Signature": signature, 
    "Content-Type": "application/json" 
} 

# 6. 发送请求 
try: 
    response = requests.get( 
        url="http://localhost:8080" + path, 
        params=query_params,  # GET请求通过params传递查询参数
        headers=headers 
    ) 
    print(f"GET请求响应状态码: {response.status_code}") 
    print(f"GET请求响应结果: {response.json()}") 
except Exception as e: 
    print(f"GET请求发生错误: {e}")

# 示例3: WebSocket连接请求（包含签名）
def _get_headers(self):
    """获取包含认证、时间戳、nonce和签名的请求头"""
    headers = {
        "Content-Type": "application/json"
    }
    if self.token:
        headers["Authorization"] = f"Bearer {self.token}"
        timestamp = str(int(time.time()))
        nonce = self._generate_nonce()
            
        headers["X-Timestamp"] = timestamp
        headers["X-Nonce"] = nonce
            
        # 计算签名
        sign_string = self._build_sign_string("GET", "/ws/ai", timestamp, nonce)
        # 创建HMAC签名
        signature = hmac.new(
            self.secret_key.encode('utf-8'),
            sign_string.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        headers["X-Signature"] = signature
                
    return headers

def connect(self):
    """连接到WebSocket服务器"""
    headers = self._get_headers()

    # 创建WebSocket连接
    self.ws = websocket.WebSocketApp(
        self.ws_url,
        header=headers,
        on_message=self.on_message,
        on_error=self.on_error,
        on_close=self.on_close
    )
    self.ws.on_open = self.on_open

    # 运行WebSocket客户端
    try:
        print(f"正在连接到 {self.ws_url}...")
        self.ws.run_forever()
    except KeyboardInterrupt:
        print("\n程序被用户中断")
        if self.ws:
            self.ws.close()
```

#### 2.2.2 客户端收到的响应种类
客户端在与服务端交互时，会收到以下几种类型的响应：

1. **认证响应**（连接建立后立即返回）
   - 成功响应：
     ```json
     {
       "success": true,
       "message": "认证成功，可以开始聊天"
     }
     ```
   - 失败响应：
     ```json
     {
       "success": false,
       "error": "无效或过期的Token"
     }
     ```

2. **流式响应**（AI回答过程中持续返回）
   - 中间响应（包含回答片段）：
     ```json
     {
       "text": "这是AI回答的一部分内容",
       "done": false
     }
     ```
   - 完成响应（标记回答结束）：
     ```json
     {
       "text": "",
       "done": true
     }
     ```

3. **错误响应**（发生错误时返回）
   - 一般错误：
     ```json
     {
       "error": "token数量不足，请充值",
       "done": true
     }
     ```
   - 系统错误：
     ```json
     {
       "error": "服务器内部错误",
       "done": true
     }
     ```

4. **API接口响应**（HTTP请求的响应）
   - 成功响应：
     ```json
     {
       "message": "操作成功",
       "data": {...}
     }
     ```
   - 错误响应：
     ```json
     {
       "error": "具体错误信息"
     }
     ```

### 3. API 接口详情
#### 3.1 认证相关接口（无需前置认证）
##### 3.1.1 用户注册
###### URL：/api/auth/register
###### 方法：POST
###### 请求头：
```
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 请求体：
```
json
{
  "username": "test_user",    // 用户名（必填，唯一）
  "email": "test@example.com",// 邮箱（必填，唯一，需符合邮箱格式）
  "password": "Test123456"    // 密码（必填，最少6位）
}
```
###### 成功响应（201 Created）：
```
json
{
  "id": 1,
  "username": "test_user",
  "email": "test@example.com"
}
```
###### 失败响应：
```
400 Bad Request：请求参数错误（如邮箱格式无效、密码长度不足）
json
{
  "error": "请求参数错误: Key: 'RegisterPayload.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
409 Conflict：邮箱已被注册
json
{
  "error": "该邮箱已被注册"
}
500 Internal Server Error：服务器内部错误（如密码加密失败、数据库写入失败）
json
{
  "error": "密码加密失败"
}
```
#####  3.1.2 用户登录
###### URL：/api/auth/login
###### 方法：POST
###### 请求头：
```
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 请求体：
```
json
{
  "email": "test@example.com",// 登录邮箱（必填）
  "password": "Test123456"    // 登录密码（必填）
}
```
###### 成功响应（200 OK）：
```
json
{
  "message": "登录成功",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // JWT Token
  "user": {
    "id": 1,
    "username": "test_user",
    "email": "test@example.com",
    "tokenCount": 10 // 初始Token数量（默认10）
  }
}
```
###### 失败响应：
```
400 Bad Request：请求参数错误
json
{
  "error": "请求参数错误: Key: 'LoginPayload.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
401 Unauthorized：邮箱或密码错误（为安全起见，统一返回此提示，不区分具体错误）
json
{
  "error": "邮箱或密码不正确"
}
500 Internal Server Error：Token 生成失败
json
{
  "error": "生成 Token 失败"
}
```
#### 3.2 用户相关接口（需认证）
##### 3.2.1 删除用户数据
###### URL：/api/user
###### 方法：DELETE
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 成功响应（200 OK）：
```
json
{
  "message": "用户数据已成功删除"
}
```
###### 失败响应：
```
401 Unauthorized：未认证（无 Token 或 Token 无效）
json
{
  "error": "用户未认证"
}
500 Internal Server Error：数据库事务失败（如删除对话、消息时出错）
json
{
  "error": "查找用户对话时出错: record not found"
}
```
##### 3.2.5 获取用户 Token 信息
##### 3.2.3 更新用户信息
###### URL：/api/user
###### 方法：PUT
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 请求体：
```
json
{
  "username": "new_username",  // 新用户名（可选）
  "email": "new@example.com"   // 新邮箱（可选）
}
```
###### 成功响应（200 OK）：
```
json
{
  "message": "用户信息更新成功",
  "user": {
    "id": 1,
    "username": "new_username",
    "email": "new@example.com",
    "tokenCount": 10
  }
}
```
###### 失败响应：
```
400 Bad Request：请求参数错误
json
{
  "error": "请求参数错误"
}
409 Conflict：邮箱已被注册
json
{
  "error": "该邮箱已被注册"
}
500 Internal Server Error：更新用户信息失败
json
{
  "error": "更新用户信息失败"
}
```
##### 3.2.4 获取用户 Token 信息
###### URL：/api/token-info
###### 方法：GET
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 成功响应（200 OK）：
```
json
{
  "tokenCount": 8,
  "username": "test_user",
  "email": "test@example.com"
}
```
###### 失败响应：
```
401 Unauthorized：未认证
json
{
  "error": "未认证的用户"
}
404 Not Found：用户不存在（如用户已被删除）
json
{
  "error": "用户不存在"
}
```
#### 3.3 模型相关接口（需认证）
##### 3.3.1 获取可用 AI 模型列表
###### URL：/api/models
###### 方法：GET
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 成功响应（200 OK）：
```json
{
  "models": [
    "qwen:7b",
    "deepseek-r1:8b",
    "deepseek-r1:14b"
  ],
  "count": 3
}
```
#### 3.4 对话相关接口（需认证）
##### 3.4.1 创建新对话
###### URL：/api/conversations
###### 方法：POST
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 请求体：
```
json
{
  "title": "新对话标题" // 对话标题（可选，默认为"新对话"）
}
```
###### 成功响应（201 Created）：
```
json
{
  "id": 1,
  "title": "新对话",
  "messageCount": 0,
  "tokenUsed": 0,
  "createdAt": "2024-05-20T14:30:00Z"
}
```
###### 失败响应：
```
400 Bad Request：请求参数错误
```json
{
  "error": "请求参数错误"
}
```
500 Internal Server Error：创建对话失败
```json
{
  "error": "创建对话失败"
}
```
##### 3.4.2 获取对话列表（分页）
###### URL：/api/conversations
###### 方法：GET
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 查询参数：
###### page：页码（默认 1，正整数）
###### pageSize：每页条数（默认 10，最大 100）
###### 成功响应（200 OK）：
```
json
{
  "conversations": [
    {
      "id": 1,
      "title": "新对话",
      "messageCount": 2,
      "tokenUsed": 5,
      "createdAt": "2024-05-20T14:30:00Z"
    }
  ],
  "total": 1,          // 总对话数
  "page": 1,           // 当前页码
  "pageSize": 10,      // 每页条数
  "totalPages": 1      // 总页数
}
```
###### 失败响应：
```
401 Unauthorized：未认证
```json
{
  "error": "未认证的用户"
}
```
##### 3.4.2 获取对话详情（含消息列表）
###### URL：/api/conversations/{id}（{id} 为对话 ID）
###### 方法：GET
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 成功响应（200 OK）：
```
json
{
  "id": 1,
  "title": "新对话",
  "messageCount": 2,
  "tokenUsed": 5,
  "createdAt": "2024-05-20T14:30:00Z",
  "messages": [
    {
      "id": 1,
      "role": "user",
      "content": "你好",
      "tokenCount": 0,
      "createdAt": "2024-05-20T14:30:00Z"
    },
    {
      "id": 2,
      "role": "assistant",
      "content": "你好！我是小玉，有什么可以帮你的吗？",
      "tokenCount": 5,
      "createdAt": "2024-05-20T14:30:05Z"
    }
  ]
}
```
###### 失败响应：
```
400 Bad Request：缺少对话 ID
```json
{
  "error": "缺少对话ID"
}
```
404 Not Found：对话不存在（或不属于当前用户）
```json
{
  "error": "对话记录不存在"
}
```
##### 3.4.4 删除对话
###### URL：/api/conversations/{id}（{id} 为对话 ID）
###### 方法：DELETE
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 成功响应（200 OK）：
```
json
{
  "message": "删除成功"
}
```
###### 失败响应：
```
404 Not Found：对话不存在
```json
{
  "error": "对话记录不存在"
}
```
500 Internal Server Error：事务提交失败
```json
{
  "error": "事务提交失败: deadlock found when trying to get lock; try restarting transaction"
}
```
##### 3.4.5 更新对话标题
###### URL：/api/conversations/{id}/title（{id} 为对话 ID）
###### 方法：PUT
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 请求体：
```json
{
  "title": "AI助手使用指南" // 新标题（必填）
}
```
###### 成功响应（200 OK）：
```json
{
  "message": "更新成功",
  "title": "AI助手使用指南"
}
```
###### 失败响应：
```
400 Bad Request：标题为空
```json
{
  "error": "请求参数错误: Key: 'req.Title' Error:Field validation for 'Title' failed on the 'required' tag"
}
```
#### 3.5 Token 相关接口（需认证）
##### 3.5.1 Token 充值
###### URL：/api/recharge
###### 方法：POST
###### 请求头：
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
Content-Type: application/json
```
###### 请求体：
```json
{
  "tokenAmount": 100 // 充值数量（必填，最少1）
}
```
###### 成功响应（200 OK）：
```json
{
  "message": "充值成功",
  "tokenCount": 108, // 充值后的总Token数
  "added": 100       // 本次充值数量
}
```
###### 失败响应：
```
400 Bad Request：充值数量无效
```json
{
  "error": "请求参数错误: Key: 'RechargeRequest.TokenAmount' Error:Field validation for 'TokenAmount' failed on the 'min' tag"
}
```
### 4. WebSocket 实时通信接口
#### 4.1 连接信息
###### 连接地址：ws://localhost:8080/ws/ai
###### 认证要求：连接时需携带以下请求头
```
Authorization: Bearer <your_jwt_token>
X-Timestamp: <current_unix_timestamp>
X-Nonce: <random_string>
X-Signature: <hmac_sha256_signature>
```
#### 4.2 消息格式
##### 4.2.1 认证响应（连接建立后立即返回）
###### 成功：
```json
{
  "success": true,
  "message": "认证成功，可以开始聊天"
}
```
###### 失败：
```json
{
  "success": false,
  "error": "无效或过期的Token" // 具体错误信息（如Token无效、时间戳过期等）
}
```
##### 4.2.2 客户端请求格式（发送给服务器）
```json
{
  "prompt": "请解释什么是AI？",  // 提问内容（必填，除非仅加载对话）
  "model": "deepseek-r1:8b",     // AI模型（可选，默认使用deepseek-r1:8b）
  "useContext": true,            // 是否使用对话上下文（可选，默认true）
  "clearContext": false,         // 是否清空当前上下文（可选，默认false）
  "conversationId": 1            // 对话ID（可选，用于加载历史对话）
}
```
##### 4.2.3 服务器响应格式
###### 中间响应（流式返回 AI 回答片段）：
```json
{
  "text": "AI（人工智能）是指", // 回答片段
  "done": false                 // 标记是否完成（false=未完成，true=完成）
}
```
###### 完成响应（回答结束时返回）：
```json
{
  "text": "",  // 完成时text为空
  "done": true
}
```
###### 错误响应（发生错误时返回）：
```json
{
  "error": "token数量不足，请充值", // 具体错误信息
  "done": true
}
```
### 5. 错误处理规范
###### 系统使用标准 HTTP 状态码区分错误类型，响应体统一包含 error 字段描述具体原因：
###### 状态码	含义	常见场景
###### 400	Bad Request（请求参数错误）	邮箱格式无效、密码长度不足、必填参数缺失等
###### 401	Unauthorized（未认证）	无 Token、Token 无效 / 过期、时间戳过期、签名验证失败等
###### 404	Not Found（资源不存在）	用户不存在、对话不存在等
###### 409	Conflict（资源冲突）	邮箱已注册
###### 429	Too Many Requests（限流）	每秒请求数超过 5 次（系统限流策略）
###### 500	服务器内部错误	数据库事务失败、Token 生成失败等

###### 常见错误响应示例：
###### 401 Unauthorized（签名验证失败）：
```json
{
  "error": "签名验证失败"
}
```
###### 429 Too Many Requests（请求过于频繁）：
```json
{
  "error": "请求过于频繁"
}
```
### 6. 系统限流策略
###### 为保障服务稳定性，系统全局采用请求限流策略：
###### 限流规则：每秒最多允许 5 个请求（所有接口共享此限制）
###### 限流响应（429 Too Many Requests）：
```json
{
  "error": "请求过于频繁"
}
```
### 7. 部署与启动
#### 7.1 环境要求
###### Go 1.18+
######  MySql 8.0+
###### Ollama 服务（默认地址：http://10.150.28.241:11434，可在 API_response/client.go 中修改）
#### 7.2 启动步骤
###### 克隆代码到本地：
```bash
git clone <repository_url>
cd MyGo
```
###### 初始化 Go 模块：
```bash
go mod tidy
```
###### 配置数据库（可选，默认使用 root:Yyz123456@tcp(127.0.0.1:3306)/User_System）：
```bash
# 方式1：通过环境变量设置
export DB_DSN="root:your_password@tcp(127.0.0.1:3306)/your_db_name?charset=utf8mb4&parseTime=True&loc=Local"

# 方式2：直接修改database/setup.go中的默认DSN
```
###### 启动服务：
```bash
go run main.go
```
##### 验证服务：
###### HTTP 服务地址：http://localhost:8080
###### WebSocket 地址：ws://localhost:8080/ws/ai
### 8. 注意事项
###### JWT 密钥安全：生产环境中需替换 controllers/auth.go 中的 jwtSecret（当前为 your_very_secret_jwt_key），建议使用环境变量或配置文件存储。
###### 时间戳精度：客户端生成的 X-Timestamp 需为秒级 Unix 时间戳，且确保客户端与服务器时间同步（偏差不超过 10 秒）。
###### Token 消耗规则：AI 回答会消耗 Token（按回答片段数量计数），用户 Token 不足时无法发起新请求，需通过 /api/recharge 接口充值。
###### 数据安全：删除用户 / 对话时会硬删除关联数据（含消息记录），操作前需确认数据重要性
