package usecase

import (
	"context"
	"fmt"

	"github.com/erlitx/mcp_server/internal/domain"
)

// ManifestCache returns the manifest cache instance.
func (uc *UseCase) GetManifestCache(ctx context.Context) (*domain.DBTManifest, error) {
	manifest, err := uc.manifestCache.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest from cache: %w", err)
	}
	return manifest, nil
}
