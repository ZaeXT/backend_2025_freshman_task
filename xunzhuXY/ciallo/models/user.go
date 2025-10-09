// models/user.go
package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// 用户等级定义
const (
	UserLevelFree    = "free"    // 免费用户
	UserLevelBasic   = "basic"   // 基础用户
	UserLevelPremium = "premium" // 高级用户
	UserLevelAdmin   = "admin"   // 管理员
)

// AI模型定义
const (
	AIModelBasic    = "basic"    // 基础模型
	AIModelAdvanced = "advanced" // 高级模型
	AIModelPremium  = "premium"  // 旗舰模型
)

// 用户等级配置
var UserLevelConfig = map[string]struct {
	Name                   string
	MaxConversations       int
	MaxMessagesPerConv     int
	AllowedModels          []string
	CanCreateConversations bool
}{
	UserLevelFree: {
		Name:                   "免费用户",
		MaxConversations:       3,
		MaxMessagesPerConv:     50,
		AllowedModels:          []string{AIModelBasic},
		CanCreateConversations: true,
	},
	UserLevelBasic: {
		Name:                   "基础用户",
		MaxConversations:       10,
		MaxMessagesPerConv:     200,
		AllowedModels:          []string{AIModelBasic, AIModelAdvanced},
		CanCreateConversations: true,
	},
	UserLevelPremium: {
		Name:                   "高级用户",
		MaxConversations:       50,
		MaxMessagesPerConv:     1000,
		AllowedModels:          []string{AIModelBasic, AIModelAdvanced, AIModelPremium},
		CanCreateConversations: true,
	},
	UserLevelAdmin: {
		Name:                   "管理员",
		MaxConversations:       1000,
		MaxMessagesPerConv:     5000,
		AllowedModels:          []string{AIModelBasic, AIModelAdvanced, AIModelPremium},
		CanCreateConversations: true,
	},
}

// AI模型配置
var AIModelConfig = map[string]struct {
	Name          string
	Description   string
	MaxTokens     int
	Temperature   float64
	RequiresLevel string
}{
	AIModelBasic: {
		Name:          "基础模型",
		Description:   "适合日常对话和简单问答",
		MaxTokens:     1024,
		Temperature:   0.7,
		RequiresLevel: UserLevelFree,
	},
	AIModelAdvanced: {
		Name:          "高级模型",
		Description:   "适合复杂问题分析和创意写作",
		MaxTokens:     2048,
		Temperature:   0.7,
		RequiresLevel: UserLevelBasic,
	},
	AIModelPremium: {
		Name:          "旗舰模型",
		Description:   "适合专业领域和深度思考",
		MaxTokens:     4096,
		Temperature:   0.8,
		RequiresLevel: UserLevelPremium,
	},
}

// 用户性别定义
const (
	GenderMale    = "male"
	GenderFemale  = "female"
	GenderUnknown = "unknown"
)

// 升级密码
const UpgradePassword = "114514"

// 用户结构
type User struct {
	ID            string         `json:"id"`
	Username      string         `json:"username"`
	PasswordHash  string         `json:"password_hash"` // 密码哈希值
	Level         string         `json:"level"`
	Gender        string         `json:"gender"`
	Nickname      string         `json:"nickname"`
	CreatedAt     time.Time      `json:"created_at"`
	LastLogin     time.Time      `json:"last_login"`
	Conversations []Conversation `json:"conversations"`
	CurrentModel  string         `json:"current_model"`
}

// 对话会话
type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"` // 对话使用的模型
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages"`
}

// 用户管理器
type UserManager struct {
	users    map[string]*User
	mutex    sync.RWMutex
	dataFile string
}

// 创建新用户管理器
func NewUserManager(dataFile string) *UserManager {
	um := &UserManager{
		users:    make(map[string]*User),
		dataFile: dataFile,
	}
	um.loadUsers()
	return um
}

