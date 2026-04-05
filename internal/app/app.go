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
	"github.com/erlitx/mcp_server/internal/adapter/dbt"
	"github.com/erlitx/mcp_server/internal/adapter/nop"
	httpcontroller "github.com/erlitx/mcp_server/internal/controller/http"
	"github.com/erlitx/mcp_server/internal/usecase"
	clickhousepkg "github.com/erlitx/mcp_server/pkg/clickhouse"
	"github.com/erlitx/mcp_server/pkg/httpserver"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
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

	// DBT
	dbtAdapter := dbt.New("/home/db_admin/Projects/Centaur/DWH/Source/dwh_dbt/centaur_dwh/target/manifest.json")
	dbtCache := dbt.NewManifestCache(dbtAdapter)

	// Warmup cache on startup
	log.Info().Msg("Warming up DBT manifest cache...")
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
		deps.Clock,
		dbtAdapter,
		dbtCache,
	)

	router := chi.NewRouter()
	err = httpcontroller.ProfileRouter(cfg, router, uc)
	if err != nil {
		return fmt.Errorf("failed http router profile: %w", err)
	}

	httpSrv := httpserver.New(router, cfg.HTTP.Port)

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
