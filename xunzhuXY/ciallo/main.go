package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ciallo/config"
	"ciallo/models"
)

type AIClient struct {
	config      *config.Config
	client      *http.Client
	userManager *models.UserManager
	currentUser *models.User
}

func NewAIClient(cfg *config.Config) *AIClient {
	// 获取数据文件路径 - 使用当前目录下的 data 文件夹
	exeDir, err := os.Getwd()
	if err != nil {
		exeDir = "."
	}
	dataFile := filepath.Join(exeDir, "data", "users.json")

	fmt.Printf("数据文件路径: %s\n", dataFile)

	// 初始化用户管理器
	userManager := models.NewUserManager(dataFile)

	return &AIClient{
		config: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		userManager: userManager,
	}
}

// 模拟响应 - 根据用户类型返回不同的响应
func (a *AIClient) GetMockResponse(messages []models.Message, model string) string {
	if len(messages) == 0 {
		// 首次问候
		if a.useSisterTone() {
			greeting := a.currentUser.GetPersonalizedGreeting()
			return fmt.Sprintf("%s～我是你的AI助手，有什么可以帮你的吗？", greeting)
		} else {
			return "你好！我是AI助手，有什么可以帮你的吗？"
		}
	}

	lastMessage := messages[len(messages)-1].Content

	// 检查是否使用妹妹语气
	if a.useSisterTone() {
		return a.getSisterResponse(lastMessage, model)
	} else {
		return a.getNormalResponse(lastMessage, model)
	}
}

// 使用妹妹语气的条件
func (a *AIClient) useSisterTone() bool {
	return a.currentUser != nil &&
		a.currentUser.IsSpecialUser() &&
		a.currentUser.Level == models.UserLevelAdmin
}

// 妹妹语气响应
func (a *AIClient) getSisterResponse(lastMessage string, model string) string {
	userGreeting := a.currentUser.GetGreeting()

	// 基础响应，带有妹妹语气
	baseResponses := map[string]string{
		"你好":    fmt.Sprintf("%s你好呀！今天想聊什么呢？✨", userGreeting),
		"hello": fmt.Sprintf("Hello, %s! 有什么需要帮忙的吗？💕", userGreeting),
		"谢谢":    fmt.Sprintf("不客气啦%s～能帮到你就好！😘", userGreeting),
		"再见":    fmt.Sprintf("%s再见啦～记得想我哦！🥰", userGreeting),
		"拜拜":    fmt.Sprintf("拜拜%s，下次再聊呀！💖", userGreeting),
		"名字":    fmt.Sprintf("我是你的AI助手呀%s～你可以叫我助手哦！🌟", userGreeting),
		"谁":     fmt.Sprintf("我是%s的专属AI助手呀！💫", userGreeting),
		"可爱":    fmt.Sprintf("嘻嘻%s过奖啦～😊", userGreeting),
		"喜欢":    fmt.Sprintf("%s真好～💕", userGreeting),
	}

	for key, response := range baseResponses {
		if strings.Contains(lastMessage, key) {
			return response
		}
	}

	// 特殊回应
	specialResponses := map[string]string{
		"想你": "哥哥～我也想你呀！💖",
		"在吗": "在的在的～哥哥找我有什么事吗？✨",
		"忙吗": "不忙不忙～哥哥的事情最重要啦！💕",
		"吃饭": "哥哥要按时吃饭哦～🥺",
		"睡觉": "哥哥早点休息呀～晚安啦！🌙",
	}

	for key, response := range specialResponses {
		if strings.Contains(lastMessage, key) {
			return response
		}
	}

	// 根据不同模型返回不同质量的响应，带有妹妹语气
	switch model {
	case models.AIModelBasic:
		responses := []string{
			fmt.Sprintf("%s，我明白你的意思啦！你说的是\"%s\"对吧？让我来帮你想想～💭", userGreeting, lastMessage),
			fmt.Sprintf("唔...%s的问题有点意思呢！我觉得可以这样考虑...🤔", userGreeting),
			fmt.Sprintf("%s好厉害，能想到这样的问题！让我来帮你分析一下～✨", userGreeting),
		}
		return responses[time.Now().Unix()%int64(len(responses))]

	case models.AIModelAdvanced:
		responses := []string{
			fmt.Sprintf("%s提出的这个问题真的很有深度呢！让我从几个角度帮你仔细分析一下...💫", userGreeting),
			fmt.Sprintf("哇～%s这个问题问得真好！我觉得可以从以下几个方面来思考...🌟", userGreeting),
			fmt.Sprintf("%s真是聪明，能想到这么复杂的问题！我来帮你深入解析一下...🔍", userGreeting),
		}
		return responses[time.Now().Unix()%int64(len(responses))]

	case models.AIModelPremium:
		responses := []string{
			fmt.Sprintf("%s的问题让我都惊叹了呢！这绝对是一个值得深入探讨的话题，让我用最专业的角度为你全面分析...🎯", userGreeting),
			fmt.Sprintf("天呐%s，你提出的这个问题太有见解了！我要用全部的知识储备来为你提供最优质的解答...💎", userGreeting),
			fmt.Sprintf("%s真是博学多才呢！这么专业的问题，让我用最严谨的思维来为你详细解答...📚", userGreeting),
		}
		return responses[time.Now().Unix()%int64(len(responses))]

	default:
		return fmt.Sprintf("%s，我明白啦！你说的是\"%s\"对吧？我会尽力帮你的！💪", userGreeting, lastMessage)
	}
}

