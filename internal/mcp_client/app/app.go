package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_client/adapter/claude"
	"github.com/erlitx/mcp_server/internal/mcp_client/adapter/mcpclient"
	httpcontroller "github.com/erlitx/mcp_server/internal/mcp_client/controller/http"
	"github.com/erlitx/mcp_server/internal/mcp_client/usecase"
	"github.com/erlitx/mcp_server/pkg/httpserver"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context, cfg config.Config) error {
	// --- Build adapters ---

	// MCP Client (to communicate with mcp_server)
	mcpClient, err := mcpclient.New(cfg.MCPServerConnection.Addr)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer mcpClient.Close()

	// Claude API Client
	claudeClient := claude.New(cfg.Claude.APIKey)

	// --- Usecase layer ---
	uc := usecase.New(mcpClient, claudeClient)

	// --- HTTP Router ---
	router := chi.NewRouter()
	err = httpcontroller.ProfileRouter(cfg, router, uc)
	if err != nil {
		return fmt.Errorf("failed to profile HTTP router: %w", err)
	}

	// --- HTTP Server ---
	httpSrv := httpserver.New(router, cfg.HTTP.ClientPort)

	log.Info().Msgf("MCP Client started on port %s!", cfg.HTTP.ClientPort)

	// --- Signal handling ---
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	select {
	case s := <-sig:
		log.Info().Str("signal", s.String()).Msg("MCP Client got signal to stop")
	case err := <-httpSrv.Notify():
		if err != nil {
			return fmt.Errorf("app - Run - httpSrv.Notify: %w", err)
		}
	case <-ctx.Done():
		log.Info().Msg("MCP Client context canceled, stopping")
	}

	// --- Graceful shutdown ---
	httpSrv.Close()

	log.Info().Msg("MCP Client stopped!")

	return nil
}
