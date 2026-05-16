package v1

import (
	"github.com/erlitx/mcp_server/config"
	"github.com/erlitx/mcp_server/internal/mcp_client/usecase"
)

// HTTP-обработчики REST API
type Handlers struct {
	usecase *usecase.UseCase
	config  config.Config
}

// Создаёт новый экземпляр
func New(uc *usecase.UseCase, cfg config.Config) *Handlers {
	return &Handlers{
		usecase: uc,
		config:  cfg,
	}
}
