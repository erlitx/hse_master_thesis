package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)


func (ch *ClickHouse) PingTest(ctx context.Context) error {
	var now time.Time
	if err := ch.conn.QueryRow(ctx, "SELECT now()").Scan(&now); err != nil {
		log.Error().Err(err).Msg("ClickHouse ping query failed")
		return fmt.Errorf("clickhouse ping query failed: %w", err)
	}

	log.Info().Msgf("ClickHouse connection OK, server time: %s", now.Format(time.RFC3339))
	return nil
}

