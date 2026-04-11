package v1

import (
	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_client/usecase"
)

type Handlers struct {
	usecase *usecase.UseCase
	config  config.Config
}

func New(uc *usecase.UseCase, cfg config.Config) *Handlers {
	return &Handlers{
		usecase: uc,
		config:  cfg,
	}
}
