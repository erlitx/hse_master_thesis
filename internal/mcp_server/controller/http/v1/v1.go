package v1

import "github.com/erlitx/mcp_server/internal/mcp_server/usecase"

// HTTP-обработчики REST API
type Handlers struct {
	usecase *usecase.UseCase
}

// Создаёт новый экземпляр
func New(uc *usecase.UseCase) *Handlers {
	return &Handlers{
		usecase: uc,
	}
}
