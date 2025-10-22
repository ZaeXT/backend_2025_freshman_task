

## ✨ 核心功能

-   **用户账户系统**:
    -   **用户注册与登录**（密码使用 bcrypt 加密）。
    -   基于 **JWT (JSON Web Token)** 的无状态 API 认证。
-   **对话体验**:
    -   支持与大语言模型（**火山引擎方舟大模型**）进行**流式对话 (SSE)**。
    -   **上下文记忆**，支持流畅的多轮对话。
    -   **用户记忆功能**，允许用户保存个人信息，让 AI 提供更具个性化的回答。
-   **模型权限管理**:
    -   **多等级模型访问**：可配置不同用户等级（如 `free`, `premium`）可使用的 AI 模型。
    -   **继承式权限**：高级用户自动获得所有低级用户的模型使用权限。
-   **对话管理**:
    -   **AI 自动生成标题**：在新对话开始时，后台会自动调用 AI 为对话生成一个简洁的摘要标题。
    -   **多级对话分类**：用户可以创建树状结构的分类来组织对话。
    -   **AI 自动分类**：一键调用 AI，智能地将当前对话归入最合适的分类。
-   **数据管理**:
    -   **对话回收站**：删除的对话会先进入回收站，可恢复或永久删除。
    -   **自动清理机制**：后台定时任务（Cron Job）会自动永久删除回收站中的过期对话（默认30天）。
-   **架构**:
    -   **分层架构** (Handler, Service, Repository)。
    -   支持**安全停机**，确保服务在关停时数据的一致性和安全性。

## 🛠️ 技术栈

-   **语言**: Go
-   **Web 框架**: Gin
-   **数据库 ORM**: GORM
-   **配置管理**: Viper
-   **API 认证**: JWT for Go
-   **定时任务**: robfig/cron
-   **密码加密**: golang.org/x/crypto/bcrypt

## 🚀 快速开始

### 1. 先决条件

-   Go (版本 1.18 或更高)
-   MySQL 或 PostgreSQL 数据库
-   一个有效的火山引擎方舟大模型 API Key

### 2. 配置

1.  将项目克隆到本地：
    ```bash
    git clone https://github.com/ZaeXT/backend_2025_freshman_task.git
    cd backend_2025_freshman_task/ZaeXT
    ```
2.  复制配置文件模板：
    ```bash
    cp configs/config.yaml.example configs/config.yaml
    ```
3.  编辑 `configs/config.yaml` 文件，填入您的配置信息：
    -   **`database`**: 数据库连接信息。
    -   **`jwt.secret`**: 设置一个长且随机的 JWT 密钥。
    -   **`volcengine.api_key`**: 填入火山引擎 API Key。
    -   **`volcengine.available_models`**: 根据火山引擎平台开通的模型，配置模型 ID、名称和访问等级。

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 运行服务

```bash
go run cmd/server/main.go
```
服务将在 `config.yaml` 中配置的端口上启动（默认为 `8080`）。

## 📁 项目结构

本项目采用清晰、可扩展的分层架构。主要目录结构和职责如下：

```
.
├── cmd/server/            # 程序主入口
│   └── main.go            # 负责初始化所有组件并启动HTTP服务
├── configs/               # 配置文件
│   ├── config.yaml        # 项目的核心配置文件
│   └── config.yaml.example # 配置模板
├── internal/              # 项目内部代码 (核心逻辑)
│   ├── adapter/           # 适配器层 (与外部服务交互)
│   │   └── volcengine/    # 封装火山引擎API的客户端
│   ├── handler/           # HTTP处理层 (控制器)
│   │   ├── middleware/    # 中间件 (如: JWT认证)
│   │   ├── request/       # 定义请求体的JSON结构
│   │   ├── response/      # 定义响应体的JSON结构
│   │   └── *.go           # 具体的业务模块Handler
│   ├── model/             # 数据模型层 (GORM模型)
│   ├── pkg/               # 内部共享工具包
│   │   ├── e/             # 错误码定义
│   │   ├── hash/          # 密码加密
│   │   └── jwt/           # JWT生成与解析
│   ├── repository/        # 数据仓库层 (数据库操作)
│   │   └── db/            # 数据库初始化
│   ├── service/           # 业务逻辑层 (核心业务处理)
│   └── tasks/             # 定时任务
├── test-scripts.py        # 功能测试脚本
├── go.mod                 # Go模块依赖
└── README.md              # 项目说明
```

---

### 代码分层逻辑

-   **请求流**: 一个HTTP请求会从 `handler` 进入，`handler` 调用 `service` 来处理业务逻辑，`service` 调用 `repository` 来操作数据库，`repository` 使用 `model` 来映射数据表。
-   **外部服务调用**: 当 `service` 需要调用外部AI服务时，它会通过 `adapter` 层，`adapter` 负责处理所有与第三方API的通信细节。
-   **关注点分离**: 每一层都只关心自己的职责，例如 `repository` 只管数据库的增删改查，而不知道这些操作是因何而起；`service` 只管业务规则，而不知道数据具体是如何存储的。
-   **可测试性**: 这种清晰的分层使得每一层都可以被独立测试，极大地提高了代码质量和可维护性。


