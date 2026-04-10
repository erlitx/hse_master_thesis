package dto

// SendMessageRequest represents the HTTP API request for sending a message to Claude
type SendMessageRequest struct {
	Message string `json:"message" validate:"required"`
}

// SendMessageResponse represents the HTTP API response from Claude
type SendMessageResponse struct {
	Response      string         `json:"response"`
	ResourcesUsed []ResourceInfo `json:"resources_used"`
	Model         string         `json:"model"`
	Usage         *UsageInfo     `json:"usage,omitempty"`
}

// ResourceInfo contains metadata about MCP resources used in the request
type ResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UsageInfo contains token usage information from Claude API
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
