package v1

import "github.com/erlitx/mcp_server/internal/mcp_server/usecase"

type Handlers struct {
	usecase *usecase.UseCase
}

func New(uc *usecase.UseCase) *Handlers {
	return &Handlers{
		usecase: uc,
	}
}