// 加载用户数据
func (um *UserManager) loadUsers() {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	// 确保目录存在
	dir := filepath.Dir(um.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	data, err := os.ReadFile(um.dataFile)
	if err != nil {
		// 如果文件不存在，初始化空用户列表
		if os.IsNotExist(err) {
			fmt.Printf("用户数据文件不存在，将创建新文件: %s\n", um.dataFile)
			return
		}
		// 其他错误则打印但不panic
		fmt.Printf("读取用户数据文件失败: %v\n", err)
		return
	}

	if len(data) == 0 {
		fmt.Println("用户数据文件为空")
		return
	}

	err = json.Unmarshal(data, &um.users)
	if err != nil {
		fmt.Printf("解析用户数据失败: %v\n", err)
		// 不panic，继续使用空的用户列表
	}

	// 确保所有用户都有等级和当前模型
	for _, user := range um.users {
		if user.Level == "" {
			user.Level = UserLevelFree
		}
		if user.CurrentModel == "" {
			user.CurrentModel = AIModelBasic
		}
	}
}

// 保存用户数据
func (um *UserManager) SaveUsers() error {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	// 确保目录存在
	dir := filepath.Dir(um.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	data, err := json.MarshalIndent(um.users, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化用户数据失败: %v", err)
	}

	// 创建临时文件，然后重命名，避免写入过程中损坏原文件
	tempFile := um.dataFile + ".tmp"
	err = os.WriteFile(tempFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 重命名临时文件为正式文件
	err = os.Rename(tempFile, um.dataFile)
	if err != nil {
		return fmt.Errorf("重命名文件失败: %v", err)
	}

	fmt.Printf("用户数据已保存到: %s\n", um.dataFile)
	return nil
}

// 创建用户
func (um *UserManager) CreateUser(username, password string) (*User, error) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	// 验证密码强度
	if len(password) < 6 {
		return nil, fmt.Errorf("密码长度至少6位")
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	user := &User{
		ID:           generateID(),
		Username:     username,
		PasswordHash: string(passwordHash),
		Level:        UserLevelFree,
		Gender:       GenderUnknown,
		Nickname:     username,
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
		CurrentModel: AIModelBasic,
		Conversations: []Conversation{
			{
				ID:        generateID(),
				Title:     "默认对话",
				Model:     AIModelBasic,
				CreatedAt: time.Now(),
				Messages:  []Message{},
			},
		},
	}

	um.users[user.ID] = user

	// 立即保存
	go func() {
		if err := um.SaveUsers(); err != nil {
			fmt.Printf("保存用户数据失败: %v\n", err)
		}
	}()

	return user, nil
}

// 验证用户密码
func (um *UserManager) VerifyPassword(username, password string) (*User, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	// 查找用户
	var user *User
	for _, u := range um.users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("密码错误")
	}

	return user, nil
}

// 更新用户密码
func (um *UserManager) UpdateUserPassword(userID, newPassword string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 验证新密码强度
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}

	// 生成新密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	user.PasswordHash = string(passwordHash)
	return nil
}

// 验证升级密码
func (um *UserManager) ValidateUpgradePassword(password string) bool {
	return password == UpgradePassword
}

// 更新用户性别
func (um *UserManager) UpdateUserGender(userID, gender string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 验证性别是否有效
	if gender != GenderMale && gender != GenderFemale && gender != GenderUnknown {
		return fmt.Errorf("无效的性别: %s", gender)
	}

	user.Gender = gender
	return nil
}

// 更新用户昵称
func (um *UserManager) UpdateUserNickname(userID, nickname string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	if nickname == "" {
		return fmt.Errorf("昵称不能为空")
	}

	user.Nickname = nickname
	return nil
}

// 检查是否是特殊用户（xunzhu）
func (u *User) IsSpecialUser() bool {
	return u.Username == "xunzhu"
}

// 获取用户称呼（只有xunzhu有特殊称呼）
func (u *User) GetGreeting() string {
	if u.IsSpecialUser() && u.Level == UserLevelAdmin {
		return "哥哥"
	}
	return "" // 其他用户没有特殊称呼
}

// 获取个性化问候语（只有xunzhu有特殊问候）
func (u *User) GetPersonalizedGreeting() string {
	if u.IsSpecialUser() && u.Level == UserLevelAdmin {
		// 根据一天中的时间返回不同的问候语
		hour := time.Now().Hour()
		var timeGreeting string
		switch {
		case hour < 6:
			timeGreeting = "这么晚还在呀"
		case hour < 12:
			timeGreeting = "早上好"
		case hour < 18:
			timeGreeting = "下午好"
		default:
			timeGreeting = "晚上好"
		}

		return fmt.Sprintf("%s，%s", u.GetGreeting(), timeGreeting)
	}
	return "你好" // 其他用户使用普通问候
}

// 根据用户名查找用户
func (um *UserManager) FindUserByUsername(username string) *User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	for _, user := range um.users {
		if user.Username == username {
			return user
		}
	}
	return nil
}

