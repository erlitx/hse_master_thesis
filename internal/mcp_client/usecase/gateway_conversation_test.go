package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/erlitx/mcp_server/internal/mcp_client/adapter/storage/memory"
	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/erlitx/mcp_server/internal/mcp_client/usecase"
)

// --- Claude response templates (named keys for stubClaudeClient.templates) ---

func claudeTemplateToolUseListDWH() *dto.ClaudeResponse {
	return &dto.ClaudeResponse{
		Model:      "claude-opus-4-6",
		ID:         "msg_01GpgKXLKECF9syy6Z4WMyh6",
		Type:       "message",
		Role:       "assistant",
		StopReason: "tool_use",
		Content: []dto.ClaudeContent{
			{Type: "text", Text: "\n\nLet me first explore the available models to understand the data structure."},
			{
				Type:  "tool_use",
				ID:    "toolu_01CvEJC6ZoLAoyiFNpN9EkVV",
				Name:  "list_dwh_models",
				Input: map[string]interface{}{},
			},
		},
		Usage: dto.ClaudeUsage{InputTokens: 939, OutputTokens: 54},
	}
}

func claudeTemplateEndTurnPlain() *dto.ClaudeResponse {
	return &dto.ClaudeResponse{
		Model:      "claude-test",
		Type:       "message",
		Role:       "assistant",
		StopReason: "end_turn",
		Content: []dto.ClaudeContent{
			{Type: "text", Text: "Here is the final reply."},
		},
		Usage: dto.ClaudeUsage{InputTokens: 20, OutputTokens: 8},
	}
}

func claudeTemplateHiGreeting() *dto.ClaudeResponse {
	return &dto.ClaudeResponse{
		Model:      "claude-opus-4-6",
		Type:       "message",
		Role:       "assistant",
		StopReason: "end_turn",
		Content: []dto.ClaudeContent{
			{Type: "text", Text: "Hello, how are you."},
		},
		Usage: dto.ClaudeUsage{InputTokens: 4, OutputTokens: 6},
	}
}

// stubClaudeClient implements usecase.ClaudeClient for tests.
// templates holds named *dto.ClaudeResponse values; SendGatewayMessage picks one by switch logic.
type stubClaudeClient struct {
	templates map[string]*dto.ClaudeResponse
	callN     int
	scenario  string // "tool_then_end_turn" | "hi_greeting"
}

func (s *stubClaudeClient) SendMessage(_ context.Context, _ dto.ClaudeRequest) (*dto.ClaudeResponse, error) {
	return nil, fmt.Errorf("SendMessage not used in gateway tests")
}

func (s *stubClaudeClient) SendGatewayMessage(_ context.Context, _ dto.ManualGatewayRequest) (*dto.ClaudeResponse, error) {
	s.callN++
	switch s.scenario {
	case "tool_then_end_turn":
		switch s.callN {
		case 1:
			return s.templates["tool_use"], nil
		case 2:
			return s.templates["end_turn_plain"], nil
		default:
			return nil, fmt.Errorf("stub Claude: unexpected SendGatewayMessage call #%d", s.callN)
		}
	case "hi_greeting":
		return s.templates["hi_greeting"], nil
	default:
		return nil, fmt.Errorf("stub Claude: unknown scenario %q", s.scenario)
	}
}

func newStubClaudeClient() *stubClaudeClient {
	return &stubClaudeClient{
		scenario: "tool_then_end_turn",
		templates: map[string]*dto.ClaudeResponse{
			"tool_use":       claudeTemplateToolUseListDWH(),
			"end_turn_plain": claudeTemplateEndTurnPlain(),
		},
	}
}

func newStubClaudeClientHiGreeting() *stubClaudeClient {
	return &stubClaudeClient{
		scenario: "hi_greeting",
		templates: map[string]*dto.ClaudeResponse{
			"hi_greeting": claudeTemplateHiGreeting(),
		},
	}
}

// stubMCPServer implements usecase.MCPClient for tests.
type stubMCPServer struct {
	tools          []dto.MCPToolInfo
	callToolResp   map[string]interface{}
	callToolErr    error
	CallToolCount  int
	ListToolsCount int
}

func (s *stubMCPServer) ListResources(context.Context) ([]dto.MCPResource, error) {
	return nil, nil
}

func (s *stubMCPServer) ReadResource(context.Context, string) (*dto.MCPResource, error) {
	return nil, nil
}

func (s *stubMCPServer) ListTools(context.Context) ([]dto.MCPToolInfo, error) {
	s.ListToolsCount++
	return s.tools, nil
}

func (s *stubMCPServer) CallTool(_ context.Context, name string, arguments map[string]interface{}) (map[string]interface{}, error) {
	s.CallToolCount++
	if s.callToolErr != nil {
		return nil, s.callToolErr
	}
	if s.callToolResp != nil {
		return s.callToolResp, nil
	}
	return map[string]interface{}{"tool": name, "args": arguments}, nil
}

