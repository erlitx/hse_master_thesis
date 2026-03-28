package old

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/adapter/clickhouse"
	"github.com/erlitx/mcp_server/internal/adapter/clock"
	"github.com/erlitx/mcp_server/internal/adapter/nop"
	"github.com/erlitx/mcp_server/internal/controller/mcpserver"
	"github.com/erlitx/mcp_server/internal/usecase"
	"github.com/rs/zerolog/log"
	clickhousepkg "github.com/erlitx/mcp_server/pkg/clickhouse"
)

type Dependencies struct {
	Clock      clock.Clock
}

func Run(ctx context.Context, c config.Config) error {
	// --- Build dependencies (adapters) ---
	deps := Dependencies{
		Clock: clock.NewSystemClock(),
	}

	// CLICKHOUSE
	chPool, err := clickhousepkg.New(ctx, c.ClickHouse)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init ClickHouse")
	}

	clickhouseUc := clickhouse.New(chPool.Conn())

	// --- Usecase layer (interfaces only; implementations come from adapters) ---
	uc := usecase.New(
		nop.Storage{},
		nop.Repository{},
		nop.Postgres{},
		clickhouseUc,
		deps.Clock,
	)

	// --- Controller layer (MCP server) ---
	h := mcpserver.New(c, uc)

	// --- Run transport ---
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		log.Info().Msg("shutdown signal received")
		cancel()
	}()

	switch c.MCP.Transport {
	case "http", "":
		log.Info().Str("addr", c.MCP.Addr).Str("path", c.MCP.Path).Msg("MCP transport: http")
		mux := http.NewServeMux()
		mux.Handle(c.MCP.Path, h.HTTPHandler())

		httpSrv := &http.Server{Addr: c.MCP.Addr, Handler: mux}
		errCh := make(chan error, 1)
		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()

		select {
		case <-ctx.Done():
			log.Info().Msg("shutting down HTTP server")
			_ = httpSrv.Shutdown(context.Background())
			return nil
		case err := <-errCh:
			return fmt.Errorf("http server: %w", err)
		}
	default:
		return fmt.Errorf("unsupported MCP_TRANSPORT=%q (supported: http)", c.MCP.Transport)
	}
}
