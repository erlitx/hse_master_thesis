package dto

import (
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/google/uuid"
)

// Запрос диалога
type ConversationRequest struct {
	Message   string  `json:"message" validate:"required"`
	SessionID *string `json:"session_id,omitempty"`
	Model     string  `json:"model,omitempty"`
	MaxTokens int     `json:"max_tokens,omitempty"`
	System    string  `json:"system,omitempty"`
}

// Ответ диалога
type ConversationResponse struct {
	SessionID  string     `json:"session_id"`
	Response   string     `json:"response"`
	StopReason string     `json:"stop_reason"`
	Model      string     `json:"model"`
	Usage      *UsageInfo `json:"usage,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Преобразует запрос в domain.Message
func (r *ConversationRequest) ToUserMessage(sessionID string) domain.Message {
	return domain.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content: []domain.Content{
			{
				Type: "text",
				Text: r.Message,
			},
		},
		CreatedAt:    time.Now(),
		InputTokens:  0,
		OutputTokens: 0,
	}
}

// Формирует ответ диалога из сессии
func FromSession(session *domain.Session) *ConversationResponse {
	response := ""
	var inputTokens, outputTokens int

	if len(session.Messages) > 0 {
		lastMsg := session.Messages[len(session.Messages)-1]
		if lastMsg.Role == "assistant" {
			response = extractTextFromContent(lastMsg.Content)
			inputTokens = lastMsg.InputTokens
			outputTokens = lastMsg.OutputTokens
		}
	}

	return &ConversationResponse{
		SessionID:  session.ID,
		Response:   response,
		StopReason: session.StopReason,
		Model:      session.Model,
		Usage: &UsageInfo{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		},
		CreatedAt: session.UpdatedAt,
	}
}

// Извлекает текст из блоков контента
func extractTextFromContent(content []domain.Content) string {
	if len(content) == 0 {
		return ""
	}

	var result string
	for _, c := range content {
		if c.Type == "text" {
			result += c.Text
		}
	}
	return result
}

// Конвертирует сообщения в ClaudeMessage
func FromDomainMessages(messages []domain.Message) []ClaudeMessage {
	claudeMessages := make([]ClaudeMessage, 0, len(messages))
	for _, msg := range messages {
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    msg.Role,
			Content: fromDomainContent(msg.Content),
		})
	}
	return claudeMessages
}

// Конвертирует ClaudeContent в domain.Content
func ToDomainContent(dtoContent []ClaudeContent) []domain.Content {
	if len(dtoContent) == 0 {
		return nil
	}

	result := make([]domain.Content, 0, len(dtoContent))
	for _, c := range dtoContent {
		input := c.Input
		// Claude may omit empty object in responses, but when replaying tool_use
		// we must preserve it as {} for the next request.
		if c.Type == "tool_use" && input == nil {
			input = map[string]interface{}{}
		}

		domainContent := domain.Content{
			Type:       c.Type,
			Text:       c.Text,
			ID:         c.ID,
			ToolUseID:  c.ToolUseID,
			Name:       c.Name,
			Input:      input,
			IsError:    c.IsError,
			ServerName: c.ServerName,
		}

		// Recursively convert nested content (for tool_result)
		if len(c.Content) > 0 {
			domainContent.Content = ToDomainContent(c.Content)
		}

		result = append(result, domainContent)
	}
	return result
}

// Конвертирует domain.Content для Claude
func fromDomainContent(domainContent []domain.Content) interface{} {
	if len(domainContent) == 0 {
		return ""
	}

	if len(domainContent) == 1 && domainContent[0].Type == "text" {
		return domainContent[0].Text
	}

	return fromDomainContentBlocks(domainContent)
}

// Конвертирует блоки контента для gateway
func fromDomainContentBlocks(domainContent []domain.Content) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(domainContent))
	for _, c := range domainContent {
		entry := map[string]interface{}{
			"type": c.Type,
		}

		switch c.Type {
		case "text":
			entry["text"] = c.Text
		case "tool_use":
			entry["id"] = c.ID
			entry["name"] = c.Name
			if c.Input == nil {
				entry["input"] = map[string]interface{}{}
			} else {
				entry["input"] = c.Input
			}
			if c.ServerName != "" {
				entry["server_name"] = c.ServerName
			}
		case "tool_result":
			entry["tool_use_id"] = c.ToolUseID
			if c.IsError != nil {
				entry["is_error"] = *c.IsError
			}
			if len(c.Content) > 0 {
				entry["content"] = fromDomainContentBlocks(c.Content)
			}
		default:
			if c.Text != "" {
				entry["text"] = c.Text
			}
		}

		result = append(result, entry)
	}
	return result
}
