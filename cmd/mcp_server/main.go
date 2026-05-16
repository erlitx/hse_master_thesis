package main

import (
	"context"
	"os"

	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_server/app"
	"github.com/erlitx/mcp_server/pkg/logger"
	"github.com/rs/zerolog/log"
)

// Точка входа приложения
func main() {
	logger.Init(logger.Config{
		AppName:       "MCP_SERVER",
		AppVersion:    "v0.1.0",
		PrettyConsole: true,
		Level:         "debug", // force
		Env:           os.Getenv("APP_ENV"),
	})

	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("config.New")
	}

	log.Info().Msgf("Starting %s environment", cfg.MCPServer.Env)

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatal().Err(err).Msg("app.Run")
	}
}
