package domain

import "time"

// Message represents a single message in a conversation with Claude
type Message struct {
	ID           string
	SessionID    string
	Role         string // "user" or "assistant"
	Content      string
	CreatedAt    time.Time
	InputTokens  int
	OutputTokens int
}


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