// stubListToolsResult mirrors a real MCP tools/list payload (result.tools).
func stubListToolsResult() []dto.MCPToolInfo {
	ann := map[string]interface{}{
		"readOnlyHint":    false,
		"destructiveHint": true,
		"idempotentHint":  false,
		"openWorldHint":   true,
	}
	emptyObjectSchema := map[string]interface{}{
		"properties": map[string]interface{}{},
		"required":   []interface{}{},
		"type":       "object",
	}
	chQuerySchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"description": "SQL query to execute",
				"type":        "string",
			},
		},
		"required": []interface{}{"query"},
		"type":     "object",
	}
	getDWHModelSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"model_name": map[string]interface{}{
				"description": "Name of the DWH model",
				"type":        "string",
			},
		},
		"required": []interface{}{"model_name"},
		"type":     "object",
	}
	return []dto.MCPToolInfo{
		{
			Name:        "ch_ping",
			Description: "Ping ClickHouse using the configured adapter.",
			Annotations: ann,
			InputSchema: emptyObjectSchema,
		},
		{
			Name:        "ch_query",
			Description: "Execute a read-only ClickHouse query (SELECT/SHOW/DESCRIBE/EXPLAIN).",
			Annotations: ann,
			InputSchema: chQuerySchema,
		},
		{
			Name:        "get_dwh_model",
			Description: "Return full DWH model metadata by model name from the cached manifest",
			Annotations: ann,
			InputSchema: getDWHModelSchema,
		},
		{
			Name:        "list_dwh_models",
			Description: "Return all DWH models (tables) names and business descriptions",
			Annotations: ann,
			InputSchema: emptyObjectSchema,
		},
	}
}

func newStubMCPServer() *stubMCPServer {
	return &stubMCPServer{
		tools:        stubListToolsResult(),
		callToolResp: map[string]interface{}{"ok": true, "value": 42},
	}
}

func TestHandleGatewayConversation_toolThenFinalText(t *testing.T) {
	ctx := context.Background()
	mcp := newStubMCPServer()
	claude := newStubClaudeClient()
	repo := memory.New()
	uc := usecase.New(mcp, claude, repo)

	out, err := uc.HandleGatewayConversation(ctx, dto.GatewayConversationInput{
		Message:   "Hello gateway",
		SessionID: nil,
		Model:     "claude-3-test",
		MaxTokens: 1024,
		System:    "You are a test assistant.",
	})
	if err != nil {
		t.Fatalf("HandleGatewayConversation: %v", err)
	}

	if mcp.ListToolsCount != 1 {
		t.Errorf("ListTools calls = %d, want 1", mcp.ListToolsCount)
	}
	if mcp.CallToolCount != 1 {
		t.Errorf("CallTool calls = %d, want 1", mcp.CallToolCount)
	}
	if claude.callN != 2 {
		t.Errorf("SendGatewayMessage calls = %d, want 2", claude.callN)
	}

	if len(out.Tools) != 4 {
		t.Fatalf("session tools count = %d, want 4: %+v", len(out.Tools), out.Tools)
	}
	if out.Tools[3].Name != "list_dwh_models" {
		t.Fatalf("expected list_dwh_models in stub tools, got %+v", out.Tools[3])
	}

	if len(out.Messages) != 4 {
		t.Fatalf("messages count = %d, want 4: %v", len(out.Messages), messageRoles(out.Messages))
	}
	if out.Messages[0].Role != "user" || out.Messages[1].Role != "assistant" ||
		out.Messages[2].Role != "user" || out.Messages[3].Role != "assistant" {
		t.Fatalf("unexpected roles: %v", messageRoles(out.Messages))
	}

	wantTokens := 939 + 54 + 20 + 8
	if out.TotalTokens != wantTokens {
		t.Errorf("TotalTokens = %d, want %d", out.TotalTokens, wantTokens)
	}
	if out.StopReason != "end_turn" {
		t.Errorf("StopReason = %q, want end_turn", out.StopReason)
	}
	if out.Model != "claude-test" {
		t.Errorf("Model = %q, want claude-test", out.Model)
	}

	last := out.Messages[len(out.Messages)-1].Content
	if len(last) != 1 || last[0].Type != "text" || last[0].Text != "Here is the final reply." {
		t.Fatalf("last assistant content: %+v", last)
	}
}

func TestHandleGatewayConversation_simple(t *testing.T) {
	ctx := context.Background()
	mcp := newStubMCPServer()
	claude := newStubClaudeClientHiGreeting()
	repo := memory.New()
	uc := usecase.New(mcp, claude, repo)

	out, err := uc.HandleGatewayConversation(ctx, dto.GatewayConversationInput{
		Message:   "hi",
		SessionID: nil,
		Model:     "claude-3-test",
		MaxTokens: 512,
		System:    "You are helpful.",
	})
	if err != nil {
		t.Fatalf("HandleGatewayConversation: %v", err)
	}

	if mcp.CallToolCount != 0 {
		t.Errorf("CallTool calls = %d, want 0", mcp.CallToolCount)
	}
	if claude.callN != 1 {
		t.Errorf("SendGatewayMessage calls = %d, want 1", claude.callN)
	}

	if len(out.Messages) != 2 {
		t.Fatalf("messages count = %d, want 2: %v", len(out.Messages), messageRoles(out.Messages))
	}
	if out.Messages[1].Role != "assistant" || len(out.Messages[1].Content) != 1 ||
		out.Messages[1].Content[0].Text != "Hello, how are you." {
		t.Fatalf("assistant message: %+v", out.Messages[1])
	}
	if out.StopReason != "end_turn" || out.Model != "claude-opus-4-6" {
		t.Fatalf("session metadata: stop=%q model=%q", out.StopReason, out.Model)
	}
	if out.TotalTokens != 4+6 {
		t.Errorf("TotalTokens = %d, want 10", out.TotalTokens)
	}
}

func messageRoles(msgs []domain.Message) []string {
	r := make([]string, len(msgs))
	for i := range msgs {
		r[i] = msgs[i].Role
	}
	return r
}
