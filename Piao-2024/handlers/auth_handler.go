package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"Piao/config"
	"Piao/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register 用户注册
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ 注册请求解析失败: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("📝 注册请求: username=%s\n", req.Username)

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ 密码加密失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// 插入数据库
	_, err = config.DB.Exec("INSERT INTO users (username, password, level) VALUES (?, ?, ?)",
		req.Username, string(hashedPassword), 1)
	if err != nil {
		log.Printf("❌ 用户注册失败: %v\n", err)
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	log.Printf("✅ 用户注册成功: %s\n", req.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "注册成功"})
}

// Login 用户登录
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("❌ 登录请求解析失败: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("🔐 登录请求: username=%s\n", credentials.Username)

	// 查询用户
	var user models.User
	err := config.DB.QueryRow("SELECT id, username, password, level FROM users WHERE username = ?",
		credentials.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Level)
	if err != nil {
		log.Printf("❌ 用户不存在: %s\n", credentials.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		log.Printf("❌ 密码错误: username=%s\n", credentials.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Level:    user.Level,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		log.Printf("❌ Token生成失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ 登录成功: %s (Level: %d)\n", user.Username, user.Level)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    tokenString,
		"username": user.Username,
		"level":    user.Level,
	})
}

// Upgrade 用户升级
func Upgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	currentLevel := r.Header.Get("X-User-Level")

	if currentLevel != "1" {
		http.Error(w, "您已经是高级用户了", http.StatusBadRequest)
		return
	}

	var data struct {
		Answer string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 验证答案
	correctAnswer := "杭电助手"
	if strings.TrimSpace(data.Answer) != correctAnswer {
		log.Printf("❌ 升级失败: userID=%s, 错误答案\n", userID)
		http.Error(w, "答案错误", http.StatusUnauthorized)
		return
	}

	// 更新用户等级
	result, err := config.DB.Exec("UPDATE users SET level = 2 WHERE id = ?", userID)
	if err != nil {
		log.Printf("❌ 升级失败: %v\n", err)
		http.Error(w, "升级失败", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "用户不存在", http.StatusNotFound)
		return
	}

	log.Printf("✅ 用户升级成功: userID=%s\n", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "升级成功！",
		"level":   2,
	})
}
