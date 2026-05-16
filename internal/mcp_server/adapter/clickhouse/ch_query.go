package clickhouse

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const maxQueryRows = 1000

// Выполняет read-only SELECT в ClickHouse
func (ch *ClickHouse) QueryReadOnly(ctx context.Context, query string) (*domain.QueryResult, error) {
	rows, err := ch.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("clickhouse - QueryReadOnly - conn.Query: %w", err)
	}
	defer rows.Close()

	columns := rows.Columns()
	columnTypes := rows.ColumnTypes()

	scanDest := make([]any, len(columnTypes))
	for i, ct := range columnTypes {
		scanDest[i] = reflect.New(ct.ScanType()).Interface()
	}

	result := &domain.QueryResult{
		Columns: columns,
		Rows:    make([]map[string]any, 0),
	}

	for rows.Next() {
		if len(result.Rows) >= maxQueryRows {
			result.Truncated = true
			break
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("clickhouse - QueryReadOnly - rows.Scan: %w", err)
		}

		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = normalizeCellValue(reflect.ValueOf(scanDest[i]).Elem().Interface())
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("clickhouse - QueryReadOnly - rows.Err: %w", err)
	}

	result.RowCount = len(result.Rows)
	log.Info().Msgf("ClickHouse query result: %+v", result)
	return result, nil
}

// Нормализует значение ячейки для JSON
func normalizeCellValue(v any) any {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch val := rv.Interface().(type) {
	case time.Time:
		return val.Format(time.RFC3339Nano)
	case []byte:
		return string(val)
	case uuid.UUID:
		return val.String()
	case fmt.Stringer:
		return val.String()
	default:
		return rv.Interface()
	}
}