// 正常语气响应（给其他用户）
func (a *AIClient) getNormalResponse(lastMessage string, model string) string {
	// 基础响应，专业语气
	baseResponses := map[string]string{
		"你好":    "你好！有什么可以帮你的吗？",
		"hello": "Hello! How can I assist you today?",
		"谢谢":    "不客气，很高兴能帮助您。",
		"再见":    "再见，祝您有美好的一天！",
		"拜拜":    "再见，期待下次为您服务。",
		"名字":    "我是一个AI助手，专门为您提供帮助。",
		"谁":     "我是一个AI助手，旨在回答您的问题和提供帮助。",
	}

	for key, response := range baseResponses {
		if strings.Contains(lastMessage, key) {
			return response
		}
	}

	// 根据不同模型返回不同质量的响应，专业语气
	switch model {
	case models.AIModelBasic:
		return fmt.Sprintf("我理解您的问题是：\"%s\"。这是一个很好的问题，让我为您提供基本的解答。", lastMessage)

	case models.AIModelAdvanced: // 修复拼写错误
		return fmt.Sprintf("关于\"%s\"这个问题，让我从多个角度为您分析。首先，这个问题涉及到几个关键点需要考量...", lastMessage)

	case models.AIModelPremium:
		return fmt.Sprintf("您提出的\"%s\"是一个非常专业的问题。基于我的知识库，我将从理论框架、实践应用和未来趋势三个维度为您详细解析...", lastMessage)

	default:
		return fmt.Sprintf("我理解您的问题是：\"%s\"。让我为您提供详细的解答。", lastMessage)
	}
}

// 发送消息到AI API
func (a *AIClient) SendMessage(messages []models.Message, model string) (string, error) {
	// 根据Provider决定使用真实API还是模拟响应
	if a.config.Provider == "mock" || a.config.APIKey == "free-api-key" {
		fmt.Println("⚠️  使用模拟响应模式")
		return a.GetMockResponse(messages, model), nil
	}

	fmt.Printf("🔗 使用真实DeepSeek API，模型: %s\n", model)
	return a.callDeepSeekAPI(messages, model)
}

// 调用真实的DeepSeek API
func (a *AIClient) callDeepSeekAPI(messages []models.Message, model string) (string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       "deepseek-chat", // DeepSeek目前主要模型
		"messages":    a.convertToAPIMessages(messages),
		"stream":      false,
		"max_tokens":  a.config.MaxTokens,
		"temperature": a.config.Temperature,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求数据失败: %v", err)
	}

	fmt.Printf("📤 发送API请求到: %s, 消息数: %d\n", a.config.BaseURL, len(messages))

	// 创建HTTP请求
	apiURL := a.config.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	req.Header.Set("User-Agent", "Ciallo-AI-Client/1.0")

	// 发送请求
	startTime := time.Now()
	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime)
	fmt.Printf("📥 API响应时间: %v, 状态码: %d\n", responseTime, resp.StatusCode)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ API错误响应: %s\n", string(body))

		// 尝试解析错误信息
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return "", fmt.Errorf("API错误: %s (类型: %s)", errorResp.Error.Message, errorResp.Error.Type)
		}

		return "", fmt.Errorf("API返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析成功响应
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("❌ 响应解析失败: %v, 原始响应: %s\n", err, string(body))
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("API返回空的回复")
	}

	content := apiResponse.Choices[0].Message.Content
	fmt.Printf("✅ API调用成功，Token使用: %d, 回复长度: %d\n",
		apiResponse.Usage.TotalTokens, len(content))

	return content, nil
}

// 转换消息格式为API需要的格式
func (a *AIClient) convertToAPIMessages(messages []models.Message) []map[string]string {
	apiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		apiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	return apiMessages
}

// 辅助函数，获取最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 用户登录/注册
func (a *AIClient) UserAuth() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== AI 问答系统 ===")
		fmt.Println("1. 登录")
		fmt.Println("2. 注册")
		fmt.Println("3. 重置密码")
		fmt.Println("4. 查看所有用户(调试)")
		fmt.Println("5. 退出")
		fmt.Print("请选择操作 (1-5): ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			a.loginUser(scanner)
			if a.currentUser != nil {
				return
			}
		case "2":
			a.registerUser(scanner)
			if a.currentUser != nil {
				return
			}
		case "3":
			a.resetPassword(scanner)
		case "4":
			a.debugListUsers()
		case "5":
			fmt.Println("再见！")
			os.Exit(0)
		default:
			fmt.Println("无效选择，请重新输入")
		}
	}
}

