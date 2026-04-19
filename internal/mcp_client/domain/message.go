package domain

import "time"

// Session represents a conversation session with Claude
type Session struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Messages    []Message
	// Tools holds MCP tool definitions (from tools/list) for gateway flows; converted to API DTOs when calling Claude.
	Tools       []ToolDefinition
	// GatewayModel, GatewayMaxTokens, GatewaySystem are the caller-selected Claude request options for the manual gateway flow.
	GatewayModel     string
	GatewayMaxTokens int
	GatewaySystem    string
	StopReason       string // "end_turn", "max_tokens", "stop_sequence", etc.
	Model            string
	TotalTokens      int
}

// ToolDefinition is a normalized MCP tool descriptor stored on the session transcript model.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// Message represents a single message in a conversation with Claude
type Message struct {
	ID           string
	SessionID    string
	Role         string // "user" or "assistant"
	Content      []Content
	CreatedAt    time.Time
	InputTokens  int
	OutputTokens int
}

type Content struct {
	Type      string
	Text      string
	ID        string
	ToolUseID string
	Name      string
	Input     map[string]interface{}
	IsError   *bool
	ServerName string
	Content    []Content
}

