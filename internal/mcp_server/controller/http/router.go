package http

import (
	"github.com/erlitx/mcp_server/config"
	v1 "github.com/erlitx/mcp_server/internal/mcp_server/controller/http/v1"
	"github.com/erlitx/mcp_server/internal/mcp_server/controller/mcpserver"
	"github.com/erlitx/mcp_server/internal/mcp_server/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Регистрирует HTTP-маршруты
func ProfileRouter(cfg config.Config, r *chi.Mux, uc *usecase.UseCase) error {
	v1Handler := v1.New(uc)
	mcpHandler := mcpserver.New(cfg, uc)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", v1Handler.Health)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Process MCP JSON-RPC request to find registered tools/resources/prompts and execute them
			r.Handle("/", mcpHandler.HTTPHandler())
			r.Handle("/mcp", mcpHandler.HTTPHandler())

			// DBT manifest endpoint
			r.Get("/dbt/manifest", v1Handler.GetManifest)
			r.Post("/bitool/create_dataset", v1Handler.CreateDataset)
		})
	})

	return nil
}
