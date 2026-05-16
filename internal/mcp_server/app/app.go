package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_server/adapter/clickhouse"
	"github.com/erlitx/mcp_server/internal/mcp_server/adapter/dbt"
	"github.com/erlitx/mcp_server/internal/mcp_server/adapter/nop"
	"github.com/erlitx/mcp_server/internal/mcp_server/adapter/superset"
	httpcontroller "github.com/erlitx/mcp_server/internal/mcp_server/controller/http"
	"github.com/erlitx/mcp_server/internal/mcp_server/usecase"
	clickhousepkg "github.com/erlitx/mcp_server/pkg/clickhouse"
	"github.com/erlitx/mcp_server/pkg/httpserver"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// Запускает приложение
func Run(ctx context.Context, cfg config.Config) error {
	chPool, err := clickhousepkg.New(ctx, cfg.ClickHouse)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init ClickHouse")
	}
	defer chPool.Close()

	clickhouseUc := clickhouse.New(chPool.Conn())

	dbtAdapter := dbt.New(cfg.DBT.ManifestPath)
	dbtCache := dbt.NewManifestCache(dbtAdapter)
	bitoolAdapter := superset.New(cfg.Superset)

	log.Info().Str("path", cfg.DBT.ManifestPath).Msg("Warming up DBT manifest cache...")
	if err := dbtCache.Warmup(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to warmup DBT manifest cache")
	}
	log.Info().Msg("DBT manifest cache warmed up successfully")

	// --- Usecase layer ---
	uc := usecase.New(
		nop.Storage{},
		nop.Repository{},
		nop.Postgres{},
		clickhouseUc,
		dbtAdapter,
		dbtCache,
		bitoolAdapter,
	)

	router := chi.NewRouter()
	err = httpcontroller.ProfileRouter(cfg, router, uc)
	if err != nil {
		return fmt.Errorf("failed http router profile: %w", err)
	}

	httpSrv := httpserver.New(router, cfg.HTTP.ServerPort)

	log.Info().Msg("App started!")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

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

	//nolint:contextcheck
	httpSrv.Close()
	if err := clickhouseUc.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close ClickHouse connection")
	}

	log.Info().Msg("App stopped!")

	return nil
}
