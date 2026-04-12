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
		ID:           uuid.New().String(),
		SessionID:    sessionID,
		Role:         "user",
		Content:      r.Message,
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
			response = lastMsg.Content
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

// FromDomainMessages converts domain Messages to Claude API messages format
func FromDomainMessages(messages []domain.Message) []ClaudeMessage {
	claudeMessages := make([]ClaudeMessage, 0, len(messages))
	for _, msg := range messages {
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return claudeMessages
}
