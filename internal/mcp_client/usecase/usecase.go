package usecase

import (
	"context"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
)

// Метаданные MCP-сервера
type MCPServer interface {
	// Список ресурсов MCP
	ListResources(ctx context.Context) ([]dto.MCPResource, error)
	// Чтение ресурса по URI
	ReadResource(ctx context.Context, uri string) (*dto.MCPResource, error)
	// Список инструментов MCP
	ListTools(ctx context.Context) ([]dto.MCPToolInfo, error)
	// Вызов MCP-инструмента
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (map[string]interface{}, error)
}

// Интерфейс ClaudeClient
type ClaudeClient interface {
	// Отправка сообщения в Claude
	SendMessage(ctx context.Context, req dto.ClaudeRequest) (*dto.ClaudeResponse, error)
	// Отправка gateway-запроса в Claude
	SendGatewayMessage(ctx context.Context, req dto.ManualGatewayRequest) (*dto.ClaudeResponse, error)
}

// Хранилище сессий
type SessionRepository interface {
	// Создаёт сессию
	CreateSession(ctx context.Context, session *domain.Session) error
	// Загружает сессию по ID
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	// Сохраняет сессию
	SaveSession(ctx context.Context, session *domain.Session) error
}

// Слой бизнес-логики
type UseCase struct {
	MCPServer    MCPServer
	claudeClient ClaudeClient
	sessionRepo  SessionRepository
}

// Создаёт новый экземпляр
func New(mcp MCPServer, claude ClaudeClient, sessionRepo SessionRepository) *UseCase {
	return &UseCase{
		MCPServer:    mcp,
		claudeClient: claude,
		sessionRepo:  sessionRepo,
	}
}
