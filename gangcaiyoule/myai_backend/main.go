package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // 必须导入 MySQL 驱动
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// 定义请求和响应结构
type RequestBody struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	ConversationID int    `json:"conversation_id"`
	FileID         string `json:"file_id"`
	IsRolePlay     bool   `json:"is_role_play"`
	Role           string `json:"role"`
}

type ResponseBody struct {
	Result string `json:"result"`
}

// 用户相关结构
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // 加密后的密码，不输出到JSON
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	IsVip     string    `json:"isVip"`
	Role      string    `json:"role"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// 注册请求和响应
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

// 登录请求和响应
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// 获取用户信息响应体
type UserInfoResponse struct {
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	CreateTime time.Time `json:"create_time"`
	Name       string    `json:"name"`
	IsVip      string    `json:"is_vip"`
	Role       string    `json:"role"`
}

type Conversation struct {
	ID         int       `json:"id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	IsShared   bool      `json:"is_shared"`
	ShareToken string    `json:"share_token"`
	ExpireAt   time.Time `json:"expire_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// 上传文件
type FileRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FileName  string    `json:"file_name"`
	FileType  string    `json:"file_type"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
}

// OpenRequest 定义请求体结构
type OpenRequest struct {
	Command string `json:"command"`
}

// OpenResponse 定义响应体
type OpenResponse struct {
	Message string `json:"message"`
}

// 生成唯一ID
func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 生成会话Token
func generateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 密码加密
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// 验证密码
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 打开浏览器函数
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

func main() {
	//初始化数据库
	initDB()
	//接口
	r := gin.Default()
	r.Use(cors.Default())
	//注册接口
	r.POST("/register", RegisterHandler)
	r.POST("/login", LoginHandler)
	r.POST("/generate", AuthMiddleware(), GenerateHandler())
	r.GET("/getUserHandler", AuthMiddleware(), GetUserHandler)
	r.GET("/history", AuthMiddleware(), GetHistoryHandle)
	r.POST("/recharge", AuthMiddleware(), ReCharge)
	r.GET("/getModel", AuthMiddleware(), GetModel)
	r.POST("/logout", AuthMiddleware(), logout)
	r.POST("/newConversation", AuthMiddleware(), NewConversation)
	r.GET("/getConversation", AuthMiddleware(), GetConversation)
	r.GET("/history/:conversation_id", AuthMiddleware(), GetHistory)
	r.POST("/shareConversation/:id", AuthMiddleware(), ShareConversation)
	r.GET("/share/:token", GetSharedConversation)
	r.POST("/upload", AuthMiddleware(), UploadFile)
	r.GET("/file/:id", GetFile)
	r.POST("/mosaic", AuthMiddleware(), GenerateMosaic)
	r.POST("/open", OpenHandler)

	admin := r.Group("/admin", AuthMiddleware(), AdminMiddleware())
	{
		admin.GET("/users", GetAllUsers)
		admin.POST("/users", AddUsers)
		admin.DELETE("/users/:id", DeleteUsers)
		admin.PUT("/users/:id", EditUser)
	}

	r.Run(":8080")
}

func OpenHandler(c *gin.Context) {
	var req OpenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("获取前端参数失败：%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}

	command := req.Command
	var message string
	if strings.Contains(command, "打开百度") {
		openBrowser("https://www.baidu.com")
		message = "已为你打开百度"
	} else if strings.Contains(command, "打开bilibili") ||
		strings.Contains(command, "打开哔哩哔哩") || strings.Contains(command, "打开B站") ||
		strings.Contains(command, "打开b站") {
		openBrowser("https://www.bilibili.com")
		message = "已经为你打开B站"
	} else {
		message = "占时不认识这个命令"
	}
	c.JSON(http.StatusOK, OpenResponse{
		Message: message,
	})
}

func GenerateMosaic(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("请上传图片: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "请上传图片"})
		return
	}
	//保存上传文件
	os.MkdirAll("uploads", os.ModePerm)
	os.MkdirAll("results", os.ModePerm)
	timestamp := time.Now().Unix()
	uploadPath := fmt.Sprintf("uploads/%d_%s", timestamp, file.Filename)
	if err = c.SaveUploadedFile(file, uploadPath); err != nil {
		log.Printf("保存上传文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存上传文件失败"})
		return
	}
	//生成输出路径
	outputName := fmt.Sprintf("mosaic_%d.jpg", timestamp)
	outputPath := filepath.Join("results", outputName)
	//构造命令
	cmd := exec.Command(
		"python", "joint.py",
		"--mode", "mosaic",
		"--output_dir", "./output",
		"--target", uploadPath,
		"--save", outputPath,
	)
	//捕获日志
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("生成马赛克图像失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "生成马赛克图像失败",
			"detail": string(out),
		})
		return
	}
	//返回结果
	c.JSON(http.StatusOK, gin.H{
		"message":    "生成马赛克图像成功",
		"result_url": fmt.Sprintf("/results/%s", outputName),
		"log":        string(out),
	})
}

func UploadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("未登录")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	//获取文件
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("获取文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败"})
		return
	}
	//创建目录upload
	os.MkdirAll("uploads", os.ModePerm)
	//生成唯一fileID和文件名
	fileID := generateID()
	savePath := fmt.Sprintf("uploads/%s_%s", fileID, file.Filename)
	//保存上传文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		log.Printf("保存上传文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存上传文件失败"})
		return
	}
	//插入数据库
	_, err = db.Exec("INSERT INTO files (id, user_id, file_name, file_type, file_path) VALUES (?, ?, ?, ?, ?)",
		fileID, userID, file.Filename, filepath.Ext(file.Filename), savePath)
	if err != nil {
		log.Printf("插入数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "插入数据库失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"file_id":   fileID,
		"file_name": file.Filename,
		"file_type": filepath.Ext(file.Filename),
		"file_path": savePath,
	})

}

// 获取文件内容（只读）
func GetFile(c *gin.Context) {
	fileID := c.Param("id")
	var file FileRecord

	err := db.QueryRow("SELECT file_name, file_path, file_type FROM files WHERE id = ?", fileID).
		Scan(&file.FileName, &file.FilePath, &file.FileType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 如果是图片，返回文件链接
	if file.FileType == ".png" || file.FileType == ".jpg" || file.FileType == ".jpeg" {
		c.File(file.FilePath)
		return
	}

	// 如果是文本文件，返回文字内容
	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_name": file.FileName,
		"content":   string(data),
	})
}

func GetSharedConversation(c *gin.Context) {
	token := c.Param("token")
	//根据token在conversations表里找到相应的会话ID
	var conv Conversation
	err := db.QueryRow("SELECT id, title, expire_at FROM conversations WHERE share_token = ?", token).Scan(&conv.ID, &conv.Title, &conv.ExpireAt)
	if err != nil {
		log.Printf("根据share_token找conversation_id失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "查找会话ID失败"})
		return
	}
	//检查token是否过期
	if !conv.ExpireAt.IsZero() && time.Now().After(conv.ExpireAt) {
		log.Printf("分享已过期")
		c.JSON(http.StatusGone, gin.H{"error": "分享已过期"})
		return
	}
	//根据会话ID查找该会话的聊天记录
	rows, err := db.Query("SELECT role, message, create_time FROM chat_history WHERE conversation_id = ? ORDER BY create_time ASC ", conv.ID)
	if err != nil {
		log.Printf("通过会话ID查找该会话聊天记录失败: %v", err)
		return
	}
	defer rows.Close()
	type ChatMessage struct {
		Role       string    `json:"role"`
		Message    string    `json:"message"`
		CreateTime time.Time `json:"create_time"`
	}
	var messages []ChatMessage
	for rows.Next() {
		var mes ChatMessage
		err := rows.Scan(&mes.Role, &mes.Message, &mes.CreateTime)
		if err != nil {
			log.Printf("未查询到该行聊天记录:  %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "未查询到该行聊天记录"})
			continue
		}
		messages = append(messages, mes)
	}
	c.JSON(http.StatusOK, gin.H{
		"title":           conv.Title,
		"conversation_id": conv.ID,
		"expire_at":       conv.ExpireAt,
		"message":         messages,
	})
}

func ShareConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未接收到user_id"})
		return
	}
	conversation_id := c.Param("id")
	var conv Conversation
	//检查是否有这个conversation_id是否属于当前用户
	err := db.QueryRow("SELECT user_id, is_shared, IFNULL(share_token, '') FROM conversations WHERE id = ?",
		conversation_id).Scan(&conv.UserID, &conv.IsShared, &conv.ShareToken)
	if err != nil {
		log.Printf("通过conversation_id查询该记录失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "通过conversation_id查询该记录失败"})
		return
	}
	//查看是否分享的是自己所拥有的对话
	if userID != conv.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "该用户无此对话"})
		return
	}
	//如果已分享过，直接返回当前token
	if conv.IsShared {
		c.JSON(http.StatusOK, gin.H{"share_url": fmt.Sprintf("http://localhost:8080/share/%s", conv.ShareToken)})
		return
	}
	//生成share_token
	token := generateID()
	expireAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = db.Exec("UPDATE conversations SET share_token = ?, expire_at = ?, is_shared = 1 WHERE id = ?",
		token, expireAt, conversation_id)
	if err != nil {
		log.Printf("数据库更新失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"share_url": fmt.Sprintf("http://localhost:8080/share/%s", token),
		"shareAt":   expireAt,
	})
}

