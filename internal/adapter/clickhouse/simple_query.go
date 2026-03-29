package clickhouse

import (
	"context"

	"github.com/rs/zerolog/log"
)

func (ch *ClickHouse) QueryReadOnly(ctx context.Context, query string) ([]map[string]any, error) {
	log.Info().Msgf("TEST QUERY: %s", query)
	return nil, nil
}
