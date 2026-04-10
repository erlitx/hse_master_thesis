package http

import (
	"github.com/erlitx/mcp_server/config"
	v1 "github.com/erlitx/mcp_server/internal/mcp_client/controller/http/v1"
	"github.com/erlitx/mcp_server/internal/mcp_client/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func ProfileRouter(cfg config.Config, r *chi.Mux, uc *usecase.UseCase) error {
	v1Handler := v1.New(uc, cfg)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", v1Handler.Health)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Main endpoint to send messages to Claude with MCP resources
			r.Post("/send_message", v1Handler.SendMessage)
		})
	})

	return nil
}
