package main

import (
    "houduan_from/config"   // 导入本地 config 包
    "houduan_from/routes"   // 导入本地 routes 包
    "github.com/gin-gonic/gin"
)

func main() {
    config.InitDB()      // 初始化数据库（需要你在 config 包里实现）
    r := gin.Default()   // 创建 Gin 实例
    routes.InitRoutes(r) // 注册路由（需要你在 routes 包里实现）
    r.Run(":8080")       // 启动 HTTP 服务，监听 8080 端口
}
