package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"webtest/database"
	"webtest/utils"

	"github.com/gin-gonic/gin"
	ark "github.com/sashabaranov/go-openai"
)

type ModelConfig struct {
	ApiKey  string
	BaseURL string
	Model   string
	context []ark.ChatCompletionMessage
}

func AIChatRequest(c *gin.Context) {
	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取token，优先从Authorization头部获取，其次从查询参数获取，兼容不支持自定义请求头的情况
	token := ""
	authHeader := c.GetHeader("Authorization")

	if authHeader != "" {
		// 检查Bearer前缀
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	// 如果头部没有token，则尝试从查询参数获取
	if token == "" {
		token = c.Query("token")
	}

	// 如果仍然没有token，则返回错误
	if token == "" {
		c.JSON(401, gin.H{"error": "Authorization header or token parameter is required"})
		return
	}

	// 验证token并获取用户名
	username, err := utils.VerifyToken(token)
	fmt.Println("username:", username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	Permission, _ := GetUserPermission(username)
	// 获取参数
	// 提交的文本
	content := c.Query("content")
	// 对话ID
	cid := c.Query("cid")
	// 模型名称
	modelName := c.Query("MODELNAME")

	// 从配置文件加载模型配置
	modelsConfig, err := utils.LoadModelsConfig("models_config.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load models configuration"})
		return
	}

	// 获取指定模型的配置
	modelConfig, exists := modelsConfig[modelName]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model not found: " + modelName})
		return
	}

	// 检查用户权限是否足够使用该模型
	if modelConfig.Permission > Permission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to use model: " + modelName})
		return
	}
	//保存记录
	err = database.SaveChatRecord(username, modelConfig.ModelFullName, "user", time.Now().UnixMilli(), content, cid)
	if err != nil {
		fmt.Println("Failed to save chat record:", err)
	}

	chatHistory, err := database.GetChatHistory(username, cid, 10)
	var ChatContext []ark.ChatCompletionMessage
	for _, record := range chatHistory {
		ChatContext = append(ChatContext, ark.ChatCompletionMessage{
			Role:    record.Role,
			Content: record.Content,
		})
	}
	// 创建通道传递生成结果
	dataChan := make(chan any)
	Config := ModelConfig{
		ApiKey:  modelConfig.APIKeyStore,
		BaseURL: modelConfig.BaseURL,
		Model:   modelConfig.ModelFullName,
		context: ChatContext,
	}

	//使用OPENAI SDK
	go ChatRequest(content, Config, dataChan, username, cid)

	// 使用 c.SSEvent 实现 SSE 流式传输，符合 OpenAI API 格式
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-dataChan; ok {
			c.SSEvent("message", msg)
			c.Writer.Flush()
			return true
		}
		c.SSEvent("message", "[DONE]")
		c.Writer.Flush()
		return false
	})
}

func ChatRequest(prompt string, Config ModelConfig, ch chan<- any, username string, cid string) {
	config := ark.DefaultConfig(os.Getenv(Config.ApiKey))
	config.BaseURL = Config.BaseURL
	client := ark.NewClientWithConfig(config)
	//通道关闭
	defer close(ch)

	// 添加当前用户消息到上下文
	Config.context = append(Config.context, ark.ChatCompletionMessage{
		Role:    "user",
		Content: prompt,
	})

	// 流式返回
	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		ark.ChatCompletionRequest{
			Model:    Config.Model,
			Messages: Config.context,
		},
	)
	if err != nil {
		fmt.Printf("stream chat error: %v\n", err)
		// 发送错误信息给客户端
		//ch <- fmt.Sprintf("{\"error\": \"%v\"}", err)
		return
	}
	defer func(stream *ark.ChatCompletionStream) {
		_ = stream.Close()
	}(stream)
	Result := ""
	for {
		rec, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Stream chat error: %v\n", err)
			// 发送错误信息给客户端
			ch <- fmt.Sprintf("{\"error\": \"%v\"}", err)
			break
		}

		if len(rec.Choices) > 0 && rec.Choices[0].Delta.Content != "" {
			Result += rec.Choices[0].Delta.Content
			ch <- rec
		}
	}
	err = database.SaveChatRecord(username, Config.Model, "assistant", time.Now().UnixMilli(), Result, cid)
	if err != nil {
		return
	}
}
