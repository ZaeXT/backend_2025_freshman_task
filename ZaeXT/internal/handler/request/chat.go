package request

type ChatMessage struct {
	Message        string `json:"message" binding:"required,max=5000"`
	ModelID        string `json:"model_id,omitempty"`
	EnableThinking bool   `json:"enable_thinking,omitempty"`
}

type CreateConversation struct {
	IsTemporary bool  `json:"is_temporary"`
	CategoryID  *uint `json:"category_id,omitempty"`
}

type UpdateTitle struct {
	Title string `json:"title" binding:"required,max=255"`
}