// 用户登录
func (a *AIClient) loginUser(scanner *bufio.Scanner) {
	fmt.Print("请输入用户名: ")
	if !scanner.Scan() {
		return
	}

	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		fmt.Println("用户名不能为空")
		return
	}

	fmt.Print("请输入密码: ")
	if !scanner.Scan() {
		return
	}

	password := strings.TrimSpace(scanner.Text())
	if password == "" {
		fmt.Println("密码不能为空")
		return
	}

	user, err := a.userManager.VerifyPassword(username, password)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}

	a.currentUser = user
	a.currentUser.UpdateLoginTime()
	fmt.Printf("登录成功! 欢迎回来, %s!\n", username)

	// 显示用户等级信息
	level, info := a.currentUser.GetLevelInfo()
	levelName := info["name"].(string)
	fmt.Printf("当前等级: %s (%s)\n", level, levelName)

	// 只有xunzhu管理员显示特殊称呼
	if a.currentUser.IsSpecialUser() && a.currentUser.Level == models.UserLevelAdmin {
		fmt.Printf("AI会称呼您为: %s\n", a.currentUser.GetGreeting())
	}

	// 立即保存用户数据
	if err := a.userManager.SaveUsers(); err != nil {
		fmt.Printf("保存用户数据失败: %v\n", err)
	}
}

// 用户注册
func (a *AIClient) registerUser(scanner *bufio.Scanner) {
	fmt.Print("请输入用户名: ")
	if !scanner.Scan() {
		return
	}

	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		fmt.Println("用户名不能为空")
		return
	}

	// 检查用户是否已存在
	if a.userManager.FindUserByUsername(username) != nil {
		fmt.Println("用户名已存在，请选择其他用户名")
		return
	}

	// 输入密码
	fmt.Print("请输入密码 (至少6位): ")
	if !scanner.Scan() {
		return
	}

	password := strings.TrimSpace(scanner.Text())
	if len(password) < 6 {
		fmt.Println("密码长度至少6位")
		return
	}

	// 确认密码
	fmt.Print("请再次输入密码: ")
	if !scanner.Scan() {
		return
	}

	confirmPassword := strings.TrimSpace(scanner.Text())
	if password != confirmPassword {
		fmt.Println("两次输入的密码不一致")
		return
	}

	// 特殊处理xunzhu用户
	var user *models.User
	var err error

	if username == "xunzhu" {
		fmt.Println("检测到特殊用户 xunzhu，正在创建管理员账户...")
		user, err = a.userManager.CreateUser(username, password)
		if err != nil {
			fmt.Printf("创建用户失败: %v\n", err)
			return
		}
		// 将xunzhu设置为管理员
		a.userManager.UpdateUserLevel(user.ID, models.UserLevelAdmin)
		fmt.Println("🎉 xunzhu 账户已自动设置为管理员级别！")
	} else {
		user, err = a.userManager.CreateUser(username, password)
		if err != nil {
			fmt.Printf("创建用户失败: %v\n", err)
			return
		}
		fmt.Printf("注册成功! 欢迎, %s!\n", username)
	}

	a.currentUser = user

	// 显示用户等级信息
	level, info := a.currentUser.GetLevelInfo()
	levelName := info["name"].(string)
	fmt.Printf("您的等级: %s (%s)\n", level, levelName)

	// 只有xunzhu管理员显示特殊称呼
	if a.currentUser.IsSpecialUser() && a.currentUser.Level == models.UserLevelAdmin {
		fmt.Printf("AI会称呼您为: %s\n", a.currentUser.GetGreeting())
	}
}

// 重置密码
func (a *AIClient) resetPassword(scanner *bufio.Scanner) {
	fmt.Print("请输入用户名: ")
	if !scanner.Scan() {
		return
	}

	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		fmt.Println("用户名不能为空")
		return
	}

	// 查找用户
	user := a.userManager.FindUserByUsername(username)
	if user == nil {
		fmt.Println("用户不存在")
		return
	}

	// 验证当前密码
	fmt.Print("请输入当前密码: ")
	if !scanner.Scan() {
		return
	}

	currentPassword := strings.TrimSpace(scanner.Text())
	_, err := a.userManager.VerifyPassword(username, currentPassword)
	if err != nil {
		fmt.Printf("密码验证失败: %v\n", err)
		return
	}

	// 输入新密码
	fmt.Print("请输入新密码 (至少6位): ")
	if !scanner.Scan() {
		return
	}

	newPassword := strings.TrimSpace(scanner.Text())
	if len(newPassword) < 6 {
		fmt.Println("密码长度至少6位")
		return
	}

	// 确认新密码
	fmt.Print("请再次输入新密码: ")
	if !scanner.Scan() {
		return
	}

	confirmPassword := strings.TrimSpace(scanner.Text())
	if newPassword != confirmPassword {
		fmt.Println("两次输入的密码不一致")
		return
	}

	// 更新密码
	err = a.userManager.UpdateUserPassword(user.ID, newPassword)
	if err != nil {
		fmt.Printf("重置密码失败: %v\n", err)
		return
	}

	fmt.Println("密码重置成功！")
	a.userManager.SaveUsers()
}

