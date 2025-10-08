# README.md


## 项目说明

简陋的AI聊天系统后端，使用OPENAI SDK

### 功能特点

- 用户注册、登录和密码修改
- 使用OPENAI SDK
- 流式传输对话（SSE流式传输）
- 聊天历史记录保存和查询
- 多轮对话支持
- JWT Token身份验证
- 用户权限管理

### 技术栈

- 语言：GO
- 框架：GIN

### 安装步骤

1. 克隆项目：

    ```bash
    git clone <repository-url>
    cd webtest
    ```
2. 配置模型访问密钥：  
    在环境变量中设置所需的API密钥：

    > 你需要为不同的服务来源配置好不同的APIKEY环境变量。


## 配置说明

### 模型配置 (models_config.json)

根据需要配置可用的模型，每种模型可以配置不同的访问密钥、基础URL（使用OPENAI API）和权限等级。
permission: 可访问的最低权限要求
API_KEY_STORE: 从环境变量中获取key
```json
{
  "DEEPSEEKR1LOCAL": {  
    "model_full_name": "deepseek-r1:8b",  
    "API_KEY_STORE": "ollama_key",		
    "base_url": "http://localhost:11434/v1/",
    "permission": 0	
  }
}
```

