package response

type OpenAIStreamChoiceDelta struct {
	Content string `json:"content"`
}

type OpenAIStreamChoice struct {
	Delta OpenAIStreamChoiceDelta `json:"delta"`
}

type OpenAIStreamResponse struct {
	Choices []OpenAIStreamChoice `json:"choices"`
}