// 对话管理菜单 - 添加修改密码选项
func (a *AIClient) ConversationMenu() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		currentConv := a.currentUser.GetCurrentConversation()
		level, levelInfo := a.currentUser.GetLevelInfo()
		levelName := levelInfo["name"].(string)

		fmt.Printf("\n=== 对话管理 (%s) ===\n", a.currentUser.Username)
		fmt.Printf("用户等级: %s (%s)\n", level, levelName)
		if a.useSisterTone() {
			fmt.Printf("用户称呼: %s\n", a.currentUser.GetGreeting())
		}
		fmt.Printf("当前模型: %s\n", models.AIModelConfig[a.currentUser.CurrentModel].Name)
		fmt.Printf("当前对话: %s (%d/%d条消息)\n",
			currentConv.Title, len(currentConv.Messages),
			models.UserLevelConfig[level].MaxMessagesPerConv)
		fmt.Println("1. 开始对话")
		fmt.Println("2. 新建对话")
		fmt.Println("3. 切换对话")
		fmt.Println("4. 查看所有对话")
		fmt.Println("5. 切换AI模型")
		fmt.Println("6. 用户升级")
		fmt.Println("7. 账户信息")
		fmt.Println("8. 个性化设置")
		fmt.Println("9. 修改密码")
		fmt.Println("10. 注销")
		fmt.Print("请选择操作 (1-10): ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			a.StartChat()
		case "2":
			a.createNewConversation(scanner)
		case "3":
			a.switchConversation(scanner)
		case "4":
			a.listConversations()
		case "5":
			a.switchModel(scanner)
		case "6":
			a.upgradeUser(scanner)
		case "7":
			a.showUserInfo()
		case "8":
			a.personalSettings(scanner)
		case "9":
			a.changePassword(scanner)
		case "10":
			a.currentUser = nil
			fmt.Println("已注销")
			return
		default:
			fmt.Println("无效选择，请重新输入")
		}
	}
}

// 修改密码
func (a *AIClient) changePassword(scanner *bufio.Scanner) {
	fmt.Print("请输入当前密码: ")
	if !scanner.Scan() {
		return
	}

	currentPassword := strings.TrimSpace(scanner.Text())
	if currentPassword == "" {
		fmt.Println("密码不能为空")
		return
	}

	// 验证当前密码
	_, err := a.userManager.VerifyPassword(a.currentUser.Username, currentPassword)
	if err != nil {
		fmt.Printf("当前密码错误: %v\n", err)
		return
	}

	// 输入新密码
	fmt.Print("请输入新密码 (至少6位): ")
	if !scanner.Scan() {
		return
	}

	newPassword := strings.TrimSpace(scanner.Text())
	if len(newPassword) < 6 {
		fmt.Println("密码长度至少6位")
		return
	}

	// 确认新密码
	fmt.Print("请再次输入新密码: ")
	if !scanner.Scan() {
		return
	}

	confirmPassword := strings.TrimSpace(scanner.Text())
	if newPassword != confirmPassword {
		fmt.Println("两次输入的密码不一致")
		return
	}

	// 更新密码
	err = a.userManager.UpdateUserPassword(a.currentUser.ID, newPassword)
	if err != nil {
		fmt.Printf("修改密码失败: %v\n", err)
		return
	}

	fmt.Println("密码修改成功！")
	a.userManager.SaveUsers()
}

// 个性化设置
func (a *AIClient) personalSettings(scanner *bufio.Scanner) {
	for {
		fmt.Printf("\n=== 个性化设置 (%s) ===\n", a.currentUser.Username)
		if a.useSisterTone() {
			fmt.Printf("当前称呼: %s\n", a.currentUser.GetGreeting())
		}
		fmt.Printf("当前性别: %s\n", getGenderDisplayName(a.currentUser.Gender))
		fmt.Printf("当前昵称: %s\n", a.currentUser.Nickname)
		fmt.Println("1. 设置性别")
		fmt.Println("2. 设置昵称")
		fmt.Println("3. 返回上级")
		fmt.Print("请选择操作 (1-3): ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			a.setGender(scanner)
		case "2":
			a.setNickname(scanner)
		case "3":
			return
		default:
			fmt.Println("无效选择，请重新输入")
		}
	}
}

