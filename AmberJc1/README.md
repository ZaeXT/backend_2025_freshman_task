AmberJc1 — Backend 2025 Freshman Task
项目简介

本项目为 2025 杭助后端招新任务，使用 Go 语言 (Golang) 编写，基于 Gin + GORM + MySQL 实现基础用户系统和简易对话接口。

项目包含以下主要功能接口：
用户注册 /api/v1/register
用户登录 /api/v1/login
聊天接口 /api/v1/chat（需登录后携带 token）

1.克隆仓库
git clone https://github.com/AmberJc1/backend_2025_freshman_task.git
cd backend_2025_freshman_task/AmberJc1


2.配置数据库连接
在 config/config.go 中修改你的数据库配置

3.运行项目
go run main.go


4.接口测试
使用APIFOX进行测试。
注册用户
用户登录
发送聊天请求

5.目录结构
.
├── main.go            # 程序入口
├── config/            # 数据库连接配置
├── controllers/       # 路由控制器
├── models/            # 数据模型
└── README.md

作者
AmberJc1