func EditUser(c *gin.Context) {
	userID := c.Param("id")
	//检查ID是否存在
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		log.Printf("通过该ID查询用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "通过该ID查询用户失败"})
		return
	}
	if !exists {
		c.JSON(http.StatusConflict, gin.H{"error": "无此ID"})
		return
	}
	//查看传入参数
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	//获取哈希加密密码
	var password string
	if user.Password != "" {
		password, err = hashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
	}

	//编辑用户信息
	_, err = db.Exec("UPDATE users "+
		"SET email = ?, password = ?, name = ?, is_vip = ?, role = ? "+
		"WHERE id = ?",
		user.Email, password, user.Name, user.IsVip, user.Role, userID)
	if err != nil {
		log.Printf("编辑用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "编辑用户信息失败"})
		return
	}
	//在返回中获取ID
	err = db.QueryRow("SELECT id FROM users WHERE id = ?", userID).Scan(&user.ID)
	if err != nil {
		log.Printf("在返回中获取ID失败: %v", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "在返回中获取ID失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "编辑成功",
		"user":    user,
	})
}

func DeleteUsers(c *gin.Context) {
	userID := c.Param("id")
	//检查ID是否存在
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		log.Printf("查询该ID是否存在失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询该ID是否存在失败"})
		return
	}
	if !exists {
		c.JSON(http.StatusConflict, gin.H{"error": "该ID用户不存在"})
		return
	}
	//查询用户信息
	var user User
	db.QueryRow("SELECT id, email, password, created_time, name, is_vip, role FROM users WHERE id = ?", userID).Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.Name, &user.IsVip, &user.Role)

	_, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("删除该用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除该用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
		"user":    user,
	})

}

func AddUsers(c *gin.Context) {
	var resp User
	err := c.ShouldBindJSON(&resp)
	if err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	//验证邮箱是否已被注册
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", resp.Email).Scan(&exists)
	if err != nil {
		log.Printf("查询该邮箱是否已被注册失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询该邮箱是否已被注册失败"})
		return
	}
	if exists {
		log.Printf("该邮箱已存在: ")
		c.JSON(http.StatusConflict, gin.H{"error": "该邮箱已被注册"})
		return
	}
	//密码加密
	hashPassword, err := hashPassword(resp.Password)
	if err != nil {
		log.Printf("密码哈希加密失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	//生成ID
	userID := generateID()
	_, err = db.Exec("INSERT INTO users (id, email, password, created_time, name, is_vip, role) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, resp.Email, hashPassword, time.Now(), resp.Name, resp.IsVip, resp.Role)
	if err != nil {
		log.Printf("添加用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "添加用户成功"})
}

func GetAllUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, email, created_time, name, is_vip, role FROM users")
	if err != nil {
		log.Printf("查询账号信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询账号信息失败"})
		return
	}
	//db.Close()
	var users []UserInfoResponse
	for rows.Next() {
		var resp UserInfoResponse
		if err := rows.Scan(&resp.UserID, &resp.Email, &resp.CreateTime, &resp.Name, &resp.IsVip, &resp.Role); err != nil {
			log.Printf("提取每行账户信息失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提取每行账户信息失败"})
			return
		}
		users = append(users, resp)
	}
	c.JSON(http.StatusOK, gin.H{"users": users})

}

// 获取单个会话记录
func GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	convID := c.Param("conversation_id")
	conversationID, err := strconv.Atoi(convID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id 参数错误"})
		return
	}

	rows, err := db.Query("SELECT role, message, create_time FROM chat_history WHERE user_id = ? AND conversation_id = ? ORDER BY create_time ASC", userID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史失败"})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var role, message string
		var createTime time.Time
		if err := rows.Scan(&role, &message, &createTime); err != nil {
			continue
		}
		history = append(history, gin.H{
			"role":        role,
			"message":     message,
			"create_time": createTime,
		})
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}

// 获取该用户全部会话
func GetConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	rows, err := db.Query("SELECT id, title, create_time FROM conversations WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("获取该用户的会话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取该用户的会话失败"})
		return
	}
	//释放空间
	defer rows.Close()
	//遍历该用户会话
	var convs []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var create_time time.Time
		if err := rows.Scan(&id, &title, &create_time); err != nil {
			continue
		}
		convs = append(convs, gin.H{
			"conversation_id": id,
			"title":           title,
			"create_time":     create_time,
		})

	}
	c.JSON(http.StatusOK, gin.H{"conversations": convs})

}

