package domain

import "time"


// Session represents a conversation session with Claude
type Session struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Messages    []Message
	StopReason  string // "end_turn", "max_tokens", "stop_sequence", etc.
	Model       string
	TotalTokens int
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

