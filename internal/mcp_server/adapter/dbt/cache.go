package dbt

import (
	"context"
	"fmt"
	"sync"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/rs/zerolog/log"
)

// ManifestCache implements usecase.ManifestCache interface.
// It provides thread-safe cached access to parsed DBT manifest data.
type ManifestCache struct {
	parser *DbtParser
	mu     sync.RWMutex
	cache  *domain.DBTManifest
}

// NewManifestCache creates a new cache instance that uses the provided parser.
func NewManifestCache(parser *DbtParser) *ManifestCache {
	return &ManifestCache{
		parser: parser,
	}
}

// Get returns the cached manifest. If cache is empty, it loads it first.
func (c *ManifestCache) Get(ctx context.Context) (*domain.DBTManifest, error) {
	log.Debug().Msg("Getting DBT manifest from cache...")
	c.mu.RLock()
	if c.cache != nil {
		defer c.mu.RUnlock()
		log.Debug().Msg("Cache exists, returning cached manifest")
		return c.cache, nil
	}
	c.mu.RUnlock()

	// Cache is empty, need to load
	log.Debug().Msg("Cache is empty, loading manifest from parser")
	return c.load(ctx)
}

// Warmup loads the manifest into cache during application startup.
func (c *ManifestCache) Warmup(ctx context.Context) error {
	_, err := c.load(ctx)
	if err != nil {
		return fmt.Errorf("warmup failed: %w", err)
	}
	return nil
}

// Refresh reloads the manifest from source and updates the cache.
func (c *ManifestCache) Refresh(ctx context.Context) error {
	_, err := c.load(ctx)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	return nil
}

// load parses the manifest and stores it in cache (thread-safe).
func (c *ManifestCache) load(ctx context.Context) (*domain.DBTManifest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine might have loaded while we waited for lock
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
