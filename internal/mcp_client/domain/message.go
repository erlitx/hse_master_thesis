package domain

import "time"

// Сессия диалога
type Session struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Messages    []Message
	Tools       []ToolDefinition
	GatewayModel     string
	GatewayMaxTokens int
	GatewaySystem    string
	StopReason       string // "end_turn", "max_tokens", "stop_sequence", etc.
	Model            string
	TotalTokens      int
}

// Описание MCP-инструмента
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// Сообщение в сессии
type Message struct {
	ID           string
	SessionID    string
	Role         string // "user" or "assistant"
	Content      []Content
	CreatedAt    time.Time
	InputTokens  int
	OutputTokens int
}

// Блок контента сообщения
type Content struct {
	Type      string	// "text", "tool_use", "tool_result"
	Text      string
	ID        string
	ToolUseID string
	Name      string
	Input     map[string]interface{}
	IsError   *bool
	ServerName string
	Content    []Content
}

