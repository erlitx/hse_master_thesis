package dbt

import (
	"context"
	"fmt"
	"sync"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/rs/zerolog/log"
)

// Кэш manifest DBT
type ManifestCache struct {
	parser *DbtParser
	mu     sync.RWMutex
	cache  *domain.DBTManifest
}

// Создаёт кэш manifest DBT
func NewManifestCache(parser *DbtParser) *ManifestCache {
	return &ManifestCache{
		parser: parser,
	}
}

// Возвращает manifest из кэша
func (c *ManifestCache) Get(ctx context.Context) (*domain.DBTManifest, error) {
	log.Debug().Msg("Getting DBT manifest from cache...")
	c.mu.RLock()
	if c.cache != nil {
		defer c.mu.RUnlock()
		log.Debug().Msg("Cache exists, returning cached manifest")
		return c.cache, nil
	}
	c.mu.RUnlock()

	log.Debug().Msg("Cache is empty, loading manifest from parser")
	return c.load(ctx)
}

// Прогревает кэш manifest
func (c *ManifestCache) Warmup(ctx context.Context) error {
	_, err := c.load(ctx)
	if err != nil {
		return fmt.Errorf("warmup failed: %w", err)
	}
	return nil
}

// Обновляет кэш manifest из источника
func (c *ManifestCache) Refresh(ctx context.Context) error {
	_, err := c.load(ctx)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	return nil
}

// Загружает manifest в кэш
func (c *ManifestCache) load(ctx context.Context) (*domain.DBTManifest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache != nil {
		return c.cache, nil
	}

	manifest, err := c.parser.ParseManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	c.cache = manifest
	return c.cache, nil
}
