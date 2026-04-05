package usecase

import (
	"context"
	"time"

	"github.com/erlitx/mcp_server/internal/domain"
)

// Storage is a placeholder interface to mirror your architecture.
// Implementations live in adapters.
type Storage interface{}

// Repository is a placeholder interface to mirror your architecture.
// Implementations live in adapters.
type Repository interface{}

// Postgres is a placeholder interface to mirror your architecture.
// Implementations live in adapters.
type Postgres interface{}

// Clock abstracts time for testability.
type Clock interface {
	Now() time.Time
}

// ClickHouse provides read-only access to ClickHouse.
type ClickHouse interface {
	PingTest(ctx context.Context) error
	// QueryReadOnly executes a read-only query (SELECT/SHOW/DESCRIBE/EXPLAIN) and returns rows.
	QueryReadOnly(ctx context.Context, query string) ([]map[string]any, error)

	// // LoadFromS3ToCH is included to match your desired shape; template keeps it unimplemented.
	// LoadFromS3ToCH(ctx context.Context, load any) error
}

// DBTAdapter provides access to DBT manifest parsing.
type DBTAdapter interface {
	ParseManifest(ctx context.Context) (*domain.DBTManifest, error)
}

// ManifestCache provides cached access to DBT manifest data.
type ManifestCache interface {
	// Get returns the cached manifest. If cache is empty, it will load it first.
	Get(ctx context.Context) (*domain.DBTManifest, error)

	// Warmup loads the manifest into cache.
	Warmup(ctx context.Context) error

	// Refresh reloads the manifest from source and updates the cache.
	Refresh(ctx context.Context) error
}

type UseCase struct {
	storageminio  Storage
	repository    Repository
	postgres      Postgres
	clickhouse    ClickHouse
	clock         Clock
	dbtAdapter    DBTAdapter
	manifestCache ManifestCache
}

func New(s Storage, r Repository, p Postgres, ch ClickHouse, clk Clock, dbt DBTAdapter, cache ManifestCache) *UseCase {
	return &UseCase{
		storageminio:  s,
		repository:    r,
		postgres:      p,
		clickhouse:    ch,
		clock:         clk,
		dbtAdapter:    dbt,
		manifestCache: cache,
	}
}