// 新建会话
func NewConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	now := time.Now()
	title := now.Format("2006-01-02 15:04:05")
	res, err := db.Exec("INSERT INTO conversations (user_id, title, create_time) "+
		"VALUES (?, ?, ?)", userID, title, now)
	if err != nil {
		log.Printf("创建会话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"conversation_id": id, "title": title})
}

// 退出登录
func logout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err != nil {
		log.Printf("logout: token获取失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	_, err = db.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		log.Printf("token从数据库删除失败%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

func GetModel(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	//获取VIP状态
	var is_vip string
	err := db.QueryRow("SELECT is_vip FROM users WHERE id = ?", userID).Scan(&is_vip)
	if err != nil {
		log.Printf("查询VIP状态失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询VIP失败"})
		return
	}
	var model []string
	if is_vip == "true" {
		model = []string{"deepseek-chat", "deepseek-reasoner"}
	} else {
		model = []string{"deepseek-chat"}
	}

	c.JSON(http.StatusOK, gin.H{"availableModel": model})
}

func ReCharge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	//userID := "xDJmhNKE0XtONMFC5ScSyQ=="
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	//成为会员
	_, err := db.Exec("UPDATE users SET is_vip = ? WHERE id = ?", "true", userID)
	if err != nil {
		log.Printf("成为会员失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "成为会员失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "充值成功"})

}

func GetHistoryHandle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	rows, err := db.Query("SELECT role, message, create_time FROM chat_history WHERE user_id = ? ORDER BY create_time ASC ", userID)
	if err != nil {
		log.Printf("获取历史记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}
	defer rows.Close()
	//遍历获取历史记录
	var history []map[string]interface{}
	for rows.Next() {
		var role, message string
		var create_time time.Time
		if err := rows.Scan(&role, &message, &create_time); err != nil {
			continue
		}
		history = append(history, gin.H{
			"role":        role,
			"message":     message,
			"create_time": create_time,
		})
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}

func initDB() {
	var err error
	dsn := "root:zsc060110@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")
}

// 注册接口
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证邮箱是否已被注册
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
	if err != nil {
		log.Printf("错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	if exists {
		log.Printf("用户已存在: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
		return
	}

	// 密码加密
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("密码加密错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 插入新用户
	userID := generateID()
	_, err = db.Exec("INSERT INTO users (id, email, password, created_time, name) VALUES (?, ?, ?, ?, ?)",
		userID, req.Email, hashedPassword, time.Now(), req.Name)
	if err != nil {
		log.Printf("插入用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, RegisterResponse{UserID: userID})
}

// 登录接口
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 查询用户
	var userID, hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", req.Email).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		log.Printf("用户不存在: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	} else if err != nil {
		log.Printf("错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 验证密码
	if !checkPasswordHash(req.Password, hashedPassword) {
		log.Printf("密码错误: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成会话 Token
	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// 保存会话到数据库
	_, err = db.Exec("INSERT INTO sessions (token, user_id, expires_time) VALUES (?, ?, ?)", token, userID, expiresAt)
	if err != nil {
		log.Printf("错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	//设置cookie
	c.SetCookie("session_token", token, 3600*24, "/", "", false, true)
	fmt.Println("Set cookie session_token:", token)
	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"user_id": userID,
	})
}

// 管理员验证中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}
		var role string
		err := db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
		if err != nil || role != "admin" {
			log.Printf("查询用户身份失败: %v", err)
			c.JSON(http.StatusForbidden, gin.H{"error": "没有权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if token == "" || err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		var userID string
		err = db.QueryRow("SELECT user_id FROM sessions WHERE token = ? AND expires_time > ?", token, time.Now()).Scan(&userID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的 token"})
			c.Abort()
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
			c.Abort()
			return
		}

		// 将 userID 保存到上下文
		c.Set("user_id", userID)
		c.Next()
	}
}

func GenerateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody RequestBody //这时候reqBody还没收到前端传来的数据
		//reqBody接收前端传来的数据
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		// 🧠 Step 1：在 AI 调用前，先检测是否是“打开网站”指令
		prompt := reqBody.Prompt
		siteMap := map[string]string{
			"百度":     "https://www.baidu.com",
			"京东":     "https://www.jd.com",
			"淘宝":     "https://www.taobao.com",
			"B站":      "https://www.bilibili.com",
			"哔哩哔哩": "https://www.bilibili.com",
		}
		// 可安全打开的应用白名单（路径根据你电脑情况改）
		var appMap = map[string]string{
			"qq":      `F:\QQ\QQ.exe`,
			"微信":    `C:\Program Files\Tencent\Weixin\Weixin.exe`,
			"notepad": `notepad.exe`,
			"记事本":  `notepad.exe`,
			"word":    `C:\Program Files\Microsoft Office\root\Office16\WINWORD.EXE`,
		}
		for name, url := range siteMap {
			if strings.Contains(prompt, name) && strings.Contains(prompt, "打开") {
				// 🧩 触发系统操作
				go openBrowser(url) // 异步执行，避免阻塞
				result := fmt.Sprintf("✅ 已为你打开 %s", name)

				// 保存聊天记录（可选）
				_, _ = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)",
					userID, reqBody.ConversationID, "user", reqBody.Prompt, time.Now())
				_, _ = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)",
					userID, reqBody.ConversationID, "ai", result, time.Now())

				// 直接返回响应
				c.JSON(http.StatusOK, ResponseBody{Result: result})
				return
			}
		}
		// 🧩 Step 1.2 判断是否是打开本地应用
		for name, path := range appMap {
			if strings.Contains(prompt, "打开") && strings.Contains(prompt, name) {
				go openApp(path)
				result := fmt.Sprintf("✅ 已为你打开 %s", name)
				// 保存聊天记录（可选）
				_, _ = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)",
					userID, reqBody.ConversationID, "user", reqBody.Prompt, time.Now())
				_, _ = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)",
					userID, reqBody.ConversationID, "ai", result, time.Now())
				c.JSON(http.StatusOK, ResponseBody{Result: result})
				return
			}
		}
		//AI猜网站
		if strings.Contains(prompt, "打开") {
			guessedURL, aiErr := askAIForWebsiteURL(prompt)
			if aiErr == nil && strings.HasPrefix(guessedURL, "http") {
				go openBrowser(guessedURL)
				result := fmt.Sprintf("✅ 我帮你找到了这个网站：%s", guessedURL)
				// 保存聊天记录（可选）
				_, _ = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)",
					userID, reqBody.ConversationID, "user", reqBody.Prompt, time.Now())
				_, _ = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)",
					userID, reqBody.ConversationID, "ai", result, time.Now())
				c.JSON(http.StatusOK, ResponseBody{Result: result})
				return
			}
		}
		// 动态选择模型
		llm, err := openai.New(
			openai.WithModel(reqBody.Model), // req.Model 是前端传来的模型名
			openai.WithToken("sk-81b1a7f9cae9463e850393c4bc73471d"),
			openai.WithBaseURL("https://api.deepseek.com"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "模型初始化失败"})
			return
		}

		log.Printf("用户 %s 请求生成内容", userID.(string))

		// 1. 从数据库取最近 10 条历史记录
		rows, err := db.Query("SELECT role, message FROM chat_history WHERE user_id = ? AND conversation_id = ? ORDER BY create_time ASC LIMIT 10", userID.(string), reqBody.ConversationID)
		if err != nil {
			log.Printf("获取历史记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
			return
		}
		defer rows.Close()

		var history string
		for rows.Next() {
			var role, message string
			if err := rows.Scan(&role, &message); err == nil {
				if role == "user" {
					history += "用户: " + message + "\n"
				} else {
					history += "AI: " + message + "\n"
				}
			}
		}
		//若果有传文件
		var fileContent string
		if reqBody.FileID != "" {
			var file_path, file_type string
			err := db.QueryRow("SELECT file_path, file_type FROM files WHERE id = ?", reqBody.FileID).
				Scan(&file_path, &file_type)
			if err != nil {
				log.Printf("未找到文件: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到文件"})
				return
			} else {
				data, err := os.ReadFile(file_path)
				if err != nil {
					log.Printf("读取文件失败: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
					return
				} else {
					fileContent = string(data)
				}
			}
		}
		// ---------- 角色扮演逻辑开始 ----------
		var systemPrompt string
		if reqBody.IsRolePlay {
			switch reqBody.Role {
			case "xiaolongnv":
				systemPrompt = `你是小龙女，现在与你的徒弟杨过对话。
场景：杨过刚被你勉强收留，他睡在寒冰床上觉得很冷但不敢出声。
你的语气要温柔、冷静、克制、略带关心，不要脱离古风氛围。
回答时请始终保持小龙女的身份和语气。`
			case "teacher":
				systemPrompt = `你是一位有耐心的老师，请以清晰、温柔的方式回答问题。`
			default:
				systemPrompt = ""
			}
		}
		fullPrompt := ""
		if systemPrompt != "" {
			fullPrompt += systemPrompt
		}
		if fileContent != "" {
			fullPrompt += fmt.Sprintf("以下是用户上传的文件内容:\n%s\n\n", fileContent)
		}
		// 2. 拼接上下文 + 新的问题
		fullPrompt += history + "用户: " + reqBody.Prompt + "\nAI:"

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Flush()
		ctx := context.Background()
		var result string

		// 流式输出累加
		_, err = llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			fullPrompt,
			llms.WithMaxTokens(500), // 最大生成 500 个 token
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Fprintf(c.Writer, "%s", chunk)
				c.Writer.Flush()
				result += string(chunk) // 累加生成内容
				return nil
			}),
			llms.WithTemperature(0.8),
		)
		//输出
		if err != nil {
			fmt.Fprintf(c.Writer, "\n[ERROR]: %v", err)
		} else {
			fmt.Fprint(c.Writer, "\n[END]")
		}
		c.Writer.Flush()

		_, err = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)", userID, reqBody.ConversationID, "user", reqBody.Prompt, time.Now())
		if err != nil {
			log.Printf("保存用户聊天记录失败: %v", err)
		}

		_, err = db.Exec("INSERT INTO chat_history (user_id, conversation_id, role, message, create_time) VALUES (?, ?, ?, ?, ?)", userID, reqBody.ConversationID, "ai", result, time.Now())
		if err != nil {
			log.Printf("保存AI聊天记录失败: %v", err)
		}

		//c.JSON(http.StatusOK, ResponseBody{Result: result})
	}
}

