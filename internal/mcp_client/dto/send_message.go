package dto

// Запрос отправки сообщения
type SendMessageRequest struct {
	Message string `json:"message" validate:"required"`
}

// Ответ отправки сообщения
type SendMessageResponse struct {
	Response      string         `json:"response"`
	ResourcesUsed []ResourceInfo `json:"resources_used"`
	Model         string         `json:"model"`
	Usage         *UsageInfo     `json:"usage,omitempty"`
}

// Метаданные ресурса в ответе
type ResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Статистика токенов
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
