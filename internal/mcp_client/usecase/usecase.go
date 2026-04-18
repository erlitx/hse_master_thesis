package usecase

import (
	"context"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
)

// MCPClient provides access to the MCP server to fetch resources
type MCPClient interface {
	// ListResources returns all available resources from the MCP server
	ListResources(ctx context.Context) ([]dto.MCPResource, error)
	// ReadResource fetches the content of a specific resource
	ReadResource(ctx context.Context, uri string) (*dto.MCPResource, error)
	// ListTools returns all available tools from the MCP server
	ListTools(ctx context.Context) ([]dto.MCPToolInfo, error)
	// CallTool executes a specific MCP tool with arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (map[string]interface{}, error)
}

// ClaudeClient provides access to Claude API
type ClaudeClient interface {
	// SendMessage sends a message to Claude with context from MCP resources
	SendMessage(ctx context.Context, req dto.ClaudeRequest) (*dto.ClaudeResponse, error)
	// SendGatewayMessage sends a request to Claude with Claude-native tools format
	SendGatewayMessage(ctx context.Context, req dto.ManualGatewayRequest) (*dto.ClaudeResponse, error)
}

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	// CreateSession creates a new session in the repository
	CreateSession(ctx context.Context, session *domain.Session) error

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)

	// SaveSession updates an existing session
	SaveSession(ctx context.Context, session *domain.Session) error
}

type UseCase struct {
	mcpClient    MCPClient
	claudeClient ClaudeClient
	sessionRepo  SessionRepository
}

func New(mcp MCPClient, claude ClaudeClient, sessionRepo SessionRepository) *UseCase {
	return &UseCase{
		mcpClient:    mcp,
		claudeClient: claude,
		sessionRepo:  sessionRepo,
	}
}