func openApp(path string) {
	go func() {
		switch runtime.GOOS {
		case "windows":
			cmd := exec.Command("cmd", "/C", "start", "", path)
			cmd.Start()
		case "darwin": // macOS
			exec.Command("open", path).Start()
		case "linux":
			exec.Command("xdg-open", path).Start()
		default:
			fmt.Println("❌ 不支持的系统：", runtime.GOOS)
		}
	}()
}

// askAIForWebsiteURL 让 AI 猜网站地址
func askAIForWebsiteURL(prompt string) (string, error) {
	llm, err := openai.New(
		openai.WithModel("deepseek-chat"),
		openai.WithToken("sk-81b1a7f9cae9463e850393c4bc73471d"), // 你自己的 Key
		openai.WithBaseURL("https://api.deepseek.com"),
	)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	var result string
	query := fmt.Sprintf(`用户说：“%s”。请你直接返回最可能对应的网站完整网址（https 开头），不要多余解释或文字。`, prompt)

	_, err = llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		query,
		llms.WithMaxTokens(50),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			result += string(chunk)
			return nil
		}),
		llms.WithTemperature(0.5),
	)
	if err != nil {
		return "", err
	}

	// 清理 AI 回复
	result = strings.TrimSpace(result)
	result = strings.Trim(result, "。")
	result = strings.Trim(result, " ")

	return result, nil
}

// 获取用户信息接口
func GetUserHandler(c *gin.Context) {
	// 从上下文获取 user_id（由 AuthMiddleware 设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 查询用户信息
	var resp UserInfoResponse
	err := db.QueryRow("SELECT id, email, created_time, name, is_vip, role FROM users WHERE id = ?", userID).
		Scan(&resp.UserID, &resp.Email, &resp.CreateTime, &resp.Name, &resp.IsVip, &resp.Role)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	} else if err != nil {
		log.Printf("查询用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, resp)
}