---


## 📖 API 文档

所有需要认证的 API 都需要在请求头中包含 `Authorization` 字段，格式为 `Bearer [your_jwt_token]`。

---

### 认证 (Auth)

-   `POST /api/v1/register`
    -   **功能**: 注册一个新用户。
    -   **请求体**: `{"username": "your_username", "password": "your_password"}`
    -   **成功响应**: `200 OK`
-   `POST /api/v1/login`
    -   **功能**: 用户登录。
    -   **请求体**: `{"username": "your_username", "password": "your_password"}`
    -   **成功响应**: `200 OK`, `{"data": {"token": "..."}}`

---

### 用户 (User)

-   `GET /api/v1/profile`
    -   **功能**: 获取当前登录用户的个人信息。
    -   **成功响应**: `200 OK`, `{"data": {"id": 1, "username": "...", "tier": "free", ...}}`
-   `PUT /api/v1/profile/memory`
    -   **功能**: 更新用户的记忆信息。
    -   **请求体**: `{"memory_info": "我是...，我喜欢..."}`
    -   **成功响应**: `200 OK`

---

### AI 模型 (Models)

-   `GET /api/v1/models`
    -   **功能**: 获取当前用户可用的所有 AI 模型列表。
    -   **成功响应**: `200 OK`, `{"data": [{"id": "...", "name": "..."}, ...]}`

---

### 对话 (Conversations)

-   `POST /api/v1/conversations`
    -   **功能**: 创建一个新的对话。
    -   **请求体 (可选)**: `{"is_temporary": false, "category_id": 123}`
    -   **成功响应**: `200 OK`, `{"data": {"id": 1, "title": "New Chat", ...}}`
-   `GET /api/v1/conversations`
    -   **功能**: 获取当前用户的所有对话列表。
    -   **成功响应**: `200 OK`, `{"data": [...]}`
-   `POST /api/v1/conversations/:id/messages`
    -   **功能**: 在指定对话中发送消息并获取**流式响应**。
    -   **请求体**: `{"message": "你好", "model_id": "your_chosen_model_id"}`
    -   **成功响应**: `200 OK` (SSE stream)
-   `PUT /api/v1/conversations/:id/title`
    -   **功能**: 手动更新对话标题。
    -   **请求体**: `{"title": "我的新标题"}`
    -   **成功响应**: `200 OK`
-   `POST /api/v1/conversations/:id/auto-classify`
    -   **功能**: 请求 AI 自动为该对话进行分类。
    -   **成功响应**: `200 OK`
-   `DELETE /api/v1/conversations/:id`
    -   **功能**: 将对话移入回收站（软删除）。
    -   **成功响应**: `200 OK`

---

### 分类 (Categories)

-   `POST /api/v1/categories`
    -   **功能**: 创建一个新的分类。
    -   **请求体**: `{"name": "我的分类", "parent_id": 456}` (`parent_id` 可选)
    -   **成功响应**: `200 OK`, `{"data": {"id": 1, "name": "...", ...}}`
-   `GET /api/v1/categories`
    -   **功能**: 获取用户的所有分类（树状结构）。
    -   **成功响应**: `200 OK`, `{"data": [...]}`
-   `PUT /api/v1/categories/:id`
    -   **功能**: 更新一个分类。
    -   **请求体**: `{"name": "新名字", "parent_id": 789}`
    -   **成功响应**: `200 OK`
-   `DELETE /api/v1/categories/:id`
    -   **功能**: 删除一个分类及其所有子分类（级联删除）。
    -   **成功响应**: `200 OK`

---

### 回收站 (Recycle Bin)

-   `GET /api/v1/recycle-bin`
    -   **功能**: 查看回收站中的所有对话。
    -   **成功响应**: `200 OK`, `{"data": [...]}`
-   `POST /api/v1/recycle-bin/restore/:id`
    -   **功能**: 从回收站恢复一个对话。
    -   **成功响应**: `200 OK`
-   `DELETE /api/v1/recycle-bin/permanent/:id`
    -   **功能**: 永久删除一个对话。
    -   **成功响应**: `200 OK`

## 🧪 测试

项目内置了多套 Python 测试脚本，用于端到端地验证所有功能和安全性。

1.  **安装测试依赖**:
    ```bash
    pip install requests sseclient-py
    ```
2.  **运行功能测试**: (测试所有成功路径)
    ```bash
    python test_backend.py
    ```
3.  **运行失败与安全测试**: (测试错误处理和权限隔离)
    ```bash
    python test_failures.py
    ```
4.  **运行模型权限继承测试**: (测试不同等级模型调用限制与权限继承)
    ```bash
    python test_premissions.py
    ```
5.  **运行分类递归删除测试**: (测试删除父级分类时自动删除子分类)
    ```bash
    python test_cascade_delete.py
    ```
