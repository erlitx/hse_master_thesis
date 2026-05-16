package usecase

import (
	"context"
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
)

// Заглушка хранилища (адаптер)
type Storage interface{}

// Заглушка репозитория (адаптер)
type Repository interface{}

// Заглушка Postgres (адаптер)
type Postgres interface{}

// Абстракция времени для тестов
type Clock interface {
	Now() time.Time
}

// Адаптер ClickHouse
type ClickHouse interface {
	// Проверяет доступность ClickHouse
	PingTest(ctx context.Context) error
	// Выполняет read-only запрос
	QueryReadOnly(ctx context.Context, query string) (*domain.QueryResult, error)
}

// Интерфейс DBTAdapter
type DBTAdapter interface {
	// Парсит manifest DBT
	ParseManifest(ctx context.Context) (*domain.DBTManifest, error)
}

// Кэш manifest DBT
type ManifestCache interface {
	// Возвращает manifest из кэша
	Get(ctx context.Context) (*domain.DBTManifest, error)
	// Прогревает кэш
	Warmup(ctx context.Context) error
	// Обновляет кэш из источника
	Refresh(ctx context.Context) error
}

// Клиент Superset (BI)
type BiTool interface {
	// Создаёт датасет
	CreateDataset(ctx context.Context, input domain.CreateDatasetInput) (*domain.Dataset, error)
	// Создаёт дашборд
	CreateDashboard(ctx context.Context) error
	// Создаёт чарт
	CreateChart(ctx context.Context) error
}

// Слой бизнес-логики
type UseCase struct {
	storageminio  Storage
	repository    Repository
	postgres      Postgres
	clickhouse    ClickHouse
	dbtAdapter    DBTAdapter
	manifestCache ManifestCache
	bitool        BiTool
}

// Создаёт новый экземпляр
func New(s Storage, r Repository, p Postgres, ch ClickHouse, dbt DBTAdapter, cache ManifestCache, bitool BiTool) *UseCase {
	return &UseCase{
		storageminio:  s,
		repository:    r,
		postgres:      p,
		clickhouse:    ch,
		dbtAdapter:    dbt,
		manifestCache: cache,
		bitool:        bitool,
	}
}