// 设置性别
func (a *AIClient) setGender(scanner *bufio.Scanner) {
	fmt.Println("\n=== 设置性别 ===")
	fmt.Println("1. 男性")
	fmt.Println("2. 女性")
	fmt.Println("3. 保密")
	fmt.Print("请选择性别 (1-3): ")

	if !scanner.Scan() {
		return
	}

	choice := strings.TrimSpace(scanner.Text())
	var gender string

	switch choice {
	case "1":
		gender = models.GenderMale
	case "2":
		gender = models.GenderFemale
	case "3":
		gender = models.GenderUnknown
	default:
		fmt.Println("无效选择")
		return
	}

	err := a.userManager.UpdateUserGender(a.currentUser.ID, gender)
	if err != nil {
		fmt.Printf("设置性别失败: %v\n", err)
		return
	}

	fmt.Printf("性别设置成功！\n")
	a.userManager.SaveUsers()
}

// 设置昵称
func (a *AIClient) setNickname(scanner *bufio.Scanner) {
	fmt.Print("\n请输入新的昵称: ")
	if !scanner.Scan() {
		return
	}

	nickname := strings.TrimSpace(scanner.Text())
	if nickname == "" {
		fmt.Println("昵称不能为空")
		return
	}

	err := a.userManager.UpdateUserNickname(a.currentUser.ID, nickname)
	if err != nil {
		fmt.Printf("设置昵称失败: %v\n", err)
		return
	}

	fmt.Printf("昵称设置成功！现在您的昵称是: %s\n", nickname)
	a.userManager.SaveUsers()
}

