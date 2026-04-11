package mcpserver

import (
	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/erlitx/mcp_server/internal/mcp_server/usecase"
	"github.com/mark3labs/mcp-go/server"
)

// Handler wires MCP protocol handlers to the usecase layer.
// All registration methods (tools/resources/prompts) are methods on this struct.
type Handler struct {
	cfg config.Config
	uc  *usecase.UseCase
	srv *server.MCPServer
	modelByName map[string]domain.DBTModel
}

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

func (h *Handler) registerTools() {
	h.registerClickHouseTools()
	h.registerManifestTools()
}

func (h *Handler) registerResources() {
	h.registerTimeResources()
	h.registerManifestResources()
}

func (h *Handler) registerPrompts() {
	h.registerGreetingPrompt()
}

func (h *Handler) Server() *server.MCPServer { return h.srv }
