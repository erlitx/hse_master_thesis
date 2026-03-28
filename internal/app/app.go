package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/adapter/clickhouse"
	"github.com/erlitx/mcp_server/internal/adapter/clock"
	"github.com/erlitx/mcp_server/internal/adapter/nop"
	"github.com/erlitx/mcp_server/internal/controller/mcpserver"
	"github.com/erlitx/mcp_server/internal/usecase"
	clickhousepkg "github.com/erlitx/mcp_server/pkg/clickhouse"
	"github.com/erlitx/mcp_server/pkg/httpserver"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/go-chi/chi/v5/middleware"

)

type Dependencies struct {
	Clock clock.Clock
}

func Run(ctx context.Context, cfg config.Config) error {
	// --- Build dependencies (adapters) ---
	deps := Dependencies{
		Clock: clock.NewSystemClock(),
	}

	// CLICKHOUSE
	chPool, err := clickhousepkg.New(ctx, cfg.ClickHouse)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init ClickHouse")
	}
	defer chPool.Close()

	clickhouseUc := clickhouse.New(chPool.Conn())

	// --- Usecase layer ---
	uc := usecase.New(
		nop.Storage{},
		nop.Repository{},
		nop.Postgres{},
		clickhouseUc,
		deps.Clock,
	)

	// --- Controller layer (MCP server) ---
	mcpHandler := mcpserver.New(cfg, uc)

	switch cfg.MCP.Transport {
	case "http", "":
		// HTTP router
		router := chi.NewRouter()
		router.Use(middleware.RequestID)
		router.Use(middleware.RealIP)
		router.Use(middleware.Logger)
		router.Use(middleware.Recoverer)

		router.Handle(cfg.MCP.Path, mcpHandler.HTTPHandler())

		router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		httpSrv := httpserver.New(router, cfg.HTTP.Port)

		log.Info().
			Str("port", cfg.HTTP.Port).
			Str("path", cfg.MCP.Path).
			Msg("App started!")

		// STOPPING
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		select {
		case s := <-sig:
			log.Info().Str("signal", s.String()).Msg("App got signal to stop")
		case err := <-httpSrv.Notify():
			if err != nil {
				return fmt.Errorf("app - Run - httpSrv.Notify: %w", err)
			}
		case <-ctx.Done():
			log.Info().Msg("App context canceled, stopping")
		}

		httpSrv.Close()
		log.Info().Msg("App stopped!")

		return nil

	default:
		return fmt.Errorf("unsupported MCP_TRANSPORT=%q (supported: http)", cfg.MCP.Transport)
	}
}