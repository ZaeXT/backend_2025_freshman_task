package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// 设置静态文件目录为Web文件夹
	fs := http.FileServer(http.Dir("./Web"))

	// 创建自定义文件服务器处理器，处理SPA路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 获取请求的文件路径
		path := r.URL.Path

		// 检查请求的文件是否存在
		filePath := filepath.Join("./Web", path)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 文件不存在，重定向到登录页面
			http.Redirect(w, r, "/login_test.html", http.StatusFound)
			return
		}

		// 如果是根路径，重定向到登录页面
		if path == "/" {
			http.Redirect(w, r, "/login_test.html", http.StatusFound)
			return
		}

		// 文件存在，正常服务
		fs.ServeHTTP(w, r)
	})

	// 添加聊天室页面路由
	http.HandleFunc("/chatroom", func(w http.ResponseWriter, r *http.Request) {
		// 检查用户是否已登录（通过JWT token）
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// 如果没有token，重定向到登录页面
			http.Redirect(w, r, "/login_test.html", http.StatusFound)
			return
		}

		// 读取聊天室HTML文件并返回
		htmlContent, err := os.ReadFile("./Web/chatroom.html")
		if err != nil {
			http.Error(w, "聊天室页面加载失败", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(htmlContent)
	})

	// 启动服务器
	port := 8000

	// 使用现有的正式证书文件
	certFile := "20735405_www.yyzyyz.click_nginx/www.yyzyyz.click.pem"
	keyFile := "20735405_www.yyzyyz.click_nginx/www.yyzyyz.click.key"

	// 检查证书文件是否存在
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("证书文件不存在: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("私钥文件不存在: %s", keyFile)
	}

	log.Printf("HTML静态文件服务启动在 https://localhost:%d", port)
	log.Printf("访问地址: https://localhost:%d/ai.html", port)
	log.Printf("WebSocket后端服务需要单独启动在端口8080")

	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, nil)
	if err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
