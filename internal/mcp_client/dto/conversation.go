package dto

import (
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/google/uuid"
)

// ConversationRequest represents the HTTP request for conversation endpoint
type ConversationRequest struct {
	Message   string  `json:"message" validate:"required"`
	SessionID *string `json:"session_id,omitempty"`
	Model     string  `json:"model,omitempty"`
	MaxTokens int     `json:"max_tokens,omitempty"`
	System    string  `json:"system,omitempty"`
}

// ConversationResponse represents the HTTP response for conversation endpoint
type ConversationResponse struct {
	SessionID  string     `json:"session_id"`
	Response   string     `json:"response"`
	StopReason string     `json:"stop_reason"`
	Model      string     `json:"model"`
	Usage      *UsageInfo `json:"usage,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToUserMessage converts the request message to a domain Message
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

// FromSession converts a domain Session to ConversationResponse DTO
func FromSession(session *domain.Session) *ConversationResponse {
	// Get the last assistant message as response
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

// extractTextFromContent extracts only text content from Content array
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

// FromDomainMessages converts domain Messages to Claude API messages format
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

// ToDomainContent converts DTO ClaudeContent to domain Content
func ToDomainContent(dtoContent []ClaudeContent) []domain.Content {
	if len(dtoContent) == 0 {
		return nil
	}

	result := make([]domain.Content, 0, len(dtoContent))
	for _, c := range dtoContent {
		domainContent := domain.Content{
			Type:       c.Type,
			Text:       c.Text,
			ID:         c.ID,
			ToolUseID:  c.ToolUseID,
			Name:       c.Name,
			Input:      c.Input,
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

// fromDomainContent converts domain Content to DTO format for Claude API
func fromDomainContent(domainContent []domain.Content) interface{} {
	if len(domainContent) == 0 {
		return ""
	}

	// If there's only one text content, return as string for simplicity
	if len(domainContent) == 1 && domainContent[0].Type == "text" {
		return domainContent[0].Text
	}

	// Otherwise, return as array of ClaudeContent
	result := make([]ClaudeContent, 0, len(domainContent))
	for _, c := range domainContent {
		claudeContent := ClaudeContent{
			Type:       c.Type,
			Text:       c.Text,
			ID:         c.ID,
			ToolUseID:  c.ToolUseID,
			Name:       c.Name,
			Input:      c.Input,
			ServerName: c.ServerName,
		}

		// Recursively convert nested content
		if len(c.Content) > 0 {
			claudeContent.Content = make([]ClaudeContent, 0, len(c.Content))
			for _, nested := range c.Content {
				claudeContent.Content = append(claudeContent.Content, ClaudeContent{
					Type: nested.Type,
					Text: nested.Text,
				})
			}
		}

		result = append(result, claudeContent)
	}
	return result
}