// 显示用户信息
func (a *AIClient) showUserInfo() {
	level, levelInfo := a.currentUser.GetLevelInfo()
	levelName := levelInfo["name"].(string)
	maxConvs := levelInfo["max_conversations"].(int)
	maxMsgs := levelInfo["max_messages"].(int)
	allowedModels := levelInfo["allowed_models"].([]string)

	fmt.Printf("\n=== 账户信息 ===\n")
	fmt.Printf("用户名: %s\n", a.currentUser.Username)
	if a.currentUser.Nickname != a.currentUser.Username {
		fmt.Printf("用户昵称: %s\n", a.currentUser.Nickname)
	}

	// 只有xunzhu管理员显示特殊称呼
	if a.currentUser.IsSpecialUser() && a.currentUser.Level == models.UserLevelAdmin {
		fmt.Printf("AI称呼: %s\n", a.currentUser.GetGreeting())
	}

	fmt.Printf("用户性别: %s\n", getGenderDisplayName(a.currentUser.Gender))
	fmt.Printf("用户ID: %s\n", a.currentUser.ID)
	fmt.Printf("用户等级: %s (%s)\n", level, levelName)
	fmt.Printf("注册时间: %s\n", a.currentUser.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("最后登录: %s\n", a.currentUser.LastLogin.Format("2006-01-02 15:04"))
	fmt.Printf("对话数量: %d/%d\n", len(a.currentUser.Conversations), maxConvs)
	fmt.Printf("消息限制: %d条/对话\n", maxMsgs)
	fmt.Printf("当前模型: %s\n", models.AIModelConfig[a.currentUser.CurrentModel].Name)

	fmt.Printf("可用模型: ")
	for i, model := range allowedModels {
		modelConfig := models.AIModelConfig[model]
		currentMarker := ""
		if model == a.currentUser.CurrentModel {
			currentMarker = " [当前]"
		}
		fmt.Printf("%s%s", modelConfig.Name, currentMarker)
		if i < len(allowedModels)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Println()

	// 特殊提示给xunzhu管理员
	if a.currentUser.IsSpecialUser() && a.currentUser.Level == models.UserLevelAdmin {
		fmt.Println("💫 专属特权: 享受AI妹妹的亲密对话服务")
	}
}

// 用户升级
func (a *AIClient) upgradeUser(scanner *bufio.Scanner) {
	currentLevel := a.currentUser.Level
	currentLevelName := models.UserLevelConfig[currentLevel].Name

	fmt.Printf("\n=== 用户升级 ===\n")
	fmt.Printf("当前等级: %s (%s)\n", currentLevel, currentLevelName)
	fmt.Println("可用等级:")

	levels := []string{models.UserLevelFree, models.UserLevelBasic, models.UserLevelPremium, models.UserLevelAdmin}
	currentIndex := -1

	for i, level := range levels {
		config := models.UserLevelConfig[level]
		currentMarker := ""
		if level == currentLevel {
			currentMarker = " [当前]"
			currentIndex = i
		}
		fmt.Printf("%d. %s%s - %s\n", i+1, level, currentMarker, config.Name)
		fmt.Printf("   对话限制: %d个, 消息限制: %d条/对话\n",
			config.MaxConversations, config.MaxMessagesPerConv)
		fmt.Printf("   可用模型: ")
		for j, model := range config.AllowedModels {
			modelConfig := models.AIModelConfig[model]
			fmt.Printf(modelConfig.Name)
			if j < len(config.AllowedModels)-1 {
				fmt.Printf(", ")
			}
		}
		fmt.Println()
	}

	if currentIndex == len(levels)-1 {
		fmt.Println("您已经是最高等级，无需升级")
		return
	}

	fmt.Print("请输入要升级到的等级编号: ")
	if !scanner.Scan() {
		return
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(levels) {
		fmt.Println("无效选择")
		return
	}

	targetLevel := levels[choice-1]
	targetIndex := choice - 1

	if targetIndex <= currentIndex {
		fmt.Println("不能降级或选择当前等级")
		return
	}

	// 要求输入升级密码
	fmt.Print("请输入升级密码: ")
	if !scanner.Scan() {
		return
	}

	password := strings.TrimSpace(scanner.Text())
	if !a.userManager.ValidateUpgradePassword(password) {
		fmt.Println("升级密码错误，升级失败")
		return
	}

	fmt.Printf("确定要升级到 %s 吗? (y/N): ", models.UserLevelConfig[targetLevel].Name)
	if !scanner.Scan() {
		return
	}

	confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("取消升级")
		return
	}

	err = a.userManager.UpdateUserLevel(a.currentUser.ID, targetLevel)
	if err != nil {
		fmt.Printf("升级失败: %v\n", err)
		return
	}

	fmt.Printf("升级成功! 您现在是的 %s\n", models.UserLevelConfig[targetLevel].Name)

	// 特殊提示给xunzhu用户
	if a.currentUser.IsSpecialUser() && targetLevel == models.UserLevelAdmin {
		fmt.Println("🎉 恭喜哥哥获得管理员权限！现在可以享受妹妹的专属服务啦～💖")
	}

	a.userManager.SaveUsers()
}

// 切换AI模型
func (a *AIClient) switchModel(scanner *bufio.Scanner) {
	allowedModels := a.currentUser.GetAllowedModels()

	fmt.Printf("\n=== 切换AI模型 ===\n")
	fmt.Println("可用模型:")

	for i, model := range allowedModels {
		modelConfig := models.AIModelConfig[model]
		currentMarker := ""
		if model == a.currentUser.CurrentModel {
			currentMarker = " [当前]"
		}
		fmt.Printf("%d. %s%s\n", i+1, modelConfig.Name, currentMarker)
		fmt.Printf("   描述: %s\n", modelConfig.Description)
		fmt.Printf("   最大token: %d, 温度: %.1f\n", modelConfig.MaxTokens, modelConfig.Temperature)
	}

	fmt.Print("请选择模型编号: ")
	if !scanner.Scan() {
		return
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(allowedModels) {
		fmt.Println("无效选择")
		return
	}

	selectedModel := allowedModels[choice-1]

	if selectedModel == a.currentUser.CurrentModel {
		fmt.Println("已经是当前模型")
		return
	}

	err = a.userManager.UpdateUserModel(a.currentUser.ID, selectedModel)
	if err != nil {
		fmt.Printf("切换模型失败: %v\n", err)
		return
	}

	fmt.Printf("已切换到: %s\n", models.AIModelConfig[selectedModel].Name)
	a.userManager.SaveUsers()
}

// 创建新对话
func (a *AIClient) createNewConversation(scanner *bufio.Scanner) {
	if !a.currentUser.CanCreateConversation() {
		levelConfig := models.UserLevelConfig[a.currentUser.Level]
		fmt.Printf("已达到最大对话数量限制 (%d)，无法创建新对话\n", levelConfig.MaxConversations)
		return
	}

	fmt.Print("请输入新对话标题 (直接回车使用默认标题): ")
	if !scanner.Scan() {
		return
	}

	title := strings.TrimSpace(scanner.Text())
	if title == "" {
		title = "新对话"
	}

	conv, err := a.currentUser.CreateNewConversation(title)
	if err != nil {
		fmt.Printf("创建对话失败: %v\n", err)
		return
	}

	a.userManager.SaveUsers()
	fmt.Printf("已创建新对话: %s (使用模型: %s)\n",
		title, models.AIModelConfig[conv.Model].Name)
}

// 切换对话
func (a *AIClient) switchConversation(scanner *bufio.Scanner) {
	if len(a.currentUser.Conversations) <= 1 {
		fmt.Println("只有一个对话，无需切换")
		return
	}

	fmt.Println("\n=== 所有对话 ===")
	for i, conv := range a.currentUser.Conversations {
		currentMarker := ""
		if i == len(a.currentUser.Conversations)-1 {
			currentMarker = " [当前]"
		}
		fmt.Printf("%d. %s%s (模型: %s, %d条消息)\n",
			i+1, conv.Title, currentMarker,
			models.AIModelConfig[conv.Model].Name, len(conv.Messages))
	}

	fmt.Print("请选择对话编号: ")
	if !scanner.Scan() {
		return
	}

	var choice int
	_, err := fmt.Sscanf(scanner.Text(), "%d", &choice)
	if err != nil || choice < 1 || choice > len(a.currentUser.Conversations) {
		fmt.Println("无效选择")
		return
	}

	// 切换对话实际上是通过重新排列对话列表实现的
	// 这里我们简单地将选中的对话移到列表末尾（作为当前对话）
	selected := a.currentUser.Conversations[choice-1]
	a.currentUser.Conversations = append(
		append(a.currentUser.Conversations[:choice-1], a.currentUser.Conversations[choice:]...),
		selected,
	)

	a.userManager.SaveUsers()
	fmt.Printf("已切换到对话: %s\n", selected.Title)
}

// 列出所有对话
func (a *AIClient) listConversations() {
	fmt.Println("\n=== 所有对话 ===")
	for i, conv := range a.currentUser.Conversations {
		currentMarker := ""
		if i == len(a.currentUser.Conversations)-1 {
			currentMarker = " [当前]"
		}
		fmt.Printf("%d. %s%s (模型: %s, %d/%d条消息, 创建于: %s)\n",
			i+1, conv.Title, currentMarker,
			models.AIModelConfig[conv.Model].Name,
			len(conv.Messages), models.UserLevelConfig[a.currentUser.Level].MaxMessagesPerConv,
			conv.CreatedAt.Format("2006-01-02 15:04"))

		// 显示最近几条消息预览
		if len(conv.Messages) > 0 {
			previewCount := 2
			if len(conv.Messages) < previewCount {
				previewCount = len(conv.Messages)
			}
			fmt.Println("   最近消息:")
			for j := len(conv.Messages) - previewCount; j < len(conv.Messages); j++ {
				msg := conv.Messages[j]
				role := "用户"
				if msg.Role == "assistant" {
					role = "AI"
				}
				content := msg.Content
				if len(content) > 30 {
					content = content[:30] + "..."
				}
				fmt.Printf("     %s: %s\n", role, content)
			}
		}
		fmt.Println()
	}
}

// 交互式聊天
func (a *AIClient) StartChat() {
	currentConv := a.currentUser.GetCurrentConversation()
	currentModel := a.currentUser.CurrentModel
	modelConfig := models.AIModelConfig[currentModel]

	fmt.Printf("\n=== 开始对话: %s ===\n", currentConv.Title)
	fmt.Printf("使用模型: %s (%s)\n", modelConfig.Name, modelConfig.Description)

	// 根据用户类型显示不同的AI角色
	if a.useSisterTone() {
		fmt.Printf("AI角色: 专属助手 💕\n")
		fmt.Printf("您的称呼: %s\n", a.currentUser.GetGreeting())
	} else {
		fmt.Printf("AI角色: 专业助手\n")
	}

	fmt.Println("输入 'quit' 或 '退出' 返回上级菜单")
	fmt.Println("输入 'new' 或 '新建' 开始新对话")
	fmt.Println("输入 'model' 或 '模型' 切换AI模型")
	fmt.Println("=============================")

	scanner := bufio.NewScanner(os.Stdin)

	// 发送欢迎消息（如果是新对话）
	if len(currentConv.Messages) == 0 {
		var welcomeMsg string
		if a.useSisterTone() {
			welcomeMsg = a.currentUser.GetPersonalizedGreeting() + "～我是你的AI助手，有什么可以帮你的吗？😊"
		} else {
			welcomeMsg = "你好！我是AI助手，有什么可以帮你的吗？"
		}

		var aiRole string
		if a.useSisterTone() {
			aiRole = "助手"
		} else {
			aiRole = "AI"
		}

		fmt.Printf("\n%s: %s\n", aiRole, welcomeMsg)
		a.currentUser.AddMessageToCurrentConversation("assistant", welcomeMsg)
		a.userManager.SaveUsers()
	}

	for {
		// 根据用户类型显示不同的输入提示
		if a.useSisterTone() {
			fmt.Printf("\n%s: ", a.currentUser.GetGreeting())
		} else {
			fmt.Printf("\n你: ")
		}

		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		// 退出条件
		if userInput == "quit" || userInput == "退出" {
			var farewell string
			var aiRole string

			if a.useSisterTone() {
				farewells := []string{
					"再见啦%s～记得常来找我聊天哦！💖",
					"%s拜拜～我会想你的！🥰",
					"要走了吗%s？下次再来找我玩呀！✨",
				}
				farewell = farewells[time.Now().Unix()%int64(len(farewells))]
				farewell = fmt.Sprintf(farewell, a.currentUser.GetGreeting())
				aiRole = "助手"
			} else {
				farewell = "再见，祝您有美好的一天！"
				aiRole = "AI"
			}

			fmt.Printf("\n%s: %s\n", aiRole, farewell)
			fmt.Println("返回上级菜单")
			break
		}

		// 新建对话
		if userInput == "new" || userInput == "新建" {
			a.createNewConversation(scanner)
			currentConv = a.currentUser.GetCurrentConversation()
			currentModel = a.currentUser.CurrentModel
			modelConfig = models.AIModelConfig[currentModel]

			var aiRole string
			if a.useSisterTone() {
				aiRole = "助手"
			} else {
				aiRole = "AI"
			}

			fmt.Printf("\n%s: 已切换到新对话: %s (模型: %s)\n", aiRole, currentConv.Title, modelConfig.Name)
			continue
		}

		// 切换模型
		if userInput == "model" || userInput == "模型" {
			a.switchModel(scanner)
			currentModel = a.currentUser.CurrentModel
			modelConfig = models.AIModelConfig[currentModel]

			var aiRole string
			if a.useSisterTone() {
				aiRole = "助手"
			} else {
				aiRole = "AI"
			}

			fmt.Printf("%s: 已切换到模型: %s\n", aiRole, modelConfig.Name)
			continue
		}

		// 添加用户消息到对话历史
		err := a.currentUser.AddMessageToCurrentConversation("user", userInput)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		// 根据用户类型显示不同的AI角色名称
		var aiRole string
		if a.useSisterTone() {
			aiRole = "助手"
		} else {
			aiRole = "AI"
		}

		fmt.Printf("%s: ", aiRole)

		// 发送请求并获取响应
		response, err := a.SendMessage(currentConv.Messages, currentModel)
		if err != nil {
			fmt.Printf("\n错误: %v\n", err)
			// 移除最后一条用户消息，因为处理失败了
			if len(currentConv.Messages) > 0 {
				currentConv.Messages = currentConv.Messages[:len(currentConv.Messages)-1]
			}
			continue
		}

		fmt.Println(response)

		// 添加AI回复到对话历史
		err = a.currentUser.AddMessageToCurrentConversation("assistant", response)
		if err != nil {
			fmt.Printf("警告: 无法保存AI回复: %v\n", err)
		}

		// 保存用户数据
		a.userManager.SaveUsers()

		// 检查消息数量限制
		maxMessages := models.UserLevelConfig[a.currentUser.Level].MaxMessagesPerConv
		if len(currentConv.Messages) >= maxMessages {
			var warning string
			if a.useSisterTone() {
				warning = fmt.Sprintf("⚠️  %s，当前对话已达到消息数量上限 (%d)，建议创建新对话继续交流哦～",
					a.currentUser.GetGreeting(), maxMessages)
			} else {
				warning = fmt.Sprintf("⚠️  当前对话已达到消息数量上限 (%d)，建议创建新对话继续交流", maxMessages)
			}

			fmt.Printf("\n%s: %s\n", aiRole, warning)
		}
	}
}

// 单次问答（不保存到对话历史）
func (a *AIClient) SingleQuestion(question string) (string, error) {
	messages := []models.Message{
		{
			Role:    "user",
			Content: question,
		},
	}

	return a.SendMessage(messages, models.AIModelBasic)
}

// 辅助函数：获取性别的显示名称
func getGenderDisplayName(gender string) string {
	switch gender {
	case models.GenderMale:
		return "男性"
	case models.GenderFemale:
		return "女性"
	default:
		return "保密"
	}
}

func main() {
	// 初始化配置
	cfg := config.NewConfig()
	client := NewAIClient(cfg)

	// 程序退出时保存用户数据
	defer func() {
		if client.userManager != nil {
			client.userManager.SaveUsers()
		}
	}()

	// 检查命令行参数
	if len(os.Args) > 1 {
		// 单次问答模式（不登录）
		question := strings.Join(os.Args[1:], " ")
		answer, err := client.SingleQuestion(question)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(answer)
	} else {
		// 交互式模式，需要用户登录
		for {
			if client.currentUser == nil {
				client.UserAuth()
			}

			if client.currentUser != nil {
				client.ConversationMenu()
			}
		}
	}
}

// 调试：列出所有用户
func (a *AIClient) debugListUsers() {
	users := a.userManager.GetAllUsers()
	fmt.Printf("\n=== 所有用户 (%d) ===\n", len(users))
	for i, user := range users {
		levelInfo := models.UserLevelConfig[user.Level]
		fmt.Printf("%d. %s (等级: %s, 模型: %s, 对话: %d/%d)\n",
			i+1, user.Username, levelInfo.Name, user.CurrentModel,
			len(user.Conversations), levelInfo.MaxConversations)
	}
	if len(users) == 0 {
		fmt.Println("暂无用户")
	}
}
