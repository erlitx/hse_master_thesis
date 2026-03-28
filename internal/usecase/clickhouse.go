package usecase

import (
	"context"
	"fmt"
	"strings"
)

func (u *UseCase) ClickHousePing(ctx context.Context) error {
	if u.clickhouse == nil {
		return fmt.Errorf("clickhouse adapter is not configured")
	}
	return u.clickhouse.PingTest(ctx)
}

func (u *UseCase) ClickHouseQueryReadOnly(ctx context.Context, query string) ([]map[string]any, error) {
	if u.clickhouse == nil {
		return nil, fmt.Errorf("clickhouse adapter is not configured")
	}
	q := strings.TrimSpace(strings.ToUpper(query))
	// A simple safety gate (template-level) for read-only.
	allowed := []string{"SELECT", "SHOW", "DESCRIBE", "EXPLAIN"}
	ok := false
	for _, a := range allowed {
		if strings.HasPrefix(q, a) {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("only read-only queries are allowed (SELECT/SHOW/DESCRIBE/EXPLAIN)")
	}

	return u.clickhouse.QueryReadOnly(ctx, query)
}
