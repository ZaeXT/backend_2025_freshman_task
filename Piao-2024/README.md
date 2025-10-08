# 杭助聊天机器人项目 

## 项目结构

```
chatbot-project/
├── main.go                          # 程序入口
├── go.mod                           # Go模块定义
├── go.sum                           # 依赖锁定
├── .env.example                     # 环境变量配置
├── index.html                       # 前端页面
│
├── config/                          # 配置模块
│   └── config.go                    # 配置加载和管理
│
├── models/                          # 数据模型
│   ├── user.go                      # 用户相关模型
│   └── volcengine.go                # API相关模型
│
├── database/                        # 数据库模块
│   └── database.go                  # 数据库初始化
│
├── middleware/                      # 中间件
│   └── auth.go                      # JWT认证中间件
│
├── handlers/                        # 请求处理器
│   ├── auth_handler.go              # 用户认证
│   ├── conversation_handler.go      # 对话管理
│   ├── chat_handler.go              # 聊天功能
│   └── static_handler.go            # 静态文件服务
│
└── services/                        # 业务逻辑层
    └── volcengine_service.go        # 火山引擎API服务
```

## 🚀 快速开始

### 1. 下载依赖

```bash
   ```

## 🎯 模块功能说明

### config - 配置管理
- 统一管理所有配置项
- 加载环境变量
- 提供全局访问接口

### models - 数据模型
- 定义业务实体
- API请求/响应结构
- 保持数据结构的一致性

### database - 数据库层
- 数据库连接管理
- 自动创建表结构
- 连接池管理

### middleware - 中间件
- JWT token验证
- 请求预处理
- 用户信息注入

### handlers - 控制器
- 处理HTTP请求
- 参数验证
- 响应格式化

### services - 服务层
- 业务逻辑封装
- 第三方API调用
- 可复用的功能模块

## 🔧 开发建议

1. **添加新功能**：在相应模块添加新文件
2. **修改业务逻辑**：优先修改services层
3. **添加新接口**：在handlers中添加，main.go中注册路由
4. **添加新模型**：在models目录添加对应文件

