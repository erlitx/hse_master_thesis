package usecase

import (
	"context"
	"fmt"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/erlitx/mcp_server/internal/mcp_server/service"
)

// Проверяет доступность ClickHouse
func (u *UseCase) ClickHousePing(ctx context.Context) error {
	if u.clickhouse == nil {
		return fmt.Errorf("clickhouse adapter is not configured")
	}
	return u.clickhouse.PingTest(ctx)
}

// Выполняет read-only запрос к ClickHouse
func (u *UseCase) ClickHouseQueryReadOnly(ctx context.Context, query string) (*domain.QueryResult, error) {
	if u.clickhouse == nil {
		return nil, fmt.Errorf("clickhouse adapter is not configured")
	}
	if err := service.ValidateReadOnlySQL(query); err != nil {
		return nil, err
	}

	return u.clickhouse.QueryReadOnly(ctx, query)
}
