package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// 调试：列出所有用户
func (a *AIClient) DebugListUsers() {
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

// 发送消息到AI API
func (a *AIClient) SendMessage(messages []models.Message, model string) (string, error) {
	// 根据Provider决定使用真实API还是模拟响应
	if a.config.Provider == "mock" || a.config.APIKey == "free-api-key" {
		fmt.Println("⚠️  Web版本: 使用模拟响应模式")
		return a.GetMockResponse(messages, model), nil
	}

	fmt.Printf("🔗 Web版本: 使用真实DeepSeek API，模型: %s\n", model)
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

	fmt.Printf("📤 Web版本: 发送API请求到: %s, 消息数: %d\n", a.config.BaseURL, len(messages))

	// 创建HTTP请求
	apiURL := a.config.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	req.Header.Set("User-Agent", "Ciallo-Web-Client/1.0")

	// 发送请求
	startTime := time.Now()
	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime)
	fmt.Printf("📥 Web版本: API响应时间: %v, 状态码: %d\n", responseTime, resp.StatusCode)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Web版本: API错误响应: %s\n", string(body))

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
		fmt.Printf("❌ Web版本: 响应解析失败: %v, 原始响应: %s\n", err, string(body))
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("API返回空的回复")
	}

	content := apiResponse.Choices[0].Message.Content
	fmt.Printf("✅ Web版本: API调用成功，Token使用: %d, 回复长度: %d\n",
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

// 模拟响应 - 根据用户类型返回不同的响应
func (a *AIClient) GetMockResponse(messages []models.Message, model string) string {
	if len(messages) == 0 {
		// 首次问候
		if a.UseSisterTone() {
			greeting := a.currentUser.GetPersonalizedGreeting()
			return fmt.Sprintf("%s～我是你的AI助手，有什么可以帮你的吗？", greeting)
		} else {
			return "你好！我是AI助手，有什么可以帮你的吗？"
		}
	}

	lastMessage := messages[len(messages)-1].Content

	// 检查是否使用妹妹语气
	if a.UseSisterTone() {
		return a.getSisterResponse(lastMessage, model)
	} else {
		return a.getNormalResponse(lastMessage, model)
	}
}

// 使用妹妹语气的条件
func (a *AIClient) UseSisterTone() bool {
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

	case models.AIModelAdvanced:
		return fmt.Sprintf("关于\"%s\"这个问题，让我从多个角度为您分析。首先，这个问题涉及到几个关键点需要考量...", lastMessage)

	case models.AIModelPremium:
		return fmt.Sprintf("您提出的\"%s\"是一个非常专业的问题。基于我的知识库，我将从理论框架、实践应用和未来趋势三个维度为您详细解析...", lastMessage)

	default:
		return fmt.Sprintf("我理解您的问题是：\"%s\"。让我为您提供详细的解答。", lastMessage)
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
func GetGenderDisplayName(gender string) string {
	switch gender {
	case models.GenderMale:
		return "男性"
	case models.GenderFemale:
		return "女性"
	default:
		return "保密"
	}
}
