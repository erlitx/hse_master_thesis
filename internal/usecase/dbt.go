package usecase

import (
	"context"
	"fmt"

	"github.com/erlitx/mcp_server/internal/domain"
)

// GetDBTManifest retrieves and parses the DBT manifest
func (uc *UseCase) GetDBTManifest(ctx context.Context) (*domain.DBTManifest, error) {
	manifest, err := uc.dbtAdapter.ParseManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase - GetDBTManifest - uc.dbtAdapter.ParseManifest: %w", err)
	}

	return manifest, nil
}
