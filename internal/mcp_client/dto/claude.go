package dto

// Запрос к Claude API
type ClaudeRequest struct {
	Model      string           `json:"model"`
	MaxTokens  int              `json:"max_tokens"`
	Messages   []ClaudeMessage  `json:"messages"`
	System     string           `json:"system,omitempty"`
	MCPServers []MCPServerInfo  `json:"mcp_servers,omitempty"`
	Tools      []ClaudeToolInfo `json:"tools,omitempty"`
}

// Описание MCP-сервера для Claude
type MCPServerInfo struct {
	Type string `json:"type"` // e.g., "url"
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Инструмент в запросе Claude
type ClaudeToolInfo struct {
	Type          string `json:"type"` // e.g., "mcp_toolset"
	MCPServerName string `json:"mcp_server_name"`
}

// Сообщение Claude
type ClaudeMessage struct {
	Role    string      `json:"role"` // "user" or "assistant"
	Content interface{} `json:"content"` // string for user messages, []ClaudeContent for assistant messages
}

// Ответ Claude API
type ClaudeResponse struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      []ClaudeContent `json:"content"`
	Model        string          `json:"model"`
	StopReason   string          `json:"stop_reason"`
	StopSequence *string         `json:"stop_sequence"`
	Usage        ClaudeUsage     `json:"usage"`
}

// Блок контента Claude
type ClaudeContent struct {
	Type       string                 `json:"type"` // "text", "mcp_tool_use", "mcp_tool_result"
	Text       string                 `json:"text,omitempty"`
	ID         string                 `json:"id,omitempty"`          // for tool_use
	Name       string                 `json:"name,omitempty"`        // for tool_use
	Input      map[string]interface{} `json:"input,omitempty"`       // for tool_use
	ServerName string                 `json:"server_name,omitempty"` // for tool_use
	ToolUseID  string                 `json:"tool_use_id,omitempty"` // for tool_result
	IsError    *bool                  `json:"is_error,omitempty"`    // for tool_result
	Content    []ClaudeContent        `json:"content,omitempty"`     // for tool_result (nested content)
}

// Использование токенов Claude
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
