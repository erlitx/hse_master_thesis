package usecase

import (
	"context"

	"github.com/erlitx/mcp_server/internal/dto"
)

// MCPClient provides access to the MCP server to fetch resources
type MCPClient interface {
	// ListResources returns all available resources from the MCP server
	ListResources(ctx context.Context) ([]dto.MCPResource, error)
	// ReadResource fetches the content of a specific resource
	ReadResource(ctx context.Context, uri string) (*dto.MCPResource, error)
}

// ClaudeClient provides access to Claude API
type ClaudeClient interface {
	// SendMessage sends a message to Claude with context from MCP resources
	SendMessage(ctx context.Context, req dto.ClaudeRequest) (*dto.ClaudeResponse, error)
}

type UseCase struct {
	mcpClient    MCPClient
	claudeClient ClaudeClient
}

func New(mcp MCPClient, claude ClaudeClient) *UseCase {
	return &UseCase{
		mcpClient:    mcp,
		claudeClient: claude,
	}
}
