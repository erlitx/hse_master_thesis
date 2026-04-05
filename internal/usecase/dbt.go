package usecase

import (
	"context"
	"fmt"

	"github.com/erlitx/mcp_server/internal/domain"
)

// GetDBTManifest retrieves the cached DBT manifest
func (uc *UseCase) GetDBTManifest(ctx context.Context) (*domain.DBTManifest, error) {
	manifest, err := uc.manifestCache.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase - GetDBTManifest - uc.manifestCache.Get: %w", err)
	}

	return manifest, nil
}
