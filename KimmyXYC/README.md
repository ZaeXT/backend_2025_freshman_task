AIBackend - 简易AI问答后端 (Golang + PostgreSQL)

简介
- 使用 Gin + GORM + PostgreSQL 实现的简易问答后端，支持：
  - 用户注册/登录（JWT）
  - 会话与消息存储（PostgreSQL）
  - 基于角色的模型访问控制（free/pro/admin）
  - 上下文对话与流式输出（SSE）
  - 可插拔的模型提供方接口（默认 Mock，便于本地演示）

快速开始
1) 准备 Postgres 并创建数据库，例如：aibackend

2) 配置环境变量（可创建 .env 文件）
- DATABASE_URL=postgres://<user>:<pass>@localhost:5432/aibackend?sslmode=disable
- JWT_SECRET=change-me
- ADDR=:8080

3) 运行
- go run ./cmd/server

4) 健康检查
- GET http://localhost:8080/health -> {"status":"ok"}

API 文档
- 见 docs/api.md，包含注册、登录、会话、聊天（支持 SSE 流式）等接口说明与 curl 示例。

模型提供方
- 默认使用 MockProvider（无需外部 Key，本地直接演示）。
- OpenAI 兼容：当设置 OPENAI_API_KEY 时，自动切换为 OpenAI Chat Completions 协议（支持流式）。
  - 环境变量：
    - OPENAI_API_KEY=你的密钥
    - OPENAI_API_BASE=自定义 Endpoint（可选，默认 https://api.openai.com）
  - Chat 请求将发送到 {OPENAI_API_BASE}/v1/chat/completions（stream=true），并解析 SSE 的 data: 行，提取 choices[0].delta.content。
- 若需接入火山引擎（Volcengine）或其他厂商：
  1. 在 internal/provider 中实现 LLMProvider 接口。
  2. 在 NewProviderFromEnv 中根据环境变量选择对应 Provider。
  3. 在 ChatService 中无需改动，保持调用接口不变。

项目结构
- cmd/server/main.go           程序入口
- internal/db                  数据库连接与迁移
- internal/models              GORM 模型（User/Conversation/Message）
- internal/provider            模型提供方接口与 Mock 实现
- internal/services            Auth 与 Chat 业务逻辑
- internal/httpserver          Gin 路由与 HTTP 处理器
- pkg/auth                     JWT 生成与解析
- pkg/middleware               Gin 中间件（鉴权与模型权限）
- docs/api.md                  API 文档

角色与模型权限（示例）
- free:  [mock-mini]
- pro:   [mock-mini, mock-pro]
- admin: [mock-mini, mock-pro, mock-admin]

注意
- 初次运行会自动迁移数据库表结构。
- 流式输出采用 SSE（text/event-stream）。
- 若未设置 DATABASE_URL，程序会尝试使用本地默认串，但数据库必须实际可连接。



前端（Web）
- 本项目内置了一个简单的 Web 前端，覆盖了后端的全部核心功能：
  - 注册/登录（JWT 持久化在 localStorage）
  - 查看会话列表、查看会话消息
  - 发起聊天（支持非流式与流式输出）
  - 模型选择（会提示当前角色可用的模型范围，越权将被后端拦截）
- 代码位置：web/
  - index.html          页面骨架
  - css/styles.css      样式
  - js/state.js         本地状态与 Token 管理
  - js/api.js           封装所有后端 API 调用（含流式解析）
  - js/auth.js          登录与注册表单逻辑
  - js/chat.js          会话/消息渲染与发送消息（含流式处理）
  - js/main.js          应用入口与视图切换

如何使用前端
1) 启动后端：
   - go run ./cmd/server
2) 打开浏览器访问：
   - http://localhost:8080
3) 在首页进行注册或登录：
   - 你可以选择角色（free/pro/admin），不同角色允许的模型不同：
     - free: [mock-mini]
     - pro:  [mock-mini, mock-pro]
     - admin:[mock-mini, mock-pro, mock-admin]
4) 进入应用后：
   - 左侧为会话列表，可点击切换；
   - 右侧为消息区与输入框；
   - 顶部可切换模型、开启/关闭“流式输出”；
   - 点击“+ 新建对话”开始新的会话；
   - 输入问题后按“发送”即可，开启“流式输出”时可看到回复逐步出现。

注意事项
- 前端与后端同域部署（由 Gin 静态服务提供），无需额外配置 CORS；
- 流式回复采用 fetch ReadableStream 解析服务端的 text/event-stream（后端使用 POST 返回 SSE 风格数据，无法直接用 EventSource，因此用 fetch 流解析）；
- 若模型越权，后端会返回 403，前端会在输入区下方给出提示；
- Mock 模型会回显你的输入，便于本地验证体验。
