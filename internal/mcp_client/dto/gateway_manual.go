package dto

// GatewayConversationRequest is the request payload for manual gateway orchestration.
// It mirrors ConversationRequest but is dedicated to the /gateway flow.
type GatewayConversationRequest struct {
	Message   string  `json:"message" validate:"required"`
	SessionID *string `json:"session_id,omitempty"`
	Model     string  `json:"model,omitempty"`
	MaxTokens int     `json:"max_tokens,omitempty"`
	System    string  `json:"system,omitempty"`
}

// ClaudeToolDefinition describes a Claude-native tool in manual mode.
type ClaudeToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ManualGatewayRequest is the Claude request shape used by manual /gateway flow.
// It intentionally uses Claude-native tool definitions (not mcp_servers).
type ManualGatewayRequest struct {
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	System    string                 `json:"system,omitempty"`
	Messages  []ClaudeMessage        `json:"messages"`
	Tools     []ClaudeToolDefinition `json:"tools,omitempty"`
}

// ToolUseContent represents assistant tool invocation block returned by Claude.
// For manual tool orchestration the type is expected to be "tool_use".
type ToolUseContent struct {
	Type  string                 `json:"type"` // "tool_use"
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResultContent is sent back to Claude as a result for a specific tool_use id.
// For manual tool orchestration the type is expected to be "tool_result".
type ToolResultContent struct {
	Type      string          `json:"type"` // "tool_result"
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error,omitempty"`
	Content   []ClaudeContent `json:"content"`
}

// MCPToolCallRequest is the JSON-RPC params payload for tools/call.
type MCPToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// MCPToolInfo is a normalized tool descriptor returned from MCP tools/list.
type MCPToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	// MCP tools/list returns this field as "inputSchema" (camelCase).
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}
