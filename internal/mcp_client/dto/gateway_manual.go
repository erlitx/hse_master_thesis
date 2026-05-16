package dto

import "github.com/erlitx/mcp_server/internal/mcp_client/domain"

// HTTP-запрос gateway-диалога
type GatewayConversationRequest struct {
	Message   string  `json:"message" validate:"required"`
	SessionID *string `json:"session_id,omitempty"`
	Model     string  `json:"model,omitempty"`
	MaxTokens int     `json:"max_tokens,omitempty"`
	System    string  `json:"system,omitempty"`
}

// Вход gateway-диалога
type GatewayConversationInput struct {
	Message   string
	SessionID *string
	Model     string
	MaxTokens int
	System    string
}

// Описание инструмента для Claude
type ClaudeToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// Запрос Claude с инструментами
type ManualGatewayRequest struct {
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	System    string                 `json:"system,omitempty"`
	Messages  []ClaudeMessage        `json:"messages"`
	Tools     []ClaudeToolDefinition `json:"tools,omitempty"`
}

// Собирает gateway-запрос из сессии
func ManualGatewayRequestFromSession(s *domain.Session) ManualGatewayRequest {
	if s == nil {
		return ManualGatewayRequest{}
	}
	return ManualGatewayRequest{
		Model:     s.GatewayModel,
		MaxTokens: s.GatewayMaxTokens,
		System:    s.GatewaySystem,
		Messages:  FromDomainMessages(s.Messages),
		Tools:     ClaudeToolDefinitionsFromDomain(s.Tools),
	}
}

// Блок tool_use
type ToolUseContent struct {
	Type  string                 `json:"type"` // "tool_use"
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// Блок tool_result
type ToolResultContent struct {
	Type      string          `json:"type"` // "tool_result"
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error,omitempty"`
	Content   []ClaudeContent `json:"content"`
}

// Запрос вызова MCP-инструмента
type MCPToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// Метаданные MCP-инструмента
type MCPToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// Конвертирует MCP-инструменты в domain
func DomainToolsFromMCPTools(tools []MCPToolInfo) []domain.ToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	out := make([]domain.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		out = append(out, domain.ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	return out
}

// Конвертирует инструменты в формат Claude
func ClaudeToolDefinitionsFromDomain(tools []domain.ToolDefinition) []ClaudeToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	out := make([]ClaudeToolDefinition, 0, len(tools))
	for _, t := range tools {
		out = append(out, ClaudeToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	return out
}