// 根据用户ID查找用户
func (um *UserManager) FindUserByID(userID string) *User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	return um.users[userID]
}

// 更新用户等级
func (um *UserManager) UpdateUserLevel(userID, newLevel string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 验证等级是否有效
	if _, valid := UserLevelConfig[newLevel]; !valid {
		return fmt.Errorf("无效的用户等级: %s", newLevel)
	}

	user.Level = newLevel

	// 如果新等级不支持当前模型，切换到基础模型
	allowedModels := UserLevelConfig[newLevel].AllowedModels
	canUseCurrentModel := false
	for _, model := range allowedModels {
		if model == user.CurrentModel {
			canUseCurrentModel = true
			break
		}
	}
	if !canUseCurrentModel && len(allowedModels) > 0 {
		user.CurrentModel = allowedModels[0]
	}

	return nil
}

// 更新用户当前模型
func (um *UserManager) UpdateUserModel(userID, newModel string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 验证用户是否有权限使用该模型
	if !user.CanUseModel(newModel) {
		return fmt.Errorf("用户等级 %s 无权使用模型 %s", user.Level, newModel)
	}

	user.CurrentModel = newModel
	return nil
}

// 检查用户是否可以使用某个模型
func (u *User) CanUseModel(model string) bool {
	allowedModels := UserLevelConfig[u.Level].AllowedModels
	for _, allowedModel := range allowedModels {
		if allowedModel == model {
			return true
		}
	}
	return false
}

// 获取用户可用的模型列表
func (u *User) GetAllowedModels() []string {
	return UserLevelConfig[u.Level].AllowedModels
}

// 获取用户等级信息
func (u *User) GetLevelInfo() (string, map[string]interface{}) {
	levelConfig := UserLevelConfig[u.Level]
	info := map[string]interface{}{
		"name":              levelConfig.Name,
		"max_conversations": levelConfig.MaxConversations,
		"max_messages":      levelConfig.MaxMessagesPerConv,
		"allowed_models":    levelConfig.AllowedModels,
	}
	return u.Level, info
}

// 检查用户是否可以创建新对话
func (u *User) CanCreateConversation() bool {
	if !UserLevelConfig[u.Level].CanCreateConversations {
		return false
	}
	return len(u.Conversations) < UserLevelConfig[u.Level].MaxConversations
}

// 获取用户当前对话
func (u *User) GetCurrentConversation() *Conversation {
	if len(u.Conversations) == 0 {
		// 如果没有对话，创建一个默认对话
		u.Conversations = append(u.Conversations, Conversation{
			ID:        generateID(),
			Title:     "默认对话",
			Model:     u.CurrentModel,
			CreatedAt: time.Now(),
			Messages:  []Message{},
		})
	}
	return &u.Conversations[len(u.Conversations)-1]
}

// 创建新对话
func (u *User) CreateNewConversation(title string) (*Conversation, error) {
	if !u.CanCreateConversation() {
		return nil, fmt.Errorf("已达到最大对话数量限制")
	}

	conv := Conversation{
		ID:        generateID(),
		Title:     title,
		Model:     u.CurrentModel,
		CreatedAt: time.Now(),
		Messages:  []Message{},
	}
	u.Conversations = append(u.Conversations, conv)
	return &conv, nil
}

// 添加消息到当前对话
func (u *User) AddMessageToCurrentConversation(role, content string) error {
	conv := u.GetCurrentConversation()

	// 检查消息数量限制
	maxMessages := UserLevelConfig[u.Level].MaxMessagesPerConv
	if len(conv.Messages) >= maxMessages {
		return fmt.Errorf("对话消息数量已达到上限 (%d)", maxMessages)
	}

	conv.Messages = append(conv.Messages, Message{
		Role:    role,
		Content: content,
	})

	// 如果对话标题还是默认的，且这是AI的第一个回复，用用户的第一条消息作为标题
	if conv.Title == "默认对话" && len(conv.Messages) == 2 && role == "assistant" {
		if len(conv.Messages) > 0 {
			firstUserMsg := conv.Messages[0].Content
			if len(firstUserMsg) > 20 {
				conv.Title = firstUserMsg[:20] + "..."
			} else {
				conv.Title = firstUserMsg
			}
		}
	}

	return nil
}

// 更新用户登录时间
func (u *User) UpdateLoginTime() {
	u.LastLogin = time.Now()
}

// 获取所有用户（用于调试）
func (um *UserManager) GetAllUsers() []*User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	users := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}
	return users
}

// 生成简单ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
