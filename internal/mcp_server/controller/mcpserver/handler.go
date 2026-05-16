package mcpserver

import (
	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/erlitx/mcp_server/internal/mcp_server/usecase"
	"github.com/mark3labs/mcp-go/server"
)

// Обработчик MCP: связывает протокол со слоем usecase
type Handler struct {
	cfg config.Config
	uc  *usecase.UseCase
	srv *server.MCPServer
	modelByName map[string]domain.DBTModel
}

// Создаёт новый экземпляр
func New(cfg config.Config, uc *usecase.UseCase) *Handler {
	s := server.NewMCPServer(
		cfg.MCPServer.Name,
		cfg.MCPServer.Version,
		server.WithLogging(),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)
	
	h := &Handler{cfg: cfg, uc: uc, srv: s}
	h.registerTools()
	h.registerResources()
	h.registerPrompts()

	return h
}

// Регистрирует MCP-инструменты
func (h *Handler) registerTools() {
	h.registerClickHouseTools()
	h.registerManifestTools()
}

// Регистрирует MCP-ресурсы
func (h *Handler) registerResources() {
	h.registerManifestResources()
}

// Регистрирует MCP-промпты
func (h *Handler) registerPrompts() {
	h.registerGreetingPrompt()
}

// Возвращает экземпляр MCP-сервера
func (h *Handler) Server() *server.MCPServer { return h.srv }
